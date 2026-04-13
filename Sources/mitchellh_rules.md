# Go Testing Rules
*Derived from Mitchell Hashimoto's "Advanced Testing with Go"*

These rules guide how to write tests and how to structure production code so it
can be tested well. Both are required — good test technique applied to
untestable code produces poor results.

---

## Part One: Test Methodology

### Table-Driven Tests

Always use table-driven tests. Set up the table structure even for a single case
if the function could plausibly grow. It is low overhead to add a row; it is
high overhead to retrofit the structure later.

```go
func TestAdd(t *testing.T) {
    cases := []struct{ A, B, Expected int }{
        {1, 1, 2},
        {1, -1, 0},
        {1, 0, 1},
        {0, 0, 0},
    }
    for _, tc := range cases {
        actual := tc.A + tc.B
        if actual != tc.Expected {
            t.Errorf("%d + %d = %d, expected %d",
                tc.A, tc.B, actual, tc.Expected)
        }
    }
}
```

**Name every case.** Use a `Name` field in the struct or a `map[string]struct{...}`
so that failure output identifies the case immediately. Never rely on array indices.

```go
cases := map[string]struct{ A, B, Expected int }{
    "positive":    {1, 1, 2},
    "negative":    {-1, -2, -3},
    "mixed":       {1, -1, 0},
    "both zero":   {0, 0, 0},
}
for name, tc := range cases {
    t.Run(name, func(t *testing.T) {
        actual := tc.A + tc.B
        if actual != tc.Expected {
            t.Errorf("expected %d, got %d", tc.Expected, actual)
        }
    })
}
```

**Do Not:**
- Use `t.Errorf("case %d failed", i)` — index numbers are useless in failure output
- Duplicate test logic across separate `TestFoo_case1`, `TestFoo_case2` functions
  when a table would work

### Subtests

Use `t.Run` to wrap each table case. Each subtest is a closure, so `defer` works
correctly within it and resources are released after each case rather than
accumulating until the parent test ends.

```go
for _, tc := range cases {
    tc := tc // capture loop variable
    t.Run(tc.Name, func(t *testing.T) {
        // defer works correctly here
        f, cleanup := testTempFile(t)
        defer cleanup()
        // …
    })
}
```

Subtests can be targeted individually: `go test -run TestFoo/positive`.

### Test Fixtures

Store test data in a `test-fixtures/` directory alongside the test file. Use
relative paths — `go test` always sets the working directory to the package
directory being tested, so relative paths work correctly regardless of where
`go test` is invoked.

```go
func TestParseConfig(t *testing.T) {
    data := filepath.Join("test-fixtures", "valid_config.hcl")
    f, err := os.Open(data)
    if err != nil {
        t.Fatalf("failed to open fixture: %s", err)
    }
    defer f.Close()
    // …
}
```

**Do Not:**
- Use `runtime.Caller` or `os.Getwd` to compute absolute fixture paths
- Embed large binary fixtures as base64 strings in test source

### Golden Files

Use golden files to test complex output (formatted text, serialized structs,
generated code). Store the expected output in `test-fixtures/<name>.golden` and
provide an `-update` flag to regenerate them.

```go
var update = flag.Bool("update", false, "update golden files")

func TestFormat(t *testing.T) {
    cases := []struct{ Name, Input string }{
        {"basic", "input.hcl"},
        {"empty", "empty.hcl"},
    }
    for _, tc := range cases {
        t.Run(tc.Name, func(t *testing.T) {
            input, _ := ioutil.ReadFile(filepath.Join("test-fixtures", tc.Input))
            actual := Format(input)

            golden := filepath.Join("test-fixtures", tc.Name+".golden")
            if *update {
                ioutil.WriteFile(golden, actual, 0644)
            }

            expected, _ := ioutil.ReadFile(golden)
            if !bytes.Equal(actual, expected) {
                t.Errorf("output mismatch for %s\ngot:\n%s\nwant:\n%s",
                    tc.Name, actual, expected)
            }
        })
    }
}
```

Workflow: run `go test -update` to generate the golden files, inspect them by
eye, commit them when they look correct, then run `go test` normally thereafter.

**Do Not:**
- Hardcode expected multi-line output as string constants in the test
- Commit golden files without reviewing them

### Test Helpers

**Never return an error from a test helper.** Accept `*testing.T` and call
`t.Fatalf` internally. If a helper returns an error, every call site requires an
error check, which defeats the purpose of having a helper.

```go
// Wrong — callers must check the error
func testTempFile(t *testing.T) (string, error) { … }

// Right — helper handles the error itself
func testTempFile(t *testing.T) (string, func()) {
    tf, err := ioutil.TempFile("", "test")
    if err != nil {
        t.Fatalf("testTempFile: %s", err)
    }
    tf.Close()
    return tf.Name(), func() { os.Remove(tf.Name()) }
}
```

**Always call `t.Helper()`** at the top of every test helper. This causes failure
output to point to the caller of the helper, not to the line inside the helper.

```go
func testTempFile(t *testing.T) (string, func()) {
    t.Helper()
    // …
}
```

**Return a `func()` for cleanup** instead of requiring the caller to manage
teardown separately. The closure captures `t` so it can also call `t.Fatalf` if
cleanup fails.

```go
func TestSomething(t *testing.T) {
    path, cleanup := testTempFile(t)
    defer cleanup()
    // …
}
```

When a helper has no meaningful return value (it only sets up side effects),
the cleanup can be one-lined:

```go
func testChdir(t *testing.T, dir string) func() {
    t.Helper()
    old, err := os.Getwd()
    if err != nil {
        t.Fatalf("testChdir: %s", err)
    }
    if err := os.Chdir(dir); err != nil {
        t.Fatalf("testChdir: %s", err)
    }
    return func() { os.Chdir(old) }
}

func TestThing(t *testing.T) {
    defer testChdir(t, "/tmp/testdir")()
    // …
}
```

**Do Not:**
- Return errors from test helpers
- Omit `t.Helper()` from test helpers
- Use `t.Error` (non-fatal) in helpers when `t.Fatal` is appropriate — a helper
  that continues running after a failure produces confusing output

### Timing-Dependent Tests

For tests that wait on asynchronous behavior, use a `select` with `time.After`.
Never use `time.Sleep` with a fixed duration.

```go
func TestAsyncThing(t *testing.T) {
    done := make(chan struct{})
    go doAsyncWork(done)

    timeout := 5 * time.Second * timeMultiplier
    select {
    case <-done:
    case <-time.After(timeout):
        t.Fatal("timed out waiting for async work")
    }
}
```

Define `timeMultiplier` as a package-level variable so CI environments can
increase it without changing test logic:

```go
var timeMultiplier = time.Duration(1)
```

Do not use fake time. A `timeMultiplier` is less intrusive and sufficient for
most cases.

**Do Not:**
- Use `time.Sleep` in tests
- Hard-code timeouts without a multiplier

### Parallelization

Do not use `t.Parallel()`. Parallel tests make failures ambiguous: you cannot
tell whether a failure is a logic bug or a race condition.

If you need to validate concurrent behavior, write dedicated tests using `go`
goroutines and `-race` explicitly, and run them as separate processes rather than
via `t.Parallel()`.

If you must run parallel and serial modes, run `go test -parallel=1` and
`go test -parallel=N` as separate invocations.

**Do Not:**
- Call `t.Parallel()` in standard test functions

---

## Part Two: Writing Testable Code

### Global State

Avoid global state. Tests that depend on global state may pass or fail depending
on execution order.

When global state is unavoidable, use this hierarchy:

```go
// Worst — constant cannot be changed in tests
const port = 1000

// Better — variable can be overridden in tests
var port = 1000

// Best — constant default, configurable via struct
const defaultPort = 1000

type ServerOpts struct {
    Port int // initialize to defaultPort in constructor
}
```

**Do Not:**
- Use `init()` to set global state that tests cannot override
- Use `sync.Once` to initialize global singletons that tests need to reset

### Packages and Functions

Break functionality into packages judiciously. Doing it correctly aids testing
by providing a clean, bounded surface area. Overdoing it creates import cycles
and makes tests harder to write.

**Test the exported API, not internals.** Treat unexported functions and types
as implementation details. If the exported API is correct, the internals are
correct by definition. Only unit-test an unexported function when it is
extremely complex and cannot be adequately exercised through the exported API.

```go
// Test this:
func TestServer_CreateUser(t *testing.T) { … }

// Not this (unless the function is extremely complex):
func Test_hashPassword(t *testing.T) { … }
```

Use `internal/` packages to over-package safely. Internal packages cannot be
imported by external consumers, so you can create many of them to achieve good
test boundaries without committing to a public API.

**Do Not:**
- Test unexported functions as the primary testing strategy
- Create a `utils` or `helpers` package shared across unrelated packages
- Under-package to the point where everything imports everything (creates
  import cycles that prevent future refactoring)

### Networking

Always make real network connections in tests. Do not mock `net.Conn`.

Use a helper that creates a real listener on an OS-assigned port and returns both
ends of the connection:

```go
// Error checking omitted for brevity
func testConn(t *testing.T) (client, server net.Conn) {
    t.Helper()
    ln, _ := net.Listen("tcp", "127.0.0.1:0")
    var srv net.Conn
    go func() {
        defer ln.Close()
        srv, _ = ln.Accept()
    }()
    cli, _ := net.Dial("tcp", ln.Addr().String())
    return cli, srv
}
```

This pattern works for any protocol (TCP, Unix sockets, TLS, SSH) and any
address family (IPv4, IPv6). Return the listener too if the test needs to accept
multiple connections.

**Do Not:**
- Mock `net.Conn` — there is no reason to
- Hardcode port numbers — always use `"127.0.0.1:0"` and let the OS assign one

### Configurability

Production code that has hardcoded behavior (fixed ports, fixed paths, fixed
timeouts) is difficult to test. Overparameterize structs to allow tests to
fine-tune behavior. Unexported fields are fine — only internal tests need access.

```go
// Do this even if CachePath and Port never change in production.
// Tests can use different values.
type ServerOpts struct {
    CachePath string
    Port      int
    // unexported, test-only fields are acceptable
    testSkipAuth bool
}
```

Use the unexported `test`-prefixed field pattern to bypass expensive external
dependencies in tests (OAuth flows, external API calls, etc.):

```go
func (s *Server) authenticate(r *http.Request) (User, error) {
    if s.opts.testSkipAuth {
        return User{ID: "test-user"}, nil
    }
    return s.oauthProvider.Validate(r)
}
```

**Do Not:**
- Hardcode ports, paths, or timeouts as unoverridable constants in production structs
- Use environment variables as the only way to configure test behavior

### Complex Structs

For testing equality of complex data structures, choose based on what gives the
most useful failure output:

**Option 1 — `reflect.DeepEqual`**: the blunt instrument. Works, but when it
returns false the error message often does not tell you why.

```go
if !reflect.DeepEqual(got, want) {
    t.Errorf("got %+v, want %+v", got, want)
}
```

**Option 2 — third-party diff library**: generates a human-readable diff on
mismatch. Prefer this over raw `reflect.DeepEqual` for non-trivial structs.

**Option 3 — `testString()` pattern**: for large or deeply nested structures,
add an unexported method that renders only the fields that matter to the test
as a human-readable string, then compare strings.

```go
// In the package under test (in a _test.go file):
func (g *Graph) testString() string {
    var buf bytes.Buffer
    for _, node := range g.nodes {
        fmt.Fprintf(&buf, "%s -> %v\n", node.Name, node.Deps)
    }
    return buf.String()
}

// In the test:
if got.testString() != want.testString() {
    t.Errorf("graph mismatch:\ngot:\n%s\nwant:\n%s",
        got.testString(), want.testString())
}
```

The `testString()` pattern is especially valuable when a struct has 2,000+ nodes
and a `reflect.DeepEqual` failure is unreadable.

### Subprocessing

When testing code that shells out to an external binary, two options are available.

**Option 1 — Execute the real binary.** Guard with a boolean initialized in `init`:

```go
var testHasGit bool

func init() {
    if _, err := exec.LookPath("git"); err == nil {
        testHasGit = true
    }
}

func TestGitStatus(t *testing.T) {
    if !testHasGit {
        t.Log("git not found, skipping")
        t.Skip()
    }
    // use real git binary
}
```

**Option 2 — Mock the subprocess** (from the Go standard library's `os/exec`
tests). Make `*exec.Cmd` configurable in the struct under test. In the test,
substitute a `helperProcess` command that re-enters the test binary:

```go
func helperProcess(s ...string) *exec.Cmd {
    cs := []string{"-test.run=TestHelperProcess", "--"}
    cs = append(cs, s...)
    env := []string{"GO_WANT_HELPER_PROCESS=1"}
    cmd := exec.Command(os.Args[0], cs...)
    cmd.Env = append(env, os.Environ()...)
    return cmd
}

func TestHelperProcess(*testing.T) {
    if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
        return
    }
    defer os.Exit(0)

    args := os.Args
    for len(args) > 0 {
        if args[0] == "--" {
            args = args[1:]
            break
        }
        args = args[1:]
    }

    cmd, args := args[0], args[1:]
    switch cmd {
    case "status":
        fmt.Println("nothing to commit")
    default:
        fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
        os.Exit(1)
    }
}
```

`TestHelperProcess` is a no-op during normal `go test` runs (the env var is not
set). When called via `helperProcess`, it implements mock behavior for the
subprocess.

**Do Not:**
- Hardcode the path to an external binary — always use `exec.LookPath`
- Mock the *output* of a subprocess by returning fake bytes — mock the subprocess
  itself using the `helperProcess` pattern

### Interfaces

Interfaces are mocking points. Define interfaces at the point of use (the caller
defines what it needs), not in the package that implements the behavior.

Keep interfaces as small as possible. If a function only needs `io.Closer`,
accept `io.Closer`, not `net.Conn`. A smaller interface requires less mock
implementation in tests.

```go
// Too broad — tests must implement the full net.Conn interface
func ServeConn(c net.Conn) error { … }

// Better — tests only need to implement io.ReadWriteCloser
func ServeConn(c io.ReadWriteCloser) error { … }
```

Create interfaces where you expect alternate implementations or where the
interface is the natural mocking seam for tests. Do not create an interface for
every type — overdoing it complicates readability without benefit.

**Do Not:**
- Define interfaces in the implementing package and export them for callers to use
- Create interfaces wider than the function actually needs

### Testing as a Public API

For packages consumed by other packages, export test helpers in a `testing.go`
file (not `_test.go`). This file is compiled into the package and importable by
consumers, allowing them to test against your package without rebuilding all the
scaffolding themselves.

Common exports:

```go
// testing.go

// TestConfig returns a valid configuration suitable for use in tests.
func TestConfig(t testing.T) *Config {
    t.Helper()
    return &Config{
        Addr:    "127.0.0.1:0",
        Timeout: 100 * time.Millisecond,
    }
}

// TestConfigInvalid returns a configuration guaranteed to fail validation.
func TestConfigInvalid(t testing.T) *Config {
    t.Helper()
    return &Config{} // missing required fields
}

// TestServer starts an in-memory server and returns its address and a closer.
func TestServer(t testing.T) (net.Addr, io.Closer) {
    t.Helper()
    srv := newInMemoryServer()
    if err := srv.Start(); err != nil {
        t.Fatalf("TestServer: %s", err)
    }
    return srv.Addr(), srv
}
```

Use the [`go-testing-interface`](https://github.com/mitchellh/go-testing-interface)
`T` interface instead of `*testing.T` in exported test helpers. Importing
`testing` from a non-test file injects `go test` flags into every program that
imports your package.

```go
import testing "github.com/mitchellh/go-testing-interface"

func TestServer(t testing.T) (net.Addr, io.Closer) { … }
```

Also export mock structs for interfaces your package defines, so consumers can
record and replay calls in their own tests.

**Do Not:**
- Import `*testing.T` directly in `testing.go` (use the interface instead)
- Force consumers to build their own in-memory servers or fixture constructors

### Custom Frameworks

Always use `go test` as the test runner, even for acceptance tests, integration
tests, and tests that provision real infrastructure. Build a custom framework
*inside* `go test`, not a separate harness alongside it.

Guard expensive test suites with a flag:

```go
var flagAcceptance = flag.Bool("acceptance", false, "run acceptance tests")

func TestProvider_basic(t *testing.T) {
    if !*flagAcceptance {
        t.Skip("skipping acceptance test; run with -acceptance")
    }
    // provision real resources, make real API calls
}
```

The custom framework handles setup/teardown, assertions, and lifecycle
management. The `go test` workflow handles invocation, filtering, output,
and CI integration.

```go
// Example pattern: custom test harness invoked via go test
logicaltest.Test(t, logicaltest.TestCase{
    PreCheck: func() { checkEnvVars(t) },
    Steps: []logicaltest.TestStep{
        testStepCreate(t),
        testStepRead(t),
        testStepDelete(t),
    },
})
```

**Do Not:**
- Write a separate test runner script that invokes binaries directly
- Skip `go test` in favor of `go run` for integration tests

### Repeat Yourself in Tests

Prefer a long, flat, self-contained test over a short test that delegates to
many helpers in many files. When a test fails months later during a refactor, the
reader should be able to understand the full test by reading one function without
jumping across files.

```go
// Prefer this (200 lines, everything in one place) …
func TestContext_Apply_basicDestroy(t *testing.T) {
    m := testModule(t, "apply-destroy")
    p := testProvider("aws")
    p.ApplyFn = testApplyFn
    p.DiffFn = testDiffFn
    state := &State{ … }
    ctx := testContext(t, &ContextOpts{
        Module:    m,
        Providers: map[string]ResourceProviderFactory{ "aws": testProviderFuncFixed(p) },
        State:     state,
    })
    // … 180 more explicit lines
}

// Over this (20 lines calling 7 helpers defined in 4 files):
func TestContext_Apply_basicDestroy(t *testing.T) {
    ctx := defaultTestContext(t)
    applyAndCheck(t, ctx, expectedDestroyState())
}
```

When writing a new test that resembles an existing one, copy the existing test
and modify the relevant lines. It feels wasteful but it keeps each test
independently readable.

**Do Not:**
- Abstract test setup into shared helpers unless the helper is truly universal
  (like `testTempFile`)
- Create test packages or test utility files to share logic across tests that
  are not meaningfully related

---

## Checklist

### Test methodology
- [ ] Table-driven test structure used, even for single cases
- [ ] Every table case has a name (no index-based names)
- [ ] `t.Run` used to scope each case so `defer` works and cases are targetable
- [ ] Test fixtures stored in `test-fixtures/`, accessed with relative paths
- [ ] Complex output tested with golden files and an `-update` flag
- [ ] Test helpers accept `*testing.T`, never return errors, call `t.Helper()`
- [ ] Cleanup returned as `func()` closure, deferred at call site
- [ ] Async tests use `select` + `time.After` + `timeMultiplier`, not `time.Sleep`
- [ ] `t.Parallel()` not used

### Writing testable code
- [ ] No unoverridable global constants controlling runtime behavior
- [ ] Configurable defaults follow `const defaultX` + `struct { X type }` pattern
- [ ] Only exported API tested; unexported functions treated as implementation details
- [ ] Large packages split via `internal/` to create test boundaries without public API
- [ ] Real network connections used in tests; `net.Conn` not mocked
- [ ] Structs overparameterized so tests can override ports, paths, timeouts
- [ ] Complex struct equality tested via diff library or `testString()` pattern
- [ ] Subprocess tests use `testHasX bool` guard or `helperProcess` mock pattern
- [ ] Interfaces defined at point of use, as narrow as the caller needs
- [ ] Public packages export test helpers in `testing.go` using `go-testing-interface`
- [ ] Acceptance/integration tests guarded by a flag, run via `go test`
