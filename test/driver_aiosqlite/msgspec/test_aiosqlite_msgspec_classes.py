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

import collections.abc
import datetime
import decimal
import json
import math
import random
from collections import UserString

import aiosqlite
import pytest
import pytest_asyncio

from test.driver_aiosqlite.msgspec.classes import models
from test.driver_aiosqlite.msgspec.classes import queries


@pytest.mark.asyncio(loop_scope="session")
class TestMsgspecClasses:
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

    @pytest_asyncio.fixture(scope="class", loop_scope="session")
    async def queries_obj(self, aiosqlite_conn: aiosqlite.Connection) -> queries.Queries:
        return queries.Queries(conn=aiosqlite_conn)

    @pytest.mark.asyncio(loop_scope="session")
    async def test_conn_attr(self, queries_obj: queries.Queries) -> None:
        assert isinstance(queries_obj.conn, aiosqlite.Connection)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AiosqliteTestMsgspecClasses::insert")
    async def test_insert(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        await queries_obj.insert_one_sqlite_type(
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
        name="AiosqliteTestMsgspecClasses::inner_insert", depends=["AiosqliteTestMsgspecClasses::insert"]
    )
    async def test_inner_insert(
        self,
        queries_obj: queries.Queries,
        inner_model: models.TestInnerSqliteType,
    ) -> None:
        await queries_obj.insert_one_inner_sqlite_type(
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
        name="AiosqliteTestMsgspecClasses::get_one", depends=["AiosqliteTestMsgspecClasses::inner_insert"]
    )
    async def test_get_one(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.get_one_sqlite_type(id_=model.id)

        assert result is not None

        assert isinstance(result, models.TestSqliteType)

        assert result == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_one_none", depends=["AiosqliteTestMsgspecClasses::get_one"]
    )
    async def test_get_one_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_sqlite_type(id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_one_inner", depends=["AiosqliteTestMsgspecClasses::get_one_none"]
    )
    async def test_get_one_inner(
        self,
        queries_obj: queries.Queries,
        inner_model: models.TestInnerSqliteType,
    ) -> None:
        result = await queries_obj.get_one_inner_sqlite_type(table_id=inner_model.table_id)

        assert result is not None

        assert isinstance(result, models.TestInnerSqliteType)
        assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_one_inner_none", depends=["AiosqliteTestMsgspecClasses::get_one_inner"]
    )
    async def test_get_one_inner_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_inner_sqlite_type(table_id=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_date", depends=["AiosqliteTestMsgspecClasses::get_one_inner_none"]
    )
    async def test_get_date(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.get_one_date(id_=model.id, date_test=model.date_test)

        assert result is not None

        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_date_none", depends=["AiosqliteTestMsgspecClasses::get_date"]
    )
    async def test_get_date_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_date(id_=0, date_test=datetime.date.today())

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_datetime", depends=["AiosqliteTestMsgspecClasses::get_date_none"]
    )
    async def test_get_datetime(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.get_one_datetime(id_=model.id, datetime_test=model.datetime_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_datetime_none", depends=["AiosqliteTestMsgspecClasses::get_datetime"]
    )
    async def test_get_datetime_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_datetime(id_=0, datetime_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_timestamp", depends=["AiosqliteTestMsgspecClasses::get_datetime_none"]
    )
    async def test_get_timestamp(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.get_one_timestamp(id_=model.id, timestamp_test=model.timestamp_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.timestamp_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_timestamp_none", depends=["AiosqliteTestMsgspecClasses::get_timestamp"]
    )
    async def test_get_timestamp_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_timestamp(id_=0, timestamp_test=datetime.datetime.now())

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_bool", depends=["AiosqliteTestMsgspecClasses::get_timestamp_none"]
    )
    async def test_get_bool(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.get_one_bool(id_=model.id, bool_test=model.bool_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.bool_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_bool_none", depends=["AiosqliteTestMsgspecClasses::get_bool"]
    )
    async def test_get_bool_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_bool(id_=0, bool_test=False)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_boolean", depends=["AiosqliteTestMsgspecClasses::get_bool_none"]
    )
    async def test_get_boolean(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.get_one_boolean(id_=model.id, boolean_test=model.boolean_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result == model.boolean_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_boolean_none", depends=["AiosqliteTestMsgspecClasses::get_boolean"]
    )
    async def test_get_boolean_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_boolean(id_=0, boolean_test=True)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_decimal", depends=["AiosqliteTestMsgspecClasses::get_boolean_none"]
    )
    async def test_get_decimal(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.get_one_decimal(id_=model.id, decimal_test=model.decimal_test)

        assert result is not None

        assert isinstance(result, decimal.Decimal)
        assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_decimal_none", depends=["AiosqliteTestMsgspecClasses::get_decimal"]
    )
    async def test_get_decimal_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_decimal(id_=0, decimal_test=decimal.Decimal("0.1"))

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_blob", depends=["AiosqliteTestMsgspecClasses::get_decimal_none"]
    )
    async def test_get_blob(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.get_one_blob(id_=model.id, blob_test=model.blob_test)

        assert result is not None

        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_blob_none", depends=["AiosqliteTestMsgspecClasses::get_blob"]
    )
    async def test_get_blob_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_blob(id_=0, blob_test=memoryview(b"test"))

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many", depends=["AiosqliteTestMsgspecClasses::get_blob_none"]
    )
    async def test_get_many(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        result = await queries_obj.get_many_sqlite_type(id_=model.id)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], models.TestSqliteType)

        assert result[0] == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_iter", depends=["AiosqliteTestMsgspecClasses::get_many"]
    )
    async def test_get_many_iter(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        async for result in queries_obj.get_many_sqlite_type(id_=model.id):
            assert result is not None
            assert isinstance(result, models.TestSqliteType)

            assert result == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_inner", depends=["AiosqliteTestMsgspecClasses::get_many_iter"]
    )
    async def test_get_many_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerSqliteType) -> None:
        result = await queries_obj.get_many_inner_sqlite_type(table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], models.TestInnerSqliteType)

        assert result[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_inner_iter", depends=["AiosqliteTestMsgspecClasses::get_many_inner"]
    )
    async def test_get_many_inner_iter(
        self, queries_obj: queries.Queries, inner_model: models.TestInnerSqliteType
    ) -> None:
        async for result in queries_obj.get_many_inner_sqlite_type(table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerSqliteType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_nullable_inner",
        depends=["AiosqliteTestMsgspecClasses::get_many_inner_iter"],
    )
    async def test_get_many_nullable_inner(
        self, queries_obj: queries.Queries, inner_model: models.TestInnerSqliteType
    ) -> None:
        result = await queries_obj.get_many_nullable_inner_sqlite_type(table_id=inner_model.table_id, int_test=None)
        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], models.TestInnerSqliteType)

        assert result[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_nullable_inner_iter",
        depends=["AiosqliteTestMsgspecClasses::get_many_nullable_inner"],
    )
    async def test_get_many_nullable_inner_iter(
        self, queries_obj: queries.Queries, inner_model: models.TestInnerSqliteType
    ) -> None:
        async for result in queries_obj.get_many_nullable_inner_sqlite_type(
            table_id=inner_model.table_id, int_test=inner_model.int_test
        ):
            assert result is not None
            assert isinstance(result, models.TestInnerSqliteType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_date",
        depends=["AiosqliteTestMsgspecClasses::get_many_nullable_inner_iter"],
    )
    async def test_get_many_date(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        result = await queries_obj.get_many_date(id_=model.id, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], datetime.date)

        assert result[0] == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_date_iter", depends=["AiosqliteTestMsgspecClasses::get_many_date"]
    )
    async def test_get_many_date_iter(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        async for result in queries_obj.get_many_date(id_=model.id, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)

            assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_datetime",
        depends=["AiosqliteTestMsgspecClasses::get_many_date_iter"],
    )
    async def test_get_many_datetime(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        result = await queries_obj.get_many_datetime(id_=model.id, datetime_test=model.datetime_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], datetime.datetime)

        assert result[0] == model.datetime_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_datetime_iter",
        depends=["AiosqliteTestMsgspecClasses::get_many_datetime"],
    )
    async def test_get_many_datetime_iter(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        async for result in queries_obj.get_many_datetime(id_=model.id, datetime_test=model.datetime_test):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.datetime_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_timestamp",
        depends=["AiosqliteTestMsgspecClasses::get_many_datetime_iter"],
    )
    async def test_get_many_timestamp(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        result = await queries_obj.get_many_timestamp(id_=model.id, timestamp_test=model.timestamp_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], datetime.datetime)

        assert result[0] == model.timestamp_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_timestamp_iter",
        depends=["AiosqliteTestMsgspecClasses::get_many_timestamp"],
    )
    async def test_get_many_timestamp_iter(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        async for result in queries_obj.get_many_timestamp(id_=model.id, timestamp_test=model.timestamp_test):
            assert result is not None
            assert isinstance(result, datetime.datetime)

            assert result == model.timestamp_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_bool",
        depends=["AiosqliteTestMsgspecClasses::get_many_timestamp_iter"],
    )
    async def test_get_many_bool(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        result = await queries_obj.get_many_bool(id_=model.id, bool_test=model.bool_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], bool)

        assert result[0] == model.bool_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_bool_iter", depends=["AiosqliteTestMsgspecClasses::get_many_bool"]
    )
    async def test_get_many_bool_iter(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        async for result in queries_obj.get_many_bool(id_=model.id, bool_test=model.bool_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.bool_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_boolean",
        depends=["AiosqliteTestMsgspecClasses::get_many_bool_iter"],
    )
    async def test_get_many_boolean(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        result = await queries_obj.get_many_boolean(id_=model.id, boolean_test=model.boolean_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], bool)

        assert result[0] == model.boolean_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_boolean_iter",
        depends=["AiosqliteTestMsgspecClasses::get_many_boolean"],
    )
    async def test_get_many_boolean_iter(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        async for result in queries_obj.get_many_boolean(id_=model.id, boolean_test=model.boolean_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result == model.boolean_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_decimal",
        depends=["AiosqliteTestMsgspecClasses::get_many_boolean_iter"],
    )
    async def test_get_many_decimal(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        result = await queries_obj.get_many_decimal(id_=model.id, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], decimal.Decimal)

        assert result[0] == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_decimal_iter",
        depends=["AiosqliteTestMsgspecClasses::get_many_decimal"],
    )
    async def test_get_many_decimal_iter(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        async for result in queries_obj.get_many_decimal(id_=model.id, decimal_test=model.decimal_test):
            assert result is not None
            assert isinstance(result, decimal.Decimal)

            assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_blob",
        depends=["AiosqliteTestMsgspecClasses::get_many_decimal_iter"],
    )
    async def test_get_many_blob(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        result = await queries_obj.get_many_blob(id_=model.id, blob_test=model.blob_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], memoryview)

        assert result[0] == model.blob_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_blob_iter", depends=["AiosqliteTestMsgspecClasses::get_many_blob"]
    )
    async def test_get_many_blob_iter(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        async for result in queries_obj.get_many_blob(id_=model.id, blob_test=model.blob_test):
            assert result is not None
            assert isinstance(result, memoryview)

            assert result == model.blob_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::insert_result", depends=["AiosqliteTestMsgspecClasses::get_many_blob_iter"]
    )
    async def test_insert_result(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.insert_result_one_sqlite_type(
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
        name="AiosqliteTestMsgspecClasses::update_result", depends=["AiosqliteTestMsgspecClasses::insert_result"]
    )
    async def test_update_result(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.update_result_one_sqlite_type(id_=model.id + 1)
        assert isinstance(result, aiosqlite.Cursor)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::delete_result", depends=["AiosqliteTestMsgspecClasses::update_result"]
    )
    async def test_delete_result(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.delete_result_one_sqlite_type(id_=model.id + 1)
        assert isinstance(result, aiosqlite.Cursor)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::insert_rows", depends=["AiosqliteTestMsgspecClasses::delete_result"]
    )
    async def test_insert_rows(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.insert_rows_one_sqlite_type(
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
        name="AiosqliteTestMsgspecClasses::update_rows", depends=["AiosqliteTestMsgspecClasses::insert_rows"]
    )
    async def test_update_rows(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.update_rows_one_sqlite_type(id_=model.id + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::delete_rows", depends=["AiosqliteTestMsgspecClasses::update_rows"]
    )
    async def test_delete_rows(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.delete_rows_one_sqlite_type(id_=model.id + 2)
        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::create_table_rows", depends=["AiosqliteTestMsgspecClasses::delete_rows"]
    )
    async def test_create_table_rows(
        self,
        queries_obj: queries.Queries,
        aiosqlite_conn: aiosqlite.Connection,
    ) -> None:
        result = await queries_obj.create_rows_table()
        assert isinstance(result, int)
        await aiosqlite_conn.execute("DROP TABLE test_create_rows_table;")
        await aiosqlite_conn.commit()
        assert result == -1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::insert_last_id", depends=["AiosqliteTestMsgspecClasses::create_table_rows"]
    )
    async def test_insert_last_id(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.insert_last_id_one_sqlite_type(
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
        name="AiosqliteTestMsgspecClasses::update_last_id", depends=["AiosqliteTestMsgspecClasses::insert_last_id"]
    )
    async def test_update_last_id(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.update_last_id_one_sqlite_type(id_=model.id + 3)
        assert isinstance(result, int)
        assert result == model.id + 3

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::delete_last_id", depends=["AiosqliteTestMsgspecClasses::update_last_id"]
    )
    async def test_delete_last_id(
        self,
        queries_obj: queries.Queries,
        model: models.TestSqliteType,
    ) -> None:
        result = await queries_obj.delete_last_id_one_sqlite_type(id_=model.id + 3)
        assert isinstance(result, int)
        assert result == model.id + 3

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::delete_sqlite_type", depends=["AiosqliteTestMsgspecClasses::delete_last_id"]
    )
    async def test_delete_sqlite_type(self, queries_obj: queries.Queries, model: models.TestSqliteType) -> None:
        await queries_obj.delete_one_sqlite_type(id_=model.id)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::delete_inner_sqlite_type",
        depends=["AiosqliteTestMsgspecClasses::delete_sqlite_type"],
    )
    async def test_delete_inner_sqlite_type(
        self, queries_obj: queries.Queries, inner_model: models.TestInnerSqliteType
    ) -> None:
        await queries_obj.delete_one_test_inner_sqlite_type(table_id=inner_model.table_id)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::insert_type_override",
    )
    async def test_insert_type_override(
        self, queries_obj: queries.Queries, override_model: models.TestTypeOverride
    ) -> None:
        await queries_obj.insert_type_override(id_=override_model.id, text_test=override_model.text_test)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_one_type_override",
        depends=["AiosqliteTestMsgspecClasses::insert_type_override"],
    )
    async def test_get_one_type_override(
        self, queries_obj: queries.Queries, override_model: models.TestTypeOverride
    ) -> None:
        result = await queries_obj.get_one_type_override(id_=override_model.id)
        assert result is not None
        assert result == override_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_one_type_override_none",
        depends=["AiosqliteTestMsgspecClasses::get_one_type_override"],
    )
    async def test_get_one_type_override_none(
        self, queries_obj: queries.Queries, override_model: models.TestTypeOverride
    ) -> None:
        result = await queries_obj.get_one_type_override(id_=override_model.id - 1)
        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_type_override",
        depends=["AiosqliteTestMsgspecClasses::get_one_type_override_none"],
    )
    async def test_get_many_type_override(
        self, queries_obj: queries.Queries, override_model: models.TestTypeOverride
    ) -> None:
        result = await queries_obj.get_many_type_override(id_=override_model.id)
        assert isinstance(result, collections.abc.Sequence)
        assert result[0] == override_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_one_text_type_override",
        depends=["AiosqliteTestMsgspecClasses::get_many_type_override"],
    )
    async def test_get_one_text_type_override(
        self, queries_obj: queries.Queries, override_model: models.TestTypeOverride
    ) -> None:
        result = await queries_obj.get_one_text_type_override(id_=override_model.id)
        assert result is not None
        assert result == override_model.text_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_one_text_type_override_none",
        depends=["AiosqliteTestMsgspecClasses::get_one_text_type_override"],
    )
    async def test_get_one_text_type_override_none(
        self, queries_obj: queries.Queries, override_model: models.TestTypeOverride
    ) -> None:
        result = await queries_obj.get_one_text_type_override(id_=override_model.id - 1)
        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::get_many_text_type_override",
        depends=["AiosqliteTestMsgspecClasses::get_one_text_type_override_none"],
    )
    async def test_get_many_text_type_override(
        self, queries_obj: queries.Queries, override_model: models.TestTypeOverride
    ) -> None:
        result = await queries_obj.get_many_text_type_override(id_=override_model.id)
        assert isinstance(result, collections.abc.Sequence)
        assert result[0] == override_model.text_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AiosqliteTestMsgspecClasses::delete_type_override",
        depends=["AiosqliteTestMsgspecClasses::get_many_text_type_override"],
    )
    async def test_delete_type_override(
        self, queries_obj: queries.Queries, override_model: models.TestTypeOverride
    ) -> None:
        await queries_obj.delete_type_override(id_=override_model.id)
