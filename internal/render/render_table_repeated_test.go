package render

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/driver"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
)

// A repeated MySQL parameter shares the first occurrence's field, so the
// bundled Params class must declare it once even though the driver still
// binds every occurrence. Only the MySQL transform produces such a table,
// so the renderer is driven directly.
func TestRenderTableSkipsRepeatedColumns(t *testing.T) {
	t.Parallel()
	conf := &config.Config{
		Package:             "db",
		SqlDriver:           config.SQLDriverPymysql,
		ModelType:           config.ModelTypeDataclass,
		EmitDocstrings:      config.DocstringConventionNone,
		EmitDocstringsSQL:   utils.ToPtr(true),
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
	}
	drv, err := driver.New(conf)
	if err != nil {
		t.Fatalf("driver.New() error: %v", err)
	}
	body := writer.NewCodeWriter(conf)
	New(conf, drv).renderTable(body, model.Table{
		Name: "RenameParams",
		Columns: []model.Column{
			{Name: "term", Type: model.PyType{Type: "str"}},
			{Name: "term", Type: model.PyType{Type: "str"}, Repeated: true},
			{Name: "id_", Type: model.PyType{Type: "int"}},
		},
	})

	want := "@dataclasses.dataclass()\nclass RenameParams:\n    term: str\n    id_: int\n"
	if got := body.String(); got != want {
		t.Errorf("renderTable() = %q, want %q", got, want)
	}
}
