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
from __future__ import annotations

import datetime
import decimal
import json
import math
import pathlib
import random
import sqlite3
from collections import UserString

import pytest

from test.driver_sqlite3.dataclass.functions import models
from test.driver_sqlite3.dataclass.functions import queries
from test.driver_sqlite3.dataclass.functions import queries_any_param
from test.driver_sqlite3.dataclass.functions import queries_case
from test.driver_sqlite3.dataclass.functions import queries_override_adapter
from test.driver_sqlite3.dataclass.functions import queries_override_converter
from test.driver_sqlite3.dataclass.functions import queries_slice
from test.driver_sqlite3.dataclass.functions import queries_unknown_override

OVERRIDE_PRICE = 12.5
OVERRIDE_HAPPENED_AT = datetime.datetime(2026, 7, 19, 12, 30)
CASE_DT = datetime.datetime(2026, 7, 19, 8, 15)
CASE_DEC = decimal.Decimal("12.34")
RESERVED_ARG_ID = 525252
UNKNOWN_OVERRIDE_ID = 545454
ANY_PARAM_ID = 565656
SLICE_ID_BASE = 585858
SLICE_ROW_COUNT = 4
SLICE_NAME_MATCH_COUNT = 3


class TestSqlite3DataclassFunctions:
    @pytest.fixture(scope="session")
    def override_model(self) -> models.TestTypeOverride:
        return models.TestTypeOverride(id_=random.randint(1, 10000000), text_test=UserString("Test"))

    @pytest.fixture(scope="session")
    def model(self) -> models.TestSqliteType:
        return models.TestSqliteType(
            id_=random.randint(1, 10000000),
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
            table_id=model.id_,
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

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::insert")
    def test_insert(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        queries.insert_one_sqlite_type(
            conn=sqlite3_conn,
            id_=model.id_,
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

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::inner_insert", depends=["Sqlite3TestDataclassFunctions::insert"])
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

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_one", depends=["Sqlite3TestDataclassFunctions::inner_insert"])
    def test_get_one(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_sqlite_type(conn=sqlite3_conn, id_=model.id_)

        assert result is not None

        assert isinstance(result, models.TestSqliteType)

        assert result == model

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_one_none", depends=["Sqlite3TestDataclassFunctions::get_one"])
    def test_get_one_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_sqlite_type(conn=sqlite3_conn, id_=0)

        assert result is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_one_inner", depends=["Sqlite3TestDataclassFunctions::get_one_none"])
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
        name="Sqlite3TestDataclassFunctions::get_one_inner_none",
        depends=["Sqlite3TestDataclassFunctions::get_one_inner"],
    )
    def test_get_one_inner_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_inner_sqlite_type(conn=sqlite3_conn, table_id=0)

        assert result is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_date", depends=["Sqlite3TestDataclassFunctions::get_one_inner_none"])
    def test_get_date(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_date(conn=sqlite3_conn, id_=model.id_, date_test=model.date_test)

        assert result is not None

        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_date_none", depends=["Sqlite3TestDataclassFunctions::get_date"])
    def test_get_date_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_date(conn=sqlite3_conn, id_=0, date_test=datetime.date.today())

        assert result is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_datetime", depends=["Sqlite3TestDataclassFunctions::get_date_none"])
    def test_get_datetime(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_datetime(conn=sqlite3_conn, id_=model.id_, datetime_test=model.datetime_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_datetime_none", depends=["Sqlite3TestDataclassFunctions::get_datetime"])
    def test_get_datetime_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_datetime(conn=sqlite3_conn, id_=0, datetime_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_timestamp",
        depends=["Sqlite3TestDataclassFunctions::get_datetime_none"],
    )
    def test_get_timestamp(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_timestamp(conn=sqlite3_conn, id_=model.id_, timestamp_test=model.timestamp_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.timestamp_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_timestamp_none",
        depends=["Sqlite3TestDataclassFunctions::get_timestamp"],
    )
    def test_get_timestamp_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_timestamp(conn=sqlite3_conn, id_=0, timestamp_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_bool", depends=["Sqlite3TestDataclassFunctions::get_timestamp_none"])
    def test_get_bool(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_bool(conn=sqlite3_conn, id_=model.id_, bool_test=model.bool_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.bool_test

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_bool_none", depends=["Sqlite3TestDataclassFunctions::get_bool"])
    def test_get_bool_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_bool(conn=sqlite3_conn, id_=0, bool_test=False)

        assert result is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_boolean", depends=["Sqlite3TestDataclassFunctions::get_bool_none"])
    def test_get_boolean(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_boolean(conn=sqlite3_conn, id_=model.id_, boolean_test=model.boolean_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.boolean_test

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_boolean_none", depends=["Sqlite3TestDataclassFunctions::get_boolean"])
    def test_get_boolean_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_boolean(conn=sqlite3_conn, id_=0, boolean_test=True)

        assert result is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_decimal", depends=["Sqlite3TestDataclassFunctions::get_boolean_none"])
    def test_get_decimal(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_decimal(conn=sqlite3_conn, id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None

        assert isinstance(result, decimal.Decimal)
        assert result == model.decimal_test

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_decimal_none", depends=["Sqlite3TestDataclassFunctions::get_decimal"])
    def test_get_decimal_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_decimal(conn=sqlite3_conn, id_=0, decimal_test=decimal.Decimal("0.1"))

        assert result is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_blob", depends=["Sqlite3TestDataclassFunctions::get_decimal_none"])
    def test_get_blob(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_blob(conn=sqlite3_conn, id_=model.id_, blob_test=model.blob_test)

        assert result is not None

        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_blob_none", depends=["Sqlite3TestDataclassFunctions::get_blob"])
    def test_get_blob_none(
        self,
        sqlite3_conn: sqlite3.Connection,
    ) -> None:
        result = queries.get_one_blob(conn=sqlite3_conn, id_=0, blob_test=memoryview(b"test"))

        assert result is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_many", depends=["Sqlite3TestDataclassFunctions::get_blob_none"])
    def test_get_many(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_sqlite_type(conn=sqlite3_conn, id_=model.id_)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestSqliteType)

        assert results[0] == model
        results = result()
        assert isinstance(results[0], models.TestSqliteType)

        assert results[0] == model

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_many_iter", depends=["Sqlite3TestDataclassFunctions::get_many"])
    def test_get_many_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_sqlite_type(conn=sqlite3_conn, id_=model.id_):
            assert result is not None
            assert isinstance(result, models.TestSqliteType)

            assert result == model

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_many_inner", depends=["Sqlite3TestDataclassFunctions::get_many_iter"])
    def test_get_many_inner(self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType) -> None:
        result = queries.get_many_inner_sqlite_type(conn=sqlite3_conn, table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestInnerSqliteType)

        assert results[0] == inner_model

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_inner_iter",
        depends=["Sqlite3TestDataclassFunctions::get_many_inner"],
    )
    def test_get_many_inner_iter(self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType) -> None:
        for result in queries.get_many_inner_sqlite_type(conn=sqlite3_conn, table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerSqliteType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_nullable_inner",
        depends=["Sqlite3TestDataclassFunctions::get_many_inner_iter"],
    )
    async def test_get_many_nullable_inner(self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType) -> None:
        result = queries.get_many_nullable_inner_sqlite_type(conn=sqlite3_conn, table_id=inner_model.table_id, int_test=inner_model.int_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestInnerSqliteType)

        assert results[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_nullable_inner_iter",
        depends=["Sqlite3TestDataclassFunctions::get_many_nullable_inner"],
    )
    async def test_get_many_nullable_inner_iter(self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType) -> None:
        for result in queries.get_many_nullable_inner_sqlite_type(conn=sqlite3_conn, table_id=inner_model.table_id, int_test=inner_model.int_test):
            assert result is not None
            assert isinstance(result, models.TestInnerSqliteType)

            assert result == inner_model

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_date",
        depends=["Sqlite3TestDataclassFunctions::get_many_nullable_inner_iter"],
    )
    def test_get_many_date(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_date(conn=sqlite3_conn, id_=model.id_, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.date)

        assert results[0] == model.date_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_date_iter",
        depends=["Sqlite3TestDataclassFunctions::get_many_date"],
    )
    def test_get_many_date_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_date(conn=sqlite3_conn, id_=model.id_, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)

            assert result == model.date_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_datetime",
        depends=["Sqlite3TestDataclassFunctions::get_many_date_iter"],
    )
    def test_get_many_datetime(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_datetime(conn=sqlite3_conn, id_=model.id_, datetime_test=model.datetime_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.datetime)

        assert results[0] == model.datetime_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_datetime_iter",
        depends=["Sqlite3TestDataclassFunctions::get_many_datetime"],
    )
    def test_get_many_datetime_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_datetime(conn=sqlite3_conn, id_=model.id_, datetime_test=model.datetime_test):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.datetime_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_timestamp",
        depends=["Sqlite3TestDataclassFunctions::get_many_datetime_iter"],
    )
    def test_get_many_timestamp(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_timestamp(conn=sqlite3_conn, id_=model.id_, timestamp_test=model.timestamp_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.datetime)

        assert results[0] == model.timestamp_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_timestamp_iter",
        depends=["Sqlite3TestDataclassFunctions::get_many_timestamp"],
    )
    def test_get_many_timestamp_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_timestamp(conn=sqlite3_conn, id_=model.id_, timestamp_test=model.timestamp_test):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.timestamp_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_bool",
        depends=["Sqlite3TestDataclassFunctions::get_many_timestamp_iter"],
    )
    def test_get_many_bool(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_bool(conn=sqlite3_conn, id_=model.id_, bool_test=model.bool_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], bool)

        assert results[0] == model.bool_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_bool_iter",
        depends=["Sqlite3TestDataclassFunctions::get_many_bool"],
    )
    def test_get_many_bool_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_bool(conn=sqlite3_conn, id_=model.id_, bool_test=model.bool_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.bool_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_boolean",
        depends=["Sqlite3TestDataclassFunctions::get_many_bool_iter"],
    )
    def test_get_many_boolean(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_boolean(conn=sqlite3_conn, id_=model.id_, boolean_test=model.boolean_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], bool)

        assert results[0] == model.boolean_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_boolean_iter",
        depends=["Sqlite3TestDataclassFunctions::get_many_boolean"],
    )
    def test_get_many_boolean_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_boolean(conn=sqlite3_conn, id_=model.id_, boolean_test=model.boolean_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.boolean_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_decimal",
        depends=["Sqlite3TestDataclassFunctions::get_many_boolean_iter"],
    )
    def test_get_many_decimal(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_decimal(conn=sqlite3_conn, id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], decimal.Decimal)

        assert results[0] == model.decimal_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_decimal_iter",
        depends=["Sqlite3TestDataclassFunctions::get_many_decimal"],
    )
    def test_get_many_decimal_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_decimal(conn=sqlite3_conn, id_=model.id_, decimal_test=model.decimal_test):
            assert result is not None
            assert isinstance(result, decimal.Decimal)

            assert result == model.decimal_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_blob",
        depends=["Sqlite3TestDataclassFunctions::get_many_decimal_iter"],
    )
    def test_get_many_blob(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_blob(conn=sqlite3_conn, id_=model.id_, blob_test=model.blob_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], memoryview)

        assert results[0] == model.blob_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_blob_iter",
        depends=["Sqlite3TestDataclassFunctions::get_many_blob"],
    )
    def test_get_many_blob_iter(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_blob(conn=sqlite3_conn, id_=model.id_, blob_test=model.blob_test):
            assert result is not None
            assert isinstance(result, memoryview)

            assert result == model.blob_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::insert_result",
        depends=["Sqlite3TestDataclassFunctions::get_many_blob_iter"],
    )
    def test_insert_result(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.insert_result_one_sqlite_type(
            conn=sqlite3_conn,
            id_=model.id_ + 1,
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

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::update_result", depends=["Sqlite3TestDataclassFunctions::insert_result"])
    def test_update_result(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.update_result_one_sqlite_type(conn=sqlite3_conn, id_=model.id_ + 1)
        assert isinstance(result, sqlite3.Cursor)

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::delete_result", depends=["Sqlite3TestDataclassFunctions::update_result"])
    def test_delete_result(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.delete_result_one_sqlite_type(conn=sqlite3_conn, id_=model.id_ + 1)
        assert isinstance(result, sqlite3.Cursor)

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::insert_rows", depends=["Sqlite3TestDataclassFunctions::delete_result"])
    def test_insert_rows(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.insert_rows_one_sqlite_type(
            conn=sqlite3_conn,
            id_=model.id_ + 2,
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

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::update_rows", depends=["Sqlite3TestDataclassFunctions::insert_rows"])
    def test_update_rows(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.update_rows_one_sqlite_type(conn=sqlite3_conn, id_=model.id_ + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::delete_rows", depends=["Sqlite3TestDataclassFunctions::update_rows"])
    def test_delete_rows(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.delete_rows_one_sqlite_type(conn=sqlite3_conn, id_=model.id_ + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::create_table_rows", depends=["Sqlite3TestDataclassFunctions::delete_rows"])
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
        name="Sqlite3TestDataclassFunctions::insert_last_id",
        depends=["Sqlite3TestDataclassFunctions::create_table_rows"],
    )
    def test_insert_last_id(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.insert_last_id_one_sqlite_type(
            conn=sqlite3_conn,
            id_=model.id_ + 3,
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
        assert result == model.id_ + 3

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::update_last_id", depends=["Sqlite3TestDataclassFunctions::insert_last_id"])
    def test_update_last_id(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.update_last_id_one_sqlite_type(conn=sqlite3_conn, id_=model.id_ + 3)
        assert isinstance(result, int)
        assert result == model.id_ + 3

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::delete_last_id", depends=["Sqlite3TestDataclassFunctions::update_last_id"])
    def test_delete_last_id(
        self,
        sqlite3_conn: sqlite3.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.delete_last_id_one_sqlite_type(conn=sqlite3_conn, id_=model.id_ + 3)
        assert isinstance(result, int)
        assert result == model.id_ + 3

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::delete_sqlite_type",
        depends=["Sqlite3TestDataclassFunctions::delete_last_id"],
    )
    def test_delete_sqlite_type(self, sqlite3_conn: sqlite3.Connection, model: models.TestSqliteType) -> None:
        queries.delete_one_sqlite_type(conn=sqlite3_conn, id_=model.id_)

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::delete_inner_sqlite_type",
        depends=["Sqlite3TestDataclassFunctions::delete_sqlite_type"],
    )
    def test_delete_inner_sqlite_type(self, sqlite3_conn: sqlite3.Connection, inner_model: models.TestInnerSqliteType) -> None:
        queries.delete_one_test_inner_sqlite_type(conn=sqlite3_conn, table_id=inner_model.table_id)

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::insert_type_override",
    )
    def test_insert_type_override(self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride) -> None:
        queries.insert_type_override(conn=sqlite3_conn, id_=override_model.id_, text_test=override_model.text_test)

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_one_type_override",
        depends=["Sqlite3TestDataclassFunctions::insert_type_override"],
    )
    def test_get_one_type_override(self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_one_type_override(conn=sqlite3_conn, id_=override_model.id_)
        assert result is not None
        assert result == override_model

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_one_type_override_none",
        depends=["Sqlite3TestDataclassFunctions::get_one_type_override"],
    )
    def test_get_one_type_override_none(self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_one_type_override(conn=sqlite3_conn, id_=override_model.id_ - 1)
        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_type_override",
        depends=["Sqlite3TestDataclassFunctions::get_one_type_override_none"],
    )
    def test_get_many_type_override(self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_many_type_override(conn=sqlite3_conn, id_=override_model.id_)
        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestTypeOverride)

        assert results[0] == override_model

        results = result()
        assert isinstance(results[0], models.TestTypeOverride)

        assert results[0] == override_model

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_one_text_type_override",
        depends=["Sqlite3TestDataclassFunctions::get_many_type_override"],
    )
    def test_get_one_text_type_override(self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_one_text_type_override(conn=sqlite3_conn, id_=override_model.id_)
        assert result is not None
        assert result == override_model.text_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_one_text_type_override_none",
        depends=["Sqlite3TestDataclassFunctions::get_one_text_type_override"],
    )
    def test_get_one_text_type_override_none(self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_one_text_type_override(conn=sqlite3_conn, id_=override_model.id_ - 1)
        assert result is None

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_many_text_type_override",
        depends=["Sqlite3TestDataclassFunctions::get_one_text_type_override_none"],
    )
    def test_get_many_text_type_override(self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_many_text_type_override(conn=sqlite3_conn, id_=override_model.id_)
        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], UserString)

        assert results[0] == override_model.text_test

        results = result()
        assert isinstance(results[0], UserString)

        assert results[0] == override_model.text_test

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::delete_type_override",
        depends=["Sqlite3TestDataclassFunctions::get_many_text_type_override"],
    )
    def test_delete_type_override(self, sqlite3_conn: sqlite3.Connection, override_model: models.TestTypeOverride) -> None:
        queries.delete_type_override(conn=sqlite3_conn, id_=override_model.id_)

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::insert_override_conversion")
    def test_insert_override_conversion(self, sqlite3_conn: sqlite3.Connection) -> None:
        # The overridden price parameter is a plain float; the generated code
        # converts it back to decimal.Decimal, which the registered adapter
        # then serializes.
        queries_override_adapter.insert_override_conversion(
            conn=sqlite3_conn,
            id_=434343,
            price=OVERRIDE_PRICE,
            happened_at=OVERRIDE_HAPPENED_AT,
        )

    @pytest.mark.dependency(
        name="Sqlite3TestDataclassFunctions::get_override_price",
        depends=["Sqlite3TestDataclassFunctions::insert_override_conversion"],
    )
    def test_get_override_price(self, sqlite3_conn: sqlite3.Connection) -> None:
        price = queries_override_adapter.get_override_price(conn=sqlite3_conn, id_=434343)
        assert price is not None
        assert isinstance(price, float)
        assert price == OVERRIDE_PRICE

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::get_override_price"])
    def test_get_override_price_not_found(self, sqlite3_conn: sqlite3.Connection) -> None:
        assert queries_override_adapter.get_override_price(conn=sqlite3_conn, id_=434342) is None

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::get_override_price"])
    def test_get_override_happened_at(self, sqlite3_conn: sqlite3.Connection) -> None:
        happened_at = queries_override_converter.get_override_happened_at(conn=sqlite3_conn, id_=434343)
        assert happened_at is not None
        assert isinstance(happened_at, datetime.datetime)
        assert happened_at == OVERRIDE_HAPPENED_AT

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::get_override_price"])
    def test_get_override_happened_at_not_found(self, sqlite3_conn: sqlite3.Connection) -> None:
        assert queries_override_converter.get_override_happened_at(conn=sqlite3_conn, id_=434342) is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::insert_case_row")
    def test_insert_case_row(self, sqlite3_conn: sqlite3.Connection) -> None:
        # The schema declares the columns as DATETIME and decimal(10,2); both
        # must round-trip through the registered adapters and converters.
        queries_case.insert_case_row(conn=sqlite3_conn, id_=515151, upper_dt=CASE_DT, prec_dec=CASE_DEC)

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_case_row"])
    def test_get_case_row(self, sqlite3_conn: sqlite3.Connection) -> None:
        row = queries_case.get_case_row(conn=sqlite3_conn, id_=515151)
        assert row is not None
        assert isinstance(row.upper_dt, datetime.datetime)
        assert row.upper_dt == CASE_DT
        assert isinstance(row.prec_dec, decimal.Decimal)
        assert row.prec_dec == CASE_DEC

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_case_row"])
    def test_get_case_row_not_found(self, sqlite3_conn: sqlite3.Connection) -> None:
        assert queries_case.get_case_row(conn=sqlite3_conn, id_=515150) is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::insert_reserved_arg")
    def test_insert_reserved_arg(self, sqlite3_conn: sqlite3.Connection) -> None:
        # The column is literally named "conn"; the generated parameter must
        # be deduplicated against the implicit connection argument.
        queries_case.insert_reserved_arg(conn=sqlite3_conn, id_=RESERVED_ARG_ID, conn_2="reserved-arg-value")

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_reserved_arg"])
    def test_get_reserved_arg(self, sqlite3_conn: sqlite3.Connection) -> None:
        found_id = queries_case.get_reserved_arg(conn=sqlite3_conn, conn_2="reserved-arg-value")
        assert found_id == RESERVED_ARG_ID

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_reserved_arg"])
    def test_get_reserved_arg_not_found(self, sqlite3_conn: sqlite3.Connection) -> None:
        assert queries_case.get_reserved_arg(conn=sqlite3_conn, conn_2="missing-reserved-arg-value") is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::insert_unknown_override")
    def test_insert_unknown_override(self, sqlite3_conn: sqlite3.Connection) -> None:
        # Overridden unknown SQL type (JULIANDAY): the value must pass
        # through unconverted instead of being wrapped in typing.Any(...).
        queries_unknown_override.insert_unknown_override(conn=sqlite3_conn, id_=UNKNOWN_OVERRIDE_ID, happened_at="2460500.5")

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_unknown_override"])
    def test_get_unknown_override(self, sqlite3_conn: sqlite3.Connection) -> None:
        happened_at = queries_unknown_override.get_unknown_override(conn=sqlite3_conn, id_=UNKNOWN_OVERRIDE_ID)
        assert happened_at == "2460500.5"

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_unknown_override"])
    def test_get_unknown_override_not_found(self, sqlite3_conn: sqlite3.Connection) -> None:
        assert queries_unknown_override.get_unknown_override(conn=sqlite3_conn, id_=UNKNOWN_OVERRIDE_ID - 1) is None

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_unknown_override"])
    def test_get_unknown_override_null_value(self, sqlite3_conn: sqlite3.Connection) -> None:
        queries_unknown_override.insert_unknown_override(conn=sqlite3_conn, id_=UNKNOWN_OVERRIDE_ID + 1, happened_at=None)
        assert queries_unknown_override.get_unknown_override(conn=sqlite3_conn, id_=UNKNOWN_OVERRIDE_ID + 1) is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::insert_any_param")
    def test_insert_any_param(self, sqlite3_conn: sqlite3.Connection) -> None:
        # The override maps an unknown SQL type to a type the driver cannot
        # encode on its own, so the caller registers the adapter.
        sqlite3.register_adapter(pathlib.PurePosixPath, str)
        queries_any_param.insert_any_param(conn=sqlite3_conn, id_=ANY_PARAM_ID, tag=pathlib.PurePosixPath("a/b"))

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_any_param"])
    def test_list_any_param_ids(self, sqlite3_conn: sqlite3.Connection) -> None:
        # Passed through unconverted, so PurePosixPath must be a valid
        # QueryResults argument type.
        results = queries_any_param.list_any_param_ids(conn=sqlite3_conn, tag=pathlib.PurePosixPath("a/b"))
        assert list(results()) == [ANY_PARAM_ID]

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_any_param"])
    def test_iterate_any_param_ids(self, sqlite3_conn: sqlite3.Connection) -> None:
        seen = list(queries_any_param.list_any_param_ids(conn=sqlite3_conn, tag=pathlib.PurePosixPath("a/b")))
        assert seen == [ANY_PARAM_ID]

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::insert_slice_rows")
    def test_insert_slice_rows(self, sqlite3_conn: sqlite3.Connection) -> None:
        for offset, name in enumerate(("a", "b", "c", "b")):
            queries_slice.insert_slice_row(conn=sqlite3_conn, id_=SLICE_ID_BASE + offset, name=name)

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_slice_rows", depends=["Sqlite3TestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_rows(self, sqlite3_conn: sqlite3.Connection) -> None:
        result = queries_slice.get_slice_rows(conn=sqlite3_conn, ids=[SLICE_ID_BASE, SLICE_ID_BASE + 2])
        assert isinstance(result, queries_slice.QueryResults)
        rows = result()
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a"),
            models.TestSlice(id_=SLICE_ID_BASE + 2, name="c"),
        ]
        assert list(result) == rows

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_rows_empty_slice(self, sqlite3_conn: sqlite3.Connection) -> None:
        # An empty sequence expands the placeholder to NULL: IN (NULL)
        # matches no rows instead of raising.
        assert queries_slice.get_slice_rows(conn=sqlite3_conn, ids=[])() == []

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::get_slice_row_filtered", depends=["Sqlite3TestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_row_filtered(self, sqlite3_conn: sqlite3.Connection) -> None:
        # Plain params surround the slice, so this proves the flattened
        # argument tuple binds in SQL text order.
        row = queries_slice.get_slice_row_filtered(
            conn=sqlite3_conn,
            name="b",
            ids=[SLICE_ID_BASE + 1, SLICE_ID_BASE + 3],
            id_=SLICE_ID_BASE + 1,
        )
        assert row == models.TestSlice(id_=SLICE_ID_BASE + 3, name="b")

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_row_filtered_not_found(self, sqlite3_conn: sqlite3.Connection) -> None:
        assert queries_slice.get_slice_row_filtered(conn=sqlite3_conn, name="a", ids=[SLICE_ID_BASE], id_=SLICE_ID_BASE) is None

    @pytest.mark.dependency(name="Sqlite3TestDataclassFunctions::count_slice_rows_two_slices", depends=["Sqlite3TestDataclassFunctions::insert_slice_rows"])
    def test_count_slice_rows_two_slices(self, sqlite3_conn: sqlite3.Connection) -> None:
        count = queries_slice.count_slice_rows(conn=sqlite3_conn, ids=[SLICE_ID_BASE], names=["b"])
        assert count == SLICE_NAME_MATCH_COUNT
        assert queries_slice.count_slice_rows(conn=sqlite3_conn, ids=[], names=[]) == 0

    @pytest.mark.dependency(depends=["Sqlite3TestDataclassFunctions::get_slice_rows", "Sqlite3TestDataclassFunctions::get_slice_row_filtered", "Sqlite3TestDataclassFunctions::count_slice_rows_two_slices"])
    def test_delete_slice_rows(self, sqlite3_conn: sqlite3.Connection) -> None:
        assert queries_slice.delete_slice_rows(conn=sqlite3_conn, ids=[]) == 0
        deleted = queries_slice.delete_slice_rows(conn=sqlite3_conn, ids=[SLICE_ID_BASE + offset for offset in range(SLICE_ROW_COUNT)])
        assert deleted == SLICE_ROW_COUNT
