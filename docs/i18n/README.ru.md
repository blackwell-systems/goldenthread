[English](../../README.md) · [简体中文](README.zh-CN.md) · **Русский** · [हिन्दी](README.hi.md) · [العربية](README.ar.md)

<p align="center">
  <img src="asset-banner.png" alt="goldenthread">
</p>

> Компилятор схем: структуры Go → TypeScript/Zod. Автоматически поддерживает синхронность валидации на бэкенде и фронтенде.

[![Blackwell Systems™](https://raw.githubusercontent.com/blackwell-systems/blackwell-docs-theme/main/badge-trademark.svg)](https://github.com/blackwell-systems) 
[![Go Reference](https://pkg.go.dev/badge/github.com/blackwell-systems/goldenthread.svg)](https://pkg.go.dev/github.com/blackwell-systems/goldenthread) 
[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://go.dev/) 
[![CI](https://github.com/blackwell-systems/goldenthread/workflows/CI/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/ci.yml)
[![Lint](https://github.com/blackwell-systems/goldenthread/workflows/Lint/badge.svg)](https://github.com/blackwell-systems/goldenthread/actions/workflows/lint.yml)

[![License](https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue.svg)](LICENSE-APACHE) 
[![Sponsor](https://img.shields.io/badge/Sponsor-Buy%20Me%20a%20Coffee-yellow?logo=buy-me-a-coffee&logoColor=white)](https://buymeacoffee.com/blackwellsystems)

**goldenthread** — это компилятор схем времени сборки, который генерирует готовую к продакшену валидацию Zod из структур Go. Опишите правила валидации один раз в тегах Go и автоматически получите типобезопасные схемы TypeScript. Встроенное обнаружение расхождений выявляет несоответствия схем в CI. Никакой ручной синхронизации. Никаких накладных расходов во время выполнения.

## Обзор

Поддержание валидации между бэкендами на Go и фронтендами на TypeScript означает синхронизацию нескольких представлений. goldenthread генерирует схемы Zod напрямую из тегов структур Go:

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

Изменения в структурах Go автоматически перегенерируют схемы TypeScript. Компилятор гарантирует, что они остаются синхронными.

## Возможности

**goldenthread генерирует готовые к продакшену схемы Zod с полной типобезопасностью.**

### Что вы получаете

- **Полная поддержка типов Go**: примитивы, массивы, отображения (maps), перечисления, вложенные объекты, указатели
- **Всесторонняя валидация**: границы длины, числовые диапазоны, регулярные выражения, валидаторы форматов (email, UUID, URL, IPv4/IPv6, datetime)
- **Генерация перечислений**: `z.enum(['pending', 'completed'])` из строковых полей Go
- **Размеченные объединения**: `z.discriminatedUnion('kind', [...])` из one-of-структуры Go
- **Поддержка отображений**: `z.record(z.string(), T)` для отображений Go
- **Валидация массивов**: ограничения на минимальную/максимальную длину
- **Вложенные объекты**: типобезопасные ссылки на другие схемы
- **Уплощение встроенных структур**: анонимные поля продвигаются автоматически
- **Обнаружение коллизий**: повторяющиеся ключи JSON выявляются на этапе компиляции (а не во время выполнения)
- **Обнаружение расхождений**: `goldenthread check` завершает CI ошибкой, если схемы рассинхронизированы
- **Нулевые накладные расходы во время выполнения**: чистая генерация кода, без рефлексии, без «магии»

### Правила валидации

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

Полный синтаксис тегов см. в [docs/TAG_SPEC.md](docs/TAG_SPEC.md).

## Установка

```bash
go install github.com/blackwell-systems/goldenthread/cmd/goldenthread@latest
```

## Быстрый старт

### 1. Определите свои модели

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

### 2. Сгенерируйте схемы

```bash
# Generate from current directory
goldenthread generate ./models

# Generate recursively
goldenthread generate ./models --recursive

# Specify output directory
goldenthread generate ./models --out ./frontend/src/schemas
```

Вывод:
```
gen/
  user.ts                    # Generated Zod schema
  .goldenthread.json         # Metadata for drift detection
```

### 3. Используйте в TypeScript

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

### 4. Проверьте, что схемы остаются синхронными

```bash
# Check if generated schemas match source
goldenthread check ./models

# Returns exit code 1 if out of sync (CI-friendly)
```

Добавьте в CI:
```yaml
- name: Check schema drift
  run: goldenthread check ./models
```

## Архитектура

goldenthread — это компилятор схем с тремя этапами:

```
Go Source → Parser → Intermediate Representation → Emitter → Generated Code
            (AST)         (Language-agnostic)        (Zod)    (TypeScript)
```

### Компоненты конвейера

**1. Парсер** (`internal/parser`)

Использует `go/packages` и `go/types` для корректного разрешения типов. Извлекает определения структур с тегами `gt:`, проверяет синтаксис тегов и конфликты, обрабатывает встроенные структуры и межпакетные ссылки.

**2. Промежуточное представление** (`internal/schema`)

Независимый от языка формат схемы, отделяющий разбор от генерации кода. Позволяет добавлять будущие эмиттеры (OpenAPI, JSON Schema и т. д.) без изменения парсера.

**3. Нормализация** (`internal/normalize`)

Уплощает поля встроенных структур, обнаруживает коллизии имён полей Go, проверяет коллизии имён JSON, обеспечивает корректность схемы перед эмиссией.

**4. Эмиттер** (`internal/emitter/zod`)

Генерирует TypeScript со схемами Zod. Сохраняет документацию Go в виде JSDoc, эмитирует типобезопасные типы `z.infer<>`, выдаёт детерминированный вывод для стабильных диффов.

**5. Хеширование/обнаружение расхождений** (`internal/hash`)

Хеширование содержимого схемы по SHA-256 с отслеживанием метаданных (`.goldenthread.json`). Обнаруживает, когда исходный и сгенерированный код расходятся.

## Поддерживаемые типы

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

Полный охват типов см. в [docs/FEATURES.md](docs/FEATURES.md).

## Теги валидации

### Наличие

- `required` — поле должно присутствовать (по умолчанию для не-указателей)
- `optional` — поле можно опустить

### Правила для строк

- `len:M..N` — длина от M до N символов
- `min:N` — минимальная длина (сокращение для `len:N..`)
- `max:N` — максимальная длина (сокращение для `len:..N`)
- `pattern:REGEX` — должно соответствовать регулярному выражению

### Числовые правила

- `min:N` — минимальное значение
- `max:N` — максимальное значение

### Валидаторы форматов

- `email` — корректный адрес электронной почты
- `uuid` — корректная строка UUID
- `url` — корректный URL
- `date` — дата в формате ISO 8601 (YYYY-MM-DD)
- `datetime` — datetime в формате ISO 8601
- `ipv4` — адрес IPv4
- `ipv6` — адрес IPv6

### Перечисления

- `enum:value1,value2,value3` — одно из указанных значений

### Правила для массивов

- `min:N` — минимальная длина массива
- `max:N` — максимальная длина массива

Подробную спецификацию см. в [docs/TAG_SPEC.md](docs/TAG_SPEC.md).

## Примеры

### Вложенные объекты

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

Генерирует вложенную схему Zod:
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

### Встроенные структуры

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

Генерирует уплощённую схему:
```typescript
const ArticleSchema = z.object({
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
  title: z.string().min(1).max(200),
  content: z.string()
})
```

### Перечисления

```go
type Task struct {
  Title    string `json:"title" gt:"required"`
  Status   string `json:"status" gt:"enum:pending,in_progress,completed,cancelled"`
  Priority string `json:"priority" gt:"enum:low,medium,high"`
}
```

Генерирует схемы перечислений:
```typescript
const TaskSchema = z.object({
  title: z.string(),
  status: z.enum(['pending', 'in_progress', 'completed', 'cancelled']),
  priority: z.enum(['low', 'medium', 'high'])
})
```

### Размеченные объединения

Структура Go с полем `discriminator` и одним полем полезной нагрузки
`variant:<name>` на каждый вариант компилируется в размеченное объединение Zod:

```go
// WiringElement is exactly one of edge, switch, or join.
type WiringElement struct {
  Kind   string      `json:"kind" gt:"discriminator"`
  Edge   *EdgeSpec   `json:"edge,omitempty"   gt:"variant:edge"`
  Switch *SwitchSpec `json:"switch,omitempty" gt:"variant:switch"`
  Join   *JoinSpec   `json:"join,omitempty"   gt:"variant:join"`
}
```

Генерирует размеченное объединение, которое TypeScript сужает по `kind`:

```typescript
const WiringElementSchema = z.discriminatedUnion('kind', [
  z.object({ kind: z.literal('edge'),   edge: EdgeSpecSchema }),
  z.object({ kind: z.literal('switch'), switch: SwitchSpecSchema }),
  z.object({ kind: z.literal('join'),   join: JoinSpecSchema })
])
```

Полезная нагрузка эмитируется как обязательная внутри своего варианта: элемент,
помеченный как `edge`, но без полезной нагрузки `edge`, не проходит валидацию.
См. [examples/wiring](examples/wiring/).

### Отображения

```go
type Config struct {
  Name     string            `json:"name" gt:"required"`
  Settings map[string]string `json:"settings" gt:"required"`
  Metadata map[string]any    `json:"metadata" gt:"optional"`
}
```

Генерирует схемы record:
```typescript
const ConfigSchema = z.object({
  name: z.string(),
  settings: z.record(z.string(), z.string()),
  metadata: z.record(z.string(), z.any()).optional()
})
```

## Команды CLI

### `generate`

Сгенерировать схемы Zod из исходного кода Go:

```bash
goldenthread generate [flags] <directory>

Flags:
  --out <dir>       Output directory (default: ./gen)
  --recursive       Process subdirectories recursively
  --target <name>   Target format (default: zod)
  --infer-json      Infer schemas from structs carrying only json: tags
```

### `check`

Проверить, что сгенерированные схемы соответствуют исходному коду:

```bash
goldenthread check [flags] <directory>

Flags:
  --metadata <file>  Metadata file to check against
  --recursive        Process subdirectories recursively
  --infer-json       Infer schemas from structs carrying only json: tags
```

Коды выхода:
- `0` — схемы синхронны
- `1` — схемы рассинхронизированы или произошла ошибка

### `init` _(запланировано на v0.2)_

Инициализировать конфигурацию goldenthread для проекта.

> **Примечание.** Эта команда ещё не реализована. `goldenthread init --wails`
> (автонастройка проекта Wails) запланирована на v0.2. Сейчас запуск
> `goldenthread init` лишь выводит сообщение-заглушку.

## Вывод из тегов json:

По умолчанию goldenthread генерирует схему для структуры, только если одно из её
полей несёт тег `gt:`. Структуры, у которых нет ничего, кроме стандартных тегов
`json:`, пропускаются. Флаг `--infer-json` включает второй режим: когда он
установлен, goldenthread также генерирует схемы из структур, несущих только теги
`json:`, выводя каждое поле из его json-тега.

```bash
goldenthread generate ./models --infer-json
goldenthread check ./models --infer-json
```

Используйте его, чтобы навести мост к типам из внешнего фреймворка, который уже
помечает свои структуры для JSON-кодирования, и получить Zod без добавления
аннотаций `gt:` к каждому полю. При выводе:

- Имя поля — это имя json-тега. Поле, помеченное `json:"-"`, исключается.
- Поле необязательно, когда его json-тег содержит `,omitempty` или само поле Go
  является указателем; в противном случае оно обязательно.
- Базовый тип берётся из того же отображения Go-в-Zod, что и для полей с тегами
  gt: примитивы, `[]T`, `map[string]T`, `*T` и ссылки на именованные структуры.
  Поле, тип которого — другая структура (или её срез), эмитирует ссылку на схему
  этой структуры, и упомянутая структура тоже генерируется.

Теги `gt:` всегда имеют приоритет. Поле или структура с тегами `gt:` сохраняют
своё точное существующее поведение: правила валидации, перечисления и размеченные
объединения не меняются, а вывод лишь заполняет поля и структуры, у которых нет
`gt:`. Смешивание допускается: структура может иметь одни поля с тегами gt и
другие — только с json, и с `--infer-json` оба вида появляются в выводе. Без
этого флага поведение не меняется: структуры только с json по-прежнему ничего не
дают.

## Текущие ограничения

<details>
<summary>goldenthread v0.1 focuses on the core use case: struct validation for APIs and forms. Click to see what's not yet supported.</summary>

### Система типов

- Объединённые типы Go (неразмеченные интерфейсы с приведением типов) — используйте вместо этого размеченное объединение
- Массивы фиксированной длины (`[3]int`)
- Литеральные константные значения
- Рекурсивные/самоссылающиеся типы

### Валидация

- Пользовательские функции валидации
- Межполевая валидация (подтверждение пароля и т. п.)
- Уникальность элементов в массивах
- Условная валидация

### Эмиттеры

- Генерация OpenAPI/Swagger
- Генерация JSON Schema
- Типы TypeScript (отдельно от Zod)
- Другие библиотеки валидации

### Возможности тегов

- Наследование/композиция тегов
- Условные правила
- Несколько наборов валидации на структуру

Эти ограничения намеренны для v0.1. Инструмент **делает одно дело хорошо**: генерирует типобезопасные схемы Zod из структур Go (включая примитивы, перечисления, массивы, отображения и вложенные объекты). Будущие версии могут расширить охват на основе реального использования.

</details>

## Сравнение с альтернативами

| Tool | Go → TS | Validation | Approach | Use Case |
|------|---------|-----------|----------|----------|
| `validator.v10` | No | Yes | Runtime tags | Go-only validation |
| `swaggo/swag` | No | No | Comments | OpenAPI from Go |
| `oapi-codegen` | Yes | Yes | OpenAPI → Go | Contract-first APIs |
| `protobuf` | Yes | Limited | `.proto` files | Cross-language RPC |
| **goldenthread** | Yes | Yes | Go → Zod | Go-first validation |

goldenthread оптимизирован для команд, которые:
- Пишут бэкенд на Go
- Пишут фронтенд на TypeScript
- Хотят использовать структуры Go как источник истины
- Нуждаются в валидации во время выполнения на обоих языках
- Предпочитают типобезопасность гибкости

## Документация

- [Tag Specification](docs/TAG_SPEC.md) — полный справочник по синтаксису тегов
- [Feature Matrix](docs/FEATURES.md) — подробная разбивка возможностей
- [Architecture](docs/ARCHITECTURE.md) — архитектура системы и детали реализации
- [Testing Strategy](docs/TESTING.md) — набор тестов и руководство по непрерывному фаззингу
- [Fuzzing Bug Log](docs/FUZZING_BUGS.md) — ошибки, обнаруженные непрерывным фаззингом
- [Brand Guidelines](BRAND.md) — использование товарного знака и логотипа
- [Changelog](CHANGELOG.md) — история версий и изменения

## Разработка

### Структура проекта

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

### Сборка

```bash
# Build CLI tool
go build ./cmd/goldenthread

# Run tests
go test ./...

# Run with coverage
go test ./... -cover
```

### Участие в разработке

Мы рады вкладу! Интересующие направления:

1. Дополнительные эмиттеры (OpenAPI, JSON Schema)
2. Больше правил валидации
3. Улучшение покрытия тестами
4. Документация и примеры

## Философия

goldenthread следует этим принципам:

1. **Структуры Go каноничны** — не файлы конфигурации, не DSL, не аннотации
2. **Генерация во время сборки** — не рефлексия во время выполнения и не прокси
3. **Типобезопасность повсюду** — выявляйте ошибки на этапе компиляции
4. **Просто лучше, чем хитро** — предсказуемое поведение важнее гибкости
5. **Отказывать быстро** — неизвестные токены вызывают ошибку немедленно

Название «goldenthread» происходит из традиционного ткачества: единая непрерывная нить, которая держит ткань вместе. В программном обеспечении это нить типобезопасности, которая должна проходить непрерывно от доменных моделей через API к фронтендам.

## Лицензия

Двойное лицензирование на ваш выбор:

- **Apache License 2.0** ([LICENSE-APACHE](LICENSE-APACHE))
- **MIT License** ([LICENSE-MIT](LICENSE-MIT))

Большинство пользователей предпочитают MIT ради простоты. Apache 2.0 предоставляет дополнительную патентную защиту.

## Товарные знаки

**Blackwell Systems™** и **логотип Blackwell Systems** являются товарными знаками Dayna Blackwell. Вы можете использовать название «Blackwell Systems» для обозначения этого проекта, но не вправе использовать это название или логотип так, чтобы это создавало впечатление одобрения или официальной аффилиации, без предварительного письменного разрешения. Рекомендации по использованию см. в [BRAND.md](BRAND.md).
