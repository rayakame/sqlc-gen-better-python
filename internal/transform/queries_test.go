package transform_test

import (
	"reflect"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/transform"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

var (
	pyInt = model.PyType{SQLType: "int4", Type: "int", DefaultType: "int"}
	pyStr = model.PyType{SQLType: "text", Type: "str", DefaultType: "str"}
)

func queryCol(name, sqlType string, table *plugin.Identifier) *plugin.Column {
	return &plugin.Column{Name: name, NotNull: true, Type: &plugin.Identifier{Name: sqlType}, Table: table}
}

func authorsIdent() *plugin.Identifier {
	return &plugin.Identifier{Name: "test_authors"}
}

// queriesRequest builds a request whose catalog produces two models:
// TestAuthor (id int, name str) and TestPref (id int, mood enums.Mood).
func queriesRequest(queries []*plugin.Query) *plugin.GenerateRequest {
	return &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				{
					Name:  "public",
					Enums: []*plugin.Enum{{Name: "mood", Vals: []string{"happy"}}},
					Tables: []*plugin.Table{
						{
							Rel: &plugin.Identifier{Name: "test_authors"},
							Columns: []*plugin.Column{
								queryCol("id", "int4", nil),
								queryCol("name", "text", nil),
							},
						},
						{
							Rel: &plugin.Identifier{Name: "test_prefs"},
							Columns: []*plugin.Column{
								queryCol("id", "int4", nil),
								queryCol("mood", "mood", nil),
							},
						},
					},
				},
			},
		},
		Queries: queries,
	}
}

func buildQueries(t *testing.T, conf *config.Config, pluginQueries []*plugin.Query) []model.Query {
	t.Helper()
	tf := transform.NewTransformer(conf, queriesRequest(pluginQueries), types.PostgresTypeToPython)

	return tf.BuildQueries(tf.BuildTables())
}

func buildSingleQuery(t *testing.T, conf *config.Config, pluginQuery *plugin.Query) model.Query {
	t.Helper()
	queries := buildQueries(t, conf, []*plugin.Query{pluginQuery})
	if len(queries) != 1 {
		t.Fatalf("BuildQueries returned %d queries, want 1", len(queries))
	}

	return queries[0]
}

func TestBuildQueriesSkipsNamelessAndCmdless(t *testing.T) {
	t.Parallel()
	queries := buildQueries(t, &config.Config{}, []*plugin.Query{
		{Name: "", Cmd: ":one"},
		{Name: "NoCmd", Cmd: ""},
		{Name: "Ping", Cmd: ":exec", Text: "SELECT 1", Filename: "queries.sql"},
	})

	if len(queries) != 1 {
		t.Fatalf("BuildQueries returned %d queries, want 1 (nameless and cmdless must be skipped)", len(queries))
	}
	if queries[0].QueryName != "Ping" {
		t.Errorf("QueryName = %q, want %q", queries[0].QueryName, "Ping")
	}
}

func TestBuildQueriesModuleName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		filename string
		want     string
	}{
		{name: "extension stripped", filename: "queries.sql", want: "queries"},
		{name: "no dot kept as is", filename: "plain", want: "plain"},
		{name: "only last dot stripped", filename: "multi.dot.sql", want: "multi.dot"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			query := buildSingleQuery(t, &config.Config{}, &plugin.Query{Name: "Ping", Cmd: ":exec", Filename: tc.filename})

			if query.ModuleName != tc.want {
				t.Errorf("ModuleName = %q, want %q", query.ModuleName, tc.want)
			}
			if query.FileName != tc.filename {
				t.Errorf("FileName = %q, want %q", query.FileName, tc.filename)
			}
		})
	}
}

func TestBuildQueriesExecBasics(t *testing.T) {
	t.Parallel()
	query := buildSingleQuery(t, &config.Config{}, &plugin.Query{
		Name:     "MakeAuthor",
		Cmd:      ":exec",
		Text:     "INSERT INTO test_authors (name) VALUES ($1)",
		Filename: "authors.sql",
		Params: []*plugin.Parameter{
			{Number: 1, Column: queryCol("name", "text", nil)},
			{Number: 2, Column: queryCol("", "int4", nil)},
			{Number: 3, Column: queryCol("for", "text", nil)},
		},
	})

	if query.Cmd != ":exec" {
		t.Errorf("Cmd = %q, want %q", query.Cmd, ":exec")
	}
	if query.SQL != "INSERT INTO test_authors (name) VALUES ($1)" {
		t.Errorf("SQL = %q, want the raw query text", query.SQL)
	}
	if query.ConstantName != "MAKE_AUTHOR" {
		t.Errorf("ConstantName = %q, want %q", query.ConstantName, "MAKE_AUTHOR")
	}
	if query.FuncName != "make_author" {
		t.Errorf("FuncName = %q, want %q", query.FuncName, "make_author")
	}
	if want := (model.QueryValue{Type: model.PyType{Type: "None"}}); query.Returns != want {
		t.Errorf("Returns = %+v, want %+v", query.Returns, want)
	}
	wantParams := []model.QueryValue{
		{Name: "name", Type: pyStr, Number: 1},
		{Name: "dollar_2", Type: pyInt, Number: 2},
		{Name: "for_", Type: pyStr, Number: 3},
	}
	if len(query.Params) != len(wantParams) {
		t.Fatalf("Params = %+v, want %d params", query.Params, len(wantParams))
	}
	for i, want := range wantParams {
		if query.Params[i] != want {
			t.Errorf("Params[%d] = %+v, want %+v", i, query.Params[i], want)
		}
	}
}

func TestBuildQueriesImplicitArgCollision(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		emitClasses bool
		driver      config.SQLDriver
		cmd         string
		column      string
		sqlcSlice   bool
		want        string
	}{
		{name: "conn collides in functions mode", emitClasses: false, column: "conn", want: "conn_2"},
		{name: "self is free in functions mode", emitClasses: false, column: "self", want: "self"},
		{name: "self collides in classes mode", emitClasses: true, column: "self", want: "self_2"},
		{name: "conn is free in classes mode", emitClasses: true, column: "conn", want: "conn"},
		// Slice queries write the expanded SQL into a local named "sql"
		// before binding, so the name is reserved exactly there.
		{name: "sql collides in a slice query", emitClasses: false, column: "sql", sqlcSlice: true, want: "sql_2"},
		{name: "sql is free without slices", emitClasses: false, column: "sql", want: "sql"},
		// psycopg query bodies introduce locals of their own: the hoisted
		// params dict, the :execrows cursor, and the :one row.
		{
			name:   "sql_params collides for psycopg",
			driver: config.SQLDriverPsycopgAsync,
			column: "sql_params",
			want:   "sql_params_2",
		},
		{name: "sql_params is free for asyncpg", driver: config.SQLDriverAsyncpg, column: "sql_params", want: "sql_params"},
		{
			name:   "cur collides in a psycopg execrows query",
			driver: config.SQLDriverPsycopgAsync,
			cmd:    ":execrows",
			column: "cur",
			want:   "cur_2",
		},
		{name: "cur is free in a psycopg exec query", driver: config.SQLDriverPsycopgAsync, column: "cur", want: "cur"},
		{
			name:   "row collides in a psycopg one query",
			driver: config.SQLDriverPsycopgAsync,
			cmd:    ":one",
			column: "row",
			want:   "row_2",
		},
		{name: "row is free in a psycopg exec query", driver: config.SQLDriverPsycopgAsync, column: "row", want: "row"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			column := queryCol(tc.column, "int4", nil)
			column.IsSqlcSlice = tc.sqlcSlice
			cmd := tc.cmd
			if cmd == "" {
				cmd = ":exec"
			}
			query := buildSingleQuery(t, &config.Config{EmitClasses: tc.emitClasses, SqlDriver: tc.driver}, &plugin.Query{
				Name:   "Ping",
				Cmd:    cmd,
				Params: []*plugin.Parameter{{Number: 1, Column: column}},
			})

			if query.Params[0].Name != tc.want {
				t.Errorf("param name = %q, want %q", query.Params[0].Name, tc.want)
			}
		})
	}
}

func TestBuildQueriesReturnKindsByCmd(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		cmd  string
		want model.PyType
	}{
		{name: "exec returns None", cmd: ":exec", want: model.PyType{Type: "None"}},
		{name: "execlastid returns optional int", cmd: ":execlastid", want: model.PyType{Type: "int", IsNullable: true}},
		{name: "execrows returns int", cmd: ":execrows", want: model.PyType{Type: "int"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			query := buildSingleQuery(t, &config.Config{}, &plugin.Query{Name: "Ping", Cmd: tc.cmd})

			if query.Returns.Type != tc.want {
				t.Errorf("Returns.Type = %+v, want %+v", query.Returns.Type, tc.want)
			}
		})
	}
}

func TestBuildQueriesCopyFrom(t *testing.T) {
	t.Parallel()
	query := buildSingleQuery(t, &config.Config{}, &plugin.Query{
		Name:            "CopyAuthors",
		Cmd:             ":copyfrom",
		InsertIntoTable: authorsIdent(),
		Params: []*plugin.Parameter{
			{Number: 1, Column: queryCol("id", "int4", nil)},
			{Number: 2, Column: queryCol("name", "text", nil)},
			{Number: 3, Column: queryCol("name", "text", nil)},
		},
	})

	if want := (model.PyType{Type: "int"}); query.Returns.Type != want {
		t.Errorf("Returns.Type = %+v, want %+v", query.Returns.Type, want)
	}
	if query.Table == nil || query.Table.Name != "test_authors" {
		t.Errorf("Table = %+v, want the InsertIntoTable identifier", query.Table)
	}
	if len(query.Params) != 1 {
		t.Fatalf("Params = %+v, want a single bundled params value", query.Params)
	}
	param := query.Params[0]
	if param.Name != "params" || !param.EmitTable {
		t.Errorf("param = %+v, want name %q with EmitTable", param, "params")
	}
	if want := (model.PyType{Type: "CopyAuthorsParams", IsList: true}); param.Type != want {
		t.Errorf("param type = %+v, want %+v", param.Type, want)
	}
	if param.Table == nil {
		t.Fatal("param.Table is nil, want the generated params class")
	}
	if param.Table.Name != "CopyAuthorsParams" {
		t.Errorf("params class name = %q, want %q", param.Table.Name, "CopyAuthorsParams")
	}
	if param.Table.Identifier == nil || param.Table.Identifier.Name != "" {
		t.Errorf("params class identifier = %+v, want an empty identifier", param.Table.Identifier)
	}
	wantColumns := []model.Column{
		{Name: "id_", DBName: "id", Type: pyInt, Number: 1},
		{Name: "name", DBName: "name", Type: pyStr, Number: 2},
		{Name: "name_2", DBName: "name", Type: pyStr, Number: 3},
	}
	if len(param.Table.Columns) != len(wantColumns) {
		t.Fatalf("params class columns = %+v, want %d columns", param.Table.Columns, len(wantColumns))
	}
	for i, want := range wantColumns {
		if param.Table.Columns[i] != want {
			t.Errorf("params class column[%d] = %+v, want %+v", i, param.Table.Columns[i], want)
		}
	}
}

func TestBuildQueriesParamsClassOverLimit(t *testing.T) {
	t.Parallel()
	conf := &config.Config{QueryParameterLimit: utils.ToPtr(0)}
	query := buildSingleQuery(t, conf, &plugin.Query{
		Name:   "TouchAuthor",
		Cmd:    ":exec",
		Params: []*plugin.Parameter{{Number: 1, Column: queryCol("name", "text", nil)}},
	})

	if len(query.Params) != 1 {
		t.Fatalf("Params = %+v, want a single bundled params value", query.Params)
	}
	param := query.Params[0]
	if want := (model.PyType{Type: "TouchAuthorParams", IsList: false}); param.Type != want {
		t.Errorf("param type = %+v, want %+v (no list outside copyfrom)", param.Type, want)
	}
	if !param.EmitTable || param.Table == nil || param.Table.Name != "TouchAuthorParams" {
		t.Errorf("param = %+v, want an emitted TouchAuthorParams class", param)
	}
}

func TestBuildQueriesOneSingleColumn(t *testing.T) {
	t.Parallel()
	query := buildSingleQuery(t, &config.Config{}, &plugin.Query{
		Name:    "GetAuthorID",
		Cmd:     ":one",
		Columns: []*plugin.Column{queryCol("id", "int4", authorsIdent())},
	})

	if want := (model.QueryValue{Type: pyInt}); query.Returns != want {
		t.Errorf("Returns = %+v, want the bare column type %+v", query.Returns, want)
	}
}

func TestBuildQueriesManyMatchesModel(t *testing.T) {
	t.Parallel()
	query := buildSingleQuery(t, &config.Config{}, &plugin.Query{
		Name: "ListAuthors",
		Cmd:  ":many",
		Columns: []*plugin.Column{
			queryCol("id", "int4", authorsIdent()),
			queryCol("name", "text", authorsIdent()),
		},
	})

	if query.Returns.EmitTable {
		t.Error("Returns.EmitTable = true, want false for a matched model")
	}
	if query.Returns.Table == nil || query.Returns.Table.Name != "TestAuthor" {
		t.Errorf("Returns.Table = %+v, want the TestAuthor model", query.Returns.Table)
	}
	if want := (model.PyType{Type: "models.TestAuthor"}); query.Returns.Type != want {
		t.Errorf("Returns.Type = %+v, want %+v", query.Returns.Type, want)
	}
}

func TestBuildQueriesRowClassWhenNoModelMatches(t *testing.T) {
	t.Parallel()
	prefsIdent := func() *plugin.Identifier { return &plugin.Identifier{Name: "test_prefs"} }
	arrayCol := queryCol("name", "text", authorsIdent())
	arrayCol.IsArray = true
	nullableCol := queryCol("name", "text", authorsIdent())
	nullableCol.NotNull = false
	// buildPyType looks the enum up by the raw identifier name while the
	// conversion func parses "schema.name", so this column keeps the enum
	// Python type but loses the IsEnum flag - only that flag differs.
	fakeEnumCol := queryCol("mood", "public.mood", prefsIdent())
	cases := []struct {
		name    string
		columns []*plugin.Column
	}{
		{name: "column name differs", columns: []*plugin.Column{
			queryCol("id", "int4", authorsIdent()),
			queryCol("title", "text", authorsIdent()),
		}},
		{name: "python type differs", columns: []*plugin.Column{
			queryCol("id", "int4", authorsIdent()),
			queryCol("name", "int4", authorsIdent()),
		}},
		{name: "nullability differs", columns: []*plugin.Column{
			queryCol("id", "int4", authorsIdent()),
			nullableCol,
		}},
		{name: "list flag differs", columns: []*plugin.Column{
			queryCol("id", "int4", authorsIdent()),
			arrayCol,
		}},
		{name: "enum flag differs", columns: []*plugin.Column{
			queryCol("id", "int4", prefsIdent()),
			fakeEnumCol,
		}},
		{name: "source table missing", columns: []*plugin.Column{
			queryCol("id", "int4", nil),
			queryCol("name", "text", nil),
		}},
		{name: "column count differs", columns: []*plugin.Column{
			queryCol("id", "int4", authorsIdent()),
			queryCol("name", "text", authorsIdent()),
			queryCol("extra", "text", nil),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			query := buildSingleQuery(t, &config.Config{}, &plugin.Query{Name: "FetchData", Cmd: ":many", Columns: tc.columns})

			if !query.Returns.EmitTable {
				t.Error("Returns.EmitTable = false, want an emitted row class")
			}
			if query.Returns.Table == nil || query.Returns.Table.Name != "FetchDataRow" {
				t.Fatalf("Returns.Table = %+v, want the FetchDataRow class", query.Returns.Table)
			}
			if want := (model.PyType{Type: "FetchDataRow"}); query.Returns.Type != want {
				t.Errorf("Returns.Type = %+v, want %+v", query.Returns.Type, want)
			}
		})
	}
}

func TestBuildQueriesRowClassColumns(t *testing.T) {
	t.Parallel()
	// Duplicate column names must dedup identically during table matching
	// and row class construction.
	query := buildSingleQuery(t, &config.Config{}, &plugin.Query{
		Name: "PairIDs",
		Cmd:  ":many",
		Columns: []*plugin.Column{
			queryCol("id", "int4", authorsIdent()),
			queryCol("id", "int4", authorsIdent()),
		},
	})

	if query.Returns.Table == nil {
		t.Fatal("Returns.Table is nil, want the PairIDsRow class")
	}
	// "id" is a Python builtin, so it escapes to "id_" before dedup.
	wantColumns := []model.Column{
		{Name: "id_", DBName: "id", Type: pyInt},
		{Name: "id__2", DBName: "id", Type: pyInt},
	}
	if len(query.Returns.Table.Columns) != len(wantColumns) {
		t.Fatalf("row columns = %+v, want %d columns", query.Returns.Table.Columns, len(wantColumns))
	}
	for i, want := range wantColumns {
		if query.Returns.Table.Columns[i] != want {
			t.Errorf("row column[%d] = %+v, want %+v", i, query.Returns.Table.Columns[i], want)
		}
	}
}

func TestBuildQueriesEmbeds(t *testing.T) {
	t.Parallel()
	authorEmbed := queryCol("test_authors", "int4", nil)
	authorEmbed.EmbedTable = authorsIdent()
	prefEmbed := queryCol("test_prefs", "int4", nil)
	prefEmbed.EmbedTable = &plugin.Identifier{Name: "test_prefs"}
	query := buildSingleQuery(t, &config.Config{}, &plugin.Query{
		Name:    "GetAuthorWithPref",
		Cmd:     ":one",
		Columns: []*plugin.Column{queryCol("id", "int4", authorsIdent()), authorEmbed, prefEmbed},
	})

	if query.Returns.Table == nil || query.Returns.Table.Name != "GetAuthorWithPrefRow" {
		t.Fatalf("Returns.Table = %+v, want the GetAuthorWithPrefRow class", query.Returns.Table)
	}
	columns := query.Returns.Table.Columns
	if len(columns) != 3 {
		t.Fatalf("row columns = %+v, want 3 columns", columns)
	}
	if want := (model.Column{Name: "id_", DBName: "id", Type: pyInt}); columns[0] != want {
		t.Errorf("row column[0] = %+v, want %+v", columns[0], want)
	}
	embedCases := []struct {
		index       int
		fieldName   string
		dbName      string
		modelName   string
		wantColumns []model.Column
	}{
		{
			index: 1, fieldName: "test_author", dbName: "test_authors", modelName: "TestAuthor",
			wantColumns: []model.Column{
				{Name: "id_", DBName: "id", Type: pyInt},
				{Name: "name", DBName: "name", Type: pyStr},
			},
		},
		{
			index: 2, fieldName: "test_pref", dbName: "test_prefs", modelName: "TestPref",
			wantColumns: []model.Column{
				{Name: "id_", DBName: "id", Type: pyInt},
				{Name: "mood", DBName: "mood", Type: model.PyType{
					SQLType: "mood", Type: "enums.Mood", IsEnum: true, DefaultType: "enums.Mood",
				}},
			},
		},
	}
	for _, tc := range embedCases {
		column := columns[tc.index]
		if column.Name != tc.fieldName || column.DBName != tc.dbName {
			t.Errorf("row column[%d] = %+v, want name %q for db column %q", tc.index, column, tc.fieldName, tc.dbName)
		}
		if column.Type.Type != "models."+tc.modelName {
			t.Errorf("row column[%d] type = %q, want %q", tc.index, column.Type.Type, "models."+tc.modelName)
		}
		if column.Embed == nil || column.Embed.ModelName != tc.modelName {
			t.Fatalf("row column[%d] embed = %+v, want model %q", tc.index, column.Embed, tc.modelName)
		}
		if !reflect.DeepEqual(column.Embed.Columns, tc.wantColumns) {
			t.Errorf("row column[%d] embed columns = %+v, want %+v", tc.index, column.Embed.Columns, tc.wantColumns)
		}
	}
}

func TestBuildQueriesEmbedUnknownTable(t *testing.T) {
	t.Parallel()
	ghost := queryCol("data", "text", nil)
	ghost.EmbedTable = &plugin.Identifier{Name: "missing"}
	// A single embed column must not take the bare-column return path, and
	// an embed of an unknown table degrades to a plain column.
	query := buildSingleQuery(t, &config.Config{}, &plugin.Query{
		Name:    "GetGhost",
		Cmd:     ":one",
		Columns: []*plugin.Column{ghost},
	})

	if query.Returns.Table == nil || query.Returns.Table.Name != "GetGhostRow" {
		t.Fatalf("Returns.Table = %+v, want the GetGhostRow class", query.Returns.Table)
	}
	if len(query.Returns.Table.Columns) != 1 {
		t.Fatalf("row columns = %+v, want 1 column", query.Returns.Table.Columns)
	}
	if want := (model.Column{Name: "data", DBName: "data", Type: pyStr}); query.Returns.Table.Columns[0] != want {
		t.Errorf("row column[0] = %+v, want plain column %+v", query.Returns.Table.Columns[0], want)
	}
}

func TestBuildQueriesPsycopgSQLRewrite(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		driver  config.SQLDriver
		query   *plugin.Query
		wantSQL string
	}{
		{
			name:   "parameterized query is rewritten for psycopg",
			driver: config.SQLDriverPsycopgAsync,
			query: &plugin.Query{
				Name: "GetAuthor",
				Cmd:  ":one",
				Text: "SELECT name FROM test_authors WHERE id = $1 AND name LIKE 'a%'",
				Params: []*plugin.Parameter{
					{Number: 1, Column: queryCol("id", "int4", nil)},
				},
				Columns: []*plugin.Column{queryCol("name", "text", nil)},
			},
			wantSQL: "SELECT name FROM test_authors WHERE id = %(p1)s AND name LIKE 'a%%'",
		},
		{
			name:   "parameterless query stays untouched",
			driver: config.SQLDriverPsycopgAsync,
			query: &plugin.Query{
				Name:    "CountAuthors",
				Cmd:     ":one",
				Text:    "SELECT count(*) FROM test_authors WHERE name LIKE 'a%'",
				Columns: []*plugin.Column{queryCol("count", "int8", nil)},
			},
			wantSQL: "SELECT count(*) FROM test_authors WHERE name LIKE 'a%'",
		},
		{
			name:   "copyfrom stays untouched",
			driver: config.SQLDriverPsycopgAsync,
			query: &plugin.Query{
				Name: "CopyAuthors",
				Cmd:  ":copyfrom",
				Text: "INSERT INTO test_authors (id) VALUES ($1)",
				Params: []*plugin.Parameter{
					{Number: 1, Column: queryCol("id", "int4", nil)},
				},
				InsertIntoTable: &plugin.Identifier{Name: "test_authors"},
			},
			wantSQL: "INSERT INTO test_authors (id) VALUES ($1)",
		},
		{
			name:   "asyncpg keeps native placeholders",
			driver: config.SQLDriverAsyncpg,
			query: &plugin.Query{
				Name: "GetAuthor",
				Cmd:  ":one",
				Text: "SELECT name FROM test_authors WHERE id = $1",
				Params: []*plugin.Parameter{
					{Number: 1, Column: queryCol("id", "int4", nil)},
				},
				Columns: []*plugin.Column{queryCol("name", "text", nil)},
			},
			wantSQL: "SELECT name FROM test_authors WHERE id = $1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			query := buildSingleQuery(t, &config.Config{SqlDriver: tc.driver}, tc.query)
			if query.SQL != tc.wantSQL {
				t.Errorf("SQL = %q, want %q", query.SQL, tc.wantSQL)
			}
		})
	}
}
