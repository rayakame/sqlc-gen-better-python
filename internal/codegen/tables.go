package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
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

func BuildPyTabel(modelType string, table *core.Table, body *builders.IndentStringBuilder) {
	if modelType == core.ModelTypeDataclass {
		body.WriteLine("@dataclasses.dataclass()")
	} else if modelType == core.ModelTypeAttrs {
		body.WriteLine("@attrs.define()")
	}
	inheritance := ""
	if modelType == core.ModelTypeMsgspec {
		inheritance = "(msgspec.Struct)"
	}
	body.WriteLine(fmt.Sprintf("class %s%s:", table.Name, inheritance))
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
		} else if modelType == core.ModelTypeMsgspec {
			body.WriteString(" = msgspec.field()")
		}
		body.WriteString("\n")
	}
}

func (dr *Driver) buildPyTables(imp *core.Importer, tables []core.Table) (string, []byte, error) {
	fileName := "models.sql"
	body := builders.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	body.WriteImportAnnotations()
	body.WriteLine("__all__: collections.abc.Sequence[str] = (")
	for _, table := range tables {
		body.WriteIndentedLine(1, fmt.Sprintf("\"%s\",", table.Name))
	}
	body.WriteLine(")")
	body.WriteString("\n")
	std, tye, pkg := imp.Imports(fileName)
	for _, imp := range std {
		body.WriteLine(imp)
	}
	if len(tye) != 0 {
		body.WriteLine("if typing.TYPE_CHECKING:")
		for _, imp := range tye {
			body.WriteIndentedLine(1, imp)
		}
	}
	body.WriteLine("")
	for _, imp := range pkg {
		body.WriteLine(imp)
	}
	for _, table := range tables {
		body.WriteString("\n")
		body.WriteString("\n")
		BuildPyTabel(imp.C.ModelType, &table, body)
	}
	return fileName, []byte(body.String()), nil
}
