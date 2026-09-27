# goldenthread v0.1 Tag Specification

This document defines the canonical behavior of `gt:` struct tags.

## Tag Key

Struct tag key: `gt:"..."`

## Grammar

- Tokens are comma-separated: `gt:"token,token,key:value,key:value"`
- Whitespace around tokens is ignored
- Token names are lowercase ASCII

## Tokens

### Presence

- `required` - marks field as required
- `optional` - marks field as optional

**Precedence (final Optional flag):**

1. If `required` present → `Optional = false`
2. Else if `optional` present → `Optional = true`
3. Else if field is pointer → `Optional = true`
4. Else if `json` tag includes `omitempty` → `Optional = true`
5. Else → `Optional = false` (required by default)

**Conflicts:**

- If both `required` and `optional` → **error**
- If `required` and (`omitempty` or pointer) → allowed, but **warn** ("required overrides omitempty/pointer")

### Numeric

- `min:N` - minimum value (signed decimal)
- `max:N` - maximum value (signed decimal)

`N` is stored as float64 in IR.

**Conflicts:**

- `min > max` → **error**

**Array applicability:** `min:N` and `max:N` also apply to array/slice fields,
constraining the minimum and maximum number of items.

### String Length

- `len:M..N` - length range where M and N are non-negative integers, M ≤ N

Maps to:
- `Rules.MinLength = M`
- `Rules.MaxLength = N`

**Conflicts:**

- If used on non-string → **error**

### Pattern

- `pattern:REGEX` - regex pattern

Semantics:
- Treated as raw regex string
- For Zod: embedded in `/.../` with proper escaping

**Conflicts:**

- If used on non-string → **error**

### Enum (string-only)

- `enum:val1,val2,val3` - restricts field to one of the specified string values

Maps to `FieldRules.Enum []string` in IR.

**Validation:**

- Must have at least one value
- Values are trimmed of whitespace
- Empty values after trimming are filtered out

**Conflicts:**

- If used on non-string → **error**
- Enum values with commas not currently supported (limitation)

**Examples:**

```go
Status string `gt:"enum:active,inactive,pending"`  // ✓ Valid
Role   string `gt:"enum:admin,user"`              // ✓ Valid
Count  int    `gt:"enum:1,2,3"`                    // ✗ Error: enum only for strings
```

### Discriminated Unions

Two tags describe a tagged one-of on a Go struct. The struct becomes a Zod
`z.discriminatedUnion` instead of a `z.object`.

- `discriminator` (string-only) - marks the field whose literal value selects
  the active variant (e.g., the `kind` field). Exactly one per struct.
- `variant:NAME` - marks an optional payload field as the payload for the
  variant selected when the discriminator equals `NAME`.

**Shape:**

A discriminated-union struct has one `discriminator` field plus one or more
`variant:NAME` payload fields. Payload fields are typically pointers with
`omitempty`, since a Go struct expresses "present only for this variant" as an
optional field.

**Validation:**

- `discriminator` on a non-string field → **error**
- `variant:` with an empty name → **error**
- More than one `discriminator` field in a struct → **error**
- `variant` fields present but no `discriminator` field → **error**
- A `discriminator` field but no `variant` fields → **error**
- Duplicate variant values → **error**

**Emission:**

Each variant emits an object of the discriminator literal plus its payload:

```
z.object({ kind: z.literal('edge'), edge: EdgeSpecSchema })
```

The payload is emitted as **required** inside its variant object even when the
Go field is a pointer with `omitempty`: within the `edge` variant the payload is
present by definition, so forcing required keeps the union narrow.

**Example:**

```go
// WiringElement is exactly one of edge, switch, or join.
type WiringElement struct {
    Kind   string      `json:"kind" gt:"discriminator"`
    Edge   *EdgeSpec   `json:"edge,omitempty"   gt:"variant:edge"`
    Switch *SwitchSpec `json:"switch,omitempty" gt:"variant:switch"`
    Join   *JoinSpec   `json:"join,omitempty"   gt:"variant:join"`
}
```

Emits:

```typescript
export const WiringElementSchema = z.discriminatedUnion('kind', [
  z.object({ kind: z.literal('edge'),   edge: EdgeSpecSchema }),
  z.object({ kind: z.literal('switch'), switch: SwitchSpecSchema }),
  z.object({ kind: z.literal('join'),   join: JoinSpecSchema })
])
```

### Discriminated Unions

A discriminated union (tagged one-of) is expressed as a Go struct with a
discriminator field plus one optional payload field per variant.

- `discriminator` (string-only) - marks the field whose literal value selects
  the active variant. Its JSON name becomes the discriminator key.
- `variant:<name>` - marks a payload field belonging to the variant selected
  when the discriminator equals `<name>`.

A struct is treated as a discriminated union when it has a `discriminator`
field. Each variant object is emitted as the discriminator literal plus its
payload, wrapped in `z.discriminatedUnion`.

The payload is emitted as **required** inside its variant, even when the Go
field is a pointer or has `omitempty`: a `*EdgeSpec` with `omitempty` is how a
Go struct expresses "present only for the edge variant", not a genuinely
optional field.

**Validation:**

- `discriminator` applies only to string fields → **error** otherwise
- `variant:<name>` must name a non-empty value
- Exactly one `discriminator` field per struct → **error** on more than one
- At least one `variant` field is required when a `discriminator` is present,
  and a `discriminator` is required when any `variant` field is present
- Variant values must be unique within a struct → **error** on duplicates

**Example:**

```go
// WiringElement is exactly one of edge, switch, or join.
type WiringElement struct {
    Kind   string      `json:"kind" gt:"discriminator"`
    Edge   *EdgeSpec   `json:"edge,omitempty" gt:"variant:edge"`
    Switch *SwitchSpec `json:"switch,omitempty" gt:"variant:switch"`
    Join   *JoinSpec   `json:"join,omitempty" gt:"variant:join"`
}
```

Emits:

```typescript
export const WiringElementSchema = z.discriminatedUnion('kind', [
  z.object({ kind: z.literal('edge'), edge: EdgeSpecSchema }),
  z.object({ kind: z.literal('switch'), switch: SwitchSpecSchema }),
  z.object({ kind: z.literal('join'), join: JoinSpecSchema })
])
```

A variant may carry no payload (discriminator literal only); omit the
`variant` payload field and use a two- or more-variant union.

### Formats (string-only)

- `email` - email address
- `uuid` - UUID string
- `url` - URL
- `date` - ISO 8601 date (YYYY-MM-DD)
- `datetime` - ISO 8601 datetime
- `ipv4` - IPv4 address
- `ipv6` - IPv6 address

**Conflicts:**

- Multiple formats on one field → **error**
- Format on non-string → **error**

## Zod Mappings

| Token      | Zod Output                               |
|------------|------------------------------------------|
| `email`    | `z.string().email()`                     |
| `uuid`     | `z.string().uuid()`                      |
| `url`      | `z.string().url()`                       |
| `datetime` | `z.string().datetime()`                  |
| `date`     | `z.string().regex(/^\d{4}-\d{2}-\d{2}$/)` |
| `ipv4`     | `z.string().ip({ version: 'v4' })`       |
| `ipv6`     | `z.string().ip({ version: 'v6' })`       |
| `enum:a,b` | `z.enum(['a', 'b'])`                     |
| `discriminator` + `variant:x` | `z.discriminatedUnion('kind', [ z.object({ kind: z.literal('x'), ... }) ])` |

## Parsing Table

| Token form      | Parsed as | Applies to | IR field(s)                          | Notes                     |
|-----------------|-----------|------------|--------------------------------------|---------------------------|
| `required`      | flag      | any        | affects `Field.Optional`             | conflicts with `optional` |
| `optional`      | flag      | any        | affects `Field.Optional`             | conflicts with `required` |
| `min:N`         | kv        | numeric, array | `Rules.Min`                      | error if non-array/non-numeric type |
| `max:N`         | kv        | numeric, array | `Rules.Max`                      | error if non-array/non-numeric type |
| `len:M..N`      | kv        | string     | `Rules.MinLength`, `Rules.MaxLength` | parse `..` range          |
| `pattern:REGEX` | kv        | string     | `Rules.Pattern`                      | store raw                 |
| `enum:a,b,c`    | kv        | string     | `Rules.Enum`                         | comma-separated values    |
| `discriminator` | flag      | string     | `Rules.IsDiscriminator`, `Schema.Discriminator` | one per struct |
| `variant:NAME`  | kv        | any        | `Rules.Variant`, `Schema.Discriminator` | payload for variant NAME  |
| `email`         | flag      | string     | `Rules.Format=FormatEmail`           | only one format           |
| `uuid`          | flag      | string     | `Rules.Format=FormatUUID`            | only one format           |
| `url`           | flag      | string     | `Rules.Format=FormatURL`             | only one format           |
| `date`          | flag      | string     | `Rules.Format=FormatDate`            | only one format           |
| `datetime`      | flag      | string     | `Rules.Format=FormatDateTime`        | only one format           |
| `ipv4`          | flag      | string     | `Rules.Format=FormatIPv4`            | only one format           |
| `ipv6`          | flag      | string     | `Rules.Format=FormatIPv6`            | only one format           |

## Unknown Token Policy

**v0.1 behavior: Unknown token → error**

This prevents silent drift and makes goldenthread feel "compiler-like". Unknown tokens and unknown keys in `key:value` pairs will fail generation with a clear error message.

## Fallback Tags

When a struct field has no `gt:` tag, goldenthread checks fallback tag keys
in order. The default fallback is `validate:` (compatible with
[go-playground/validator](https://github.com/go-playground/validator)).

Fallback tag parsing uses the same grammar as `gt:` tags. This allows
gradual adoption: existing `validate:` annotations are recognized without
requiring migration.

**Default fallback tags:** `validate`

**Example:**

```go
type User struct {
    // No gt: tag — falls back to validate: tag
    Username string `validate:"required,min=3,max=20"`
}
```

> **Note:** Fallback tags are parsed using goldenthread's tag grammar, not
> go-playground/validator syntax. Token names must match gt: token names.
> `min=3` (validator style) is not the same as `min:3` (goldenthread style).

## Examples

### Valid

```go
type User struct {
    // Required by default
    Username string `json:"username" gt:"len:3..20"`
    
    // Optional via pointer
    Email *string `json:"email" gt:"email"`
    
    // Optional via omitempty
    Age int `json:"age,omitempty" gt:"min:13,max:130"`
    
    // Explicit optional
    Bio string `json:"bio" gt:"optional,len:0..500"`
    
    // Required overrides pointer (with warning)
    ID *string `json:"id" gt:"required,uuid"`
}
```

### Invalid (errors)

```go
type Bad struct {
    // Error: both required and optional
    Field1 string `gt:"required,optional"`
    
    // Error: min > max
    Field2 int `gt:"min:10,max:5"`
    
    // Error: multiple formats
    Field3 string `gt:"email,uuid"`
    
    // Error: format on non-string
    Field4 int `gt:"email"`
    
    // Error: unknown token
    Field5 string `gt:"foobar"`
}
```
