package internal

import (
	"context"
	"fmt"

	configPackage "github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/log"
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

	transformer := transform.NewTransformer(config, req, typeConversionFunc)
	enums := transformer.BuildEnums()
	tables := transformer.BuildTables()
	queries := transformer.BuildQueries()

	log.L().LogAny(enums)
	log.L().LogAny(tables)
	log.L().LogAny(queries)

	outputFiles := make([]*plugin.File, 0)
	if config.Debug {
		fileName, fileContent := log.L().Export()
		outputFiles = append(outputFiles, &plugin.File{
			Name:     fileName,
			Contents: fileContent,
		})
	}

	return &plugin.GenerateResponse{Files: outputFiles}, nil
}
