package driver

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/rayakame/sqlc-gen-better-python/internal/driver/aiosqlite"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type TypeBuildPyQueryFunc func(*core.Importer, *core.Query, *codegen.IndentStringBuilder) error
type TypeAcceptedDriverCMDs func() []string

type Driver struct {
	buildPyQueryFunc   TypeBuildPyQueryFunc
	acceptedDriverCMDs TypeAcceptedDriverCMDs

	//BuildPyQueriesFiles(*core.Importer, []core.Query) ([]*plugin.File, error)
}

func NewDriver(driver core.SQLDriverType) (*Driver, error) {
	var buildPyQueryFunc TypeBuildPyQueryFunc
	var acceptedDriverCMDs TypeAcceptedDriverCMDs
	switch driver {
	case core.SQLDriverAioSQLite:
		buildPyQueryFunc = aiosqlite.BuildPyQueryFunc
		acceptedDriverCMDs = aiosqlite.AcceptedDriverCMDs
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver.String())
	}

	return &Driver{buildPyQueryFunc: buildPyQueryFunc, acceptedDriverCMDs: acceptedDriverCMDs}, nil
}

func (dr *Driver) BuildPyTablesFile(imp *core.Importer, tables []core.Table) (*plugin.File, error) {
	fileName, fileContent, err := dr.buildPyTables(imp, tables)
	if err != nil {
		return nil, err
	}
	return &plugin.File{
		Name:     fmt.Sprintf("%s.py", fileName),
		Contents: fileContent,
	}, nil
}

func (dr *Driver) buildQueryHeader(query *core.Query, body *codegen.IndentStringBuilder) {
	body.WriteLine(fmt.Sprintf(`%s = """-- name: %s %s`, query.ConstantName, query.MethodName, query.Cmd))
	body.WriteLine(query.SQL)
	body.WriteLine(`"""`)
}

func (dr *Driver) BuildPyQueriesFile(imp *core.Importer, queries []core.Query, sourceName string) ([]byte, error) {
	body := codegen.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	body.WriteImportAnnotations()

	funcNames := make([]string, 0)
	queryBody := codegen.NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	for i, query := range queries {
		funcNames = append(funcNames, query.FuncName)
		if i != 0 {
			queryBody.WriteString("\n\n")
		}
		dr.buildQueryHeader(&query, queryBody)
		queryBody.WriteString("\n\n")
		err := dr.buildPyQueryFunc(imp, &query, queryBody)
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
		data, err := dr.BuildPyQueriesFile(imp, queries, sourceName)
		if err != nil {
			return nil, err
		}
		files = append(files, &plugin.File{
			Name:     fmt.Sprintf("%s.py", sourceName),
			Contents: data,
		})
	}

	return files, nil
}

func (dr *Driver) supportedCMD(command string) error {
	cmds := dr.acceptedDriverCMDs()
	for _, cmd := range cmds {
		if cmd == command {
			return nil
		}
	}
	return fmt.Errorf("unsupported command for selected driver: %s", command)
}
