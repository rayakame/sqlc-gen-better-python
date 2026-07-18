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
	var annotation, docRetType string
	switch query.Cmd {
	case metadata.CmdExec:
		annotation, docRetType = query.Returns.Type.Print(), ""
	case metadata.CmdExecResult:
		annotation, docRetType = "aiosqlite.Cursor", "aiosqlite.Cursor"
	case metadata.CmdExecRows, metadata.CmdExecLastId:
		annotation, docRetType = query.Returns.Type.Print(), query.Returns.Type.Type
	case metadata.CmdOne:
		annotation, docRetType = query.Returns.Type.PrintOptional(), query.Returns.Type.Type
	case metadata.CmdMany:
		annotation, docRetType = "QueryResults["+query.Returns.Type.Type+"]", query.Returns.Type.Type
	}

	conn := writeFuncSignature(body, d, config, indent, query, annotation)

	indent++
	writeQueryDocstring(body, d, config, query, indent, docRetType)
	switch query.Cmd {
	case metadata.CmdExec:
		writeSqliteCall(body, indent, query, fmt.Sprintf("await %s.execute(%s", conn, query.ConstantName), ")")

	case metadata.CmdExecResult:
		writeSqliteCall(body, indent, query, fmt.Sprintf("return await %s.execute(%s", conn, query.ConstantName), ")")

	case metadata.CmdExecRows:
		writeSqliteCall(body, indent, query, fmt.Sprintf("return (await %s.execute(%s", conn, query.ConstantName), ")).rowcount")

	case metadata.CmdExecLastId:
		writeSqliteCall(body, indent, query, fmt.Sprintf("return (await %s.execute(%s", conn, query.ConstantName), ")).lastrowid")

	case metadata.CmdOne:
		writeSqliteCall(body, indent, query, fmt.Sprintf("row = await (await %s.execute(%s", conn, query.ConstantName), ")).fetchone()")
		body.WriteIndentedLine(indent, "if row is None:")
		body.WriteIndentedLine(indent+1, "return None")

		if query.Returns.IsStruct() {
			d.rows.WriteStructReturn(body, indent, query.Returns)
		} else {
			d.rows.WriteScalarReturn(body, indent, query.Returns)
		}

	case metadata.CmdMany:
		decodeHook := d.rows.WriteDecodeHook(body, indent, query, sqliteResultType)
		manyArgs := append([]string{conn, query.ConstantName, decodeHook}, expandParams(query)...)
		body.WriteWrappedCall(indent, fmt.Sprintf("return QueryResults[%s](", query.Returns.Type.Type), manyArgs, ")")
	}
}
