package writer_test

import (
	"strings"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
)

func TestNewQueryResultsWriter(t *testing.T) {
	t.Parallel()
	w := newWriter(config.DocstringConventionNone)
	qr := writer.NewQueryResultsWriter(w)
	if qr == nil {
		t.Fatal("NewQueryResultsWriter() = nil, want non-nil")
	}
	// The returned writer must emit into the CodeWriter it wraps.
	qr.WriteQueryResultsCallFunction(nil)
	want := strings.Join([]string{
		"    def __call__(",
		"        self,",
		"    ) -> collections.abc.Sequence[T]:",
		"",
	}, "\n")
	if got := w.String(); got != want {
		t.Errorf("WriteQueryResultsCallFunction(nil) = %q, want %q", got, want)
	}
}

func TestWriteQueryResultsClassHeader(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		connType   string
		initFields []string
		returnType string
		async      bool
		want       string
	}{
		{
			name:       "sync without init fields",
			connType:   "sqlite3.Connection",
			returnType: "sqlite3.Row",
			async:      false,
			want: strings.Join([]string{
				"class QueryResults[T]:",
				"    __slots__ = (\"_args\", \"_conn\", \"_cursor\", \"_decode_hook\", \"_iterator\", \"_sql\")",
				"",
				"    def __init__(",
				"        self,",
				"        conn: sqlite3.Connection,",
				"        sql: str,",
				"        decode_hook: collections.abc.Callable[[sqlite3.Row], T],",
				"        *args: QueryResultsArgsType,",
				"    ) -> None:",
				"        self._conn = conn",
				"        self._sql = sql",
				"        self._decode_hook = decode_hook",
				"        self._args = args",
				"",
				"    def __iter__(self) -> QueryResults[T]:",
				"        return self",
				"",
				"",
			}, "\n"),
		},
		{
			name:       "async with init fields",
			connType:   "asyncpg.Connection",
			initFields: []string{"self._cursor = None", "self._iterator = None"},
			returnType: "asyncpg.Record",
			async:      true,
			want: strings.Join([]string{
				"class QueryResults[T]:",
				"    __slots__ = (\"_args\", \"_conn\", \"_cursor\", \"_decode_hook\", \"_iterator\", \"_sql\")",
				"",
				"    def __init__(",
				"        self,",
				"        conn: asyncpg.Connection,",
				"        sql: str,",
				"        decode_hook: collections.abc.Callable[[asyncpg.Record], T],",
				"        *args: QueryResultsArgsType,",
				"    ) -> None:",
				"        self._conn = conn",
				"        self._sql = sql",
				"        self._decode_hook = decode_hook",
				"        self._args = args",
				"        self._cursor = None",
				"        self._iterator = None",
				"",
				"    def __aiter__(self) -> QueryResults[T]:",
				"        return self",
				"",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := newWriter(config.DocstringConventionNone)
			w.QueryResults.WriteQueryResultsClassHeader(tc.connType, tc.initFields, tc.returnType, tc.async)
			if got := w.String(); got != tc.want {
				t.Errorf("WriteQueryResultsClassHeader() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestWriteQueryResultsClassHeaderDocstrings(t *testing.T) {
	t.Parallel()
	w := newWriter(config.DocstringConventionGoogle)
	w.QueryResults.WriteQueryResultsClassHeader("asyncpg.Connection", []string{"self._cursor = None"}, "asyncpg.Record", true)
	want := strings.Join([]string{
		"class QueryResults[T]:",
		"    \"\"\"Helper class that allows both iteration and normal fetching of data from the db.\"\"\"",
		"",
		"    __slots__ = (\"_args\", \"_conn\", \"_cursor\", \"_decode_hook\", \"_iterator\", \"_sql\")",
		"",
		"    def __init__(",
		"        self,",
		"        conn: asyncpg.Connection,",
		"        sql: str,",
		"        decode_hook: collections.abc.Callable[[asyncpg.Record], T],",
		"        *args: QueryResultsArgsType,",
		"    ) -> None:",
		"        \"\"\"Initialize the QueryResults instance.",
		"",
		"        Args:",
		"            conn:",
		"                The connection object of type `asyncpg.Connection` used to execute queries.",
		"            sql:",
		"                The SQL statement that will be executed when fetching/iterating.",
		"            decode_hook:",
		"                A callback that turns an `asyncpg.Record` object into `T` that will be returned.",
		"            *args:",
		"                Arguments that should be sent when executing the sql query.",
		"        \"\"\"",
		"        self._conn = conn",
		"        self._sql = sql",
		"        self._decode_hook = decode_hook",
		"        self._args = args",
		"        self._cursor = None",
		"",
		"    def __aiter__(self) -> QueryResults[T]:",
		"        \"\"\"Initialize iteration support for `async for`.",
		"",
		"        Returns:",
		"            Self as an asynchronous iterator.",
		"        \"\"\"",
		"        return self",
		"",
		"",
	}, "\n")
	if got := w.String(); got != want {
		t.Errorf("WriteQueryResultsClassHeader() = %q, want %q", got, want)
	}
}

func TestWriteQueryResultsCallFunction(t *testing.T) {
	t.Parallel()
	w := newWriter(config.DocstringConventionNone)
	w.QueryResults.WriteQueryResultsCallFunction([]string{"records = self._conn.execute()", "return records"})
	want := strings.Join([]string{
		"    def __call__(",
		"        self,",
		"    ) -> collections.abc.Sequence[T]:",
		"        records = self._conn.execute()",
		"        return records",
		"",
	}, "\n")
	if got := w.String(); got != want {
		t.Errorf("WriteQueryResultsCallFunction() = %q, want %q", got, want)
	}
}

func TestWriteQueryResultsAwaitFunction(t *testing.T) {
	t.Parallel()
	t.Run("docstrings disabled", func(t *testing.T) {
		t.Parallel()
		w := newWriter(config.DocstringConventionNone)
		w.QueryResults.WriteQueryResultsAwaitFunction([]string{"return await _fetch()"})
		want := strings.Join([]string{
			"    def __await__(",
			"        self,",
			"    ) -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:",
			"        async def _wrapper() -> collections.abc.Sequence[T]:",
			"            return await _fetch()",
			"",
			"        return _wrapper().__await__()",
			"",
		}, "\n")
		if got := w.String(); got != want {
			t.Errorf("WriteQueryResultsAwaitFunction() = %q, want %q", got, want)
		}
	})
	t.Run("docstrings enabled add blank line", func(t *testing.T) {
		t.Parallel()
		w := newWriter(config.DocstringConventionGoogle)
		w.QueryResults.WriteQueryResultsAwaitFunction([]string{"return await _fetch()"})
		got := w.String()
		wantPrefix := strings.Join([]string{
			"    def __await__(",
			"        self,",
			"    ) -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:",
			"",
		}, "\n")
		if !strings.HasPrefix(got, wantPrefix) {
			t.Errorf("WriteQueryResultsAwaitFunction() = %q, want prefix %q", got, wantPrefix)
		}
		// The DocstringsEnabled branch inserts a blank line between the
		// docstring and the nested wrapper definition.
		wantGap := "\"\"\"\n\n        async def _wrapper() -> collections.abc.Sequence[T]:\n"
		if !strings.Contains(got, wantGap) {
			t.Errorf("WriteQueryResultsAwaitFunction() output missing %q\ngot: %q", wantGap, got)
		}
	})
}
