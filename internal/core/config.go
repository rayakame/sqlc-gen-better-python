package core

import (
	"encoding/json"
	"fmt"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

const PluginVersion = "v0.1.0"

type Config struct {
	Package                     string        `json:"package" yaml:"package"`
	SqlDriver                   SQLDriverType `json:"sql_driver" yaml:"sql_driver"`
	ModelType                   string        `json:"model_type" yaml:"model_type"`
	Initialisms                 *[]string     `json:"initialisms,omitempty" yaml:"initialisms"`
	EmitExactTableNames         bool          `json:"emit_exact_table_names,omitempty" yaml:"emit_exact_table_names"`
	EmitClasses                 bool          `json:"emit_classes" yaml:"emit_classes"`
	InflectionExcludeTableNames []string      `json:"inflection_exclude_table_names,omitempty" yaml:"inflection_exclude_table_names"`
	OmitUnusedStructs           bool          `json:"omit_unused_structs,omitempty" yaml:"omit_unused_structs"`
	QueryParameterLimit         *int32        `json:"query_parameter_limit,omitempty" yaml:"query_parameter_limit"`

	IndentChar          string `json:"indent_char" yaml:"indent_char"`
	CharsPerIndentLevel int    `json:"chars_per_indent_level" yaml:"chars_per_indent_level"`

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
	if config.QueryParameterLimit == nil {
		config.QueryParameterLimit = new(int32)
		*config.QueryParameterLimit = 1
	}
	if config.Initialisms == nil {
		config.Initialisms = new([]string)
		*config.Initialisms = []string{"id"}
	}
	if config.IndentChar == "" {
		config.IndentChar = " "
	}
	if config.CharsPerIndentLevel == 0 {
		config.CharsPerIndentLevel = 4
	}

	config.InitialismsMap = map[string]struct{}{}
	for _, initial := range *config.Initialisms {
		config.InitialismsMap[initial] = struct{}{}
	}
	return &config, nil
}
func ValidateConf(conf *Config, engine string) error {
	if *conf.QueryParameterLimit < 0 {
		return fmt.Errorf("invalid options: query parameter limit must not be negative")
	}

	if conf.Package == "" {
		return fmt.Errorf("invalid options: package must not be empty")
	}

	if err := isDriverValid(conf.SqlDriver, engine); err != nil {
		return err
	}

	if err := isModelTypeValid(conf.ModelType); err != nil {

		return fmt.Errorf("invalid options: %s", err)
	}

	return nil
}
