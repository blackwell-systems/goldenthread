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

## Parsing Table

| Token form      | Parsed as | Applies to | IR field(s)                          | Notes                     |
|-----------------|-----------|------------|--------------------------------------|---------------------------|
| `required`      | flag      | any        | affects `Field.Optional`             | conflicts with `optional` |
| `optional`      | flag      | any        | affects `Field.Optional`             | conflicts with `required` |
| `min:N`         | kv        | numeric    | `Rules.Min`                          | error if non-numeric type |
| `max:N`         | kv        | numeric    | `Rules.Max`                          | error if non-numeric type |
| `len:M..N`      | kv        | string     | `Rules.MinLength`, `Rules.MaxLength` | parse `..` range          |
| `pattern:REGEX` | kv        | string     | `Rules.Pattern`                      | store raw                 |
| `enum:a,b,c`    | kv        | string     | `Rules.Enum`                         | comma-separated values    |
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
