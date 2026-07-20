package render_test

import (
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestRenderTables(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		options string
		catalog *plugin.Catalog
		want    string
	}{
		{
			name:    "dataclass",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false,"model_type":"dataclass"}`,
			catalog: pgItemsCatalog(),
			want: sqlcFileHeader("") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("TestItem",)

import dataclasses
import typing

if typing.TYPE_CHECKING:
    import collections.abc


@dataclasses.dataclass()
class TestItem:
    id_: int
    name: str | None
`,
		},
		{
			name:    "attrs",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false,"model_type":"attrs"}`,
			catalog: pgItemsCatalog(),
			want: sqlcFileHeader("") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("TestItem",)

import attrs
import typing

if typing.TYPE_CHECKING:
    import collections.abc


@attrs.define()
class TestItem:
    id_: int
    name: str | None
`,
		},
		{
			name:    "msgspec",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false,"model_type":"msgspec"}`,
			catalog: pgItemsCatalog(),
			want: sqlcFileHeader("") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("TestItem",)

import msgspec
import typing

if typing.TYPE_CHECKING:
    import collections.abc


class TestItem(msgspec.Struct):
    id_: int
    name: str | None
`,
		},
		{
			name:    "pydantic emits model_config",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false,"model_type":"pydantic"}`,
			catalog: pgItemsCatalog(),
			want: sqlcFileHeader("") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("TestItem",)

import pydantic
import typing

if typing.TYPE_CHECKING:
    import collections.abc


class TestItem(pydantic.BaseModel):
    model_config = pydantic.ConfigDict(arbitrary_types_allowed=True)

    id_: int
    name: str | None
`,
		},
		{
			name:    "dataclass with google docstrings",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false,"docstrings":"google"}`,
			catalog: pgItemsCatalog(),
			want: sqlcFileHeader("") + `"""Module containing models."""

from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("TestItem",)

import dataclasses
import typing

if typing.TYPE_CHECKING:
    import collections.abc


@dataclasses.dataclass()
class TestItem:
    """Model representing TestItem.

    Attributes:
        id_: int
        name: str | None
    """

    id_: int
    name: str | None
`,
		},
		{
			name:    "no tables still emits models file",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false}`,
			catalog: nil,
			want: sqlcFileHeader("") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = ()

import typing

if typing.TYPE_CHECKING:
    import collections.abc
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := newRenderRequest("postgresql", tc.options, tc.catalog, nil)

			got := renderedFile(t, mustRenderFiles(t, req), "models.py")
			if got != tc.want {
				t.Errorf("models.py mismatch\ngot:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}
