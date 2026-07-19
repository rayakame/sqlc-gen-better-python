package driver

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
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
	first := fmt.Sprintf("conn: %s", d.ConnType())
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
	parts := make([]string, 0, len(query.Params))
	for _, param := range query.Params {
		if param.IsEmpty() {
			continue
		}
		if param.EmitTable && param.Table != nil {
			for _, col := range param.Table.Columns {
				parts = append(parts, convertParamExpr(fmt.Sprintf("%s.%s", param.Name, col.Name), col.Type))
			}

			continue
		}
		parts = append(parts, convertParamExpr(param.Name, param.Type))
	}

	return parts
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
	if !typ.DoOverride() || typ.DefaultType == "typing.Any" {
		return expr
	}
	converted := fmt.Sprintf("%s(%s)", typ.DefaultType, expr)
	if typ.IsList {
		converted = fmt.Sprintf("[%s(v) for v in %s]", typ.DefaultType, expr)
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
