package transform

import (
	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Transformer struct {
	typeConversionFunc types.TypeConversionFunc
	config             *config.Config
	req                *plugin.GenerateRequest
}

func NewTransformer(conf *config.Config, req *plugin.GenerateRequest, convFunc types.TypeConversionFunc) *Transformer {
	return &Transformer{typeConversionFunc: convFunc, config: conf, req: req}
}
