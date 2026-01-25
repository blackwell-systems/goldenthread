# Changelog

All notable changes to goldenthread will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Core schema compiler architecture**
  - Three-stage pipeline: Parse → IR → Emit
  - Language-agnostic intermediate representation (IR)
  - Source position tracking for all schemas and fields

- **Tag specification (v0.1)**
  - Comprehensive tag grammar with validation
  - Presence flags: `required`, `optional`
  - Numeric rules: `min:N`, `max:N`
  - String length: `len:M..N`
  - Pattern matching: `pattern:REGEX`
  - Format validators: `email`, `uuid`, `url`, `date`, `datetime`, `ipv4`, `ipv6`
  - Conflict detection: unknown tokens, multiple formats, type mismatches
  - Unknown token policy: error immediately (compiler-like behavior)

- **Parser implementation**
  - AST-based Go source parsing with `go/parser`
  - Struct tag extraction using `reflect.StructTag` (handles spaces/quotes correctly)
  - Pointer type detection (`*T`)
  - Selector expression handling (`time.Time`, `uuid.UUID`)
  - Named type detection (custom types become `TypeNamed`)
  - Documentation extraction (Doc comments + trailing Comment fallback)
  - Optional semantics with 5-level precedence
  - Type-aware validation (string rules only on strings, numeric rules only on numbers)
  - Comprehensive error messages with file:line positions

- **Zod emitter**
  - TypeScript/Zod schema generation
  - JSDoc comment generation from Go documentation
  - Source origin comments (`Generated from file:line`)
  - Proper enum handling (`z.enum([...])`)
  - Format mappings (email → `.email()`, uuid → `.uuid()`, etc.)
  - Type inference exports (`export type User = z.infer<typeof UserSchema>`)
  - Comment escaping (handles `*/` in docs)

- **CLI tool**
  - `goldenthread generate` command with flag parsing
    - `--out` flag for output directory (default: `./gen`)
    - `--target` flag for generation target (zod only in v0.1)
    - `--recursive` flag for subdirectory processing
    - Progress reporting and success/failure summary
    - Kebab-case filename generation (User → user.ts)
    - Automatic metadata generation for drift detection
  - `goldenthread check` command with drift detection
    - Hash-based schema comparison
    - Detects changed, added, and removed schemas
    - Clear visual output (✓ up to date, ✗ changed, + added, - removed)
    - Exit code 1 if out of sync (CI-friendly)
    - Reads `.goldenthread.json` metadata file

- **Hash-based drift detection**
  - Deterministic SHA-256 hashing of schema content
  - Metadata file (`.goldenthread.json`) tracking
  - Version tracking in metadata
  - Excludes documentation and positions for stable hashes
  - Sorted fields and enum values for deterministic output

- **go/packages integration** (proper type resolution)
  - Uses `golang.org/x/tools/go/packages` for package loading
  - Full type information via `go/types`
  - Real import path resolution (`time.Time` → `"time"`, `uuid.UUID` → `"github.com/google/uuid"`)
  - Handles modules, build tags, generated code, type aliases
  - Full package paths in registry for embedded struct resolution
  - Pattern support: `./models` or `./...` for recursive

- **Embedded struct flattening**
  - Automatic field promotion from embedded structs
  - Two-pass algorithm: register schemas, then flatten
  - Recursive flattening with cycle detection
  - Collision detection for GoName and JSONName conflicts
  - Preserves field documentation, validation rules, tags
  - Cross-package support when types in registry
  - Example: `type User struct { Base; Username string }` promotes Base fields into User schema

- **First-class collision detection** (must-have for v0.1)
  - JSON name collision detection across all fields (including embedded)
  - Go field name collision detection during parsing
  - Runs validation before and after embedded struct flattening
  - Clear error messages showing both conflicting fields with positions
  - Prevents silently generating invalid APIs with duplicate keys
  - Example: `UserID` and `UserId` both mapping to `"userId"` → compile error

- **Documentation**
  - Comprehensive tag specification (docs/TAG_SPEC.md)
  - Parsing table with IR mappings
  - Zod output mappings
  - Valid and invalid examples
  - Architecture documentation
  - Design philosophy and decision rationale

- **Dual licensing**
  - Apache-2.0 OR MIT
  - SPDX identifiers in all source files
  - Clear license selection guidance

### Changed

- **IR refactoring**
  - Split `Type` from `TypeRef` for proper named type resolution
  - Renamed `SourceLocation` to `SourcePos` for brevity
  - Changed `Field.Name` to `Field.GoName` for clarity
  - Replaced `Field.Required` with `Field.Optional` (clearer semantics)
  - Renamed `TypeRef.PackagePath` to `PackageQualifier` (AST-ready, go/packages-ready)
  - Added `TypeNamed` kind for named type references
  - Added `FieldRules.Enum` for oneof support

- **Optional semantics precedence** (explicit 5-level hierarchy):
  1. `required` tag → `Optional = false`
  2. `optional` tag → `Optional = true`
  3. Pointer type → `Optional = true`
  4. `omitempty` → `Optional = true`
  5. Default → `Optional = false` (required by default)

### Fixed

- **Critical parser bugs**
  - String conversion using `string(rune(n))` producing control characters instead of numbers
  - Tag parsing breaking on quoted values and spaces (now uses `reflect.StructTag`)
  - Optional semantics backwards (everything optional by default)
  - Field positions never captured from AST
  - Silent field drops on parse errors (now propagates errors)
  - Unknown identifiers becoming `TypeAny` instead of `TypeNamed`

- **Validation improvements**
  - Parse failures for min/max/len now error instead of silent ignore
  - `len:M..N` validates both sides parse and M ≤ N
  - Format count increment syntax (assignment not `++`)
  - Conflict detection for `required+optional` (now errors)
  - All type mismatches report clear error messages

- **Zod emitter corrections**
  - Enum generation using correct top-level constructor `z.enum([...])`
  - Date format using regex pattern not non-existent `.date()` method

### Internal

- Custom `intToString` helper to avoid stdlib dependency
- Normalized documentation strings with `normalizeDoc()` helper
- JSDoc writing helper with proper escaping (`writeJSDoc()`)
- Kebab-case filename conversion (`toKebabCase()`)
- Error propagation from parser through to CLI

## [0.1.0] - TBD

Initial release (in development)

[Unreleased]: https://github.com/blackwell-systems/goldenthread/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/blackwell-systems/goldenthread/releases/tag/v0.1.0
