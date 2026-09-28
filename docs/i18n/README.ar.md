[English](../../README.md) · [简体中文](README.zh-CN.md) · [Русский](README.ru.md) · [हिन्दी](README.hi.md) · **العربية**

<p align="center">
  <img src="asset-banner.png" alt="goldenthread">
</p>

> مُصرِّف schema: بُنى Go (structs) → TypeScript/Zod. حافظ على تزامن التحقق بين الخلفية والواجهة الأمامية تلقائيًا.

[![Blackwell Systems™](https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg)](https://github.com/blackwell-systems) 
[![Go Reference](https://pkg.go.dev/badge/github.com/blackwell-systems/goldenthread.svg)](https://pkg.go.dev/github.com/blackwell-systems/goldenthread) 
[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://go.dev/) 
[![CI](https://github.com/blackwell-systems/goldenthread/workflows/CI/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/ci.yml)
[![Lint](https://github.com/blackwell-systems/goldenthread/workflows/Lint/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/lint.yml)

[![License](https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue.svg)](LICENSE-APACHE) 
[![Sponsor](https://img.shields.io/badge/Sponsor-Buy%20Me%20a%20Coffee-yellow?logo=buy-me-a-coffee&logoColor=white)](https://buymeacoffee.com/blackwellsystems)

**goldenthread** هو مُصرِّف schema يعمل في وقت البناء ويولّد تحقق Zod جاهزًا للإنتاج انطلاقًا من بُنى Go. اكتب قواعد التحقق مرة واحدة في وسوم Go، واحصل تلقائيًا على schemas من TypeScript آمنة على مستوى النوع. يكتشف اكتشاف الانحراف المدمج عدم تطابق الـ schema في CI. لا مزامنة يدوية. لا عبء في وقت التشغيل.

## نظرة عامة

يعني الحفاظ على التحقق بين خلفيات Go وواجهات TypeScript الأمامية إبقاء عدة تمثيلات متزامنة. يولّد goldenthread schemas من Zod مباشرةً انطلاقًا من وسوم بُنى Go:

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

تؤدي التغييرات على بُنى Go إلى إعادة توليد schemas من TypeScript تلقائيًا. ويضمن المُصرِّف بقاءها متزامنة.

## المزايا

**يولّد goldenthread schemas من Zod جاهزة للإنتاج مع أمان كامل على مستوى النوع.**

### ما الذي تحصل عليه

- **دعم كامل لأنواع Go**: الأنواع الأولية، والمصفوفات، والخرائط (maps)، والتعدادات (enums)، والكائنات المتداخلة، والمؤشرات (pointers)
- **تحقق شامل**: حدود الطول، والنطاقات العددية، وأنماط regex، ومُدقِّقات الصيغ (email، UUID، URL، IPv4/IPv6، datetime)
- **توليد التعدادات**: `z.enum(['pending', 'completed'])` من حقول Go النصية
- **الاتحادات المُميَّزة**: `z.discriminatedUnion('kind', [...])` من بنية one-of في Go
- **دعم الخرائط**: `z.record(z.string(), T)` لخرائط Go
- **تحقق المصفوفات**: قيود الطول الأدنى/الأقصى
- **الكائنات المتداخلة**: مراجع آمنة على مستوى النوع إلى schemas أخرى
- **تسطيح البُنى المُضمَّنة**: تتم ترقية الحقول المجهولة تلقائيًا
- **اكتشاف التعارضات**: تُلتقط مفاتيح JSON المكررة في وقت التصريف (لا في وقت التشغيل)
- **اكتشاف الانحراف**: يُفشل `goldenthread check` عملية CI إذا كانت الـ schemas غير متزامنة
- **عبء صفري في وقت التشغيل**: توليد كود خالص، بلا reflection، بلا سحر

### قواعد التحقق

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

راجع [docs/TAG_SPEC.md](docs/TAG_SPEC.md) للاطلاع على صياغة الوسوم الكاملة.

## التثبيت

```bash
go install github.com/blackwell-systems/goldenthread/cmd/goldenthread@latest
```

## البدء السريع

### 1. عرّف نماذجك

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

### 2. ولّد الـ schemas

```bash
# Generate from current directory
goldenthread generate ./models

# Generate recursively
goldenthread generate ./models --recursive

# Specify output directory
goldenthread generate ./models --out ./frontend/src/schemas
```

المُخرَج:
```
gen/
  user.ts                    # Generated Zod schema
  .goldenthread.json         # Metadata for drift detection
```

### 3. استخدمها في TypeScript

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

### 4. تحقق من بقاء الـ schemas متزامنة

```bash
# Check if generated schemas match source
goldenthread check ./models

# Returns exit code 1 if out of sync (CI-friendly)
```

أضِف إلى CI:
```yaml
- name: Check schema drift
  run: goldenthread check ./models
```

## البنية المعمارية

goldenthread هو مُصرِّف schema من ثلاث مراحل:

```
Go Source → Parser → Intermediate Representation → Emitter → Generated Code
            (AST)         (Language-agnostic)        (Zod)    (TypeScript)
```

### مكوّنات خط المعالجة

**1. المُحلِّل (Parser)** (`internal/parser`)

يستخدم `go/packages` و`go/types` لأجل تحليل الأنواع بشكل سليم. يستخرج تعريفات البُنى ذات وسوم `gt:`، ويتحقق من صياغة الوسوم وتعارضاتها، ويتعامل مع البُنى المُضمَّنة والمراجع عبر الحزم.

**2. التمثيل الوسيط (Intermediate Representation)** (`internal/schema`)

صيغة schema مستقلة عن اللغة تفصل التحليل عن توليد الكود. تتيح مُصدِّرات (emitters) مستقبلية (OpenAPI، JSON Schema، إلخ) دون تعديل المُحلِّل.

**3. التطبيع (Normalization)** (`internal/normalize`)

يُسطِّح حقول البُنى المُضمَّنة، ويكتشف تعارضات أسماء حقول Go، ويتحقق من تعارضات أسماء JSON، ويضمن صحة الـ schema قبل الإصدار.

**4. المُصدِّر (Emitter)** (`internal/emitter/zod`)

يولّد TypeScript مع schemas من Zod. يحفظ توثيق Go على هيئة JSDoc، ويُصدِر أنواع `z.infer<>` آمنة على مستوى النوع، وينتج مُخرَجًا حتميًا لأجل diffs مستقرة.

**5. التجزئة/اكتشاف الانحراف (Hash/Drift Detection)** (`internal/hash`)

تجزئة SHA-256 لمحتوى الـ schema مع تتبُّع البيانات الوصفية (`.goldenthread.json`). يكتشف متى يتباعد الكود المصدري والكود المُولَّد.

## الأنواع المدعومة

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

راجع [docs/FEATURES.md](docs/FEATURES.md) للاطلاع على التغطية الكاملة للأنواع.

## وسوم التحقق

### الوجود

- `required` - يجب أن يكون الحقل موجودًا (الإعداد الافتراضي لغير المؤشرات)
- `optional` - يمكن حذف الحقل

### قواعد السلاسل النصية

- `len:M..N` - الطول بين M وN من الأحرف
- `min:N` - الطول الأدنى (اختصار لـ `len:N..`)
- `max:N` - الطول الأقصى (اختصار لـ `len:..N`)
- `pattern:REGEX` - يجب أن يطابق تعبيرًا نمطيًا

### القواعد العددية

- `min:N` - القيمة الدنيا
- `max:N` - القيمة القصوى

### مُدقِّقات الصيغ

- `email` - عنوان بريد إلكتروني صالح
- `uuid` - سلسلة UUID صالحة
- `url` - عنوان URL صالح
- `date` - تاريخ بصيغة ISO 8601 (YYYY-MM-DD)
- `datetime` - datetime بصيغة ISO 8601
- `ipv4` - عنوان IPv4
- `ipv6` - عنوان IPv6

### التعدادات

- `enum:value1,value2,value3` - إحدى القيم المحددة

### قواعد المصفوفات

- `min:N` - الطول الأدنى للمصفوفة
- `max:N` - الطول الأقصى للمصفوفة

راجع [docs/TAG_SPEC.md](docs/TAG_SPEC.md) للاطلاع على المواصفة التفصيلية.

## أمثلة

### الكائنات المتداخلة

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

يولّد schema متداخلة من Zod:
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

### البُنى المُضمَّنة

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

يولّد schema مُسطَّحة:
```typescript
const ArticleSchema = z.object({
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
  title: z.string().min(1).max(200),
  content: z.string()
})
```

### التعدادات

```go
type Task struct {
  Title    string `json:"title" gt:"required"`
  Status   string `json:"status" gt:"enum:pending,in_progress,completed,cancelled"`
  Priority string `json:"priority" gt:"enum:low,medium,high"`
}
```

يولّد schemas للتعدادات:
```typescript
const TaskSchema = z.object({
  title: z.string(),
  status: z.enum(['pending', 'in_progress', 'completed', 'cancelled']),
  priority: z.enum(['low', 'medium', 'high'])
})
```

### الاتحادات المُميَّزة

بنية Go تحتوي على حقل `discriminator` وحقل حمولة واحد `variant:<name>`
لكل مُتغيِّر (variant) تُصرَّف إلى اتحاد مُميَّز من Zod:

```go
// WiringElement is exactly one of edge, switch, or join.
type WiringElement struct {
  Kind   string      `json:"kind" gt:"discriminator"`
  Edge   *EdgeSpec   `json:"edge,omitempty"   gt:"variant:edge"`
  Switch *SwitchSpec `json:"switch,omitempty" gt:"variant:switch"`
  Join   *JoinSpec   `json:"join,omitempty"   gt:"variant:join"`
}
```

يولّد اتحادًا مُميَّزًا تُضيِّقه TypeScript حسب `kind`:

```typescript
const WiringElementSchema = z.discriminatedUnion('kind', [
  z.object({ kind: z.literal('edge'),   edge: EdgeSpecSchema }),
  z.object({ kind: z.literal('switch'), switch: SwitchSpecSchema }),
  z.object({ kind: z.literal('join'),   join: JoinSpecSchema })
])
```

تُصدَر الحمولة بوصفها مطلوبة داخل مُتغيِّرها: عنصر موسوم بـ `edge`
بلا حمولة `edge` يفشل في التحقق. راجع [examples/wiring](examples/wiring/).

### الخرائط

```go
type Config struct {
  Name     string            `json:"name" gt:"required"`
  Settings map[string]string `json:"settings" gt:"required"`
  Metadata map[string]any    `json:"metadata" gt:"optional"`
}
```

يولّد schemas من نوع record:
```typescript
const ConfigSchema = z.object({
  name: z.string(),
  settings: z.record(z.string(), z.string()),
  metadata: z.record(z.string(), z.any()).optional()
})
```

## أوامر CLI

### `generate`

توليد schemas من Zod انطلاقًا من مصدر Go:

```bash
goldenthread generate [flags] <directory>

Flags:
  --out <dir>       Output directory (default: ./gen)
  --recursive       Process subdirectories recursively
  --target <name>   Target format (default: zod)
  --infer-json      Infer schemas from structs carrying only json: tags
```

### `check`

التحقق من أن الـ schemas المُولَّدة تطابق المصدر:

```bash
goldenthread check [flags] <directory>

Flags:
  --metadata <file>  Metadata file to check against
  --recursive        Process subdirectories recursively
  --infer-json       Infer schemas from structs carrying only json: tags
```

رموز الخروج:
- `0` - الـ schemas متزامنة
- `1` - الـ schemas غير متزامنة أو حدث خطأ

### `init` _(مُخطَّط له في الإصدار v0.2)_

تهيئة إعدادات goldenthread لمشروعٍ ما.

> **ملاحظة:** لم يُنفَّذ هذا الأمر بعد. الأمر `goldenthread init --wails`
> (الإعداد التلقائي لمشروع Wails) مُخطَّط له في الإصدار v0.2. وتشغيل `goldenthread init`
> حاليًا يطبع رسالة نائبة فحسب.

## الاستنتاج من وسوم json:

افتراضيًا، لا يولّد goldenthread schema لبنيةٍ إلا عندما يحمل أحد حقولها وسم
`gt:`. أما البُنى التي لا تحمل سوى وسوم `json:` القياسية فيُتجاوَز عنها. ويُفعِّل
العلَم `--infer-json` وضعًا ثانيًا: عند ضبطه، يولّد goldenthread أيضًا schemas من
البُنى التي تحمل وسوم `json:` فقط، مُشتقًّا كل حقل من وسم json الخاص به.

```bash
goldenthread generate ./models --infer-json
goldenthread check ./models --infer-json
```

استخدمه للربط بين أنواع قادمة من إطار عمل خارجي يَسِم بُناه أصلًا لأجل ترميز
JSON، لكي تُنتِج Zod دون إضافة تعليقات `gt:` إلى كل حقل. في وضع الاستنتاج:

- اسم الحقل هو اسم وسم json. ويُستبعَد أي حقل موسوم بـ `json:"-"`.
- يكون الحقل اختياريًا عندما يحتوي وسم json الخاص به على `,omitempty` أو عندما
  يكون حقل Go مؤشرًا؛ وإلا فهو مطلوب.
- يأتي النوع الأساسي من نفس تخطيط Go-إلى-Zod المستخدَم لحقول gt الموسومة:
  الأنواع الأولية، و`[]T`، و`map[string]T`، و`*T`، ومراجع البُنى المُسمَّاة. والحقل
  الذي يكون نوعه بنية أخرى (أو شريحة منها) يُصدِر مرجعًا إلى schema تلك البنية،
  وتُولَّد البنية المُشار إليها أيضًا.

تحظى وسوم `gt:` دومًا بالأولوية. فالحقل أو البنية الذي يحمل وسوم `gt:` يحتفظ
بسلوكه الحالي بالضبط: تبقى قواعد التحقق والتعدادات والاتحادات المُميَّزة دون تغيير،
ولا يملأ الاستنتاج سوى الحقول والبُنى التي تفتقر إلى `gt:`. ويُسمَح بالخلط: قد تحتوي
بنية على بعض الحقول الموسومة بـ gt وبعضها الآخر بوسوم json فقط، ومع `--infer-json`
يظهر كلاهما في المُخرَج. ودون العلَم، يبقى السلوك دون تغيير: بُنى json-فقط ما زالت
لا تُنتِج شيئًا.

## القيود الحالية

<details>
<summary>goldenthread v0.1 focuses on the core use case: struct validation for APIs and forms. Click to see what's not yet supported.</summary>

### نظام الأنواع

- أنواع الاتحاد في Go (الواجهات غير الموسومة مع تأكيدات النوع)، استخدم بدلًا منها اتحادًا مُميَّزًا
- المصفوفات ثابتة الطول (`[3]int`)
- القيم الثابتة الحرفية
- الأنواع العَوْدية/ذاتية المرجع

### التحقق

- دوال التحقق المخصصة
- التحقق عبر الحقول (تأكيد كلمة المرور، إلخ)
- العناصر الفريدة في المصفوفات
- التحقق الشرطي

### المُصدِّرات

- توليد OpenAPI/Swagger
- توليد JSON Schema
- أنواع TypeScript (منفصلة عن Zod)
- مكتبات تحقق أخرى

### مزايا الوسوم

- وراثة/تركيب الوسوم
- القواعد الشرطية
- مجموعات تحقق متعددة لكل بنية

هذه القيود مقصودة في الإصدار v0.1. فالأداة **تُتقِن أمرًا واحدًا**: توليد schemas من Zod آمنة على مستوى النوع انطلاقًا من بُنى Go (بما في ذلك الأنواع الأولية والتعدادات والمصفوفات والخرائط والكائنات المتداخلة). وقد تُوسِّع الإصدارات المستقبلية النطاق بناءً على الاستخدام الواقعي.

</details>

## المقارنة مع البدائل

| Tool | Go → TS | Validation | Approach | Use Case |
|------|---------|-----------|----------|----------|
| `validator.v10` | No | Yes | Runtime tags | Go-only validation |
| `swaggo/swag` | No | No | Comments | OpenAPI from Go |
| `oapi-codegen` | Yes | Yes | OpenAPI → Go | Contract-first APIs |
| `protobuf` | Yes | Limited | `.proto` files | Cross-language RPC |
| **goldenthread** | Yes | Yes | Go → Zod | Go-first validation |

goldenthread مُحسَّن للفرق التي:
- تكتب الخلفية بلغة Go
- تكتب الواجهة الأمامية بلغة TypeScript
- تريد أن تكون بُنى Go هي مصدر الحقيقة
- تحتاج إلى تحقق في وقت التشغيل في كلتا اللغتين
- تُفضِّل الأمان على مستوى النوع على المرونة

## التوثيق

- [Tag Specification](docs/TAG_SPEC.md) - مرجع كامل لصياغة الوسوم
- [Feature Matrix](docs/FEATURES.md) - تفصيل مُفصَّل للقدرات
- [Architecture](docs/ARCHITECTURE.md) - تصميم النظام وتفاصيل التنفيذ
- [Testing Strategy](docs/TESTING.md) - مجموعة الاختبارات ودليل الـ fuzzing المستمر
- [Fuzzing Bug Log](docs/FUZZING_BUGS.md) - العلل المُكتشَفة عبر الـ fuzzing المستمر
- [Brand Guidelines](BRAND.md) - استخدام العلامة التجارية والشعار
- [Changelog](CHANGELOG.md) - سجل الإصدارات والتغييرات

## التطوير

### بنية المشروع

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

### البناء

```bash
# Build CLI tool
go build ./cmd/goldenthread

# Run tests
go test ./...

# Run with coverage
go test ./... -cover
```

### المساهمة

المساهمات مُرحَّب بها! مجالات الاهتمام:

1. مُصدِّرات إضافية (OpenAPI، JSON Schema)
2. المزيد من قواعد التحقق
3. تحسينات تغطية الاختبارات
4. التوثيق والأمثلة

## الفلسفة

يتبع goldenthread هذه المبادئ:

1. **بُنى Go هي المرجع القانوني** - ليست ملفات إعدادات، ولا DSLs، ولا تعليقات توضيحية
2. **التوليد في وقت البناء** - لا reflection ولا وكلاء (proxies) في وقت التشغيل
3. **الأمان على مستوى النوع في كل مكان** - التقط الأخطاء في وقت التصريف
4. **البساطة أفضل من الذكاء** - السلوك المُتوقَّع خير من المرونة
5. **الفشل السريع** - الرموز المجهولة تُخطئ فورًا

يأتي الاسم "goldenthread" من النسيج التقليدي: الخيط المفرد المتصل الذي يُمسك القماش معًا. وفي البرمجيات، هو خيط الأمان على مستوى النوع الذي ينبغي أن يمتد دون انقطاع من نماذج المجال مرورًا بواجهات API وصولًا إلى الواجهات الأمامية.

## الترخيص

مُرخَّص ترخيصًا مزدوجًا حسب اختيارك من بين:

- **Apache License 2.0** ([LICENSE-APACHE](LICENSE-APACHE))
- **MIT License** ([LICENSE-MIT](LICENSE-MIT))

يُفضِّل معظم المستخدمين MIT لبساطته. ويوفّر Apache 2.0 حمايات إضافية لبراءات الاختراع.

## العلامات التجارية

**Blackwell Systems™** و**شعار Blackwell Systems** علامتان تجاريتان لـ Dayna Blackwell. يجوز لك استخدام الاسم "Blackwell Systems" للإشارة إلى هذا المشروع، لكن لا يجوز لك استخدام الاسم أو الشعار بطريقة توحي بالمصادقة أو الانتساب الرسمي دون إذن خطي مسبق. راجع [BRAND.md](BRAND.md) للاطلاع على إرشادات الاستخدام.
