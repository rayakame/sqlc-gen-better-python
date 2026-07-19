package writer

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
)

type QueryResultsWriter struct {
	writer *CodeWriter
}

func NewQueryResultsWriter(writer *CodeWriter) *QueryResultsWriter {
	return utils.ToPtr(QueryResultsWriter{writer: writer})
}

func (w *QueryResultsWriter) WriteQueryResultsClassHeader(
	connType string,
	initFields []string,
	driverReturnType string,
	async bool,
) {
	// PEP 695 class-scoped type parameter: no module-level TypeVar and no
	// typing.Generic base needed on Python 3.12+.
	w.writer.WriteLine("class QueryResults[T]:")
	w.writer.WriteQueryResultsClassDocstring(connType, driverReturnType)
	w.writer.WriteIndentedLine(1, `__slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_iterator", "_sql")`)
	w.writer.NewLine()
	w.writer.WriteIndentedLine(1, "def __init__(")
	w.writer.WriteIndentedLine(2, "self,")
	w.writer.WriteIndentedLine(2, fmt.Sprintf("conn: %s,", connType))
	w.writer.WriteIndentedLine(2, "sql: str,")
	w.writer.WriteIndentedLine(2, fmt.Sprintf("decode_hook: collections.abc.Callable[[%s], T],", driverReturnType))
	w.writer.WriteIndentedLine(2, "*args: QueryResultsArgsType,")
	w.writer.WriteIndentedLine(1, ") -> None:")
	w.writer.WriteQueryResultsInitDocstring(connType, driverReturnType)
	w.writer.WriteIndentedLine(2, "self._conn = conn")
	w.writer.WriteIndentedLine(2, "self._sql = sql")
	w.writer.WriteIndentedLine(2, "self._decode_hook = decode_hook")
	w.writer.WriteIndentedLine(2, "self._args = args")
	for _, line := range initFields {
		w.writer.WriteIndentedLine(2, line)
	}
	w.writer.NewLine()

	if async {
		w.writer.WriteIndentedLine(1, "def __aiter__(self) -> QueryResults[T]:")
	} else {
		w.writer.WriteIndentedLine(1, "def __iter__(self) -> QueryResults[T]:")
	}
	w.writer.WriteQueryResultsIterDocstring(async)
	w.writer.WriteIndentedLine(2, "return self")
	w.writer.NewLine()
}

// WriteQueryResultsCallFunction writes the synchronous __call__ method.
func (w *QueryResultsWriter) WriteQueryResultsCallFunction(wrapperLines []string) {
	w.writer.WriteIndentedLine(1, "def __call__(")
	w.writer.WriteIndentedLine(2, "self,")
	w.writer.WriteIndentedLine(1, ") -> collections.abc.Sequence[T]:")
	w.writer.WriteQueryResultsFetchDocstring(false)
	for _, line := range wrapperLines {
		w.writer.WriteIndentedLine(2, line)
	}
}

// WriteQueryResultsAwaitFunction writes the async __await__ method.
func (w *QueryResultsWriter) WriteQueryResultsAwaitFunction(wrapperLines []string) {
	w.writer.WriteIndentedLine(1, "def __await__(")
	w.writer.WriteIndentedLine(2, "self,")
	w.writer.WriteIndentedLine(1, ") -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:")
	w.writer.WriteQueryResultsFetchDocstring(true)
	if w.writer.DocstringsEnabled() {
		w.writer.NewLine()
	}
	w.writer.WriteIndentedLine(2, "async def _wrapper() -> collections.abc.Sequence[T]:")
	for _, line := range wrapperLines {
		w.writer.WriteIndentedLine(3, line)
	}
	w.writer.NewLine()
	w.writer.WriteIndentedLine(2, "return _wrapper().__await__()")
}
