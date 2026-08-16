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
import random
from collections import UserString

import pytest
import turso

from test.driver_turso_sync.dataclass.functions import models
from test.driver_turso_sync.dataclass.functions import queries
from test.driver_turso_sync.dataclass.functions import queries_case
from test.driver_turso_sync.dataclass.functions import queries_named_slice
from test.driver_turso_sync.dataclass.functions import queries_override_adapter
from test.driver_turso_sync.dataclass.functions import queries_override_converter
from test.driver_turso_sync.dataclass.functions import queries_slice
from test.driver_turso_sync.dataclass.functions import queries_unknown_override

OVERRIDE_PRICE = 12.5
OVERRIDE_HAPPENED_AT = datetime.datetime(2026, 7, 19, 12, 30)
CASE_DT = datetime.datetime(2026, 7, 19, 8, 15)
CASE_DEC = decimal.Decimal("12.34")
RESERVED_ARG_ID = 525252
UNKNOWN_OVERRIDE_ID = 545454
SLICE_ID_BASE = 585858
SLICE_ROW_COUNT = 4


class TestTursoSyncDataclassFunctions:
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
            nvarchar_test="Ola mundo",
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

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::insert")
    def test_insert(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        queries.insert_one_sqlite_type(
            conn=turso_sync_conn,
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

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::inner_insert", depends=["TursoSyncTestDataclassFunctions::insert"])
    def test_inner_insert(
        self,
        turso_sync_conn: turso.Connection,
        inner_model: models.TestInnerSqliteType,
    ) -> None:
        queries.insert_one_inner_sqlite_type(
            conn=turso_sync_conn,
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

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_one", depends=["TursoSyncTestDataclassFunctions::inner_insert"])
    def test_get_one(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_sqlite_type(conn=turso_sync_conn, id_=model.id_)

        assert result is not None

        assert isinstance(result, models.TestSqliteType)

        assert result == model

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_one_none", depends=["TursoSyncTestDataclassFunctions::get_one"])
    def test_get_one_none(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.get_one_sqlite_type(conn=turso_sync_conn, id_=0)

        assert result is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_one_inner", depends=["TursoSyncTestDataclassFunctions::get_one_none"])
    def test_get_one_inner(
        self,
        turso_sync_conn: turso.Connection,
        inner_model: models.TestInnerSqliteType,
    ) -> None:
        result = queries.get_one_inner_sqlite_type(conn=turso_sync_conn, table_id=inner_model.table_id)

        assert result is not None

        assert isinstance(result, models.TestInnerSqliteType)
        assert result == inner_model

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_one_inner_none",
        depends=["TursoSyncTestDataclassFunctions::get_one_inner"],
    )
    def test_get_one_inner_none(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.get_one_inner_sqlite_type(conn=turso_sync_conn, table_id=0)

        assert result is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_date", depends=["TursoSyncTestDataclassFunctions::get_one_inner_none"])
    def test_get_date(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_date(conn=turso_sync_conn, id_=model.id_, date_test=model.date_test)

        assert result is not None

        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_date_none", depends=["TursoSyncTestDataclassFunctions::get_date"])
    def test_get_date_none(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.get_one_date(conn=turso_sync_conn, id_=0, date_test=datetime.date.today())

        assert result is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_datetime", depends=["TursoSyncTestDataclassFunctions::get_date_none"])
    def test_get_datetime(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_datetime(conn=turso_sync_conn, id_=model.id_, datetime_test=model.datetime_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_datetime_none", depends=["TursoSyncTestDataclassFunctions::get_datetime"])
    def test_get_datetime_none(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.get_one_datetime(conn=turso_sync_conn, id_=0, datetime_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_timestamp",
        depends=["TursoSyncTestDataclassFunctions::get_datetime_none"],
    )
    def test_get_timestamp(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_timestamp(conn=turso_sync_conn, id_=model.id_, timestamp_test=model.timestamp_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.timestamp_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_timestamp_none",
        depends=["TursoSyncTestDataclassFunctions::get_timestamp"],
    )
    def test_get_timestamp_none(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.get_one_timestamp(conn=turso_sync_conn, id_=0, timestamp_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_bool", depends=["TursoSyncTestDataclassFunctions::get_timestamp_none"])
    def test_get_bool(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_bool(conn=turso_sync_conn, id_=model.id_, bool_test=model.bool_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.bool_test

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_bool_none", depends=["TursoSyncTestDataclassFunctions::get_bool"])
    def test_get_bool_none(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.get_one_bool(conn=turso_sync_conn, id_=0, bool_test=False)

        assert result is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_boolean", depends=["TursoSyncTestDataclassFunctions::get_bool_none"])
    def test_get_boolean(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_boolean(conn=turso_sync_conn, id_=model.id_, boolean_test=model.boolean_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.boolean_test

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_boolean_none", depends=["TursoSyncTestDataclassFunctions::get_boolean"])
    def test_get_boolean_none(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.get_one_boolean(conn=turso_sync_conn, id_=0, boolean_test=True)

        assert result is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_decimal", depends=["TursoSyncTestDataclassFunctions::get_boolean_none"])
    def test_get_decimal(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_decimal(conn=turso_sync_conn, id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None

        assert isinstance(result, decimal.Decimal)
        assert result == model.decimal_test

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_decimal_none", depends=["TursoSyncTestDataclassFunctions::get_decimal"])
    def test_get_decimal_none(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.get_one_decimal(conn=turso_sync_conn, id_=0, decimal_test=decimal.Decimal("0.1"))

        assert result is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_blob", depends=["TursoSyncTestDataclassFunctions::get_decimal_none"])
    def test_get_blob(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.get_one_blob(conn=turso_sync_conn, id_=model.id_, blob_test=model.blob_test)

        assert result is not None

        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_blob_none", depends=["TursoSyncTestDataclassFunctions::get_blob"])
    def test_get_blob_none(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.get_one_blob(conn=turso_sync_conn, id_=0, blob_test=memoryview(b"test"))

        assert result is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_many", depends=["TursoSyncTestDataclassFunctions::get_blob_none"])
    def test_get_many(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_sqlite_type(conn=turso_sync_conn, id_=model.id_)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestSqliteType)

        assert results[0] == model
        results = result()
        assert isinstance(results[0], models.TestSqliteType)

        assert results[0] == model

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_many_iter", depends=["TursoSyncTestDataclassFunctions::get_many"])
    def test_get_many_iter(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_sqlite_type(conn=turso_sync_conn, id_=model.id_):
            assert result is not None
            assert isinstance(result, models.TestSqliteType)

            assert result == model

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_many_inner", depends=["TursoSyncTestDataclassFunctions::get_many_iter"])
    def test_get_many_inner(self, turso_sync_conn: turso.Connection, inner_model: models.TestInnerSqliteType) -> None:
        result = queries.get_many_inner_sqlite_type(conn=turso_sync_conn, table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestInnerSqliteType)

        assert results[0] == inner_model

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_inner_iter",
        depends=["TursoSyncTestDataclassFunctions::get_many_inner"],
    )
    def test_get_many_inner_iter(self, turso_sync_conn: turso.Connection, inner_model: models.TestInnerSqliteType) -> None:
        for result in queries.get_many_inner_sqlite_type(conn=turso_sync_conn, table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerSqliteType)

            assert result == inner_model

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_nullable_inner",
        depends=["TursoSyncTestDataclassFunctions::get_many_inner_iter"],
    )
    def test_get_many_nullable_inner(self, turso_sync_conn: turso.Connection, inner_model: models.TestInnerSqliteType) -> None:
        result = queries.get_many_nullable_inner_sqlite_type(conn=turso_sync_conn, table_id=inner_model.table_id, int_test=inner_model.int_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestInnerSqliteType)

        assert results[0] == inner_model

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_nullable_inner_iter",
        depends=["TursoSyncTestDataclassFunctions::get_many_nullable_inner"],
    )
    def test_get_many_nullable_inner_iter(self, turso_sync_conn: turso.Connection, inner_model: models.TestInnerSqliteType) -> None:
        for result in queries.get_many_nullable_inner_sqlite_type(conn=turso_sync_conn, table_id=inner_model.table_id, int_test=inner_model.int_test):
            assert result is not None
            assert isinstance(result, models.TestInnerSqliteType)

            assert result == inner_model

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_date",
        depends=["TursoSyncTestDataclassFunctions::get_many_nullable_inner_iter"],
    )
    def test_get_many_date(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_date(conn=turso_sync_conn, id_=model.id_, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.date)

        assert results[0] == model.date_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_date_iter",
        depends=["TursoSyncTestDataclassFunctions::get_many_date"],
    )
    def test_get_many_date_iter(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_date(conn=turso_sync_conn, id_=model.id_, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)

            assert result == model.date_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_datetime",
        depends=["TursoSyncTestDataclassFunctions::get_many_date_iter"],
    )
    def test_get_many_datetime(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_datetime(conn=turso_sync_conn, id_=model.id_, datetime_test=model.datetime_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.datetime)

        assert results[0] == model.datetime_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_datetime_iter",
        depends=["TursoSyncTestDataclassFunctions::get_many_datetime"],
    )
    def test_get_many_datetime_iter(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_datetime(conn=turso_sync_conn, id_=model.id_, datetime_test=model.datetime_test):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.datetime_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_timestamp",
        depends=["TursoSyncTestDataclassFunctions::get_many_datetime_iter"],
    )
    def test_get_many_timestamp(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_timestamp(conn=turso_sync_conn, id_=model.id_, timestamp_test=model.timestamp_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.datetime)

        assert results[0] == model.timestamp_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_timestamp_iter",
        depends=["TursoSyncTestDataclassFunctions::get_many_timestamp"],
    )
    def test_get_many_timestamp_iter(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_timestamp(conn=turso_sync_conn, id_=model.id_, timestamp_test=model.timestamp_test):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.timestamp_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_bool",
        depends=["TursoSyncTestDataclassFunctions::get_many_timestamp_iter"],
    )
    def test_get_many_bool(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_bool(conn=turso_sync_conn, id_=model.id_, bool_test=model.bool_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], bool)

        assert results[0] == model.bool_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_bool_iter",
        depends=["TursoSyncTestDataclassFunctions::get_many_bool"],
    )
    def test_get_many_bool_iter(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_bool(conn=turso_sync_conn, id_=model.id_, bool_test=model.bool_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.bool_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_boolean",
        depends=["TursoSyncTestDataclassFunctions::get_many_bool_iter"],
    )
    def test_get_many_boolean(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_boolean(conn=turso_sync_conn, id_=model.id_, boolean_test=model.boolean_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], bool)

        assert results[0] == model.boolean_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_boolean_iter",
        depends=["TursoSyncTestDataclassFunctions::get_many_boolean"],
    )
    def test_get_many_boolean_iter(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_boolean(conn=turso_sync_conn, id_=model.id_, boolean_test=model.boolean_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.boolean_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_decimal",
        depends=["TursoSyncTestDataclassFunctions::get_many_boolean_iter"],
    )
    def test_get_many_decimal(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_decimal(conn=turso_sync_conn, id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], decimal.Decimal)

        assert results[0] == model.decimal_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_decimal_iter",
        depends=["TursoSyncTestDataclassFunctions::get_many_decimal"],
    )
    def test_get_many_decimal_iter(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_decimal(conn=turso_sync_conn, id_=model.id_, decimal_test=model.decimal_test):
            assert result is not None
            assert isinstance(result, decimal.Decimal)

            assert result == model.decimal_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_blob",
        depends=["TursoSyncTestDataclassFunctions::get_many_decimal_iter"],
    )
    def test_get_many_blob(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        result = queries.get_many_blob(conn=turso_sync_conn, id_=model.id_, blob_test=model.blob_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], memoryview)

        assert results[0] == model.blob_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_blob_iter",
        depends=["TursoSyncTestDataclassFunctions::get_many_blob"],
    )
    def test_get_many_blob_iter(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        for result in queries.get_many_blob(conn=turso_sync_conn, id_=model.id_, blob_test=model.blob_test):
            assert result is not None
            assert isinstance(result, memoryview)

            assert result == model.blob_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::insert_result",
        depends=["TursoSyncTestDataclassFunctions::get_many_blob_iter"],
    )
    def test_insert_result(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.insert_result_one_sqlite_type(
            conn=turso_sync_conn,
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
        assert isinstance(result, turso.Cursor)

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::update_result", depends=["TursoSyncTestDataclassFunctions::insert_result"])
    def test_update_result(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.update_result_one_sqlite_type(conn=turso_sync_conn, id_=model.id_ + 1)
        assert isinstance(result, turso.Cursor)

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::delete_result", depends=["TursoSyncTestDataclassFunctions::update_result"])
    def test_delete_result(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.delete_result_one_sqlite_type(conn=turso_sync_conn, id_=model.id_ + 1)
        assert isinstance(result, turso.Cursor)

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::insert_rows", depends=["TursoSyncTestDataclassFunctions::delete_result"])
    def test_insert_rows(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.insert_rows_one_sqlite_type(
            conn=turso_sync_conn,
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

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::update_rows", depends=["TursoSyncTestDataclassFunctions::insert_rows"])
    def test_update_rows(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.update_rows_one_sqlite_type(conn=turso_sync_conn, id_=model.id_ + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::delete_rows", depends=["TursoSyncTestDataclassFunctions::update_rows"])
    def test_delete_rows(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.delete_rows_one_sqlite_type(conn=turso_sync_conn, id_=model.id_ + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::create_table_rows", depends=["TursoSyncTestDataclassFunctions::delete_rows"])
    def test_create_table_rows(
        self,
        turso_sync_conn: turso.Connection,
    ) -> None:
        result = queries.create_rows_table(conn=turso_sync_conn)
        assert isinstance(result, int)
        turso_sync_conn.execute("DROP TABLE test_create_rows_table;")
        turso_sync_conn.commit()
        assert result == 0

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::insert_last_id",
        depends=["TursoSyncTestDataclassFunctions::create_table_rows"],
    )
    def test_insert_last_id(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.insert_last_id_one_sqlite_type(
            conn=turso_sync_conn,
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

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::update_last_id", depends=["TursoSyncTestDataclassFunctions::insert_last_id"])
    def test_update_last_id(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.update_last_id_one_sqlite_type(conn=turso_sync_conn, id_=model.id_ + 3)
        # Unlike sqlite3, turso's cursor.lastrowid only reflects the most
        # recent INSERT on that statement; it is None after an UPDATE
        # (empirically verified) rather than the connection's last insert id.
        assert result is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::delete_last_id", depends=["TursoSyncTestDataclassFunctions::update_last_id"])
    def test_delete_last_id(
        self,
        turso_sync_conn: turso.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = queries.delete_last_id_one_sqlite_type(conn=turso_sync_conn, id_=model.id_ + 3)
        # Same turso lastrowid semantics as test_update_last_id: None after a
        # non-INSERT statement.
        assert result is None

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::delete_sqlite_type",
        depends=["TursoSyncTestDataclassFunctions::delete_last_id"],
    )
    def test_delete_sqlite_type(self, turso_sync_conn: turso.Connection, model: models.TestSqliteType) -> None:
        queries.delete_one_sqlite_type(conn=turso_sync_conn, id_=model.id_)

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::delete_inner_sqlite_type",
        depends=["TursoSyncTestDataclassFunctions::delete_sqlite_type"],
    )
    def test_delete_inner_sqlite_type(self, turso_sync_conn: turso.Connection, inner_model: models.TestInnerSqliteType) -> None:
        queries.delete_one_test_inner_sqlite_type(conn=turso_sync_conn, table_id=inner_model.table_id)

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::insert_type_override",
    )
    def test_insert_type_override(self, turso_sync_conn: turso.Connection, override_model: models.TestTypeOverride) -> None:
        queries.insert_type_override(conn=turso_sync_conn, id_=override_model.id_, text_test=override_model.text_test)

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_one_type_override",
        depends=["TursoSyncTestDataclassFunctions::insert_type_override"],
    )
    def test_get_one_type_override(self, turso_sync_conn: turso.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_one_type_override(conn=turso_sync_conn, id_=override_model.id_)
        assert result is not None
        assert result == override_model

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_one_type_override_none",
        depends=["TursoSyncTestDataclassFunctions::get_one_type_override"],
    )
    def test_get_one_type_override_none(self, turso_sync_conn: turso.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_one_type_override(conn=turso_sync_conn, id_=override_model.id_ - 1)
        assert result is None

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_type_override",
        depends=["TursoSyncTestDataclassFunctions::get_one_type_override_none"],
    )
    def test_get_many_type_override(self, turso_sync_conn: turso.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_many_type_override(conn=turso_sync_conn, id_=override_model.id_)
        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestTypeOverride)

        assert results[0] == override_model

        results = result()
        assert isinstance(results[0], models.TestTypeOverride)

        assert results[0] == override_model

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_one_text_type_override",
        depends=["TursoSyncTestDataclassFunctions::get_many_type_override"],
    )
    def test_get_one_text_type_override(self, turso_sync_conn: turso.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_one_text_type_override(conn=turso_sync_conn, id_=override_model.id_)
        assert result is not None
        assert result == override_model.text_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_one_text_type_override_none",
        depends=["TursoSyncTestDataclassFunctions::get_one_text_type_override"],
    )
    def test_get_one_text_type_override_none(self, turso_sync_conn: turso.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_one_text_type_override(conn=turso_sync_conn, id_=override_model.id_ - 1)
        assert result is None

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_many_text_type_override",
        depends=["TursoSyncTestDataclassFunctions::get_one_text_type_override_none"],
    )
    def test_get_many_text_type_override(self, turso_sync_conn: turso.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_many_text_type_override(conn=turso_sync_conn, id_=override_model.id_)
        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], UserString)

        assert results[0] == override_model.text_test

        results = result()
        assert isinstance(results[0], UserString)

        assert results[0] == override_model.text_test

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::delete_type_override",
        depends=["TursoSyncTestDataclassFunctions::get_many_text_type_override"],
    )
    def test_delete_type_override(self, turso_sync_conn: turso.Connection, override_model: models.TestTypeOverride) -> None:
        queries.delete_type_override(conn=turso_sync_conn, id_=override_model.id_)

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::insert_override_conversion")
    def test_insert_override_conversion(self, turso_sync_conn: turso.Connection) -> None:
        # The overridden price parameter is a plain float; the generated code
        # converts it back to decimal.Decimal and serializes it inline via
        # str() before binding (turso has no adapter registration).
        queries_override_adapter.insert_override_conversion(
            conn=turso_sync_conn,
            id_=434343,
            price=OVERRIDE_PRICE,
            happened_at=OVERRIDE_HAPPENED_AT,
        )

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::get_override_price",
        depends=["TursoSyncTestDataclassFunctions::insert_override_conversion"],
    )
    def test_get_override_price(self, turso_sync_conn: turso.Connection) -> None:
        price = queries_override_adapter.get_override_price(conn=turso_sync_conn, id_=434343)
        assert price is not None
        assert isinstance(price, float)
        assert price == OVERRIDE_PRICE

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::get_override_price"])
    def test_get_override_price_not_found(self, turso_sync_conn: turso.Connection) -> None:
        assert queries_override_adapter.get_override_price(conn=turso_sync_conn, id_=434342) is None

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::get_override_price"])
    def test_get_override_happened_at(self, turso_sync_conn: turso.Connection) -> None:
        happened_at = queries_override_converter.get_override_happened_at(conn=turso_sync_conn, id_=434343)
        assert happened_at is not None
        assert isinstance(happened_at, datetime.datetime)
        assert happened_at == OVERRIDE_HAPPENED_AT

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::get_override_price"])
    def test_get_override_happened_at_not_found(self, turso_sync_conn: turso.Connection) -> None:
        assert queries_override_converter.get_override_happened_at(conn=turso_sync_conn, id_=434342) is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::insert_case_row")
    def test_insert_case_row(self, turso_sync_conn: turso.Connection) -> None:
        # The schema declares the columns as DATETIME and decimal(10,2); both
        # must round-trip through the generated module's inline conversion
        # (isoformat/fromisoformat and str/Decimal, no registration involved).
        queries_case.insert_case_row(conn=turso_sync_conn, id_=515151, upper_dt=CASE_DT, prec_dec=CASE_DEC)

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::insert_case_row"])
    def test_get_case_row(self, turso_sync_conn: turso.Connection) -> None:
        row = queries_case.get_case_row(conn=turso_sync_conn, id_=515151)
        assert row is not None
        assert isinstance(row.upper_dt, datetime.datetime)
        assert row.upper_dt == CASE_DT
        assert isinstance(row.prec_dec, decimal.Decimal)
        assert row.prec_dec == CASE_DEC

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::insert_case_row"])
    def test_get_case_row_not_found(self, turso_sync_conn: turso.Connection) -> None:
        assert queries_case.get_case_row(conn=turso_sync_conn, id_=515150) is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::insert_reserved_arg")
    def test_insert_reserved_arg(self, turso_sync_conn: turso.Connection) -> None:
        # The column is literally named "conn"; the generated parameter must
        # be deduplicated against the implicit connection argument.
        queries_case.insert_reserved_arg(conn=turso_sync_conn, id_=RESERVED_ARG_ID, conn_2="reserved-arg-value")

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::insert_reserved_arg"])
    def test_get_reserved_arg(self, turso_sync_conn: turso.Connection) -> None:
        found_id = queries_case.get_reserved_arg(conn=turso_sync_conn, conn_2="reserved-arg-value")
        assert found_id == RESERVED_ARG_ID

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::insert_reserved_arg"])
    def test_get_reserved_arg_not_found(self, turso_sync_conn: turso.Connection) -> None:
        assert queries_case.get_reserved_arg(conn=turso_sync_conn, conn_2="missing-reserved-arg-value") is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::insert_unknown_override")
    def test_insert_unknown_override(self, turso_sync_conn: turso.Connection) -> None:
        # Overridden unknown SQL type (JULIANDAY): the value must pass
        # through unconverted instead of being wrapped in typing.Any(...).
        queries_unknown_override.insert_unknown_override(conn=turso_sync_conn, id_=UNKNOWN_OVERRIDE_ID, happened_at="2460500.5")

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::insert_unknown_override"])
    def test_get_unknown_override(self, turso_sync_conn: turso.Connection) -> None:
        happened_at = queries_unknown_override.get_unknown_override(conn=turso_sync_conn, id_=UNKNOWN_OVERRIDE_ID)
        assert happened_at == "2460500.5"

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::insert_unknown_override"])
    def test_get_unknown_override_not_found(self, turso_sync_conn: turso.Connection) -> None:
        assert queries_unknown_override.get_unknown_override(conn=turso_sync_conn, id_=UNKNOWN_OVERRIDE_ID - 1) is None

    @pytest.mark.dependency(depends=["TursoSyncTestDataclassFunctions::insert_unknown_override"])
    def test_get_unknown_override_null_value(self, turso_sync_conn: turso.Connection) -> None:
        queries_unknown_override.insert_unknown_override(conn=turso_sync_conn, id_=UNKNOWN_OVERRIDE_ID + 1, happened_at=None)
        assert queries_unknown_override.get_unknown_override(conn=turso_sync_conn, id_=UNKNOWN_OVERRIDE_ID + 1) is None

    # NOTE: the any_param scenario (queries_any_param.py, sqlc db_type override
    # TAGTYPE -> pathlib.PurePosixPath) is intentionally not ported here. On
    # sqlite3 it works because the caller registers a global adapter
    # (sqlite3.register_adapter) that converts PurePosixPath to str before
    # binding. pyturso has no adapter-registration API at all (empirically
    # verified: turso.lib.DatabaseError: "unexpected parameter value, only
    # None, numbers, strings and bytes are supported" when a PurePosixPath is
    # bound directly, and PEP 246 __conform__ is not honored either), and the
    # generated turso module passes the override value through unconverted
    # just like sqlite3 does. There is no substitute mechanism to make this
    # scenario work against turso, so the test is dropped rather than adapted.

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::insert_slice_rows")
    def test_insert_slice_rows(self, turso_sync_conn: turso.Connection) -> None:
        for offset, (name, note) in enumerate((("a", "x"), ("b", "y"), ("c", None), ("b", "y"))):
            queries_slice.insert_slice_row(conn=turso_sync_conn, id_=SLICE_ID_BASE + offset, name=name, note=note)

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_slice_rows", depends=["TursoSyncTestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_rows(self, turso_sync_conn: turso.Connection) -> None:
        result = queries_slice.get_slice_rows(conn=turso_sync_conn, ids=[SLICE_ID_BASE, SLICE_ID_BASE + 2])
        assert isinstance(result, queries_slice.QueryResults)
        rows = result()
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a", note="x"),
            models.TestSlice(id_=SLICE_ID_BASE + 2, name="c", note=None),
        ]
        assert list(result) == rows

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_slice_rows_empty_slice", depends=["TursoSyncTestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_rows_empty_slice(self, turso_sync_conn: turso.Connection) -> None:
        # An empty sequence expands the placeholder to NULL: IN (NULL)
        # matches no rows instead of raising.
        assert queries_slice.get_slice_rows(conn=turso_sync_conn, ids=[])() == []

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_slice_row_filtered", depends=["TursoSyncTestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_row_filtered(self, turso_sync_conn: turso.Connection) -> None:
        # Plain params surround the slice, so this proves the flattened
        # argument tuple binds in SQL text order.
        row = queries_slice.get_slice_row_filtered(
            conn=turso_sync_conn,
            name="b",
            ids=[SLICE_ID_BASE + 1, SLICE_ID_BASE + 3],
            id_=SLICE_ID_BASE + 1,
        )
        assert row == models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y")

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_slice_row_filtered_not_found", depends=["TursoSyncTestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_row_filtered_not_found(self, turso_sync_conn: turso.Connection) -> None:
        assert queries_slice.get_slice_row_filtered(conn=turso_sync_conn, name="a", ids=[SLICE_ID_BASE], id_=SLICE_ID_BASE) is None

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_slice_rows_by_notes", depends=["TursoSyncTestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_rows_by_notes(self, turso_sync_conn: turso.Connection) -> None:
        # The slice targets a nullable column; the parameter is still a plain
        # Sequence, and rows whose note is NULL never match.
        rows = queries_slice.get_slice_rows_by_notes(conn=turso_sync_conn, notes=["y"])()
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE + 1, name="b", note="y"),
            models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y"),
        ]
        assert queries_slice.get_slice_rows_by_notes(conn=turso_sync_conn, notes=[])() == []

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_slice_rows_by_name_or_note", depends=["TursoSyncTestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_rows_by_name_or_note(self, turso_sync_conn: turso.Connection) -> None:
        # The same slice name is used twice, so every marker occurrence is
        # expanded and the sequence is bound once per occurrence.
        rows = queries_slice.get_slice_rows_by_name_or_note(conn=turso_sync_conn, names=["b", "x"])()
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a", note="x"),
            models.TestSlice(id_=SLICE_ID_BASE + 1, name="b", note="y"),
            models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y"),
        ]
        assert queries_slice.get_slice_rows_by_name_or_note(conn=turso_sync_conn, names=[])() == []

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_slice_rows_by_name_or_note_filtered", depends=["TursoSyncTestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_rows_by_name_or_note_filtered(self, turso_sync_conn: turso.Connection) -> None:
        # A plain parameter sits between the two uses of the slice, so this
        # proves the flattened arguments follow SQL text order.
        rows = queries_slice.get_slice_rows_by_name_or_note_filtered(conn=turso_sync_conn, names=["b", "x"], id_=SLICE_ID_BASE + 1)()
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a", note="x"),
            models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y"),
        ]
        assert queries_slice.get_slice_rows_by_name_or_note_filtered(conn=turso_sync_conn, names=[], id_=SLICE_ID_BASE)() == []

    @pytest.mark.dependency(name="TursoSyncTestDataclassFunctions::get_first_slice_name_two_slices", depends=["TursoSyncTestDataclassFunctions::insert_slice_rows"])
    def test_get_first_slice_name_two_slices(self, turso_sync_conn: turso.Connection) -> None:
        name = queries_slice.get_first_slice_name(conn=turso_sync_conn, ids=[SLICE_ID_BASE + 1], names=["a"])
        assert name == "a"
        assert queries_slice.get_first_slice_name(conn=turso_sync_conn, ids=[], names=[]) is None

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::named_slice_rows",
        depends=["TursoSyncTestDataclassFunctions::insert_slice_rows"],
    )
    def test_named_slice_rows(self, turso_sync_conn: turso.Connection) -> None:
        # sqlc numbers the placeholders of a query using a named argument, and
        # those indexes no longer line up once the marker expands. Every
        # length matters: one element used to pass by accident.
        # Rows BASE..BASE+3 are named a, b, c, b - so "b" discriminates and
        # the id list is not simply echoed back.
        for ids, want in (
            ([], []),
            ([SLICE_ID_BASE + 1], [SLICE_ID_BASE + 1]),
            ([SLICE_ID_BASE + 1, SLICE_ID_BASE + 3], [SLICE_ID_BASE + 1, SLICE_ID_BASE + 3]),
            ([SLICE_ID_BASE, SLICE_ID_BASE + 1, SLICE_ID_BASE + 3], [SLICE_ID_BASE + 1, SLICE_ID_BASE + 3]),
        ):
            rows = queries_named_slice.get_named_slice_rows(conn=turso_sync_conn, ids=ids, wanted="b")()
            assert [row.id_ for row in rows] == want
            reused = queries_named_slice.get_named_slice_rows_reused(conn=turso_sync_conn, ids=ids, wanted="b")()
            assert [row.id_ for row in reused] == want
            first = queries_named_slice.get_named_slice_rows_arg_first(conn=turso_sync_conn, wanted="b", ids=ids)()
            assert [row.id_ for row in first] == want
            one = queries_named_slice.get_named_slice_row(conn=turso_sync_conn, ids=ids, wanted="b")
            assert (one.id_ if one is not None else None) == (want[0] if want else None)

    @pytest.mark.dependency(
        name="TursoSyncTestDataclassFunctions::named_slice_rows_iter",
        depends=["TursoSyncTestDataclassFunctions::named_slice_rows"],
    )
    def test_named_slice_rows_iter(self, turso_sync_conn: turso.Connection) -> None:
        ids = [SLICE_ID_BASE + 1, SLICE_ID_BASE + 3]
        rows = list(queries_named_slice.get_named_slice_rows_reused(conn=turso_sync_conn, ids=ids, wanted="b"))
        assert [row.id_ for row in rows] == ids

    @pytest.mark.dependency(
        depends=[
            "TursoSyncTestDataclassFunctions::get_slice_rows",
            "TursoSyncTestDataclassFunctions::get_slice_rows_empty_slice",
            "TursoSyncTestDataclassFunctions::get_slice_row_filtered",
            "TursoSyncTestDataclassFunctions::get_slice_row_filtered_not_found",
            "TursoSyncTestDataclassFunctions::get_slice_rows_by_notes",
            "TursoSyncTestDataclassFunctions::get_slice_rows_by_name_or_note",
            "TursoSyncTestDataclassFunctions::get_slice_rows_by_name_or_note_filtered",
            "TursoSyncTestDataclassFunctions::get_first_slice_name_two_slices",
            "TursoSyncTestDataclassFunctions::named_slice_rows",
            "TursoSyncTestDataclassFunctions::named_slice_rows_iter",
        ]
    )
    def test_delete_slice_rows(self, turso_sync_conn: turso.Connection) -> None:
        assert queries_slice.delete_slice_rows(conn=turso_sync_conn, ids=[]) == 0
        deleted = queries_slice.delete_slice_rows(conn=turso_sync_conn, ids=[SLICE_ID_BASE + offset for offset in range(SLICE_ROW_COUNT)])
        assert deleted == SLICE_ROW_COUNT
