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
		return p.Column.Name
	}
	return fmt.Sprintf("dollar_%d", p.Number)
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

func SQLToPyFileName(s string) string {
	return strings.ReplaceAll(s, ".sql", ".py")
}
