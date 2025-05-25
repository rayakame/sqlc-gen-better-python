# Copyright (c) 2025 Rayakame

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
from __future__ import annotations

import datetime
import decimal
import json
import math
import random
import sqlite3
from collections import UserString

import pytest

from test.driver_sqlite3.attrs.functions import models
from test.driver_sqlite3.attrs.functions import queries


class TestSqlite3AttrsFunctions:
    @pytest.fixture(scope="session")
    def override_model(self) -> models.TestTypeOverride:
        return models.TestTypeOverride(id=random.randint(1, 10000000), text_test=UserString("Test"))

    @pytest.fixture(scope="session")
    def model(self) -> models.TestSqliteType:
        return models.TestSqliteType(
            id=random.randint(1, 10000000),
            int_test=42,
            bigint_test=9_007_199_254_740_991,
            smallint_test=32_767,
            tinyint_test=255,
            int2_test=12_345,
            int8_test=123_456_789,
            bigserial_test=1,
            blob_test=memoryview(b"\x00\x01\x02hello"),
            real_test=math.pi,
            double_test=math.e,
            double_precision_test=1.41421,
            float_test=9.81,
            numeric_test=123.456,
            decimal_test=decimal.Decimal("789.0123"),
            bool_test=True,
            boolean_test=False,
            date_test=datetime.date(2025, 1, 1),
            datetime_test=datetime.datetime(2025, 1, 1, 12),
            timestamp_test=datetime.datetime.now(),
            character_test="ABCDEFGHIJ",
            varchar_test="Hello varchar",
            varyingcharacter_test="VarChar variant",
            nchar_test="ABCDEFGHIJ",
            nativecharacter_test="NativeChar",
            nvarchar_test="Olá mundo",
            text_test="Some text",
            clob_test="Some clob data",
            json_test=json.dumps({"foo": "bar"}),
        )

    @pytest.fixture(scope="session")
    def inner_model(self, model: models.TestSqliteType) -> models.TestInnerSqliteType:
        return models.TestInnerSqliteType(
            table_id=model.id,
            int_test=None,
            bigint_test=model.bigint_test,
            smallint_test=model.smallint_test,
            tinyint_test=model.tinyint_test,
            int2_test=model.int2_test,
            int8_test=model.int8_test,
            bigserial_test=model.bigserial_test,
            blob_test=None,
            real_test=model.real_test,
            double_test=model.double_test,
            double_precision_test=model.double_precision_test,
            float_test=model.float_test,
            numeric_test=model.numeric_test,
            decimal_test=None,
            bool_test=None,
            boolean_test=None,
            date_test=None,
            datetime_test=None,
            timestamp_test=None,
            character_test=model.character_test,
            varchar_test=model.varchar_test,
            varyingcharacter_test=model.varyingcharacter_test,
            nchar_test=model.nchar_test,
            nativecharacter_test=model.nativecharacter_test,
            nvarchar_test=model.nvarchar_test,
            text_test=model.text_test,
            clob_test=model.clob_test,
            json_test=model.json_test,
        )

    @pytest.mark.dependency(name="Sqlite3TestAttrsFunctions::insert")
    def test_insert(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        queries.insert_one_sqlite_type(
            conn=sqlite3_conn,
            id_=model.id,
            int_test=model.int_test,
            bigint_test=model.bigint_test,
            smallint_test=model.smallint_test,
            tinyint_test=model.tinyint_test,
            int2_test=model.int2_test,
            int8_test=model.int8_test,
            bigserial_test=model.bigserial_test,
            blob_test=model.blob_test,
            real_test=model.real_test,
            double_test=model.double_test,
            double_precision_test=model.double_precision_test,
            float_test=model.float_test,
            numeric_test=model.numeric_test,
            decimal_test=model.decimal_test,
            bool_test=model.bool_test,
            boolean_test=model.boolean_test,
            date_test=model.date_test,
            datetime_test=model.datetime_test,
            timestamp_test=model.timestamp_test,
            character_test=model.character_test,
            varchar_test=model.varchar_test,
            varyingcharacter_test=model.varyingcharacter_test,
            nchar_test=model.nchar_test,
            nativecharacter_test=model.nativecharacter_test,
            nvarchar_test=model.nvarchar_test,
            text_test=model.text_test,
            clob_test=model.clob_test,
            json_test=model.json_test,
        )

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::inner_insert", depends=["Sqlite3TestAttrsFunctions::insert"]
    )
    def test_inner_insert(
        self,
        sqlite3_conn: sqlite3.Connection,
        inner_model: models.TestInnerSqliteType,
    ) -> None:
        queries.insert_one_inner_sqlite_type(
            conn=sqlite3_conn,
            table_id=inner_model.table_id,
            int_test=inner_model.int_test,
            bigint_test=inner_model.bigint_test,
            smallint_test=inner_model.smallint_test,
            tinyint_test=inner_model.tinyint_test,
            int2_test=inner_model.int2_test,
            int8_test=inner_model.int8_test,
            bigserial_test=inner_model.bigserial_test,
            blob_test=inner_model.blob_test,
            real_test=inner_model.real_test,
            double_test=inner_model.double_test,
            double_precision_test=inner_model.double_precision_test,
            float_test=inner_model.float_test,
            numeric_test=inner_model.numeric_test,
            decimal_test=inner_model.decimal_test,
            bool_test=inner_model.bool_test,
            boolean_test=inner_model.boolean_test,
            date_test=inner_model.date_test,
            datetime_test=inner_model.datetime_test,
            timestamp_test=inner_model.timestamp_test,
            character_test=inner_model.character_test,
            varchar_test=inner_model.varchar_test,
            varyingcharacter_test=inner_model.varyingcharacter_test,
            nchar_test=inner_model.nchar_test,
            nativecharacter_test=inner_model.nativecharacter_test,
            nvarchar_test=inner_model.nvarchar_test,
            text_test=inner_model.text_test,
            clob_test=inner_model.clob_test,
            json_test=inner_model.json_test,
        )

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_one", depends=["Sqlite3TestAttrsFunctions::inner_insert"]
    )
    def test_get_one(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_sqlite_type(conn=sqlite3_conn, id_=model.id)

        assert result is not None

        assert isinstance(result, models.TestSqliteType)

        assert result == model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_one_none", depends=["Sqlite3TestAttrsFunctions::get_one"]
    )
    def test_get_one_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_sqlite_type(conn=sqlite3_conn, id_=0)

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_one_inner", depends=["Sqlite3TestAttrsFunctions::get_one_none"]
    )
    def test_get_one_inner(
        self,
        sqlite3_conn: sqlite3.Connection,
        inner_model: models.TestInnerSqliteType,
    ) -> None:
        result = queries.get_one_inner_sqlite_type(conn=sqlite3_conn, table_id=inner_model.table_id)

        assert result is not None

        assert isinstance(result, models.TestInnerSqliteType)
        assert result == inner_model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_one_inner_none", depends=["Sqlite3TestAttrsFunctions::get_one_inner"]
    )
    def test_get_one_inner_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_inner_sqlite_type(conn=sqlite3_conn, table_id=0)

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_date", depends=["Sqlite3TestAttrsFunctions::get_one_inner_none"]
    )
    def test_get_date(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_date(conn=sqlite3_conn, id_=model.id, date_test=model.date_test)

        assert result is not None

        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_date_none", depends=["Sqlite3TestAttrsFunctions::get_date"]
    )
    def test_get_date_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_date(conn=sqlite3_conn, id_=0, date_test=datetime.date.today())

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_datetime", depends=["Sqlite3TestAttrsFunctions::get_date_none"]
    )
    def test_get_datetime(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_datetime(conn=sqlite3_conn, id_=model.id, datetime_test=model.datetime_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_datetime_none", depends=["Sqlite3TestAttrsFunctions::get_datetime"]
    )
    def test_get_datetime_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_datetime(conn=sqlite3_conn, id_=0, datetime_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_timestamp", depends=["Sqlite3TestAttrsFunctions::get_datetime_none"]
    )
    def test_get_timestamp(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_timestamp(conn=sqlite3_conn, id_=model.id, timestamp_test=model.timestamp_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.timestamp_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_timestamp_none", depends=["Sqlite3TestAttrsFunctions::get_timestamp"]
    )
    def test_get_timestamp_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_timestamp(conn=sqlite3_conn, id_=0, timestamp_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_bool", depends=["Sqlite3TestAttrsFunctions::get_timestamp_none"]
    )
    def test_get_bool(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_bool(conn=sqlite3_conn, id_=model.id, bool_test=model.bool_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.bool_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_bool_none", depends=["Sqlite3TestAttrsFunctions::get_bool"]
    )
    def test_get_bool_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_bool(conn=sqlite3_conn, id_=0, bool_test=False)

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_boolean", depends=["Sqlite3TestAttrsFunctions::get_bool_none"]
    )
    def test_get_boolean(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_boolean(conn=sqlite3_conn, id_=model.id, boolean_test=model.boolean_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.boolean_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_boolean_none", depends=["Sqlite3TestAttrsFunctions::get_boolean"]
    )
    def test_get_boolean_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_boolean(conn=sqlite3_conn, id_=0, boolean_test=True)

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_decimal", depends=["Sqlite3TestAttrsFunctions::get_boolean_none"]
    )
    def test_get_decimal(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_decimal(conn=sqlite3_conn, id_=model.id, decimal_test=model.decimal_test)

        assert result is not None

        assert isinstance(result, decimal.Decimal)
        assert result == model.decimal_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_decimal_none", depends=["Sqlite3TestAttrsFunctions::get_decimal"]
    )
    def test_get_decimal_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_decimal(conn=sqlite3_conn, id_=0, decimal_test=decimal.Decimal("0.1"))

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_blob", depends=["Sqlite3TestAttrsFunctions::get_decimal_none"]
    )
    def test_get_blob(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_blob(conn=sqlite3_conn, id_=model.id, blob_test=model.blob_test)

        assert result is not None

        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_blob_none", depends=["Sqlite3TestAttrsFunctions::get_blob"]
    )
    def test_get_blob_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_blob(conn=sqlite3_conn, id_=0, blob_test=memoryview(b"test"))

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many", depends=["Sqlite3TestAttrsFunctions::get_blob_none"]
    )
    def test_get_many(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_sqlite_type(conn=sqlite3_conn, id_=model.id)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestSqliteType)

        assert results[0] == model

        results = result()
        assert isinstance(results[0], models.TestSqliteType)

        assert results[0] == model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_iter", depends=["Sqlite3TestAttrsFunctions::get_many"]
    )
    def test_get_many_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_sqlite_type(conn=sqlite3_conn, id_=model.id):
            assert result is not None
            assert isinstance(result, models.TestSqliteType)

            assert result == model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_inner", depends=["Sqlite3TestAttrsFunctions::get_many_iter"]
    )
    def test_get_many_inner(self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType) -> None:
        result = queries.get_many_inner_sqlite_type(conn=sqlite3_conn, table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestInnerSqliteType)

        assert results[0] == inner_model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_inner_iter", depends=["Sqlite3TestAttrsFunctions::get_many_inner"]
    )
    def test_get_many_inner_iter(
        self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType
    ) -> None:
        for result in queries.get_many_inner_sqlite_type(conn=sqlite3_conn, table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerSqliteType)

            assert result == inner_model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_nullable_inner",
        depends=["Sqlite3TestAttrsFunctions::get_many_inner_iter"],
    )
    def test_get_many_nullable_inner(
        self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType
    ) -> None:
        result = queries.get_many_nullable_inner_sqlite_type(
            conn=sqlite3_conn, table_id=inner_model.table_id, int_test=inner_model.int_test
        )

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestInnerSqliteType)

        assert results[0] == inner_model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_nullable_inner_iter",
        depends=["Sqlite3TestAttrsFunctions::get_many_nullable_inner"],
    )
    def test_get_many_nullable_inner_iter(
        self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType
    ) -> None:
        for result in queries.get_many_nullable_inner_sqlite_type(
            conn=sqlite3_conn, table_id=inner_model.table_id, int_test=inner_model.int_test
        ):
            assert result is not None
            assert isinstance(result, models.TestInnerSqliteType)

            assert result == inner_model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_date",
        depends=["Sqlite3TestAttrsFunctions::get_many_nullable_inner_iter"],
    )
    def test_get_many_date(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_date(conn=sqlite3_conn, id_=model.id, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.date)

        assert results[0] == model.date_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_date_iter", depends=["Sqlite3TestAttrsFunctions::get_many_date"]
    )
    def test_get_many_date_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_date(conn=sqlite3_conn, id_=model.id, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)

            assert result == model.date_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_datetime", depends=["Sqlite3TestAttrsFunctions::get_many_date_iter"]
    )
    def test_get_many_datetime(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_datetime(conn=sqlite3_conn, id_=model.id, datetime_test=model.datetime_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.datetime)

        assert results[0] == model.datetime_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_datetime_iter",
        depends=["Sqlite3TestAttrsFunctions::get_many_datetime"],
    )
    def test_get_many_datetime_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_datetime(conn=sqlite3_conn, id_=model.id, datetime_test=model.datetime_test):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.datetime_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_timestamp",
        depends=["Sqlite3TestAttrsFunctions::get_many_datetime_iter"],
    )
    def test_get_many_timestamp(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_timestamp(conn=sqlite3_conn, id_=model.id, timestamp_test=model.timestamp_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.datetime)

        assert results[0] == model.timestamp_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_timestamp_iter",
        depends=["Sqlite3TestAttrsFunctions::get_many_timestamp"],
    )
    def test_get_many_timestamp_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_timestamp(conn=sqlite3_conn, id_=model.id, timestamp_test=model.timestamp_test):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.timestamp_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_bool", depends=["Sqlite3TestAttrsFunctions::get_many_timestamp_iter"]
    )
    def test_get_many_bool(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_bool(conn=sqlite3_conn, id_=model.id, bool_test=model.bool_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], bool)

        assert results[0] == model.bool_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_bool_iter", depends=["Sqlite3TestAttrsFunctions::get_many_bool"]
    )
    def test_get_many_bool_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_bool(conn=sqlite3_conn, id_=model.id, bool_test=model.bool_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.bool_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_boolean", depends=["Sqlite3TestAttrsFunctions::get_many_bool_iter"]
    )
    def test_get_many_boolean(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_boolean(conn=sqlite3_conn, id_=model.id, boolean_test=model.boolean_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], bool)

        assert results[0] == model.boolean_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_boolean_iter", depends=["Sqlite3TestAttrsFunctions::get_many_boolean"]
    )
    def test_get_many_boolean_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_boolean(conn=sqlite3_conn, id_=model.id, boolean_test=model.boolean_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.boolean_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_decimal", depends=["Sqlite3TestAttrsFunctions::get_many_boolean_iter"]
    )
    def test_get_many_decimal(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_decimal(conn=sqlite3_conn, id_=model.id, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], decimal.Decimal)

        assert results[0] == model.decimal_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_decimal_iter", depends=["Sqlite3TestAttrsFunctions::get_many_decimal"]
    )
    def test_get_many_decimal_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_decimal(conn=sqlite3_conn, id_=model.id, decimal_test=model.decimal_test):
            assert result is not None
            assert isinstance(result, decimal.Decimal)

            assert result == model.decimal_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_blob", depends=["Sqlite3TestAttrsFunctions::get_many_decimal_iter"]
    )
    def test_get_many_blob(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_blob(conn=sqlite3_conn, id_=model.id, blob_test=model.blob_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], memoryview)

        assert results[0] == model.blob_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_blob_iter", depends=["Sqlite3TestAttrsFunctions::get_many_blob"]
    )
    def test_get_many_blob_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_blob(conn=sqlite3_conn, id_=model.id, blob_test=model.blob_test):
            assert result is not None
            assert isinstance(result, memoryview)

            assert result == model.blob_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::insert_result", depends=["Sqlite3TestAttrsFunctions::get_many_blob_iter"]
    )
    def test_insert_result(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.insert_result_one_sqlite_type(
            conn=sqlite3_conn,
            id_=model.id + 1,
            int_test=model.int_test,
            bigint_test=model.bigint_test,
            smallint_test=model.smallint_test,
            tinyint_test=model.tinyint_test,
            int2_test=model.int2_test,
            int8_test=model.int8_test,
            bigserial_test=model.bigserial_test,
            blob_test=model.blob_test,
            real_test=model.real_test,
            double_test=model.double_test,
            double_precision_test=model.double_precision_test,
            float_test=model.float_test,
            numeric_test=model.numeric_test,
            decimal_test=model.decimal_test,
            bool_test=model.bool_test,
            boolean_test=model.boolean_test,
            date_test=model.date_test,
            datetime_test=model.datetime_test,
            timestamp_test=model.timestamp_test,
            character_test=model.character_test,
            varchar_test=model.varchar_test,
            varyingcharacter_test=model.varyingcharacter_test,
            nchar_test=model.nchar_test,
            nativecharacter_test=model.nativecharacter_test,
            nvarchar_test=model.nvarchar_test,
            text_test=model.text_test,
            clob_test=model.clob_test,
            json_test=model.json_test,
        )
        assert isinstance(result, sqlite3.Cursor)

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::update_result", depends=["Sqlite3TestAttrsFunctions::insert_result"]
    )
    def test_update_result(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.update_result_one_sqlite_type(conn=sqlite3_conn, id_=model.id + 1)
        assert isinstance(result, sqlite3.Cursor)

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::delete_result", depends=["Sqlite3TestAttrsFunctions::update_result"]
    )
    def test_delete_result(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.delete_result_one_sqlite_type(conn=sqlite3_conn, id_=model.id + 1)
        assert isinstance(result, sqlite3.Cursor)

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::insert_rows", depends=["Sqlite3TestAttrsFunctions::delete_result"]
    )
    def test_insert_rows(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.insert_rows_one_sqlite_type(
            conn=sqlite3_conn,
            id_=model.id + 2,
            int_test=model.int_test,
            bigint_test=model.bigint_test,
            smallint_test=model.smallint_test,
            tinyint_test=model.tinyint_test,
            int2_test=model.int2_test,
            int8_test=model.int8_test,
            bigserial_test=model.bigserial_test,
            blob_test=model.blob_test,
            real_test=model.real_test,
            double_test=model.double_test,
            double_precision_test=model.double_precision_test,
            float_test=model.float_test,
            numeric_test=model.numeric_test,
            decimal_test=model.decimal_test,
            bool_test=model.bool_test,
            boolean_test=model.boolean_test,
            date_test=model.date_test,
            datetime_test=model.datetime_test,
            timestamp_test=model.timestamp_test,
            character_test=model.character_test,
            varchar_test=model.varchar_test,
            varyingcharacter_test=model.varyingcharacter_test,
            nchar_test=model.nchar_test,
            nativecharacter_test=model.nativecharacter_test,
            nvarchar_test=model.nvarchar_test,
            text_test=model.text_test,
            clob_test=model.clob_test,
            json_test=model.json_test,
        )
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::update_rows", depends=["Sqlite3TestAttrsFunctions::insert_rows"]
    )
    def test_update_rows(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.update_rows_one_sqlite_type(conn=sqlite3_conn, id_=model.id + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::delete_rows", depends=["Sqlite3TestAttrsFunctions::update_rows"]
    )
    def test_delete_rows(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.delete_rows_one_sqlite_type(conn=sqlite3_conn, id_=model.id + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::create_table_rows", depends=["Sqlite3TestAttrsFunctions::delete_rows"]
    )
    def test_create_table_rows(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.create_rows_table(conn=sqlite3_conn)
        assert isinstance(result, int)
        sqlite3_conn.execute("DROP TABLE test_create_rows_table;")
        sqlite3_conn.commit()
        assert result == -1

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::insert_last_id", depends=["Sqlite3TestAttrsFunctions::create_table_rows"]
    )
    def test_insert_last_id(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.insert_last_id_one_sqlite_type(
            conn=sqlite3_conn,
            id_=model.id + 3,
            int_test=model.int_test,
            bigint_test=model.bigint_test,
            smallint_test=model.smallint_test,
            tinyint_test=model.tinyint_test,
            int2_test=model.int2_test,
            int8_test=model.int8_test,
            bigserial_test=model.bigserial_test,
            blob_test=model.blob_test,
            real_test=model.real_test,
            double_test=model.double_test,
            double_precision_test=model.double_precision_test,
            float_test=model.float_test,
            numeric_test=model.numeric_test,
            decimal_test=model.decimal_test,
            bool_test=model.bool_test,
            boolean_test=model.boolean_test,
            date_test=model.date_test,
            datetime_test=model.datetime_test,
            timestamp_test=model.timestamp_test,
            character_test=model.character_test,
            varchar_test=model.varchar_test,
            varyingcharacter_test=model.varyingcharacter_test,
            nchar_test=model.nchar_test,
            nativecharacter_test=model.nativecharacter_test,
            nvarchar_test=model.nvarchar_test,
            text_test=model.text_test,
            clob_test=model.clob_test,
            json_test=model.json_test,
        )
        assert isinstance(result, int)
        assert result == model.id + 3

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::update_last_id", depends=["Sqlite3TestAttrsFunctions::insert_last_id"]
    )
    def test_update_last_id(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.update_last_id_one_sqlite_type(conn=sqlite3_conn, id_=model.id + 3)
        assert isinstance(result, int)
        assert result == model.id + 3

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::delete_last_id", depends=["Sqlite3TestAttrsFunctions::update_last_id"]
    )
    def test_delete_last_id(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.delete_last_id_one_sqlite_type(conn=sqlite3_conn, id_=model.id + 3)
        assert isinstance(result, int)
        assert result == model.id + 3

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::delete_sqlite_type", depends=["Sqlite3TestAttrsFunctions::delete_last_id"]
    )
    def test_delete_sqlite_type(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        queries.delete_one_sqlite_type(conn=sqlite3_conn, id_=model.id)

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::delete_inner_sqlite_type",
        depends=["Sqlite3TestAttrsFunctions::delete_sqlite_type"],
    )
    def test_delete_inner_sqlite_type(
        self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType
    ) -> None:
        queries.delete_one_test_inner_sqlite_type(conn=sqlite3_conn, table_id=inner_model.table_id)

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::insert_type_override",
    )
    def test_insert_type_override(
        self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride
    ) -> None:
        queries.insert_type_override(conn=sqlite3_conn, id_=override_model.id, text_test=override_model.text_test)

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_one_type_override",
        depends=["Sqlite3TestAttrsFunctions::insert_type_override"],
    )
    def test_get_one_type_override(
        self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride
    ) -> None:
        result = queries.get_one_type_override(conn=sqlite3_conn, id_=override_model.id)
        assert result is not None
        assert result == override_model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_one_type_override_none",
        depends=["Sqlite3TestAttrsFunctions::get_one_type_override"],
    )
    def test_get_one_type_override_none(
        self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride
    ) -> None:
        result = queries.get_one_type_override(conn=sqlite3_conn, id_=override_model.id - 1)
        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_type_override",
        depends=["Sqlite3TestAttrsFunctions::get_one_type_override_none"],
    )
    def test_get_many_type_override(
        self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride
    ) -> None:
        result = queries.get_many_type_override(conn=sqlite3_conn, id_=override_model.id)
        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestTypeOverride)

        assert results[0] == override_model

        results = result()
        assert isinstance(results[0], models.TestTypeOverride)

        assert results[0] == override_model

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_one_text_type_override",
        depends=["Sqlite3TestAttrsFunctions::get_many_type_override"],
    )
    def test_get_one_text_type_override(
        self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride
    ) -> None:
        result = queries.get_one_text_type_override(conn=sqlite3_conn, id_=override_model.id)
        assert result is not None
        assert result == override_model.text_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_one_text_type_override_none",
        depends=["Sqlite3TestAttrsFunctions::get_one_text_type_override"],
    )
    def test_get_one_text_type_override_none(
        self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride
    ) -> None:
        result = queries.get_one_text_type_override(conn=sqlite3_conn, id_=override_model.id - 1)
        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::get_many_text_type_override",
        depends=["Sqlite3TestAttrsFunctions::get_one_text_type_override_none"],
    )
    def test_get_many_text_type_override(
        self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride
    ) -> None:
        result = queries.get_many_text_type_override(conn=sqlite3_conn, id_=override_model.id)
        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], UserString)

        assert results[0] == override_model.text_test

        results = result()
        assert isinstance(results[0], UserString)

        assert results[0] == override_model.text_test

    @pytest.mark.dependency(
        name="Sqlite3TestAttrsFunctions::delete_type_override",
        depends=["Sqlite3TestAttrsFunctions::get_many_text_type_override"],
    )
    def test_delete_type_override(
        self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride
    ) -> None:
        queries.delete_type_override(conn=sqlite3_conn, id_=override_model.id)
