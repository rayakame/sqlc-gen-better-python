package core

import (
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"strings"
)

type Table struct {
	Table   *plugin.Identifier
	Name    string
	Columns []Column
	Comment string
}

type PyType struct {
	SqlType    string
	Type       string
	IsList     bool
	IsNullable bool
	IsEnum     bool
}
type Constant struct {
	Name  string
	Type  string
	Value string
}

type Enum struct {
	Name      string
	Comment   string
	Constants []Constant
}

func enumReplacer(r rune) rune {
	if strings.ContainsRune("-/:_", r) {
		return '_'
	} else if (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') {
		return r
	} else {
		return -1
	}
}

// EnumReplace removes all non ident symbols (all but letters, numbers and
// underscore) and returns valid ident name for provided name.
func EnumReplace(value string) string {
	return strings.Map(enumReplacer, value)
}

type QueryValue struct {
	Emit        bool
	EmitPointer bool
	Name        string
	DBName      string // The name of the field in the database. Only set if Struct==nil.
	Table       *Table
	Typ         PyType
	SQLDriver   string

	// Column is kept so late in the generation process around to differentiate
	// between mysql slices and pg arrays
	Column *plugin.Column
}

type Query struct {
	Cmd          string
	Comments     []string
	MethodName   string
	FieldName    string
	ConstantName string
	SQL          string
	SourceName   string
	Ret          QueryValue
	Arg          QueryValue
	// Used for :copyfrom
	Table *plugin.Identifier
}
