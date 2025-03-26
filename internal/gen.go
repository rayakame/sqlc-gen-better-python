package internal

import (
	"context"
	"github.com/rayakame/sqlc-gen-better-python/plugin/log"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	outputFiles := make([]*plugin.File, 0)
	log.GlobalLogger.Log(req.String())
	fileName, fileContent := log.GlobalLogger.Print()
	outputFiles = append(outputFiles, &plugin.File{
		Name:     fileName,
		Contents: fileContent,
	})
	return &plugin.GenerateResponse{Files: outputFiles}, nil
}
