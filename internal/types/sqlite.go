package types

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/log"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func SqliteTypeToPython(_ *plugin.GenerateRequest, _ *config.Config, pluginType *plugin.Identifier) string {
	columnType := sdk.DataType(pluginType)

	switch columnType {
	case Int, "integer", "tinyint", "smallint", "mediumint", "bigint", "unsignedbigint", "int2", "int8", "bigserial":
		return Int
	case "blob":
		return "memoryview"
	case "real", "double", "double precision", "doubleprecision", Float, "numeric":
		return Float
	case Boolean, Bool:
		return Bool
	case "date":
		return "datetime.date"
	case "datetime", "timestamp":
		return "datetime.datetime"
	case "decimal":
		return "decimal.Decimal"
	}

	switch {
	case strings.HasPrefix(columnType, "character"),
		strings.HasPrefix(columnType, "varchar"),
		strings.HasPrefix(columnType, "varyingcharacter"),
		strings.HasPrefix(columnType, "nchar"),
		strings.HasPrefix(columnType, "nativecharacter"),
		strings.HasPrefix(columnType, "nvarchar"),
		columnType == "text",
		columnType == "clob",
		columnType == "json":
		return Str

	default:
		log.L().Log("unknown SQLite type: " + columnType)

		return "typing.Any"
	}
}
