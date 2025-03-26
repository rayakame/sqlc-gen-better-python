package core

import (
	"encoding/json"
	"fmt"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Config struct {
	SqlDriver                   string    `json:"sql_driver" yaml:"sql_driver"`
	ModelType                   string    `json:"model_type" yaml:"model_type"`
	Initialisms                 *[]string `json:"initialisms,omitempty" yaml:"initialisms"`
	EmitExactTableNames         bool      `json:"emit_exact_table_names,omitempty" yaml:"emit_exact_table_names"`
	InflectionExcludeTableNames []string  `json:"inflection_exclude_table_names,omitempty" yaml:"inflection_exclude_table_names"`
	OmitUnusedStructs           bool      `json:"omit_unused_structs,omitempty" yaml:"omit_unused_structs"`
	QueryParameterLimit         *int32    `json:"query_parameter_limit,omitempty" yaml:"query_parameter_limit"`

	InitialismsMap map[string]struct{} `json:"-" yaml:"-"`
	Async          bool
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
	if config.Initialisms == nil {
		config.Initialisms = new([]string)
		*config.Initialisms = []string{"id"}
	}

	config.InitialismsMap = map[string]struct{}{}
	for _, initial := range *config.Initialisms {
		config.InitialismsMap[initial] = struct{}{}
	}
	return &config, nil
}
