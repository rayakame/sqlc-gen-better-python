---
title: Model types
weight: 30
prev: /docs/guide/drivers
next: /docs/guide/writing-queries
---

The `model_type` option controls what kind of class the plugin generates for each
table. All four produce the same fields with the same annotations - they differ
only in the class machinery and dependencies.

```yaml
options:
  model_type: "msgspec"   # dataclass | attrs | msgspec | pydantic
```

## The four forms

Given this table:

```sql
CREATE TABLE test_field_namings
(
    id      bigint PRIMARY KEY NOT NULL,
    outputs jsonb              NOT NULL
);
```

the plugin generates (shown with the default `docstrings: none`):

{{< tabs >}}

  {{< tab name="dataclass" >}}

```python
@dataclasses.dataclass()
class TestFieldNaming:
    id_: int
    outputs: str
```

Standard library, no dependencies. The safe default.
  {{< /tab >}}

  {{< tab name="attrs" >}}

```python
@attrs.define()
class TestFieldNaming:
    id_: int
    outputs: str
```

Requires `attrs`. Slotted classes with concise definitions.
  {{< /tab >}}

  {{< tab name="msgspec" >}}

```python
class TestFieldNaming(msgspec.Struct):
    id_: int
    outputs: str
```

Requires `msgspec`. Very fast serialization/validation; pairs well with
[working with JSON](/docs/guide/working-with-json).
  {{< /tab >}}

  {{< tab name="pydantic" >}}

```python
class TestFieldNaming(pydantic.BaseModel):
    model_config = pydantic.ConfigDict(arbitrary_types_allowed=True)

    id_: int
    outputs: str
```

Requires `pydantic >= 2.9`. Full runtime validation on construction.
  {{< /tab >}}

{{< /tabs >}}

Which to pick: `dataclass` if you want zero dependencies, `msgspec` for speed and
JSON, `pydantic` for runtime validation, `attrs` for its ergonomics.

## Pydantic specifics

`model_type: pydantic` differs from the others in two ways worth knowing:

- **`arbitrary_types_allowed=True`** is set on every model. It is required for
  field types pydantic has no core schema for (`memoryview` for `bytea`/`blob`,
  and custom [override](/docs/guide/type-overrides) types). Those fields are still
  validated with an `isinstance` check; all other fields get full validation.
- **Runtime annotations.** Unlike the other model types, pydantic resolves field
  annotations at runtime, so generated files import referenced modules at module
  level instead of inside `if typing.TYPE_CHECKING:`.

{{< callout type="info" >}}
  If you lint generated pydantic code with ruff, set
  `lint.flake8-type-checking.runtime-evaluated-base-classes = ["pydantic.BaseModel"]`
  so it does not try to move those runtime imports into a `TYPE_CHECKING` block.
{{< /callout >}}
