# Architecture

goldenthread is a **schema compiler** that transforms Go domain models into cross-platform validation and type definitions.

## Core Principle

```
Go structs → AST Analysis → Schema IR → Code Generation → Target Languages
```

goldenthread is **build-time only**. There is no runtime library, no reflection overhead, no magic. It's pure code generation, like `sqlc`, `gqlgen`, or `protoc`.

## System Components

### 1. Parser (`internal/parser`)

**Responsibility**: Extract schema definitions from Go source code.

**Input**: Go files containing structs with `gt:` tags
**Output**: Schema IR (intermediate representation)

**Implementation**:
- Uses `go/parser` and `go/ast` to parse Go source
- Walks AST to find struct definitions
- Extracts field types, tags, and documentation
- Resolves type references (including embedded structs)

**Example**:

```go
// Input Go code
type User struct {
    Username string `gt:"required,len:3..20"`
    Email    string `gt:"email"`
}

// Parser extracts
Schema{
    Name: "User",
    Fields: []Field{
        {Name: "Username", Type: "string", Rules: {Required: true, MinLen: 3, MaxLen: 20}},
        {Name: "Email", Type: "string", Rules: {Format: "email"}},
    },
}
```

### 2. Schema IR (`internal/schema`)

**Responsibility**: Language-agnostic intermediate representation.

The Schema IR is the core data structure that decouples parsing from code generation. It represents validation rules in a way that can be emitted to any target language.

**Key Types**:

```go
type Schema struct {
    Name        string
    Package     string
    Fields      []Field
    Rules       []Rule
    Documentation string
}

type Field struct {
    Name          string
    Type          FieldType
    Required      bool
    Rules         FieldRules
    Documentation string
    Tags          map[string]string // Original Go tags
}

type FieldType struct {
    Kind       TypeKind // string, int, float, bool, array, object
    Element    *FieldType // For arrays: element type
    Properties []Field // For objects: nested fields
}

type FieldRules struct {
    // String rules
    MinLength *int
    MaxLength *int
    Pattern   *string
    Format    *string // email, uuid, url, etc.
    
    // Numeric rules
    Min *float64
    Max *float64
    
    // Array rules
    MinItems *int
    MaxItems *int
    UniqueItems bool
}
```

**Why a separate IR?**

1. **Decouples concerns**: Parser doesn't know about Zod, emitters don't know about Go
2. **Extensibility**: New emitters only need to understand IR, not Go AST
3. **Validation**: IR can validate semantic correctness before codegen
4. **Testing**: Can test parser and emitters independently

### 3. Emitters (`internal/emitter/*`)

**Responsibility**: Generate target language code from Schema IR.

Each emitter is independent and generates code for one target language/framework.

#### Zod Emitter (`internal/emitter/zod`)

Generates Zod validation schemas.

**Input**: `Schema` IR
**Output**: TypeScript file with Zod schemas

**Example**:

```typescript
// Generated from User schema
import { z } from 'zod'

export const UserSchema = z.object({
  username: z.string().min(3).max(20),
  email: z.string().email(),
})

export type User = z.infer<typeof UserSchema>
```

**Design Decisions**:
- Always export both schema and inferred type
- Use const assertions for better type inference
- Include source comments showing origin

#### TypeScript Emitter (`internal/emitter/typescript`)

Generates TypeScript interface definitions (no validation).

**Why separate from Zod?** Some codebases use TypeScript without Zod. This emitter provides plain types.

**Example**:

```typescript
// Generated from User schema
export interface User {
  username: string
  email: string
}
```

#### OpenAPI Emitter (`internal/emitter/openapi`)

Generates OpenAPI 3.0 specification YAML/JSON.

**Example**:

```yaml
# Generated from User schema
components:
  schemas:
    User:
      type: object
      required:
        - username
      properties:
        username:
          type: string
          minLength: 3
          maxLength: 20
        email:
          type: string
          format: email
```

### 4. CLI Tool (`cmd/goldenthread`)

**Responsibility**: User-facing command-line interface.

**Commands**:

```bash
# Generate all schemas
goldenthread generate ./models

# Generate specific targets
goldenthread generate --target=zod ./models
goldenthread generate --target=openapi ./models

# Verify schemas are up-to-date (CI)
goldenthread check ./models

# Initialize goldenthread.yaml config
goldenthread init
```

**Configuration** (`goldenthread.yaml`):

```yaml
version: 1

# Source directories to scan
sources:
  - ./models
  - ./api

# Output configuration
output:
  dir: ./gen
  targets:
    - zod
    - typescript
    - openapi

# Options
options:
  zod:
    import_path: "zod"
    export_types: true
  
  openapi:
    version: "3.0.0"
    info:
      title: "API"
      version: "1.0.0"
```

## Data Flow

### Generation Flow

```
1. CLI invoked
   ↓
2. Load configuration (goldenthread.yaml)
   ↓
3. Discover Go files in source directories
   ↓
4. For each file:
   a. Parse Go AST
   b. Find structs with gt: tags
   c. Extract schema into IR
   ↓
5. Validate schemas (check for conflicts, invalid rules)
   ↓
6. For each target (zod, typescript, openapi):
   a. Pass Schema IR to emitter
   b. Generate code
   c. Write to output directory
   ↓
7. Write metadata (hashes for sync checking)
```

### Check Flow (CI)

```
1. CLI invoked with 'check'
   ↓
2. Parse Go files → Schema IR
   ↓
3. Hash Schema IR
   ↓
4. Read metadata from previous generation
   ↓
5. Compare hashes
   ↓
6. If mismatch: exit 1 (fail CI)
   If match: exit 0 (pass)
```

## Design Principles

### 1. No Runtime Dependencies

Generated code should have **zero** dependency on goldenthread. Users only depend on:
- `zod` (if using Zod emitter)
- Standard TypeScript (if using TypeScript emitter)
- Nothing (if using OpenAPI emitter)

### 2. Readable Generated Code

Generated code should look like it was written by a human. Include:
- Comments showing origin
- Formatted properly
- Clear naming
- No magic

Example:

```typescript
// Generated by goldenthread from models/user.go:15
// Source: type User struct { Username string `gt:"required,len:3..20"` }
export const UserSchema = z.object({
  username: z.string().min(3).max(20),
})
```

### 3. Gradual Adoption

Should work alongside existing validation libraries:

```go
type User struct {
    // validator.v10 tags still work
    Username string `validate:"required,min=3,max=20" gt:"export"`
    // gt can read validator tags
    Email string `validate:"email" gt:"export"`
}
```

The `gt:"export"` tag signals "generate from validator tags".

### 4. Explicit, Not Magic

Avoid surprises. Generated code location is explicit, regeneration is explicit, no hidden conventions.

## Testing Strategy

### Parser Tests

```go
func TestParseUserStruct(t *testing.T) {
    src := `
    package models
    type User struct {
        Username string ` + "`gt:\"required,len:3..20\"`" + `
    }
    `
    schema, err := parser.Parse(src)
    assert.NoError(t, err)
    assert.Equal(t, "User", schema.Name)
    assert.Len(t, schema.Fields, 1)
    assert.True(t, schema.Fields[0].Required)
}
```

### Emitter Tests

```go
func TestZodEmitter(t *testing.T) {
    schema := &schema.Schema{
        Name: "User",
        Fields: []schema.Field{
            {Name: "Username", Type: schema.TypeString, Required: true},
        },
    }
    
    code, err := zod.Emit(schema)
    assert.NoError(t, err)
    assert.Contains(t, code, "z.string()")
}
```

### Integration Tests

```go
func TestEndToEnd(t *testing.T) {
    // Write Go file
    // Run goldenthread generate
    // Verify Zod file exists
    // Verify TypeScript compiles
    // Verify OpenAPI is valid
}
```

## Future Extensions

### Phase 2: DSL Support

```go
var UserSchema = gt.Object("User",
    gt.Field("username", gt.String().Min(3).Max(20)),
)
```

This requires:
- DSL parser (separate from struct tag parser)
- Registration mechanism
- Same Schema IR output

### Phase 3: Custom Validators

```go
type User struct {
    Email string `gt:"email,custom:corporate_email"`
}

func init() {
    gt.RegisterValidator("corporate_email", func(v string) bool {
        return strings.HasSuffix(v, "@corp.com")
    })
}
```

This generates:

```typescript
export const UserSchema = z.object({
  email: z.string().email().refine(
    v => v.endsWith('@corp.com'),
    { message: 'Must be corporate email' }
  ),
})
```

### Phase 4: Bidirectional Sync

Support marking schemas as "client-extendable":

```go
type User struct {
    Password string `gt:"min:8,client_extend"`
}
```

Generates a hook point:

```typescript
export const UserSchema = z.object({
  password: z.string().min(8),
}).refine(/* YOUR CUSTOM VALIDATION HERE */)
```

## Performance Considerations

### Build Time

For 1000 structs:
- Parse: <1s (Go's parser is fast)
- Emit: <1s (string generation)
- Write: <100ms (disk I/O)

**Total: ~2 seconds for 1000 schemas**

This is acceptable for build-time tooling.

### Memory

Schema IR is lightweight. Each schema ~1KB in memory.

For 1000 schemas: ~1MB total.

No memory concerns.

## Comparison with Alternatives

### vs Runtime Reflection

**goldenthread** (build-time):
- ✅ Zero runtime overhead
- ✅ Type-safe generated code
- ✅ Works with static TypeScript

**Reflection-based** (runtime):
- ❌ Runtime performance cost
- ❌ Cannot generate TypeScript at build time
- ❌ Requires Go server running

### vs OpenAPI-First

**goldenthread** (code-first):
- ✅ Go structs are source of truth
- ✅ No separate schema files
- ✅ Type-safe in Go

**OpenAPI-first** (schema-first):
- ❌ YAML/JSON as source of truth
- ❌ Go code generated from schemas
- ❌ Lose Go's type system benefits

### vs Manual Sync

**goldenthread** (automated):
- ✅ Single command regenerates everything
- ✅ CI detects drift
- ✅ Impossible to forget

**Manual** (error-prone):
- ❌ Must update 3+ places
- ❌ Easy to forget
- ❌ No drift detection

## Summary

goldenthread is designed as a **compiler, not a library**. It transforms Go domain models into cross-platform schemas using a clean three-stage pipeline: parse → IR → emit. This architecture makes it extensible, testable, and maintainable while keeping generated code clean and dependency-free.
