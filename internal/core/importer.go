package core

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/typeConversion"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"sort"
	"strings"
)

type importSpec struct {
	Module       string
	Name         string
	Alias        string
	TypeChecking bool
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

func (i *Importer) Imports(fileName string) ([]string, []string, []string) {
	if fileName == "models.sql" {
		return i.modelImports()
	}
	return i.queryImports(fileName)
}

func tableUses(name string, s Table) (bool, string) {
	for _, col := range s.Columns {
		if col.Type.Type == name {
			return true, col.Type.SqlType
		}
	}
	return false, ""

}

func (i *Importer) getModelImportSpec() (string, importSpec, error) {
	switch i.C.ModelType {
	case ModelTypeAttrs:
		return "attrs", importSpec{Module: "attrs"}, nil
	case ModelTypeDataclass:
		return "dataclasses", importSpec{Module: "dataclasses"}, nil
	case ModelTypeMsgspec:
		return "msgspec", importSpec{Module: "msgspec"}, nil
	default:
		return "", importSpec{}, fmt.Errorf("unknown model type: %s", i.C.ModelType)
	}
}

func (i *Importer) splitTypeChecking(pks map[string]importSpec) (map[string]importSpec, map[string]importSpec) {
	normalImports := make(map[string]importSpec)
	typeChecking := make(map[string]importSpec)
	for name, val := range pks {
		if val.TypeChecking {
			typeChecking[name] = val
		} else {
			normalImports[name] = val
		}
	}
	return normalImports, typeChecking
}

func (i *Importer) modelImportSpecs() (map[string]importSpec, map[string]importSpec, map[string]importSpec) {
	modelUses := func(name string) (bool, bool) {
		for _, table := range i.Tables {
			if val, _ := tableUses(name, table); val {
				return true, true
			}
		}
		return false, false
	}

	std := stdImports(modelUses)
	std, typeChecking := i.splitTypeChecking(std)
	if len(typeChecking) != 0 {
		std["typing"] = importSpec{Module: "typing"}
	}
	modelName, modelImport, err := i.getModelImportSpec()
	if err == nil {
		std[modelName] = modelImport
	}
	if len(i.Enums) > 0 {
		std["enum"] = importSpec{Module: fmt.Sprintf("from %s import enums", i.C.Package)}
	}

	pkg := make(map[string]importSpec)

	return std, typeChecking, pkg
}

func (i *Importer) queryValueUses(name string, qv QueryValue) (bool, bool) {
	if !qv.IsEmpty() {
		if qv.IsStruct() && qv.EmitStruct() {
			if val, sqlType := tableUses(name, *qv.Table); val {
				if i.C.SqlDriver == SQLDriverAsyncpg {
					if _, found := typeConversion.AsyncpgDoTypeConversion()[sqlType]; found {
						return true, false
					} else {
						return true, true
					}
				}
				return true, false
			}
		} else {
			if qv.Typ.Type == name {
				if i.C.SqlDriver == SQLDriverAsyncpg {
					if _, found := typeConversion.AsyncpgDoTypeConversion()[qv.Typ.SqlType]; found {
						return true, false
					} else {
						return true, true
					}
				}
				return true, false
			}
		}
	}
	return false, false
}

func (i *Importer) queryImportSpecs(fileName string) (map[string]importSpec, map[string]importSpec, map[string]importSpec, map[string]importSpec) {
	queryUses := func(name string) (bool, bool) {
		var uses *bool = nil
		var typeChecking *bool = nil

		helper := func(val1, val2 bool) {
			if uses == nil || typeChecking == nil {
				uses = new(bool)
				typeChecking = new(bool)
				*uses = val1
				*typeChecking = val2
			} else if *typeChecking == true {
				*uses = val1
				*typeChecking = val2
			}
		}
		for _, q := range i.Queries {
			//if q.SourceName != fileName { TODO q.SourceName is the name of the sql file
			//	continue
			//}
			if val1, val2 := i.queryValueUses(name, q.Ret); val1 {
				if q.Cmd == metadata.CmdMany {
					helper(val1, false)
				}
				helper(val1, val2)
			}
			for _, arg := range q.Args {
				if val1, val2 := i.queryValueUses(name, arg); val1 {
					helper(val1, val2)
				}
			}
		}
		if uses == nil || typeChecking == nil {
			return false, false
		}
		return *uses, *typeChecking
	}

	std := stdImports(queryUses)
	std, typeChecking := i.splitTypeChecking(std)
	typeChecking[i.C.SqlDriver.String()] = importSpec{Module: i.C.SqlDriver.String()}
	if IsAnyQueryMany(i.Queries) && i.C.SqlDriver == SQLDriverAsyncpg {
		typeChecking[i.C.SqlDriver.String()+".cursor"] = importSpec{Module: i.C.SqlDriver.String() + ".cursor"}
	}
	std["typing"] = importSpec{Module: "typing"}

	pkg := make(map[string]importSpec)
	loc := make(map[string]importSpec)

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
	}

	loc["models"] = importSpec{Module: i.C.Package, Name: "models"}

	return std, typeChecking, pkg, loc
}

func (i *Importer) queryImports(fileName string) ([]string, []string, []string) {
	std, typeCheck, pkg, loc := i.queryImportSpecs(fileName)

	importLines := make([]string, 0)
	typeLines := make([]string, 0)
	packageLines := make([]string, 0)
	if len(std) != 0 {
		importLines = append(importLines, buildImportBlock(std)...)
	}
	if len(typeCheck) != 0 {
		typeLines = append(typeLines, buildImportBlock(typeCheck)...)
	}
	if IsAnyQueryMany(i.Queries) {
		if len(typeCheck) != 0 {
			typeLines[len(typeLines)-1] = typeLines[len(typeLines)-1] + "\n"
		}
		queryResultsArgsType := "QueryResultsArgsType: typing.TypeAlias = int | float | str | memoryview"
		if IsInMultipleMaps("decimal", std, typeCheck) {
			queryResultsArgsType += " | decimal.Decimal"
		}
		if IsInMultipleMaps("uuid", std, typeCheck) {
			queryResultsArgsType += " | uuid.UUID"
		}
		if IsInMultipleMaps("datetime", std, typeCheck) {
			queryResultsArgsType += " | datetime.date | datetime.time | datetime.datetime | datetime.timedelta"
		}
		typeLines = append(typeLines, queryResultsArgsType)
	}

	if len(pkg) != 0 {
		packageLines = append(packageLines, buildImportBlock(pkg)...)
	}
	if len(loc) != 0 {
		if len(packageLines) != 0 {
			packageLines = append(packageLines, "")
		}
		packageLines = append(packageLines, buildImportBlock(loc)...)
	}
	return importLines, typeLines, packageLines
}

func (i *Importer) modelImports() ([]string, []string, []string) {
	std, typeCheck, pkg := i.modelImportSpecs()
	importLines := make([]string, 0)
	typeLines := make([]string, 0)
	packageLines := make([]string, 0)
	if len(std) != 0 {
		importLines = append(importLines, buildImportBlock(std)...)
	}
	if len(typeCheck) != 0 {
		typeLines = append(typeLines, buildImportBlock(typeCheck)...)
	}
	if len(pkg) != 0 {
		packageLines = append(packageLines, buildImportBlock(pkg)...)
	}
	return importLines, typeLines, packageLines
}

func buildImportBlock(pkgs map[string]importSpec) []string {
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
	return importStrings
}

// typeCheckingOverwriteProtection function that takes in importSpec map and adds/replaced imports.
// Important here is that importSpec's with TypeChecking set to false have higher priority then
// type checking imports.
func typeCheckingOverwriteProtection(std map[string]importSpec, name string, newImport importSpec) {
	if val, found := std[name]; found {
		if val.TypeChecking == true {
			std[name] = newImport
		}
	} else {
		std[name] = newImport
	}
}

func stdImports(uses func(name string) (bool, bool)) map[string]importSpec {
	std := make(map[string]importSpec)
	std["collections"] = importSpec{Module: "collections.abc", TypeChecking: true}
	add := func(name, module string) {
		if use, typeChecking := uses(name); use {
			typeCheckingOverwriteProtection(std, module, importSpec{Module: module, TypeChecking: typeChecking})
		}
	}

	add("decimal.Decimal", "decimal")

	add("datetime.date", "datetime")
	add("datetime.time", "datetime")
	add("datetime.datetime", "datetime")
	add("datetime.timedelta", "datetime")

	add("uuid.UUID", "uuid")
	return std
}
