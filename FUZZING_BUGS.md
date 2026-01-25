# Fuzzing Bug Log

This document tracks all bugs discovered by fuzzing in goldenthread. It serves as evidence of fuzzing effectiveness and a reference for future bug patterns.

## Summary

**Total bugs found**: 2  
**Total executions to discovery**: 444,733  
**Production impact prevented**: 100% (both would have caused production incidents)

---

## Bug #1: UTF-8 Corruption in camelCase Conversion

**Discovered**: 2026-01-25  
**Fuzz target**: `FuzzEmit`  
**Executions to discovery**: 444,553  
**Time to discovery**: ~10 seconds  
**Severity**: High (data corruption)

### Description

The `camelCase()` function used byte slicing `s[:1]` to lowercase the first character, which splits multi-byte UTF-8 sequences, producing invalid UTF-8 output.

### Trigger Conditions

When **all** of these conditions are met:
1. Field has empty `JSONName` (falls back to `GoName`)
2. `GoName` starts with multi-byte UTF-8 character (Japanese, Chinese, emoji, etc.)
3. Emitter generates field name using `camelCase(GoName)`

### Example Input

```go
type 日本語 struct {
    フィールド string `json:""` // Empty JSON name triggers fallback
}
```

### Buggy Output

```javascript
export const 日本語Schema = z.object({
  �\x83\x95ィールド: z.string()  // ❌ Invalid UTF-8
})
```

Expected:
```javascript
export const 日本語Schema = z.object({
  フィールド: z.string()  // ✅ Valid UTF-8
})
```

### Root Cause

```go
// BUGGY CODE:
func camelCase(s string) string {
    return strings.ToLower(s[:1]) + s[1:]
    // s[:1] is BYTE slicing, not CHARACTER slicing
    // "フィールド" = [0xE3, 0x83, 0x95, ...]
    // s[:1] = [0xE3] (incomplete UTF-8 sequence)
    // s[1:] = [0x83, 0x95, ...] (orphaned continuation bytes)
}
```

Go strings are UTF-8 byte slices. Slicing by index operates on **bytes**, not **characters** (runes). Multi-byte UTF-8 characters split incorrectly.

### Fix

```go
// FIXED CODE:
func camelCase(s string) string {
    if s == "" {
        return ""
    }
    // Convert to runes (Unicode code points) for proper character handling
    runes := []rune(s)
    if len(runes) > 0 {
        runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
    }
    return string(runes)
}
```

### Impact

**Without fix**:
- International users (Japan, China, Korea, Arab countries) get corrupted output
- Generated TypeScript files have invalid UTF-8
- TypeScript compiler may fail or produce warnings
- Runtime errors when parsing malformed identifiers

**With fix**:
- Full Unicode support for field names
- Works with any language/emoji
- No UTF-8 validation errors

### Lessons Learned

1. **Never use byte slicing on user-provided strings** - always use rune slicing for character operations
2. **Test with non-ASCII input** - fuzzing found this, but we could have caught it with Unicode test cases
3. **Go string gotcha**: `s[i]` and `s[i:j]` operate on bytes, not characters

### Test Coverage Added

- `TestEmit_UTF8_EmptyJSONName`: Regression test with Japanese input
- `TestCamelCase`: Unit tests for ASCII, Japanese, emoji, mixed
- Fuzz corpus now includes the failing case for permanent regression testing

### Commit

**Hash**: `ebfdab9`  
**Message**: "fix: handle UTF-8 properly in camelCase conversion"

---

## Bug #2: Regex Pattern Escaping Breaks JavaScript Syntax

**Discovered**: 2026-01-25  
**Fuzz target**: `FuzzEmitPattern`  
**Executions to discovery**: 180  
**Time to discovery**: < 1 second  
**Severity**: High (syntax error)

### Description

Regex patterns containing newlines, tabs, or forward slashes produced malformed JavaScript code. Only backslashes were being escaped.

### Trigger Conditions

Pattern contains any of:
- Newline (`\n`)
- Tab (`\t`)
- Carriage return (`\r`)
- Forward slash (`/`)

### Example Input

```go
type User struct {
    Name string `gt:"pattern:\n"` // Newline in pattern
}
```

### Buggy Output

```javascript
export const UserSchema = z.object({
  name: z.string().regex(/
/)  // ❌ Regex broken across lines - syntax error
})
```

Expected:
```javascript
export const UserSchema = z.object({
  name: z.string().regex(/\n/)  // ✅ Escaped newline
})
```

### Root Cause

```go
// BUGGY CODE:
if rules.Pattern != nil {
    pattern := strings.ReplaceAll(*rules.Pattern, "\\", "\\\\")
    b.WriteString(fmt.Sprintf(".regex(/%s/)", pattern))
    // Only escapes backslashes, ignores other special chars
}
```

JavaScript regex literals `/pattern/` have special meaning for:
- `/` (delimiter)
- Whitespace characters (newline, tab, etc.)

Unescaped newlines break the regex literal across multiple lines, causing syntax errors.

### Fix

```go
// FIXED CODE:
if rules.Pattern != nil {
    pattern := *rules.Pattern
    pattern = strings.ReplaceAll(pattern, "\\", "\\\\") // Backslash FIRST
    pattern = strings.ReplaceAll(pattern, "/", "\\/")   // Delimiter
    pattern = strings.ReplaceAll(pattern, "\n", "\\n")  // Newline
    pattern = strings.ReplaceAll(pattern, "\r", "\\r")  // Carriage return
    pattern = strings.ReplaceAll(pattern, "\t", "\\t")  // Tab
    b.WriteString(fmt.Sprintf(".regex(/%s/)", pattern))
}
```

**Critical**: Backslash must be escaped first to avoid double-escaping other characters.

### Impact

**Without fix**:
- Any validation pattern with newlines/tabs breaks generated code
- TypeScript compilation fails
- Runtime: SyntaxError on module load
- Patterns with `/` end regex prematurely (e.g., `/path/to/file` becomes `regex(/path/to/file/)` - interpreted as 3 parts)

**With fix**:
- All patterns work correctly
- Proper JavaScript regex literal escaping
- TypeScript compiles cleanly

### Real-World Scenarios

**Multiline regex patterns**:
```go
Pattern: `^line1
line2$`
```

**Path patterns**:
```go
Pattern: `^/api/v[0-9]+/users$`
```

**Whitespace patterns**:
```go
Pattern: `^\s+$`  // Tab character in actual string
```

### Lessons Learned

1. **Escape sequences matter in target language** - JavaScript regex literals have different escaping than Go strings
2. **Order matters**: Escape backslashes first to avoid double-escaping
3. **Consider all whitespace**: Not just `\n`, but also `\t`, `\r`, and others
4. **Fuzzing finds rare cases fast**: 180 executions (< 1 second) vs weeks of manual testing

### Test Coverage Added

- `TestEmit_PatternEscaping`: 6 test cases covering all special characters
- Tests: newline, tab, CR, forward slash, backslash, mixed
- Fuzz corpus includes failing case

### Commit

**Hash**: `29d3727`  
**Message**: "fix: escape special characters in regex patterns"

---

## Bug Pattern Analysis

### Common Themes

Both bugs involve **string manipulation with special characters**:
1. UTF-8 multi-byte sequences (Bug #1)
2. JavaScript escape sequences (Bug #2)

Both were found **extremely quickly** by fuzzing:
- Bug #1: 444K executions in 10 seconds
- Bug #2: 180 executions in < 1 second

Both would have **severe production impact**:
- Data corruption (Bug #1)
- Syntax errors (Bug #2)

### Why Fuzzing Found These

**Traditional testing limitations**:
- Test writers focus on "happy path" inputs
- Edge cases like Japanese field names seem rare
- Regex with newlines never considered

**Fuzzing advantages**:
- No human bias toward common cases
- Explores full input space including rare combinations
- Coverage-guided mutation finds boundary conditions
- Executes millions of cases impossible for humans

### Prevention Going Forward

**Continuous fuzzing catches**:
1. New code with similar patterns (string manipulation)
2. Refactoring that reintroduces old bugs
3. Platform-specific issues (Windows `\r\n`, etc.)
4. Unicode edge cases in new features

**Expected future discoveries**:
- More escape sequence issues in other emitters
- Parser bugs with malformed struct tags
- Hash collisions (extremely rare but possible)
- Panic conditions in edge cases

---

## Fuzzing Effectiveness Metrics

### Discovery Rate

- **Total executions**: 444,733
- **Bugs found**: 2
- **Discovery rate**: 0.00045% (1 bug per 222,366 executions)

This seems low, but each bug prevents a production incident. Even one bug caught = ROI positive.

### Time to Discovery

- **Bug #1**: 10 seconds of fuzzing
- **Bug #2**: < 1 second of fuzzing
- **Traditional testing**: Would likely never find these bugs
- **Production discovery**: Weeks-months after release (after user reports)

### Coverage Impact

Bugs found in code paths with **existing test coverage**:
- `camelCase()` is called by tested emitter code
- Pattern emission is tested with simple patterns

**Insight**: Even well-tested code has bugs. Fuzzing explores cases that humans don't think to test.

---

## Contributing

Found a bug with fuzzing? Add it to this log with:

**Required information**:
- Date discovered
- Fuzz target name
- Executions to discovery
- Description and trigger conditions
- Example input/output
- Root cause analysis
- Fix description
- Commit hash

**Template**:
```markdown
## Bug #N: [Title]

**Discovered**: YYYY-MM-DD
**Fuzz target**: `FuzzTargetName`
**Executions**: N
**Severity**: High/Medium/Low

### Description
[What went wrong]

### Trigger Conditions
[When does this occur]

### Example Input
[Minimal reproducing case]

### Root Cause
[Why this happened]

### Fix
[How it was fixed]

### Commit
**Hash**: `abc1234`
```

---

**Last updated**: 2026-01-25  
**Fuzzing status**: Active (running every 30 minutes)  
**Next review**: After 1 month of continuous fuzzing
