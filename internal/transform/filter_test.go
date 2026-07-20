package transform_test

import (
	"reflect"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/transform"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
)

func TestFilterUnusedModelsKeepsReferencedModels(t *testing.T) {
	t.Parallel()
	enums := []model.Enum{{Name: "Mood"}, {Name: "UnusedEnum"}}
	tables := []model.Table{{Name: "Author"}, {Name: "Book"}, {Name: "Unused"}}
	queries := []model.Query{
		{
			// Scalar return without a row class: only the bare type is kept.
			Returns: model.QueryValue{Name: "total", Type: model.PyType{Type: types.Int}},
		},
		{
			Params: []model.QueryValue{{
				Name: "arg",
				Type: model.PyType{Type: "models.Author"},
				Table: &model.Table{
					Name: "Author",
					Columns: []model.Column{
						{Type: model.PyType{Type: "enums.Mood"}},
						{Embed: &model.Embed{
							ModelName: "models.Book",
							Columns:   []model.Column{{Type: model.PyType{Type: types.Int}}},
						}},
					},
				},
			}},
		},
	}

	gotEnums, gotTables := transform.FilterUnusedModels(enums, tables, queries)

	wantEnums := []model.Enum{{Name: "Mood"}}
	wantTables := []model.Table{{Name: "Author"}, {Name: "Book"}}
	if !reflect.DeepEqual(gotEnums, wantEnums) {
		t.Errorf("FilterUnusedModels() enums = %+v, want %+v", gotEnums, wantEnums)
	}
	if !reflect.DeepEqual(gotTables, wantTables) {
		t.Errorf("FilterUnusedModels() tables = %+v, want %+v", gotTables, wantTables)
	}
}

func TestFilterUnusedModelsDropsEverything(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		enums   []model.Enum
		tables  []model.Table
		queries []model.Query
	}{
		{name: "nil inputs"},
		{
			name:   "no queries",
			enums:  []model.Enum{{Name: "Mood"}},
			tables: []model.Table{{Name: "Author"}},
		},
		{
			name:    "empty query values keep nothing",
			enums:   []model.Enum{{Name: "Mood"}},
			tables:  []model.Table{{Name: "Author"}},
			queries: []model.Query{{}},
		},
		{
			name:    "query referencing missing model",
			enums:   []model.Enum{{Name: "Mood"}},
			tables:  []model.Table{{Name: "Author"}},
			queries: []model.Query{{Returns: model.QueryValue{Name: "row", Type: model.PyType{Type: "models.Ghost"}}}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotEnums, gotTables := transform.FilterUnusedModels(tc.enums, tc.tables, tc.queries)
			if len(gotEnums) != 0 {
				t.Errorf("FilterUnusedModels() enums = %+v, want empty", gotEnums)
			}
			if len(gotTables) != 0 {
				t.Errorf("FilterUnusedModels() tables = %+v, want empty", gotTables)
			}
		})
	}
}
