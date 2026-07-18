package utils

import (
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func ToPtr[T any](t T) *T {
	return &t
}

func SameTableName(table1, table2 *plugin.Identifier, defaultSchema string) bool {
	if table1 == nil || table2 == nil {
		return false
	}
	schema1, schema2 := table1.Schema, table2.Schema
	if schema1 == "" {
		schema1 = defaultSchema
	}
	if schema2 == "" {
		schema2 = defaultSchema
	}
	return table1.Catalog == table2.Catalog && schema1 == schema2 && table1.Name == table2.Name
}
