# Ruleset: Structuring Applications in Go

```
Source: "Structuring Applications in Go" — Ben Johnson (2014)
Scope:  Go; application architecture; HTTP handlers; package layout;
        database wrapping; multi-binary projects; testability
```

---

## 1. Eliminate Global State

Handler functions that rely on global variables cannot be tested in isolation
and create hidden coupling between otherwise independent parts of the application.

---

§1.1  **[MUST][ARCH]**  Do not use global variables to hold application state
      (database connections, configuration, loggers) that HTTP handlers or other
      components depend on.
      Global state makes unit tests order-dependent and mutually interfering;
      a test that modifies a global connection or config value contaminates every
      other test that runs in the same process.
      ```go
      // ✗  Handler reaches for global state
      var db *sql.DB

      func hello(w http.ResponseWriter, r *http.Request) {
          row := db.QueryRow("SELECT myname FROM mytable")
          ...
      }

      // ✓  Dependencies carried on the handler struct
      type HelloHandler struct {
          db *sql.DB
      }

      func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
          row := h.db.QueryRow("SELECT myname FROM mytable")
          ...
      }
      ```

§1.2  **[MUST][CODE]**  Implement HTTP handlers as struct types with
      `ServeHTTP(w, r)` methods, not as bare functions registered via
      `http.HandleFunc`.
      A bare function handler has no receiver and therefore no way to hold
      injected dependencies; the only escape is a global variable, which §1.1
      prohibits.
      *(Overrides the net/http package's own documentation examples — apply as
      stated.)*

§1.3  **[MUST][CODE]**  Construct handlers in `main` (or an equivalent wiring
      layer) and inject dependencies as struct fields; do not have the handler
      open its own database connection or read its own config.
      A handler that sources its own dependencies cannot have those dependencies
      substituted in tests; the test is forced to provision real infrastructure.
      ```go
      // ✓  Dependencies injected at construction in main
      db, err := sql.Open("postgres", os.Getenv("DB"))
      ...
      http.Handle("/hello", &HelloHandler{db: db})
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `var db *sql.DB` at package level used by handlers | ARCH | §1.1 | Tests share global state; parallel tests corrupt each other | Struct field injected at construction |
| `http.HandleFunc("/path", handlerFunc)` where `handlerFunc` uses globals | CODE | §1.2 | No injection point; global dependency is the only option | `http.Handle("/path", &MyHandler{dep: dep})` |
| Handler calls `sql.Open(...)` inside `ServeHTTP` | CODE | §1.3 | New connection per request; untestable without real DB | Inject `*sql.DB` as a struct field |
| Global logger or config var read inside handler | ARCH | §1.1 | Tests cannot control logger/config without mutating global | Pass logger/config as struct fields |

---

## 2. Separate Binary from Application Library

Placing `main.go` at the repository root conflates application entry with
library API and prevents the package from being imported as a library or
extended with additional binaries.

---

§2.1  **[MUST][ARCH]**  Place each application binary in its own subdirectory
      under `cmd/`; do not place a `main` package at the repository root.
      A `main` package at the root prevents `go get ./...` from installing cleanly
      as a library, limits the project to one binary, and conflates entry-point
      concerns with core logic.
      ```
      ✗  myapp/main.go          // root is both library and binary

      ✓  myapp/cmd/myapp/main.go
      ✓  myapp/cmd/myapp-ctl/main.go
      ```

§2.2  **[MUST][ARCH]**  Write all core logic as if it were a library that
      `cmd/*/main.go` consumes; the binary is a thin client of the library.
      Core logic inside a `main` package cannot be imported or tested by anything
      other than the binary itself; it cannot be reused across multiple binaries
      and cannot be tested without running the full binary.

§2.3  **[SHOULD][ARCH]**  When the same logic should be accessible in multiple
      ways (CLI and HTTP server, for example), represent each as a separate binary
      under `cmd/`, all importing the same core library package.
      A single binary that tries to be both CLI and server requires flags or env
      vars to switch modes, making the entry-point complex and the two interfaces
      harder to evolve independently.

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `main.go` at repo root | ARCH | §2.1 | Package unimportable as library; only one binary possible | `cmd/<name>/main.go` |
| Business logic defined in `package main` | ARCH | §2.2 | Logic untestable and unreusable across binaries | Extract to importable library package |
| Single binary with `--mode=cli\|server` flag | ARCH | §2.3 | Both interfaces entangled; harder to evolve, test, or deploy independently | Separate `cmd/myapp` and `cmd/myapp-server` |

---

## 3. Wrap External Types for Application Context

Generic types from external packages (e.g. `*sql.DB`, `*sql.Tx`) carry no
application semantics. Wrapping them adds application-level methods, enforces
validation, and hides the underlying technology from callers.

---

§3.1  **[SHOULD][CODE]**  When an external type (database connection, transaction,
      client) is used throughout the application, wrap it in an application-specific
      struct and expose only application-level methods on that struct.
      Callers that use the raw external type directly are forced to know the
      external API; adding application-level behaviour (validation, instrumentation)
      then requires changes at every call site.
      ```go
      // ✓  Application wrapper adds domain methods; raw *sql.DB is unexposed
      type DB struct{ *sql.DB }
      type Tx struct{ *sql.Tx }

      func (tx *Tx) CreateUser(u *User) error {
          if u == nil { return errors.New("user required") }
          if u.Name == "" { return errors.New("name required") }
          return tx.Exec(`INSERT INTO users ...`)
      }
      ```

§3.2  **[SHOULD][CODE]**  Place validation logic for a write operation on the
      transaction method that performs the write, not in the caller.
      Validation scattered across callers is duplicated and inconsistently applied;
      centralising it on the write method means any caller that uses the method
      gets the validation for free.

§3.3  **[SHOULD][CODE]**  Compose multiple write operations within a single
      transaction by calling the transaction's application methods repeatedly
      rather than providing batch variants of each method.
      Batch methods (`CreateUsers`) duplicate the logic of the single method and
      become unnecessary when the transaction is the unit of composition; callers
      simply loop over the single method.
      ```go
      // ✗  Separate batch method duplicates CreateUser logic
      func (tx *Tx) CreateUsers(users []*User) error { ... }

      // ✓  Loop over the single method inside one transaction
      tx, _ := db.Begin()
      for _, u := range users {
          tx.CreateUser(u)
      }
      tx.Commit()
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Validation inline in every caller before `db.Exec(...)` | CODE | §3.2 | Validation duplicated; one caller can omit it | Move validation into `Tx.CreateUser()` |
| `CreateUsers(users []*User)` alongside `CreateUser(u *User)` | CODE | §3.3 | Logic duplicated; both methods must be maintained in sync | Loop over `CreateUser` within one transaction |
| Raw `*sql.DB` passed as a parameter throughout the application | CODE | §3.1 | Application-level behaviour (caching, instrumentation, validation) cannot be added without modifying every call site | Wrap in `myapp.DB`; pass the wrapper |
| Raw `*sql.Tx` used directly in service methods | CODE | §3.1 | Callers know about `database/sql`; swapping the data store changes all callers | Wrap in `myapp.Tx` with domain methods |

---

## 4. Resist Premature Subpackage Proliferation

Splitting into subpackages before a clear dependency boundary exists creates
cyclic import pressure, forces unexported symbols to become exported, and makes
the codebase harder to navigate.

---

§4.1  **[MUST][ARCH]**  Do not create a subpackage solely because a source file
      or type count feels large; create a subpackage only when a clear dependency
      boundary exists that justifies the isolation.
      Packages split by file count force types that naturally call each other
      to become exported (to cross package lines), enlarging the public API and
      creating cyclic import candidates.

§4.2  **[SHOULD][CODE]**  Group related types and functions together in one file
      rather than one-type-per-file; target 200–500 SLOC per file, with 1,000
      SLOC as the upper limit before splitting the file.
      One-type-per-file scatters related code; a reader tracing a call must open
      multiple files to build a mental model. Files in the 200–500 SLOC range
      contain enough context to be self-explanatory without being overwhelming.
      *(Rationale is stated as the author's experience, not universal fact —
      applied as a [SHOULD].)*

§4.3  **[SHOULD][CODE]**  Order types within a file from most important to least
      important (top to bottom).
      Readers scanning a file for its primary purpose find it immediately at the
      top; secondary helpers do not crowd out the main type.

§4.4  **[SHOULD][ARCH]**  When a project grows beyond approximately 10,000 SLOC,
      evaluate whether it should be split into separate, independently importable
      packages or repositories rather than adding more subpackages within the
      existing module.
      Internal subpackage splits do not reduce the cognitive load of the whole
      module for downstream consumers; an independent package does, and it forces
      the API boundary to be explicitly designed.

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| New subpackage created because a package has "too many files" | ARCH | §4.1 | Cyclic import pressure; unexported symbols must be exported to cross package lines | Stay in one package until a dependency boundary emerges |
| One file per type (`user.go`, `order.go`, `account.go` each with one type) | CODE | §4.2 | Fragmented; reader must open multiple files to trace a single operation | Group related types in one file |
| Helper types at top of file, primary domain type buried at bottom | CODE | §4.3 | Reader cannot orient to the file's purpose without scrolling | Most important type first |
| Large project with dozens of internal subpackages all in one module | ARCH | §4.4 | Downstream consumers must understand the entire module; circular import risk grows | Evaluate splitting into independent importable packages |

---

## Master Anti-patterns Table

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Global `var db *sql.DB` used by handlers | ARCH | §1.1 | Tests share mutable global; parallel tests corrupt each other | Struct-field injection |
| `http.HandleFunc` with a bare function using globals | CODE | §1.2 | No injection point; globals are the only option | `http.Handle` with struct handler |
| Handler opens its own DB connection inside `ServeHTTP` | CODE | §1.3 | Untestable without real DB; new connection per request | Inject `*sql.DB` at construction |
| `main.go` at repository root | ARCH | §2.1 | Package unimportable as library; single binary only | `cmd/<name>/main.go` |
| Business logic in `package main` | ARCH | §2.2 | Untestable; unreusable across binaries | Extract to importable library package |
| Single binary with mode-switching flag | ARCH | §2.3 | Both modes entangled; hard to evolve or test independently | Separate binaries under `cmd/` |
| Validation inline at every call site before `db.Exec` | CODE | §3.2 | Duplicated; one caller can omit it | Validate inside the transaction method |
| Batch method (`CreateUsers`) alongside single method (`CreateUser`) | CODE | §3.3 | Logic duplication; both must be maintained | Loop single method in one transaction |
| Raw `*sql.DB` / `*sql.Tx` passed throughout application | CODE | §3.1 | Callers coupled to `database/sql`; no place to add domain behaviour | Wrap in `myapp.DB` / `myapp.Tx` |
| Subpackage created for file-count reasons | ARCH | §4.1 | Cyclic import pressure; API surface forced larger | One package until a real boundary appears |
| One type per file | CODE | §4.2 | Related code fragmented across files | Group related types in one file (200–500 SLOC) |
| Helpers at top, primary type at bottom | CODE | §4.3 | Reader cannot orient without scrolling | Primary type first |
| Many internal subpackages in one large module | ARCH | §4.4 | Cognitive load not reduced for consumers | Split into independent packages at ~10K SLOC |
