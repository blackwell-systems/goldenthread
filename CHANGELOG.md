# Changelog

All notable changes to goldenthread will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **Array `min`/`max` constraints now correctly emitted** — `gt:"min:N"` and `gt:"max:N"`
  on array/slice fields were silently ignored in generated Zod output. The parser was
  storing array item constraints in the numeric IR fields (`Rules.Min`/`Rules.Max`)
  instead of the array-specific fields (`Rules.MinItems`/`Rules.MaxItems`). The Zod
  emitter reads `MinItems`/`MaxItems` for arrays, so the constraints were never reached.
  Fixed by routing array `min`/`max` to the correct IR fields in the parser.

## [0.1.2] - 2026-01-25

### Added

- **Enhanced pkg.go.dev documentation**
  - Root doc.go with comprehensive overview and quick start
  - Example tests demonstrating programmatic API usage
  - Enhanced CLI package documentation
- **CODEOWNERS file** for automatic review requests

## [0.1.1] - 2026-01-25

### Added

- **Release automation** with goreleaser
  - Cross-platform binary builds (Linux, macOS, Windows on AMD64/ARM64)
  - Automatic checksums and archives
  - Pre-built binaries attached to GitHub Releases
- **Enhanced version command** showing commit hash and build date

### Fixed

- Go 1.24 compatibility issues in CI (dropped from test matrix)
- Linter configuration for pragmatic v0.1 release

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

- **Map type support**
  - Full support for Go maps: `map[K]V` → `z.record(K, V)`
  - Works with any key and value types (string keys, struct values, nested maps)
  - Automatically handled in type resolution and emission
  - Example: `map[string][]User` → `z.record(z.string(), z.array(UserSchema))`

- **Enum specification finalized**
  - Tag syntax: `enum:value1,value2,value3`
  - Smart token parser handles commas in enum values
  - Supports values with underscores, hyphens, etc.
  - Validates at least one value required
  - String-only (type-checked at parse time)
  - Emits correct `z.enum([...])` syntax
  - Example: `Status string \`gt:"enum:pending,in_progress,completed"\`` → `z.enum(['pending', 'in_progress', 'completed'])`

- **Comprehensive test suite**
  - Unit tests: 39 test functions across all packages
  - Integration tests: 8 end-to-end parser tests with real Go packages
  - Fuzz tests: 12 fuzz targets for continuous bug discovery
  - Test coverage: 53.4% overall (84.8% emitter, 96.4% normalize, 71.7% parser)
  - setupTestModule helper for isolated go/packages testing

- **Continuous fuzzing infrastructure**
  - GitHub Actions workflow running every 30 minutes (48x per day)
  - 12 fuzz targets across parser, emitter, and hash packages
  - Corpus caching for compound growth over time
  - Automatic GitHub issue creation on failure with reproduction steps
  - Detailed statistics in job summaries
  - Found and fixed 2 production bugs before release (UTF-8 and regex escaping)

- **CI/CD infrastructure**
  - Comprehensive test suite running on Go 1.24 and 1.25.6
  - Code linting with golangci-lint (20+ linters enabled)
  - Automated formatting checks (gofmt, goimports)
  - Coverage tracking with Codecov integration
  - CI and Lint badges in README

- **Documentation**
  - Comprehensive tag specification (docs/TAG_SPEC.md)
  - Architecture guide reflecting v0.1 implementation (docs/ARCHITECTURE.md)
  - Complete testing strategy guide (docs/TESTING.md)
  - Fuzzing bug log with technical analysis (docs/FUZZING_BUGS.md)
  - Feature matrix (docs/FEATURES.md)
  - All documentation organized in docs/ directory

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

- **UTF-8 handling** (discovered by fuzzing)
  - camelCase conversion now uses rune slicing instead of byte slicing
  - Fixes invalid UTF-8 output when field names start with multi-byte characters
  - Example bug: "フィールド" → "�\x83\x95ィールド" (now fixed)

- **Regex pattern escaping** (discovered by fuzzing)
  - Properly escape special characters in regex patterns: `/`, `\n`, `\r`, `\t`
  - Fixes broken JavaScript when patterns contain newlines or slashes
  - Example bug: pattern "\n" broke regex across multiple lines (now fixed)

### Internal

- Custom `intToString` helper to avoid stdlib dependency
- Normalized documentation strings with `normalizeDoc()` helper
- JSDoc writing helper with proper escaping (`writeJSDoc()`)
- Kebab-case filename conversion (`toKebabCase()`)
- Error propagation from parser through to CLI

## [0.1.0] - 2026-01-25

First stable release of goldenthread - a schema compiler that generates TypeScript/Zod validation from Go structs.

**Core features:**
- Generate Zod schemas from Go struct tags
- Full type system support (primitives, arrays, maps, nested objects, time.Time)
- Comprehensive validation rules (string length, numeric bounds, formats, enums, patterns)
- Embedded struct flattening with collision detection
- Hash-based drift detection for CI integration
- 53.4% test coverage with continuous fuzzing

**What's included:**
- `goldenthread generate` - Generate Zod schemas from Go code
- `goldenthread check` - Verify schemas are in sync (CI-ready)
- Comprehensive documentation and examples
- Production-ready with 2 bugs found and fixed by fuzzing before release

[Unreleased]: https://github.com/blackwell-systems/goldenthread/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/blackwell-systems/goldenthread/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/blackwell-systems/goldenthread/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/blackwell-systems/goldenthread/releases/tag/v0.1.0
