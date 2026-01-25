# MVP Roadmap

This document outlines the path to v0.1 - a working schema compiler with Zod generation.

## MVP Scope (v0.1)

**Goal**: Prove the concept with minimal but complete functionality.

**Features**:
- Parse Go structs with `gt:` tags
- Generate Zod schemas
- Generate TypeScript types
- Basic CLI (`goldenthread generate`)

**Non-Goals** (defer to later versions):
- OpenAPI generation
- DSL syntax
- Custom validators
- CI helpers (`check` command)

## Implementation Phases

### Phase 1: Core Schema IR ✅

**Status**: Complete

Files created:
- `internal/schema/schema.go` - IR types and validation

**What it defines**:
- `Schema` - Represents a complete domain model
- `Field` - Individual struct field
- `FieldType` - Type system representation
- `FieldRules` - Validation rules

### Phase 2: Tag Parser 🚧

**Status**: In progress

Files to implement:
- `internal/parser/parser.go` - AST parsing ✅ (skeleton)
- `internal/parser/parser_test.go` - Parser tests
- `internal/parser/tags.go` - Tag parsing logic
- `internal/parser/tags_test.go` - Tag parsing tests

**Tasks**:
1. Implement `parseTags()` - Extract tag map from string
2. Implement `parseRules()` - Convert tag values to FieldRules
3. Implement `extractType()` - Map Go types to TypeKind
4. Add comprehensive tests for edge cases

**Test cases needed**:
```go
// Basic types
type Simple struct {
    Name string `gt:"required"`
}

// Validation rules
type Validated struct {
    Username string `gt:"required,len:3..20"`
    Email    string `gt:"email"`
    Age      int    `gt:"min:13,max:130"`
}

// Nested structs
type Nested struct {
    Profile Profile `gt:"required"`
}

// Arrays
type WithArray struct {
    Tags []string `gt:"minitems:1,maxitems:10"`
}
```

### Phase 3: Zod Emitter 🚧

**Status**: In progress

Files to implement:
- `internal/emitter/zod/emitter.go` - Zod generation ✅ (skeleton)
- `internal/emitter/zod/emitter_test.go` - Emitter tests

**Tasks**:
1. Implement `Emit()` - Main generation function
2. Implement `emitField()` - Field-level generation
3. Implement `emitType()` - Type translation
4. Implement rule methods (`emitStringRules()`, etc.)
5. Add tests for all type/rule combinations

**Test cases needed**:
```go
// String with rules
z.string().min(3).max(20)

// Email format
z.string().email()

// Number with range
z.number().min(0).max(100)

// Optional field
z.string().optional()

// Array
z.array(z.string()).min(1).max(10)
```

### Phase 4: TypeScript Emitter 📋

**Status**: Not started

Files to create:
- `internal/emitter/typescript/emitter.go`
- `internal/emitter/typescript/emitter_test.go`

**Tasks**:
1. Implement interface generation
2. Map Go types to TypeScript types
3. Handle optional fields (`?` syntax)
4. Add JSDoc comments

**Example output**:
```typescript
/**
 * User represents a user account
 */
export interface User {
  /** Unique identifier */
  id: string
  username: string
  email?: string
}
```

### Phase 5: CLI Tool 📋

**Status**: Skeleton created

Files to implement:
- `cmd/goldenthread/main.go` ✅ (skeleton)
- `cmd/goldenthread/generate.go` - Generate command
- `cmd/goldenthread/config.go` - Configuration loading

**Tasks**:
1. Implement `generate` command
   - Parse flags (`--output`, `--target`)
   - Discover Go files
   - Run parser
   - Run emitters
   - Write output files
2. Implement configuration loading
3. Add error handling and user feedback

**CLI UX goals**:
- Clear progress output
- Helpful error messages
- Fast execution (<200ms for small projects)

### Phase 6: Integration Testing 📋

**Status**: Not started

Files to create:
- `integration_test.go`

**Test flow**:
1. Create temp directory
2. Write Go file with test structs
3. Run `goldenthread generate`
4. Verify Zod file exists
5. Verify TypeScript file exists
6. Compile TypeScript with `tsc`
7. Run Zod validations
8. Clean up

**Test cases**:
- Basic struct with primitives
- Struct with validation rules
- Nested structs
- Arrays and slices
- All supported formats

## Success Criteria for v0.1

Before releasing v0.1, must achieve:

- [ ] Parse 10+ test structs correctly
- [ ] Generate valid Zod schemas
- [ ] Generate valid TypeScript interfaces
- [ ] CLI tool works end-to-end
- [ ] Generated TypeScript compiles without errors
- [ ] Generated Zod validates test data correctly
- [ ] Documentation is complete
- [ ] README has working examples
- [ ] 80%+ test coverage

## Post-MVP (v0.2+)

Deferred to v0.2:
- OpenAPI emitter
- `check` command for CI
- DSL syntax
- Custom validators
- goldenthread.yaml configuration

Deferred to v0.3:
- Cross-field validation
- Schema composition
- Conditional validation
- Client-side extension hooks

## Development Workflow

### Quick iteration cycle

```bash
# 1. Write parser code
vim internal/parser/parser.go

# 2. Test immediately
go test ./internal/parser -v

# 3. Try on example
go run cmd/goldenthread/main.go generate examples/basic

# 4. Check generated output
cat examples/basic/gen/user.schema.ts
```

### Key development principles

1. **Test-driven**: Write tests before implementation
2. **Small commits**: Each phase should be committable
3. **Documentation-first**: README examples before implementation
4. **Real usage**: Test on actual Go structs, not just toys

## Timeline Estimate

**Phase 2** (Parser): 4-6 hours
- Tag parsing: 2 hours
- Type extraction: 2 hours
- Tests: 2 hours

**Phase 3** (Zod Emitter): 3-4 hours
- Basic emission: 2 hours
- Rule translation: 1 hour
- Tests: 1 hour

**Phase 4** (TypeScript Emitter): 2-3 hours
- Interface generation: 1 hour
- Type mapping: 1 hour
- Tests: 1 hour

**Phase 5** (CLI): 3-4 hours
- Command implementation: 2 hours
- File discovery: 1 hour
- Error handling: 1 hour

**Phase 6** (Integration): 2-3 hours
- Test harness: 1 hour
- Test cases: 1 hour
- CI setup: 1 hour

**Total**: ~15-20 hours to working MVP

## Next Immediate Steps

1. Complete tag parsing implementation
2. Add parser tests
3. Test parser on examples/basic/models.go
4. Implement Zod emitter
5. Test end-to-end generation

Once parser + Zod emitter work, the rest is straightforward.
