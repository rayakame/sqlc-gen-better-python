package builders

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

var docstringConfig *string

func SetDocstringConfig(c *string) {
	docstringConfig = c
}

func (b *IndentStringBuilder) WriteQueryClassDocstring(lvl int, sourceName string) {
	if docstringConfig == nil {
		return
	}
	b.WriteIndentedString(lvl, fmt.Sprintf(`"""Queries from file %s.`, sourceName))
	if *docstringConfig == core.DocstringConventionNumpy {
		b.NewLine()
		b.NewLine()
		b.WriteIndentedLine(lvl, "Parameters")
		b.WriteIndentedLine(lvl, "----------")
		b.WriteIndentedLine(lvl, "conn")
		b.WriteIndentedLine(lvl+1, "The connection object used to execute queries.")
		b.NewLine()
		b.WriteIndentedLine(lvl, `"""`)
	} else {
		b.WriteLine(`"""`)
	}
	b.NewLine()
}

func (b *IndentStringBuilder) WriteQueryClassInitDocstring(lvl int) {
	if docstringConfig == nil {
		return
	}
	b.WriteIndentedString(lvl, fmt.Sprintf(`"""Initializes the instance using the connection.`))
	if *docstringConfig == core.DocstringConventionNumpy {
		b.WriteLine(`"""`)
	} else if *docstringConfig == core.DocstringConventionGoogle {
		b.NNewLine(2)
		b.WriteIndentedLine(lvl, "Args:")
		b.WriteIndentedLine(lvl+1, "conn: Connection object used to execute queries.")
		b.WriteIndentedLine(lvl, `"""`)
	} else if *docstringConfig == core.DocstringConventionPEP257 {
		b.NNewLine(2)
		b.WriteIndentedLine(lvl, "Arguments:")
		b.WriteIndentedLine(lvl, "conn -- Connection object used to execute queries.")
		b.WriteIndentedLine(lvl, `"""`)
	}
}

func (b *IndentStringBuilder) WriteQueryFileModuleDocstring(sourceName string) {
	if docstringConfig == nil {
		return
	}
	b.WriteLine(fmt.Sprintf(`"""Module containing queries from file %s."""`, sourceName))
}

func (b *IndentStringBuilder) WriteInitFileModuleDocstring() {
	if docstringConfig == nil {
		return
	}
	b.WriteLine(`"""Package containing queries and models automatically generated using sqlc-gen-better-python."""`)
}

func (b *IndentStringBuilder) WriteQueryFunctionDocstring(lvl int, query *core.Query, docstringConnType string, queryArgs []core.FunctionArg) {
	if docstringConfig == nil {
		return
	}

	if query.Cmd == metadata.CmdExec {
		b.WriteIndentedLine(lvl, fmt.Sprintf(`"""Execute SQL query with name: %s`, query.MethodName))
		b.NewLine()
		if *docstringConfig == core.DocstringConventionNumpy {
			b.WriteIndentedLine(lvl, "Parameters")
			b.WriteIndentedLine(lvl, "----------")
			if docstringConnType != "" {
				b.WriteIndentedLine(lvl, fmt.Sprintf("conn : %s", docstringConnType))
				b.WriteIndentedLine(lvl+1, fmt.Sprintf("Connection oject of type `%s` used to execute the query.", docstringConnType))
			}
			for _, arg := range queryArgs {
				b.WriteIndentedLine(lvl, fmt.Sprintf("%s : %s", arg.Name, arg.Type))
			}
			b.NewLine()
		} else if *docstringConfig == core.DocstringConventionGoogle {
			b.WriteIndentedLine(lvl, "Args:")
			if docstringConnType != "" {
				b.WriteIndentedLine(lvl, "conn:")
				b.WriteIndentedLine(lvl+1, fmt.Sprintf("Connection oject of type `%s` used to execute the query.", docstringConnType))
			}
			for _, arg := range queryArgs {
				b.WriteIndentedLine(lvl, fmt.Sprintf("%s: %s.", arg.Name, arg.Type))
			}
		} else if *docstringConfig == core.DocstringConventionPEP257 {
			b.WriteIndentedLine(lvl, "Arguments:")
			if docstringConnType != "" {
				b.WriteIndentedLine(lvl, fmt.Sprintf("conn -- Connection oject of type `%s` used to execute the query.", docstringConnType))
			}
			for _, arg := range queryArgs {
				b.WriteIndentedLine(lvl, fmt.Sprintf("%s -- %s.", arg.Name, arg.Type))
			}
		}
		b.WriteIndentedLine(lvl, `"""`)
	}

}
