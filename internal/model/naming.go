package model

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func SnakeToCamel(conf *config.Config, s string) string {
	out := ""
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return r
		}
		if unicode.IsDigit(r) {
			return r
		}

		return rune('_')
	}, s)
	for _, p := range strings.Split(s, "_") {
		if _, found := conf.InitialismsMap[p]; found {
			out += strings.ToUpper(p)
		} else {
			out += cases.Title(language.Und, cases.NoLower).String(p)
		}
	}
	// A Model prefix, not an underscore: pyright strict flags references to
	// leading-underscore class names as private. IsReserved catches the only
	// CapWords-shaped keywords (True/False/None).
	r, _ := utf8.DecodeRuneInString(out)
	switch {
	case out == "", unicode.IsDigit(r), IsReserved(out):
		return "Model" + out
	default:
		return out
	}
}

func UpperSnakeCase(s string) string {
	result := ""
	for i, r := range s {
		if unicode.IsUpper(r) && i != 0 {
			result += "_" + string(r)
		} else {
			result += string(r)
		}
	}
	result = strings.ToUpper(result)

	return result
}

func ColumnName(pluginColumn *plugin.Column, pos int) string {
	if pluginColumn.Name != "" {
		return pluginColumn.Name
	}

	return fmt.Sprintf("column_%d", pos+1)
}

func sanitizeIdentifier(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return r
		}

		return '_'
	}, name)
}

func EscapedColumnName(pluginColumn *plugin.Column, pos int) string {
	name := sanitizeIdentifier(ColumnName(pluginColumn, pos))
	// attrs strips leading underscores from init params and pydantic treats
	// such fields as private, so they get the column_ prefix like digits.
	if r, _ := utf8.DecodeRuneInString(name); unicode.IsDigit(r) || r == '_' {
		name = "column_" + name
	}

	return Escape(name)
}

// ModelName builds the class name for a table. Singularization runs on the
// raw snake_case table name (before camel-casing) so that
// inflection_exclude_table_names entries, which users write in snake_case,
// match correctly. For non-default-schema tables exclusions match BOTH the
// bare table name ("events") and the schema-qualified form
// ("analytics_events") - v0.4.x singularized the qualified string, so
// existing configs use the qualified spelling.
func ModelName(config *config.Config, modelName string, schemaName string) string {
	if !config.EmitExactTableNames && !inflectionExcluded(config, modelName, schemaName) {
		modelName = Singular(SingularParams{
			Name:       modelName,
			Exclusions: nil,
		})
	}

	return qualifiedClassName(config, modelName, schemaName)
}

// inflectionExcluded reports whether the table name is excluded from
// singularization, matching the bare or schema-qualified form.
func inflectionExcluded(config *config.Config, modelName, schemaName string) bool {
	qualified := modelName
	if schemaName != "" {
		qualified = schemaName + "_" + modelName
	}
	for _, exclusion := range config.InflectionExcludeTableNames {
		if strings.EqualFold(exclusion, modelName) || strings.EqualFold(exclusion, qualified) {
			return true
		}
	}

	return false
}

// EnumName builds the class name for a SQL enum. Enum type names are never
// singularized - they are type names, not table names.
func EnumName(config *config.Config, enumName string, schemaName string) string {
	return qualifiedClassName(config, enumName, schemaName)
}

func qualifiedClassName(config *config.Config, name, schemaName string) string {
	if schemaName != "" {
		name = schemaName + "_" + name
	}

	return SnakeToCamel(config, name)
}

// EnumConstantName converts an enum value into a valid, unique Python constant
// name: non-alphanumeric characters become underscores, empty results fall
// back to VALUE_N, digit-leading names get an underscore prefix, and
// duplicates get a numeric suffix. seen tracks names across one enum.
func EnumConstantName(value string, index int, seen map[string]int) string {
	name := strings.ToUpper(sanitizeIdentifier(value))

	if strings.Trim(name, "_") == "" {
		name = fmt.Sprintf("VALUE_%d", index+1)
	} else if r, _ := utf8.DecodeRuneInString(name); unicode.IsDigit(r) || r == '_' {
		// pyright strict flags references to leading-underscore members as
		// private; at runtime they would work fine.
		name = "VALUE_" + name
	}

	return DedupName(name, seen)
}

// DedupName makes repeated Python identifiers unique by appending a numeric
// suffix ("name", "name_2", "name_3", ...). Suffixes are probed until an
// unused identifier is found, so a literal "name_2" that appeared earlier can
// never collide with a generated one. seen tracks usage counts per scope.
func DedupName(name string, seen map[string]int) string {
	return dedup(name, "%s_%d", seen)
}

// DedupClassName is DedupName with a bare digit suffix ("Name", "Name2",
// ...): an underscore suffix would violate the CapWords convention (N801).
func DedupClassName(name string, seen map[string]int) string {
	return dedup(name, "%s%d", seen)
}

func dedup(name, format string, seen map[string]int) string {
	seen[name]++
	if seen[name] == 1 {
		return name
	}
	for i := seen[name]; ; i++ {
		candidate := fmt.Sprintf(format, name, i)
		if seen[candidate] == 0 {
			seen[candidate]++

			return candidate
		}
	}
}

// ParamName allows leading underscores: parameters are plain kwargs, never
// attrs/pydantic fields, so only digit-leading names need the prefix.
func ParamName(p *plugin.Parameter) string {
	name := sanitizeIdentifier(p.Column.GetName())
	if r, _ := utf8.DecodeRuneInString(name); name == "" {
		name = fmt.Sprintf("dollar_%d", p.GetNumber())
	} else if unicode.IsDigit(r) {
		name = "column_" + name
	}

	return Escape(name)
}
