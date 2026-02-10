package types

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/rayakame/sqlc-gen-better-python/internal/log"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func SqliteTypeToPython(_ *plugin.GenerateRequest, typ *plugin.Identifier, _ *core.Config) string {
	columnType := strings.ToLower(sdk.DataType(typ))

	switch columnType {
	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "unsignedbigint", "int2", "int8", "bigserial":
		return "int"
	case "blob":
		return "memoryview"
	case "real", "double", "double precision", "doubleprecision", "float", "numeric":
		return "float"
	case "boolean", "bool":
		return "bool"
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
		return "str"

	default:
		log.GlobalLogger.Log(fmt.Sprintf("unknown SQLite type: %s", columnType))
		return "typing.Any"
	}
}
