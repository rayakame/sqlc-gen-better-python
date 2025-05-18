package typeConversion

import "strings"

var conversions = map[string]struct{}{
	"boolean":   {},
	"bool":      {},
	"date":      {},
	"datetime":  {},
	"timestamp": {},
}

func SqliteDoTypeConversion(name string) bool {
	_, found := conversions[name]
	if found {
		return found
	} else if strings.HasPrefix(name, "decimal") {
		return true
	}
	return false
}

func SqliteGetConversions() map[string]struct{} {
	return conversions
}
