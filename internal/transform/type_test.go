package transform

import (
	"reflect"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

const typeTestBaseOptions = `{"package":"db","sql_driver":"asyncpg","emit_init_file":true}`

const typeTestOverrideOptions = `{"package":"db","sql_driver":"asyncpg","emit_init_file":true,"overrides":[` +
	`{"db_type":"pg_catalog.UUID","py_type":{"type":"str"}},` +
	`{"db_type":"mood","py_type":{"import":"custom","type":"custom.Mood"}},` +
	`{"column":"authors.name","py_type":{"import":"collections","type":"collections.UserString"}},` +
	`{"column":"analytics.events.payload","py_type":{"type":"bytes"}}]}`

const typeTestConverterOptions = `{"package":"db","sql_driver":"asyncpg","emit_init_file":true,` +
	`"converters":[{"name":"money","py_type":{"import":"myapp.money","type":"myapp.money.Money"},` +
	`"to_db":"myapp.converters.encode","from_db":"myapp.converters.decode"}],` +
	`"overrides":[{"db_type":"numeric","converter":"money"},` +
	`{"db_type":"text","py_type":{"type":"bytes"}}]}`

func typeTestRequest(options string) *plugin.GenerateRequest {
	return &plugin.GenerateRequest{
		PluginOptions: []byte(options),
		Settings:      &plugin.Settings{Engine: "postgresql"},
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				{Name: types.PgCatalog, Enums: []*plugin.Enum{{Name: "mood"}}},
				{Name: types.InformationSchema},
				{Name: "public", Enums: []*plugin.Enum{{Name: "mood"}}},
				{Name: "analytics", Enums: []*plugin.Enum{{Name: "level"}}},
			},
		},
	}
}

func TestBuildPyType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		options string
		column  *plugin.Column
		want    model.PyType
	}{
		{
			name:    "not null scalar",
			options: typeTestBaseOptions,
			column:  &plugin.Column{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
			want:    model.PyType{SQLType: "int4", Type: types.Int, DefaultType: types.Int},
		},
		{
			name:    "nullable scalar",
			options: typeTestBaseOptions,
			column:  &plugin.Column{Name: "title", Type: &plugin.Identifier{Name: "text"}},
			want:    model.PyType{SQLType: "text", Type: "str", IsNullable: true, DefaultType: "str"},
		},
		{
			name:    "array column",
			options: typeTestBaseOptions,
			column:  &plugin.Column{Name: "tags", Type: &plugin.Identifier{Name: "text"}, NotNull: true, IsArray: true},
			want:    model.PyType{SQLType: "text", Type: "str", IsList: true, DefaultType: "str"},
		},
		{
			name:    "sqlc slice parameter",
			options: typeTestBaseOptions,
			column:  &plugin.Column{Name: "ids", Type: &plugin.Identifier{Name: "int4"}, NotNull: true, IsSqlcSlice: true},
			want:    model.PyType{SQLType: "int4", Type: types.Int, IsList: true, DefaultType: types.Int, SqlcSliceName: "ids"},
		},
		{
			// The generated expansion calls len() on the sequence, so a slice
			// against a nullable column must not become "Sequence[T] | None".
			name:    "sqlc slice parameter on nullable column stays required",
			options: typeTestBaseOptions,
			column:  &plugin.Column{Name: "notes", Type: &plugin.Identifier{Name: "text"}, IsSqlcSlice: true},
			want:    model.PyType{SQLType: "text", Type: "str", IsList: true, DefaultType: "str", SqlcSliceName: "notes"},
		},
		{
			// DDL casing survives into the identifier; SQLType must come out
			// lowercased regardless of what the conversion func returns.
			name:    "sql type is lowercased",
			options: typeTestBaseOptions,
			column:  &plugin.Column{Name: "raw", Type: &plugin.Identifier{Name: "TEXT"}, NotNull: true},
			want:    model.PyType{SQLType: "text", Type: "typing.Any", DefaultType: "typing.Any"},
		},
		{
			name:    "enum in default schema",
			options: typeTestBaseOptions,
			column:  &plugin.Column{Name: "mood", Type: &plugin.Identifier{Name: "mood"}, NotNull: true},
			want:    model.PyType{SQLType: "mood", Type: "enums.Mood", IsEnum: true, DefaultType: "enums.Mood"},
		},
		{
			name:    "enum in non-default schema",
			options: typeTestBaseOptions,
			column:  &plugin.Column{Name: "level", Type: &plugin.Identifier{Schema: "analytics", Name: "level"}, NotNull: true},
			want: model.PyType{
				SQLType:     "analytics.level",
				Type:        "enums.AnalyticsLevel",
				IsEnum:      true,
				DefaultType: "enums.AnalyticsLevel",
			},
		},
		{
			name:    "non-enum type in non-default schema",
			options: typeTestBaseOptions,
			column:  &plugin.Column{Name: "x", Type: &plugin.Identifier{Schema: "analytics", Name: "mood"}, NotNull: true},
			want:    model.PyType{SQLType: "analytics.mood", Type: "typing.Any", DefaultType: "typing.Any"},
		},
		{
			name:    "db_type override matches case-insensitively",
			options: typeTestOverrideOptions,
			column:  &plugin.Column{Name: "uid", Type: &plugin.Identifier{Schema: "pg_catalog", Name: "uuid"}, NotNull: true},
			want:    model.PyType{SQLType: "pg_catalog.uuid", Type: "str", IsOverride: true, DefaultType: "uuid.UUID"},
		},
		{
			name:    "db_type override on sqlc slice parameter keeps slice name",
			options: typeTestOverrideOptions,
			column: &plugin.Column{
				Name:        "uids",
				Type:        &plugin.Identifier{Schema: "pg_catalog", Name: "uuid"},
				NotNull:     true,
				IsSqlcSlice: true,
			},
			want: model.PyType{
				SQLType:       "pg_catalog.uuid",
				Type:          "str",
				IsList:        true,
				IsOverride:    true,
				DefaultType:   "uuid.UUID",
				SqlcSliceName: "uids",
			},
		},
		{
			name:    "db_type override on enum column disables enum handling",
			options: typeTestOverrideOptions,
			column:  &plugin.Column{Name: "current_mood", Type: &plugin.Identifier{Name: "mood"}, NotNull: true},
			want:    model.PyType{SQLType: "mood", Type: "custom.Mood", IsOverride: true, DefaultType: "enums.Mood"},
		},
		{
			name:    "column override keeps nullability and list flags",
			options: typeTestOverrideOptions,
			column: &plugin.Column{
				Name:    "name",
				Type:    &plugin.Identifier{Name: "text"},
				IsArray: true,
				Table:   &plugin.Identifier{Name: "authors"},
			},
			want: model.PyType{
				SQLType:     "text",
				Type:        "collections.UserString",
				IsNullable:  true,
				IsList:      true,
				IsOverride:  true,
				DefaultType: "str",
			},
		},
		{
			name:    "column override matches original name",
			options: typeTestOverrideOptions,
			column: &plugin.Column{
				Name:         "renamed",
				OriginalName: "name",
				Type:         &plugin.Identifier{Name: "text"},
				NotNull:      true,
				Table:        &plugin.Identifier{Name: "authors"},
			},
			want: model.PyType{SQLType: "text", Type: "collections.UserString", IsOverride: true, DefaultType: "str"},
		},
		{
			name:    "original name takes precedence and does not match",
			options: typeTestOverrideOptions,
			column: &plugin.Column{
				Name:         "name",
				OriginalName: "other",
				Type:         &plugin.Identifier{Name: "text"},
				NotNull:      true,
				Table:        &plugin.Identifier{Name: "authors"},
			},
			want: model.PyType{SQLType: "text", Type: "str", DefaultType: "str"},
		},
		{
			name:    "column override rejects other tables",
			options: typeTestOverrideOptions,
			column: &plugin.Column{
				Name:    "name",
				Type:    &plugin.Identifier{Name: "text"},
				NotNull: true,
				Table:   &plugin.Identifier{Name: "books"},
			},
			want: model.PyType{SQLType: "text", Type: "str", DefaultType: "str"},
		},
		{
			name:    "column override rejects columns without a table",
			options: typeTestOverrideOptions,
			column:  &plugin.Column{Name: "name", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
			want:    model.PyType{SQLType: "text", Type: "str", DefaultType: "str"},
		},
		{
			name:    "schema qualified column override",
			options: typeTestOverrideOptions,
			column: &plugin.Column{
				Name:    "payload",
				Type:    &plugin.Identifier{Name: "text"},
				NotNull: true,
				Table:   &plugin.Identifier{Schema: "analytics", Name: "events"},
			},
			want: model.PyType{SQLType: "text", Type: "bytes", IsOverride: true, DefaultType: "str"},
		},
		{
			name:    "no override matches",
			options: typeTestOverrideOptions,
			column: &plugin.Column{
				Name:    "id",
				Type:    &plugin.Identifier{Name: "int4"},
				NotNull: true,
				Table:   &plugin.Identifier{Name: "authors"},
			},
			want: model.PyType{SQLType: "int4", Type: types.Int, DefaultType: types.Int},
		},
		{
			name:    "resolved converter sets both conversion functions",
			options: typeTestConverterOptions,
			column:  &plugin.Column{Name: "price", Type: &plugin.Identifier{Name: "numeric"}, NotNull: true},
			want: model.PyType{
				SQLType:       "numeric",
				Type:          "myapp.money.Money",
				IsOverride:    true,
				DefaultType:   types.Decimal,
				ConverterTo:   "myapp.converters.encode",
				ConverterFrom: "myapp.converters.decode",
			},
		},
		{
			name:    "override without a converter leaves conversion functions empty",
			options: typeTestConverterOptions,
			column:  &plugin.Column{Name: "title", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
			want:    model.PyType{SQLType: "text", Type: "bytes", IsOverride: true, DefaultType: "str"},
		},
		{
			name:    "no override leaves conversion functions empty",
			options: typeTestConverterOptions,
			column:  &plugin.Column{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
			want:    model.PyType{SQLType: "int4", Type: types.Int, DefaultType: types.Int},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := typeTestRequest(tc.options)
			conf, err := config.NewConfig(req)
			if err != nil {
				t.Fatalf("NewConfig() error = %v, want nil", err)
			}
			tf := NewTransformer(conf, req, types.PostgresTypeToPython)
			if got := tf.buildPyType(tc.column); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("buildPyType() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestBuildPyTypeDoesNotMutateColumn(t *testing.T) {
	t.Parallel()
	req := typeTestRequest(typeTestBaseOptions)
	conf, err := config.NewConfig(req)
	if err != nil {
		t.Fatalf("NewConfig() error = %v, want nil", err)
	}
	tf := NewTransformer(conf, req, types.PostgresTypeToPython)
	column := &plugin.Column{Name: "mood", Type: &plugin.Identifier{Name: "mood"}, NotNull: true}

	first := tf.buildPyType(column)

	if column.Type.Schema != "" {
		t.Errorf("buildPyType() wrote schema %q back into the shared column", column.Type.Schema)
	}
	if second := tf.buildPyType(column); !reflect.DeepEqual(first, second) {
		t.Errorf("second buildPyType() = %+v, want %+v (must be idempotent)", second, first)
	}
}

func TestMatchOverrideSkipsEmptyPyType(t *testing.T) {
	t.Parallel()
	req := typeTestRequest(typeTestBaseOptions)
	// config.NewConfig rejects overrides without a py_type, so build the
	// config by hand to reach the defensive skip.
	conf := &config.Config{Overrides: []config.Override{{DBType: "text"}}}
	tf := NewTransformer(conf, req, types.PostgresTypeToPython)

	got := tf.buildPyType(&plugin.Column{Name: "title", Type: &plugin.Identifier{Name: "text"}, NotNull: true})

	want := model.PyType{SQLType: "text", Type: "str", DefaultType: "str"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildPyType() = %+v, want %+v", got, want)
	}
}
