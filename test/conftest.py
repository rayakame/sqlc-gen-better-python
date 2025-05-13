from __future__ import annotations

import asyncio
import typing

import asyncpg

if typing.TYPE_CHECKING:
    import pytest
import pytest_asyncio


def pytest_addoption(parser: pytest.Parser) -> None:
    parser.addoption(
        "--db",
        action="store",
        help="the db uri needed to connect to the db",
        required=True,
    )


def get_dsn(config: pytest.Config) -> str:
    dsn = config.getoption("--db")
    if dsn is None or not isinstance(dsn, str):
        msg = "--db option is missing"
        raise ValueError(msg)
    return dsn


@pytest_asyncio.fixture(scope="session", loop_scope="session")
async def asyncpg_conn(request: pytest.FixtureRequest) -> asyncpg.Connection[asyncpg.Record]:
    dsn = get_dsn(request.config)
    return await asyncpg.connect(dsn)


async def asyncpg_delete_all(dsn: str) -> None:
    conn = await asyncpg.connect(dsn)

    await conn.execute("""
    DELETE FROM test_postgres_types;
    DELETE FROM test_inner_postgres_types;
    """)
    await conn.close()


def pytest_sessionfinish(session: pytest.Session, exitstatus: pytest.ExitCode) -> None:  # noqa: ARG001
    dsn = get_dsn(session.config)

    asyncio.run(asyncpg_delete_all(dsn))
