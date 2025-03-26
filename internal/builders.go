package internal

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/rayakame/sqlc-gen-better-python/internal/inflection"
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
	for _, column := range table.Columns {
		t.Columns = append(t.Columns, core.Column{
			Name:    core.SnakeToCamel(column.Name, gen.config),
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
	strType := gen.typeConversionFunc(gen.req, col, gen.config)
	return core.PyType{
		SqlType:    sdk.DataType(col.Type),
		Type:       strType,
		IsNullable: !col.NotNull,
		IsList:     col.IsArray,
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

func (gen *PythonGenerator) buildQueries(tables []core.Table) ([]core.Query, error) {
	qs := make([]core.Query, 0, len(gen.req.Queries))
	for _, query := range gen.req.Queries {
		if query.Name == "" {
			continue
		}
		if query.Cmd == "" {
			continue
		}

		constantName := sdk.LowerTitle(query.Name)

		comments := query.Comments

		gq := core.Query{
			Cmd:          query.Cmd,
			ConstantName: constantName,
			FieldName:    sdk.LowerTitle(query.Name) + "Stmt",
			MethodName:   query.Name,
			SourceName:   query.Filename,
			SQL:          query.Text,
			Comments:     comments,
			Table:        query.InsertIntoTable,
		}

		qpl := int(*gen.config.QueryParameterLimit)

		if len(query.Params) == 1 && qpl != 0 {
			p := query.Params[0]
			gq.Arg = core.QueryValue{
				Name:      core.Escape(core.ParamName(p)),
				DBName:    p.Column.GetName(),
				Typ:       gen.makePythonType(p.Column),
				SQLDriver: gen.config.SqlDriver,
				Column:    p.Column,
			}
		} else if len(query.Params) >= 1 {
			var cols []goColumn
			for _, p := range query.Params {
				cols = append(cols, goColumn{
					id:     int(p.Number),
					Column: p.Column,
				})
			}
			s, err := columnsToStruct(req, options, gq.MethodName+"Params", cols, false)
			if err != nil {
				return nil, err
			}
			gq.Arg = QueryValue{
				Emit:        true,
				Name:        "arg",
				Struct:      s,
				SQLDriver:   sqlpkg,
				EmitPointer: options.EmitParamsStructPointers,
			}

			// if query params is 2, and query params limit is 4 AND this is a copyfrom, we still want to emit the query's model
			// otherwise we end up with a copyfrom using a struct without the struct definition
			if len(query.Params) <= qpl && query.Cmd != ":copyfrom" {
				gq.Arg.Emit = false
			}
		}

		if len(query.Columns) == 1 && query.Columns[0].EmbedTable == nil {
			c := query.Columns[0]
			name := columnName(c, 0)
			name = strings.Replace(name, "$", "_", -1)
			gq.Ret = QueryValue{
				Name:      escape(name),
				DBName:    name,
				Typ:       goType(req, options, c),
				SQLDriver: sqlpkg,
			}
		} else if putOutColumns(query) {
			var gs *Struct
			var emit bool

			for _, s := range structs {
				if len(s.Fields) != len(query.Columns) {
					continue
				}
				same := true
				for i, f := range s.Fields {
					c := query.Columns[i]
					sameName := f.Name == StructName(columnName(c, i), options)
					sameType := f.Type == goType(req, options, c)
					sameTable := sdk.SameTableName(c.Table, s.Table, req.Catalog.DefaultSchema)
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
						embed:  newGoEmbed(c.EmbedTable, structs, req.Catalog.DefaultSchema),
					})
				}
				var err error
				gs, err = columnsToStruct(req, options, gq.MethodName+"Row", columns, true)
				if err != nil {
					return nil, err
				}
				emit = true
			}
			gq.Ret = QueryValue{
				Emit:        emit,
				Name:        "i",
				Struct:      gs,
				SQLDriver:   sqlpkg,
				EmitPointer: options.EmitResultStructPointers,
			}
		}

		qs = append(qs, gq)
	}
	sort.Slice(qs, func(i, j int) bool { return qs[i].MethodName < qs[j].MethodName })
	return qs, nil
}
