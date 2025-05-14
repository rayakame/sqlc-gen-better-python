package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/drivers"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
)

type TypeBuildPyQueryFunc func(*core.Query, *builders.IndentStringBuilder, []core.FunctionArg, core.PyType, bool) error
type TypeAcceptedDriverCMDs func() []string

type Driver struct {
	conf *core.Config

	connType           string
	buildPyQueryFunc   TypeBuildPyQueryFunc
	acceptedDriverCMDs TypeAcceptedDriverCMDs

	//BuildPyQueriesFiles(*core.Importer, []core.Query) ([]*plugin.File, error)
}

func NewDriver(conf *core.Config) (*Driver, error) {
	var buildPyQueryFunc TypeBuildPyQueryFunc
	var acceptedDriverCMDs TypeAcceptedDriverCMDs
	var connType string
	switch conf.SqlDriver {
	case core.SQLDriverAioSQLite:
		buildPyQueryFunc = drivers.AioSQLiteBuildPyQueryFunc
		acceptedDriverCMDs = drivers.AioSQLiteAcceptedDriverCMDs
		connType = drivers.AioSQLiteConn
	case core.SQLDriverSQLite:
		buildPyQueryFunc = drivers.SQLite3BuildPyQueryFunc
		acceptedDriverCMDs = drivers.SQLite3AcceptedDriverCMDs
		connType = drivers.SQLite3Conn
	case core.SQLDriverAsyncpg:
		buildPyQueryFunc = drivers.AsyncpgBuildPyQueryFunc
		acceptedDriverCMDs = drivers.AsyncpgAcceptedDriverCMDs
		connType = drivers.AsyncpgConn
	default:
		return nil, fmt.Errorf("unsupported driver: %s", conf.SqlDriver.String())
	}
	builders.SetDocstringConfig(conf.EmitDocstrings)

	return &Driver{buildPyQueryFunc: buildPyQueryFunc, acceptedDriverCMDs: acceptedDriverCMDs, conf: conf, connType: connType}, nil
}

func (dr *Driver) supportedCMD(command string) error {
	cmds := dr.acceptedDriverCMDs()
	for _, cmd := range cmds {
		if cmd == command {
			return nil
		}
	}
	return fmt.Errorf("unsupported command for selected driver: %s", command)
}
