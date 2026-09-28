[English](../../README.md) · [简体中文](README.zh-CN.md) · [Русский](README.ru.md) · **हिन्दी**

<p align="center">
  <img src="asset-banner.png" alt="goldenthread">
</p>

> Schema कंपाइलर: Go structs → TypeScript/Zod। बैकएंड/फ्रंटएंड सत्यापन को अपने आप समकालिक रखें।

[![Blackwell Systems™](https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg)](https://github.com/blackwell-systems) 
[![Go Reference](https://pkg.go.dev/badge/github.com/blackwell-systems/goldenthread.svg)](https://pkg.go.dev/github.com/blackwell-systems/goldenthread) 
[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://go.dev/) 
[![CI](https://github.com/blackwell-systems/goldenthread/workflows/CI/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/ci.yml)
[![Lint](https://github.com/blackwell-systems/goldenthread/workflows/Lint/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/lint.yml)

[![License](https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue.svg)](LICENSE-APACHE) 
[![Sponsor](https://img.shields.io/badge/Sponsor-Buy%20Me%20a%20Coffee-yellow?logo=buy-me-a-coffee&logoColor=white)](https://buymeacoffee.com/blackwellsystems)

**goldenthread** एक build-time schema कंपाइलर है जो Go structs से production-ready Zod सत्यापन उत्पन्न करता है। सत्यापन नियम Go tags में एक बार लिखें, और अपने आप type-safe TypeScript schemas पाएँ। अंतर्निहित drift detection CI में schema बेमेल को पकड़ता है। कोई मैनुअल समकालन नहीं। कोई runtime overhead नहीं।

## अवलोकन

Go बैकएंड और TypeScript फ्रंटएंड के बीच सत्यापन बनाए रखने का अर्थ है कई प्रतिनिधित्वों को समकालिक रखना। goldenthread सीधे Go struct tags से Zod schemas उत्पन्न करता है:

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

Go structs में बदलाव TypeScript schemas को अपने आप पुनः उत्पन्न करते हैं। कंपाइलर सुनिश्चित करता है कि वे समकालिक बने रहें।

## विशेषताएँ

**goldenthread पूर्ण type safety के साथ production-ready Zod schemas उत्पन्न करता है।**

### आपको क्या मिलता है

- **पूर्ण Go type समर्थन**: primitives, arrays, maps, enums, nested objects, pointers
- **व्यापक सत्यापन**: लंबाई सीमाएँ, संख्यात्मक श्रेणियाँ, regex patterns, format validators (email, UUID, URL, IPv4/IPv6, datetime)
- **Enum उत्पादन**: Go string fields से `z.enum(['pending', 'completed'])`
- **Discriminated unions**: एक Go one-of struct से `z.discriminatedUnion('kind', [...])`
- **Map समर्थन**: Go maps के लिए `z.record(z.string(), T)`
- **Array सत्यापन**: न्यूनतम/अधिकतम लंबाई की बाध्यताएँ
- **Nested objects**: अन्य schemas के type-safe संदर्भ
- **Embedded struct समतलन**: anonymous fields अपने आप प्रोन्नत होते हैं
- **Collision detection**: दोहराई गई JSON keys compile time पर पकड़ी जाती हैं (runtime पर नहीं)
- **Drift detection**: schemas समकालन से बाहर होने पर `goldenthread check` CI को विफल करता है
- **शून्य runtime overhead**: शुद्ध कोड उत्पादन, कोई reflection नहीं, कोई जादू नहीं

### सत्यापन नियम

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

पूर्ण tag syntax के लिए [docs/TAG_SPEC.md](docs/TAG_SPEC.md) देखें।

## स्थापना

```bash
go install github.com/blackwell-systems/goldenthread/cmd/goldenthread@latest
```

## त्वरित प्रारंभ

### 1. अपने models परिभाषित करें

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

### 2. Schemas उत्पन्न करें

```bash
# Generate from current directory
goldenthread generate ./models

# Generate recursively
goldenthread generate ./models --recursive

# Specify output directory
goldenthread generate ./models --out ./frontend/src/schemas
```

आउटपुट:
```
gen/
  user.ts                    # Generated Zod schema
  .goldenthread.json         # Metadata for drift detection
```

### 3. TypeScript में उपयोग करें

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

### 4. सत्यापित करें कि schemas समकालिक रहें

```bash
# Check if generated schemas match source
goldenthread check ./models

# Returns exit code 1 if out of sync (CI-friendly)
```

CI में जोड़ें:
```yaml
- name: Check schema drift
  run: goldenthread check ./models
```

## आर्किटेक्चर

goldenthread तीन चरणों वाला एक schema कंपाइलर है:

```
Go Source → Parser → Intermediate Representation → Emitter → Generated Code
            (AST)         (Language-agnostic)        (Zod)    (TypeScript)
```

### पाइपलाइन घटक

**1. Parser** (`internal/parser`)

उचित type resolution के लिए `go/packages` और `go/types` का उपयोग करता है। `gt:` tags वाली struct परिभाषाएँ निकालता है, tag syntax और टकरावों को सत्यापित करता है, embedded structs और cross-package संदर्भों को संभालता है।

**2. Intermediate Representation** (`internal/schema`)

एक language-agnostic schema प्रारूप जो parsing को code generation से अलग करता है। parser में बदलाव किए बिना भविष्य के emitters (OpenAPI, JSON Schema, आदि) को सक्षम करता है।

**3. Normalization** (`internal/normalize`)

embedded struct fields को समतल करता है, Go field name टकरावों का पता लगाता है, JSON name टकरावों को सत्यापित करता है, emission से पहले schema की शुद्धता सुनिश्चित करता है।

**4. Emitter** (`internal/emitter/zod`)

Zod schemas के साथ TypeScript उत्पन्न करता है। Go documentation को JSDoc के रूप में संरक्षित करता है, type-safe `z.infer<>` types उत्सर्जित करता है, स्थिर diffs के लिए deterministic output देता है।

**5. Hash/Drift Detection** (`internal/hash`)

metadata tracking (`.goldenthread.json`) के साथ schema सामग्री का SHA-256 hashing। पता लगाता है कि source और generated code कब अलग हो जाते हैं।

## समर्थित Types

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

पूर्ण type coverage के लिए [docs/FEATURES.md](docs/FEATURES.md) देखें।

## सत्यापन Tags

### उपस्थिति

- `required` - field उपस्थित होना चाहिए (non-pointers के लिए डिफ़ॉल्ट)
- `optional` - field छोड़ा जा सकता है

### String नियम

- `len:M..N` - लंबाई M से N वर्णों के बीच
- `min:N` - न्यूनतम लंबाई (`len:N..` का संक्षिप्त रूप)
- `max:N` - अधिकतम लंबाई (`len:..N` का संक्षिप्त रूप)
- `pattern:REGEX` - regular expression से मेल खाना चाहिए

### संख्यात्मक नियम

- `min:N` - न्यूनतम मान
- `max:N` - अधिकतम मान

### Format Validators

- `email` - वैध ईमेल पता
- `uuid` - वैध UUID string
- `url` - वैध URL
- `date` - ISO 8601 तिथि (YYYY-MM-DD)
- `datetime` - ISO 8601 datetime
- `ipv4` - IPv4 पता
- `ipv6` - IPv6 पता

### Enums

- `enum:value1,value2,value3` - निर्दिष्ट मानों में से एक

### Array नियम

- `min:N` - न्यूनतम array लंबाई
- `max:N` - अधिकतम array लंबाई

विस्तृत विनिर्देश के लिए [docs/TAG_SPEC.md](docs/TAG_SPEC.md) देखें।

## उदाहरण

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

nested Zod schema उत्पन्न करता है:
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

समतल किया गया schema उत्पन्न करता है:
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

enum schemas उत्पन्न करता है:
```typescript
const TaskSchema = z.object({
  title: z.string(),
  status: z.enum(['pending', 'in_progress', 'completed', 'cancelled']),
  priority: z.enum(['low', 'medium', 'high'])
})
```

### Discriminated Unions

एक Go struct जिसमें एक `discriminator` field और प्रति variant एक `variant:<name>`
payload field हो, वह एक Zod discriminated union में कंपाइल होता है:

```go
// WiringElement is exactly one of edge, switch, or join.
type WiringElement struct {
  Kind   string      `json:"kind" gt:"discriminator"`
  Edge   *EdgeSpec   `json:"edge,omitempty"   gt:"variant:edge"`
  Switch *SwitchSpec `json:"switch,omitempty" gt:"variant:switch"`
  Join   *JoinSpec   `json:"join,omitempty"   gt:"variant:join"`
}
```

एक discriminated union उत्पन्न करता है जिसे TypeScript `kind` पर संकीर्ण करता है:

```typescript
const WiringElementSchema = z.discriminatedUnion('kind', [
  z.object({ kind: z.literal('edge'),   edge: EdgeSpecSchema }),
  z.object({ kind: z.literal('switch'), switch: SwitchSpecSchema }),
  z.object({ kind: z.literal('join'),   join: JoinSpecSchema })
])
```

payload अपने variant के भीतर required के रूप में उत्सर्जित होता है: `edge` के रूप
में टैग किया गया एक element जिसमें कोई `edge` payload न हो, सत्यापन में विफल हो
जाता है। देखें [examples/wiring](examples/wiring/)।

### Maps

```go
type Config struct {
  Name     string            `json:"name" gt:"required"`
  Settings map[string]string `json:"settings" gt:"required"`
  Metadata map[string]any    `json:"metadata" gt:"optional"`
}
```

record schemas उत्पन्न करता है:
```typescript
const ConfigSchema = z.object({
  name: z.string(),
  settings: z.record(z.string(), z.string()),
  metadata: z.record(z.string(), z.any()).optional()
})
```

## CLI कमांड

### `generate`

Go source से Zod schemas उत्पन्न करें:

```bash
goldenthread generate [flags] <directory>

Flags:
  --out <dir>       Output directory (default: ./gen)
  --recursive       Process subdirectories recursively
  --target <name>   Target format (default: zod)
  --infer-json      Infer schemas from structs carrying only json: tags
```

### `check`

सत्यापित करें कि उत्पन्न schemas source से मेल खाते हैं:

```bash
goldenthread check [flags] <directory>

Flags:
  --metadata <file>  Metadata file to check against
  --recursive        Process subdirectories recursively
  --infer-json       Infer schemas from structs carrying only json: tags
```

Exit codes:
- `0` - schemas समकालिक
- `1` - schemas समकालन से बाहर या त्रुटि

### `init` _(v0.2 के लिए नियोजित)_

किसी project के लिए goldenthread configuration आरंभ करें।

> **नोट:** यह कमांड अभी लागू नहीं हुआ है। `goldenthread init --wails`
> (Wails project auto-setup) v0.2 के लिए नियोजित है। वर्तमान में `goldenthread init`
> चलाने पर केवल एक placeholder संदेश प्रिंट होता है।

## json: tags से अनुमान लगाना

डिफ़ॉल्ट रूप से, goldenthread किसी struct के लिए schema तभी उत्पन्न करता है जब उसके
किसी एक field पर `gt:` tag हो। जिन structs में केवल मानक `json:` tags होते हैं उन्हें
छोड़ दिया जाता है। `--infer-json` flag एक दूसरे mode को सक्षम करता है: जब यह सेट होता
है, तो goldenthread उन structs से भी schemas उत्पन्न करता है जिनमें केवल `json:`
tags होते हैं, और प्रत्येक field को उसके json tag से व्युत्पन्न करता है।

```bash
goldenthread generate ./models --infer-json
goldenthread check ./models --infer-json
```

इसका उपयोग किसी बाहरी framework से types को जोड़ने के लिए करें जो पहले से ही अपने
structs को JSON encoding के लिए tag करता है, ताकि आप हर field में `gt:` annotations
जोड़े बिना Zod उत्पन्न कर सकें। अनुमान के अंतर्गत:

- field का नाम json tag का नाम है। `json:"-"` के रूप में tag किया गया field बाहर रखा
  जाता है।
- कोई field तब optional होता है जब उसके json tag में `,omitempty` हो या वह Go field
  एक pointer हो; अन्यथा वह required होता है।
- base type उसी Go-to-Zod mapping से आता है जो gt-tagged fields के लिए प्रयुक्त
  होता है: primitives, `[]T`, `map[string]T`, `*T`, और named-struct संदर्भ। जिस
  field का type कोई अन्य struct (या उसका slice) हो, वह उस struct के schema का एक
  संदर्भ उत्सर्जित करता है, और संदर्भित struct भी उत्पन्न होता है।

`gt:` tags हमेशा प्राथमिकता रखते हैं। जिस field या struct में `gt:` tags होते हैं
वह अपना ठीक वही मौजूदा व्यवहार बनाए रखता है: सत्यापन नियम, enums, और discriminated
unions अपरिवर्तित रहते हैं, और अनुमान केवल उन fields और structs को भरता है जिनमें
`gt:` नहीं है। मिश्रण की अनुमति है: एक struct में कुछ gt-tagged fields और कुछ
केवल-json fields हो सकते हैं, और `--infer-json` के साथ दोनों output में दिखाई देते
हैं। इस flag के बिना, व्यवहार अपरिवर्तित रहता है: केवल-json structs अब भी कुछ उत्पन्न
नहीं करते।

## वर्तमान सीमाएँ

<details>
<summary>goldenthread v0.1 focuses on the core use case: struct validation for APIs and forms. Click to see what's not yet supported.</summary>

### Type System

- Go union types (type assertions वाले untagged interfaces), इसके बजाय एक discriminated union का उपयोग करें
- निश्चित-लंबाई वाले arrays (`[3]int`)
- Literal constant मान
- Recursive/self-referential types

### सत्यापन

- Custom validation functions
- Cross-field validation (password confirmation, आदि)
- Arrays में unique items
- Conditional validation

### Emitters

- OpenAPI/Swagger उत्पादन
- JSON Schema उत्पादन
- TypeScript types (Zod से अलग)
- अन्य validation libraries

### Tag विशेषताएँ

- Tag inheritance/composition
- Conditional नियम
- प्रति struct अनेक validation sets

ये सीमाएँ v0.1 के लिए जानबूझकर रखी गई हैं। यह tool **एक काम अच्छी तरह करता है**: Go structs (जिनमें primitives, enums, arrays, maps, और nested objects शामिल हैं) से type-safe Zod schemas उत्पन्न करना। भविष्य के संस्करण वास्तविक उपयोग के आधार पर दायरे का विस्तार कर सकते हैं।

</details>

## विकल्पों के साथ तुलना

| Tool | Go → TS | Validation | Approach | Use Case |
|------|---------|-----------|----------|----------|
| `validator.v10` | No | Yes | Runtime tags | Go-only validation |
| `swaggo/swag` | No | No | Comments | OpenAPI from Go |
| `oapi-codegen` | Yes | Yes | OpenAPI → Go | Contract-first APIs |
| `protobuf` | Yes | Limited | `.proto` files | Cross-language RPC |
| **goldenthread** | Yes | Yes | Go → Zod | Go-first validation |

goldenthread उन टीमों के लिए अनुकूलित है जो:
- बैकएंड Go में लिखती हैं
- फ्रंटएंड TypeScript में लिखती हैं
- Go structs को सत्य के स्रोत के रूप में चाहती हैं
- दोनों भाषाओं में runtime सत्यापन की आवश्यकता रखती हैं
- लचीलेपन की तुलना में type safety को प्राथमिकता देती हैं

## दस्तावेज़ीकरण

- [Tag Specification](docs/TAG_SPEC.md) - पूर्ण tag syntax संदर्भ
- [Feature Matrix](docs/FEATURES.md) - विस्तृत क्षमता विवरण
- [Architecture](docs/ARCHITECTURE.md) - सिस्टम डिज़ाइन और कार्यान्वयन विवरण
- [Testing Strategy](docs/TESTING.md) - test suite और continuous fuzzing मार्गदर्शिका
- [Fuzzing Bug Log](docs/FUZZING_BUGS.md) - continuous fuzzing द्वारा खोजी गई बग्स
- [Brand Guidelines](BRAND.md) - Trademark और logo उपयोग
- [Changelog](CHANGELOG.md) - संस्करण इतिहास और परिवर्तन

## विकास

### Project संरचना

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

### निर्माण

```bash
# Build CLI tool
go build ./cmd/goldenthread

# Run tests
go test ./...

# Run with coverage
go test ./... -cover
```

### योगदान

योगदान का स्वागत है! रुचि के क्षेत्र:

1. अतिरिक्त emitters (OpenAPI, JSON Schema)
2. अधिक सत्यापन नियम
3. Test coverage में सुधार
4. दस्तावेज़ीकरण और उदाहरण

## दर्शन

goldenthread इन सिद्धांतों का पालन करता है:

1. **Go structs आधिकारिक हैं** - config files नहीं, DSLs नहीं, annotations नहीं
2. **Build-time उत्पादन** - runtime reflection या proxies नहीं
3. **हर जगह type safety** - त्रुटियाँ compile time पर पकड़ें
4. **चतुर से बेहतर सरल** - अनुमानित व्यवहार लचीलेपन से बेहतर है
5. **तेज़ी से विफल हों** - अज्ञात tokens तुरंत त्रुटि देते हैं

नाम "goldenthread" पारंपरिक बुनाई से आता है: वह एकल निरंतर धागा जो कपड़े को एक साथ बाँधे रखता है। सॉफ़्टवेयर में, यह type safety का वह धागा है जो domain models से लेकर APIs होते हुए frontends तक अटूट चलना चाहिए।

## लाइसेंस

आपकी पसंद के अनुसार दोहरे-लाइसेंस के अंतर्गत:

- **Apache License 2.0** ([LICENSE-APACHE](LICENSE-APACHE))
- **MIT License** ([LICENSE-MIT](LICENSE-MIT))

अधिकांश उपयोगकर्ता सरलता के लिए MIT को प्राथमिकता देते हैं। Apache 2.0 अतिरिक्त patent सुरक्षा प्रदान करता है।

## ट्रेडमार्क

**Blackwell Systems™** और **Blackwell Systems logo**, Dayna Blackwell के trademarks हैं। आप इस project को संदर्भित करने के लिए "Blackwell Systems" नाम का उपयोग कर सकते हैं, परंतु पूर्व लिखित अनुमति के बिना आप इस नाम या logo का उपयोग इस प्रकार नहीं कर सकते जिससे endorsement या आधिकारिक संबद्धता का आभास हो। उपयोग दिशानिर्देशों के लिए [BRAND.md](BRAND.md) देखें।
