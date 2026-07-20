package render_test

import (
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestRenderEnums(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		options string
		enums   []*plugin.Enum
		want    string
	}{
		{
			name:    "constant naming and multiple enums",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false}`,
			enums: []*plugin.Enum{
				{Name: "test_mood", Vals: []string{"sad", "24h", "_hidden", ""}},
				{Name: "status", Vals: []string{"active"}},
			},
			want: sqlcFileHeader("") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = (
    "Status",
    "TestMood",
)

import enum
import typing

if typing.TYPE_CHECKING:
    import collections.abc


class Status(enum.StrEnum):
    ACTIVE = "active"


class TestMood(enum.StrEnum):
    SAD = "sad"
    VALUE_24H = "24h"
    VALUE__HIDDEN = "_hidden"
    VALUE_4 = ""
`,
		},
		{
			name:    "google docstrings",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false,"docstrings":"google"}`,
			enums: []*plugin.Enum{
				{Name: "status", Vals: []string{"active", "deleted"}},
			},
			want: sqlcFileHeader("") + `"""Module containing enums."""

from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("Status",)

import enum
import typing

if typing.TYPE_CHECKING:
    import collections.abc


class Status(enum.StrEnum):
    """Enum representing Status."""

    ACTIVE = "active"
    DELETED = "deleted"
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			catalog := &plugin.Catalog{
				DefaultSchema: "public",
				Schemas:       []*plugin.Schema{{Name: "public", Enums: tc.enums}},
			}
			req := newRenderRequest("postgresql", tc.options, catalog, nil)

			got := renderedFile(t, mustRenderFiles(t, req), "enums.py")
			if got != tc.want {
				t.Errorf("enums.py mismatch\ngot:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}
