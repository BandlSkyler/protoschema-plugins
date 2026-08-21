# protoschema-plugins

[![Build](https://github.com/BandlSkyler/protoschema-plugins/actions/workflows/ci.yaml/badge.svg?branch=main)][badges_ci]
[![Report Card](https://goreportcard.com/badge/github.com/BandlSkyler/protoschema-plugins)][badges_goreportcard]
[![GoDoc](https://pkg.go.dev/badge/github.com/BandlSkyler/protoschema-plugins.svg)][badges_godoc]
[![Slack](https://img.shields.io/badge/slack-buf-%23e01563)][badges_slack]

protoschema-plugins 仓库包含一组 Protobuf 插件,用于从 protobuf 文件生成不同类型的 schema。包括:

- [PubSub](#pubsub-protobuf-schema)
- [JSON Schema](#json-schema)

## PubSub Protobuf Schema

根据给定的 protobuf 文件生成一个 schema,可作为 PubSub schema 使用,形式为单个自包含的、
归一化为 proto2 的 message。

直接安装 `protoc-gen-pubsub` 插件:

```sh
go install github.com/BandlSkyler/protoschema-plugins/cmd/protoc-gen-pubsub@latest
```

或在 `buf.gen.yaml` 中作为 [远程插件](https://buf.build/docs/generate/remote-plugins) 引用:

```yaml
version: v1
plugins:
  - plugin: buf.build/bufbuild/protoschema-pubsub
    out: ./gen
```

示例见 [testdata](/internal/testdata/pubsub/),其中包含针对 [proto](/proto/)
中测试用例定义生成的 schema。

## JSON Schema

根据给定的 protobuf 文件生成 [JSON Schema](https://json-schema.org/)。本实现使用最新的
[JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12/release-notes)。

直接安装 `protoc-gen-jsonschema`:

```sh
go install github.com/BandlSkyler/protoschema-plugins/cmd/protoc-gen-jsonschema@latest
```

或在 `buf.gen.yaml` 中作为 [远程插件](https://buf.build/docs/generate/remote-plugins) 引用:

```yaml
version: v1
plugins:
  - plugin: buf.build/bufbuild/protoschema-jsonschema
    out: ./gen
```

示例见 [testdata](/internal/testdata/jsonschema/),其中包含针对 [proto](/proto/)
中测试用例定义生成的 schema。

以下 protobuf 生成的一个简单示例 schema:

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

默认情况下,会生成以下 JSON Schema 文件:

- `*.schema.json` — 使用 protobuf 字段名生成(如 `product_id`、`product_name`)。
- `*.schema.bundle.json` — 将所有依赖合并到单个文件中,使用 protobuf 字段名。
- `*.schema.strict.json` — 使用 protobuf 字段名生成,但不允许别名、字符串数字或其他任何非归一化表示。
- `*.schema.strict.bundle.json` — 包含 strict schema,并将所有依赖合并到单个文件中,使用 protobuf 字段名。
- `*.jsonschema.json` — 使用 JSON 字段名生成(如 `productId`、`productName`)。
- `*.jsonschema.bundle.json` — 将所有依赖合并到单个文件中,使用 JSON 字段名。
- `*.jsonschema.strict.json` — 使用 JSON 字段名生成,但不允许别名、字符串数字或其他任何非归一化表示。
- `*.jsonschema.strict.bundle.json` — 包含 strict JSON schema,并将所有依赖合并到单个文件中,使用 JSON 字段名。

例如,上述 protobuf 生成以下 `*.schema.json` 文件:

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

或者以下 `*.jsonschema.strict.bundle.json` 文件:

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

### 选项(Options)

JSON Schema 插件支持以下选项:

- `target` — 为 `proto`、`json`、`proto-bundle`、`json-bundle`、`proto-strict`、`json-strict`、
  `proto-strict-bundle`、`json-strict-bundle` 或 `all` 中的任意值,用 `+` 分隔(如 `proto+json`)。默认为 `all`。
  - 如果为 `proto`,schema 将使用 Protobuf 字段名生成(如 `product_id`、`product_name`)。
  - 如果为 `json`,schema 将使用 JSON 字段名生成(如 `productId`、`productName`)。
  - 如果后缀为 `-bundle`,schema 会将所有依赖合并到单个文件中。
  - 如果后缀为 `-strict`,schema 将不允许别名、字符串数字或其他任何非归一化表示。当被校验的 JSON 数据直接使用、
    而非转换为 Protobuf message 时,strict 模式很有用。使用 [Protobuf JSON](https://protobuf.dev/programming-guides/json/#json-options)
    时需要开启 "always emit fields without presence" 选项。
  - 如果后缀为 `-strict-bundle`,schema 将同时满足 strict 和 bundle(合并依赖)两种行为。
- `additional_properties` — 如果为 `true`,生成的 schema 会将 `additionalProperties` 设为 `true`,
  使未知字段被忽略而不是报错。默认为 `false`。当客户端/发送方与服务器/接收方的 schema 版本不同时很有用,
  与 [Protobuf JSON](https://protobuf.dev/programming-guides/json/#json-options) 的
  "ignore unknown fields" 选项类似。
- `non_required_default` — 如果为 `true`,字段默认非必填:只有显式声明
  `(buf.validate.field).required = true` 的字段才会进入 `required` 数组。同时会禁用 strict 模式对
  非 optional(隐式默认值)字段的强制必填。默认为 `false`。
- `schema_version` — 生成的 JSON Schema 草稿版本。`2020-12`(默认)或 `draft-07`。
  当消费方只支持较旧的草稿时,`draft-07` 很有用。
- `enum_oneof` — 如果为 `true`,枚举字段将被渲染为字符串类型 + 一个 `oneOf` 的
  `{const, title, description}` 分支列表,每个允许的枚举值一个分支,替代默认的 `anyOf`(字符串枚举名
  + 整数范围)。`const` 为 proto 枚举值名(即 protobuf JSON 的标准表示)。每个枚举值的前置注释按空行
  拆成两段:第一段作为分支的 `title`(可作为值的"别名"),第二段作为分支的 `description`。如果没有
  两段注释,分支 `title` 默认用枚举值名。`enum.const`/`enum.in`/`enum.not_in` 等规则仍会过滤最终
  输出的分支。枚举零值(编号 0,如 `*_UNSPECIFIED`/`*_UNKNOWN`)分支默认被剔除,需用
  `enum_zero_value=true` 显式开启;`required` 不再影响零值分支的取舍。非 strict 模式下,隐式默认值
  (零值)仍会以字符串枚举值名的形式渲染,与 `const` 分支保持一致。默认为 `false`。
- `enum_zero_value` — 如果为 `true`,枚举零值(编号 0)分支会作为允许值出现在生成的 schema 中。
  默认(未指定或 `false`)时,无论字段是否 `required`,零值分支都会被剔除。默认为 `false`。

> **message 级 oneof**：proto 中声明的 message 级 oneof（非 proto3 `optional` 生成的 synthetic
> oneof）会渲染为 schema 顶层的 `oneOf` 列表：每个成员一个分支 `{type: object, required:
> [member]}`，强制"恰好一个成员存在"。成员字段仍保留在顶层 `properties` 中（title/description/
> format/自定义扩展照常生效），因此字段约束在顶层校验、错误路径清晰，不会被 oneOf 匹配吞并。
> proto3 `optional` 字段生成的 synthetic oneof 会被排除，仍作为普通可选字段渲染。这是默认行为，
> 无需开关。

### 自定义扩展属性(Custom extension properties)

该插件支持通过 [自定义 options](https://protobuf.dev/programming-guides/proto2/#customoptions)
定制生成的 JSON Schema。`title`、`description`、`default` 三个字段会直接覆盖对应的标准 JSON
Schema 关键字,`properties` 可以注入任意 vendor 扩展关键字(如 `x-foo`),同时支持外部自定义
扩展类型嵌入自定义 JSON。

这些选项定义在 `buf/protoschema/custom/v1/custom_options.proto` 中。在你的 proto 文件中引入:

```proto
import "buf/protoschema/custom/v1/custom_options.proto";
```

**title、description、default。** 这三个字段直接覆盖标准关键字,message 级和 field 级均适用:

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

`default` 是 `google.protobuf.Value`,因此可以表达任意 JSON 值:数字、字符串、布尔、null、对象或数组。

**properties。** 一个 `google.protobuf.Struct`,用于没有专属字段的自定义关键字。关键字会原样
写入 schema,因此请使用 `x-` 前缀,避免与标准 JSON Schema 关键字冲突:

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

**外部扩展类型。** `CustomOptions` 开放了扩展。你可以定义自己的 message 并
`extend buf.protoschema.custom.v1.CustomOptions`。注意:proto3 只允许扩展 descriptor options,
因此外部扩展必须在 **proto2 文件**中声明:

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

然后在任意 message 或字段上使用:

```proto
message User {
  option (buf.protoschema.custom.v1.message_options) = {
    title: "User"
    [my.pkg.vendor]: {entity: "user", tags: ["a", "b"]}
  };
}
```

扩展的 message 会被序列化为 JSON,其键值合并进 schema。

例如 [testdata](/internal/testdata/jsonschema/) 中的 `CustomVendor` message 会生成如下
schema(略简):

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

自定义关键字会最后合并,因此冲突的 key 会覆盖与之冲突的标准关键字。请优先使用 `x-` 前缀约定
(或将扩展字段命名为 `x-*`),确保自定义关键字永远不会遮蔽标准关键字。

### 结构化 DSL:allOf、if/then/else 与显隐控制

上面的 `properties` 需要手写嵌套的 `google.protobuf.Struct`,表达复杂结构时较长。
为此 `CustomOptions` 提供三组结构化字段,把高频场景压缩到几行:

- `all_of`(`repeated Constraints`)→ 编译为 schema 的 `allOf` 数组,每个成员与生成的
  schema 取 AND。
- `conditional`(`Conditional`)→ 编译为 `if`/`then`/`else`,用于条件必填等跨字段约束。
- `display_if`(`Condition`)→ 编译为 `x-display-if` vendor 关键字,作为 UI 层的显隐
  提示(不是校验断言)。

`Constraints` 支持 `required`、`field_constraints`(字段级 `minItems`/`const`)和任意
`properties` 透传;`Condition` 支持点号字段路径与 `exists`/`equals_bool`/`equals_int`/
`equals_string`/`equals_value`。

例如,"仅当 `base.enabled` 为 `true` 时才要求 `rules` 非空":

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

生成:

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

显隐控制(仅当 `base.enabled` 为 `true` 时才显示该字段):

```proto
string reward_rule = 3 [(buf.protoschema.custom.v1.field_options) = {
  display_if: { field: "base.enabled" equals_bool: true }
}];
```

生成 `"reward_rule": {"type": "string", "x-display-if": {"field": "base.enabled", "equals": true}}`。

`allOf` 组合:

```proto
option (buf.protoschema.custom.v1.message_options) = {
  all_of: { required: "foo" field_constraints: { field: "bar" min_items: 2 } }
  all_of: { properties: { fields: { key: "x-meta", value: { string_value: "v" } } } }
};
```

生成 `"allOf": [{"required": ["foo"], "properties": {"bar": {"minItems": 2}}}, {"x-meta": "v"}]`。

### 常用关键字的简写字段

常用 JSON Schema 关键字与项目自定义 vendor 扩展在 `CustomOptions` 上都有对应的简写字段，
无需再在 `properties` 里嵌套 `google.protobuf.Struct`：

| 字段 | Schema 关键字 | 类型 |
|---|---|---|
| `permission_level` | `x-permission-level` | `int32`（1/2/3；0 或未设置则不输出） |
| `datasource` | `x-datasource` | `string` |
| `items_from` | `x-items-from` | `string` |
| `renderer` | `x-renderer` | `string` |
| `examples` | `examples` | `repeated google.protobuf.Value` |
| `enum` | `enum` | `repeated google.protobuf.Value` |
| `unique_items` | `uniqueItems` | `bool` |
| `unique_item_properties` | `uniqueItemProperties` | `repeated string` |
| `format` | `format` | `string` |

简写展开在 `properties` 透传**之后**执行，因此显式简写会覆盖 `properties` 里同名 vendor 关键字。示例：

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

生成 `item_id` 为 `"item_id": {"type": "string", "x-items-from": "#/module/tg_backpack/item_groups/*/items/*|item_id|name", "x-permission-level": 2, "format": "uri"}`。

## 社区

关于 Protobuf、最佳实践等更多帮助与讨论,欢迎加入我们的 [Slack][badges_slack]。

## 状态

本项目目前处于 **alpha** 阶段。API 应视为不稳定,且可能发生变化。

## 法律声明

依据 [Apache 2 license][license] 提供。

[badges_ci]: https://github.com/BandlSkyler/protoschema-plugins/actions/workflows/ci.yaml
[badges_goreportcard]: https://goreportcard.com/report/github.com/BandlSkyler/protoschema-plugins
[badges_godoc]: https://pkg.go.dev/github.com/BandlSkyler/protoschema-plugins
[badges_slack]: https://join.slack.com/t/bufbuild/shared_invite/zt-f5k547ki-dW9LjSwEnl6qTzbyZtPojw
[license]: https://github.com/BandlSkyler/protoschema-plugins/blob/main/LICENSE.txt
