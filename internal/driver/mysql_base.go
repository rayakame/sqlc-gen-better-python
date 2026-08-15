package driver

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

// mysqlResultType is the row type of the default (tuple) cursors of both
// MySQL modules.
const mysqlResultType = "tuple[typing.Any, ...]"

// mysqlBinaryTypes are the SQL types carried as raw bytes on the wire.
var mysqlBinaryTypes = map[string]struct{}{
	sqlTypeBlob:  {},
	"binary":     {},
	"varbinary":  {},
	"tinyblob":   {},
	"mediumblob": {},
	"longblob":   {},
	"bit":        {},
}

// mysqlNeedsConversion reports the SQL types whose values convert inline.
// The PyMySQL family returns datetime/date/timedelta/Decimal natively, so
// only the binary family (bytes -> memoryview) and the tinyint spellings
// (int -> bool for tinyint(1) columns) remain. No decode templates are
// needed: RowBuilder's constructor-call fallback produces memoryview(...)
// and bool(...) - a plain tinyint column gets a redundant but harmless
// int(...) wrap, the price of keying conversions off the SQL type alone.
func mysqlNeedsConversion(sqlType string) bool {
	if _, found := mysqlBinaryTypes[sqlType]; found {
		return true
	}
	switch sqlType {
	case "tinyint", types.Bool, types.Boolean:
		return true
	default:
		return false
	}
}

// mysqlWire converts parameters the drivers cannot bind natively: the
// PyMySQL encoder table has no memoryview entry and silently stringifies
// one, so binary values bind as bytes.
func mysqlWire(sqlType string) (string, bool) {
	if _, found := mysqlBinaryTypes[sqlType]; found {
		return wireBytes, true
	}

	return "", false
}

// mysqlBase is the complete driver implementation for both MySQL modules -
// pymysql (sync) and asyncmy (async). All emission differences between the
// two are derived from moduleName and the async flag.
type mysqlBase struct {
	moduleName string // "pymysql" or "asyncmy"
	async      bool
	rows       *RowBuilder
}

var _ Driver = (*mysqlBase)(nil)

func newMysqlDriver(moduleName string, async bool) *mysqlBase {
	return &mysqlBase{
		moduleName: moduleName,
		async:      async,
		rows:       newRowBuilder(mysqlNeedsConversion),
	}
}

// Name returns the Python module name ("pymysql" or "asyncmy").
func (mb *mysqlBase) Name() string { return mb.moduleName }

// ConnType returns the connection type annotation, e.g. "pymysql.Connection".
func (mb *mysqlBase) ConnType() string { return mb.moduleName + ".Connection" }

// IsAsync reports whether this is the asyncmy (async) driver.
func (mb *mysqlBase) IsAsync() bool { return mb.async }

// SupportsCommand returns if the driver supports the command; the set
// matches the sqlite drivers (:copyfrom has no honest MySQL-driver
// equivalent - LOAD DATA LOCAL INFILE is not expressible through them).
func (mb *mysqlBase) SupportsCommand(cmd string) bool {
	switch cmd {
	case metadata.CmdExec,
		metadata.CmdExecResult,
		metadata.CmdExecLastId,
		metadata.CmdExecRows,
		metadata.CmdOne,
		metadata.CmdMany:
		return true
	default:
		return false
	}
}

// TypeCheckingHook returns nil (no type-checking hook for MySQL drivers).
func (mb *mysqlBase) TypeCheckingHook() []string {
	return nil
}

// NeedsConversion reports whether a SQL type needs runtime conversion.
func (mb *mysqlBase) NeedsConversion(sqlType string) bool {
	return mysqlNeedsConversion(sqlType)
}

// ConvertsInline reports whether a SQL type converts inline in decode code;
// MySQL conversions are all inline (the drivers' converter maps are
// per-connection, so generated modules cannot register anything).
func (mb *mysqlBase) ConvertsInline(sqlType string) bool {
	return mysqlNeedsConversion(sqlType)
}

// WriteConversionSetup is a no-op for MySQL: nothing runs at module import
// time.
func (mb *mysqlBase) WriteConversionSetup(_ *writer.CodeWriter, _ *config.Config, _ []model.Query) bool {
	return false
}

// WriteQueryResultsClass writes the QueryResults class for :many queries, in
// its sync (pymysql) or async (asyncmy) variant. Neither module's connection
// has an execute method, so both paths open a cursor first; iteration steps
// with fetchone like the turso drivers, keeping the state machine identical
// in both flavors.
func (mb *mysqlBase) WriteQueryResultsClass(body *writer.CodeWriter) string {
	awaitKw, nextDef, stopExc, article := "", defNextSync, stopIteration, "a "
	if mb.async {
		awaitKw, nextDef, stopExc, article = awaitPrefix, defNextAsync, stopAsyncIteration, "an "
	}

	body.QueryResults.WriteQueryResultsClassHeaderNoIterator(mb.ConnType(), []string{
		"self._cursor: " + mb.cursorType() + " | None = None",
	}, mysqlResultType, mb.async)
	fetchLines := []string{
		"cur = self._conn.cursor()",
		awaitKw + "cur.execute(self._sql, self._args)",
		fmt.Sprintf("result = %scur.fetchall()", awaitKw),
		awaitKw + "cur.close()",
		decodeRowsExpr,
	}
	if mb.async {
		body.QueryResults.WriteQueryResultsAwaitFunction(fetchLines)
	} else {
		body.QueryResults.WriteQueryResultsCallFunction(fetchLines)
	}
	body.NewLine()
	body.WriteIndentedLine(1, nextDef+"(self) -> T:")
	body.WriteQueryResultsNextDocstring(article+mb.moduleName+" cursor", mb.async)
	body.WriteIndentedLine(2, "if self._cursor is None:")
	body.WriteIndentedLine(3, "self._cursor = self._conn.cursor()")
	body.WriteIndentedLine(3, awaitKw+"self._cursor.execute(self._sql, self._args)")
	body.WriteIndentedLine(2, "record = "+awaitKw+"self._cursor.fetchone()")
	body.WriteIndentedLine(2, "if record is None:")
	body.WriteIndentedLine(3, "self._cursor = None")
	body.WriteIndentedLine(3, "raise "+stopExc)
	body.WriteIndentedLine(2, "return self._decode_hook(record)")

	return queryResultsClassName
}

func (mb *mysqlBase) WriteQueryFunc(body *writer.CodeWriter, config *config.Config, query model.Query, indent int) {
	var annotation, docRetType string
	switch query.Cmd {
	case metadata.CmdExec:
		annotation, docRetType = query.Returns.Type.Print(), ""
	case metadata.CmdExecResult:
		annotation, docRetType = mb.cursorType(), mb.cursorType()
	case metadata.CmdExecRows, metadata.CmdExecLastId:
		annotation, docRetType = query.Returns.Type.Print(), query.Returns.Type.Type
	case metadata.CmdOne:
		annotation, docRetType = query.Returns.Type.PrintOptional(), query.Returns.Type.Type
	case metadata.CmdMany:
		annotation, docRetType = "QueryResults["+query.Returns.Type.Print()+"]", query.Returns.Type.Print()
	}

	conn := writeFuncSignature(body, mb, config, indent, query, annotation)

	indent++
	writeQueryDocstring(body, mb, config, query, indent, docRetType)
	// :many delays this until after the decode hook, whose trailing blank
	// line keeps the assignment from touching the nested def (ruff E306).
	sqlRef := query.ConstantName
	if query.Cmd != metadata.CmdMany {
		sqlRef = writeSliceExpansion(body, indent, query, pyformatSliceJoin)
	}

	// Neither MySQL module has conn.execute: every body opens a cursor. The
	// with-block closes it; :execresult hands the open cursor to the caller
	// instead.
	withStmt, execKw := "with ", ""
	if mb.async {
		withStmt, execKw = "async with ", awaitPrefix
	}
	withStmt += conn + ".cursor() as cur:"
	parts := expandParamsFlattenSlicesWire(query, mysqlWire)
	writeExec := func(indent int, prefix string) {
		writeCursorCall(body, indent, parts, prefix+execKw+"cur.execute("+sqlRef, ")")
	}

	switch query.Cmd {
	case metadata.CmdExec:
		body.WriteIndentedLine(indent, withStmt)
		writeExec(indent+1, "")

	case metadata.CmdExecResult:
		body.WriteIndentedLine(indent, "cur = "+conn+".cursor()")
		writeExec(indent, "")
		body.WriteIndentedLine(indent, "return cur")

	case metadata.CmdExecRows:
		// execute() returns the affected-row count typed int in both
		// modules' stubs; asyncmy's cursor stub types rowcount as object,
		// which pyright strict rejects as a return value.
		body.WriteIndentedLine(indent, withStmt)
		writeExec(indent+1, "return ")

	case metadata.CmdExecLastId:
		// lastrowid is 0 (never None) when the statement inserted nothing;
		// AUTO_INCREMENT ids start at 1, so 0 maps to the documented None.
		body.WriteIndentedLine(indent, withStmt)
		writeExec(indent+1, "")
		body.WriteIndentedLine(indent+1, "return cur.lastrowid or None")

	case metadata.CmdOne:
		body.WriteIndentedLine(indent, withStmt)
		writeExec(indent+1, "")
		body.WriteIndentedLine(indent+1, fmt.Sprintf("row = %scur.fetchone()", execKw))
		body.WriteIndentedLine(indent, "if row is None:")
		body.WriteIndentedLine(indent+1, "return None")

		if query.Returns.IsStruct() {
			mb.rows.WriteStructReturn(body, indent, query.Returns)
		} else {
			mb.rows.WriteScalarReturn(body, indent, query.Returns)
		}

	case metadata.CmdMany:
		decodeHook := mb.rows.WriteDecodeHook(body, indent, query, mysqlResultType)
		sqlRef = writeSliceExpansion(body, indent, query, pyformatSliceJoin)
		manyArgs := append([]string{conn, sqlRef, decodeHook}, parts...)
		// Deliberately unsubscripted: QueryResults[T](...) would go through
		// typing's _GenericAlias.__call__ on every invocation (~10x call
		// overhead) for zero benefit - the return annotation carries the type.
		body.WriteWrappedCall(indent, "return QueryResults(", manyArgs, ")")
	}
}

// cursorType returns the annotation of the default cursor class.
func (mb *mysqlBase) cursorType() string { return mb.moduleName + ".cursors.Cursor" }
