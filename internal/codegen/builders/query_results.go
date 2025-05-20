package builders

import "fmt"

func (b *IndentStringBuilder) WriteQueryResultsClassHeader(connType string, initFields []string, driverReturnType string) {
	b.WriteLine(`T = typing.TypeVar("T")`)
	b.NewLine()
	b.WriteLine("class QueryResults(typing.Generic[T]):")
	b.WriteQueryResultsClassDocstring(connType, driverReturnType)
	b.WriteIndentedLine(1, `__slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_iterator", "_sql")`)
	b.NewLine()
	b.WriteIndentedLine(1, "def __init__(")
	b.WriteIndentedLine(2, "self,")
	b.WriteIndentedLine(2, fmt.Sprintf("conn: %s,", connType))
	b.WriteIndentedLine(2, "sql: str,")
	b.WriteIndentedLine(2, fmt.Sprintf("decode_hook: collections.abc.Callable[[%s], T],", driverReturnType))
	b.WriteIndentedLine(2, "*args: QueryResultsArgsType,")
	b.WriteIndentedLine(1, ") -> None:")
	b.WriteQueryResultsInitDocstring(connType, driverReturnType)
	b.WriteIndentedLine(2, "self._conn = conn")
	b.WriteIndentedLine(2, "self._sql = sql")
	b.WriteIndentedLine(2, "self._decode_hook = decode_hook")
	b.WriteIndentedLine(2, "self._args = args")
	for _, line := range initFields {
		b.WriteIndentedLine(2, line)
	}
	b.NewLine()
	b.WriteIndentedLine(1, "def __aiter__(self) -> QueryResults[T]:")
	b.WriteQueryResultsAiterDocstring()
	b.WriteIndentedLine(2, "return self")
	b.NewLine()
}

func (b *IndentStringBuilder) WriteQueryResultsAwaitFunction(wrapperLines []string) {
	b.WriteIndentedLine(1, "def __await__(")
	b.WriteIndentedLine(2, "self,")
	b.WriteIndentedLine(1, ") -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:")
	b.WriteQueryResultsAwaitDocstring()
	b.WriteIndentedLine(2, "async def _wrapper() -> collections.abc.Sequence[T]:")
	for _, line := range wrapperLines {
		b.WriteIndentedLine(3, line)
	}
	b.WriteIndentedLine(2, "return _wrapper().__await__()")

}
