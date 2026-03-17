# Advanced Testing with Go

*Mitchell Hashimoto — [@mitchellh](https://twitter.com/mitchellh)*

---

## HashiCorp and Go

Go has been HashiCorp's primary language for approximately four years. Our
projects are deployed by millions of users and significantly in enterprise
environments. The software we build spans several demanding categories:

- **Distributed systems** — Consul, Serf, Nomad
- **Extreme performance** — Consul read performance, Nomad scheduling
- **Security** — Vault
- **Correctness** — Terraform, Consul, Nomad, Vault

---

## Two Parts of Testing

Good testing requires two things that are equally important:

**Test Methodology** — Methods to test specific cases; techniques to write better
tests. There is a lot more to testing than `assert(func() == expected)`.

**Writing Testable Code** — How to write code that can be tested well and easily.
Many developers who tell me "this can't be tested" aren't wrong—they just wrote
the code in a way that made it so. We very rarely see cases at HashiCorp that
truly can't be tested well. Rewriting existing code to be testable is a pain,
but it is worth it.

---

## Part One: Test Methodology

### Table-Driven Tests

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
            t.Errorf(
                "%d + %d = %d, expected %d",
                tc.A, tc.B, actual, tc.Expected)
        }
    }
}
```

- Low overhead to add new test cases
- Makes testing exhaustive scenarios simple
- Makes reproducing reported issues simple
- Follow this pattern a lot; use it even for single cases if the function could
  plausibly need more cases in the future

**Consider naming cases.** Using array indices for names produces failure messages
like "test index 3014 failed," which is useless. A map or a named field makes
failures immediately identifiable:

```go
func TestAdd(t *testing.T) {
    cases := map[string]struct{ A, B, Expected int }{
        "foo": {1, 1, 2},
        "bar": {1, -1, 0},
    }
    for k, tc := range cases {
        actual := tc.A + tc.B
        if actual != tc.Expected {
            t.Errorf(
                "%s: %d + %d = %d, expected %d",
                k, tc.A, tc.B, actual, tc.Expected)
        }
    }
}
```

### Test Fixtures

```go
func TestAdd(t *testing.T) {
    data := filepath.Join("test-fixtures", "add_data.json")
    // … do something with data
}
```

- `go test` sets the working directory to the package directory being tested
- Use a relative `test-fixtures/` directory as a place to store test data
- Very useful for loading config files, model data, binary data, etc.

### Golden Files (and Test Flags)

```go
var update = flag.Bool("update", false, "update golden files")

func TestAdd(t *testing.T) {
    // … table (probably!)
    for _, tc := range cases {
        actual := doSomething(tc)
        golden := filepath.Join("test-fixtures", tc.Name+".golden")
        if *update {
            ioutil.WriteFile(golden, actual, 0644)
        }
        expected, _ := ioutil.ReadFile(golden)
        if !bytes.Equal(actual, expected) {
            // FAIL
        }
    }
}
```

Usage:

```
$ go test          # compare against golden files
$ go test -update  # regenerate golden files
```

- Test complex output without manually hardcoding expected bytes
- Human-eyeball the generated golden data; if it is correct, commit it
- Very scalable way to test complex structures (write a `String()` method and
  use it as the golden output)

### Test Helpers

Test helpers should **never return errors**. Accept `*testing.T` and call
`t.Fatalf` internally. By not returning errors, usage is much prettier since
error checking is eliminated. Test helpers exist to make the test clear on what
it is actually testing, not to add boilerplate.

```go
func testTempFile(t *testing.T) string {
    tf, err := ioutil.TempFile("", "test")
    if err != nil {
        t.Fatalf("err: %s", err)
    }
    tf.Close()
    return tf.Name()
}
```

**Return a `func()` for cleanup.** The cleanup closure is an elegant way to
bundle teardown with setup. Because it is a closure, it retains access to `*testing.T`
and can call `t.Fatalf` if cleanup itself fails.

```go
func testTempFile(t *testing.T) (string, func()) {
    tf, err := ioutil.TempFile("", "test")
    if err != nil {
        t.Fatalf("err: %s", err)
    }
    tf.Close()
    return tf.Name(), func() { os.Remove(tf.Name()) }
}

func TestThing(t *testing.T) {
    tf, tfclose := testTempFile(t)
    defer tfclose()
}
```

When the cleanup has no meaningful return value, you can one-line the `defer`:

```go
func testChdir(t *testing.T, dir string) func() {
    old, err := os.Getwd()
    if err != nil {
        t.Fatalf("err: %s", err)
    }
    if err := os.Chdir(dir); err != nil {
        t.Fatalf("err: %s", err)
    }
    return func() { os.Chdir(old) }
}

func TestThing(t *testing.T) {
    defer testChdir(t, "/other")()
    // …
}
```

Proper setup and teardown for `testChdir` without the helper would be at least
10 lines in every test. The helper eliminates that across all tests.

### Timing-Dependent Tests

For tests that wait on asynchronous behavior, use a `select` with a timeout:

```go
func TestThing(t *testing.T) {
    // …
    select {
    case <-thingHappened:
    case <-time.After(timeout):
        t.Fatal("timeout")
    }
}
```

We do not use "fake time." Instead we have a multiplier available that can be
set to increase timeouts in slow environments:

```go
func TestThing(t *testing.T) {
    // …
    timeout := 3 * time.Minute * timeMultiplier
    select {
    case <-thingHappened:
    case <-time.After(timeout):
        t.Fatal("timeout")
    }
}
```

This is not perfect, but it is less intrusive than fake time. Fake time could
be better, but we have not found an effective way to use it yet.

### Parallelization

```go
func TestThing(t *testing.T) {
    t.Parallel()
}
```

**Don't do it. Run multiple processes instead.**

- Parallel tests make failures uncertain: is the failure due to a pure logic
  bug, or a race condition?
- Prefer running tests with both `-parallel=1` and `-parallel=N` if you need
  to check both
- We have preferred not to use parallelization. We use multiple processes and
  unit tests specifically written to test for races.

---

## Part Two: Writing Testable Code

### Global State

Avoid global state as much as possible. Instead of hard global state, make
whatever is global a configuration option that uses global state only as a
default, allowing tests to modify it.

If global state is truly necessary, make it a `var` (not a `const`) so it can
be modified in tests—but this is a last resort.

```go
// Not good on its own
const port = 1000

// Better
var port = 1000

// Best
const defaultPort = 1000

type ServerOpts struct {
    Port int // default to defaultPort somewhere
}
```

### Packages and Functions

- Break down functionality into packages and functions judiciously
- **Don't overdo it.** Do it where it makes sense. Doing this correctly aids
  testing while also improving organization; overdoing it will complicate both
  testing and readability
- Qualitative—practice will make perfect

Unless a function is extremely complex, test only the exported functions—the
exported API. Treat unexported functions and structs as implementation details:
they are a means to an end. As long as the end is tested and behaves within
spec, the means don't matter.

Some people take this too far and choose to only integration/acceptance test
(the ultimate "test the end, ignore the means"). We disagree with this approach.

### Networking

Testing networking? **Make a real network connection. Don't mock `net.Conn`.**

```go
// Error checking omitted for brevity
func testConn(t *testing.T) (client, server net.Conn) {
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    var server net.Conn
    go func() {
        defer ln.Close()
        server, err = ln.Accept()
    }()
    client, err := net.Dial("tcp", ln.Addr().String())
    return client, server
}
```

- That was a one-connection example; easy to make an N-connection helper
- Easy to test any protocol
- Easy to return the listener as well
- Easy to test IPv6 if needed
- There is no reason to ever mock `net.Conn`

### Configurability

Unconfigurable behavior is often a point of difficulty for tests. Examples:
ports, timeouts, paths.

Overparameterize structs to allow tests to fine-tune behavior. It is fine to
make these configurations unexported so only tests (within the same package)
can set them.

```go
// Do this, even if CachePath and Port are always the same
// in practice. For testing, it lets us be more careful.
type ServerOpts struct {
    CachePath string
    Port      int
}
```

### Subprocessing

Subprocessing is typically a point of difficult-to-test behavior. Two options:

1. Actually execute the subprocess
2. Mock the output or behavior

#### Option 1: Execute the Real Binary

Actually executing the subprocess is ideal when possible. Guard the test for
the existence of the binary and make sure side effects don't affect other tests.

```go
var testHasGit bool

func init() {
    if _, err := exec.LookPath("git"); err == nil {
        testHasGit = true
    }
}

func TestGitGetter(t *testing.T) {
    if !testHasGit {
        t.Log("git not found, skipping")
        t.Skip()
    }
    // …
}
```

#### Option 2: Mock the Subprocess

You still actually execute something—but you are executing a mock. Make the
`*exec.Cmd` configurable and pass in a custom one. This technique comes from
the Go standard library; it is how they test `os/exec`. HashiCorp uses it for
`go-plugin` and more.

**Get the `*exec.Cmd`:**

```go
func helperProcess(s ...string) *exec.Cmd {
    cs := []string{"-test.run=TestHelperProcess", "--"}
    cs = append(cs, s...)
    env := []string{
        "GO_WANT_HELPER_PROCESS=1",
    }
    cmd := exec.Command(os.Args[0], cs...)
    cmd.Env = append(env, os.Environ()...)
    return cmd
}
```

**What it executes (`TestHelperProcess`):**

```go
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
    case "foo":
        // implement mock behavior for "foo"
    }
}
```

`TestHelperProcess` returns immediately when `GO_WANT_HELPER_PROCESS` is not
set, so normal `go test` runs are unaffected. When it is set, the function
implements mock subprocess behavior by switching on the command arguments.

### Interfaces

Interfaces are **mocking points**. Behavior can be defined regardless of
implementation and exposed via a custom framework or `testing.go` (see below).

Similar to packages and functions: use interfaces judiciously. Overdoing it
will complicate readability.

### Testing as a Public API

Newer HashiCorp projects have adopted the practice of creating `testing.go` or
`testing_*.go` files. These export APIs for the sole purpose of providing mocks,
test harnesses, and helpers. They allow other packages to test using our package
without reinventing the components needed to meaningfully use our package in a
test.

**Example: config file parser**

```go
TestConfig(t)        // Returns a valid, complete configuration for tests
TestConfigInvalid(t) // Returns an invalid configuration
```

**Example: API server**

```go
TestServer(t) (net.Addr, io.Closer)
// Returns a fully started in-memory server (address to connect to)
// and a closer to shut it down.
```

**Example: interface for downloading files**

```go
TestDownloader(t, Downloader)
// Tests all the properties a Downloader should have.

type DownloaderMock struct{}
// Implements Downloader as a mock, allowing recording and replaying of calls.
```

### Custom Frameworks

`go test` is an incredible workflow tool. For complex, pluggable systems, write
a custom framework *within* `go test` rather than building a separate test
harness. Examples: Terraform providers, Vault backends, Nomad schedulers.

```go
// Example from Vault
func TestBackend_basic(t *testing.T) {
    b, _ := Factory(logical.TestBackendConfig())
    logicaltest.Test(t, logicaltest.TestCase{
        PreCheck: func() { testAccPreCheck(t) },
        Backend:  b,
        Steps: []logicaltest.TestStep{
            testAccStepConfig(t, false),
            testAccStepRole(t),
            testAccStepReadCreds(t, b, "web"),
            testAccStepConfig(t, false),
            testAccStepRole(t),
            testAccStepReadCreds(t, b, "web"),
        },
    })
}
```

`logicaltest.Test` is a custom harness that handles repeated setup/teardown,
assertions, and so on. Terraform provider acceptance tests follow the same
pattern. We can still use `go test` to run them all.

---

*Thank you — [hashicorp.com](https://hashicorp.com)*
