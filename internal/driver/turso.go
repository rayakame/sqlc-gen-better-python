package driver

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

// tursoResultType is the row type of pyturso's default row factory. The
// bindings annotate fetches as Any, so the tuple form is the honest contract
// for decode hooks without needing extra imports.
const tursoResultType = "tuple[typing.Any, ...]"

// tursoConversion describes one inline turso type conversion: pyturso has no
// register_adapter/register_converter machinery and binds only None, numbers,
// strings, and bytes, so both directions convert inline in generated code.
// decodeFmt is the sqlite-converter equivalent (wire value -> Python type),
// wireFmt the sqlite-adapter equivalent (Python type -> bindable value); an
// empty wireFmt means the Python type binds natively.
type tursoConversion struct {
	sqlTypes  []string
	decodeFmt string
	wireFmt   string
}

const (
	tursoWireIsoformat  = "%s.isoformat()"
	tursoDecodeDatetime = "datetime.datetime.fromisoformat(%s)"
	tursoDecodeDecimal  = "decimal.Decimal(str(%s))"
	tursoDecodeBool     = "bool(%s)"
	tursoWireStr        = "str(%s)"
)

// tursoConversions mirrors the SQL type set of sqliteConversions: the same
// declared types convert, just inline instead of via registration. Decimal
// fidelity matches the sqlite drivers exactly - NUMERIC affinity stores the
// value as REAL either way, so both round-trip at double precision.
var tursoConversions = []tursoConversion{
	{sqlTypes: []string{sqlTypeDate}, decodeFmt: "datetime.date.fromisoformat(%s)", wireFmt: tursoWireIsoformat},
	{
		sqlTypes:  []string{sqlTypeDatetime, sqlTypeTimestamp},
		decodeFmt: tursoDecodeDatetime,
		wireFmt:   tursoWireIsoformat,
	},
	{sqlTypes: []string{sqlTypeDecimal}, decodeFmt: tursoDecodeDecimal, wireFmt: tursoWireStr},
	// bool binds natively (it is a Python int); only the return direction
	// needs the wrap back from the stored integer.
	{sqlTypes: []string{types.Bool, types.Boolean}, decodeFmt: tursoDecodeBool, wireFmt: ""},
	// pyturso rejects memoryview parameters, so blobs bind as bytes.
	{sqlTypes: []string{sqlTypeBlob}, decodeFmt: "memoryview(%s)", wireFmt: "bytes(%s)"},
}

// findTursoConversion returns the conversion spec for a SQL type, or nil.
// Precision variants like "decimal(10,5)" resolve through the "decimal" key,
// matching findSqliteConversion.
func findTursoConversion(sqlType string) *tursoConversion {
	for i := range tursoConversions {
		for _, name := range tursoConversions[i].sqlTypes {
			if name == sqlType {
				return &tursoConversions[i]
			}
		}
	}
	if strings.HasPrefix(sqlType, sqlTypeDecimal) {
		return findTursoConversion(sqlTypeDecimal)
	}

	return nil
}

// tursoNeedsConversion reports whether a SQL type decodes inline for turso.
func tursoNeedsConversion(sqlType string) bool {
	return findTursoConversion(sqlType) != nil
}

// tursoDecodeExpr returns the decode-expression template for a SQL type.
func tursoDecodeExpr(sqlType string) (string, bool) {
	if spec := findTursoConversion(sqlType); spec != nil {
		return spec.decodeFmt, true
	}

	return "", false
}

// tursoWire returns the wire-conversion template for a SQL type's parameters.
func tursoWire(sqlType string) (string, bool) {
	if spec := findTursoConversion(sqlType); spec != nil && spec.wireFmt != "" {
		return spec.wireFmt, true
	}

	return "", false
}

// tursoBase is the driver implementation for both pyturso flavors -
// turso_sync (module turso) and turso_async (module turso.aio). The two APIs
// mirror each other, so all emission differences derive from the async flag.
type tursoBase struct {
	async bool
	rows  *RowBuilder
}

var _ Driver = (*tursoBase)(nil)

// newTursoDriver creates the driver for one turso flavor. All type
// conversion happens inline: returns decode through the RowBuilder's decode
// expressions, parameters convert to their wire type via tursoWire.
func newTursoDriver(async bool) *tursoBase {
	return &tursoBase{
		async: async,
		rows:  newRowBuilderWithDecode(tursoNeedsConversion, tursoDecodeExpr),
	}
}

// Name returns the Python module name ("turso" or "turso.aio").
func (tb *tursoBase) Name() string {
	if tb.async {
		return "turso.aio"
	}

	return "turso"
}

// ConnType returns the connection type annotation, e.g. "turso.Connection".
func (tb *tursoBase) ConnType() string { return tb.Name() + ".Connection" }

// IsAsync reports whether this is the turso_async flavor.
func (tb *tursoBase) IsAsync() bool { return tb.async }

// SupportsCommand returns if the driver supports the command; the set
// matches the sqlite drivers (:copyfrom has no sqlite-dialect equivalent).
func (tb *tursoBase) SupportsCommand(cmd string) bool {
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

// TypeCheckingHook returns nil (no type-checking hook for turso drivers).
func (tb *tursoBase) TypeCheckingHook() []string {
	return nil
}

// NeedsConversion reports whether a SQL type needs runtime conversion.
func (tb *tursoBase) NeedsConversion(sqlType string) bool {
	return tursoNeedsConversion(sqlType)
}

// ConvertsInline reports whether a SQL type converts inline in decode code;
// turso converts everything inline (pyturso has no registration mechanism).
func (tb *tursoBase) ConvertsInline(sqlType string) bool {
	return tursoNeedsConversion(sqlType)
}

// WriteConversionSetup is a no-op for turso: with no adapter/converter
// registry, nothing runs at module import time.
func (tb *tursoBase) WriteConversionSetup(_ *writer.CodeWriter, _ *config.Config, _ []model.Query) bool {
	return false
}

// WriteQueryResultsClass writes the QueryResults class for :many queries, in
// its sync (turso) or async (turso.aio) variant. Both step through the
// cursor with fetchone: pyturso's async cursor does not declare the
// iteration dunders, and the sync flavor mirrors that shape so both emit the
// same state machine - iteration state is just the cursor itself.
func (tb *tursoBase) WriteQueryResultsClass(body *writer.CodeWriter) string {
	cursorType := tb.Name() + ".Cursor"
	awaitKw, nextDef, stopExc, article := "", defNextSync, stopIteration, "a "
	if tb.async {
		awaitKw, nextDef, stopExc, article = awaitPrefix, defNextAsync, stopAsyncIteration, "a "
	}

	body.QueryResults.WriteQueryResultsClassHeaderNoIterator(tb.ConnType(), []string{
		"self._cursor: " + cursorType + " | None = None",
	}, tursoResultType, tb.async)
	if tb.async {
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
	body.WriteQueryResultsNextDocstring(article+tb.Name()+" cursor", tb.async)
	body.WriteIndentedLine(2, "if self._cursor is None:")
	body.WriteIndentedLine(3, "self._cursor = "+awaitKw+"self._conn.execute(self._sql, self._args)")
	body.WriteIndentedLine(2, "record = "+awaitKw+"self._cursor.fetchone()")
	body.WriteIndentedLine(2, "if record is None:")
	body.WriteIndentedLine(3, "self._cursor = None")
	body.WriteIndentedLine(3, "raise "+stopExc)
	body.WriteIndentedLine(2, "return self._decode_hook(record)")

	return queryResultsClassName
}

func (tb *tursoBase) WriteQueryFunc(body *writer.CodeWriter, config *config.Config, query model.Query, indent int) {
	cursorType := tb.Name() + ".Cursor"
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

	conn := writeFuncSignature(body, tb, config, indent, query, annotation)

	indent++
	writeQueryDocstring(body, tb, config, query, indent, docRetType)
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
		base := conn + ".execute(" + sqlRef
		switch {
		case !tb.async:
			return prefix + base, ")" + attribute
		case attribute == "":
			return prefix + awaitPrefix + base, ")"
		default:
			return prefix + "(await " + base, "))" + attribute
		}
	}

	parts := expandParamsFlattenSlicesWire(query, tursoWire)
	switch query.Cmd {
	case metadata.CmdExec:
		head, tail := stmt("", "")
		writeSqliteCall(body, indent, parts, head, tail)

	case metadata.CmdExecResult:
		head, tail := stmt("return ", "")
		writeSqliteCall(body, indent, parts, head, tail)

	case metadata.CmdExecRows:
		head, tail := stmt("return ", ".rowcount")
		writeSqliteCall(body, indent, parts, head, tail)

	case metadata.CmdExecLastId:
		head, tail := stmt("return ", ".lastrowid")
		writeSqliteCall(body, indent, parts, head, tail)

	case metadata.CmdOne:
		prefix := "row = "
		if tb.async {
			// turso.aio's fetchone is itself a coroutine.
			prefix = "row = await "
		}
		head, tail := stmt(prefix, ".fetchone()")
		writeSqliteCall(body, indent, parts, head, tail)
		body.WriteIndentedLine(indent, "if row is None:")
		body.WriteIndentedLine(indent+1, "return None")

		if query.Returns.IsStruct() {
			tb.rows.WriteStructReturn(body, indent, query.Returns)
		} else {
			tb.rows.WriteScalarReturn(body, indent, query.Returns)
		}

	case metadata.CmdMany:
		decodeHook := tb.rows.WriteDecodeHook(body, indent, query, tursoResultType)
		sqlRef = writeSliceExpansion(body, indent, query)
		manyArgs := append([]string{conn, sqlRef, decodeHook}, parts...)
		// Deliberately unsubscripted: QueryResults[T](...) would go through
		// typing's _GenericAlias.__call__ on every invocation (~10x call
		// overhead) for zero benefit - the return annotation carries the type.
		body.WriteWrappedCall(indent, "return QueryResults(", manyArgs, ")")
	}
}
