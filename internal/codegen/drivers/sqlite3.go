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

const Sqlite3Result = "sqlite3.Row"
const SQLite3Conn = "sqlite3.Connection"

func SQLite3BuildTypeConvFunc(queries []core.Query, body *builders.IndentStringBuilder, conf *core.Config) {
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
		body.NNewLine(2)
		adapters = append(adapters, "sqlite3.register_adapter(datetime.date, _adapt_date)")
		body.WriteLine("def _convert_date(val: bytes) -> datetime.date:")
		if conf.Speedups {
			body.WriteIndentedLine(1, "return ciso8601.parse_datetime(val.decode()).date()")
		} else {
			body.WriteIndentedLine(1, "return datetime.date.fromisoformat(val.decode())")
		}
		body.NNewLine(2)
		converters = append(converters, `sqlite3.register_converter("date", _convert_date)`)
	}
	if _, found := toConvert["decimal.Decimal"]; found {
		body.WriteLine("def _adapt_decimal(val: decimal.Decimal) -> str:")
		body.WriteIndentedLine(1, "return str(val)")
		body.NNewLine(2)
		adapters = append(adapters, "sqlite3.register_adapter(decimal.Decimal, _adapt_decimal)")
		body.WriteLine("def _convert_decimal(val: bytes) -> decimal.Decimal:")
		body.WriteIndentedLine(1, "return decimal.Decimal(val.decode())")
		body.NNewLine(2)
		converters = append(converters, `sqlite3.register_converter("decimal", _convert_decimal)`)
	}
	if _, found := toConvert["datetime.datetime"]; found {
		body.WriteLine("def _adapt_datetime(val: datetime.datetime) -> str:")
		body.WriteIndentedLine(1, "return val.isoformat()")
		body.NNewLine(2)
		adapters = append(adapters, "sqlite3.register_adapter(datetime.datetime, _adapt_datetime)")
		body.WriteLine("def _convert_datetime(val: bytes) -> datetime.datetime:")
		if conf.Speedups {
			body.WriteIndentedLine(1, "return ciso8601.parse_datetime(val.decode())")
		} else {
			body.WriteIndentedLine(1, "return datetime.datetime.fromisoformat(val.decode())")
		}
		body.NNewLine(2)
		converters = append(converters, `sqlite3.register_converter("datetime", _convert_datetime)`)
		converters = append(converters, `sqlite3.register_converter("timestamp", _convert_datetime)`)
	}
	if _, found := toConvert["bool"]; found {
		body.WriteLine("def _adapt_bool(val: bool) -> int:")
		body.WriteIndentedLine(1, "return int(val)")
		body.NNewLine(2)
		adapters = append(adapters, "sqlite3.register_adapter(bool, _adapt_bool)")
		body.WriteLine("def _convert_bool(val: bytes) -> bool:")
		body.WriteIndentedLine(1, "return bool(int(val))")
		body.NNewLine(2)
		converters = append(converters, `sqlite3.register_converter("bool", _convert_bool)`)
		converters = append(converters, `sqlite3.register_converter("boolean", _convert_bool)`)
	}
	if _, found := toConvert["memoryview"]; found {
		body.WriteLine("def _adapt_memoryview(val: memoryview) -> bytes:")
		body.WriteIndentedLine(1, "return val.tobytes()")
		body.NNewLine(2)
		adapters = append(adapters, "sqlite3.register_adapter(memoryview, _adapt_memoryview)")
		body.WriteLine("def _convert_memoryview(val: bytes) -> memoryview:")
		body.WriteIndentedLine(1, "return memoryview(val)")
		body.NNewLine(2)
		converters = append(converters, `sqlite3.register_converter("blob", _convert_memoryview)`)
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

func SQLite3BuildQueryResults(body *builders.IndentStringBuilder) string {
	body.WriteSyncQueryResultsClassHeader(SQLite3Conn, []string{
		"self._cursor: sqlite3.Cursor | None = None",
		fmt.Sprintf("self._iterator: collections.abc.Iterator[%s] | None = None", Sqlite3Result),
	}, Sqlite3Result)
	body.WriteQueryResultsCallFunction([]string{
		"result = self._conn.execute(self._sql, self._args).fetchall()",
		"return [self._decode_hook(row) for row in result]",
	})
	body.NewLine()
	body.WriteIndentedLine(1, "def __next__(self) -> T:")
	body.WriteQueryResultsNextDocstringSqlite()
	body.WriteIndentedLine(2, "if self._cursor is None or self._iterator is None:")
	body.WriteIndentedLine(3, "self._cursor: sqlite3.Cursor | None = self._conn.execute(self._sql, self._args)")
	body.WriteIndentedLine(3, "self._iterator = self._cursor.__iter__()")
	body.WriteIndentedLine(2, "try:")
	body.WriteIndentedLine(3, "record = self._iterator.__next__()")
	body.WriteIndentedLine(2, "except StopIteration:")
	body.WriteIndentedLine(3, "self._cursor = None")
	body.WriteIndentedLine(3, "self._iterator = None")
	body.WriteIndentedLine(3, "raise")
	body.WriteIndentedLine(2, "return self._decode_hook(record)")
	return "QueryResults"
}

func SQLite3BuildPyQueryFunc(query *core.Query, body *builders.IndentStringBuilder, args []core.FunctionArg, retType core.PyType, isClass bool) error {
	indentLevel := 0
	params := fmt.Sprintf("conn: %s", SQLite3Conn)
	conn := "conn"
	docstringConnType := SQLite3Conn
	if isClass {
		params = "self"
		conn = "self._conn"
		indentLevel = 1
		docstringConnType = ""
	}
	body.WriteIndentedString(indentLevel, fmt.Sprintf("def %s(%s", query.FuncName, params))
	for i, arg := range args {
		if i == 0 {
			body.WriteString(", *")
		}
		body.WriteString(fmt.Sprintf(", %s", arg.FunctionFormat))
	}
	if query.Cmd == metadata.CmdExec {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, retType)
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("%s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecResult {
		body.WriteLine(fmt.Sprintf(") -> %s:", "sqlite3.Cursor"))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, core.PyType{Type: "sqlite3.Cursor"})
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(")")
	} else if query.Cmd == metadata.CmdExecRows {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, retType)
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(").rowcount")
	} else if query.Cmd == metadata.CmdExecLastId {
		body.WriteLine(fmt.Sprintf(") -> %s:", retType.Type))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, retType)
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return %s.execute(%s", conn, query.ConstantName))
		sqlite3WriteParams(query, body)
		body.WriteLine(").lastrowid")
	} else if query.Cmd == metadata.CmdOne {
		body.WriteLine(fmt.Sprintf(") -> %s | None:", retType.Type))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, retType)
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
			body.WriteIndentedLine(indentLevel+1, "return row[0]")
		}
	} else if query.Cmd == metadata.CmdMany {
		body.WriteLine(fmt.Sprintf(") -> QueryResults[%s]:", retType.Type))
		body.WriteQueryFunctionDocstring(indentLevel+1, query, docstringConnType, args, retType)
		decode_hook := "_decode_hook"
		if !query.Ret.IsStruct() {
			decode_hook = "operator.itemgetter(0)"
		} else {
			body.WriteIndentedLine(indentLevel+1, fmt.Sprintf("def _decode_hook(row: %s) -> %s:", Sqlite3Result, retType.Type))
			body.WriteIndentedString(indentLevel+2, fmt.Sprintf("return %s(", retType.Type))
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
		}
		body.WriteIndentedString(indentLevel+1, fmt.Sprintf("return QueryResults[%s](%s, %s, %s", retType.Type, conn, query.ConstantName, decode_hook))
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
		if params != "" {
			body.WriteString("," + params)
		}
		body.WriteLine(")")
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
	body.WriteString(", " + params + ")")
}
