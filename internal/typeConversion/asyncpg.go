package typeConversion

func AsyncpgDoTypeConversion() map[string]struct{} {
	return map[string]struct{}{
		"bytea":            {},
		"blob":             {},
		"pg_catalog.bytea": {},
		"inet":             {},
		"cidr":             {},
	}
}
