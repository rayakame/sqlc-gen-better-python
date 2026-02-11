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

func ModelName(conf *config.Config, modelName string, schemaName string) string {
	name := ""
	if schemaName != "" {
		name += schemaName + "_"
	}
	name += modelName

	return SnakeToCamel(conf, name)
}
