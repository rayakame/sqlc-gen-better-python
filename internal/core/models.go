package core

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/typeConversion"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Table struct {
	Table   *plugin.Identifier
	Name    string
	Columns []Column
	Comment string
}

type PyType struct {
	SqlType     string
	Type        string
	DefaultType string
	IsList      bool
	IsNullable  bool
	IsEnum      bool
	IsOverride  bool
	Override    *Override
}

func (p *PyType) DoConversion(conversion typeConversion.TypeDoTypeConversion) bool {
	if p.DoOverride() {
		return true
	}
	return conversion(p.SqlType)
}
func (p *PyType) DoOverride() bool {
	return p.IsOverride && p.Override != nil
}

type Constant struct {
	Name  string
	Type  PyType
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
	Emit   bool
	Name   string
	DBName string // The name of the field in the database. Only set if Struct==nil.
	Table  *Table
	Typ    PyType

	// Column is kept so late in the generation process around to differentiate
	// between mysql slices and pg arrays
	Column *plugin.Column
}

func (v QueryValue) EmitStruct() bool {
	return v.Emit
}

func (v QueryValue) IsStruct() bool {
	return v.Table != nil
}

func (v QueryValue) IsEmpty() bool {
	return v.Typ.Type == "" && v.Name == "" && v.Table == nil
}

func (v QueryValue) Type() string {
	if v.Typ.Type != "" {
		return v.Typ.Type
	}
	if v.Table != nil {
		return v.Table.Name
	}
	panic("no type for QueryValue: " + v.Name)
}

type Query struct {
	Cmd          string
	Comments     []string
	MethodName   string
	FuncName     string
	FieldName    string
	ConstantName string
	SQL          string
	SourceName   string
	Ret          QueryValue
	Args         []QueryValue

	// Used for :copyfrom
	Table *plugin.Identifier
}

func (q Query) HasRetType() bool {
	scanned := q.Cmd == metadata.CmdOne || q.Cmd == metadata.CmdMany ||
		q.Cmd == metadata.CmdBatchMany || q.Cmd == metadata.CmdBatchOne
	return scanned && !q.Ret.IsEmpty()
}

func IsAnyQueryMany(queries []Query) bool {
	for _, query := range queries {
		if query.Cmd == metadata.CmdMany {
			return true
		}
	}
	return false
}
