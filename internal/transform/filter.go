package transform

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
)

// FilterUnusedModels removes tables and enums that are not referenced by any
// query (as argument, return value, row/params class column, or embed).
// Used when the omit_unused_models option is enabled.
func FilterUnusedModels(enums []model.Enum, tables []model.Table, queries []model.Query) ([]model.Enum, []model.Table) {
	keep := make(map[string]struct{})
	addType := func(typeName string) {
		typeName = strings.TrimPrefix(typeName, "models.")
		typeName = strings.TrimPrefix(typeName, "enums.")
		keep[typeName] = struct{}{}
	}
	addPyType := func(typ model.PyType) {
		addType(typ.Type)
		// An overridden enum column still calls the enum class at runtime:
		// parameters convert back through their DefaultType.
		if typ.DoOverride() {
			addType(typ.DefaultType)
		}
	}
	collect := func(qv model.QueryValue) {
		if qv.IsEmpty() {
			return
		}
		addPyType(qv.Type)
		if qv.Table == nil {
			return
		}
		for _, col := range qv.Table.Columns {
			if col.Embed != nil {
				addType(col.Embed.ModelName)
				for _, embedCol := range col.Embed.Columns {
					addPyType(embedCol.Type)
				}

				continue
			}
			addPyType(col.Type)
		}
	}
	for _, query := range queries {
		collect(query.Returns)
		for _, param := range query.Params {
			collect(param)
		}
	}

	keptEnums := make([]model.Enum, 0, len(enums))
	for _, enum := range enums {
		if _, ok := keep[enum.Name]; ok {
			keptEnums = append(keptEnums, enum)
		}
	}

	keptTables := make([]model.Table, 0, len(tables))
	for _, table := range tables {
		if _, ok := keep[table.Name]; ok {
			keptTables = append(keptTables, table)
		}
	}

	return keptEnums, keptTables
}
