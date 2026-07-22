package driver

import (
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
)

const (
	decodeRowsExpr        = "return [self._decode_hook(row) for row in result]"
	queryResultsClassName = "QueryResults"
)

type Driver interface {
	// Name returns the Python module name (e.g., "asyncpg", "aiosqlite", "sqlite3").
	Name() string

	// ConnType returns the Python type annotation for the connection parameter.
	ConnType() string

	// SupportsCommand returns if the driver supports the command.
	SupportsCommand(cmd string) bool

	// IsAsync reports whether this driver uses async/await.
	IsAsync() bool

	// NeedsConversion reports whether a SQL type needs explicit Python-side conversion,
	// meaning the type's module must be imported at runtime (not TYPE_CHECKING-only).
	NeedsConversion(sqlType string) bool

	// ConvertsInline reports whether values of this SQL type are converted inline in
	// generated decode code. Drivers that register converters instead return false.
	// Must stay in sync with the RowBuilder's conversion check.
	ConvertsInline(sqlType string) bool

	// WriteConversionSetup writes module-level type conversion setup (e.g. sqlite
	// adapter/converter registration) and reports whether anything was written.
	WriteConversionSetup(body *writer.CodeWriter, config *config.Config, queries []model.Query) bool

	// WriteQueryFunc writes the Python function body for a single query.
	WriteQueryFunc(body *writer.CodeWriter, config *config.Config, query model.Query, indent int)

	// WriteQueryResultsClass writes the QueryResults helper class for :many queries.
	// Returns the class name (typically "QueryResults").
	WriteQueryResultsClass(w *writer.CodeWriter) string

	// TypeCheckingHook returns additional lines to emit in the TYPE_CHECKING
	// block. The lines must also be runtime-safe: with omit_typechecking_block
	// they are emitted at module level and actually execute.
	TypeCheckingHook() []string
}

func New(conf *config.Config) (Driver, error) {
	switch conf.SqlDriver {
	case config.SQLDriverAsyncpg:
		return newAsyncpgDriver(), nil
	case config.SQLDriverPsycopgAsync:
		return newPsycopgDriver(true), nil
	case config.SQLDriverAioSQLite:
		return newSqliteDriver("aiosqlite", true), nil
	case config.SQLDriverSQLite:
		return newSqliteDriver("sqlite3", false), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", conf.SqlDriver)
	}
}
