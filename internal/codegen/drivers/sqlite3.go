package drivers

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"strconv"
)

const SQLite3Conn = "sqlite3.Connection"

func SQLite3BuildPyQueryFunc(query *core.Query, body *builders.IndentStringBuilder, argType string, retType string, isClass bool) error {
	indentLevel := 0
	params := fmt.Sprintf("conn: %s", SQLite3Conn)
	conn := "conn"
	if isClass {
		params = "self"
		conn = "self._conn"
		indentLevel = 1
	}
	body.WriteIndentedString(indentLevel, fmt.Sprintf("def %s(%s", query.FuncName, params))
	if argType != "" {
		body.WriteString(fmt.Sprintf(", %s: %s", query.Arg.Name, argType))
	}
	if query.Cmd == metadata.CmdExec {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("%s.execute(%s", conn, query.ConstantName))
		writeParams(query, body, argType)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecResult {
		body.WriteLine(fmt.Sprintf(") -> %s:", "sqlite3.Cursor"))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("%s.execute(%s", conn, query.ConstantName))
		writeParams(query, body, argType)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecRows {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("%s.execute(%s", conn, query.ConstantName))
		writeParams(query, body, argType)
		body.WriteLine(").rowcount")
	} else if query.Cmd == metadata.CmdExecLastId {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("%s.execute(%s", conn, query.ConstantName))
		writeParams(query, body, argType)
		body.WriteLine(").lastrowid")
	} else if query.Cmd == metadata.CmdOne {
		body.WriteLine(fmt.Sprintf(") -> typing.Optional[%s]:", retType))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("row = %s.execute(%s", conn, query.ConstantName))
		writeParams(query, body, argType)
		body.WriteLine(").fetchone()")
		body.WriteIndentedLine(indentLevel+1, "if row is None:")
		body.WriteIndentedLine(indentLevel+2, "return None")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s(", retType))
			for i, col := range query.Ret.Table.Columns {
				if i != 0 {
					body.WriteString(", ")
				}
				body.WriteString(fmt.Sprintf("%s=row[%s]", col.Name, strconv.Itoa(i)))
			}
			body.WriteLine(")")
		} else {
			body.WriteIndentedLine(indentLevel+1, fmt.Sprintf("return %s(row[0])", retType))
		}
	} else if query.Cmd == metadata.CmdMany {
		body.WriteLine(fmt.Sprintf(") -> typing.List[%s]:", retType))
		body.WriteIndentedLine(indentLevel+1, fmt.Sprintf("rows: typing.List[%s] = []", retType))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("for row in %s.execute(%s", conn, query.ConstantName))
		writeParams(query, body, argType)
		body.WriteLine(").fetchall():")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(indentLevel+2, fmt.Sprintf("rows.append(%s(", retType))
			for i, col := range query.Ret.Table.Columns {
				if i != 0 {
					body.WriteString(", ")
				}
				body.WriteString(fmt.Sprintf("%s=row[%s]", col.Name, strconv.Itoa(i)))
			}
			body.WriteLine("))")
		} else {
			body.WriteIndentedLine(indentLevel+2, fmt.Sprintf("rows.append(%s(row[0]))", retType))
		}
		body.WriteIndentedLine(indentLevel+1, "return rows")
	}
	return nil
}

func SQLite3AcceptedDriverCMDs() []string {
	return []string{
		metadata.CmdExec,
		metadata.CmdExecResult,
		metadata.CmdExecLastId,
		metadata.CmdExecRows,
		metadata.CmdOne,
		metadata.CmdMany,
	}
}

func writeParams(query *core.Query, body *builders.IndentStringBuilder, argType string) {
	if argType != "" {
		params := "("
		if query.Arg.IsStruct() {
			for _, col := range query.Arg.Table.Columns {
				params += fmt.Sprintf("%s.%s, ", query.Arg.Name, col.Name)
			}
		} else {
			params += fmt.Sprintf("%s, ", query.Arg.Name)
		}
		body.WriteString("," + params + ")")
	}
}
