package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
)

func BuildModelFile(config *core.Config, tables []core.Table) (string, []byte, error) {
	imports := make([]string, 0)

	for _, table := range tables {
		for _, col := range table.Columns {
			if imp := core.ExtractImport(col.Type); len(imp) != 0 {
				imports = core.AppendUniqueString(imports, imp)
			}
		}
	}

	switch config.ModelType {
	case core.ModelTypeDataclass:
		imports = append(imports, "from dataclasses import dataclass")
	case core.ModelTypeAttrs:
		imports = append(imports, "import attrs")
	default:
		return "", nil, fmt.Errorf("unsupported model type: %s", config.ModelType)
	}

	body := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	for _, imp := range imports {
		body.WriteLine(imp)
	}
	for _, table := range tables {
		body.WriteString("\n")
		body.WriteString("\n")
		buildModelHeader(config, table, body)
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
	return "models.py", []byte(body.String()), nil
}

func buildModelHeader(config *core.Config, table core.Table, body *IndentStringBuilder) {
	if config.ModelType == core.ModelTypeDataclass {
		body.WriteLine("@dataclass()")
	} else if config.ModelType == core.ModelTypeAttrs {
		body.WriteLine("@attrs.define()")
	}
	body.WriteLine("class " + table.Name + ":")
}
