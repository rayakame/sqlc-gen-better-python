package driver

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

const sqliteResultType = "sqlite3.Row"

// sqliteBase contains shared logic between the sqlite3 and aiosqlite drivers.
type sqliteBase struct {
	moduleName string // "sqlite3" or "aiosqlite"
	rows       *RowBuilder
}

// newSqliteBase creates a shared sqlite base. The RowBuilder never converts
// inline (except overrides/enums): registered converters handle the raw values,
// see WriteConversionSetup.
func newSqliteBase(moduleName string) sqliteBase {
	return sqliteBase{
		moduleName: moduleName,
		rows:       newRowBuilder(func(string) bool { return false }),
	}
}

// SupportsCommand returns if the driver supports the command.
func (sb *sqliteBase) SupportsCommand(cmd string) bool {
	switch cmd {
	case metadata.CmdExec,
		metadata.CmdExecResult,
		metadata.CmdExecLastId,
		metadata.CmdExecRows,
		metadata.CmdOne,
		metadata.CmdMany:
		return true
	default:
		return false
	}
}

// TypeCheckingHook returns nil (no type-checking hook for sqlite drivers).
func (sb *sqliteBase) TypeCheckingHook() []string {
	return nil
}

// NeedsConversion reports whether a SQL type needs runtime conversion for sqlite.
func (sb *sqliteBase) NeedsConversion(sqlType string) bool {
	return sqliteNeedsConversion(sqlType)
}

// ConvertsInline always returns false: sqlite drivers convert via registered
// adapters/converters, not inline in decode code.
func (sb *sqliteBase) ConvertsInline(_ string) bool {
	return false
}

// sqliteConvSpec describes one adapter/converter pair for a Python type.
type sqliteConvSpec struct {
	suffix        string   // function name suffix, e.g. "date"
	pyType        string   // Python type to register the adapter for
	adaptRet      string   // adapter return annotation
	adaptBody     string   // adapter body expression
	convBody      string   // converter body expression
	speedupsBody  string   // converter body when speedups are enabled ("" = same as convBody)
	converterKeys []string // sqlite declared type names to register the converter under
}

// sqliteConvSpecs maps Python type names to their conversion spec.
var sqliteConvSpecs = map[string]sqliteConvSpec{
	"datetime.date": {
		suffix:        "date",
		pyType:        "datetime.date",
		adaptRet:      "str",
		adaptBody:     "val.isoformat()",
		convBody:      "datetime.date.fromisoformat(val.decode())",
		speedupsBody:  "ciso8601.parse_datetime(val.decode()).date()",
		converterKeys: []string{"date"},
	},
	"decimal.Decimal": {
		suffix:        "decimal",
		pyType:        "decimal.Decimal",
		adaptRet:      "str",
		adaptBody:     "str(val)",
		convBody:      "decimal.Decimal(val.decode())",
		speedupsBody:  "",
		converterKeys: []string{"decimal"},
	},
	"datetime.datetime": {
		suffix:        "datetime",
		pyType:        "datetime.datetime",
		adaptRet:      "str",
		adaptBody:     "val.isoformat()",
		convBody:      "datetime.datetime.fromisoformat(val.decode())",
		speedupsBody:  "ciso8601.parse_datetime(val.decode())",
		converterKeys: []string{"datetime", "timestamp"},
	},
	"bool": {
		suffix:        "bool",
		pyType:        "bool",
		adaptRet:      "int",
		adaptBody:     "int(val)",
		convBody:      "bool(int(val))",
		speedupsBody:  "",
		converterKeys: []string{"bool", "boolean"},
	},
	"memoryview": {
		suffix:        "memoryview",
		pyType:        "memoryview",
		adaptRet:      "bytes",
		adaptBody:     "val.tobytes()",
		convBody:      "memoryview(val)",
		speedupsBody:  "",
		converterKeys: []string{"blob"},
	},
}

// WriteConversionSetup writes the adapter/converter functions and their
// registrations for every conversion type used by the given queries.
// Values written by adapters and read back by converters require the user's
// connection to be opened with detect_types=sqlite3.PARSE_DECLTYPES.
func (sb *sqliteBase) WriteConversionSetup(body *writer.CodeWriter, config *config.Config, queries []model.Query) bool {
	usedTypes := SqliteConversionsUsed(queries)
	if len(usedTypes) == 0 {
		return false
	}

	adapters := make([]string, 0, len(usedTypes))
	converters := make([]string, 0, len(usedTypes))
	for _, pyType := range usedTypes {
		spec := sqliteConvSpecs[pyType]

		body.WriteLine(fmt.Sprintf("def _adapt_%s(val: %s) -> %s:", spec.suffix, spec.pyType, spec.adaptRet))
		body.WriteIndentedLine(1, "return "+spec.adaptBody)
		body.NNewLine(2)

		convBody := spec.convBody
		if config.Speedups && spec.speedupsBody != "" {
			convBody = spec.speedupsBody
		}
		body.WriteLine(fmt.Sprintf("def _convert_%s(val: bytes) -> %s:", spec.suffix, spec.pyType))
		body.WriteIndentedLine(1, "return "+convBody)
		body.NNewLine(2)

		adapters = append(adapters, fmt.Sprintf("%s.register_adapter(%s, _adapt_%s)", sb.moduleName, spec.pyType, spec.suffix))
		for _, key := range spec.converterKeys {
			converters = append(converters, fmt.Sprintf(`%s.register_converter("%s", _convert_%s)`, sb.moduleName, key, spec.suffix))
		}
	}

	for _, line := range adapters {
		body.WriteLine(line)
	}
	body.NewLine()
	for _, line := range converters {
		body.WriteLine(line)
	}

	return true
}

// writeSqliteCall writes stmtHead+argsSegment+stmtTail on one line, hoisting a
// too-long parameter tuple into a local _args variable first so the statement
// stays within the line limit.
func writeSqliteCall(body *writer.CodeWriter, indent int, query model.Query, stmtHead, stmtTail string) {
	parts := expandParams(query)
	segment := ""
	switch {
	case len(parts) == 1:
		segment = fmt.Sprintf(", (%s,)", parts[0])
	case len(parts) > 1:
		segment = fmt.Sprintf(", (%s)", strings.Join(parts, ", "))
	}

	stmt := stmtHead + segment + stmtTail
	if body.FitsLine(indent, stmt) {
		body.WriteIndentedLine(indent, stmt)

		return
	}

	body.WriteIndentedLine(indent, "sql_args = (")
	for _, part := range parts {
		body.WriteIndentedLine(indent+1, part+",")
	}
	body.WriteIndentedLine(indent, ")")
	body.WriteIndentedLine(indent, stmtHead+", sql_args"+stmtTail)
}
