package types

import (
	"log"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func SqliteTypeToPython(req *plugin.GenerateRequest, col *plugin.Column) string {
	columnType := strings.ToLower(sdk.DataType(col.Type))

	// see: https://github.com/sqlc-dev/sqlc/blob/main/internal/codegen/golang/sqlite_type.go
	switch columnType {
	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "unsignedbigint", "int2", "int8":
		return "int"
	case "blob":
		return "bytes"
	case "real", "double", "double precision", "float", "numeric":
		return "float"
	case "boolean", "bool":
		return "bool"
	case "date":
		return "datetime.date"
	case "datetime", "timestamp":
		return "datetime.datetime"
	case "any":
		return "Any"
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
	case strings.HasPrefix(columnType, "decimal"):
		return "decimal.Decimal"

	default:
		log.Printf("unknown SQLite type: %s\n", columnType)
		return "Any"
	}
}
