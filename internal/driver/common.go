package driver

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/sqllex"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

func writeFuncSignature(
	body *writer.CodeWriter,
	d Driver,
	config *config.Config,
	indent int,
	query model.Query,
	returnAnnotation string,
) string {
	conn := "conn"
	first := "conn: " + d.ConnType()
	if config.EmitClasses {
		first = "self"
		conn = "self._conn"
	}
	asyncPrefix := ""
	if d.IsAsync() && query.Cmd != metadata.CmdMany {
		asyncPrefix = "async "
	}

	signatureParams := 0
	for _, param := range query.Params {
		if !param.Repeated {
			signatureParams++
		}
	}
	args := []string{first}
	if signatureParams > config.OmitKwargsLimit {
		args = append(args, "*")
	}
	for _, param := range query.Params {
		// A repeated MySQL parameter binds again but is one argument.
		if param.Repeated {
			continue
		}
		args = append(args, fmt.Sprintf("%s: %s", param.Name, param.Type.Print()))
	}
	body.WriteWrappedCall(indent,
		fmt.Sprintf("%sdef %s(", asyncPrefix, query.FuncName),
		args,
		fmt.Sprintf(") -> %s:", returnAnnotation),
	)

	return conn
}

// wireConvertFunc returns a driver's wire-conversion template (with one %s
// verb for the element expression) for a SQL type - the inline equivalent of
// a registered sqlite adapter - or false when the driver binds the Python
// type natively.
type wireConvertFunc func(sqlType string) (string, bool)

// placeholderStyle pairs the dialect a driver's SQL is written in with the
// expression that expands a sqlc.slice at call time. The sqlite family
// multiplies the single-character "?"; pyformat needs a tuple repeat because
// join would otherwise walk "%s" character by character.
type placeholderStyle struct {
	dialect  sqllex.Dialect
	joinExpr string
}

var (
	questionStyle = placeholderStyle{ //nolint:gochecknoglobals
		dialect:  sqllex.SQLite,
		joinExpr: `",".join("?" * len(%s)) or "NULL"`,
	}
	pyformatStyle = placeholderStyle{ //nolint:gochecknoglobals
		dialect:  sqllex.MySQLPyformat,
		joinExpr: `",".join(("%%s",) * len(%s)) or "NULL"`,
	}
)

// querySlots scans a query's final SQL for its bindable positions. Both the
// argument ordering and the slice expansion read it, so a marker inside a
// string or comment can never be counted by one and skipped by the other.
func querySlots(query model.Query, ph placeholderStyle) []sqllex.Slot {
	return sqllex.Slots(query.SQL, ph.dialect)
}

// expandParams returns the Python argument expressions for a query's parameters.
// Bundled Params classes (query_parameter_limit) are expanded into their fields
// ("params.a, params.b") so drivers receive positional values. :copyfrom params
// are never passed through here - writeCopyFromBody builds its own records list.
func expandParams(query model.Query) []string {
	return expandParamsImpl(query, false, nil, questionStyle)
}

// expandParamsFlattenSlices additionally star-unpacks sqlc.slice parameters
// ("*ids"), so after runtime placeholder expansion every "?" binds one element.
func expandParamsFlattenSlices(query model.Query) []string {
	return expandParamsImpl(query, true, nil, questionStyle)
}

// expandParamsFlattenSlicesWire is expandParamsFlattenSlices for the turso
// drivers, which additionally convert parameters to their wire type inline.
func expandParamsFlattenSlicesWire(query model.Query, wire wireConvertFunc) []string {
	return expandParamsImpl(query, true, wire, questionStyle)
}

// expandParamsPyformat is the MySQL variant: wire conversion over text whose
// placeholders were rewritten to pyformat.
func expandParamsPyformat(query model.Query, wire wireConvertFunc) []string {
	return expandParamsImpl(query, true, wire, pyformatStyle)
}

func expandParamsImpl(query model.Query, flattenSlices bool, wire wireConvertFunc, ph placeholderStyle) []string {
	type part struct {
		expr string
		// slice is the raw marker name for slice params, "" otherwise.
		slice string
	}
	parts := make([]part, 0, len(query.Params))
	appendPart := func(expr string, typ model.PyType) {
		converted := convertParamExprWire(expr, typ, wire)
		slice := ""
		if flattenSlices && typ.SqlcSliceName != "" {
			converted = "*" + converted
			slice = typ.SqlcSliceName
		}
		parts = append(parts, part{expr: converted, slice: slice})
	}
	for _, param := range query.Params {
		if param.IsEmpty() {
			continue
		}
		if param.EmitTable && param.Table != nil {
			for _, col := range param.Table.Columns {
				appendPart(fmt.Sprintf("%s.%s", param.Name, col.Name), col.Type)
			}

			continue
		}
		appendPart(param.Name, param.Type)
	}

	slots := querySlots(query, ph)
	reused := false
	for _, p := range parts {
		if p.slice != "" && slotMarkerCount(slots, p.slice) > 1 {
			reused = true

			break
		}
	}
	if !reused {
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			out = append(out, p.expr)
		}

		return out
	}

	// A reused slice binds once per marker occurrence, and other placeholders
	// may sit between the use sites, so arguments must follow the SQL text
	// order rather than the parameter order.
	plain := make([]string, 0, len(parts))
	starred := make(map[string]string, len(parts))
	for _, p := range parts {
		if p.slice == "" {
			plain = append(plain, p.expr)
		} else {
			starred[p.slice] = p.expr
		}
	}
	if ordered, ok := orderByPlaceholders(slots, plain, starred); ok {
		return ordered
	}

	// Unmatchable SQL (hand-built IR in tests): consecutive copies keep the
	// argument count right even if the interleaving cannot be derived.
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p.slice != "" {
			for range slotMarkerCount(slots, p.slice) {
				out = append(out, p.expr)
			}

			continue
		}
		out = append(out, p.expr)
	}

	return out
}

// orderByPlaceholders lines the flattened arguments up with the query's bind
// order: plain expressions fill plain slots in order, and every marker
// occurrence gets its slice's starred copy. Reports false when the bind order
// does not account for exactly the given arguments - only reachable from
// hand-built IR, since transform derives SQL and its bind order together.
func orderByPlaceholders(slots []sqllex.Slot, plain []string, starred map[string]string) ([]string, bool) {
	out := make([]string, 0, len(slots))
	next := 0
	used := make(map[string]struct{}, len(starred))
	for _, slot := range slots {
		name := slot.Name
		if name == "" {
			if next >= len(plain) {
				return nil, false
			}
			out = append(out, plain[next])
			next++

			continue
		}
		expr, found := starred[name]
		if !found {
			return nil, false
		}
		used[name] = struct{}{}
		out = append(out, expr)
	}
	// A slice whose marker the scan never saw must fail too, or a truncated
	// scan would silently drop its arguments instead of using the fallback.
	if next != len(plain) || len(used) != len(starred) {
		return nil, false
	}

	return out, true
}

type sliceParam struct {
	// marker is the raw sqlc.slice name inside the /*SLICE:name*/? placeholder.
	marker string
	// expr is the Python expression holding the passed sequence.
	expr string
}

// slotMarkerCount reports how often a slice parameter's marker occurs in the
// scanned text. sqlc merges same-named sqlc.slice uses into ONE parameter but
// keeps a marker per use site, so each occurrence needs its own expansion and
// its own copy of the arguments. Clamped to 1 for hand-built IR without SQL.
func slotMarkerCount(slots []sqllex.Slot, name string) int {
	count := 0
	for _, slot := range slots {
		if slot.Name == name {
			count++
		}
	}
	if count > 1 {
		return count
	}

	return 1
}

// slotMarkerText returns a slice parameter's marker exactly as it appears in
// the SQL. Taking the scanned text rather than rebuilding it from the raw
// sqlc name is what keeps a name containing "%" - which the MySQL rewriter
// doubles - replaceable at runtime.
func slotMarkerText(slots []sqllex.Slot, name string) string {
	for _, slot := range slots {
		if slot.Name == name {
			return slot.Marker
		}
	}

	return ""
}

// sliceParams collects the sqlc.slice parameters of a query, including fields
// of a bundled Params class.
func sliceParams(query model.Query) []sliceParam {
	var params []sliceParam
	for _, param := range query.Params {
		if param.IsEmpty() {
			continue
		}
		if param.EmitTable && param.Table != nil {
			for _, col := range param.Table.Columns {
				if col.Type.SqlcSliceName != "" {
					params = append(
						params,
						sliceParam{marker: col.Type.SqlcSliceName, expr: fmt.Sprintf("%s.%s", param.Name, col.Name)},
					)
				}
			}

			continue
		}
		if param.Type.SqlcSliceName != "" {
			params = append(params, sliceParam{marker: param.Type.SqlcSliceName, expr: param.Name})
		}
	}

	return params
}

// writeCursorNextMethod writes the cursor-backed __next__/__anext__ shared by
// the asyncpg and psycopg QueryResults classes (the sqlite drivers emit their
// own variant inline): open the cursor lazily via cursorInit, forward one
// record, and reset both fields on exhaustion so iteration can restart.
func writeCursorNextMethod(body *writer.CodeWriter, async bool, cursorDesc, cursorInit string) {
	nextDef, iterDunder, nextDunder, stopExc, awaitKw := defNextSync, "__iter__", "__next__", stopIteration, ""
	if async {
		nextDef, iterDunder, nextDunder, stopExc, awaitKw = defNextAsync, "__aiter__", "__anext__", stopAsyncIteration, awaitPrefix
	}
	body.NewLine()
	body.WriteIndentedLine(1, nextDef+"(self) -> T:")
	body.WriteQueryResultsNextDocstring(cursorDesc, async)
	body.WriteIndentedLine(2, "if self._cursor is None or self._iterator is None:")
	body.WriteIndentedLine(3, cursorInit)
	body.WriteIndentedLine(3, fmt.Sprintf("self._iterator = self._cursor.%s()", iterDunder))
	body.WriteIndentedLine(2, "try:")
	body.WriteIndentedLine(3, fmt.Sprintf("record = %sself._iterator.%s()", awaitKw, nextDunder))
	body.WriteIndentedLine(2, "except "+stopExc+":")
	body.WriteIndentedLine(3, "self._cursor = None")
	body.WriteIndentedLine(3, "self._iterator = None")
	body.WriteIndentedLine(3, "raise")
	body.WriteIndentedLine(2, "return self._decode_hook(record)")
}

// writeQueryDocstring writes the docstring for a generated query function.
// retType is the type shown in the Returns section (driver-specific for some
// commands); pass "" for commands without one (:exec).
func writeQueryDocstring(body *writer.CodeWriter, d Driver, cfg *config.Config, query model.Query, indent int, retType string) {
	connType := ""
	if !cfg.EmitClasses {
		connType = d.ConnType()
	}
	args := make([]writer.DocArg, 0, len(query.Params))
	for _, param := range query.Params {
		if param.IsEmpty() || param.Repeated {
			continue
		}
		extra := ""
		if query.Cmd == metadata.CmdCopyFrom {
			extra = "A list of params for rows that should be inserted."
		}
		args = append(args, writer.DocArg{Name: param.Name, Type: param.Type.Print(), Extra: extra})
	}
	body.WriteQueryFunctionDocstring(indent, &query, connType, args, retType)
}

// convertParamExpr converts an overridden argument back to the type the driver
// expects (its DefaultType) before passing it on. List values convert
// element-wise, mirroring RowBuilder.convertExpr on the return side.
// Overrides on SQL types the plugin does not know map to typing.Any, which
// is not instantiable - those values pass through unconverted (there is no
// registered adapter for unknown types either).
func convertParamExpr(expr string, typ model.PyType) string {
	return convertParamExprWire(expr, typ, nil)
}

// convertParamExprWire is convertParamExpr with a driver wire conversion
// composed on top: the override conversion (back to DefaultType, or through
// the user's to_db function) runs first, then the wire template wraps the
// result - mirroring how sqlite adapters receive already-override-converted
// values. List and nullable wrapping applies once around the composed
// element expression.
func convertParamExprWire(expr string, typ model.PyType, wire wireConvertFunc) string {
	elemFmt := "%s"
	if typ.DoOverride() {
		callable := typ.DefaultType
		if typ.ConverterTo != "" {
			callable = typ.ConverterTo
		} else if typ.DefaultType == types.Any {
			// A typing.Any default is not instantiable, so the value passes
			// through unconverted and skips the wire conversion too - the
			// wire templates assume the default Python type.
			return expr
		}
		elemFmt = callable + "(%s)"
	}
	if wire != nil {
		if wireFmt, ok := wire(typ.SQLType); ok {
			elemFmt = fmt.Sprintf(wireFmt, elemFmt)
		}
	}
	if elemFmt == "%s" {
		return expr
	}
	converted := fmt.Sprintf(elemFmt, expr)
	if typ.IsList {
		converted = fmt.Sprintf("[%s for v in %s]", fmt.Sprintf(elemFmt, "v"), expr)
	}
	if typ.IsNullable {
		return fmt.Sprintf("%s if %s is not None else None", converted, expr)
	}

	return converted
}

func writeExecRowsReturn(body *writer.CodeWriter, config *config.Config, indent int) {
	if config.Speedups {
		body.WriteIndentedLine(indent, "return int(n) if (n := r.split()[-1]).isdigit() else 0")
	} else {
		body.WriteIndentedLine(indent, "return int(n) if (p := r.split()) and (n := p[-1]).isdigit() else 0")
	}
}

// writeSliceExpansion writes the runtime replacement of every sqlc.slice
// placeholder - one placeholder per element, or "NULL" for an empty sequence
// so that "IN (NULL)" matches no rows - and returns the expression holding
// the final SQL: a local "sql" variable, or the untouched constant without
// slices.
func writeSliceExpansion(body *writer.CodeWriter, indent int, query model.Query, ph placeholderStyle) string {
	params := sliceParams(query)
	if len(params) == 0 {
		return query.ConstantName
	}
	slots := querySlots(query, ph)
	src := query.ConstantName
	for _, param := range params {
		args := []string{
			writer.PyQuote(slotMarkerText(slots, param.marker)),
			fmt.Sprintf(ph.joinExpr, param.expr),
		}
		// A reused slice has one marker per use site: replace them all, with
		// the flattening param expansion supplying a copy of the args for each.
		if slotMarkerCount(slots, param.marker) == 1 {
			args = append(args, "1")
		}
		body.WriteWrappedCall(indent, "sql = "+src+".replace(", args, ")")
		src = "sql"
	}

	return "sql"
}

// writeCursorCall writes stmtHead+argsSegment+stmtTail on one line, hoisting a
// too-long parameter tuple into a local "sql_args" variable first so the
// statement stays within the line limit. parts are the already-expanded (and,
// for wire-converting drivers, already-converted) argument expressions. Shared
// by the sqlite, turso, and MySQL drivers.
func writeCursorCall(body *writer.CodeWriter, indent int, parts []string, stmtHead, stmtTail string) {
	// A parameterless statement passes no args at all: an empty tuple would
	// still make the MySQL drivers interpolate the (undoubled) SQL text.
	if len(parts) == 0 {
		body.WriteIndentedLine(indent, stmtHead+stmtTail)

		return
	}

	segment := fmt.Sprintf(", (%s,)", parts[0])
	if len(parts) > 1 {
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
