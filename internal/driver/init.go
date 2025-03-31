package driver

import (
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func (dr *Driver) BuildInitFile() *plugin.File {
	body := codegen.NewIndentStringBuilder(dr.conf.IndentChar, dr.conf.CharsPerIndentLevel)
	body.WriteSqlcHeader()
	return &plugin.File{
		Name:     "__init__.py",
		Contents: []byte(body.String()),
	}
}
