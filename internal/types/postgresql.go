package types

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/log"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

const (
	identifierPartsSchemaName        = 2
	identifierPartsCatalogSchemaName = 3
)

func parseIdentifierString(name string) (*plugin.Identifier, error) {
	parts := strings.Split(name, ".")
	switch len(parts) {
	case 1:
		return &plugin.Identifier{
			Name: parts[0],
		}, nil
	case identifierPartsSchemaName:
		return &plugin.Identifier{
			Schema: parts[0],
			Name:   parts[1],
		}, nil
	case identifierPartsCatalogSchemaName:
		return &plugin.Identifier{
			Catalog: parts[0],
			Schema:  parts[1],
			Name:    parts[2],
		}, nil
	default:
		return nil, fmt.Errorf("invalid name: %s", name)
	}
}

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
		return Decimal
	case "money":
		return Str
	case Boolean, Bool, "pg_catalog.bool":
		return Bool
	case "pg_catalog.json", "json", "jsonb":
		return Str
	case "bytea", "blob", "pg_catalog.bytea":
		return Memoryview
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
	case "uuid", "pg_catalog.uuid":
		return "uuid.UUID"
	case "inet", "cidr", "macaddr", "macaddr8":
		// psycopg2 does have support for ipaddress objects, but it is not enabled by default
		//
		// https://www.psycopg.org/docs/extras.html#adapt-network
		return Str
	case "ltree", "lquery", "ltxtquery":
		return Str
	default:
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
					return "enums." + model.EnumName(config, enum.Name, "")
				}

				return "enums." + model.EnumName(config, enum.Name, schema.Name)
			}
		}
		log.L().Log("unknown PostgreSQL type: " + columnType)

		return Any
	}
}
