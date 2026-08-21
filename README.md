# protoschema-plugins

[![Build](https://github.com/BandlSkyler/protoschema-plugins/actions/workflows/ci.yaml/badge.svg?branch=main)][badges_ci]
[![Report Card](https://goreportcard.com/badge/github.com/BandlSkyler/protoschema-plugins)][badges_goreportcard]
[![GoDoc](https://pkg.go.dev/badge/github.com/BandlSkyler/protoschema-plugins.svg)][badges_godoc]
[![Slack](https://img.shields.io/badge/slack-buf-%23e01563)][badges_slack]

The protoschema-plugins repository contains a collection of Protobuf plugins that generate different
types of schema from protobuf files. This includes:

- [PubSub](#pubsub-protobuf-schema)
- [JSON Schema](#json-schema)

## PubSub Protobuf Schema

Generates a schema for a given protobuf file that can be used as a PubSub schema in the form of a
single self-contained messaged normalized to proto2.

Install the `protoc-gen-pubsub` plugin directly:

```sh
go install github.com/BandlSkyler/protoschema-plugins/cmd/protoc-gen-pubsub@latest
```

Or reference it as a [Remote Plugin](https://buf.build/docs/generate/remote-plugins) in `buf.gen.yaml`:

```yaml
version: v1
plugins:
  - plugin: buf.build/bufbuild/protoschema-pubsub
    out: ./gen
```

For examples see [testdata](/internal/testdata/pubsub/) which contains the generated schema for
test case definitions found in [proto](/proto/).

## JSON Schema

Generates a [JSON Schema](https://json-schema.org/) for a given protobuf file. This implementation
uses [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12/release-notes) by default,
and can generate [Draft-07](https://json-schema.org/draft-07/json-schema-release-notes.html) schemas
via the `schema_version` option.

Install the `protoc-gen-jsonschema` directly:

```sh
go install github.com/BandlSkyler/protoschema-plugins/cmd/protoc-gen-jsonschema@latest
```

Or reference it as a [Remote Plugin](https://buf.build/docs/generate/remote-plugins) in `buf.gen.yaml`:

```yaml
version: v1
plugins:
  - plugin: buf.build/bufbuild/protoschema-jsonschema
    out: ./gen
```

For examples see [testdata](/internal/testdata/jsonschema/) which contains the generated schema for
test case definitions found in [proto](/proto/).

Here is a simple generated schema from the following protobuf:

```proto
// A product.
//
// A product is a good or service that is offered for sale.
message Product {
  // A point on the earth's surface.
  message Location {
    double lat = 1 [
      (buf.validate.field).double.finite = true,
      (buf.validate.field).double.gte = -90,
      (buf.validate.field).double.lte = 90
    ];
    double long = 2 [
      (buf.validate.field).double.finite = true,
      (buf.validate.field).double.gte = -180,
      (buf.validate.field).double.lte = 180
    ];
  }

  // The unique identifier for the product.
  int32 product_id = 1 [(buf.validate.field).required = true];
  // The name of the product.
  string product_name = 2 [(buf.validate.field).required = true];
  // The price of the product.
  float price = 3 [
    (buf.validate.field).float.finite = true,
    (buf.validate.field).float.gte = 0
  ];
  // The tags associated with the product.
  repeated string tags = 4;
  // The location of the product.
  Location location = 5 [(buf.validate.field).required = true];
}

```

By default, results in the following JSON Schema files:

- `*.schema.json` files are generated with protobuf field names (e.g. `product_id`, `product_name`)
- `*.schema.bundle.json` files include all dependencies in a single file with protobuf field names.
- `*.schema.strict.json` files are generated with protobuf field names, but do not allow aliases, string numbers, or any other non-normalized representation.
- `*.schema.strict.bundle.json` files include the strict schema with all dependencies in a single file with protobuf field names.
- `*.jsonschema.json` files are generated with JSON field names (e.g. `productId`, `productName`)
  other non-normalized representation.
- `*.jsonschema.bundle.json` files include all dependencies in a single file with the JSON field names.
- `*.jsonschema.strict.json` files are generated with JSON field names, but do not allow aliases, string numbers, or any other non-normalized representation.
- `*.jsonschema.strict.bundle.json` files include the strict JSON schema with all dependencies in a single file with JSON field names.

For example, the above protobuf generates the following `*.schema.json` files:

<details>
<summary>Product.schema.json</summary>

```json
{
  "$id": "Product.schema.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "additionalProperties": false,
  "title": "A product.",
  "description": "A product is a good or service that is offered for sale.",
  "type": "object",
  "properties": {
    "product_id": {
      "description": "The unique identifier for the product.",
      "maximum": 2147483647,
      "minimum": -2147483648,
      "type": "integer"
    },
    "product_name": {
      "description": "The name of the product.",
      "type": "string"
    },
    "price": {
      "anyOf": [
        {
          "maximum": 3.4028234663852886e38,
          "minimum": 0,
          "type": "number"
        },
        {
          "pattern": "^-?[0-9]+(\\.[0-9]+)?([eE][+-]?[0-9]+)?$",
          "type": "string"
        }
      ],
      "default": 0,
      "description": "The price of the product."
    },
    "tags": {
      "description": "The tags associated with the product.",
      "items": {
        "type": "string"
      },
      "type": "array"
    },
    "location": {
      "$ref": "Product.Location.schema.json",
      "description": "The location of the product."
    }
  },
  "required": ["product_id", "product_name", "location"],
  "patternProperties": {
    "^(productId)$": {
      "description": "The unique identifier for the product.",
      "maximum": 2147483647,
      "minimum": -2147483648,
      "type": "integer"
    },
    "^(productName)$": {
      "description": "The name of the product.",
      "type": "string"
    }
  }
}
```

</details>

<details>
<summary>Product.Location.schema.json</summary>

```json
{
  "$id": "Location.schema.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "additionalProperties": false,
  "title": "Location",
  "description": "A point on the earth's surface.",
  "type": "object",
  "properties": {
    "lat": {
      "anyOf": [
        {
          "maximum": 90,
          "minimum": -90,
          "type": "number"
        },
        {
          "pattern": "^-?[0-9]+(\\.[0-9]+)?([eE][+-]?[0-9]+)?$",
          "type": "string"
        }
      ],
      "default": 0
    },
    "long": {
      "anyOf": [
        {
          "maximum": 180,
          "minimum": -180,
          "type": "number"
        },
        {
          "pattern": "^-?[0-9]+(\\.[0-9]+)?([eE][+-]?[0-9]+)?$",
          "type": "string"
        }
      ],
      "default": 0
    }
  }
}
```

</details>

Or the following `*.jsonschema.strict.bundle.json` file:

<details>
<summary>Product.jsonschema.strict.bundle.json</summary>

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "buf.protoschema.test.v1.Product.jsonschema.strict.bundle.json",
  "$ref": "#/$defs/buf.protoschema.test.v1.Product.jsonschema.strict.json",
  "$defs": {
    "buf.protoschema.test.v1.Product.jsonschema.strict.json": {
      "$schema": "https://json-schema.org/draft/2020-12/schema",
      "title": "A product.",
      "description": "A product is a good or service that is offered for sale.",
      "type": "object",
      "properties": {
        "productId": {
          "description": "The unique identifier for the product.",
          "maximum": 2147483647,
          "minimum": -2147483648,
          "type": "integer"
        },
        "productName": {
          "description": "The name of the product.",
          "type": "string"
        },
        "price": {
          "description": "The price of the product.",
          "maximum": 3.4028234663852886e38,
          "minimum": 0,
          "type": "number"
        },
        "tags": {
          "description": "The tags associated with the product.",
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "location": {
          "$ref": "#/$defs/buf.protoschema.test.v1.Product.Location.jsonschema.strict.json",
          "description": "The location of the product."
        }
      },
      "required": ["productId", "productName", "price", "location"],
      "additionalProperties": false
    },
    "buf.protoschema.test.v1.Product.Location.jsonschema.strict.json": {
      "$schema": "https://json-schema.org/draft/2020-12/schema",
      "additionalProperties": true,
      "description": "A point on the earth's surface.",
      "properties": {
        "lat": {
          "maximum": 90,
          "minimum": -90,
          "type": "number"
        },
        "long": {
          "maximum": 180,
          "minimum": -180,
          "type": "number"
        }
      },
      "required": ["lat", "long"],
      "title": "Location",
      "type": "object"
    }
  }
}
```

</details>

### Options

The JSON Schema plugin supports the following options:

- `target` - Any of `proto`, `json`, `proto-bundle`, `json-bundle`, `proto-strict`, `json-strict`,
  `proto-strict-bundle`, `json-strict-bundle`, or `all` separated by `+` (e.g. `proto+json`). Defaults to `all`.
  - If `proto`, the schema will be generated with Protobuf field names (e.g. `product_id`,
    `product_name`).
  - If `json`, the schema will be generated with JSON field names (e.g. `productId`, `productName`).
  - If suffixed with `-bundle`, the schema will include all dependencies in a single file.
  - If suffixed with `-strict`, the schema will not allow aliases, string numbers, or any other
    non-normalized representation. Strict is useful when the validated JSON data is used directly
    instead of being converted to a Protobuf message. Requires the "always emit fields without
    presence" option when using [Protobuf JSON](https://protobuf.dev/programming-guides/json/#json-options).
  - If suffixed with `-strict-bundle`, the schema will be strict and include all dependencies in a single file.
- `additional_properties` - If `true`, the generated schema will set `additionalProperties` to
  `true`, causing unknown fields to be ignored instead of erroring. Defaults to `false`. Useful when a
  client/sender may have a different version the schema than the server/receiver. Similar to the
  "ignore unknown fields" option in [Protobuf JSON](https://protobuf.dev/programming-guides/json/#json-options).
- `non_required_default` - If `true`, fields are non-required by default: only fields explicitly
  marked `(buf.validate.field).required = true` are listed in the `required` array. This also
  disables the strict-mode behavior of requiring non-optional (implicit default) fields. Defaults to
  `false`.
- `schema_version` - The JSON Schema draft version to generate. `2020-12` (default) or `draft-07`.
  Draft-07 is useful for consumers that only support the older draft.
- `enum_oneof` - If `true`, enum fields are rendered as a string type with a `oneOf` list of
  `{const, title, description}` branches, one per allowed enum value, instead of the default
  `anyOf` of string names and integer ranges. The `const` is the proto enum value name (the
  canonical protobuf JSON representation). Each value's leading comments are split into two
  sections separated by a blank line: the first section becomes the branch `title` (an alias for
  the value), and the second becomes the branch `description`. Without two-section comments, the
  branch title defaults to the enum value name. Rules such as `enum.const`/`enum.in`/`enum.not_in`
  still filter which branches are emitted. The enum zero value (number 0, e.g. the
  `*_UNSPECIFIED` or `*_UNKNOWN` branch) is dropped by default; use `enum_zero_value=true` to emit
  it, and `required` no longer affects whether it is emitted. In non-strict mode the implicit
  default value (the zero value) is still rendered as the string enum value name so it is
  consistent with the `const` branches. For example, given `// High` `//` `// An urgent problem.`
  on `SEVERITY_HIGH = 2;`, the generated branch is
  `{"const": "SEVERITY_HIGH", "title": "High", "description": "An urgent problem."}`.
- `enum_zero_value` - If `true`, the enum zero value (number 0) branch is included in the
  generated schema as an allowed value. By default (unset or `false`) the zero value branch is
  dropped regardless of whether the field is `required`. Defaults to `false`.

### Custom extension properties

The plugin supports customizing the generated JSON Schema via
[custom options](https://protobuf.dev/programming-guides/proto2/#customoptions). The `title`,
`description`, and `default` fields directly override the corresponding standard JSON Schema
keywords, `properties` injects arbitrary vendor extension keywords (e.g. `x-foo`), and external
extension types can embed custom JSON.

The options are defined in `buf/protoschema/custom/v1/custom_options.proto`. Import it in your
proto file:

```proto
import "buf/protoschema/custom/v1/custom_options.proto";
```

**title, description, and default.** These fields directly override the standard keywords and
apply to both messages and fields:

```proto
message User {
  option (buf.protoschema.custom.v1.message_options) = {
    title: "User"
    description: "A user of the system."
  };

  int32 score = 1 [(buf.protoschema.custom.v1.field_options) = {
    title: "Score"
    default: {number_value: 0}
  }];
}
```

`default` is a `google.protobuf.Value`, so any JSON value can be expressed: number, string, bool,
null, object, or array.

**properties.** A `google.protobuf.Struct` of arbitrary keywords, useful for custom keywords that
have no dedicated field. Keys are written into the schema verbatim, so use the `x-` prefix to
avoid colliding with the standard JSON Schema keywords:

```proto
message User {
  string name = 1 [(buf.protoschema.custom.v1.field_options).properties = {
    fields: {
      key: "x-ref"
      value: {string_value: "/users/1"}
    }
  }];
}
```

**External extension types.** `CustomOptions` is open to extension. Define your own message and
`extend buf.protoschema.custom.v1.CustomOptions`. Note that proto3 only allows extending descriptor
options, so external extensions must be declared in a proto2 file:

```proto
// In a proto2 file.
message VendorExt {
  optional string entity = 1;
  repeated string tags = 2;
}

extend buf.protoschema.custom.v1.CustomOptions {
  optional VendorExt vendor = 100;
}
```

Then use it on any message or field:

```proto
message User {
  option (buf.protoschema.custom.v1.message_options) = {
    title: "User"
    [my.pkg.vendor]: {entity: "user", tags: ["a", "b"]}
  };
}
```

The extension's message is serialized to JSON and its keys are merged into the schema.

For example, the `CustomVendor` message in [testdata](/internal/testdata/jsonschema/) generates the
following schema (abbreviated):

```json
{
  "title": "A Custom Vendor",
  "description": "A vendor whose schema is customized.",
  "entity": "user",
  "tags": ["a", "b"],
  "properties": {
    "name": { "type": "string", "x-example": "alice" },
    "price": { "type": "integer", "default": 5, "title": "Price" }
  }
}
```

Custom keywords are merged last, so a conflicting key overrides the standard keyword it collides
with. Prefer the `x-` prefix convention (or name extension fields `x-*`) so custom keywords never
shadow standard ones.

### Structured DSL: allOf, if/then/else, and visibility control

Writing nested `google.protobuf.Struct` values for complex structures is verbose. To cover the
common cases concisely, `CustomOptions` adds three structured fields:

- `all_of` (`repeated Constraints`) → compiles to the schema's `allOf` array; each member is ANDed
  with the generated schema.
- `conditional` (`Conditional`) → compiles to `if`/`then`/`else` for cross-field constraints such
  as conditional required fields.
- `display_if` (`Condition`) → compiles to the `x-display-if` vendor keyword as a UI visibility
  hint (not a validation assertion).

`Constraints` supports `required`, `field_constraints` (per-field `minItems`/`const`), and verbatim
`properties`. `Condition` supports dotted field paths with `exists`/`equals_bool`/`equals_int`/
`equals_string`/`equals_value`.

For example, "require `rules` to be non-empty only when `base.enabled` is `true`":

```proto
message WinLimitStrategy {
  option (buf.protoschema.custom.v1.message_options) = {
    conditional: {
      if:   { field: "base.enabled" equals_bool: true }
      then: { required: "rules" field_constraints: { field: "rules" min_items: 1 } }
    }
  };
  StrategyBase base = 1;
  repeated WinLimitRule rules = 2;
}
```

generates:

```json
{
  "type": "object",
  "properties": {
    "base":  { "$ref": "demo.WinLimitStrategy.StrategyBase.schema.json" },
    "rules": { "type": "array", "items": { "$ref": "demo.WinLimitStrategy.WinLimitRule.schema.json" } }
  },
  "if": {
    "required": ["base"],
    "properties": {
      "base": { "required": ["enabled"], "properties": { "enabled": { "const": true } } }
    }
  },
  "then": { "required": ["rules"], "properties": { "rules": { "minItems": 1 } } }
}
```

Visibility control (show a field only when `base.enabled` is `true`):

```proto
string reward_rule = 3 [(buf.protoschema.custom.v1.field_options) = {
  display_if: { field: "base.enabled" equals_bool: true }
}];
```

generates `"reward_rule": {"type": "string", "x-display-if": {"field": "base.enabled", "equals": true}}`.

`allOf` combination:

```proto
option (buf.protoschema.custom.v1.message_options) = {
  all_of: { required: "foo" field_constraints: { field: "bar" min_items: 2 } }
  all_of: { properties: { fields: { key: "x-meta", value: { string_value: "v" } } } }
};
```

generates `"allOf": [{"required": ["foo"], "properties": {"bar": {"minItems": 2}}}, {"x-meta": "v"}]`.

### Shorthand fields for common keywords

Common JSON Schema keywords and the project's vendor extensions get dedicated shorthand fields on
`CustomOptions`, so you don't have to nest a `google.protobuf.Struct` inside `properties`:

| Field | Schema keyword | Type |
|---|---|---|
| `permission_level` | `x-permission-level` | `int32` (1/2/3; 0 or unset emits nothing) |
| `datasource` | `x-datasource` | `string` |
| `items_from` | `x-items-from` | `string` |
| `renderer` | `x-renderer` | `string` |
| `examples` | `examples` | `repeated google.protobuf.Value` |
| `enum` | `enum` | `repeated google.protobuf.Value` |
| `unique_items` | `uniqueItems` | `bool` |
| `unique_item_properties` | `uniqueItemProperties` | `repeated string` |
| `format` | `format` | `string` |

These run after the `properties` pass-through, so an explicit shorthand overrides a vendor key
injected via `properties` with the same name. Example:

```proto
message PoolItem {
  option (buf.protoschema.custom.v1.message_options) = {
    permission_level: 3
    unique_item_properties: "item_id"
  };

  string item_id = 1 [(buf.protoschema.custom.v1.field_options) = {
    items_from: "#/module/tg_backpack/item_groups/*/items/*|item_id|name"
    permission_level: 2
    format: "uri"
  }];
}
```

generates `item_id` as `"item_id": {"type": "string", "x-items-from": "#/module/tg_backpack/item_groups/*/items/*|item_id|name", "x-permission-level": 2, "format": "uri"}`.

## Community

For help and discussion around Protobuf, best practices, and more, join us
on [Slack][badges_slack].

## Status

This project is currently in **alpha**. The API should be considered unstable and likely to change.

## Legal

Offered under the [Apache 2 license][license].

[badges_ci]: https://github.com/BandlSkyler/protoschema-plugins/actions/workflows/ci.yaml
[badges_goreportcard]: https://goreportcard.com/report/github.com/BandlSkyler/protoschema-plugins
[badges_godoc]: https://pkg.go.dev/github.com/BandlSkyler/protoschema-plugins
[badges_slack]: https://join.slack.com/t/bufbuild/shared_invite/zt-f5k547ki-dW9LjSwEnl6qTzbyZtPojw
[license]: https://github.com/BandlSkyler/protoschema-plugins/blob/main/LICENSE.txt
