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
the query functions and the QueryResults helper (both the call path and the
cursor-based for path) of the classes and functions packages.
"""

from __future__ import annotations

import typing

import pytest

from test.driver_psycopg_sync.no_row_conn import NoRowConn
from test.driver_psycopg_sync.omit_tc.classes import enums as classes_enums
from test.driver_psycopg_sync.omit_tc.classes import models as classes_models
from test.driver_psycopg_sync.omit_tc.classes import queries_enum_override as classes_queries
from test.driver_psycopg_sync.omit_tc.functions import enums as functions_enums
from test.driver_psycopg_sync.omit_tc.functions import models as functions_models
from test.driver_psycopg_sync.omit_tc.functions import queries_enum_override as functions_queries

if typing.TYPE_CHECKING:
    import psycopg
    import psycopg.rows

# Ids reserved for this file; all suites share one database sequentially, so
# every enum_override chain uses unique ids and deletes its rows at the end.
CLASSES_IDS: typing.Final[tuple[int, int]] = (510010, 520010)
FUNCTIONS_IDS: typing.Final[tuple[int, int]] = (510011, 520011)
MISSING_ID: typing.Final[int] = 987654321


class TestOmitTcClasses:
    @pytest.fixture(scope="session")
    def queries_obj(self, psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow]) -> classes_queries.QueriesEnumOverride:
        return classes_queries.QueriesEnumOverride(conn=psycopg_sync_conn)

    @pytest.mark.dependency(name="TestOmitTcClasses::insert_enum_override")
    def test_insert_enum_override(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        # The overridden parameter is a plain str; the generated code converts
        # it back to enums.TestMood before it reaches the driver.
        queries_obj.insert_enum_override(id_=CLASSES_IDS[0], mood_test="happy")
        queries_obj.insert_enum_override(id_=CLASSES_IDS[1], mood_test="sad")

    @pytest.mark.dependency(name="TestOmitTcClasses::get_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    def test_get_enum_override_mood(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        mood = queries_obj.get_enum_override_mood(id_=CLASSES_IDS[0])
        assert mood is not None
        assert isinstance(mood, str)
        assert mood == "happy"

    def test_get_enum_override_mood_not_found(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        assert queries_obj.get_enum_override_mood(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="TestOmitTcClasses::list_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    def test_list_enum_override_by_ids(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        # Calling the QueryResults object fetches all rows in one go.
        rows = queries_obj.list_enum_override_by_ids(dollar_1=list(CLASSES_IDS))()
        assert all(isinstance(row, classes_models.TestEnumOverride) for row in rows)
        assert {row.id_: row.mood_test for row in rows} == {CLASSES_IDS[0]: "happy", CLASSES_IDS[1]: "sad"}

    @pytest.mark.dependency(name="TestOmitTcClasses::iterate_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    def test_iterate_enum_override_by_ids(
        self,
        queries_obj: classes_queries.QueriesEnumOverride,
        psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow],
    ) -> None:
        assert queries_obj.conn is psycopg_sync_conn
        results = queries_obj.list_enum_override_by_ids(dollar_1=list(CLASSES_IDS))
        seen: dict[int, str] = {}
        # Exercise the cursor-based for path.
        with queries_obj.conn.transaction():
            for row in results:
                assert isinstance(row, classes_models.TestEnumOverride)
                seen[row.id_] = row.mood_test
        assert seen == {CLASSES_IDS[0]: "happy", CLASSES_IDS[1]: "sad"}

    @pytest.mark.dependency(name="TestOmitTcClasses::count_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    def test_count_enum_override_by_moods(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        count = queries_obj.count_enum_override_by_moods(dollar_1=[classes_enums.TestMood.HAPPY, classes_enums.TestMood.SAD])
        assert count == len(CLASSES_IDS)
        assert queries_obj.count_enum_override_by_moods(dollar_1=[classes_enums.TestMood.VALUE_24H]) == 0

    def test_count_enum_override_no_row(self) -> None:
        conn = typing.cast("psycopg.Connection[psycopg.rows.TupleRow]", NoRowConn())
        stub_queries_obj = classes_queries.QueriesEnumOverride(conn=conn)
        count = stub_queries_obj.count_enum_override_by_moods(dollar_1=[classes_enums.TestMood.HAPPY])
        assert count is None

    @pytest.mark.dependency(depends=["TestOmitTcClasses::insert_enum_override"])
    def test_delete_enum_override(self, psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow]) -> None:
        # count_enum_override_by_moods asserts exact counts; remove the rows so
        # later suites against the shared database start clean.
        for row_id in CLASSES_IDS:
            psycopg_sync_conn.execute("DELETE FROM test_enum_override WHERE id = %(id)s", {"id": row_id})


class TestOmitTcFunctions:
    @pytest.mark.dependency(name="TestOmitTcFunctions::insert_enum_override")
    def test_insert_enum_override(self, psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow]) -> None:
        # The overridden parameter is a plain str; the generated code converts
        # it back to enums.TestMood before it reaches the driver.
        functions_queries.insert_enum_override(conn=psycopg_sync_conn, id_=FUNCTIONS_IDS[0], mood_test="happy")
        functions_queries.insert_enum_override(conn=psycopg_sync_conn, id_=FUNCTIONS_IDS[1], mood_test="sad")

    @pytest.mark.dependency(name="TestOmitTcFunctions::get_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_get_enum_override_mood(self, psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow]) -> None:
        mood = functions_queries.get_enum_override_mood(conn=psycopg_sync_conn, id_=FUNCTIONS_IDS[0])
        assert mood is not None
        assert isinstance(mood, str)
        assert mood == "happy"

    def test_get_enum_override_mood_not_found(self, psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow]) -> None:
        assert functions_queries.get_enum_override_mood(conn=psycopg_sync_conn, id_=MISSING_ID) is None

    @pytest.mark.dependency(name="TestOmitTcFunctions::list_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_list_enum_override_by_ids(self, psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow]) -> None:
        # Calling the QueryResults object fetches all rows in one go.
        rows = functions_queries.list_enum_override_by_ids(conn=psycopg_sync_conn, dollar_1=list(FUNCTIONS_IDS))()
        assert all(isinstance(row, functions_models.TestEnumOverride) for row in rows)
        assert {row.id_: row.mood_test for row in rows} == {FUNCTIONS_IDS[0]: "happy", FUNCTIONS_IDS[1]: "sad"}

    @pytest.mark.dependency(name="TestOmitTcFunctions::iterate_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_iterate_enum_override_by_ids(self, psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow]) -> None:
        results = functions_queries.list_enum_override_by_ids(conn=psycopg_sync_conn, dollar_1=list(FUNCTIONS_IDS))
        seen: dict[int, str] = {}
        # Exercise the cursor-based for path.
        with psycopg_sync_conn.transaction():
            for row in results:
                assert isinstance(row, functions_models.TestEnumOverride)
                seen[row.id_] = row.mood_test
        assert seen == {FUNCTIONS_IDS[0]: "happy", FUNCTIONS_IDS[1]: "sad"}

    @pytest.mark.dependency(name="TestOmitTcFunctions::count_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_count_enum_override_by_moods(self, psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow]) -> None:
        count = functions_queries.count_enum_override_by_moods(conn=psycopg_sync_conn, dollar_1=[functions_enums.TestMood.HAPPY, functions_enums.TestMood.SAD])
        assert count == len(FUNCTIONS_IDS)
        assert functions_queries.count_enum_override_by_moods(conn=psycopg_sync_conn, dollar_1=[functions_enums.TestMood.VALUE_24H]) == 0

    def test_count_enum_override_no_row(self) -> None:
        conn = typing.cast("psycopg.Connection[psycopg.rows.TupleRow]", NoRowConn())
        count = functions_queries.count_enum_override_by_moods(conn=conn, dollar_1=[functions_enums.TestMood.HAPPY])
        assert count is None

    @pytest.mark.dependency(depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_delete_enum_override(self, psycopg_sync_conn: psycopg.Connection[psycopg.rows.TupleRow]) -> None:
        # count_enum_override_by_moods asserts exact counts; remove the rows so
        # later suites against the shared database start clean.
        for row_id in FUNCTIONS_IDS:
            psycopg_sync_conn.execute("DELETE FROM test_enum_override WHERE id = %(id)s", {"id": row_id})
