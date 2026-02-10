package codegen

import (
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func (dr *Driver) BuildInitFile() *plugin.File {
	body := dr.GetStringBuilder()
	body.WriteSqlcHeader()
	body.WriteInitFileModuleDocstring()
	return &plugin.File{
		Name:     "__init__.py",
		Contents: []byte(body.String()),
	}
}
