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

func EscapedColumnName(pluginColumn *plugin.Column, pos int) string {
	return Escape(ColumnName(pluginColumn, pos))
}

// ModelName builds the class name for a table. Singularization runs on the
// raw snake_case table name (before camel-casing) so that
// inflection_exclude_table_names entries, which users write in snake_case,
// match correctly.
func ModelName(config *config.Config, modelName string, schemaName string) string {
	if !config.EmitExactTableNames {
		modelName = Singular(SingularParams{
			Name:       modelName,
			Exclusions: config.InflectionExcludeTableNames,
		})
	}

	return qualifiedClassName(config, modelName, schemaName)
}

// EnumName builds the class name for a SQL enum. Enum type names are never
// singularized — they are type names, not table names.
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
	if r, _ := utf8.DecodeRuneInString(name); unicode.IsDigit(r) {
		name = "_" + name
	}

	seen[name]++
	if seen[name] > 1 {
		name = fmt.Sprintf("%s_%d", name, seen[name])
		// Reserve the suffixed name so a literal collision later gets its own suffix.
		seen[name]++
	}

	return name
}

func ParamName(p *plugin.Parameter) string {
	var name string
	if p.Column.GetName() != "" {
		name = p.Column.Name
	} else {
		name = fmt.Sprintf("dollar_%d", p.GetNumber())
	}

	return Escape(name)
}
