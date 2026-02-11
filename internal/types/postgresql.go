package types

import (
	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/log"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func PostgresTypeToPython(req *plugin.GenerateRequest, config *config.Config, pluginType *plugin.Identifier) string {
	columnType := sdk.DataType(pluginType)
	switch columnType {
	case "serial",
		"serial4",
		"pg_catalog.serial4",
		"bigserial",
		"serial8",
		"pg_catalog.serial8",
		"smallserial",
		"serial2",
		"pg_catalog.serial2",
		"integer",
		Int,
		"int4",
		"pg_catalog.int4",
		"bigint",
		"int8",
		"pg_catalog.int8",
		"smallint",
		"int2",
		"pg_catalog.int2":
		return Int
	case Float, "double precision", "float8", "pg_catalog.float8", "real", "float4", "pg_catalog.float4":
		return Float
	case "numeric", "pg_catalog.numeric":
		return "decimal.Decimal"
	case "money":
		return Str
	case Boolean, Bool, "pg_catalog.bool":
		return Bool
	case "pg_catalog.json", "json", "jsonb":
		return Str
	case "bytea", "blob", "pg_catalog.bytea":
		return "memoryview"
	case "date":
		return "datetime.date"
	case "pg_catalog.time", "pg_catalog.timetz", "timetz":
		return "datetime.time"
	case "pg_catalog.timestamp", "pg_catalog.timestamptz", "timestamptz":
		return "datetime.datetime"
	case "interval", "pg_catalog.interval":
		return "datetime.timedelta"
	case "text", "pg_catalog.varchar", "bpchar", "pg_catalog.bpchar", "char", "string", "citext":
		return Str
	case "uuid":
		return "uuid.UUID"
	case "inet", "cidr", "macaddr", "macaddr8":
		// psycopg2 does have support for ipaddress objects, but it is not enabled by default
		//
		// https://www.psycopg.org/docs/extras.html#adapt-network
		return Str
	case "ltree", "lquery", "ltxtquery":
		return Str
	default:
		if pluginType.Schema == "" {
			pluginType.Schema = req.Catalog.DefaultSchema
		}
		for _, schema := range req.Catalog.Schemas {
			if schema.Name == utils.PgCatalog || schema.Name == utils.InformationSchema {
				continue
			}
			if schema.Name != pluginType.Schema {
				continue
			}
			for _, enum := range schema.Enums {
				if pluginType.Name != enum.Name {
					continue
				}
				if schema.Name == req.Catalog.DefaultSchema {
					return "enums." + model.ModelName(config, enum.Name, "")
				}

				return "enums." + model.ModelName(config, enum.Name, schema.Name)
			}
		}
		log.L().Log("unknown PostgreSQL type: " + columnType)

		return "typing.Any"
	}
}
