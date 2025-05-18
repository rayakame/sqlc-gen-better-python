package drivers

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/rayakame/sqlc-gen-better-python/internal/typeConversion"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"strconv"
	"strings"
)

const AsyncpgConn = "ConnectionLike"
const AsyncpgResult = "asyncpg.Record"

func AsyncpgTypeCheckingHook() []string {
	return []string{
		fmt.Sprintf(
			"ConnectionLike: typing.TypeAlias = asyncpg.Connection[%[1]s] | asyncpg.pool.PoolConnectionProxy[%[1]s]",
			AsyncpgResult,
		),
	}
}

func AsyncpgBuildQueryResults(body *builders.IndentStringBuilder) string {
	body.WriteQueryResultsClassHeader(AsyncpgConn, []string{
		fmt.Sprintf("self._cursor: asyncpg.cursor.CursorFactory[%s] | None = None", AsyncpgResult),
		fmt.Sprintf("self._iterator: asyncpg.cursor.CursorIterator[%s] | None = None", AsyncpgResult),
	}, AsyncpgResult)
	body.WriteQueryResultsAwaitFunction([]string{
		"result = await self._conn.fetch(self._sql, *self._args)",
		"return [self._decode_hook(row) for row in result]",
	})
	body.NewLine()
	body.WriteIndentedLine(1, "async def __anext__(self) -> T:")
	body.WriteQueryResultsAnextDocstringAsyncpg()
	body.WriteIndentedLine(2, "if self._cursor is None or self._iterator is None:")
	body.WriteIndentedLine(3, "self._cursor = self._conn.cursor(self._sql, *self._args)")
	body.WriteIndentedLine(3, "self._iterator = self._cursor.__aiter__()")
	body.WriteIndentedLine(2, "try:")
	body.WriteIndentedLine(3, "record = await self._iterator.__anext__()")
	body.WriteIndentedLine(2, "except StopAsyncIteration:")
	body.WriteIndentedLine(3, "self._cursor = None")
	body.WriteIndentedLine(3, "self._iterator = None")
	body.WriteIndentedLine(3, "raise")
	body.WriteIndentedLine(2, "return self._decode_hook(record)")
	return "QueryResults"
}

func AsyncpgBuildPyQueryFunc(query *core.Query, body *builders.IndentStringBuilder, args []core.FunctionArg, retType core.PyType, isClass bool) error {
	indentLevel := 0
	params := fmt.Sprintf("conn: %s", AsyncpgConn)
	conn := "conn"
	asyncFunc := "async "
	docstringConnType := AsyncpgConn
	if isClass {
		params = "self"
		conn = "self._conn"
		indentLevel = 1
		docstringConnType = ""
	}
	if query.Cmd == metadata.CmdMany {
		asyncFunc = ""
	}
	body.WriteIndentedString(indentLevel, fmt.Sprintf("%sdef %s(%s", asyncFunc, query.FuncName, params))
	for i, arg := range args {
		if i == 0 {
			body.WriteString(", *")
		}
		body.WriteString(fmt.Sprintf(", %s", arg.FunctionFormat))
	}
	if query.Cmd == metadata.CmdExec {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, retType)
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("await %s.execute(%s", conn, query.ConstantName))
		asyncpgWriteParams(query, body)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecResult {
		body.WriteLine(") -> str:")
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, core.PyType{Type: "str"})
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return await %s.execute(%s", conn, query.ConstantName))
		asyncpgWriteParams(query, body)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecRows {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, retType)
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("result = await %s.execute(%s", conn, query.ConstantName))
		asyncpgWriteParams(query, body)
		body.WriteLine(")")
		body.WriteIndentedLine(indentLevel+1, "return int(result.split()[-1]) if result.split()[-1].isdigit() else 0")
	} else if query.Cmd == metadata.CmdOne {
		body.WriteLine(fmt.Sprintf(") -> %s | None:", retType.Type))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, retType)
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("row = await %s.fetchrow(%s", conn, query.ConstantName))
		asyncpgWriteParams(query, body)
		body.WriteLine(")")
		body.WriteIndentedLine(indentLevel+1, "if row is None:")
		body.WriteIndentedLine(indentLevel+2, "return None")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s(", retType.Type))
			i := 0
			for _, col := range query.Ret.Table.Columns {
				if i != 0 {
					body.WriteString(", ")
				}
				if len(col.EmbedFields) != 0 {
					var inner []string
					body.WriteString(fmt.Sprintf("%s=%s(", col.Name, col.Type.Type))
					for _, embedCol := range col.EmbedFields {
						if _, found := typeConversion.AsyncpgDoTypeConversion()[embedCol.Type.SqlType]; found {
							inner = append(inner, fmt.Sprintf("%s=%s(row[%s])", embedCol.Name, embedCol.Type.Type, strconv.Itoa(i)))
						} else {
							inner = append(inner, fmt.Sprintf("%s=row[%s]", embedCol.Name, strconv.Itoa(i)))
						}
						i++
					}
					body.WriteString(strings.Join(inner, ", ") + ")")
				} else {
					if _, found := typeConversion.AsyncpgDoTypeConversion()[col.Type.SqlType]; found {
						body.WriteString(fmt.Sprintf("%s=%s(row[%s])", col.Name, col.Type.Type, strconv.Itoa(i)))
					} else {
						body.WriteString(fmt.Sprintf("%s=row[%s]", col.Name, strconv.Itoa(i)))
					}
					i++
				}
			}
			body.WriteString("   ")
			body.WriteLine(")")
		} else {
			if _, found := typeConversion.AsyncpgDoTypeConversion()[retType.SqlType]; found {
				body.WriteIndentedLine(indentLevel+1, fmt.Sprintf("return %s(row[0])", retType.Type))
			} else {
				body.WriteIndentedLine(indentLevel+1, "return row[0]")
			}
		}
	} else if query.Cmd == metadata.CmdMany {
		body.WriteLine(fmt.Sprintf(") -> QueryResults[%s]:", retType.Type))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, retType)
		body.WriteIndentedLine(indentLevel+1, fmt.Sprintf("def _decode_hook(row: %s) -> %s:", AsyncpgResult, retType.Type))
		if query.Ret.IsStruct() {
			body.WriteIndentedString(indentLevel+2, fmt.Sprintf("return %s(", retType.Type))
			i := 0
			for _, col := range query.Ret.Table.Columns {
				if i != 0 {
					body.WriteString(", ")
				}
				if len(col.EmbedFields) != 0 {
					var inner []string
					body.WriteString(fmt.Sprintf("%s=%s(", col.Name, col.Type.Type))
					for _, embedCol := range col.EmbedFields {
						if _, found := typeConversion.AsyncpgDoTypeConversion()[embedCol.Type.SqlType]; found {
							inner = append(inner, fmt.Sprintf("%s=%s(row[%s])", embedCol.Name, embedCol.Type.Type, strconv.Itoa(i)))
						} else {
							inner = append(inner, fmt.Sprintf("%s=row[%s]", embedCol.Name, strconv.Itoa(i)))
						}
						i++
					}
					body.WriteString(strings.Join(inner, ", ") + ")")
				} else {
					if _, found := typeConversion.AsyncpgDoTypeConversion()[col.Type.SqlType]; found {
						body.WriteString(fmt.Sprintf("%s=%s(row[%s])", col.Name, col.Type.Type, strconv.Itoa(i)))
					} else {
						body.WriteString(fmt.Sprintf("%s=row[%s]", col.Name, strconv.Itoa(i)))
					}
					i++
				}
			}
			body.WriteLine(")")
		} else {
			if _, found := typeConversion.AsyncpgDoTypeConversion()[retType.SqlType]; found {
				body.WriteIndentedLine(indentLevel+2, fmt.Sprintf("return %s(row[0])", retType.Type))
			} else {
				body.WriteIndentedLine(indentLevel+2, "return row[0]")
			}
		}
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return QueryResults[%s](%s, %s, _decode_hook", retType.Type, conn, query.ConstantName))
		asyncpgWriteParams(query, body)
		body.WriteLine(")")
	}
	return nil
}

func AsyncpgAcceptedDriverCMDs() []string {
	return []string{
		metadata.CmdExec,
		metadata.CmdExecResult,
		metadata.CmdExecRows,
		metadata.CmdOne,
		metadata.CmdMany,
	}
}

func asyncpgWriteParams(query *core.Query, body *builders.IndentStringBuilder) {
	if len(query.Args) == 0 {
		return
	}
	params := ""
	for i, arg := range query.Args {
		if !arg.IsEmpty() {
			if i == len(query.Args)-1 {
				params += fmt.Sprintf(" %s", arg.Name)
			} else {
				params += fmt.Sprintf(" %s,", arg.Name)
			}
		}
	}
	body.WriteString("," + params)
}
