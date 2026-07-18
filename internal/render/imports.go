package render

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/driver"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

type ImportResult struct {
	Std          []string // Standard library imports (e.g., "import typing").
	TypeChecking []string // Imports inside "if TYPE_CHECKING:" block.
	Package      []string // Local package imports (e.g., "from mypackage import models").
}

func (r ImportResult) Write(body *writer.CodeWriter, omitTypeChecking bool, typeCheckingLines []string) {
	for _, line := range r.Std {
		body.WriteLine(line)
	}
	if len(r.Std) != 0 && len(r.TypeChecking) != 0 {
		body.NewLine()
	}
	indentLevel := 0
	if (len(r.TypeChecking) != 0 || len(typeCheckingLines) != 0) && !omitTypeChecking {
		body.WriteLine("if typing.TYPE_CHECKING:")
		indentLevel = 1
	}
	for _, line := range r.TypeChecking {
		body.WriteIndentedLine(indentLevel, line)
	}
	for i, line := range typeCheckingLines {
		if i == 0 && len(r.TypeChecking) != 0 {
			body.NewLine()
		}
		body.WriteIndentedLine(indentLevel, line)
	}
	for i, line := range r.Package {
		if i == 0 {
			body.NewLine()
		}
		body.WriteLine(line)
	}
}

// ImportResolver computes Python import statements for generated files.
// It is stateless — all data is passed as arguments.
type ImportResolver struct {
	conf *config.Config
	drv  driver.Driver
}

// NewImportResolver creates a new ImportResolver.
func NewImportResolver(conf *config.Config, drv driver.Driver) *ImportResolver {
	return &ImportResolver{conf: conf, drv: drv}
}

type importSpec struct {
	Module       string
	Name         string
	Alias        string
	TypeChecking bool
}

func (s importSpec) String() string {
	if s.Alias != "" {
		if s.Name == "" {
			return fmt.Sprintf("import %s as %s", s.Module, s.Alias)
		}

		return fmt.Sprintf("from %s import %s as %s", s.Module, s.Name, s.Alias)
	}
	if s.Name == "" {
		return "import " + s.Module
	}

	return fmt.Sprintf("from %s import %s", s.Module, s.Name)
}

func (r *ImportResolver) ModelImports(tables []model.Table) ImportResult {
	// "uses" checks whether any table column has a given Python type.
	uses := func(name string) (bool, bool) {
		for _, table := range tables {
			for _, col := range table.Columns {
				if col.Type.Type == name {
					return true, true
				}
			}
		}

		return false, false
	}

	// Scan enum/list usage in a dedicated pass: the uses closure early-returns
	// on the first type match, so side-effect flags inside it would miss
	// columns positioned after a match.
	usesEnum := false
	hasList := false
	for _, table := range tables {
		for _, col := range table.Columns {
			if col.Type.IsEnum {
				usesEnum = true
			}
			if col.Type.IsList {
				hasList = true
			}
		}
	}

	std := r.stdImports(uses)
	r.addOverrideImports(std, uses)
	r.forcePydanticRuntimeImports(std, hasList)

	std, typeChecking := splitTypeChecking(std)

	r.addModelImport(std)

	local := make(map[string]importSpec)
	if usesEnum {
		if r.conf.ModelType == config.ModelTypePydantic {
			// pydantic evaluates field annotations when building schemas.
			local["enum"] = importSpec{Module: r.conf.Package, Name: "enums", Alias: "", TypeChecking: false}
		} else {
			typeChecking["enums"] = importSpec{Module: r.conf.Package, Name: "enums", Alias: "", TypeChecking: true}
		}
	}
	if r.conf.ModelType == config.ModelTypePydantic && len(typeChecking) == 0 {
		// Without a TYPE_CHECKING block, typing itself is unused in models.py.
		delete(std, "typing")
	}

	return buildResult(std, typeChecking, local)
}

func (r *ImportResolver) QueryImports(queries []model.Query) ImportResult {
	addCiso := false

	// "uses" checks whether any query arg/return uses a given Python type.
	// Returns (isUsed, goesInTypeChecking).
	uses := func(name string) (bool, bool) {
		var bestUsed, bestTC *bool

		update := func(used, tc bool) {
			if bestUsed == nil {
				bestUsed = &used
				bestTC = &tc
			} else if *bestTC {
				// Runtime import (tc=false) takes priority over TYPE_CHECKING.
				*bestUsed = used
				*bestTC = tc
			}
		}

		for _, query := range queries {
			if used, tc := r.queryValueUses(name, query.Returns); used {
				// Scalar :many returns subscript QueryResults[T] at runtime,
				// so the type itself must be imported at runtime. Struct
				// returns only reference the type inside annotations.
				if query.Cmd == metadata.CmdMany && !query.Returns.IsStruct() {
					update(used, false)
				}
				update(used, tc)
				// The datetime import stays at runtime even with speedups:
				// register_adapter(datetime.date, ...) needs it at import time.
				// ciso8601 is only used inside the converter bodies.
				if !tc && r.isSpeedupDatetime(name) {
					addCiso = true
				}
			}
			for _, arg := range query.Params {
				if used, tc := r.queryValueUses(name, arg); used {
					update(used, tc)
					if !tc && r.isSpeedupDatetime(name) {
						addCiso = true
					}
				}
				// Overridden params are converted back to their DefaultType at
				// runtime (e.g. "decimal.Decimal(params.rating)"), so that
				// type's module must be imported at runtime too.
				if overrideDefaultTypeUses(name, arg) {
					update(true, false)
				}
			}
		}

		if bestUsed == nil {
			return false, false
		}
		return *bestUsed, *bestTC
	}

	hasList := anyQueryType(queries, func(typ model.PyType) bool { return typ.IsList })

	std := r.stdImports(uses)
	r.addOverrideImports(std, uses)
	r.forcePydanticRuntimeImports(std, hasList)

	std, typeChecking := splitTypeChecking(std)
	r.addDriverImports(std, typeChecking, queries)

	if addCiso {
		std["ciso8601"] = importSpec{Module: "ciso8601"}
	}

	// Model import if any query emits a struct or uses copyfrom.
	for _, query := range queries {
		if (query.EmitsTable()) || query.Cmd == metadata.CmdCopyFrom {
			r.addModelImport(std)
			break
		}
	}

	// Only import models/enums when THIS module's queries actually reference
	// them — a global flag would emit unused imports in multi-file projects.
	local := map[string]importSpec{}
	if anyQueryType(queries, func(typ model.PyType) bool { return strings.HasPrefix(typ.Type, "models.") }) {
		local["models"] = importSpec{Module: r.conf.Package, Name: "models"}
	}
	if anyQueryType(queries, func(typ model.PyType) bool { return typ.IsEnum || strings.HasPrefix(typ.Type, "enums.") }) {
		local["enums"] = importSpec{Module: r.conf.Package, Name: "enums"}
	}

	return r.buildQueryResult(std, typeChecking, local, queries)
}

// anyQueryType reports whether pred matches any Python type used by the
// queries: scalar params/returns, row/params class columns, embed field types,
// and embed columns.
func anyQueryType(queries []model.Query, pred func(model.PyType) bool) bool {
	checkValue := func(qv model.QueryValue) bool {
		if qv.IsEmpty() {
			return false
		}
		if pred(qv.Type) {
			return true
		}
		if qv.Table == nil {
			return false
		}
		for _, col := range qv.Table.Columns {
			if pred(col.Type) {
				return true
			}
			if col.Embed != nil {
				for _, embedColumn := range col.Embed.Columns {
					if pred(embedColumn.Type) {
						return true
					}
				}
			}
		}

		return false
	}
	for _, query := range queries {
		if checkValue(query.Returns) {
			return true
		}
		for _, param := range query.Params {
			if checkValue(param) {
				return true
			}
		}
	}

	return false
}

func (r *ImportResolver) EnumImports() ImportResult {
	uses := func(name string) (bool, bool) {
		return false, false
	}
	std := r.stdImports(uses)
	std["enum"] = importSpec{Module: "enum", Name: "", Alias: "", TypeChecking: false}
	std, typeChecking := splitTypeChecking(std)

	return buildResult(std, typeChecking, nil)
}

// isSpeedupDatetime checks if speedups are enabled and this is a datetime type
// on a sqlite driver (where ciso8601 replaces datetime at runtime).
func (r *ImportResolver) isSpeedupDatetime(name string) bool {
	if !r.conf.Speedups {
		return false
	}
	isSqlite := r.conf.SqlDriver == config.SQLDriverAioSQLite || r.conf.SqlDriver == config.SQLDriverSQLite
	isDatetime := name == "datetime.datetime" || name == "datetime.date"
	return isSqlite && isDatetime
}

// forcePydanticRuntimeImports moves every type import to runtime for pydantic
// models: pydantic resolves field annotations when building the model schema,
// so TYPE_CHECKING-only imports would rely on pydantic's (version-dependent)
// TYPE_CHECKING-block resolution instead of plain module imports.
// collections.abc is only needed at runtime when a list field exists.
func (r *ImportResolver) forcePydanticRuntimeImports(std map[string]importSpec, hasListColumns bool) {
	if r.conf.ModelType != config.ModelTypePydantic {
		return
	}
	for key, spec := range std {
		if key == "collections" && !hasListColumns {
			continue
		}
		spec.TypeChecking = false
		std[key] = spec
	}
}

// overrideDefaultTypeUses reports whether the query value contains an overridden
// type whose DefaultType is `name` — those are converted back to DefaultType at
// runtime when passed to the driver (see driver.convertParamExpr).
func overrideDefaultTypeUses(name string, qv model.QueryValue) bool {
	if qv.IsEmpty() {
		return false
	}
	if qv.IsStruct() {
		for _, col := range qv.Table.Columns {
			if col.Type.DoOverride() && col.Type.DefaultType == name {
				return true
			}
		}

		return false
	}

	return qv.Type.DoOverride() && qv.Type.DefaultType == name
}

// addOverrideImports adds imports contributed by configured type overrides.
func (r *ImportResolver) addOverrideImports(std map[string]importSpec, uses func(string) (bool, bool)) {
	for _, override := range r.conf.Overrides {
		if override.PyType.Type == "" || override.PyType.Import == "" {
			continue
		}
		if used, tc := uses(override.PyType.Type); used {
			addWithPriority(std, override.PyType.Type, importSpec{
				Module: override.PyType.Import, Name: override.PyType.Package, Alias: "", TypeChecking: tc,
			})
		}
	}
}

// addDriverImports adds driver-specific imports to the std/typeChecking maps.
func (r *ImportResolver) addDriverImports(std, typeChecking map[string]importSpec, queries []model.Query) {
	driverName := string(r.conf.SqlDriver)
	hasMany := isAnyQueryMany(queries)

	switch r.conf.SqlDriver {
	case config.SQLDriverAsyncpg:
		typeChecking[driverName] = importSpec{Module: driverName}
		if hasMany {
			typeChecking[driverName+".cursor"] = importSpec{Module: driverName + ".cursor"}
			if r.hasSimpleReturn(queries) {
				std["operator"] = importSpec{Module: "operator"}
			}
		}

	case config.SQLDriverAioSQLite:
		// register_adapter/register_converter calls need the module at runtime.
		if len(driver.SqliteConversionsUsed(queries)) > 0 {
			std[driverName] = importSpec{Module: driverName}
		} else {
			typeChecking[driverName] = importSpec{Module: driverName}
		}
		if hasMany {
			typeChecking["sqlite3"] = importSpec{Module: "sqlite3"}
			if r.hasSimpleReturn(queries) {
				std["operator"] = importSpec{Module: "operator"}
			}
		}

	case config.SQLDriverSQLite:
		if len(driver.SqliteConversionsUsed(queries)) > 0 {
			std[driverName] = importSpec{Module: driverName}
		} else {
			typeChecking[driverName] = importSpec{Module: driverName}
		}
		if hasMany && r.hasSimpleReturn(queries) {
			std["operator"] = importSpec{Module: "operator"}
		}
	}
}

// hasSimpleReturn checks if any query has a non-struct return that doesn't need
// conversion. Must mirror RowBuilder.columnNeedsConversion: only these returns
// use operator.itemgetter instead of a _decode_hook.
func (r *ImportResolver) hasSimpleReturn(queries []model.Query) bool {
	for _, query := range queries {
		if query.Cmd != metadata.CmdMany {
			continue
		}
		if query.Returns.IsStruct() || query.Returns.Type.IsEnum || query.Returns.Type.DoOverride() {
			continue
		}
		if !r.drv.ConvertsInline(query.Returns.Type.SQLType) {
			return true
		}
	}
	return false
}

func (r *ImportResolver) queryValueUses(name string, queryValue model.QueryValue) (bool, bool) {
	if queryValue.IsEmpty() {
		return false, false
	}

	if queryValue.IsStruct() {
		// Scan ALL columns (including embed columns): any occurrence that
		// needs runtime conversion must force a runtime import, even when an
		// earlier annotation-only occurrence of the same type exists.
		used := false
		typeChecking := true
		check := func(typ model.PyType) {
			if typ.Type != name {
				return
			}
			used = true
			if r.drv.NeedsConversion(typ.SQLType) || typ.DoOverride() {
				typeChecking = false
			}
		}
		for _, column := range queryValue.Table.Columns {
			if column.Embed != nil {
				for _, embedColumn := range column.Embed.Columns {
					check(embedColumn.Type)
				}

				continue
			}
			check(column.Type)
		}
		if !used {
			return false, false
		}

		return true, typeChecking
	}

	if queryValue.Type.Type == name {
		needsConv := r.drv.NeedsConversion(queryValue.Type.SQLType) || queryValue.Type.DoOverride()
		return true, !needsConv
	}

	return false, false
}

func (r *ImportResolver) addModelImport(std map[string]importSpec) {
	switch r.conf.ModelType {
	case config.ModelTypeAttrs:
		std["attrs"] = importSpec{Module: "attrs", Name: "", Alias: "", TypeChecking: false}
	case config.ModelTypeDataclass:
		std["dataclasses"] = importSpec{Module: "dataclasses", Name: "", Alias: "", TypeChecking: false}
	case config.ModelTypeMsgspec:
		std["msgspec"] = importSpec{Module: "msgspec", Name: "", Alias: "", TypeChecking: false}
	case config.ModelTypePydantic:
		std["pydantic"] = importSpec{Module: "pydantic", Name: "", Alias: "", TypeChecking: false}
	}
}

func buildResult(std, typeChecking, local map[string]importSpec) ImportResult {
	return ImportResult{
		Std:          buildImportBlock(std),
		TypeChecking: buildImportBlock(typeChecking),
		Package:      buildImportBlock(local),
	}
}

// buildQueryResult is like buildResult but also appends QueryResultsArgsType.
func (r *ImportResolver) buildQueryResult(std, typeChecking, local map[string]importSpec, queries []model.Query) ImportResult {
	result := buildResult(std, typeChecking, local)

	if isAnyQueryMany(queries) {
		if len(result.TypeChecking) != 0 {
			result.TypeChecking[len(result.TypeChecking)-1] += "\n"
		}
		argsType := "QueryResultsArgsType: typing.TypeAlias = int | float | str | memoryview"
		allSpecs := mergeMaps(std, typeChecking)
		if _, ok := allSpecs["decimal"]; ok {
			argsType += " | decimal.Decimal"
		}
		if _, ok := allSpecs["uuid"]; ok {
			argsType += " | uuid.UUID"
		}
		if _, ok := allSpecs["datetime"]; ok {
			argsType += " | datetime.date | datetime.time | datetime.datetime | datetime.timedelta"
		}
		argsType += " | None"
		result.TypeChecking = append(result.TypeChecking, argsType)
	}

	return result
}

func buildImportBlock(specs map[string]importSpec) []string {
	if len(specs) == 0 {
		return nil
	}

	lines := make([]string, 0, len(specs))
	for _, spec := range specs {
		lines = append(lines, spec.String())
	}

	sort.Strings(lines)

	// Different specs can render to the same line (e.g. an override importing
	// a module the std scan also imports) — drop exact duplicates.
	return slices.Compact(lines)
}

// stdImports returns the base set of standard library imports.
// The uses function should return if the type is used and if it is only
// used for typechecking or not.
func (r *ImportResolver) stdImports(uses func(string) (bool, bool)) map[string]importSpec {
	std := map[string]importSpec{
		"collections": {Module: "collections.abc", TypeChecking: true, Name: "", Alias: ""},
		"typing":      {Module: "typing", TypeChecking: false, Name: "", Alias: ""},
	}

	// Check which standard types are used.
	for _, check := range []struct{ typeName, module string }{
		{"decimal.Decimal", "decimal"},
		{"datetime.date", "datetime"},
		{"datetime.time", "datetime"},
		{"datetime.datetime", "datetime"},
		{"datetime.timedelta", "datetime"},
		{"uuid.UUID", "uuid"},
	} {
		if used, tc := uses(check.typeName); used {
			addWithPriority(std, check.module, importSpec{Module: check.module, TypeChecking: tc, Name: "", Alias: ""})
		}
	}

	return std
}

// addWithPriority adds an import, but runtime imports (TypeChecking=false)
// take priority over TYPE_CHECKING imports.
func addWithPriority(m map[string]importSpec, key string, spec importSpec) {
	if existing, ok := m[key]; ok && !existing.TypeChecking {
		return // Existing runtime import has priority.
	}
	m[key] = spec
}

// splitTypeChecking separates imports into runtime and TYPE_CHECKING groups.
func splitTypeChecking(specs map[string]importSpec) (map[string]importSpec, map[string]importSpec) {
	runtime := make(map[string]importSpec)
	typeChecking := make(map[string]importSpec)
	for name, spec := range specs {
		if spec.TypeChecking {
			typeChecking[name] = spec
		} else {
			runtime[name] = spec
		}
	}

	return runtime, typeChecking
}

func mergeMaps(maps ...map[string]importSpec) map[string]importSpec {
	result := make(map[string]importSpec)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
