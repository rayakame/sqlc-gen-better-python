package internal

import (
	"context"
	"fmt"

	configPackage "github.com/rayakame/sqlc-gen-better-python/internal/config"
	driverPackage "github.com/rayakame/sqlc-gen-better-python/internal/driver"
	"github.com/rayakame/sqlc-gen-better-python/internal/log"
	"github.com/rayakame/sqlc-gen-better-python/internal/render"
	"github.com/rayakame/sqlc-gen-better-python/internal/transform"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func Handler(_ context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	config, err := configPackage.NewConfig(req)
	if err != nil {
		return nil, fmt.Errorf("error trying to parse config: %w", err)
	}

	typeConversionFunc, err := types.GetTypeConversionFunc(req.Settings.Engine)
	if err != nil {
		return nil, fmt.Errorf("error trying to parse config: %w", err)
	}

	driver, err := driverPackage.New(config)
	if err != nil {
		return nil, fmt.Errorf("error trying to parse config: %w", err)
	}

	transformer := transform.NewTransformer(config, req, typeConversionFunc)
	enums := transformer.BuildEnums()
	tables := transformer.BuildTables()
	queries := transformer.BuildQueries(tables)

	if config.OmitUnusedModels {
		enums, tables = transform.FilterUnusedModels(enums, tables, queries)
	}

	renderer := render.New(config, driver)
	/*
		log.L().LogAny(enums)
		log.L().LogAny(tables)
		log.L().LogAny(queries)
	*/
	outputFiles, err := renderer.RenderAll(enums, tables, queries)
	if err != nil {
		return nil, fmt.Errorf("error building queries: %w", err)
	}
	if config.Debug {
		fileName, fileContent := log.L().Export()
		outputFiles = append(outputFiles, &plugin.File{
			Name:     fileName,
			Contents: fileContent,
		})
	}

	return &plugin.GenerateResponse{Files: outputFiles}, nil
}
