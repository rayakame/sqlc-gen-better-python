package driver

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
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

// placeholderStyle describes how bindable placeholders appear in a query's
// final SQL text. The sqlite-family drivers keep sqlc's native "?"; the
// MySQL drivers rewrite to pyformat "%s" at IR build time, which also
// changes the lexing rules for the surrounding text.
type placeholderStyle struct {
	// token is one bindable placeholder as it appears in the SQL.
	token string
	// joinExpr is the Sprintf template (one %s verb: the sequence
	// expression) for the runtime slice expansion - one comma-joined
	// placeholder per element, "NULL" for an empty sequence.
	joinExpr string
	// numbered marks placeholders that may carry a digit suffix ("?2",
	// sqlite only); the digits belong to the token.
	numbered bool
	// backslashEscapes marks '...' and "..." literals as honoring
	// backslash escapes in addition to doubled quotes (MySQL).
	backslashEscapes bool
	// hashComments marks "#" as a line-comment introducer (MySQL).
	hashComments bool
	// dashCommentNeedsGap requires whitespace (or end of input) after "--"
	// for it to start a comment (MySQL; "a--1" is arithmetic).
	dashCommentNeedsGap bool
	// backtickIdents marks `...` as quoted identifiers (MySQL).
	backtickIdents bool
	// versionComments marks /*! comment bodies as live SQL that can hold
	// placeholders (MySQL; sqlc's parser emits parameters for them).
	versionComments bool
	// doubledToken is a non-placeholder escape sequence to skip as a unit
	// ("%%" in pyformat text); empty when not applicable.
	doubledToken string
}

var (
	questionPlaceholders = placeholderStyle{
		token:    "?",
		joinExpr: `",".join("?" * len(%s)) or "NULL"`,
		numbered: true,
	}
	pyformatPlaceholders = placeholderStyle{
		token: "%s",
		// A tuple repeat, not a string repeat: join iterates strings
		// per-character, which only works for one-char placeholders.
		joinExpr:            `",".join(("%%s",) * len(%s)) or "NULL"`,
		backslashEscapes:    true,
		hashComments:        true,
		dashCommentNeedsGap: true,
		backtickIdents:      true,
		versionComments:     true,
		doubledToken:        "%%",
	}
)

// expandParams returns the Python argument expressions for a query's parameters.
// Bundled Params classes (query_parameter_limit) are expanded into their fields
// ("params.a, params.b") so drivers receive positional values. :copyfrom params
// are never passed through here - writeCopyFromBody builds its own records list.
func expandParams(query model.Query) []string {
	return expandParamsImpl(query, false, nil, questionPlaceholders)
}

// expandParamsFlattenSlices additionally star-unpacks sqlc.slice parameters
// ("*ids"), so after runtime placeholder expansion every "?" binds one element.
func expandParamsFlattenSlices(query model.Query) []string {
	return expandParamsImpl(query, true, nil, questionPlaceholders)
}

// expandParamsFlattenSlicesWire is expandParamsFlattenSlices for drivers that
// additionally convert parameters to their wire type inline.
func expandParamsFlattenSlicesWire(query model.Query, wire wireConvertFunc) []string {
	return expandParamsImpl(query, true, wire, questionPlaceholders)
}

// expandParamsPyformat is the MySQL variant: wire conversion plus the
// pyformat placeholder style of the rewritten SQL text.
func expandParamsPyformat(query model.Query, wire wireConvertFunc) []string {
	return expandParamsImpl(query, true, wire, pyformatPlaceholders)
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

	reused := false
	for _, p := range parts {
		if p.slice != "" && sliceMarkerCount(query, p.slice, ph) > 1 {
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
	if ordered, ok := orderByPlaceholders(query.SQL, plain, starred, ph); ok {
		return ordered
	}

	// Unmatchable SQL (hand-built IR in tests): consecutive copies keep the
	// argument count right even if the interleaving cannot be derived.
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p.slice != "" {
			for range sliceMarkerCount(query, p.slice, ph) {
				out = append(out, p.expr)
			}

			continue
		}
		out = append(out, p.expr)
	}

	return out
}

// orderByPlaceholders lines the flattened arguments up with the SQL text's
// placeholder sequence: plain expressions fill "?" slots in order, and every
// marker occurrence gets its slice's starred copy. Reports false when the SQL
// does not account for exactly the given arguments.
func orderByPlaceholders(sql string, plain []string, starred map[string]string, ph placeholderStyle) ([]string, bool) {
	seq := placeholderSequence(sql, ph)
	out := make([]string, 0, len(seq))
	next := 0
	used := make(map[string]struct{}, len(starred))
	for _, name := range seq {
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

// placeholderSequence scans the SQL for bindable placeholders in text order:
// the raw slice name for a /*SLICE:name*/<token> marker, "" for a plain
// (possibly numbered) token. String literals, quoted identifiers, and
// comments are skipped, so a token inside them never counts as a
// placeholder. The lexing rules follow the style: MySQL text adds backslash
// escapes, backtick identifiers, "#" comments, the "--"+whitespace rule,
// and the "%%" literal escape.
func placeholderSequence(sql string, ph placeholderStyle) []string {
	var seq []string
	for i := 0; i < len(sql); {
		rest := sql[i:]
		switch {
		case strings.HasPrefix(rest, "/*SLICE:"):
			end := strings.Index(rest, "*/"+ph.token)
			if end == -1 {
				return seq
			}
			seq = append(seq, rest[len("/*SLICE:"):end])
			i += end + len("*/") + len(ph.token)
		case ph.versionComments && strings.HasPrefix(rest, "/*!"):
			// The body is live SQL: keep scanning it; the closing */ passes
			// through the default case as ordinary text.
			i += len("/*!")
		case strings.HasPrefix(rest, "/*"):
			end := strings.Index(rest[len("/*"):], "*/")
			if end == -1 {
				return seq
			}
			i += len("/*") + end + len("*/")
		case strings.HasPrefix(rest, "--"):
			if ph.dashCommentNeedsGap && len(rest) > 2 && rest[2] > ' ' {
				// MySQL: "--x" is double unary minus, not a comment. Advance
				// one byte, not two: in an odd-length dash run the comment
				// starts mid-run, and the rewriter re-examines every
				// position the same way.
				i++

				continue
			}
			end := strings.IndexByte(rest, '\n')
			if end == -1 {
				return seq
			}
			i += end + 1
		case ph.hashComments && rest[0] == '#':
			end := strings.IndexByte(rest, '\n')
			if end == -1 {
				return seq
			}
			i += end + 1
		case rest[0] == '\'' || rest[0] == '"' || (ph.backtickIdents && rest[0] == '`'):
			// Backslash escapes never apply inside backticks.
			i = scanQuotedRegion(sql, i, ph.backslashEscapes && rest[0] != '`')
		case ph.doubledToken != "" && strings.HasPrefix(rest, ph.doubledToken):
			i += len(ph.doubledToken)
		case strings.HasPrefix(rest, ph.token):
			seq = append(seq, "")
			i += len(ph.token)
			if ph.numbered {
				for i < len(sql) && sql[i] >= '0' && sql[i] <= '9' {
					i++
				}
			}
		default:
			i++
		}
	}

	return seq
}

// scanQuotedRegion returns the index just past the closing quote of the
// quoted region starting at sql[i]. A doubled quote is an escape; with
// escapes, a backslash escapes the following byte. An unterminated region
// consumes the rest of the input.
func scanQuotedRegion(sql string, i int, escapes bool) int {
	quote := sql[i]
	j := i + 1
	for j < len(sql) {
		switch {
		case escapes && sql[j] == '\\' && j+1 < len(sql):
			j += 2
		case sql[j] != quote:
			j++
		case j+1 < len(sql) && sql[j+1] == quote:
			j += 2
		default:
			return j + 1
		}
	}

	return j
}

type sliceParam struct {
	// marker is the raw sqlc.slice name inside the /*SLICE:name*/? placeholder.
	marker string
	// expr is the Python expression holding the passed sequence.
	expr string
}

// sliceMarker returns the placeholder left in the SQL for a slice name:
// sqlc's raw marker for "?" styles, its rewritten form for pyformat.
func sliceMarker(name string, ph placeholderStyle) string {
	return "/*SLICE:" + name + "*/" + ph.token
}

// sliceMarkerCount reports how often a slice parameter's placeholder occurs in
// the query. sqlc merges same-named sqlc.slice uses into ONE parameter but
// keeps a marker per use site, so each occurrence needs its own expansion and
// its own copy of the arguments. Clamped to 1 for queries without the marker.
func sliceMarkerCount(query model.Query, name string, ph placeholderStyle) int {
	if count := strings.Count(query.SQL, sliceMarker(name, ph)); count > 1 {
		return count
	}

	return 1
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
	src := query.ConstantName
	for _, param := range params {
		args := []string{
			writer.PyQuote(sliceMarker(param.marker, ph)),
			fmt.Sprintf(ph.joinExpr, param.expr),
		}
		// A reused slice has one marker per use site: replace them all, with
		// the flattening param expansion supplying a copy of the args for each.
		if sliceMarkerCount(query, param.marker, ph) == 1 {
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
