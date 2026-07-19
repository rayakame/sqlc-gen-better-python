package render

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/driver"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Renderer struct {
	config         *config.Config
	driver         driver.Driver
	importResolver *ImportResolver
}

func New(cfg *config.Config, drv driver.Driver) *Renderer {
	return &Renderer{
		config:         cfg,
		driver:         drv,
		importResolver: NewImportResolver(cfg, drv),
	}
}

func (r *Renderer) RenderAll(enums []model.Enum, tables []model.Table, queries []model.Query) ([]*plugin.File, error) {
	outputFiles := make([]*plugin.File, 0)
	if len(enums) > 0 {
		outputFiles = append(outputFiles, r.renderEnums(enums))
	}
	// models.py is emitted even when every table was filtered out: users
	// import it directly, and v0.4.x always shipped it.
	outputFiles = append(outputFiles, r.renderTables(tables))

	queriesModuleMap := make(map[string][]model.Query)
	for _, query := range queries {
		if !r.driver.SupportsCommand(query.Cmd) {
			return nil, fmt.Errorf(`unsupported cmd "%s" for driver "%s"`, query.Cmd, r.driver.Name())
		}

		innerQueries, ok := queriesModuleMap[query.ModuleName]
		if ok {
			queriesModuleMap[query.ModuleName] = append(innerQueries, query)
		} else {
			queriesModuleMap[query.ModuleName] = []model.Query{query}
		}
	}

	for module, innerQueries := range queriesModuleMap {
		outputFiles = append(outputFiles, r.renderQueriesModule(module, innerQueries))
	}

	if r.config.EmitInitFile != nil && *r.config.EmitInitFile {
		outputFiles = append(outputFiles, r.renderInitFile())
	}

	return outputFiles, nil
}

// renderInitFile renders the package __init__.py.
func (r *Renderer) renderInitFile() *plugin.File {
	fileBody := r.getCodeWriter()
	fileBody.WriteSqlcHeader(nil)
	fileBody.WriteInitFileModuleDocstring()

	return &plugin.File{
		Name:     "__init__.py",
		Contents: fileBody.Bytes(),
	}
}

func (r *Renderer) getCodeWriter() *writer.CodeWriter {
	return writer.NewCodeWriter(r.config)
}
