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
	r, _ := utf8.DecodeRuneInString(out)
	if unicode.IsDigit(r) {
		return "_" + out
	} else {
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

// sanitizePyIdentifier maps invalid runes to "_" and returns "" when nothing
// usable remains. Digit-leading results get digitPrefix, not "_": attrs and
// pydantic treat leading-underscore fields specially.
func sanitizePyIdentifier(name, digitPrefix string) string {
	var builder strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	sanitized := builder.String()
	if strings.Trim(sanitized, "_") == "" {
		return ""
	}
	if r, _ := utf8.DecodeRuneInString(sanitized); unicode.IsDigit(r) {
		return digitPrefix + "_" + sanitized
	}

	return sanitized
}

func EscapedColumnName(pluginColumn *plugin.Column, pos int) string {
	name := sanitizePyIdentifier(ColumnName(pluginColumn, pos), "column")
	if name == "" {
		name = fmt.Sprintf("column_%d", pos+1)
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
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToUpper(r))
		} else {
			builder.WriteRune('_')
		}
	}
	name := builder.String()

	if strings.Trim(name, "_") == "" {
		name = fmt.Sprintf("VALUE_%d", index+1)
	}
	// A VALUE_ prefix, not an underscore: enum treats leading-underscore
	// names as private attributes instead of members.
	if r, _ := utf8.DecodeRuneInString(name); unicode.IsDigit(r) {
		name = "VALUE_" + name
	}

	return DedupName(name, seen)
}

// DedupName makes repeated Python identifiers unique by appending a numeric
// suffix ("name", "name_2", "name_3", ...). Suffixes are probed until an
// unused identifier is found, so a literal "name_2" that appeared earlier can
// never collide with a generated one. seen tracks usage counts per scope.
func DedupName(name string, seen map[string]int) string {
	seen[name]++
	if seen[name] == 1 {
		return name
	}
	for i := seen[name]; ; i++ {
		candidate := fmt.Sprintf("%s_%d", name, i)
		if seen[candidate] == 0 {
			seen[candidate]++

			return candidate
		}
	}
}

func ParamName(p *plugin.Parameter) string {
	name := sanitizePyIdentifier(p.Column.GetName(), "arg")
	if name == "" {
		name = fmt.Sprintf("dollar_%d", p.GetNumber())
	}

	return Escape(name)
}
