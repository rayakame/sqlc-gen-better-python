package codegen

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"strconv"
)

func cmdNotSupported(cmd string) bool {
	switch cmd {
	case metadata.CmdBatchOne:
		return true
	case metadata.CmdBatchMany:
		return true
	case metadata.CmdBatchExec:
		return true
	default:
		return false
	}
}

func buildAioSQLiteFunction(imp *core.Importer, query *core.Query, body *IndentStringBuilder) error {
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
	} else if query.Cmd == metadata.CmdExecResult {
		body.WriteLine(fmt.Sprintf(") -> %s:", "aiosqlite.Cursor"))
		body.WriteIndentedString(1, fmt.Sprintf("return await conn.execute(%s", query.ConstantName))
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
	} else if query.Cmd == metadata.CmdExecRows {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType))
		body.WriteIndentedString(1, fmt.Sprintf("return await conn.execute(%s", query.ConstantName))
		if argType != "" {
			if query.Arg.IsStruct() {
				for _, col := range query.Arg.Table.Columns {
					body.WriteString(fmt.Sprintf(", %s.%s", query.Arg.Name, col.Name))
				}
			} else {
				body.WriteString(fmt.Sprintf(", %s", query.Arg.Name))
			}
		}
		body.WriteLine(").rowcount")
	} else if query.Cmd == metadata.CmdExecLastId {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType))
		body.WriteIndentedString(1, fmt.Sprintf("return await conn.execute(%s", query.ConstantName))
		if argType != "" {
			if query.Arg.IsStruct() {
				for _, col := range query.Arg.Table.Columns {
					body.WriteString(fmt.Sprintf(", %s.%s", query.Arg.Name, col.Name))
				}
			} else {
				body.WriteString(fmt.Sprintf(", %s", query.Arg.Name))
			}
		}
		body.WriteLine(").lastrowid")
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
	} else if query.Cmd == metadata.CmdCopyFrom || cmdNotSupported(query.Cmd) {
		return fmt.Errorf("command %s is not supported by %s", query.Cmd, imp.C.SqlDriver)
	} else {
		body.WriteLine(fmt.Sprintf(") -> None:"))
		body.WriteIndentedLine(1, "return None")
	}
	return nil
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
		if query.Arg.Typ.IsList {
			argType = fmt.Sprintf("typing.Sequence[%s]", argType)
		}
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
	if query.Cmd == metadata.CmdExecLastId || query.Cmd == metadata.CmdExecRows {
		retType = "int"
	}
	return argType, retType
}

func BuildQueriesFile(imp *core.Importer, queries []core.Query, tables []core.Table) (string, []byte, error) {
	fileName := "queries.py"
	body := NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	body.WriteImportAnnotations()

	funcNames := make([]string, 0)
	queryBody := NewIndentStringBuilder(imp.C.IndentChar, imp.C.CharsPerIndentLevel)
	for i, query := range queries {
		funcNames = append(funcNames, query.FuncName)
		if i != 0 {
			queryBody.WriteString("\n\n")
		}
		buildQueryHeader(&query, queryBody)
		queryBody.WriteString("\n\n")

		err := buildAioSQLiteFunction(imp, &query, queryBody)
		if err != nil {
			return "", nil, err
		}
	}
	body.WriteLine("__all__: typing.Sequence[str] = (")
	for _, n := range funcNames {
		body.WriteIndentedLine(1, fmt.Sprintf("\"%s\",", n))
	}
	body.WriteLine(")")
	body.WriteString("\n")
	for _, imp := range imp.Imports(fileName) {
		body.WriteLine(imp)
	}
	body.WriteString("\n")

	return fileName, []byte(body.String() + queryBody.String()), nil
}
