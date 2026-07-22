package driver

import (
	"fmt"
	"slices"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

const (
	psycopgConnType   = "ConnectionLike"
	psycopgResultType = "psycopg.rows.TupleRow"
)

// psycopgBase is the driver implementation for psycopg (Psycopg 3).
type psycopgBase struct {
	rows *RowBuilder
}

var _ Driver = (*psycopgBase)(nil)

// newPsycopgDriver creates the psycopg driver. Runtime value conversion is
// identical to asyncpg: bytea, inet, and cidr convert inline; json and jsonb
// keep their str wire type via registered loaders, see WriteConversionSetup.
func newPsycopgDriver() *psycopgBase {
	return &psycopgBase{
		rows: newRowBuilder(asyncpgNeedsConversion),
	}
}

// Name returns the Python module name, "psycopg".
func (p *psycopgBase) Name() string { return "psycopg" }

// ConnType returns "ConnectionLike".
func (p *psycopgBase) ConnType() string { return psycopgConnType }

// IsAsync returns true; only the async psycopg flavor exists.
func (p *psycopgBase) IsAsync() bool { return true }

// NeedsConversion reports whether a SQL type needs runtime conversion.
func (p *psycopgBase) NeedsConversion(sqlType string) bool {
	return asyncpgNeedsConversion(sqlType)
}

// ConvertsInline mirrors asyncpg: bytea, inet, and cidr convert inline in
// decode code; json and jsonb are handled by registered loaders instead.
func (p *psycopgBase) ConvertsInline(sqlType string) bool {
	return asyncpgNeedsConversion(sqlType)
}

// SupportsCommand returns if the driver supports the command.
func (p *psycopgBase) SupportsCommand(cmd string) bool {
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

// TypeCheckingHook returns the ConnectionLike type alias. TupleRow is the
// default row factory's type, so the alias also documents that generated code
// expects tuple rows, and pyright rejects e.g. dict_row connections.
func (p *psycopgBase) TypeCheckingHook() []string {
	return []string{
		fmt.Sprintf("type ConnectionLike = psycopg.AsyncConnection[%s]", psycopgResultType),
	}
}

// PsycopgJSONTypesReturned collects the distinct json/jsonb type names a
// module's queries return, which decide the loader registrations.
func PsycopgJSONTypesReturned(queries []model.Query) []string {
	seen := make(map[string]struct{})
	for _, query := range queries {
		collect := func(typ model.PyType) {
			switch typ.SQLType {
			case "json", "pg_catalog.json":
				seen["json"] = struct{}{}
			case "jsonb":
				seen["jsonb"] = struct{}{}
			}
		}
		if query.Returns.IsStruct() {
			for _, col := range query.Returns.Table.Columns {
				if col.Embed != nil {
					for _, embedColumn := range col.Embed.Columns {
						collect(embedColumn.Type)
					}

					continue
				}
				collect(col.Type)
			}

			continue
		}
		collect(query.Returns.Type)
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)

	return names
}

// WriteConversionSetup registers a raw-text loader for returned json/jsonb
// columns, keeping their str wire type in line with the other drivers -
// psycopg would otherwise parse them into Python objects that cannot be
// passed back as parameters. Like sqlite converter registration, the adapters
// map is process-global, and each module registers only what it returns.
func (p *psycopgBase) WriteConversionSetup(body *writer.CodeWriter, _ *config.Config, queries []model.Query) bool {
	names := PsycopgJSONTypesReturned(queries)
	for _, name := range names {
		body.WriteLine(fmt.Sprintf(`psycopg.adapters.register_loader("%s", psycopg.types.string.TextLoader)`, name))
	}

	return len(names) != 0
}

// WriteQueryResultsClass writes the async QueryResults class for psycopg
// :many queries. Note the default cursor buffers the full result set client
// side on execute; iteration decodes row by row but does not stream from the
// server.
func (p *psycopgBase) WriteQueryResultsClass(body *writer.CodeWriter) string {
	body.QueryResults.WriteQueryResultsClassHeaderNamedParams(psycopgConnType, []string{
		fmt.Sprintf("self._cursor: psycopg.AsyncCursor[%s] | None = None", psycopgResultType),
		fmt.Sprintf("self._iterator: collections.abc.AsyncIterator[%s] | None = None", psycopgResultType),
	}, psycopgResultType, true)
	body.QueryResults.WriteQueryResultsAwaitFunction([]string{
		"result = await (await self._conn.execute(self._sql, self._params)).fetchall()",
		decodeRowsExpr,
	})
	writeAsyncNextMethod(body, "a psycopg cursor", "self._cursor = await self._conn.execute(self._sql, self._params)")

	return queryResultsClassName
}

// psycopgParamValue converts a parameter expression for the binding dict.
// Beyond the shared override conversion, unconverted sequence parameters are
// copied into a list: psycopg only dumps lists as arrays (a tuple becomes a
// composite record), while the annotation permits any sequence like asyncpg.
func psycopgParamValue(expr string, typ model.PyType) string {
	converted := convertParamExpr(expr, typ)
	if !typ.IsList || converted != expr {
		return converted
	}
	if typ.IsNullable {
		return fmt.Sprintf("list(%s) if %s is not None else None", expr, expr)
	}

	return "list(" + expr + ")"
}

// psycopgParamEntries returns the named-binding dict entries for a query's
// parameters, keyed by sqlc parameter number to match the %(pN)s rewrite.
func psycopgParamEntries(query model.Query) []string {
	entries := make([]string, 0, len(query.Params))
	appendEntry := func(number int32, expr string, typ model.PyType) {
		entries = append(entries, fmt.Sprintf(`"p%d": %s`, number, psycopgParamValue(expr, typ)))
	}
	for _, param := range query.Params {
		if param.IsEmpty() {
			continue
		}
		if param.EmitTable && param.Table != nil {
			for _, col := range param.Table.Columns {
				appendEntry(col.Number, fmt.Sprintf("%s.%s", param.Name, col.Name), col.Type)
			}

			continue
		}
		appendEntry(param.Number, param.Name, param.Type)
	}

	return entries
}

// writePsycopgCall writes head+leadArgs+dict+")" on one line, hoisting a
// too-long params dict into a local sql_params variable first; overlong
// statements wrap through WriteWrappedCall like every other driver. Only
// :many modules define QueryResultsArgsType, so only the :many hoist is
// annotated with it - there the declared type must match the QueryResults
// parameter exactly (dict is invariant), while conn.execute() accepts any
// string mapping.
func writePsycopgCall(body *writer.CodeWriter, indent int, query model.Query, head string, leadArgs []string) {
	entries := psycopgParamEntries(query)
	if len(entries) == 0 {
		body.WriteWrappedCall(indent, head, leadArgs, ")")

		return
	}

	dict := "{" + strings.Join(entries, ", ") + "}"
	stmt := head + strings.Join(append(slices.Clone(leadArgs), dict), ", ") + ")"
	if body.FitsLine(indent, stmt) {
		body.WriteIndentedLine(indent, stmt)

		return
	}

	hoist := "sql_params = {"
	if query.Cmd == metadata.CmdMany {
		hoist = "sql_params: dict[str, QueryResultsArgsType] = {"
	}
	body.WriteIndentedLine(indent, hoist)
	for _, entry := range entries {
		body.WriteIndentedLine(indent+1, entry+",")
	}
	body.WriteIndentedLine(indent, "}")
	body.WriteWrappedCall(indent, head, append(slices.Clone(leadArgs), "sql_params"), ")")
}

// writePsycopgOneCall writes the :one fetch statement. Its tail closes two
// parentheses, which WriteWrappedCall's exploded form cannot express in a
// ruff-stable way, so the overlong case emits ruff format's nested-await
// layout instead.
func writePsycopgOneCall(body *writer.CodeWriter, indent int, query model.Query, conn string) {
	entries := psycopgParamEntries(query)
	head := fmt.Sprintf("row = await (await %s.execute(", conn)
	args := []string{query.ConstantName}
	if len(entries) != 0 {
		dict := "{" + strings.Join(entries, ", ") + "}"
		stmt := head + query.ConstantName + ", " + dict + ")).fetchone()"
		if body.FitsLine(indent, stmt) {
			body.WriteIndentedLine(indent, stmt)

			return
		}
		body.WriteIndentedLine(indent, "sql_params = {")
		for _, entry := range entries {
			body.WriteIndentedLine(indent+1, entry+",")
		}
		body.WriteIndentedLine(indent, "}")
		args = append(args, "sql_params")
	}

	stmt := head + strings.Join(args, ", ") + ")).fetchone()"
	if body.FitsLine(indent, stmt) {
		body.WriteIndentedLine(indent, stmt)

		return
	}
	body.WriteIndentedLine(indent, "row = await (")
	body.WriteIndentedLine(indent+1, fmt.Sprintf("await %s.execute(", conn))
	for _, arg := range args {
		body.WriteIndentedLine(indent+2, arg+",")
	}
	body.WriteIndentedLine(indent+1, ")")
	body.WriteIndentedLine(indent, ").fetchone()")
}

func (p *psycopgBase) WriteQueryFunc(body *writer.CodeWriter, config *config.Config, query model.Query, indent int) {
	cursorType := fmt.Sprintf("psycopg.AsyncCursor[%s]", psycopgResultType)
	var annotation, docRetType string
	switch query.Cmd {
	case metadata.CmdExec:
		annotation, docRetType = query.Returns.Type.Print(), ""
	case metadata.CmdExecResult:
		annotation, docRetType = cursorType, cursorType
	case metadata.CmdExecRows, metadata.CmdCopyFrom:
		annotation, docRetType = query.Returns.Type.Print(), query.Returns.Type.Type
	case metadata.CmdOne:
		annotation, docRetType = query.Returns.Type.PrintOptional(), query.Returns.Type.Type
	case metadata.CmdMany:
		annotation, docRetType = "QueryResults["+query.Returns.Type.Print()+"]", query.Returns.Type.Print()
	}

	conn := writeFuncSignature(body, p, config, indent, query, annotation)

	indent++
	writeQueryDocstring(body, p, config, query, indent, docRetType)

	execHead := fmt.Sprintf("await %s.execute(", conn)
	constArg := []string{query.ConstantName}
	switch query.Cmd {
	case metadata.CmdExec:
		writePsycopgCall(body, indent, query, execHead, constArg)

	case metadata.CmdExecResult:
		writePsycopgCall(body, indent, query, "return "+execHead, constArg)

	case metadata.CmdExecRows:
		writePsycopgCall(body, indent, query, "cur = "+execHead, constArg)
		body.WriteIndentedLine(indent, "return cur.rowcount")

	case metadata.CmdCopyFrom:
		p.writeCopyFromBody(body, query, conn, indent)

	case metadata.CmdOne:
		writePsycopgOneCall(body, indent, query, conn)
		body.WriteIndentedLine(indent, "if row is None:")
		body.WriteIndentedLine(indent+1, "return None")

		if query.Returns.IsStruct() {
			p.rows.WriteStructReturn(body, indent, query.Returns)
		} else {
			p.rows.WriteScalarReturn(body, indent, query.Returns)
		}

	case metadata.CmdMany:
		decodeHook := p.rows.WriteDecodeHook(body, indent, query, psycopgResultType)
		writePsycopgCall(
			body,
			indent,
			query,
			"return QueryResults(",
			[]string{conn, query.ConstantName, decodeHook},
		)
	}
}

// writeCopyFromBody writes the body for a psycopg :copyfrom command: rows
// stream through cursor.copy(), and the cursor reports the inserted count.
func (p *psycopgBase) writeCopyFromBody(body *writer.CodeWriter, query model.Query, conn string, indent int) {
	columns := query.Params[0].Table.Columns
	rowParts := make([]string, 0, len(columns))
	columnParts := make([]string, 0, len(columns))
	for _, col := range columns {
		// Overridden columns convert back to their DefaultType here too:
		// copy() receives the raw row values, so this is the only place the
		// conversion can happen for :copyfrom.
		rowParts = append(rowParts, psycopgParamValue("param."+col.Name, col.Type))
		columnParts = append(columnParts, quoteSQLIdent(col.DBName))
	}

	table := quoteSQLIdent(query.Table.Name)
	if query.Table.Schema != "" {
		table = quoteSQLIdent(query.Table.Schema) + "." + table
	}
	copyStmt := fmt.Sprintf("COPY %s (%s) FROM STDIN", table, strings.Join(columnParts, ", "))

	rowTuple := "(" + strings.Join(rowParts, ", ") + ")"
	if len(rowParts) == 1 {
		// A one-element tuple needs the trailing comma, otherwise the
		// parentheses are just grouping and the row is a bare value.
		rowTuple = "(" + rowParts[0] + ",)"
	}

	copyIndent := indent + 1
	loopIndent := copyIndent + 1
	rowIndent := loopIndent + 1
	body.WriteIndentedLine(indent, fmt.Sprintf("async with %s.cursor() as cur:", conn))
	copyCall := fmt.Sprintf("async with cur.copy(%s) as copy:", writer.PyQuote(copyStmt))
	if body.FitsLine(copyIndent, copyCall) {
		body.WriteIndentedLine(copyIndent, copyCall)
	} else {
		// Matches ruff format's layout for an overlong single-string call:
		// the string moves to its own line WITHOUT a magic trailing comma.
		body.WriteIndentedLine(copyIndent, "async with cur.copy(")
		body.WriteIndentedLine(loopIndent, writer.PyQuote(copyStmt))
		body.WriteIndentedLine(copyIndent, ") as copy:")
	}
	body.WriteIndentedLine(loopIndent, "for param in "+query.Params[0].Name+":")
	writeRow := "await copy.write_row(" + rowTuple + ")"
	switch {
	case body.FitsLine(rowIndent, writeRow):
		body.WriteIndentedLine(rowIndent, writeRow)
	case len(rowParts) == 1 && body.FitsLine(rowIndent+1, rowTuple):
		// ruff format keeps a fitting one-element tuple on a single line -
		// its required trailing comma is not a magic one.
		body.WriteIndentedLine(rowIndent, "await copy.write_row(")
		body.WriteIndentedLine(rowIndent+1, rowTuple)
		body.WriteIndentedLine(rowIndent, ")")
	default:
		// ruff format's stable layout: the tuple opens on its own line inside
		// the call and the magic trailing comma keeps it exploded.
		body.WriteIndentedLine(rowIndent, "await copy.write_row(")
		body.WriteIndentedLine(rowIndent+1, "(")
		for _, part := range rowParts {
			body.WriteIndentedLine(rowIndent+2, part+",")
		}
		body.WriteIndentedLine(rowIndent+1, ")")
		body.WriteIndentedLine(rowIndent, ")")
	}
	body.WriteIndentedLine(copyIndent, "return cur.rowcount")
}

// quoteSQLIdent double-quotes a SQL identifier for the generated COPY
// statement, escaping embedded quotes by doubling them.
func quoteSQLIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
