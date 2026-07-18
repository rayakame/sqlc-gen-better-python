package render

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func (r *Renderer) renderQueriesModule(moduleName string, queries []model.Query, hasTables, hasEnums bool) *plugin.File {
	fileBody := r.getCodeWriter()
	fileBody.WriteSqlcHeader(utils.ToPtr(queries[0]))
	fileBody.WriteQueryFileModuleDocstring(queries[0].FileName)
	fileBody.WriteFutureImport()

	tablesBody := r.getCodeWriter()
	constantsBody := r.getCodeWriter()
	functionsBody := r.getCodeWriter()

	all := make([]string, 0, len(queries)*2) //nolint:mnd

	indentLevel := 0
	newLines := 2

	if isAnyQueryMany(queries) {
		r.driver.WriteQueryResultsClass(functionsBody)
		all = append(all, "QueryResults")
		// In classes mode the Querier class follows directly; in functions
		// mode each query already writes its own two leading blank lines.
		if r.config.EmitClasses {
			functionsBody.NNewLine(2)
		}
	}

	if r.config.EmitClasses {
		newLines = 1
		indentLevel = 1
		className := model.SnakeToCamel(r.config, moduleName)
		functionsBody.WriteLine(fmt.Sprintf("class %s:", className))
		functionsBody.WriteQueryClassDocstring(queries[0].FileName, r.driver.ConnType())
		functionsBody.WriteIndentedLine(1, `__slots__ = ("_conn",)`)
		functionsBody.NewLine()
		functionsBody.WriteIndentedLine(1, fmt.Sprintf(`def __init__(self, conn: %s) -> None:`, r.driver.ConnType()))
		functionsBody.WriteQueryClassInitDocstring(2, r.driver.ConnType()) //nolint:mnd
		functionsBody.WriteIndentedLine(2, "self._conn = conn")
		functionsBody.NewLine()
		functionsBody.WriteIndentedLine(1, "@property")
		functionsBody.WriteIndentedLine(1, fmt.Sprintf(`def conn(self) -> %s:`, r.driver.ConnType()))
		functionsBody.WriteQueryClassConnDocstring(r.driver.ConnType())
		functionsBody.WriteIndentedLine(2, `return self._conn`)
		all = append(all, className)
	}

	for _, query := range queries {
		constantsBody.WriteLine(fmt.Sprintf(`%s: typing.Final[str] = """-- name: %s %s`, query.ConstantName, query.QueryName, query.Cmd))
		constantsBody.WriteLine(query.SQL)
		constantsBody.WriteLine(`"""`)
		constantsBody.NewLine()

		if query.Returns.EmitTable {
			r.renderTable(tablesBody, *query.Returns.Table)
			all = append(all, query.Returns.Table.Name)
			tablesBody.NNewLine(2)
		}

		for _, param := range query.Params {
			if !param.EmitTable {
				continue
			}
			r.renderTable(tablesBody, *param.Table)
			tablesBody.NNewLine(2)
			all = append(all, param.Table.Name)
		}

		if !r.config.EmitClasses {
			all = append(all, query.FuncName)
		}
		functionsBody.NNewLine(newLines)
		r.driver.WriteQueryFunc(functionsBody, r.config, query, indentLevel)
	}
	fileBody.WriteAll(all)
	fileBody.NewLine()
	r.importResolver.QueryImports(queries, hasTables, hasEnums).Write(fileBody, r.config.OmitTypecheckingBlock, r.driver.TypeCheckingHook())
	fileBody.NNewLine(2)
	conversionBody := r.getCodeWriter()
	if r.driver.WriteConversionSetup(conversionBody, r.config, queries) {
		fileBody.WriteString(conversionBody.String())
		fileBody.NNewLine(2)
	}
	fileBody.WriteString(tablesBody.String())
	fileBody.WriteString(constantsBody.String())
	fileBody.NewLine()
	fileBody.WriteString(strings.TrimLeft(functionsBody.String(), "\n"))

	return &plugin.File{
		Name:     moduleName + ".py",
		Contents: fileBody.Bytes(),
	}
}

func isAnyQueryMany(queries []model.Query) bool {
	for _, query := range queries {
		if query.Cmd == metadata.CmdMany {
			return true
		}
	}
	return false
}
