package driver

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func (dr *Driver) BuildPyTablesFile(imp *core.Importer, tables []core.Table) (*plugin.File, error) {
	fileName, fileContent, err := dr.buildPyTables(imp, tables)
	if err != nil {
		return nil, err
	}
	return &plugin.File{
		Name:     core.SQLToPyFileName(fileName),
		Contents: fileContent,
	}, nil
}

func BuildPyTabel(modelType string, table *core.Table, body *codegen.IndentStringBuilder) {
	if modelType == core.ModelTypeDataclass {
		body.WriteLine("@dataclasses.dataclass()")
	} else if modelType == core.ModelTypeAttrs {
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
		if modelType == core.ModelTypeAttrs {
			body.WriteString(" = attrs.field()")
		}
		body.WriteString("\n")
	}
}

func (dr *Driver) buildPyTables(imp *core.Importer, tables []core.Table) (string, []byte, error) {
	fileName := "models.sql"
	body := codegen.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
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
		BuildPyTabel(imp.C.ModelType, &table, body)
	}
	return fileName, []byte(body.String()), nil
}
