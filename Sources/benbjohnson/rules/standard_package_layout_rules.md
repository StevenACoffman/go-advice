# Ruleset: Standard Package Layout

```
Source: "Standard Package Layout" — Ben Johnson (2016)
Scope:  Go; application package structure; dependency isolation;
        domain-driven design; multi-binary applications
```

---

## 1. Root Package as Domain Layer

The root package is the application's domain: the shared vocabulary all other
packages speak. It must be dependency-free so that every other package can
import it without creating cycles.

---

§1.1  **[MUST][ARCH]**  Place all domain types — structs, interfaces, error
      types, and domain constants — in the root application package.
      Scattering domain types across subpackages forces every package that needs
      them to import a peer, which creates circular import candidates and requires
      callers to learn multiple import paths for conceptually unified types.
      ```
      ✗  package users; type User struct { ... }     // users.User stutter
      ✗  package models; type User struct { ... }    // models.User stutter

      ✓  package myapp; type User struct { ... }     // myapp.User — clear domain ownership
      ```

§1.2  **[MUST][ARCH]**  The root package must not import any other package
      within the same application.
      Any intra-application import in the root package creates a dependency on an
      implementation detail; swapping that implementation then requires changing
      the domain itself, and circular imports become possible as subpackages grow.

§1.3  **[MUST][ARCH]**  The root package must not import any external third-party
      or standard-library packages that represent I/O, persistence, or transport
      concerns (e.g. `database/sql`, `net/http`, `github.com/lib/pq`).
      External dependency imports in the root package couple the domain to a
      specific technology; unit-testing domain logic or swapping the implementation
      then requires the full technology stack to be available.

§1.4  **[SHOULD][CODE]**  Include a domain type in the root package only if it
      depends solely on other domain types; exclude any type that performs I/O,
      network calls, or database operations.
      A type that reaches outside the domain is an adapter, not a domain type —
      it belongs in the subpackage that owns the corresponding dependency.

§1.5  **[MUST][CODE]**  Do not produce stutter names (package name repeated in
      the type name, e.g. `users.User`, `controller.UserController`). When a name
      stutters, the type is in the wrong package.
      Stutter names are the observable symptom of type-by-function or
      type-by-module grouping; they pollute call sites and signal that the
      package boundary was drawn around the wrong concern.

---

**Flawed approaches ruled out by §1.x:**

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Group all handlers in `handlers/`, controllers in `controllers/`, models in `models/` (Rails-style) | ARCH | §1.1, §1.5 | Stutter names (`controller.UserController`); circular imports when types reference each other | Domain types in root package; implementations in dependency subpackages |
| Group all user code in `users/`, account code in `accounts/` (module-style) | ARCH | §1.1, §1.5 | `users.User` stutter; circular imports when modules reference each other | Domain types in root package |
| External dependency import in root package | ARCH | §1.3 | Domain coupled to technology; full stack required to test domain | Push dependency into named subpackage |
| Domain type that calls a database or HTTP endpoint | ARCH | §1.4 | Root package now depends on external package | Move to the appropriate dependency subpackage |
| Intra-application import in root package | ARCH | §1.2 | Circular imports; domain coupled to implementation | Root package may not import any sibling package |

---

## 2. Subpackages Grouped by Dependency

Subpackages are adapters between the domain and the outside world. Each one
owns exactly one external dependency and implements domain interfaces.

---

§2.1  **[MUST][ARCH]**  Create one subpackage per external dependency; name each
      subpackage after the dependency it wraps (e.g. `postgres`, `http`, `stripe`,
      `bolt`).
      Grouping by dependency makes every import path self-documenting and confines
      all code that can break when a dependency upgrades or is swapped to a single,
      clearly named package.
      ```
      ✗  myapp/db/           // ambiguous — which database?
      ✗  myapp/storage/      // generic — does not identify the dependency

      ✓  myapp/postgres/     // clearly owns the PostgreSQL dependency
      ✓  myapp/http/         // clearly owns the net/http dependency
      ```

§2.2  **[MUST][ARCH]**  Each dependency subpackage must implement domain
      interfaces defined in the root package; it must not expose its
      implementation-specific types (driver errors, ORM types, HTTP primitives) to
      callers outside the subpackage.
      Leaking implementation types forces callers to import the implementing
      package, coupling them to a technology they are not responsible for.

§2.3  **[MUST][ARCH]**  When two dependency subpackages need to share data, route
      that communication through root-package domain interfaces, not through direct
      imports between subpackages.
      A direct import from `postgres` into `stripe` (or vice versa) couples two
      implementations; changing either then requires understanding both, and
      circular imports become likely as the application grows.
      ```
      ✗  type UserService struct {
             DB             *sql.DB
             StripeClient   stripe.Client   // direct dependency import
         }

      ✓  type UserService struct {
             DB                 *sql.DB
             TransactionService myapp.TransactionService  // domain interface
         }
      ```

§2.4  **[MUST][ARCH]**  Treat standard-library packages that represent I/O or
      transport (`net/http`, `os`, `bufio`, etc.) as external dependencies and
      isolate them in their own subpackage by the same rule as third-party
      dependencies.
      Allowing `net/http` to be used freely across the application means HTTP
      concerns (routing, request parsing, response writing) leak into layers that
      should know nothing about the transport protocol; swapping or testing the
      transport then requires changes across the whole codebase.
      *(Overrides common practice — apply as stated.)*
      ```go
      // ✓  myapp/http/handler.go — all net/http code confined here
      package http

      import (
          "net/http"
          "github.com/benbjohnson/myapp"
      )

      type Handler struct {
          UserService myapp.UserService
      }

      func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { ... }
      ```

§2.5  **[CONSIDER][ARCH]**  Layer multiple implementations of the same domain
      interface (e.g. an in-memory LRU cache wrapping a database implementation)
      using the domain interface as the layering point, not struct embedding or
      package-level wiring.
      Using the domain interface as the layer boundary means each decorator is
      independently testable and the layering order can change without modifying
      either layer.
      ```go
      // ✓  Cache wraps any UserService — the type has no knowledge of postgres
      type UserCache struct {
          cache   map[int]*User
          service UserService   // domain interface, not *postgres.UserService
      }
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `myapp/storage/` or `myapp/db/` grouping multiple backends | ARCH | §2.1 | Ambiguous ownership; changes to one backend risk breaking others | One subpackage per backend (`postgres/`, `bolt/`) |
| Subpackage exposes `*sql.DB`, `*pq.Error`, or driver types in its public API | ARCH | §2.2 | Callers must import the implementing package; technology swap requires call-site changes | Return and accept domain types only |
| `postgres` package imports `stripe` package | ARCH | §2.3 | Two implementations coupled; circular import risk | Communicate via root-package domain interface |
| `net/http` used freely in service or domain packages | ARCH | §2.4 | HTTP concerns bleed into business logic; testing requires HTTP stack | Confine all `net/http` to `myapp/http/` |
| Cache wraps `*postgres.UserService` directly | ARCH | §2.5 | Cache coupled to one implementation; cannot layer differently | Wrap via `myapp.UserService` interface |

---

## 3. Shared Mock Subpackage

Mocks are the test-time implementations of domain interfaces. They must live in
one shared location so every consuming package can import them without cycles.

---

§3.1  **[MUST][ARCH]**  Place all mock implementations of domain interfaces in a
      single shared `mock` subpackage (`myapp/mock`); do not scatter test doubles
      into `_test.go` files in consuming packages or inline them as unexported
      test helpers.
      Scattered mocks are duplicated across packages; when a domain interface gains
      a method, every scattered mock must be updated independently, and the
      inconsistency is not caught until a test runs.

§3.2  **[SHOULD][CODE]**  Implement mock structs using exported function fields
      (one per interface method) and a corresponding boolean invocation flag:
      ```go
      // ✓  Function-field mock pattern
      type UserService struct {
          UserFn      func(id int) (*myapp.User, error)
          UserInvoked bool
          UsersFn     func() ([]*myapp.User, error)
          UsersInvoked bool
      }

      func (s *UserService) User(id int) (*myapp.User, error) {
          s.UserInvoked = true
          return s.UserFn(id)
      }
      ```
      This pattern lets each test define exactly the behaviour it needs without
      inheriting unrelated behaviour from base types or generated code; the
      invocation flags verify the interface was called without additional assertion
      libraries.

§3.3  **[MUST][ARCH]**  Tests must receive mocks through domain interfaces, not
      through concrete implementation types.
      Accepting `*postgres.UserService` in a test couples the test to an
      implementation; the test then requires a live database even for logic that
      has nothing to do with persistence.
      ```go
      // ✗  Test directly instantiates implementation
      svc := &postgres.UserService{DB: testDB}
      h.UserService = svc

      // ✓  Test injects mock through the domain interface
      var us mock.UserService
      us.UserFn = func(id int) (*myapp.User, error) { ... }
      h.UserService = &us
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Mock defined as unexported type in `handler_test.go` | ARCH | §3.1 | Duplicated when another package needs the same mock; interface changes missed | Shared `myapp/mock` subpackage |
| Using a generated mock library (GoMock etc.) as the only mock strategy | ARCH | §3.1 (spirit) | Generated mocks couple tests to a code-gen pipeline; function-field pattern is simpler | Function-field mocks in `mock/` |
| Test takes `*postgres.UserService` as parameter | ARCH | §3.3 | Test requires real database; not a unit test | Inject via `myapp.UserService` interface |
| No invocation flag on mock methods | CODE | §3.2 | Cannot verify the interface was called; test misses integration errors | Add `XxxInvoked bool` per method |

---

## 4. Main Package as Wiring Layer

The `main` package is the only package that knows about all dependencies
simultaneously. Its sole responsibility is to instantiate and connect them.

---

§4.1  **[MUST][ARCH]**  Place each binary's entry point in
      `cmd/<binaryname>/main.go`; do not place `main` packages at the repository
      root or at the root of any subpackage.
      A `main` package at the repository root makes it impossible to import the
      root package as a library and conflates application entry with domain
      definition.
      ```
      ✗  myapp/main.go

      ✓  myapp/cmd/myapp/main.go
      ✓  myapp/cmd/myappctl/main.go
      ```

§4.2  **[MUST][ARCH]**  Instantiate and wire all concrete implementations
      exclusively in the `main` package; no other package may construct a concrete
      implementation of a domain interface and assign it to a dependency field.
      Constructing implementations outside `main` couples the constructing package
      to both the domain and the implementation simultaneously, violating the
      dependency-isolation model; it also makes it impossible to swap implementations
      without changing non-wiring code.

§4.3  **[SHOULD][ARCH]**  The `main` package should contain only wiring code
      (open connections, construct services, attach handlers, start servers). If
      it contains business logic, that logic has leaked from the wrong layer.
      Business logic in `main` cannot be unit-tested because `main` cannot be
      imported; it also cannot be reused by other binaries in `cmd/`.

§4.4  **[MUST][ARCH]**  The `main` package is the only package permitted to
      import both root-domain packages and dependency-subpackages simultaneously.
      Any non-`main` package that imports both is performing wiring — a
      responsibility that belongs exclusively to `main`.

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `myapp/main.go` at repository root | ARCH | §4.1 | Root package cannot be imported as a library; `go get` fetches a binary, not a library | `myapp/cmd/myapp/main.go` |
| Service package constructs its own `postgres.UserService` | ARCH | §4.2 | Service package now imports both domain and implementation; swapping requires changing non-wiring code | Inject via constructor parameter or field assignment in `main` |
| Business logic inside `main()` | ARCH | §4.3 | Logic untestable and unreusable across binaries | Extract to domain or adapter layer; `main` only wires |
| Non-`main` package imports both `myapp` and `myapp/postgres` | ARCH | §4.4 | Wiring concern has leaked; package now owns both domain definition and implementation choice | Move wiring to `main` |

---

## Master Anti-patterns Table

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Rails-style grouping: `handlers/`, `controllers/`, `models/` | ARCH | §1.1, §1.5 | Stutter names; circular imports | Domain types in root; implementations in dependency subpackages |
| Module-style grouping: `users/`, `accounts/` | ARCH | §1.1, §1.5 | `users.User` stutter; cross-module circular imports | Domain types in root package |
| Intra-application import in root package | ARCH | §1.2 | Circular import risk; domain coupled to implementation | Root package has no intra-app imports |
| External I/O dependency in root package | ARCH | §1.3 | Domain coupled to technology | Push dependency into named subpackage |
| Domain type that performs I/O or network calls | ARCH | §1.4 | Root package acquires an external dependency | Move to the appropriate dependency subpackage |
| Stutter name (`controller.UserController`) | CODE | §1.5 | Wrong package boundary; name quality degrades | Move type to root package or rename subpackage |
| Generic subpackage name (`storage/`, `db/`) | ARCH | §2.1 | Ownership ambiguous; multiple backends collide | Name subpackage after the dependency it wraps |
| Subpackage exposes driver/ORM types in its public API | ARCH | §2.2 | Callers coupled to implementation | Return and accept domain types only |
| Direct import between dependency subpackages | ARCH | §2.3 | Two implementations tightly coupled | Communicate via root-package domain interface |
| `net/http` used freely outside `http/` subpackage | ARCH | §2.4 | HTTP concerns bleed into business logic | Confine `net/http` to `myapp/http/` |
| Mock defined inside a `_test.go` file | ARCH | §3.1 | Duplicated across packages; interface changes missed | Shared `myapp/mock` subpackage |
| Test accepts concrete implementation type | ARCH | §3.3 | Requires live dependency (database, HTTP server) | Inject mock via domain interface |
| `myapp/main.go` at repository root | ARCH | §4.1 | Root package unimportable as library | `cmd/<name>/main.go` |
| Concrete implementation constructed outside `main` | ARCH | §4.2 | Non-wiring package owns both domain and implementation | Construction exclusively in `main` |
| Business logic in `main()` | ARCH | §4.3 | Untestable; unreusable across binaries | Extract to domain or adapter layer |
| Non-`main` package imports both domain and its implementation | ARCH | §4.4 | Wiring leaked out of `main` | Move wiring to `main` |
