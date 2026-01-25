# goldenthread

> The golden thread of truth that runs through your system.

[![Blackwell Systems™](https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg)](https://github.com/blackwell-systems) 
[![Go Reference](https://pkg.go.dev/badge/github.com/blackwell-systems/goldenthread.svg)](https://pkg.go.dev/github.com/blackwell-systems/goldenthread) 
[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://go.dev/) 

[![License](https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue.svg)](LICENSE-APACHE) 
[![Sponsor](https://img.shields.io/badge/Sponsor-Buy%20Me%20a%20Coffee-yellow?logo=buy-me-a-coffee&logoColor=white)](https://buymeacoffee.com/blackwellsystems)

**goldenthread** is a schema compiler for Go that maintains a single source of truth across your entire stack. Define your domain models once in Go, and compile them into validation, types, and APIs everywhere else.

## The Problem

Modern full-stack development breaks the "golden thread" of type safety at API boundaries:

```
Go Backend              API Boundary           TypeScript Frontend
───────────            ─────────────           ───────────────────
struct User {          ❌ Manual sync          interface User {
  Username string        required                username: string
  Email string         ❌ Validation             email: string
  Age int                duplicated            }
}                      ❌ Documentation        
                         out of date           const schema = z.object({
                                                 username: z.string()
                                                 email: z.string()
                                               })
```

Every change requires updating:
1. Go struct
2. OpenAPI documentation  
3. TypeScript types
4. Zod schemas
5. Client-side validation

Miss one? Runtime errors, integration bugs, documentation drift.

## The Solution

goldenthread maintains the golden thread:

```
Go Backend                                     TypeScript Frontend
───────────                                    ───────────────────
type User struct {                             // Generated automatically
  Username string `gt:"required,len:3..20"`    export const UserSchema = z.object({
  Email    string `gt:"email"`                   username: z.string().min(3).max(20),
  Age      int    `gt:"min:13,max:130"`         email: z.string().email(),
}                                                 age: z.number().min(13).max(130),
                                               })
                                               
                                               export type User = z.infer<typeof UserSchema>
```

One change in Go → everything regenerates → impossible to drift.

## Features

- **Single Source of Truth**: Go structs are canonical
- **Automatic Generation**: Zod schemas, TypeScript types, OpenAPI specs
- **Type-Safe**: Compile-time guarantees across the stack
- **CI Integration**: Tests fail if schemas are out of sync
- **Go Idiomatic**: Feels like `sqlc`, `gqlgen`, `protoc`
- **Zero Runtime**: Pure code generation, no reflection overhead

## Quick Start

### Installation

```bash
go install github.com/blackwell-systems/goldenthread/cmd/goldenthread@latest
```

### Basic Usage

1. **Define your domain models in Go:**

```go
// models/user.go
package models

type User struct {
    ID       string `json:"id" gt:"uuid,required"`
    Username string `json:"username" gt:"required,len:3..20"`
    Email    string `json:"email" gt:"email"`
    Age      int    `json:"age" gt:"min:13,max:130"`
}
```

2. **Generate schemas:**

```bash
goldenthread generate ./models
```

3. **Use generated code in TypeScript:**

```typescript
import { UserSchema, User } from './gen/user'

// Validation
const result = UserSchema.safeParse(data)

// Type inference
const user: User = {
  id: '123e4567-e89b-12d3-a456-426614174000',
  username: 'alice',
  email: 'alice@example.com',
  age: 25
}
```

## Syntax Design

goldenthread supports two modes: **tag-based** (simple) and **DSL** (advanced).

### Tag-Based (Level 1)

Familiar struct tags for 80% of use cases:

```go
type Product struct {
    Name  string  `gt:"required,len:1..100"`
    Price float64 `gt:"required,min:0"`
    SKU   string  `gt:"pattern:^[A-Z0-9]{8}$"`
}
```

### DSL (Level 2)

Explicit schema definition for complex validation:

```go
var PasswordResetSchema = gt.Object("PasswordReset",
    gt.Field("password", gt.String().Min(8)),
    gt.Field("confirm", gt.String()),
).Refine(func(v gt.Value) bool {
    return v["password"] == v["confirm"]
}, "Passwords must match")
```

### Hybrid (Level 3)

Combine both for maximum flexibility:

```go
type User struct {
    Username string `gt:"required,len:3..20"`
    Email    string `gt:"email"`
}

var UserRules = gt.For[User]().
    Refine(func(u User) bool {
        return !strings.HasSuffix(u.Email, "@tempmail.com")
    }, "Disposable emails not allowed")
```

## Architecture

goldenthread is a **schema compiler**, not a runtime library:

```
Go Source Code → AST Parser → Schema IR → Emitters → Generated Code
                                               ├── Zod
                                               ├── TypeScript
                                               └── OpenAPI
```

### Internal Components

- **Parser** (`internal/parser`): Extracts schema definitions from Go AST
- **Schema IR** (`internal/schema`): Language-agnostic intermediate representation
- **Emitters** (`internal/emitter/*`): Generate target language code
  - `zod`: Generates Zod validation schemas
  - `typescript`: Generates TypeScript types
  - `openapi`: Generates OpenAPI 3.0 specifications

## CI Integration

Ensure schemas never drift:

```go
func TestSchemasInSync(t *testing.T) {
    gt.AssertGenerated(t, "./gen")
}
```

This test:
- Hashes your Go source
- Hashes generated files
- Fails if out of sync

Run in CI to make schema drift impossible.

## Comparison with Existing Tools

| Tool | Validation | OpenAPI | Zod/TS Export | Approach |
|------|-----------|---------|---------------|----------|
| `validator.v10` | ✅ | ❌ | ❌ | Runtime validation |
| `swaggo/swag` | ❌ | ✅ | ❌ | Comment annotations |
| `oapi-codegen` | ❌ | ✅ | ❌ | OpenAPI → Go |
| **goldenthread** | ✅ | ✅ | ✅ | Go → Everything |

goldenthread is the missing piece: **Go structs as the source of truth**.

## Why "goldenthread"?

In traditional weaving, the golden thread is the single continuous strand that holds the fabric together. In software, it's the thread of type safety and validation that should run unbroken from your domain models through your APIs to your frontend.

Most systems today have a *broken thread*—manual synchronization between backend and frontend that inevitably drifts. goldenthread restores that continuous connection.

## License

Dual-licensed under your choice of:

- **Apache License 2.0** ([LICENSE-APACHE](LICENSE-APACHE))
- **MIT License** ([LICENSE-MIT](LICENSE-MIT))

This means you can choose either license for your use case. Most users prefer MIT for simplicity, while Apache 2.0 provides additional patent protections.

---

**goldenthread** - One source of truth, from backend to frontend.
# goldenthread
