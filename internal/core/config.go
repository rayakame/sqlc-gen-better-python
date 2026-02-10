package core

import (
	"encoding/json"
	"fmt"

	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

const PluginVersion = "v0.4.5"

type Config struct {
	Package                     string              `json:"package" yaml:"package"`
	SqlDriver                   SQLDriver           `json:"sql_driver" yaml:"sql_driver"`
	ModelType                   ModelType           `json:"model_type" yaml:"model_type"`
	Initialisms                 *[]string           `json:"initialisms,omitempty" yaml:"initialisms,omitempty"`
	EmitExactTableNames         bool                `json:"emit_exact_table_names" yaml:"emit_exact_table_names"`
	EmitClasses                 bool                `json:"emit_classes" yaml:"emit_classes"`
	InflectionExcludeTableNames []string            `json:"inflection_exclude_table_names,omitempty" yaml:"inflection_exclude_table_names,omitempty"`
	OmitUnusedModels            bool                `json:"omit_unused_models" yaml:"omit_unused_models"`
	OmitTypecheckingBlock       bool                `json:"omit_typechecking_block" yaml:"omit_typechecking_block"`
	QueryParameterLimit         *int32              `json:"query_parameter_limit,omitempty" yaml:"query_parameter_limit"`
	OmitKwargsLimit             *int32              `json:"omit_kwargs_limit,omitempty" yaml:"omit_kwargs_limit"`
	EmitInitFile                *bool               `json:"emit_init_file" yaml:"emit_init_file"`
	EmitDocstrings              DocstringConvention `json:"docstrings" yaml:"docstrings"`
	OmitDocstringsSQL           bool                `json:"docstrings_emit_sql" yaml:"docstrings_emit_sql"`
	Speedups                    bool                `json:"speedups" yaml:"speedups"`
	Overrides                   []Override          `json:"overrides,omitempty" yaml:"overrides"`

	Debug bool `json:"debug" yaml:"debug"`

	IndentChar          string `json:"indent_char" yaml:"indent_char"`
	CharsPerIndentLevel int    `json:"chars_per_indent_level" yaml:"chars_per_indent_level"`

	InitialismsMap map[string]struct{} `json:"-" yaml:"-"`
	Async          bool
}

func NewConfig(req *plugin.GenerateRequest) (*Config, error) {
	config, err := parseConfig(req)
	if err != nil {
		return nil, err
	}
	err = validateConf(config, req.Settings.Engine)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func parseConfig(req *plugin.GenerateRequest) (*Config, error) {
	var config Config
	if len(req.PluginOptions) == 0 {
		return &config, nil
	}
	if err := json.Unmarshal(req.PluginOptions, &config); err != nil {
		return nil, fmt.Errorf("unmarshalling plugin options: %w", err)
	}
	config.Async = config.SqlDriver.Async()

	for i := range config.Overrides {
		if err := config.Overrides[i].parse(req); err != nil {
			return nil, err
		}
	}

	if config.ModelType == "" {
		config.ModelType = ModelTypeDataclass
	}
	if config.QueryParameterLimit == nil {
		config.QueryParameterLimit = utils.ToPtr(int32(1))
	}
	if config.OmitKwargsLimit == nil {
		config.OmitKwargsLimit = new(int32)
	}
	if config.Initialisms == nil {
		config.Initialisms = utils.ToPtr([]string{"id"})
	}
	if config.IndentChar == "" {
		config.IndentChar = " "
	}
	if config.CharsPerIndentLevel <= 0 {
		config.CharsPerIndentLevel = 4
	}
	if config.EmitDocstrings == "" {
		config.EmitDocstrings = DocstringConventionNone
	}

	config.InitialismsMap = map[string]struct{}{}
	for _, initial := range *config.Initialisms {
		config.InitialismsMap[initial] = struct{}{}
	}
	return &config, nil
}
func validateConf(conf *Config, engine string) error {
	if *conf.QueryParameterLimit < 0 {
		return fmt.Errorf("invalid options: query parameter limit must not be negative")
	}
	if *conf.OmitKwargsLimit < 0 {
		return fmt.Errorf("invalid options: omit kwarg limit must not be negative")
	}

	if conf.EmitInitFile == nil {
		return fmt.Errorf("invalid options: you need to specify emit_init_file")
	}

	if conf.Package == "" {
		return fmt.Errorf("invalid options: package must not be empty")
	}

	if err := conf.SqlDriver.Validate(engine); err != nil {
		return fmt.Errorf("invalid options: unknown model type: %e", err)
	}

	if !conf.ModelType.Valid() {
		return fmt.Errorf("invalid options: unknown model type: %s", conf.ModelType)
	}

	if !conf.EmitDocstrings.Valid() {
		return fmt.Errorf("invalid options: unknown docstring convention: %s", conf.EmitDocstrings)
	}

	return nil
}
