package typeConversion

import "strings"

func SqliteDoTypeConversion(name string) bool {
	_, found := map[string]struct{}{
		"boolean":   {},
		"bool":      {},
		"date":      {},
		"datetime":  {},
		"timestamp": {},
	}[name]
	if found {
		return found
	} else if strings.HasPrefix(name, "decimal") {
		return true
	}
	return false
}
