package types

// Python type names emitted into generated code. Bool and Boolean double as
// SQL type spellings in the conversion tables.
const (
	Bool    = "bool"
	Boolean = "boolean"
	Str     = "str"
	Int     = "int"
	Float   = "float"
	Decimal = "decimal.Decimal"
	Any     = "typing.Any"
)

// SQL system schemas whose objects are never rendered as models or enums.
const (
	InformationSchema = "information_schema"
	PgCatalog         = "pg_catalog"
)
