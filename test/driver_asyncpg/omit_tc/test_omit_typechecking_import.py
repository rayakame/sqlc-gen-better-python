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
"""With omit_typechecking_block the driver hook executes at module level.

Importing the generated modules is the whole test: stub-only generics like
``asyncpg.Connection[...]`` must never be evaluated eagerly, which the lazy
PEP 695 alias form guarantees.
"""

from __future__ import annotations

from test.driver_asyncpg.omit_tc.classes import queries_enum_override as classes_module
from test.driver_asyncpg.omit_tc.functions import queries_enum_override as functions_module


def test_omit_typechecking_modules_import_at_runtime() -> None:
    assert classes_module.INSERT_ENUM_OVERRIDE
    assert functions_module.INSERT_ENUM_OVERRIDE
