package drivers

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"strconv"
)

func BuildPyQueryFunc(query *core.Query, body *builders.IndentStringBuilder, argType string, retType string) error {
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
	}
	return nil
}

func AcceptedDriverCMDs() []string {
	return []string{
		metadata.CmdExec,
		metadata.CmdExecResult,
		metadata.CmdExecLastId,
		metadata.CmdExecRows,
		metadata.CmdOne,
		metadata.CmdMany,
	}
}
