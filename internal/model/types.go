package model

type PyType struct {
	SQLType    string
	Type       string
	IsNullable bool
	IsList     bool
	IsEnum     bool
}

type Enum struct {
	Name      string
	Constants []EnumConstants
}

type EnumConstants struct {
	Name  string
	Value string
}

type Table struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name string
	Type PyType
}

type Query struct {
	Cmd          string // The command of the query: https://docs.sqlc.dev/en/latest/reference/query-annotations.html
	SQL          string // The raw SQL of the query
	ConstantName string // The name of the constant where the raw SQL will be saved in python
	FuncName     string // The name of the python function
}
