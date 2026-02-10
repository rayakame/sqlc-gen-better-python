package internal

import (
	"context"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func Handler(_ context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	pythonGenerator, err := NewPythonGenerator(req)
	if err != nil {
		return nil, err
	}
	return pythonGenerator.Run()
}
