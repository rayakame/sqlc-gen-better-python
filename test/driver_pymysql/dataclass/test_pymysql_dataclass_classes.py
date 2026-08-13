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

import pymysql
import pymysql.cursors
import pytest

from test.driver_pymysql.dataclass.classes import enums
from test.driver_pymysql.dataclass.classes import models
from test.driver_pymysql.dataclass.classes import queries
from test.driver_pymysql.dataclass.classes import queries_case
from test.driver_pymysql.dataclass.classes import queries_enum_override
from test.driver_pymysql.dataclass.classes import queries_field_namings
from test.driver_pymysql.dataclass.classes import queries_invalid_identifiers

# Fixed ids: the MySQL tables are shared by every pymysql/asyncmy suite in the
# session, so each test file owns a distinct id range. This file: 1000-1499.
MAIN_ID = 1000
TYPE_OVERRIDE_ID = 1100
TYPE_OVERRIDE_NONE_ID = 1101
ENUM_OVERRIDE_ID = 1150
ENUM_OVERRIDE_ID_2 = 1151
CASE_ID = 1200
RESERVED_ARG_ID = 1250
FIELD_NAMING_ID = 1300
INVALID_IDENTIFIER_ID = 1350
THIRD_PARTY_ID = 1360
THIRD_PARTY_TOTAL = 9000

CASE_DT = datetime.datetime(2026, 7, 19, 8, 15)
CASE_DEC = decimal.Decimal("12.34")
RESERVED_ARG_VALUE = "pymysql-dataclass-classes-conn"
EXEC_LAST_ID_NAME = "pymysql-dataclass-classes"
UPDATED_VARCHAR = "updated varchar"
# decimal(12,4) and numeric(10,2) come back padded to their full scale.
DECIMAL_PADDED = "1234.5000"
NUMERIC_PADDED = "87.60"
EXPECTED_MONTH = "2026-01"


class TestPymysqlDataclassClasses:
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

    @pytest.fixture(scope="class")
    def queries_obj(self, pymysql_conn: pymysql.Connection) -> queries.Queries:
        return queries.Queries(conn=pymysql_conn)

    @pytest.fixture(scope="class")
    def case_obj(self, pymysql_conn: pymysql.Connection) -> queries_case.QueriesCase:
        return queries_case.QueriesCase(conn=pymysql_conn)

    @pytest.fixture(scope="class")
    def enum_override_obj(self, pymysql_conn: pymysql.Connection) -> queries_enum_override.QueriesEnumOverride:
        return queries_enum_override.QueriesEnumOverride(conn=pymysql_conn)

    @pytest.fixture(scope="class")
    def field_namings_obj(self, pymysql_conn: pymysql.Connection) -> queries_field_namings.QueriesFieldNamings:
        return queries_field_namings.QueriesFieldNamings(conn=pymysql_conn)

    @pytest.fixture(scope="class")
    def invalid_identifiers_obj(self, pymysql_conn: pymysql.Connection) -> queries_invalid_identifiers.QueriesInvalidIdentifiers:
        return queries_invalid_identifiers.QueriesInvalidIdentifiers(conn=pymysql_conn)

    def test_conn_attr(self, queries_obj: queries.Queries) -> None:
        assert isinstance(queries_obj.conn, pymysql.Connection)

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::insert")
    def test_insert(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
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

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::inner_insert", depends=["PymysqlTestDataclassClasses::insert"])
    def test_inner_insert(
        self,
        queries_obj: queries.Queries,
        inner_model: models.TestInnerMysqlType,
    ) -> None:
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

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_one", depends=["PymysqlTestDataclassClasses::inner_insert"])
    def test_get_one(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_mysql_type(id_=model.id_)

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

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_one_none", depends=["PymysqlTestDataclassClasses::get_one"])
    def test_get_one_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = queries_obj.get_one_mysql_type(id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_one_inner", depends=["PymysqlTestDataclassClasses::get_one_none"])
    def test_get_one_inner(
        self,
        queries_obj: queries.Queries,
        inner_model: models.TestInnerMysqlType,
    ) -> None:
        result = queries_obj.get_one_inner_mysql_type(table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, models.TestInnerMysqlType)
        assert result == inner_model

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_one_inner_none", depends=["PymysqlTestDataclassClasses::get_one_inner"])
    def test_get_one_inner_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = queries_obj.get_one_inner_mysql_type(table_id=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_date", depends=["PymysqlTestDataclassClasses::get_one_inner_none"])
    def test_get_date(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_date(id_=model.id_, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, datetime.date)
        assert result == model.date_test

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_date_none", depends=["PymysqlTestDataclassClasses::get_date"])
    def test_get_date_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_date(id_=0, date_test=model.date_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_datetime", depends=["PymysqlTestDataclassClasses::get_date_none"])
    def test_get_datetime(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_datetime(id_=model.id_, datetime_test=model.datetime_test)

        assert result is not None
        assert isinstance(result, datetime.datetime)
        assert result == model.datetime_test

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_datetime_none", depends=["PymysqlTestDataclassClasses::get_datetime"])
    def test_get_datetime_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_datetime(id_=0, datetime_test=model.datetime_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_time", depends=["PymysqlTestDataclassClasses::get_datetime_none"])
    def test_get_time(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_time(id_=model.id_, time_test=model.time_test)

        assert result is not None
        assert isinstance(result, datetime.timedelta)
        assert result == model.time_test

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_time_none", depends=["PymysqlTestDataclassClasses::get_time"])
    def test_get_time_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_time(id_=0, time_test=model.time_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_bool", depends=["PymysqlTestDataclassClasses::get_time_none"])
    def test_get_bool(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_bool(id_=model.id_, tinyint1_test=model.tinyint1_test)

        assert result is not None
        assert isinstance(result, bool)
        assert result is True

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_bool_none", depends=["PymysqlTestDataclassClasses::get_bool"])
    def test_get_bool_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = queries_obj.get_one_bool(id_=0, tinyint1_test=False)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_decimal", depends=["PymysqlTestDataclassClasses::get_bool_none"])
    def test_get_decimal(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_decimal(id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, decimal.Decimal)
        assert result == model.decimal_test
        assert str(result) == DECIMAL_PADDED

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_decimal_none", depends=["PymysqlTestDataclassClasses::get_decimal"])
    def test_get_decimal_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_decimal(id_=0, decimal_test=model.decimal_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_blob", depends=["PymysqlTestDataclassClasses::get_decimal_none"])
    def test_get_blob(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_blob(id_=model.id_, blob_test=model.blob_test)

        assert result is not None
        assert isinstance(result, memoryview)
        assert result == model.blob_test

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_blob_none", depends=["PymysqlTestDataclassClasses::get_blob"])
    def test_get_blob_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_blob(id_=0, blob_test=model.blob_test)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_bit", depends=["PymysqlTestDataclassClasses::get_blob_none"])
    def test_get_bit(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_bit(id_=model.id_)

        assert result is not None
        assert isinstance(result, memoryview)
        assert bytes(result) == b"\x80"

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_bit_none", depends=["PymysqlTestDataclassClasses::get_bit"])
    def test_get_bit_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = queries_obj.get_one_bit(id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_year", depends=["PymysqlTestDataclassClasses::get_bit_none"])
    def test_get_year(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_year(id_=model.id_)

        assert result is not None
        assert isinstance(result, int)
        assert result == model.year_test

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_year_none", depends=["PymysqlTestDataclassClasses::get_year"])
    def test_get_year_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = queries_obj.get_one_year(id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_json", depends=["PymysqlTestDataclassClasses::get_year_none"])
    def test_get_json(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_json(id_=model.id_)

        assert result is not None
        assert isinstance(result, str)
        # MySQL normalizes json spacing; never compare the raw strings.
        assert json.loads(result) == json.loads(model.json_test)

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_json_none", depends=["PymysqlTestDataclassClasses::get_json"])
    def test_get_json_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = queries_obj.get_one_json(id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_mood", depends=["PymysqlTestDataclassClasses::get_json_none"])
    def test_get_mood(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_mood(id_=model.id_, mood=model.mood)

        assert result is not None
        assert isinstance(result, enums.TestMysqlTypesMood)
        assert result is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_mood_none", depends=["PymysqlTestDataclassClasses::get_mood"])
    def test_get_mood_none(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_mood(id_=0, mood=model.mood)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_tag", depends=["PymysqlTestDataclassClasses::get_mood_none"])
    def test_get_tag(
        self,
        queries_obj: queries.Queries,
        model: models.TestMysqlType,
    ) -> None:
        result = queries_obj.get_one_tag(id_=model.id_)

        assert result is not None
        assert isinstance(result, enums.TestMysqlTypesTag)
        assert result is enums.TestMysqlTypesTag.BETA
        assert result == model.tag

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_tag_none", depends=["PymysqlTestDataclassClasses::get_tag"])
    def test_get_tag_none(
        self,
        queries_obj: queries.Queries,
    ) -> None:
        result = queries_obj.get_one_tag(id_=0)

        assert result is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many", depends=["PymysqlTestDataclassClasses::get_tag_none"])
    def test_get_many(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_mysql_type(id_=model.id_)

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

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many_iter", depends=["PymysqlTestDataclassClasses::get_many"])
    def test_get_many_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        for result in queries_obj.get_many_mysql_type(id_=model.id_):
            assert result is not None
            assert isinstance(result, models.TestMysqlType)
            assert dataclasses.replace(result, json_test=model.json_test) == model

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many_inner", depends=["PymysqlTestDataclassClasses::get_many_iter"])
    def test_get_many_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        result = queries_obj.get_many_inner_mysql_type(table_id=inner_model.table_id)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], models.TestInnerMysqlType)
        assert results[0] == inner_model

        results = result()
        assert results[0] == inner_model

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many_inner_iter", depends=["PymysqlTestDataclassClasses::get_many_inner"])
    def test_get_many_inner_iter(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        for result in queries_obj.get_many_inner_mysql_type(table_id=inner_model.table_id):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)
            assert result == inner_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_many_nullable_inner",
        depends=["PymysqlTestDataclassClasses::get_many_inner_iter"],
    )
    def test_get_many_nullable_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        # int_test is None; the query uses the NULL-safe <=> comparison.
        result = queries_obj.get_many_nullable_inner_mysql_type(table_id=inner_model.table_id, int_test=inner_model.int_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = result()
        assert isinstance(results[0], models.TestInnerMysqlType)
        assert results[0] == inner_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_many_nullable_inner_iter",
        depends=["PymysqlTestDataclassClasses::get_many_nullable_inner"],
    )
    def test_get_many_nullable_inner_iter(self, queries_obj: queries.Queries, inner_model: models.TestInnerMysqlType) -> None:
        for result in queries_obj.get_many_nullable_inner_mysql_type(table_id=inner_model.table_id, int_test=inner_model.int_test):
            assert result is not None
            assert isinstance(result, models.TestInnerMysqlType)
            assert result == inner_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_many_date",
        depends=["PymysqlTestDataclassClasses::get_many_nullable_inner_iter"],
    )
    def test_get_many_date(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_date(id_=model.id_, date_test=model.date_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.date)
        assert results[0] == model.date_test

        results = result()
        assert results[0] == model.date_test

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many_date_iter", depends=["PymysqlTestDataclassClasses::get_many_date"])
    def test_get_many_date_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        for result in queries_obj.get_many_date(id_=model.id_, date_test=model.date_test):
            assert result is not None
            assert isinstance(result, datetime.date)
            assert result == model.date_test

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many_time", depends=["PymysqlTestDataclassClasses::get_many_date_iter"])
    def test_get_many_time(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_time(id_=model.id_, time_test=model.time_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], datetime.timedelta)
        assert results[0] == model.time_test

        results = result()
        assert results[0] == model.time_test

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many_time_iter", depends=["PymysqlTestDataclassClasses::get_many_time"])
    def test_get_many_time_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        for result in queries_obj.get_many_time(id_=model.id_, time_test=model.time_test):
            assert result is not None
            assert isinstance(result, datetime.timedelta)
            assert result == model.time_test

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many_bool", depends=["PymysqlTestDataclassClasses::get_many_time_iter"])
    def test_get_many_bool(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_bool(id_=model.id_, tinyint1_test=model.tinyint1_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], bool)
        assert results[0] is True

        results = result()
        assert results[0] is True

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many_bool_iter", depends=["PymysqlTestDataclassClasses::get_many_bool"])
    def test_get_many_bool_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        for result in queries_obj.get_many_bool(id_=model.id_, tinyint1_test=model.tinyint1_test):
            assert result is not None
            assert isinstance(result, bool)
            assert result is True

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_many_decimal",
        depends=["PymysqlTestDataclassClasses::get_many_bool_iter"],
    )
    def test_get_many_decimal(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_decimal(id_=model.id_, decimal_test=model.decimal_test)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert isinstance(results[0], decimal.Decimal)
        assert results[0] == model.decimal_test
        assert str(results[0]) == DECIMAL_PADDED

        results = result()
        assert results[0] == model.decimal_test

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_many_decimal_iter",
        depends=["PymysqlTestDataclassClasses::get_many_decimal"],
    )
    def test_get_many_decimal_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        for result in queries_obj.get_many_decimal(id_=model.id_, decimal_test=model.decimal_test):
            assert result is not None
            assert isinstance(result, decimal.Decimal)
            assert result == model.decimal_test

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_many_mood",
        depends=["PymysqlTestDataclassClasses::get_many_decimal_iter"],
    )
    def test_get_many_mood(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.get_many_mood(mood=model.mood)

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        results = list(result)
        assert len(results) == 1
        assert isinstance(results[0], enums.TestMysqlTypesMood)
        assert results[0] is enums.TestMysqlTypesMood.VALUE_24H

        results = result()
        assert results[0] is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_many_mood_iter", depends=["PymysqlTestDataclassClasses::get_many_mood"])
    def test_get_many_mood_iter(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        for result in queries_obj.get_many_mood(mood=model.mood):
            assert result is not None
            assert isinstance(result, enums.TestMysqlTypesMood)
            assert result is enums.TestMysqlTypesMood.VALUE_24H

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::list_months", depends=["PymysqlTestDataclassClasses::get_many_mood_iter"])
    def test_list_months(self, queries_obj: queries.Queries) -> None:
        # DATE_FORMAT emits %% in the stored SQL; the empty argument tuple
        # still goes through pymysql's %-substitution, halving it back.
        result = queries_obj.list_months()

        assert result is not None
        assert isinstance(result, queries.QueryResults)
        months = result()
        assert list(months) == [EXPECTED_MONTH]

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::list_months_iter", depends=["PymysqlTestDataclassClasses::list_months"])
    def test_list_months_iter(self, queries_obj: queries.Queries) -> None:
        months = list(queries_obj.list_months())
        assert months == [EXPECTED_MONTH]

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::count", depends=["PymysqlTestDataclassClasses::list_months_iter"])
    def test_count_mysql_types(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.count_mysql_types()

        assert result == 1

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::all_cursor", depends=["PymysqlTestDataclassClasses::count"])
    def test_all_mysql_types_cursor(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        cur = queries_obj.all_mysql_types_cursor()

        assert isinstance(cur, pymysql.cursors.Cursor)
        rows = cur.fetchall()
        assert len(rows) == 1
        assert rows[0][0] == model.id_
        cur.close()

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::update_rows", depends=["PymysqlTestDataclassClasses::all_cursor"])
    def test_update_varchar_rows(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        result = queries_obj.update_varchar_test(varchar_test=UPDATED_VARCHAR, id_=model.id_)

        assert isinstance(result, int)
        assert result == 1

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::update_rows_noop", depends=["PymysqlTestDataclassClasses::update_rows"])
    def test_update_varchar_rows_noop(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        # Without CLIENT.FOUND_ROWS MySQL reports changed rows, so setting
        # the same value again affects nothing.
        result = queries_obj.update_varchar_test(varchar_test=UPDATED_VARCHAR, id_=model.id_)

        assert result == 0

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::insert_last_id", depends=["PymysqlTestDataclassClasses::update_rows_noop"])
    def test_insert_exec_last_id(self, queries_obj: queries.Queries) -> None:
        # The AUTO_INCREMENT counter persists across runs; never assert an
        # exact id.
        new_id = queries_obj.insert_exec_last_id(name=EXEC_LAST_ID_NAME)

        assert new_id is not None
        assert isinstance(new_id, int)
        assert new_id > 0
        assert queries_obj.get_exec_last_id_name(id_=new_id) == EXEC_LAST_ID_NAME

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::delete", depends=["PymysqlTestDataclassClasses::insert_last_id"])
    def test_delete_mysql_type(self, queries_obj: queries.Queries, model: models.TestMysqlType) -> None:
        queries_obj.delete_one_mysql_type(id_=model.id_)

        assert queries_obj.get_one_mysql_type(id_=model.id_) is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::insert_type_override")
    def test_insert_type_override(self, queries_obj: queries.Queries, override_model: models.TestTypeOverride) -> None:
        queries_obj.insert_type_override(id_=override_model.id_, text_test=override_model.text_test)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_type_override",
        depends=["PymysqlTestDataclassClasses::insert_type_override"],
    )
    def test_get_type_override(self, queries_obj: queries.Queries, override_model: models.TestTypeOverride) -> None:
        result = queries_obj.get_type_override(id_=override_model.id_)

        assert result is not None
        assert isinstance(result.text_test, UserString)
        assert result == override_model

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_type_override_none_value",
        depends=["PymysqlTestDataclassClasses::get_type_override"],
    )
    def test_get_type_override_none_value(self, queries_obj: queries.Queries) -> None:
        # The override target is nullable: NULL must come back as None
        # without passing through UserString.
        queries_obj.insert_type_override(id_=TYPE_OVERRIDE_NONE_ID, text_test=None)
        result = queries_obj.get_type_override(id_=TYPE_OVERRIDE_NONE_ID)

        assert result is not None
        assert result.text_test is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_type_override_not_found",
        depends=["PymysqlTestDataclassClasses::get_type_override_none_value"],
    )
    def test_get_type_override_not_found(self, queries_obj: queries.Queries) -> None:
        assert queries_obj.get_type_override(id_=0) is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::insert_reserved_arg")
    def test_insert_reserved_arg(self, queries_obj: queries.Queries) -> None:
        # The column is literally named "conn"; on methods the parameter
        # keeps its name because self is the only implicit argument.
        queries_obj.insert_reserved_arg(id_=RESERVED_ARG_ID, conn=RESERVED_ARG_VALUE)

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_reserved_arg", depends=["PymysqlTestDataclassClasses::insert_reserved_arg"])
    def test_get_reserved_arg(self, queries_obj: queries.Queries) -> None:
        result = queries_obj.get_reserved_arg(conn=RESERVED_ARG_VALUE)

        assert result == models.TestReservedArg(id_=RESERVED_ARG_ID, conn=RESERVED_ARG_VALUE)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_reserved_arg_not_found",
        depends=["PymysqlTestDataclassClasses::get_reserved_arg"],
    )
    def test_get_reserved_arg_not_found(self, queries_obj: queries.Queries) -> None:
        assert queries_obj.get_reserved_arg(conn="missing-reserved-arg-value") is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::insert_case_rows")
    def test_insert_case_rows(self, case_obj: queries_case.QueriesCase) -> None:
        case_obj.insert_case_row(id_=CASE_ID, upper_dt=CASE_DT, prec_dec=CASE_DEC)
        case_obj.insert_case_row(id_=CASE_ID + 1, upper_dt=CASE_DT, prec_dec=CASE_DEC)

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_case_row", depends=["PymysqlTestDataclassClasses::insert_case_rows"])
    def test_get_case_row(self, case_obj: queries_case.QueriesCase) -> None:
        row = case_obj.get_case_row(id_=CASE_ID)

        assert row is not None
        assert isinstance(row.upper_dt, datetime.datetime)
        assert row.upper_dt == CASE_DT
        assert isinstance(row.prec_dec, decimal.Decimal)
        assert row.prec_dec == CASE_DEC

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_case_row_not_found",
        depends=["PymysqlTestDataclassClasses::get_case_row"],
    )
    def test_get_case_row_not_found(self, case_obj: queries_case.QueriesCase) -> None:
        assert case_obj.get_case_row(id_=0) is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::count_case_rows_filters",
        depends=["PymysqlTestDataclassClasses::get_case_row_not_found"],
    )
    def test_count_case_rows_filters(self, case_obj: queries_case.QueriesCase) -> None:
        # The WHERE clause lives inside an executable /*! version comment;
        # raising the threshold by one must drop exactly the first row.
        count_ge_first = case_obj.count_case_rows(id_=CASE_ID)
        count_ge_second = case_obj.count_case_rows(id_=CASE_ID + 1)

        assert count_ge_first is not None
        assert count_ge_second is not None
        assert count_ge_first - count_ge_second == 1

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::insert_enum_override")
    def test_insert_enum_override(self, enum_override_obj: queries_enum_override.QueriesEnumOverride) -> None:
        enum_override_obj.insert_enum_override(id_=ENUM_OVERRIDE_ID, mood_test="happy")
        enum_override_obj.insert_enum_override(id_=ENUM_OVERRIDE_ID_2, mood_test="sad")

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::count_enum_override_by_moods",
        depends=["PymysqlTestDataclassClasses::insert_enum_override"],
    )
    def test_count_enum_override_by_moods(self, enum_override_obj: queries_enum_override.QueriesEnumOverride) -> None:
        # Each slice element converts back through the enum class before
        # binding; an invalid member raises before any SQL runs. Counts are
        # relative: the shared table carries other files' rows too.
        both = enum_override_obj.count_enum_override_by_moods(moods=("happy", "sad"))
        happy = enum_override_obj.count_enum_override_by_moods(moods=["happy"])
        sad = enum_override_obj.count_enum_override_by_moods(moods=["sad"])
        assert both is not None
        assert happy is not None
        assert sad is not None
        assert happy >= 1
        assert sad >= 1
        assert both == happy + sad
        assert enum_override_obj.count_enum_override_by_moods(moods=[]) == 0
        with pytest.raises(ValueError, match="not-a-mood"):
            enum_override_obj.count_enum_override_by_moods(moods=["not-a-mood"])

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_enum_override_mood",
        depends=["PymysqlTestDataclassClasses::insert_enum_override"],
    )
    def test_get_enum_override_mood(self, enum_override_obj: queries_enum_override.QueriesEnumOverride) -> None:
        result = enum_override_obj.get_enum_override_mood(id_=ENUM_OVERRIDE_ID)

        assert result == "happy"
        assert isinstance(result, str)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_enum_override_mood_not_found",
        depends=["PymysqlTestDataclassClasses::get_enum_override_mood"],
    )
    def test_get_enum_override_mood_not_found(self, enum_override_obj: queries_enum_override.QueriesEnumOverride) -> None:
        assert enum_override_obj.get_enum_override_mood(id_=0) is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::list_enum_override_by_ids",
        depends=["PymysqlTestDataclassClasses::get_enum_override_mood_not_found"],
    )
    def test_list_enum_override_by_ids(self, enum_override_obj: queries_enum_override.QueriesEnumOverride) -> None:
        result = enum_override_obj.list_enum_override_by_ids(ids=[ENUM_OVERRIDE_ID, ENUM_OVERRIDE_ID_2])

        assert isinstance(result, queries_enum_override.QueryResults)
        rows = result()
        assert rows == [
            models.TestEnumOverride(id_=ENUM_OVERRIDE_ID, mood_test="happy"),
            models.TestEnumOverride(id_=ENUM_OVERRIDE_ID_2, mood_test="sad"),
        ]
        assert list(result) == rows

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::list_enum_override_by_ids_empty",
        depends=["PymysqlTestDataclassClasses::list_enum_override_by_ids"],
    )
    def test_list_enum_override_by_ids_empty(self, enum_override_obj: queries_enum_override.QueriesEnumOverride) -> None:
        # An empty slice expands the placeholder to NULL: IN (NULL) matches
        # no rows instead of raising.
        assert enum_override_obj.list_enum_override_by_ids(ids=[])() == []

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::seed_field_naming")
    def test_seed_field_naming(self, pymysql_conn: pymysql.Connection) -> None:
        # No generated insert exists for this table; seed it directly.
        with pymysql_conn.cursor() as cur:
            cur.execute("INSERT INTO test_field_namings (id, outputs) VALUES (%s, %s)", (FIELD_NAMING_ID, json.dumps({"first": 1})))

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::get_field_naming", depends=["PymysqlTestDataclassClasses::seed_field_naming"])
    def test_get_field_naming(self, field_namings_obj: queries_field_namings.QueriesFieldNamings) -> None:
        result = field_namings_obj.get_field_naming(id_=FIELD_NAMING_ID)

        assert result is not None
        assert result.id_ == FIELD_NAMING_ID
        assert json.loads(result.outputs) == {"first": 1}

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_field_naming_not_found",
        depends=["PymysqlTestDataclassClasses::get_field_naming"],
    )
    def test_get_field_naming_not_found(self, field_namings_obj: queries_field_namings.QueriesFieldNamings) -> None:
        assert field_namings_obj.get_field_naming(id_=0) is None

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_joined_field_namings",
        depends=["PymysqlTestDataclassClasses::get_field_naming_not_found"],
    )
    def test_get_joined_field_namings(self, field_namings_obj: queries_field_namings.QueriesFieldNamings) -> None:
        result = field_namings_obj.get_joined_field_namings(id_=FIELD_NAMING_ID)

        assert result is not None
        assert isinstance(result, queries_field_namings.GetJoinedFieldNamingsRow)
        assert json.loads(result.outputs) == {"first": 1}
        assert result.outputs == result.outputs_2

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::set_field_naming_outputs",
        depends=["PymysqlTestDataclassClasses::get_joined_field_namings"],
    )
    def test_set_field_naming_outputs(self, field_namings_obj: queries_field_namings.QueriesFieldNamings) -> None:
        field_namings_obj.set_field_naming_outputs(outputs=json.dumps({"second": 2}), id_=FIELD_NAMING_ID)
        result = field_namings_obj.get_field_naming(id_=FIELD_NAMING_ID)

        assert result is not None
        assert json.loads(result.outputs) == {"second": 2}

    @pytest.mark.dependency(depends=["PymysqlTestDataclassClasses::set_field_naming_outputs"])
    def test_delete_field_naming(self, pymysql_conn: pymysql.Connection) -> None:
        with pymysql_conn.cursor() as cur:
            cur.execute("DELETE FROM test_field_namings WHERE id = %s", (FIELD_NAMING_ID,))

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::insert_invalid_identifiers")
    def test_insert_invalid_identifiers(self, invalid_identifiers_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers) -> None:
        invalid_identifiers_obj.insert_invalid_identifiers(id_=INVALID_IDENTIFIER_ID, column_3p_="3p-value", new_notes="some new notes")

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_invalid_identifiers",
        depends=["PymysqlTestDataclassClasses::insert_invalid_identifiers"],
    )
    def test_get_invalid_identifiers(self, invalid_identifiers_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers) -> None:
        result = invalid_identifiers_obj.get_invalid_identifiers(id_=INVALID_IDENTIFIER_ID)

        # The insert never sets `%pct`, so it stays NULL.
        assert result == models.TestInvalidIdentifier(
            id_=INVALID_IDENTIFIER_ID,
            column_3p_="3p-value",
            new_notes="some new notes",
            column__pct=None,
        )

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_invalid_identifiers_not_found",
        depends=["PymysqlTestDataclassClasses::get_invalid_identifiers"],
    )
    def test_get_invalid_identifiers_not_found(self, invalid_identifiers_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers) -> None:
        assert invalid_identifiers_obj.get_invalid_identifiers(id_=0) is None

    @pytest.mark.dependency(name="PymysqlTestDataclassClasses::insert_third_party_stat")
    def test_insert_third_party_stat(self, invalid_identifiers_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers) -> None:
        invalid_identifiers_obj.insert_third_party_stat(id_=THIRD_PARTY_ID, total=THIRD_PARTY_TOTAL)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_third_party_stat",
        depends=["PymysqlTestDataclassClasses::insert_third_party_stat"],
    )
    def test_get_third_party_stat(self, invalid_identifiers_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers) -> None:
        result = invalid_identifiers_obj.get_third_party_stat(id_=THIRD_PARTY_ID)

        assert result == models.Model3RdPartyStat(id_=THIRD_PARTY_ID, total=THIRD_PARTY_TOTAL)

    @pytest.mark.dependency(
        name="PymysqlTestDataclassClasses::get_third_party_stat_not_found",
        depends=["PymysqlTestDataclassClasses::get_third_party_stat"],
    )
    def test_get_third_party_stat_not_found(self, invalid_identifiers_obj: queries_invalid_identifiers.QueriesInvalidIdentifiers) -> None:
        assert invalid_identifiers_obj.get_third_party_stat(id_=0) is None
