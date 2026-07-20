# Copyright (c) 2025-present Rayakame

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
"""User-defined converters exercised by the generated test packages."""

from __future__ import annotations

import dataclasses
import json
import pathlib
import typing


@dataclasses.dataclass()
class Preferences:
    """Value stored as jsonb and converted through the functions below."""

    theme: str
    notifications: bool


def encode_preferences(value: Preferences) -> str:
    """Serialize preferences into the column's natural type.

    Returns
    -------
    str
        The JSON encoded preferences.
    """
    return json.dumps({"theme": value.theme, "notifications": value.notifications})


def decode_preferences(value: str) -> Preferences:
    """Deserialize preferences, never receiving None.

    Returns
    -------
    Preferences
        The decoded preferences.
    """
    raw: dict[str, typing.Any] = json.loads(value)
    return Preferences(theme=raw["theme"], notifications=raw["notifications"])


def encode_tags(value: frozenset[str]) -> str:
    """Serialize tags into a JSON array, preserving tags with commas.

    Returns
    -------
    str
        The JSON encoded tags.
    """
    return json.dumps(sorted(value))


def decode_tags(value: str) -> frozenset[str]:
    """Deserialize a JSON array of tags.

    Returns
    -------
    frozenset[str]
        The parsed tags.
    """
    raw: list[str] = json.loads(value)
    return frozenset(raw)


def encode_label(value: pathlib.PurePosixPath) -> str:
    """Serialize a label for a domain-typed (text) column.

    Returns
    -------
    str
        The label as text.
    """
    return str(value)


def decode_label(value: str) -> pathlib.PurePosixPath:
    """Deserialize a label from a domain-typed (text) column.

    Returns
    -------
    pathlib.PurePosixPath
        The parsed label.
    """
    return pathlib.PurePosixPath(value)
