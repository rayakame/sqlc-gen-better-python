package driver

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

const sqliteResultType = "sqlite3.Row"

// sqliteBase is the complete driver implementation for both sqlite modules -
// sqlite3 (sync) and aiosqlite (async). All emission differences between the
// two are derived from moduleName and the async flag.
type sqliteBase struct {
	moduleName string // "sqlite3" or "aiosqlite"
	async      bool
	rows       *RowBuilder
}

var _ Driver = (*sqliteBase)(nil)

// newSqliteDriver creates the driver for one of the two sqlite modules. The
// RowBuilder never converts inline (except overrides/enums): registered
// converters handle the raw values, see WriteConversionSetup.
func newSqliteDriver(moduleName string, async bool) *sqliteBase {
	return &sqliteBase{
		moduleName: moduleName,
		async:      async,
		rows:       newRowBuilder(func(string) bool { return false }),
	}
}

// Name returns the Python module name ("sqlite3" or "aiosqlite").
func (sb *sqliteBase) Name() string { return sb.moduleName }

// ConnType returns the connection type annotation, e.g. "sqlite3.Connection".
func (sb *sqliteBase) ConnType() string { return sb.moduleName + ".Connection" }

// IsAsync reports whether this is the aiosqlite (async) driver.
func (sb *sqliteBase) IsAsync() bool { return sb.async }

// SupportsCommand returns if the driver supports the command.
func (sb *sqliteBase) SupportsCommand(cmd string) bool {
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

// TypeCheckingHook returns nil (no type-checking hook for sqlite drivers).
func (sb *sqliteBase) TypeCheckingHook() []string {
	return nil
}

// NeedsConversion reports whether a SQL type needs runtime conversion for sqlite.
func (sb *sqliteBase) NeedsConversion(sqlType string) bool {
	return sqliteNeedsConversion(sqlType)
}

// ConvertsInline always returns false: sqlite drivers convert via registered
// adapters/converters, not inline in decode code.
func (sb *sqliteBase) ConvertsInline(_ string) bool {
	return false
}

// WriteConversionSetup writes the adapter/converter functions and their
// registrations for every conversion type used by the given queries.
// Values written by adapters and read back by converters require the user's
// connection to be opened with detect_types=sqlite3.PARSE_DECLTYPES.
func (sb *sqliteBase) WriteConversionSetup(body *writer.CodeWriter, config *config.Config, queries []model.Query) bool {
	usage := SqliteConversionsUsed(queries)
	if !usage.Any() {
		return false
	}

	adapters := make([]string, 0, len(usage.uses))
	converters := make([]string, 0, len(usage.uses))
	for _, use := range usage.uses {
		spec := use.spec

		if use.adapter {
			body.WriteLine(fmt.Sprintf("def _adapt_%s(val: %s) -> %s:", spec.suffix, spec.pyType, spec.adaptRet))
			body.WriteIndentedLine(1, "return "+spec.adaptBody)
			body.NNewLine(2)
			adapters = append(
				adapters,
				fmt.Sprintf("%s.register_adapter(%s, _adapt_%s)", sb.moduleName, spec.pyType, spec.suffix),
			)
		}

		if use.converter {
			convBody := spec.convBody
			if config.Speedups && spec.speedupsBody != "" {
				convBody = spec.speedupsBody
			}
			body.WriteLine(fmt.Sprintf("def _convert_%s(val: bytes) -> %s:", spec.suffix, spec.pyType))
			body.WriteIndentedLine(1, "return "+convBody)
			body.NNewLine(2)
			for _, key := range spec.sqlTypes {
				converters = append(
					converters,
					fmt.Sprintf(`%s.register_converter("%s", _convert_%s)`, sb.moduleName, key, spec.suffix),
				)
			}
		}
	}

	for _, line := range adapters {
		body.WriteLine(line)
	}
	if len(adapters) != 0 && len(converters) != 0 {
		body.NewLine()
	}
	for _, line := range converters {
		body.WriteLine(line)
	}

	return true
}

// WriteQueryResultsClass writes the QueryResults class for :many queries,
// in its sync (sqlite3) or async (aiosqlite) variant.
func (sb *sqliteBase) WriteQueryResultsClass(body *writer.CodeWriter) string {
	cursorType := sb.moduleName + ".Cursor"
	awaitKw, iteratorType, nextDef, iterDunder, nextDunder, stopExc, article := "", "Iterator", "def __next__", "__iter__", "__next__", "StopIteration", "a "
	if sb.async {
		awaitKw, iteratorType, nextDef, iterDunder, nextDunder, stopExc, article = "await ", "AsyncIterator", "async def __anext__", "__aiter__", "__anext__", "StopAsyncIteration", "an "
	}

	body.QueryResults.WriteQueryResultsClassHeader(sb.ConnType(), []string{
		fmt.Sprintf("self._cursor: %s | None = None", cursorType),
		fmt.Sprintf("self._iterator: collections.abc.%s[%s] | None = None", iteratorType, sqliteResultType),
	}, sqliteResultType, sb.async)
	if sb.async {
		body.QueryResults.WriteQueryResultsAwaitFunction([]string{
			"result = await (await self._conn.execute(self._sql, self._args)).fetchall()",
			decodeRowsExpr,
		})
	} else {
		body.QueryResults.WriteQueryResultsCallFunction([]string{
			"result = self._conn.execute(self._sql, self._args).fetchall()",
			decodeRowsExpr,
		})
	}
	body.NewLine()
	body.WriteIndentedLine(1, nextDef+"(self) -> T:")
	body.WriteQueryResultsNextDocstring(article+sb.moduleName+" cursor", sb.async)
	body.WriteIndentedLine(2, "if self._cursor is None or self._iterator is None:")
	body.WriteIndentedLine(
		3,
		fmt.Sprintf("self._cursor: %s | None = %sself._conn.execute(self._sql, self._args)", cursorType, awaitKw),
	)
	body.WriteIndentedLine(3, fmt.Sprintf("self._iterator = self._cursor.%s()", iterDunder))
	body.WriteIndentedLine(2, "try:")
	body.WriteIndentedLine(3, fmt.Sprintf("record = %sself._iterator.%s()", awaitKw, nextDunder))
	body.WriteIndentedLine(2, "except "+stopExc+":")
	body.WriteIndentedLine(3, "self._cursor = None")
	body.WriteIndentedLine(3, "self._iterator = None")
	body.WriteIndentedLine(3, "raise")
	body.WriteIndentedLine(2, "return self._decode_hook(record)")

	return queryResultsClassName
}

func (sb *sqliteBase) WriteQueryFunc(body *writer.CodeWriter, config *config.Config, query model.Query, indent int) {
	cursorType := sb.moduleName + ".Cursor"
	var annotation, docRetType string
	switch query.Cmd {
	case metadata.CmdExec:
		annotation, docRetType = query.Returns.Type.Print(), ""
	case metadata.CmdExecResult:
		annotation, docRetType = cursorType, cursorType
	case metadata.CmdExecRows, metadata.CmdExecLastId:
		annotation, docRetType = query.Returns.Type.Print(), query.Returns.Type.Type
	case metadata.CmdOne:
		annotation, docRetType = query.Returns.Type.PrintOptional(), query.Returns.Type.Type
	case metadata.CmdMany:
		annotation, docRetType = "QueryResults["+query.Returns.Type.Print()+"]", query.Returns.Type.Print()
	}

	conn := writeFuncSignature(body, sb, config, indent, query, annotation)

	indent++
	writeQueryDocstring(body, sb, config, query, indent, docRetType)
	// :many delays this until after the decode hook, whose trailing blank
	// line keeps the assignment from touching the nested def (ruff E306).
	sqlRef := query.ConstantName
	if query.Cmd != metadata.CmdMany {
		sqlRef = writeSliceExpansion(body, indent, query)
	}

	// stmt builds the execute-statement head/tail with the correct await
	// wrapping for the async driver: accessing an attribute or method of the
	// cursor requires parenthesizing the awaited execute call.
	stmt := func(prefix, attribute string) (string, string) {
		base := fmt.Sprintf("%s.execute(%s", conn, sqlRef)
		switch {
		case !sb.async:
			return prefix + base, ")" + attribute
		case attribute == "":
			return prefix + "await " + base, ")"
		default:
			return prefix + "(await " + base, "))" + attribute
		}
	}

	switch query.Cmd {
	case metadata.CmdExec:
		head, tail := stmt("", "")
		writeSqliteCall(body, indent, query, head, tail)

	case metadata.CmdExecResult:
		head, tail := stmt("return ", "")
		writeSqliteCall(body, indent, query, head, tail)

	case metadata.CmdExecRows:
		head, tail := stmt("return ", ".rowcount")
		writeSqliteCall(body, indent, query, head, tail)

	case metadata.CmdExecLastId:
		head, tail := stmt("return ", ".lastrowid")
		writeSqliteCall(body, indent, query, head, tail)

	case metadata.CmdOne:
		prefix := "row = "
		if sb.async {
			// aiosqlite's fetchone is itself a coroutine.
			prefix = "row = await "
		}
		head, tail := stmt(prefix, ".fetchone()")
		writeSqliteCall(body, indent, query, head, tail)
		body.WriteIndentedLine(indent, "if row is None:")
		body.WriteIndentedLine(indent+1, "return None")

		if query.Returns.IsStruct() {
			sb.rows.WriteStructReturn(body, indent, query.Returns)
		} else {
			sb.rows.WriteScalarReturn(body, indent, query.Returns)
		}

	case metadata.CmdMany:
		decodeHook := sb.rows.WriteDecodeHook(body, indent, query, sqliteResultType)
		sqlRef = writeSliceExpansion(body, indent, query)
		manyArgs := append([]string{conn, sqlRef, decodeHook}, expandParamsFlattenSlices(query)...)
		// Deliberately unsubscripted: QueryResults[T](...) would go through
		// typing's _GenericAlias.__call__ on every invocation (~10x call
		// overhead) for zero benefit - the return annotation carries the type.
		body.WriteWrappedCall(indent, "return QueryResults(", manyArgs, ")")
	}
}

// writeSliceExpansion writes the runtime replacement of every sqlc.slice
// placeholder - one "?" per element, or "NULL" for an empty sequence so that
// "IN (NULL)" matches no rows - and returns the expression holding the final
// SQL: a local "sql" variable, or the untouched constant without slices.
func writeSliceExpansion(body *writer.CodeWriter, indent int, query model.Query) string {
	params := sliceParams(query)
	if len(params) == 0 {
		return query.ConstantName
	}
	src := query.ConstantName
	for _, param := range params {
		args := []string{
			writer.PyQuote(sliceMarker(param.marker)),
			fmt.Sprintf(`",".join("?" * len(%s)) or "NULL"`, param.expr),
		}
		// A reused slice has one marker per use site: replace them all, with
		// expandParamsFlattenSlices supplying a copy of the args for each.
		if sliceMarkerCount(query, param.marker) == 1 {
			args = append(args, "1")
		}
		body.WriteWrappedCall(indent, "sql = "+src+".replace(", args, ")")
		src = "sql"
	}

	return "sql"
}

// writeSqliteCall writes stmtHead+argsSegment+stmtTail on one line, hoisting a
// too-long parameter tuple into a local _args variable first so the statement
// stays within the line limit.
func writeSqliteCall(body *writer.CodeWriter, indent int, query model.Query, stmtHead, stmtTail string) {
	parts := expandParamsFlattenSlices(query)
	segment := ""
	switch {
	case len(parts) == 1:
		segment = fmt.Sprintf(", (%s,)", parts[0])
	case len(parts) > 1:
		segment = fmt.Sprintf(", (%s)", strings.Join(parts, ", "))
	}

	stmt := stmtHead + segment + stmtTail
	if body.FitsLine(indent, stmt) {
		body.WriteIndentedLine(indent, stmt)

		return
	}

	body.WriteIndentedLine(indent, "sql_args = (")
	for _, part := range parts {
		body.WriteIndentedLine(indent+1, part+",")
	}
	body.WriteIndentedLine(indent, ")")
	body.WriteIndentedLine(indent, stmtHead+", sql_args"+stmtTail)
}
