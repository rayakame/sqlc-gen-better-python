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
import typing
from collections import UserString

import pymysql
import pymysql.cursors
import pytest

from test.driver_pymysql import no_row_conn
from test.driver_pymysql.pydantic.classes import enums
from test.driver_pymysql.pydantic.classes import models
from test.driver_pymysql.pydantic.classes import queries
from test.driver_pymysql.pydantic.classes import queries_case
from test.driver_pymysql.pydantic.classes import queries_enum_override
from test.driver_pymysql.pydantic.classes import queries_field_namings
from test.driver_pymysql.pydantic.classes import queries_invalid_identifiers

# Ids fixed and unique across the pymysql suites (pydantic owns 4000-4999);
# every chain deletes its rows at the end so reruns start clean.
TYPE_ID: typing.Final = 4000
OVERRIDE_ID: typing.Final = 4010
OVERRIDE_NONE_ID: typing.Final = 4011
RESERVED_ID: typing.Final = 4020
CASE_IDS: typing.Final = (4030, 4031)
ENUM_IDS: typing.Final = (4040, 4041)
FIELD_ID: typing.Final = 4050
INVALID_ID: typing.Final = 4060
THIRD_PARTY_ID: typing.Final = 4061
MISSING_ID: typing.Final = 4999
RESERVED_CONN: typing.Final = "pydantic-classes-conn"
EXEC_LAST_ID_NAME: typing.Final = "pydantic-classes-lastid"
CASE_DT: typing.Final = datetime.datetime(2026, 7, 19, 8, 15)
CASE_DEC: typing.Final = decimal.Decimal("12.34")


def _without_json(row: models.TestMysqlType) -> models.TestMysqlType:
    # MySQL normalizes JSON spacing, so json_test never compares as a string.
    return row.model_copy(update={"json_test": ""})


class TestPymysqlPydanticClasses:
    @pytest.fixture(scope="session")
    def override_model(self) -> models.TestTypeOverride:
        return models.TestTypeOverride(id_=OVERRIDE_ID, text_test=UserString("Test"))

    @pytest.fixture(scope="session")
    def model(self) -> models.TestMysqlType:
        return models.TestMysqlType(
            id_=TYPE_ID,
            int_test=42,
            integer_test=-42,
            mediumint_test=8_388_607,
            smallint_test=-32_768,
            tinyint_test=-128,
            bigint_test=9_223_372_036_854_775_807,
            int_unsigned_test=4_294_967_295,
            bigint_unsigned_test=2**63 + 11,
            year_test=2026,
            tinyint1_test=True,
            bool_test=True,
            boolean_test=False,
            float_test=2.5,
            double_test=math.pi,
            double_precision_test=math.e,
            real_test=1.5,
            decimal_test=decimal.Decimal("12.34"),
            numeric_test=decimal.Decimal("99.99"),
            char_test="ABCDEFGHIJ",
            varchar_test="Hello varchar",
            tinytext_test="tiny text",
            text_test="Some text",
            mediumtext_test="medium text",
            longtext_test="long text",
            binary_test=memoryview(b"0123456789abcdef"),
            varbinary_test=memoryview(b"\x00\x01varbinary"),
            tinyblob_test=memoryview(b"tinyblob"),
            blob_test=memoryview(b"\x00\x01\x02hello"),
            mediumblob_test=memoryview(b"mediumblob"),
            longblob_test=memoryview(b"longblob"),
            bit_test=memoryview(b"\x80"),
            date_test=datetime.date(2026, 1, 15),
            datetime_test=datetime.datetime(2026, 1, 15, 12, 30, 45),
            datetime6_test=datetime.datetime(2026, 1, 15, 12, 30, 45, 123456),
            timestamp_test=datetime.datetime(2026, 1, 2, 3, 4, 5),
            time_test=datetime.timedelta(hours=1, minutes=2, seconds=3),
            json_test=json.dumps({"foo": "bar"}),
            mood=enums.TestMysqlTypesMood.VALUE_24H,
            tag=enums.TestMysqlTypesTag.BETA,
        )

    @pytest.fixture(scope="session")
    def inner_model(self, model: models.TestMysqlType) -> models.TestInnerMysqlType:
        return models.TestInnerMysqlType(
            table_id=model.id_,
            int_test=None,
            integer_test=7,
            mediumint_test=None,
            smallint_test=3,
            tinyint_test=127,
            bigint_test=None,
            int_unsigned_test=None,
            bigint_unsigned_test=2**63 + 42,
            year_test=1901,
            tinyint1_test=False,
            bool_test=None,
            boolean_test=True,
            float_test=None,
            double_test=0.25,
            double_precision_test=None,
            real_test=None,
            decimal_test=decimal.Decimal("0.5000"),
            numeric_test=None,
            char_test=None,
            varchar_test="inner varchar",
            tinytext_test=None,
            text_test=None,
            mediumtext_test=None,
            longtext_test=None,
            binary_test=None,
            varbinary_test=memoryview(b"inner"),
            tinyblob_test=None,
            blob_test=None,
            mediumblob_test=None,
            longblob_test=None,
            bit_test=memoryview(b"\x01"),
            date_test=None,
            datetime_test=None,
            datetime6_test=None,
            timestamp_test=None,
            time_test=datetime.timedelta(hours=8, minutes=30),
            json_test=None,
            mood=enums.TestInnerMysqlTypesMood.VALUE__HIDDEN,
            tag=None,
        )

    @pytest.fixture(scope="session")
    def queries_obj(self, pymysql_conn: pymysql.Connection) -> queries.Queries:
        return queries.Queries(conn=pymysql_conn)

    @pytest.fixture(scope="session")
    def case_obj(self, pymysql_conn: pymysql.Connection) -> queries_case.QueriesCase:
        return queries_case.QueriesCase(conn=pymysql_conn)

    @pytest.fixture(scope="session")
    def enum_obj(self, pymysql_conn: pymysql.Connection) -> queries_enum_override.QueriesEnumOverride:
        return queries_enum_override.QueriesEnumOverride(conn=pymysql_conn)

    @pytest.fixture(scope="session")
    def field_obj(self, pymysql_conn: pymysql.Connection) -> queries_field_namings.QueriesFieldNamings:
        return queries_field_namings.QueriesFieldNamings(conn=pymysql_conn)

    @pytest.fixture(scope="session")
    def invalid_obj(self, pymysql_conn: pymysql.Connection) -> queries_invalid_identifiers.QueriesInvalidIdentifiers:
        return queries_invalid_identifiers.QueriesInvalidIdentifiers(conn=pymysql_conn)

    def test_conn_attr(self, queries_obj: queries.Queries, pymysql_conn: pymysql.Connection) -> None:
        assert isinstance(queries_obj.conn, pymysql.Connection)
        assert queries_obj.conn is pymysql_conn

    @pytest.mark.dependency(name="PymysqlPydanticClasses::insert")
    def test_insert(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        queries_obj.insert_one_mysql_type(
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

    @pytest.mark.dependency(name="PymysqlPydanticClasses::inner_insert", depends=["PymysqlPydanticClasses::insert"])
    def test_inner_insert(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        queries_obj.insert_one_inner_mysql_type(
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

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_one", depends=["PymysqlPydanticClasses::inner_insert"])
    def test_get_one(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_one_mysql_type(id_=TYPE_ID)

        assert result is not None
        assert isinstance(result, models.TestMysqlType)
        assert json.loads(result.json_test) == json.loads(model.json_test)
        assert _without_json(result) == _without_json(model)
        assert result.tinyint1_test is True
        assert result.bool_test is True
        assert result.boolean_test is False
        # plain datetime drops microseconds, datetime(6) keeps them
        assert result.datetime_test.microsecond == 0
        assert result.datetime6_test.microsecond == model.datetime6_test.microsecond
        assert result.bigint_unsigned_test == 2**63 + 11
        assert result.mood is enums.TestMysqlTypesMood.VALUE_24H
        assert result.tag is enums.TestMysqlTypesTag.BETA

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_one_none", depends=["PymysqlPydanticClasses::get_one"])
    def test_get_one_none(self, queries_obj: queries.Queries) -> None:
        assert queries_obj.get_one_mysql_type(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_one_inner", depends=["PymysqlPydanticClasses::get_one_none"])
    def test_get_one_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        result = queries_obj.get_one_inner_mysql_type(table_id=TYPE_ID)

        assert result is not None
        assert isinstance(result, models.TestInnerMysqlType)
        assert result == inner_model
        assert result.tinyint1_test is False
        assert result.boolean_test is True
        assert result.json_test is None
        assert result.mood is enums.TestInnerMysqlTypesMood.VALUE__HIDDEN
        assert result.tag is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_one_inner_none", depends=["PymysqlPydanticClasses::get_one_inner"])
    def test_get_one_inner_none(self, queries_obj: queries.Queries) -> None:
        assert queries_obj.get_one_inner_mysql_type(table_id=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_date", depends=["PymysqlPydanticClasses::get_one_inner_none"])
    def test_get_date(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_one_date(id_=TYPE_ID, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, datetime.date)
        assert result == model.date_test
        assert queries_obj.get_one_date(id_=MISSING_ID, date_test=model.date_test) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_datetime", depends=["PymysqlPydanticClasses::get_date"])
    def test_get_datetime(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_one_datetime(id_=TYPE_ID, datetime_test=model.datetime_test)

        assert result is not None
        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test
        assert queries_obj.get_one_datetime(id_=MISSING_ID, datetime_test=model.datetime_test) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_time", depends=["PymysqlPydanticClasses::get_datetime"])
    def test_get_time(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_one_time(id_=TYPE_ID, time_test=model.time_test)

        assert result is not None
        # MySQL time maps to timedelta, not datetime.time
        assert isinstance(result, datetime.timedelta)
        assert result == model.time_test
        assert queries_obj.get_one_time(id_=MISSING_ID, time_test=model.time_test) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_bool", depends=["PymysqlPydanticClasses::get_time"])
    def test_get_bool(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.get_one_bool(id_=TYPE_ID, tinyint1_test=True)

        assert result is True
        assert queries_obj.get_one_bool(id_=MISSING_ID, tinyint1_test=True) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_decimal", depends=["PymysqlPydanticClasses::get_bool"])
    def test_get_decimal(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_one_decimal(id_=TYPE_ID, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, decimal.Decimal)
        # decimal(12,4) comes back padded to scale 4
        assert result == decimal.Decimal("12.3400")
        assert str(result) == "12.3400"
        assert queries_obj.get_one_decimal(id_=MISSING_ID, decimal_test=model.decimal_test) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_blob", depends=["PymysqlPydanticClasses::get_decimal"])
    def test_get_blob(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_one_blob(id_=TYPE_ID, blob_test=model.blob_test)

        assert result is not None
        assert isinstance(result, memoryview)
        assert result == model.blob_test
        assert queries_obj.get_one_blob(id_=MISSING_ID, blob_test=model.blob_test) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_bit", depends=["PymysqlPydanticClasses::get_blob"])
    def test_get_bit(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.get_one_bit(id_=TYPE_ID)

        assert result is not None
        # bit(8) comes back as a single byte
        assert isinstance(result, memoryview)
        assert bytes(result) == b"\x80"
        assert queries_obj.get_one_bit(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_year", depends=["PymysqlPydanticClasses::get_bit"])
    def test_get_year(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_one_year(id_=TYPE_ID)

        assert result is not None
        assert isinstance(result, int)
        assert result == model.year_test
        assert queries_obj.get_one_year(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_json", depends=["PymysqlPydanticClasses::get_year"])
    def test_get_json(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.get_one_json(id_=TYPE_ID)

        assert result is not None
        assert isinstance(result, str)
        assert json.loads(result) == {"foo": "bar"}
        assert queries_obj.get_one_json(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_mood", depends=["PymysqlPydanticClasses::get_json"])
    def test_get_mood(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.get_one_mood(id_=TYPE_ID, mood=enums.TestMysqlTypesMood.VALUE_24H)

        assert result is not None
        assert isinstance(result, enums.TestMysqlTypesMood)
        assert result is enums.TestMysqlTypesMood.VALUE_24H
        assert result == "24h"
        assert queries_obj.get_one_mood(id_=MISSING_ID, mood=enums.TestMysqlTypesMood.VALUE_24H) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_tag", depends=["PymysqlPydanticClasses::get_mood"])
    def test_get_tag(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.get_one_tag(id_=TYPE_ID)

        assert result is not None
        assert isinstance(result, enums.TestMysqlTypesTag)
        assert result is enums.TestMysqlTypesTag.BETA
        assert queries_obj.get_one_tag(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_many", depends=["PymysqlPydanticClasses::get_tag"])
    def test_get_many(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_mysql_type(id_=TYPE_ID)

        assert isinstance(result, queries.QueryResults)
        results = result()
        assert len(results) == 1
        assert isinstance(results[0], models.TestMysqlType)
        assert _without_json(results[0]) == _without_json(model)

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_many_iter", depends=["PymysqlPydanticClasses::get_many"])
    def test_get_many_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        for result in queries_obj.get_many_mysql_type(id_=TYPE_ID):
            assert isinstance(result, models.TestMysqlType)
            assert _without_json(result) == _without_json(model)

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_many_inner", depends=["PymysqlPydanticClasses::get_many_iter"])
    def test_get_many_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        result = queries_obj.get_many_inner_mysql_type(table_id=TYPE_ID)

        assert isinstance(result, queries.QueryResults)
        results = result()
        assert list(results) == [inner_model]
        for row in queries_obj.get_many_inner_mysql_type(table_id=TYPE_ID):
            assert isinstance(row, models.TestInnerMysqlType)
            assert row == inner_model

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_many_nullable_inner", depends=["PymysqlPydanticClasses::get_many_inner"])
    def test_get_many_nullable_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        # int_test is compared with <=>, so None matches the NULL row.
        result = queries_obj.get_many_nullable_inner_mysql_type(table_id=TYPE_ID, int_test=None)

        results = result()
        assert list(results) == [inner_model]
        assert list(queries_obj.get_many_nullable_inner_mysql_type(table_id=TYPE_ID, int_test=0)()) == []

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_many_date", depends=["PymysqlPydanticClasses::get_many_nullable_inner"])
    def test_get_many_date(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_date(id_=TYPE_ID, date_test=model.date_test)

        assert isinstance(result, queries.QueryResults)
        assert list(result()) == [model.date_test]
        assert list(queries_obj.get_many_date(id_=TYPE_ID, date_test=model.date_test)) == [model.date_test]

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_many_time", depends=["PymysqlPydanticClasses::get_many_date"])
    def test_get_many_time(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_time(id_=TYPE_ID, time_test=model.time_test)

        assert list(result()) == [model.time_test]
        assert list(queries_obj.get_many_time(id_=TYPE_ID, time_test=model.time_test)) == [model.time_test]

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_many_bool", depends=["PymysqlPydanticClasses::get_many_time"])
    def test_get_many_bool(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.get_many_bool(id_=TYPE_ID, tinyint1_test=True)

        results = result()
        assert len(results) == 1
        assert results[0] is True
        for row in queries_obj.get_many_bool(id_=TYPE_ID, tinyint1_test=True):
            assert row is True

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_many_decimal", depends=["PymysqlPydanticClasses::get_many_bool"])
    def test_get_many_decimal(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_decimal(id_=TYPE_ID, decimal_test=model.decimal_test)

        assert list(result()) == [decimal.Decimal("12.3400")]
        for row in queries_obj.get_many_decimal(id_=TYPE_ID, decimal_test=model.decimal_test):
            assert str(row) == "12.3400"

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_many_mood", depends=["PymysqlPydanticClasses::get_many_decimal"])
    def test_get_many_mood(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.get_many_mood(mood=enums.TestMysqlTypesMood.VALUE_24H)

        assert list(result()) == [enums.TestMysqlTypesMood.VALUE_24H]
        assert list(queries_obj.get_many_mood(mood=enums.TestMysqlTypesMood.VALUE_24H)) == [enums.TestMysqlTypesMood.VALUE_24H]

    @pytest.mark.dependency(name="PymysqlPydanticClasses::list_months", depends=["PymysqlPydanticClasses::get_many_mood"])
    def test_list_months(self, queries_obj: queries.Queries) -> None:
        # Regression for the percent-doubling bug: the parameterless :many
        # query contains literal % signs in DATE_FORMAT.
        result = queries_obj.list_months()

        assert list(result()) == ["2026-01"]
        assert list(queries_obj.list_months()) == ["2026-01"]

    @pytest.mark.dependency(name="PymysqlPydanticClasses::count", depends=["PymysqlPydanticClasses::list_months"])
    def test_count(self, queries_obj: queries.Queries) -> None:
        # The shared table may carry other files' rows; only a lower bound is safe.
        count = queries_obj.count_mysql_types()
        assert count is not None
        assert count >= 1

    @pytest.mark.dependency(name="PymysqlPydanticClasses::update_varchar", depends=["PymysqlPydanticClasses::count"])
    def test_update_varchar(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.update_varchar_test(varchar_test="updated varchar", id_=TYPE_ID)

        assert isinstance(result, int)
        # The shared table may carry other files' rows; only a lower bound is safe.
        assert result is not None
        assert result >= 1
        assert queries_obj.update_varchar_test(varchar_test="updated varchar", id_=MISSING_ID) == 0

    @pytest.mark.dependency(name="PymysqlPydanticClasses::all_cursor", depends=["PymysqlPydanticClasses::update_varchar"])
    def test_all_cursor(self, queries_obj: queries.Queries) -> None:
        cursor = queries_obj.all_mysql_types_cursor()

        assert isinstance(cursor, pymysql.cursors.Cursor)
        rows = cursor.fetchall()
        cursor.close()
        # The shared table may carry other files' rows; assert on our own.
        assert TYPE_ID in {row[0] for row in rows}

    @pytest.mark.dependency(name="PymysqlPydanticClasses::delete", depends=["PymysqlPydanticClasses::all_cursor"])
    def test_delete(self, queries_obj: queries.Queries, pymysql_conn: pymysql.Connection) -> None:
        queries_obj.delete_one_mysql_type(id_=TYPE_ID)

        assert queries_obj.get_one_mysql_type(id_=TYPE_ID) is None
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_inner_mysql_types WHERE table_id = %s", (TYPE_ID,))
        assert queries_obj.get_one_inner_mysql_type(table_id=TYPE_ID) is None

    def test_exec_last_id(self, queries_obj: queries.Queries, pymysql_conn: pymysql.Connection) -> None:
        # The AUTO_INCREMENT counter persists across runs, so only > 0 holds.
        last_id = queries_obj.insert_exec_last_id(name=EXEC_LAST_ID_NAME)

        assert isinstance(last_id, int)
        assert last_id > 0
        assert queries_obj.get_exec_last_id_name(id_=last_id) == EXEC_LAST_ID_NAME
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_execlastid WHERE id = %s", (last_id,))
        assert queries_obj.get_exec_last_id_name(id_=last_id) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::insert_type_override")
    def test_insert_type_override(self, queries_obj: queries.Queries, override_model: models.TestTypeOverride) -> None:
        queries_obj.insert_type_override(id_=override_model.id_, text_test=override_model.text_test)
        queries_obj.insert_type_override(id_=OVERRIDE_NONE_ID, text_test=None)

    @pytest.mark.dependency(name="PymysqlPydanticClasses::get_type_override", depends=["PymysqlPydanticClasses::insert_type_override"])
    def test_get_type_override(self, queries_obj: queries.Queries, override_model: models.TestTypeOverride) -> None:
        result = queries_obj.get_type_override(id_=OVERRIDE_ID)

        assert result is not None
        assert isinstance(result.text_test, UserString)
        assert result == override_model

        none_result = queries_obj.get_type_override(id_=OVERRIDE_NONE_ID)
        assert none_result is not None
        assert none_result.text_test is None
        assert queries_obj.get_type_override(id_=MISSING_ID) is None

    @pytest.mark.dependency(depends=["PymysqlPydanticClasses::get_type_override"])
    def test_delete_type_override(self, queries_obj: queries.Queries, pymysql_conn: pymysql.Connection) -> None:
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_type_override WHERE id IN (%s, %s)", (OVERRIDE_ID, OVERRIDE_NONE_ID))
        assert queries_obj.get_type_override(id_=OVERRIDE_ID) is None

    def test_reserved_arg(self, queries_obj: queries.Queries, pymysql_conn: pymysql.Connection) -> None:
        queries_obj.insert_reserved_arg(id_=RESERVED_ID, conn=RESERVED_CONN)

        result = queries_obj.get_reserved_arg(conn=RESERVED_CONN)
        assert result == models.TestReservedArg(id_=RESERVED_ID, conn=RESERVED_CONN)
        assert queries_obj.get_reserved_arg(conn="pydantic-classes-missing") is None
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_reserved_args WHERE id = %s", (RESERVED_ID,))

    @pytest.mark.dependency(name="PymysqlPydanticClasses::case_insert")
    def test_case_insert(self, case_obj: queries_case.QueriesCase) -> None:
        case_obj.insert_case_row(id_=CASE_IDS[0], upper_dt=CASE_DT, prec_dec=CASE_DEC)
        case_obj.insert_case_row(id_=CASE_IDS[1], upper_dt=CASE_DT, prec_dec=CASE_DEC)

    @pytest.mark.dependency(name="PymysqlPydanticClasses::case_get", depends=["PymysqlPydanticClasses::case_insert"])
    def test_case_get(self, case_obj: queries_case.QueriesCase) -> None:
        result = case_obj.get_case_row(id_=CASE_IDS[0])

        assert result is not None
        assert result == models.TestCaseSensitivity(id_=CASE_IDS[0], upper_dt=CASE_DT, prec_dec=CASE_DEC)
        assert case_obj.get_case_row(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::case_count", depends=["PymysqlPydanticClasses::case_get"])
    def test_case_count(self, case_obj: queries_case.QueriesCase) -> None:
        # The WHERE clause lives inside an executable /*! version comment; if
        # MySQL ignored it both counts would be 2.
        # Range-scoped asserts: the shared table may carry other files' rows,
        # so counts outside [CASE_IDS[0], CASE_IDS[1]] must cancel out.
        beyond = case_obj.count_case_rows(id_=CASE_IDS[1] + 1)
        high = case_obj.count_case_rows(id_=CASE_IDS[1])
        low = case_obj.count_case_rows(id_=CASE_IDS[0])
        assert beyond is not None
        assert high is not None
        assert low is not None
        assert high - beyond == 1
        assert low - beyond == len(CASE_IDS)

    @pytest.mark.dependency(depends=["PymysqlPydanticClasses::case_count"])
    def test_case_delete(self, case_obj: queries_case.QueriesCase, pymysql_conn: pymysql.Connection) -> None:
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_case_sensitivity WHERE id IN (%s, %s)", CASE_IDS)
        beyond = case_obj.count_case_rows(id_=CASE_IDS[1] + 1)
        low = case_obj.count_case_rows(id_=CASE_IDS[0])
        assert beyond is not None
        assert low is not None
        assert low - beyond == 0

    @pytest.mark.dependency(name="PymysqlPydanticClasses::enum_insert")
    def test_enum_override_insert(self, enum_obj: queries_enum_override.QueriesEnumOverride) -> None:
        # The overridden parameter is a plain str; the generated code converts
        # it back through enums.TestEnumOverrideMoodTest.
        enum_obj.insert_enum_override(id_=ENUM_IDS[0], mood_test="happy")
        enum_obj.insert_enum_override(id_=ENUM_IDS[1], mood_test="sad")
        with pytest.raises(ValueError, match="angry"):
            enum_obj.insert_enum_override(id_=MISSING_ID, mood_test="angry")

    @pytest.mark.dependency(name="PymysqlPydanticClasses::enum_get", depends=["PymysqlPydanticClasses::enum_insert"])
    def test_enum_override_get(self, enum_obj: queries_enum_override.QueriesEnumOverride) -> None:
        mood = enum_obj.get_enum_override_mood(id_=ENUM_IDS[0])

        assert mood is not None
        assert isinstance(mood, str)
        assert mood == "happy"
        assert enum_obj.get_enum_override_mood(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::enum_list", depends=["PymysqlPydanticClasses::enum_insert"])
    def test_enum_override_list(self, enum_obj: queries_enum_override.QueriesEnumOverride) -> None:
        rows = enum_obj.list_enum_override_by_ids(ids=list(ENUM_IDS))()

        assert all(isinstance(row, models.TestEnumOverride) for row in rows)
        assert {row.id_: row.mood_test for row in rows} == {ENUM_IDS[0]: "happy", ENUM_IDS[1]: "sad"}

    @pytest.mark.dependency(name="PymysqlPydanticClasses::enum_iterate", depends=["PymysqlPydanticClasses::enum_insert"])
    def test_enum_override_iterate(self, enum_obj: queries_enum_override.QueriesEnumOverride) -> None:
        seen: dict[int, str] = {}
        for row in enum_obj.list_enum_override_by_ids(ids=list(ENUM_IDS)):
            assert isinstance(row, models.TestEnumOverride)
            seen[row.id_] = row.mood_test
        assert seen == {ENUM_IDS[0]: "happy", ENUM_IDS[1]: "sad"}

    @pytest.mark.dependency(name="PymysqlPydanticClasses::enum_empty", depends=["PymysqlPydanticClasses::enum_insert"])
    def test_enum_override_empty_slice(self, enum_obj: queries_enum_override.QueriesEnumOverride) -> None:
        assert list(enum_obj.list_enum_override_by_ids(ids=[])()) == []
        assert list(enum_obj.list_enum_override_by_ids(ids=[])) == []

    @pytest.mark.dependency(depends=["PymysqlPydanticClasses::enum_empty"])
    def test_enum_override_delete(self, enum_obj: queries_enum_override.QueriesEnumOverride, pymysql_conn: pymysql.Connection) -> None:
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_enum_override WHERE id IN (%s, %s)", ENUM_IDS)
        assert enum_obj.get_enum_override_mood(id_=ENUM_IDS[0]) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::field_insert")
    def test_field_naming_insert(self, pymysql_conn: pymysql.Connection) -> None:
        # There is no generated insert for this table.
        with pymysql_conn.cursor() as cur:
            cur.execute("INSERT INTO test_field_namings (id, outputs) VALUES (%s, %s)", (FIELD_ID, json.dumps(["first", "second"])))

    @pytest.mark.dependency(name="PymysqlPydanticClasses::field_get", depends=["PymysqlPydanticClasses::field_insert"])
    def test_field_naming_get(self, field_obj: queries_field_namings.QueriesFieldNamings) -> None:
        result = field_obj.get_field_naming(id_=FIELD_ID)

        assert result is not None
        assert isinstance(result, models.TestFieldNaming)
        assert result.id_ == FIELD_ID
        assert json.loads(result.outputs) == ["first", "second"]
        assert field_obj.get_field_naming(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::field_joined", depends=["PymysqlPydanticClasses::field_get"])
    def test_field_naming_joined(self, field_obj: queries_field_namings.QueriesFieldNamings) -> None:
        result = field_obj.get_joined_field_namings(id_=FIELD_ID)

        assert result is not None
        assert isinstance(result, queries_field_namings.GetJoinedFieldNamingsRow)
        assert json.loads(result.outputs) == ["first", "second"]
        assert json.loads(result.outputs_2) == ["first", "second"]

    @pytest.mark.dependency(name="PymysqlPydanticClasses::field_set", depends=["PymysqlPydanticClasses::field_joined"])
    def test_field_naming_set(self, field_obj: queries_field_namings.QueriesFieldNamings) -> None:
        field_obj.set_field_naming_outputs(outputs=json.dumps({"count": 2}), id_=FIELD_ID)

        result = field_obj.get_field_naming(id_=FIELD_ID)
        assert result is not None
        assert json.loads(result.outputs) == {"count": 2}

    @pytest.mark.dependency(depends=["PymysqlPydanticClasses::field_set"])
    def test_field_naming_delete(self, field_obj: queries_field_namings.QueriesFieldNamings, pymysql_conn: pymysql.Connection) -> None:
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_field_namings WHERE id = %s", (FIELD_ID,))
        assert field_obj.get_field_naming(id_=FIELD_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::invalid_insert")
    def test_invalid_identifiers_insert(self, invalid_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers) -> None:
        invalid_obj.insert_invalid_identifiers(id_=INVALID_ID, column_3p_="3p value", new_notes="note value")
        invalid_obj.insert_third_party_stat(id_=THIRD_PARTY_ID, total=987)

    @pytest.mark.dependency(name="PymysqlPydanticClasses::invalid_get", depends=["PymysqlPydanticClasses::invalid_insert"])
    def test_invalid_identifiers_get(self, invalid_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers) -> None:
        result = invalid_obj.get_invalid_identifiers(id_=INVALID_ID)

        assert result == models.TestInvalidIdentifier(id_=INVALID_ID, column_3p_="3p value", new_notes="note value", column__pct=None)
        assert invalid_obj.get_invalid_identifiers(id_=MISSING_ID) is None

    @pytest.mark.dependency(name="PymysqlPydanticClasses::third_party_get", depends=["PymysqlPydanticClasses::invalid_insert"])
    def test_third_party_stat_get(self, invalid_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers) -> None:
        result = invalid_obj.get_third_party_stat(id_=THIRD_PARTY_ID)

        assert result == models.Model3RdPartyStat(id_=THIRD_PARTY_ID, total=987)
        assert invalid_obj.get_third_party_stat(id_=MISSING_ID) is None

    @pytest.mark.dependency(depends=["PymysqlPydanticClasses::invalid_get", "PymysqlPydanticClasses::third_party_get"])
    def test_invalid_identifiers_delete(self, invalid_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers, pymysql_conn: pymysql.Connection) -> None:
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_invalid_identifiers WHERE id = %s", (INVALID_ID,))
            cur.execute("DELETE FROM `3rd_party_stats` WHERE id = %s", (THIRD_PARTY_ID,))
        assert invalid_obj.get_invalid_identifiers(id_=INVALID_ID) is None
        assert invalid_obj.get_third_party_stat(id_=THIRD_PARTY_ID) is None

    def test_one_missing_rows_return_none(self, pymysql_conn: pymysql.Connection) -> None:
        # Every :one not-found branch, plus the no-insert :execlastid. The
        # count queries always return a row, so their miss branch needs the
        # no-row stub; the sub-module Querier conn properties ride along.
        obj = queries.Queries(conn=pymysql_conn)
        assert obj.get_one_mysql_type(id_=-1) is None
        assert obj.get_one_inner_mysql_type(table_id=-1) is None
        assert obj.get_one_date(id_=-1, date_test=datetime.date(1970, 1, 1)) is None
        assert obj.get_one_datetime(id_=-1, datetime_test=datetime.datetime(1970, 1, 1)) is None
        assert obj.get_one_time(id_=-1, time_test=datetime.timedelta()) is None
        assert obj.get_one_bool(id_=-1, tinyint1_test=False) is None
        assert obj.get_one_decimal(id_=-1, decimal_test=decimal.Decimal(0)) is None
        assert obj.get_one_blob(id_=-1, blob_test=memoryview(b"")) is None
        assert obj.get_one_bit(id_=-1) is None
        assert obj.get_one_year(id_=-1) is None
        assert obj.get_one_json(id_=-1) is None
        assert obj.get_one_mood(id_=-1, mood=enums.TestMysqlTypesMood.SAD) is None
        assert obj.get_one_tag(id_=-1) is None
        assert obj.get_exec_last_id_name(id_=-1) is None
        assert obj.get_type_override(id_=-1) is None
        assert obj.get_reserved_arg(conn="missing") is None
        assert obj.touch_exec_last_id(name="untouched", id_=-1) is None

        case_obj = queries_case.QueriesCase(conn=pymysql_conn)
        naming_obj = queries_field_namings.QueriesFieldNamings(conn=pymysql_conn)
        invalid_obj = queries_invalid_identifiers.QueriesInvalidIdentifiers(conn=pymysql_conn)
        enum_obj = queries_enum_override.QueriesEnumOverride(conn=pymysql_conn)
        assert case_obj.conn is pymysql_conn
        assert naming_obj.conn is pymysql_conn
        assert invalid_obj.conn is pymysql_conn
        assert enum_obj.conn is pymysql_conn
        assert case_obj.get_case_row(id_=-1) is None
        assert naming_obj.get_field_naming(id_=-1) is None
        assert naming_obj.get_joined_field_namings(id_=-1) is None
        assert invalid_obj.get_invalid_identifiers(id_=-1) is None
        assert enum_obj.get_enum_override_mood(id_=-1) is None
        assert enum_obj.count_enum_override_by_moods(moods=[]) == 0

        stub = typing.cast("pymysql.Connection", no_row_conn.NoRowConn())
        assert queries.Queries(conn=stub).count_mysql_types() is None
        assert queries_case.QueriesCase(conn=stub).count_case_rows(id_=0) is None
        assert queries_enum_override.QueriesEnumOverride(conn=stub).count_enum_override_by_moods(moods=[]) is None
