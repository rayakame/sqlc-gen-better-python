package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"strconv"
)

func buildAioSQLiteFunction(imp *core.Importer, query *core.Query, body *IndentStringBuilder) {
	argType, retType := prepareFunctionHeader(imp, query, body)
	body.WriteString(fmt.Sprintf("async def %s(conn: aiosqlite.Connection", query.FuncName))
	if argType != "" {
		body.WriteString(fmt.Sprintf(", %s: %s", query.Arg.Name, argType))
	}
	if query.Cmd == metadata.CmdExec {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType))
		body.WriteIndentedString(1, fmt.Sprintf("await conn.execute(%s", query.ConstantName))
		if argType != "" {
			if query.Arg.IsStruct() {
				for _, col := range query.Arg.Table.Columns {
					body.WriteString(fmt.Sprintf(", %s.%s", query.Arg.Name, col.Name))
				}
			} else {
				body.WriteString(fmt.Sprintf(", %s", query.Arg.Name))
			}
		}
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdOne {
		body.WriteLine(fmt.Sprintf(") -> typing.Optional[%s]:", retType))
		body.WriteIndentedString(1, fmt.Sprintf("row = await (await conn.execute(%s", query.ConstantName))
		if argType != "" {
			if query.Arg.IsStruct() {
				for _, col := range query.Arg.Table.Columns {
					body.WriteString(fmt.Sprintf(", %s.%s", query.Arg.Name, col.Name))
				}
			} else {
				body.WriteString(fmt.Sprintf(", %s", query.Arg.Name))
			}
		}
		body.WriteLine(")).fetchone()")
		body.WriteIndentedLine(1, "if row is None:")
		body.WriteIndentedLine(2, "return None")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(1, fmt.Sprintf("return %s(", retType))
			for i, col := range query.Ret.Table.Columns {
				if i != 0 {
					body.WriteString(", ")
				}
				body.WriteString(fmt.Sprintf("%s=row[%s]", col.Name, strconv.Itoa(i)))
			}
			body.WriteLine(")")
		} else {
			body.WriteIndentedLine(1, fmt.Sprintf("return %s(row[0])", retType))
		}
	} else if query.Cmd == metadata.CmdMany {
		body.WriteLine(fmt.Sprintf(") -> typing.AsyncIterator[%s]:", retType))
		body.WriteIndentedString(1, fmt.Sprintf("stream = await conn.execute(%s", query.ConstantName))
		if argType != "" {
			if query.Arg.IsStruct() {
				for _, col := range query.Arg.Table.Columns {
					body.WriteString(fmt.Sprintf(", %s.%s", query.Arg.Name, col.Name))
				}
			} else {
				body.WriteString(fmt.Sprintf(", %s", query.Arg.Name))
			}
		}
		body.WriteLine(")")
		body.WriteIndentedLine(1, "async for row in stream:")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(2, fmt.Sprintf("yield %s(", retType))
			for i, col := range query.Ret.Table.Columns {
				if i != 0 {
					body.WriteString(", ")
				}
				body.WriteString(fmt.Sprintf("%s=row[%s]", col.Name, strconv.Itoa(i)))
			}
			body.WriteLine(")")
		} else {
			body.WriteIndentedLine(2, fmt.Sprintf("yield %s(row[0])", retType))
		}
	} else {
		body.WriteLine(fmt.Sprintf(") -> None:"))
		body.WriteIndentedLine(1, "return None")
	}
}

func buildQueryHeader(query *core.Query, body *IndentStringBuilder) {
	body.WriteLine(fmt.Sprintf(`%s = """-- name: %s %s`, query.ConstantName, query.MethodName, query.Cmd))
	body.WriteLine(query.SQL)
	body.WriteLine(`"""`)
}

func prepareFunctionHeader(imp *core.Importer, query *core.Query, body *IndentStringBuilder) (string, string) {
	argType := ""
	if query.Arg.EmitStruct() && query.Arg.IsStruct() {
		BuildModel(imp.C, query.Arg.Table, body)
		body.WriteString("\n\n")
		argType = query.Arg.Table.Name
	} else if !query.Arg.IsEmpty() {
		argType = query.Arg.Typ.Type
	}
	retType := "None"
	if query.Ret.EmitStruct() && query.Ret.IsStruct() {
		BuildModel(imp.C, query.Ret.Table, body)
		body.WriteString("\n\n")
		retType = query.Ret.Table.Name
	} else if !query.Ret.IsEmpty() {
		if query.Ret.IsStruct() {
			retType = fmt.Sprintf("models.%s", query.Ret.Table.Name)
		} else {
			retType = query.Ret.Typ.Type
		}
	}
	return argType, retType
}

func BuildQueriesFile(imp *core.Importer, queries []core.Query, tables []core.Table) (string, []byte, error) {
	fileName := "queries.py"
	body := NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	body.WriteImportAnnotations()
	for _, imp := range imp.Imports(fileName) {
		body.WriteLine(imp)
	}
	body.WriteString("\n")

	for i, query := range queries {
		if i != 0 {
			body.WriteString("\n\n")
		}
		buildQueryHeader(&query, body)
		body.WriteString("\n\n")

		buildAioSQLiteFunction(imp, &query, body)
	}

	return fileName, []byte(body.String()), nil
}
