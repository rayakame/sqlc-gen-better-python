package transform

import (
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

	if pluginColumn.Type.Schema == "" {
		pluginColumn.Type.Schema = t.req.Catalog.DefaultSchema
	}

	for _, schema := range t.req.Catalog.Schemas {
		if schema.Name == utils.PgCatalog || schema.Name == utils.InformationSchema {
			continue
		}
		if pluginColumn.Type.Schema != schema.GetName() {
			continue
		}

		for _, enum := range schema.Enums {
			if pluginColumn.Type.Name == enum.Name {
				isEnum = true
			}
		}
	}

	return model.PyType{
		SQLType:    columnType,
		Type:       strType,
		IsNullable: !pluginColumn.GetNotNull(),
		IsList:     pluginColumn.GetIsArray() || pluginColumn.GetIsSqlcSlice(),
		IsEnum:     isEnum,
	}
}
