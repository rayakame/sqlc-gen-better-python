package driver

import (
	"strings"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
)

const decodeHookName = "_decode_hook"

func newRowBuilderWriter(convention config.DocstringConvention) *writer.CodeWriter {
	return writer.NewCodeWriter(&config.Config{
		SqlDriver:           config.SQLDriverAsyncpg,
		EmitDocstrings:      convention,
		EmitDocstringsSQL:   utils.ToPtr(true),
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
	})
}

func TestConvertExpr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		// driverConverts is the constant the RowBuilder's needsConversion func returns.
		driverConverts bool
		typ            model.PyType
		src            string
		want           string
	}{
		{
			name: "no conversion returns source",
			typ:  model.PyType{Type: "int"},
			src:  "row[0]",
			want: "row[0]",
		},
		{
			name: "enum scalar constructor",
			typ:  model.PyType{Type: "Status", IsEnum: true},
			src:  "row[0]",
			want: "Status(row[0])",
		},
		{
			name: "override scalar constructor",
			typ:  model.PyType{Type: "uuid.UUID", IsOverride: true, DefaultType: "str"},
			src:  "row[1]",
			want: "uuid.UUID(row[1])",
		},
		{
			name:           "driver conversion via sql type",
			driverConverts: true,
			typ:            model.PyType{Type: "datetime.datetime", SQLType: "timestamp"},
			src:            "row[0]",
			want:           "datetime.datetime(row[0])",
		},
		{
			name: "enum list comprehension",
			typ:  model.PyType{Type: "Status", IsEnum: true, IsList: true},
			src:  "row[0]",
			want: "[Status(v) for v in row[0]]",
		},
		{
			name: "nullable scalar guarded",
			typ:  model.PyType{Type: "Status", IsEnum: true, IsNullable: true},
			src:  "row[0]",
			want: "Status(row[0]) if row[0] is not None else None",
		},
		{
			name: "nullable override list guarded comprehension",
			typ:  model.PyType{Type: "uuid.UUID", IsOverride: true, IsList: true, IsNullable: true},
			src:  "row[2]",
			want: "[uuid.UUID(v) for v in row[2]] if row[2] is not None else None",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rb := newRowBuilder(func(string) bool { return tc.driverConverts })
			if got := rb.convertExpr(tc.typ, tc.src); got != tc.want {
				t.Errorf("convertExpr(%+v, %q) = %q, want %q", tc.typ, tc.src, got, tc.want)
			}
		})
	}
}

func TestColumnNeedsConversion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		driverConverts bool
		typ            model.PyType
		want           bool
	}{
		{name: "override converts", typ: model.PyType{Type: "uuid.UUID", IsOverride: true}, want: true},
		{name: "enum converts", typ: model.PyType{Type: "Status", IsEnum: true}, want: true},
		{
			name:           "driver sql type converts",
			driverConverts: true,
			typ:            model.PyType{Type: "float", SQLType: "real"},
			want:           true,
		},
		{name: "plain type does not convert", typ: model.PyType{Type: "int", SQLType: "integer"}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rb := newRowBuilder(func(string) bool { return tc.driverConverts })
			if got := rb.columnNeedsConversion(tc.typ); got != tc.want {
				t.Errorf("columnNeedsConversion(%+v) = %v, want %v", tc.typ, got, tc.want)
			}
		})
	}
}

func TestWriteScalarReturn(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		indent int
		ret    model.QueryValue
		want   string
	}{
		{
			name:   "converted enum",
			indent: 1,
			ret:    model.QueryValue{Name: "status", Type: model.PyType{Type: "Status", IsEnum: true}},
			want:   "    return Status(row[0])\n",
		},
		{
			name:   "plain value",
			indent: 2,
			ret:    model.QueryValue{Name: "id", Type: model.PyType{Type: "int"}},
			want:   "        return row[0]\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rb := newRowBuilder(func(string) bool { return false })
			body := newRowBuilderWriter(config.DocstringConventionNone)
			rb.WriteScalarReturn(body, tc.indent, tc.ret)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteScalarReturn() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestWriteStructReturn(t *testing.T) {
	t.Parallel()
	// longType forces the single-line rendering over MaxLineLength.
	longType := strings.Repeat("A", writer.MaxLineLength)
	cases := []struct {
		name   string
		indent int
		ret    model.QueryValue
		want   string
	}{
		{
			name:   "plain columns fit on one line",
			indent: 1,
			ret: model.QueryValue{
				Name: "author",
				Type: model.PyType{Type: "Author"},
				Table: &model.Table{
					Name: "author",
					Columns: []model.Column{
						{Name: "id", Type: model.PyType{Type: "int"}},
						{Name: "name", Type: model.PyType{Type: "str"}},
					},
				},
			},
			want: "    return Author(id=row[0], name=row[1])\n",
		},
		{
			name:   "embed and conversions fit on one line",
			indent: 1,
			ret: model.QueryValue{
				Name: "post",
				Type: model.PyType{Type: "Post"},
				Table: &model.Table{
					Name: "post",
					Columns: []model.Column{
						{Name: "id", Type: model.PyType{Type: "int"}},
						{
							Name: "author",
							Type: model.PyType{Type: "Author"},
							Embed: &model.Embed{
								ModelName: "Author",
								Columns: []model.Column{
									{Name: "name", Type: model.PyType{Type: "str"}},
									{Name: "status", Type: model.PyType{Type: "Status", IsEnum: true}},
								},
							},
						},
						{Name: "tag", Type: model.PyType{Type: "Tag", IsEnum: true}},
					},
				},
			},
			want: "    return Post(id=row[0], author=Author(name=row[1], status=Status(row[2])), tag=Tag(row[3]))\n",
		},
		{
			name:   "long construction explodes with trailing commas",
			indent: 1,
			ret: model.QueryValue{
				Name: "wide",
				Type: model.PyType{Type: longType},
				Table: &model.Table{
					Name: "wide",
					Columns: []model.Column{
						{
							Name: "author",
							Type: model.PyType{Type: "Author"},
							Embed: &model.Embed{
								ModelName: "Author",
								Columns: []model.Column{
									{Name: "name", Type: model.PyType{Type: "str"}},
									{Name: "email", Type: model.PyType{Type: "str"}},
								},
							},
						},
						{Name: "id", Type: model.PyType{Type: "int"}},
					},
				},
			},
			want: strings.Join([]string{
				"    return " + longType + "(",
				"        author=Author(name=row[0], email=row[1]),",
				"        id=row[2],",
				"    )",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rb := newRowBuilder(func(string) bool { return false })
			body := newRowBuilderWriter(config.DocstringConventionNone)
			rb.WriteStructReturn(body, tc.indent, tc.ret)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteStructReturn() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestWriteDecodeHook(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		convention config.DocstringConvention
		indent     int
		returns    model.QueryValue
		wantHook   string
		wantBody   string
	}{
		{
			name:       "plain scalar uses itemgetter",
			convention: config.DocstringConventionNone,
			indent:     1,
			returns:    model.QueryValue{Name: "id", Type: model.PyType{Type: "int"}},
			wantHook:   "operator.itemgetter(0)",
			wantBody:   "",
		},
		{
			name:       "converted scalar emits hook",
			convention: config.DocstringConventionNone,
			indent:     1,
			returns:    model.QueryValue{Name: "status", Type: model.PyType{Type: "Status", IsEnum: true}},
			wantHook:   decodeHookName,
			wantBody: strings.Join([]string{
				"    def _decode_hook(row: asyncpg.Record) -> Status:",
				"        return Status(row[0])",
				"",
				"",
			}, "\n"),
		},
		{
			name:       "struct with docstrings gets leading blank line",
			convention: config.DocstringConventionGoogle,
			indent:     1,
			returns: model.QueryValue{
				Name: "author",
				Type: model.PyType{Type: "Author"},
				Table: &model.Table{
					Name: "author",
					Columns: []model.Column{
						{Name: "id", Type: model.PyType{Type: "int"}},
						{Name: "name", Type: model.PyType{Type: "str"}},
					},
				},
			},
			wantHook: decodeHookName,
			wantBody: strings.Join([]string{
				"",
				"    def _decode_hook(row: asyncpg.Record) -> Author:",
				"        return Author(id=row[0], name=row[1])",
				"",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rb := newRowBuilder(func(string) bool { return false })
			body := newRowBuilderWriter(tc.convention)
			query := model.Query{Returns: tc.returns}
			got := rb.WriteDecodeHook(body, tc.indent, query, "asyncpg.Record")
			if got != tc.wantHook {
				t.Errorf("WriteDecodeHook() = %q, want %q", got, tc.wantHook)
			}
			if gotBody := body.String(); gotBody != tc.wantBody {
				t.Errorf("WriteDecodeHook() body = %q, want %q", gotBody, tc.wantBody)
			}
		})
	}
}
