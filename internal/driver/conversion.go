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

// asyncpgNeedsConversion reports whether a SQL type needs runtime conversion for asyncpg.
func asyncpgNeedsConversion(sqlType string) bool {
	_, ok := asyncpgConversions[sqlType]
	return ok
}

// sqliteConversion fully describes one sqlite type conversion: the SQL type
// names that trigger it (which double as the register_converter keys), the
// Python type, and how to emit its adapter/converter pair. Adding a new
// convertible type means adding exactly one entry here.
type sqliteConversion struct {
	pyType       string   // Python type the adapter is registered for
	suffix       string   // function name suffix, e.g. "date"
	sqlTypes     []string // SQL type names mapping to this conversion / converter keys
	adaptRet     string   // adapter return annotation
	adaptBody    string   // adapter body expression
	convBody     string   // converter body expression
	speedupsBody string   // converter body when speedups are enabled ("" = same as convBody)
}

// sqliteConversions is ordered: the slice order defines the emission order of
// the adapter/converter pairs in generated modules.
var sqliteConversions = []sqliteConversion{
	{
		pyType:       "datetime.date",
		suffix:       "date",
		sqlTypes:     []string{"date"},
		adaptRet:     "str",
		adaptBody:    "val.isoformat()",
		convBody:     "datetime.date.fromisoformat(val.decode())",
		speedupsBody: "ciso8601.parse_datetime(val.decode()).date()",
	},
	{
		pyType:       "decimal.Decimal",
		suffix:       "decimal",
		sqlTypes:     []string{"decimal"},
		adaptRet:     "str",
		adaptBody:    "str(val)",
		convBody:     "decimal.Decimal(val.decode())",
		speedupsBody: "",
	},
	{
		pyType:       "datetime.datetime",
		suffix:       "datetime",
		sqlTypes:     []string{"datetime", "timestamp"},
		adaptRet:     "str",
		adaptBody:    "val.isoformat()",
		convBody:     "datetime.datetime.fromisoformat(val.decode())",
		speedupsBody: "ciso8601.parse_datetime(val.decode())",
	},
	{
		pyType:       "bool",
		suffix:       "bool",
		sqlTypes:     []string{"bool", "boolean"},
		adaptRet:     "int",
		adaptBody:    "int(val)",
		convBody:     "bool(int(val))",
		speedupsBody: "",
	},
	{
		pyType:       "memoryview",
		suffix:       "memoryview",
		sqlTypes:     []string{"blob"},
		adaptRet:     "bytes",
		adaptBody:    "val.tobytes()",
		convBody:     "memoryview(val)",
		speedupsBody: "",
	},
}

// findSqliteConversion returns the conversion spec for a SQL type, or nil.
func findSqliteConversion(sqlType string) *sqliteConversion {
	for i := range sqliteConversions {
		for _, name := range sqliteConversions[i].sqlTypes {
			if name == sqlType {
				return &sqliteConversions[i]
			}
		}
	}
	// Precision variants like "decimal(10,5)" keep their prefix.
	if strings.HasPrefix(sqlType, "decimal") {
		for i := range sqliteConversions {
			if sqliteConversions[i].pyType == "decimal.Decimal" {
				return &sqliteConversions[i]
			}
		}
	}

	return nil
}

// sqliteNeedsConversion reports whether a SQL type needs runtime conversion for sqlite.
func sqliteNeedsConversion(sqlType string) bool {
	return findSqliteConversion(sqlType) != nil
}

// SqliteConversionsUsed returns the Python types used by the given queries that
// need a registered sqlite adapter/converter pair, in canonical emission order.
// Overridden RETURN columns are excluded — those are converted inline with the
// override type. Overridden PARAMS are included: convertParamExpr converts
// them back to their DefaultType, which still needs the registered adapter.
func SqliteConversionsUsed(queries []model.Query) []string {
	used := make(map[string]struct{})
	add := func(typ model.PyType, skipOverride bool) {
		if skipOverride && typ.DoOverride() {
			return
		}
		if spec := findSqliteConversion(typ.SQLType); spec != nil {
			used[spec.pyType] = struct{}{}
		}
	}
	collect := func(qv model.QueryValue, skipOverride bool) {
		if qv.IsEmpty() {
			return
		}
		if qv.IsStruct() {
			for _, col := range qv.Table.Columns {
				if col.Embed != nil {
					for _, embedCol := range col.Embed.Columns {
						add(embedCol.Type, skipOverride)
					}

					continue
				}
				add(col.Type, skipOverride)
			}

			return
		}
		add(qv.Type, skipOverride)
	}
	for _, query := range queries {
		collect(query.Returns, true)
		for _, param := range query.Params {
			collect(param, false)
		}
	}

	result := make([]string, 0, len(used))
	for i := range sqliteConversions {
		if _, ok := used[sqliteConversions[i].pyType]; ok {
			result = append(result, sqliteConversions[i].pyType)
		}
	}

	return result
}
