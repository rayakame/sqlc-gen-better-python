package render

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

const enumFileName = "enums.py"

func (r *Renderer) renderEnums(enums []model.Enum) *plugin.File {
	fileBody := r.getCodeWriter()
	fileBody.WriteSqlcHeader(nil)
	fileBody.WriteEnumsFileModuleDocstring()
	fileBody.WriteFutureImport()

	all := make([]string, len(enums))
	for i, enum := range enums {
		all[i] = enum.Name
	}
	fileBody.WriteAll(all)
	fileBody.NewLine()

	r.importResolver.EnumImports().Write(fileBody, r.config.OmitTypecheckingBlock, nil)

	for _, enum := range enums {
		fileBody.NNewLine(2)
		fileBody.WriteLine(fmt.Sprintf("class %s(enum.StrEnum):", enum.Name))
		fileBody.WriteEnumClassDocstring(enum.Name)
		for _, constant := range enum.Constants {
			fileBody.WriteIndentedLine(1, fmt.Sprintf("%s = %s", constant.Name, writer.PyQuote(constant.Value)))
		}
	}

	return &plugin.File{
		Name:     enumFileName,
		Contents: fileBody.Bytes(),
	}
}

