package types

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
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
