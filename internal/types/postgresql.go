package types

import (
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"log"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func PostgresTypeToPython(req *plugin.GenerateRequest, col *plugin.Column, conf *core.Config) string {
	columnType := sdk.DataType(col.Type)

	switch columnType {
	case "serial", "serial4", "pg_catalog.serial4", "bigserial", "serial8", "pg_catalog.serial8", "smallserial", "serial2", "pg_catalog.serial2", "integer", "int", "int4", "pg_catalog.int4", "bigint", "int8", "pg_catalog.int8", "smallint", "int2", "pg_catalog.int2":
		return "int"
	case "float", "double precision", "float8", "pg_catalog.float8", "real", "float4", "pg_catalog.float4":
		return "float"
	case "numeric", "pg_catalog.numeric", "money":
		return "decimal.Decimal"
	case "boolean", "bool", "pg_catalog.bool":
		return "bool"
	case "json", "jsonb":
		return "Any"
	case "bytea", "blob", "pg_catalog.bytea":
		return "memoryview"
	case "date":
		return "datetime.date"
	case "pg_catalog.time", "pg_catalog.timetz":
		return "datetime.time"
	case "pg_catalog.timestamp", "pg_catalog.timestamptz", "timestamptz":
		return "datetime.datetime"
	case "interval", "pg_catalog.interval":
		return "datetime.timedelta"
	case "text", "pg_catalog.varchar", "pg_catalog.bpchar", "string", "citext":
		return "str"
	case "uuid":
		return "uuid.UUID"
	case "inet", "cidr", "macaddr", "macaddr8":
		// psycopg2 does have support for ipaddress objects, but it is not enabled by default
		//
		// https://www.psycopg.org/docs/extras.html#adapt-network
		return "str"
	case "ltree", "lquery", "ltxtquery":
		return "str"
	default:
		for _, schema := range req.Catalog.Schemas {
			if schema.Name == "pg_catalog" || schema.Name == "information_schema" {
				continue
			}
			for _, enum := range schema.Enums {
				if columnType == enum.Name {
					if schema.Name == req.Catalog.DefaultSchema {
						return "models." + core.ModelName(enum.Name, "", conf)
					}
					return "models." + core.ModelName(enum.Name, schema.Name, conf)
				}
			}
		}
		log.Printf("unknown PostgreSQL type: %s\n", columnType)
		return "Any"
	}
}
