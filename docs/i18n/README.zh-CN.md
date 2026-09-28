[English](../../README.md) · **简体中文** · [Русский](README.ru.md) · [हिन्दी](README.hi.md)

<p align="center">
  <img src="asset-banner.png" alt="goldenthread">
</p>

> Schema 编译器：Go 结构体 → TypeScript/Zod。自动保持后端/前端校验同步。

[![Blackwell Systems™](https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg)](https://github.com/blackwell-systems) 
[![Go Reference](https://pkg.go.dev/badge/github.com/blackwell-systems/goldenthread.svg)](https://pkg.go.dev/github.com/blackwell-systems/goldenthread) 
[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://go.dev/) 
[![CI](https://github.com/blackwell-systems/goldenthread/workflows/CI/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/ci.yml)
[![Lint](https://github.com/blackwell-systems/goldenthread/workflows/Lint/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/lint.yml)

[![License](https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue.svg)](LICENSE-APACHE) 
[![Sponsor](https://img.shields.io/badge/Sponsor-Buy%20Me%20a%20Coffee-yellow?logo=buy-me-a-coffee&logoColor=white)](https://buymeacoffee.com/blackwellsystems)

**goldenthread** 是一款构建期 schema 编译器，可从 Go 结构体生成生产就绪的 Zod 校验。校验规则只需在 Go 标签中编写一次，即可自动获得类型安全的 TypeScript schema。内置的漂移检测能在 CI 中捕获 schema 不一致。无需手动同步。无运行时开销。

## 概述

在 Go 后端与 TypeScript 前端之间维护校验，意味着要让多套表示保持同步。goldenthread 直接从 Go 结构体标签生成 Zod schema：

```go
// Go: Define once with validation tags
type User struct {
  Username string `json:"username" gt:"required,len:3..20"`
  Email    string `json:"email" gt:"email"`
  Age      int    `json:"age" gt:"min:13,max:130"`
}
```

```bash
# Generate Zod schemas
goldenthread generate ./models
```

```typescript
// TypeScript: Use generated schemas
import { UserSchema, User } from './gen/user'

// Runtime validation
const result = UserSchema.safeParse(data)

// Type inference
const user: User = {
  username: 'alice',
  email: 'alice@example.com',
  age: 25
}
```

对 Go 结构体的更改会自动重新生成 TypeScript schema。编译器确保二者始终保持同步。

## 特性

**goldenthread 生成具备完整类型安全的生产就绪 Zod schema。**

### 你将获得

- **完整的 Go 类型支持**：基本类型、数组、映射、枚举、嵌套对象、指针
- **全面的校验**：长度边界、数值范围、正则模式、格式校验器（email、UUID、URL、IPv4/IPv6、datetime）
- **枚举生成**：从 Go 字符串字段生成 `z.enum(['pending', 'completed'])`
- **可辨识联合**：从 Go 的 one-of 结构体生成 `z.discriminatedUnion('kind', [...])`
- **映射支持**：为 Go 映射生成 `z.record(z.string(), T)`
- **数组校验**：最小/最大长度约束
- **嵌套对象**：对其他 schema 的类型安全引用
- **嵌入结构体扁平化**：匿名字段自动提升
- **冲突检测**：在编译期（而非运行期）捕获重复的 JSON 键
- **漂移检测**：当 schema 失去同步时，`goldenthread check` 会使 CI 失败
- **零运行时开销**：纯代码生成，无反射，无魔法

### 校验规则

```go
type Product struct {
  // String validation
  Name  string `gt:"required,len:1..100"`
  SKU   string `gt:"pattern:^[A-Z0-9]{8}$"`
  
  // Numeric validation  
  Price float64 `gt:"required,min:0,max:999999.99"`
  Stock int     `gt:"min:0"`
  
  // Format validation
  Email  string `gt:"email"`
  WebURL string `gt:"url"`
  ID     string `gt:"uuid,required"`
  
  // Enum validation
  Status string `gt:"enum:draft,published,archived"`
  
  // Optional fields (pointer or omitempty)
  Notes *string `gt:"len:0..500"`
  Tags  []string `json:"tags,omitempty"`
  
  // Map support
  Metadata map[string]string `gt:"optional"`
}
```

完整的标签语法见 [docs/TAG_SPEC.md](docs/TAG_SPEC.md)。

## 安装

```bash
go install github.com/blackwell-systems/goldenthread/cmd/goldenthread@latest
```

## 快速开始

### 1. 定义你的模型

```go
// models/user.go
package models

// User represents a system user with validated fields.
type User struct {
  // ID is the unique identifier
  ID string `json:"id" gt:"uuid,required"`
  
  // Username must be 3-20 characters
  Username string `json:"username" gt:"required,len:3..20"`
  
  // Email must be valid format
  Email string `json:"email" gt:"email"`
  
  // Age must be between 13 and 130
  Age int `json:"age" gt:"min:13,max:130"`
  
  // Bio is optional with max length
  Bio *string `json:"bio" gt:"len:0..500"`
}
```

### 2. 生成 schema

```bash
# Generate from current directory
goldenthread generate ./models

# Generate recursively
goldenthread generate ./models --recursive

# Specify output directory
goldenthread generate ./models --out ./frontend/src/schemas
```

输出：
```
gen/
  user.ts                    # Generated Zod schema
  .goldenthread.json         # Metadata for drift detection
```

### 3. 在 TypeScript 中使用

```typescript
import { UserSchema, User } from './gen/user'

// Parse and validate API response
const response = await fetch('/api/users/123')
const data = await response.json()
const result = UserSchema.safeParse(data)

if (!result.success) {
  console.error('Validation failed:', result.error)
  return
}

// Type-safe access
const user: User = result.data
console.log(user.username) // Type: string
console.log(user.bio)      // Type: string | undefined
```

### 4. 验证 schema 保持同步

```bash
# Check if generated schemas match source
goldenthread check ./models

# Returns exit code 1 if out of sync (CI-friendly)
```

添加到 CI：
```yaml
- name: Check schema drift
  run: goldenthread check ./models
```

## 架构

goldenthread 是一个包含三个阶段的 schema 编译器：

```
Go Source → Parser → Intermediate Representation → Emitter → Generated Code
            (AST)         (Language-agnostic)        (Zod)    (TypeScript)
```

### 流水线组件

**1. 解析器**（`internal/parser`）

使用 `go/packages` 和 `go/types` 进行正确的类型解析。提取带 `gt:` 标签的结构体定义，校验标签语法与冲突，处理嵌入结构体和跨包引用。

**2. 中间表示**（`internal/schema`）

一种与语言无关的 schema 格式，将解析与代码生成分离。使得将来能在不修改解析器的情况下增加新的 emitter（OpenAPI、JSON Schema 等）。

**3. 归一化**（`internal/normalize`）

扁平化嵌入结构体字段，检测 Go 字段名冲突，校验 JSON 名冲突，在发射前确保 schema 的正确性。

**4. 发射器**（`internal/emitter/zod`）

生成带 Zod schema 的 TypeScript。将 Go 文档保留为 JSDoc，发射类型安全的 `z.infer<>` 类型，产出确定性输出以获得稳定的 diff。

**5. 哈希/漂移检测**（`internal/hash`）

对 schema 内容进行 SHA-256 哈希，并跟踪元数据（`.goldenthread.json`）。检测源代码与生成代码何时出现分歧。

## 支持的类型

| Go Type | Zod Output | Notes |
|---------|-----------|-------|
| `string` | `z.string()` | With validation rules |
| `int`, `int64`, `float64`, etc. | `z.number()` | All numeric types |
| `bool` | `z.boolean()` | Boolean values |
| `[]T` | `z.array(T)` | Arrays/slices |
| `map[string]T` | `z.record(z.string(), T)` | String-keyed maps |
| `time.Time` | `z.string().datetime()` | ISO 8601 datetime |
| `*T` | `T.optional()` | Pointers become optional |
| Named struct | `TypeSchema` | References to other schemas |
| Embedded struct | Fields flattened | Anonymous fields promoted |

完整的类型覆盖见 [docs/FEATURES.md](docs/FEATURES.md)。

## 校验标签

### 存在性

- `required` - 字段必须存在（非指针的默认值）
- `optional` - 字段可省略

### 字符串规则

- `len:M..N` - 长度介于 M 到 N 个字符之间
- `min:N` - 最小长度（`len:N..` 的简写）
- `max:N` - 最大长度（`len:..N` 的简写）
- `pattern:REGEX` - 必须匹配正则表达式

### 数值规则

- `min:N` - 最小值
- `max:N` - 最大值

### 格式校验器

- `email` - 有效的电子邮件地址
- `uuid` - 有效的 UUID 字符串
- `url` - 有效的 URL
- `date` - ISO 8601 日期（YYYY-MM-DD）
- `datetime` - ISO 8601 datetime
- `ipv4` - IPv4 地址
- `ipv6` - IPv6 地址

### 枚举

- `enum:value1,value2,value3` - 指定值之一

### 数组规则

- `min:N` - 最小数组长度
- `max:N` - 最大数组长度

详细规范见 [docs/TAG_SPEC.md](docs/TAG_SPEC.md)。

## 示例

### 嵌套对象

```go
type Address struct {
  Street  string `json:"street" gt:"required"`
  City    string `json:"city" gt:"required"`
  ZipCode string `json:"zip_code" gt:"pattern:^[0-9]{5}$"`
}

type User struct {
  Name    string  `json:"name" gt:"required"`
  Address Address `json:"address" gt:"required"`
}
```

生成嵌套的 Zod schema：
```typescript
const AddressSchema = z.object({
  street: z.string(),
  city: z.string(),
  zip_code: z.string().regex(/^[0-9]{5}$/)
})

const UserSchema = z.object({
  name: z.string(),
  address: AddressSchema
})
```

### 嵌入结构体

```go
type Timestamps struct {
  CreatedAt string `json:"created_at" gt:"datetime,required"`
  UpdatedAt string `json:"updated_at" gt:"datetime,required"`
}

type Article struct {
  Timestamps  // Fields automatically flattened
  
  Title   string `json:"title" gt:"required,len:1..200"`
  Content string `json:"content" gt:"required"`
}
```

生成扁平化的 schema：
```typescript
const ArticleSchema = z.object({
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
  title: z.string().min(1).max(200),
  content: z.string()
})
```

### 枚举

```go
type Task struct {
  Title    string `json:"title" gt:"required"`
  Status   string `json:"status" gt:"enum:pending,in_progress,completed,cancelled"`
  Priority string `json:"priority" gt:"enum:low,medium,high"`
}
```

生成枚举 schema：
```typescript
const TaskSchema = z.object({
  title: z.string(),
  status: z.enum(['pending', 'in_progress', 'completed', 'cancelled']),
  priority: z.enum(['low', 'medium', 'high'])
})
```

### 可辨识联合

一个 Go 结构体，若带有 `discriminator` 字段，并且每个变体都有一个 `variant:<name>`
载荷字段，会被编译为 Zod 可辨识联合：

```go
// WiringElement is exactly one of edge, switch, or join.
type WiringElement struct {
  Kind   string      `json:"kind" gt:"discriminator"`
  Edge   *EdgeSpec   `json:"edge,omitempty"   gt:"variant:edge"`
  Switch *SwitchSpec `json:"switch,omitempty" gt:"variant:switch"`
  Join   *JoinSpec   `json:"join,omitempty"   gt:"variant:join"`
}
```

生成一个 TypeScript 会依据 `kind` 收窄的可辨识联合：

```typescript
const WiringElementSchema = z.discriminatedUnion('kind', [
  z.object({ kind: z.literal('edge'),   edge: EdgeSpecSchema }),
  z.object({ kind: z.literal('switch'), switch: SwitchSpecSchema }),
  z.object({ kind: z.literal('join'),   join: JoinSpecSchema })
])
```

载荷在其变体内以必填形式发射：一个标记为 `edge` 且没有 `edge` 载荷的元素
将无法通过校验。参见 [examples/wiring](examples/wiring/)。

### 映射

```go
type Config struct {
  Name     string            `json:"name" gt:"required"`
  Settings map[string]string `json:"settings" gt:"required"`
  Metadata map[string]any    `json:"metadata" gt:"optional"`
}
```

生成 record schema：
```typescript
const ConfigSchema = z.object({
  name: z.string(),
  settings: z.record(z.string(), z.string()),
  metadata: z.record(z.string(), z.any()).optional()
})
```

## CLI 命令

### `generate`

从 Go 源代码生成 Zod schema：

```bash
goldenthread generate [flags] <directory>

Flags:
  --out <dir>       Output directory (default: ./gen)
  --recursive       Process subdirectories recursively
  --target <name>   Target format (default: zod)
  --infer-json      Infer schemas from structs carrying only json: tags
```

### `check`

验证生成的 schema 与源代码匹配：

```bash
goldenthread check [flags] <directory>

Flags:
  --metadata <file>  Metadata file to check against
  --recursive        Process subdirectories recursively
  --infer-json       Infer schemas from structs carrying only json: tags
```

退出码：
- `0` - schema 已同步
- `1` - schema 失去同步或出错

### `init` _(计划于 v0.2)_

为项目初始化 goldenthread 配置。

> **注意：** 此命令尚未实现。`goldenthread init --wails`
> （Wails 项目自动设置）计划于 v0.2 推出。当前运行 `goldenthread init`
> 只会打印一条占位消息。

## 从 json: 标签推断

默认情况下，goldenthread 只有在结构体的某个字段带有 `gt:` 标签时才为其生成
schema。只带标准 `json:` 标签的结构体会被跳过。`--infer-json` 标志启用第二种
模式：当它被设置时，goldenthread 也会从只带 `json:` 标签的结构体生成 schema，
并根据每个字段的 json 标签推导该字段。

```bash
goldenthread generate ./models --infer-json
goldenthread check ./models --infer-json
```

用它来桥接来自外部框架的类型，这些框架已经为其结构体标注了 JSON 编码标签，
从而无需为每个字段添加 `gt:` 注解即可产出 Zod。在推断模式下：

- 字段名即 json 标签名。标记为 `json:"-"` 的字段被排除。
- 当字段的 json 标签带有 `,omitempty` 或该 Go 字段是指针时，字段为可选；
  否则为必填。
- 基础类型来自与 gt 标签字段相同的 Go-to-Zod 映射：基本类型、`[]T`、
  `map[string]T`、`*T` 以及命名结构体引用。类型为另一个结构体（或其切片）的
  字段会发射对该结构体 schema 的引用，且被引用的结构体也会被生成。

`gt:` 标签始终优先。带有 `gt:` 标签的字段或结构体保持其原有的确切行为：
校验规则、枚举和可辨识联合均不变，推断只填补缺少 `gt:` 的字段和结构体。
允许混用：一个结构体可以有一些 gt 标签字段和一些仅带 json 的字段，在
`--infer-json` 下两者都会出现在输出中。不带该标志时，行为不变：仅带 json 的
结构体仍不产生任何内容。

## 当前限制

<details>
<summary>goldenthread v0.1 focuses on the core use case: struct validation for APIs and forms. Click to see what's not yet supported.</summary>

### 类型系统

- Go 联合类型（未加标签、使用类型断言的接口），请改用可辨识联合
- 定长数组（`[3]int`）
- 字面量常量值
- 递归/自引用类型

### 校验

- 自定义校验函数
- 跨字段校验（密码确认等）
- 数组中的唯一项
- 条件校验

### 发射器

- OpenAPI/Swagger 生成
- JSON Schema 生成
- TypeScript 类型（独立于 Zod）
- 其他校验库

### 标签特性

- 标签继承/组合
- 条件规则
- 每个结构体多套校验集

这些限制对于 v0.1 是有意为之。该工具**把一件事做好**：从 Go 结构体（包括基本类型、枚举、数组、映射和嵌套对象）生成类型安全的 Zod schema。未来版本可能会根据实际使用情况扩展范围。

</details>

## 与替代方案的比较

| Tool | Go → TS | Validation | Approach | Use Case |
|------|---------|-----------|----------|----------|
| `validator.v10` | No | Yes | Runtime tags | Go-only validation |
| `swaggo/swag` | No | No | Comments | OpenAPI from Go |
| `oapi-codegen` | Yes | Yes | OpenAPI → Go | Contract-first APIs |
| `protobuf` | Yes | Limited | `.proto` files | Cross-language RPC |
| **goldenthread** | Yes | Yes | Go → Zod | Go-first validation |

goldenthread 为以下团队优化：
- 后端用 Go 编写
- 前端用 TypeScript 编写
- 希望以 Go 结构体作为唯一可信来源
- 需要在两种语言中进行运行时校验
- 偏好类型安全甚于灵活性

## 文档

- [Tag Specification](docs/TAG_SPEC.md) - 完整的标签语法参考
- [Feature Matrix](docs/FEATURES.md) - 详细的能力细分
- [Architecture](docs/ARCHITECTURE.md) - 系统设计与实现细节
- [Testing Strategy](docs/TESTING.md) - 测试套件与持续模糊测试指南
- [Fuzzing Bug Log](docs/FUZZING_BUGS.md) - 持续模糊测试发现的缺陷
- [Brand Guidelines](BRAND.md) - 商标与徽标使用
- [Changelog](CHANGELOG.md) - 版本历史与变更

## 开发

### 项目结构

```
goldenthread/
├── cmd/goldenthread/      # CLI tool
├── internal/
│   ├── parser/            # Go AST parsing with go/packages
│   ├── schema/            # Intermediate representation
│   ├── normalize/         # Schema validation and flattening
│   ├── emitter/zod/       # Zod schema generation
│   ├── hash/              # Drift detection
│   └── load/              # Package loading wrapper
├── examples/              # Test cases
└── docs/                  # Documentation

```

### 构建

```bash
# Build CLI tool
go build ./cmd/goldenthread

# Run tests
go test ./...

# Run with coverage
go test ./... -cover
```

### 贡献

欢迎贡献！关注方向：

1. 更多 emitter（OpenAPI、JSON Schema）
2. 更多校验规则
3. 测试覆盖率改进
4. 文档与示例

## 理念

goldenthread 遵循以下原则：

1. **Go 结构体是权威来源** - 不是配置文件，不是 DSL，不是注解
2. **构建期生成** - 不是运行时反射或代理
3. **处处类型安全** - 在编译期捕获错误
4. **简单胜于聪明** - 可预测的行为胜过灵活性
5. **快速失败** - 未知的 token 立即报错

“goldenthread”这个名字源自传统织造：那根贯穿始终、将织物维系在一起的连续丝线。在软件中，它是类型安全之线，应当从领域模型经由 API 一直不间断地延伸到前端。

## 许可证

在以下两者中任选其一进行双重许可：

- **Apache License 2.0** ([LICENSE-APACHE](LICENSE-APACHE))
- **MIT License** ([LICENSE-MIT](LICENSE-MIT))

大多数用户为求简便而选择 MIT。Apache 2.0 则提供额外的专利保护。

## 商标

**Blackwell Systems™** 和 **Blackwell Systems 徽标**是 Dayna Blackwell 的商标。你可以使用“Blackwell Systems”这一名称来指代本项目，但未经事先书面许可，你不得以暗示背书或官方从属关系的方式使用该名称或徽标。使用指南见 [BRAND.md](BRAND.md)。
