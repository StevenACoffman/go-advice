# Ruleset: Structuring Tests in Go

```
Source: "Structuring Tests in Go" — Ben Johnson (2014)
Scope:  Go; unit testing; test package layout; mocking external dependencies;
        test helpers; interface design for testability
```

---

## 1. Test Foundations: Self-Contained and Reproducible

Every structural decision in the test suite follows from two invariants: tests
must not affect each other, and anyone must be able to run them without
environment setup beyond what the test itself provides.

---

§1.1  **[MUST][METHOD]**  Design every test to be self-contained: it must not
      depend on state created or modified by another test, and changing or
      removing one test must not break another.
      Tests that share mutable state (globals, shared DB fixtures, shared temp
      files) produce order-dependent failures; a failing test then requires
      understanding the entire suite to diagnose rather than just the failing
      function.

§1.2  **[MUST][METHOD]**  Design every test to be independently reproducible: a
      developer must be able to run any single test in isolation and get the same
      result as running the full suite.
      Tests that require manual environment setup (seeded databases, external
      services, specific file paths) fail silently in CI or on a fresh checkout,
      and the failure is in the environment rather than the code under test.

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Tests share a package-level DB or client variable | METHOD | §1.1 | One test's mutation corrupts another's state | Each test constructs its own instance (see §3.x) |
| Test requires external service or manually seeded DB | METHOD | §1.2 | Fails on fresh checkout or in CI | Use temp files, in-process databases, or mocks |
| `TestMain` that seeds shared state consumed by all tests | METHOD | §1.1 | Suite-wide state causes order-dependent failures | Per-test setup/teardown via test-specific types |

---

## 2. Use Only the Standard Library Testing Package

Third-party testing frameworks add a dependency barrier, impose non-idiomatic
assertion styles, and obscure the standard test output format that Go tooling
and CI systems expect.

---

§2.1  **[MUST][CODE]**  Do not use third-party testing frameworks (BDD runners,
      assertion libraries, setup/teardown frameworks); use only the standard
      `testing` package.
      A framework that another contributor must `go get` before they can run
      tests raises the contribution barrier; framework-specific assertion syntax
      also produces unfamiliar failure messages that slow diagnosis.
      *(Overrides common practice of recommending `testify` or similar — apply as
      stated.)*

§2.2  **[SHOULD][CODE]**  Reduce test verbosity with three small, locally defined
      helper functions rather than importing an assertion library:
      ```go
      // assert fails the test if condition is false.
      func assert(tb testing.TB, condition bool, msg string, v ...interface{})

      // ok fails the test if err is not nil.
      func ok(tb testing.TB, err error)

      // equals fails the test if exp != act.
      func equals(tb testing.TB, exp, act interface{})
      ```
      These three functions cover the vast majority of assertion needs without
      introducing an external dependency; they can be copied into any package
      and are immediately readable by any Go developer.
      ```go
      // ✗  Verbose but dependency-free stdlib approach
      value, err := DoSomething()
      if err != nil { t.Fatalf("DoSomething() failed: %s", err) }
      if value != 100 { t.Fatalf("expected 100, got: %d", value) }

      // ✓  Same coverage, local helpers, no framework
      value, err := DoSomething()
      ok(t, err)
      equals(t, 100, value)
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `github.com/stretchr/testify` or similar as test assertion library | CODE | §2.1 | External dependency; non-standard failure output; contributor barrier | Local `ok`/`equals`/`assert` helpers |
| BDD framework (`ginkgo`, `goconvey`) | CODE | §2.1 | Non-idiomatic; framework-specific runner required | Standard `testing` package with table-driven tests |
| Verbose `if err != nil { t.Fatalf(...) }` repeated throughout | CODE | §2.2 | Noise obscures test logic | Local `ok(t, err)` helper |

---

## 3. Use the `_test` External Package for Exported API Tests

Testing from inside the package under test gives access to unexported symbols,
which encourages testing implementation details rather than the public contract.

---

§3.1  **[MUST][CODE]**  Name test files that test a package's exported API with
      the `package foo_test` declaration (the "underscore test" package), not
      `package foo`.
      Testing from inside the package tempts the test author to access unexported
      fields and functions; tests that rely on unexported internals are brittle —
      any refactor that preserves exported behaviour but changes internals breaks
      them.
      ```go
      // ✗  package myapp — test has access to unexported symbols
      // user_test.go
      package myapp
      func TestUser_create(t *testing.T) { u := &User{}; u.create() }  // tests internal

      // ✓  package myapp_test — test sees only exported API
      // user_test.go
      package myapp_test
      import "github.com/benbjohnson/myapp"
      func TestUser_Save(t *testing.T) {
          u := &myapp.User{Name: "Susy Queue"}
          ok(t, u.Save())
      }
      ```

§3.2  **[SHOULD][CODE]**  Do not use dot-imports (`. "pkg/path"`) in external
      test packages even though they syntactically resemble being inside the
      package.
      A dot-import obscures which package a symbol comes from; tests are the
      primary documentation of API usage, and a test that hides its import paths
      misleads readers about how to call the package.
      *(Source updated its own guidance against dot-imports — applied as stated.)*

§3.3  **[CONSIDER][CODE]**  Reserve `package foo` (same-package) test files for
      testing unexported helpers that have no exported surface to test through.
      Same-package tests are appropriate when the logic under test is deliberately
      unexported and has no observable effect on the exported API; in all other
      cases, prefer the external test package.

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `package myapp` in `myapp_test.go` testing exported behaviour | CODE | §3.1 | Test couples to unexported internals; refactors break it | `package myapp_test` |
| Dot-import in external test package: `. "github.com/foo/bar"` | CODE | §3.2 | Symbol origin obscured; misleads readers about API usage | Explicit `myapp.User{...}` |
| Testing unexported functions via same-package test when exported path exists | CODE | §3.1 | Implementation details tested; internals cannot be refactored freely | Test through exported API |

---

## 4. Use Test-Specific Types for Setup and Teardown

Ad-hoc setup/teardown code duplicated in every test function is a maintenance
liability and a source of subtle inconsistency.

---

§4.1  **[MUST][CODE]**  For any resource that requires setup and cleanup (temp
      files, database connections, servers), define a test-specific wrapper type
      with a constructor that opens the resource and a `Close()` method that
      cleans it up; use `defer db.Close()` immediately after construction.
      Setup/teardown code inline in each test function is duplicated; when the
      resource's lifecycle changes, every test must be updated independently, and
      one missed cleanup leaks resources or leaves state that contaminates other
      tests.
      ```go
      // ✓  Test-specific type handles full lifecycle
      type TestDB struct{ *DB }

      func NewTestDB() *TestDB {
          f, _ := ioutil.TempFile("", "")
          path := f.Name()
          f.Close(); os.Remove(path)
          db, err := Open(path, 0600)
          if err != nil { panic(err) }
          return &TestDB{db}
      }

      func (db *TestDB) Close() {
          defer os.Remove(db.Path())
          db.DB.Close()
      }

      // Test is now two lines of lifecycle management
      func TestDB_DoSomething(t *testing.T) {
          db := NewTestDB()
          defer db.Close()
          // ... test logic
      }
      ```

§4.2  **[SHOULD][CODE]**  Use `panic` (not `t.Fatal`) inside test helper
      constructors when resource initialisation fails.
      A constructor that calls `t.Fatal` requires `*testing.T` as a parameter,
      couples the constructor to the test framework, and cannot be used in
      `TestMain` or package-level init; `panic` surfaces the failure immediately
      with a clear message and requires no parameter threading.

§4.3  **[SHOULD][CODE]**  Add application-specific convenience methods to test
      wrapper types instead of repeating setup logic inline in each test.
      When multiple tests need the same precondition (a user already created, a
      transaction already started), a method on the test type centralises that
      logic; tests that call it are shorter and the precondition is consistently
      applied.

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `os.MkdirTemp` / `sql.Open` inline in each test function | CODE | §4.1 | Duplicated; one test's missed cleanup leaks into others | `NewTestDB()` + `defer db.Close()` |
| `TestMain` that opens one shared DB for all tests | CODE | §1.1, §4.1 | Shared mutable state; tests not self-contained | Per-test `NewTestDB()` |
| Test helper constructor takes `*testing.T` and calls `t.Fatal` | CODE | §4.2 | Constructor coupled to test framework; cannot be used outside test functions | `panic` on init failure |
| Precondition logic (create seed user) repeated in every test | CODE | §4.3 | Duplication; inconsistent setup when one copy diverges | Method on test wrapper type |

---

## 5. Define Interfaces in the Caller, Not the Callee

An interface defined by the package that provides a type forces all callers to
use the full surface area even when they need only a subset; it also couples
callers to the provider's API evolution.

---

§5.1  **[MUST][ARCH]**  Define an interface for an external dependency inside the
      package that *uses* it, not in the package that *provides* it; the interface
      should contain only the methods the consuming package actually calls.
      An interface declared in the provider exposes the full API surface; the
      consumer declares exactly what it needs, making the mock smaller, the
      dependency explicit, and the consumer independently testable.
      ```go
      // ✗  Consumer accepts the full concrete type (or an interface from provider)
      type MyApplication struct {
          YoClient *yo.Client   // coupled to concrete type
      }

      // ✓  Consumer declares the minimal interface it needs
      type MyApplication struct {
          YoClient interface {
              Send(string) error
          }
      }
      ```

§5.2  **[MUST][CODE]**  Write the mock for an inline interface in the test file
      (or `_test` package) that uses it, not in a shared mock package, when the
      interface is declared inline in the consumer.
      A shared mock for an inline interface requires the consumer's package to
      export the interface, which defeats the purpose of declaring it inline; the
      mock belongs with the test that uses it.
      ```go
      // ✓  Mock defined in the test file, not exported to a shared package
      type testYoClient struct {
          SendFunc func(string) error
      }
      func (c *testYoClient) Send(s string) error { return c.SendFunc(s) }
      ```

§5.3  **[SHOULD][CODE]**  Give mock function fields the suffix `Func` (e.g.
      `SendFunc func(string) error`) so that the test can assign only the methods
      it needs and leave others as nil; unused methods that panic on nil confirm
      they were not called.
      A mock where every method must be pre-populated requires the test to stub
      methods it has no interest in; the `Func` suffix pattern lets each test
      configure exactly the behaviour it exercises.

§5.4  **[CONSIDER][CODE]**  Use the same inline-interface pattern to mock standard
      library packages (`os`, `io`) by declaring a minimal interface for the
      filesystem or I/O operations the consumer calls.
      Replacing `os.File` with an interface that has `Create(string) (io.Writer,
      error)` removes the filesystem dependency from any test that exercises that
      code path.

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Struct field typed as `*yo.Client` (concrete third-party type) | ARCH | §5.1 | Test must construct or stub a real `yo.Client`; consumer coupled to provider API | Inline interface with only the methods needed |
| Interface for external dependency declared in the provider package | ARCH | §5.1 | Consumer forced to satisfy full provider interface; mock must implement unused methods | Caller-side inline interface |
| Mock placed in shared `mock/` package for an inline consumer interface | CODE | §5.2 | Consumer interface must be exported; defeats inline design | Mock defined in test file or `_test` package |
| Mock methods without `Func` suffix, all pre-populated | CODE | §5.3 | Every test must stub every method even when not called | `SendFunc func(...)` — nil = not called |
| `os.Create(path)` called directly in code under test | CODE | §5.4 | Test requires real filesystem; temp-file cleanup needed | Accept `interface{ Create(string) (io.Writer, error) }` |

---

## Master Anti-patterns Table

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Tests share a package-level variable for DB or config | METHOD | §1.1 | Order-dependent failures; parallel tests corrupt each other | Per-test construction |
| Test requires manually seeded DB or external service | METHOD | §1.2 | Fails on fresh checkout; not reproducible | Temp files, in-process DB, or mocks |
| `github.com/stretchr/testify` or BDD framework | CODE | §2.1 | External dependency; contributor barrier | Local `ok`/`equals`/`assert` helpers |
| Verbose `if err != nil { t.Fatalf(...) }` repeated | CODE | §2.2 | Noise obscures intent | `ok(t, err)` |
| `package myapp` in test file testing exported behaviour | CODE | §3.1 | Internals accessible; refactors break tests | `package myapp_test` |
| Dot-import in `_test` package | CODE | §3.2 | Symbol origin obscured | Explicit qualified names |
| Setup/teardown logic inline in each test function | CODE | §4.1 | Duplication; missed cleanup leaks | Test wrapper type with `NewTestX()` + `defer x.Close()` |
| Shared `TestMain` DB for all tests | CODE | §1.1, §4.1 | Shared mutable state | Per-test `NewTestDB()` |
| Test helper constructor calls `t.Fatal` | CODE | §4.2 | Coupled to test framework; unusable in `TestMain` | `panic` on init failure |
| Concrete third-party type as struct field | ARCH | §5.1 | Consumer coupled to provider; mock requires real dependency | Inline consumer-side interface |
| Interface declared in provider package | ARCH | §5.1 | Full surface area forced on consumers | Caller-side minimal interface |
| Shared mock package for inline consumer interface | CODE | §5.2 | Forces export of inline interface | Mock in test file |
| Mock fields without `Func` suffix, all pre-populated | CODE | §5.3 | Every test must stub every method | `XxxFunc func(...)` pattern |
