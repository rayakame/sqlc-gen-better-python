package core

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

func ModelName(enumName string, schemaName string) string {
	if schemaName != "" {
		enumName = schemaName + "_" + enumName
	}
	out := ""
	for _, p := range strings.Split(enumName, "_") {
		out += cases.Title(language.Und, cases.NoLower).String(p)
	}
	return out
}
