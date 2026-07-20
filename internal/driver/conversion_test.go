package driver_test

import (
	"maps"
	"strings"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/driver"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
)

// convParamQuery builds a query whose parameters have the given types and
// whose return value is unset, mirroring an :exec query shape.
func convParamQuery(types ...model.PyType) model.Query {
	params := make([]model.QueryValue, 0, len(types))
	for _, typ := range types {
		params = append(params, model.QueryValue{Name: "p", Type: typ})
	}

	return model.Query{Params: params}
}

// convReturnQuery builds a query returning a single scalar of the given type.
func convReturnQuery(typ model.PyType) model.Query {
	return model.Query{Returns: model.QueryValue{Type: typ}}
}

func convDateType() model.PyType {
	return model.PyType{Type: "datetime.date", SQLType: "date"}
}

func convDatetimeType() model.PyType {
	return model.PyType{Type: "datetime.datetime", SQLType: "datetime"}
}

func convOverriddenDateType() model.PyType {
	return model.PyType{Type: "float", SQLType: "date", IsOverride: true, DefaultType: "datetime.date"}
}

func TestAsyncpgConversionTypes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		sqlType string
		want    bool
	}{
		{sqlType: "bytea", want: true},
		{sqlType: "blob", want: true},
		{sqlType: "pg_catalog.bytea", want: true},
		{sqlType: "inet", want: true},
		{sqlType: "cidr", want: true},
		{sqlType: "text", want: false},
		{sqlType: "", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.sqlType, func(t *testing.T) {
			t.Parallel()
			d, err := driver.New(&config.Config{SqlDriver: config.SQLDriverAsyncpg})
			if err != nil {
				t.Fatalf("driver.New() error = %v", err)
			}
			if got := d.NeedsConversion(tc.sqlType); got != tc.want {
				t.Errorf("NeedsConversion(%q) = %v, want %v", tc.sqlType, got, tc.want)
			}
		})
	}
}

func TestSqliteConversionUsageAny(t *testing.T) {
	t.Parallel()
	if got := driver.SqliteConversionsUsed(nil).Any(); got {
		t.Errorf("SqliteConversionsUsed(nil).Any() = %v, want false", got)
	}
	usage := driver.SqliteConversionsUsed([]model.Query{convParamQuery(convDateType())})
	if got := usage.Any(); !got {
		t.Errorf("Any() = %v, want true", got)
	}
}

func TestSqliteConversionUsageRuntimeModules(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		queries  []model.Query
		speedups bool
		want     map[string]struct{}
	}{
		{
			name:    "adapter needs module",
			queries: []model.Query{convParamQuery(convDateType())},
			want:    map[string]struct{}{"datetime": {}},
		},
		{
			name:     "adapter with speedups still needs module",
			queries:  []model.Query{convParamQuery(convDateType())},
			speedups: true,
			want:     map[string]struct{}{"datetime": {}},
		},
		{
			name:    "converter needs module",
			queries: []model.Query{convReturnQuery(convDateType())},
			want:    map[string]struct{}{"datetime": {}},
		},
		{
			name:     "speedups converter drops module",
			queries:  []model.Query{convReturnQuery(convDateType())},
			speedups: true,
			want:     map[string]struct{}{},
		},
		{
			name:     "speedups converter without variant keeps module",
			queries:  []model.Query{convReturnQuery(model.PyType{Type: "decimal.Decimal", SQLType: "decimal"})},
			speedups: true,
			want:     map[string]struct{}{"decimal": {}},
		},
		{
			name: "builtin types need no module",
			queries: []model.Query{
				convParamQuery(model.PyType{Type: "memoryview", SQLType: "blob"}),
				convReturnQuery(model.PyType{Type: "bool", SQLType: "boolean"}),
			},
			want: map[string]struct{}{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			usage := driver.SqliteConversionsUsed(tc.queries)
			if got := usage.RuntimeModules(tc.speedups); !maps.Equal(got, tc.want) {
				t.Errorf("RuntimeModules(%v) = %v, want %v", tc.speedups, got, tc.want)
			}
		})
	}
}

func TestSqliteConversionUsageSpeedupConverterUsed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		queries []model.Query
		want    bool
	}{
		{
			name:    "converter with speedups variant",
			queries: []model.Query{convReturnQuery(convDateType())},
			want:    true,
		},
		{
			name:    "adapter only never uses speedups",
			queries: []model.Query{convParamQuery(convDateType())},
			want:    false,
		},
		{
			name:    "converter without speedups variant",
			queries: []model.Query{convReturnQuery(model.PyType{Type: "decimal.Decimal", SQLType: "decimal"})},
			want:    false,
		},
		{
			name:    "no conversions",
			queries: nil,
			want:    false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			usage := driver.SqliteConversionsUsed(tc.queries)
			if got := usage.SpeedupConverterUsed(); got != tc.want {
				t.Errorf("SpeedupConverterUsed() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSqliteWriteConversionSetup(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		sqlDriver   config.SQLDriver
		speedups    bool
		queries     []model.Query
		wantWritten bool
		want        string
	}{
		{
			name:      "no conversions used",
			sqlDriver: config.SQLDriverSQLite,
			queries: []model.Query{
				{
					Params:  []model.QueryValue{{Name: "n", Type: model.PyType{Type: "int", SQLType: "integer"}}},
					Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
				},
			},
			wantWritten: false,
			want:        "",
		},
		{
			name:        "adapter only",
			sqlDriver:   config.SQLDriverSQLite,
			queries:     []model.Query{convParamQuery(convDateType())},
			wantWritten: true,
			want: strings.Join([]string{
				"def _adapt_date(val: datetime.date) -> str:",
				"    return val.isoformat()",
				"",
				"",
				"sqlite3.register_adapter(datetime.date, _adapt_date)",
				"",
			}, "\n"),
		},
		{
			name:        "converter only",
			sqlDriver:   config.SQLDriverSQLite,
			queries:     []model.Query{convReturnQuery(convDateType())},
			wantWritten: true,
			want: strings.Join([]string{
				"def _convert_date(val: bytes) -> datetime.date:",
				"    return datetime.date.fromisoformat(val.decode())",
				"",
				"",
				`sqlite3.register_converter("date", _convert_date)`,
				"",
			}, "\n"),
		},
		{
			name:      "adapter and converter aiosqlite multi keys",
			sqlDriver: config.SQLDriverAioSQLite,
			queries: []model.Query{
				convParamQuery(convDatetimeType()),
				convReturnQuery(convDatetimeType()),
			},
			wantWritten: true,
			want: strings.Join([]string{
				"def _adapt_datetime(val: datetime.datetime) -> str:",
				"    return val.isoformat()",
				"",
				"",
				"def _convert_datetime(val: bytes) -> datetime.datetime:",
				"    return datetime.datetime.fromisoformat(val.decode())",
				"",
				"",
				"aiosqlite.register_adapter(datetime.datetime, _adapt_datetime)",
				"",
				`aiosqlite.register_converter("datetime", _convert_datetime)`,
				`aiosqlite.register_converter("timestamp", _convert_datetime)`,
				"",
			}, "\n"),
		},
		{
			name:        "speedups converter uses ciso8601",
			sqlDriver:   config.SQLDriverSQLite,
			speedups:    true,
			queries:     []model.Query{convReturnQuery(convDatetimeType())},
			wantWritten: true,
			want: strings.Join([]string{
				"def _convert_datetime(val: bytes) -> datetime.datetime:",
				"    return ciso8601.parse_datetime(val.decode())",
				"",
				"",
				`sqlite3.register_converter("datetime", _convert_datetime)`,
				`sqlite3.register_converter("timestamp", _convert_datetime)`,
				"",
			}, "\n"),
		},
		{
			name:        "speedups without variant keeps body",
			sqlDriver:   config.SQLDriverSQLite,
			speedups:    true,
			queries:     []model.Query{convReturnQuery(model.PyType{Type: "decimal.Decimal", SQLType: "decimal"})},
			wantWritten: true,
			want: strings.Join([]string{
				"def _convert_decimal(val: bytes) -> decimal.Decimal:",
				"    return decimal.Decimal(val.decode())",
				"",
				"",
				`sqlite3.register_converter("decimal", _convert_decimal)`,
				"",
			}, "\n"),
		},
		{
			name:      "overridden return excluded overridden param kept",
			sqlDriver: config.SQLDriverSQLite,
			queries: []model.Query{
				convParamQuery(convOverriddenDateType()),
				convReturnQuery(convOverriddenDateType()),
			},
			wantWritten: true,
			want: strings.Join([]string{
				"def _adapt_date(val: datetime.date) -> str:",
				"    return val.isoformat()",
				"",
				"",
				"sqlite3.register_adapter(datetime.date, _adapt_date)",
				"",
			}, "\n"),
		},
		{
			name:      "struct return with embed",
			sqlDriver: config.SQLDriverSQLite,
			queries: []model.Query{
				{
					Returns: model.QueryValue{
						Table: &model.Table{
							Name: "AuthorRow",
							Columns: []model.Column{
								{
									Name: "author",
									Type: model.PyType{Type: "models.Author"},
									Embed: &model.Embed{
										ModelName: "Author",
										Columns: []model.Column{
											{Name: "created_at", Type: convDateType()},
										},
									},
								},
								{Name: "active", Type: model.PyType{Type: "bool", SQLType: "boolean"}},
							},
						},
						Type: model.PyType{Type: "AuthorRow"},
					},
				},
			},
			wantWritten: true,
			want: strings.Join([]string{
				"def _convert_date(val: bytes) -> datetime.date:",
				"    return datetime.date.fromisoformat(val.decode())",
				"",
				"",
				"def _convert_bool(val: bytes) -> bool:",
				"    return bool(int(val))",
				"",
				"",
				`sqlite3.register_converter("date", _convert_date)`,
				`sqlite3.register_converter("bool", _convert_bool)`,
				`sqlite3.register_converter("boolean", _convert_bool)`,
				"",
			}, "\n"),
		},
		{
			name:      "struct param bundle",
			sqlDriver: config.SQLDriverSQLite,
			queries: []model.Query{
				{
					Params: []model.QueryValue{
						{
							EmitTable: true,
							Table: &model.Table{
								Name: "InsertParams",
								Columns: []model.Column{
									{Name: "created_at", Type: convDateType()},
									{Name: "n", Type: model.PyType{Type: "int", SQLType: "integer"}},
								},
							},
							Name: "params",
							Type: model.PyType{Type: "InsertParams"},
						},
					},
				},
			},
			wantWritten: true,
			want: strings.Join([]string{
				"def _adapt_date(val: datetime.date) -> str:",
				"    return val.isoformat()",
				"",
				"",
				"sqlite3.register_adapter(datetime.date, _adapt_date)",
				"",
			}, "\n"),
		},
		{
			name:      "emission follows spec order not query order",
			sqlDriver: config.SQLDriverSQLite,
			queries: []model.Query{
				convParamQuery(model.PyType{Type: "memoryview", SQLType: "blob"}),
				convReturnQuery(convDateType()),
			},
			wantWritten: true,
			want: strings.Join([]string{
				"def _convert_date(val: bytes) -> datetime.date:",
				"    return datetime.date.fromisoformat(val.decode())",
				"",
				"",
				"def _adapt_memoryview(val: memoryview) -> bytes:",
				"    return val.tobytes()",
				"",
				"",
				"sqlite3.register_adapter(memoryview, _adapt_memoryview)",
				"",
				`sqlite3.register_converter("date", _convert_date)`,
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			conf := &config.Config{
				SqlDriver:           tc.sqlDriver,
				EmitDocstrings:      config.DocstringConventionNone,
				EmitDocstringsSQL:   utils.ToPtr(true),
				Speedups:            tc.speedups,
				IndentChar:          " ",
				CharsPerIndentLevel: 4,
			}
			d, err := driver.New(conf)
			if err != nil {
				t.Fatalf("driver.New() error = %v", err)
			}
			body := writer.NewCodeWriter(conf)
			if got := d.WriteConversionSetup(body, conf, tc.queries); got != tc.wantWritten {
				t.Errorf("WriteConversionSetup() = %v, want %v", got, tc.wantWritten)
			}
			if got := body.String(); got != tc.want {
				t.Errorf("WriteConversionSetup() wrote %q, want %q", got, tc.want)
			}
		})
	}
}
