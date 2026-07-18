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
	// Precision variants like "decimal(10,5)" keep their prefix; resolve them
	// through the exact "decimal" key.
	if strings.HasPrefix(sqlType, "decimal") {
		return findSqliteConversion("decimal")
	}

	return nil
}

// sqliteNeedsConversion reports whether a SQL type needs runtime conversion for sqlite.
func sqliteNeedsConversion(sqlType string) bool {
	return findSqliteConversion(sqlType) != nil
}

// sqliteConversionUse marks which half of a conversion spec's adapter/converter
// pair the queries actually need.
type sqliteConversionUse struct {
	spec      *sqliteConversion
	adapter   bool
	converter bool
}

// SqliteConversionUsage lists the conversion specs a module's queries need, in
// canonical emission order, split by direction: parameters need a registered
// adapter (Python value -> SQL), returns need a registered converter (SQL value
// -> Python). Registering only what is needed matters because sqlite3
// converters are global: an unnecessary register_converter would change what
// overridden return columns receive under PARSE_DECLTYPES.
type SqliteConversionUsage struct {
	uses []sqliteConversionUse
}

// Any reports whether at least one adapter or converter must be registered.
func (u SqliteConversionUsage) Any() bool {
	return len(u.uses) > 0
}

// RuntimeModules returns the Python modules the emitted conversion setup
// references at runtime: register_adapter needs the adapted type's module,
// and converter bodies reference their type's module unless the speedups
// variant (which references only ciso8601) replaces them. Builtin types
// (bool, memoryview) need no import.
func (u SqliteConversionUsage) RuntimeModules(speedups bool) map[string]struct{} {
	modules := make(map[string]struct{})
	for _, use := range u.uses {
		module, _, found := strings.Cut(use.spec.pyType, ".")
		if !found {
			continue
		}
		if use.adapter || (use.converter && (!speedups || use.spec.speedupsBody == "")) {
			modules[module] = struct{}{}
		}
	}

	return modules
}

// SpeedupConverterUsed reports whether any needed converter has a speedups
// variant - i.e. whether the generated module references ciso8601 when the
// speedups option is enabled.
func (u SqliteConversionUsage) SpeedupConverterUsed() bool {
	for _, use := range u.uses {
		if use.converter && use.spec.speedupsBody != "" {
			return true
		}
	}

	return false
}

// SqliteConversionsUsed collects the conversion specs used by the queries.
// Overridden RETURN columns need no converter - they are converted inline with
// the override type, and registering one anyway would hand the override
// constructor an already-converted value. Overridden PARAMS do need the
// adapter: convertParamExpr converts them back to their DefaultType before
// they reach the driver.
func SqliteConversionsUsed(queries []model.Query) SqliteConversionUsage {
	adapters := make(map[string]struct{})
	converters := make(map[string]struct{})
	addParam := func(typ model.PyType) {
		if spec := findSqliteConversion(typ.SQLType); spec != nil {
			adapters[spec.pyType] = struct{}{}
		}
	}
	addReturn := func(typ model.PyType) {
		if typ.DoOverride() {
			return
		}
		if spec := findSqliteConversion(typ.SQLType); spec != nil {
			converters[spec.pyType] = struct{}{}
		}
	}
	collect := func(qv model.QueryValue, add func(model.PyType)) {
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
		collect(query.Returns, addReturn)
		for _, param := range query.Params {
			collect(param, addParam)
		}
	}

	usage := SqliteConversionUsage{uses: make([]sqliteConversionUse, 0, len(adapters)+len(converters))}
	for i := range sqliteConversions {
		spec := &sqliteConversions[i]
		_, adapter := adapters[spec.pyType]
		_, converter := converters[spec.pyType]
		if adapter || converter {
			usage.uses = append(usage.uses, sqliteConversionUse{spec: spec, adapter: adapter, converter: converter})
		}
	}

	return usage
}
