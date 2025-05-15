# sqlc-gen-better-python contribution guidelines

Thank you for investing your time trying to improve this plugin. We have some contribution guidelines
that you should follow to ensure that your contribution is at its best.

# CI

We have a basic CI to ensure that the plugin generates working code without any obvious errors.
The CI is build using `nox` which makes running pipelines for python code much easier.

To get our pipelines running you will need to first install `nox`.

```bash
uv sync --group dev
```

<details>
    <summary>Equivalent pip command</summary>
    
```bash
pip install 'nox[uv]'
```
</details>

You will also need to have [sqlc installed](https://docs.sqlc.dev/en/latest/overview/install.html) locally before running some of the pipelines.

The `pytest` pipeline requires you to have a local postgres db running. To change the default connection URI,
nox looks for a `POSTGRES_URI` enviourment variable.

To start a postgres instance with docker, run

```bash
docker run --name sqlc-gen-better-python-postgres \
  -e POSTGRES_USER=root \
  -e POSTGRES_PASSWORD=187187 \
  -e POSTGRES_DB=root \
  -p 5432:5432 \
  -d postgres
```

and stop it (after running the tests) with

```bash
docker stop sqlc-gen-better-python-postgres
```

Before committing we recommend you to run `nox` to run all important pipelines and make sure the pipelines won't fail.

You may run a single pipeline with `nox -s name` or multiple pipelines with `nox -s name1 name3 name9`.

# Changelog fragments

We use [changie](https://changie.dev/) to manage changelog creation.

Every PR needs to have a changelog fragment for that to work.
Please refer to the [changie documentation](https://changie.dev/guide/installation/) for information about installing changie.

After installing changie you can run

```cmd
changie new
```

To create the needed changelog fragment. Changie will ask you for the following fields:

- Kind: The kind of changes, should be self explanatory
- Body: A short description about the made changes.
- PR: The number of the pull request associated to the changes.
- Github Name: The **username** of the github account that made the changes. This is used for giving credits to contributors in the changelog.
