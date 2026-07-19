package config_test

import (
	"reflect"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func newRequest(options string, engine string) *plugin.GenerateRequest {
	return &plugin.GenerateRequest{
		PluginOptions: []byte(options),
		Settings:      &plugin.Settings{Engine: engine},
	}
}

func TestNewConfigErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		options string
		engine  string
		wantErr string
	}{
		{
			name:    "invalid json",
			options: "{",
			engine:  "postgresql",
			wantErr: "unmarshalling plugin options: unexpected end of JSON input",
		},
		{
			name:    "empty options fail validation",
			options: "",
			engine:  "postgresql",
			wantErr: "invalid options: you need to specify emit_init_file",
		},
		{
			name: "override parse error propagates",
			options: `{"package":"db","sql_driver":"asyncpg","emit_init_file":true,` +
				`"overrides":[{"db_type":"text","column":"authors.name","py_type":{"type":"str"}}]}`,
			engine:  "postgresql",
			wantErr: "override specifying both `column` (\"authors.name\") and `db_type` (\"text\") is not valid",
		},
		{
			name:    "negative omit_kwargs_limit",
			options: `{"package":"db","sql_driver":"asyncpg","emit_init_file":true,"omit_kwargs_limit":-1}`,
			engine:  "postgresql",
			wantErr: "invalid options: omit kwarg limit must not be negative",
		},
		{
			name:    "missing emit_init_file",
			options: `{"package":"db","sql_driver":"asyncpg"}`,
			engine:  "postgresql",
			wantErr: "invalid options: you need to specify emit_init_file",
		},
		{
			name:    "empty package",
			options: `{"sql_driver":"asyncpg","emit_init_file":true}`,
			engine:  "postgresql",
			wantErr: "invalid options: package must not be empty",
		},
		{
			name:    "unknown sql driver",
			options: `{"package":"db","sql_driver":"mysql","emit_init_file":true}`,
			engine:  "postgresql",
			wantErr: "invalid options: invalid sql driver: unknown SQL driver: mysql",
		},
		{
			name:    "asyncpg does not support sqlite",
			options: `{"package":"db","sql_driver":"asyncpg","emit_init_file":true}`,
			engine:  "sqlite",
			wantErr: "invalid options: invalid sql driver: SQL driver asyncpg does not support sqlite",
		},
		{
			name:    "sqlite3 does not support postgresql",
			options: `{"package":"db","sql_driver":"sqlite3","emit_init_file":true}`,
			engine:  "postgresql",
			wantErr: "invalid options: invalid sql driver: SQL driver sqlite3 does not support postgresql",
		},
		{
			name:    "aiosqlite does not support postgresql",
			options: `{"package":"db","sql_driver":"aiosqlite","emit_init_file":true}`,
			engine:  "postgresql",
			wantErr: "invalid options: invalid sql driver: SQL driver aiosqlite does not support postgresql",
		},
		{
			name:    "unknown model type",
			options: `{"package":"db","sql_driver":"asyncpg","emit_init_file":true,"model_type":"struct"}`,
			engine:  "postgresql",
			wantErr: "invalid options: unknown model type: struct",
		},
		{
			name:    "unknown docstring convention",
			options: `{"package":"db","sql_driver":"asyncpg","emit_init_file":true,"docstrings":"rst"}`,
			engine:  "postgresql",
			wantErr: "invalid options: unknown docstring convention: rst",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			conf, err := config.NewConfig(newRequest(tc.options, tc.engine))
			if err == nil {
				t.Fatalf("NewConfig() error = nil, want %q", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Errorf("NewConfig() error = %q, want %q", err.Error(), tc.wantErr)
			}
			if conf != nil {
				t.Errorf("NewConfig() = %v, want nil on error", conf)
			}
		})
	}
}

func TestNewConfigDefaults(t *testing.T) {
	t.Parallel()
	req := newRequest(`{"package":"db","sql_driver":"asyncpg","emit_init_file":true}`, "postgresql")
	conf, err := config.NewConfig(req)
	if err != nil {
		t.Fatalf("NewConfig() error = %v, want nil", err)
	}
	want := &config.Config{
		Package:             "db",
		SqlDriver:           config.SQLDriverAsyncpg,
		ModelType:           config.ModelTypeDataclass,
		Initialisms:         utils.ToPtr([]string{"id"}),
		EmitInitFile:        utils.ToPtr(true),
		EmitDocstrings:      config.DocstringConventionNone,
		EmitDocstringsSQL:   utils.ToPtr(true),
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
		InitialismsMap:      map[string]struct{}{"id": {}},
	}
	if !reflect.DeepEqual(conf, want) {
		t.Errorf("NewConfig() = %+v, want %+v", conf, want)
	}
}

func TestNewConfigAllOptions(t *testing.T) {
	t.Parallel()
	const (
		wantQueryParameterLimit = 2
		wantOmitKwargsLimit     = 3
		wantCharsPerIndentLevel = 8
	)
	options := `{
		"package": "queries",
		"sql_driver": "sqlite3",
		"model_type": "msgspec",
		"initialisms": ["id", "sql"],
		"emit_exact_table_names": true,
		"emit_classes": true,
		"inflection_exclude_table_names": ["users"],
		"omit_unused_models": true,
		"omit_typechecking_block": true,
		"query_parameter_limit": 2,
		"omit_kwargs_limit": 3,
		"emit_init_file": false,
		"docstrings": "numpy",
		"docstrings_emit_sql": false,
		"speedups": true,
		"debug": true,
		"indent_char": "\t",
		"chars_per_indent_level": 8,
		"overrides": [{"db_type": "text", "py_type": {"import": "collections", "type": "UserString", "package": "UserString"}}]
	}`
	conf, err := config.NewConfig(newRequest(options, "sqlite"))
	if err != nil {
		t.Fatalf("NewConfig() error = %v, want nil", err)
	}
	want := &config.Config{
		Package:                     "queries",
		SqlDriver:                   config.SQLDriverSQLite,
		ModelType:                   config.ModelTypeMsgspec,
		Initialisms:                 utils.ToPtr([]string{"id", "sql"}),
		EmitExactTableNames:         true,
		EmitClasses:                 true,
		InflectionExcludeTableNames: []string{"users"},
		OmitUnusedModels:            true,
		OmitTypecheckingBlock:       true,
		QueryParameterLimit:         utils.ToPtr(wantQueryParameterLimit),
		OmitKwargsLimit:             wantOmitKwargsLimit,
		EmitInitFile:                utils.ToPtr(false),
		EmitDocstrings:              config.DocstringConventionNumpy,
		EmitDocstringsSQL:           utils.ToPtr(false),
		Speedups:                    true,
		Debug:                       true,
		IndentChar:                  "\t",
		CharsPerIndentLevel:         wantCharsPerIndentLevel,
		Overrides: []config.Override{{
			PyType: config.OverridePyType{Import: "collections", Type: "UserString", Package: "UserString"},
			DBType: "text",
		}},
		InitialismsMap: map[string]struct{}{"id": {}, "sql": {}},
	}
	if !reflect.DeepEqual(conf, want) {
		t.Errorf("NewConfig() = %+v, want %+v", conf, want)
	}
}

func TestIsOverQueryParameterLimit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		limit    int
		hasLimit bool
		num      int
		want     bool
	}{
		{name: "nil limit is opt-out", num: 5, want: false},
		{name: "negative limit is opt-out", limit: -1, hasLimit: true, num: 5, want: false},
		{name: "zero limit with zero params", limit: 0, hasLimit: true, num: 0, want: false},
		{name: "zero limit with one param", limit: 0, hasLimit: true, num: 1, want: true},
		{name: "under limit", limit: 3, hasLimit: true, num: 2, want: false},
		{name: "exactly at limit", limit: 3, hasLimit: true, num: 3, want: false},
		{name: "over limit", limit: 3, hasLimit: true, num: 4, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			conf := &config.Config{}
			if tc.hasLimit {
				limit := tc.limit
				conf.QueryParameterLimit = &limit
			}
			if got := conf.IsOverQueryParameterLimit(tc.num); got != tc.want {
				t.Errorf("IsOverQueryParameterLimit(%d) with limit %v = %v, want %v", tc.num, tc.limit, got, tc.want)
			}
		})
	}
}
