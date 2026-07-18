package transform

import (
	"cmp"
	"slices"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func (t *Transformer) BuildTables() []model.Table {
	tables := make([]model.Table, 0)
	for _, schema := range t.req.Catalog.Schemas {
		if schema.Name == utils.PgCatalog || schema.Name == utils.InformationSchema {
			continue
		}
		for _, table := range schema.Tables {
			tables = append(tables, t.buildTable(schema, table))
		}
	}
	slices.SortFunc(tables, func(a, b model.Table) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return tables
}

func (t *Transformer) buildTable(pluginSchema *plugin.Schema, pluginTable *plugin.Table) model.Table {
	var schemaName string
	if pluginSchema.Name != t.req.Catalog.DefaultSchema {
		schemaName = pluginSchema.Name
	}
	tableName := model.ModelName(t.config, pluginTable.Rel.Name, schemaName)
	table := model.Table{
		Name:    tableName,
		Columns: make([]model.Column, 0, len(pluginTable.Columns)),
		Identifier: utils.ToPtr(plugin.Identifier{
			Catalog: "",
			Schema:  pluginSchema.Name,
			Name:    pluginTable.Rel.Name,
		}),
	}
	for i, column := range pluginTable.Columns {
		table.Columns = append(table.Columns, model.Column{
			Name:   model.EscapedColumnName(column, i),
			DBName: model.ColumnName(column, i),
			Type:   t.buildPyType(column),
			Embed:  nil,
		})
	}

	return table
}
