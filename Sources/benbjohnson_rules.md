# Ben Johnson's Go Application Design Spec

Rules distilled from Ben Johnson's gobeyond.dev series on Go application design.

---

## Do Not

- Do not put domain types in subpackages — they belong in the root package.
- Do not group packages by type (`models/`, `controllers/`, `handlers/`) — this is Rails-style layout and it causes circular dependency problems in Go. Group by dependency instead.
- Do not allow the root package to import any other package in the application.
- Do not use global variables for application state (db connections, config, etc.).
- Do not use third-party testing frameworks — use the stdlib `testing` package only.
- Do not use ORMs — use `database/sql` directly.
- Do not expose transactions to callers of a service — transactions are an implementation detail.
- Do not return `(nil, nil)` from a function that looks up a single entity by ID.
- Do not allow arbitrary sort columns from callers — map a fixed set of named values to SQL.
- Do not put the `main` package in the project root — put binaries under `cmd/`.
- Do not put business logic in `main` — it only wires dependencies together.
- Do not enforce authorization in HTTP handlers or middleware — enforce it in service implementations, embedded in SQL where possible.
- Do not interpolate caller-supplied strings into SQL queries — use parameterized queries and fixed mapped values.

---

## Project layout

```
myapp/                    — root package: domain types, interfaces, Error type
    user.go               — User struct, UserService interface
    dial.go               — Dial struct, DialService interface
    error.go              — Error type, error codes, ErrorCode(), ErrorMessage()
    context.go            — NewContextWithUser(), UserIDFromContext()
    cmd/
        myapp/
            main.go       — wires dependencies, starts server
        myappctl/
            main.go       — optional CLI binary
    sqlite/               — SQLite implementation of domain interfaces
        sqlite.go         — DB type, Open(), Close()
        user.go           — sqlite.UserService
        dial.go           — sqlite.DialService
    http/                 — HTTP adapter
        server.go         — http.Server struct, NewServer(), route registration
        user.go           — HTTP handlers for users
        dial.go           — HTTP handlers for dials
    mock/                 — shared mock implementations for testing
        user_service.go   — mock.UserService
        dial_service.go   — mock.DialService
```

**Rules:**
- The root package name matches the application name (e.g., `package myapp`).
- Each subpackage is named after the dependency it wraps: `sqlite`, `http`, `postgres`, `mock`.
- It is acceptable — even preferred — to name a package the same as its wrapped stdlib package
  (e.g., `http`) because the two are never used in the same file.
- Multiple binaries: one subdirectory per binary under `cmd/`.

### File organization within packages

- Group related types and functions together in one file. One major concept per file
  (e.g., `user.go` contains `User`, `UserService`, `UserFilter`, `UserUpdate`).
- Put the most important type at the top of the file; lesser types below.
- Target 200–500 SLOC per file. 1000 SLOC is the hard limit for a single file.
- If a package exceeds ~10,000 SLOC total, evaluate whether it should be split into
  separate projects.

---

## Root package: domain types and interfaces

The root package defines the application's domain language. It contains only:
- Plain data structs with no external dependencies
- Service interfaces
- The `Error` type and error helpers
- Context helpers for passing the authenticated user

```go
// myapp/user.go
package myapp

import "context"

// User represents an application user.
type User struct {
    ID    int
    Name  string
    Email string
}

// UserService defines operations on users.
// Implementations live in subpackages (sqlite, postgres, mock).
type UserService interface {
    // FindUserByID retrieves a user by ID.
    // Returns ENOTFOUND if the user does not exist.
    FindUserByID(ctx context.Context, id int) (*User, error)

    // FindUsers retrieves a list of users matching filter.
    // Also returns the total count of matching users regardless of Limit/Offset.
    FindUsers(ctx context.Context, filter UserFilter) ([]*User, int, error)

    // CreateUser creates a new user.
    // On success, dial.ID and timestamps are populated on the input struct.
    CreateUser(ctx context.Context, user *User) error

    // UpdateUser updates an existing user by ID.
    // Returns the updated user even if an error occurs.
    // Returns ENOTFOUND if the user does not exist.
    // Returns EUNAUTHORIZED if the caller does not own the user.
    UpdateUser(ctx context.Context, id int, upd UserUpdate) (*User, error)

    // DeleteUser permanently removes a user by ID.
    // Returns ENOTFOUND if the user does not exist.
    // Returns EUNAUTHORIZED if the caller does not own the user.
    DeleteUser(ctx context.Context, id int) error
}

// UserFilter filters results from FindUsers.
type UserFilter struct {
    ID    *int    // optional
    Email *string // optional

    Offset int
    Limit  int
}

// UserUpdate holds the fields that can be updated on a user.
// Pointer fields are optional — nil means do not change.
type UserUpdate struct {
    Name  *string
    Email *string
}
```

**Rules:**
- Domain structs reference only primitive types and other domain types.
- No `database/sql`, `net/http`, or any other external import in the root package.
- Service interfaces live alongside the types they operate on, in the same file.
- Every interface method has a godoc comment that documents which error codes it can return.
- Filter structs use pointer fields so each field is independently optional.
- Update structs use pointer fields so partial updates are expressible without a separate endpoint.

---

## Error type

Define one `Error` type in the root package. It serves three consumers:
the application (via `Code`), the end user (via `Message`), and the operator (via `Op`/`Err` chain).

```go
// myapp/error.go
package myapp

import (
    "bytes"
    "fmt"
)

// Application error codes.
const (
    ECONFLICT      = "conflict"      // action cannot be performed
    EINTERNAL      = "internal"      // internal error
    EINVALID       = "invalid"       // validation failed
    ENOTFOUND      = "not_found"     // entity does not exist
    EUNAUTHORIZED  = "unauthorized"  // caller lacks permission
)

// Error defines a standard application error.
type Error struct {
    Code    string // machine-readable error code
    Message string // human-readable message for end users
    Op      string // logical operation (e.g. "sqlite.UserService.FindUserByID")
    Err     error  // nested error
}

// Error returns a string suitable for operators, including the logical stack trace.
func (e *Error) Error() string {
    var buf bytes.Buffer
    if e.Op != "" {
        fmt.Fprintf(&buf, "%s: ", e.Op)
    }
    if e.Err != nil {
        buf.WriteString(e.Err.Error())
    } else {
        if e.Code != "" {
            fmt.Fprintf(&buf, "<%s> ", e.Code)
        }
        buf.WriteString(e.Message)
    }
    return buf.String()
}

// ErrorCode returns the code of the root error, or EINTERNAL if none is set.
func ErrorCode(err error) string {
    if err == nil {
        return ""
    }
    if e, ok := err.(*Error); ok && e.Code != "" {
        return e.Code
    } else if ok && e.Err != nil {
        return ErrorCode(e.Err)
    }
    return EINTERNAL
}

// ErrorMessage returns the human-readable message of the error, or a generic message.
func ErrorMessage(err error) string {
    if err == nil {
        return ""
    }
    if e, ok := err.(*Error); ok && e.Message != "" {
        return e.Message
    } else if ok && e.Err != nil {
        return ErrorMessage(e.Err)
    }
    return "An internal error has occurred. Please contact technical support."
}
```

**Rules:**
- `Error` is defined in the root package, not in an `errors` subpackage — avoids the stutter `errors.Error`.
- `Code` and `Message` are never set on the same `Error` as `Err`. A wrapping error carries `Op` + `Err`; a leaf error carries `Code` + `Message`.
- Start with five codes (ECONFLICT, EINTERNAL, EINVALID, ENOTFOUND, EUNAUTHORIZED). Add more only as needed.
- Translate all external errors (e.g., `sql.ErrNoRows`) to domain error codes at the implementation boundary, before they escape to callers.
- Use `ErrorCode(err)` and `ErrorMessage(err)` in calling code — never type-assert `*Error` directly.

### Using Op for logical stack traces

Every significant function wraps errors with its `Op` name using the format `"package.Type.Method"`:

```go
func (s *UserService) CreateUser(ctx context.Context, user *myapp.User) error {
    const op = "sqlite.UserService.CreateUser"

    if err := s.insertUser(ctx, user); err != nil {
        return &myapp.Error{Op: op, Err: err}
    }
    return nil
}

func (s *UserService) insertUser(ctx context.Context, user *myapp.User) error {
    const op = "sqlite.insertUser"
    if _, err := s.db.ExecContext(ctx, `INSERT INTO users ...`); err != nil {
        return &myapp.Error{Op: op, Err: err}
    }
    return nil
}
```

The resulting `Error()` string is a single-line logical stack trace:
```
sqlite.UserService.CreateUser: sqlite.insertUser: near "INSERT": syntax error
```

---

## Authentication via context

The authenticated user is stored in `context.Context` at the request boundary and
extracted inside service implementations to enforce authorization.

```go
// myapp/context.go
package myapp

import "context"

type contextKey int

const userContextKey contextKey = iota

// NewContextWithUser returns a new context with the given user attached.
func NewContextWithUser(ctx context.Context, user *User) context.Context {
    return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext returns the user stored in ctx, or nil if none.
func UserFromContext(ctx context.Context) *User {
    u, _ := ctx.Value(userContextKey).(*User)
    return u
}

// UserIDFromContext returns the ID of the user stored in ctx, or 0 if none.
func UserIDFromContext(ctx context.Context) int {
    if u := UserFromContext(ctx); u != nil {
        return u.ID
    }
    return 0
}
```

**Rules:**
- The HTTP layer (or equivalent transport layer) sets the user on the context after authentication.
- Service implementations extract the user from context and apply authorization rules.
- Authorization is enforced at the lowest level possible — embedded in SQL `WHERE` clauses
  so the database engine enforces it, not application-level filtering after the fact.

Example: enforce ownership inside a SQL helper, not in the HTTP handler:

```go
// sqlite/dial.go — authorization embedded in the query
func findDials(ctx context.Context, tx *Tx, filter myapp.DialFilter) ([]*myapp.Dial, int, error) {
    userID := myapp.UserIDFromContext(ctx)

    where := []string{"1 = 1"}
    args := []interface{}{}

    // Restrict to dials the current user is a member of.
    where = append(where, `id IN (SELECT dial_id FROM dial_memberships WHERE user_id = ?)`)
    args = append(args, userID)

    // ... additional filter predicates
}
```

---

## Subpackages: dependency adapters

Each subpackage wraps one external dependency and implements one or more domain interfaces.
The subpackage depends on the root package; the root package never depends on subpackages.

```go
// sqlite/user.go
package sqlite

import (
    "context"
    "myapp"
)

// UserService implements myapp.UserService using SQLite.
type UserService struct {
    db *DB
}

func (s *UserService) FindUserByID(ctx context.Context, id int) (*myapp.User, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    user, err := findUserByID(ctx, tx, id)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

**Rules:**
- The subpackage struct (`sqlite.UserService`) holds only the dependencies it needs.
- All subpackage service methods follow the same shape: open transaction → call helpers → commit.
- Subpackages never import each other. If two implementations need shared logic, it goes in the root package.

### Layering implementations

Because all implementations satisfy the same root-package interface, they can be
stacked. A caching layer in the root package wraps any `UserService`:

```go
// myapp/user_cache.go
package myapp

// UserCache wraps a UserService to provide an in-memory read-through cache.
type UserCache struct {
    cache   map[int]*User
    service UserService
}

func NewUserCache(service UserService) *UserCache {
    return &UserCache{cache: make(map[int]*User), service: service}
}

func (c *UserCache) FindUserByID(ctx context.Context, id int) (*User, error) {
    if u := c.cache[id]; u != nil {
        return u, nil
    }
    u, err := c.service.FindUserByID(ctx, id)
    if err != nil {
        return nil, err
    }
    c.cache[id] = u
    return u, nil
}

// ... implement remaining UserService methods by delegating to c.service
```

`main` composes the layers:
```go
userService := myapp.NewUserCache(&sqlite.UserService{DB: db})
```

---

## HTTP layer

The `http` subpackage is an adapter between the domain and the HTTP protocol.
It holds service interfaces as fields — never as globals or closures over package-level vars.

```go
// http/server.go
package http

import (
    "net/http"
    "myapp"
)

// Server is the HTTP adapter. It holds all service dependencies.
type Server struct {
    UserService myapp.UserService
    DialService myapp.DialService

    router *http.ServeMux
}

// NewServer creates a new HTTP server and registers all routes.
func NewServer(userService myapp.UserService, dialService myapp.DialService) *Server {
    s := &Server{
        UserService: userService,
        DialService: dialService,
        router:      http.NewServeMux(),
    }
    s.router.HandleFunc("/users", s.handleUsersGet)
    s.router.HandleFunc("/users/", s.handleUserGet)
    // ... additional routes
    return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.router.ServeHTTP(w, r)
}
```

**Rules:**
- Handlers are methods on `*Server`, which gives them access to service dependencies without globals.
- The HTTP layer authenticates the request and calls `myapp.NewContextWithUser` before invoking services.
- The HTTP layer translates domain errors to HTTP status codes using `myapp.ErrorCode(err)`.
- The HTTP layer never contains business logic — only protocol translation.

HTTP error translation example:
```go
func errorStatusCode(err error) int {
    switch myapp.ErrorCode(err) {
    case myapp.ENOTFOUND:
        return http.StatusNotFound
    case myapp.EINVALID:
        return http.StatusBadRequest
    case myapp.EUNAUTHORIZED:
        return http.StatusUnauthorized
    case myapp.ECONFLICT:
        return http.StatusConflict
    default:
        return http.StatusInternalServerError
    }
}
```

---

## SQL layer

### DB and Tx types

Wrap `*sql.DB` and `*sql.Tx` in your own types to add application-specific helpers:

```go
// sqlite/sqlite.go
package sqlite

import "database/sql"

type DB struct {
    db  *sql.DB
    DSN string
}

func (db *DB) Open() error  { ... }
func (db *DB) Close() error { ... }

type Tx struct {
    *sql.Tx
}
```

### Service methods: transaction boundary only

Service methods are thin. They own the transaction; helper functions own the SQL.

```go
func (s *DialService) CreateDial(ctx context.Context, dial *myapp.Dial) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if err := createDial(ctx, tx, dial); err != nil {
        return err
    }
    return tx.Commit()
}
```

### Helper functions: unexported, accept `*Tx`, reusable

Helper functions are package-level (not attached to a service type) so multiple
service methods can call them within the same transaction.

```go
func createDial(ctx context.Context, tx *Tx, dial *myapp.Dial) error {
    const op = "sqlite.createDial"
    result, err := tx.ExecContext(ctx, `
        INSERT INTO dials (user_id, name, invite_code, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)`,
        dial.UserID, dial.Name, dial.InviteCode, dial.CreatedAt, dial.UpdatedAt,
    )
    if err != nil {
        return &myapp.Error{Op: op, Err: err}
    }
    dial.ID, err = result.LastInsertId()
    return err
}
```

### Loading associated data

Use a dedicated `attachXAssociations` helper to load related objects after the
primary query. Keep it separate from the main query so it is reusable across
`FindByID` and `FindMany`.

```go
func (s *DialService) FindDials(ctx context.Context, filter myapp.DialFilter) ([]*myapp.Dial, int, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, 0, err
    }
    defer tx.Rollback()

    dials, n, err := findDials(ctx, tx, filter)
    if err != nil {
        return dials, n, err
    }
    for _, dial := range dials {
        if err := attachDialAssociations(ctx, tx, dial); err != nil {
            return dials, n, err
        }
    }
    return dials, n, nil
}

// attachDialAssociations loads the owner user and memberships onto dial.
func attachDialAssociations(ctx context.Context, tx *Tx, dial *myapp.Dial) error {
    var err error
    if dial.User, err = findUserByID(ctx, tx, dial.UserID); err != nil {
        return fmt.Errorf("attach dial user: %w", err)
    }
    return nil
}
```

**Return parent associations always. Include child collections only when they are
small and almost always needed by the caller.**

### Building WHERE clauses dynamically

```go
where := []string{"1 = 1"}
args := []interface{}{}

if v := filter.ID; v != nil {
    where = append(where, "id = ?")
    args = append(args, *v)
}
if v := filter.Email; v != nil {
    where = append(where, "email = ?")
    args = append(args, *v)
}

query := `SELECT id, name, email, COUNT(*) OVER() FROM users WHERE ` +
    strings.Join(where, " AND ") +
    ` ORDER BY ` + orderBy +
    ` LIMIT ? OFFSET ?`
args = append(args, filter.Limit, filter.Offset)
```

- Start with `"1 = 1"` so the slice is never empty and `strings.Join` always produces valid SQL.
- Each filter field is a pointer; nil means not set.

### Row iteration: always close and check error

```go
rows, err := tx.QueryContext(ctx, query, args...)
if err != nil {
    return nil, 0, err
}
defer rows.Close()

dials := make([]*myapp.Dial, 0)  // use make, not var — encodes as [] not null in JSON
var n int
for rows.Next() {
    var dial myapp.Dial
    if err := rows.Scan(&dial.ID, &dial.Name, &n); err != nil {
        return nil, 0, err
    }
    dials = append(dials, &dial)
}
return dials, n, rows.Err()  // always return rows.Err()
```

**Three rules:**
1. `defer rows.Close()` immediately after a successful `QueryContext`.
2. Initialize result slices with `make([]*T, 0)` — nil slices encode as JSON `null`; empty slices encode as `[]`.
3. Return `rows.Err()` after the loop — it captures errors that occurred mid-iteration.

### Pagination with total count in one query

```SQL
SELECT id, name, COUNT(*) OVER()
FROM dials
WHERE user_id = ?
ORDER BY id ASC
LIMIT ? OFFSET ?
```

`COUNT(*) OVER()` is a SQL window function that returns the total matching row count
on every row, ignoring `LIMIT`/`OFFSET`. Scan it on every iteration — the value is
the same each time. This avoids a second `COUNT` query.

### Sort order: fixed named values only

```go
var orderBy string
switch filter.SortBy {
case "name_asc":
    orderBy = "name ASC"
case "updated_at_desc":
    orderBy = "updated_at DESC"
default:
    orderBy = "id ASC"
}
```

Never interpolate caller-supplied strings into SQL. Map a closed set of named values
to SQL fragments. Always provide a safe default.

---

## CRUD conventions

### FindByID

```go
// FindUserByID retrieves a user by ID.
// Returns ENOTFOUND if the user does not exist.
FindUserByID(ctx context.Context, id int) (*User, error)
```

- Never return `(nil, nil)`. If the entity does not exist, return `ENOTFOUND`.
- Return parent associations attached to the object.
- Include bounded child collections only if they are almost always needed and small.

Implementation — wrap the list helper and change the empty-result semantics:

```go
func findUserByID(ctx context.Context, tx *Tx, id int) (*myapp.User, error) {
    users, _, err := findUsers(ctx, tx, myapp.UserFilter{ID: &id})
    if err != nil {
        return nil, err
    } else if len(users) == 0 {
        return nil, &myapp.Error{Code: myapp.ENOTFOUND, Message: "User not found."}
    }
    return users[0], nil
}
```

### FindMany

```go
// FindUsers retrieves users matching filter.
// Also returns the total count ignoring Limit/Offset.
FindUsers(ctx context.Context, filter UserFilter) ([]*User, int, error)
```

- `(nil, 0, nil)` is a valid return — an empty result set is not an error.
- Second return is the total count for pagination.

### Create

```go
// CreateUser creates a new user.
// On success, user.ID and timestamps are populated on the input struct.
CreateUser(ctx context.Context, user *User) error
```

- Accept a pointer to the domain struct.
- Mutate the input pointer in place with generated values (ID, CreatedAt, UpdatedAt).
- Nest related object creation in the same transaction by accepting populated child slices.

### Update

```go
// UpdateUser updates a user by ID. Returns the updated user even on error.
// Returns ENOTFOUND if the user does not exist.
// Returns EUNAUTHORIZED if the caller does not own the user.
UpdateUser(ctx context.Context, id int, upd UserUpdate) (*User, error)
```

- Accept an `Update` struct with pointer fields; nil means unchanged.
- Return the updated object even when an error occurs — web UIs need it to replay the form.
- The `id` is separate from `UserUpdate` so the same update struct can target multiple IDs.

### Delete

```go
// DeleteUser permanently removes a user by ID.
// Returns ENOTFOUND if the user does not exist.
// Returns EUNAUTHORIZED if the caller does not own the user.
DeleteUser(ctx context.Context, id int) error
```

- Delete by primary key.
- Enforce authorization inside the implementation using `UserIDFromContext`.

---

## mock package

The mock package provides simple hand-written mocks for testing. Each mock holds
function fields that tests assign. No third-party mock libraries.

```go
// mock/user_service.go
package mock

import (
    "context"
    "myapp"
)

// UserService is a mock implementation of myapp.UserService.
type UserService struct {
    FindUserByIDFn      func(ctx context.Context, id int) (*myapp.User, error)
    FindUserByIDInvoked bool

    CreateUserFn      func(ctx context.Context, user *myapp.User) error
    CreateUserInvoked bool

    // one Fn + Invoked pair per interface method
}

func (s *UserService) FindUserByID(ctx context.Context, id int) (*myapp.User, error) {
    s.FindUserByIDInvoked = true
    return s.FindUserByIDFn(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, user *myapp.User) error {
    s.CreateUserInvoked = true
    return s.CreateUserFn(ctx, user)
}
```

Usage in tests:

```go
var svc mock.UserService
svc.FindUserByIDFn = func(ctx context.Context, id int) (*myapp.User, error) {
    if id != 100 {
        t.Fatalf("unexpected id: %d", id)
    }
    return &myapp.User{ID: 100, Name: "Susy"}, nil
}
```

**Rules:**
- One mock per domain interface, in the `mock` package.
- Do not use third-party mock generation tools. Write them by hand.
- The `Invoked` booleans let tests assert that a method was or was not called.

---

## main package: wiring dependencies

`main` wires concrete implementations to interfaces and starts the server.
It contains no business logic.

```go
// cmd/myapp/main.go
package main

import (
    "log"
    "os"

    "github.com/myorg/myapp"
    "github.com/myorg/myapp/http"
    "github.com/myorg/myapp/sqlite"
)

func main() {
    // Open database.
    db := &sqlite.DB{DSN: os.Getenv("DATABASE_URL")}
    if err := db.Open(); err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create service implementations, optionally wrapped with caching layers.
    userService := myapp.NewUserCache(&sqlite.UserService{DB: db})
    dialService := &sqlite.DialService{DB: db}

    // Wire into HTTP server and start.
    srv := http.NewServer(userService, dialService)
    if err := srv.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

**Rules:**
- `main` is the only place where concrete implementation packages (`sqlite`, `http`) are imported together.
- Dependency injection is manual — no framework.
- `main` is also an adapter: it connects the OS environment (env vars, args) to the domain.

---

## Testing

### Use the `_test` package

Test files use the external test package to test the exported API, not internals:

```go
// sqlite/user_test.go
package sqlite_test
```

This prevents tests from accessing unexported fields, which keeps them from being brittle
and ensures the exported API is actually usable.

### Test-specific helper types

Wrap the real type to handle setup and teardown:

```go
// sqlite/testing_test.go
package sqlite_test

import "testing"

type TestDB struct {
    *sqlite.DB
}

func MustOpenDB(t *testing.T) *TestDB {
    t.Helper()
    db := &sqlite.DB{DSN: ":memory:"}
    if err := db.Open(); err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { db.Close() })
    return &TestDB{db}
}
```

### Test assertion helpers

Define three simple helpers in the test file. Do not import an assertion library.

```go
// assert fails the test if condition is false.
func assert(t *testing.T, condition bool, msg string) {
    t.Helper()
    if !condition {
        t.Fatal(msg)
    }
}

// ok fails the test if err is non-nil.
func ok(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %s", err)
    }
}

// equals fails the test if exp and act are not equal.
func equals(t *testing.T, exp, act interface{}) {
    t.Helper()
    if exp != act {
        t.Fatalf("expected %v, got %v", exp, act)
    }
}
```

Use these instead of verbose `if err != nil { t.Fatalf(...) }` blocks throughout tests.

### Inline interfaces in tests (caller defines the interface)

The caller — not the callee — defines the interface. Declare only the methods the
test actually needs:

```go
type fakeMailer struct {
    SendFunc func(to, subject, body string) error
}

func (f *fakeMailer) Send(to, subject, body string) error {
    return f.SendFunc(to, subject, body)
}
```

---

## Checklist

- [ ] Root package contains domain types, service interfaces, Error type — no external imports
- [ ] Error type in root package with Code, Message, Op, Err fields
- [ ] Five base error codes: ECONFLICT, EINTERNAL, EINVALID, ENOTFOUND, EUNAUTHORIZED
- [ ] ErrorCode() and ErrorMessage() helpers in root package
- [ ] Every interface method has godoc documenting which error codes it returns
- [ ] External errors translated to domain error codes at the implementation boundary
- [ ] NewContextWithUser() and UserIDFromContext() in root package context.go
- [ ] Authorization enforced inside service implementations via UserIDFromContext, embedded in SQL
- [ ] Subpackages named after wrapped dependency (sqlite, http, mock)
- [ ] HTTP layer holds service interfaces as Server struct fields, not globals
- [ ] HTTP layer translates domain error codes to HTTP status codes
- [ ] Service methods are thin: begin tx → helpers → commit; defer tx.Rollback()
- [ ] SQL helpers are unexported package-level functions accepting *Tx
- [ ] defer rows.Close() after every successful QueryContext
- [ ] rows.Err() returned after every row iteration loop
- [ ] Result slices initialized with make([]*T, 0), not var
- [ ] WHERE clauses built with []string{"1 = 1"} + strings.Join
- [ ] Sort order mapped from fixed named values, never interpolated from caller
- [ ] COUNT(*) OVER() used for pagination total count in one query
- [ ] attachXAssociations helper used to load related data after primary query
- [ ] FindByID never returns (nil, nil); returns ENOTFOUND if missing
- [ ] FindMany returns ([]T, int, error); int is total count for pagination
- [ ] Filter structs use pointer fields; Update structs use pointer fields
- [ ] Create mutates the input pointer (sets ID, timestamps)
- [ ] UpdateX returns updated object even on error
- [ ] Caching/layering via wrapper types that implement domain interfaces
- [ ] mock package has hand-written mocks with Fn + Invoked fields per method
- [ ] main only wires dependencies; no business logic
- [ ] Binaries live under cmd/
- [ ] Test files use _test package suffix
- [ ] Three test helpers defined: assert, ok, equals
- [ ] No global variables; no testing frameworks; no ORM
