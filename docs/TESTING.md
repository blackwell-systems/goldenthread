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

**Schedule**: Every 30 minutes + on every push to main + on every PR

**Strategy**: 
- Runs 10 fuzz targets in parallel
- Each runs for 5-10 minutes
- Total: ~50-100 minutes of fuzzing per run
- Corpus is cached and grows over time

**Why This Approach Works**:

Continuous fuzzing is fundamentally different from one-time testing. Here's why it's effective:

**1. Corpus Evolution**
- Each run adds "interesting" inputs to the corpus (inputs that increase coverage)
- Next run uses previous corpus as starting point + explores new mutations
- Over weeks/months, corpus becomes highly optimized for finding bugs
- Like compound interest: each run builds on all previous runs

**Example progression**:
- Week 1: Discovers basic edge cases (empty strings, max values)
- Week 2: Mutates edge cases, finds combinations (empty + special chars)
- Week 3: Deeper mutations find rare paths (nested edge cases)
- Month 3: Corpus has 1000s of interesting cases covering obscure paths

**2. Coverage-Guided Exploration**
- Go's fuzzer instruments code to track which branches execute
- Prioritizes inputs that explore new code paths
- Automatically finds rare conditions (e.g., "if len == 42 && first_char == '🎉'")
- No human could think of these combinations

**3. Time Advantage**
- Humans write ~10-20 test cases per feature
- Fuzzer executes 100K-1M cases per minute
- Per CI run: ~20M-40M test cases executed (10 minutes)
- Per day: 960M-1.9B test cases (48 runs every 30 minutes)
- Per month: 28.8B-57.6B test cases

**4. Zero Maintenance**
- No test cases to write for new code paths
- Automatically explores new features added to codebase
- Corpus naturally adapts to code changes
- Only action needed: fix bugs when found

**Real Results from goldenthread**:
- **180 executions** to find regex escaping bug (< 1 second)
- **444,553 executions** to find UTF-8 camelCase bug (< 10 seconds)
- Traditional testing would never find these (who tests newlines in regex patterns?)

**Execution matrix**:
| Package | Target | Duration | Execs/Run (est) |
|---------|--------|----------|-----------------|
| emitter/zod | FuzzEmit | 10m | ~2-5M |
| emitter/zod | FuzzEmitPattern | 10m | ~500K-1M |
| emitter/zod | FuzzEmitFieldName | 5m | ~1-2M |
| emitter/zod | FuzzEmitValidation | 5m | ~1-2M |
| emitter/zod | FuzzEmitEnum | 5m | ~500K-1M |
| hash | FuzzComputeSchemaHash | 10m | ~5-10M |
| hash | FuzzComputeSchemaHash_Stability | 5m | ~2-5M |
| hash | FuzzComputeSchemaHash_TypeChanges | 5m | ~2-5M |
| hash | FuzzComputeSchemaHash_FieldOrder | 5m | ~2-5M |
| parser | FuzzParsePackages | 5m | ~50-100 |

**Total per CI run**: ~20-40 million test cases (parser is slow due to go/packages)

**Benefits of CI fuzzing**:
1. **Corpus growth**: Each run discovers new interesting inputs and caches them
2. **Regression prevention**: Finds bugs in new code before merging to main
3. **Coverage expansion**: Explores code paths developers never consider
4. **Zero developer effort**: Runs automatically every 30 minutes, 48x per day
5. **Compound returns**: Gets more effective over time as corpus grows
6. **Bug discovery timeline**: Finds bugs in minutes-hours instead of months/years in production
7. **PR validation**: Runs on every pull request before merge

**Why 30-minute intervals?**:
- Aggressive bug discovery (< 30 minute latency from code change)
- Rapid corpus growth from frequent updates
- Free for open source on GitHub Actions
- Catches bugs before developers even switch tasks
- 48 runs/day = maximum practical fuzzing frequency
- Each run completes in ~10 minutes (parallel execution)

**GitHub Actions Limits** (for public repositories):
- Concurrent jobs: 20 (we use 10 per run)
- Minutes per month: Unlimited for public repos
- Storage: 500MB artifacts (we cache small corpus files)
- Perfect fit for aggressive fuzzing strategy

### How You'll Know When Fuzzing Finds a Bug

**Automatic notifications** when fuzzing discovers issues:

**1. GitHub Issue Created Automatically**
- Title: "Fuzzing found bug in [FuzzTestName]"
- Labels: `bug`, `fuzzing`, `automated`
- Contains:
  - Exact command to reproduce locally
  - Link to failing test case artifact
  - Full fuzz output (last 100 lines)
  - Workflow run link

**2. GitHub Actions Failure**
- Workflow status turns red
- Email notification if you watch the repo
- Shows in GitHub UI notifications
- Blocks merging if on PR

**3. Workflow Summary**
- Clear PASS or FAIL status per target
- Execution statistics
- Direct links to artifacts

**Example notification flow**:
1. Fuzzing runs every 30 minutes
2. Bug discovered in `FuzzEmitPattern`
3. **GitHub issue created** with title "Fuzzing found bug in FuzzEmitPattern"
4. **Email sent** (if watching repo)
5. **Artifact uploaded** with failing test case
6. You see issue in your notifications
7. Click issue → get reproduction command
8. Fix bug → update FUZZING_BUGS.md
9. Close issue

**Manual checking** (if you don't enable notifications):
- Visit: https://github.com/[your-repo]/actions
- Check "Continuous Fuzzing" workflow status
- Red = bug found, Green = all clear

**Recommended: Enable watch notifications**
- Go to repository settings
- Click "Watch" → "All Activity"
- Enables email on CI failures
- Get notified within minutes of bug discovery

**Alternative: OSS-Fuzz**
For even more coverage, consider integrating with [OSS-Fuzz](https://github.com/google/oss-fuzz):
- 24/7 continuous fuzzing on Google infrastructure
- Free for open source projects
- Automatic bug reporting
- Corpus shared with community
- Used by major projects (Go stdlib, Chrome, LLVM, etc.)

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
