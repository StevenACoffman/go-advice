# Ben Johnson Go Architecture — Unified Ruleset

```
Sources:          "Common CRUD Design in Go" — Ben Johnson (Jan 2021)
                  "Failure Is Your Domain" — Ben Johnson (2018)
                  "Packages as Layers, Not Groups" — Ben Johnson (Jan 2021)
                  "Real-World SQL in Go: Part I" — Ben Johnson (Jan 2021)
                  "Standard Package Layout" — Ben Johnson (2016)
                  "Structuring Applications in Go" — Ben Johnson (2014)
                  "Structuring Tests in Go" — Ben Johnson (2014)
                  "Introducing WTF Dial" — Ben Johnson (Jan 2021)
Scope:            Go; domain-driven application architecture; HTTP services;
                  SQL persistence; unit testing
Synthesized from: 8 rulesets (6 distilled, 2 derived from source documents)
```

---

## 1. Package Architecture

The root package is the application's shared vocabulary. Every other package is an adapter between that vocabulary and an external dependency. The relationship is always one-directional: adapters depend on the domain; the domain depends on nothing.

---

§1.1  `[MUST][ARCH]`  Place all domain types — structs, service interfaces, the application `Error` type, and domain constants — in the root application package.
      When domain types are scattered across subpackages, every package that needs them must import a peer, creating circular import candidates and multiple authoritative sources for what should be one vocabulary. Go packages are layers, not groups; the root package is the layer every other package depends on.
      ✗  `package users; type User struct { ... }` — `users.User` stutter; `sqlite.DialService` owns its own interface
      ✓  `package myapp; type User struct { ... }; type DialService interface { ... }; type Error struct { ... }`

§1.2  `[MUST][ARCH]`  The root package must import no sibling application packages and no external packages representing I/O, persistence, or transport (e.g. `database/sql`, `net/http`, `github.com/lib/pq`).
      Any intra-application import in the root creates a dependency on an implementation; any external I/O import couples the domain to a technology. Both prevent the root package from being loaded in isolation for testing or for use in alternate contexts.
      ✗  `import "net/http"` or `import "database/sql"` in the root domain package
      ✓  Root package contains only types defined in terms of other domain types and primitive/stdlib value types

§1.3  `[MUST][ARCH]`  Do not produce stutter names (package name repeated in the type name, e.g. `users.User`, `controller.UserController`). When a name stutters, the type is in the wrong package.
      A stutter name is the observable symptom of a wrong architectural boundary: the package was drawn around a domain concept that belongs in the root layer, not around an external dependency. Left uncorrected, the misplaced type becomes a peer-import anchor that grows into circular import pressure as the application adds more packages.
      ✗  `type UserService struct { ... }` in `package users` — callers write `users.UserService`
      ✓  `type UserService interface { ... }` in root `package myapp` — callers write `myapp.UserService`

§1.4  `[MUST][ARCH]`  Create one subpackage per external dependency; name each subpackage after the dependency it wraps (e.g. `postgres`, `http`, `stripe`, `bolt`). Treat standard-library I/O packages (`net/http`, `os`, `bufio`) as external dependencies by the same rule.
      A generic subpackage name (`storage/`, `db/`) hides which technology is wrapped. When the technology changes, identifying affected code requires reading rather than navigating. Naming by dependency makes every import path self-documenting and confines all breakage from a dependency upgrade to a single, clearly-named package.
      ✗  `myapp/storage/` or `myapp/db/` containing both postgres and redis code
      ✓  `myapp/postgres/`, `myapp/http/`, `myapp/stripe/` — one dependency per package

§1.5  `[MUST][ARCH]`  Each dependency subpackage must implement domain interfaces defined in the root package and must not expose its implementation-specific types (driver errors, ORM types, HTTP primitives) to callers outside the subpackage.
      Leaking implementation types forces callers to import the implementing package, coupling them to a technology they are not responsible for. Swapping the implementation then requires changes across every caller.
      ✗  `func (s *UserService) Find(id int) (*pq.Row, error)` — leaks postgres type into calling code
      ✓  `func (s *UserService) FindUserByID(ctx context.Context, id int) (*myapp.User, error)`

§1.6  `[MUST][ARCH]`  When two dependency subpackages need to share data, route that communication through root-package domain interfaces; do not import one subpackage from another.
      A direct import from `postgres` into `stripe` couples two implementations; changing either requires understanding both, and circular imports become likely as the application grows. Routing via the domain interface also enables layering: a cache decorator can wrap any `UserService` without knowing whether the underlying implementation is postgres or in-memory.
      ✗  `type UserService struct { DB *sql.DB; StripeClient stripe.Client }` — direct implementation import
      ✓  `type UserService struct { DB *sql.DB; TransactionService myapp.TransactionService }` — domain interface

§1.7  `[SHOULD][ARCH]`  Design every Go application to be distributed as a single statically-compiled binary. For applications serving hundreds of users or fewer, prefer a single-process deployment on a small host over a distributed microservices architecture.
      A shared-library or runtime dependency introduces a deployment failure mode CI cannot validate: the binary is present, tests pass, but the application fails to start on the target host. A distributed deployment adds service discovery, inter-service latency, distributed tracing, and independent failure domains — operational concerns that provide value only beyond a team and load threshold the application may not have reached.
      ✗  Binary `dlopen`s a system `.so` at startup, or requires JVM/Python on the target host
      ✗  Microservices deployment for an application with a small user base and a single team
      ✓  `go build -o myapp ./cmd/myapp` — uploads and runs on a single VPS without installation steps

§1.8  `[MUST][ARCH]`  Create a subpackage only when a clear dependency boundary exists that justifies the isolation; do not split for file-count or line-count reasons alone. When a single module grows beyond approximately 10,000 SLOC, evaluate splitting it into independently importable packages.
      Premature package splits force types that naturally call each other to become exported (to cross package lines), enlarging the public API and creating cyclic import candidates. An oversized module forces downstream consumers to load everything to use any part of it; an independently importable package makes the boundary explicit by design.
      ✗  New subpackage created because a source file has "too many types"
      ✓  New subpackage created because a new external dependency (redis, S3) requires isolation

§1.9  `[SHOULD][ARCH]`  Group related types and functions together in one file; target 200–500 SLOC per file, with 1,000 SLOC as the upper limit before splitting; within a file, place the most important type first.
      One-type-per-file scatters related code; a reader tracing a call must open multiple files to build a mental model. A reader who opens a file should find its primary type immediately at the top; secondary helpers should not crowd it out.
      ✗  `user.go`, `order.go`, `account.go` each containing a single type definition
      ✓  `user.go` containing `User`, `UserFilter`, and helpers used only with `User`; primary type at the top

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Domain types in `users/`, `models/`, `controllers/` | ARCH | §1.1 | Stutter names; circular import candidates | All domain types in root package |
| External I/O import in root package | ARCH | §1.2 | Domain coupled to technology; full stack required to test | Push dependency into named subpackage |
| Intra-application import in root package | ARCH | §1.2 | Circular import risk; domain coupled to implementation | Root package imports nothing from the same module |
| Stutter name (`users.UserService`, `controller.UserController`) | ARCH | §1.3 | Wrong architectural boundary; circular import pressure | Move type to root package or rename subpackage |
| Generic subpackage name (`storage/`, `db/`) | ARCH | §1.4 | Ownership ambiguous; multiple backends collide | Name subpackage after the dependency it wraps |
| Subpackage exposes driver or ORM types in its public API | ARCH | §1.5 | Callers coupled to technology; swap requires all-callers change | Return and accept domain types only |
| Direct import between dependency subpackages | ARCH | §1.6 | Two implementations coupled; circular import risk as app grows | Communicate via root-package domain interface |
| Binary requires system runtime or `dlopen`-loaded `.so` | ARCH | §1.7 | Deployment fails on clean host; CI does not validate host state | Statically compiled single binary |
| Microservices for a small-scale application | ARCH | §1.7 | Operational overhead dwarfs benefit at small scale | Single-process deployment until team or load justifies split |
| Subpackage created for file-count reasons | ARCH | §1.8 | Cyclic import pressure; unexported types must be exported | One package until a real dependency boundary exists |
| One type per file | ARCH | §1.9 | Related code fragmented; reader opens many files to trace one call | Group related types in one file |

---

## 2. Dependency Injection and Wiring

The application binary exists to wire together implementations that satisfy domain interfaces. Wiring is a concern distinct from both domain logic and adapter logic; confining it to `main` lets every other package be tested independently.

---

§2.1  `[MUST][ARCH]`  Do not use global variables to hold application state (database connections, configuration, loggers). Carry all dependencies as struct fields; inject them at construction time. HTTP handlers must be struct types with `ServeHTTP` methods, not bare functions registered via `http.HandleFunc`.
      Global state makes tests order-dependent and mutually interfering: a test that sets a global database connection or config value contaminates every other test in the same process. A bare function handler has no receiver and therefore no injection point; the only escape is a global variable.
      ✗  `var db *sql.DB` at package scope; `http.HandleFunc("/path", fn)` where `fn` reads `db`
      ✓  `type HelloHandler struct { db *sql.DB }; func (h *HelloHandler) ServeHTTP(w, r) { ... }`

§2.2  `[MUST][ARCH]`  Place each application binary in `cmd/<binaryname>/main.go`; do not place a `main` package at the repository root.
      A `main` package at the root prevents the repository from being imported as a library, limits the project to one binary, and conflates entry-point concerns with domain definition.
      ✗  `myapp/main.go` at the repository root
      ✓  `myapp/cmd/myapp/main.go`; `myapp/cmd/myapp-ctl/main.go`

§2.3  `[MUST][ARCH]`  Instantiate and wire all concrete implementations exclusively in the `main` package; never construct a concrete implementation of a domain interface outside of `main`. `main` must contain only wiring code — opening connections, constructing services, attaching handlers, starting servers. Business logic in `main` cannot be unit-tested and cannot be reused across binaries.
      Constructing implementations outside `main` couples a non-wiring package to both the domain and the implementation simultaneously. Any non-`main` package that imports both the root domain package and a dependency subpackage is performing wiring — a responsibility that belongs to `main` alone.
      ✗  Service package calls `postgres.NewUserService(db)` to provision its own dependency
      ✗  Authorization or domain logic written inside `main()`
      ✓  `main()` opens DB, calls `postgres.NewUserService(db)`, injects it into `http.NewHandler(svc)`, calls `srv.ListenAndServe()`

§2.4  `[SHOULD][ARCH]`  When the same core logic is needed in multiple distinct entry points (CLI and HTTP server, for example), represent each as a separate binary under `cmd/`, all importing the same core library package.
      A single binary that switches between modes via flags entangles both modes; each is harder to evolve independently, and the switching logic adds complexity that belongs to neither.
      ✗  `myapp --mode=server` vs `myapp --mode=migrate`
      ✓  `cmd/myapp-server/` and `cmd/myapp-migrate/` both importing `package myapp`

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `var db *sql.DB` at package level used by handlers | ARCH | §2.1 | Tests share mutable global; parallel tests corrupt each other | Struct-field injection |
| `http.HandleFunc` with a bare function reading globals | CODE | §2.1 | No injection point; global dependency is the only option | `http.Handle` with a struct handler |
| `main.go` at repository root | ARCH | §2.2 | Unimportable as library; only one binary possible | `cmd/<name>/main.go` |
| Concrete implementation constructed outside `main` | ARCH | §2.3 | Non-wiring package couples domain and implementation | Construction exclusively in `main` |
| Business logic inside `main()` | ARCH | §2.3 | Untestable; unreusable across binaries | Extract to domain or adapter layer |
| Single binary with mode-switching flag | ARCH | §2.4 | Both modes entangled; hard to evolve or test independently | Separate binaries under `cmd/` |

---

## 3. Error Handling

Application errors are domain artifacts, not infrastructure concerns. They must be owned by the domain package, carry exactly what each consumer role needs, and form a chain traversable by helpers — not by type assertions.

---

§3.1  `[MUST][ARCH]`  Define the application `Error` type in the root domain package with exactly four fields: `Code string` (machine-readable), `Message string` (human-readable, end-user-facing), `Op string` (logical operation name), `Err error` (wrapped cause). Do not use a third-party error-wrapping library as the primary domain error type.
      A type in a subpackage named `errors` produces a stutter name and removes the error vocabulary from the domain. A third-party type cannot carry application-specific codes or messages and couples the domain's error contract to an external dependency. Typed wrappers for `Op` or `Code` add indirection without benefit; plain strings suffice.
      ✗  `package errors; type Error struct { ... }` — callers write `errors.Error{...}`; stutter
      ✗  `pkg/errors.Wrap(err, "op")` as the primary error for domain conditions
      ✓  `package myapp; type Error struct { Code, Message, Op string; Err error }` *(overrides common practice — apply as stated)*

§3.2  `[MUST][CODE]`  Never set both `Err` and (`Code` or `Message`) on a single `Error` instance. `Err` marks a wrapping node — meaning is delegated to the nested error. `Code`/`Message` mark a leaf node — they name the origin of the condition. Mixing them on one struct breaks the `ErrorCode()` and `ErrorMessage()` chain traversals.
      ```go
      // ✗  Mixing wrapping and leaf roles
      return &myapp.Error{Op: op, Code: myapp.EINTERNAL, Err: err}
      // ✓  Wrapping node: Op + Err only
      return &myapp.Error{Op: op, Err: err}
      // ✓  Leaf node: Code + Message only (no Err)
      return &myapp.Error{Code: myapp.EINVALID, Message: "Username is required."}
      ```

§3.3  `[SHOULD][CODE]`  Start with four generic error codes — `ECONFLICT`, `EINTERNAL`, `EINVALID`, `ENOTFOUND` — and expand only when a concrete use case cannot be described by an existing code. Document which codes each public function can return.
      Fine-grained codes added before any caller needs to distinguish them produce dead `case` branches — `case ECONFLICT_USERNAME:` paths that can never be executed because the code is never returned — and enlarge the interface contract irreversibly without proof of need. Undocumented codes mean callers cannot write exhaustive handling; a code that is not in the function comment is a surprise at runtime.
      ✗  `ECONFLICT_USERNAME`, `ECONFLICT_EMAIL` defined before any caller needed to distinguish them
      ✓  `// FindUserByID returns ENOTFOUND if the user does not exist.`

§3.4  `[MUST][CODE]`  Provide an `ErrorCode(err error) string` function that (1) returns `""` for `nil`, (2) returns the first non-empty `Code` found by recursively following `Err`, and (3) returns `EINTERNAL` when no `Code` is found anywhere in the chain.
      Without this helper, callers must type-assert to `*myapp.Error`, cannot traverse wrapped errors, and write verbose branching that breaks when an error gains another wrapping layer. Returning `""` (not `EINTERNAL`) for `nil` is mandatory: returning `EINTERNAL` for nil fires spurious internal-error branches on every nil-error path.
      ```go
      func ErrorCode(err error) string {
          if err == nil { return "" }
          if e, ok := err.(*Error); ok && e.Code != "" { return e.Code }
          if e, ok := err.(*Error); ok && e.Err != nil { return ErrorCode(e.Err) }
          return EINTERNAL
      }
      // ✗  err.(*myapp.Error).Code — panics if outer error is not *myapp.Error; doesn't traverse chain
      // ✓  myapp.ErrorCode(err) == myapp.ENOTFOUND
      ```

§3.5  `[MUST][CODE]`  Provide an `ErrorMessage(err error) string` function that (1) returns `""` for `nil`, (2) returns the first non-empty `Message` found by recursively following `Err`, and (3) returns a generic contact-support message when none is found. Set `Message` only on leaf `Error` nodes that describe a condition the end user can act on.
      Presentation-layer code without this helper must replicate chain traversal or receive empty strings for wrapped errors. A `Message` on a wrapping node causes `ErrorMessage()` to return the outermost (least specific) message instead of the actionable one from the origin.
      ```go
      // ✗  Message on a wrapping node — ErrorMessage() returns this instead of the leaf message
      return &myapp.Error{Op: op, Message: "something went wrong", Err: err}
      // ✓  Message only at the leaf where the actionable condition was detected
      return &myapp.Error{Code: myapp.EINVALID, Message: "Username is required."}
      ```

§3.6  `[MUST][CODE]`  When branching on an error in application logic, call `ErrorCode()`; never type-assert to an implementation-specific type to extract error details, and never match on `err.Error()` strings.
      Type-asserting to an implementation type couples calling code to the implementing package; swapping the implementation silently breaks every caller that relies on the concrete type. String-matching on `err.Error()` is equally fragile — it breaks on library version changes.
      ✗  `if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { ... }`
      ✓  `if myapp.ErrorCode(err) == myapp.ECONFLICT { ... }`

§3.7  `[MUST][CODE]`  Never return a raw database, network, or library error directly to an end user or across a domain interface boundary. At every domain-interface implementation boundary, translate all implementation-specific errors to `myapp.Error` with an appropriate `Code` before returning.
      Raw errors from external libraries can expose schema details or query text to attackers. Domain callers have no knowledge of implementation-specific error types (`sql.ErrNoRows`, `*pq.Error`); an untranslated error forces callers to import the implementing package to inspect it.
      ```go
      // ✗  Leaks sql.ErrNoRows into the domain
      if err := row.Scan(...); err != nil { return nil, err }
      // ✓  Translated at the implementation boundary
      if err := row.Scan(...); err == sql.ErrNoRows {
          return nil, &myapp.Error{Code: myapp.ENOTFOUND}
      } else if err != nil {
          return nil, &myapp.Error{Code: myapp.EINTERNAL, Err: err}
      }
      ```

§3.8  `[MUST][CODE]`  In every function that wraps an error, set `Op` to the logical operation name (`"TypeName.MethodName"` for methods, `"functionName"` for unexported functions) before returning. Implement `Error() string` to emit the `Op` chain first (colon-separated), followed by the leaf `Code` and `Message`, on a single line. Do not embed runtime stack traces in `Error` values.
      Without `Op` at each layer, a wrapped `EINTERNAL` that surfaces at the HTTP boundary carries no function name; the operator cannot tell from the error alone which call failed and must attach a debugger or scatter log lines to locate it — debug work that `Op`-chaining eliminates. A multi-line `Error()` format prevents grepping logs. Runtime stack traces expose irrelevant library frames and add significant allocation pressure at high request rates.
      ```go
      // ✗  No Op — operator cannot locate the failure
      return &myapp.Error{Err: err}
      // ✗  Op only at the outermost layer — intermediate operations invisible
      // ✓  Op at every wrapping layer; single-line output
      const op = "UserService.CreateUser"
      return &myapp.Error{Op: op, Err: err}
      // Error() output: "UserService.CreateUser: attachRole: <internal> syntax error near INSERT"
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `package errors; type Error struct { ... }` | ARCH | §3.1 | Stutter; error type leaves domain vocabulary | `package myapp; type Error struct { ... }` |
| `pkg/errors.Wrap(err, "msg")` as primary domain error | ARCH | §3.1 | Domain coupled to external library | Domain `myapp.Error` with `Op`+`Err` wrapping |
| `&myapp.Error{Op: op, Code: EINTERNAL, Err: err}` | CODE | §3.2 | `ErrorCode()` / `ErrorMessage()` traverse incorrectly | Separate wrapping node from leaf node |
| Premature fine-grained codes (`ECONFLICT_USERNAME`) | CODE | §3.3 | Proliferates vocabulary before need | Start with 4 generic codes; specialize on demand |
| Undocumented error codes on public functions | CODE | §3.3 | Callers cannot write exhaustive handling | Document every returned `Code` in function comment |
| Direct type assertion `err.(*myapp.Error).Code` | CODE | §3.4 | Panics if outer error is not `*myapp.Error`; doesn't traverse | `myapp.ErrorCode(err)` |
| `ErrorCode(nil)` returns `EINTERNAL` | CODE | §3.4 | Spurious internal-error branches on nil path | Return `""` for nil |
| `Message` on every wrapping node | CODE | §3.5 | `ErrorMessage()` returns outermost (least specific) message | `Message` only at origin of user-visible condition |
| `err.(*pq.Error)` in service or handler | CODE | §3.6 | Couples domain/handler to postgres; breaks on DB swap | `myapp.ErrorCode(err)` |
| `err.Error()` string-matching for branching | CODE | §3.6 | Fragile; breaks on library version changes | `myapp.ErrorCode(err)` |
| `return nil, sql.ErrNoRows` from repository | ARCH | §3.7 | Domain callers must import `database/sql` to inspect | `&myapp.Error{Code: myapp.ENOTFOUND}` |
| `return nil, err` without `Op` for a library error | CODE | §3.8 | `ErrorCode()` returns `EINTERNAL` but operator has no location context | `&myapp.Error{Op: op, Err: err}` |
| `Op` set only at the outermost wrapping layer | CODE | §3.8 | Intermediate operations invisible | `Op` at every function that wraps |
| `runtime.Stack()` captured into error values | CODE | §3.8 | Library frames obscure flow; high allocation cost | `Op`-chain logical trace |

---

## 4. CRUD Service Design

Service interfaces define the contract between the domain and the outside world. Each method is a self-contained unit of work. Callers interact with contracts, not implementations.

---

§4.1  `[MUST][ARCH]`  Each service method is a self-contained transactional unit. Never expose transactions, connections, or other implementation details across the service boundary — the caller must not be able to compose cross-service calls that appear atomic but are not.
      Leaking transactions allows callers to build pseudo-atomic sequences with no real atomicity guarantees. It also makes the underlying technology visible to every caller, preventing substitution.
      ✗  `BeginTx() *sql.Tx` method on the service interface
      ✓  Each `CreateDial`, `UpdateDial`, etc. opens and commits or rolls back its own transaction internally

§4.2  `[MUST][ARCH]`  Enforce authorization inside the service implementation; propagate the authenticated user through `context.Context` and push authorization restrictions into the data query rather than filtering in application code after retrieval. Every delete operation must enforce ownership or authorization — never delete by ID alone.
      A check placed only in an HTTP handler or middleware is bypassed by any direct caller of the service. Post-retrieval filtering fetches unauthorized rows, wastes I/O, and can leak existence information through timing. A delete with no ownership check allows any authenticated user to delete any record by guessing IDs.
      ✗  Middleware enforces access; service method trusts the header value directly
      ✗  `DELETE FROM dials WHERE id = ?` with no `user_id` restriction
      ✓  `WHERE user_id = ?` applied in SQL; `wtf.UserIDFromContext(ctx)` supplies the value

§4.3  `[MUST][CODE]`  When a caller requests a specific object by primary key, return either the object or an error — never `(nil, nil)`. When a collection search finds no matching records, return an empty slice and `nil` error — never a not-found error for an empty result.
      A `nil` object with a `nil` error compiles without forcing the caller to check for nil; callers that only check `err != nil` will panic on dereference. Conversely, returning `ENOTFOUND` for an empty search result forces callers to special-case a non-error condition — the absence of matches is not an error.
      ✗  `FindDialByID` returns `(nil, nil)` when no row exists → callers panic on dereference
      ✗  `FindDials` returns `(nil, ErrNotFound)` when the filter matches nothing → callers must special-case
      ✓  `FindDialByID` returns `&myapp.Error{Code: myapp.ENOTFOUND}`; `FindDials` returns `(nil, 0, nil)`

§4.4  `[SHOULD][CODE]`  Always populate parent-relationship objects on the returned entity; include child collections only when their size is bounded and they are nearly always needed alongside the parent.
      Parent objects are almost always required for display or authorization — fetching them inline avoids a second round-trip. Unbounded child collections can blow up payload size and query cost and should be fetched separately.
      ✗  Return `Dial` with no `User` field — every caller must make a second fetch for the owner
      ✓  Return `Dial` with `User` populated; include bounded `Memberships` but not open-ended lists

§4.5  `[MUST][CODE]`  Represent all search parameters in a filter struct with pointer fields for optional criteria; nil means "apply no restriction on this dimension." Return the total match count alongside the current page's results.
      Each new filter criterion added to a positional signature is a breaking API change. Non-pointer fields cannot distinguish "not supplied" from "zero value" — `ID = 0` means either "filter by ID zero" or "no ID filter." Pagination UI requires the total count regardless of page size; omitting it forces a second count query on every render.
      ✗  `FindDials(ctx, userID int, name string, offset, limit int)` — every new filter is breaking
      ✗  `type DialFilter struct { ID int }` — zero value is ambiguous
      ✓  `FindDials(ctx, filter DialFilter) ([]*Dial, int, error)` with `type DialFilter struct { ID *int; Limit int }`

§4.6  `[MUST][CODE]`  Map sort criteria to a fixed enumeration of SQL expressions; never interpolate an unvalidated sort string into a query.
      Passing raw user-supplied sort strings into SQL is an injection vector. Unindexed columns produce full-table scans that degrade under load.
      ```go
      // ✗  String interpolation — SQL injection; sort column chosen by attacker
      q := "SELECT ... FROM dials ORDER BY " + filter.SortBy

      // ✓  Enumerated sort expressions — validated, always a known indexed column
      switch filter.SortBy {
      case "updated_at_desc": sortBy = "dm.updated_at DESC"
      default:                sortBy = "dm.id ASC"
      }
      ```

§4.7  `[SHOULD][CODE]`  For create operations, accept a pointer to the new object and write generated fields (primary key, timestamps) back onto the pointed-to struct. Create parent and children atomically in a single service call.
      A function that returns a copy of the entity compels the caller to reconcile two representations; a caller that continues referencing the original struct reads stale field values — missing primary key, zero timestamps — and may silently pass them downstream. Requiring separate service calls for each child places the atomicity burden on the caller: a failure between calls leaves a parent entity with no children and no rollback path.
      ✗  `CreateDial(ctx, dial Dial) (Dial, error)` — caller must capture and switch to the returned copy
      ✓  `CreateDial(ctx, dial *Dial) error` — `dial.ID` populated in place; `dial.Memberships` persisted atomically

§4.8  `[MUST][CODE]`  For update operations, use a dedicated Update struct with pointer fields for each updatable attribute; pass the target ID as a separate argument. Return the updated (or attempted) object even when an error occurs.
      Pointer fields distinguish "set this field to its zero value" from "leave this field unchanged." Passing the full entity for update overwrites all fields unconditionally. Embedding the ID in the Update struct prevents reuse for bulk updates. A stateless HTTP handler needs the attempted state to re-render a form without a second fetch.
      ✗  `UpdateDial(ctx, dial *Dial) error` — all fields always overwritten; embedded ID prevents bulk
      ✓  `UpdateDial(ctx, id int, upd DialUpdate) (*Dial, error)` with `type DialUpdate struct { Name *string }`

§4.9  `[CONSIDER][CODE]`  Design delete and update signatures to accept a slice of IDs rather than a single ID so that bulk operations require no future API change.
      Changing `id int` to `ids []int` is a breaking change; designing for slices from the start keeps the option open at zero cost.
      ✗  `DeleteDial(ctx context.Context, id int) error` — a later bulk-delete requirement forces a breaking signature change
      ✓  `DeleteDials(ctx context.Context, ids []int) error` — single deletion passes `[]int{id}`

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `BeginTx()` method on service interface | ARCH | §4.1 | Pseudo-atomic cross-service calls with no real guarantees | Each service method owns its own transaction |
| Authorization only in HTTP middleware | ARCH | §4.2 | Direct service callers bypass the check | Enforce inside service implementation |
| Post-retrieval authorization filtering in application code | ARCH | §4.2 | Fetches unauthorized rows; existence leak via timing | Push restriction into SQL `WHERE` |
| `DELETE FROM table WHERE id = ?` with no ownership check | ARCH | §4.2 | Any authenticated user deletes any record | Restrict by `user_id` in query or service |
| `FindByID` returns `(nil, nil)` when no row exists | CODE | §4.3 | Callers that check only `err != nil` panic | Return typed not-found error |
| `FindAll` returns `(nil, ErrNotFound)` on empty result | CODE | §4.3 | Callers must special-case a non-error condition | Return empty slice + nil error |
| Bare entity returned with no parent fields | CODE | §4.4 | N+1 fetches in every caller | Populate parent relationships inline |
| Positional filter arguments | CODE | §4.5 | Each new filter is a breaking signature change | Filter struct |
| Non-pointer optional filter fields | CODE | §4.5 | Zero value is ambiguous; can't express "unset" | Pointer fields; nil = no restriction |
| Paged results returned without total count | CODE | §4.5 | Pagination impossible without a second query | Return count alongside results |
| Raw sort string interpolated into SQL | CODE | §4.6 | SQL injection; full-table scan | Fixed enumeration of sort expressions |
| `CreateDial(ctx, Dial) (Dial, error)` returning a copy | CODE | §4.7 | Caller must discard original; extra allocation | `CreateDial(ctx, *Dial) error`; mutate in place |
| Multi-step creation exposed to callers | CODE | §4.7 | Partial failure leaves graph inconsistent | Parent + children atomically in one service call |
| Full entity struct passed for update | CODE | §4.8 | All fields unconditionally overwritten | Dedicated Update struct with pointer fields |
| Target ID embedded in Update struct | CODE | §4.8 | Prevents reuse for bulk updates | ID as separate argument |
| `(nil, err)` returned on validation failure in update | CODE | §4.8 | Stateless callers lose submitted state | Return attempted object alongside error |

---

## 5. SQL Implementation

The service interface boundary is also the transactional boundary. Below it, SQL logic lives in package-private helpers that are composable, reusable, and independent of any particular service.

---

§5.1  `[MUST][CODE]`  Service methods are thin transactional shells: begin transaction, call package-private helper functions, commit or rollback. SQL query logic lives in helpers that accept `*Tx` as a parameter, not as a service receiver; helpers are reusable across service methods and callable by other helpers.
      Embedding SQL logic directly in service methods makes it unreusable: when two service methods both need to attach an owner `User` to an entity, duplicating the SQL produces two sources of truth. Helpers attached to `*Tx` rather than to a service receiver can be called by any service or by other helpers within an existing transaction.
      ```go
      // ✓  Service = thin shell; SQL logic in package-private *Tx helper
      func (s *DialService) FindDialByID(ctx context.Context, id int) (*wtf.Dial, error) {
          tx, err := s.db.BeginTx(ctx, nil)
          if err != nil { return nil, err }
          defer tx.Rollback()
          dial, err := findDialByID(ctx, tx, id)   // package-private helper
          if err != nil { return nil, err }
          return dial, tx.Commit()
      }
      func findDialByID(ctx context.Context, tx *Tx, id int) (*wtf.Dial, error) { ... }
      ```

§5.2  `[SHOULD][CODE]`  Implement single-ID lookup helpers by wrapping the collection helper and translating an empty result to `ENOTFOUND`; do not duplicate the SQL in an "optimized" version.
      Duplicating query code to save one slice allocation makes it harder to maintain two sources of truth. Optimize only after profiling identifies a hot path. The semantic difference between collection and single-ID lookup is cleanly expressed at one point in the single-ID helper.
      ```go
      // ✓  Reuses findDials; translates empty result to ENOTFOUND at one point
      func findDialByID(ctx context.Context, tx *Tx, id int) (*wtf.Dial, error) {
          dials, _, err := findDials(ctx, tx, wtf.DialFilter{ID: &id})
          if err != nil { return nil, err }
          if len(dials) == 0 { return nil, &wtf.Error{Code: wtf.ENOTFOUND, Message: "Dial not found."} }
          return dials[0], nil
      }
      ```

§5.3  `[MUST][CODE]`  Build SQL `WHERE` clauses by appending predicate strings to a slice and joining with `AND`; use `1 = 1` as the base predicate; bind all values through `?` (or `$N`) placeholders — never by string interpolation. (Sort criteria cannot be parameterized; map them to a fixed enumeration per §4.6.)
      String interpolation is the primary SQL injection vector. An always-true base predicate simplifies joining: the clause is valid regardless of how many predicates are appended, and SQL optimizers ignore trivially-true conditions.
      ```go
      where := []string{"1 = 1"}
      var args []interface{}
      if v := filter.ID; v != nil {
          where = append(where, "id = ?")
          args = append(args, *v)
      }
      q := `SELECT ... FROM dials WHERE ` + strings.Join(where, " AND ")
      ```
      ✗  `"WHERE name = '" + filter.Name + "'"` — SQL injection
      ✓  Predicate slice + `strings.Join(where, " AND ")` + parameterized bind variables

§5.4  `[SHOULD][CODE]`  Use `COUNT(*) OVER()` SQL windowing to compute the total row count in the same query as the paginated results; do not execute a separate `COUNT` query.
      A second count query doubles database round-trips per paginated request and can return a stale count if a write occurs between the two queries. `COUNT(*) OVER()` is supported by SQLite, PostgreSQL, and most modern RDBMS; scan its value on every row and use the last.
      ✗  Two separate queries — one for the page, one for the total:
      ```go
      db.QueryRowContext(ctx, "SELECT COUNT(*) FROM dials WHERE user_id = ?", uid).Scan(&total)
      rows, _ = db.QueryContext(ctx, "SELECT id, name FROM dials WHERE user_id = ? OFFSET ? LIMIT ?", uid, offset, limit)
      ```
      ✓  `SELECT id, name, COUNT(*) OVER() FROM dials WHERE user_id = ? OFFSET 40 LIMIT 20`

§5.5  `[SHOULD][CODE]`  Wrap `*sql.DB` and `*sql.Tx` in application-specific types that add domain-level methods; place validation for a write operation on the transaction method that performs the write. Do not provide batch variants of single-item transaction methods — callers compose multiple calls within one transaction.
      Callers using raw `*sql.DB`/`*sql.Tx` directly are forced to know the `database/sql` API; adding application-level behaviour (instrumentation, validation) then requires changes at every call site. Validation scattered across callers is duplicated and inconsistently applied. Batch methods duplicate the single method's logic and must be maintained in sync.
      ```go
      type Tx struct{ *sql.Tx }
      func (tx *Tx) CreateUser(u *User) error {
          if u == nil  { return errors.New("user required") }
          if u.Name == "" { return errors.New("name required") }
          _, err := tx.Exec(`INSERT INTO users ...`)
          return err
      }
      // ✗  func (tx *Tx) CreateUsers(users []*User) error — duplicates CreateUser logic
      // ✓  Callers loop over CreateUser inside one transaction
      ```

§5.6  `[SHOULD][CODE]`  Use `database/sql` directly; avoid ORM libraries.
      ORM tools help with trivial access patterns but make advanced queries and debugging more difficult. Direct `database/sql` keeps SQL visible, debuggable, and free of framework conventions.
      ✗  `db.Where("user_id = ?", uid).Find(&dials)` — ORM hides the SQL; advanced queries require ORM idioms
      ✓  `rows, err := tx.QueryContext(ctx, "SELECT ... FROM dials WHERE user_id = ?", uid)`

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| SQL query logic embedded in service methods | CODE | §5.1 | Unreusable; duplicated when needed across services | Package-private helper functions accepting `*Tx` |
| Separate "optimized" `findByIDFast` duplicating collection SQL | CODE | §5.2 | Two sources of truth; must be maintained in sync | Wrap collection helper; translate empty result to ENOTFOUND |
| String interpolation in SQL predicates | CODE | §5.3 | SQL injection | Parameterized placeholders + predicate slice |
| Separate `COUNT(*)` query for total pagination count | CODE | §5.4 | Extra round-trip; stale count if write occurs between queries | `COUNT(*) OVER()` in original query |
| Validation inline at every call site before `db.Exec` | CODE | §5.5 | Duplicated; one caller can omit it | Validate inside the transaction method |
| `CreateUsers([]*User)` alongside `CreateUser(*User)` | CODE | §5.5 | Logic duplicated; both must be maintained in sync | Loop single-item method inside one transaction |
| ORM library for data access | CODE | §5.6 | SQL obscured; advanced queries require ORM idioms | `database/sql` directly |

---

## 6. Testing

Tests are the first callers of the application's public API. Their structural requirements — isolation, reproducibility, minimal dependencies — follow directly from treating them as first-class callers.

---

§6.1  `[MUST][METHOD]`  Design every test to be self-contained and independently reproducible: it must not depend on state created or modified by another test, and it must produce the same result whether run in isolation or in the full suite, without any manual environment setup.
      Tests that share mutable state produce order-dependent failures; diagnosing a failure requires understanding the entire suite rather than just the failing function. Tests that require manual setup fail silently in CI and on fresh checkouts.
      ✗  Package-level `var db *DB` shared by all tests; `TestMain` that seeds shared state
      ✓  Each test constructs its own independent resource via `NewTestDB()` + `defer db.Close()`

§6.2  `[MUST][CODE]`  Use only the standard `testing` package; do not use third-party testing frameworks (BDD runners, assertion libraries such as `testify`, `ginkgo`, `goconvey`).
      A framework another contributor must install before running tests raises the contribution barrier. Framework-specific assertion syntax produces unfamiliar failure messages that slow diagnosis. *(Overrides the common recommendation of `testify` — apply as stated.)*
      ✗  `assert.Equal(t, expected, actual)` — requires `testify` import
      ✓  `if expected != actual { t.Fatalf("expected %v, got %v", expected, actual) }`

§6.3  `[SHOULD][CODE]`  Reduce test verbosity with three locally-defined helper functions rather than importing an assertion library:
      ```go
      func assert(tb testing.TB, condition bool, msg string, v ...interface{})
      func ok(tb testing.TB, err error)
      func equals(tb testing.TB, exp, act interface{})
      ```
      These three cover the vast majority of assertion needs without an external dependency and can be copied into any package unchanged.

§6.4  `[MUST][CODE]`  Name test files that test a package's exported API with the `package foo_test` declaration; do not use dot-imports in external test packages.
      Testing from inside the package (`package foo`) gives access to unexported fields and functions; tests that rely on unexported internals are brittle — any refactor that preserves exported behaviour but changes internals breaks them. Dot-imports obscure which package a symbol comes from; tests are the primary documentation of API usage.
      ✗  `package myapp` in `user_test.go` — unexported fields accessible; tests couple to internals
      ✗  `. "github.com/benbjohnson/myapp"` — symbol origin hidden from readers
      ✓  `package myapp_test; import "github.com/benbjohnson/myapp"`

§6.5  `[MUST][CODE]`  For any resource that requires setup and cleanup (temp files, database connections, test servers), define a test-specific wrapper type with a constructor that opens the resource and a `Close()` method that cleans it up; call `defer resource.Close()` immediately after construction. Use `panic` (not `t.Fatal`) in test helper constructors when initialization fails.
      Setup/teardown code duplicated inline in every test is a maintenance liability; one missed cleanup leaks state into other tests. A constructor that calls `t.Fatal` requires `*testing.T` as a parameter, couples the constructor to the test framework, and cannot be used in `TestMain`.
      ```go
      type TestDB struct{ *DB }
      func NewTestDB() *TestDB {
          db, err := Open(tempPath(), 0600)
          if err != nil { panic(err) }
          return &TestDB{db}
      }
      func (db *TestDB) Close() { defer os.Remove(db.Path()); db.DB.Close() }
      // Usage: db := NewTestDB(); defer db.Close()
      ```

§6.6  `[MUST][ARCH]`  Define interfaces for external dependencies in the package that *uses* them, not in the package that *provides* them; the interface should contain only the methods the consuming package actually calls.
      An interface declared in the provider exposes the full API surface and forces all consumers to satisfy it even if they use only a subset. A consumer-side interface with only the required methods makes the mock smaller and keeps the consumer independently testable without any code change in the provider.
      ✗  `type MyApplication struct { YoClient *yo.Client }` — concrete external type; test must construct real client
      ✓  `type MyApplication struct { YoClient interface{ Send(string) error } }` — minimal, consumer-declared

§6.7  `[MUST][CODE]`  Place mocks according to where the interface is defined: when an interface is declared in the root domain package, place its mock in a shared `myapp/mock/` subpackage using the function-field pattern; when an interface is declared inline in a consuming package, place the mock in the test file that uses it.
      Scattered mocks for root-domain interfaces are duplicated across packages; when the interface gains a method, every scattered copy must be updated independently. Conversely, a shared mock for an inline consumer interface forces the consumer to export that interface, defeating the purpose of declaring it inline.
      ⚠ "Standard Package Layout" prescribes shared `mock/` for all interfaces; "Structuring Tests in Go" prescribes test-file mocks for inline interfaces. This conditional form is the synthesis.
      ```go
      // ✓  Shared mock/ for root-domain interface (function-field pattern)
      type UserService struct {
          FindUserByIDFn      func(ctx context.Context, id int) (*myapp.User, error)
          FindUserByIDInvoked bool
      }
      func (s *UserService) FindUserByID(ctx context.Context, id int) (*myapp.User, error) {
          s.FindUserByIDInvoked = true
          return s.FindUserByIDFn(ctx, id)
      }
      // ✓  Test-file mock for inline consumer interface
      type testYoClient struct{ SendFunc func(string) error }
      func (c *testYoClient) Send(s string) error { return c.SendFunc(s) }
      ```

§6.8  `[SHOULD][CODE]`  Give mock function fields the suffix `Func` (e.g. `SendFunc func(string) error`) so that each test assigns only the methods it needs, leaving others as `nil`; a nil `Func` that is called panics immediately and identifies the unexpected call. Add application-specific convenience methods to test wrapper types for preconditions that many tests share.
      A mock where every method must be pre-populated requires the test to stub methods it does not exercise. Precondition logic repeated in every test function is a maintenance liability.
      ✗  All mock methods pre-populated; every test stubs `ListUsers` even when only `FindUserByID` is tested
      ✓  `svc.FindUserByIDFunc = func(ctx, id) (*myapp.User, error) { return testUser, nil }` — other fields left nil

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Tests share a package-level DB or config variable | METHOD | §6.1 | Order-dependent failures; parallel tests corrupt each other | Per-test `NewTestDB()` |
| Test requires external service or manually seeded DB | METHOD | §6.1 | Fails on fresh checkout; not reproducible in CI | Temp files, in-process DB, or mocks |
| `github.com/stretchr/testify` or BDD framework | CODE | §6.2 | External dependency; contributor barrier | Local `ok`/`equals`/`assert` helpers |
| `package myapp` in test file testing exported API | CODE | §6.4 | Unexported internals accessible; refactors break tests | `package myapp_test` |
| Dot-import in `_test` package | CODE | §6.4 | Symbol origin obscured; API usage misleading to readers | Explicit qualified names |
| Setup/teardown logic inline in each test function | CODE | §6.5 | Duplication; missed cleanup leaks state | Test wrapper type with `NewTestX()` + `defer x.Close()` |
| Test helper constructor calls `t.Fatal` | CODE | §6.5 | Coupled to test framework; unusable in `TestMain` | `panic` on init failure |
| Struct field typed as `*yo.Client` (concrete external type) | ARCH | §6.6 | Test must construct or stub a real external dependency | Inline consumer-side interface |
| Interface for external dependency declared in provider package | ARCH | §6.6 | Consumer forced to satisfy full surface; unused methods on mock | Caller-side minimal interface |
| Root-domain interface mock defined in each consuming `_test.go` | ARCH | §6.7 | Duplicated; interface changes require updating all copies | Shared `myapp/mock/` subpackage |
| Shared `mock/` package for an inline consumer interface | CODE | §6.7 | Forces export of inline interface; defeats its purpose | Mock in test file |
| Mock function fields all pre-populated; no `Func` suffix | CODE | §6.8 | Every test stubs every method even when only one is tested | `XxxFunc func(...)` pattern; nil = not called |

---

## Master Anti-patterns Table

This table synthesizes the per-section tables. Rows that address the same cross-section failure are merged; only concerns not already captured by a more specific row from a single section appear here.

| Anti-pattern | Levels | Rules violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Domain types outside root package (in `users/`, `models/`, etc.) or stutter names (`users.UserService`) | ARCH | §1.1, §1.3 | Wrong architectural boundary; circular import pressure as app grows | All domain types and service interfaces in root package |
| External I/O or intra-app import in root package | ARCH | §1.2 | Domain coupled to technology; full stack required to unit-test | Push all I/O into named dependency subpackage |
| Generic or ambiguous subpackage name (`storage/`, `db/`) | ARCH | §1.4 | Multiple backends collide; ownership unclear | One subpackage per dependency, named after it |
| Subpackage exposes driver or ORM types in public API | ARCH | §1.5 | Every caller imports the implementation; swap requires all-caller changes | Accept and return domain types only |
| Direct import between two dependency subpackages | ARCH | §1.6 | Two implementations coupled; circular import risk on growth | Route all inter-subpackage communication through root-package domain interface |
| Binary requires a system runtime, shared library, or non-portable runtime | ARCH | §1.7 | Deployment fails on clean host; CI cannot validate host state | Statically compiled single binary; single-process deployment at small scale |
| Subpackage created for file-count or line-count reasons | ARCH | §1.8 | Unexported types forced public; cyclic import pressure | Create subpackage only when a real dependency boundary exists |
| Global mutable state (`var db`) or bare handler functions reading globals | ARCH, CODE | §2.1 | Tests share mutable state; parallel tests corrupt each other; no dependency injection point | Struct-field injection; struct handler types with `ServeHTTP` |
| Wiring outside `main` (`main.go` at root; construction or business logic inside/outside `main`) | ARCH | §2.2, §2.3 | Packages couple domain + implementation; logic untestable and non-reusable across binaries | `cmd/<name>/main.go`; construction and wiring exclusively in `main` |
| Application `Error` type defined outside root package or outsourced to a third-party library | ARCH | §3.1 | Stutter name; error vocabulary leaves domain; domain coupled to external library | `package myapp; type Error struct { Code, Message, Op string; Err error }` |
| Wrapping-node and leaf-node fields mixed on one `Error` instance | CODE | §3.2 | `ErrorCode()` / `ErrorMessage()` chain traversal returns wrong result | Set either `Err` (wrapping node) or `Code`/`Message` (leaf) — never both |
| Fine-grained error codes added before a caller needs to distinguish them, or undocumented codes on public functions | CODE | §3.3 | Dead `case` branches; unstable API contract; callers cannot write exhaustive handling | Start with 4 generic codes; add only on proven caller need; document all returned codes |
| Error inspected via type assertion (`err.(*pq.Error)`, `err.(*myapp.Error).Code`) or `err.Error()` string match | CODE | §3.4, §3.6 | Panics if type does not match; couples caller to implementation; breaks on library version change | `myapp.ErrorCode(err)` / `myapp.ErrorMessage(err)` |
| Implementation-specific error (`sql.ErrNoRows`, `*pq.Error`) surfaced across domain-interface boundary | ARCH | §3.7 | Domain callers must import the implementing package to inspect the error; schema details leak | Translate to `myapp.Error` with appropriate `Code` at every implementation boundary |
| `Op` missing or set only at the outermost wrapping layer | CODE | §3.8 | Wrapped `EINTERNAL` carries no function name; operator cannot locate failure without a debugger | Set `Op` in every function that wraps an error |
| Service interface exposes transactions or implementation details (`BeginTx()` method) | ARCH | §4.1 | Callers build pseudo-atomic sequences with no real atomicity guarantee; technology visible to caller | Each service method opens, commits, and rolls back its own transaction internally |
| Authorization enforced outside service (middleware only, post-retrieval filter, or delete with no ownership check) | ARCH | §4.2 | Direct service callers bypass check; unauthorized rows fetched; any user deletes any record | Enforce inside service; push restriction into SQL `WHERE`; require `user_id` on every delete |
| `FindByID` returns `(nil, nil)` for missing row, or `FindAll` returns a not-found error for empty result | CODE | §4.3 | `nil` with no error causes panic on dereference; not-found error forces callers to special-case a non-error condition | `FindByID` → typed `ENOTFOUND` error; `FindAll` → empty slice + `nil` error |
| Filter arguments passed positionally, or filter struct with non-pointer fields, or paged results returned without total count | CODE | §4.5 | Every new filter is a breaking change; zero value is ambiguous; pagination UI requires a second query | Filter struct with pointer fields; `nil` = no restriction; return count alongside results |
| String interpolation in SQL predicates or `ORDER BY` expressions | CODE | §4.6, §5.3 | SQL injection; unindexed columns produce full-table scans | Parameterized placeholders for values; fixed enumeration of sort expressions |
| Create returns a copy of the entity, or children created in separate service calls | CODE | §4.7 | Caller referencing original struct reads stale fields; separate calls leave graph inconsistent on partial failure | `Create(ctx, *Entity) error` — mutate in place; create parent + children in one atomic service call |
| Full entity struct passed for update, or ID embedded in Update struct, or `(nil, err)` on validation failure | CODE | §4.8 | All fields unconditionally overwritten; bulk updates require duplicate signatures; stateless caller loses submitted state | Dedicated `EntityUpdate` struct with pointer fields; ID as separate argument; return attempted object alongside error |
| SQL query logic embedded in service methods or duplicated for "optimized" single-ID variants | CODE | §5.1, §5.2 | Two sources of truth; any change must be applied in every duplicate; single-ID variant drifts from collection query | Package-private helpers accepting `*Tx`; single-ID helper wraps collection helper |
| Separate `COUNT(*)` query for pagination total | CODE | §5.4 | Extra round-trip doubles DB cost; stale count if write occurs between queries | `COUNT(*) OVER()` in the same paginated query |
| Validation scattered across callers, or `CreateUsers` batch method duplicating `CreateUser` logic | CODE | §5.5 | One caller omits validation; batch and single methods diverge silently | Validate inside the `Tx` method; callers loop single-item method in one transaction |
| ORM library for data access | CODE | §5.6 | SQL obscured; advanced queries require ORM idioms; debugging requires framework knowledge | `database/sql` directly |
| Tests share mutable state or depend on external services / manually seeded data | METHOD | §6.1 | Order-dependent failures; suite does not reproduce on fresh checkout or in CI | Per-test construction via `NewTestDB()`; teardown via `defer db.Close()` |
| Third-party testing framework or wrong test-file package declaration | CODE | §6.2, §6.4 | External dependency raises contribution bar; internals accessible from test file; refactors break tests | `testing` package only; `package myapp_test` for exported-API tests |
| Setup/teardown inline in each test function, or test constructor calling `t.Fatal` | CODE | §6.5 | Duplication; missed cleanup leaks state; constructor coupled to test framework, unusable in `TestMain` | Wrapper type with `NewTestX()` + `defer x.Close()`; `panic` on init failure |
| Dependency injected as concrete external type or interface declared in provider package | ARCH | §6.6 | Test must construct real dependency; consumer forced to satisfy full interface surface | Consumer-side minimal interface with only the methods the consumer calls |
| Root-domain interface mock scattered across test files, or inline-interface mock placed in shared `mock/` package | ARCH, CODE | §6.7 | Interface changes require updating every scattered copy; shared mock forces export of inline interface | Root-domain mocks in `myapp/mock/`; inline-interface mocks in the test file that uses them |
| Mock `Func` fields all pre-populated; no `Func` suffix convention | CODE | §6.8 | Every test must stub every method even when only one is exercised; unexpected calls fail silently | `XxxFunc func(...)` fields; leave unused fields `nil` — nil call panics and identifies the unexpected call immediately |

---

## Appendix: Coverage Accounting

Every rule from every input ruleset is accounted for below.

### CRUD Rules (18 input rules)

| Input | Output | Case |
|---|---|---|
| crud §1.1 — service interface in root domain package | §1.1 | B — subsumed; root package owns all domain interfaces |
| crud §1.2 — service is black box; no exposed transactions | §4.1 | E |
| crud §1.3 — auth via context; enforced inside service | §4.2 | A — merged with §1.4 and §6.1 |
| crud §1.4 — push auth restrictions into SQL WHERE | §4.2 | A — merged |
| crud §2.1 — keyed lookup: entity or error, never (nil,nil) | §4.3 | C — complementary to §3.1; merged into one rule |
| crud §2.2 — populate parent relationships inline | §4.4 | E |
| crud §3.1 — empty search is not an error | §4.3 | C — merged with §2.1 |
| crud §3.2 — filter struct for search params | §4.5 | A — merged with §3.3 and §3.4 |
| crud §3.3 — pointer fields in filter struct | §4.5 | A — merged |
| crud §3.4 — return total count alongside paged results | §4.5 | A — merged |
| crud §3.5 — sort criteria mapped to enumeration; no interpolation | §4.6 | A — merged with RWS WHERE clause building |
| crud §4.1 — accept *entity for create; write generated fields back | §4.7 | A — merged with §4.2 |
| crud §4.2 — create parent+children atomically | §4.7 | A — merged |
| crud §5.1 — dedicated Update struct with pointer fields | §4.8 | E |
| crud §5.2 — return attempted object even on update failure | §4.8 | B — subsumed into §4.8 as requirement |
| crud §5.3 — separate ID arg enables future bulk update | §4.9 | A — merged with §6.2 |
| crud §6.1 — ownership check on every delete | §4.2 | B — subsumed into §4.2 auth rule |
| crud §6.2 — accept slice of IDs for delete | §4.9 | A — merged with §5.3 |

### Failure Is Your Domain (17 input rules)

| Input | Output | Case |
|---|---|---|
| failure §1.1 — Error type in root domain package | §3.1 | A — merged with §1.1 and §2.1 |
| failure §1.2 — no third-party error wrapping library | §3.1 | E — retained in §3.1 |
| failure §2.1 — Error struct with 4 fields (Code, Message, Op, Err) | §3.1 | A — merged |
| failure §2.2 — never mix Err and Code/Message on one instance | §3.2 | E |
| failure §2.3 — document which codes each function returns | §3.3 | B — subsumed into §3.3 |
| failure §3.1 — start with 4 generic codes; expand on demand | §3.3 | A — merged with §2.3 |
| failure §3.2 — ErrorCode() traversal helper with specified behaviors | §3.4 | E |
| failure §3.3 — ErrorCode(nil) returns "" not EINTERNAL | §3.4 | B — subsumed into §3.4 spec |
| failure §4.1 — branch on ErrorCode(), never type-assert | §3.6 | E |
| failure §4.2 — never surface raw implementation error to user | §3.7 | A — merged with §5.1 |
| failure §4.3 — ErrorMessage() traversal helper | §3.5 | E |
| failure §4.4 — Message only on leaf nodes | §3.5 | B — subsumed into §3.5 |
| failure §5.1 — translate implementation errors at domain boundary | §3.7 | A — merged with §4.2 |
| failure §5.2 — wrap unknown errors as EINTERNAL with Op | §3.8 | B — subsumed into §3.8 Op-wrapping rule |
| failure §6.1 — set Op at every wrapping layer | §3.8 | A — merged with §6.2 and §6.3 |
| failure §6.2 — Error() emits Op chain on single line | §3.8 | A — merged |
| failure §6.3 — no runtime stack traces | §3.8 | A — merged |

### Standard Package Layout (17 input rules)

| Input | Output | Case |
|---|---|---|
| standard §1.1 — domain types in root package | §1.1 | E |
| standard §1.2 — root package no intra-app imports | §1.2 | A — merged with §1.3 |
| standard §1.3 — root package no external I/O imports | §1.2 | A — merged |
| standard §1.4 — domain type must be pure domain (no I/O) | §1.2 | B — subsumed by §1.2 |
| standard §1.5 — no stutter names | §1.3 | E |
| standard §2.1 — one subpackage per external dependency | §1.4 | A — merged with §2.4 |
| standard §2.2 — subpackage exposes only domain types | §1.5 | E |
| standard §2.3 — cross-subpackage via domain interface | §1.6 | A — merged with packages_as_layers flat-hierarchy |
| standard §2.4 — stdlib I/O packages treated as external deps | §1.4 | A — merged with §2.1 |
| standard §2.5 — layer implementations via domain interface | §1.6 | B — subsumed into §1.6 |
| standard §3.1 — mocks in shared `mock/` subpackage | §6.7 | D — conflict with tests §5.2; conditional form resolves |
| standard §3.2 — function-field mock pattern | §6.7–§6.8 | E |
| standard §3.3 — tests receive mocks via domain interface | §6.6 | B — subsumed by caller-side interface rule |
| standard §4.1 — binary in `cmd/<name>/main.go` | §2.2 | A — merged with structuring_apps §2.1 |
| standard §4.2 — wire exclusively in main | §2.3 | A — merged with structuring_apps §1.3 |
| standard §4.3 — main contains only wiring code | §2.3 | B — subsumed into §2.3 |
| standard §4.4 — only main imports both domain and implementation | §2.3 | B — subsumed into §2.3 |

### Structuring Applications in Go (13 input rules)

| Input | Output | Case |
|---|---|---|
| apps §1.1 — no global application state | §2.1 | E |
| apps §1.2 — HTTP handlers as structs with ServeHTTP | §2.1 | B — special case of §2.1; subsumed |
| apps §1.3 — inject dependencies at construction in main | §2.3 | B — subsumed into §2.3 wiring rule |
| apps §2.1 — binary under `cmd/`; not at repo root | §2.2 | A — merged with standard §4.1 |
| apps §2.2 — core logic as library consumed by cmd/ | §2.3 | B — subsumed: main=wiring-only implies core is elsewhere |
| apps §2.3 — separate binaries for separate entry points | §2.4 | E |
| apps §3.1 — wrap *sql.DB/*sql.Tx in application types | §5.5 | E |
| apps §3.2 — validation on the transaction method | §5.5 | B — subsumed into §5.5 |
| apps §3.3 — no batch methods; loop single method in one transaction | §5.5 | B — subsumed into §5.5 anti-patterns |
| apps §4.1 — no subpackage for file-count reasons | §1.8 | A — merged with §4.4 |
| apps §4.2 — group related types; 200–500 SLOC target | §1.9 | A — merged with §4.3 |
| apps §4.3 — most important type first in file | §1.9 | B — subsumed into §1.9 |
| apps §4.4 — evaluate independent packages at ~10K SLOC | §1.8 | A — merged with §4.1 |

### Structuring Tests in Go (14 input rules)

| Input | Output | Case |
|---|---|---|
| tests §1.1 — self-contained tests | §6.1 | A — merged with §1.2 |
| tests §1.2 — independently reproducible tests | §6.1 | A — merged |
| tests §2.1 — stdlib testing package only | §6.2 | E |
| tests §2.2 — three local helpers (assert, ok, equals) | §6.3 | E |
| tests §3.1 — `package foo_test` for exported API tests | §6.4 | A — merged with §3.2 |
| tests §3.2 — no dot-imports in external test packages | §6.4 | A — merged |
| tests §3.3 — same-package test for unexported helpers only | §6.4 | B — subsumed as CONSIDER annotation |
| tests §4.1 — test wrapper types with NewTestX + defer Close | §6.5 | A — merged with §4.2 |
| tests §4.2 — panic (not t.Fatal) in test constructors | §6.5 | A — merged |
| tests §4.3 — convenience methods on test wrappers | §6.8 | B — subsumed into §6.8 |
| tests §5.1 — caller-side interface with only required methods | §6.6 | E |
| tests §5.2 — inline interface mock belongs in test file | §6.7 | D — conflict with standard §3.1; conditional form resolves |
| tests §5.3 — Func suffix on mock function fields | §6.8 | E |
| tests §5.4 — inline interfaces for stdlib mocking | §6.8 | B — subsumed as extension of §6.8 |

### WTF Dial (3 input rules)

| Input | Output | Case |
|---|---|---|
| wtf_dial §1.1 — single statically-compiled binary | §1.7 | A — merged with §1.2 |
| wtf_dial §1.2 — single-process at small scale | §1.7 | A — merged |
| wtf_dial §2.1 — build before document | dropped | E — fails replacement test: "document as you build" is defensible; not a Go code constraint |

### Packages as Layers (derived from source; 4 rules)

| Input | Output | Case |
|---|---|---|
| PAL-1 — packages are layers not groups (mental model) | §1.4, §1.6 | B — naming by dependency (§1.4) and routing via domain interface (§1.6) express this concretely |
| PAL-2 — root package as business domain layer | §1.1 | A — identical to standard §1.1 |
| PAL-3 — service interfaces in domain package enable flat hierarchy | §1.1, §1.6 | B — subsumed by §1.1 and §1.6 |
| PAL-4 — domain types named after domain, not implementation | §1.3 | A — identical to standard §1.5 (no stutter names) |

### Real-World SQL Part I (derived from source; 7 rules)

| Input | Output | Case |
|---|---|---|
| RWS-1 — service methods = transactional boundaries | §4.1 | A — identical to crud §1.2; merged |
| RWS-2 — service = thin shell; SQL in package-private *Tx helpers | §5.1 | E |
| RWS-3 — helper functions reusable across multiple services | §5.2 | E |
| RWS-4 — WHERE clause: predicate slice + 1=1 + parameterized values | §5.3 | C — complements crud §3.5 (sort enumeration); addresses SQL injection from the value side; merged into §5.3 with §4.6 covering sort |
| RWS-5 — COUNT(*) OVER() for total pagination count | §5.4 | E |
| RWS-6 — use database/sql directly; no ORM | §5.6 | E |
| RWS-7 — don't duplicate SQL for single-ID optimization; reuse helpers | §5.2 | A — merged with RWS-3 |
