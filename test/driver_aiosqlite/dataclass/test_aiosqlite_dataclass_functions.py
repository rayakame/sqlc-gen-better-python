from __future__ import annotations

import collections.abc
import datetime
import decimal
import json
import random
import aiosqlite

import pytest

from test.driver_aiosqlite.dataclass.functions import models
from test.driver_aiosqlite.dataclass.functions import queries


@pytest.mark.asyncio(loop_scope="session")
class TestDataclassFunctions:
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
            real_test=3.14,
            double_test=2.71828,
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

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="SqliteTestDataclassFunctions::insert")
    async def test_insert(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        await queries.insert_one_sqlite_type(
            conn=aiosqlite_conn,
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

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::inner_insert", depends=["SqliteTestDataclassFunctions::insert"]
    )
    async def test_inner_insert(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        inner_model: models.TestInnerSqliteType,
    ) -> None:
        await queries.insert_one_inner_sqlite_type(
            conn=aiosqlite_conn,
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

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_one", depends=["SqliteTestDataclassFunctions::inner_insert"]
    )
    async def test_get_one(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.get_one_sqlite_type(conn=aiosqlite_conn, id_=model.id)

        assert result is not None

        assert isinstance(result, models.TestSqliteType)

        assert result == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_one_none", depends=["SqliteTestDataclassFunctions::get_one"]
    )
    async def test_get_one_none(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.get_one_sqlite_type(conn=aiosqlite_conn, id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_one_inner", depends=["SqliteTestDataclassFunctions::get_one_none"]
    )
    async def test_get_one_inner(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        inner_model: models.TestInnerSqliteType,
    ) -> None:
        result = await queries.get_one_inner_sqlite_type(conn=aiosqlite_conn, table_id=inner_model.table_id)

        assert result is not None

        assert isinstance(result, models.TestInnerSqliteType)
        assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_one_inner_none", depends=["SqliteTestDataclassFunctions::get_one_inner"]
    )
    async def test_get_one_inner_none(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.get_one_inner_sqlite_type(conn=aiosqlite_conn, table_id=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_date", depends=["SqliteTestDataclassFunctions::get_one_inner_none"]
    )
    async def test_get_date(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.get_one_date(conn=aiosqlite_conn, id_=model.id, date_test=model.date_test)

        assert result is not None

        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_date_none", depends=["SqliteTestDataclassFunctions::get_date"]
    )
    async def test_get_date_none(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.get_one_date(conn=aiosqlite_conn, id_=0, date_test=datetime.date.today())

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_datetime", depends=["SqliteTestDataclassFunctions::get_date_none"]
    )
    async def test_get_datetime(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.get_one_datetime(conn=aiosqlite_conn, id_=model.id, datetime_test=model.datetime_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_datetime_none", depends=["SqliteTestDataclassFunctions::get_datetime"]
    )
    async def test_get_datetime_none(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.get_one_datetime(conn=aiosqlite_conn, id_=0, datetime_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_timestamp", depends=["SqliteTestDataclassFunctions::get_datetime_none"]
    )
    async def test_get_timestamp(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.get_one_timestamp(conn=aiosqlite_conn, id_=model.id, timestamp_test=model.timestamp_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.timestamp_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_timestamp_none", depends=["SqliteTestDataclassFunctions::get_timestamp"]
    )
    async def test_get_timestamp_none(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.get_one_timestamp(conn=aiosqlite_conn, id_=0, timestamp_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_bool", depends=["SqliteTestDataclassFunctions::get_timestamp_none"]
    )
    async def test_get_bool(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.get_one_bool(conn=aiosqlite_conn, id_=model.id, bool_test=model.bool_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.bool_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_bool_none", depends=["SqliteTestDataclassFunctions::get_bool"]
    )
    async def test_get_bool_none(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.get_one_bool(conn=aiosqlite_conn, id_=0, bool_test=False)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_boolean", depends=["SqliteTestDataclassFunctions::get_bool_none"]
    )
    async def test_get_boolean(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.get_one_boolean(conn=aiosqlite_conn, id_=model.id, boolean_test=model.boolean_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.boolean_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_boolean_none", depends=["SqliteTestDataclassFunctions::get_boolean"]
    )
    async def test_get_boolean_none(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.get_one_boolean(conn=aiosqlite_conn, id_=0, boolean_test=True)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_decimal", depends=["SqliteTestDataclassFunctions::get_boolean_none"]
    )
    async def test_get_decimal(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.get_one_decimal(conn=aiosqlite_conn, id_=model.id, decimal_test=model.decimal_test)

        assert result is not None

        assert isinstance(result, decimal.Decimal)
        assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_decimal_none", depends=["SqliteTestDataclassFunctions::get_decimal"]
    )
    async def test_get_decimal_none(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.get_one_decimal(conn=aiosqlite_conn, id_=0, decimal_test=decimal.Decimal("0.1"))

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_blob", depends=["SqliteTestDataclassFunctions::get_decimal_none"]
    )
    async def test_get_blob(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.get_one_blob(conn=aiosqlite_conn, id_=model.id, blob_test=model.blob_test)

        assert result is not None

        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_blob_none", depends=["SqliteTestDataclassFunctions::get_blob"]
    )
    async def test_get_blob_none(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.get_one_blob(conn=aiosqlite_conn, id_=0, blob_test=memoryview(b"test"))

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many", depends=["SqliteTestDataclassFunctions::get_blob_none"]
    )
    async def test_get_many(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        result = await queries.get_many_sqlite_type(conn=aiosqlite_conn, id_=model.id)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], models.TestSqliteType)

        assert result[0] == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_iter", depends=["SqliteTestDataclassFunctions::get_many"]
    )
    async def test_get_many_iter(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        async for result in queries.get_many_sqlite_type(conn=aiosqlite_conn, id_=model.id):
            assert result is not None
            assert isinstance(result, models.TestSqliteType)

            assert result == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_inner", depends=["SqliteTestDataclassFunctions::get_many_iter"]
    )
    async def test_get_many_inner(
        self, aiosqlite_conn: aiosqlite.Connection, inner_model: models.TestInnerSqliteType
    ) -> None:
        result = await queries.get_many_inner_sqlite_type(conn=aiosqlite_conn, table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], models.TestInnerSqliteType)

        assert result[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_inner_iter",
        depends=["SqliteTestDataclassFunctions::get_many_inner"],
    )
    async def test_get_many_inner_iter(
        self, aiosqlite_conn: aiosqlite.Connection, inner_model: models.TestInnerSqliteType
    ) -> None:
        async for result in queries.get_many_inner_sqlite_type(conn=aiosqlite_conn, table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerSqliteType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_date",
        depends=["SqliteTestDataclassFunctions::get_many_inner_iter"],
    )
    async def test_get_many_date(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        result = await queries.get_many_date(conn=aiosqlite_conn, id_=model.id, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], datetime.date)

        assert result[0] == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_date_iter", depends=["SqliteTestDataclassFunctions::get_many_date"]
    )
    async def test_get_many_date_iter(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        async for result in queries.get_many_date(conn=aiosqlite_conn, id_=model.id, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)

            assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_datetime",
        depends=["SqliteTestDataclassFunctions::get_many_date_iter"],
    )
    async def test_get_many_datetime(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        result = await queries.get_many_datetime(conn=aiosqlite_conn, id_=model.id, datetime_test=model.datetime_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], datetime.datetime)

        assert result[0] == model.datetime_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_datetime_iter",
        depends=["SqliteTestDataclassFunctions::get_many_datetime"],
    )
    async def test_get_many_datetime_iter(
        self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType
    ) -> None:
        async for result in queries.get_many_datetime(
            conn=aiosqlite_conn, id_=model.id, datetime_test=model.datetime_test
        ):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.datetime_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_timestamp",
        depends=["SqliteTestDataclassFunctions::get_many_datetime_iter"],
    )
    async def test_get_many_timestamp(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        result = await queries.get_many_timestamp(
            conn=aiosqlite_conn, id_=model.id, timestamp_test=model.timestamp_test
        )

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], datetime.datetime)

        assert result[0] == model.timestamp_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_timestamp_iter",
        depends=["SqliteTestDataclassFunctions::get_many_timestamp"],
    )
    async def test_get_many_timestamp_iter(
        self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType
    ) -> None:
        async for result in queries.get_many_timestamp(
            conn=aiosqlite_conn, id_=model.id, timestamp_test=model.timestamp_test
        ):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.timestamp_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_bool",
        depends=["SqliteTestDataclassFunctions::get_many_timestamp_iter"],
    )
    async def test_get_many_bool(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        result = await queries.get_many_bool(conn=aiosqlite_conn, id_=model.id, bool_test=model.bool_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], bool)

        assert result[0] == model.bool_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_bool_iter", depends=["SqliteTestDataclassFunctions::get_many_bool"]
    )
    async def test_get_many_bool_iter(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        async for result in queries.get_many_bool(conn=aiosqlite_conn, id_=model.id, bool_test=model.bool_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.bool_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_boolean",
        depends=["SqliteTestDataclassFunctions::get_many_bool_iter"],
    )
    async def test_get_many_boolean(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        result = await queries.get_many_boolean(conn=aiosqlite_conn, id_=model.id, boolean_test=model.boolean_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], bool)

        assert result[0] == model.boolean_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_boolean_iter",
        depends=["SqliteTestDataclassFunctions::get_many_boolean"],
    )
    async def test_get_many_boolean_iter(
        self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType
    ) -> None:
        async for result in queries.get_many_boolean(
            conn=aiosqlite_conn, id_=model.id, boolean_test=model.boolean_test
        ):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.boolean_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_decimal",
        depends=["SqliteTestDataclassFunctions::get_many_boolean_iter"],
    )
    async def test_get_many_decimal(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        result = await queries.get_many_decimal(conn=aiosqlite_conn, id_=model.id, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], decimal.Decimal)

        assert result[0] == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_decimal_iter",
        depends=["SqliteTestDataclassFunctions::get_many_decimal"],
    )
    async def test_get_many_decimal_iter(
        self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType
    ) -> None:
        async for result in queries.get_many_decimal(
            conn=aiosqlite_conn, id_=model.id, decimal_test=model.decimal_test
        ):
            assert result is not None
            assert isinstance(result, decimal.Decimal)

            assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_blob",
        depends=["SqliteTestDataclassFunctions::get_many_decimal_iter"],
    )
    async def test_get_many_blob(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        result = await queries.get_many_blob(conn=aiosqlite_conn, id_=model.id, blob_test=model.blob_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], memoryview)

        assert result[0] == model.blob_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::get_many_blob_iter", depends=["SqliteTestDataclassFunctions::get_many_blob"]
    )
    async def test_get_many_blob_iter(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        async for result in queries.get_many_blob(conn=aiosqlite_conn, id_=model.id, blob_test=model.blob_test):
            assert result is not None
            assert isinstance(result, memoryview)

            assert result == model.blob_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::insert_result", depends=["SqliteTestDataclassFunctions::get_many_blob_iter"]
    )
    async def test_insert_result(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.insert_result_one_sqlite_type(
            conn=aiosqlite_conn,
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
        assert isinstance(result, aiosqlite.Cursor)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::update_result", depends=["SqliteTestDataclassFunctions::insert_result"]
    )
    async def test_update_result(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.update_result_one_sqlite_type(conn=aiosqlite_conn, id_=model.id + 1)
        assert isinstance(result, aiosqlite.Cursor)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::delete_result", depends=["SqliteTestDataclassFunctions::update_result"]
    )
    async def test_delete_result(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.delete_result_one_sqlite_type(conn=aiosqlite_conn, id_=model.id + 1)
        assert isinstance(result, aiosqlite.Cursor)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::insert_rows", depends=["SqliteTestDataclassFunctions::delete_result"]
    )
    async def test_insert_rows(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.insert_rows_one_sqlite_type(
            conn=aiosqlite_conn,
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

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::update_rows", depends=["SqliteTestDataclassFunctions::insert_rows"]
    )
    async def test_update_rows(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.update_rows_one_sqlite_type(conn=aiosqlite_conn, id_=model.id + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::delete_rows", depends=["SqliteTestDataclassFunctions::update_rows"]
    )
    async def test_delete_rows(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.delete_rows_one_sqlite_type(conn=aiosqlite_conn, id_=model.id + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::create_table_rows", depends=["SqliteTestDataclassFunctions::delete_rows"]
    )
    async def test_create_table_rows(
        self,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries.create_rows_table(conn=aiosqlite_conn)
        assert isinstance(result, int)
        await aiosqlite_conn.execute("DROP TABLE test_create_rows_table;")
        await aiosqlite_conn.commit()
        assert result == -1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::insert_last_id", depends=["SqliteTestDataclassFunctions::create_table_rows"]
    )
    async def test_insert_last_id(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.insert_last_id_one_sqlite_type(
            conn=aiosqlite_conn,
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

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::update_last_id", depends=["SqliteTestDataclassFunctions::insert_last_id"]
    )
    async def test_update_last_id(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.update_last_id_one_sqlite_type(conn=aiosqlite_conn, id_=model.id + 3)
        assert isinstance(result, int)
        assert result == model.id + 3

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::delete_last_id", depends=["SqliteTestDataclassFunctions::update_last_id"]
    )
    async def test_delete_last_id(
        self,
        aiosqlite_conn: aiosqlite.Connection,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries.delete_last_id_one_sqlite_type(conn=aiosqlite_conn, id_=model.id + 3)
        assert isinstance(result, int)
        assert result == model.id + 3

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::delete_sqlite_type",
        depends=["SqliteTestDataclassFunctions::delete_last_id"],
    )
    async def test_delete_sqlite_type(self, aiosqlite_conn: aiosqlite.Connection, model: models.TestSqliteType) -> None:
        await queries.delete_one_sqlite_type(conn=aiosqlite_conn, id_=model.id)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="SqliteTestDataclassFunctions::delete_inner_sqlite_type",
        depends=["SqliteTestDataclassFunctions::delete_sqlite_type"],
    )
    async def test_delete_inner_sqlite_type(
        self, aiosqlite_conn: aiosqlite.Connection, inner_model: models.TestInnerSqliteType
    ) -> None:
        await queries.delete_one_test_inner_sqlite_type(conn=aiosqlite_conn, table_id=inner_model.table_id)
