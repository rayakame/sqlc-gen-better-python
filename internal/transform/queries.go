package transform

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func (t *Transformer) BuildQueries(tables []model.Table) []model.Query {
	queries := make([]model.Query, 0, len(t.req.Queries))
	for _, pluginQuery := range t.req.Queries {
		if pluginQuery.Name == "" {
			continue
		}
		if pluginQuery.Cmd == "" {
			continue
		}

		constantName := model.UpperSnakeCase(pluginQuery.Name)

		moduleName := pluginQuery.Filename
		lastDot := strings.LastIndex(moduleName, ".")
		if lastDot != -1 {
			moduleName = moduleName[:lastDot]
		}

		query := model.Query{
			Cmd:          pluginQuery.Cmd,
			SQL:          pluginQuery.Text,
			ConstantName: constantName,
			FuncName:     strings.ToLower(constantName),
			QueryName:    pluginQuery.Name,
			Params:       make([]model.QueryValue, 0),
			Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			FileName:     pluginQuery.Filename,
			ModuleName:   moduleName,
			Table:        pluginQuery.InsertIntoTable,
		}

		if query.Cmd == metadata.CmdCopyFrom || t.config.IsOverQueryParameterLimit(len(pluginQuery.Params)) {
			columns := make([]pyColumn, 0, len(pluginQuery.Params))
			for _, param := range pluginQuery.Params {
				columns = append(columns, pyColumn{
					column: param.Column,
					embed:  nil,
				})
			}
			table := t.columnsToClass(query.QueryName+"Params", columns)
			query.Params = []model.QueryValue{
				{
					Table:     utils.ToPtr(table),
					Name:      "params",
					Type:      model.PyType{Type: table.Name, IsList: query.Cmd == metadata.CmdCopyFrom},
					EmitTable: true,
				},
			}
		} else {
			query.Params = make([]model.QueryValue, 0, len(pluginQuery.Params))
			seen := make(map[string]int, len(pluginQuery.Params))
			for _, param := range pluginQuery.Params {
				query.Params = append(query.Params, model.QueryValue{
					Name: dedupName(model.ParamName(param), seen),
					Type: t.buildPyType(param.Column),
				})
			}
		}

		if query.Cmd == metadata.CmdExecLastId {
			query.Returns.Type = model.PyType{Type: "int", IsNullable: true}
		}
		if query.Cmd == metadata.CmdExecRows || query.Cmd == metadata.CmdCopyFrom {
			query.Returns.Type = model.PyType{Type: "int"}
		}

		if pluginQuery.Cmd != metadata.CmdOne && pluginQuery.Cmd != metadata.CmdMany {
			queries = append(queries, query)

			continue
		}

		if len(pluginQuery.Columns) == 1 && pluginQuery.Columns[0].EmbedTable == nil {
			column := pluginQuery.Columns[0]
			query.Returns = model.QueryValue{Type: t.buildPyType(column)}
			queries = append(queries, query)

			continue
		}

		// Precompute the query's column names/types once — they do not depend
		// on the candidate table — instead of rebuilding them per candidate.
		queryColumnNames := make([]string, len(pluginQuery.Columns))
		queryColumnTypes := make([]string, len(pluginQuery.Columns))
		for i, column := range pluginQuery.Columns {
			queryColumnNames[i] = model.EscapedColumnName(column, i)
			queryColumnTypes[i] = t.buildPyType(column).Type
		}

		var tableFound bool
		for _, table := range tables {
			if len(table.Columns) != len(pluginQuery.Columns) {
				continue
			}
			// A table only matches when EVERY column matches by name, type,
			// and source table — otherwise a dedicated Row class is needed.
			same := true
			for i, tableColumn := range table.Columns {
				queryColumn := pluginQuery.Columns[i]

				sameName := tableColumn.Name == queryColumnNames[i]
				sameType := tableColumn.Type.Type == queryColumnTypes[i]
				sameTable := utils.SameTableName(queryColumn.Table, table.Identifier, t.req.Catalog.DefaultSchema)
				if !sameName || !sameType || !sameTable {
					same = false

					break
				}
			}
			if same {
				query.Returns = model.QueryValue{
					Table: utils.ToPtr(table),
					Type:  model.PyType{Type: "models." + table.Name},
				}
				tableFound = true
				break
			}
		}

		if !tableFound {
			columns := make([]pyColumn, 0, len(pluginQuery.Columns))
			for _, column := range pluginQuery.Columns {
				columns = append(columns, pyColumn{
					column: column,
					embed:  t.newGoEmbed(column.EmbedTable, tables),
				})
			}
			returnTable := t.columnsToClass(query.QueryName+"Row", columns)
			query.Returns = model.QueryValue{
				Table:     utils.ToPtr(returnTable),
				Type:      model.PyType{Type: returnTable.Name},
				EmitTable: true,
			}
		}

		queries = append(queries, query)
	}

	return queries
}

type pyColumn struct {
	column *plugin.Column
	embed  *model.Embed
}

// dedupName makes repeated Python identifiers unique by appending a numeric
// suffix ("name", "name_2", "name_3", ...), so duplicate columns or
// parameters never generate duplicate fields or arguments.
func dedupName(name string, seen map[string]int) string {
	seen[name]++
	if seen[name] > 1 {
		name = fmt.Sprintf("%s_%d", name, seen[name])
		// Reserve the suffixed name so a literal collision later gets its own suffix.
		seen[name]++
	}

	return name
}

func (t *Transformer) columnsToClass(name string, columns []pyColumn) model.Table {
	table := model.Table{
		Name:       name,
		Columns:    make([]model.Column, 0, len(columns)),
		Identifier: utils.ToPtr(plugin.Identifier{}),
	}
	seen := make(map[string]int, len(columns))
	for i, column := range columns {
		columnName := model.EscapedColumnName(column.column, i)
		if column.embed != nil {
			// Embed fields are named after their table; use the singular
			// form ("test_inner_postgres_type"), matching the model naming.
			columnName = model.Singular(model.SingularParams{
				Name:       columnName,
				Exclusions: t.config.InflectionExcludeTableNames,
			})
		}
		tableColumn := model.Column{
			Name:   dedupName(columnName, seen),
			DBName: model.ColumnName(column.column, i),
		}

		if column.embed == nil {
			tableColumn.Type = t.buildPyType(column.column)
		} else {
			tableColumn.Embed = column.embed
			tableColumn.Type.Type = "models." + column.embed.ModelName
		}

		table.Columns = append(table.Columns, tableColumn)
	}
	return table
}

func (t *Transformer) newGoEmbed(embedTable *plugin.Identifier, tables []model.Table) *model.Embed {
	if embedTable == nil {
		return nil
	}

	for _, table := range tables {
		if !utils.SameTableName(embedTable, table.Identifier, t.req.Catalog.DefaultSchema) {
			continue
		}
		columns := make([]model.Column, len(table.Columns))
		for i, column := range table.Columns {
			columns[i] = column
		}
		return utils.ToPtr(model.Embed{ModelName: table.Name, Columns: columns})
	}
	return nil
}
