package render

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

const tableFileName = "models.py"

func (r *Renderer) renderTable(body *writer.CodeWriter, table model.Table) {
	inheritance := ""
	switch r.config.ModelType {
	case config.ModelTypeAttrs:
		body.WriteLine("@attrs.define()")
	case config.ModelTypeDataclass:
		body.WriteLine("@dataclasses.dataclass()")
	case config.ModelTypeMsgspec:
		inheritance = "(msgspec.Struct)"
	case config.ModelTypePydantic:
		inheritance = "(pydantic.BaseModel)"
	}
	body.WriteLine(fmt.Sprintf("class %s%s:", table.Name, inheritance))
	body.WriteModelClassDocstring(&table)
	if r.config.ModelType == config.ModelTypePydantic {
		// Without this, class definition fails for field types pydantic has no
		// core schema for (memoryview, override types); isinstance validation
		// still applies to them.
		body.WriteIndentedLine(1, "model_config = pydantic.ConfigDict(arbitrary_types_allowed=True)")
		body.NewLine()
	}
	for _, column := range table.Columns {
		body.WriteIndentedLine(1, column.Name+": "+column.Type.Print())
	}
}

func (r *Renderer) renderTables(tables []model.Table) *plugin.File {
	fileBody := r.getCodeWriter()
	fileBody.WriteSqlcHeader(nil)
	fileBody.WriteModelFileModuleDocstring()
	fileBody.WriteFutureImport()

	all := make([]string, len(tables))
	for i, table := range tables {
		all[i] = table.Name
	}
	fileBody.WriteAll(all)
	fileBody.NewLine()

	r.importResolver.ModelImports(tables).Write(fileBody, r.config.OmitTypecheckingBlock, nil)

	for _, table := range tables {
		fileBody.NNewLine(2)
		r.renderTable(fileBody, table)
	}

	return &plugin.File{
		Name:     tableFileName,
		Contents: fileBody.Bytes(),
	}
}
