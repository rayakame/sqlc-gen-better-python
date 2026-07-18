package writer

import (
	"fmt"
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

// DocArg describes one function argument in a docstring.
type DocArg struct {
	Name  string
	Type  string
	Extra string // optional extra description line (used by :copyfrom)
}

// DocstringsEnabled reports whether docstring generation is active. Emitters
// use this to add the blank line ruff format requires between a docstring and
// a following nested function definition.
func (w *CodeWriter) DocstringsEnabled() bool {
	return w.docstringConvention != config.DocstringConventionNone
}

// --- Module docstrings ------------------------------------------------------

// WriteModelFileModuleDocstring writes the models.py module docstring.
func (w *CodeWriter) WriteModelFileModuleDocstring() {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteLine(`"""Module containing models."""`)
	w.NewLine()
}

// WriteEnumsFileModuleDocstring writes the enums.py module docstring.
func (w *CodeWriter) WriteEnumsFileModuleDocstring() {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteLine(`"""Module containing enums."""`)
	w.NewLine()
}

// WriteQueryFileModuleDocstring writes a query module docstring.
func (w *CodeWriter) WriteQueryFileModuleDocstring(sourceName string) {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteLine(fmt.Sprintf(`"""Module containing queries from file %s."""`, sourceName))
	w.NewLine()
}

// WriteInitFileModuleDocstring writes the __init__.py module docstring.
func (w *CodeWriter) WriteInitFileModuleDocstring() {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteLine(`"""Package containing queries and models automatically generated using sqlc-gen-better-python."""`)
}

// --- Class docstrings -------------------------------------------------------

// WriteModelClassDocstring writes a model/row/params class docstring with an
// attribute list, followed by a blank line.
func (w *CodeWriter) WriteModelClassDocstring(table *model.Table) {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteIndentedLine(1, `"""`+fmt.Sprintf("Model representing %s.", table.Name))
	w.NewLine()
	switch w.docstringConvention {
	case config.DocstringConventionNumpy:
		w.WriteIndentedLine(1, "Attributes")
		w.WriteIndentedLine(1, "----------")
		for _, col := range table.Columns {
			w.WriteIndentedLine(1, fmt.Sprintf("%s : %s", col.Name, col.Type.Print()))
		}
		w.NewLine()
	case config.DocstringConventionGoogle:
		w.WriteIndentedLine(1, "Attributes:")
		for _, col := range table.Columns {
			w.WriteIndentedLine(2, fmt.Sprintf("%s: %s", col.Name, col.Type.Print()))
		}
	case config.DocstringConventionPEP257:
		w.WriteIndentedLine(1, "Attributes:")
		for _, col := range table.Columns {
			w.WriteIndentedLine(1, fmt.Sprintf("%s -- %s", col.Name, col.Type.Print()))
		}
	}
	w.WriteIndentedLine(1, `"""`)
	w.NewLine()
}

// WriteEnumClassDocstring writes a one-line enum class docstring.
func (w *CodeWriter) WriteEnumClassDocstring(name string) {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteIndentedLine(1, fmt.Sprintf(`"""Enum representing %s."""`, name))
	w.NewLine()
}

// --- Querier class docstrings -----------------------------------------------

// WriteQueryClassDocstring writes the Querier class docstring.
func (w *CodeWriter) WriteQueryClassDocstring(sourceName, connType string) {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteIndentedString(1, fmt.Sprintf(`"""Queries from file %s.`, sourceName))
	if w.docstringConvention == config.DocstringConventionNumpy {
		w.NNewLine(2)
		w.WriteIndentedLine(1, "Parameters")
		w.WriteIndentedLine(1, "----------")
		w.WriteIndentedLine(1, "conn : "+connType)
		w.WriteIndentedLine(2, "The connection object used to execute queries.")
		w.NewLine()
		w.WriteIndentedLine(1, `"""`)
	} else {
		w.WriteLine(`"""`)
	}
	w.NewLine()
}

// WriteQueryClassInitDocstring writes the Querier __init__ docstring.
func (w *CodeWriter) WriteQueryClassInitDocstring(lvl int, connType string) {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteIndentedString(lvl, `"""Initialize the instance using the connection.`)
	switch w.docstringConvention {
	case config.DocstringConventionNumpy:
		w.WriteLine(`"""`)
	case config.DocstringConventionGoogle:
		w.NNewLine(2)
		w.WriteIndentedLine(lvl, "Args:")
		w.WriteIndentedLine(lvl+1, "conn:")
		w.WriteIndentedLine(lvl+2, fmt.Sprintf("Connection object of type `%s` used to execute the query.", connType))
		w.WriteIndentedLine(lvl, `"""`)
	case config.DocstringConventionPEP257:
		w.NNewLine(2)
		w.WriteIndentedLine(lvl, "Arguments:")
		w.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection object of type `%s` used to execute queries.", connType))
		w.WriteIndentedLine(lvl, `"""`)
	}
}

// WriteQueryClassConnDocstring writes the Querier conn property docstring.
func (w *CodeWriter) WriteQueryClassConnDocstring(connType string) {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteIndentedLine(2, `"""Connection object used to make queries.`)
	w.NewLine()
	switch w.docstringConvention {
	case config.DocstringConventionNumpy:
		w.WriteIndentedLine(2, "Returns")
		w.WriteIndentedLine(2, "-------")
		w.WriteIndentedLine(2, connType)
		w.NewLine()
	case config.DocstringConventionGoogle:
		w.WriteIndentedLine(2, "Returns:")
		w.WriteIndentedLine(3, fmt.Sprintf("Connection object of type `%s` used to make queries.", connType))
	case config.DocstringConventionPEP257:
		w.WriteIndentedLine(2, "Returns:")
		w.WriteIndentedLine(2, fmt.Sprintf("%s -- Connection object used to make queries.", connType))
	}
	w.WriteIndentedLine(2, `"""`)
}

// --- QueryResults docstrings --------------------------------------------------

// WriteQueryResultsClassDocstring writes the QueryResults class docstring.
func (w *CodeWriter) WriteQueryResultsClassDocstring(connType, resultType string) {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteIndentedString(1, `"""Helper class that allows both iteration and normal fetching of data from the db.`)
	if w.docstringConvention == config.DocstringConventionNumpy {
		w.NNewLine(2)
		w.WriteIndentedLine(1, "Parameters")
		w.WriteIndentedLine(1, "----------")
		w.WriteIndentedLine(1, "conn")
		w.WriteIndentedLine(2, fmt.Sprintf("The connection object of type `%s` used to execute queries.", connType))
		w.WriteIndentedLine(1, "sql")
		w.WriteIndentedLine(2, "The SQL statement that will be executed when fetching/iterating.")
		w.WriteIndentedLine(1, "decode_hook")
		w.WriteIndentedLine(2, fmt.Sprintf("A callback that turns an `%s` object into `T` that will be returned.", resultType))
		w.WriteIndentedLine(1, "*args")
		w.WriteIndentedLine(2, "Arguments that should be sent when executing the sql query.")
		w.NewLine()
		w.WriteIndentedLine(1, `"""`)
	} else {
		w.WriteLine(`"""`)
	}
	w.NewLine()
}

// WriteQueryResultsInitDocstring writes the QueryResults __init__ docstring.
func (w *CodeWriter) WriteQueryResultsInitDocstring(connType, resultType string) {
	if !w.DocstringsEnabled() {
		return
	}
	w.WriteIndentedString(2, `"""Initialize the QueryResults instance.`)
	switch w.docstringConvention {
	case config.DocstringConventionNumpy:
		w.WriteLine(`"""`)
	case config.DocstringConventionGoogle:
		w.NNewLine(2)
		w.WriteIndentedLine(2, "Args:")
		w.WriteIndentedLine(3, "conn:")
		w.WriteIndentedLine(4, fmt.Sprintf("The connection object of type `%s` used to execute queries.", connType))
		w.WriteIndentedLine(3, "sql:")
		w.WriteIndentedLine(4, "The SQL statement that will be executed when fetching/iterating.")
		w.WriteIndentedLine(3, "decode_hook:")
		w.WriteIndentedLine(4, fmt.Sprintf("A callback that turns an `%s` object into `T` that will be returned.", resultType))
		w.WriteIndentedLine(3, "*args:")
		w.WriteIndentedLine(4, "Arguments that should be sent when executing the sql query.")
		w.WriteIndentedLine(2, `"""`)
	case config.DocstringConventionPEP257:
		w.NNewLine(2)
		w.WriteIndentedLine(2, "Arguments:")
		w.WriteIndentedLine(2, fmt.Sprintf("conn -- The connection object of type `%s` used to execute queries.", connType))
		w.WriteIndentedLine(2, "sql -- The SQL statement that will be executed when fetching/iterating.")
		w.WriteIndentedLine(2, fmt.Sprintf("decode_hook -- A callback that turns an `%s` object into `T` that will be returned.", resultType))
		w.WriteIndentedLine(2, "*args -- Arguments that should be sent when executing the sql query.")
		w.WriteIndentedLine(2, `"""`)
	}
}

// WriteQueryResultsIterDocstring writes the __iter__/__aiter__ docstring.
func (w *CodeWriter) WriteQueryResultsIterDocstring(async bool) {
	if !w.DocstringsEnabled() {
		return
	}
	summary := "Initialize iteration support."
	returns := "Self as an iterator."
	if async {
		summary = "Initialize iteration support for `async for`."
		returns = "Self as an asynchronous iterator."
	}
	w.WriteIndentedLine(2, `"""`+summary)
	w.NewLine()
	switch w.docstringConvention {
	case config.DocstringConventionNumpy:
		w.WriteIndentedLine(2, "Returns")
		w.WriteIndentedLine(2, "-------")
		w.WriteIndentedLine(2, "QueryResults[T]")
		w.WriteIndentedLine(3, returns)
	case config.DocstringConventionGoogle:
		w.WriteIndentedLine(2, "Returns:")
		w.WriteIndentedLine(3, returns)
	case config.DocstringConventionPEP257:
		w.WriteIndentedLine(2, "Returns:")
		w.WriteIndentedLine(2, returns)
	}
	w.WriteIndentedLine(2, `"""`)
}

// WriteQueryResultsNextDocstring writes the __next__/__anext__ docstring.
// cursorPhrase is e.g. "an asyncpg cursor" or "a sqlite3 cursor".
func (w *CodeWriter) WriteQueryResultsNextDocstring(cursorPhrase string, async bool) {
	if !w.DocstringsEnabled() {
		return
	}
	raises := "StopIteration"
	if async {
		raises = "StopAsyncIteration"
	}
	w.WriteIndentedLine(2, fmt.Sprintf(`"""Yield the next item in the query result using %s.`, cursorPhrase))
	w.NewLine()
	switch w.docstringConvention {
	case config.DocstringConventionNumpy:
		w.WriteIndentedLine(2, "Returns")
		w.WriteIndentedLine(2, "-------")
		w.WriteIndentedLine(2, "T")
		w.WriteIndentedLine(3, "The next decoded result.")
		w.NewLine()
		w.WriteIndentedLine(2, "Raises")
		w.WriteIndentedLine(2, "------")
		w.WriteIndentedLine(2, raises)
		w.WriteIndentedLine(3, "When no more records are available.")
	case config.DocstringConventionGoogle:
		w.WriteIndentedLine(2, "Returns:")
		w.WriteIndentedLine(3, "The next decoded result of type `T`.")
		w.NewLine()
		w.WriteIndentedLine(2, "Raises:")
		w.WriteIndentedLine(3, raises+": When no more records are available.")
	case config.DocstringConventionPEP257:
		w.WriteIndentedLine(2, "Returns:")
		w.WriteIndentedLine(2, "The next decoded result of type `T`.")
		w.NewLine()
		w.WriteIndentedLine(2, "Raises:")
		w.WriteIndentedLine(2, raises+" -- When no more records are available.")
	}
	w.WriteIndentedLine(2, `"""`)
}

// WriteQueryResultsFetchDocstring writes the __await__/__call__ docstring.
func (w *CodeWriter) WriteQueryResultsFetchDocstring(async bool) {
	if !w.DocstringsEnabled() {
		return
	}
	summary := "Allow calling the object to return all rows as a fully decoded sequence."
	if async {
		summary = "Allow `await` on the object to return all rows as a fully decoded sequence."
	}
	w.WriteIndentedLine(2, `"""`+summary)
	w.NewLine()
	switch w.docstringConvention {
	case config.DocstringConventionNumpy:
		w.WriteIndentedLine(2, "Returns")
		w.WriteIndentedLine(2, "-------")
		w.WriteIndentedLine(2, "collections.abc.Sequence[T]")
		w.WriteIndentedLine(3, "A sequence of decoded objects of type `T`.")
	case config.DocstringConventionGoogle:
		w.WriteIndentedLine(2, "Returns:")
		w.WriteIndentedLine(3, "A sequence of decoded objects of type `T`.")
	case config.DocstringConventionPEP257:
		w.WriteIndentedLine(2, "Returns:")
		w.WriteIndentedLine(2, "A sequence of decoded objects of type `T`.")
	}
	w.WriteIndentedLine(2, `"""`)
}

// --- Query function docstrings ------------------------------------------------

// retDoc describes the Returns section of a query function docstring.
type retDoc struct {
	numpyType string // type line for numpy, e.g. "int" or "QueryResults[models.Author]"
	text      string // description used by numpy/pep257
	google    string // full google-style description line
}

// WriteQueryFunctionDocstring writes the docstring for a generated query
// function. retType is the return type used in the Returns section; its
// exact value is driver-specific for some commands (e.g. :execresult).
func (w *CodeWriter) WriteQueryFunctionDocstring(lvl int, query *model.Query, connType string, args []DocArg, retType string) {
	if !w.DocstringsEnabled() {
		return
	}

	var summaryFmt string
	var ret *retDoc
	emitSQL := true
	switch query.Cmd {
	case metadata.CmdExec:
		summaryFmt = "Execute SQL query with `name: %s %s`."
	case metadata.CmdExecRows:
		summaryFmt = "Execute SQL query with `name: %s %s` and return the number of affected rows."
		noRows := "0"
		if w.docstringDriver == config.SQLDriverAioSQLite {
			noRows = "-1"
		}
		ret = &retDoc{
			numpyType: retType,
			text:      fmt.Sprintf("The number of affected rows. This will be %s for queries like `CREATE TABLE`.", noRows),
			google:    fmt.Sprintf("The number (`%s`) of affected rows. This will be %s for queries like `CREATE TABLE`.", retType, noRows),
		}
	case metadata.CmdCopyFrom:
		summaryFmt = "Execute COPY FROM query to insert rows into a table with `name: %s %s` and return the number of affected rows."
		emitSQL = false
		ret = &retDoc{
			numpyType: retType,
			text:      "The number of affected rows.",
			google:    fmt.Sprintf("The number (`%s`) of affected rows.", retType),
		}
	case metadata.CmdExecResult:
		summaryFmt = "Execute and return the result of SQL query with `name: %s %s`."
		ret = &retDoc{
			numpyType: retType,
			text:      "The result returned when executing the query.",
			google:    fmt.Sprintf("The result of type `%s` returned when executing the query.", retType),
		}
	case metadata.CmdExecLastId:
		summaryFmt = "Execute SQL query with `name: %s %s` and return the id of the last affected row."
		ret = &retDoc{
			numpyType: retType,
			text:      "The id of the last affected row. Will be `None` if no rows are affected.",
			google:    fmt.Sprintf("The id (`%s`) of the last affected row. Will be `None` if no rows are affected.", retType),
		}
	case metadata.CmdOne:
		summaryFmt = "Fetch one from the db using the SQL query with `name: %s %s`."
		ret = &retDoc{
			numpyType: retType,
			text:      "Result fetched from the db. Will be `None` if not found.",
			google:    fmt.Sprintf("Result of type `%s` fetched from the db. Will be `None` if not found.", retType),
		}
	case metadata.CmdMany:
		summaryFmt = "Fetch many from the db using the SQL query with `name: %s %s`."
		ret = &retDoc{
			numpyType: "QueryResults[" + retType + "]",
			text:      "Helper class that allows both iteration and normal fetching of data from the db.",
			google:    fmt.Sprintf("Helper class of type `QueryResults[%s]` that allows both iteration and normal fetching of data from the db.", retType),
		}
	default:
		return
	}

	w.WriteIndentedLine(lvl, `"""`+fmt.Sprintf(summaryFmt, query.QueryName, query.Cmd))
	w.NewLine()
	if emitSQL && !w.docstringOmitSQL {
		w.WriteIndentedLine(lvl, "```sql")
		for _, line := range strings.Split(strings.ReplaceAll(query.SQL, "\r\n", "\n"), "\n") {
			// Never write indentation-only lines (ruff W293).
			if strings.TrimSpace(line) == "" {
				w.NewLine()
			} else {
				w.WriteIndentedLine(lvl, line)
			}
		}
		w.WriteIndentedLine(lvl, "```")
		w.NewLine()
	}

	wroteArgs := w.writeDocArgsSection(lvl, connType, args)
	if ret == nil {
		// Commands without a Returns section (:exec): numpy keeps a trailing
		// blank line after the parameters.
		if wroteArgs && w.docstringConvention == config.DocstringConventionNumpy {
			w.NewLine()
		}
		w.WriteIndentedLine(lvl, `"""`)

		return
	}

	if wroteArgs {
		w.NewLine()
	}
	w.writeDocReturnsSection(lvl, ret)
	w.WriteIndentedLine(lvl, `"""`)
}

// writeDocArgsSection writes the Parameters/Args/Arguments section and reports
// whether anything was written. No trailing blank line is emitted.
func (w *CodeWriter) writeDocArgsSection(lvl int, connType string, args []DocArg) bool {
	if connType == "" && len(args) == 0 {
		return false
	}
	connDesc := fmt.Sprintf("Connection object of type `%s` used to execute the query.", connType)
	switch w.docstringConvention {
	case config.DocstringConventionNumpy:
		w.WriteIndentedLine(lvl, "Parameters")
		w.WriteIndentedLine(lvl, "----------")
		if connType != "" {
			w.WriteIndentedLine(lvl, "conn : "+connType)
			w.WriteIndentedLine(lvl+1, connDesc)
		}
		for _, arg := range args {
			w.WriteIndentedLine(lvl, fmt.Sprintf("%s : %s", arg.Name, arg.Type))
			if arg.Extra != "" {
				w.WriteIndentedLine(lvl+1, arg.Extra)
			}
		}
	case config.DocstringConventionGoogle:
		w.WriteIndentedLine(lvl, "Args:")
		if connType != "" {
			w.WriteIndentedLine(lvl+1, "conn:")
			w.WriteIndentedLine(lvl+2, connDesc)
		}
		for _, arg := range args {
			w.WriteIndentedLine(lvl+1, fmt.Sprintf("%s: %s.", arg.Name, arg.Type))
			if arg.Extra != "" {
				w.WriteIndentedLine(lvl+2, arg.Extra)
			}
		}
	case config.DocstringConventionPEP257:
		w.WriteIndentedLine(lvl, "Arguments:")
		if connType != "" {
			w.WriteIndentedLine(lvl, "conn -- "+connDesc)
		}
		for _, arg := range args {
			line := fmt.Sprintf("%s -- %s.", arg.Name, arg.Type)
			if arg.Extra != "" {
				line += " " + arg.Extra
			}
			w.WriteIndentedLine(lvl, line)
		}
	}

	return true
}

// writeDocReturnsSection writes the Returns section for a query function.
func (w *CodeWriter) writeDocReturnsSection(lvl int, ret *retDoc) {
	switch w.docstringConvention {
	case config.DocstringConventionNumpy:
		w.WriteIndentedLine(lvl, "Returns")
		w.WriteIndentedLine(lvl, "-------")
		w.WriteIndentedLine(lvl, ret.numpyType)
		w.WriteIndentedLine(lvl+1, ret.text)
		w.NewLine()
	case config.DocstringConventionGoogle:
		w.WriteIndentedLine(lvl, "Returns:")
		w.WriteIndentedLine(lvl+1, ret.google)
	case config.DocstringConventionPEP257:
		w.WriteIndentedLine(lvl, "Returns:")
		w.WriteIndentedLine(lvl, fmt.Sprintf("%s -- %s", ret.numpyType, ret.text))
	}
}
