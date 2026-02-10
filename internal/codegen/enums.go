package codegen

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func (dr *Driver) BuildPyEnumsFile(imp *core.Importer, enums []core.Enum) (*plugin.File, error) {
	fileName, fileContent, err := dr.buildPyEnums(imp, enums)
	if err != nil {
		return nil, err
	}
	return &plugin.File{
		Name:     core.SQLToPyFileName(fileName),
		Contents: fileContent,
	}, nil
}

func (dr *Driver) buildPyEnum(enum *core.Enum, body *builders.IndentStringBuilder) {
	body.WriteLine(fmt.Sprintf("class %s(enum.StrEnum):", enum.Name))
	for _, constant := range enum.Constants {
		body.WriteIndentedLine(1, constant.Name+" = \""+constant.Value+"\"")
	}
}

func (dr *Driver) buildPyEnums(imp *core.Importer, enums []core.Enum) (string, []byte, error) {
	fileName := "enums.sql"
	body := dr.GetStringBuilder()
	body.WriteSqlcHeader()
	body.WriteModelFileModuleDocstring()
	body.WriteImportAnnotations()
	body.WriteLine("__all__: collections.abc.Sequence[str] = (")
	for _, table := range enums {
		body.WriteIndentedLine(1, fmt.Sprintf("\"%s\",", table.Name))
	}
	body.WriteLine(")")
	body.WriteString("\n")
	std, tye, pkg := imp.Imports(fileName)
	for _, imp := range std {
		body.WriteLine(imp)
	}
	if len(tye) != 0 {
		if len(std) != 0 {
			body.NewLine()
		}
		if !dr.conf.OmitTypecheckingBlock {
			body.WriteLine("if typing.TYPE_CHECKING:")
			for _, imp := range tye {
				body.WriteIndentedLine(1, imp)
			}
		} else {
			for _, imp := range tye {
				body.WriteLine(imp)
			}
		}
	}
	for i, imp := range pkg {
		if i == 0 {
			body.NewLine()
		}
		body.WriteLine(imp)
	}

	for _, enum := range enums {
		body.WriteString("\n")
		body.WriteString("\n")
		dr.buildPyEnum(&enum, body)
	}
	return fileName, body.Bytes(), nil
}
