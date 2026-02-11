package transform

import (
	"strings"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
)

func (t *Transformer) BuildQueries() []model.Query {
	queries := make([]model.Query, 0, len(t.req.Queries))
	for _, pluginQuery := range t.req.Queries {
		if pluginQuery.Name == "" {
			continue
		}
		if pluginQuery.Cmd == "" {
			continue
		}

		constantName := model.UpperSnakeCase(pluginQuery.Name)

		query := model.Query{
			Cmd:          pluginQuery.Cmd,
			SQL:          pluginQuery.Text,
			ConstantName: constantName,
			FuncName:     strings.ToLower(constantName),
		}

		queries = append(queries, query)
	}

	return queries
}
