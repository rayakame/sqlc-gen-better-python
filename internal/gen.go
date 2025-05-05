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
	"strings"
)

type PythonGenerator struct {
	req    *plugin.GenerateRequest
	config *core.Config

	typeConversionFunc types.TypeConversionFunc
	sqlDriver          *codegen.Driver
}

func NewPythonGenerator(req *plugin.GenerateRequest) (*PythonGenerator, error) {
	config, err := core.ParseConfig(req)
	if err != nil {
		return nil, err
	}
	if err = core.ValidateConf(config, req.Settings.Engine); err != nil {
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

	sqlDriver, err := codegen.NewDriver(config)
	if err != nil {
		return nil, err
	}

	return &PythonGenerator{
		req:                req,
		config:             config,
		typeConversionFunc: typeConversionFunc,
		sqlDriver:          sqlDriver,
	}, nil
}

func (gen *PythonGenerator) Run() (*plugin.GenerateResponse, error) {
	outputFiles := make([]*plugin.File, 0)
	log.GlobalLogger.LogByte(gen.req.PluginOptions)
	enums := gen.buildEnums()
	tables := gen.buildTables()
	queries, err := gen.buildQueries(tables)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(gen.req)
	log.GlobalLogger.LogByte(jsonData)
	jsonData, _ = json.Marshal(gen.config)
	log.GlobalLogger.LogByte(jsonData)
	jsonData, _ = json.Marshal(enums)
	log.GlobalLogger.LogByte(jsonData)
	jsonData, _ = json.Marshal(tables)
	log.GlobalLogger.LogByte(jsonData)
	jsonData, _ = json.Marshal(queries)
	log.GlobalLogger.LogByte(jsonData)

	if gen.config.OmitUnusedModels {
		enums, tables = filterUnusedStructs(enums, tables, queries)
	}
	if err := gen.validate(enums, tables); err != nil {
		return nil, err
	}
	importer := core.Importer{
		Tables:  tables,
		Queries: queries,
		Enums:   enums,
		C:       gen.config,
	}
	if file, err := gen.sqlDriver.BuildPyTablesFile(&importer, tables); err != nil {
		return nil, err
	} else {
		outputFiles = append(outputFiles, file)
	}
	if files, err := gen.sqlDriver.BuildPyQueriesFiles(&importer, queries); err != nil {
		return nil, err
	} else {
		outputFiles = append(outputFiles, files...)
	}
	if *gen.config.EmitInitFile {
		outputFiles = append(outputFiles, gen.sqlDriver.BuildInitFile())
	}
	jsonData, _ = json.Marshal(outputFiles)
	log.GlobalLogger.LogByte(jsonData)
	if gen.config.Debug {
		fileName, fileContent := log.GlobalLogger.Print()
		outputFiles = append(outputFiles, &plugin.File{
			Name:     fileName,
			Contents: fileContent,
		})
	}
	return &plugin.GenerateResponse{Files: outputFiles}, nil
}

func Generate(_ context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	pythonGenerator, err := NewPythonGenerator(req)
	if err != nil {
		return nil, err
	}
	return pythonGenerator.Run()
}

func (gen *PythonGenerator) validate(enums []core.Enum, structs []core.Table) error {
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
		for _, arg := range query.Args {
			if !arg.IsEmpty() {
				keepTypes[arg.Type()] = struct{}{}
			}
		}
		if query.HasRetType() {
			keepTypes[query.Ret.Type()] = struct{}{}
			if query.Ret.IsStruct() {
				for _, field := range query.Ret.Table.Columns {
					keepTypes[strings.ReplaceAll(field.Type.Type, "models.", "")] = struct{}{}
					for _, embedField := range field.EmbedFields {
						keepTypes[strings.ReplaceAll(embedField.Type.Type, "models.", "")] = struct{}{}
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
