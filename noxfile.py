from __future__ import annotations

import pathlib
import typing as t

import nox
from nox import options

PATH_TO_PROJECT = pathlib.Path(__name__).parent
SCRIPT_PATHS = ["noxfile.py", PATH_TO_PROJECT / "scripts"]

DRIVER_PATHS = {"asyncpg": PATH_TO_PROJECT / "test" / "driver_asyncpg"}

SQLC_CONFIGS = ["sqlc.yaml"]

options.default_venv_backend = "uv"
options.sessions = ["pyright", "ruff"]


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
    extras: t.Sequence[str] = (),
    groups: t.Sequence[str] = (),
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


@nox.session()
def asyncpg(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pyright"])

    sqlc_generate(session, "asyncpg")
    session.run("pyright", DRIVER_PATHS["asyncpg"])


@nox.session()
def pyright(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["pyright"])

    session.run("pyright", *SCRIPT_PATHS)


@nox.session()
def ruff(session: nox.Session) -> None:
    uv_sync(session, include_self=True, groups=["ruff"])

    session.run("ruff", "format", *SCRIPT_PATHS)
    session.run("ruff", "check")
