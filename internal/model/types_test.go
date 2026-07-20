package model_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
)

func TestPyTypePrint(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		typ  model.PyType
		want string
	}{
		{
			name: "plain type",
			typ:  model.PyType{Type: "int"},
			want: "int",
		},
		{
			name: "list wraps in sequence",
			typ:  model.PyType{Type: "str", IsList: true},
			want: "collections.abc.Sequence[str]",
		},
		{
			name: "nullable appends none",
			typ:  model.PyType{Type: "str", IsNullable: true},
			want: "str | None",
		},
		{
			name: "nullable list wraps then appends none",
			typ:  model.PyType{Type: "bytes", IsList: true, IsNullable: true},
			want: "collections.abc.Sequence[bytes] | None",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.typ.Print(); got != tc.want {
				t.Errorf("PyType%+v.Print() = %q, want %q", tc.typ, got, tc.want)
			}
		})
	}
}

func TestPyTypePrintOptional(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		typ  model.PyType
		want string
	}{
		{
			name: "nullable type is printed unchanged",
			typ:  model.PyType{Type: "int", IsNullable: true},
			want: "int | None",
		},
		{
			name: "non-nullable type gets none appended",
			typ:  model.PyType{Type: "int"},
			want: "int | None",
		},
		{
			name: "non-nullable list gets none appended",
			typ:  model.PyType{Type: "int", IsList: true},
			want: "collections.abc.Sequence[int] | None",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.typ.PrintOptional(); got != tc.want {
				t.Errorf("PyType%+v.PrintOptional() = %q, want %q", tc.typ, got, tc.want)
			}
		})
	}
}

func TestPyTypeDoOverride(t *testing.T) {
	t.Parallel()
	overridden := model.PyType{Type: "int", IsOverride: true, DefaultType: "str"}
	if !overridden.DoOverride() {
		t.Error("DoOverride() = false for a type with IsOverride set, want true")
	}
	plain := model.PyType{Type: "int"}
	if plain.DoOverride() {
		t.Error("DoOverride() = true for a type without IsOverride, want false")
	}
}

func TestPyTypeHasConverter(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		typ  model.PyType
		want bool
	}{
		{name: "no converter", typ: model.PyType{Type: "int"}, want: false},
		{name: "both directions", typ: model.PyType{Type: "M", ConverterTo: "m.to", ConverterFrom: "m.from"}, want: true},
		{name: "only to_db", typ: model.PyType{Type: "M", ConverterTo: "m.to"}, want: true},
		{name: "only from_db", typ: model.PyType{Type: "M", ConverterFrom: "m.from"}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.typ.HasConverter(); got != tc.want {
				t.Errorf("HasConverter() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestQueryEmitsTable(t *testing.T) {
	t.Parallel()
	table := &model.Table{Name: "User"}
	cases := []struct {
		name  string
		query model.Query
		want  bool
	}{
		{
			name:  "returns emit a table",
			query: model.Query{Returns: model.QueryValue{EmitTable: true, Table: table}},
			want:  true,
		},
		{
			name: "a param emits a table",
			query: model.Query{
				Params: []model.QueryValue{
					{Name: "id", Type: model.PyType{Type: "int"}},
					{EmitTable: true, Table: table},
				},
			},
			want: true,
		},
		{
			name: "params without emitted tables",
			query: model.Query{
				Params: []model.QueryValue{{Name: "id", Type: model.PyType{Type: "int"}}},
			},
			want: false,
		},
		{
			name:  "no params and no emitted return",
			query: model.Query{},
			want:  false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.query.EmitsTable(); got != tc.want {
				t.Errorf("EmitsTable() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestQueryValueIsEmpty(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value model.QueryValue
		want  bool
	}{
		{
			name:  "zero value is empty",
			value: model.QueryValue{},
			want:  true,
		},
		{
			name:  "type set",
			value: model.QueryValue{Type: model.PyType{Type: "int"}},
			want:  false,
		},
		{
			name:  "name set",
			value: model.QueryValue{Name: "id"},
			want:  false,
		},
		{
			name:  "table set",
			value: model.QueryValue{Table: &model.Table{Name: "User"}},
			want:  false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.value.IsEmpty(); got != tc.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestQueryValueIsStruct(t *testing.T) {
	t.Parallel()
	withTable := model.QueryValue{Table: &model.Table{Name: "User"}}
	if !withTable.IsStruct() {
		t.Error("IsStruct() = false for a value with a table, want true")
	}
	withoutTable := model.QueryValue{Name: "id"}
	if withoutTable.IsStruct() {
		t.Error("IsStruct() = true for a value without a table, want false")
	}
}
