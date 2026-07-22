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
"""Runtime coverage for the psycopg omit_typechecking_block query modules.

The generated code must behave exactly like the regular variants even though
all imports and type aliases execute at module level. These tests exercise
the query functions and the QueryResults helper (both the await path and the
cursor-based async-for path) of the classes and functions packages.
"""

from __future__ import annotations

import asyncio
import typing

import pytest
import pytest_asyncio

from test.driver_psycopg_async.omit_tc.classes import enums as classes_enums
from test.driver_psycopg_async.omit_tc.classes import models as classes_models
from test.driver_psycopg_async.omit_tc.classes import queries_enum_override as classes_queries
from test.driver_psycopg_async.omit_tc.functions import enums as functions_enums
from test.driver_psycopg_async.omit_tc.functions import models as functions_models
from test.driver_psycopg_async.omit_tc.functions import queries_enum_override as functions_queries

if typing.TYPE_CHECKING:
    import psycopg
    import psycopg.rows

# Ids reserved for this file; all suites share one database sequentially, so
# every enum_override chain uses unique ids and deletes its rows at the end.
CLASSES_IDS: typing.Final[tuple[int, int]] = (510010, 520010)
FUNCTIONS_IDS: typing.Final[tuple[int, int]] = (510011, 520011)
MISSING_ID: typing.Final[int] = 987654321


class _NoRowCursor:
    async def fetchone(self) -> None:
        await asyncio.sleep(0)


class _NoRowConn:
    # `SELECT count(*)` always returns exactly one row, so the generated
    # not-found branch of the count queries needs a connection stub that
    # misses.
    async def execute(self, _query: str, _params: object = None) -> _NoRowCursor:
        await asyncio.sleep(0)
        return _NoRowCursor()


@pytest.mark.asyncio(loop_scope="session")
class TestOmitTcClasses:
    @pytest_asyncio.fixture(scope="session", loop_scope="session")
    async def queries_obj(self, psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow]) -> classes_queries.QueriesEnumOverride:
        return classes_queries.QueriesEnumOverride(conn=psycopg_async_conn)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcClasses::insert_enum_override")
    async def test_insert_enum_override(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        # The overridden parameter is a plain str; the generated code converts
        # it back to enums.TestMood before it reaches the driver.
        await queries_obj.insert_enum_override(id_=CLASSES_IDS[0], mood_test="happy")
        await queries_obj.insert_enum_override(id_=CLASSES_IDS[1], mood_test="sad")

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcClasses::get_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    async def test_get_enum_override_mood(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        mood = await queries_obj.get_enum_override_mood(id_=CLASSES_IDS[0])
        assert mood is not None
        assert isinstance(mood, str)
        assert mood == "happy"

    @pytest.mark.asyncio(loop_scope="session")
    async def test_get_enum_override_mood_not_found(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        assert await queries_obj.get_enum_override_mood(id_=MISSING_ID) is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcClasses::list_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    async def test_list_enum_override_by_ids(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        # Awaiting the QueryResults object fetches all rows in one go.
        rows = await queries_obj.list_enum_override_by_ids(dollar_1=list(CLASSES_IDS))
        assert all(isinstance(row, classes_models.TestEnumOverride) for row in rows)
        assert {row.id_: row.mood_test for row in rows} == {CLASSES_IDS[0]: "happy", CLASSES_IDS[1]: "sad"}

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcClasses::iterate_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    async def test_iterate_enum_override_by_ids(
        self,
        queries_obj: classes_queries.QueriesEnumOverride,
        psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow],
    ) -> None:
        assert queries_obj.conn is psycopg_async_conn
        results = queries_obj.list_enum_override_by_ids(dollar_1=list(CLASSES_IDS))
        seen: dict[int, str] = {}
        # Exercise the cursor-based async-for path.
        async with queries_obj.conn.transaction():
            async for row in results:
                assert isinstance(row, classes_models.TestEnumOverride)
                seen[row.id_] = row.mood_test
        assert seen == {CLASSES_IDS[0]: "happy", CLASSES_IDS[1]: "sad"}

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcClasses::count_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    async def test_count_enum_override_by_moods(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        count = await queries_obj.count_enum_override_by_moods(dollar_1=[classes_enums.TestMood.HAPPY, classes_enums.TestMood.SAD])
        assert count == len(CLASSES_IDS)
        assert await queries_obj.count_enum_override_by_moods(dollar_1=[classes_enums.TestMood.VALUE_24H]) == 0

    @pytest.mark.asyncio(loop_scope="session")
    async def test_count_enum_override_no_row(self) -> None:
        conn = typing.cast("psycopg.AsyncConnection[psycopg.rows.TupleRow]", _NoRowConn())
        stub_queries_obj = classes_queries.QueriesEnumOverride(conn=conn)
        count = await stub_queries_obj.count_enum_override_by_moods(dollar_1=[classes_enums.TestMood.HAPPY])
        assert count is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestOmitTcClasses::insert_enum_override"])
    async def test_delete_enum_override(self, psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow]) -> None:
        # count_enum_override_by_moods asserts exact counts; remove the rows so
        # later suites against the shared database start clean.
        for row_id in CLASSES_IDS:
            await psycopg_async_conn.execute("DELETE FROM test_enum_override WHERE id = %(id)s", {"id": row_id})


@pytest.mark.asyncio(loop_scope="session")
class TestOmitTcFunctions:
    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcFunctions::insert_enum_override")
    async def test_insert_enum_override(self, psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow]) -> None:
        # The overridden parameter is a plain str; the generated code converts
        # it back to enums.TestMood before it reaches the driver.
        await functions_queries.insert_enum_override(conn=psycopg_async_conn, id_=FUNCTIONS_IDS[0], mood_test="happy")
        await functions_queries.insert_enum_override(conn=psycopg_async_conn, id_=FUNCTIONS_IDS[1], mood_test="sad")

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcFunctions::get_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    async def test_get_enum_override_mood(self, psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow]) -> None:
        mood = await functions_queries.get_enum_override_mood(conn=psycopg_async_conn, id_=FUNCTIONS_IDS[0])
        assert mood is not None
        assert isinstance(mood, str)
        assert mood == "happy"

    @pytest.mark.asyncio(loop_scope="session")
    async def test_get_enum_override_mood_not_found(self, psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow]) -> None:
        assert await functions_queries.get_enum_override_mood(conn=psycopg_async_conn, id_=MISSING_ID) is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcFunctions::list_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    async def test_list_enum_override_by_ids(self, psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow]) -> None:
        # Awaiting the QueryResults object fetches all rows in one go.
        rows = await functions_queries.list_enum_override_by_ids(conn=psycopg_async_conn, dollar_1=list(FUNCTIONS_IDS))
        assert all(isinstance(row, functions_models.TestEnumOverride) for row in rows)
        assert {row.id_: row.mood_test for row in rows} == {FUNCTIONS_IDS[0]: "happy", FUNCTIONS_IDS[1]: "sad"}

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcFunctions::iterate_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    async def test_iterate_enum_override_by_ids(self, psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow]) -> None:
        results = functions_queries.list_enum_override_by_ids(conn=psycopg_async_conn, dollar_1=list(FUNCTIONS_IDS))
        seen: dict[int, str] = {}
        # Exercise the cursor-based async-for path.
        async with psycopg_async_conn.transaction():
            async for row in results:
                assert isinstance(row, functions_models.TestEnumOverride)
                seen[row.id_] = row.mood_test
        assert seen == {FUNCTIONS_IDS[0]: "happy", FUNCTIONS_IDS[1]: "sad"}

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestOmitTcFunctions::count_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    async def test_count_enum_override_by_moods(self, psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow]) -> None:
        count = await functions_queries.count_enum_override_by_moods(conn=psycopg_async_conn, dollar_1=[functions_enums.TestMood.HAPPY, functions_enums.TestMood.SAD])
        assert count == len(FUNCTIONS_IDS)
        assert await functions_queries.count_enum_override_by_moods(conn=psycopg_async_conn, dollar_1=[functions_enums.TestMood.VALUE_24H]) == 0

    @pytest.mark.asyncio(loop_scope="session")
    async def test_count_enum_override_no_row(self) -> None:
        conn = typing.cast("psycopg.AsyncConnection[psycopg.rows.TupleRow]", _NoRowConn())
        count = await functions_queries.count_enum_override_by_moods(conn=conn, dollar_1=[functions_enums.TestMood.HAPPY])
        assert count is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestOmitTcFunctions::insert_enum_override"])
    async def test_delete_enum_override(self, psycopg_async_conn: psycopg.AsyncConnection[psycopg.rows.TupleRow]) -> None:
        # count_enum_override_by_moods asserts exact counts; remove the rows so
        # later suites against the shared database start clean.
        for row_id in FUNCTIONS_IDS:
            await psycopg_async_conn.execute("DELETE FROM test_enum_override WHERE id = %(id)s", {"id": row_id})
