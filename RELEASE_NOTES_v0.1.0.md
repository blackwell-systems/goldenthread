# goldenthread v0.1.0

**First stable release** - A schema compiler that generates TypeScript/Zod validation from Go structs.

## 🎯 Core Features

- **Generate Zod schemas** from Go struct tags
- **Full type system support**: primitives, arrays, maps, nested objects, time.Time
- **Comprehensive validation**: string length, numeric bounds, formats (email, uuid, url, datetime), enums, patterns
- **Embedded struct flattening** with collision detection
- **Hash-based drift detection** for CI integration
- **53.4% test coverage** with continuous fuzzing

## 🚀 Installation

```bash
go install github.com/blackwell-systems/goldenthread/cmd/goldenthread@v0.1.0
```

**Requirements:** Go 1.25.6 or later

## 📦 What's Included

### Commands

- `goldenthread generate` - Generate Zod schemas from Go code
- `goldenthread check` - Verify schemas are in sync (CI-ready)
- `goldenthread version` - Display version information

### Example

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

const result = UserSchema.safeParse(data)
if (!result.success) {
  console.error('Validation failed:', result.error)
}
```

## 🧪 Quality Assurance

- **Continuous fuzzing**: 12 fuzz targets running every 30 minutes
- **Bugs found before release**: 2 (UTF-8 corruption, regex escaping)
- **CI/CD**: Automated testing and linting on every commit
- **Test coverage**: 89.4% emitter, 96.4% normalize, 75.1% parser

## 📚 Documentation

- [Tag Specification](https://github.com/blackwell-systems/goldenthread/blob/v0.1.0/docs/TAG_SPEC.md) - Complete tag syntax reference
- [Architecture](https://github.com/blackwell-systems/goldenthread/blob/v0.1.0/docs/ARCHITECTURE.md) - System design details
- [Testing Strategy](https://github.com/blackwell-systems/goldenthread/blob/v0.1.0/docs/TESTING.md) - Test suite and fuzzing guide
- [Feature Matrix](https://github.com/blackwell-systems/goldenthread/blob/v0.1.0/docs/FEATURES.md) - Detailed capability breakdown

## 🐛 Known Limitations

goldenthread v0.1 focuses on the core use case. Not yet supported:

- Union types / discriminated unions
- Custom validation functions
- Cross-field validation
- OpenAPI/JSON Schema generation (Zod only for now)

These may be added in future versions based on real-world usage.

## 🔧 CI Integration

```yaml
# .github/workflows/schemas.yml
- name: Check schema drift
  run: goldenthread check ./models
```

Exit code 1 if schemas are out of sync, perfect for CI pipelines.

## 📝 Changelog

See [CHANGELOG.md](https://github.com/blackwell-systems/goldenthread/blob/v0.1.0/CHANGELOG.md) for complete release notes.

## 🙏 Credits

- Built with [go/packages](https://pkg.go.dev/golang.org/x/tools/go/packages) for type-aware parsing
- Inspired by similar tools: tygo, typescriptify, oapi-codegen
- Tested with Go's native fuzzing (introduced in Go 1.18)

## 📄 License

Dual-licensed under Apache-2.0 OR MIT - your choice.

---

**Next Steps:**
1. Try the [Quick Start](https://github.com/blackwell-systems/goldenthread#quick-start) guide
2. Read the [Tag Specification](https://github.com/blackwell-systems/goldenthread/blob/v0.1.0/docs/TAG_SPEC.md)
3. Report issues at [github.com/blackwell-systems/goldenthread/issues](https://github.com/blackwell-systems/goldenthread/issues)

**goldenthread** - Go structs to TypeScript validation, automatically.
