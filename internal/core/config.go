package core

import (
	"encoding/json"
	"fmt"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Config struct {
	SqlDriver string `json:"sql_driver" yaml:"sql_driver"`
	ModelType string `json:"model_type" yaml:"model_type"`

	Async bool
}

func ParseConfig(req *plugin.GenerateRequest) (*Config, error) {
	var config Config
	if len(req.PluginOptions) == 0 {
		return &config, nil
	}
	if err := json.Unmarshal(req.PluginOptions, &config); err != nil {
		return nil, fmt.Errorf("unmarshalling plugin options: %w", err)
	}
	if config.SqlDriver == "" {
		config.SqlDriver = SQLDriverAioSQLite
	}
	val, err := isDriverAsync(config.SqlDriver)
	if err != nil {
		return nil, fmt.Errorf("invalid options: %s", err)
	}
	config.Async = val
	if config.ModelType == "" {
		config.ModelType = ModelTypeDataclass
	}
	if err := isModelTypeValid(config.ModelType); err != nil {
		return nil, fmt.Errorf("invalid options: %s", err)
	}
	return &config, nil
}
