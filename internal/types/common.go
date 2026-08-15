package types

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/log"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// TypeConversionFunc maps one column to its Python type annotation. It
// receives the whole column, not just the type identifier: MySQL needs
// Column.Length to tell tinyint(1) (bool) from tinyint (int).
type TypeConversionFunc func(*plugin.GenerateRequest, *config.Config, *plugin.Column) string

func GetTypeConversionFunc(engine string) (TypeConversionFunc, error) {
	switch engine {
	case "postgresql":
		return PostgresTypeToPython, nil
	case "sqlite":
		return SqliteTypeToPython, nil
	case "mysql":
		return MysqlTypeToPython, nil
	default:
		return nil, fmt.Errorf("engine %q is not supported", engine)
	}
}

// resolveCatalogEnum resolves an unrecognized column type against the
// catalog's enums, shared by the PostgreSQL and MySQL mappers: parse the
// (possibly schema-qualified) identifier, fall back to the default schema,
// skip the system schemas, and qualify the enum name only outside the
// default schema. engineName appears in the unknown-type debug log; Any is
// returned when nothing matches.
func resolveCatalogEnum(req *plugin.GenerateRequest, config *config.Config, engineName, columnType string) string {
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
	log.L().Log("unknown " + engineName + " type: " + columnType)

	return Any
}
