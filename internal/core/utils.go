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
