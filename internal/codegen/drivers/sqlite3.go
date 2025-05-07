package drivers

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"strconv"
	"strings"
)

const SQLite3Conn = "sqlite3.Connection"

func SQLite3BuildPyQueryFunc(query *core.Query, body *builders.IndentStringBuilder, args []string, retType core.PyType, isClass bool) error {
	indentLevel := 0
	params := fmt.Sprintf("conn: %s", SQLite3Conn)
	conn := "conn"
	if isClass {
		params = "self"
		conn = "self._conn"
		indentLevel = 1
	}
	body.WriteIndentedString(indentLevel, fmt.Sprintf("def %s(%s", query.FuncName, params))
	for i, arg := range args {
		if i == 0 {
			body.WriteString(", *")
		}
		body.WriteString(fmt.Sprintf(", %s", arg))
	}
	if query.Cmd == metadata.CmdExec {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("%s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecResult {
		body.WriteLine(fmt.Sprintf(") -> %s:", "sqlite3.Cursor"))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecRows {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(").rowcount")
	} else if query.Cmd == metadata.CmdExecLastId {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(").lastrowid")
	} else if query.Cmd == metadata.CmdOne {
		body.WriteLine(fmt.Sprintf(") -> typing.Optional[%s]:", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("row = %s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(").fetchone()")
		body.WriteIndentedLine(indentLevel+1, "if row is None:")
		body.WriteIndentedLine(indentLevel+2, "return None")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s(", retType.Type))
			i := 0
			for _, col := range query.Ret.Table.Columns {
				if i != 0 {
					body.WriteString(", ")
				}
				if len(col.EmbedFields) != 0 {
					var inner []string
					body.WriteString(fmt.Sprintf("%s=%s(", col.Name, col.Type.Type))
					for _, embedCol := range col.EmbedFields {
						inner = append(inner, fmt.Sprintf("%s=row[%s]", embedCol.Name, strconv.Itoa(i)))
						i++
					}
					body.WriteString(strings.Join(inner, ", ") + ")")
				} else {
					body.WriteString(fmt.Sprintf("%s=row[%s]", col.Name, strconv.Itoa(i)))
					i++
				}
			}
			body.WriteLine(")")
		} else {
			body.WriteIndentedLine(indentLevel+1, fmt.Sprintf("return %s(row[0])", retType.Type))
		}
	} else if query.Cmd == metadata.CmdMany {
		body.WriteLine(fmt.Sprintf(") -> typing.List[%s]:", retType.Type))
		body.WriteIndentedLine(indentLevel+1, fmt.Sprintf("rows: typing.List[%s] = []", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("for row in %s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(").fetchall():")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(indentLevel+2, fmt.Sprintf("rows.append(%s(", retType.Type))
			i := 0
			for _, col := range query.Ret.Table.Columns {
				if i != 0 {
					body.WriteString(", ")
				}
				if len(col.EmbedFields) != 0 {
					var inner []string
					body.WriteString(fmt.Sprintf("%s=%s(", col.Name, col.Type.Type))
					for _, embedCol := range col.EmbedFields {
						inner = append(inner, fmt.Sprintf("%s=row[%s]", embedCol.Name, strconv.Itoa(i)))
						i++
					}
					body.WriteString(strings.Join(inner, ", ") + ")")
				} else {
					body.WriteString(fmt.Sprintf("%s=row[%s]", col.Name, strconv.Itoa(i)))
					i++
				}
			}
			body.WriteLine("))")
		} else {
			body.WriteIndentedLine(indentLevel+2, fmt.Sprintf("rows.append(%s(row[0]))", retType.Type))
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

/*
func SQLite3SkipTypeConversion() []string {
	return []string{}
}*/

func sqlite3WriteParams(query *core.Query, body *builders.IndentStringBuilder) {
	if len(query.Args) == 0 {
		return
	}
	params := "("
	for i, arg := range query.Args {
		if !arg.IsEmpty() {
			if i == len(query.Args)-1 && i != 0 {
				params += fmt.Sprintf("%s", arg.Name)
			} else {
				params += fmt.Sprintf("%s, ", arg.Name)
			}
		}
	}
	body.WriteString("," + params + ")")
}
