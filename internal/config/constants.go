package config

import "fmt"

const PluginVersion = "v0.6.0"

type (
	SQLDriver           string
	DocstringConvention string
	ModelType           string
)

func (dr SQLDriver) String() string {
	return string(dr)
}

const (
	SQLDriverSQLite    SQLDriver = "sqlite3"
	SQLDriverAioSQLite SQLDriver = "aiosqlite"
	SQLDriverAsyncpg   SQLDriver = "asyncpg"
)

const (
	ModelTypeDataclass ModelType = "dataclass"
	ModelTypeAttrs     ModelType = "attrs"
	ModelTypeMsgspec   ModelType = "msgspec"
	ModelTypePydantic  ModelType = "pydantic"
)

var driversEngine = map[SQLDriver]string{
	SQLDriverSQLite:    "sqlite",
	SQLDriverAioSQLite: "sqlite",
	SQLDriverAsyncpg:   "postgresql",
}

const (
	DocstringConventionNone   DocstringConvention = "none"
	DocstringConventionGoogle DocstringConvention = "google"
	DocstringConventionNumpy  DocstringConvention = "numpy"
	DocstringConventionPEP257 DocstringConvention = "pep257"
)

func (dr SQLDriver) Validate(engine string) error {
	val, found := driversEngine[dr]
	if !found {
		return fmt.Errorf("unknown SQL driver: %s", dr)
	}
	if val != engine {
		return fmt.Errorf("SQL driver %s does not support %s", dr, engine)
	}

	return nil
}

func (modelType ModelType) Valid() bool {
	switch modelType {
	case ModelTypeDataclass, ModelTypeMsgspec, ModelTypeAttrs, ModelTypePydantic:
		return true
	default:
		return false
	}
}

func (ds DocstringConvention) Valid() bool {
	switch ds {
	case DocstringConventionNone, DocstringConventionNumpy, DocstringConventionGoogle, DocstringConventionPEP257:
		return true
	default:
		return false
	}
}
