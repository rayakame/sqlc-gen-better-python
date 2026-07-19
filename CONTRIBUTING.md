# sqlc-gen-better-python contribution guidelines

Thank you for investing your time trying to improve this plugin. We have some contribution guidelines
that you should follow to ensure that your contribution is at its best.

## Prerequisites

- **Go** - the version from [`go.mod`](go.mod). Any recent Go installation works, the toolchain
  is downloaded automatically if yours is older.
- **Python >= 3.12** and [**uv**](https://docs.astral.sh/uv/) - all Python tooling runs through uv.
- [**sqlc**](https://docs.sqlc.dev/en/latest/overview/install.html) on your PATH.
- **Docker** (or a local PostgreSQL) - only needed for the runtime tests.

One-time setup for the Python tooling:

```bash
uv sync --group dev
```

## How the repo fits together

The plugin itself is Go code under `internal/`, compiled to a WASM binary that `sqlc generate`
executes. The Python code it generates is committed on purpose: `test/driver_<driver>/` contains
a `sqlc.yaml` that generates a matrix of model types into committed subdirectories. CI regenerates
them and fails if the output differs, type-checks them with pyright (strict), lints them with ruff
and runs runtime test suites against real databases. Never edit generated files by hand - change
the generator and regenerate.

## Building the WASM plugin

After any Go change the plugin must be rebuilt, otherwise `sqlc generate` keeps using the old
binary:

```bash
./scripts/build/build.sh      # Linux/macOS
.\scripts\build\build.bat     # Windows
```

The script builds with `GOOS=wasip1 GOARCH=wasm`, computes the SHA-256 of the new binary, patches
the `sha256:` field in the root `sqlc.yaml` and every `test/driver_*/sqlc.yaml` (sqlc refuses to
run a plugin whose hash does not match), and copies the `.wasm` into each test driver directory.
Commit the updated binaries and yaml files together with your Go change.

## Go checks

```bash
make tests      # go test -shuffle=on ./...
make fmt        # golangci-lint fmt
make lint       # golangci-lint run
make pipelines  # lint-fix + fmt + lint (default goal)
```

`make lint` passes with zero issues on a clean checkout; please keep it that way.

## Python pipelines

The pipelines are built with `nox`. `uv run nox` runs the default sessions (regeneration,
pyright, ruff, ruff_format and pytest - everything except the `_check` variants), single
sessions run with `uv run nox -s name1 name2`:

| Session                                             | What it does                                                                           |
|-----------------------------------------------------|----------------------------------------------------------------------------------------|
| `asyncpg`, `sqlite3`, `aiosqlite`                   | Regenerate the driver's test fixtures via sqlc, then pyright + ruff                    |
| `asyncpg_check`, `sqlite3_check`, `aiosqlite_check` | `sqlc diff` variant: verify the committed generated code is up to date (CI uses these) |
| `pyright`, `ruff`, `ruff_format`                    | Type-check / lint / format-check the repository                                        |
| `pytest`                                            | Runtime tests against real databases                                                   |

The `pytest` session needs a local PostgreSQL. The connection URI is read from the
`POSTGRES_URI` environment variable and defaults to
`postgresql://root:187187@localhost:5432/root`; set the variable only if your instance
differs from that. To start a matching instance with docker, run

```bash
docker run --rm --name sqlc-gen-better-python-postgres \
  -e POSTGRES_USER=root \
  -e POSTGRES_PASSWORD=187187 \
  -e POSTGRES_DB=root \
  -p 5432:5432 \
  -d postgres
```

and stop it (after running the tests) with the command below; `--rm` removes the container
on stop, so the `docker run` command above can be reused as is next time.

```bash
docker stop sqlc-gen-better-python-postgres
```

Extra pytest arguments pass through after `--`, e.g.
`uv run nox -s pytest -- test/driver_asyncpg/msgspec/test_msgspec_classes.py -k test_name`.

## The full loop for generator changes

1. Change the Go code and run `make tests` / `make lint`.
2. Rebuild the WASM plugin (see above).
3. `uv run nox` - regenerates the fixtures and runs every check on them.
4. If your change affects generated output, add coverage: a query/schema case in the test matrix
   that pins the new behavior, plus a runtime test where it makes sense. CI gates pull requests
   on patch coverage, so aim for covering every branch of code your PR adds.
5. Commit the regenerated fixtures, wasm binaries and yaml files together with the Go change.

Two gotchas worth knowing: generated output must be byte-identical across runs AND a no-op under
`ruff format`, and `.sql` files must stay ASCII-only - multi-byte characters in comments corrupt
sqlc's byte-offset parameter rewriting and can silently drop `?` placeholders.

## Changelog fragments

We use [changie](https://changie.dev/) to manage changelog creation, and every PR needs a
changelog fragment. changie is set up as a Go tool, so no separate installation is needed:

```bash
make changelog    # or: go tool changie new
```

Changie will ask you for the following fields:

- Kind: The kind of changes, should be self-explanatory
- Body: A short description of the changes.
- PR: The number of the pull request associated to the changes.
- Github Name: The **username** of the github account that made the changes. This is used for
  giving credits to contributors in the changelog. When a fix was diagnosed by someone else
  (e.g. an issue reporter), feel free to credit them here instead.
