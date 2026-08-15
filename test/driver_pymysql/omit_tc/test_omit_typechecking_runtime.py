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
"""Runtime coverage for the pymysql omit_typechecking_block query modules.

The generated code must behave exactly like the regular variants even though
all imports and type aliases execute at module level. These tests exercise
the query functions and the QueryResults helper (both the call path and the
cursor-based for path) of the classes and functions packages.
"""

from __future__ import annotations

import typing

import pytest

from test.driver_pymysql import no_row_conn
from test.driver_pymysql.omit_tc.classes import models as classes_models
from test.driver_pymysql.omit_tc.classes import queries_enum_override as classes_queries
from test.driver_pymysql.omit_tc.functions import models as functions_models
from test.driver_pymysql.omit_tc.functions import queries_enum_override as functions_queries

if typing.TYPE_CHECKING:
    import pymysql

# Ids reserved for this file (omit_tc owns 5000-5099); all suites share one
# database sequentially, so every enum_override chain uses unique ids and
# deletes its rows at the end.
CLASSES_IDS: typing.Final[tuple[int, int]] = (5000, 5001)
FUNCTIONS_IDS: typing.Final[tuple[int, int]] = (5010, 5011)
MISSING_ID: typing.Final[int] = 5099


class TestOmitTcClasses:
    @pytest.fixture(scope="session")
    def queries_obj(self, pymysql_conn: pymysql.Connection) -> classes_queries.QueriesEnumOverride:
        return classes_queries.QueriesEnumOverride(conn=pymysql_conn)

    @pytest.mark.dependency(name="TestOmitTcClasses::insert_enum_override")
    def test_insert_enum_override(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        # The overridden parameter is a plain str; the generated code converts
        # it back to enums.TestEnumOverrideMoodTest before it reaches the
        # driver.
        queries_obj.insert_enum_override(id_=CLASSES_IDS[0], mood_test="happy")
        queries_obj.insert_enum_override(id_=CLASSES_IDS[1], mood_test="sad")
        with pytest.raises(ValueError, match="angry"):
            queries_obj.insert_enum_override(id_=MISSING_ID, mood_test="angry")

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
        rows = queries_obj.list_enum_override_by_ids(ids=list(CLASSES_IDS))()
        assert all(isinstance(row, classes_models.TestEnumOverride) for row in rows)
        assert {row.id_: row.mood_test for row in rows} == {CLASSES_IDS[0]: "happy", CLASSES_IDS[1]: "sad"}

    @pytest.mark.dependency(name="TestOmitTcClasses::iterate_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    def test_iterate_enum_override_by_ids(
        self,
        queries_obj: classes_queries.QueriesEnumOverride,
        pymysql_conn: pymysql.Connection,
    ) -> None:
        assert queries_obj.conn is pymysql_conn
        results = queries_obj.list_enum_override_by_ids(ids=list(CLASSES_IDS))
        seen: dict[int, str] = {}
        # Exercise the cursor-based for path.
        for row in results:
            assert isinstance(row, classes_models.TestEnumOverride)
            seen[row.id_] = row.mood_test
        assert seen == {CLASSES_IDS[0]: "happy", CLASSES_IDS[1]: "sad"}

    @pytest.mark.dependency(name="TestOmitTcClasses::empty_enum_override", depends=["TestOmitTcClasses::insert_enum_override"])
    def test_list_enum_override_by_ids_empty(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        # An empty slice expands to IN (NULL), which matches no rows.
        assert list(queries_obj.list_enum_override_by_ids(ids=[])()) == []
        assert list(queries_obj.list_enum_override_by_ids(ids=[])) == []

    @pytest.mark.dependency(depends=["TestOmitTcClasses::insert_enum_override"])
    def test_count_enum_override_by_moods(self, queries_obj: classes_queries.QueriesEnumOverride) -> None:
        # An empty slice expands to IN (NULL); count(*) still returns a row.
        assert queries_obj.count_enum_override_by_moods(moods=[]) == 0
        stub = typing.cast("pymysql.Connection", no_row_conn.NoRowConn())
        assert classes_queries.QueriesEnumOverride(conn=stub).count_enum_override_by_moods(moods=[]) is None

    @pytest.mark.dependency(depends=["TestOmitTcClasses::insert_enum_override"])
    def test_delete_enum_override(self, pymysql_conn: pymysql.Connection) -> None:
        # Remove the rows so later suites against the shared database start
        # clean.
        with pymysql_conn.cursor() as cur:
            for row_id in CLASSES_IDS:
                cur.execute("DELETE FROM test_enum_override WHERE id = %s", (row_id,))


class TestOmitTcFunctions:
    @pytest.mark.dependency(name="TestOmitTcFunctions::insert_enum_override")
    def test_insert_enum_override(self, pymysql_conn: pymysql.Connection) -> None:
        # The overridden parameter is a plain str; the generated code converts
        # it back to enums.TestEnumOverrideMoodTest before it reaches the
        # driver.
        functions_queries.insert_enum_override(conn=pymysql_conn, id_=FUNCTIONS_IDS[0], mood_test="happy")
        functions_queries.insert_enum_override(conn=pymysql_conn, id_=FUNCTIONS_IDS[1], mood_test="sad")
        with pytest.raises(ValueError, match="angry"):
            functions_queries.insert_enum_override(conn=pymysql_conn, id_=MISSING_ID, mood_test="angry")

    @pytest.mark.dependency(name="TestOmitTcFunctions::get_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_get_enum_override_mood(self, pymysql_conn: pymysql.Connection) -> None:
        mood = functions_queries.get_enum_override_mood(conn=pymysql_conn, id_=FUNCTIONS_IDS[0])
        assert mood is not None
        assert isinstance(mood, str)
        assert mood == "happy"

    def test_get_enum_override_mood_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert functions_queries.get_enum_override_mood(conn=pymysql_conn, id_=MISSING_ID) is None

    def test_count_enum_override_by_moods(self, pymysql_conn: pymysql.Connection) -> None:
        # An empty slice expands to IN (NULL); count(*) still returns a row.
        assert functions_queries.count_enum_override_by_moods(conn=pymysql_conn, moods=[]) == 0
        stub = typing.cast("pymysql.Connection", no_row_conn.NoRowConn())
        assert functions_queries.count_enum_override_by_moods(conn=stub, moods=[]) is None

    @pytest.mark.dependency(name="TestOmitTcFunctions::list_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_list_enum_override_by_ids(self, pymysql_conn: pymysql.Connection) -> None:
        # Calling the QueryResults object fetches all rows in one go.
        rows = functions_queries.list_enum_override_by_ids(conn=pymysql_conn, ids=list(FUNCTIONS_IDS))()
        assert all(isinstance(row, functions_models.TestEnumOverride) for row in rows)
        assert {row.id_: row.mood_test for row in rows} == {FUNCTIONS_IDS[0]: "happy", FUNCTIONS_IDS[1]: "sad"}

    @pytest.mark.dependency(name="TestOmitTcFunctions::iterate_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_iterate_enum_override_by_ids(self, pymysql_conn: pymysql.Connection) -> None:
        results = functions_queries.list_enum_override_by_ids(conn=pymysql_conn, ids=list(FUNCTIONS_IDS))
        seen: dict[int, str] = {}
        # Exercise the cursor-based for path.
        for row in results:
            assert isinstance(row, functions_models.TestEnumOverride)
            seen[row.id_] = row.mood_test
        assert seen == {FUNCTIONS_IDS[0]: "happy", FUNCTIONS_IDS[1]: "sad"}

    @pytest.mark.dependency(name="TestOmitTcFunctions::empty_enum_override", depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_list_enum_override_by_ids_empty(self, pymysql_conn: pymysql.Connection) -> None:
        # An empty slice expands to IN (NULL), which matches no rows.
        assert list(functions_queries.list_enum_override_by_ids(conn=pymysql_conn, ids=[])()) == []
        assert list(functions_queries.list_enum_override_by_ids(conn=pymysql_conn, ids=[])) == []

    @pytest.mark.dependency(depends=["TestOmitTcFunctions::insert_enum_override"])
    def test_delete_enum_override(self, pymysql_conn: pymysql.Connection) -> None:
        # Remove the rows so later suites against the shared database start
        # clean.
        with pymysql_conn.cursor() as cur:
            for row_id in FUNCTIONS_IDS:
                cur.execute("DELETE FROM test_enum_override WHERE id = %s", (row_id,))
