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
"""Runtime tests for the DATETIME db_type override of the pymysql driver."""

from __future__ import annotations

import datetime
import typing

import pytest

from test.driver_pymysql.dbtype.functions import models
from test.driver_pymysql.dbtype.functions import queries_dbtype_override

if typing.TYPE_CHECKING:
    import pymysql

DBTYPE_ID: typing.Final = 5100
STAMP: typing.Final = "2026-01-02T03:04:05"


class TestPymysqlDbtypeOverride:
    @pytest.mark.dependency(name="PymysqlDbtypeOverride::insert")
    def test_insert_dbtype_override(self, pymysql_conn: pymysql.Connection) -> None:
        # The str parameter converts to a datetime through the stamp
        # converter before binding.
        queries_dbtype_override.insert_dbtype_override(conn=pymysql_conn, id_=DBTYPE_ID, happened_at=STAMP)

    @pytest.mark.dependency(name="PymysqlDbtypeOverride::get", depends=["PymysqlDbtypeOverride::insert"])
    def test_get_dbtype_override(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_dbtype_override.get_dbtype_override(conn=pymysql_conn, id_=DBTYPE_ID)

        assert result is not None
        assert result == models.TestDbtypeOverride(id_=DBTYPE_ID, happened_at=STAMP)
        assert isinstance(result.happened_at, str)
        # The database stored a real datetime, not the string.
        with pymysql_conn.cursor() as cur:
            cur.execute("SELECT happened_at FROM test_dbtype_override WHERE id = %s", (DBTYPE_ID,))
            row = cur.fetchone()
        assert row is not None
        assert row[0] == datetime.datetime.fromisoformat(STAMP)

    @pytest.mark.dependency(depends=["PymysqlDbtypeOverride::get"])
    def test_delete_dbtype_override(self, pymysql_conn: pymysql.Connection) -> None:
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_dbtype_override WHERE id = %s", (DBTYPE_ID,))
        assert queries_dbtype_override.get_dbtype_override(conn=pymysql_conn, id_=DBTYPE_ID) is None
