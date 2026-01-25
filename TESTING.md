# Testing Strategy

This document describes goldenthread's comprehensive testing approach, with emphasis on continuous fuzzing.

## Test Coverage Overview

Current coverage: **53.4%** overall
- **Emitter**: 84.8% (critical code generation paths)
- **Parser**: 71.7% (Go source parsing)
- **Hash**: 47.6% (deterministic drift detection)
- **Normalize**: 96.4% (embedded struct flattening)

## Testing Layers

### 1. Unit Tests
**Purpose**: Verify individual function correctness with known inputs.

**Location**: `*_test.go` files alongside source code

**Examples**:
- `emitter_test.go`: 14 test functions, 51 test cases
- `hash_test.go`: 8 test functions covering determinism
- `flatten_test.go`: 11 test functions for embedding logic

**Run**:
```bash
go test ./...
```

### 2. Integration Tests
**Purpose**: Test components working together with real Go packages.

**Key tests**:
- `parser/integration_test.go`: End-to-end parsing with `go/packages`
- Creates temporary Go modules to test realistic scenarios
- Tests: struct parsing, validation, enums, arrays, maps

**Run**:
```bash
go test ./internal/parser -run Integration
```

### 3. Fuzz Testing
**Purpose**: Discover edge cases, crashes, and bugs through randomized input generation.

This is our primary defense against production bugs in untrusted input handling.

## Fuzzing Deep Dive

### What is Fuzzing?

Fuzzing automatically generates millions of random test inputs to find:
- **Crashes and panics**: Unhandled edge cases
- **Invalid output**: UTF-8 corruption, syntax errors
- **Security issues**: Code injection, infinite loops
- **Logic errors**: Incorrect behavior on boundary values

### Why Fuzzing?

Traditional unit tests check **expected** inputs. Fuzzing explores **unexpected** inputs.

**Real bugs found by fuzzing**:
1. **UTF-8 corruption**: Japanese field names with empty JSON names produced invalid UTF-8
2. **Regex escaping**: Newlines in patterns broke JavaScript syntax across multiple lines

These bugs would never have been caught by human-written tests.

### Fuzz Test Targets

#### Emitter Fuzzing (`internal/emitter/zod/fuzz_test.go`)
```go
FuzzEmit                 // Random schema configurations
FuzzEmitFieldName        // Field name edge cases
FuzzEmitValidation       // Numeric boundaries
FuzzEmitPattern          // Regex pattern escaping
FuzzEmitEnum             // Enum value generation
```

**Why**: Emitter generates code executed in production. Malformed output could break applications.

#### Parser Fuzzing (`internal/parser/fuzz_test.go`)
```go
FuzzParsePackages        // Random Go struct definitions with gt: tags
FuzzNormalizeDoc         // Documentation handling edge cases
```

**Why**: Parser handles untrusted input (user Go code). Must never panic.

#### Hash Fuzzing (`internal/hash/fuzz_test.go`)
```go
FuzzComputeSchemaHash              // Determinism verification
FuzzComputeSchemaHash_Stability    // Doc/position shouldn't affect hash
FuzzComputeSchemaHash_TypeChanges  // Type changes MUST affect hash
FuzzComputeSchemaHash_FieldOrder   // Field order independence
FuzzComputeSchemaHash_OptionalChange // Optional flag detection
```

**Why**: Hash must be deterministic for CI drift detection. Even subtle bugs break the entire feature.

### Running Fuzz Tests Locally

**Quick test (10 seconds)**:
```bash
go test ./internal/emitter/zod -fuzz=FuzzEmit -fuzztime=10s
```

**Longer run (5 minutes)**:
```bash
go test ./internal/emitter/zod -fuzz=FuzzEmit -fuzztime=5m
```

**Run all fuzz targets** (use with `parallel` or manually):
```bash
# Emitter
go test ./internal/emitter/zod -fuzz=FuzzEmit -fuzztime=1m
go test ./internal/emitter/zod -fuzz=FuzzEmitPattern -fuzztime=1m
go test ./internal/emitter/zod -fuzz=FuzzEmitEnum -fuzztime=1m

# Hash
go test ./internal/hash -fuzz='^FuzzComputeSchemaHash$' -fuzztime=1m

# Parser (slow due to go/packages overhead)
go test ./internal/parser -fuzz=FuzzParsePackages -fuzztime=30s
```

**Re-run a specific failing case**:
```bash
go test ./internal/emitter/zod -run=FuzzEmit/89831cc049267b2c
```

### Continuous Fuzzing in CI

**Workflow**: `.github/workflows/fuzz.yml`

**Schedule**: Every 6 hours + on every push to main

**Strategy**: 
- Runs 10 fuzz targets in parallel
- Each runs for 5-10 minutes
- Total: ~50-100 minutes of fuzzing per run
- Corpus is cached and grows over time

**Execution matrix**:
| Package | Target | Duration |
|---------|--------|----------|
| emitter/zod | FuzzEmit | 10m |
| emitter/zod | FuzzEmitPattern | 10m |
| emitter/zod | FuzzEmitFieldName | 5m |
| emitter/zod | FuzzEmitValidation | 5m |
| emitter/zod | FuzzEmitEnum | 5m |
| hash | FuzzComputeSchemaHash | 10m |
| hash | FuzzComputeSchemaHash_Stability | 5m |
| hash | FuzzComputeSchemaHash_TypeChanges | 5m |
| hash | FuzzComputeSchemaHash_FieldOrder | 5m |
| parser | FuzzParsePackages | 5m |

**Benefits of CI fuzzing**:
1. **Corpus growth**: Each run discovers new interesting inputs
2. **Regression prevention**: Finds bugs before they reach production
3. **Coverage expansion**: Explores code paths humans miss
4. **Zero developer effort**: Runs automatically

### Understanding Fuzz Output

```
fuzz: elapsed: 3s, execs: 1021405 (340350/sec), new interesting: 1 (total: 63)
```

- **execs**: Number of test cases executed
- **execs/sec**: Execution rate (higher is better)
- **new interesting**: Inputs that increased code coverage
- **total**: Cumulative interesting inputs in corpus

**Good signs**:
- High exec/sec (>100k for simple functions)
- Low "new interesting" after initial spike (coverage plateau)
- No failures

**Warning signs**:
- Exec/sec drops to near zero (infinite loop or hang)
- Continuous "new interesting" discovery (unexplored code paths)
- Test failures

### Fuzz Corpus Management

**Location**: `**/testdata/fuzz/FuzzTestName/`

**Format**: Go native fuzz corpus (binary + text format)

**Growth**: Corpus grows as fuzzing discovers interesting inputs

**Cleanup**: Old corpus entries can be deleted safely - fuzzing will rediscover them

**Version control**: 
- ✅ Commit failure cases for regression tests
- ❌ Don't commit entire corpus (too large)

### Writing New Fuzz Tests

**Template**:
```go
func FuzzMyFunction(f *testing.F) {
    // Seed corpus with known interesting cases
    f.Add("valid input")
    f.Add("")
    f.Add("edge case")
    
    f.Fuzz(func(t *testing.T, input string) {
        // Skip invalid inputs if needed
        if !utf8.ValidString(input) {
            return
        }
        
        // Panic protection
        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Panic: %v", r)
            }
        }()
        
        // Call function under test
        result := MyFunction(input)
        
        // Validate output properties
        if !utf8.ValidString(result) {
            t.Error("Invalid UTF-8 output")
        }
    })
}
```

**Best practices**:
1. **Seed corpus**: Provide 5-10 diverse inputs to bootstrap
2. **Input validation**: Skip truly invalid inputs (but be minimal)
3. **Panic recovery**: Catch panics and report them
4. **Property testing**: Check output properties, not exact values
5. **Fast execution**: Keep fuzz target fast (<1ms) for high throughput

### Debugging Fuzz Failures

**Step 1: Reproduce locally**
```bash
go test ./internal/emitter/zod -run=FuzzEmit/89831cc049267b2c -v
```

**Step 2: Examine the failing input**
```bash
cat internal/emitter/zod/testdata/fuzz/FuzzEmit/89831cc049267b2c
```

**Step 3: Create a regression test**
```go
func TestEmit_FuzzRegression_Issue123(t *testing.T) {
    // Exact input that caused the failure
    input := "\n"
    
    // Test should now pass with fix
    result := MyFunction(input)
    
    // Assert expected behavior
    assert.Valid(t, result)
}
```

**Step 4: Fix the bug**

**Step 5: Verify fix**
```bash
go test ./internal/emitter/zod -run=FuzzEmit/89831cc049267b2c
go test ./internal/emitter/zod -fuzz=FuzzEmit -fuzztime=30s
```

## Test Execution Times

**Unit tests**: ~10ms total (very fast)
**Integration tests**: ~5s (go/packages overhead)
**Fuzz tests**: Configurable (10s - infinity)

**Recommendation**: 
- Run unit tests on every save (via IDE)
- Run integration tests pre-commit
- Run fuzz tests for 30s locally before PR
- Let CI run extended fuzzing continuously

## CI Test Matrix

### Pull Request Checks
```yaml
- Unit tests: all packages
- Integration tests: all packages
- Fuzz tests: 30s per target (smoke test)
- Coverage report
```

### Scheduled Fuzzing
```yaml
- Every 6 hours
- 5-10 minutes per target
- 10 parallel targets
- Total: ~60 minutes fuzzing per run
```

### On Main Branch
```yaml
- Full test suite
- Extended fuzzing (10m per target)
- Coverage tracking
```

## Measuring Success

**Coverage metrics**:
- Track coverage trends over time
- Target: >80% for critical paths (emitter, parser, hash)
- Current: 53.4% overall (excellent for v0.1)

**Fuzz metrics**:
- Corpus size growth
- New bugs discovered per month
- Executions per CI run
- Coverage increase from fuzzing

**Bug prevention**:
- Zero known panics in production
- Zero invalid UTF-8 output reports
- Zero code injection vulnerabilities

## Future Testing Additions

**Planned**:
- [ ] End-to-end CLI tests
- [ ] Performance benchmarks
- [ ] Mutation testing
- [ ] Property-based testing (gopter)
- [ ] Fuzz testing for normalize package

**Wishlist**:
- OSS-Fuzz integration for 24/7 fuzzing
- Differential testing against other schema tools
- Chaos engineering for concurrent operations

## Contributing

When adding new features:
1. ✅ Add unit tests for happy path
2. ✅ Add integration test if cross-component
3. ✅ Add fuzz test if handling external input
4. ✅ Update this document if adding new test types

**Questions?** Check the [test files](./internal/) for examples or ask in issues.

---

**Remember**: Fuzzing is not a replacement for thoughtful test design. It's a complement that finds the bugs you didn't think to test.
