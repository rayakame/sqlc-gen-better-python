package core

import (
	"fmt"
	"sort"
	"strings"
)

type importSpec struct {
	Module string
	Name   string
	Alias  string
}

func (i importSpec) String() string {
	if i.Alias != "" {
		if i.Name == "" {
			return fmt.Sprintf("import %s as %s", i.Module, i.Alias)
		}
		return fmt.Sprintf("from %s import %s as %s", i.Module, i.Name, i.Alias)
	}
	if i.Name == "" {
		return "import " + i.Module
	}
	return fmt.Sprintf("from %s import %s", i.Module, i.Name)
}

type Importer struct {
	Tables  []Table
	Queries []Query
	Enums   []Enum
	C       *Config
}

func (i *Importer) Imports(fileName string) []string {
	if fileName == "models.sql" {
		return i.modelImports()
	}
	return i.queryImports(fileName)
}

func tableUses(name string, s Table) bool {
	for _, col := range s.Columns {
		if name == "typing" && col.Type.IsList || name == "typing" && col.Type.IsNullable {
			return true
		}
		if col.Type.Type == name {
			return true
		}
	}
	return false
}

func (i *Importer) getModelImportSpec() (string, importSpec, error) {
	switch i.C.ModelType {
	case ModelTypeAttrs:
		return "attrs", importSpec{Module: "attrs"}, nil
	case ModelTypeDataclass:
		return "dataclasses", importSpec{Module: "dataclasses"}, nil
	default:
		return "", importSpec{}, fmt.Errorf("unknown model type: %s", i.C.ModelType)
	}
}

func (i *Importer) modelImportSpecs() (map[string]importSpec, map[string]importSpec) {
	modelUses := func(name string) bool {
		for _, table := range i.Tables {
			if tableUses(name, table) {
				return true
			}
		}
		return false
	}

	std := stdImports(modelUses)
	modelName, modelImport, err := i.getModelImportSpec()
	if err == nil {
		std[modelName] = modelImport
	}
	if len(i.Enums) > 0 {
		std["enum"] = importSpec{Module: fmt.Sprintf("from %s import enums", i.C.Package)}
	}

	pkg := make(map[string]importSpec)

	return std, pkg
}

func queryValueUses(name string, qv QueryValue) bool {
	if !qv.IsEmpty() {
		if name == "typing" && qv.Typ.IsList {
			return true
		}
		if name == "typing" && qv.Typ.IsNullable {
			return true
		}
		if qv.IsStruct() && qv.EmitStruct() {
			if tableUses(name, *qv.Table) {
				return true
			}
		} else {
			if qv.Typ.Type == name {
				return true
			}
		}
	}
	return false
}

func (i *Importer) queryImportSpecs(fileName string) (map[string]importSpec, map[string]importSpec, map[string]importSpec) {
	queryUses := func(name string) bool {
		for _, q := range i.Queries {
			//if q.SourceName != fileName { TODO q.SourceName is the name of the sql file
			//	continue
			//}
			if queryValueUses(name, q.Ret) {
				return true
			}
			if queryValueUses(name, q.Arg) {
				return true
			}
		}
		return false
	}

	std := stdImports(queryUses)

	pkg := make(map[string]importSpec)
	loc := make(map[string]importSpec)
	pkg[i.C.SqlDriver.String()] = importSpec{Module: i.C.SqlDriver.String()}

	queryValueModelImports := func(qv QueryValue) {
		if qv.IsStruct() && qv.EmitStruct() {
			modelName, modelImport, err := i.getModelImportSpec()
			if err == nil {
				std[modelName] = modelImport
			}
		}
	}

	for _, q := range i.Queries {
		//if q.SourceName != fileName { TODO
		//	continue
		//}
		queryValueModelImports(q.Ret)
		queryValueModelImports(q.Arg)
	}

	loc["models"] = importSpec{Module: i.C.Package, Name: "models"}

	return std, pkg, loc
}

func (i *Importer) queryImports(fileName string) []string {
	std, pkg, loc := i.queryImportSpecs(fileName)

	importLines := make([]string, 0)
	if len(std) != 0 {
		importLines = append(importLines, buildImportBlock(std))
	}
	if len(pkg) != 0 {
		if len(importLines) != 0 {
			importLines = append(importLines, "")
		}
		importLines = append(importLines, buildImportBlock(pkg))
	}
	if len(loc) != 0 {
		if len(importLines) != 0 {
			importLines = append(importLines, "")
		}
		importLines = append(importLines, buildImportBlock(loc))
	}
	return importLines
}

func (i *Importer) modelImports() []string {
	std, pkg := i.modelImportSpecs()
	importLines := make([]string, 0)
	if len(std) != 0 {
		importLines = append(importLines, buildImportBlock(std))
	}
	if len(pkg) != 0 {
		if len(importLines) != 0 {
			importLines = append(importLines, "")
		}
		importLines = append(importLines, buildImportBlock(pkg))
	}
	return importLines
}

func buildImportBlock(pkgs map[string]importSpec) string {
	pkgImports := make([]importSpec, 0)
	fromImports := make(map[string][]string)
	for _, is := range pkgs {
		if is.Name == "" || is.Alias != "" {
			pkgImports = append(pkgImports, is)
		} else {
			names, ok := fromImports[is.Module]
			if !ok {
				names = make([]string, 0, 1)
			}
			names = append(names, is.Name)
			fromImports[is.Module] = names
		}
	}

	importStrings := make([]string, 0, len(pkgImports)+len(fromImports))
	for _, is := range pkgImports {
		importStrings = append(importStrings, is.String())
	}
	for modName, names := range fromImports {
		sort.Strings(names)
		nameString := strings.Join(names, ", ")
		importStrings = append(importStrings, fmt.Sprintf("from %s import %s", modName, nameString))
	}
	sort.Strings(importStrings)
	return strings.Join(importStrings, "\n")
}

func stdImports(uses func(name string) bool) map[string]importSpec {
	std := make(map[string]importSpec)
	std["typing"] = importSpec{Module: "typing"}
	if uses("decimal.Decimal") {
		std["decimal"] = importSpec{Module: "decimal"}
	}
	if uses("datetime.date") || uses("datetime.time") || uses("datetime.datetime") || uses("datetime.timedelta") {
		std["datetime"] = importSpec{Module: "datetime"}
	}
	if uses("uuid.UUID") {
		std["uuid"] = importSpec{Module: "uuid"}
	}
	return std
}
