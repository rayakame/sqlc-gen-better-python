package internal

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/rayakame/sqlc-gen-better-python/internal/inflection"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
	"sort"
	"strings"
)

func (gen *PythonGenerator) buildTable(schema *plugin.Schema, table *plugin.Table) core.Table {
	var tableName string
	if schema.Name == gen.req.Catalog.DefaultSchema {
		tableName = table.Rel.Name
	} else {
		tableName = schema.Name + "_" + table.Rel.Name
	}
	structName := tableName
	if !gen.config.EmitExactTableNames {
		structName = inflection.Singular(inflection.SingularParams{
			Name:       structName,
			Exclusions: gen.config.InflectionExcludeTableNames,
		})
	}
	t := core.Table{
		Table:   &plugin.Identifier{Schema: schema.Name, Name: table.Rel.Name},
		Name:    core.SnakeToCamel(structName, gen.config),
		Comment: table.Comment,
	}
	for i, column := range table.Columns {
		t.Columns = append(t.Columns, core.Column{
			Name:    core.ColumnName(column, i),
			Type:    gen.makePythonType(column),
			Comment: column.Comment,
		})
	}
	return t
}

func (gen *PythonGenerator) buildTables() []core.Table {
	tables := make([]core.Table, 0)
	for _, schema := range gen.req.Catalog.Schemas {
		if schema.Name == "pg_catalog" || schema.Name == "information_schema" {
			continue
		}
		for _, table := range schema.Tables {
			t := gen.buildTable(schema, table)
			tables = append(tables, t)
		}
	}
	if len(tables) > 0 {
		sort.Slice(tables, func(i, j int) bool { return tables[i].Name < tables[j].Name })
	}
	return tables
}

func (gen *PythonGenerator) makePythonType(col *plugin.Column) core.PyType {
	columnType := sdk.DataType(col.Type)
	for _, override := range gen.config.Overrides {
		if override.PyTypeName == "" {
			continue
		}
		cname := col.Name
		if col.OriginalName != "" {
			cname = col.OriginalName
		}
		sameTable := override.Matches(col.Table, gen.req.Catalog.DefaultSchema)
		if override.Column != "" && sdk.MatchString(override.Column, cname) && sameTable {
			return core.PyType{
				SqlType:    columnType,
				Type:       override.PyTypeName,
				IsNullable: !col.NotNull,
				IsList:     col.GetIsArray() || col.GetIsSqlcSlice(),
				IsEnum:     false,
			}
		}
		if override.DBType != "" && override.DBType == columnType {
			return core.PyType{
				SqlType:    columnType,
				Type:       override.PyTypeName,
				IsNullable: !col.NotNull,
				IsList:     col.GetIsArray() || col.GetIsSqlcSlice(),
				IsEnum:     false,
			}
		}
	}
	strType := gen.typeConversionFunc(gen.req, col, gen.config)
	return core.PyType{
		SqlType:    columnType,
		Type:       strType,
		IsNullable: !col.NotNull,
		IsList:     col.GetIsArray() || col.GetIsSqlcSlice(),
		IsEnum:     false,
	}
}

func (gen *PythonGenerator) buildEnums() []core.Enum {
	var enums []core.Enum
	for _, schema := range gen.req.Catalog.Schemas {
		if schema.Name == "pg_catalog" || schema.Name == "information_schema" {
			continue
		}
		for _, enum := range schema.Enums {
			var enumName string
			if schema.Name == gen.req.Catalog.DefaultSchema {
				enumName = enum.Name
			} else {
				enumName = schema.Name + "_" + enum.Name
			}

			e := core.Enum{
				Name:    core.SnakeToCamel(enumName, gen.config),
				Comment: enum.Comment,
			}

			seen := make(map[string]struct{}, len(enum.Vals))
			for i, v := range enum.Vals {
				value := core.EnumReplace(v)
				if _, found := seen[value]; found || value == "" {
					value = fmt.Sprintf("value_%d", i)
				}
				e.Constants = append(e.Constants, core.Constant{
					Name:  core.SnakeToCamel(enumName+"_"+value, gen.config),
					Value: v,
					Type:  e.Name,
				})
				seen[value] = struct{}{}
			}
			enums = append(enums, e)
		}
	}
	if len(enums) > 0 {
		sort.Slice(enums, func(i, j int) bool { return enums[i].Name < enums[j].Name })
	}
	return enums
}

type goColumn struct {
	id int
	*plugin.Column
	embed *goEmbed
}

type goEmbed struct {
	modelType string
	modelName string
	fields    []core.Column
}

var cmdReturnsData = map[string]struct{}{
	metadata.CmdBatchMany: {},
	metadata.CmdBatchOne:  {},
	metadata.CmdMany:      {},
	metadata.CmdOne:       {},
}

func putOutColumns(query *plugin.Query) bool {
	_, found := cmdReturnsData[query.Cmd]
	return found
}

// look through all the structs and attempt to find a matching one to embed
// We need the name of the struct and its field names.
func newGoEmbed(embed *plugin.Identifier, structs []core.Table, defaultSchema string) *goEmbed {
	if embed == nil {
		return nil
	}

	for _, s := range structs {
		embedSchema := defaultSchema
		if embed.Schema != "" {
			embedSchema = embed.Schema
		}

		// compare the other attributes
		if embed.Catalog != s.Table.Catalog || embed.Name != s.Table.Name || embedSchema != s.Table.Schema {
			continue
		}

		fields := make([]core.Column, len(s.Columns))
		for i, f := range s.Columns {
			fields[i] = f
		}
		return &goEmbed{
			modelType: s.Name,
			modelName: s.Name,
			fields:    fields,
		}
	}

	return nil
}

func (gen *PythonGenerator) buildQueries(tables []core.Table) ([]core.Query, error) {
	qs := make([]core.Query, 0, len(gen.req.Queries))
	for _, query := range gen.req.Queries {
		if query.Name == "" {
			continue
		}
		if query.Cmd == "" {
			continue
		}

		constantName := core.UpperSnakeCase(query.Name)

		comments := query.Comments

		gq := core.Query{
			Cmd:          query.Cmd,
			ConstantName: constantName,
			FuncName:     strings.ToLower(constantName),
			FieldName:    sdk.LowerTitle(query.Name) + "Stmt",
			MethodName:   query.Name,
			SourceName:   query.Filename,
			SQL:          query.Text,
			Comments:     comments,
			Table:        query.InsertIntoTable,
		}

		//qpl := int(*gen.config.QueryParameterLimit) TODO maybe?

		//if len(query.Params) == 1 && qpl != 0 {
		if query.Cmd == metadata.CmdCopyFrom {
			var cols []goColumn
			for _, p := range query.Params {
				cols = append(cols, goColumn{
					id:     int(p.Number),
					Column: p.Column,
				})
			}
			s, err := gen.columnsToStruct(gq.MethodName+"Params", cols, true)
			if err != nil {
				return nil, err
			}
			gq.Args = []core.QueryValue{{
				Emit:  true,
				Name:  "params",
				Table: s,
				Typ: core.PyType{
					Type: gq.MethodName + "Params",
				},
			}}
		} else {
			if len(query.Params) == 1 {
				p := query.Params[0]
				gq.Args = []core.QueryValue{{
					Name:   core.Escape(core.ParamName(p)),
					DBName: p.Column.GetName(),
					Typ:    gen.makePythonType(p.Column),
					Column: p.Column,
				}}
			} else if len(query.Params) >= 1 {
				var values []core.QueryValue
				for _, p := range query.Params {
					values = append(values, core.QueryValue{
						Name:   core.Escape(core.ParamName(p)),
						DBName: p.Column.GetName(),
						Typ:    gen.makePythonType(p.Column),
						Column: p.Column,
					})
				}
				gq.Args = values

				// if query params is 2, and query params limit is 4 AND this is a copyfrom, we still want to emit the query's model
				// otherwise we end up with a copyfrom using a struct without the struct definition
				//if len(query.Params) <= qpl && query.Cmd != ":copyfrom" {
				//	gq.Args.Emit = false
				//}
			}
		}

		if len(query.Columns) == 1 && query.Columns[0].EmbedTable == nil {
			c := query.Columns[0]
			name := core.ColumnName(c, 0)
			name = strings.Replace(name, "$", "_", -1)
			gq.Ret = core.QueryValue{
				Name:   core.Escape(name),
				DBName: name,
				Typ:    gen.makePythonType(c),
			}
		} else if putOutColumns(query) {
			var gs *core.Table
			var emit bool

			for _, s := range tables {
				if len(s.Columns) != len(query.Columns) {
					continue
				}
				same := true
				for i, f := range s.Columns {
					c := query.Columns[i]
					sameName := f.Name == core.ColumnName(c, i)
					sameType := f.Type == gen.makePythonType(c)
					sameTable := sdk.SameTableName(c.Table, s.Table, gen.req.Catalog.DefaultSchema)
					if !sameName || !sameType || !sameTable {
						same = false
					}
				}
				if same {
					gs = &s
					break
				}
			}

			if gs == nil {
				var columns []goColumn
				for i, c := range query.Columns {
					columns = append(columns, goColumn{
						id:     i,
						Column: c,
						embed:  newGoEmbed(c.EmbedTable, tables, gen.req.Catalog.DefaultSchema),
					})
				}
				var err error
				gs, err = gen.columnsToStruct(gq.MethodName+"Row", columns, true)
				if err != nil {
					return nil, err
				}
				emit = true
			}
			gq.Ret = core.QueryValue{
				Emit:  emit,
				Name:  "i",
				Table: gs,
			}
		}

		qs = append(qs, gq)
	}
	sort.Slice(qs, func(i, j int) bool { return qs[i].MethodName < qs[j].MethodName })
	return qs, nil
}

func (gen *PythonGenerator) columnsToStruct(name string, columns []goColumn, useID bool) (*core.Table, error) {
	gs := core.Table{
		Name: name,
	}
	seen := map[string][]int{}
	suffixes := map[int]int{}
	for i, c := range columns {
		colName := core.ColumnName(c.Column, i)

		// override col/tag with expected model name
		if c.embed != nil {
			colName = c.embed.modelName
		}

		fieldName := core.SnakeToCamel(colName, gen.config)
		baseFieldName := fieldName
		// Track suffixes by the ID of the column, so that columns referring to the same numbered parameter can be
		// reused.
		suffix := 0
		if o, ok := suffixes[c.id]; ok && useID {
			suffix = o
		} else if v := len(seen[fieldName]); v > 0 && !c.IsNamedParam {
			suffix = v + 1
		}
		suffixes[c.id] = suffix
		if suffix > 0 {
			fieldName = fmt.Sprintf("%s_%d", fieldName, suffix)
		}

		f := core.Column{
			Name: inflection.Singular(inflection.SingularParams{
				Name:       core.ColumnName(c.Column, i),
				Exclusions: gen.config.InflectionExcludeTableNames,
			}),
			DBName: colName,
			Column: c.Column,
		}

		if c.embed == nil {
			f.Type = gen.makePythonType(c.Column)
		} else {
			f.Type = core.PyType{
				SqlType:    c.embed.modelType,
				Type:       "models." + c.embed.modelType,
				IsList:     false,
				IsNullable: false,
				IsEnum:     false,
			}
			f.EmbedFields = c.embed.fields
		}

		gs.Columns = append(gs.Columns, f)
		if _, found := seen[baseFieldName]; !found {
			seen[baseFieldName] = []int{i}
		} else {
			seen[baseFieldName] = append(seen[baseFieldName], i)
		}
	}

	// If a field does not have a known type, but another
	// field with the same name has a known type, assign
	// the known type to the field without a known type
	/*for i, field := range gs.Columns {
		if len(seen[field.Name]) > 1 && field.Type.Type == "interface{}" {
			for _, j := range seen[field.Name] {
				if i == j {
					continue
				}
				otherField := gs.Fields[j]
				if otherField.Type != field.Type {
					field.Type = otherField.Type
				}
				gs.Fields[i] = field
			}
		}
	}*/

	err := checkIncompatibleFieldTypes(gs.Columns)
	if err != nil {
		return nil, err
	}

	return &gs, nil
}

func checkIncompatibleFieldTypes(fields []core.Column) error {
	fieldTypes := map[string]string{}
	for _, field := range fields {
		if fieldType, found := fieldTypes[field.Name]; !found {
			fieldTypes[field.Name] = field.Type.Type
		} else if field.Type.Type != fieldType {
			return fmt.Errorf("named param %s has incompatible types: %s, %s", field.Name, field.Type.Type, fieldType)
		}
	}
	return nil
}
