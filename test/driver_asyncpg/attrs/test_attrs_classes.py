from __future__ import annotations

import collections.abc
import datetime
import decimal
import uuid

import asyncpg
import pytest
import pytest_asyncio

from test.driver_asyncpg.attrs.classes import models
from test.driver_asyncpg.attrs.classes import queries


@pytest.mark.asyncio(loop_scope="session")
class TestAttrsClasses:
    @pytest.fixture
    def model(self) -> models.TestPostgresType:
        return models.TestPostgresType(
            id=1,
            serial_test=1,
            serial4_test=2,
            bigserial_test=3,
            smallserial_test=4,
            int_test=123,
            bigint_test=123_456_789_012_345,
            smallint_test=12,
            float_test=3.14,
            double_precision_test=2.718281828459045,
            real_test=3.25,
            numeric_test=decimal.Decimal("12345.6789"),
            money_test="$99.99",
            bool_test=True,
            json_test='{"foo": "bar"}',
            jsonb_test='{"foo": "bar", "active": true}',
            bytea_test=memoryview(b"\x00\x01\x02hello"),
            date_test=datetime.date(2025, 1, 1),
            time_test=datetime.time(14, 30, 0),
            timetz_test=datetime.time(14, 30, 0, tzinfo=datetime.timezone.utc),
            timestamp_test=datetime.datetime(2025, 1, 1, 14, 30, 0),
            timestamptz_test=datetime.datetime(2025, 1, 1, 14, 30, 0, tzinfo=datetime.timezone.utc),
            interval_test=datetime.timedelta(days=1, hours=2, minutes=30),
            text_test="Lorem ipsum",
            varchar_test="Example varchar",
            bpchar_test="ABCDEFGHIJ",  # 10‑char padded string
            char_test="X",
            citext_test="CaseInsensitive",
            uuid_test=uuid.UUID("12345678-1234-5678-1234-567812345678"),
            inet_test="192.168.1.1",
            cidr_test="192.168.100.0/24",
            macaddr_test="08:00:2b:01:02:03",
            macaddr8_test="08:00:2b:ff:fe:01:02:03",
            ltree_test="Top.Science.Astronomy",
            lquery_test="*.Astronomy.*",
            ltxtquery_test="Astro* & Stars",
        )

    @pytest.fixture
    def inner_model(self) -> models.TestInnerPostgresType:
        return models.TestInnerPostgresType(
            table_id=1,
            serial_test=1,
            serial4_test=2,
            bigserial_test=3,
            smallserial_test=4,
            int_test=123,
            bigint_test=123_456_789_012_345,
            smallint_test=12,
            float_test=3.14,
            double_precision_test=2.718281828459045,
            real_test=3.25,
            numeric_test=decimal.Decimal("12345.6789"),
            money_test="$99.99",
            bool_test=True,
            json_test='{"foo": "bar"}',
            jsonb_test='{"foo": "bar", "active": true}',
            bytea_test=memoryview(b"\x00\x01\x02hello"),
            date_test=datetime.date(2025, 1, 1),
            time_test=datetime.time(14, 30, 0),
            timetz_test=datetime.time(14, 30, 0, tzinfo=datetime.timezone.utc),
            timestamp_test=datetime.datetime(2025, 1, 1, 14, 30, 0),
            timestamptz_test=datetime.datetime(2025, 1, 1, 14, 30, 0, tzinfo=datetime.timezone.utc),
            interval_test=datetime.timedelta(days=1, hours=2, minutes=30),
            text_test="Lorem ipsum",
            varchar_test="Example varchar",
            bpchar_test="ABCDEFGHIJ",  # 10‑char padded string
            char_test="X",
            citext_test="CaseInsensitive",
            uuid_test=uuid.UUID("12345678-1234-5678-1234-567812345678"),
            inet_test="192.168.1.1",
            cidr_test="192.168.100.0/24",
            macaddr_test="08:00:2b:01:02:03",
            macaddr8_test="08:00:2b:ff:fe:01:02:03",
            ltree_test="Top.Science.Astronomy",
            lquery_test="*.Astronomy.*",
            ltxtquery_test="Astro* & Stars",
        )

    @pytest_asyncio.fixture(scope="session", loop_scope="session")
    async def queries_obj(self, asyncpg_conn: asyncpg.Connection[asyncpg.Record]):
        return queries.Queries(conn=asyncpg_conn)

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestAttrsClasses::create")
    async def test_create(self, queries_obj: queries.Queries, model: models.TestPostgresType) -> None:
        await queries_obj.create_one_test_postgres_type(
            id_=model.id,
            serial_test=model.serial_test,
            serial4_test=model.serial4_test,
            bigserial_test=model.bigserial_test,
            smallserial_test=model.smallserial_test,
            int_test=model.int_test,
            bigint_test=model.bigint_test,
            smallint_test=model.smallint_test,
            float_test=model.float_test,
            double_precision_test=model.double_precision_test,
            real_test=model.real_test,
            numeric_test=model.numeric_test,
            money_test=model.money_test,
            bool_test=model.bool_test,
            json_test=model.json_test,
            jsonb_test=model.jsonb_test,
            bytea_test=model.bytea_test,
            date_test=model.date_test,
            time_test=model.time_test,
            timetz_test=model.timetz_test,
            timestamp_test=model.timestamp_test,
            timestamptz_test=model.timestamptz_test,
            interval_test=model.interval_test,
            text_test=model.text_test,
            varchar_test=model.varchar_test,
            bpchar_test=model.bpchar_test,
            char_test=model.char_test,
            citext_test=model.citext_test,
            uuid_test=model.uuid_test,
            inet_test=model.inet_test,
            cidr_test=model.cidr_test,
            macaddr_test=model.macaddr_test,
            macaddr8_test=model.macaddr8_test,
            ltree_test=model.ltree_test,
            lquery_test=model.lquery_test,
            ltxtquery_test=model.ltxtquery_test,
        )

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(name="TestAttrsClasses::create_inner", depends=["TestAttrsClasses::create"])
    async def test_create_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerPostgresType) -> None:
        await queries_obj.create_one_test_postgres_inner_type(
            table_id=inner_model.table_id,
            serial_test=inner_model.serial_test,
            serial4_test=inner_model.serial4_test,
            bigserial_test=inner_model.bigserial_test,
            smallserial_test=inner_model.smallserial_test,
            int_test=inner_model.int_test,
            bigint_test=inner_model.bigint_test,
            smallint_test=inner_model.smallint_test,
            float_test=inner_model.float_test,
            double_precision_test=inner_model.double_precision_test,
            real_test=inner_model.real_test,
            numeric_test=inner_model.numeric_test,
            money_test=inner_model.money_test,
            bool_test=inner_model.bool_test,
            json_test=inner_model.json_test,
            jsonb_test=inner_model.jsonb_test,
            bytea_test=inner_model.bytea_test,
            date_test=inner_model.date_test,
            time_test=inner_model.time_test,
            timetz_test=inner_model.timetz_test,
            timestamp_test=inner_model.timestamp_test,
            timestamptz_test=inner_model.timestamptz_test,
            interval_test=inner_model.interval_test,
            text_test=inner_model.text_test,
            varchar_test=inner_model.varchar_test,
            bpchar_test=inner_model.bpchar_test,
            char_test=inner_model.char_test,
            citext_test=inner_model.citext_test,
            uuid_test=inner_model.uuid_test,
            inet_test=inner_model.inet_test,
            cidr_test=inner_model.cidr_test,
            macaddr_test=inner_model.macaddr_test,
            macaddr8_test=inner_model.macaddr8_test,
            ltree_test=inner_model.ltree_test,
            lquery_test=inner_model.lquery_test,
            ltxtquery_test=inner_model.ltxtquery_test,
        )

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::create_inner"], name="TestAttrsClasses::get_one")
    async def test_get_one(self, queries_obj: queries.Queries, model: models.TestPostgresType) -> None:
        result = await queries_obj.get_one_test_postgres_type()

        assert result is not None
        assert isinstance(result, models.TestPostgresType)

        assert result.id == model.id
        assert result.serial_test == model.serial_test
        assert result.serial4_test == model.serial4_test
        assert result.bigserial_test == model.bigserial_test
        assert result.smallserial_test == model.smallserial_test
        assert result.int_test == model.int_test
        assert result.bigint_test == model.bigint_test
        assert result.smallint_test == model.smallint_test
        assert result.float_test == model.float_test
        assert result.double_precision_test == model.double_precision_test
        assert result.real_test == model.real_test
        assert result.numeric_test == model.numeric_test
        assert result.money_test == model.money_test
        assert result.bool_test == model.bool_test
        assert result.json_test == model.json_test
        assert result.jsonb_test == model.jsonb_test
        assert result.bytea_test == model.bytea_test
        assert result.date_test == model.date_test
        assert result.time_test == model.time_test
        assert result.timetz_test == model.timetz_test
        assert result.timestamp_test == model.timestamp_test
        assert result.timestamptz_test == model.timestamptz_test
        assert result.interval_test == model.interval_test
        assert result.text_test == model.text_test
        assert result.varchar_test == model.varchar_test
        assert result.bpchar_test == model.bpchar_test
        assert result.char_test == model.char_test
        assert result.citext_test == model.citext_test
        assert result.uuid_test == model.uuid_test
        assert result.inet_test == model.inet_test
        assert result.cidr_test == model.cidr_test
        assert result.macaddr_test == model.macaddr_test
        assert result.macaddr8_test == model.macaddr8_test
        assert result.ltree_test == model.ltree_test
        assert result.lquery_test == model.lquery_test
        assert result.ltxtquery_test == model.ltxtquery_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::get_one"], name="TestAttrsClasses::get_one_inner")
    async def test_get_one_inner(self, queries_obj: queries.Queries, inner_model: models.TestInnerPostgresType) -> None:
        result = await queries_obj.get_one_inner_test_postgres_type()

        assert result is not None
        assert isinstance(result, models.TestInnerPostgresType)

        assert result.table_id == inner_model.table_id
        assert result.serial_test == inner_model.serial_test
        assert result.serial4_test == inner_model.serial4_test
        assert result.bigserial_test == inner_model.bigserial_test
        assert result.smallserial_test == inner_model.smallserial_test
        assert result.int_test == inner_model.int_test
        assert result.bigint_test == inner_model.bigint_test
        assert result.smallint_test == inner_model.smallint_test
        assert result.float_test == inner_model.float_test
        assert result.double_precision_test == inner_model.double_precision_test
        assert result.real_test == inner_model.real_test
        assert result.numeric_test == inner_model.numeric_test
        assert result.money_test == inner_model.money_test
        assert result.bool_test == inner_model.bool_test
        assert result.json_test == inner_model.json_test
        assert result.jsonb_test == inner_model.jsonb_test
        assert result.bytea_test == inner_model.bytea_test
        assert result.date_test == inner_model.date_test
        assert result.time_test == inner_model.time_test
        assert result.timetz_test == inner_model.timetz_test
        assert result.timestamp_test == inner_model.timestamp_test
        assert result.timestamptz_test == inner_model.timestamptz_test
        assert result.interval_test == inner_model.interval_test
        assert result.text_test == inner_model.text_test
        assert result.varchar_test == inner_model.varchar_test
        assert result.bpchar_test == inner_model.bpchar_test
        assert result.char_test == inner_model.char_test
        assert result.citext_test == inner_model.citext_test
        assert result.uuid_test == inner_model.uuid_test
        assert result.inet_test == inner_model.inet_test
        assert result.cidr_test == inner_model.cidr_test
        assert result.macaddr_test == inner_model.macaddr_test
        assert result.macaddr8_test == inner_model.macaddr8_test
        assert result.ltree_test == inner_model.ltree_test
        assert result.lquery_test == inner_model.lquery_test
        assert result.ltxtquery_test == inner_model.ltxtquery_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::get_one_inner"], name="TestAttrsClasses::get_one_timestamp")
    async def test_get_one_timestamp(self, queries_obj: queries.Queries, model: models.TestPostgresType) -> None:
        result = await queries_obj.get_one_test_timestamp_postgres_type()

        assert result is not None
        assert isinstance(result, datetime.datetime)
        assert result == model.timestamp_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::get_one_timestamp"], name="TestAttrsClasses::get_one_bytea")
    async def test_get_one_bytea(self, queries_obj: queries.Queries, model: models.TestPostgresType) -> None:
        result = await queries_obj.get_one_test_bytea_postgres_type()

        assert result is not None
        assert isinstance(result, memoryview)
        assert result == model.bytea_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::get_one_bytea"], name="TestAttrsClasses::get_many")
    async def test_get_many(self, queries_obj: queries.Queries, model: models.TestPostgresType) -> None:
        result = await queries_obj.get_many_test_postgres_type()

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], models.TestPostgresType)

        first_result = result[0]
        assert first_result.id == model.id
        assert first_result.serial_test == model.serial_test
        assert first_result.serial4_test == model.serial4_test
        assert first_result.bigserial_test == model.bigserial_test
        assert first_result.smallserial_test == model.smallserial_test
        assert first_result.int_test == model.int_test
        assert first_result.bigint_test == model.bigint_test
        assert first_result.smallint_test == model.smallint_test
        assert first_result.float_test == model.float_test
        assert first_result.double_precision_test == model.double_precision_test
        assert first_result.real_test == model.real_test
        assert first_result.numeric_test == model.numeric_test
        assert first_result.money_test == model.money_test
        assert first_result.bool_test == model.bool_test
        assert first_result.json_test == model.json_test
        assert first_result.jsonb_test == model.jsonb_test
        assert first_result.bytea_test == model.bytea_test
        assert first_result.date_test == model.date_test
        assert first_result.time_test == model.time_test
        assert first_result.timetz_test == model.timetz_test
        assert first_result.timestamp_test == model.timestamp_test
        assert first_result.timestamptz_test == model.timestamptz_test
        assert first_result.interval_test == model.interval_test
        assert first_result.text_test == model.text_test
        assert first_result.varchar_test == model.varchar_test
        assert first_result.bpchar_test == model.bpchar_test
        assert first_result.char_test == model.char_test
        assert first_result.citext_test == model.citext_test
        assert first_result.uuid_test == model.uuid_test
        assert first_result.inet_test == model.inet_test
        assert first_result.cidr_test == model.cidr_test
        assert first_result.macaddr_test == model.macaddr_test
        assert first_result.macaddr8_test == model.macaddr8_test
        assert first_result.ltree_test == model.ltree_test
        assert first_result.lquery_test == model.lquery_test
        assert first_result.ltxtquery_test == model.ltxtquery_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::get_many"], name="TestAttrsClasses::get_many_timestamp")
    async def test_get_many_timestamp(self, queries_obj: queries.Queries, model: models.TestPostgresType) -> None:
        result = await queries_obj.get_many_test_timestamp_postgres_type()

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], datetime.datetime)

        assert result[0] == model.timestamp_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::get_many_timestamp"], name="TestAttrsClasses::get_many_bytea")
    async def test_get_many_bytea(self, queries_obj: queries.Queries, model: models.TestPostgresType) -> None:
        result = await queries_obj.get_many_test_bytea_postgres_type()

        assert result is not None
        assert isinstance(result, collections.abc.Sequence)
        assert isinstance(result[0], memoryview)

        assert result[0] == model.bytea_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::get_many_bytea"], name="TestAttrsClasses::get_embedded")
    async def test_get_embedded(
        self, queries_obj: queries.Queries, model: models.TestPostgresType, inner_model: models.TestInnerPostgresType
    ) -> None:
        result = await queries_obj.get_embedded_test_postgres_type()

        assert result is not None
        assert isinstance(result, queries.GetEmbeddedTestPostgresTypeRow)
        assert isinstance(result.test_inner_postgres_type, models.TestInnerPostgresType)

        assert result.id == model.id
        assert result.serial_test == model.serial_test
        assert result.serial4_test == model.serial4_test
        assert result.bigserial_test == model.bigserial_test
        assert result.smallserial_test == model.smallserial_test
        assert result.int_test == model.int_test
        assert result.bigint_test == model.bigint_test
        assert result.smallint_test == model.smallint_test
        assert result.float_test == model.float_test
        assert result.double_precision_test == model.double_precision_test
        assert result.real_test, 4 == model.real_test
        assert result.numeric_test == model.numeric_test
        assert result.money_test == model.money_test
        assert result.bool_test == model.bool_test
        assert result.json_test == model.json_test
        assert result.jsonb_test == model.jsonb_test
        assert result.bytea_test == model.bytea_test
        assert result.date_test == model.date_test
        assert result.time_test == model.time_test
        assert result.timetz_test == model.timetz_test
        assert result.timestamp_test == model.timestamp_test
        assert result.timestamptz_test == model.timestamptz_test
        assert result.interval_test == model.interval_test
        assert result.text_test == model.text_test
        assert result.varchar_test == model.varchar_test
        assert result.bpchar_test == model.bpchar_test
        assert result.char_test == model.char_test
        assert result.citext_test == model.citext_test
        assert result.uuid_test == model.uuid_test
        assert result.inet_test == model.inet_test
        assert result.cidr_test == model.cidr_test
        assert result.macaddr_test == model.macaddr_test
        assert result.macaddr8_test == model.macaddr8_test
        assert result.ltree_test == model.ltree_test
        assert result.lquery_test == model.lquery_test
        assert result.ltxtquery_test == model.ltxtquery_test

        assert result.test_inner_postgres_type.table_id == inner_model.table_id
        assert result.test_inner_postgres_type.serial_test == inner_model.serial_test
        assert result.test_inner_postgres_type.serial4_test == inner_model.serial4_test
        assert result.test_inner_postgres_type.bigserial_test == inner_model.bigserial_test
        assert result.test_inner_postgres_type.smallserial_test == inner_model.smallserial_test
        assert result.test_inner_postgres_type.int_test == inner_model.int_test
        assert result.test_inner_postgres_type.bigint_test == inner_model.bigint_test
        assert result.test_inner_postgres_type.smallint_test == inner_model.smallint_test
        assert result.test_inner_postgres_type.float_test == inner_model.float_test
        assert result.test_inner_postgres_type.double_precision_test == inner_model.double_precision_test
        assert result.test_inner_postgres_type.real_test == inner_model.real_test
        assert result.test_inner_postgres_type.numeric_test == inner_model.numeric_test
        assert result.test_inner_postgres_type.money_test == inner_model.money_test
        assert result.test_inner_postgres_type.bool_test == inner_model.bool_test
        assert result.test_inner_postgres_type.json_test == inner_model.json_test
        assert result.test_inner_postgres_type.jsonb_test == inner_model.jsonb_test
        assert result.test_inner_postgres_type.bytea_test == inner_model.bytea_test
        assert result.test_inner_postgres_type.date_test == inner_model.date_test
        assert result.test_inner_postgres_type.time_test == inner_model.time_test
        assert result.test_inner_postgres_type.timetz_test == inner_model.timetz_test
        assert result.test_inner_postgres_type.timestamp_test == inner_model.timestamp_test
        assert result.test_inner_postgres_type.timestamptz_test == inner_model.timestamptz_test
        assert result.test_inner_postgres_type.interval_test == inner_model.interval_test
        assert result.test_inner_postgres_type.text_test == inner_model.text_test
        assert result.test_inner_postgres_type.varchar_test == inner_model.varchar_test
        assert result.test_inner_postgres_type.bpchar_test == inner_model.bpchar_test
        assert result.test_inner_postgres_type.char_test == inner_model.char_test
        assert result.test_inner_postgres_type.citext_test == inner_model.citext_test
        assert result.test_inner_postgres_type.uuid_test == inner_model.uuid_test
        assert result.test_inner_postgres_type.inet_test == inner_model.inet_test
        assert result.test_inner_postgres_type.cidr_test == inner_model.cidr_test
        assert result.test_inner_postgres_type.macaddr_test == inner_model.macaddr_test
        assert result.test_inner_postgres_type.macaddr8_test == inner_model.macaddr8_test
        assert result.test_inner_postgres_type.ltree_test == inner_model.ltree_test
        assert result.test_inner_postgres_type.lquery_test == inner_model.lquery_test
        assert result.test_inner_postgres_type.ltxtquery_test == inner_model.ltxtquery_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::get_embedded"], name="TestAttrsClasses::get_all_embedded")
    async def test_get_all_embedded(
        self, queries_obj: queries.Queries, model: models.TestPostgresType, inner_model: models.TestInnerPostgresType
    ) -> None:
        result = await queries_obj.get_all_embedded_test_postgres_type()

        assert result is not None
        assert isinstance(result, queries.GetAllEmbeddedTestPostgresTypeRow)
        assert isinstance(result.test_postgres_type, models.TestPostgresType)
        assert isinstance(result.test_inner_postgres_type, models.TestInnerPostgresType)

        assert result.test_postgres_type.id == model.id
        assert result.test_postgres_type.serial_test == model.serial_test
        assert result.test_postgres_type.serial4_test == model.serial4_test
        assert result.test_postgres_type.bigserial_test == model.bigserial_test
        assert result.test_postgres_type.smallserial_test == model.smallserial_test
        assert result.test_postgres_type.int_test == model.int_test
        assert result.test_postgres_type.bigint_test == model.bigint_test
        assert result.test_postgres_type.smallint_test == model.smallint_test
        assert result.test_postgres_type.float_test == model.float_test
        assert result.test_postgres_type.double_precision_test == model.double_precision_test
        assert result.test_postgres_type.real_test, 4 == model.real_test
        assert result.test_postgres_type.numeric_test == model.numeric_test
        assert result.test_postgres_type.money_test == model.money_test
        assert result.test_postgres_type.bool_test == model.bool_test
        assert result.test_postgres_type.json_test == model.json_test
        assert result.test_postgres_type.jsonb_test == model.jsonb_test
        assert result.test_postgres_type.bytea_test == model.bytea_test
        assert result.test_postgres_type.date_test == model.date_test
        assert result.test_postgres_type.time_test == model.time_test
        assert result.test_postgres_type.timetz_test == model.timetz_test
        assert result.test_postgres_type.timestamp_test == model.timestamp_test
        assert result.test_postgres_type.timestamptz_test == model.timestamptz_test
        assert result.test_postgres_type.interval_test == model.interval_test
        assert result.test_postgres_type.text_test == model.text_test
        assert result.test_postgres_type.varchar_test == model.varchar_test
        assert result.test_postgres_type.bpchar_test == model.bpchar_test
        assert result.test_postgres_type.char_test == model.char_test
        assert result.test_postgres_type.citext_test == model.citext_test
        assert result.test_postgres_type.uuid_test == model.uuid_test
        assert result.test_postgres_type.inet_test == model.inet_test
        assert result.test_postgres_type.cidr_test == model.cidr_test
        assert result.test_postgres_type.macaddr_test == model.macaddr_test
        assert result.test_postgres_type.macaddr8_test == model.macaddr8_test
        assert result.test_postgres_type.ltree_test == model.ltree_test
        assert result.test_postgres_type.lquery_test == model.lquery_test
        assert result.test_postgres_type.ltxtquery_test == model.ltxtquery_test

        assert result.test_inner_postgres_type.table_id == inner_model.table_id
        assert result.test_inner_postgres_type.serial_test == inner_model.serial_test
        assert result.test_inner_postgres_type.serial4_test == inner_model.serial4_test
        assert result.test_inner_postgres_type.bigserial_test == inner_model.bigserial_test
        assert result.test_inner_postgres_type.smallserial_test == inner_model.smallserial_test
        assert result.test_inner_postgres_type.int_test == inner_model.int_test
        assert result.test_inner_postgres_type.bigint_test == inner_model.bigint_test
        assert result.test_inner_postgres_type.smallint_test == inner_model.smallint_test
        assert result.test_inner_postgres_type.float_test == inner_model.float_test
        assert result.test_inner_postgres_type.double_precision_test == inner_model.double_precision_test
        assert result.test_inner_postgres_type.real_test == inner_model.real_test
        assert result.test_inner_postgres_type.numeric_test == inner_model.numeric_test
        assert result.test_inner_postgres_type.money_test == inner_model.money_test
        assert result.test_inner_postgres_type.bool_test == inner_model.bool_test
        assert result.test_inner_postgres_type.json_test == inner_model.json_test
        assert result.test_inner_postgres_type.jsonb_test == inner_model.jsonb_test
        assert result.test_inner_postgres_type.bytea_test == inner_model.bytea_test
        assert result.test_inner_postgres_type.date_test == inner_model.date_test
        assert result.test_inner_postgres_type.time_test == inner_model.time_test
        assert result.test_inner_postgres_type.timetz_test == inner_model.timetz_test
        assert result.test_inner_postgres_type.timestamp_test == inner_model.timestamp_test
        assert result.test_inner_postgres_type.timestamptz_test == inner_model.timestamptz_test
        assert result.test_inner_postgres_type.interval_test == inner_model.interval_test
        assert result.test_inner_postgres_type.text_test == inner_model.text_test
        assert result.test_inner_postgres_type.varchar_test == inner_model.varchar_test
        assert result.test_inner_postgres_type.bpchar_test == inner_model.bpchar_test
        assert result.test_inner_postgres_type.char_test == inner_model.char_test
        assert result.test_inner_postgres_type.citext_test == inner_model.citext_test
        assert result.test_inner_postgres_type.uuid_test == inner_model.uuid_test
        assert result.test_inner_postgres_type.inet_test == inner_model.inet_test
        assert result.test_inner_postgres_type.cidr_test == inner_model.cidr_test
        assert result.test_inner_postgres_type.macaddr_test == inner_model.macaddr_test
        assert result.test_inner_postgres_type.macaddr8_test == inner_model.macaddr8_test
        assert result.test_inner_postgres_type.ltree_test == inner_model.ltree_test
        assert result.test_inner_postgres_type.lquery_test == inner_model.lquery_test
        assert result.test_inner_postgres_type.ltxtquery_test == inner_model.ltxtquery_test

    @pytest.mark.asyncio(loop_scope="session")
    @pytest.mark.dependency(depends=["TestAttrsClasses::get_embedded"])
    async def test_delete(self, queries_obj: queries.Queries, model: models.TestPostgresType):
        await queries_obj.delete_one_test_postgres_type(id_=model.id)
