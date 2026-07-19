package transform

import (
	"cmp"
	"slices"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
)

func (t *Transformer) BuildEnums() []model.Enum {
	enums := make([]model.Enum, 0)
	seenNames := make(map[string]int)
	for _, schema := range t.req.Catalog.Schemas {
		if schema.Name == utils.PgCatalog || schema.Name == utils.InformationSchema {
			continue
		}
		for _, enum := range schema.Enums {
			var schemaName string
			if schema.Name != t.req.Catalog.DefaultSchema {
				schemaName = schema.Name
			}

			e := model.Enum{
				Name:      model.DedupClassName(model.EnumName(t.config, enum.Name, schemaName), seenNames),
				Constants: make([]model.EnumConstants, 0, len(enum.Vals)),
			}

			seen := make(map[string]int, len(enum.Vals))
			for i, v := range enum.Vals {
				e.Constants = append(e.Constants, model.EnumConstants{
					Name:  model.EnumConstantName(v, i, seen),
					Value: v,
				})
			}
			enums = append(enums, e)
		}
	}
	slices.SortFunc(enums, func(a, b model.Enum) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return enums
}
