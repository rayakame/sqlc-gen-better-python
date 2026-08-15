package transform

import (
	"reflect"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/sqllex"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
)

func TestRewriteSQLiteSQL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sql  string
		want string
	}{
		{name: "no placeholders", sql: "SELECT 1", want: "SELECT 1"},
		{name: "bare placeholders are left alone", sql: "WHERE a = ? AND b = ?", want: "WHERE a = ? AND b = ?"},
		{name: "numbered placeholders lose their index", sql: "WHERE a = ?1 AND b = ?2", want: "WHERE a = ? AND b = ?"},
		{name: "multi digit index", sql: "WHERE a = ?12", want: "WHERE a = ?"},
		{name: "reused index becomes two slots", sql: "WHERE a = ?1 AND b = ?1", want: "WHERE a = ? AND b = ?"},
		{
			name: "slice markers keep their bare placeholder",
			sql:  "WHERE id IN (/*SLICE:ids*/?) AND a = ?2",
			want: "WHERE id IN (/*SLICE:ids*/?) AND a = ?",
		},
		{
			// A "?N" inside a literal or comment is text, not a placeholder.
			name: "quoted and commented text is untouched",
			sql:  "WHERE a = '?1' AND b = ?1 -- ?2\nAND c = `?3`",
			want: "WHERE a = '?1' AND b = ? -- ?2\nAND c = `?3`",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := rewriteSQLiteSQL(tc.sql); got != tc.want {
				t.Errorf("rewriteSQLiteSQL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestOrderParamsBySlot(t *testing.T) {
	t.Parallel()
	param := func(name string, number int32) model.QueryValue {
		return model.QueryValue{Name: name, Number: number, Type: model.PyType{Type: "str"}}
	}
	slice := func(name string, number int32) model.QueryValue {
		return model.QueryValue{
			Name:   name,
			Number: number,
			Type:   model.PyType{Type: types.Int, IsList: true, SqlcSliceName: name},
		}
	}
	repeat := func(qv model.QueryValue) model.QueryValue {
		qv.Repeated = true

		return qv
	}
	cases := []struct {
		name   string
		sql    string
		params []model.QueryValue
		want   []model.QueryValue
	}{
		{
			// sqlc sorts a numbered query's parameters by index, which is not
			// the text order once an explicit index sits after a marker.
			name:   "slice before a numbered argument",
			sql:    "WHERE id IN (/*SLICE:ids*/?) AND a = ?2",
			params: []model.QueryValue{slice("ids", 1), param("x", 2)},
			want:   []model.QueryValue{slice("ids", 1), param("x", 2)},
		},
		{
			name:   "numbered argument before the slice",
			sql:    "WHERE a = ?1 AND id IN (/*SLICE:ids*/?)",
			params: []model.QueryValue{param("x", 1), slice("ids", 2)},
			want:   []model.QueryValue{param("x", 1), slice("ids", 2)},
		},
		{
			// The index sqlc assigned puts x first, but the marker binds first.
			name:   "explicit index after a marker follows text order",
			sql:    "WHERE id IN (/*SLICE:ids*/?) AND a = ?1",
			params: []model.QueryValue{param("x", 1), slice("ids", 2)},
			want:   []model.QueryValue{slice("ids", 2), param("x", 1)},
		},
		{
			// The reused argument is one parameter but two bind slots; the
			// repeat keeps the slot and drops out of the signature.
			name:   "reused argument binds at every slot",
			sql:    "WHERE id IN (/*SLICE:ids*/?) AND a = ?2 AND b = ?3 AND c = ?2",
			params: []model.QueryValue{slice("ids", 1), param("x", 2), param("y", 3)},
			want: []model.QueryValue{
				slice("ids", 1),
				param("x", 2),
				param("y", 3),
				repeat(param("x", 2)),
			},
		},
		{
			name:   "reused marker is emitted once",
			sql:    "WHERE id IN (/*SLICE:ids*/?) AND a = ?2 AND id IN (/*SLICE:ids*/?)",
			params: []model.QueryValue{slice("ids", 1), param("x", 2)},
			want:   []model.QueryValue{slice("ids", 1), param("x", 2)},
		},
		{
			// sqlc numbers every placeholder of a query that has a named
			// argument, so a bare one is a shape this does not know.
			name:   "unnumbered plain slot is rejected",
			sql:    "WHERE id IN (/*SLICE:ids*/?) AND a = ?2 AND b = ?",
			params: []model.QueryValue{slice("ids", 1), param("x", 2), param("y", 3)},
		},
		{
			// sqlc aliases a slice and an argument sharing a name into one
			// parameter, leaving a slot nothing can bind.
			name:   "slot resolving to a slice is rejected",
			sql:    "WHERE id IN (/*SLICE:v*/?) AND a = ?1",
			params: []model.QueryValue{slice("v", 1)},
		},
		{
			// sqlc dropped the parameter the slot names: emitting an argument
			// for it is not an option.
			name:   "slot without a parameter is rejected",
			sql:    "WHERE id IN (/*SLICE:ids*/?) AND a = ?3",
			params: []model.QueryValue{slice("ids", 1), param("x", 2)},
		},
		{
			name:   "parameter binding nowhere is rejected",
			sql:    "WHERE id IN (/*SLICE:ids*/?) AND a = ?2",
			params: []model.QueryValue{slice("ids", 1), param("x", 2), param("ghost", 3)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := orderParamsBySlot(tc.params, sqllex.Slots(tc.sql, sqllex.SQLite))
			if ok != (tc.want != nil) {
				t.Fatalf("orderParamsBySlot() ok = %v, want %v", ok, tc.want != nil)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("orderParamsBySlot() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestOrderColumnsBySlot(t *testing.T) {
	t.Parallel()
	column := func(name string, number int32) model.Column {
		return model.Column{Name: name, Number: number, Type: model.PyType{Type: "str"}}
	}
	slice := func(name string, number int32) model.Column {
		return model.Column{
			Name:   name,
			Number: number,
			Type:   model.PyType{Type: types.Int, IsList: true, SqlcSliceName: name},
		}
	}
	cases := []struct {
		name    string
		sql     string
		columns []model.Column
		want    []model.Column
	}{
		{
			// The class is expanded field by field, so the marker's fields
			// have to come first even though sqlc numbered them second.
			name:    "explicit index after a marker follows text order",
			sql:     "WHERE id IN (/*SLICE:ids*/?) AND a = ?1",
			columns: []model.Column{column("x", 1), slice("ids", 2)},
			want:    []model.Column{slice("ids", 2), column("x", 1)},
		},
		{
			// The repeat keeps its binding slot; the class emits one field.
			name:    "reused argument binds at every slot",
			sql:     "WHERE id IN (/*SLICE:ids*/?) AND a = ?2 AND b = ?3 AND c = ?2",
			columns: []model.Column{slice("ids", 1), column("x", 2), column("y", 3)},
			want: []model.Column{
				slice("ids", 1),
				column("x", 2),
				column("y", 3),
				{Name: "x", Number: 2, Type: model.PyType{Type: "str"}, Repeated: true},
			},
		},
		{
			name:    "slot without a field is rejected",
			sql:     "WHERE id IN (/*SLICE:ids*/?) AND a = ?3",
			columns: []model.Column{slice("ids", 1), column("x", 2)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := orderColumnsBySlot(tc.columns, sqllex.Slots(tc.sql, sqllex.SQLite))
			if ok != (tc.want != nil) {
				t.Fatalf("orderColumnsBySlot() ok = %v, want %v", ok, tc.want != nil)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("orderColumnsBySlot() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestNeedsSlotRewrite(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sql  string
		want bool
	}{
		{name: "bare only", sql: "WHERE a = ? AND b = ?", want: false},
		{name: "slice marker only", sql: "WHERE id IN (/*SLICE:ids*/?)", want: false},
		{name: "numbered without a marker", sql: "WHERE a = ?1 AND b = ?1", want: false},
		{name: "numbered after a marker", sql: "WHERE id IN (/*SLICE:ids*/?) AND a = ?2", want: true},
		{name: "numbered before a marker", sql: "WHERE a = ?1 AND id IN (/*SLICE:ids*/?)", want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := needsSlotRewrite(sqllex.Slots(tc.sql, sqllex.SQLite)); got != tc.want {
				t.Errorf("needsSlotRewrite(%q) = %v, want %v", tc.sql, got, tc.want)
			}
		})
	}
}
