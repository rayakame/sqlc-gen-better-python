package types

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/log"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

// Spellings shared with postgresql.go and sqlite.go, hoisted so goconst does
// not flag the third literal occurrence.
const (
	sqlInteger         = "integer"
	sqlSmallint        = "smallint"
	sqlBigint          = "bigint"
	sqlDoublePrecision = "double precision"
	sqlReal            = "real"
	sqlDecimal         = "decimal"
	sqlNumeric         = "numeric"
	sqlText            = "text"
	sqlJSON            = "json"
	sqlBlob            = "blob"
	sqlDate            = "date"
	pyDate             = "datetime.date"
	pyDatetime         = "datetime.datetime"
	enumsPrefix        = "enums."
)

func MysqlTypeToPython(req *plugin.GenerateRequest, config *config.Config, pluginColumn *plugin.Column) string {
	columnType := strings.ToLower(sdk.DataType(pluginColumn.Type))

	switch columnType {
	case "tinyint":
		// MySQL bool columns are tinyint(1); any other length is a plain int.
		if pluginColumn.GetLength() == 1 {
			return Bool
		}

		return Int
	case Bool, Boolean:
		return Bool
	case Int, sqlInteger, "mediumint", sqlSmallint, sqlBigint, "year", "serial", "bigint unsigned", "bigint signed":
		return Int
	case Float, "double", sqlDoublePrecision, sqlReal:
		return Float
	case sqlDecimal, "dec", "fixed", sqlNumeric:
		// Drivers return decimal.Decimal for all of these. Differs from
		// sqlite, where numeric maps to float.
		return Decimal
	case "varchar", "char", sqlText, "tinytext", "mediumtext", "longtext":
		return Str
	case "set":
		// PyMySQL-family drivers return SET as a comma-joined str.
		return Str
	case "enum":
		// Expression-derived enum columns lose their column identity and
		// arrive as the bare "enum" type. Column-typed enums arrive as
		// synthesized named types handled in the default branch.
		return Str
	case sqlBlob, "binary", "varbinary", "tinyblob", "mediumblob", "longblob":
		return Memoryview
	case "bit":
		// Drivers return BIT as raw bytes.
		return Memoryview
	case sqlDate:
		return pyDate
	case "datetime", "timestamp":
		return pyDatetime
	case "time":
		// PyMySQL-family drivers return TIME columns as timedelta, not
		// datetime.time.
		return "datetime.timedelta"
	case sqlJSON:
		return Str
	case "any":
		return Any
	default:
		// sqlc's dolphin engine materializes each MySQL enum column as a
		// catalog enum named "{table}_{column}" in the default schema
		// ("public"), so the same catalog scan as PostgreSQL resolves them.
		columnRelation, err := parseIdentifierString(columnType)
		if err != nil {
			log.L().LogErr("error trying to parse identifier string", err)

			return Any
		}
		if columnRelation.Schema == "" {
			columnRelation.Schema = req.Catalog.DefaultSchema
		}
		for _, schema := range req.Catalog.Schemas {
			if schema.Name == PgCatalog || schema.Name == InformationSchema {
				continue
			}
			if schema.Name != columnRelation.Schema {
				continue
			}
			for _, enum := range schema.Enums {
				if columnRelation.Name != enum.Name {
					continue
				}
				if schema.Name == req.Catalog.DefaultSchema {
					return enumsPrefix + model.EnumName(config, enum.Name, "")
				}

				return enumsPrefix + model.EnumName(config, enum.Name, schema.Name)
			}
		}
		log.L().Log("unknown MySQL type: " + columnType)

		return Any
	}
}
