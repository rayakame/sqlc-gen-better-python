package writer_test

import (
	"strings"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

// methodLvl is the indent level of a method body inside a class.
const methodLvl = 2

// docCase drives one docstring emission through a fresh writer and compares
// the exact output.
type docCase struct {
	name    string
	conv    config.DocstringConvention
	driver  config.SQLDriver
	omitSQL bool
	write   func(w *writer.CodeWriter)
	want    string
}

func newDocWriter(conv config.DocstringConvention, driver config.SQLDriver, emitSQL bool) *writer.CodeWriter {
	return writer.NewCodeWriter(&config.Config{
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
		EmitDocstrings:      conv,
		SqlDriver:           driver,
		EmitDocstringsSQL:   &emitSQL,
	})
}

func runDocCases(t *testing.T, cases []docCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			driver := tc.driver
			if driver == "" {
				driver = config.SQLDriverAsyncpg
			}
			w := newDocWriter(tc.conv, driver, !tc.omitSQL)
			tc.write(w)
			if got := w.String(); got != tc.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

// lines joins output lines with newlines and adds the trailing newline every
// emitted docstring ends with.
func lines(ls ...string) string {
	return strings.Join(ls, "\n") + "\n"
}

func TestDocstringsEnabled(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		conv config.DocstringConvention
		want bool
	}{
		{name: "none", conv: config.DocstringConventionNone, want: false},
		{name: "google", conv: config.DocstringConventionGoogle, want: true},
		{name: "numpy", conv: config.DocstringConventionNumpy, want: true},
		{name: "pep257", conv: config.DocstringConventionPEP257, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := newDocWriter(tc.conv, config.SQLDriverAsyncpg, true)
			if got := w.DocstringsEnabled(); got != tc.want {
				t.Errorf("DocstringsEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestModuleDocstrings(t *testing.T) {
	t.Parallel()
	runDocCases(t, []docCase{
		{
			name:  "model file none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteModelFileModuleDocstring() },
			want:  "",
		},
		{
			name:  "model file google",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteModelFileModuleDocstring() },
			want:  lines(`"""Module containing models."""`, ``),
		},
		{
			name:  "enums file none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteEnumsFileModuleDocstring() },
			want:  "",
		},
		{
			name:  "enums file numpy",
			conv:  config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) { w.WriteEnumsFileModuleDocstring() },
			want:  lines(`"""Module containing enums."""`, ``),
		},
		{
			name:  "query file none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteQueryFileModuleDocstring("authors.sql") },
			want:  "",
		},
		{
			name:  "query file pep257",
			conv:  config.DocstringConventionPEP257,
			write: func(w *writer.CodeWriter) { w.WriteQueryFileModuleDocstring("authors.sql") },
			want:  lines(`"""Module containing queries from file authors.sql."""`, ``),
		},
		{
			name:  "init file none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteInitFileModuleDocstring() },
			want:  "",
		},
		{
			name:  "init file google",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteInitFileModuleDocstring() },
			want:  lines(`"""Package containing queries and models automatically generated using sqlc-gen-better-python."""`),
		},
	})
}

func TestWriteModelClassDocstring(t *testing.T) {
	t.Parallel()
	table := &model.Table{Name: "Author", Columns: []model.Column{
		{Name: "id", Type: model.PyType{Type: "int"}},
		{Name: "tags", Type: model.PyType{Type: "str", IsList: true, IsNullable: true}},
	}}
	runDocCases(t, []docCase{
		{
			name:  "none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteModelClassDocstring(table) },
			want:  "",
		},
		{
			name:  "numpy",
			conv:  config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) { w.WriteModelClassDocstring(table) },
			want: lines(
				`    """Model representing Author.`,
				``,
				`    Attributes`,
				`    ----------`,
				`    id : int`,
				`    tags : collections.abc.Sequence[str] | None`,
				``,
				`    """`,
				``,
			),
		},
		{
			name:  "google",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteModelClassDocstring(table) },
			want: lines(
				`    """Model representing Author.`,
				``,
				`    Attributes:`,
				`        id: int`,
				`        tags: collections.abc.Sequence[str] | None`,
				`    """`,
				``,
			),
		},
		{
			name:  "pep257",
			conv:  config.DocstringConventionPEP257,
			write: func(w *writer.CodeWriter) { w.WriteModelClassDocstring(table) },
			want: lines(
				`    """Model representing Author.`,
				``,
				`    Attributes:`,
				`    id -- int`,
				`    tags -- collections.abc.Sequence[str] | None`,
				`    """`,
				``,
			),
		},
		{
			name: "numpy without columns",
			conv: config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) {
				w.WriteModelClassDocstring(&model.Table{Name: "Empty"})
			},
			want: lines(
				`    """Model representing Empty.`,
				``,
				`    Attributes`,
				`    ----------`,
				``,
				`    """`,
				``,
			),
		},
	})
}

func TestWriteEnumClassDocstring(t *testing.T) {
	t.Parallel()
	runDocCases(t, []docCase{
		{
			name:  "none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteEnumClassDocstring("BookStatus") },
			want:  "",
		},
		{
			name:  "google",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteEnumClassDocstring("BookStatus") },
			want:  lines(`    """Enum representing BookStatus."""`, ``),
		},
	})
}

func TestWriteQueryClassDocstring(t *testing.T) {
	t.Parallel()
	runDocCases(t, []docCase{
		{
			name:  "none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassDocstring("authors.sql", "asyncpg.Connection") },
			want:  "",
		},
		{
			name:  "numpy has parameters section",
			conv:  config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassDocstring("authors.sql", "asyncpg.Connection") },
			want: lines(
				`    """Queries from file authors.sql.`,
				``,
				`    Parameters`,
				`    ----------`,
				`    conn : asyncpg.Connection`,
				`        The connection object used to execute queries.`,
				``,
				`    """`,
				``,
			),
		},
		{
			name:  "google stays one line",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassDocstring("authors.sql", "asyncpg.Connection") },
			want:  lines(`    """Queries from file authors.sql."""`, ``),
		},
	})
}

func TestWriteQueryClassInitDocstring(t *testing.T) {
	t.Parallel()
	runDocCases(t, []docCase{
		{
			name:  "none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassInitDocstring(methodLvl, "asyncpg.Connection") },
			want:  "",
		},
		{
			name:  "numpy",
			conv:  config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassInitDocstring(methodLvl, "asyncpg.Connection") },
			want:  lines(`        """Initialize the instance using the connection."""`),
		},
		{
			name:  "google",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassInitDocstring(methodLvl, "asyncpg.Connection") },
			want: lines(
				`        """Initialize the instance using the connection.`,
				``,
				`        Args:`,
				`            conn:`,
				"                Connection object of type `asyncpg.Connection` used to execute the query.",
				`        """`,
			),
		},
		{
			name:  "pep257",
			conv:  config.DocstringConventionPEP257,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassInitDocstring(methodLvl, "asyncpg.Connection") },
			want: lines(
				`        """Initialize the instance using the connection.`,
				``,
				`        Arguments:`,
				"        conn -- Connection object of type `asyncpg.Connection` used to execute queries.",
				`        """`,
			),
		},
	})
}

func TestWriteQueryClassConnDocstring(t *testing.T) {
	t.Parallel()
	runDocCases(t, []docCase{
		{
			name:  "none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassConnDocstring("asyncpg.Connection") },
			want:  "",
		},
		{
			name:  "numpy",
			conv:  config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassConnDocstring("asyncpg.Connection") },
			want: lines(
				`        """Connection object used to make queries.`,
				``,
				`        Returns`,
				`        -------`,
				`        asyncpg.Connection`,
				``,
				`        """`,
			),
		},
		{
			name:  "google",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassConnDocstring("asyncpg.Connection") },
			want: lines(
				`        """Connection object used to make queries.`,
				``,
				`        Returns:`,
				"            Connection object of type `asyncpg.Connection` used to make queries.",
				`        """`,
			),
		},
		{
			name:  "pep257",
			conv:  config.DocstringConventionPEP257,
			write: func(w *writer.CodeWriter) { w.WriteQueryClassConnDocstring("asyncpg.Connection") },
			want: lines(
				`        """Connection object used to make queries.`,
				``,
				`        Returns:`,
				`        asyncpg.Connection -- Connection object used to make queries.`,
				`        """`,
			),
		},
	})
}

func TestWriteQueryResultsClassDocstring(t *testing.T) {
	t.Parallel()
	runDocCases(t, []docCase{
		{
			name: "none",
			conv: config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryResultsClassDocstring("asyncpg.Connection", "asyncpg.Record", false)
			},
			want: "",
		},
		{
			name: "numpy has parameters section",
			conv: config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryResultsClassDocstring("asyncpg.Connection", "asyncpg.Record", false)
			},
			want: lines(
				`    """Helper class that allows both iteration and normal fetching of data from the db.`,
				``,
				`    Parameters`,
				`    ----------`,
				`    conn`,
				"        The connection object of type `asyncpg.Connection` used to execute queries.",
				`    sql`,
				`        The SQL statement that will be executed when fetching/iterating.`,
				`    decode_hook`,
				"        A callback that turns an `asyncpg.Record` object into `T` that will be returned.",
				`    *args`,
				`        Arguments that should be sent when executing the sql query.`,
				``,
				`    """`,
				``,
			),
		},
		{
			name: "pep257 stays one line",
			conv: config.DocstringConventionPEP257,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryResultsClassDocstring("asyncpg.Connection", "asyncpg.Record", false)
			},
			want: lines(`    """Helper class that allows both iteration and normal fetching of data from the db."""`, ``),
		},
	})
}

func TestWriteQueryResultsInitDocstring(t *testing.T) {
	t.Parallel()
	runDocCases(t, []docCase{
		{
			name: "none",
			conv: config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryResultsInitDocstring("asyncpg.Connection", "asyncpg.Record", false)
			},
			want: "",
		},
		{
			name: "numpy",
			conv: config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryResultsInitDocstring("asyncpg.Connection", "asyncpg.Record", false)
			},
			want: lines(`        """Initialize the QueryResults instance."""`),
		},
		{
			name: "google",
			conv: config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryResultsInitDocstring("asyncpg.Connection", "asyncpg.Record", false)
			},
			want: lines(
				`        """Initialize the QueryResults instance.`,
				``,
				`        Args:`,
				`            conn:`,
				"                The connection object of type `asyncpg.Connection` used to execute queries.",
				`            sql:`,
				`                The SQL statement that will be executed when fetching/iterating.`,
				`            decode_hook:`,
				"                A callback that turns an `asyncpg.Record` object into `T` that will be returned.",
				`            *args:`,
				`                Arguments that should be sent when executing the sql query.`,
				`        """`,
			),
		},
		{
			name: "pep257",
			conv: config.DocstringConventionPEP257,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryResultsInitDocstring("asyncpg.Connection", "asyncpg.Record", false)
			},
			want: lines(
				`        """Initialize the QueryResults instance.`,
				``,
				`        Arguments:`,
				"        conn -- The connection object of type `asyncpg.Connection` used to execute queries.",
				`        sql -- The SQL statement that will be executed when fetching/iterating.`,
				"        decode_hook -- A callback that turns an `asyncpg.Record` object into `T` that will be returned.",
				`        *args -- Arguments that should be sent when executing the sql query.`,
				`        """`,
			),
		},
	})
}

func TestWriteQueryResultsIterAndFetchDocstrings(t *testing.T) {
	t.Parallel()
	runDocCases(t, []docCase{
		{
			name:  "iter none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsIterDocstring(false) },
			want:  "",
		},
		{
			name:  "iter numpy sync",
			conv:  config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsIterDocstring(false) },
			want: lines(
				`        """Initialize iteration support.`,
				``,
				`        Returns`,
				`        -------`,
				`        QueryResults[T]`,
				`            Self as an iterator.`,
				`        """`,
			),
		},
		{
			name:  "iter google async",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsIterDocstring(true) },
			want: lines(
				"        \"\"\"Initialize iteration support for `async for`.",
				``,
				`        Returns:`,
				`            Self as an asynchronous iterator.`,
				`        """`,
			),
		},
		{
			name:  "iter pep257 sync",
			conv:  config.DocstringConventionPEP257,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsIterDocstring(false) },
			want: lines(
				`        """Initialize iteration support.`,
				``,
				`        Returns:`,
				`        Self as an iterator.`,
				`        """`,
			),
		},
		{
			name:  "fetch none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsFetchDocstring(false) },
			want:  "",
		},
		{
			name:  "fetch numpy async",
			conv:  config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsFetchDocstring(true) },
			want: lines(
				"        \"\"\"Allow `await` on the object to return all rows as a fully decoded sequence.",
				``,
				`        Returns`,
				`        -------`,
				`        collections.abc.Sequence[T]`,
				"            A sequence of decoded objects of type `T`.",
				`        """`,
			),
		},
		{
			name:  "fetch google sync",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsFetchDocstring(false) },
			want: lines(
				`        """Allow calling the object to return all rows as a fully decoded sequence.`,
				``,
				`        Returns:`,
				"            A sequence of decoded objects of type `T`.",
				`        """`,
			),
		},
		{
			name:  "fetch pep257 sync",
			conv:  config.DocstringConventionPEP257,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsFetchDocstring(false) },
			want: lines(
				`        """Allow calling the object to return all rows as a fully decoded sequence.`,
				``,
				`        Returns:`,
				"        A sequence of decoded objects of type `T`.",
				`        """`,
			),
		},
	})
}

func TestWriteQueryResultsNextDocstring(t *testing.T) {
	t.Parallel()
	runDocCases(t, []docCase{
		{
			name:  "none",
			conv:  config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsNextDocstring("an asyncpg cursor", true) },
			want:  "",
		},
		{
			name:  "numpy async",
			conv:  config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsNextDocstring("an asyncpg cursor", true) },
			want: lines(
				`        """Yield the next item in the query result using an asyncpg cursor.`,
				``,
				`        Returns`,
				`        -------`,
				`        T`,
				`            The next decoded result.`,
				``,
				`        Raises`,
				`        ------`,
				`        StopAsyncIteration`,
				`            When no more records are available.`,
				`        """`,
			),
		},
		{
			name:  "google sync",
			conv:  config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsNextDocstring("a sqlite3 cursor", false) },
			want: lines(
				`        """Yield the next item in the query result using a sqlite3 cursor.`,
				``,
				`        Returns:`,
				"            The next decoded result of type `T`.",
				``,
				`        Raises:`,
				`            StopIteration: When no more records are available.`,
				`        """`,
			),
		},
		{
			name:  "pep257 async",
			conv:  config.DocstringConventionPEP257,
			write: func(w *writer.CodeWriter) { w.WriteQueryResultsNextDocstring("an asyncpg cursor", true) },
			want: lines(
				`        """Yield the next item in the query result using an asyncpg cursor.`,
				``,
				`        Returns:`,
				"        The next decoded result of type `T`.",
				``,
				`        Raises:`,
				`        StopAsyncIteration -- When no more records are available.`,
				`        """`,
			),
		},
	})
}

func TestWriteQueryFunctionDocstring(t *testing.T) {
	t.Parallel()
	execQuery := &model.Query{
		Cmd:       metadata.CmdExec,
		QueryName: "DeleteAuthor",
		SQL:       "DELETE FROM authors WHERE id = $1",
	}
	execRowsQuery := &model.Query{
		Cmd:       metadata.CmdExecRows,
		QueryName: "TouchAuthors",
		SQL:       "UPDATE authors SET name = $1",
	}
	runDocCases(t, []docCase{
		{
			name: "none",
			conv: config.DocstringConventionNone,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryFunctionDocstring(1, execQuery, "asyncpg.Connection", nil, "")
			},
			want: "",
		},
		{
			name: "unsupported command writes nothing",
			conv: config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) {
				batch := &model.Query{Cmd: metadata.CmdBatchExec, QueryName: "BatchDelete", SQL: "SELECT 1"}
				w.WriteQueryFunctionDocstring(1, batch, "asyncpg.Connection", nil, "")
			},
			want: "",
		},
		{
			name: "exec google with conn and arg",
			conv: config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) {
				args := []writer.DocArg{{Name: "author_id", Type: "int"}}
				w.WriteQueryFunctionDocstring(1, execQuery, "asyncpg.Connection", args, "")
			},
			want: lines(
				"    \"\"\"Execute SQL query with `name: DeleteAuthor :exec`.",
				``,
				"    ```sql",
				`    DELETE FROM authors WHERE id = $1`,
				"    ```",
				``,
				`    Args:`,
				`        conn:`,
				"            Connection object of type `asyncpg.Connection` used to execute the query.",
				`        author_id: int.`,
				`    """`,
			),
		},
		{
			name: "exec numpy keeps blank line after parameters",
			conv: config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) {
				args := []writer.DocArg{{Name: "author_id", Type: "int"}}
				w.WriteQueryFunctionDocstring(1, execQuery, "asyncpg.Connection", args, "")
			},
			want: lines(
				"    \"\"\"Execute SQL query with `name: DeleteAuthor :exec`.",
				``,
				"    ```sql",
				`    DELETE FROM authors WHERE id = $1`,
				"    ```",
				``,
				`    Parameters`,
				`    ----------`,
				`    conn : asyncpg.Connection`,
				"        Connection object of type `asyncpg.Connection` used to execute the query.",
				`    author_id : int`,
				``,
				`    """`,
			),
		},
		{
			name: "exec google without conn and args",
			conv: config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) {
				query := &model.Query{Cmd: metadata.CmdExec, QueryName: "Cleanup", SQL: "DELETE FROM logs"}
				w.WriteQueryFunctionDocstring(1, query, "", nil, "")
			},
			want: lines(
				"    \"\"\"Execute SQL query with `name: Cleanup :exec`.",
				``,
				"    ```sql",
				`    DELETE FROM logs`,
				"    ```",
				``,
				`    """`,
			),
		},
		{
			name: "execrows asyncpg google",
			conv: config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryFunctionDocstring(1, execRowsQuery, "asyncpg.Connection", nil, "int")
			},
			want: lines(
				"    \"\"\"Execute SQL query with `name: TouchAuthors :execrows` and return the number of affected rows.",
				``,
				"    ```sql",
				`    UPDATE authors SET name = $1`,
				"    ```",
				``,
				`    Args:`,
				`        conn:`,
				"            Connection object of type `asyncpg.Connection` used to execute the query.",
				``,
				`    Returns:`,
				"        The number (`int`) of affected rows. This will be 0 for queries like `CREATE TABLE`.",
				`    """`,
			),
		},
		{
			name:    "execrows aiosqlite numpy without sql",
			conv:    config.DocstringConventionNumpy,
			driver:  config.SQLDriverAioSQLite,
			omitSQL: true,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryFunctionDocstring(1, execRowsQuery, "aiosqlite.Connection", nil, "int")
			},
			want: lines(
				"    \"\"\"Execute SQL query with `name: TouchAuthors :execrows` and return the number of affected rows.",
				``,
				`    Parameters`,
				`    ----------`,
				`    conn : aiosqlite.Connection`,
				"        Connection object of type `aiosqlite.Connection` used to execute the query.",
				``,
				`    Returns`,
				`    -------`,
				`    int`,
				"        The number of affected rows. This will be -1 for queries like `CREATE TABLE`.",
				``,
				`    """`,
			),
		},
		{
			name:    "execrows psycopg_sync documents rowcount's -1",
			conv:    config.DocstringConventionGoogle,
			driver:  config.SQLDriverPsycopgSync,
			omitSQL: true,
			write: func(w *writer.CodeWriter) {
				w.WriteQueryFunctionDocstring(1, execRowsQuery, "ConnectionLike", nil, "int")
			},
			want: lines(
				"    \"\"\"Execute SQL query with `name: TouchAuthors :execrows` and return the number of affected rows.",
				``,
				`    Args:`,
				`        conn:`,
				"            Connection object of type `ConnectionLike` used to execute the query.",
				``,
				`    Returns:`,
				"        The number (`int`) of affected rows. This will be -1 for queries like `CREATE TABLE`.",
				`    """`,
			),
		},
		{
			name:   "execrows sqlite3 pep257 normalizes sql lines",
			conv:   config.DocstringConventionPEP257,
			driver: config.SQLDriverSQLite,
			write: func(w *writer.CodeWriter) {
				query := &model.Query{
					Cmd:       metadata.CmdExecRows,
					QueryName: "RenameAuthor",
					SQL:       "UPDATE authors\r\nSET name = ?\r\n   \r\nWHERE id = ?",
				}
				args := []writer.DocArg{{Name: "id", Type: "int"}}
				w.WriteQueryFunctionDocstring(1, query, "sqlite3.Connection", args, "int")
			},
			want: lines(
				"    \"\"\"Execute SQL query with `name: RenameAuthor :execrows` and return the number of affected rows.",
				``,
				"    ```sql",
				`    UPDATE authors`,
				`    SET name = ?`,
				``,
				`    WHERE id = ?`,
				"    ```",
				``,
				`    Arguments:`,
				"    conn -- Connection object of type `sqlite3.Connection` used to execute the query.",
				`    id -- int.`,
				``,
				`    Returns:`,
				"    int -- The number of affected rows. This will be -1 for queries like `CREATE TABLE`.",
				`    """`,
			),
		},
		{
			name: "copyfrom google omits sql and adds extra line",
			conv: config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) {
				query := &model.Query{Cmd: metadata.CmdCopyFrom, QueryName: "Copy", SQL: "COPY authors FROM STDIN"}
				args := []writer.DocArg{{
					Name:  "rows",
					Type:  "collections.abc.Sequence[CopyParams]",
					Extra: "Each element will be inserted as one row.",
				}}
				w.WriteQueryFunctionDocstring(1, query, "asyncpg.Connection", args, "int")
			},
			want: lines(
				"    \"\"\"Execute COPY FROM query to insert rows into a table with `name: Copy :copyfrom` "+
					"and return the number of affected rows.",
				``,
				`    Args:`,
				`        conn:`,
				"            Connection object of type `asyncpg.Connection` used to execute the query.",
				`        rows: collections.abc.Sequence[CopyParams].`,
				`            Each element will be inserted as one row.`,
				``,
				`    Returns:`,
				"        The number (`int`) of affected rows.",
				`    """`,
			),
		},
		{
			name: "execresult numpy without conn and with extra line",
			conv: config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) {
				query := &model.Query{Cmd: metadata.CmdExecResult, QueryName: "GetStatus", SQL: "SELECT 1"}
				args := []writer.DocArg{{Name: "id", Type: "int", Extra: "The author id."}}
				w.WriteQueryFunctionDocstring(1, query, "", args, "str")
			},
			want: lines(
				"    \"\"\"Execute and return the result of SQL query with `name: GetStatus :execresult`.",
				``,
				"    ```sql",
				`    SELECT 1`,
				"    ```",
				``,
				`    Parameters`,
				`    ----------`,
				`    id : int`,
				`        The author id.`,
				``,
				`    Returns`,
				`    -------`,
				`    str`,
				`        The result returned when executing the query.`,
				``,
				`    """`,
			),
		},
		{
			name:   "execlastid pep257 appends extra to arg line",
			conv:   config.DocstringConventionPEP257,
			driver: config.SQLDriverAioSQLite,
			write: func(w *writer.CodeWriter) {
				query := &model.Query{
					Cmd:       metadata.CmdExecLastId,
					QueryName: "InsertAuthor",
					SQL:       "INSERT INTO authors (name) VALUES (?)",
				}
				args := []writer.DocArg{{Name: "name", Type: "str", Extra: "Name of the author."}}
				w.WriteQueryFunctionDocstring(1, query, "", args, "int")
			},
			want: lines(
				"    \"\"\"Execute SQL query with `name: InsertAuthor :execlastid` and return the id of the last affected row.",
				``,
				"    ```sql",
				`    INSERT INTO authors (name) VALUES (?)`,
				"    ```",
				``,
				`    Arguments:`,
				`    name -- str. Name of the author.`,
				``,
				`    Returns:`,
				"    int -- The id of the last affected row. Will be `None` if no rows are affected.",
				`    """`,
			),
		},
		{
			name: "one google without conn and args",
			conv: config.DocstringConventionGoogle,
			write: func(w *writer.CodeWriter) {
				query := &model.Query{Cmd: metadata.CmdOne, QueryName: "GetFirstAuthor", SQL: "SELECT id FROM authors LIMIT 1"}
				w.WriteQueryFunctionDocstring(1, query, "", nil, "models.Author")
			},
			want: lines(
				"    \"\"\"Fetch one from the db using the SQL query with `name: GetFirstAuthor :one`.",
				``,
				"    ```sql",
				`    SELECT id FROM authors LIMIT 1`,
				"    ```",
				``,
				`    Returns:`,
				"        Result of type `models.Author` fetched from the db. Will be `None` if not found.",
				`    """`,
			),
		},
		{
			name:    "one google at method level without sql",
			conv:    config.DocstringConventionGoogle,
			omitSQL: true,
			write: func(w *writer.CodeWriter) {
				query := &model.Query{Cmd: metadata.CmdOne, QueryName: "GetAuthor", SQL: "SELECT id FROM authors WHERE id = $1"}
				args := []writer.DocArg{{Name: "author_id", Type: "int"}}
				w.WriteQueryFunctionDocstring(methodLvl, query, "", args, "models.Author")
			},
			want: lines(
				"        \"\"\"Fetch one from the db using the SQL query with `name: GetAuthor :one`.",
				``,
				`        Args:`,
				`            author_id: int.`,
				``,
				`        Returns:`,
				"            Result of type `models.Author` fetched from the db. Will be `None` if not found.",
				`        """`,
			),
		},
		{
			name: "many numpy wraps ret type in QueryResults",
			conv: config.DocstringConventionNumpy,
			write: func(w *writer.CodeWriter) {
				query := &model.Query{Cmd: metadata.CmdMany, QueryName: "ListAuthors", SQL: "SELECT id FROM authors"}
				w.WriteQueryFunctionDocstring(1, query, "asyncpg.Connection", nil, "models.Author")
			},
			want: lines(
				"    \"\"\"Fetch many from the db using the SQL query with `name: ListAuthors :many`.",
				``,
				"    ```sql",
				`    SELECT id FROM authors`,
				"    ```",
				``,
				`    Parameters`,
				`    ----------`,
				`    conn : asyncpg.Connection`,
				"        Connection object of type `asyncpg.Connection` used to execute the query.",
				``,
				`    Returns`,
				`    -------`,
				`    QueryResults[models.Author]`,
				`        Helper class that allows both iteration and normal fetching of data from the db.`,
				``,
				`    """`,
			),
		},
	})
}
