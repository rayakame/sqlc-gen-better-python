package driver

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

const (
	asyncpgConnType   = "ConnectionLike"
	asyncpgResultType = "asyncpg.Record"
)

// AsyncpgDriver generates Python code for the asyncpg (async PostgreSQL) driver.
type AsyncpgDriver struct {
	rows *RowBuilder
}

func newAsyncpgDriver() *AsyncpgDriver {
	return &AsyncpgDriver{
		rows: newRowBuilder(asyncpgNeedsConversion),
	}
}

// Name returns "asyncpg".
func (d *AsyncpgDriver) Name() string { return "asyncpg" }

// ConnType returns "ConnectionLike".
func (d *AsyncpgDriver) ConnType() string { return asyncpgConnType }

// IsAsync returns true.
func (d *AsyncpgDriver) IsAsync() bool { return true }

// NeedsConversion reports whether a SQL type needs runtime conversion for asyncpg.
func (d *AsyncpgDriver) NeedsConversion(sqlType string) bool {
	return asyncpgNeedsConversion(sqlType)
}

// ConvertsInline reports whether a SQL type is converted inline; asyncpg converts
// everything inline (no registration mechanism).
func (d *AsyncpgDriver) ConvertsInline(sqlType string) bool {
	return asyncpgNeedsConversion(sqlType)
}

// WriteConversionSetup is a no-op for asyncpg.
func (d *AsyncpgDriver) WriteConversionSetup(_ *writer.CodeWriter, _ *config.Config, _ []model.Query) bool {
	return false
}

// TypeCheckingHook returns the ConnectionLike type alias for the TYPE_CHECKING block.
func (d *AsyncpgDriver) TypeCheckingHook() []string {
	return []string{
		fmt.Sprintf(
			"ConnectionLike: typing.TypeAlias = asyncpg.Connection[%[1]s] | asyncpg.pool.PoolConnectionProxy[%[1]s]",
			asyncpgResultType,
		),
	}
}

// WriteQueryResultsClass writes the async QueryResults class for asyncpg :many queries.
func (d *AsyncpgDriver) WriteQueryResultsClass(body *writer.CodeWriter) string {
	body.QueryResults.WriteQueryResultsClassHeader(asyncpgConnType, []string{
		fmt.Sprintf("self._cursor: asyncpg.cursor.CursorFactory[%s] | None = None", asyncpgResultType),
		fmt.Sprintf("self._iterator: asyncpg.cursor.CursorIterator[%s] | None = None", asyncpgResultType),
	}, asyncpgResultType, d.IsAsync())
	body.QueryResults.WriteQueryResultsAwaitFunction([]string{
		"result = await self._conn.fetch(self._sql, *self._args)",
		"return [self._decode_hook(row) for row in result]",
	})
	body.NewLine()
	body.WriteIndentedLine(1, "async def __anext__(self) -> T:")
	body.WriteQueryResultsNextDocstring("an asyncpg cursor", d.IsAsync())
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

// SupportsCommand returns if the driver supports the command.
func (d *AsyncpgDriver) SupportsCommand(cmd string) bool {
	switch cmd {
	case metadata.CmdExec,
		metadata.CmdExecResult,
		metadata.CmdExecRows,
		metadata.CmdOne,
		metadata.CmdMany,
		metadata.CmdCopyFrom:
		return true
	default:
		return false
	}
}

func (d *AsyncpgDriver) WriteQueryFunc(body *writer.CodeWriter, config *config.Config, query model.Query, indent int) {
	var annotation, docRetType string
	switch query.Cmd {
	case metadata.CmdExec:
		annotation, docRetType = query.Returns.Type.Print(), ""
	case metadata.CmdExecResult:
		annotation, docRetType = "str", "str"
	case metadata.CmdExecRows, metadata.CmdCopyFrom:
		annotation, docRetType = query.Returns.Type.Print(), query.Returns.Type.Type
	case metadata.CmdOne:
		annotation, docRetType = query.Returns.Type.PrintOptional(), query.Returns.Type.Type
	case metadata.CmdMany:
		annotation, docRetType = "QueryResults["+query.Returns.Type.Type+"]", query.Returns.Type.Type
	}

	conn := writeFuncSignature(body, d, config, indent, query, annotation)

	indent++
	writeQueryDocstring(body, d, config, query, indent, docRetType)
	callArgs := append([]string{query.ConstantName}, expandParams(query)...)
	switch query.Cmd {
	case metadata.CmdExec:
		body.WriteWrappedCall(indent, fmt.Sprintf("await %s.execute(", conn), callArgs, ")")

	case metadata.CmdExecResult:
		body.WriteWrappedCall(indent, fmt.Sprintf("return await %s.execute(", conn), callArgs, ")")

	case metadata.CmdExecRows:
		body.WriteWrappedCall(indent, fmt.Sprintf("r = await %s.execute(", conn), callArgs, ")")
		writeExecRowsReturn(body, config, indent)

	case metadata.CmdCopyFrom:
		writeCopyFromBody(body, config, query, conn, indent)

	case metadata.CmdOne:
		body.WriteWrappedCall(indent, fmt.Sprintf("row = await %s.fetchrow(", conn), callArgs, ")")
		body.WriteIndentedLine(indent, "if row is None:")
		body.WriteIndentedLine(indent+1, "return None")

		if query.Returns.IsStruct() {
			d.rows.WriteStructReturn(body, indent, query.Returns)
		} else {
			d.rows.WriteScalarReturn(body, indent, query.Returns)
		}

	case metadata.CmdMany:
		decodeHook := d.rows.WriteDecodeHook(body, indent, query, asyncpgResultType)
		manyArgs := append([]string{conn, query.ConstantName, decodeHook}, expandParams(query)...)
		body.WriteWrappedCall(indent, fmt.Sprintf("return QueryResults[%s](", query.Returns.Type.Type), manyArgs, ")")
	}
}

// writeCopyFromBody writes the body for an asyncpg :copyfrom command.
func writeCopyFromBody(body *writer.CodeWriter, config *config.Config, query model.Query, conn string, indent int) {
	var paramParts []string
	var columnParts []string
	for _, col := range query.Params[0].Table.Columns {
		paramParts = append(paramParts, fmt.Sprintf("param.%s", col.Name))
		columnParts = append(columnParts, fmt.Sprintf(`"%s"`, col.DBName))
	}

	paramsName := query.Params[0].Name
	singleComprehension := fmt.Sprintf("records = [(%s) for param in %s]", strings.Join(paramParts, ", "), paramsName)
	switch {
	case body.FitsLine(indent, singleComprehension):
		body.WriteIndentedLine(indent, singleComprehension)
	case body.FitsLine(indent+1, "("+strings.Join(paramParts, ", ")+")"):
		body.WriteIndentedLine(indent, "records = [")
		body.WriteIndentedLine(indent+1, fmt.Sprintf("(%s)", strings.Join(paramParts, ", ")))
		body.WriteIndentedLine(indent+1, "for param in "+paramsName)
		body.WriteIndentedLine(indent, "]")
	default:
		body.WriteIndentedLine(indent, "records = [")
		body.WriteIndentedLine(indent+1, "(")
		for _, part := range paramParts {
			body.WriteIndentedLine(indent+2, part+",")
		}
		body.WriteIndentedLine(indent+1, ")")
		body.WriteIndentedLine(indent+1, "for param in "+paramsName)
		body.WriteIndentedLine(indent, "]")
	}
	columnsArg := fmt.Sprintf("columns=[%s]", strings.Join(columnParts, ", "))
	copyArgs := []string{fmt.Sprintf(`"%s"`, query.Table.Name), columnsArg, "records=records"}
	if query.Table.Schema != "" {
		copyArgs = append(copyArgs, fmt.Sprintf(`schema_name="%s"`, query.Table.Schema))
	}

	head := fmt.Sprintf("r = await %s.copy_records_to_table(", conn)
	single := head + strings.Join(copyArgs, ", ") + ")"
	switch {
	case body.FitsLine(indent, single):
		body.WriteIndentedLine(indent, single)
	default:
		body.WriteIndentedLine(indent, head)
		for _, arg := range copyArgs {
			if arg == columnsArg && !body.FitsLine(indent+1, arg+",") {
				body.WriteIndentedLine(indent+1, "columns=[")
				for _, col := range columnParts {
					body.WriteIndentedLine(indent+2, col+",")
				}
				body.WriteIndentedLine(indent+1, "],")

				continue
			}
			body.WriteIndentedLine(indent+1, arg+",")
		}
		body.WriteIndentedLine(indent, ")")
	}
	writeExecRowsReturn(body, config, indent)
}
