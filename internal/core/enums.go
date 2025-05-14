package core

import "fmt"

type SQLDriverType string

func (dr *SQLDriverType) String() string {
	return string(*dr)
}

const (
	SQLDriverSQLite    SQLDriverType = "sqlite3"
	SQLDriverAioSQLite SQLDriverType = "aiosqlite"
	SQLDriverAsyncpg   SQLDriverType = "asyncpg"
)

const (
	ModelTypeDataclass = "dataclass"
	ModelTypeAttrs     = "attrs"
	ModelTypeMsgspec   = "msgspec"
)

var asyncDrivers = map[SQLDriverType]bool{
	SQLDriverSQLite:    false,
	SQLDriverAioSQLite: true,
	SQLDriverAsyncpg:   true,
}

var driversEngine = map[SQLDriverType]string{
	SQLDriverSQLite:    "sqlite",
	SQLDriverAioSQLite: "sqlite",
	SQLDriverAsyncpg:   "postgresql",
}

var validModelTypes = map[string]struct{}{
	string(ModelTypeDataclass): {},
	string(ModelTypeAttrs):     {},
	string(ModelTypeMsgspec):   {},
}

const (
	DocstringConventionGoogle = "google"
	DocstringConventionNumpy  = "numpy"
	DocstringConventionPEP257 = "pep257"
)

var validDocstringConventions = map[string]struct{}{
	DocstringConventionGoogle: {},
	DocstringConventionNumpy:  {},
	DocstringConventionPEP257: {},
}

func isDriverAsync(sqlDriver SQLDriverType) (bool, error) {
	val, found := asyncDrivers[sqlDriver]
	if !found {
		return false, fmt.Errorf("unknown SQL driver: %s", sqlDriver)
	}
	return val, nil
}

func isDriverValid(sqlDriver SQLDriverType, engine string) error {
	val, found := driversEngine[sqlDriver]
	if !found {
		return fmt.Errorf("unknown SQL driver: %s", sqlDriver)
	}
	if val != engine {
		return fmt.Errorf("SQL driver %s does not support %s", sqlDriver, engine)
	}
	return nil
}

func isModelTypeValid(modelType string) error {
	if _, found := validModelTypes[modelType]; !found {
		return fmt.Errorf("unknown model type: %s", modelType)
	}
	return nil
}

func isDocstringValid(ds *string) error {
	if ds == nil {
		return nil
	}
	if _, found := validDocstringConventions[*ds]; !found {
		return fmt.Errorf("unknown docstring convention: %s", ds)
	}
	return nil
}
