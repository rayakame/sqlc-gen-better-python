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
import typing
from collections import UserString

import asyncmy
import asyncmy.cursors
import pytest
import pytest_asyncio

from test.driver_asyncmy import no_row_conn
from test.driver_asyncmy.dataclass.classes import enums
from test.driver_asyncmy.dataclass.classes import models
from test.driver_asyncmy.dataclass.classes import queries

MODEL_ID = 6000
NORMALIZATION_ID = 6001
OVERRIDE_ID = 6100
OVERRIDE_NULL_ID = 6101
RESERVED_ARG_ID = 6150
RESERVED_ARG_VALUE = "dataclass-classes-conn"


@pytest.mark.asyncio(loop_scope="session")
class TestAsyncmyDataclassClasses:
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

    @pytest_asyncio.fixture(scope="class", loop_scope="session")
    async def queries_obj(self, asyncmy_conn: asyncmy.Connection) -> queries.Queries:
        return queries.Queries(conn=asyncmy_conn)

    @pytest.mark.asyncio(loop_scope="session")
    async def test_conn_attr(self, queries_obj: queries.Queries) -> None:
        assert isinstance(queries_obj.conn, asyncmy.Connection)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::insert")
    async def test_insert(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        await queries_obj.insert_one_mysql_type(
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
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::inner_insert", depends=["AsyncmyTestDataclassClasses::insert"])
    async def test_inner_insert(
        self,
        queries_obj: queries.Queries,
        inner_model: models.TestInnerMysqlType,
    ) -> None:
        await queries_obj.insert_one_inner_mysql_type(
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
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_one", depends=["AsyncmyTestDataclassClasses::inner_insert"])
    async def test_get_one(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_mysql_type(id_=model.id_)

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
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_one_none", depends=["AsyncmyTestDataclassClasses::get_one"])
    async def test_get_one_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_mysql_type(id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_one_inner", depends=["AsyncmyTestDataclassClasses::get_one_none"])
    async def test_get_one_inner(
        self,
        queries_obj: queries.Queries,
        inner_model: models.TestInnerMysqlType,
    ) -> None:
        result = await queries_obj.get_one_inner_mysql_type(table_id=inner_model.table_id)

        assert result is not None

        assert isinstance(result, models.TestInnerMysqlType)
        assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_one_inner_none", depends=["AsyncmyTestDataclassClasses::get_one_inner"])
    async def test_get_one_inner_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_inner_mysql_type(table_id=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_date", depends=["AsyncmyTestDataclassClasses::get_one_inner_none"])
    async def test_get_date(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_date(id_=model.id_, date_test=model.date_test)

        assert result is not None

        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_date_none", depends=["AsyncmyTestDataclassClasses::get_date"])
    async def test_get_date_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_date(id_=0, date_test=model.date_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_datetime", depends=["AsyncmyTestDataclassClasses::get_date_none"])
    async def test_get_datetime(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_datetime(id_=model.id_, datetime_test=model.datetime_test)

        assert result is not None

        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_datetime_none", depends=["AsyncmyTestDataclassClasses::get_datetime"])
    async def test_get_datetime_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_datetime(id_=0, datetime_test=model.datetime_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_time", depends=["AsyncmyTestDataclassClasses::get_datetime_none"])
    async def test_get_time(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_time(id_=model.id_, time_test=model.time_test)

        assert result is not None

        # MySQL time columns arrive as timedelta, not datetime.time.
        assert isinstance(result, datetime.timedelta)
        assert result == model.time_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_time_none", depends=["AsyncmyTestDataclassClasses::get_time"])
    async def test_get_time_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_time(id_=0, time_test=model.time_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_bool", depends=["AsyncmyTestDataclassClasses::get_time_none"])
    async def test_get_bool(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_bool(id_=model.id_, tinyint1_test=model.tinyint1_test)

        assert result is not None

        assert isinstance(result, bool)
        assert result is True

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_bool_none", depends=["AsyncmyTestDataclassClasses::get_bool"])
    async def test_get_bool_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_bool(id_=0, tinyint1_test=False)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_decimal", depends=["AsyncmyTestDataclassClasses::get_bool_none"])
    async def test_get_decimal(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_decimal(id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None

        assert isinstance(result, decimal.Decimal)
        # decimal(12,4) always comes back padded to scale.
        assert str(result) == "12.3400"
        assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_decimal_none", depends=["AsyncmyTestDataclassClasses::get_decimal"])
    async def test_get_decimal_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_decimal(id_=0, decimal_test=model.decimal_test)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_blob", depends=["AsyncmyTestDataclassClasses::get_decimal_none"])
    async def test_get_blob(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_blob(id_=model.id_, blob_test=model.blob_test)

        assert result is not None

        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_blob_none", depends=["AsyncmyTestDataclassClasses::get_blob"])
    async def test_get_blob_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_blob(id_=0, blob_test=memoryview(b"test"))

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_bit", depends=["AsyncmyTestDataclassClasses::get_blob_none"])
    async def test_get_bit(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_bit(id_=model.id_)

        assert result is not None

        # bit(8) arrives as a single byte of raw bits.
        assert isinstance(result, memoryview)
        assert len(result) == 1
        assert bytes(result) == b"\x80"

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_bit_none", depends=["AsyncmyTestDataclassClasses::get_bit"])
    async def test_get_bit_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_bit(id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_year", depends=["AsyncmyTestDataclassClasses::get_bit_none"])
    async def test_get_year(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_year(id_=model.id_)

        assert result is not None

        assert isinstance(result, int)
        assert result == model.year_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_year_none", depends=["AsyncmyTestDataclassClasses::get_year"])
    async def test_get_year_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_year(id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_json", depends=["AsyncmyTestDataclassClasses::get_year_none"])
    async def test_get_json(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_json(id_=model.id_)

        assert result is not None

        assert isinstance(result, str)
        assert json.loads(result) == json.loads(model.json_test)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_json_none", depends=["AsyncmyTestDataclassClasses::get_json"])
    async def test_get_json_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_json(id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_mood", depends=["AsyncmyTestDataclassClasses::get_json_none"])
    async def test_get_mood(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_mood(id_=model.id_, mood=model.mood)

        assert result is not None

        assert isinstance(result, enums.TestMysqlTypesMood)
        assert result is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_mood_none", depends=["AsyncmyTestDataclassClasses::get_mood"])
    async def test_get_mood_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = await queries_obj.get_one_mood(id_=0, mood=model.mood)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_tag", depends=["AsyncmyTestDataclassClasses::get_mood_none"])
    async def test_get_tag(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_tag(id_=MODEL_ID)

        assert result is not None

        assert isinstance(result, enums.TestMysqlTypesTag)
        assert result is enums.TestMysqlTypesTag.ALPHA

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_tag_none", depends=["AsyncmyTestDataclassClasses::get_tag"])
    async def test_get_tag_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = await queries_obj.get_one_tag(id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::value_normalization", depends=["AsyncmyTestDataclassClasses::get_tag_none"])
    async def test_value_normalization(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        # decimal(12,4) pads to scale, char(10) strips trailing spaces and
        # binary(16) is right-padded with NUL bytes on return.
        await queries_obj.insert_one_mysql_type(
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
        result = await queries_obj.get_one_mysql_type(id_=NORMALIZATION_ID)
        assert result is not None
        assert str(result.decimal_test) == "12.3400"
        assert str(result.numeric_test) == "3.50"
        assert result.char_test == "AB"
        assert bytes(result.binary_test) == b"abc" + b"\x00" * 13
        await queries_obj.delete_one_mysql_type(id_=NORMALIZATION_ID)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many", depends=["AsyncmyTestDataclassClasses::value_normalization"])
    async def test_get_many(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_mysql_type(id_=model.id_)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], models.TestMysqlType)

        assert dataclasses.replace(results[0], json_test=model.json_test) == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_iter", depends=["AsyncmyTestDataclassClasses::get_many"])
    async def test_get_many_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        async for result in queries_obj.get_many_mysql_type(id_=model.id_):
            assert result is not None
            assert isinstance(result, models.TestMysqlType)

            assert dataclasses.replace(result, json_test=model.json_test) == model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_inner", depends=["AsyncmyTestDataclassClasses::get_many_iter"])
    async def test_get_many_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        result = queries_obj.get_many_inner_mysql_type(table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], models.TestInnerMysqlType)

        assert results[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_inner_iter", depends=["AsyncmyTestDataclassClasses::get_many_inner"])
    async def test_get_many_inner_iter(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        async for result in queries_obj.get_many_inner_mysql_type(table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::get_many_nullable_inner",
        depends=["AsyncmyTestDataclassClasses::get_many_inner_iter"],
    )
    async def test_get_many_nullable_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        # int_test is None; the <=> in the query is NULL-safe equality.
        result = queries_obj.get_many_nullable_inner_mysql_type(table_id=inner_model.table_id, int_test=inner_model.int_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], models.TestInnerMysqlType)

        assert results[0] == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::get_many_nullable_inner_iter",
        depends=["AsyncmyTestDataclassClasses::get_many_nullable_inner"],
    )
    async def test_get_many_nullable_inner_iter(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        async for result in queries_obj.get_many_nullable_inner_mysql_type(table_id=inner_model.table_id, int_test=inner_model.int_test):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)

            assert result == inner_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::get_many_date",
        depends=["AsyncmyTestDataclassClasses::get_many_nullable_inner_iter"],
    )
    async def test_get_many_date(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_date(id_=model.id_, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], datetime.date)

        assert results[0] == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_date_iter", depends=["AsyncmyTestDataclassClasses::get_many_date"])
    async def test_get_many_date_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        async for result in queries_obj.get_many_date(id_=model.id_, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)

            assert result == model.date_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_time", depends=["AsyncmyTestDataclassClasses::get_many_date_iter"])
    async def test_get_many_time(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_time(id_=model.id_, time_test=model.time_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], datetime.timedelta)

        assert results[0] == model.time_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_time_iter", depends=["AsyncmyTestDataclassClasses::get_many_time"])
    async def test_get_many_time_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        async for result in queries_obj.get_many_time(id_=model.id_, time_test=model.time_test):
            assert result is not None
            assert isinstance(result, datetime.timedelta)

            assert result == model.time_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_bool", depends=["AsyncmyTestDataclassClasses::get_many_time_iter"])
    async def test_get_many_bool(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_bool(id_=model.id_, tinyint1_test=model.tinyint1_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], bool)

        assert results[0] is True

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_bool_iter", depends=["AsyncmyTestDataclassClasses::get_many_bool"])
    async def test_get_many_bool_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        async for result in queries_obj.get_many_bool(id_=model.id_, tinyint1_test=model.tinyint1_test):
            assert result is not None
            assert isinstance(result, bool)

            assert result is True

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_decimal", depends=["AsyncmyTestDataclassClasses::get_many_bool_iter"])
    async def test_get_many_decimal(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_decimal(id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert len(results) == 1
        assert isinstance(results[0], decimal.Decimal)

        assert results[0] == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::get_many_decimal_iter",
        depends=["AsyncmyTestDataclassClasses::get_many_decimal"],
    )
    async def test_get_many_decimal_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        async for result in queries_obj.get_many_decimal(id_=model.id_, decimal_test=model.decimal_test):
            assert result is not None
            assert isinstance(result, decimal.Decimal)

            assert result == model.decimal_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_mood", depends=["AsyncmyTestDataclassClasses::get_many_decimal_iter"])
    async def test_get_many_mood(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_mood(mood=model.mood)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        # The query is not id-filtered; other rows may exist, so assert containment.
        assert len(results) >= 1
        assert all(isinstance(mood, enums.TestMysqlTypesMood) for mood in results)

        assert all(mood is enums.TestMysqlTypesMood.VALUE_24H for mood in results)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::get_many_mood_iter", depends=["AsyncmyTestDataclassClasses::get_many_mood"])
    async def test_get_many_mood_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        async for result in queries_obj.get_many_mood(mood=model.mood):
            assert result is not None
            assert isinstance(result, enums.TestMysqlTypesMood)

            assert result is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::list_months", depends=["AsyncmyTestDataclassClasses::get_many_mood_iter"])
    async def test_list_months(self, queries_obj: queries.Queries) -> None:
        # Parameterless :many with literal percents in the SQL; regression
        # test for the percent-doubling bug.
        result = queries_obj.list_months()

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = await result
        assert results == ["2026-01"]

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::list_months_iter", depends=["AsyncmyTestDataclassClasses::list_months"])
    async def test_list_months_iter(self, queries_obj: queries.Queries) -> None:
        months = [month async for month in queries_obj.list_months()]
        assert months == ["2026-01"]

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::count", depends=["AsyncmyTestDataclassClasses::list_months_iter"])
    async def test_count(self, queries_obj: queries.Queries) -> None:
        result = await queries_obj.count_mysql_types()

        # The shared table may carry other files' rows; only a lower bound is safe.
        assert result is not None
        assert result >= 1

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::update_varchar_rows", depends=["AsyncmyTestDataclassClasses::count"])
    async def test_update_varchar_rows(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = await queries_obj.update_varchar_test(varchar_test="updated varchar", id_=model.id_)
        assert isinstance(result, int)
        assert result == 1

        result = await queries_obj.update_varchar_test(varchar_test="updated varchar", id_=0)
        assert result == 0

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::all_types_cursor", depends=["AsyncmyTestDataclassClasses::update_varchar_rows"])
    async def test_all_types_cursor(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        cursor = await queries_obj.all_mysql_types_cursor()
        assert isinstance(cursor, asyncmy.cursors.Cursor)

        rows = await cursor.fetchall()
        await cursor.close()
        # The shared table may carry other files' rows; assert on our own.
        assert model.id_ in {row[0] for row in rows}

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::insert_exec_last_id", depends=["AsyncmyTestDataclassClasses::all_types_cursor"])
    async def test_insert_exec_last_id(self, queries_obj: queries.Queries) -> None:
        # AUTO_INCREMENT counters persist across runs, so only relative
        # assertions are safe.
        first_id = await queries_obj.insert_exec_last_id(name="dataclass-classes-first")
        assert first_id is not None
        assert isinstance(first_id, int)
        assert first_id > 0

        name = await queries_obj.get_exec_last_id_name(id_=first_id)
        assert name == "dataclass-classes-first"

        second_id = await queries_obj.insert_exec_last_id(name="dataclass-classes-second")
        assert second_id is not None
        assert second_id > first_id

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::get_exec_last_id_name_none",
        depends=["AsyncmyTestDataclassClasses::insert_exec_last_id"],
    )
    async def test_get_exec_last_id_name_none(self, queries_obj: queries.Queries) -> None:
        result = await queries_obj.get_exec_last_id_name(id_=0)

        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::delete_mysql_type",
        depends=["AsyncmyTestDataclassClasses::get_exec_last_id_name_none"],
    )
    async def test_delete_mysql_type(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        await queries_obj.delete_one_mysql_type(id_=model.id_)

        result = await queries_obj.get_one_mysql_type(id_=model.id_)
        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::insert_type_override",
    )
    async def test_insert_type_override(self, queries_obj: queries.Queries, override_model: models.TestTypeOverride) -> None:
        await queries_obj.insert_type_override(id_=override_model.id_, text_test=override_model.text_test)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::get_type_override",
        depends=["AsyncmyTestDataclassClasses::insert_type_override"],
    )
    async def test_get_type_override(self, queries_obj: queries.Queries, override_model: models.TestTypeOverride) -> None:
        result = await queries_obj.get_type_override(id_=override_model.id_)
        assert result is not None
        assert isinstance(result.text_test, UserString)
        assert result == override_model

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::get_type_override_none",
        depends=["AsyncmyTestDataclassClasses::get_type_override"],
    )
    async def test_get_type_override_none(self, queries_obj: queries.Queries, override_model: models.TestTypeOverride) -> None:
        result = await queries_obj.get_type_override(id_=override_model.id_ - 1)
        assert result is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::type_override_null_value",
        depends=["AsyncmyTestDataclassClasses::get_type_override_none"],
    )
    async def test_type_override_null_value(self, queries_obj: queries.Queries) -> None:
        # The UserString override sits on a nullable column.
        await queries_obj.insert_type_override(id_=OVERRIDE_NULL_ID, text_test=None)

        result = await queries_obj.get_type_override(id_=OVERRIDE_NULL_ID)
        assert result is not None
        assert result.text_test is None

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="AsyncmyTestDataclassClasses::insert_reserved_arg")
    async def test_insert_reserved_arg(self, queries_obj: queries.Queries) -> None:
        # The column is literally named "conn"; on methods no deduplication
        # against an implicit connection argument is needed.
        await queries_obj.insert_reserved_arg(id_=RESERVED_ARG_ID, conn=RESERVED_ARG_VALUE)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(
        name="AsyncmyTestDataclassClasses::get_reserved_arg",
        depends=["AsyncmyTestDataclassClasses::insert_reserved_arg"],
    )
    async def test_get_reserved_arg(self, queries_obj: queries.Queries) -> None:
        result = await queries_obj.get_reserved_arg(conn=RESERVED_ARG_VALUE)
        assert result == models.TestReservedArg(id_=RESERVED_ARG_ID, conn=RESERVED_ARG_VALUE)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["AsyncmyTestDataclassClasses::insert_reserved_arg"])
    async def test_get_reserved_arg_not_found(self, queries_obj: queries.Queries) -> None:
        assert await queries_obj.get_reserved_arg(conn="missing-reserved-arg-value") is None

    @pytest.mark.asyncio(loop_scope="session")
    async def test_one_missing_rows_return_none(self, asyncmy_conn: asyncmy.Connection) -> None:
        # Every :one not-found branch, plus the no-insert :execlastid.
        obj = queries.Queries(conn=asyncmy_conn)
        assert await obj.get_one_mysql_type(id_=-1) is None
        assert await obj.get_one_inner_mysql_type(table_id=-1) is None
        assert await obj.get_one_date(id_=-1, date_test=datetime.date(1970, 1, 1)) is None
        assert await obj.get_one_datetime(id_=-1, datetime_test=datetime.datetime(1970, 1, 1)) is None
        assert await obj.get_one_time(id_=-1, time_test=datetime.timedelta()) is None
        assert await obj.get_one_bool(id_=-1, tinyint1_test=False) is None
        assert await obj.get_one_decimal(id_=-1, decimal_test=decimal.Decimal(0)) is None
        assert await obj.get_one_blob(id_=-1, blob_test=memoryview(b"")) is None
        assert await obj.get_one_bit(id_=-1) is None
        assert await obj.get_one_year(id_=-1) is None
        assert await obj.get_one_json(id_=-1) is None
        assert await obj.get_one_mood(id_=-1, mood=enums.TestMysqlTypesMood.SAD) is None
        assert await obj.get_one_tag(id_=-1) is None
        assert await obj.get_exec_last_id_name(id_=-1) is None
        assert await obj.get_type_override(id_=-1) is None
        assert await obj.get_reserved_arg(conn="missing") is None
        assert await obj.touch_exec_last_id(name="untouched", id_=-1) is None

        # count(*) always returns a row; its miss branch needs the stub.
        stub = typing.cast("asyncmy.Connection", no_row_conn.NoRowConn())
        assert await queries.Queries(conn=stub).count_mysql_types() is None
