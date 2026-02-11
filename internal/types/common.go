package types

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type TypeConversionFunc func(*plugin.GenerateRequest, *config.Config, *plugin.Identifier) string

func GetTypeConversionFunc(engine string) (TypeConversionFunc, error) {
	switch engine {
	case "postgresql":
		return PostgresTypeToPython, nil
	case "sqlite":
		return SqliteTypeToPython, nil
	default:
		return nil, fmt.Errorf("engine %q is not supported", engine)
	}
}
