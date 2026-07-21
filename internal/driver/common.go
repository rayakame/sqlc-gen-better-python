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

	args := []string{first}
	if len(query.Params) > config.OmitKwargsLimit {
		args = append(args, "*")
	}
	for _, param := range query.Params {
		args = append(args, fmt.Sprintf("%s: %s", param.Name, param.Type.Print()))
	}
	body.WriteWrappedCall(indent,
		fmt.Sprintf("%sdef %s(", asyncPrefix, query.FuncName),
		args,
		fmt.Sprintf(") -> %s:", returnAnnotation),
	)

	return conn
}

// expandParams returns the Python argument expressions for a query's parameters.
// Bundled Params classes (query_parameter_limit) are expanded into their fields
// ("params.a, params.b") so drivers receive positional values. :copyfrom params
// are never passed through here - writeCopyFromBody builds its own records list.
func expandParams(query model.Query) []string {
	return expandParamsImpl(query, false)
}

// expandParamsFlattenSlices additionally star-unpacks sqlc.slice parameters
// ("*ids"), so after runtime placeholder expansion every "?" binds one element.
func expandParamsFlattenSlices(query model.Query) []string {
	return expandParamsImpl(query, true)
}

func expandParamsImpl(query model.Query, flattenSlices bool) []string {
	type part struct {
		expr string
		// slice is the raw marker name for slice params, "" otherwise.
		slice string
	}
	parts := make([]part, 0, len(query.Params))
	appendPart := func(expr string, typ model.PyType) {
		converted := convertParamExpr(expr, typ)
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
		if p.slice != "" && sliceMarkerCount(query, p.slice) > 1 {
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
	if ordered, ok := orderByPlaceholders(query.SQL, plain, starred); ok {
		return ordered
	}

	// Unmatchable SQL (hand-built IR in tests): consecutive copies keep the
	// argument count right even if the interleaving cannot be derived.
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p.slice != "" {
			for range sliceMarkerCount(query, p.slice) {
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
func orderByPlaceholders(sql string, plain []string, starred map[string]string) ([]string, bool) {
	seq := placeholderSequence(sql)
	out := make([]string, 0, len(seq))
	next := 0
	for _, name := range seq {
		if name == "" {
			if next >= len(plain) {
				return nil, false
			}
			out = append(out, plain[next])
			next++

			continue
		}
		expr, ok := starred[name]
		if !ok {
			return nil, false
		}
		out = append(out, expr)
	}
	if next != len(plain) {
		return nil, false
	}

	return out, true
}

// placeholderSequence scans the SQL for bindable placeholders in text order:
// the raw slice name for a /*SLICE:name*/? marker, "" for a plain (possibly
// numbered) "?". String literals, quoted identifiers, and comments are
// skipped, so a "?" inside them never counts as a placeholder.
func placeholderSequence(sql string) []string {
	var seq []string
	for i := 0; i < len(sql); {
		rest := sql[i:]
		switch {
		case strings.HasPrefix(rest, "/*SLICE:"):
			end := strings.Index(rest, "*/?")
			if end == -1 {
				return seq
			}
			seq = append(seq, rest[len("/*SLICE:"):end])
			i += end + len("*/?")
		case strings.HasPrefix(rest, "/*"):
			end := strings.Index(rest[len("/*"):], "*/")
			if end == -1 {
				return seq
			}
			i += len("/*") + end + len("*/")
		case strings.HasPrefix(rest, "--"):
			end := strings.IndexByte(rest, '\n')
			if end == -1 {
				return seq
			}
			i += end + 1
		case rest[0] == '\'' || rest[0] == '"':
			quote := rest[0]
			j := i + 1
			for j < len(sql) {
				if sql[j] != quote {
					j++

					continue
				}
				if j+1 < len(sql) && sql[j+1] == quote {
					// A doubled quote is an escape, not the end.
					j += 2

					continue
				}

				break
			}
			i = j + 1
		case rest[0] == '?':
			seq = append(seq, "")
			i++
			for i < len(sql) && sql[i] >= '0' && sql[i] <= '9' {
				i++
			}
		default:
			i++
		}
	}

	return seq
}

type sliceParam struct {
	// marker is the raw sqlc.slice name inside the /*SLICE:name*/? placeholder.
	marker string
	// expr is the Python expression holding the passed sequence.
	expr string
}

// sliceMarker returns the placeholder sqlc leaves in the SQL for a slice name.
func sliceMarker(name string) string {
	return "/*SLICE:" + name + "*/?"
}

// sliceMarkerCount reports how often a slice parameter's placeholder occurs in
// the query. sqlc merges same-named sqlc.slice uses into ONE parameter but
// keeps a marker per use site, so each occurrence needs its own expansion and
// its own copy of the arguments. Clamped to 1 for queries without the marker.
func sliceMarkerCount(query model.Query, name string) int {
	if count := strings.Count(query.SQL, sliceMarker(name)); count > 1 {
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
		if param.IsEmpty() {
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
	if !typ.DoOverride() {
		return expr
	}
	callable := typ.DefaultType
	if typ.ConverterTo != "" {
		callable = typ.ConverterTo
	} else if typ.DefaultType == types.Any {
		return expr
	}
	converted := fmt.Sprintf("%s(%s)", callable, expr)
	if typ.IsList {
		converted = fmt.Sprintf("[%s(v) for v in %s]", callable, expr)
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
