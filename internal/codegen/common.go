package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/drivers"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
)

type TypeBuildPyQueryFunc func(*core.Query, *builders.IndentStringBuilder, []core.FunctionArg, core.PyType, *core.Config) error
type TypeAcceptedDriverCMDs func() []string
type TypeDriverTypeCheckingHook func() []string
type TypeDriverBuildQueryResults func(*builders.IndentStringBuilder) string

func defaultDriverTypeCheckingHook() []string {
	return nil
}
func defaultDriverBuildQueryResults(_ *builders.IndentStringBuilder) string {
	return ""
}

type Driver struct {
	conf *core.Config

	connType           string
	buildPyQueryFunc   TypeBuildPyQueryFunc
	acceptedDriverCMDs TypeAcceptedDriverCMDs

	driverTypeCheckingHook  TypeDriverTypeCheckingHook
	driverBuildQueryResults TypeDriverBuildQueryResults

	//BuildPyQueriesFiles(*core.Importer, []core.Query) ([]*plugin.File, error)
}

func NewDriver(conf *core.Config) (*Driver, error) {
	var buildPyQueryFunc TypeBuildPyQueryFunc
	var acceptedDriverCMDs TypeAcceptedDriverCMDs
	var connType string
	var driverTypeCheckingHook TypeDriverTypeCheckingHook = defaultDriverTypeCheckingHook
	var driverBuildQueryResults TypeDriverBuildQueryResults = defaultDriverBuildQueryResults
	switch conf.SqlDriver {
	case core.SQLDriverAioSQLite:
		buildPyQueryFunc = drivers.AioSQLiteBuildPyQueryFunc
		acceptedDriverCMDs = drivers.AioSQLiteAcceptedDriverCMDs
		connType = drivers.AioSQLiteConn
		driverBuildQueryResults = drivers.AiosqliteBuildQueryResults
	case core.SQLDriverSQLite:
		buildPyQueryFunc = drivers.SQLite3BuildPyQueryFunc
		acceptedDriverCMDs = drivers.SQLite3AcceptedDriverCMDs
		connType = drivers.SQLite3Conn
		driverBuildQueryResults = drivers.SQLite3BuildQueryResults
	case core.SQLDriverAsyncpg:
		buildPyQueryFunc = drivers.AsyncpgBuildPyQueryFunc
		acceptedDriverCMDs = drivers.AsyncpgAcceptedDriverCMDs
		connType = drivers.AsyncpgConn
		driverTypeCheckingHook = drivers.AsyncpgTypeCheckingHook
		driverBuildQueryResults = drivers.AsyncpgBuildQueryResults
	default:
		return nil, fmt.Errorf("unsupported driver: %s", conf.SqlDriver.String())
	}
	builders.SetDocstringConfig(conf.EmitDocstrings, conf.EmitDocstringsSQL, conf.SqlDriver)

	return &Driver{
		buildPyQueryFunc:        buildPyQueryFunc,
		acceptedDriverCMDs:      acceptedDriverCMDs,
		conf:                    conf,
		connType:                connType,
		driverTypeCheckingHook:  driverTypeCheckingHook,
		driverBuildQueryResults: driverBuildQueryResults,
	}, nil
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
