package render

import (
	"fmt"
	"strconv"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
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
			fileBody.WriteIndentedLine(1, fmt.Sprintf(`%s = "%s"`, constant.Name, escapePyString(constant.Value)))
		}
	}

	return &plugin.File{
		Name:     enumFileName,
		Contents: fileBody.Bytes(),
	}
}

// escapePyString escapes a value for embedding in a double-quoted Python
// string literal, covering backslashes, quotes, and control characters.
// Go's Quote escaping (\", \\, \n, \t, \xNN, \uNNNN, ...) is a compatible
// subset of Python's string-literal escapes.
func escapePyString(value string) string {
	quoted := strconv.Quote(value)

	return quoted[1 : len(quoted)-1]
}
