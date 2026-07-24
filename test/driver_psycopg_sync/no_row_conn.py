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
"""Shared connection stub for exercising the generated not-found branches."""

from __future__ import annotations


class NoRowCursor:
    """Cursor stub whose fetchone never finds a row."""

    @staticmethod
    def fetchone() -> None:
        """Return None, exactly like a cursor over an empty result set."""


class NoRowConn:
    """Connection stub whose queries never find a row."""

    # `SELECT count(*)` always returns exactly one row, so the generated
    # not-found branch of the count queries needs a connection stub that
    # misses.
    @staticmethod
    def execute(_query: str, _params: object = None) -> NoRowCursor:
        """Return a cursor that finds no row.

        Returns
        -------
        NoRowCursor
            The cursor stub missing every row.
        """
        return NoRowCursor()
