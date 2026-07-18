package driver

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
)

// RowBuilder generates Python code that constructs a model instance from a database row.
// It handles embeds, nullable conversions, and overrides uniformly across all drivers.
type RowBuilder struct {
	// needsConversion checks whether a SQL type requires explicit conversion.
	needsConversion func(string) bool
}

// newRowBuilder creates a RowBuilder with the given conversion check function.
func newRowBuilder(needsConversion func(string) bool) *RowBuilder {
	return &RowBuilder{needsConversion: needsConversion}
}

// WriteStructReturn writes "return ModelType(col1=row[0], col2=row[1], ...)"
// handling embeds, nullable wrapping, and type conversions. Constructions that
// would exceed the line limit are exploded with magic trailing commas.
func (rb *RowBuilder) WriteStructReturn(body *writer.CodeWriter, indent int, ret model.QueryValue) {
	head := fmt.Sprintf("return %s(", ret.Type.Type)

	args := make([]string, 0, len(ret.Table.Columns))
	idx := 0
	for _, col := range ret.Table.Columns {
		if col.Embed != nil {
			args = append(args, rb.formatEmbedConstruction(col, &idx))
		} else {
			args = append(args, rb.formatColumnValue(col, idx))
			idx++
		}
	}

	single := head + strings.Join(args, ", ") + ")"
	if body.FitsLine(indent, single) {
		body.WriteIndentedLine(indent, single)

		return
	}

	body.WriteIndentedLine(indent, head)
	idx = 0
	for _, col := range ret.Table.Columns {
		if col.Embed != nil {
			embedHead := fmt.Sprintf("%s=%s(", col.Name, col.Type.Type)
			embedArgs := make([]string, 0, len(col.Embed.Columns))
			for _, embedCol := range col.Embed.Columns {
				embedArgs = append(embedArgs, rb.formatColumnValue(embedCol, idx))
				idx++
			}
			body.WriteWrappedCall(indent+1, embedHead, embedArgs, "),")

			continue
		}
		body.WriteIndentedLine(indent+1, rb.formatColumnValue(col, idx)+",")
		idx++
	}
	body.WriteIndentedLine(indent, ")")
}

// formatEmbedConstruction returns "name=EmbedType(field1=row[i], ...)".
func (rb *RowBuilder) formatEmbedConstruction(col model.Column, idx *int) string {
	inner := make([]string, 0, len(col.Embed.Columns))
	for _, embedCol := range col.Embed.Columns {
		inner = append(inner, rb.formatColumnValue(embedCol, *idx))
		*idx++
	}

	return fmt.Sprintf("%s=%s(%s)", col.Name, col.Type.Type, strings.Join(inner, ", "))
}

// WriteDecodeHook writes a _decode_hook function for :many queries or returns
// "operator.itemgetter(0)" for simple non-converted scalar returns. The blank
// lines around the nested def match ruff format's layout.
func (rb *RowBuilder) WriteDecodeHook(body *writer.CodeWriter, indent int, query model.Query, resultType string) string {
	// Simple scalar without conversion: use itemgetter.
	if !query.Returns.IsStruct() && !rb.columnNeedsConversion(query.Returns.Type) {
		return "operator.itemgetter(0)"
	}

	if body.DocstringsEnabled() {
		body.NewLine()
	}
	body.WriteIndentedLine(indent, fmt.Sprintf("def _decode_hook(row: %s) -> %s:", resultType, query.Returns.Type.Print()))
	if query.Returns.IsStruct() {
		rb.WriteStructReturn(body, indent+1, query.Returns)
	} else {
		body.WriteIndentedLine(indent+1, "return "+rb.convertExpr(query.Returns.Type, "row[0]"))
	}
	body.NewLine()

	return "_decode_hook"
}

// WriteScalarReturn writes the return statement for a non-struct :one query.
func (rb *RowBuilder) WriteScalarReturn(body *writer.CodeWriter, indent int, ret model.QueryValue) {
	if rb.columnNeedsConversion(ret.Type) {
		body.WriteIndentedLine(indent, "return "+rb.convertExpr(ret.Type, "row[0]"))
	} else {
		body.WriteIndentedLine(indent, "return row[0]")
	}
}

// convertExpr returns the Python expression converting a raw row value into
// its target type: constructor call for scalars, an element-wise comprehension
// for lists, both guarded against None for nullable values.
func (rb *RowBuilder) convertExpr(typ model.PyType, src string) string {
	if !rb.columnNeedsConversion(typ) {
		return src
	}
	expr := fmt.Sprintf("%s(%s)", typ.Type, src)
	if typ.IsList {
		expr = fmt.Sprintf("[%s(v) for v in %s]", typ.Type, src)
	}
	if typ.IsNullable {
		expr = fmt.Sprintf("%s if %s is not None else None", expr, src)
	}

	return expr
}

// columnNeedsConversion reports whether a column type needs explicit conversion.
// Enum columns always convert: the driver returns the raw value (e.g. str),
// which must be wrapped in the generated enum class.
func (rb *RowBuilder) columnNeedsConversion(typ model.PyType) bool {
	return typ.DoOverride() || typ.IsEnum || rb.needsConversion(typ.SQLType)
}

// formatColumnValue returns the Python expression for accessing a single column from a row.
func (rb *RowBuilder) formatColumnValue(col model.Column, idx int) string {
	return fmt.Sprintf("%s=%s", col.Name, rb.convertExpr(col.Type, "row["+strconv.Itoa(idx)+"]"))
}
