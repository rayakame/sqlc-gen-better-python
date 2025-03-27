package core

import (
	"fmt"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ModelName(enumName string, schemaName string, conf *Config) string {
	if schemaName != "" {
		enumName = schemaName + "_" + enumName
	}
	return SnakeToCamel(enumName, conf)
}

func SnakeToCamel(s string, conf *Config) string {
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

func ColumnName(c *plugin.Column, pos int) string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("column_%d", pos+1)
}

func ParamName(p *plugin.Parameter) string {
	if p.Column.Name != "" {
		return ArgName(p.Column.Name)
	}
	return fmt.Sprintf("dollar_%d", p.Number)
}

func ArgName(name string) string {
	out := ""
	for _, p := range strings.Split(name, "_") {
		out += strings.ToLower(p)
	}
	return out
}

func ExtractImport(pyType PyType) []string {
	imports := make([]string, 0)
	if pyType.IsNullable {
		imports = append(imports, "import typing")
	}
	parts := strings.Split(pyType.Type, ".")
	if len(parts) == 1 {
		return imports
	}
	return append(imports, "import "+parts[0])
}

func AppendUniqueString(list []string, newItems []string) []string {
	seen := make(map[string]struct{}, len(list))

	// Bestehende Elemente merken
	for _, item := range list {
		seen[item] = struct{}{}
	}

	// Neue Elemente nur hinzufügen, wenn sie nicht existieren
	for _, item := range newItems {
		if _, exists := seen[item]; !exists {
			list = append(list, item)
			seen[item] = struct{}{}
		}
	}
	return list
}
