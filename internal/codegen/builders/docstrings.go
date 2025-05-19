package builders

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

var docstringConfig *string
var docstringConfigEmitSQL *bool
var docstringConfigDriver core.SQLDriverType = core.SQLDriverAsyncpg

func SetDocstringConfig(c *string, b *bool, d core.SQLDriverType) {
	docstringConfig = c
	docstringConfigEmitSQL = b
	docstringConfigDriver = d
}

func (b *IndentStringBuilder) WriteQueryResultsAiterDocstring() {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedLine(2, `"""`+"Initialize iteration support for `async for`.")
	b.NewLine()
	if *docstringConfig == core.DocstringConventionNumpy {
		b.WriteIndentedLine(2, "Returns")
		b.WriteIndentedLine(2, "-------")
		b.WriteIndentedLine(2, "QueryResults[T]")
		b.WriteIndentedLine(3, "Self as an asynchronous iterator.")
	} else if *docstringConfig == core.DocstringConventionGoogle {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(3, "Self as an asynchronous iterator.")
	} else if *docstringConfig == core.DocstringConventionPEP257 {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(2, "Self as an asynchronous iterator.")
	}
	b.WriteIndentedLine(2, `"""`)
}

func (b *IndentStringBuilder) WriteQueryResultsAnextDocstringAiosqlite() {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedLine(2, `"""Yield the next item in the query result using an aiosqlite cursor.`)
	b.NewLine()
	if *docstringConfig == core.DocstringConventionNumpy {
		b.WriteIndentedLine(2, "Returns")
		b.WriteIndentedLine(2, "-------")
		b.WriteIndentedLine(2, "T")
		b.WriteIndentedLine(3, "The next decoded result.")
		b.NewLine()
		b.WriteIndentedLine(2, "Raises")
		b.WriteIndentedLine(2, "------")
		b.WriteIndentedLine(2, "StopAsyncIteration")
		b.WriteIndentedLine(3, "When no more records are available.")
	} else if *docstringConfig == core.DocstringConventionGoogle {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(3, "The next decoded result of type `T`.")
		b.NewLine()
		b.WriteIndentedLine(2, "Raises:")
		b.WriteIndentedLine(3, "StopAsyncIteration: When no more records are available.")
	} else if *docstringConfig == core.DocstringConventionPEP257 {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(2, "The next decoded result of type `T`.")
		b.NewLine()
		b.WriteIndentedLine(2, "Raises:")
		b.WriteIndentedLine(2, "StopAsyncIteration -- When no more records are available.")
	}
	b.WriteIndentedLine(2, `"""`)
}

func (b *IndentStringBuilder) WriteQueryResultsAnextDocstringAsyncpg() {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedLine(2, `"""Yield the next item in the query result using an asyncpg cursor.`)
	b.NewLine()
	if *docstringConfig == core.DocstringConventionNumpy {
		b.WriteIndentedLine(2, "Returns")
		b.WriteIndentedLine(2, "-------")
		b.WriteIndentedLine(2, "T")
		b.WriteIndentedLine(3, "The next decoded result.")
		b.NewLine()
		b.WriteIndentedLine(2, "Raises")
		b.WriteIndentedLine(2, "------")
		b.WriteIndentedLine(2, "StopAsyncIteration")
		b.WriteIndentedLine(3, "When no more records are available.")
	} else if *docstringConfig == core.DocstringConventionGoogle {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(3, "The next decoded result of type `T`.")
		b.NewLine()
		b.WriteIndentedLine(2, "Raises:")
		b.WriteIndentedLine(3, "StopAsyncIteration: When no more records are available.")
	} else if *docstringConfig == core.DocstringConventionPEP257 {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(2, "The next decoded result of type `T`.")
		b.NewLine()
		b.WriteIndentedLine(2, "Raises:")
		b.WriteIndentedLine(2, "StopAsyncIteration -- When no more records are available.")
	}
	b.WriteIndentedLine(2, `"""`)
}

func (b *IndentStringBuilder) WriteQueryResultsAwaitDocstring() {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedLine(2, `"""`+"Allow `await` on the object to return all rows as a fully decoded sequence.")
	b.NewLine()
	if *docstringConfig == core.DocstringConventionNumpy {
		b.WriteIndentedLine(2, "Returns")
		b.WriteIndentedLine(2, "-------")
		b.WriteIndentedLine(2, "collections.abc.Sequence[T]")
		b.WriteIndentedLine(3, "A sequence of decoded objects of type `T`.")
	} else if *docstringConfig == core.DocstringConventionGoogle {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(3, "A sequence of decoded objects of type `T`.")
	} else if *docstringConfig == core.DocstringConventionPEP257 {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(2, "A sequence of decoded objects of type `T`.")
	}
	b.WriteIndentedLine(2, `"""`)
}

func (b *IndentStringBuilder) WriteQueryResultsInitDocstring(docstringConnType string, docstringDriverReturnType string) {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedString(2, fmt.Sprintf(`"""Initialize the QueryResults instance.`))
	if *docstringConfig == core.DocstringConventionNumpy {
		b.WriteLine(`"""`)
	} else if *docstringConfig == core.DocstringConventionGoogle {
		b.NNewLine(2)
		b.WriteIndentedLine(2, "Args:")
		b.WriteIndentedLine(3, "conn:")
		b.WriteIndentedLine(4, fmt.Sprintf("The connection object of type `%s` used to execute queries.", docstringConnType))
		b.WriteIndentedLine(3, "sql:")
		b.WriteIndentedLine(4, "The SQL statement that will be executed when fetching/iterating.")
		b.WriteIndentedLine(3, "decode_hook:")
		b.WriteIndentedLine(4, fmt.Sprintf("A callback that turns an `%s` object into `T` that will be returned.", docstringDriverReturnType))
		b.WriteIndentedLine(3, "*args:")
		b.WriteIndentedLine(4, "Arguments that should be sent when executing the sql query.")
		b.WriteIndentedLine(2, `"""`)
	} else if *docstringConfig == core.DocstringConventionPEP257 {
		b.NNewLine(2)
		b.WriteIndentedLine(2, "Arguments:")
		b.WriteIndentedLine(2, fmt.Sprintf("conn -- The connection object of type `%s` used to execute queries.", docstringConnType))
		b.WriteIndentedLine(2, "sql -- The SQL statement that will be executed when fetching/iterating.")
		b.WriteIndentedLine(2, fmt.Sprintf("decode_hook -- A callback that turns an `%s` object into `T` that will be returned.", docstringDriverReturnType))
		b.WriteIndentedLine(2, "*args -- Arguments that should be sent when executing the sql query.")
		b.WriteIndentedLine(2, `"""`)
	}
}

func (b *IndentStringBuilder) WriteQueryResultsClassDocstring(docstringConnType string, docstringDriverReturnType string) {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedString(1, `"""Helper class that allows both iteration and normal fetching of data from the db.`)
	if *docstringConfig == core.DocstringConventionNumpy {
		b.NewLine()
		b.NewLine()
		b.WriteIndentedLine(1, "Parameters")
		b.WriteIndentedLine(1, "----------")
		b.WriteIndentedLine(1, "conn")
		b.WriteIndentedLine(2, fmt.Sprintf("The connection object of type `%s` used to execute queries.", docstringConnType))
		b.WriteIndentedLine(1, "sql")
		b.WriteIndentedLine(2, "The SQL statement that will be executed when fetching/iterating.")
		b.WriteIndentedLine(1, "decode_hook")
		b.WriteIndentedLine(2, fmt.Sprintf("A callback that turns an `%s` object into `T` that will be returned.", docstringDriverReturnType))
		b.WriteIndentedLine(1, "*args")
		b.WriteIndentedLine(2, "Arguments that should be sent when executing the sql query.")
		b.NewLine()
		b.WriteIndentedLine(1, `"""`)
	} else {
		b.WriteLine(`"""`)
	}
	b.NewLine()
}

func (b *IndentStringBuilder) WriteQueryClassConnDocstring(docstringConnType string) {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedLine(2, `"""Connection object used to make queries.`)
	b.NewLine()
	if *docstringConfig == core.DocstringConventionNumpy {
		b.WriteIndentedLine(2, "Returns")
		b.WriteIndentedLine(2, "-------")
		b.WriteIndentedLine(2, docstringConnType)
		b.NewLine()
	} else if *docstringConfig == core.DocstringConventionGoogle {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(3, fmt.Sprintf("Connection object of type `%s` used to make queries.", docstringConnType))
	} else if *docstringConfig == core.DocstringConventionPEP257 {
		b.WriteIndentedLine(2, "Returns:")
		b.WriteIndentedLine(2, fmt.Sprintf("%s -- Connection object used to make queries.", docstringConnType))
	}
	b.WriteIndentedLine(2, `"""`)
}

func (b *IndentStringBuilder) WriteQueryClassDocstring(sourceName string, docstringConnType string) {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedString(1, fmt.Sprintf(`"""Queries from file %s.`, sourceName))
	if *docstringConfig == core.DocstringConventionNumpy {
		b.NewLine()
		b.NewLine()
		b.WriteIndentedLine(1, "Parameters")
		b.WriteIndentedLine(1, "----------")
		b.WriteIndentedLine(1, fmt.Sprintf("conn : %s", docstringConnType))
		b.WriteIndentedLine(2, "The connection object used to execute queries.")
		b.NewLine()
		b.WriteIndentedLine(1, `"""`)
	} else {
		b.WriteLine(`"""`)
	}
	b.NewLine()
}

func (b *IndentStringBuilder) WriteQueryClassInitDocstring(lvl int, docstringConnType string) {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedString(lvl, fmt.Sprintf(`"""Initialize the instance using the connection.`))
	if *docstringConfig == core.DocstringConventionNumpy {
		b.WriteLine(`"""`)
	} else if *docstringConfig == core.DocstringConventionGoogle {
		b.NNewLine(2)
		b.WriteIndentedLine(lvl, "Args:")
		b.WriteIndentedLine(lvl+1, "conn:")
		b.WriteIndentedLine(lvl+2, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
		b.WriteIndentedLine(lvl, `"""`)
	} else if *docstringConfig == core.DocstringConventionPEP257 {
		b.NNewLine(2)
		b.WriteIndentedLine(lvl, "Arguments:")
		b.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection object of type `%s` used to execute queries.", docstringConnType))
		b.WriteIndentedLine(lvl, `"""`)
	}
}

func (b *IndentStringBuilder) WriteModelClassDocstring(table *core.Table) {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedLine(1, `"""`+fmt.Sprintf("Model representing %s.", table.Name))
	if *docstringConfig == core.DocstringConventionNumpy {
		b.NewLine()
		b.WriteIndentedLine(1, "Attributes")
		b.WriteIndentedLine(1, "----------")
		for _, col := range table.Columns {
			type_ := col.Type.Type
			if col.Type.IsList {
				type_ = "collections.abc.Sequence[" + type_ + "]"
			}
			if col.Type.IsNullable {
				type_ = type_ + " | None"
			}
			b.WriteIndentedLine(1, fmt.Sprintf("%s : %s", col.Name, type_))
		}
		b.NewLine()
	} else if *docstringConfig == core.DocstringConventionGoogle {
		b.NewLine()
		b.WriteIndentedLine(1, "Attributes:")
		for _, col := range table.Columns {
			type_ := col.Type.Type
			if col.Type.IsList {
				type_ = "collections.abc.Sequence[" + type_ + "]"
			}
			if col.Type.IsNullable {
				type_ = type_ + " | None"
			}
			b.WriteIndentedLine(2, fmt.Sprintf("%s: %s", col.Name, type_))
		}
	} else if *docstringConfig == core.DocstringConventionPEP257 {
		b.NewLine()
		b.WriteIndentedLine(1, "Attributes:")
		for _, col := range table.Columns {
			type_ := col.Type.Type
			if col.Type.IsList {
				type_ = "collections.abc.Sequence[" + type_ + "]"
			}
			if col.Type.IsNullable {
				type_ = type_ + " | None"
			}
			b.WriteIndentedLine(1, fmt.Sprintf("%s -- %s", col.Name, type_))
		}
	}
	b.WriteIndentedLine(1, `"""`)
	b.NewLine()
}

func (b *IndentStringBuilder) WriteModelFileModuleDocstring() {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteLine(`"""Module containing models."""`)
}

func (b *IndentStringBuilder) WriteQueryFileModuleDocstring(sourceName string) {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteLine(fmt.Sprintf(`"""Module containing queries from file %s."""`, sourceName))
}

func (b *IndentStringBuilder) WriteInitFileModuleDocstring() {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteLine(`"""Package containing queries and models automatically generated using sqlc-gen-better-python."""`)
}

func (b *IndentStringBuilder) writeQueryFunctionSQL(lvl int, query *core.Query) {
	if *docstringConfigEmitSQL {
		b.WriteIndentedLine(lvl, "```sql")
		for _, line := range core.SplitLines(query.SQL) {
			b.WriteIndentedLine(lvl, line)
		}
		b.WriteIndentedLine(lvl, "```")
		b.NewLine()
	}
}

func (b *IndentStringBuilder) WriteQueryFunctionDocstring(lvl int, query *core.Query, docstringConnType string, queryArgs []core.FunctionArg, returnType core.PyType) {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}

	if query.Cmd == metadata.CmdExec {
		b.WriteIndentedLine(lvl, `"""`+fmt.Sprintf("Execute SQL query with `name: %s %s`.", query.MethodName, query.Cmd))
		b.NewLine()
		b.writeQueryFunctionSQL(lvl, query)
		if len(queryArgs) == 0 && docstringConnType == "" {
			b.WriteIndentedLine(lvl, `"""`)
			return
		}
		if *docstringConfig == core.DocstringConventionNumpy {
			b.WriteIndentedLine(lvl, "Parameters")
			b.WriteIndentedLine(lvl, "----------")
			if docstringConnType != "" {
				b.WriteIndentedLine(lvl, fmt.Sprintf("conn : %s", docstringConnType))
				b.WriteIndentedLine(lvl+1, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
			}
			for _, arg := range queryArgs {
				b.WriteIndentedLine(lvl, fmt.Sprintf("%s : %s", arg.Name, arg.Type))
			}
			b.NewLine()
		} else if *docstringConfig == core.DocstringConventionGoogle {
			b.WriteIndentedLine(lvl, "Args:")
			if docstringConnType != "" {
				b.WriteIndentedLine(lvl+1, "conn:")
				b.WriteIndentedLine(lvl+2, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
			}
			for _, arg := range queryArgs {
				b.WriteIndentedLine(lvl+1, fmt.Sprintf("%s: %s.", arg.Name, arg.Type))
			}
		} else if *docstringConfig == core.DocstringConventionPEP257 {
			b.WriteIndentedLine(lvl, "Arguments:")
			if docstringConnType != "" {
				b.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection object of type `%s` used to execute the query.", docstringConnType))
			}
			for _, arg := range queryArgs {
				b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- %s.", arg.Name, arg.Type))
			}
		}
		b.WriteIndentedLine(lvl, `"""`)
	} else if query.Cmd == metadata.CmdExecRows {
		b.WriteIndentedLine(lvl, `"""`+fmt.Sprintf("Execute SQL query with `name: %s %s` and return the number of affected rows.", query.MethodName, query.Cmd))
		b.NewLine()
		b.writeQueryFunctionSQL(lvl, query)
		if *docstringConfig == core.DocstringConventionNumpy {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Parameters")
				b.WriteIndentedLine(lvl, "----------")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn : %s", docstringConnType))
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s : %s", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns")
			b.WriteIndentedLine(lvl, "-------")
			b.WriteIndentedLine(lvl, returnType.Type)
			if docstringConfigDriver == core.SQLDriverAioSQLite {
				b.WriteIndentedLine(lvl+1, "The number of affected rows. This will be -1 for queries like `CREATE TABLE`.")
			} else {
				b.WriteIndentedLine(lvl+1, "The number of affected rows. This will be 0 for queries like `CREATE TABLE`.")
			}
			b.NewLine()
		} else if *docstringConfig == core.DocstringConventionGoogle {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Args:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl+1, "conn:")
					b.WriteIndentedLine(lvl+2, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("%s: %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			if docstringConfigDriver == core.SQLDriverAioSQLite {
				b.WriteIndentedLine(lvl+1, fmt.Sprintf("The number (`%s`) of affected rows. This will be -1 for queries like `CREATE TABLE`.", returnType.Type))
			} else {
				b.WriteIndentedLine(lvl+1, fmt.Sprintf("The number (`%s`) of affected rows. This will be 0 for queries like `CREATE TABLE`.", returnType.Type))
			}
		} else if *docstringConfig == core.DocstringConventionPEP257 {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Arguments:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			if docstringConfigDriver == core.SQLDriverAioSQLite {
				b.WriteIndentedLine(lvl+1, fmt.Sprintf("%s -- The number of affected rows. This will be -1 for queries like `CREATE TABLE`.", returnType.Type))
			} else {
				b.WriteIndentedLine(lvl+1, fmt.Sprintf("%s -- The number of affected rows. This will be 0 for queries like `CREATE TABLE`.", returnType.Type))
			}
		}
		b.WriteIndentedLine(lvl, `"""`)
	} else if query.Cmd == metadata.CmdExecResult {
		b.WriteIndentedLine(lvl, `"""`+fmt.Sprintf("Execute and return the result of SQL query with `name: %s %s`.", query.MethodName, query.Cmd))
		b.NewLine()
		b.writeQueryFunctionSQL(lvl, query)
		if *docstringConfig == core.DocstringConventionNumpy {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Parameters")
				b.WriteIndentedLine(lvl, "----------")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn : %s", docstringConnType))
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s : %s", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns")
			b.WriteIndentedLine(lvl, "-------")
			b.WriteIndentedLine(lvl, returnType.Type)
			b.WriteIndentedLine(lvl+1, "The result returned when executing the query.")
			b.NewLine()
		} else if *docstringConfig == core.DocstringConventionGoogle {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Args:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl+1, "conn:")
					b.WriteIndentedLine(lvl+2, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("%s: %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl+1, fmt.Sprintf("The result of type `%s` returned when executing the query.", returnType.Type))
		} else if *docstringConfig == core.DocstringConventionPEP257 {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Arguments:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- The result returned when executing the query.", returnType.Type))
		}
		b.WriteIndentedLine(lvl, `"""`)
	} else if query.Cmd == metadata.CmdExecLastId {
		b.WriteIndentedLine(lvl, `"""`+fmt.Sprintf("Execute SQL query with `name: %s %s` and return the id of the last affected row.", query.MethodName, query.Cmd))
		b.NewLine()
		b.writeQueryFunctionSQL(lvl, query)
		if *docstringConfig == core.DocstringConventionNumpy {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Parameters")
				b.WriteIndentedLine(lvl, "----------")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn : %s", docstringConnType))
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s : %s", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns")
			b.WriteIndentedLine(lvl, "-------")
			b.WriteIndentedLine(lvl, returnType.Type)
			b.WriteIndentedLine(lvl+1, "The id of the last affected row. Will be `None` if no rows are affected.")
			b.NewLine()
		} else if *docstringConfig == core.DocstringConventionGoogle {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Args:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl+1, "conn:")
					b.WriteIndentedLine(lvl+2, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("%s: %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl+1, fmt.Sprintf("The id (`%s`) of the last affected row. Will be `None` if no rows are affected.", returnType.Type))
		} else if *docstringConfig == core.DocstringConventionPEP257 {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Arguments:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- The id of the last affected row. Will be `None` if no rows are affected.", returnType.Type))
		}
		b.WriteIndentedLine(lvl, `"""`)
	} else if query.Cmd == metadata.CmdOne {
		b.WriteIndentedLine(lvl, `"""`+fmt.Sprintf("Fetch one from the db using the SQL query with `name: %s %s`.", query.MethodName, query.Cmd))
		b.NewLine()
		b.writeQueryFunctionSQL(lvl, query)
		if *docstringConfig == core.DocstringConventionNumpy {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Parameters")
				b.WriteIndentedLine(lvl, "----------")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn : %s", docstringConnType))
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s : %s", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns")
			b.WriteIndentedLine(lvl, "-------")
			b.WriteIndentedLine(lvl, returnType.Type)
			b.WriteIndentedLine(lvl+1, "Result fetched from the db. Will be `None` if not found.")
			b.NewLine()

		} else if *docstringConfig == core.DocstringConventionGoogle {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Args:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl+1, "conn:")
					b.WriteIndentedLine(lvl+2, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("%s: %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl+1, fmt.Sprintf("Result of type `%s` fetched from the db. Will be `None` if not found.", returnType.Type))
		} else if *docstringConfig == core.DocstringConventionPEP257 {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Arguments:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- Result fetched from the db. Will be `None` if not found.", returnType.Type))
		}
		b.WriteIndentedLine(lvl, `"""`)
	} else if query.Cmd == metadata.CmdMany {
		b.WriteIndentedLine(lvl, `"""`+fmt.Sprintf("Fetch many from the db using the SQL query with `name: %s %s`.", query.MethodName, query.Cmd))
		b.NewLine()
		b.writeQueryFunctionSQL(lvl, query)
		if *docstringConfig == core.DocstringConventionNumpy {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Parameters")
				b.WriteIndentedLine(lvl, "----------")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn : %s", docstringConnType))
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s : %s", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns")
			b.WriteIndentedLine(lvl, "-------")
			b.WriteIndentedLine(lvl, fmt.Sprintf("QueryResults[%s]", returnType.Type))
			b.WriteIndentedLine(lvl+1, "Helper class that allows both iteration and normal fetching of data from the db.")
			b.NewLine()
		} else if *docstringConfig == core.DocstringConventionGoogle {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Args:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl+1, "conn:")
					b.WriteIndentedLine(lvl+2, fmt.Sprintf("Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl+1, fmt.Sprintf("%s: %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl+1, fmt.Sprintf("Helper class of type `QueryResults[%s]` that allows both iteration and normal fetching of data from the db.", returnType.Type))
		} else if *docstringConfig == core.DocstringConventionPEP257 {
			if len(queryArgs) != 0 || docstringConnType != "" {
				b.WriteIndentedLine(lvl, "Arguments:")
				if docstringConnType != "" {
					b.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection object of type `%s` used to execute the query.", docstringConnType))
				}
				for _, arg := range queryArgs {
					b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- %s.", arg.Name, arg.Type))
				}
				b.NewLine()
			}
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl, fmt.Sprintf("QueryResults[%s] -- Helper class that allows both iteration and normal fetching of data from the db.", returnType.Type))
		}
		b.WriteIndentedLine(lvl, `"""`)
	}
}
