package core

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

func ModelName(name string) string {
	out := ""
	for _, p := range strings.Split(name, "_") {
		out += cases.Title(language.Und, cases.NoLower).String(p)
	}
	return out
}
