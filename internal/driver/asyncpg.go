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
	conn := writeFuncSignature(body, d, config, indent, query)

	indent++
	switch query.Cmd {
	case metadata.CmdExec:
		body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.Print()))
		writeQueryDocstring(body, d, config, query, indent, "")
		body.WriteIndentedString(indent, fmt.Sprintf("await %s.execute(%s", conn, query.ConstantName))
		writeAsyncpgParams(body, query)
		body.WriteLine(")")

	case metadata.CmdExecResult:
		body.WriteLine(") -> str:")
		writeQueryDocstring(body, d, config, query, indent, "str")
		body.WriteIndentedString(indent, fmt.Sprintf("return await %s.execute(%s", conn, query.ConstantName))
		writeAsyncpgParams(body, query)
		body.WriteLine(")")

	case metadata.CmdExecRows:
		body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.Print()))
		writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)
		body.WriteIndentedString(indent, fmt.Sprintf("r = await %s.execute(%s", conn, query.ConstantName))
		writeAsyncpgParams(body, query)
		body.WriteLine(")")
		writeExecRowsReturn(body, config, indent)

	case metadata.CmdCopyFrom:
		body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.Print()))
		writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)
		writeCopyFromBody(body, config, query, conn, indent)

	case metadata.CmdOne:
		d.writeOneBody(body, config, query, conn, indent)

	case metadata.CmdMany:
		d.writeManyBody(body, config, query, conn, indent)
	}
}

// writeOneBody writes the body for a :one query.
func (d *AsyncpgDriver) writeOneBody(body *writer.CodeWriter, config *config.Config, query model.Query, conn string, indent int) {
	body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.PrintOptional()))
	writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)
	body.WriteIndentedString(indent, fmt.Sprintf("row = await %s.fetchrow(%s", conn, query.ConstantName))
	writeAsyncpgParams(body, query)
	body.WriteLine(")")
	body.WriteIndentedLine(indent, "if row is None:")
	body.WriteIndentedLine(indent+1, "return None")

	if query.Returns.IsStruct() {
		d.rows.WriteStructReturn(body, indent, query.Returns)
	} else {
		d.rows.WriteScalarReturn(body, indent, query.Returns)
	}
}

// writeManyBody writes the body for a :many query.
func (d *AsyncpgDriver) writeManyBody(body *writer.CodeWriter, config *config.Config, query model.Query, conn string, indent int) {
	body.WriteLine(fmt.Sprintf(") -> QueryResults[%s]:", query.Returns.Type.Type))
	writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)

	decodeHook := d.rows.WriteDecodeHook(body, indent, query, asyncpgResultType)

	body.WriteIndentedString(indent, fmt.Sprintf("return QueryResults[%s](%s, %s, %s", query.Returns.Type.Type, conn, query.ConstantName, decodeHook))
	writeAsyncpgParams(body, query)
	body.WriteLine(")")
}

// writeAsyncpgParams writes asyncpg-style parameters: ", arg1, arg2".
func writeAsyncpgParams(w *writer.CodeWriter, query model.Query) {
	if len(query.Params) == 0 {
		return
	}
	parts := expandParams(query)
	if len(parts) > 0 {
		w.WriteString(", " + strings.Join(parts, ", "))
	}
}

// writeCopyFromBody writes the body for an asyncpg :copyfrom command.
func writeCopyFromBody(body *writer.CodeWriter, config *config.Config, query model.Query, conn string, indent int) {
	body.WriteIndentedLine(indent, "records = [")

	var paramParts []string
	var columnParts []string
	for _, col := range query.Params[0].Table.Columns {
		paramParts = append(paramParts, fmt.Sprintf("param.%s", col.Name))
		columnParts = append(columnParts, fmt.Sprintf(`"%s"`, col.DBName))
	}

	body.WriteIndentedLine(indent+1, fmt.Sprintf("(%s)", strings.Join(paramParts, ", ")))
	body.WriteIndentedLine(indent+1, fmt.Sprintf("for param in %s", query.Params[0].Name))
	body.WriteIndentedLine(indent, "]")
	body.WriteIndentedString(indent, fmt.Sprintf(
		`r = await %s.copy_records_to_table("%s", columns=[%s], records=records`,
		conn, query.Table.Name, strings.Join(columnParts, ", "),
	))
	if query.Table.Schema != "" {
		body.WriteString(fmt.Sprintf(`, schema_name="%s"`, query.Table.Schema))
	}
	body.WriteLine(")")
	writeExecRowsReturn(body, config, indent)
}
