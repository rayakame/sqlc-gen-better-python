package internal

import (
	"context"
	"fmt"
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
	fileName, fileContent := log.GlobalLogger.Print()
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
