package transform

import (
	"slices"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// plainParams builds the expanded parameter list of a query, keeping names
// clear of the implicit first argument and of the locals generated bodies
// introduce.
func (t *Transformer) plainParams(pluginQuery *plugin.Query, pluginParams []*plugin.Parameter) []model.QueryValue {
	params := make([]model.QueryValue, 0, len(pluginParams))
	seen := make(map[string]int, len(pluginParams)+1)
	// The implicit first argument of every generated function must never
	// collide with a parameter name: a column literally named "conn" (or
	// "self" in classes mode) would otherwise produce a duplicate argument
	// and a SyntaxError in the generated module.
	if t.config.EmitClasses {
		seen["self"]++
	} else {
		seen["conn"]++
	}
	// Slice queries materialize the expanded SQL into a local named "sql"
	// before any parameter is read; a parameter with that name would be
	// silently overwritten by the query text.
	for _, param := range pluginParams {
		if param.GetColumn().GetIsSqlcSlice() {
			seen["sql"]++

			break
		}
	}
	// psycopg query bodies introduce locals a parameter must not clobber: an
	// overlong params dict hoists into "sql_params", :execrows reads the
	// cursor via "cur", :one fetches into "row", and :many may define a
	// nested "_decode_hook".
	if t.config.SqlDriver.IsPsycopg() {
		seen["sql_params"]++
		switch pluginQuery.Cmd {
		case metadata.CmdExecRows:
			seen["cur"]++
		case metadata.CmdOne:
			seen["row"]++
		case metadata.CmdMany:
			seen["_decode_hook"]++
		}
	}
	// MySQL bodies are cursor-based: every command except :many opens "cur",
	// :one fetches into "row", and :many may define a nested "_decode_hook".
	if t.config.SqlDriver.IsMysql() {
		switch pluginQuery.Cmd {
		case metadata.CmdMany:
			seen["_decode_hook"]++
		case metadata.CmdOne:
			seen["cur"]++
			seen["row"]++
		default:
			seen["cur"]++
		}
	}
	// sqlc's MySQL engine emits one parameter per occurrence of a reused
	// named argument too (its dialect binds every use site separately).
	// Same-named NAMED parameters are one logical argument - sqlc's own
	// MySQL codegen merges them - so repeats keep their positional binding
	// slot but drop out of the signature. A use site that rejects NULL
	// makes the merged parameter non-optional. Bare "?" parameters that
	// merely inherit the same column name stay distinct (IsNamedParam is
	// false for them; sqlc generates name/name_2 arguments).
	firstIdx := make(map[string]int, len(pluginParams))
	for _, param := range pluginParams {
		typ := t.buildPyType(param.Column)
		rawName := model.ParamName(param)
		if t.config.SqlDriver.IsMysql() && typ.SqlcSliceName == "" && param.GetColumn().GetIsNamedParam() {
			if idx, found := firstIdx[rawName]; found {
				if !typ.IsNullable {
					params[idx].Type.IsNullable = false
				}
				params = append(params, model.QueryValue{
					Name:     params[idx].Name,
					Type:     params[idx].Type,
					Number:   param.Number,
					Repeated: true,
				})

				continue
			}
			firstIdx[rawName] = len(params)
		}
		params = append(params, model.QueryValue{
			Name:   model.DedupName(rawName, seen),
			Type:   typ,
			Number: param.Number,
		})
	}
	// A later occurrence can have tightened the first one's nullability
	// after repeats were copied; conversions must use one type everywhere.
	firstByName := make(map[string]model.PyType, len(firstIdx))
	for _, idx := range firstIdx {
		firstByName[params[idx].Name] = params[idx].Type
	}
	for i := range params {
		if params[i].Repeated {
			params[i].Type = firstByName[params[i].Name]
		}
	}

	return params
}

// mergeRepeatedFields is plainParams' reused-NAMED-argument merge for a
// bundled Params class: sqlc's MySQL engine emits one parameter per use site,
// so without it one sqlc.arg(term) becomes two fields ("term" and "term_2")
// the caller has to fill identically. Later occurrences take the first
// field's name and type, keep their binding slot, and are skipped when the
// class is emitted. Bare "?" parameters that merely share a column name stay
// distinct, exactly as in plainParams.
func (t *Transformer) mergeRepeatedFields(table *model.Table, pluginParams []*plugin.Parameter) {
	if !t.config.SqlDriver.IsMysql() {
		return
	}
	firstIdx := make(map[string]int, len(pluginParams))
	repeats := make(map[int]int, len(pluginParams))
	for i, param := range pluginParams {
		if param.GetColumn().GetIsSqlcSlice() || !param.GetColumn().GetIsNamedParam() {
			continue
		}
		name := model.ParamName(param)
		idx, found := firstIdx[name]
		if !found {
			firstIdx[name] = i

			continue
		}
		// A use site that rejects NULL makes the merged field non-optional.
		if !table.Columns[i].Type.IsNullable {
			table.Columns[idx].Type.IsNullable = false
		}
		repeats[i] = idx
	}
	// Second pass: a later occurrence can have tightened the first one's
	// nullability, and conversions must use one type everywhere.
	for i, idx := range repeats {
		table.Columns[i].Name = table.Columns[idx].Name
		table.Columns[i].Type = table.Columns[idx].Type
		table.Columns[i].Repeated = true
	}
}

// logicalParamCount counts parameters as the generated signature will show
// them: MySQL's per-occurrence duplicates of one reused NAMED argument count
// once; bare "?" parameters count per occurrence even when they share a name.
func (t *Transformer) logicalParamCount(params []*plugin.Parameter) int {
	if !t.config.SqlDriver.IsMysql() {
		return len(params)
	}
	count := 0
	names := make(map[string]struct{}, len(params))
	for _, param := range params {
		if !param.GetColumn().GetIsNamedParam() {
			count++

			continue
		}
		name := model.ParamName(param)
		if _, found := names[name]; found {
			continue
		}
		names[name] = struct{}{}
		count++
	}

	return count
}

// hasSlice reports whether any parameter is a sqlc.slice, i.e. whether the
// query's bind order is ever consulted.
func hasSlice(params []*plugin.Parameter) bool {
	return slices.ContainsFunc(params, func(param *plugin.Parameter) bool {
		return param.GetColumn().GetIsSqlcSlice()
	})
}

// dedupSliceParams collapses the per-occurrence duplicates sqlc's MySQL
// engine emits for a reused sqlc.slice (one parameter per marker use site;
// the other engines merge them). Without the collapse both the plain
// signature and a bundled Params class would repeat the argument. The
// driver still binds one copy of the sequence per marker occurrence.
func (t *Transformer) dedupSliceParams(params []*plugin.Parameter) []*plugin.Parameter {
	if !t.config.SqlDriver.IsMysql() {
		return params
	}
	seen := make(map[string]struct{})
	out := make([]*plugin.Parameter, 0, len(params))
	for _, param := range params {
		if param.GetColumn().GetIsSqlcSlice() {
			name := param.GetColumn().GetName()
			if _, found := seen[name]; found {
				continue
			}
			seen[name] = struct{}{}
		}
		out = append(out, param)
	}

	return out
}

// bundledParams builds the single Params-class parameter used by :copyfrom
// and query_parameter_limit queries. Field order follows the sqlc parameter
// array, and each field keeps its parameter number for name-binding drivers.
func (t *Transformer) bundledParams(pluginParams []*plugin.Parameter, queryName string, isCopyFrom bool) []model.QueryValue {
	columns := make([]pyColumn, 0, len(pluginParams))
	for _, param := range pluginParams {
		columns = append(columns, pyColumn{
			column: param.Column,
			embed:  nil,
		})
	}
	table := t.columnsToClass(queryName+"Params", columns)
	for i := range table.Columns {
		table.Columns[i].Number = pluginParams[i].Number
	}
	t.mergeRepeatedFields(&table, pluginParams)

	return []model.QueryValue{
		{
			Table:     utils.ToPtr(table),
			Name:      "params",
			Type:      model.PyType{Type: table.Name, IsList: isCopyFrom},
			EmitTable: true,
		},
	}
}

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
			Placeholders: nil,
		}
		// psycopg rejects PostgreSQL-native $N placeholders and scans the
		// whole text for pyformat ones as soon as parameters are passed, so
		// parameterized queries are rewritten once here. :copyfrom never
		// executes its SQL - the driver builds a COPY statement instead.
		if t.config.SqlDriver.IsPsycopg() &&
			len(pluginQuery.Params) > 0 && query.Cmd != metadata.CmdCopyFrom {
			query.SQL = rewritePsycopgSQL(pluginQuery.Text)
		}
		// The MySQL drivers interpolate pyformat placeholders the same way,
		// but sqlc's MySQL engine emits "?": rewrite parameterized queries to
		// "%s" once here. Parameterless queries stay untouched EXCEPT :many:
		// QueryResults always passes its (possibly empty) args tuple, and the
		// drivers interpolate whenever args is not None, so a :many query
		// needs its "%" doubled even without parameters.
		var slots []model.Placeholder
		if t.config.SqlDriver.IsMysql() &&
			(len(pluginQuery.Params) > 0 || query.Cmd == metadata.CmdMany) {
			query.SQL, slots = rewriteMySQLSQL(pluginQuery.Text)
		}
		// Only the slice expansion reads the bind order, and only the
		// sqlite-family and MySQL drivers expand slices; postgres passes
		// sequences straight to = ANY($1).
		if hasSlice(pluginQuery.Params) {
			if !t.config.SqlDriver.IsMysql() {
				slots = sqlitePlaceholders(query.SQL)
			}
			query.Placeholders = slots
		}

		// Dedup before the limit check: a reused sqlc.slice or named
		// argument is one logical parameter and must not push a query into
		// bundled mode by itself.
		pluginParams := t.dedupSliceParams(pluginQuery.Params)
		if query.Cmd == metadata.CmdCopyFrom || t.config.IsOverQueryParameterLimit(t.logicalParamCount(pluginParams)) {
			query.Params = t.bundledParams(pluginParams, query.QueryName, query.Cmd == metadata.CmdCopyFrom)
		} else {
			query.Params = t.plainParams(pluginQuery, pluginParams)
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
