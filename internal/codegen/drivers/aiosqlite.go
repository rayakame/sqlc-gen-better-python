package drivers

import (
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/rayakame/sqlc-gen-better-python/internal/typeConversion"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"strconv"
	"strings"
)

const AioSQLiteConn = "aiosqlite.Connection"

func AioSQLiteBuildTypeConvFunc(queries []core.Query, body *builders.IndentStringBuilder, conf *core.Config) {
	// this function fucking got out of hand
	queryValueUses := func(name string, qv core.QueryValue) bool {
		if !qv.IsEmpty() {
			if qv.IsStruct() && qv.EmitStruct() {
				if val, sqlType := core.TableUses(name, *qv.Table); val {
					if typeConversion.SqliteDoTypeConversion(sqlType) {
						return true
					}
				}
			} else if qv.IsStruct() {
				if val, sqlType := core.TableUses(name, *qv.Table); val {
					if typeConversion.SqliteDoTypeConversion(sqlType) {
						return true
					}
				}
			} else {
				if qv.Typ.Type == name {
					if typeConversion.SqliteDoTypeConversion(qv.Typ.SqlType) {
						return true
					}
				}
			}
		}
		return false
	}
	toConvert := make(map[string]bool)
	for _, query := range queries {
		for sqlType, _ := range typeConversion.SqliteGetConversions() {
			name := types.SqliteTypeToPython(&plugin.GenerateRequest{}, &plugin.Column{Type: &plugin.Identifier{
				Catalog: "",
				Schema:  "",
				Name:    sqlType,
			}}, conf)
			if queryValueUses(name, query.Ret) {
				toConvert[name] = true
			}
			for _, arg := range query.Args {
				if queryValueUses(name, arg) {
					toConvert[name] = true
				}
			}
		}
	}
	adapters := make([]string, 0)
	converters := make([]string, 0)
	if _, found := toConvert["datetime.date"]; found {
		body.WriteLine("def _adapt_date(val: datetime.date) -> str:")
		body.WriteIndentedLine(1, "return val.isoformat()")
		body.NewLine()
		adapters = append(adapters, "aiosqlite.register_adapter(datetime.date, _adapt_date)")
		body.WriteLine("def _convert_date(val: bytes) -> datetime.date:")
		body.WriteIndentedLine(1, "return datetime.date.fromisoformat(val.decode())")
		body.NewLine()
		converters = append(converters, `aiosqlite.register_converter("date", _convert_date)`)
	}
	if _, found := toConvert["datetime.datetime"]; found {
		body.WriteLine("def _adapt_datetime(val: datetime.datetime) -> str:")
		body.WriteIndentedLine(1, "return val.isoformat()")
		body.NewLine()
		adapters = append(adapters, "aiosqlite.register_adapter(datetime.datetime, _adapt_datetime)")
		body.WriteLine("def _convert_datetime(val: bytes) -> datetime.datetime:")
		body.WriteIndentedLine(1, "return datetime.datetime.fromisoformat(val.decode())")
		body.NewLine()
		converters = append(converters, `aiosqlite.register_converter("datetime", _convert_datetime)`)
		converters = append(converters, `aiosqlite.register_converter("timestamp", _convert_datetime)`)
	}
	if _, found := toConvert["bool"]; found {
		body.WriteLine("def _adapt_bool(val: bool) -> int:")
		body.WriteIndentedLine(1, "return int(val)")
		body.NewLine()
		adapters = append(adapters, "aiosqlite.register_adapter(bool, _adapt_bool)")
		body.WriteLine("def _convert_bool(val: bytes) -> bool:")
		body.WriteIndentedLine(1, "return bool(int(val))")
		body.NewLine()
		converters = append(converters, `aiosqlite.register_converter("bool", _convert_bool)`)
		converters = append(converters, `aiosqlite.register_converter("boolean", _convert_bool)`)
	}
	for i, line := range adapters {
		body.WriteLine(line)
		if i == len(adapters)-1 {
			body.NewLine()
		}
	}
	for i, line := range converters {
		body.WriteLine(line)
		if i == len(converters)-1 {
			body.NewLine()
		}
	}
}

func AioSQLiteBuildPyQueryFunc(query *core.Query, body *builders.IndentStringBuilder, args []core.FunctionArg, retType core.PyType, isClass bool) error {
	indentLevel := 0
	params := fmt.Sprintf("conn: %s", AioSQLiteConn)
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
		body.WriteString(fmt.Sprintf(", %s", arg.FunctionFormat))
	}
	if query.Cmd == metadata.CmdExec {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("await %s.execute(%s", conn, query.ConstantName))
		aiosqliteWriteParams(query, body)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecResult {
		body.WriteLine(fmt.Sprintf(") -> %s:", "aiosqlite.Cursor"))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return await %s.execute(%s", conn, query.ConstantName))
		aiosqliteWriteParams(query, body)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecRows {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return await %s.execute(%s", conn, query.ConstantName))
		aiosqliteWriteParams(query, body)
		body.WriteLine(").rowcount")
	} else if query.Cmd == metadata.CmdExecLastId {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return await %s.execute(%s", conn, query.ConstantName))
		aiosqliteWriteParams(query, body)
		body.WriteLine(").lastrowid")
	} else if query.Cmd == metadata.CmdOne {
		body.WriteLine(fmt.Sprintf(") -> %s | None:", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("row = await (await %s.execute(%s", conn, query.ConstantName))
		aiosqliteWriteParams(query, body)
		body.WriteLine(")).fetchone()")
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
		body.WriteLine(fmt.Sprintf(") -> typing.AsyncIterator[%s]:", retType.Type))
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("stream = await %s.execute(%s", conn, query.ConstantName))
		aiosqliteWriteParams(query, body)
		body.WriteLine(")")
		body.WriteIndentedLine(indentLevel+1, "async for row in stream:")
		if query.Ret.IsStruct() {
			body.WriteIndentedString(indentLevel+2, fmt.Sprintf("yield %s(", retType.Type))
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
			body.WriteIndentedLine(indentLevel+2, fmt.Sprintf("yield %s(row[0])", retType.Type))
		}
	}
	return nil
}

func AioSQLiteAcceptedDriverCMDs() []string {
	return []string{
		metadata.CmdExec,
		metadata.CmdExecResult,
		metadata.CmdExecLastId,
		metadata.CmdExecRows,
		metadata.CmdOne,
		metadata.CmdMany,
	}
}

func aiosqliteWriteParams(query *core.Query, body *builders.IndentStringBuilder) {
	if len(query.Args) == 0 {
		return
	}
	params := "("
	for _, arg := range query.Args {
		if !arg.IsEmpty() {
			params += fmt.Sprintf("%s, ", arg.Name)
		}
	}
	body.WriteString("," + params + ")")
}
