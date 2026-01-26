# goldenthread

> Schema compiler: Go structs → TypeScript/Zod.  Keep backend/frontend validation in sync automatically.

[![Blackwell Systems™](https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg)](https://github.com/blackwell-systems) 
[![Go Reference](https://pkg.go.dev/badge/github.com/blackwell-systems/goldenthread.svg)](https://pkg.go.dev/github.com/blackwell-systems/goldenthread) 
[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://go.dev/) 
[![CI](https://github.com/blackwell-systems/goldenthread/workflows/CI/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/ci.yml)
[![Lint](https://github.com/blackwell-systems/goldenthread/workflows/Lint/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/lint.yml)

[![License](https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue.svg)](LICENSE-APACHE) 
[![Sponsor](https://img.shields.io/badge/Sponsor-Buy%20Me%20a%20Coffee-yellow?logo=buy-me-a-coffee&logoColor=white)](https://buymeacoffee.com/blackwellsystems)

**goldenthread** generates type-safe validation schemas from Go struct definitions. Define your domain models once in Go with validation tags, and compile them into Zod schemas for TypeScript—automatically.

## Overview

Maintaining validation across Go backends and TypeScript frontends means keeping multiple representations synchronized. goldenthread generates Zod schemas directly from Go struct tags:

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

Changes to Go structs regenerate TypeScript schemas automatically. The compiler ensures they stay synchronized.

## Features

### Core Capabilities

- **Go structs as source of truth**: Define validation rules once in struct tags
- **Zod schema generation**: Automatic TypeScript/Zod output with proper types
- **Comprehensive validation**: String length, numeric bounds, formats (email, UUID, URL), patterns, enums
- **Type safety**: Go types map correctly to Zod types (primitives, arrays, maps, nested objects)
- **Embedded struct flattening**: Anonymous fields automatically promoted
- **Collision detection**: Prevents duplicate JSON field names at compile time
- **Drift detection**: `check` command verifies schemas stay in sync
- **Zero runtime overhead**: Pure code generation, no reflection

### Validation Rules

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

See [docs/TAG_SPEC.md](docs/TAG_SPEC.md) for complete tag syntax.

## Installation

```bash
go install github.com/blackwell-systems/goldenthread/cmd/goldenthread@latest
```

## Quick Start

### 1. Define Your Models

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

### 2. Generate Schemas

```bash
# Generate from current directory
goldenthread generate ./models

# Generate recursively
goldenthread generate ./models --recursive

# Specify output directory
goldenthread generate ./models --out ./frontend/src/schemas
```

Output:
```
gen/
  user.ts                    # Generated Zod schema
  .goldenthread.json         # Metadata for drift detection
```

### 3. Use in TypeScript

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

### 4. Verify Schemas Stay in Sync

```bash
# Check if generated schemas match source
goldenthread check ./models

# Returns exit code 1 if out of sync (CI-friendly)
```

Add to CI:
```yaml
- name: Check schema drift
  run: goldenthread check ./models
```

## Architecture

goldenthread is a schema compiler with three stages:

```
Go Source → Parser → Intermediate Representation → Emitter → Generated Code
            (AST)         (Language-agnostic)        (Zod)    (TypeScript)
```

### Pipeline Components

1. **Parser** (`internal/parser`)
   - Uses `go/packages` and `go/types` for proper type resolution
   - Extracts struct definitions with `gt:` tags
   - Validates tag syntax and conflicts
   - Handles embedded structs, cross-package references

2. **Intermediate Representation** (`internal/schema`)
   - Language-agnostic schema format
   - Separates parsing from code generation
   - Enables future emitters (OpenAPI, JSON Schema, etc.)

3. **Normalization** (`internal/normalize`)
   - Flattens embedded struct fields
   - Detects Go field name collisions
   - Validates JSON name collisions
   - Ensures schema correctness

4. **Emitter** (`internal/emitter/zod`)
   - Generates TypeScript with Zod schemas
   - Preserves Go documentation as JSDoc
   - Emits type-safe `z.infer<>` types
   - Deterministic output for stable diffs

5. **Hash/Drift Detection** (`internal/hash`)
   - SHA-256 hashing of schema content
   - Metadata tracking (`.goldenthread.json`)
   - Detects when source and generated diverge

## Supported Types

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

See [docs/FEATURES.md](docs/FEATURES.md) for complete type coverage.

## Validation Tags

### Presence

- `required` - Field must be present (default for non-pointers)
- `optional` - Field can be omitted

### String Rules

- `len:M..N` - Length between M and N characters
- `min:N` - Minimum length (shorthand for `len:N..`)
- `max:N` - Maximum length (shorthand for `len:..N`)
- `pattern:REGEX` - Must match regular expression

### Numeric Rules

- `min:N` - Minimum value
- `max:N` - Maximum value

### Format Validators

- `email` - Valid email address
- `uuid` - Valid UUID string
- `url` - Valid URL
- `date` - ISO 8601 date (YYYY-MM-DD)
- `datetime` - ISO 8601 datetime
- `ipv4` - IPv4 address
- `ipv6` - IPv6 address

### Enums

- `enum:value1,value2,value3` - One of specified values

### Array Rules

- `min:N` - Minimum array length
- `max:N` - Maximum array length

See [docs/TAG_SPEC.md](docs/TAG_SPEC.md) for detailed specification.

## Examples

### Nested Objects

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

Generates nested Zod schema:
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

### Embedded Structs

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

Generates flattened schema:
```typescript
const ArticleSchema = z.object({
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
  title: z.string().min(1).max(200),
  content: z.string()
})
```

### Enums

```go
type Task struct {
  Title    string `json:"title" gt:"required"`
  Status   string `json:"status" gt:"enum:pending,in_progress,completed,cancelled"`
  Priority string `json:"priority" gt:"enum:low,medium,high"`
}
```

Generates enum schemas:
```typescript
const TaskSchema = z.object({
  title: z.string(),
  status: z.enum(['pending', 'in_progress', 'completed', 'cancelled']),
  priority: z.enum(['low', 'medium', 'high'])
})
```

### Maps

```go
type Config struct {
  Name     string            `json:"name" gt:"required"`
  Settings map[string]string `json:"settings" gt:"required"`
  Metadata map[string]any    `json:"metadata" gt:"optional"`
}
```

Generates record schemas:
```typescript
const ConfigSchema = z.object({
  name: z.string(),
  settings: z.record(z.string(), z.string()),
  metadata: z.record(z.string(), z.any()).optional()
})
```

## CLI Commands

### `generate`

Generate Zod schemas from Go source:

```bash
goldenthread generate [flags] <directory>

Flags:
  --out <dir>       Output directory (default: ./gen)
  --recursive       Process subdirectories recursively
  --target <name>   Target format (default: zod)
```

### `check`

Verify generated schemas match source:

```bash
goldenthread check [flags] <directory>

Flags:
  --metadata <file>  Metadata file to check against
  --recursive        Process subdirectories recursively
```

Exit codes:
- `0` - Schemas in sync
- `1` - Schemas out of sync or error

## Current Limitations

goldenthread v0.1 focuses on the core use case: struct validation for APIs and forms. The following are **not yet supported**:

### Type System

- Go union types (interfaces with type assertions)
- Discriminated unions
- Fixed-length arrays (`[3]int`)
- Literal constant values
- Recursive/self-referential types

### Validation

- Custom validation functions
- Cross-field validation (password confirmation, etc.)
- Unique items in arrays
- Conditional validation

### Emitters

- OpenAPI/Swagger generation
- JSON Schema generation
- TypeScript types (separate from Zod)
- Other validation libraries

### Tag Features

- Tag inheritance/composition
- Conditional rules
- Multiple validation sets per struct

These limitations are intentional for v0.1. The tool does **one thing well**: generate Zod schemas from struct tags. Future versions may expand scope based on real-world usage.

## Comparison with Alternatives

| Tool | Go → TS | Validation | Approach | Use Case |
|------|---------|-----------|----------|----------|
| `validator.v10` | No | Yes | Runtime tags | Go-only validation |
| `swaggo/swag` | No | No | Comments | OpenAPI from Go |
| `oapi-codegen` | Yes | Yes | OpenAPI → Go | Contract-first APIs |
| `protobuf` | Yes | Limited | `.proto` files | Cross-language RPC |
| **goldenthread** | Yes | Yes | Go → Zod | Go-first validation |

goldenthread is optimized for teams that:
- Write backend in Go
- Write frontend in TypeScript
- Want Go structs as the source of truth
- Need runtime validation in both languages
- Prefer type safety over flexibility

## Documentation

- [Tag Specification](docs/TAG_SPEC.md) - Complete tag syntax reference
- [Feature Matrix](docs/FEATURES.md) - Detailed capability breakdown
- [Architecture](docs/ARCHITECTURE.md) - System design and implementation details
- [Testing Strategy](docs/TESTING.md) - Test suite and continuous fuzzing guide
- [Fuzzing Bug Log](docs/FUZZING_BUGS.md) - Bugs discovered by continuous fuzzing
- [Brand Guidelines](BRAND.md) - Trademark and logo usage
- [Changelog](CHANGELOG.md) - Version history and changes

## Development

### Project Structure

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

### Building

```bash
# Build CLI tool
go build ./cmd/goldenthread

# Run tests
go test ./...

# Run with coverage
go test ./... -cover
```

### Contributing

Contributions welcome! Areas of interest:

1. Additional emitters (OpenAPI, JSON Schema)
2. More validation rules
3. Test coverage improvements
4. Documentation and examples

## Philosophy

goldenthread follows these principles:

1. **Go structs are canonical** - Not config files, not DSLs, not annotations
2. **Build-time generation** - Not runtime reflection or proxies
3. **Type safety everywhere** - Catch errors at compile time
4. **Simple over clever** - Predictable behavior beats flexibility
5. **Fail fast** - Unknown tokens error immediately

The name "goldenthread" comes from traditional weaving: the single continuous strand that holds fabric together. In software, it's the thread of type safety that should run unbroken from domain models through APIs to frontends.

## License

Dual-licensed under your choice of:

- **Apache License 2.0** ([LICENSE-APACHE](LICENSE-APACHE))
- **MIT License** ([LICENSE-MIT](LICENSE-MIT))

Most users prefer MIT for simplicity. Apache 2.0 provides additional patent protections.

## Trademarks

**Blackwell Systems™** and the **Blackwell Systems logo** are trademarks of Dayna Blackwell. You may use the name "Blackwell Systems" to refer to this project, but you may not use the name or logo in a way that suggests endorsement or official affiliation without prior written permission. See [BRAND.md](BRAND.md) for usage guidelines.
