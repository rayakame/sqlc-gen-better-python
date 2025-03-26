package core

import "fmt"

const (
	SQLDriverSQLite    = "sqlite"
	SQLDriverAioSQLite = "aiosqlite"
)

const (
	ModelTypeDataclass = "dataclass"
	ModelTypeAttrs     = "attrs"
)

var asyncDrivers = map[string]bool{
	string(SQLDriverSQLite):    false,
	string(SQLDriverAioSQLite): true,
}
var validModelTypes = map[string]struct{}{
	string(ModelTypeDataclass): {},
	string(ModelTypeAttrs):     {},
}

func isDriverAsync(sqlDriver string) (bool, error) {
	val, found := asyncDrivers[sqlDriver]
	if !found {
		return false, fmt.Errorf("unknown SQL driver: %s", sqlDriver)
	}
	return val, nil
}

func isModelTypeValid(modelType string) error {
	if _, found := validModelTypes[modelType]; !found {
		return fmt.Errorf("unknown model type: %s", modelType)
	}
	return nil
}
