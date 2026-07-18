package driver

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

const aiosqliteConnType = "aiosqlite.Connection"

// AiosqliteDriver generates Python code for the asynchronous aiosqlite driver.
type AiosqliteDriver struct {
	sqliteBase
}

func newAiosqliteDriver() *AiosqliteDriver {
	return &AiosqliteDriver{
		sqliteBase: newSqliteBase("aiosqlite"),
	}
}

// Name returns "aiosqlite".
func (d *AiosqliteDriver) Name() string { return "aiosqlite" }

// ConnType returns "aiosqlite.Connection".
func (d *AiosqliteDriver) ConnType() string { return aiosqliteConnType }

// IsAsync returns true.
func (d *AiosqliteDriver) IsAsync() bool { return true }

// WriteQueryResultsClass writes the async QueryResults class for aiosqlite :many queries.
func (d *AiosqliteDriver) WriteQueryResultsClass(body *writer.CodeWriter) string {
	body.QueryResults.WriteQueryResultsClassHeader(aiosqliteConnType, []string{
		"self._cursor: aiosqlite.Cursor | None = None",
		fmt.Sprintf("self._iterator: collections.abc.AsyncIterator[%s] | None = None", sqliteResultType),
	}, sqliteResultType, d.IsAsync())
	body.QueryResults.WriteQueryResultsAwaitFunction([]string{
		"result = await (await self._conn.execute(self._sql, self._args)).fetchall()",
		"return [self._decode_hook(row) for row in result]",
	})
	body.NewLine()
	body.WriteIndentedLine(1, "async def __anext__(self) -> T:")
	body.WriteQueryResultsNextDocstring("an aiosqlite cursor", d.IsAsync())
	body.WriteIndentedLine(2, "if self._cursor is None or self._iterator is None:")
	body.WriteIndentedLine(3, "self._cursor: aiosqlite.Cursor | None = await self._conn.execute(self._sql, self._args)")
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

func (d *AiosqliteDriver) WriteQueryFunc(body *writer.CodeWriter, config *config.Config, query model.Query, indent int) {
	conn := writeFuncSignature(body, d, config, indent, query)

	indent++
	switch query.Cmd {
	case metadata.CmdExec:
		body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.Print()))
		writeQueryDocstring(body, d, config, query, indent, "")
		body.WriteIndentedString(indent, fmt.Sprintf("await %s.execute(%s", conn, query.ConstantName))
		writeSqliteParams(body, query)
		body.WriteLine(")")

	case metadata.CmdExecResult:
		body.WriteLine(") -> aiosqlite.Cursor:")
		writeQueryDocstring(body, d, config, query, indent, "aiosqlite.Cursor")
		body.WriteIndentedString(indent, fmt.Sprintf("return await %s.execute(%s", conn, query.ConstantName))
		writeSqliteParams(body, query)
		body.WriteLine(")")

	case metadata.CmdExecRows:
		body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.Print()))
		writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)
		body.WriteIndentedString(indent, fmt.Sprintf("return (await %s.execute(%s", conn, query.ConstantName))
		writeSqliteParams(body, query)
		body.WriteLine(")).rowcount")

	case metadata.CmdExecLastId:
		body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.Print()))
		writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)
		body.WriteIndentedString(indent, fmt.Sprintf("return (await %s.execute(%s", conn, query.ConstantName))
		writeSqliteParams(body, query)
		body.WriteLine(")).lastrowid")

	case metadata.CmdOne:
		d.writeOneBody(body, config, query, conn, indent)

	case metadata.CmdMany:
		d.writeManyBody(body, config, query, conn, indent)
	}
}

// writeOneBody writes the body for a :one query.
func (d *AiosqliteDriver) writeOneBody(body *writer.CodeWriter, config *config.Config, query model.Query, conn string, indent int) {
	body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.PrintOptional()))
	writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)
	body.WriteIndentedString(indent, fmt.Sprintf("row = await (await %s.execute(%s", conn, query.ConstantName))
	writeSqliteParams(body, query)
	body.WriteLine(")).fetchone()")
	body.WriteIndentedLine(indent, "if row is None:")
	body.WriteIndentedLine(indent+1, "return None")

	if query.Returns.IsStruct() {
		d.rows.WriteStructReturn(body, indent, query.Returns)
	} else {
		d.rows.WriteScalarReturn(body, indent, query.Returns)
	}
}

// writeManyBody writes the body for a :many query.
func (d *AiosqliteDriver) writeManyBody(body *writer.CodeWriter, config *config.Config, query model.Query, conn string, indent int) {
	body.WriteLine(fmt.Sprintf(") -> QueryResults[%s]:", query.Returns.Type.Type))
	writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)

	decodeHook := d.rows.WriteDecodeHook(body, indent, query, sqliteResultType)

	body.WriteIndentedString(indent, fmt.Sprintf("return QueryResults[%s](%s, %s, %s", query.Returns.Type.Type, conn, query.ConstantName, decodeHook))
	writeSqliteManyParams(body, query)
	body.WriteLine(")")
}
