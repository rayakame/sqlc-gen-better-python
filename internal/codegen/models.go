package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
)

func BuildModelFile(imp *core.Importer, tables []core.Table) (string, []byte, error) {
	fileName := "models.py"
	body := NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	body.WriteImportAnnotations()
	body.WriteLine("__all__: typing.Sequence[str] = (")
	for _, table := range tables {
		body.WriteIndentedLine(1, fmt.Sprintf("\"%s\",", table.Name))
	}
	body.WriteLine(")")
	body.WriteString("\n")
	for _, imp := range imp.Imports(fileName) {
		body.WriteLine(imp)
	}
	for _, table := range tables {
		body.WriteString("\n")
		body.WriteString("\n")
		BuildModel(imp.C, &table, body)
	}
	return fileName, []byte(body.String()), nil
}

func BuildModel(config *core.Config, table *core.Table, body *IndentStringBuilder) {
	if config.ModelType == core.ModelTypeDataclass {
		body.WriteLine("@dataclasses.dataclass()")
	} else if config.ModelType == core.ModelTypeAttrs {
		body.WriteLine("@attrs.define()")
	}
	body.WriteLine("class " + table.Name + ":")
	for _, col := range table.Columns {
		type_ := col.Type.Type
		if col.Type.IsList {
			type_ = "typing.List[" + type_ + "]"
		}
		if col.Type.IsNullable {
			type_ = "typing.Optional[" + type_ + "]"
		}
		body.WriteIndentedString(1, col.Name+": "+type_)
		if config.ModelType == core.ModelTypeAttrs {
			body.WriteString(" = attrs.field()")
		}
		body.WriteString("\n")
	}
}
