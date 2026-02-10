package types

import (
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type TypeConversionFunc func(req *plugin.GenerateRequest, typ *plugin.Identifier, conf *core.Config) string
