package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func (dr *Driver) prepareFunctionHeader(query *core.Query, body *builders.IndentStringBuilder) (string, string) {
	argType := ""
	if query.Arg.EmitStruct() && query.Arg.IsStruct() {
		BuildPyTabel(dr.conf.ModelType, query.Arg.Table, body)
		body.WriteString("\n\n")
		argType = query.Arg.Table.Name
	} else if !query.Arg.IsEmpty() {
		argType = query.Arg.Typ.Type
		if query.Arg.Typ.IsList {
			argType = fmt.Sprintf("typing.Sequence[%s]", argType)
		}
	}
	retType := "None"
	if query.Ret.EmitStruct() && query.Ret.IsStruct() {
		BuildPyTabel(dr.conf.ModelType, query.Ret.Table, body)
		body.WriteString("\n\n")
		retType = query.Ret.Table.Name
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
	return argType, retType
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
	body.WriteLine(fmt.Sprintf(`%s = """-- name: %s %s`, query.ConstantName, query.MethodName, query.Cmd))
	body.WriteLine(query.SQL)
	body.WriteLine(`"""`)
}

func (dr *Driver) buildPyQueriesFile(imp *core.Importer, queries []core.Query, sourceName string) ([]byte, error) {
	body := builders.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	body.WriteImportAnnotations()

	funcNames := make([]string, 0)
	queryBody := builders.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	for i, query := range queries {
		funcNames = append(funcNames, query.FuncName)
		if i != 0 {
			queryBody.WriteString("\n\n")
		}
		dr.buildQueryHeader(&query, queryBody)
		queryBody.WriteString("\n\n")
		argType, retType := dr.prepareFunctionHeader(&query, queryBody)
		err := dr.buildPyQueryFunc(&query, queryBody, argType, retType)
		if err != nil {
			return nil, err
		}
	}
	body.WriteLine("__all__: typing.Sequence[str] = (")
	for _, n := range funcNames {
		body.WriteIndentedLine(1, fmt.Sprintf("\"%s\",", n))
	}
	body.WriteLine(")")
	body.WriteString("\n")
	for _, imp := range imp.Imports(sourceName) {
		body.WriteLine(imp)
	}
	body.WriteString("\n")

	return []byte(body.String() + queryBody.String()), nil
}
