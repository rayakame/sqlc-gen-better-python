package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/rayakame/sqlc-gen-better-python/internal/log"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type PythonGenerator struct {
	req    *plugin.GenerateRequest
	config *core.Config

	typeConversionFunc types.TypeConversionFunc
}

func NewPythonGenerator(req *plugin.GenerateRequest) (*PythonGenerator, error) {
	config, err := core.ParseConfig(req)
	if err != nil {
		return nil, err
	}
	if err = core.ValidateConf(config); err != nil {
		return nil, err
	}
	var typeConversionFunc types.TypeConversionFunc
	switch req.Settings.Engine {
	case "postgresql":
		typeConversionFunc = types.PostgresTypeToPython
	case "sqlite":
		typeConversionFunc = types.SqliteTypeToPython
	default:
		return nil, fmt.Errorf("engine %q is not supported", req.Settings.Engine)
	}

	return &PythonGenerator{
		req:                req,
		config:             config,
		typeConversionFunc: typeConversionFunc,
	}, nil
}

func (pg *PythonGenerator) Run() (*plugin.GenerateResponse, error) {
	outputFiles := make([]*plugin.File, 0)
	log.GlobalLogger.Log(pg.req.String())
	log.GlobalLogger.LogByte(pg.req.PluginOptions)
	enums := pg.buildEnums()
	tables := pg.buildTables()
	queries, err := pg.buildQueries(tables)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(pg.config)
	log.GlobalLogger.LogByte(jsonData)
	jsonData, _ = json.Marshal(enums)
	log.GlobalLogger.LogByte(jsonData)
	jsonData, _ = json.Marshal(tables)
	log.GlobalLogger.LogByte(jsonData)
	jsonData, _ = json.Marshal(queries)
	log.GlobalLogger.LogByte(jsonData)

	if pg.config.OmitUnusedStructs {
		enums, tables = filterUnusedStructs(enums, tables, queries)
	}
	if err := pg.validate(enums, tables, queries); err != nil {
		return nil, err
	}
	fileName, fileContent := log.GlobalLogger.Print()
	outputFiles = append(outputFiles, &plugin.File{
		Name:     fileName,
		Contents: fileContent,
	})
	fileName, fileContent, _ = codegen.BuildModelFile(pg.config, tables)
	outputFiles = append(outputFiles, &plugin.File{
		Name:     fileName,
		Contents: fileContent,
	})
	return &plugin.GenerateResponse{Files: outputFiles}, nil
}

func Generate(_ context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	pythonGenerator, err := NewPythonGenerator(req)
	if err != nil {
		return nil, err
	}
	return pythonGenerator.Run()
}

func (gen *PythonGenerator) validate(enums []core.Enum, structs []core.Table, queries []core.Query) error {
	enumNames := make(map[string]struct{})
	for _, enum := range enums {
		enumNames[enum.Name] = struct{}{}
		enumNames["Null"+enum.Name] = struct{}{}
	}
	structNames := make(map[string]struct{})
	for _, struckt := range structs {
		if _, ok := enumNames[struckt.Name]; ok {
			return fmt.Errorf("struct name conflicts with enum name: %s", struckt.Name)
		}
		structNames[struckt.Name] = struct{}{}
	}
	return nil
}

func filterUnusedStructs(enums []core.Enum, tables []core.Table, queries []core.Query) ([]core.Enum, []core.Table) {
	keepTypes := make(map[string]struct{})

	for _, query := range queries {
		if !query.Arg.IsEmpty() {
			keepTypes[query.Arg.Type()] = struct{}{}
			if query.Arg.IsStruct() {
				for _, field := range query.Arg.Table.Columns {
					keepTypes[field.Type.Type] = struct{}{}
				}
			}
		}
		if query.HasRetType() {
			keepTypes[query.Ret.Type()] = struct{}{}
			if query.Ret.IsStruct() {
				for _, field := range query.Ret.Table.Columns {
					keepTypes[field.Type.Type] = struct{}{}
					for _, embedField := range field.EmbedFields {
						keepTypes[embedField.Type.Type] = struct{}{}
					}
				}
			}
		}
	}

	keepEnums := make([]core.Enum, 0, len(enums))
	for _, enum := range enums {
		_, keep := keepTypes[enum.Name]
		_, keepNull := keepTypes["Null"+enum.Name]
		if keep || keepNull {
			keepEnums = append(keepEnums, enum)
		}
	}

	keepStructs := make([]core.Table, 0, len(tables))
	for _, st := range tables {
		if _, ok := keepTypes[st.Name]; ok {
			keepStructs = append(keepStructs, st)
		}
	}

	return keepEnums, keepStructs
}
