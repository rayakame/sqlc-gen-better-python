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

import collections.abc
import datetime
import decimal
import json
import math
import typing
from collections import UserString

import asyncmy
import asyncmy.cursors
import pytest

from test.driver_asyncmy import no_row_conn
from test.driver_asyncmy.pydantic.functions import enums
from test.driver_asyncmy.pydantic.functions import models
from test.driver_asyncmy.pydantic.functions import queries

MODEL_ID = 9151
OVERRIDE_ID = 9251
OVERRIDE_NONE_ID = 9252
RESERVED_ID = 9351
MISSING_ID = 9902
SUITE_TAG = "asyncmy-pydantic-functions"
EXPECTED_MONTH = "2026-01"
DECIMAL_PADDED = "12.3400"
BINARY_UNPADDED = b"\xaa\xbb\xcc"
BINARY_LENGTH = 16
BIT_BYTE = b"\x80"


@pytest.mark.asyncio(loop_scope="session")
class TestAsyncmyPydanticFunctions:
    @pytest.fixture(scope="session")
    def override_model(self) -> models.TestTypeOverride:
        return models.TestTypeOverride(id_=OVERRIDE_ID, text_test=UserString("Test"))

    @pytest.fixture(scope="session")
    def model(self) -> models.TestMysqlType:
        return models.TestMysqlType(
            id_=MODEL_ID,
            int_test=42,
            integer_test=-1_000_000,
            mediumint_test=8_388_607,
            smallint_test=32_767,
            tinyint_test=127,
            bigint_test=9_007_199_254_740_991,
            int_unsigned_test=4_000_000_000,
            bigint_unsigned_test=2**63 + 10,
            year_test=2026,
            tinyint1_test=True,
            bool_test=False,
            boolean_test=True,
            float_test=3.5,
            double_test=math.pi,
            double_precision_test=1.41421,
            real_test=math.e,
            decimal_test=decimal.Decimal("12.34"),
            numeric_test=decimal.Decimal("100.10"),
            char_test="CHAR10",
            varchar_test="Hello varchar",
            tinytext_test="tiny text",
            text_test="Some text",
            mediumtext_test="medium text",
            longtext_test="long text",
            binary_test=memoryview(BINARY_UNPADDED + b"\x00" * (BINARY_LENGTH - len(BINARY_UNPADDED))),
            varbinary_test=memoryview(b"\x00\x01\x02hello"),
            tinyblob_test=memoryview(b"tinyblob"),
            blob_test=memoryview(b"\x00\x01\x02blob"),
            mediumblob_test=memoryview(b"mediumblob"),
            longblob_test=memoryview(b"longblob"),
            bit_test=memoryview(BIT_BYTE),
            date_test=datetime.date(2026, 1, 2),
            datetime_test=datetime.datetime(2026, 1, 2, 3, 4, 5),
            datetime6_test=datetime.datetime(2026, 1, 2, 3, 4, 5, 123456),
            timestamp_test=datetime.datetime(2026, 1, 2, 3, 4, 5),
            time_test=datetime.timedelta(hours=13, minutes=45, seconds=30),
            json_test=json.dumps({"foo": "bar", "count": 3}, separators=(",", ":")),
            mood=enums.TestMysqlTypesMood.VALUE_24H,
            tag=enums.TestMysqlTypesTag.GAMMA,
        )

    @pytest.fixture(scope="session")
    def inner_model(self, model: models.TestMysqlType) -> models.TestInnerMysqlType:
        return models.TestInnerMysqlType(
            table_id=model.id_,
            int_test=None,
            integer_test=model.integer_test,
            mediumint_test=model.mediumint_test,
            smallint_test=model.smallint_test,
            tinyint_test=model.tinyint_test,
            bigint_test=model.bigint_test,
            int_unsigned_test=model.int_unsigned_test,
            bigint_unsigned_test=model.bigint_unsigned_test,
            year_test=model.year_test,
            tinyint1_test=None,
            bool_test=model.bool_test,
            boolean_test=model.boolean_test,
            float_test=None,
            double_test=model.double_test,
            double_precision_test=model.double_precision_test,
            real_test=model.real_test,
            decimal_test=None,
            numeric_test=model.numeric_test,
            char_test=model.char_test,
            varchar_test=model.varchar_test,
            tinytext_test=model.tinytext_test,
            text_test=model.text_test,
            mediumtext_test=model.mediumtext_test,
            longtext_test=model.longtext_test,
            binary_test=None,
            varbinary_test=model.varbinary_test,
            tinyblob_test=model.tinyblob_test,
            blob_test=None,
            mediumblob_test=model.mediumblob_test,
            longblob_test=model.longblob_test,
            bit_test=model.bit_test,
            date_test=None,
            datetime_test=model.datetime_test,
            datetime6_test=model.datetime6_test,
            timestamp_test=None,
            time_test=model.time_test,
            json_test=None,
            mood=enums.TestInnerMysqlTypesMood.VALUE__HIDDEN,
            tag=None,
        )

    @pytest.mark.asyncio(loop_scope="session")
    async def test_enum_members(self) -> None:
        assert enums.TestMysqlTypesMood.VALUE_24H.value == "24h"
        assert enums.TestMysqlTypesMood.VALUE__HIDDEN.value == "_hidden"
        assert enums.TestInnerMysqlTypesMood.VALUE_24H.value == "24h"
        assert enums.TestMysqlTypesTag.BETA.value == "beta"

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::insert")
    async def test_insert(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        await queries.insert_one_mysql_type(
            conn=asyncmy_conn,
            id_=model.id_,
            int_test=model.int_test,
            integer_test=model.integer_test,
            mediumint_test=model.mediumint_test,
            smallint_test=model.smallint_test,
            tinyint_test=model.tinyint_test,
            bigint_test=model.bigint_test,
            int_unsigned_test=model.int_unsigned_test,
            bigint_unsigned_test=model.bigint_unsigned_test,
            year_test=model.year_test,
            tinyint1_test=model.tinyint1_test,
            bool_test=model.bool_test,
            boolean_test=model.boolean_test,
            float_test=model.float_test,
            double_test=model.double_test,
            double_precision_test=model.double_precision_test,
            real_test=model.real_test,
            decimal_test=model.decimal_test,
            numeric_test=model.numeric_test,
            # char(10) strips trailing spaces on return; the model holds the stripped value.
            char_test=model.char_test + "    ",
            varchar_test=model.varchar_test,
            tinytext_test=model.tinytext_test,
            text_test=model.text_test,
            mediumtext_test=model.mediumtext_test,
            longtext_test=model.longtext_test,
            # binary(16) is right-padded with NUL bytes on return; the model holds the padded value.
            binary_test=memoryview(BINARY_UNPADDED),
            varbinary_test=model.varbinary_test,
            tinyblob_test=model.tinyblob_test,
            blob_test=model.blob_test,
            mediumblob_test=model.mediumblob_test,
            longblob_test=model.longblob_test,
            bit_test=model.bit_test,
            date_test=model.date_test,
            datetime_test=model.datetime_test,
            datetime6_test=model.datetime6_test,
            timestamp_test=model.timestamp_test,
            time_test=model.time_test,
            json_test=model.json_test,
            mood=model.mood,
            tag=model.tag,
        )

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::inner_insert", depends=["AsyncmyTestPydanticFunctions::insert"])
    async def test_inner_insert(
        self,
        asyncmy_conn: asyncmy.Connection,
        inner_model: models.TestInnerMysqlType,
    ) -> None:
        await queries.insert_one_inner_mysql_type(
            conn=asyncmy_conn,
            table_id=inner_model.table_id,
            int_test=inner_model.int_test,
            integer_test=inner_model.integer_test,
            mediumint_test=inner_model.mediumint_test,
            smallint_test=inner_model.smallint_test,
            tinyint_test=inner_model.tinyint_test,
            bigint_test=inner_model.bigint_test,
            int_unsigned_test=inner_model.int_unsigned_test,
            bigint_unsigned_test=inner_model.bigint_unsigned_test,
            year_test=inner_model.year_test,
            tinyint1_test=inner_model.tinyint1_test,
            bool_test=inner_model.bool_test,
            boolean_test=inner_model.boolean_test,
            float_test=inner_model.float_test,
            double_test=inner_model.double_test,
            double_precision_test=inner_model.double_precision_test,
            real_test=inner_model.real_test,
            decimal_test=inner_model.decimal_test,
            numeric_test=inner_model.numeric_test,
            char_test=inner_model.char_test,
            varchar_test=inner_model.varchar_test,
            tinytext_test=inner_model.tinytext_test,
            text_test=inner_model.text_test,
            mediumtext_test=inner_model.mediumtext_test,
            longtext_test=inner_model.longtext_test,
            binary_test=inner_model.binary_test,
            varbinary_test=inner_model.varbinary_test,
            tinyblob_test=inner_model.tinyblob_test,
            blob_test=inner_model.blob_test,
            mediumblob_test=inner_model.mediumblob_test,
            longblob_test=inner_model.longblob_test,
            bit_test=inner_model.bit_test,
            date_test=inner_model.date_test,
            datetime_test=inner_model.datetime_test,
            datetime6_test=inner_model.datetime6_test,
            timestamp_test=inner_model.timestamp_test,
            time_test=inner_model.time_test,
            json_test=inner_model.json_test,
            mood=inner_model.mood,
            tag=inner_model.tag,
        )

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_one", depends=["AsyncmyTestPydanticFunctions::inner_insert"])
    async def test_get_one(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_mysql_type(conn=asyncmy_conn, id_=MODEL_ID)

        assert result is not None
        assert isinstance(result, models.TestMysqlType)

        assert result.tinyint1_test is True
        assert result.bool_test is False
        assert result.boolean_test is True
        assert len(result.binary_test) == BINARY_LENGTH
        assert result.char_test == model.char_test
        assert result.datetime6_test.microsecond == model.datetime6_test.microsecond
        assert isinstance(result.time_test, datetime.timedelta)
        # MySQL normalizes JSON spacing; compare parsed values, never strings.
        assert json.loads(result.json_test) == json.loads(model.json_test)
        assert result == model.model_copy(update={"json_test": result.json_test})

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_one_none", depends=["AsyncmyTestPydanticFunctions::get_one"])
    async def test_get_one_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_mysql_type(conn=asyncmy_conn, id_=MISSING_ID)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_one_inner", depends=["AsyncmyTestPydanticFunctions::get_one_none"])
    async def test_get_one_inner(
        self,
        asyncmy_conn: asyncmy.Connection,
        inner_model: models.TestInnerMysqlType,
    ) -> None:
        result = await queries.get_one_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, models.TestInnerMysqlType)

        assert result.tinyint1_test is None
        assert result.bool_test is False
        assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_one_inner_none", depends=["AsyncmyTestPydanticFunctions::get_one_inner"])
    async def test_get_one_inner_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_inner_mysql_type(conn=asyncmy_conn, table_id=MISSING_ID)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_date", depends=["AsyncmyTestPydanticFunctions::get_one_inner_none"])
    async def test_get_date(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_date(conn=asyncmy_conn, id_=MODEL_ID, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_date_none", depends=["AsyncmyTestPydanticFunctions::get_date"])
    async def test_get_date_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_date(conn=asyncmy_conn, id_=MISSING_ID, date_test=model.date_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_datetime", depends=["AsyncmyTestPydanticFunctions::get_date_none"])
    async def test_get_datetime(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_datetime(conn=asyncmy_conn, id_=MODEL_ID, datetime_test=model.datetime_test)

        assert result is not None
        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_datetime_none", depends=["AsyncmyTestPydanticFunctions::get_datetime"])
    async def test_get_datetime_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_datetime(conn=asyncmy_conn, id_=MISSING_ID, datetime_test=model.datetime_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_time", depends=["AsyncmyTestPydanticFunctions::get_datetime_none"])
    async def test_get_time(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_time(conn=asyncmy_conn, id_=MODEL_ID, time_test=model.time_test)

        assert result is not None
        # MySQL time columns map to datetime.timedelta, not datetime.time.
        assert isinstance(result, datetime.timedelta)
        assert result == model.time_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_time_none", depends=["AsyncmyTestPydanticFunctions::get_time"])
    async def test_get_time_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_time(conn=asyncmy_conn, id_=MISSING_ID, time_test=model.time_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_bool", depends=["AsyncmyTestPydanticFunctions::get_time_none"])
    async def test_get_bool(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_bool(conn=asyncmy_conn, id_=MODEL_ID, tinyint1_test=model.tinyint1_test)

        assert result is not None
        assert result is True

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_bool_none", depends=["AsyncmyTestPydanticFunctions::get_bool"])
    async def test_get_bool_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_bool(conn=asyncmy_conn, id_=MISSING_ID, tinyint1_test=model.tinyint1_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_decimal", depends=["AsyncmyTestPydanticFunctions::get_bool_none"])
    async def test_get_decimal(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_decimal(conn=asyncmy_conn, id_=MODEL_ID, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, decimal.Decimal)
        # decimal(12,4) comes back padded to scale 4.
        assert str(result) == DECIMAL_PADDED
        assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_decimal_none", depends=["AsyncmyTestPydanticFunctions::get_decimal"])
    async def test_get_decimal_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_decimal(conn=asyncmy_conn, id_=MISSING_ID, decimal_test=model.decimal_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_blob", depends=["AsyncmyTestPydanticFunctions::get_decimal_none"])
    async def test_get_blob(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_blob(conn=asyncmy_conn, id_=MODEL_ID, blob_test=model.blob_test)

        assert result is not None
        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_blob_none", depends=["AsyncmyTestPydanticFunctions::get_blob"])
    async def test_get_blob_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_blob(conn=asyncmy_conn, id_=MISSING_ID, blob_test=model.blob_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_bit", depends=["AsyncmyTestPydanticFunctions::get_blob_none"])
    async def test_get_bit(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_bit(conn=asyncmy_conn, id_=MODEL_ID)

        assert result is not None
        # bit(8) comes back as a one-byte memoryview.
        assert isinstance(result, memoryview)
        assert len(result) == 1
        assert bytes(result) == BIT_BYTE

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_year", depends=["AsyncmyTestPydanticFunctions::get_bit"])
    async def test_get_year(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_year(conn=asyncmy_conn, id_=MODEL_ID)

        assert result is not None
        assert isinstance(result, int)
        assert result == model.year_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_json", depends=["AsyncmyTestPydanticFunctions::get_year"])
    async def test_get_json(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_json(conn=asyncmy_conn, id_=MODEL_ID)

        assert result is not None
        assert isinstance(result, str)
        # MySQL normalizes JSON spacing; compare parsed values, never strings.
        assert json.loads(result) == json.loads(model.json_test)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_mood", depends=["AsyncmyTestPydanticFunctions::get_json"])
    async def test_get_mood(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_mood(conn=asyncmy_conn, id_=MODEL_ID, mood=model.mood)

        assert result is not None
        assert isinstance(result, enums.TestMysqlTypesMood)
        assert result is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_tag", depends=["AsyncmyTestPydanticFunctions::get_mood"])
    async def test_get_tag(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_tag(conn=asyncmy_conn, id_=MODEL_ID)

        assert result is not None
        assert isinstance(result, enums.TestMysqlTypesTag)
        assert result is model.tag

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many", depends=["AsyncmyTestPydanticFunctions::get_tag"])
    async def test_get_many(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = await queries.get_many_mysql_type(conn=asyncmy_conn, id_=MODEL_ID)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert len(result) == 1
        assert isinstance(result[0], models.TestMysqlType)

        assert result[0] == model.model_copy(update={"json_test": result[0].json_test})

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_iter", depends=["AsyncmyTestPydanticFunctions::get_many"])
    async def test_get_many_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_mysql_type(conn=asyncmy_conn, id_=MODEL_ID):
            assert result is not None
            assert isinstance(result, models.TestMysqlType)

            assert result == model.model_copy(update={"json_test": result.json_test})

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_inner", depends=["AsyncmyTestPydanticFunctions::get_many_iter"])
    async def test_get_many_inner(self, asyncmy_conn: asyncmy.Connection, inner_model: models.TestInnerMysqlType) -> None:
        result = await queries.get_many_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], models.TestInnerMysqlType)

        assert result[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_inner_iter", depends=["AsyncmyTestPydanticFunctions::get_many_inner"])
    async def test_get_many_inner_iter(self, asyncmy_conn: asyncmy.Connection, inner_model: models.TestInnerMysqlType) -> None:
        async for result in queries.get_many_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestPydanticFunctions::get_many_nullable_inner",
        depends=["AsyncmyTestPydanticFunctions::get_many_inner_iter"],
    )
    async def test_get_many_nullable_inner(self, asyncmy_conn: asyncmy.Connection, inner_model: models.TestInnerMysqlType) -> None:
        # int_test is None; the query matches it via the NULL-safe <=> operator.
        result = await queries.get_many_nullable_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id, int_test=inner_model.int_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], models.TestInnerMysqlType)

        assert result[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestPydanticFunctions::get_many_nullable_inner_iter",
        depends=["AsyncmyTestPydanticFunctions::get_many_nullable_inner"],
    )
    async def test_get_many_nullable_inner_iter(self, asyncmy_conn: asyncmy.Connection, inner_model: models.TestInnerMysqlType) -> None:
        async for result in queries.get_many_nullable_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id, int_test=inner_model.int_test):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestPydanticFunctions::get_many_date",
        depends=["AsyncmyTestPydanticFunctions::get_many_nullable_inner_iter"],
    )
    async def test_get_many_date(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = await queries.get_many_date(conn=asyncmy_conn, id_=MODEL_ID, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], datetime.date)

        assert result[0] == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_date_iter", depends=["AsyncmyTestPydanticFunctions::get_many_date"])
    async def test_get_many_date_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_date(conn=asyncmy_conn, id_=MODEL_ID, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)

            assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_time", depends=["AsyncmyTestPydanticFunctions::get_many_date_iter"])
    async def test_get_many_time(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = await queries.get_many_time(conn=asyncmy_conn, id_=MODEL_ID, time_test=model.time_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], datetime.timedelta)

        assert result[0] == model.time_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_time_iter", depends=["AsyncmyTestPydanticFunctions::get_many_time"])
    async def test_get_many_time_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_time(conn=asyncmy_conn, id_=MODEL_ID, time_test=model.time_test):
            assert result is not None
            assert isinstance(result, datetime.timedelta)

            assert result == model.time_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_bool", depends=["AsyncmyTestPydanticFunctions::get_many_time_iter"])
    async def test_get_many_bool(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = await queries.get_many_bool(conn=asyncmy_conn, id_=MODEL_ID, tinyint1_test=model.tinyint1_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], bool)

        assert result[0] is True

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_bool_iter", depends=["AsyncmyTestPydanticFunctions::get_many_bool"])
    async def test_get_many_bool_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_bool(conn=asyncmy_conn, id_=MODEL_ID, tinyint1_test=model.tinyint1_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result is True

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_decimal", depends=["AsyncmyTestPydanticFunctions::get_many_bool_iter"])
    async def test_get_many_decimal(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = await queries.get_many_decimal(conn=asyncmy_conn, id_=MODEL_ID, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], decimal.Decimal)

        assert str(result[0]) == DECIMAL_PADDED
        assert result[0] == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestPydanticFunctions::get_many_decimal_iter",
        depends=["AsyncmyTestPydanticFunctions::get_many_decimal"],
    )
    async def test_get_many_decimal_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_decimal(conn=asyncmy_conn, id_=MODEL_ID, decimal_test=model.decimal_test):
            assert result is not None
            assert isinstance(result, decimal.Decimal)

            assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_mood", depends=["AsyncmyTestPydanticFunctions::get_many_decimal_iter"])
    async def test_get_many_mood(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = await queries.get_many_mood(conn=asyncmy_conn, mood=model.mood)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        # The query is not id-filtered; other rows may exist, so assert containment.
        assert result
        for mood in result:
            assert isinstance(mood, enums.TestMysqlTypesMood)
            assert mood is model.mood

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_many_mood_iter", depends=["AsyncmyTestPydanticFunctions::get_many_mood"])
    async def test_get_many_mood_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        results = [mood async for mood in queries.get_many_mood(conn=asyncmy_conn, mood=model.mood)]

        assert results
        for mood in results:
            assert isinstance(mood, enums.TestMysqlTypesMood)
            assert mood is model.mood

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::list_months", depends=["AsyncmyTestPydanticFunctions::get_many_mood_iter"])
    async def test_list_months(self, asyncmy_conn: asyncmy.Connection) -> None:
        # DATE_FORMAT contains literal % characters; regression for the percent-doubling bug.
        result = await queries.list_months(conn=asyncmy_conn)

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        for month in result:
            assert isinstance(month, str)
        assert EXPECTED_MONTH in result

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::list_months_iter", depends=["AsyncmyTestPydanticFunctions::list_months"])
    async def test_list_months_iter(self, asyncmy_conn: asyncmy.Connection) -> None:
        results = [month async for month in queries.list_months(conn=asyncmy_conn)]

        assert EXPECTED_MONTH in results

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::count", depends=["AsyncmyTestPydanticFunctions::list_months_iter"])
    async def test_count(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = await queries.count_mysql_types(conn=asyncmy_conn)

        assert result is not None
        assert isinstance(result, int)
        assert result >= 1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::update_rows", depends=["AsyncmyTestPydanticFunctions::count"])
    async def test_update_rows(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = await queries.update_varchar_test(conn=asyncmy_conn, varchar_test="updated varchar", id_=MODEL_ID)

        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::update_rows_none", depends=["AsyncmyTestPydanticFunctions::update_rows"])
    async def test_update_rows_none(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = await queries.update_varchar_test(conn=asyncmy_conn, varchar_test="updated varchar", id_=MISSING_ID)

        assert isinstance(result, int)
        assert result == 0

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::execresult_cursor", depends=["AsyncmyTestPydanticFunctions::update_rows_none"])
    async def test_execresult_cursor(self, asyncmy_conn: asyncmy.Connection) -> None:
        cursor = await queries.all_mysql_types_cursor(conn=asyncmy_conn)

        assert isinstance(cursor, asyncmy.cursors.Cursor)
        rows = await cursor.fetchall()
        assert any(row[0] == MODEL_ID for row in rows)
        await cursor.close()

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::exec_last_id", depends=["AsyncmyTestPydanticFunctions::execresult_cursor"])
    async def test_exec_last_id(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = await queries.insert_exec_last_id(conn=asyncmy_conn, name=SUITE_TAG)

        assert result is not None
        assert isinstance(result, int)
        # AUTO_INCREMENT counters persist across runs; never assert exact ids.
        assert result > 0

        name = await queries.get_exec_last_id_name(conn=asyncmy_conn, id_=result)
        assert name == SUITE_TAG

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::insert_reserved_arg", depends=["AsyncmyTestPydanticFunctions::exec_last_id"])
    async def test_insert_reserved_arg(self, asyncmy_conn: asyncmy.Connection) -> None:
        await queries.insert_reserved_arg(conn=asyncmy_conn, id_=RESERVED_ID, conn_2=SUITE_TAG)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::get_reserved_arg", depends=["AsyncmyTestPydanticFunctions::insert_reserved_arg"])
    async def test_get_reserved_arg(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = await queries.get_reserved_arg(conn=asyncmy_conn, conn_2=SUITE_TAG)

        assert result is not None
        assert result == models.TestReservedArg(id_=RESERVED_ID, conn=SUITE_TAG)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::delete", depends=["AsyncmyTestPydanticFunctions::get_reserved_arg"])
    async def test_delete(self, asyncmy_conn: asyncmy.Connection) -> None:
        await queries.delete_one_mysql_type(conn=asyncmy_conn, id_=MODEL_ID)

        result = await queries.get_one_mysql_type(conn=asyncmy_conn, id_=MODEL_ID)
        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestPydanticFunctions::insert_type_override")
    async def test_insert_type_override(self, asyncmy_conn: asyncmy.Connection, override_model: models.TestTypeOverride) -> None:
        await queries.insert_type_override(conn=asyncmy_conn, id_=override_model.id_, text_test=override_model.text_test)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestPydanticFunctions::get_type_override",
        depends=["AsyncmyTestPydanticFunctions::insert_type_override"],
    )
    async def test_get_type_override(self, asyncmy_conn: asyncmy.Connection, override_model: models.TestTypeOverride) -> None:
        result = await queries.get_type_override(conn=asyncmy_conn, id_=override_model.id_)

        assert result is not None
        assert isinstance(result.text_test, UserString)
        assert result == override_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestPydanticFunctions::get_type_override_none_value",
        depends=["AsyncmyTestPydanticFunctions::get_type_override"],
    )
    async def test_get_type_override_none_value(self, asyncmy_conn: asyncmy.Connection) -> None:
        # The UserString override column is nullable; None must round-trip too.
        await queries.insert_type_override(conn=asyncmy_conn, id_=OVERRIDE_NONE_ID, text_test=None)

        result = await queries.get_type_override(conn=asyncmy_conn, id_=OVERRIDE_NONE_ID)
        assert result is not None
        assert result.text_test is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestPydanticFunctions::get_type_override_missing",
        depends=["AsyncmyTestPydanticFunctions::get_type_override_none_value"],
    )
    async def test_get_type_override_missing(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = await queries.get_type_override(conn=asyncmy_conn, id_=MISSING_ID)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestPydanticFunctions::cleanup",
        depends=["AsyncmyTestPydanticFunctions::delete", "AsyncmyTestPydanticFunctions::get_type_override_missing"],
    )
    async def test_cleanup(self, asyncmy_conn: asyncmy.Connection) -> None:
        # Tables without generated delete queries are cleaned directly so the
        # fixed ids are free for the next chain.
        async with asyncmy_conn.cursor() as cur:
            await cur.execute("DELETE FROM test_inner_mysql_types WHERE table_id = %s", (MODEL_ID,))  # pyright: ignore[reportUnknownMemberType]
            await cur.execute("DELETE FROM test_type_override WHERE id IN (%s, %s)", (OVERRIDE_ID, OVERRIDE_NONE_ID))  # pyright: ignore[reportUnknownMemberType]
            await cur.execute("DELETE FROM test_reserved_args WHERE id = %s", (RESERVED_ID,))  # pyright: ignore[reportUnknownMemberType]
            await cur.execute("DELETE FROM test_execlastid WHERE name = %s", (SUITE_TAG,))  # pyright: ignore[reportUnknownMemberType]

    @pytest.mark.asyncio(loop_scope="session")
    async def test_one_missing_rows_return_none(self, asyncmy_conn: asyncmy.Connection) -> None:
        # Every :one not-found branch, plus the no-insert :execlastid.
        assert await queries.get_one_mysql_type(conn=asyncmy_conn, id_=-1) is None
        assert await queries.get_one_inner_mysql_type(conn=asyncmy_conn, table_id=-1) is None
        assert await queries.get_one_date(conn=asyncmy_conn, id_=-1, date_test=datetime.date(1970, 1, 1)) is None
        assert await queries.get_one_datetime(conn=asyncmy_conn, id_=-1, datetime_test=datetime.datetime(1970, 1, 1)) is None
        assert await queries.get_one_time(conn=asyncmy_conn, id_=-1, time_test=datetime.timedelta()) is None
        assert await queries.get_one_bool(conn=asyncmy_conn, id_=-1, tinyint1_test=False) is None
        assert await queries.get_one_decimal(conn=asyncmy_conn, id_=-1, decimal_test=decimal.Decimal(0)) is None
        assert await queries.get_one_blob(conn=asyncmy_conn, id_=-1, blob_test=memoryview(b"")) is None
        assert await queries.get_one_bit(conn=asyncmy_conn, id_=-1) is None
        assert await queries.get_one_year(conn=asyncmy_conn, id_=-1) is None
        assert await queries.get_one_json(conn=asyncmy_conn, id_=-1) is None
        assert await queries.get_one_mood(conn=asyncmy_conn, id_=-1, mood=enums.TestMysqlTypesMood.SAD) is None
        assert await queries.get_one_tag(conn=asyncmy_conn, id_=-1) is None
        assert await queries.get_exec_last_id_name(conn=asyncmy_conn, id_=-1) is None
        assert await queries.get_type_override(conn=asyncmy_conn, id_=-1) is None
        assert await queries.get_reserved_arg(conn=asyncmy_conn, conn_2="missing") is None
        assert await queries.touch_exec_last_id(conn=asyncmy_conn, name="untouched", id_=-1) is None

        # count(*) always returns a row; its miss branch needs the stub.
        stub = typing.cast("asyncmy.Connection", no_row_conn.NoRowConn())
        assert await queries.count_mysql_types(conn=stub) is None
