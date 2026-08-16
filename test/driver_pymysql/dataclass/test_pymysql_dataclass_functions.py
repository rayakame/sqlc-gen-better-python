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

import pymysql
import pymysql.cursors
import pytest

from test.converters import Preferences
from test.driver_pymysql import no_row_conn
from test.driver_pymysql.dataclass.functions import enums
from test.driver_pymysql.dataclass.functions import models
from test.driver_pymysql.dataclass.functions import queries
from test.driver_pymysql.dataclass.functions import queries_backslash
from test.driver_pymysql.dataclass.functions import queries_case
from test.driver_pymysql.dataclass.functions import queries_converters
from test.driver_pymysql.dataclass.functions import queries_enum_override
from test.driver_pymysql.dataclass.functions import queries_field_namings
from test.driver_pymysql.dataclass.functions import queries_invalid_identifiers
from test.driver_pymysql.dataclass.functions import queries_slice

# Fixed ids: the MySQL tables are shared by every pymysql/asyncmy suite in the
# session, so each test file owns a distinct id range. This file: 1500-1999.
MAIN_ID = 1500
TYPE_OVERRIDE_ID = 1600
TYPE_OVERRIDE_NONE_ID = 1601
ENUM_OVERRIDE_ID = 1650
ENUM_OVERRIDE_ID_2 = 1651
CASE_ID = 1700
RESERVED_ARG_ID = 1750
FIELD_NAMING_ID = 1800
INVALID_IDENTIFIER_ID = 1850
THIRD_PARTY_ID = 1860
THIRD_PARTY_TOTAL = 9001
SLICE_ID_BASE = 1900
BACKSLASH_ID = 1950
SLICE_ROW_COUNT = 4
CONVERTER_ID = 1950
CONVERTER_ID_2 = 1951

CASE_DT = datetime.datetime(2026, 7, 19, 8, 15)
CASE_DEC = decimal.Decimal("12.34")
RESERVED_ARG_VALUE = "pymysql-dataclass-functions-conn"
EXEC_LAST_ID_NAME = "pymysql-dataclass-functions"
UPDATED_VARCHAR = "updated varchar"
# decimal(12,4) and numeric(10,2) come back padded to their full scale.
DECIMAL_PADDED = "1234.5000"
NUMERIC_PADDED = "87.60"
EXPECTED_MONTH = "2026-01"


class TestPymysqlDataclassFunctions:
    @pytest.fixture(scope="session")
    def override_model(self) -> models.TestTypeOverride:
        return models.TestTypeOverride(id_=TYPE_OVERRIDE_ID, text_test=UserString("Test"))

    @pytest.fixture(scope="session")
    def model(self) -> models.TestMysqlType:
        return models.TestMysqlType(
            id_=MAIN_ID,
            int_test=42,
            integer_test=43,
            mediumint_test=8_388_607,
            smallint_test=32_767,
            tinyint_test=127,
            bigint_test=9_007_199_254_740_991,
            int_unsigned_test=4_294_967_295,
            bigint_unsigned_test=2**63 + 10,
            year_test=2026,
            tinyint1_test=True,
            bool_test=True,
            boolean_test=False,
            float_test=2.5,
            double_test=math.e,
            double_precision_test=1.41421,
            real_test=math.pi,
            decimal_test=decimal.Decimal("1234.5"),
            numeric_test=decimal.Decimal("87.6"),
            char_test="ABC",
            varchar_test="Hello varchar",
            tinytext_test="tiny text",
            text_test="Some text",
            mediumtext_test="medium text",
            longtext_test="long text",
            binary_test=memoryview(b"bin-test".ljust(16, b"\x00")),
            varbinary_test=memoryview(b"\x00\x01\x02hello"),
            tinyblob_test=memoryview(b"tiny blob"),
            blob_test=memoryview(b"\x00\x01\x02blob"),
            mediumblob_test=memoryview(b"medium blob"),
            longblob_test=memoryview(b"long blob"),
            bit_test=memoryview(b"\x80"),
            date_test=datetime.date(2026, 1, 1),
            datetime_test=datetime.datetime(2026, 1, 15, 12, 30, 45),
            datetime6_test=datetime.datetime(2026, 1, 15, 12, 30, 45, 123456),
            timestamp_test=datetime.datetime(2026, 1, 15, 6, 30, 45),
            time_test=datetime.timedelta(hours=13, minutes=14, seconds=15),
            json_test=json.dumps({"foo": "bar"}),
            mood=enums.TestMysqlTypesMood.VALUE_24H,
            tag=enums.TestMysqlTypesTag.BETA,
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
            year_test=None,
            tinyint1_test=None,
            bool_test=None,
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
            binary_test=None,
            varbinary_test=model.varbinary_test,
            tinyblob_test=None,
            blob_test=None,
            mediumblob_test=None,
            longblob_test=model.longblob_test,
            bit_test=None,
            date_test=None,
            datetime_test=None,
            datetime6_test=None,
            timestamp_test=None,
            time_test=model.time_test,
            json_test=None,
            mood=None,
            tag=enums.TestInnerMysqlTypesTag.ALPHA,
        )

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert")
    def test_insert(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        queries.insert_one_mysql_type(
            conn=pymysql_conn,
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

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::inner_insert", depends=["PymysqlTestDataclassFunctions::insert"])
    def test_inner_insert(
        self,
        pymysql_conn: pymysql.Connection,
        inner_model: models.TestInnerMysqlType,
    ) -> None:
        queries.insert_one_inner_mysql_type(
            conn=pymysql_conn,
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

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_one", depends=["PymysqlTestDataclassFunctions::inner_insert"])
    def test_get_one(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_mysql_type(conn=pymysql_conn, id_=model.id_)

        assert result is not None
        assert isinstance(result, models.TestMysqlType)

        # MySQL pads decimals to their declared scale and binary(16) to full
        # width, keeps datetime(6) microseconds, and normalizes json spacing.
        assert str(result.decimal_test) == DECIMAL_PADDED
        assert str(result.numeric_test) == NUMERIC_PADDED
        assert bytes(result.binary_test) == b"bin-test".ljust(16, b"\x00")
        assert bytes(result.bit_test) == b"\x80"
        assert result.tinyint1_test is True
        assert result.bool_test is True
        assert result.boolean_test is False
        assert isinstance(result.time_test, datetime.timedelta)
        assert result.datetime6_test == model.datetime6_test
        assert json.loads(result.json_test) == json.loads(model.json_test)
        assert dataclasses.replace(result, json_test=model.json_test) == model

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_one_none", depends=["PymysqlTestDataclassFunctions::get_one"])
    def test_get_one_none(
        self,
        pymysql_conn: pymysql.Connection,
    ) -> None:
        result = queries.get_one_mysql_type(conn=pymysql_conn, id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_one_inner", depends=["PymysqlTestDataclassFunctions::get_one_none"])
    def test_get_one_inner(
        self,
        pymysql_conn: pymysql.Connection,
        inner_model: models.TestInnerMysqlType,
    ) -> None:
        result = queries.get_one_inner_mysql_type(conn=pymysql_conn, table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, models.TestInnerMysqlType)
        assert result == inner_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_one_inner_none",
        depends=["PymysqlTestDataclassFunctions::get_one_inner"],
    )
    def test_get_one_inner_none(
        self,
        pymysql_conn: pymysql.Connection,
    ) -> None:
        result = queries.get_one_inner_mysql_type(conn=pymysql_conn, table_id=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_date", depends=["PymysqlTestDataclassFunctions::get_one_inner_none"])
    def test_get_date(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_date(conn=pymysql_conn, id_=model.id_, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_date_none", depends=["PymysqlTestDataclassFunctions::get_date"])
    def test_get_date_none(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_date(conn=pymysql_conn, id_=0, date_test=model.date_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_datetime", depends=["PymysqlTestDataclassFunctions::get_date_none"])
    def test_get_datetime(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_datetime(conn=pymysql_conn, id_=model.id_, datetime_test=model.datetime_test)

        assert result is not None
        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_datetime_none", depends=["PymysqlTestDataclassFunctions::get_datetime"])
    def test_get_datetime_none(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_datetime(conn=pymysql_conn, id_=0, datetime_test=model.datetime_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_time", depends=["PymysqlTestDataclassFunctions::get_datetime_none"])
    def test_get_time(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_time(conn=pymysql_conn, id_=model.id_, time_test=model.time_test)

        assert result is not None
        assert isinstance(result, datetime.timedelta)
        assert result == model.time_test

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_time_none", depends=["PymysqlTestDataclassFunctions::get_time"])
    def test_get_time_none(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_time(conn=pymysql_conn, id_=0, time_test=model.time_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_bool", depends=["PymysqlTestDataclassFunctions::get_time_none"])
    def test_get_bool(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_bool(conn=pymysql_conn, id_=model.id_, tinyint1_test=model.tinyint1_test)

        assert result is not None
        assert isinstance(result, bool)
        assert result is True

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_bool_none", depends=["PymysqlTestDataclassFunctions::get_bool"])
    def test_get_bool_none(
        self,
        pymysql_conn: pymysql.Connection,
    ) -> None:
        result = queries.get_one_bool(conn=pymysql_conn, id_=0, tinyint1_test=False)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_decimal", depends=["PymysqlTestDataclassFunctions::get_bool_none"])
    def test_get_decimal(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_decimal(conn=pymysql_conn, id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, decimal.Decimal)
        assert result == model.decimal_test
        assert str(result) == DECIMAL_PADDED

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_decimal_none", depends=["PymysqlTestDataclassFunctions::get_decimal"])
    def test_get_decimal_none(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_decimal(conn=pymysql_conn, id_=0, decimal_test=model.decimal_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_blob", depends=["PymysqlTestDataclassFunctions::get_decimal_none"])
    def test_get_blob(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_blob(conn=pymysql_conn, id_=model.id_, blob_test=model.blob_test)

        assert result is not None
        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_blob_none", depends=["PymysqlTestDataclassFunctions::get_blob"])
    def test_get_blob_none(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_blob(conn=pymysql_conn, id_=0, blob_test=model.blob_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_bit", depends=["PymysqlTestDataclassFunctions::get_blob_none"])
    def test_get_bit(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_bit(conn=pymysql_conn, id_=model.id_)

        assert result is not None
        assert isinstance(result, memoryview)
        assert bytes(result) == b"\x80"

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_bit_none", depends=["PymysqlTestDataclassFunctions::get_bit"])
    def test_get_bit_none(
        self,
        pymysql_conn: pymysql.Connection,
    ) -> None:
        result = queries.get_one_bit(conn=pymysql_conn, id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_year", depends=["PymysqlTestDataclassFunctions::get_bit_none"])
    def test_get_year(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_year(conn=pymysql_conn, id_=model.id_)

        assert result is not None
        assert isinstance(result, int)
        assert result == model.year_test

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_year_none", depends=["PymysqlTestDataclassFunctions::get_year"])
    def test_get_year_none(
        self,
        pymysql_conn: pymysql.Connection,
    ) -> None:
        result = queries.get_one_year(conn=pymysql_conn, id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_json", depends=["PymysqlTestDataclassFunctions::get_year_none"])
    def test_get_json(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_json(conn=pymysql_conn, id_=model.id_)

        assert result is not None
        assert isinstance(result, str)
        # MySQL normalizes json spacing; never compare the raw strings.
        assert json.loads(result) == json.loads(model.json_test)

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_json_none", depends=["PymysqlTestDataclassFunctions::get_json"])
    def test_get_json_none(
        self,
        pymysql_conn: pymysql.Connection,
    ) -> None:
        result = queries.get_one_json(conn=pymysql_conn, id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_mood", depends=["PymysqlTestDataclassFunctions::get_json_none"])
    def test_get_mood(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_mood(conn=pymysql_conn, id_=model.id_, mood=model.mood)

        assert result is not None
        assert isinstance(result, enums.TestMysqlTypesMood)
        assert result is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_mood_none", depends=["PymysqlTestDataclassFunctions::get_mood"])
    def test_get_mood_none(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_mood(conn=pymysql_conn, id_=0, mood=model.mood)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_tag", depends=["PymysqlTestDataclassFunctions::get_mood_none"])
    def test_get_tag(
        self,
        pymysql_conn: pymysql.Connection,
        model: models.TestMysqlType,
    ) -> None:
        result = queries.get_one_tag(conn=pymysql_conn, id_=model.id_)

        assert result is not None
        assert isinstance(result, enums.TestMysqlTypesTag)
        assert result is enums.TestMysqlTypesTag.BETA
        assert result == model.tag

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_tag_none", depends=["PymysqlTestDataclassFunctions::get_tag"])
    def test_get_tag_none(
        self,
        pymysql_conn: pymysql.Connection,
    ) -> None:
        result = queries.get_one_tag(conn=pymysql_conn, id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_many", depends=["PymysqlTestDataclassFunctions::get_tag_none"])
    def test_get_many(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_mysql_type(conn=pymysql_conn, id_=model.id_)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert len(results) == 1
        assert isinstance(results[0], models.TestMysqlType)
        assert json.loads(results[0].json_test) == json.loads(model.json_test)
        assert dataclasses.replace(results[0], json_test=model.json_test) == model

        results = result()
        assert len(results) == 1
        assert dataclasses.replace(results[0], json_test=model.json_test) == model

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_many_iter", depends=["PymysqlTestDataclassFunctions::get_many"])
    def test_get_many_iter(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        for result in queries.get_many_mysql_type(conn=pymysql_conn, id_=model.id_):
            assert result is not None
            assert isinstance(result, models.TestMysqlType)
            assert dataclasses.replace(result, json_test=model.json_test) == model

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_many_inner", depends=["PymysqlTestDataclassFunctions::get_many_iter"])
    def test_get_many_inner(self, pymysql_conn: pymysql.Connection, inner_model: models.TestInnerMysqlType) -> None:
        result = queries.get_many_inner_mysql_type(conn=pymysql_conn, table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestInnerMysqlType)
        assert results[0] == inner_model

        results = result()
        assert results[0] == inner_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_inner_iter",
        depends=["PymysqlTestDataclassFunctions::get_many_inner"],
    )
    def test_get_many_inner_iter(self, pymysql_conn: pymysql.Connection, inner_model: models.TestInnerMysqlType) -> None:
        for result in queries.get_many_inner_mysql_type(conn=pymysql_conn, table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)
            assert result == inner_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_nullable_inner",
        depends=["PymysqlTestDataclassFunctions::get_many_inner_iter"],
    )
    def test_get_many_nullable_inner(self, pymysql_conn: pymysql.Connection, inner_model: models.TestInnerMysqlType) -> None:
        # int_test is None; the query uses the NULL-safe <=> comparison.
        result = queries.get_many_nullable_inner_mysql_type(conn=pymysql_conn, table_id=inner_model.table_id, int_test=inner_model.int_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = result()
        assert isinstance(results[0], models.TestInnerMysqlType)
        assert results[0] == inner_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_nullable_inner_iter",
        depends=["PymysqlTestDataclassFunctions::get_many_nullable_inner"],
    )
    def test_get_many_nullable_inner_iter(self, pymysql_conn: pymysql.Connection, inner_model: models.TestInnerMysqlType) -> None:
        for result in queries.get_many_nullable_inner_mysql_type(conn=pymysql_conn, table_id=inner_model.table_id, int_test=inner_model.int_test):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)
            assert result == inner_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_date",
        depends=["PymysqlTestDataclassFunctions::get_many_nullable_inner_iter"],
    )
    def test_get_many_date(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_date(conn=pymysql_conn, id_=model.id_, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.date)
        assert results[0] == model.date_test

        results = result()
        assert results[0] == model.date_test

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_date_iter",
        depends=["PymysqlTestDataclassFunctions::get_many_date"],
    )
    def test_get_many_date_iter(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        for result in queries.get_many_date(conn=pymysql_conn, id_=model.id_, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)
            assert result == model.date_test

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_time",
        depends=["PymysqlTestDataclassFunctions::get_many_date_iter"],
    )
    def test_get_many_time(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_time(conn=pymysql_conn, id_=model.id_, time_test=model.time_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.timedelta)
        assert results[0] == model.time_test

        results = result()
        assert results[0] == model.time_test

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_time_iter",
        depends=["PymysqlTestDataclassFunctions::get_many_time"],
    )
    def test_get_many_time_iter(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        for result in queries.get_many_time(conn=pymysql_conn, id_=model.id_, time_test=model.time_test):
            assert result is not None
            assert isinstance(result, datetime.timedelta)
            assert result == model.time_test

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_bool",
        depends=["PymysqlTestDataclassFunctions::get_many_time_iter"],
    )
    def test_get_many_bool(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_bool(conn=pymysql_conn, id_=model.id_, tinyint1_test=model.tinyint1_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], bool)
        assert results[0] is True

        results = result()
        assert results[0] is True

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_bool_iter",
        depends=["PymysqlTestDataclassFunctions::get_many_bool"],
    )
    def test_get_many_bool_iter(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        for result in queries.get_many_bool(conn=pymysql_conn, id_=model.id_, tinyint1_test=model.tinyint1_test):
            assert result is not None
            assert isinstance(result, bool)
            assert result is True

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_decimal",
        depends=["PymysqlTestDataclassFunctions::get_many_bool_iter"],
    )
    def test_get_many_decimal(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_decimal(conn=pymysql_conn, id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], decimal.Decimal)
        assert results[0] == model.decimal_test
        assert str(results[0]) == DECIMAL_PADDED

        results = result()
        assert results[0] == model.decimal_test

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_decimal_iter",
        depends=["PymysqlTestDataclassFunctions::get_many_decimal"],
    )
    def test_get_many_decimal_iter(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        for result in queries.get_many_decimal(conn=pymysql_conn, id_=model.id_, decimal_test=model.decimal_test):
            assert result is not None
            assert isinstance(result, decimal.Decimal)
            assert result == model.decimal_test

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_mood",
        depends=["PymysqlTestDataclassFunctions::get_many_decimal_iter"],
    )
    def test_get_many_mood(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        result = queries.get_many_mood(conn=pymysql_conn, mood=model.mood)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        # The query is not id-filtered; other rows may exist, so assert containment.
        assert len(results) >= 1
        assert all(isinstance(mood, enums.TestMysqlTypesMood) for mood in results)
        assert all(mood is enums.TestMysqlTypesMood.VALUE_24H for mood in results)

        results = result()
        assert all(mood is enums.TestMysqlTypesMood.VALUE_24H for mood in results)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_many_mood_iter",
        depends=["PymysqlTestDataclassFunctions::get_many_mood"],
    )
    def test_get_many_mood_iter(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        for result in queries.get_many_mood(conn=pymysql_conn, mood=model.mood):
            assert result is not None
            assert isinstance(result, enums.TestMysqlTypesMood)
            assert result is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::list_months", depends=["PymysqlTestDataclassFunctions::get_many_mood_iter"])
    def test_list_months(self, pymysql_conn: pymysql.Connection) -> None:
        # DATE_FORMAT emits %% in the stored SQL; the empty argument tuple
        # still goes through pymysql's %-substitution, halving it back.
        result = queries.list_months(conn=pymysql_conn)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        months = result()
        assert list(months) == [EXPECTED_MONTH]

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::list_months_iter", depends=["PymysqlTestDataclassFunctions::list_months"])
    def test_list_months_iter(self, pymysql_conn: pymysql.Connection) -> None:
        months = list(queries.list_months(conn=pymysql_conn))
        # The table is shared; the doubled %% only has to survive for our row.
        assert EXPECTED_MONTH in months

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::count", depends=["PymysqlTestDataclassFunctions::list_months_iter"])
    def test_count_mysql_types(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries.count_mysql_types(conn=pymysql_conn)

        # The shared table may carry other files' rows; only a lower bound is safe.
        assert result is not None
        assert result >= 1

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::all_cursor", depends=["PymysqlTestDataclassFunctions::count"])
    def test_all_mysql_types_cursor(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        cur = queries.all_mysql_types_cursor(conn=pymysql_conn)

        assert isinstance(cur, pymysql.cursors.Cursor)
        rows = cur.fetchall()
        # The shared table may carry other files' rows; assert on our own.
        assert model.id_ in {row[0] for row in rows}
        cur.close()

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::update_rows", depends=["PymysqlTestDataclassFunctions::all_cursor"])
    def test_update_varchar_rows(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        result = queries.update_varchar_test(conn=pymysql_conn, varchar_test=UPDATED_VARCHAR, id_=model.id_)

        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::update_rows_noop", depends=["PymysqlTestDataclassFunctions::update_rows"])
    def test_update_varchar_rows_noop(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        # Without CLIENT.FOUND_ROWS MySQL reports changed rows, so setting
        # the same value again affects nothing.
        result = queries.update_varchar_test(conn=pymysql_conn, varchar_test=UPDATED_VARCHAR, id_=model.id_)

        assert result == 0

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert_last_id", depends=["PymysqlTestDataclassFunctions::update_rows_noop"])
    def test_insert_exec_last_id(self, pymysql_conn: pymysql.Connection) -> None:
        # The AUTO_INCREMENT counter persists across runs; never assert an
        # exact id.
        new_id = queries.insert_exec_last_id(conn=pymysql_conn, name=EXEC_LAST_ID_NAME)

        assert new_id is not None
        assert isinstance(new_id, int)
        assert new_id > 0
        assert queries.get_exec_last_id_name(conn=pymysql_conn, id_=new_id) == EXEC_LAST_ID_NAME

        # A statement that inserts nothing has no last row id: the OK
        # packet's 0 maps to the documented None.
        assert queries.touch_exec_last_id(conn=pymysql_conn, name="untouched", id_=new_id + 1000000) is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::delete", depends=["PymysqlTestDataclassFunctions::insert_last_id"])
    def test_delete_mysql_type(self, pymysql_conn: pymysql.Connection, model: models.TestMysqlType) -> None:
        queries.delete_one_mysql_type(conn=pymysql_conn, id_=model.id_)

        assert queries.get_one_mysql_type(conn=pymysql_conn, id_=model.id_) is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert_type_override")
    def test_insert_type_override(self, pymysql_conn: pymysql.Connection, override_model: models.TestTypeOverride) -> None:
        queries.insert_type_override(conn=pymysql_conn, id_=override_model.id_, text_test=override_model.text_test)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_type_override",
        depends=["PymysqlTestDataclassFunctions::insert_type_override"],
    )
    def test_get_type_override(self, pymysql_conn: pymysql.Connection, override_model: models.TestTypeOverride) -> None:
        result = queries.get_type_override(conn=pymysql_conn, id_=override_model.id_)

        assert result is not None
        assert isinstance(result.text_test, UserString)
        assert result == override_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_type_override_none_value",
        depends=["PymysqlTestDataclassFunctions::get_type_override"],
    )
    def test_get_type_override_none_value(self, pymysql_conn: pymysql.Connection) -> None:
        # The override target is nullable: NULL must come back as None
        # without passing through UserString.
        queries.insert_type_override(conn=pymysql_conn, id_=TYPE_OVERRIDE_NONE_ID, text_test=None)
        result = queries.get_type_override(conn=pymysql_conn, id_=TYPE_OVERRIDE_NONE_ID)

        assert result is not None
        assert result.text_test is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_type_override_not_found",
        depends=["PymysqlTestDataclassFunctions::get_type_override_none_value"],
    )
    def test_get_type_override_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries.get_type_override(conn=pymysql_conn, id_=0) is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert_reserved_arg")
    def test_insert_reserved_arg(self, pymysql_conn: pymysql.Connection) -> None:
        # The column is literally named "conn"; the generated parameter is
        # deduplicated against the implicit connection argument.
        queries.insert_reserved_arg(conn=pymysql_conn, id_=RESERVED_ARG_ID, conn_2=RESERVED_ARG_VALUE)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_reserved_arg",
        depends=["PymysqlTestDataclassFunctions::insert_reserved_arg"],
    )
    def test_get_reserved_arg(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries.get_reserved_arg(conn=pymysql_conn, conn_2=RESERVED_ARG_VALUE)

        assert result == models.TestReservedArg(id_=RESERVED_ARG_ID, conn=RESERVED_ARG_VALUE)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_reserved_arg_not_found",
        depends=["PymysqlTestDataclassFunctions::get_reserved_arg"],
    )
    def test_get_reserved_arg_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries.get_reserved_arg(conn=pymysql_conn, conn_2="missing-reserved-arg-value") is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert_case_rows")
    def test_insert_case_rows(self, pymysql_conn: pymysql.Connection) -> None:
        queries_case.insert_case_row(conn=pymysql_conn, id_=CASE_ID, upper_dt=CASE_DT, prec_dec=CASE_DEC)
        queries_case.insert_case_row(conn=pymysql_conn, id_=CASE_ID + 1, upper_dt=CASE_DT, prec_dec=CASE_DEC)

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_case_row", depends=["PymysqlTestDataclassFunctions::insert_case_rows"])
    def test_get_case_row(self, pymysql_conn: pymysql.Connection) -> None:
        row = queries_case.get_case_row(conn=pymysql_conn, id_=CASE_ID)

        assert row is not None
        assert isinstance(row.upper_dt, datetime.datetime)
        assert row.upper_dt == CASE_DT
        assert isinstance(row.prec_dec, decimal.Decimal)
        assert row.prec_dec == CASE_DEC

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_case_row_not_found",
        depends=["PymysqlTestDataclassFunctions::get_case_row"],
    )
    def test_get_case_row_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries_case.get_case_row(conn=pymysql_conn, id_=0) is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::count_case_rows_filters",
        depends=["PymysqlTestDataclassFunctions::get_case_row_not_found"],
    )
    def test_count_case_rows_filters(self, pymysql_conn: pymysql.Connection) -> None:
        # The WHERE clause lives inside an executable /*! version comment;
        # raising the threshold by one must drop exactly the first row.
        count_ge_first = queries_case.count_case_rows(conn=pymysql_conn, id_=CASE_ID)
        count_ge_second = queries_case.count_case_rows(conn=pymysql_conn, id_=CASE_ID + 1)

        assert count_ge_first is not None
        assert count_ge_second is not None
        assert count_ge_first - count_ge_second == 1

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert_enum_override")
    def test_insert_enum_override(self, pymysql_conn: pymysql.Connection) -> None:
        queries_enum_override.insert_enum_override(conn=pymysql_conn, id_=ENUM_OVERRIDE_ID, mood_test="happy")
        queries_enum_override.insert_enum_override(conn=pymysql_conn, id_=ENUM_OVERRIDE_ID_2, mood_test="sad")

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_enum_override_mood",
        depends=["PymysqlTestDataclassFunctions::insert_enum_override"],
    )
    def test_get_enum_override_mood(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_enum_override.get_enum_override_mood(conn=pymysql_conn, id_=ENUM_OVERRIDE_ID)

        assert result == "happy"
        assert isinstance(result, str)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_enum_override_mood_not_found",
        depends=["PymysqlTestDataclassFunctions::get_enum_override_mood"],
    )
    def test_get_enum_override_mood_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries_enum_override.get_enum_override_mood(conn=pymysql_conn, id_=0) is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::list_enum_override_by_ids",
        depends=["PymysqlTestDataclassFunctions::get_enum_override_mood_not_found"],
    )
    def test_list_enum_override_by_ids(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_enum_override.list_enum_override_by_ids(conn=pymysql_conn, ids=[ENUM_OVERRIDE_ID, ENUM_OVERRIDE_ID_2])

        assert isinstance(result, queries_enum_override.QueryResults)
        rows = result()
        assert rows == [
            models.TestEnumOverride(id_=ENUM_OVERRIDE_ID, mood_test="happy"),
            models.TestEnumOverride(id_=ENUM_OVERRIDE_ID_2, mood_test="sad"),
        ]
        assert list(result) == rows

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::list_enum_override_by_ids_empty",
        depends=["PymysqlTestDataclassFunctions::list_enum_override_by_ids"],
    )
    def test_list_enum_override_by_ids_empty(self, pymysql_conn: pymysql.Connection) -> None:
        # An empty slice expands the placeholder to NULL: IN (NULL) matches
        # no rows instead of raising.
        assert queries_enum_override.list_enum_override_by_ids(conn=pymysql_conn, ids=[])() == []

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::seed_field_naming")
    def test_seed_field_naming(self, pymysql_conn: pymysql.Connection) -> None:
        # No generated insert exists for this table; seed it directly.
        with pymysql_conn.cursor() as cur:
            cur.execute("INSERT INTO test_field_namings (id, outputs) VALUES (%s, %s)", (FIELD_NAMING_ID, json.dumps({"first": 1})))

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_field_naming", depends=["PymysqlTestDataclassFunctions::seed_field_naming"])
    def test_get_field_naming(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_field_namings.get_field_naming(conn=pymysql_conn, id_=FIELD_NAMING_ID)

        assert result is not None
        assert result.id_ == FIELD_NAMING_ID
        assert json.loads(result.outputs) == {"first": 1}

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_field_naming_not_found",
        depends=["PymysqlTestDataclassFunctions::get_field_naming"],
    )
    def test_get_field_naming_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries_field_namings.get_field_naming(conn=pymysql_conn, id_=0) is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_joined_field_namings",
        depends=["PymysqlTestDataclassFunctions::get_field_naming_not_found"],
    )
    def test_get_joined_field_namings(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_field_namings.get_joined_field_namings(conn=pymysql_conn, id_=FIELD_NAMING_ID)

        assert result is not None
        assert isinstance(result, queries_field_namings.GetJoinedFieldNamingsRow)
        assert json.loads(result.outputs) == {"first": 1}
        assert result.outputs == result.outputs_2

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::set_field_naming_outputs",
        depends=["PymysqlTestDataclassFunctions::get_joined_field_namings"],
    )
    def test_set_field_naming_outputs(self, pymysql_conn: pymysql.Connection) -> None:
        queries_field_namings.set_field_naming_outputs(conn=pymysql_conn, outputs=json.dumps({"second": 2}), id_=FIELD_NAMING_ID)
        result = queries_field_namings.get_field_naming(conn=pymysql_conn, id_=FIELD_NAMING_ID)

        assert result is not None
        assert json.loads(result.outputs) == {"second": 2}

    @pytest.mark.dependency(depends=["PymysqlTestDataclassFunctions::set_field_naming_outputs"])
    def test_delete_field_naming(self, pymysql_conn: pymysql.Connection) -> None:
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_field_namings WHERE id = %s", (FIELD_NAMING_ID,))

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert_invalid_identifiers")
    def test_insert_invalid_identifiers(self, pymysql_conn: pymysql.Connection) -> None:
        queries_invalid_identifiers.insert_invalid_identifiers(
            conn=pymysql_conn,
            id_=INVALID_IDENTIFIER_ID,
            column_3p_="3p-value",
            new_notes="some new notes",
        )

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_invalid_identifiers",
        depends=["PymysqlTestDataclassFunctions::insert_invalid_identifiers"],
    )
    def test_get_invalid_identifiers(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_invalid_identifiers.get_invalid_identifiers(conn=pymysql_conn, id_=INVALID_IDENTIFIER_ID)

        # The insert never sets `%pct`, so it stays NULL.
        assert result == models.TestInvalidIdentifier(
            id_=INVALID_IDENTIFIER_ID,
            column_3p_="3p-value",
            new_notes="some new notes",
            column__pct=None,
        )

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_invalid_identifiers_not_found",
        depends=["PymysqlTestDataclassFunctions::get_invalid_identifiers"],
    )
    def test_get_invalid_identifiers_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries_invalid_identifiers.get_invalid_identifiers(conn=pymysql_conn, id_=0) is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert_third_party_stat")
    def test_insert_third_party_stat(self, pymysql_conn: pymysql.Connection) -> None:
        queries_invalid_identifiers.insert_third_party_stat(conn=pymysql_conn, id_=THIRD_PARTY_ID, total=THIRD_PARTY_TOTAL)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_third_party_stat",
        depends=["PymysqlTestDataclassFunctions::insert_third_party_stat"],
    )
    def test_get_third_party_stat(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_invalid_identifiers.get_third_party_stat(conn=pymysql_conn, id_=THIRD_PARTY_ID)

        assert result == models.Model3RdPartyStat(id_=THIRD_PARTY_ID, total=THIRD_PARTY_TOTAL)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_third_party_stat_not_found",
        depends=["PymysqlTestDataclassFunctions::get_third_party_stat"],
    )
    def test_get_third_party_stat_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries_invalid_identifiers.get_third_party_stat(conn=pymysql_conn, id_=0) is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert_slice_rows")
    def test_insert_slice_rows(self, pymysql_conn: pymysql.Connection) -> None:
        for offset, (name, note) in enumerate((("a", "x"), ("b", "y"), ("c", None), ("b", "y"))):
            queries_slice.insert_slice_row(conn=pymysql_conn, id_=SLICE_ID_BASE + offset, name=name, note=note)

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_slice_rows", depends=["PymysqlTestDataclassFunctions::insert_slice_rows"])
    def test_get_slice_rows(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_slice.get_slice_rows(conn=pymysql_conn, ids=[SLICE_ID_BASE, SLICE_ID_BASE + 2])
        assert isinstance(result, queries_slice.QueryResults)
        rows = result()
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a", note="x"),
            models.TestSlice(id_=SLICE_ID_BASE + 2, name="c", note=None),
        ]
        assert list(result) == rows

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_slice_rows_empty_slice",
        depends=["PymysqlTestDataclassFunctions::insert_slice_rows"],
    )
    def test_get_slice_rows_empty_slice(self, pymysql_conn: pymysql.Connection) -> None:
        # An empty sequence expands the placeholder to NULL: IN (NULL)
        # matches no rows instead of raising.
        assert queries_slice.get_slice_rows(conn=pymysql_conn, ids=[])() == []

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_slice_row_filtered",
        depends=["PymysqlTestDataclassFunctions::insert_slice_rows"],
    )
    def test_get_slice_row_filtered(self, pymysql_conn: pymysql.Connection) -> None:
        # Plain params surround the slice, so this proves the flattened
        # argument tuple binds in SQL text order.
        row = queries_slice.get_slice_row_filtered(
            conn=pymysql_conn,
            name="b",
            ids=[SLICE_ID_BASE + 1, SLICE_ID_BASE + 3],
            id_=SLICE_ID_BASE + 1,
        )
        assert row == models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y")

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_slice_row_filtered_not_found",
        depends=["PymysqlTestDataclassFunctions::insert_slice_rows"],
    )
    def test_get_slice_row_filtered_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries_slice.get_slice_row_filtered(conn=pymysql_conn, name="a", ids=[SLICE_ID_BASE], id_=SLICE_ID_BASE) is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_slice_rows_by_notes",
        depends=["PymysqlTestDataclassFunctions::insert_slice_rows"],
    )
    def test_get_slice_rows_by_notes(self, pymysql_conn: pymysql.Connection) -> None:
        # The slice targets a nullable column; the parameter is still a plain
        # Sequence, and rows whose note is NULL never match.
        rows = queries_slice.get_slice_rows_by_notes(conn=pymysql_conn, notes=["y"])()
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE + 1, name="b", note="y"),
            models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y"),
        ]
        assert queries_slice.get_slice_rows_by_notes(conn=pymysql_conn, notes=[])() == []

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_slice_rows_by_name_or_note",
        depends=["PymysqlTestDataclassFunctions::insert_slice_rows"],
    )
    def test_get_slice_rows_by_name_or_note(self, pymysql_conn: pymysql.Connection) -> None:
        # The same slice name is used twice but the function takes ONE
        # parameter; every marker occurrence is expanded and the sequence is
        # bound once per occurrence.
        rows = queries_slice.get_slice_rows_by_name_or_note(conn=pymysql_conn, names=["b", "x"])()
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a", note="x"),
            models.TestSlice(id_=SLICE_ID_BASE + 1, name="b", note="y"),
            models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y"),
        ]
        assert queries_slice.get_slice_rows_by_name_or_note(conn=pymysql_conn, names=[])() == []

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_slice_rows_by_name_or_note_filtered",
        depends=["PymysqlTestDataclassFunctions::insert_slice_rows"],
    )
    def test_get_slice_rows_by_name_or_note_filtered(self, pymysql_conn: pymysql.Connection) -> None:
        # A plain parameter sits between the two uses of the slice, so this
        # proves the flattened arguments follow SQL text order.
        rows = queries_slice.get_slice_rows_by_name_or_note_filtered(conn=pymysql_conn, names=["b", "x"], id_=SLICE_ID_BASE + 1)()
        assert rows == [
            models.TestSlice(id_=SLICE_ID_BASE, name="a", note="x"),
            models.TestSlice(id_=SLICE_ID_BASE + 3, name="b", note="y"),
        ]
        assert queries_slice.get_slice_rows_by_name_or_note_filtered(conn=pymysql_conn, names=[], id_=SLICE_ID_BASE)() == []

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_first_slice_name_two_slices",
        depends=["PymysqlTestDataclassFunctions::insert_slice_rows"],
    )
    def test_get_first_slice_name_two_slices(self, pymysql_conn: pymysql.Connection) -> None:
        name = queries_slice.get_first_slice_name(conn=pymysql_conn, ids=[SLICE_ID_BASE + 1], names=["a"])
        assert name == "a"
        assert queries_slice.get_first_slice_name(conn=pymysql_conn, ids=[], names=[]) is None

    @pytest.mark.dependency(
        depends=[
            "PymysqlTestDataclassFunctions::get_slice_rows",
            "PymysqlTestDataclassFunctions::get_slice_rows_empty_slice",
            "PymysqlTestDataclassFunctions::get_slice_row_filtered",
            "PymysqlTestDataclassFunctions::get_slice_row_filtered_not_found",
            "PymysqlTestDataclassFunctions::get_slice_rows_by_notes",
            "PymysqlTestDataclassFunctions::get_slice_rows_by_name_or_note",
            "PymysqlTestDataclassFunctions::get_slice_rows_by_name_or_note_filtered",
            "PymysqlTestDataclassFunctions::get_first_slice_name_two_slices",
        ]
    )
    def test_delete_slice_rows(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries_slice.delete_slice_rows(conn=pymysql_conn, ids=[]) == 0
        deleted = queries_slice.delete_slice_rows(conn=pymysql_conn, ids=[SLICE_ID_BASE + offset for offset in range(SLICE_ROW_COUNT)])
        assert deleted == SLICE_ROW_COUNT

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::insert_converted")
    def test_insert_converted(self, pymysql_conn: pymysql.Connection) -> None:
        queries_converters.insert_converted(
            conn=pymysql_conn,
            id_=CONVERTER_ID,
            prefs=Preferences(theme="dark", notifications=True),
            maybe_prefs=None,
            tags=frozenset({"alpha", "beta"}),
        )
        queries_converters.insert_converted(
            conn=pymysql_conn,
            id_=CONVERTER_ID_2,
            prefs=Preferences(theme="light", notifications=False),
            maybe_prefs=Preferences(theme="light", notifications=True),
            tags=frozenset({"gamma"}),
        )

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::get_converted", depends=["PymysqlTestDataclassFunctions::insert_converted"])
    def test_get_converted(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_converters.get_converted(conn=pymysql_conn, id_=CONVERTER_ID)

        assert result is not None
        assert result.prefs == Preferences(theme="dark", notifications=True)
        # NULL never reaches the decoder; it comes back as plain None.
        assert result.maybe_prefs is None
        assert result.tags == frozenset({"alpha", "beta"})

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_converted_with_prefs",
        depends=["PymysqlTestDataclassFunctions::get_converted"],
    )
    def test_get_converted_with_prefs(self, pymysql_conn: pymysql.Connection) -> None:
        result = queries_converters.get_converted(conn=pymysql_conn, id_=CONVERTER_ID_2)

        assert result == models.TestConverter(
            id_=CONVERTER_ID_2,
            prefs=Preferences(theme="light", notifications=False),
            maybe_prefs=Preferences(theme="light", notifications=True),
            tags=frozenset({"gamma"}),
        )

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::get_converted_not_found",
        depends=["PymysqlTestDataclassFunctions::get_converted_with_prefs"],
    )
    def test_get_converted_not_found(self, pymysql_conn: pymysql.Connection) -> None:
        assert queries_converters.get_converted(conn=pymysql_conn, id_=0) is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassFunctions::list_converted_by_tags",
        depends=["PymysqlTestDataclassFunctions::get_converted_not_found"],
    )
    def test_list_converted_by_tags(self, pymysql_conn: pymysql.Connection) -> None:
        # The parameter converts through encode_tags before hitting the db.
        result = queries_converters.list_converted_by_tags(conn=pymysql_conn, tags=frozenset({"alpha", "beta"}))

        assert isinstance(result, queries_converters.QueryResults)
        assert result() == [CONVERTER_ID]
        assert list(result) == [CONVERTER_ID]

    @pytest.mark.dependency(depends=["PymysqlTestDataclassFunctions::list_converted_by_tags"])
    def test_delete_converted(self, pymysql_conn: pymysql.Connection) -> None:
        queries_converters.delete_converted(conn=pymysql_conn, id_=CONVERTER_ID)
        queries_converters.delete_converted(conn=pymysql_conn, id_=CONVERTER_ID_2)

        assert queries_converters.get_converted(conn=pymysql_conn, id_=CONVERTER_ID) is None
        assert queries_converters.get_converted(conn=pymysql_conn, id_=CONVERTER_ID_2) is None

    def test_one_missing_rows_return_none(self, pymysql_conn: pymysql.Connection) -> None:
        # Every :one not-found branch, plus the no-insert :execlastid. The
        # count queries always return a row, so their miss branch needs the
        # no-row stub.
        assert queries.get_one_mysql_type(conn=pymysql_conn, id_=-1) is None
        assert queries.get_one_inner_mysql_type(conn=pymysql_conn, table_id=-1) is None
        assert queries.get_one_date(conn=pymysql_conn, id_=-1, date_test=datetime.date(1970, 1, 1)) is None
        assert queries.get_one_datetime(conn=pymysql_conn, id_=-1, datetime_test=datetime.datetime(1970, 1, 1)) is None
        assert queries.get_one_time(conn=pymysql_conn, id_=-1, time_test=datetime.timedelta()) is None
        assert queries.get_one_bool(conn=pymysql_conn, id_=-1, tinyint1_test=False) is None
        assert queries.get_one_decimal(conn=pymysql_conn, id_=-1, decimal_test=decimal.Decimal(0)) is None
        assert queries.get_one_blob(conn=pymysql_conn, id_=-1, blob_test=memoryview(b"")) is None
        assert queries.get_one_bit(conn=pymysql_conn, id_=-1) is None
        assert queries.get_one_year(conn=pymysql_conn, id_=-1) is None
        assert queries.get_one_json(conn=pymysql_conn, id_=-1) is None
        assert queries.get_one_mood(conn=pymysql_conn, id_=-1, mood=enums.TestMysqlTypesMood.SAD) is None
        assert queries.get_one_tag(conn=pymysql_conn, id_=-1) is None
        assert queries.get_exec_last_id_name(conn=pymysql_conn, id_=-1) is None
        assert queries.get_type_override(conn=pymysql_conn, id_=-1) is None
        assert queries.get_reserved_arg(conn=pymysql_conn, conn_2="missing") is None
        assert queries.touch_exec_last_id(conn=pymysql_conn, name="untouched", id_=-1) is None
        assert queries_case.get_case_row(conn=pymysql_conn, id_=-1) is None
        assert queries_field_namings.get_field_naming(conn=pymysql_conn, id_=-1) is None
        assert queries_field_namings.get_joined_field_namings(conn=pymysql_conn, id_=-1) is None
        assert queries_invalid_identifiers.get_invalid_identifiers(conn=pymysql_conn, id_=-1) is None
        assert queries_enum_override.get_enum_override_mood(conn=pymysql_conn, id_=-1) is None
        assert queries_enum_override.count_enum_override_by_moods(conn=pymysql_conn, moods=[]) == 0

        stub = typing.cast("pymysql.Connection", no_row_conn.NoRowConn())
        assert queries.count_mysql_types(conn=stub) is None
        assert queries_case.count_case_rows(conn=stub, id_=0) is None
        assert queries_enum_override.count_enum_override_by_moods(conn=stub, moods=[]) is None

    @pytest.mark.dependency(depends=["PymysqlTestDataclassFunctions::insert_enum_override"])
    def test_enum_override_cleanup(self, pymysql_conn: pymysql.Connection) -> None:
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_enum_override WHERE id IN (%s, %s)", (ENUM_OVERRIDE_ID, ENUM_OVERRIDE_ID_2))
        assert queries_enum_override.get_enum_override_mood(conn=pymysql_conn, id_=ENUM_OVERRIDE_ID) is None

    @pytest.mark.dependency(name="PymysqlTestDataclassFunctions::backslash_sql")
    def test_backslash_sql(self, pymysql_conn: pymysql.Connection) -> None:
        # A plain Python literal would read the "\t" as a tab and drop the
        # "\d", so the constant has to be a raw string for the doubled
        # backslashes to reach MySQL, which unescapes them back to one each.
        assert queries_backslash.get_backslash_pattern(conn=pymysql_conn) == "a\\tb\\d+"
        queries_backslash.insert_backslash_row(conn=pymysql_conn, id_=BACKSLASH_ID, name="path", note="C:\\dir\\name")
        assert queries_backslash.get_backslash_note(conn=pymysql_conn, id_=BACKSLASH_ID) == "C:\\dir\\name"
