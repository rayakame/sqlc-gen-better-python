package model_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/model"
)

func TestSingular(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		params model.SingularParams
		want   string
	}{
		{
			name:   "exclusion exact match keeps name",
			params: model.SingularParams{Name: "series", Exclusions: []string{"series"}},
			want:   "series",
		},
		{
			name:   "exclusion match is case-insensitive",
			params: model.SingularParams{Name: "Series", Exclusions: []string{"series"}},
			want:   "Series",
		},
		{
			name:   "exclusion compares the full string only",
			params: model.SingularParams{Name: "public.users", Exclusions: []string{"users"}},
			want:   "public.user",
		},
		{
			name:   "schema-qualified exclusion matches schema-qualified name",
			params: model.SingularParams{Name: "public.users", Exclusions: []string{"public.users"}},
			want:   "public.users",
		},
		{
			name:   "non-matching exclusions fall through",
			params: model.SingularParams{Name: "test_items", Exclusions: []string{"users", "orders"}},
			want:   "test_item",
		},
		{
			name:   "campus fix keeps original casing",
			params: model.SingularParams{Name: "Campus"},
			want:   "Campus",
		},
		{
			name:   "meta fix keeps original casing",
			params: model.SingularParams{Name: "Meta"},
			want:   "Meta",
		},
		{
			name:   "calories fix returns lowercase calorie",
			params: model.SingularParams{Name: "Calories"},
			want:   "calorie",
		},
		{
			name:   "waves fix returns lowercase wave",
			params: model.SingularParams{Name: "Waves"},
			want:   "wave",
		},
		{
			name:   "metadata fix returns lowercase metadata",
			params: model.SingularParams{Name: "Metadata"},
			want:   "metadata",
		},
		{
			name:   "regular plural via inflection",
			params: model.SingularParams{Name: "test_items"},
			want:   "test_item",
		},
		{
			name:   "irregular plural via inflection",
			params: model.SingularParams{Name: "people"},
			want:   "person",
		},
		{
			name:   "already singular is unchanged",
			params: model.SingularParams{Name: "user"},
			want:   "user",
		},
		{
			name:   "empty string is unchanged",
			params: model.SingularParams{Name: ""},
			want:   "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := model.Singular(tc.params); got != tc.want {
				t.Errorf("Singular(%+v) = %q, want %q", tc.params, got, tc.want)
			}
		})
	}
}
