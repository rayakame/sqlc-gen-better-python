from __future__ import annotations

import hashlib
import os
import pathlib
import typing

import nox
import yaml
from nox import options

if typing.TYPE_CHECKING:
    import collections.abc

PATH_TO_PROJECT = pathlib.Path(__name__).parent.absolute()
SCRIPT_PATHS = ["noxfile.py", PATH_TO_PROJECT / "scripts", PATH_TO_PROJECT / "test"]
TESTS_PATH = PATH_TO_PROJECT / "test"

DRIVER_PATHS = {
    "asyncpg": TESTS_PATH / "driver_asyncpg",
    "aiosqlite": TESTS_PATH / "driver_aiosqlite",
    "sqlite3": TESTS_PATH / "driver_sqlite3",
}

SQLC_CONFIGS = ["sqlc.yaml"]

options.default_venv_backend = "uv"
options.sessions = ["ruff_format", "asyncpg", "sqlite3", "aiosqlite", "pyright", "ruff", "pytest"]

DEFAULT_POSTGRES_URI = os.getenv("POSTGRES_URI", "postgresql://root:187187@localhost:5432/root")


# uv_sync taken from: https://github.com/hikari-py/hikari/blob/master/pipelines/nox.py#L48
#
# Copyright (c) 2020 Nekokatt
# Copyright (c) 2021-present davfsa
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
def uv_sync(
    session: nox.Session,
    /,
    *,
    include_self: bool = False,
    extras: collections.abc.Sequence[str] = (),
    groups: collections.abc.Sequence[str] = (),
) -> None:
    if extras and not include_self:
        msg = "When specifying extras, set `include_self=True`."
        raise RuntimeError(msg)

    args: list[str] = []
    for extra in extras:
        args.extend(("--extra", extra))

    group_flag = "--group" if include_self else "--only-group"
    for group in groups:
        args.extend((group_flag, group))

    session.run_install(
        "uv",
        "sync",
        "--frozen",
        *args,
        silent=True,
        env={"UV_PROJECT_ENVIRONMENT": session.virtualenv.location},
    )


def sqlc_generate(session: nox.Session, driver: str) -> None:
    with session.chdir(DRIVER_PATHS[driver]):
        for config in SQLC_CONFIGS:
            session.run("sqlc", "generate", "-f", config, external=True)


def sqlc_check(session: nox.Session, driver: str) -> None:
    with session.chdir(DRIVER_PATHS[driver]):
        for config in SQLC_CONFIGS:
            session.run("sqlc", "diff", "-f", config, external=True)


@nox.session(reuse_venv=True)
def update_test_plugin(session: nox.Session) -> None:
    # Build the plugin
    wasm_path = TESTS_PATH / "sqlc-gen-better-python.wasm"

    with session.chdir("plugin"):
        session.run(
            "go",
            "build",
            "-o",
            str(wasm_path),
            env={"GOOS": "wasip1", "GOARCH": "wasm"},
            external=True,
        )

    # Calculate the SHA256 hash
    sha256_hasher = hashlib.sha256()
    with wasm_path.open("rb") as fp:
        while True:
            data = fp.read(65536)  # 64kb chunks
            if not data:
                break

            sha256_hasher.update(data)

    plugin_hash = sha256_hasher.hexdigest()

    # Update the SHA256 in the config files
    for driver_name, driver_path in DRIVER_PATHS.items():
        for config_filename in SQLC_CONFIGS:
            config_path = driver_path / config_filename

            with config_path.open() as fp:
                config = yaml.safe_load(fp)

            config["plugins"][0]["wasm"]["sha256"] = plugin_hash

            with config_path.open("w") as fp:
                yaml.safe_dump(config, fp)

        sqlc_generate(session, driver_name)


@nox.session(reuse_venv=True, requires=["update_test_plugin"])
def sqlite3(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pyright", "ruff"])

    sqlc_generate(session, "sqlite3")
    session.run("pyright", DRIVER_PATHS["sqlite3"])
    session.run("ruff", "check", *session.posargs, DRIVER_PATHS["sqlite3"])


@nox.session(reuse_venv=True, requires=["update_test_plugin"])
def sqlite3_check(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pyright", "ruff"])

    sqlc_check(session, "sqlite3")
    session.run("pyright", DRIVER_PATHS["sqlite3"])
    session.run("ruff", "check", *session.posargs, DRIVER_PATHS["sqlite3"])


@nox.session(reuse_venv=True, requires=["update_test_plugin"])
def aiosqlite(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pyright", "ruff"])

    sqlc_generate(session, "aiosqlite")
    session.run("pyright", DRIVER_PATHS["aiosqlite"])
    session.run("ruff", "check", *session.posargs, DRIVER_PATHS["aiosqlite"])


@nox.session(reuse_venv=True)
def aiosqlite_check(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pyright", "ruff"])

    sqlc_check(session, "aiosqlite")
    session.run("pyright", DRIVER_PATHS["aiosqlite"])
    session.run("ruff", "check", *session.posargs, DRIVER_PATHS["aiosqlite"])


@nox.session(reuse_venv=True)
def asyncpg(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pyright", "ruff"])

    sqlc_generate(session, "asyncpg")
    session.run("pyright", DRIVER_PATHS["asyncpg"])
    session.run("ruff", "check", *session.posargs, DRIVER_PATHS["asyncpg"])


@nox.session(reuse_venv=True)
def asyncpg_check(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pyright", "ruff"])

    sqlc_check(session, "asyncpg")
    session.run("pyright", DRIVER_PATHS["asyncpg"])
    session.run("ruff", "check", *session.posargs, DRIVER_PATHS["asyncpg"])


@nox.session(reuse_venv=True)
def pyright(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pyright"])

    session.run("pyright", *SCRIPT_PATHS)


@nox.session(reuse_venv=True)
def ruff_format(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["ruff"])

    session.run("ruff", "format")
    session.run("ruff", "check", "--select", "I", "--fix")


@nox.session(reuse_venv=True)
def ruff(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["ruff"])

    session.run("ruff", "format")
    session.run("ruff", "check", *session.posargs)


@nox.session(reuse_venv=True)
def ruff_check(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["ruff"])

    session.run("ruff", "format", "--check")
    session.run("ruff", "check", *session.posargs)


PYTEST_RUN_FLAGS = [
    "--showlocals",
    "--show-capture",
    "all",
    f"--db={DEFAULT_POSTGRES_URI}",
]
PYTESTCOVERAGE_FLAGS = [
    "--cov",
    "--cov-config",
    "pyproject.toml",
    "--cov-report",
    "term",
    "--cov-report",
    "html:public",
    "--cov-report",
    "xml",
]


@nox.session(reuse_venv=True)
def pytest(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pytest"])

    flags = PYTEST_RUN_FLAGS

    if "--coverage" in session.posargs:
        session.posargs.remove("--coverage")
        flags.extend(PYTESTCOVERAGE_FLAGS)

    session.run("pytest", *flags, *session.posargs)
