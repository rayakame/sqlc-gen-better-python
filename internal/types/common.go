package types

import "github.com/sqlc-dev/plugin-sdk-go/plugin"

type TypeConversionFunc func(req *plugin.GenerateRequest, col *plugin.Column) string
