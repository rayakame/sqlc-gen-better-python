package codegen

import "github.com/rayakame/sqlc-gen-better-python/internal/core"

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
	}

	body := NewIndentStringBuilder(config.IndentChar, config.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	for _, imp := range imports {
		body.WriteLine(imp)
	}
	return "models.py", []byte(body.String()), nil
}
