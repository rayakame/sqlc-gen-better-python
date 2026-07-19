package transform

import (
	"slices"
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
			seen := make(map[string]int, len(pluginQuery.Params)+1)
			// The implicit first argument of every generated function must
			// never collide with a parameter name: a column literally named
			// "conn" (or "self" in classes mode) would otherwise produce a
			// duplicate argument and a SyntaxError in the generated module.
			if t.config.EmitClasses {
				seen["self"]++
			} else {
				seen["conn"]++
			}
			for _, param := range pluginQuery.Params {
				query.Params = append(query.Params, model.QueryValue{
					Name: model.DedupName(model.ParamName(param), seen),
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

		// Precompute the query's column names/types once - they do not depend
		// on the candidate table - instead of rebuilding them per candidate.
		// Dedup mirrors buildTable so colliding sanitized names still match.
		queryColumnNames := make([]string, len(pluginQuery.Columns))
		queryColumnTypes := make([]model.PyType, len(pluginQuery.Columns))
		seenColumns := make(map[string]int, len(pluginQuery.Columns))
		for i, column := range pluginQuery.Columns {
			queryColumnNames[i] = model.DedupName(model.EscapedColumnName(column, i), seenColumns)
			queryColumnTypes[i] = t.buildPyType(column)
		}

		var tableFound bool
		for _, table := range tables {
			if len(table.Columns) != len(pluginQuery.Columns) {
				continue
			}
			// A table only matches when EVERY column matches by name, type,
			// and source table - otherwise a dedicated Row class is needed.
			same := true
			for i, tableColumn := range table.Columns {
				queryColumn := pluginQuery.Columns[i]

				sameName := tableColumn.Name == queryColumnNames[i]
				// Compare the full type semantics, not just the type name: a
				// LEFT JOIN makes columns nullable, and reusing the non-null
				// model class would produce wrongly typed fields.
				sameType := tableColumn.Type.Type == queryColumnTypes[i].Type &&
					tableColumn.Type.IsNullable == queryColumnTypes[i].IsNullable &&
					tableColumn.Type.IsList == queryColumnTypes[i].IsList &&
					tableColumn.Type.IsEnum == queryColumnTypes[i].IsEnum
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
			Name:   model.DedupName(columnName, seen),
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

		return utils.ToPtr(model.Embed{ModelName: table.Name, Columns: slices.Clone(table.Columns)})
	}

	return nil
}
