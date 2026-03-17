# goldenthread Feature Matrix

This document tracks implemented features and support coverage for goldenthread v0.1.

## Overview

goldenthread generates TypeScript/Zod validation schemas from Go structs with validation tags. The focus is on **struct validation for APIs and forms** - the most common TypeScript integration use case.

## Core Features

### Parsing & Type Resolution

| Feature | Status | Notes |
|---------|--------|-------|
| Go struct parsing | Complete | Using go/packages + go/types |
| Type inference | Complete | Full type resolution across packages |
| Cross-package references | Complete | Handles imports and module paths |
| Embedded struct flattening | Complete | Recursive with cycle detection |
| Go field collision detection | Complete | Errors on duplicate Go field names |
| JSON name collision detection | Complete | Prevents duplicate JSON keys (critical for API correctness) |
| Pointer unwrapping | Complete | `*string` → optional string |
| Documentation extraction | Complete | From Go comments to JSDoc |
| Position tracking | Complete | File:line for error messages |

### Tag System

| Feature | Status | Syntax | Notes |
|---------|--------|--------|-------|
| Tag parsing | Complete | `gt:"token,key:value"` | Uses reflect.StructTag |
| Required/optional | Complete | `required`, `optional` | 5-level precedence |
| Fallback tags | Complete | `validate:`, `binding:` | Configurable |
| Unknown token detection | Complete | Errors on typos | Fail-fast validation |
| Conflict detection | Complete | Multiple formats, etc. | Clear error messages |
| Tag specification | Complete | docs/TAG_SPEC.md | Canonical contract |

### CLI Tool

| Feature | Status | Command | Notes |
|---------|--------|---------|-------|
| Generate command | Complete | `generate <dir>` | Outputs TypeScript schemas |
| Recursive generation | Complete | `--recursive` flag | Processes subdirectories |
| Output directory | Complete | `--out <dir>` | Custom output location |
| Check command | Complete | `check <dir>` | Drift detection |
| Metadata tracking | Complete | `.goldenthread.json` | Hash-based versioning |
| Exit codes | Complete | 0=success, 1=drift/error | CI-friendly |
| Help text | Complete | `--help` | Usage documentation |

### Code Generation

| Feature | Status | Notes |
|---------|--------|-------|
| Zod emitter | Complete | Target: TypeScript + Zod |
| Type inference export | Complete | `export type T = z.infer<...>` |
| JSDoc comments | Complete | Preserves Go documentation |
| Source attribution | Complete | File:line in generated header |
| Deterministic output | Complete | Sorted for stable diffs |
| Single-file per schema | Complete | `user.ts`, `product.ts`, etc. |

### Schema Normalization

| Feature | Status | Notes |
|---------|--------|-------|
| Schema registry | Complete | Package-aware type lookups |
| Two-pass normalization | Complete | Register → flatten |
| Embedded field promotion | Complete | Flattens anonymous structs |
| Cycle detection | Complete | Prevents infinite recursion |
| Collision validation | Complete | Go names before, JSON names after flattening |

## Type Support

### Go → Zod Type Mapping

| Go Type | Zod Output | Status | Notes |
|---------|-----------|--------|-------|
| `string` | `z.string()` | Complete | With validation rules |
| `int`, `int8`, `int16`, `int32`, `int64` | `z.number()` | Complete | All signed integer types |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `z.number()` | Complete | All unsigned integer types |
| `float32`, `float64` | `z.number()` | Complete | Floating-point numbers |
| `bool` | `z.boolean()` | Complete | Boolean values |
| `time.Time` | `z.string().datetime()` | Complete | ISO 8601 datetime |
| `[]T` | `z.array(T)` | Complete | Arrays with element type |
| `map[string]T` | `z.record(z.string(), T)` | Complete | String-keyed maps |
| `struct { ... }` | `z.object({...})` | Complete | Inline nested objects |
| Named struct | `TypeSchema` | Complete | References to other schemas |
| Embedded struct | Fields flattened | Complete | Automatic field promotion |
| `*T` (pointer) | `T.optional()` | Complete | Pointer = optional semantics |
| `map[K]V` | `z.record(K, V)` | Complete | Maps with any key/value type |
| `interface{}`, `any` | `z.any()` | Complete | Fallback for unknown types |

### Validation Rules

| Rule | Tag Syntax | Zod Output | Status |
|------|-----------|-----------|--------|
| **Presence** | | | |
| Required | `required` | (omit `.optional()`) | Complete |
| Optional | `optional` | `.optional()` | Complete |
| **String** | | | |
| Min length | `len:3..` or `min:3` | `.min(3)` | Complete |
| Max length | `len:..20` or `max:20` | `.max(20)` | Complete |
| Length range | `len:3..20` | `.min(3).max(20)` | Complete |
| Regex pattern | `pattern:^[a-z]+$` | `.regex(/^[a-z]+$/)` | Complete |
| **Formats** | | | |
| Email | `email` | `.email()` | Complete |
| URL | `url` | `.url()` | Complete |
| UUID | `uuid` | `.uuid()` | Complete |
| Date | `date` | `.regex(/^\d{4}-\d{2}-\d{2}$/)` | Complete |
| DateTime | `datetime` | `.datetime()` | Complete |
| IPv4 | `ipv4` | `.ip({ version: 'v4' })` | Complete |
| IPv6 | `ipv6` | `.ip({ version: 'v6' })` | Complete |
| **Numeric** | | | |
| Min value | `min:0` | `.min(0)` | Complete |
| Max value | `max:100` | `.max(100)` | Complete |
| **Array** | | | |
| Min items | `min:1` | `.min(1)` | Complete |
| Max items | `max:10` | `.max(10)` | Complete |
| Unique items | `unique` | - | ⚠️ Parsed but not emitted — has no effect on generated Zod output in v0.1 |
| **Enum** | | | |
| Enum values | `enum:foo,bar,baz` | `z.enum(['foo', 'bar', 'baz'])` | Complete | Values with underscores supported |
| **Custom** | | | |
| Custom validators | `validator:funcName` | - | ⚠️ Parsed, not emitted |

## Drift Detection

| Feature | Status | Notes |
|---------|--------|-------|
| SHA-256 hashing | Complete | Deterministic schema fingerprints |
| Metadata file | Complete | `.goldenthread.json` tracking |
| Version tracking | Complete | goldenthread version in metadata |
| Change detection | Complete | Compares hash of source vs metadata |
| CI integration | Complete | Exit code 1 on drift |
| Change reporting | Complete | Shows which schemas changed |

## Advanced Features

### ✅ Implemented

- **Embedded struct flattening**: Recursively promotes fields from anonymous embedded structs
- **Package-aware registry**: Resolves types across package boundaries using full import paths
- **Cycle detection**: Prevents infinite recursion in self-referential structs
- **First-class collision detection**: Validates both Go field names and JSON keys before generation
  - Catches `UserID` and `UserId` both mapping to `"userId"`
  - Detects embedded field collisions after flattening
  - Critical for API correctness - prevents silently broken schemas
- **5-level optional precedence**: `required` → `optional` → pointer → `omitempty` → default
- **Position tracking**: Every schema/field tracks source file and line number
- **Documentation preservation**: Go comments become JSDoc in generated TypeScript

### ⚠️ Partially Implemented

- **Unique items**: `UniqueItems` bool is set by the parser when `gt:"unique"`
  is used, but the Zod emitter does not emit a uniqueness constraint.
  Using `gt:"unique"` in v0.1 is a no-op in generated output.
- **Custom validators**: Tag parsed (`CustomValidators []string`) but needs function registry

### Planned Emitters (v0.3)

> **Planned for v0.3:** OpenAPI 3.1 and JSON Schema emitters are on the
> roadmap. See [ROADMAP.md](../ROADMAP.md) for details.

## Not Supported (Out of Scope)

These Zod features don't have clean Go struct equivalents:

| Feature | Reason |
|---------|--------|
| `z.union()` | Go doesn't have sum types (would need interface analysis) |
| `z.intersection()` | Could add via multiple embeds, but complex |
| `z.discriminatedUnion()` | Requires tagged unions (interface with type field) |
| `z.literal()` | Needs const value analysis |
| `z.tuple()` | Needs Go array type `[3]int` (not slices) |
| `z.transform()` | Runtime transforms are out of scope |
| `z.refine()` | Custom validation requires runtime code |
| `z.lazy()` | Recursive types need special handling |
| `z.promise()`, `z.function()` | Not applicable to struct validation |
| `z.null()` vs `z.undefined()` | JavaScript-specific distinction |

## Design Principles

1. **Build-time generation** - Not runtime reflection
2. **Declarative validation** - Tags define rules, not code
3. **Portable IR** - Schema representation works across emitters
4. **Struct-focused** - Not arbitrary Go types
5. **Fail-fast parsing** - Unknown tokens error immediately
6. **Type-safe resolution** - Uses go/types for proper type inference

## Future Considerations

### Easy Additions
- Map support: `map[string]T` → `z.record(z.string(), T)`
- Literal values: Const analysis for `z.literal()`
- Null/undefined distinction: Separate pointer from omitempty

### Medium Complexity
- Tuple support: Fixed-length arrays `[3]int` → `z.tuple([...])`
- Intersection: Multiple embeds → `z.intersection()`
- Custom validator registry: Function name → actual validation code

### High Complexity
- Union types: Interface analysis for sum types
- Discriminated unions: Type field detection
- Lazy/recursive: Self-referential type handling
- Transform/refine: Custom validation logic

## Version History

### v0.1.0 (Current)
- ✅ Core parser with go/packages integration
- ✅ Zod emitter with comprehensive validation rules
- ✅ Tag specification with conflict detection
- ✅ CLI tool (generate + check commands)
- ✅ Embedded struct flattening
- ✅ Drift detection with hash-based tracking
- ✅ Full type resolution across packages

---

**Legend:**
- ✅ Complete - Fully implemented and tested
- ⚠️ Partial - Parsed but not fully utilized
- ❌ Not supported - Intentionally out of scope
- 🔜 Planned - On the roadmap
