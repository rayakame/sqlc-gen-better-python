package driver

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
)

// asyncpgConversions lists SQL types that need explicit Python-side
// type conversion when using the asyncpg driver.
var asyncpgConversions = map[string]struct{}{
	"bytea":            {},
	"blob":             {},
	"pg_catalog.bytea": {},
	"inet":             {},
	"cidr":             {},
}

// sqliteConversions lists SQL types that need explicit Python-side
// type conversion when using sqlite3 or aiosqlite drivers.
var sqliteConversions = map[string]struct{}{
	"boolean":   {},
	"bool":      {},
	"date":      {},
	"datetime":  {},
	"timestamp": {},
	"decimal":   {},
	"blob":      {},
}

// asyncpgNeedsConversion reports whether a SQL type needs runtime conversion for asyncpg.
func asyncpgNeedsConversion(sqlType string) bool {
	_, ok := asyncpgConversions[sqlType]
	return ok
}

// sqliteNeedsConversion reports whether a SQL type needs runtime conversion for sqlite.
func sqliteNeedsConversion(sqlType string) bool {
	_, ok := sqliteConversions[sqlType]
	if ok {
		return true
	}
	return strings.HasPrefix(sqlType, "decimal")
}

// sqliteConversionOrder is the canonical emission order for adapter/converter pairs.
var sqliteConversionOrder = []string{"datetime.date", "decimal.Decimal", "datetime.datetime", "bool", "memoryview"}

// SqliteConversionsUsed returns the Python types used by the given queries that
// need a registered sqlite adapter/converter pair, in canonical emission order.
// Overridden columns are excluded — those are converted inline with the override type.
func SqliteConversionsUsed(queries []model.Query) []string {
	used := make(map[string]struct{})
	add := func(typ model.PyType) {
		if typ.DoOverride() {
			return
		}
		if sqliteNeedsConversion(typ.SQLType) {
			used[typ.Type] = struct{}{}
		}
	}
	collect := func(qv model.QueryValue) {
		if qv.IsEmpty() {
			return
		}
		if qv.IsStruct() {
			for _, col := range qv.Table.Columns {
				if col.Embed != nil {
					for _, embedCol := range col.Embed.Columns {
						add(embedCol.Type)
					}

					continue
				}
				add(col.Type)
			}

			return
		}
		add(qv.Type)
	}
	for _, query := range queries {
		collect(query.Returns)
		for _, param := range query.Params {
			collect(param)
		}
	}

	result := make([]string, 0, len(used))
	for _, name := range sqliteConversionOrder {
		if _, ok := used[name]; ok {
			result = append(result, name)
		}
	}

	return result
}
