package builders

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

var docstringConfig *string
var docstringConfigEmitSQL *bool

func SetDocstringConfig(c *string, b *bool) {
	docstringConfig = c
	docstringConfigEmitSQL = b
}

func (b *IndentStringBuilder) WriteQueryClassDocstring(lvl int, sourceName string, docstringConnType string) {
	if *docstringConfig == core.DocstringConventionNone {
		return
	}
	b.WriteIndentedString(lvl, fmt.Sprintf(`"""Queries from file %s.`, sourceName))
	if *docstringConfig == core.DocstringConventionNumpy {
		b.NewLine()
		b.NewLine()
		b.WriteIndentedLine(lvl, "Parameters")
		b.WriteIndentedLine(lvl, "----------")
		b.WriteIndentedLine(lvl, fmt.Sprintf("conn : %s", docstringConnType))
		b.WriteIndentedLine(lvl+1, "The connection object used to execute queries.")
		b.NewLine()
		b.WriteIndentedLine(lvl, `"""`)
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
		// TODO add this here after finishing asyncpg :execrows
	} else if query.Cmd == metadata.CmdExecResult {
		b.WriteIndentedLine(lvl, `"""`+fmt.Sprintf("Execute and return the result of SQL query with `name: %s %s`.", query.MethodName, query.Cmd))
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
			b.WriteIndentedLine(lvl, "Returns")
			b.WriteIndentedLine(lvl, "-------")
			b.WriteIndentedLine(lvl, returnType.Type)
			b.WriteIndentedLine(lvl+1, "The result returned when executing the query.")
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
			b.NewLine()
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl+1, fmt.Sprintf("The result of type `%s` returned when executing the query.", returnType.Type))
		} else if *docstringConfig == core.DocstringConventionPEP257 {
			b.WriteIndentedLine(lvl, "Arguments:")
			if docstringConnType != "" {
				b.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection object of type `%s` used to execute the query.", docstringConnType))
			}
			for _, arg := range queryArgs {
				b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- %s.", arg.Name, arg.Type))
			}
			b.NewLine()
			b.WriteIndentedLine(lvl, "Returns:")
			b.WriteIndentedLine(lvl+1, fmt.Sprintf("%s -- The result returned when executing the query.", returnType.Type))
		}
		b.WriteIndentedLine(lvl, `"""`)
	} else if query.Cmd == metadata.CmdExecLastId {
		// TODO add this here after finishing asyncpg :execlastid
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
			b.WriteIndentedLine(lvl, fmt.Sprintf("collections.abc.Sequence[%s]", returnType.Type))
			b.WriteIndentedLine(lvl+1, "Results fetched from the db.")
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
			b.WriteIndentedLine(lvl+1, fmt.Sprintf("Results of type `collections.abc.Sequence[%s]` fetched from the db.", returnType.Type))
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
			b.WriteIndentedLine(lvl, fmt.Sprintf("collections.abc.Sequence[%s] -- Results fetched from the db.", returnType.Type))
		}
		b.WriteIndentedLine(lvl, `"""`)
	}
}
