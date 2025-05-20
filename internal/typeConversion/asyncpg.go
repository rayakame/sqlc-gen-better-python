package typeConversion

func AsyncpgDoTypeConversion(name string) bool {
	_, found := map[string]struct{}{
		"bytea":            {},
		"blob":             {},
		"pg_catalog.bytea": {},
		"inet":             {},
		"cidr":             {},
	}[name]
	return found
}
