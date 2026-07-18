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
// handling embeds, nullable wrapping, and type conversions.
func (rb *RowBuilder) WriteStructReturn(body *writer.CodeWriter, indent int, ret model.QueryValue) {
	body.WriteIndentedString(indent, fmt.Sprintf("return %s(", ret.Type.Type))
	idx := 0
	for colIdx, col := range ret.Table.Columns {
		if colIdx != 0 {
			body.WriteString(", ")
		}
		if col.Embed != nil {
			rb.writeEmbedConstruction(body, col, &idx)
		} else {
			rb.writeColumnAssignment(body, col, idx)
			idx++
		}
	}
	body.WriteLine(")")
}

// WriteDecodeHook writes a _decode_hook function for :many queries or returns
// "operator.itemgetter(0)" for simple non-converted scalar returns.
func (rb *RowBuilder) WriteDecodeHook(body *writer.CodeWriter, indent int, query model.Query, resultType string) string {
	// Simple scalar without conversion: use itemgetter.
	if !query.Returns.IsStruct() && !rb.columnNeedsConversion(query.Returns.Type) {
		return "operator.itemgetter(0)"
	}

	// Scalar with conversion: simple decode hook.
	if !query.Returns.IsStruct() {
		body.WriteIndentedLine(indent, fmt.Sprintf("def _decode_hook(row: %s) -> %s:", resultType, query.Returns.Type.Type))
		body.WriteIndentedLine(indent+1, fmt.Sprintf("return %s(row[0])", query.Returns.Type.Type))
		return "_decode_hook"
	}

	// Struct: full decode hook with column assignments.
	body.WriteIndentedLine(indent, fmt.Sprintf("def _decode_hook(row: %s) -> %s:", resultType, query.Returns.Type.Type))
	rb.WriteStructReturn(body, indent+1, query.Returns)
	return "_decode_hook"
}

// WriteScalarReturn writes the return statement for a non-struct :one query.
func (rb *RowBuilder) WriteScalarReturn(body *writer.CodeWriter, indent int, ret model.QueryValue) {
	if rb.columnNeedsConversion(ret.Type) {
		body.WriteIndentedLine(indent, fmt.Sprintf("return %s(row[0])", ret.Type.Type))
	} else {
		body.WriteIndentedLine(indent, "return row[0]")
	}
}

// columnNeedsConversion reports whether a column type needs explicit conversion.
// Enum columns always convert: the driver returns the raw value (e.g. str),
// which must be wrapped in the generated enum class.
func (rb *RowBuilder) columnNeedsConversion(typ model.PyType) bool {
	return typ.DoOverride() || typ.IsEnum || rb.needsConversion(typ.SQLType)
}

// writeEmbedConstruction writes "name=EmbedType(field1=row[i], field2=row[i+1], ...)".
func (rb *RowBuilder) writeEmbedConstruction(body *writer.CodeWriter, col model.Column, idx *int) {
	body.WriteString(fmt.Sprintf("%s=%s(", col.Name, col.Type.Type))
	var inner []string
	for _, embedCol := range col.Embed.Columns {
		inner = append(inner, rb.formatColumnValue(embedCol, *idx))
		*idx++
	}
	body.WriteString(strings.Join(inner, ", ") + ")")
}

// writeColumnAssignment writes "name=row[i]" or "name=Type(row[i])" depending on conversion needs.
func (rb *RowBuilder) writeColumnAssignment(body *writer.CodeWriter, col model.Column, idx int) {
	body.WriteString(rb.formatColumnValue(col, idx))
}

// formatColumnValue returns the Python expression for accessing a single column from a row.
func (rb *RowBuilder) formatColumnValue(col model.Column, idx int) string {
	idxStr := strconv.Itoa(idx)
	if !rb.columnNeedsConversion(col.Type) {
		return fmt.Sprintf("%s=row[%s]", col.Name, idxStr)
	}
	if col.Type.IsNullable {
		return fmt.Sprintf("%s=%s(row[%s]) if row[%s] is not None else None",
			col.Name, col.Type.Type, idxStr, idxStr)
	}
	return fmt.Sprintf("%s=%s(row[%s])", col.Name, col.Type.Type, idxStr)
}
