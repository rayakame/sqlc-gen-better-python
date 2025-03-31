package codegen

import (
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen/builders"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func (dr *Driver) BuildInitFile() *plugin.File {
	body := builders.NewIndentStringBuilder(dr.conf.IndentChar, dr.conf.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	return &plugin.File{
		Name:     "__init__.py",
		Contents: []byte(body.String()),
	}
}
