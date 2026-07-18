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
	var annotation, docRetType string
	switch query.Cmd {
	case metadata.CmdExec:
		annotation, docRetType = query.Returns.Type.Print(), ""
	case metadata.CmdExecResult:
		annotation, docRetType = "sqlite3.Cursor", "sqlite3.Cursor"
	case metadata.CmdExecRows, metadata.CmdExecLastId:
		annotation, docRetType = query.Returns.Type.Print(), query.Returns.Type.Type
	case metadata.CmdOne:
		annotation, docRetType = query.Returns.Type.PrintOptional(), query.Returns.Type.Type
	case metadata.CmdMany:
		annotation, docRetType = "QueryResults["+query.Returns.Type.Print()+"]", query.Returns.Type.Print()
	}

	conn := writeFuncSignature(body, d, config, indent, query, annotation)

	indent++
	writeQueryDocstring(body, d, config, query, indent, docRetType)
	switch query.Cmd {
	case metadata.CmdExec:
		writeSqliteCall(body, indent, query, fmt.Sprintf("%s.execute(%s", conn, query.ConstantName), ")")

	case metadata.CmdExecResult:
		writeSqliteCall(body, indent, query, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName), ")")

	case metadata.CmdExecRows:
		writeSqliteCall(body, indent, query, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName), ").rowcount")

	case metadata.CmdExecLastId:
		writeSqliteCall(body, indent, query, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName), ").lastrowid")

	case metadata.CmdOne:
		writeSqliteCall(body, indent, query, fmt.Sprintf("row = %s.execute(%s", conn, query.ConstantName), ").fetchone()")
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
		body.WriteWrappedCall(indent, fmt.Sprintf("return QueryResults[%s](", query.Returns.Type.Print()), manyArgs, ")")
	}
}
