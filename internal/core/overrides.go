package core

import (
	"fmt"
	"github.com/sqlc-dev/plugin-sdk-go/pattern"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"go/types"
	"strings"
)

type OverridePyType struct {
	Import  string `json:"import" yaml:"import"`
	Name    string `json:"type" yaml:"type"`
	Package string `json:"package" yaml:"package"`
	Spec    string `json:"-"`
	BuiltIn bool   `json:"-"`
}

type ParsedOverridePyType struct {
	ImportPath  string
	TypeName    string
	PackageName string
	BasicType   bool
}

func (gt OverridePyType) parse() (*ParsedOverridePyType, error) {
	var o ParsedOverridePyType

	if gt.Spec == "" {
		o.ImportPath = gt.Import
		o.TypeName = gt.Name
		o.PackageName = gt.Package
		o.BasicType = gt.Import == ""
		return &o, nil
	}

	input := gt.Spec
	lastDot := strings.LastIndex(input, ".")
	lastSlash := strings.LastIndex(input, "/")
	typename := input
	if lastDot == -1 && lastSlash == -1 {
		// if the type name has no slash and no dot, validate that the type is a basic Go type
		var found bool
		for _, typ := range types.Typ {
			info := typ.Info()
			if info == 0 {
				continue
			}
			if info&types.IsUntyped != 0 {
				continue
			}
			if typename == typ.Name() {
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("Package override `go_type` specifier %q is not a Go basic type e.g. 'string'", input)
		}
		o.BasicType = true
	} else {
		// assume the type lives in a Go package
		if lastDot == -1 {
			return nil, fmt.Errorf("Package override `go_type` specifier %q is not the proper format, expected 'package.type', e.g. 'github.com/segmentio/ksuid.KSUID'", input)
		}
		typename = input[lastSlash+1:]
		// a package name beginning with "go-" will give syntax errors in
		// generated code. We should do the right thing and get the actual
		// import name, but in lieu of that, stripping the leading "go-" may get
		// us what we want.
		typename = strings.TrimPrefix(typename, "go-")
		typename = strings.TrimSuffix(typename, "-go")
		o.ImportPath = input[:lastDot]
	}
	o.TypeName = typename
	isPointer := input[0] == '*'
	if isPointer {
		o.ImportPath = o.ImportPath[1:]
		o.TypeName = "*" + o.TypeName
	}
	return &o, nil
}

type Override struct {
	// name of the golang type to use, e.g. `github.com/segmentio/ksuid.KSUID`
	PyType OverridePyType `json:"py_type" yaml:"py_type"`

	// fully qualified name of the Go type, e.g. `github.com/segmentio/ksuid.KSUID`
	DBType string `json:"db_type" yaml:"db_type"`

	// fully qualified name of the column, e.g. `accounts.id`
	Column string `json:"column" yaml:"column"`

	ColumnName    *pattern.Match `json:"-"`
	TableCatalog  *pattern.Match `json:"-"`
	TableSchema   *pattern.Match `json:"-"`
	TableRel      *pattern.Match `json:"-"`
	PyImportPath  string         `json:"-"`
	PyPackageName string         `json:"-"`
	PyTypeName    string         `json:"-"`
	PyBasicType   bool           `json:"-"`
}

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

func (o *Override) parse(req *plugin.GenerateRequest) (err error) {

	schema := "public"
	if req != nil && req.Catalog != nil {
		schema = req.Catalog.DefaultSchema
	}

	// validate option combinations
	switch {
	case o.Column != "" && o.DBType != "":
		return fmt.Errorf("Override specifying both `column` (%q) and `db_type` (%q) is not valid.", o.Column, o.DBType)
	case o.Column == "" && o.DBType == "":
		return fmt.Errorf("Override must specify one of either `column` or `db_type`")
	}

	// validate Column
	if o.Column != "" {
		colParts := strings.Split(o.Column, ".")
		switch len(colParts) {
		case 2:
			if o.ColumnName, err = pattern.MatchCompile(colParts[1]); err != nil {
				return err
			}
			if o.TableRel, err = pattern.MatchCompile(colParts[0]); err != nil {
				return err
			}
			if o.TableSchema, err = pattern.MatchCompile(schema); err != nil {
				return err
			}
		case 3:
			if o.ColumnName, err = pattern.MatchCompile(colParts[2]); err != nil {
				return err
			}
			if o.TableRel, err = pattern.MatchCompile(colParts[1]); err != nil {
				return err
			}
			if o.TableSchema, err = pattern.MatchCompile(colParts[0]); err != nil {
				return err
			}
		case 4:
			if o.ColumnName, err = pattern.MatchCompile(colParts[3]); err != nil {
				return err
			}
			if o.TableRel, err = pattern.MatchCompile(colParts[2]); err != nil {
				return err
			}
			if o.TableSchema, err = pattern.MatchCompile(colParts[1]); err != nil {
				return err
			}
			if o.TableCatalog, err = pattern.MatchCompile(colParts[0]); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Override `column` specifier %q is not the proper format, expected '[catalog.][schema.]tablename.colname'", o.Column)
		}
	}

	// validate GoType
	parsed, err := o.PyType.parse()
	if err != nil {
		return err
	}
	o.PyImportPath = parsed.ImportPath
	o.PyTypeName = parsed.TypeName
	o.PyBasicType = parsed.BasicType
	o.PyPackageName = parsed.PackageName
	return nil
}
