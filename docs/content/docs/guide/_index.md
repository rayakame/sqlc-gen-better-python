---
title: Guide
weight: 2
prev: /docs/getting-started
next: /docs/guide/configuration
sidebar:
  open: true
---

This guide walks through `sqlc-gen-better-python` one feature at a time. Each
page states what a feature does, the configuration it needs, and shows real
schema, queries, and the exact Python the plugin generates for them.

## How the pieces fit

You never call this plugin directly. You write a SQL schema and queries, and
`sqlc` runs the plugin for you:

1. You write **schema** (`CREATE TABLE`, `CREATE TYPE`) and **queries**
   (annotated `SELECT`/`INSERT`/...).
2. `sqlc` parses them into a typed **catalog** and hands it to the plugin.
3. The plugin walks the catalog and emits **Python**: a `models.py` (one class
   per table), one query module per query `.sql` file (typed functions), and an
   `enums.py` when your schema has enums.

So everything in this guide is driven by two inputs - your SQL and your
`sqlc.yaml` options - and observed through one output: the generated Python.

{{< callout type="info" >}}
  New here? Start with [Getting Started](/docs/getting-started) for install and
  a first `sqlc generate`, then come back to this guide.
{{< /callout >}}

## How to read this guide

Start with the essentials, then dip into feature pages as you need them:

{{< cards >}}
  {{< card link="configuration" title="Configuration" subtitle="The sqlc.yaml plugin block and the core options." >}}
  {{< card link="drivers" title="Drivers" subtitle="asyncpg, aiosqlite, and sqlite3 - and which query commands each supports." >}}
  {{< card link="model-types" title="Model types" subtitle="dataclass, attrs, msgspec, or pydantic models." >}}
  {{< card link="writing-queries" title="Writing queries" subtitle="How query annotations become typed Python functions." >}}
{{< /cards >}}

The remaining pages - enums, type overrides, converters, working with JSON,
docstrings, SQLite type conversion, and naming - cover specific features and can
be read in any order. The [Reference](/docs/reference) section holds the
canonical option, type-mapping, and driver-support tables.
