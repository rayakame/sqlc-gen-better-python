from __future__ import annotations

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


@pytest_asyncio.fixture(scope="session", loop_scope="session")
async def asyncpg_conn(request: pytest.FixtureRequest) -> asyncpg.Connection[asyncpg.Record]:
    dsn = request.config.getoption("--db")
    if dsn is None or not isinstance(dsn, str):
        msg = "--db option is missing"
        raise ValueError(msg)
    return await asyncpg.connect(dsn)
