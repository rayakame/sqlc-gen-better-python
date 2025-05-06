package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"sort"
	"strings"
)

func (dr *Driver) prepareFunctionHeader(query *core.Query, body *builders.IndentStringBuilder) ([]string, string, []string) {
	pyTableNames := make([]string, 0)
	args := make([]string, 0)
	for _, arg := range query.Args {
		if !arg.IsEmpty() {
			argType := arg.Typ.Type
			if arg.Typ.IsList {
				argType = fmt.Sprintf("typing.Sequence[%s]", argType)
			}
			args = append(args, fmt.Sprintf("%s: %s", arg.Name, argType))
		}
	}
	retType := "None"
	if query.Ret.EmitStruct() && query.Ret.IsStruct() {
		BuildPyTabel(dr.conf.ModelType, query.Ret.Table, body)
		body.WriteString("\n\n")
		retType = query.Ret.Table.Name
		pyTableNames = append(pyTableNames, query.Ret.Table.Name)
	} else if !query.Ret.IsEmpty() {
		if query.Ret.IsStruct() {
			retType = fmt.Sprintf("models.%s", query.Ret.Table.Name)
		} else {
			retType = query.Ret.Typ.Type
		}
	}
	if query.Cmd == metadata.CmdExecLastId || query.Cmd == metadata.CmdExecRows {
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
	body.WriteIndentedLine(1, `__slots__ = ("_conn",)`)
	body.NewLine()
	body.WriteIndentedLine(1, fmt.Sprintf(`def __init__(self, conn: %s) -> None:`, dr.connType))
	body.WriteIndentedLine(2, "self._conn = conn")
	body.NewLine()
	return className
}

func (dr *Driver) buildPyQueriesFile(imp *core.Importer, queries []core.Query, sourceName string) ([]byte, error) {
	body := builders.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	body.WriteSqlcHeader()
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
	funcBody.NewLine()
	if dr.conf.EmitClasses {
		allNames = append(allNames, dr.buildClassTemplate(sourceName, funcBody))
	}
	for i, query := range queries {
		args, retType, addedPyTableNames := dr.prepareFunctionHeader(&query, pyTableBody)
		allNames = append(allNames, addedPyTableNames...)
		err := dr.buildPyQueryFunc(&query, funcBody, args, retType, dr.conf.EmitClasses)
		if err != nil {
			return nil, err
		}
		if i != len(queries)-1 {
			funcBody.NNewLine(newLines)
		}
	}
	body.WriteLine("__all__: typing.Sequence[str] = (")
	if len(allNames) > 0 {
		sort.Slice(allNames, func(i, j int) bool { return allNames[i] < allNames[j] })
	}
	for _, n := range allNames {
		body.WriteIndentedLine(1, fmt.Sprintf("\"%s\",", n))
	}
	body.WriteLine(")")
	body.NewLine()
	for _, imp := range imp.Imports(sourceName) {
		body.WriteLine(imp)
	}
	body.NNewLine(2)
	return []byte(body.String() + pyTableBody.String() + funcBody.String()), nil
}
