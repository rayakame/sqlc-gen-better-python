---
title: sqlc-gen-better-python
layout: hextra-home
---

{{< hextra/hero-badge >}}
  <div class="hx-w-2 hx-h-2 hx-rounded-full hx-bg-primary-400"></div>
  <span>Free, open source, self-hosted</span>
{{< /hextra/hero-badge >}}

<div class="hx-mt-6 hx-mb-6">
{{< hextra/hero-headline >}}
  Type-safe Python from your SQL
{{< /hextra/hero-headline >}}
</div>

<div class="hx-mb-12">
{{< hextra/hero-subtitle >}}
  A sqlc plugin that generates modern, fully typed Python database&nbsp;<br class="sm:hx-block hx-hidden" />code from plain SQL &mdash; models plus async query functions.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx-mb-6">
{{< hextra/hero-button text="Get Started" link="docs/getting-started" >}}
</div>

<div class="hx-mt-6"></div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Four model types"
    subtitle="Generate dataclass, attrs, msgspec, or pydantic models &mdash; pick per codegen block."
  >}}
  {{< hextra/feature-card
    title="Three drivers"
    subtitle="asyncpg for PostgreSQL, plus aiosqlite and sqlite3 for SQLite."
  >}}
  {{< hextra/feature-card
    title="Strictly typed output"
    subtitle="Generated code passes pyright strict and ruff, targeting Python 3.12+."
  >}}
  {{< hextra/feature-card
    title="Enums"
    subtitle="PostgreSQL enums become enum.StrEnum classes, wired through models and queries."
  >}}
  {{< hextra/feature-card
    title="Type overrides & converters"
    subtitle="Swap a column's Python type, or plug in your own encode/decode functions."
  >}}
  {{< hextra/feature-card
    title="Docstrings"
    subtitle="Optional google, numpy, or pep257 docstrings on every generated function."
  >}}
{{< /hextra/feature-grid >}}
