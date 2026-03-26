# Ryer Rules: Writing HTTP Services in Go

Rules distilled from Mat Ryer's "How I write HTTP services in Go after 13 years."

---

## Do Not

These are the patterns to avoid. They are listed first because they are the
default instincts that must be overridden.

- **Do not make handlers methods on a server struct.** If a handler needs a
  dependency, it takes it as an argument to the maker func. No hidden
  dependencies, no surprise coupling when testing a single handler.
- **Do not use the global `flag` package.** Use `flag.NewFlagSet` inside `run`
  so tests can pass different `args` without interfering with each other.
- **Do not use `t.SetEnv` in tests.** Use the `getenv` parameter to inject
  environment values — `t.SetEnv` disables `t.Parallel()`.
- **Do not store durable state in handler closures.** Closures may not survive
  restarts or scale across multiple instances. Use a database.
- **Do not call `os.Exit` anywhere except `main`.** Return errors up the stack.
- **Do not put fallible setup in `addRoutes`.** Resolve errors in `run` before
  calling `addRoutes`, so the function stays flat and error-free.
- **Do not use a named `middleware` type alias.** The return type
  `func(http.Handler) http.Handler` is explicit and readable without the
  extra indirection of a named type.
- **Do not call `json.NewEncoder` or `json.NewDecoder` inline in handlers.**
  All JSON encoding and decoding goes through the central `encode`/`decode`
  helpers.
- **Do not repeat middleware dependency arguments on every `mux.Handle` call.**
  Use a middleware constructor that closes over its dependencies.

---

## File structure

Every service follows this layout. One concern per file.

```
main.go        — main() only; calls run()
run.go         — run() function; wires dependencies, starts server
server.go      — NewServer() constructor
routes.go      — addRoutes(); every mux.Handle call lives here
middleware.go  — middleware constructor functions
encode.go      — encode, decode, decodeValid generic helpers
```

---

## Program entry point

`main` does one thing: call `run` and handle the exit code.

```go
func main() {
    ctx := context.Background()
    if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err)
        os.Exit(1)
    }
}
```

`run` accepts OS primitives as arguments so tests can substitute them without
touching global state or the real environment:

```go
func run(
    ctx    context.Context,
    args   []string,
    getenv func(string) string,
    stdin  io.Reader,
    stdout, stderr io.Writer,
) error {
    ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
    defer cancel()
    // parse flags, build dependencies, call NewServer, start httpServer
}
```

### `run` parameter reference

| Parameter | Type | `main` passes | Test passes |
|---|---|---|---|
| `ctx` | `context.Context` | `context.Background()` | `context.WithCancel(...)` |
| `args` | `[]string` | `os.Args` | custom `[]string` |
| `getenv` | `func(string) string` | `os.Getenv` | custom func |
| `stdin` | `io.Reader` | `os.Stdin` | `strings.NewReader(...)` |
| `stdout` | `io.Writer` | `os.Stdout` | `&bytes.Buffer{}` |
| `stderr` | `io.Writer` | `os.Stderr` | `io.Discard` |

### Rules

- `signal.NotifyContext` goes inside `run`, not `main`, so `cancel` is deferred.
- Use `flag.NewFlagSet(args[0], flag.ContinueOnError)` and pass `args[1:]` to
  it. Never use the global `flag` package.
- Use the `getenv` parameter instead of calling `os.Getenv` directly.

---

## Server constructor

`NewServer` is the single constructor for the top-level `http.Handler`. It
accepts all dependencies as explicit arguments, wires routes, wraps middleware,
and returns `http.Handler`.

```go
func NewServer(
    logger *Logger,
    config *Config,
    commentStore *commentStore,
    anotherStore *anotherStore,
) http.Handler {
    mux := http.NewServeMux()
    addRoutes(mux, logger, config, commentStore, anotherStore)
    var handler http.Handler = mux
    handler = someMiddleware(handler)
    handler = someMiddleware2(handler)
    handler = someMiddleware3(handler)
    return handler
}
```

- Return type is `http.Handler`, not a named struct, unless the situation
  genuinely requires more.
- Pass `nil` for dependencies a particular test does not exercise.
- Format long argument lists vertically — one argument per line.
- Prefer explicit argument lists over config structs: the compiler enforces
  completeness and rejects missing arguments.
- Global middleware (CORS, auth, logging) is applied here, not in `addRoutes`.

---

## Route registration

All routes live in `routes.go`. This is the single place to see the full API
surface of the service. Middleware applied per-route is visible here too.

```go
func addRoutes(
    mux          *http.ServeMux,
    logger       *logging.Logger,
    config       Config,
    tenantsStore *TenantsStore,
    commentsStore *CommentsStore,
) {
    mux.Handle("/api/v1/", handleTenantsGet(logger, tenantsStore))
    mux.Handle("/oauth2/", handleOAuth2Proxy(logger, authProxy))
    mux.Handle("/admin", adminOnly(handleAdminIndex(logger)))
    mux.HandleFunc("/healthz", handleHealthzPlease(logger))
    mux.Handle("/", http.NotFoundHandler())
}
```

- `addRoutes` does not return an error. Anything fallible is resolved in `run`
  before this is called.
- Always register an explicit `http.NotFoundHandler()` for `/`.
- Always include a `/healthz` or `/readyz` endpoint.
- Per-route middleware is applied inline here, making it visible at a glance.

---

## Handler functions (maker funcs)

Handlers are maker funcs: functions that take dependencies and return
`http.Handler`. They are not `http.Handler` implementations themselves, and
they are not methods on a struct.

The return type is always `http.Handler`. The inner closure is wrapped with
`http.HandlerFunc` to satisfy the interface, but the declared return type is
`http.Handler` — not `http.HandlerFunc`. Third-party libraries expect
`http.Handler` first.

```go
func handleSomething(logger *Logger, store *Store) http.Handler {
    // one-time setup — runs once at registration time, not per request
    thing := prepareThing()
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // per-request logic
        logger.Info(r.Context(), "handleSomething")
    })
}
```

- The outer function's scope is the closure environment. Use it for setup that
  runs once.
- Only read shared closure data from within concurrent handlers. Protect any
  writes with a mutex.

### Defer expensive setup with `sync.Once`

If setup is expensive, defer it to the first request rather than program
startup. This improves startup time and avoids doing work for handlers that
are never called.

```go
func handleTemplate(files ...string) http.Handler {
    var (
        init   sync.Once
        tpl    *template.Template
        tplerr error
    )
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        init.Do(func() {
            tpl, tplerr = template.ParseFiles(files...)
        })
        if tplerr != nil {
            http.Error(w, tplerr.Error(), http.StatusInternalServerError)
            return
        }
        // use tpl
    })
}
```

- The error check lives outside `init.Do` so it surfaces on every request, not
  just the first.

### Define request/response types inside the maker func

If types are only used by one handler, declare them inside the maker func.
This keeps the package namespace clean and prevents other handlers from taking
an accidental dependency on these types.

```go
func handleSomething() http.Handler {
    type request struct {
        Name string
    }
    type response struct {
        Greeting string `json:"greeting"`
    }
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
    })
}
```

---

## Encoding and decoding

Centralize JSON encode/decode in `encode.go`. All handlers call these helpers.
Never call `json.NewEncoder` or `json.NewDecoder` inline in handler code.

The `encode` function accepts `r *http.Request` for symmetry with `decode` and
to support future content negotiation without changing call sites.

```go
func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(v); err != nil {
        return fmt.Errorf("encode json: %w", err)
    }
    return nil
}

func decode[T any](r *http.Request) (T, error) {
    var v T
    if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
        return v, fmt.Errorf("decode json: %w", err)
    }
    return v, nil
}
```

Usage — the compiler infers the type for `encode` from the argument; `decode`
requires an explicit type parameter because T appears only in the return:

```go
// encode — type inferred
err := encode(w, r, http.StatusOK, obj)

// decode — type must be specified
decoded, err := decode[CreateSomethingRequest](r)
```

---

## Validation

Use a single-method `Validator` interface. Request types implement it.

```go
type Validator interface {
    Valid(ctx context.Context) (problems map[string]string)
}
```

- `Valid` returns a `map[string]string`: field name → human-readable problem.
- Return `nil` (not an empty map) when valid; `len(nil map) == 0` and does not
  panic.
- Keep `Valid` to field-level checks: empty, format, range. Checks that require
  a database call belong outside this method.

Use `decodeValid` when the request type implements `Validator`. Use `decode`
for types that require no validation. Do not call `v.Valid()` inline in handler
code — that belongs in `decodeValid`.

```go
func decodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
    var v T
    if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
        return v, nil, fmt.Errorf("decode json: %w", err)
    }
    if problems := v.Valid(r.Context()); len(problems) > 0 {
        return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
    }
    return v, nil, nil
}
```

---

## Middleware

Middleware signature: `func(http.Handler) http.Handler`. Apply it in
`routes.go` so the route table shows exactly which middleware each route uses.

```go
func adminOnly(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !currentUser(r).IsAdmin {
            http.NotFound(w, r)
            return
        }
        h(w, r)
    })
}
```

```go
// routes.go — middleware is visible per route
mux.Handle("/admin", adminOnly(handleAdminIndex(logger)))
```

### Middleware that needs dependencies

When middleware needs dependencies, use a constructor that closes over them and
returns the middleware function. Do not pass dependencies on every `mux.Handle`
call.

```go
// middleware.go
func newAuthMiddleware(logger Logger, db *DB) func(http.Handler) http.Handler {
    return func(h http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // use logger, db
            h.ServeHTTP(w, r)
        })
    }
}

// routes.go
auth := newAuthMiddleware(logger, db)
mux.Handle("/route1", auth(handleSomething(deps)))
mux.Handle("/route2", auth(handleSomething2(deps)))
```

Return type is the explicit `func(http.Handler) http.Handler` — no named type
alias is needed.

---

## Graceful shutdown

Run the HTTP server in a goroutine. A second goroutine waits for context
cancellation and calls `Shutdown` with a timeout. `wg.Wait()` blocks `run`
until shutdown completes.

```go
httpServer := &http.Server{
    Addr:    net.JoinHostPort(config.Host, config.Port),
    Handler: srv,
}
go func() {
    if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        fmt.Fprintf(stderr, "error listening and serving: %s\n", err)
    }
}()
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := httpServer.Shutdown(shutdownCtx); err != nil {
        fmt.Fprintf(stderr, "error shutting down: %s\n", err)
    }
}()
wg.Wait()
```

- Pass `ctx` through to all dependencies. Check `ctx.Err()` in loops and
  long-running operations.

---

## Testing

### Primary rule: test by calling `run`

Call `run` from test code to exercise the full program — flag parsing,
dependency wiring, database migrations, middleware, routing — exactly as it
runs in production. This catches more issues than testing layers in isolation.

```go
func TestSomething(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    t.Cleanup(cancel)
    go run(ctx, testArgs, testGetenv, nil, io.Discard, io.Discard)
    waitForReady(ctx, 5*time.Second, "http://localhost:PORT/healthz")
    // hit the API as a real client would
}
```

- Use `t.Cleanup(cancel)` — the context cancels when the test ends, which
  triggers graceful shutdown.
- Use `t.Parallel()` freely. `run` has no global state, so concurrent test
  instances do not interfere.
- Write handler-level unit tests only when there is specific complex logic that
  warrants isolated coverage.
- **Delete unit tests that assert the same thing as an end-to-end test.** One
  set of authoritative tests is easier to maintain than duplicated assertions
  spread across layers.

### Wait for readiness

Poll `/healthz` before hitting the API. The loop also provides a healthz
endpoint to real users — designing for testability exposes what users need.

```go
func waitForReady(ctx context.Context, timeout time.Duration, endpoint string) error {
    client := http.Client{}
    start := time.Now()
    for {
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
        if err != nil {
            return fmt.Errorf("failed to create request: %w", err)
        }
        resp, err := client.Do(req)
        if err == nil && resp.StatusCode == http.StatusOK {
            resp.Body.Close()
            return nil
        }
        if resp != nil {
            resp.Body.Close()
        }
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if time.Since(start) >= timeout {
                return fmt.Errorf("timeout waiting for endpoint")
            }
            time.Sleep(250 * time.Millisecond)
        }
    }
}
```

### Control flags and environment in tests

```go
// flags via args
args := []string{"myapp", "--fmt", "markdown"}
go run(ctx, args, getenv, nil, io.Discard, io.Discard)

// environment via getenv — use this instead of t.SetEnv
getenv := func(key string) string {
    switch key {
    case "MYAPP_FORMAT":
        return "markdown"
    case "MYAPP_TIMEOUT":
        return "5s"
    default:
        return ""
    }
}
```

### Use inline types in tests for storytelling

When request/response types are scoped inside a maker func, declare minimal
inline structs in test code that express only what the test cares about:

```go
person := struct {
    Name string `json:"name"`
}{Name: "Mat Ryer"}
```

This documents intent: the test is saying only `Name` matters for this
endpoint.

---

## Summary checklist

- [ ] `main` only calls `run`; passes `os.Args`, `os.Getenv`, `os.Stdin`, `os.Stdout`, `os.Stderr`
- [ ] `run` accepts `(ctx, args, getenv, stdin, stdout, stderr)` — no OS globals
- [ ] `signal.NotifyContext` and `defer cancel()` inside `run`
- [ ] `flag.NewFlagSet` inside `run`; global `flag` package never used
- [ ] `NewServer` takes all dependencies as arguments, returns `http.Handler`
- [ ] Global middleware applied in `NewServer`, per-route middleware in `routes.go`
- [ ] All routes registered in `routes.go`; explicit `http.NotFoundHandler()` for `/`
- [ ] `/healthz` endpoint present
- [ ] Handlers are maker funcs: `func handleX(deps...) http.Handler` — not struct methods
- [ ] Maker func return type is `http.Handler`, not `http.HandlerFunc`
- [ ] Request/response types declared inside maker func if handler-specific
- [ ] Expensive setup deferred with `sync.Once`
- [ ] All JSON encode/decode through `encode`/`decode`/`decodeValid` in `encode.go`
- [ ] `decodeValid` used for request types that implement `Validator`
- [ ] `Validator` interface: `Valid(ctx context.Context) map[string]string`
- [ ] Middleware signature: `func(http.Handler) http.Handler`; no named type alias
- [ ] Middleware with dependencies uses a constructor; deps not repeated per route
- [ ] Graceful shutdown: `<-ctx.Done()` → `httpServer.Shutdown` with timeout → `wg.Wait()`
- [ ] Tests call `run` end-to-end; `t.Cleanup(cancel)`; `t.Parallel()`
- [ ] `getenv` parameter used instead of `t.SetEnv`
- [ ] Unit tests that duplicate end-to-end assertions are deleted
