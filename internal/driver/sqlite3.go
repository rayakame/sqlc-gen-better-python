package driver

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

const sqlite3ConnType = "sqlite3.Connection"

// Sqlite3Driver generates Python code for the synchronous sqlite3 driver.
type Sqlite3Driver struct {
	sqliteBase
}

func newSqlite3Driver() *Sqlite3Driver {
	return &Sqlite3Driver{
		sqliteBase: newSqliteBase("sqlite3"),
	}
}

// Name returns "sqlite3".
func (d *Sqlite3Driver) Name() string { return "sqlite3" }

// ConnType returns "sqlite3.Connection".
func (d *Sqlite3Driver) ConnType() string { return sqlite3ConnType }

// IsAsync returns true.
func (d *Sqlite3Driver) IsAsync() bool { return false }

// WriteQueryResultsClass writes the synchronous QueryResults class.
func (d *Sqlite3Driver) WriteQueryResultsClass(body *writer.CodeWriter) string {
	body.QueryResults.WriteQueryResultsClassHeader(sqlite3ConnType, []string{
		"self._cursor: sqlite3.Cursor | None = None",
		fmt.Sprintf("self._iterator: collections.abc.Iterator[%s] | None = None", sqliteResultType),
	}, sqliteResultType, d.IsAsync())
	body.QueryResults.WriteQueryResultsCallFunction([]string{
		"result = self._conn.execute(self._sql, self._args).fetchall()",
		"return [self._decode_hook(row) for row in result]",
	})
	body.NewLine()
	body.WriteIndentedLine(1, "def __next__(self) -> T:")
	body.WriteQueryResultsNextDocstring("a sqlite3 cursor", d.IsAsync())
	body.WriteIndentedLine(2, "if self._cursor is None or self._iterator is None:")
	body.WriteIndentedLine(3, "self._cursor: sqlite3.Cursor | None = self._conn.execute(self._sql, self._args)")
	body.WriteIndentedLine(3, "self._iterator = self._cursor.__iter__()")
	body.WriteIndentedLine(2, "try:")
	body.WriteIndentedLine(3, "record = self._iterator.__next__()")
	body.WriteIndentedLine(2, "except StopIteration:")
	body.WriteIndentedLine(3, "self._cursor = None")
	body.WriteIndentedLine(3, "self._iterator = None")
	body.WriteIndentedLine(3, "raise")
	body.WriteIndentedLine(2, "return self._decode_hook(record)")
	return "QueryResults"
}

func (d *Sqlite3Driver) WriteQueryFunc(body *writer.CodeWriter, config *config.Config, query model.Query, indent int) {
	conn := writeFuncSignature(body, d, config, indent, query)

	indent++
	switch query.Cmd {
	case metadata.CmdExec:
		body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.Print()))
		writeQueryDocstring(body, d, config, query, indent, "")
		body.WriteIndentedString(indent, fmt.Sprintf("%s.execute(%s", conn, query.ConstantName))
		writeSqliteParams(body, query)
		body.WriteLine(")")

	case metadata.CmdExecResult:
		body.WriteLine(") -> sqlite3.Cursor:")
		writeQueryDocstring(body, d, config, query, indent, "sqlite3.Cursor")
		body.WriteIndentedString(indent, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName))
		writeSqliteParams(body, query)
		body.WriteLine(")")

	case metadata.CmdExecRows:
		body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.Print()))
		writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)
		body.WriteIndentedString(indent, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName))
		writeSqliteParams(body, query)
		body.WriteLine(").rowcount")

	case metadata.CmdExecLastId:
		body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.Print()))
		writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)
		body.WriteIndentedString(indent, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName))
		writeSqliteParams(body, query)
		body.WriteLine(").lastrowid")

	case metadata.CmdOne:
		d.writeOneBody(body, config, query, conn, indent)

	case metadata.CmdMany:
		d.writeManyBody(body, config, query, conn, indent)
	}
}

// writeOneBody writes the body for a :one query.
func (d *Sqlite3Driver) writeOneBody(body *writer.CodeWriter, config *config.Config, query model.Query, conn string, indent int) {
	body.WriteLine(fmt.Sprintf(") -> %s:", query.Returns.Type.PrintOptional()))
	writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)
	body.WriteIndentedString(indent, fmt.Sprintf("row = %s.execute(%s", conn, query.ConstantName))
	writeSqliteParams(body, query)
	body.WriteLine(").fetchone()")
	body.WriteIndentedLine(indent, "if row is None:")
	body.WriteIndentedLine(indent+1, "return None")

	if query.Returns.IsStruct() {
		d.rows.WriteStructReturn(body, indent, query.Returns)
	} else {
		d.rows.WriteScalarReturn(body, indent, query.Returns)
	}
}

// writeManyBody writes the body for a :many query.
func (d *Sqlite3Driver) writeManyBody(body *writer.CodeWriter, config *config.Config, query model.Query, conn string, indent int) {
	body.WriteLine(fmt.Sprintf(") -> QueryResults[%s]:", query.Returns.Type.Type))
	writeQueryDocstring(body, d, config, query, indent, query.Returns.Type.Type)

	decodeHook := d.rows.WriteDecodeHook(body, indent, query, sqliteResultType)

	body.WriteIndentedString(indent, fmt.Sprintf("return QueryResults[%s](%s, %s, %s", query.Returns.Type.Type, conn, query.ConstantName, decodeHook))
	writeSqliteManyParams(body, query)
	body.WriteLine(")")
}
