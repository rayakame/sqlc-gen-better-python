package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/drivers"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"sort"
	"strings"
)

func (dr *Driver) prepareFunctionHeader(query *core.Query, body *builders.IndentStringBuilder) ([]core.FunctionArg, string, []string) {
	pyTableNames := make([]string, 0)
	args := make([]core.FunctionArg, 0)
	for _, arg := range query.Args {
		if !arg.IsEmpty() {
			argType := arg.Typ.Type
			if arg.EmitStruct() && arg.IsStruct() {
				BuildPyTabel(dr.conf.ModelType, arg.Table, body)
				body.NNewLine(2)
				pyTableNames = append(pyTableNames, arg.Table.Name)
				if query.Cmd == metadata.CmdCopyFrom {
					argType = fmt.Sprintf("collections.abc.Sequence[%s]", argType)
				}
				args = append(args, core.FunctionArg{
					Name:           arg.Name,
					Type:           argType,
					FunctionFormat: fmt.Sprintf("%s: %s", arg.Name, argType),
				})
			} else {
				if arg.Typ.IsList {
					argType = fmt.Sprintf("collections.abc.Sequence[%s]", argType)
				}
				if arg.Typ.IsNullable {
					argType = fmt.Sprintf("%s | None", argType)
				}
				args = append(args, core.FunctionArg{
					Name:           arg.Name,
					Type:           argType,
					FunctionFormat: fmt.Sprintf("%s: %s", arg.Name, argType),
				})
			}
		}
	}
	retType := "None"
	if query.Ret.EmitStruct() && query.Ret.IsStruct() {
		BuildPyTabel(dr.conf.ModelType, query.Ret.Table, body)
		body.NNewLine(2)
		retType = query.Ret.Table.Name
		pyTableNames = append(pyTableNames, query.Ret.Table.Name)
	} else if !query.Ret.IsEmpty() {
		if query.Ret.IsStruct() {
			retType = fmt.Sprintf("models.%s", query.Ret.Table.Name)
		} else {
			retType = query.Ret.Typ.Type
		}
	}
	if query.Cmd == metadata.CmdExecLastId {
		retType = "int | None"
	}
	if query.Cmd == metadata.CmdExecRows || query.Cmd == metadata.CmdCopyFrom {
		retType = "int"
	}
	return args, retType, pyTableNames
}

func (dr *Driver) BuildPyQueriesFiles(imp *core.Importer, queries []core.Query) ([]*plugin.File, error) {
	files := make([]*plugin.File, 0)
	fileQueries := make(map[string][]core.Query)
	for _, query := range queries {
		if err := dr.supportedCMD(query.Cmd); err != nil {
			return nil, err
		}
		if val, found := fileQueries[query.SourceName]; found {
			fileQueries[query.SourceName] = append(val, query)
		} else {
			fileQueries[query.SourceName] = []core.Query{query}
		}
	}

	for sourceName, queries := range fileQueries {
		data, err := dr.buildPyQueriesFile(imp, queries, sourceName)
		if err != nil {
			return nil, err
		}
		files = append(files, &plugin.File{
			Name:     core.SQLToPyFileName(sourceName),
			Contents: data,
		})
	}

	return files, nil
}

func (dr *Driver) buildQueryHeader(query *core.Query, body *builders.IndentStringBuilder) {
	body.WriteLine(fmt.Sprintf(`%s: typing.Final[str] = """-- name: %s %s`, query.ConstantName, query.MethodName, query.Cmd))
	body.WriteLine(query.SQL)
	body.WriteLine(`"""`)
}

func (dr *Driver) buildClassTemplate(sourceName string, body *builders.IndentStringBuilder) string {
	className := core.SnakeToCamel(strings.ReplaceAll(sourceName, ".sql", ""), dr.conf)
	body.WriteLine(fmt.Sprintf("class %s:", className))
	body.WriteQueryClassDocstring(sourceName, dr.connType)
	body.WriteIndentedLine(1, `__slots__ = ("_conn",)`)
	body.NewLine()
	body.WriteIndentedLine(1, fmt.Sprintf(`def __init__(self, conn: %s) -> None:`, dr.connType))
	body.WriteQueryClassInitDocstring(2, dr.connType)
	body.WriteIndentedLine(2, "self._conn = conn")
	body.NewLine()
	body.WriteIndentedLine(1, "@property")
	body.WriteIndentedLine(1, fmt.Sprintf(`def conn(self) -> %s:`, dr.connType))
	body.WriteQueryClassConnDocstring(dr.connType)
	body.WriteIndentedLine(2, `return self._conn`)
	body.NewLine()
	return className
}

func (dr *Driver) buildPyQueriesFile(imp *core.Importer, queries []core.Query, sourceName string) ([]byte, error) {
	body := builders.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	body.WriteQueryFileModuleDocstring(sourceName)
	body.WriteImportAnnotations()

	newLines := 2
	if dr.conf.EmitClasses {
		newLines = 1
	}

	allNames := make([]string, 0)
	funcBody := builders.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	pyTableBody := builders.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	for _, query := range queries {
		if !dr.conf.EmitClasses {
			allNames = append(allNames, query.FuncName)
		}
		dr.buildQueryHeader(&query, funcBody)
		funcBody.NewLine()
	}
	if core.IsAnyQueryMany(queries) {
		funcBody.NewLine()
		allNames = append(allNames, dr.driverBuildQueryResults(funcBody))
		funcBody.NewLine()
	}
	funcBody.NewLine()
	if dr.conf.EmitClasses {
		allNames = append(allNames, dr.buildClassTemplate(sourceName, funcBody))
	}
	for i, query := range queries {
		args, retType, addedPyTableNames := dr.prepareFunctionHeader(&query, pyTableBody)
		returnType := core.PyType{
			SqlType: query.Ret.Typ.SqlType,
			Type:    retType,
		}
		allNames = append(allNames, addedPyTableNames...)
		err := dr.buildPyQueryFunc(&query, funcBody, args, returnType, dr.conf.EmitClasses)
		if err != nil {
			return nil, err
		}
		if i != len(queries)-1 {
			funcBody.NNewLine(newLines)
		}
	}
	body.WriteLine("__all__: collections.abc.Sequence[str] = (")
	if len(allNames) > 0 {
		sort.Slice(allNames, func(i, j int) bool { return allNames[i] < allNames[j] })
	}
	for _, n := range allNames {
		body.WriteIndentedLine(1, fmt.Sprintf("\"%s\",", n))
	}
	body.WriteLine(")")
	body.NewLine()
	std, tye, pkg := imp.Imports(sourceName)
	tyeHook := dr.driverTypeCheckingHook()
	for _, imp := range std {
		body.WriteLine(imp)
	}
	if len(tye) != 0 || len(tyeHook) != 0 {
		if len(std) != 0 {
			body.NewLine()
		}
		body.WriteLine("if typing.TYPE_CHECKING:")
		for _, imp := range tye {
			body.WriteIndentedLine(1, imp)
		}
		for i, imp := range tyeHook {
			if i == 0 && len(tye) != 0 {
				body.NewLine()
			}
			body.WriteIndentedLine(1, imp)
		}
	}
	body.WriteLine("")
	for _, imp := range pkg {
		body.WriteLine(imp)
	}
	body.NNewLine(2)
	if dr.conf.SqlDriver == core.SQLDriverAioSQLite {
		drivers.AioSQLiteBuildTypeConvFunc(queries, body, dr.conf)
	}
	if dr.conf.SqlDriver == core.SQLDriverSQLite {
		drivers.SQLite3BuildTypeConvFunc(queries, body, dr.conf)
	}
	return []byte(body.String() + pyTableBody.String() + funcBody.String()), nil
}
