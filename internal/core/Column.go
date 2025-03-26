package core

import "github.com/sqlc-dev/plugin-sdk-go/plugin"

type Column struct {
	Name    string // CamelCased name for Go
	DBName  string // Name as used in the DB
	Type    PyType
	Comment string
	Column  *plugin.Column
}
