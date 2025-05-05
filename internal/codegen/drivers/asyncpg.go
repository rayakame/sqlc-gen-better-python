package drivers

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"strconv"
	"strings"
)

const AsyncpgConn = "asyncpg.Connection"

func AsyncpgBuildPyQueryFunc(query *core.Query, body *builders.IndentStringBuilder, args []string, retType string, isClass bool) error {
	indentLevel := 0
	params := fmt.Sprintf("conn: %s", AsyncpgConn)
	conn := "conn"
	if isClass {
		params = "self"
		conn = "self._conn"
		indentLevel = 1
	}
	body.WriteIndentedString(indentLevel, fmt.Sprintf("async def %s(%s", query.FuncName, params))
	for i, arg := range args {
		if i == 0 {
			body.WriteString(", *")
		}
		body.WriteString(fmt.Sprintf(", %s", arg))
	}
	if query.Cmd == metadata.CmdExec {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("await %s.execute(%s", conn, query.ConstantName))
		asyncpgWriteParams(query, body)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdOne {
		body.WriteLine(fmt.Sprintf(") -> typing.Optional[%s]:", retType))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("row = await %s.fetchrow(%s", conn, query.ConstantName))
		asyncpgWriteParams(query, body)
		body.WriteLine(")")
		body.WriteIndentedLine(indentLevel+1, "if row is None:")
		body.WriteIndentedLine(indentLevel+2, "return None")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s(", retType))
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
			body.WriteIndentedLine(indentLevel+1, fmt.Sprintf("return %s(row[0])", retType))
		}
	} else if query.Cmd == metadata.CmdMany {
		body.WriteLine(fmt.Sprintf(") -> typing.Sequence[%s]:", retType))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("rows = await %s.fetch(%s", conn, query.ConstantName))
		asyncpgWriteParams(query, body)
		body.WriteLine(")")
		body.WriteIndentedLine(indentLevel+1, fmt.Sprintf("return_rows: typing.List[%s] = []", retType))
		body.WriteIndentedLine(indentLevel+1, "for row in rows:")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(indentLevel+2, fmt.Sprintf("rows.append(%s(", retType))
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
			body.WriteIndentedString(indentLevel+1, "return return_rows")
		} else {
			body.WriteIndentedLine(indentLevel+2, fmt.Sprintf("return_rows.append(%s(row[0]))", retType))
			body.WriteIndentedString(indentLevel+1, "return return_rows")
		}
	}
	return nil
}

func AsyncpgAcceptedDriverCMDs() []string {
	return []string{
		metadata.CmdExec,
		metadata.CmdOne,
		metadata.CmdMany,
	}
}

func asyncpgWriteParams(query *core.Query, body *builders.IndentStringBuilder) {
	if len(query.Args) == 0 {
		return
	}
	params := ""
	for i, arg := range query.Args {
		if !arg.IsEmpty() {
			if i == len(query.Args)-1 {
				params += fmt.Sprintf(" %s", arg.Name)
			} else {
				params += fmt.Sprintf(" %s,", arg.Name)
			}
		}
	}
	body.WriteString("," + params)
}
