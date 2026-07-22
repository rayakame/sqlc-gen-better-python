package driver

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

const (
	asyncpgConnType   = "ConnectionLike"
	asyncpgResultType = "asyncpg.Record"
)

// AsyncpgDriver generates Python code for the asyncpg (async PostgreSQL) driver.
type AsyncpgDriver struct {
	rows *RowBuilder
}

func newAsyncpgDriver() *AsyncpgDriver {
	return &AsyncpgDriver{
		rows: newRowBuilder(asyncpgNeedsConversion),
	}
}

// Name returns "asyncpg".
func (d *AsyncpgDriver) Name() string { return "asyncpg" }

// ConnType returns "ConnectionLike".
func (d *AsyncpgDriver) ConnType() string { return asyncpgConnType }

// IsAsync returns true.
func (d *AsyncpgDriver) IsAsync() bool { return true }

// NeedsConversion reports whether a SQL type needs runtime conversion for asyncpg.
func (d *AsyncpgDriver) NeedsConversion(sqlType string) bool {
	return asyncpgNeedsConversion(sqlType)
}

// ConvertsInline reports whether a SQL type is converted inline; asyncpg converts
// everything inline (no registration mechanism).
func (d *AsyncpgDriver) ConvertsInline(sqlType string) bool {
	return asyncpgNeedsConversion(sqlType)
}

// WriteConversionSetup is a no-op for asyncpg.
func (d *AsyncpgDriver) WriteConversionSetup(_ *writer.CodeWriter, _ *config.Config, _ []model.Query) bool {
	return false
}

// TypeCheckingHook returns the ConnectionLike type alias. The PEP 695 form
// is lazy by design, which matters here: asyncpg.Connection[...] is a
// stub-only generic that raises TypeError when subscripted at runtime, and
// with omit_typechecking_block the alias is emitted at module level where it
// actually executes.
func (d *AsyncpgDriver) TypeCheckingHook() []string {
	return []string{
		fmt.Sprintf(
			"type ConnectionLike = asyncpg.Connection[%[1]s] | asyncpg.pool.PoolConnectionProxy[%[1]s]",
			asyncpgResultType,
		),
	}
}

// WriteQueryResultsClass writes the async QueryResults class for asyncpg :many queries.
func (d *AsyncpgDriver) WriteQueryResultsClass(body *writer.CodeWriter) string {
	body.QueryResults.WriteQueryResultsClassHeader(asyncpgConnType, []string{
		fmt.Sprintf("self._cursor: asyncpg.cursor.CursorFactory[%s] | None = None", asyncpgResultType),
		fmt.Sprintf("self._iterator: asyncpg.cursor.CursorIterator[%s] | None = None", asyncpgResultType),
	}, asyncpgResultType, d.IsAsync())
	body.QueryResults.WriteQueryResultsAwaitFunction([]string{
		"result = await self._conn.fetch(self._sql, *self._args)",
		decodeRowsExpr,
	})
	writeAsyncNextMethod(body, "an asyncpg cursor", "self._cursor = self._conn.cursor(self._sql, *self._args)")

	return queryResultsClassName
}

// SupportsCommand returns if the driver supports the command.
func (d *AsyncpgDriver) SupportsCommand(cmd string) bool {
	switch cmd {
	case metadata.CmdExec,
		metadata.CmdExecResult,
		metadata.CmdExecRows,
		metadata.CmdOne,
		metadata.CmdMany,
		metadata.CmdCopyFrom:
		return true
	default:
		return false
	}
}

func (d *AsyncpgDriver) WriteQueryFunc(body *writer.CodeWriter, config *config.Config, query model.Query, indent int) {
	var annotation, docRetType string
	switch query.Cmd {
	case metadata.CmdExec:
		annotation, docRetType = query.Returns.Type.Print(), ""
	case metadata.CmdExecResult:
		annotation, docRetType = types.Str, types.Str
	case metadata.CmdExecRows, metadata.CmdCopyFrom:
		annotation, docRetType = query.Returns.Type.Print(), query.Returns.Type.Type
	case metadata.CmdOne:
		annotation, docRetType = query.Returns.Type.PrintOptional(), query.Returns.Type.Type
	case metadata.CmdMany:
		annotation, docRetType = "QueryResults["+query.Returns.Type.Print()+"]", query.Returns.Type.Print()
	}

	conn := writeFuncSignature(body, d, config, indent, query, annotation)

	indent++
	writeQueryDocstring(body, d, config, query, indent, docRetType)
	callArgs := append([]string{query.ConstantName}, expandParams(query)...)
	switch query.Cmd {
	case metadata.CmdExec:
		body.WriteWrappedCall(indent, fmt.Sprintf("await %s.execute(", conn), callArgs, ")")

	case metadata.CmdExecResult:
		body.WriteWrappedCall(indent, fmt.Sprintf("return await %s.execute(", conn), callArgs, ")")

	case metadata.CmdExecRows:
		body.WriteWrappedCall(indent, fmt.Sprintf("r = await %s.execute(", conn), callArgs, ")")
		writeExecRowsReturn(body, config, indent)

	case metadata.CmdCopyFrom:
		writeCopyFromBody(body, config, query, conn, indent)

	case metadata.CmdOne:
		body.WriteWrappedCall(indent, fmt.Sprintf("row = await %s.fetchrow(", conn), callArgs, ")")
		body.WriteIndentedLine(indent, "if row is None:")
		body.WriteIndentedLine(indent+1, "return None")

		if query.Returns.IsStruct() {
			d.rows.WriteStructReturn(body, indent, query.Returns)
		} else {
			d.rows.WriteScalarReturn(body, indent, query.Returns)
		}

	case metadata.CmdMany:
		decodeHook := d.rows.WriteDecodeHook(body, indent, query, asyncpgResultType)
		manyArgs := append([]string{conn, query.ConstantName, decodeHook}, expandParams(query)...)
		// Deliberately unsubscripted: QueryResults[T](...) would go through
		// typing's _GenericAlias.__call__ on every invocation (~10x call
		// overhead) for zero benefit - the return annotation carries the type.
		body.WriteWrappedCall(indent, "return QueryResults(", manyArgs, ")")
	}
}

// writeCopyFromBody writes the body for an asyncpg :copyfrom command.
func writeCopyFromBody(body *writer.CodeWriter, config *config.Config, query model.Query, conn string, indent int) {
	paramParts := make([]string, 0, len(query.Params[0].Table.Columns))
	columnParts := make([]string, 0, len(query.Params[0].Table.Columns))
	for _, col := range query.Params[0].Table.Columns {
		// Overridden columns convert back to their DefaultType here too:
		// copy_records_to_table receives the raw record values, so this is
		// the only place the conversion can happen for :copyfrom.
		paramParts = append(paramParts, convertParamExpr("param."+col.Name, col.Type))
		columnParts = append(columnParts, writer.PyQuote(col.DBName))
	}

	paramsName := query.Params[0].Name
	recordTuple := "(" + strings.Join(paramParts, ", ") + ")"
	if len(paramParts) == 1 {
		// A one-element tuple needs the trailing comma, otherwise the
		// parentheses are just grouping and the record is a bare value.
		recordTuple = "(" + paramParts[0] + ",)"
	}
	singleComprehension := fmt.Sprintf("records = [%s for param in %s]", recordTuple, paramsName)
	switch {
	case body.FitsLine(indent, singleComprehension):
		body.WriteIndentedLine(indent, singleComprehension)
	case body.FitsLine(indent+1, recordTuple):
		body.WriteIndentedLine(indent, "records = [")
		body.WriteIndentedLine(indent+1, recordTuple)
		body.WriteIndentedLine(indent+1, "for param in "+paramsName)
		body.WriteIndentedLine(indent, "]")
	default:
		body.WriteIndentedLine(indent, "records = [")
		body.WriteIndentedLine(indent+1, "(")
		for _, part := range paramParts {
			body.WriteIndentedLine(indent+2, part+",")
		}
		body.WriteIndentedLine(indent+1, ")")
		body.WriteIndentedLine(indent+1, "for param in "+paramsName)
		body.WriteIndentedLine(indent, "]")
	}
	columnsArg := fmt.Sprintf("columns=[%s]", strings.Join(columnParts, ", "))
	copyArgs := []string{writer.PyQuote(query.Table.Name), columnsArg, "records=records"}
	if query.Table.Schema != "" {
		copyArgs = append(copyArgs, "schema_name="+writer.PyQuote(query.Table.Schema))
	}

	head := fmt.Sprintf("r = await %s.copy_records_to_table(", conn)
	single := head + strings.Join(copyArgs, ", ") + ")"
	switch {
	case body.FitsLine(indent, single):
		body.WriteIndentedLine(indent, single)
	default:
		body.WriteIndentedLine(indent, head)
		for _, arg := range copyArgs {
			if arg == columnsArg && !body.FitsLine(indent+1, arg+",") {
				body.WriteIndentedLine(indent+1, "columns=[")
				for _, col := range columnParts {
					body.WriteIndentedLine(indent+2, col+",")
				}
				body.WriteIndentedLine(indent+1, "],")

				continue
			}
			body.WriteIndentedLine(indent+1, arg+",")
		}
		body.WriteIndentedLine(indent, ")")
	}
	writeExecRowsReturn(body, config, indent)
}
