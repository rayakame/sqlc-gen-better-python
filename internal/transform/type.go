package transform

import (
	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func (t *Transformer) convertType(columnType *plugin.Identifier) string {
	return t.typeConversionFunc(t.req, t.config, columnType)
}

func (t *Transformer) buildPyType(pluginColumn *plugin.Column) model.PyType {
	columnType := sdk.DataType(pluginColumn.Type)
	strType := t.convertType(pluginColumn.Type)

	isEnum := false

	// Never mutate pluginColumn: buildPyType runs repeatedly on the same
	// shared columns (e.g. during table matching), and writing the default
	// schema back would change sdk.DataType results on later calls.
	typeSchema := pluginColumn.Type.Schema
	if typeSchema == "" {
		typeSchema = t.req.Catalog.DefaultSchema
	}

	for _, schema := range t.req.Catalog.Schemas {
		if schema.Name == utils.PgCatalog || schema.Name == utils.InformationSchema {
			continue
		}
		if typeSchema != schema.GetName() {
			continue
		}

		for _, enum := range schema.Enums {
			if pluginColumn.Type.Name == enum.Name {
				isEnum = true
			}
		}
	}

	if override := t.matchOverride(pluginColumn, columnType); override != nil {
		return model.PyType{
			SQLType:     columnType,
			Type:        override.PyType.Type,
			IsNullable:  !pluginColumn.GetNotNull(),
			IsList:      pluginColumn.GetIsArray() || pluginColumn.GetIsSqlcSlice(),
			IsEnum:      false,
			IsOverride:  true,
			DefaultType: strType,
		}
	}

	return model.PyType{
		SQLType:     columnType,
		Type:        strType,
		IsNullable:  !pluginColumn.GetNotNull(),
		IsList:      pluginColumn.GetIsArray() || pluginColumn.GetIsSqlcSlice(),
		IsEnum:      isEnum,
		IsOverride:  false,
		DefaultType: strType,
	}
}

// matchOverride returns the first configured override matching the column,
// either by column pattern or by exact SQL type.
func (t *Transformer) matchOverride(pluginColumn *plugin.Column, columnType string) *config.Override {
	for i := range t.config.Overrides {
		override := &t.config.Overrides[i]
		if override.PyType.Type == "" {
			continue
		}
		if override.Column != "" {
			columnName := pluginColumn.Name
			if pluginColumn.OriginalName != "" {
				columnName = pluginColumn.OriginalName
			}
			if override.ColumnName.MatchString(columnName) && override.Matches(pluginColumn.Table, t.req.Catalog.DefaultSchema) {
				return override
			}
		}
		if override.DBType != "" && override.DBType == columnType {
			return override
		}
	}

	return nil
}
