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

import asyncio
import pathlib
import sqlite3
import typing

import aiosqlite
import asyncpg
import pytest

if typing.TYPE_CHECKING:
    import collections.abc
import pytest_asyncio

ASYNCPG_PATH = pathlib.Path(__file__).parent / "driver_asyncpg"
AIOSQLITE_PATH = pathlib.Path(__file__).parent / "driver_aiosqlite"
SQLITE3_PATH = pathlib.Path(__file__).parent / "driver_sqlite3"


def pytest_addoption(parser: pytest.Parser) -> None:
    parser.addoption(
        "--db",
        action="store",
        help="the db uri needed to connect to the db",
        required=True,
    )
    parser.addoption(
        "--sqlite-db",
        action="store",
        default="sqlite.db",
        help="the sqlite db uri needed to connect to the db",
    )


def get_dsn(config: pytest.Config) -> str:
    dsn = config.getoption("--db")
    if dsn is None or not isinstance(dsn, str):
        msg = "--db option is missing"
        raise ValueError(msg)
    return dsn


def get_sqlite_dsn(config: pytest.Config) -> str:
    dsn = config.getoption("--sqlite-db")
    if dsn is None or not isinstance(dsn, str):
        msg = "--sqlite-db option is missing"
        raise ValueError(msg)
    return dsn


@pytest_asyncio.fixture(scope="session", loop_scope="session")
async def asyncpg_conn(
    request: pytest.FixtureRequest,
) -> collections.abc.AsyncGenerator[asyncpg.Connection[asyncpg.Record], typing.Any]:
    dsn = get_dsn(request.config)
    conn = await asyncpg.connect(dsn)

    await conn.execute((ASYNCPG_PATH / "schema.sql").read_text())
    yield conn
    await conn.execute("""
        DELETE FROM test_postgres_types;
        DELETE FROM test_inner_postgres_types;
        DELETE FROM test_copy_from;
    """)
    await conn.close()


@pytest.fixture(scope="class")
def sqlite3_conn(
    request: pytest.FixtureRequest,
) -> collections.abc.Generator[sqlite3.Connection, typing.Any]:
    dsn = get_sqlite_dsn(request.config)
    conn = sqlite3.connect(dsn, detect_types=sqlite3.PARSE_DECLTYPES)
    conn.executescript((SQLITE3_PATH / "schema.sql").read_text())
    conn.commit()
    yield conn

    conn.executescript("""DELETE FROM test_sqlite_types;DELETE FROM test_inner_sqlite_types;""")
    conn.commit()
    conn.close()


@pytest_asyncio.fixture(scope="class", loop_scope="session")
async def aiosqlite_conn(
    request: pytest.FixtureRequest,
) -> collections.abc.AsyncGenerator[aiosqlite.Connection, typing.Any]:
    dsn = get_sqlite_dsn(request.config)
    conn = await aiosqlite.connect(dsn, detect_types=sqlite3.PARSE_DECLTYPES)
    await conn.executescript((AIOSQLITE_PATH / "schema.sql").read_text())
    await conn.commit()
    yield conn

    await conn.executescript(
        """DELETE FROM test_sqlite_types;DELETE FROM test_inner_sqlite_types;DELETE FROM test_type_override;"""
    )
    await conn.commit()
    await conn.close()


async def asyncpg_delete_all(dsn: str) -> None:
    conn = await asyncpg.connect(dsn)

    await conn.execute("""
    DELETE FROM test_postgres_types;
    DELETE FROM test_inner_postgres_types;
        DELETE FROM test_copy_from;
    """)
    await conn.close()


async def aiosqlite_delete_all(dsn: str) -> None:
    conn = await aiosqlite.connect(dsn, detect_types=sqlite3.PARSE_DECLTYPES)

    await conn.executescript("""
        DELETE FROM test_sqlite_types;
        DELETE FROM test_inner_sqlite_types;
        DELETE FROM test_type_override;
    """)
    await conn.commit()
    await conn.close()


def pytest_sessionfinish(session: pytest.Session, exitstatus: pytest.ExitCode) -> None:  # noqa: ARG001
    async def _delete_all(conf: pytest.Config) -> None:
        postgres_dsn = get_dsn(conf)
        await asyncpg_delete_all(postgres_dsn)

        aiosqlite_dsn = get_sqlite_dsn(conf)
        await aiosqlite_delete_all(aiosqlite_dsn)

    asyncio.run(_delete_all(session.config))
