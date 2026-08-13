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

import dataclasses
import datetime
import decimal
import json
import math
from collections import UserString

import asyncmy
import asyncmy.cursors
import pytest

from test.driver_asyncmy.dataclass.functions import enums
from test.driver_asyncmy.dataclass.functions import models
from test.driver_asyncmy.dataclass.functions import queries
from test.driver_asyncmy.dataclass.functions import queries_slice

MODEL_ID = 6500
NORMALIZATION_ID = 6501
OVERRIDE_ID = 6600
OVERRIDE_NULL_ID = 6601
RESERVED_ARG_ID = 6650
RESERVED_ARG_VALUE = "dataclass-functions-conn"
SLICE_ID_BASE = 6900
SLICE_ROW_COUNT = 4


@pytest.mark.asyncio(loop_scope="session")
class TestAsyncmyDataclassFunctions:
    @pytest.fixture(scope="session")
    def override_model(self) -> models.TestTypeOverride:
        return models.TestTypeOverride(id_=OVERRIDE_ID, text_test=UserString("Test"))

    @pytest.fixture(scope="session")
    def model(self) -> models.TestMysqlType:
        return models.TestMysqlType(
            id_=MODEL_ID,
            int_test=42,
            integer_test=-7,
            mediumint_test=8_388_607,
            smallint_test=32_767,
            tinyint_test=127,
            bigint_test=9_007_199_254_740_991,
            int_unsigned_test=4_294_967_295,
            bigint_unsigned_test=2**63 + 10,
            year_test=2024,
            tinyint1_test=True,
            bool_test=True,
            boolean_test=False,
            float_test=2.5,
            double_test=math.e,
            double_precision_test=1.41421,
            real_test=math.pi,
            decimal_test=decimal.Decimal("12.3400"),
            numeric_test=decimal.Decimal("3.50"),
            char_test="ABCDEFGHIJ",
            varchar_test="Hello varchar",
            tinytext_test="tiny text",
            text_test="Some text",
            mediumtext_test="medium text",
            longtext_test="long text",
            binary_test=memoryview(b"0123456789abcdef"),
            varbinary_test=memoryview(b"\x00\x01\x02hello"),
            tinyblob_test=memoryview(b"tiny blob"),
            blob_test=memoryview(b"\x00\x01\x02blob"),
            mediumblob_test=memoryview(b"medium blob"),
            longblob_test=memoryview(b"long blob"),
            bit_test=memoryview(b"\x80"),
            date_test=datetime.date(2026, 1, 5),
            datetime_test=datetime.datetime(2026, 1, 5, 12, 30, 45),
            datetime6_test=datetime.datetime(2026, 1, 5, 12, 30, 45, 123456),
            timestamp_test=datetime.datetime(2026, 1, 5, 12, 30, 45),
            time_test=datetime.timedelta(hours=13, minutes=14, seconds=15),
            json_test=json.dumps({"foo": "bar", "count": 2}),
            mood=enums.TestMysqlTypesMood.VALUE_24H,
            tag=enums.TestMysqlTypesTag.ALPHA,
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
            bool_test=True,
            boolean_test=None,
            float_test=model.float_test,
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
            binary_test=model.binary_test,
            varbinary_test=None,
            tinyblob_test=model.tinyblob_test,
            blob_test=None,
            mediumblob_test=model.mediumblob_test,
            longblob_test=model.longblob_test,
            bit_test=model.bit_test,
            date_test=model.date_test,
            datetime_test=None,
            datetime6_test=model.datetime6_test,
            timestamp_test=None,
            time_test=model.time_test,
            json_test=None,
            mood=enums.TestInnerMysqlTypesMood.VALUE__HIDDEN,
            tag=enums.TestInnerMysqlTypesTag.BETA,
        )

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::insert")
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
            char_test=model.char_test,
            varchar_test=model.varchar_test,
            tinytext_test=model.tinytext_test,
            text_test=model.text_test,
            mediumtext_test=model.mediumtext_test,
            longtext_test=model.longtext_test,
            binary_test=model.binary_test,
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
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::inner_insert", depends=["AsyncmyTestDataclassFunctions::insert"])
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
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_one", depends=["AsyncmyTestDataclassFunctions::inner_insert"])
    async def test_get_one(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_mysql_type(conn=asyncmy_conn, id_=model.id_)

        assert result is not None

        assert isinstance(result, models.TestMysqlType)

        assert result.tinyint1_test is True
        assert result.bool_test is True
        assert result.boolean_test is False
        assert result.datetime6_test.microsecond == model.datetime6_test.microsecond
        # MySQL normalizes JSON spacing, so the raw string may differ.
        assert json.loads(result.json_test) == json.loads(model.json_test)
        assert dataclasses.replace(result, json_test=model.json_test) == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_one_none", depends=["AsyncmyTestDataclassFunctions::get_one"])
    async def test_get_one_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_mysql_type(conn=asyncmy_conn, id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_one_inner", depends=["AsyncmyTestDataclassFunctions::get_one_none"])
    async def test_get_one_inner(
        self,
        asyncmy_conn: asyncmy.Connection,
        inner_model: models.TestInnerMysqlType,
    ) -> None:
        result = await queries.get_one_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id)

        assert result is not None

        assert isinstance(result, models.TestInnerMysqlType)
        assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_one_inner_none",
        depends=["AsyncmyTestDataclassFunctions::get_one_inner"],
    )
    async def test_get_one_inner_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_inner_mysql_type(conn=asyncmy_conn, table_id=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_date", depends=["AsyncmyTestDataclassFunctions::get_one_inner_none"])
    async def test_get_date(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_date(conn=asyncmy_conn, id_=model.id_, date_test=model.date_test)

        assert result is not None

        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_date_none", depends=["AsyncmyTestDataclassFunctions::get_date"])
    async def test_get_date_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_date(conn=asyncmy_conn, id_=0, date_test=model.date_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_datetime", depends=["AsyncmyTestDataclassFunctions::get_date_none"])
    async def test_get_datetime(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_datetime(conn=asyncmy_conn, id_=model.id_, datetime_test=model.datetime_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_datetime_none", depends=["AsyncmyTestDataclassFunctions::get_datetime"])
    async def test_get_datetime_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_datetime(conn=asyncmy_conn, id_=0, datetime_test=model.datetime_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_time", depends=["AsyncmyTestDataclassFunctions::get_datetime_none"])
    async def test_get_time(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_time(conn=asyncmy_conn, id_=model.id_, time_test=model.time_test)

        assert result is not None

        # MySQL time columns arrive as timedelta, not datetime.time.
        assert isinstance(result, datetime.timedelta)
        assert result == model.time_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_time_none", depends=["AsyncmyTestDataclassFunctions::get_time"])
    async def test_get_time_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_time(conn=asyncmy_conn, id_=0, time_test=model.time_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_bool", depends=["AsyncmyTestDataclassFunctions::get_time_none"])
    async def test_get_bool(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_bool(conn=asyncmy_conn, id_=model.id_, tinyint1_test=model.tinyint1_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result is True

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_bool_none", depends=["AsyncmyTestDataclassFunctions::get_bool"])
    async def test_get_bool_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_bool(conn=asyncmy_conn, id_=0, tinyint1_test=False)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_decimal", depends=["AsyncmyTestDataclassFunctions::get_bool_none"])
    async def test_get_decimal(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_decimal(conn=asyncmy_conn, id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None

        assert isinstance(result, decimal.Decimal)
        # decimal(12,4) always comes back padded to scale.
        assert str(result) == "12.3400"
        assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_decimal_none", depends=["AsyncmyTestDataclassFunctions::get_decimal"])
    async def test_get_decimal_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_decimal(conn=asyncmy_conn, id_=0, decimal_test=model.decimal_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_blob", depends=["AsyncmyTestDataclassFunctions::get_decimal_none"])
    async def test_get_blob(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_blob(conn=asyncmy_conn, id_=model.id_, blob_test=model.blob_test)

        assert result is not None

        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_blob_none", depends=["AsyncmyTestDataclassFunctions::get_blob"])
    async def test_get_blob_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_blob(conn=asyncmy_conn, id_=0, blob_test=memoryview(b"test"))

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_bit", depends=["AsyncmyTestDataclassFunctions::get_blob_none"])
    async def test_get_bit(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_bit(conn=asyncmy_conn, id_=model.id_)

        assert result is not None

        # bit(8) arrives as a single byte of raw bits.
        assert isinstance(result, memoryview)
        assert len(result) == 1
        assert bytes(result) == b"\x80"

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_bit_none", depends=["AsyncmyTestDataclassFunctions::get_bit"])
    async def test_get_bit_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_bit(conn=asyncmy_conn, id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_year", depends=["AsyncmyTestDataclassFunctions::get_bit_none"])
    async def test_get_year(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_year(conn=asyncmy_conn, id_=model.id_)

        assert result is not None

        assert isinstance(result, int)
        assert result == model.year_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_year_none", depends=["AsyncmyTestDataclassFunctions::get_year"])
    async def test_get_year_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_year(conn=asyncmy_conn, id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_json", depends=["AsyncmyTestDataclassFunctions::get_year_none"])
    async def test_get_json(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_json(conn=asyncmy_conn, id_=model.id_)

        assert result is not None

        assert isinstance(result, str)
        assert json.loads(result) == json.loads(model.json_test)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_json_none", depends=["AsyncmyTestDataclassFunctions::get_json"])
    async def test_get_json_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_json(conn=asyncmy_conn, id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_mood", depends=["AsyncmyTestDataclassFunctions::get_json_none"])
    async def test_get_mood(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_mood(conn=asyncmy_conn, id_=model.id_, mood=model.mood)

        assert result is not None

        assert isinstance(result, enums.TestMysqlTypesMood)
        assert result is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_mood_none", depends=["AsyncmyTestDataclassFunctions::get_mood"])
    async def test_get_mood_none(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries.get_one_mood(conn=asyncmy_conn, id_=0, mood=model.mood)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_tag", depends=["AsyncmyTestDataclassFunctions::get_mood_none"])
    async def test_get_tag(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_tag(conn=asyncmy_conn, id_=MODEL_ID)

        assert result is not None

        assert isinstance(result, enums.TestMysqlTypesTag)
        assert result is enums.TestMysqlTypesTag.ALPHA

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_tag_none", depends=["AsyncmyTestDataclassFunctions::get_tag"])
    async def test_get_tag_none(
        self,
        asyncmy_conn: asyncmy.Connection,
    ) -> None:
        result = await queries.get_one_tag(conn=asyncmy_conn, id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::value_normalization", depends=["AsyncmyTestDataclassFunctions::get_tag_none"])
    async def test_value_normalization(
        self,
        asyncmy_conn: asyncmy.Connection,
        model: models.TestMysqlType,
    ) -> None:
        # decimal(12,4) pads to scale, char(10) strips trailing spaces and
        # binary(16) is right-padded with NUL bytes on return.
        await queries.insert_one_mysql_type(
            conn=asyncmy_conn,
            id_=NORMALIZATION_ID,
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
            decimal_test=decimal.Decimal("12.34"),
            numeric_test=decimal.Decimal("3.5"),
            char_test="AB   ",
            varchar_test=model.varchar_test,
            tinytext_test=model.tinytext_test,
            text_test=model.text_test,
            mediumtext_test=model.mediumtext_test,
            longtext_test=model.longtext_test,
            binary_test=memoryview(b"abc"),
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
        result = await queries.get_one_mysql_type(conn=asyncmy_conn, id_=NORMALIZATION_ID)
        assert result is not None
        assert str(result.decimal_test) == "12.3400"
        assert str(result.numeric_test) == "3.50"
        assert result.char_test == "AB"
        assert bytes(result.binary_test) == b"abc" + b"\x00" * 13
        await queries.delete_one_mysql_type(conn=asyncmy_conn, id_=NORMALIZATION_ID)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_many", depends=["AsyncmyTestDataclassFunctions::value_normalization"])
    async def test_get_many(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_mysql_type(conn=asyncmy_conn, id_=model.id_)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], models.TestMysqlType)

        assert dataclasses.replace(results[0], json_test=model.json_test) == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_many_iter", depends=["AsyncmyTestDataclassFunctions::get_many"])
    async def test_get_many_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_mysql_type(conn=asyncmy_conn, id_=model.id_):
            assert result is not None
            assert isinstance(result, models.TestMysqlType)

            assert dataclasses.replace(result, json_test=model.json_test) == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_many_inner", depends=["AsyncmyTestDataclassFunctions::get_many_iter"])
    async def test_get_many_inner(self, asyncmy_conn: asyncmy.Connection, inner_model: models.TestInnerMysqlType) -> None:
        result = queries.get_many_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], models.TestInnerMysqlType)

        assert results[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_many_inner_iter",
        depends=["AsyncmyTestDataclassFunctions::get_many_inner"],
    )
    async def test_get_many_inner_iter(self, asyncmy_conn: asyncmy.Connection, inner_model: models.TestInnerMysqlType) -> None:
        async for result in queries.get_many_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_many_nullable_inner",
        depends=["AsyncmyTestDataclassFunctions::get_many_inner_iter"],
    )
    async def test_get_many_nullable_inner(self, asyncmy_conn: asyncmy.Connection, inner_model: models.TestInnerMysqlType) -> None:
        # int_test is None; the <=> in the query is NULL-safe equality.
        result = queries.get_many_nullable_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id, int_test=inner_model.int_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], models.TestInnerMysqlType)

        assert results[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_many_nullable_inner_iter",
        depends=["AsyncmyTestDataclassFunctions::get_many_nullable_inner"],
    )
    async def test_get_many_nullable_inner_iter(self, asyncmy_conn: asyncmy.Connection, inner_model: models.TestInnerMysqlType) -> None:
        async for result in queries.get_many_nullable_inner_mysql_type(conn=asyncmy_conn, table_id=inner_model.table_id, int_test=inner_model.int_test):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_many_date",
        depends=["AsyncmyTestDataclassFunctions::get_many_nullable_inner_iter"],
    )
    async def test_get_many_date(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_date(conn=asyncmy_conn, id_=model.id_, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], datetime.date)

        assert results[0] == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_many_date_iter", depends=["AsyncmyTestDataclassFunctions::get_many_date"])
    async def test_get_many_date_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_date(conn=asyncmy_conn, id_=model.id_, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)

            assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_many_time", depends=["AsyncmyTestDataclassFunctions::get_many_date_iter"])
    async def test_get_many_time(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_time(conn=asyncmy_conn, id_=model.id_, time_test=model.time_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], datetime.timedelta)

        assert results[0] == model.time_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_many_time_iter", depends=["AsyncmyTestDataclassFunctions::get_many_time"])
    async def test_get_many_time_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_time(conn=asyncmy_conn, id_=model.id_, time_test=model.time_test):
            assert result is not None
            assert isinstance(result, datetime.timedelta)

            assert result == model.time_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_many_bool", depends=["AsyncmyTestDataclassFunctions::get_many_time_iter"])
    async def test_get_many_bool(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_bool(conn=asyncmy_conn, id_=model.id_, tinyint1_test=model.tinyint1_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], bool)

        assert results[0] is True

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_many_bool_iter", depends=["AsyncmyTestDataclassFunctions::get_many_bool"])
    async def test_get_many_bool_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_bool(conn=asyncmy_conn, id_=model.id_, tinyint1_test=model.tinyint1_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result is True

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_many_decimal",
        depends=["AsyncmyTestDataclassFunctions::get_many_bool_iter"],
    )
    async def test_get_many_decimal(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_decimal(conn=asyncmy_conn, id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], decimal.Decimal)

        assert results[0] == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_many_decimal_iter",
        depends=["AsyncmyTestDataclassFunctions::get_many_decimal"],
    )
    async def test_get_many_decimal_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_decimal(conn=asyncmy_conn, id_=model.id_, decimal_test=model.decimal_test):
            assert result is not None
            assert isinstance(result, decimal.Decimal)

            assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_many_mood",
        depends=["AsyncmyTestDataclassFunctions::get_many_decimal_iter"],
    )
    async def test_get_many_mood(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_mood(conn=asyncmy_conn, mood=model.mood)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], enums.TestMysqlTypesMood)

        assert results[0] is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_many_mood_iter", depends=["AsyncmyTestDataclassFunctions::get_many_mood"])
    async def test_get_many_mood_iter(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        async for result in queries.get_many_mood(conn=asyncmy_conn, mood=model.mood):
            assert result is not None
            assert isinstance(result, enums.TestMysqlTypesMood)

            assert result is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::list_months", depends=["AsyncmyTestDataclassFunctions::get_many_mood_iter"])
    async def test_list_months(self, asyncmy_conn: asyncmy.Connection) -> None:
        # Parameterless :many with literal percents in the SQL; regression
        # test for the percent-doubling bug.
        result = queries.list_months(conn=asyncmy_conn)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert results == ["2026-01"]

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::list_months_iter", depends=["AsyncmyTestDataclassFunctions::list_months"])
    async def test_list_months_iter(self, asyncmy_conn: asyncmy.Connection) -> None:
        months = [month async for month in queries.list_months(conn=asyncmy_conn)]
        assert months == ["2026-01"]

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::count", depends=["AsyncmyTestDataclassFunctions::list_months_iter"])
    async def test_count(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = await queries.count_mysql_types(conn=asyncmy_conn)

        assert result == 1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::update_varchar_rows", depends=["AsyncmyTestDataclassFunctions::count"])
    async def test_update_varchar_rows(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        result = await queries.update_varchar_test(conn=asyncmy_conn, varchar_test="updated varchar", id_=model.id_)
        assert isinstance(result, int)
        assert result == 1

        result = await queries.update_varchar_test(conn=asyncmy_conn, varchar_test="updated varchar", id_=0)
        assert result == 0

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::all_types_cursor", depends=["AsyncmyTestDataclassFunctions::update_varchar_rows"])
    async def test_all_types_cursor(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        cursor = await queries.all_mysql_types_cursor(conn=asyncmy_conn)
        assert isinstance(cursor, asyncmy.cursors.Cursor)

        rows = await cursor.fetchall()
        await cursor.close()
        assert len(rows) == 1
        assert rows[0][0] == model.id_

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::insert_exec_last_id", depends=["AsyncmyTestDataclassFunctions::all_types_cursor"])
    async def test_insert_exec_last_id(self, asyncmy_conn: asyncmy.Connection) -> None:
        # AUTO_INCREMENT counters persist across runs, so only relative
        # assertions are safe.
        first_id = await queries.insert_exec_last_id(conn=asyncmy_conn, name="dataclass-functions-first")
        assert first_id is not None
        assert isinstance(first_id, int)
        assert first_id > 0

        name = await queries.get_exec_last_id_name(conn=asyncmy_conn, id_=first_id)
        assert name == "dataclass-functions-first"

        # A statement that inserts nothing has no last row id: the OK
        # packet's 0 maps to the documented None.
        assert await queries.touch_exec_last_id(conn=asyncmy_conn, name="untouched", id_=first_id + 1000000) is None

        second_id = await queries.insert_exec_last_id(conn=asyncmy_conn, name="dataclass-functions-second")
        assert second_id is not None
        assert second_id > first_id

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_exec_last_id_name_none",
        depends=["AsyncmyTestDataclassFunctions::insert_exec_last_id"],
    )
    async def test_get_exec_last_id_name_none(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = await queries.get_exec_last_id_name(conn=asyncmy_conn, id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::delete_mysql_type",
        depends=["AsyncmyTestDataclassFunctions::get_exec_last_id_name_none"],
    )
    async def test_delete_mysql_type(self, asyncmy_conn: asyncmy.Connection, model: models.TestMysqlType) -> None:
        await queries.delete_one_mysql_type(conn=asyncmy_conn, id_=model.id_)

        result = await queries.get_one_mysql_type(conn=asyncmy_conn, id_=model.id_)
        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::insert_type_override",
    )
    async def test_insert_type_override(self, asyncmy_conn: asyncmy.Connection, override_model: models.TestTypeOverride) -> None:
        await queries.insert_type_override(conn=asyncmy_conn, id_=override_model.id_, text_test=override_model.text_test)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_type_override",
        depends=["AsyncmyTestDataclassFunctions::insert_type_override"],
    )
    async def test_get_type_override(self, asyncmy_conn: asyncmy.Connection, override_model: models.TestTypeOverride) -> None:
        result = await queries.get_type_override(conn=asyncmy_conn, id_=override_model.id_)
        assert result is not None
        assert isinstance(result.text_test, UserString)
        assert result == override_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_type_override_none",
        depends=["AsyncmyTestDataclassFunctions::get_type_override"],
    )
    async def test_get_type_override_none(self, asyncmy_conn: asyncmy.Connection, override_model: models.TestTypeOverride) -> None:
        result = await queries.get_type_override(conn=asyncmy_conn, id_=override_model.id_ - 1)
        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::type_override_null_value",
        depends=["AsyncmyTestDataclassFunctions::get_type_override_none"],
    )
    async def test_type_override_null_value(self, asyncmy_conn: asyncmy.Connection) -> None:
        # The UserString override sits on a nullable column.
        await queries.insert_type_override(conn=asyncmy_conn, id_=OVERRIDE_NULL_ID, text_test=None)

        result = await queries.get_type_override(conn=asyncmy_conn, id_=OVERRIDE_NULL_ID)
        assert result is not None
        assert result.text_test is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::insert_reserved_arg")
    async def test_insert_reserved_arg(self, asyncmy_conn: asyncmy.Connection) -> None:
        # The column is literally named "conn"; the generated parameter must
        # be deduplicated against the implicit connection argument.
        await queries.insert_reserved_arg(conn=asyncmy_conn, id_=RESERVED_ARG_ID, conn_2=RESERVED_ARG_VALUE)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassFunctions::get_reserved_arg",
        depends=["AsyncmyTestDataclassFunctions::insert_reserved_arg"],
    )
    async def test_get_reserved_arg(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = await queries.get_reserved_arg(conn=asyncmy_conn, conn_2=RESERVED_ARG_VALUE)
        assert result == models.TestReservedArg(id_=RESERVED_ARG_ID, conn=RESERVED_ARG_VALUE)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["AsyncmyTestDataclassFunctions::insert_reserved_arg"])
    async def test_get_reserved_arg_not_found(self, asyncmy_conn: asyncmy.Connection) -> None:
        assert await queries.get_reserved_arg(conn=asyncmy_conn, conn_2="missing-reserved-arg-value") is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::insert_slice_rows")
    async def test_insert_slice_rows(self, asyncmy_conn: asyncmy.Connection) -> None:
        for offset, (name, note) in enumerate((("a", "x"), ("b", "y"), ("c", None), ("b", "y"))):
            await queries_slice.insert_slice_row(conn=asyncmy_conn, id_=SLICE_ID_BASE + offset, name=name, note=note)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_slice_rows", depends=["AsyncmyTestDataclassFunctions::insert_slice_rows"])
    async def test_get_slice_rows(self, asyncmy_conn: asyncmy.Connection) -> None:
        result = queries_slice.get_slice_rows(conn=asyncmy_conn, ids=[SLICE_ID_BASE, SLICE_ID_BASE + 2])
        assert isinstance(result, queries_slice.QueryResults)
        rows = await result
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a", note="x"),
            models.TestSlice(id_=SLICE_ID_BASE + 2, name="c", note=None),
        ]
        assert [row async for row in result] == rows

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_slice_rows_empty_slice", depends=["AsyncmyTestDataclassFunctions::insert_slice_rows"])
    async def test_get_slice_rows_empty_slice(self, asyncmy_conn: asyncmy.Connection) -> None:
        # An empty sequence expands the placeholder to NULL: IN (NULL)
        # matches no rows instead of raising.
        assert await queries_slice.get_slice_rows(conn=asyncmy_conn, ids=[]) == []

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_slice_row_filtered", depends=["AsyncmyTestDataclassFunctions::insert_slice_rows"])
    async def test_get_slice_row_filtered(self, asyncmy_conn: asyncmy.Connection) -> None:
        # Plain params surround the slice, so this proves the flattened
        # argument tuple binds in SQL text order.
        row = await queries_slice.get_slice_row_filtered(
            conn=asyncmy_conn,
            name="b",
            ids=[SLICE_ID_BASE + 1, SLICE_ID_BASE + 3],
            id_=SLICE_ID_BASE + 1,
        )
        assert row == models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y")

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_slice_row_filtered_not_found", depends=["AsyncmyTestDataclassFunctions::insert_slice_rows"])
    async def test_get_slice_row_filtered_not_found(self, asyncmy_conn: asyncmy.Connection) -> None:
        assert await queries_slice.get_slice_row_filtered(conn=asyncmy_conn, name="a", ids=[SLICE_ID_BASE], id_=SLICE_ID_BASE) is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_slice_rows_by_notes", depends=["AsyncmyTestDataclassFunctions::insert_slice_rows"])
    async def test_get_slice_rows_by_notes(self, asyncmy_conn: asyncmy.Connection) -> None:
        # The slice targets a nullable column; the parameter is still a plain
        # Sequence, and rows whose note is NULL never match.
        rows = await queries_slice.get_slice_rows_by_notes(conn=asyncmy_conn, notes=["y"])
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE + 1, name="b", note="y"),
            models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y"),
        ]
        assert await queries_slice.get_slice_rows_by_notes(conn=asyncmy_conn, notes=[]) == []

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_slice_rows_by_name_or_note", depends=["AsyncmyTestDataclassFunctions::insert_slice_rows"])
    async def test_get_slice_rows_by_name_or_note(self, asyncmy_conn: asyncmy.Connection) -> None:
        # The same slice name is used twice, so every marker occurrence is
        # expanded and the sequence is bound once per occurrence.
        rows = await queries_slice.get_slice_rows_by_name_or_note(conn=asyncmy_conn, names=["b", "x"])
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a", note="x"),
            models.TestSlice(id_=SLICE_ID_BASE + 1, name="b", note="y"),
            models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y"),
        ]
        assert await queries_slice.get_slice_rows_by_name_or_note(conn=asyncmy_conn, names=[]) == []

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_slice_rows_by_name_or_note_filtered", depends=["AsyncmyTestDataclassFunctions::insert_slice_rows"])
    async def test_get_slice_rows_by_name_or_note_filtered(self, asyncmy_conn: asyncmy.Connection) -> None:
        # A plain parameter sits between the two uses of the slice, so this
        # proves the flattened arguments follow SQL text order.
        rows = await queries_slice.get_slice_rows_by_name_or_note_filtered(conn=asyncmy_conn, names=["b", "x"], id_=SLICE_ID_BASE + 1)
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a", note="x"),
            models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y"),
        ]
        assert await queries_slice.get_slice_rows_by_name_or_note_filtered(conn=asyncmy_conn, names=[], id_=SLICE_ID_BASE) == []

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassFunctions::get_first_slice_name_two_slices", depends=["AsyncmyTestDataclassFunctions::insert_slice_rows"])
    async def test_get_first_slice_name_two_slices(self, asyncmy_conn: asyncmy.Connection) -> None:
        name = await queries_slice.get_first_slice_name(conn=asyncmy_conn, ids=[SLICE_ID_BASE + 1], names=["a"])
        assert name == "a"
        assert await queries_slice.get_first_slice_name(conn=asyncmy_conn, ids=[], names=[]) is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        depends=[
            "AsyncmyTestDataclassFunctions::get_slice_rows",
            "AsyncmyTestDataclassFunctions::get_slice_rows_empty_slice",
            "AsyncmyTestDataclassFunctions::get_slice_row_filtered",
            "AsyncmyTestDataclassFunctions::get_slice_row_filtered_not_found",
            "AsyncmyTestDataclassFunctions::get_slice_rows_by_notes",
            "AsyncmyTestDataclassFunctions::get_slice_rows_by_name_or_note",
            "AsyncmyTestDataclassFunctions::get_slice_rows_by_name_or_note_filtered",
            "AsyncmyTestDataclassFunctions::get_first_slice_name_two_slices",
        ]
    )
    async def test_delete_slice_rows(self, asyncmy_conn: asyncmy.Connection) -> None:
        assert await queries_slice.delete_slice_rows(conn=asyncmy_conn, ids=[]) == 0
        deleted = await queries_slice.delete_slice_rows(conn=asyncmy_conn, ids=[SLICE_ID_BASE + offset for offset in range(SLICE_ROW_COUNT)])
        assert deleted == SLICE_ROW_COUNT
