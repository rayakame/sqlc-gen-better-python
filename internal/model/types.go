package model

import (
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type PyType struct {
	SQLType    string
	Type       string
	IsNullable bool
	IsList     bool
	IsEnum     bool

	// IsOverride marks a type replaced via the overrides config option.
	IsOverride bool
	// DefaultType is the Python type the column would have without the
	// override. Used to convert override values before passing them to the driver.
	DefaultType string
}

func (t PyType) Print() string {
	type_ := t.Type
	if t.IsList {
		type_ = fmt.Sprintf("collections.abc.Sequence[%s]", type_)
	}
	if t.IsNullable {
		type_ += " | None"
	}
	return type_
}

// PrintOptional prints the type with a guaranteed "| None" suffix for values
// that may be absent at runtime (e.g. :one queries that match no row).
func (t PyType) PrintOptional() string {
	if t.IsNullable {
		return t.Print()
	}

	return t.Print() + " | None"
}

// DoOverride reports whether this type has an active override.
func (t PyType) DoOverride() bool {
	return t.IsOverride
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

	Identifier *plugin.Identifier
}

type Column struct {
	Name   string // The escaped Python attribute name
	DBName string // The raw database column name
	Type   PyType

	Embed *Embed
}

type Query struct {
	Cmd          string // The command of the query: https://docs.sqlc.dev/en/latest/reference/query-annotations.html
	SQL          string // The raw SQL of the query
	ConstantName string // The name of the constant where the raw SQL will be saved in python
	FuncName     string // The name of the python function
	QueryName    string // The original name of the query
	FileName     string // The original filename where the query is located
	ModuleName   string // The name of the python module in which the query will be implemented

	Params  []QueryValue
	Returns QueryValue

	Table *plugin.Identifier // The name of the table this query inserts into. Only used for :copyfrom
}

func (q Query) EmitsTable() bool {
	if q.Returns.EmitTable {
		return true
	}
	for _, param := range q.Params {
		if param.EmitTable {
			return true
		}
	}
	return false
}

type QueryValue struct {
	EmitTable bool
	Table     *Table
	Name      string
	DBName    string
	Type      PyType
}

type Embed struct {
	ModelName string
	Columns   []Column
}

// IsEmpty reports whether this value is unset.
func (v QueryValue) IsEmpty() bool {
	return v.Type.Type == "" && v.Name == "" && v.Table == nil
}

// IsStruct reports whether this value is a structured type (table reference).
func (v QueryValue) IsStruct() bool {
	return v.Table != nil
}
