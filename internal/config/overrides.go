package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/pattern"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// OverridePyType describes the Python type that replaces the default mapping.
// Import is the module to import (e.g. "collections"); Type is the type
// expression used in annotations (e.g. "UserString" or "collections.UserString");
// Package is the name imported from Import ("from <import> import <package>") —
// if empty, Import is imported as a plain module ("import <import>").
type OverridePyType struct {
	Import  string `json:"import"  yaml:"import"`
	Type    string `json:"type"    yaml:"type"`
	Package string `json:"package" yaml:"package"`
}

// Override replaces the default Python type of columns matched either by
// SQL type (DBType) or by a column pattern ("[catalog.][schema.]table.column",
// wildcards supported).
type Override struct {
	PyType OverridePyType `json:"py_type" yaml:"py_type"`

	// DBType matches the SQL data type exactly, e.g. "text" or "pg_catalog.int4".
	DBType string `json:"db_type" yaml:"db_type"`

	// Column matches a fully qualified column name, e.g. "authors.name".
	Column string `json:"column" yaml:"column"`

	ColumnName   *pattern.Match `json:"-" yaml:"-"`
	TableCatalog *pattern.Match `json:"-" yaml:"-"`
	TableSchema  *pattern.Match `json:"-" yaml:"-"`
	TableRel     *pattern.Match `json:"-" yaml:"-"`
}

// Matches reports whether the override's table pattern matches the identifier.
func (o *Override) Matches(n *plugin.Identifier, defaultSchema string) bool {
	if n == nil {
		return false
	}
	schema := n.Schema
	if n.Schema == "" {
		schema = defaultSchema
	}
	if o.TableCatalog != nil && !o.TableCatalog.MatchString(n.Catalog) {
		return false
	}
	if o.TableSchema == nil && schema != "" {
		return false
	}
	if o.TableSchema != nil && !o.TableSchema.MatchString(schema) {
		return false
	}
	if o.TableRel == nil && n.Name != "" {
		return false
	}
	if o.TableRel != nil && !o.TableRel.MatchString(n.Name) {
		return false
	}

	return true
}

const (
	overrideColumnPartsTable         = 2
	overrideColumnPartsSchemaTable   = 3
	overrideColumnPartsCatalogSchema = 4
)

func (o *Override) parse(req *plugin.GenerateRequest) error {
	schema := "public"
	if req != nil && req.Catalog != nil {
		schema = req.Catalog.DefaultSchema
	}

	switch {
	case o.Column != "" && o.DBType != "":
		return fmt.Errorf("override specifying both `column` (%q) and `db_type` (%q) is not valid", o.Column, o.DBType)
	case o.Column == "" && o.DBType == "":
		return errors.New("override must specify one of either `column` or `db_type`")
	}

	if o.PyType.Type == "" {
		return errors.New("override must specify a `py_type` with a non-empty `type`")
	}

	if o.Column != "" {
		return o.parseColumnPattern(schema)
	}

	return nil
}

// parseColumnPattern compiles the "[catalog.][schema.]tablename.colname" parts
// of the Column specifier into match patterns.
func (o *Override) parseColumnPattern(defaultSchema string) error {
	type target struct {
		dst  **pattern.Match
		expr string
	}

	colParts := strings.Split(o.Column, ".")
	var targets []target
	switch len(colParts) {
	case overrideColumnPartsTable:
		targets = []target{{&o.ColumnName, colParts[1]}, {&o.TableRel, colParts[0]}, {&o.TableSchema, defaultSchema}}
	case overrideColumnPartsSchemaTable:
		targets = []target{{&o.ColumnName, colParts[2]}, {&o.TableRel, colParts[1]}, {&o.TableSchema, colParts[0]}}
	case overrideColumnPartsCatalogSchema:
		targets = []target{
			{&o.ColumnName, colParts[3]},
			{&o.TableRel, colParts[2]},
			{&o.TableSchema, colParts[1]},
			{&o.TableCatalog, colParts[0]},
		}
	default:
		return fmt.Errorf(
			"override `column` specifier %q is not the proper format, expected '[catalog.][schema.]tablename.colname'",
			o.Column,
		)
	}

	for _, tgt := range targets {
		compiled, err := pattern.MatchCompile(tgt.expr)
		if err != nil {
			return err
		}
		*tgt.dst = compiled
	}

	return nil
}
