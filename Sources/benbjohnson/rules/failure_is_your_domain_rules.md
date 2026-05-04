# Ruleset: Failure Is Your Domain

```
Source: "Failure Is Your Domain" — Ben Johnson (2018)
Scope:  Go; application-layer error handling; domain-driven package layout;
        services with multiple consumer roles (application logic, end users, operators)
```

---

## 1. Error as a First-Class Domain Type

Errors are as important to the domain as `Customer` or `Order`. They must be
owned by the domain package, not delegated to infrastructure or third parties.

---

§1.1  **[MUST][ARCH]**  Define the application `Error` type in the root domain
      package, not in a subpackage named `errors`.
      Placing `Error` in a subpackage (`errors.Error`) produces a stuttered name
      and removes the type from the domain language; every package that constructs
      errors must import what is conceptually an infrastructure package instead of
      the domain itself.
      ```
      ✗  package errors
         type Error struct { ... }   // callers write errors.Error{...}

      ✓  package myapp
         type Error struct { ... }   // callers write myapp.Error{...}
      ```

§1.2  **[MUST][ARCH]**  Do not use a third-party error-wrapping library (e.g.
      `pkg/errors`) as the primary error representation for domain errors.
      Third-party types live outside the domain; they cannot carry application-
      specific codes or human-readable messages, and they couple the domain's error
      language to an external dependency. When that dependency adds or removes
      behaviour, the domain's error contract changes without a domain-level change.
      *(Overrides common practice — apply as stated.)*

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `package errors; type Error struct { ... }` | ARCH | §1.1 | Stutter name; errors leave the domain language | Root-package `myapp.Error` |
| `github.com/pkg/errors.Wrap(err, "op")` as primary domain error | ARCH | §1.2 | Domain error contract coupled to external library | Domain `myapp.Error` with `Op`+`Err` wrapping |
| Typed `Op string` / `Kind string` aliases as struct fields | CODE | §1.1 (spirit) | Unnecessary complexity for syntactic sugar; strings suffice | Plain `string` fields for `Op` and `Code` |

---

## 2. Error Type Structure

The `Error` type must carry exactly what each consumer role requires and no more.

---

§2.1  **[MUST][CODE]**  Implement the application `Error` struct with exactly
      these four fields: `Code string` (machine-readable error code), `Message
      string` (human-readable end-user message), `Op string` (logical operation
      name), `Err error` (nested wrapped error).
      These four fields satisfy all three consumer roles. Adding typed wrappers
      for `Op` or `Code` increases complexity without benefit; strings are
      sufficient when errors are constructed directly rather than via a builder.
      ```go
      // ✗  Upspin-style typed fields — unnecessary indirection
      type Op   string
      type Kind string
      type Error struct { Op Op; Kind Kind; Err error }

      // ✓  Plain string fields
      type Error struct {
          Code    string  // machine-readable
          Message string  // human-readable
          Op      string  // logical operation
          Err     error   // wrapped cause
      }
      ```

§2.2  **[MUST][CODE]**  Never set both `Err` and (`Code` or `Message`) on a
      single `Error` instance.
      `Err` marks a *wrapping node* — its meaning is delegated to the nested
      error. `Code`/`Message` mark a *leaf node* — it is the origin of the
      condition. Mixing them on one struct breaks the chain traversal in
      `ErrorCode()` and `ErrorMessage()`, causing both functions to return wrong
      results.
      ```go
      // ✗  Mixing wrapping and leaf roles on one struct
      return &myapp.Error{Op: op, Code: myapp.EINTERNAL, Err: err}

      // ✓  Wrapping node: Op + Err only
      return &myapp.Error{Op: op, Err: err}

      // ✓  Leaf node: Code + Message only (no Err)
      return &myapp.Error{Code: myapp.EINVALID, Message: "Username is required."}
      ```

§2.3  **[SHOULD][CODE]**  Document in every function's comment which `Code`
      values it can return.
      Without this documentation, callers cannot distinguish well-defined from
      undefined errors and cannot write exhaustive error-handling branches; the
      `Code` field's value as a stable API contract is lost.
      ```go
      // FindUserByID returns a user by ID.
      // Returns ENOTFOUND if user does not exist.
      func (s *UserService) FindUserByID(ctx context.Context, id int) (*myapp.User, error)
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Typed `Op` / `Kind` fields with builder function | CODE | §2.1 | Syntactic sugar adds indirection; strings are sufficient | Plain `string` fields |
| `&myapp.Error{Op: op, Code: EINTERNAL, Err: err}` | CODE | §2.2 | `ErrorCode()` / `ErrorMessage()` traverse incorrectly | Separate wrapping node from leaf node |
| Functions with undocumented error codes | CODE | §2.3 | Callers cannot handle specific codes; codes become undefined by default | Comment each returned `Code` value |

---

## 3. Error Codes

Error codes are the machine-readable contract between the domain and its callers.

---

§3.1  **[SHOULD][CODE]**  Start with the minimal set of four generic codes and
      expand only when an existing code cannot describe a new error condition:
      ```go
      const (
          ECONFLICT = "conflict"   // action cannot be performed
          EINTERNAL = "internal"   // internal error
          EINVALID  = "invalid"    // validation failed
          ENOTFOUND = "not_found"  // entity does not exist
      )
      ```
      Fine-grained codes proliferate the domain's error vocabulary and force
      callers to handle more cases before the use case is proven to require them.

§3.2  **[MUST][CODE]**  Provide an `ErrorCode(err error) string` function that:
      (1) returns `""` for `nil`; (2) returns the first non-empty `Code` found by
      recursively following `Err`; (3) returns `EINTERNAL` when no `Code` is found.
      Without this helper, callers must type-assert to `*myapp.Error`, cannot
      traverse wrapped errors, and write verbose branching that breaks when the
      error is wrapped in an additional layer.
      ```go
      func ErrorCode(err error) string {
          if err == nil {
              return ""
          } else if e, ok := err.(*Error); ok && e.Code != "" {
              return e.Code
          } else if ok && e.Err != nil {
              return ErrorCode(e.Err)
          }
          return EINTERNAL
      }

      // ✗  Direct type assertion — does not traverse chain, panics if wrong type
      if err.(*myapp.Error).Code == myapp.ENOTFOUND { ... }

      // ✓  Use helper
      if myapp.ErrorCode(err) == myapp.ENOTFOUND { ... }
      ```

§3.3  **[MUST][CODE]**  `ErrorCode` must return `""` (not `EINTERNAL`) for `nil`
      errors.
      Callers that check `ErrorCode(err) == myapp.EINTERNAL` to detect internal
      failures would fire spuriously on every nil-error path if nil mapped to
      `EINTERNAL`.

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Direct type assertion `err.(*myapp.Error).Code` | CODE | §3.2 | Panics if outer error is not `*myapp.Error`; does not traverse chain | `myapp.ErrorCode(err)` |
| Proliferating codes (`ECONFLICT_USERNAME`, `ECONFLICT_EMAIL`, …) upfront | CODE | §3.1 | Premature vocabulary growth; callers must handle more cases | Start with `ECONFLICT`; specialize only when a concrete use case demands it |
| `ErrorCode(nil) == EINTERNAL` | CODE | §3.3 | Spurious internal-error branches on nil | Return `""` for nil |

---

## 4. Consumer-Role Separation

Each of the three consumer roles — application, end user, operator — requires a
different view of the same error. The domain error type must satisfy all three
without leaking one role's information into another.

---

§4.1  **[MUST][CODE]**  When branching on an error in application logic, use
      `ErrorCode()`; never type-assert to a concrete implementation type to
      extract error details.
      Type-asserting to an implementation type (e.g. `*pq.Error`) couples
      calling code to the implementation package; swapping the implementation
      silently breaks all callers that rely on the concrete type.
      ```go
      // ✗  Coupled to postgres implementation
      if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { ... }

      // ✓  Implementation-agnostic
      if myapp.ErrorCode(err) == myapp.ECONFLICT { ... }
      ```

§4.2  **[MUST][CODE]**  Never surface an undefined external error (a raw database,
      network, or library error) directly to an end user.
      Undefined errors from external libraries can expose schema details, query
      text, or implementation specifics that allow attackers to enumerate system
      internals (e.g. a Postgres error revealing column names).
      ```go
      // ✗  Raw database error returned to HTTP response or user-facing log
      return err

      // ✓  Wrap before returning; user sees generic message via ErrorMessage()
      return &myapp.Error{Code: myapp.EINTERNAL, Err: err}
      ```

§4.3  **[MUST][CODE]**  Provide an `ErrorMessage(err error) string` function
      that: (1) returns `""` for `nil`; (2) returns the first non-empty `Message`
      found by recursively following `Err`; (3) returns a generic contact-support
      message when none is found.
      Presentation-layer code that tries to extract messages without this helper
      must replicate chain traversal or receive empty strings for wrapped errors.
      ```go
      func ErrorMessage(err error) string {
          if err == nil {
              return ""
          } else if e, ok := err.(*Error); ok && e.Message != "" {
              return e.Message
          } else if ok && e.Err != nil {
              return ErrorMessage(e.Err)
          }
          return "An internal error has occurred. Please contact technical support."
      }
      ```

§4.4  **[SHOULD][CODE]**  Set `Message` only on leaf `Error` nodes that describe
      a condition the end user can act on.
      Operator-facing wrapping nodes (`Op`+`Err`) should not carry a `Message`;
      attaching a message at a wrapping layer causes `ErrorMessage()` to return the
      outermost (least specific) message instead of the actionable one from the
      origin.
      ```go
      // ✗  Message on a wrapping node obscures the leaf message
      return &myapp.Error{Op: op, Message: "something went wrong", Err: err}

      // ✓  Message only at the leaf where the actionable condition is detected
      return &myapp.Error{Code: myapp.EINVALID, Message: "Username is required."}
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `err.(*pq.Error)` in service or handler code | CODE | §4.1 | Couples domain/handler to implementation; breaks on DB swap | `myapp.ErrorCode(err)` |
| Returning raw `sql` / `net` / library error to an HTTP response | CODE | §4.2 | May leak schema or query details to attacker | Wrap as `EINTERNAL` before returning |
| Checking `err.Error()` string for branching | CODE | §4.1 | String matching is fragile; breaks on library version changes | `myapp.ErrorCode(err)` |
| `Message` on every wrapping node in the chain | CODE | §4.4 | `ErrorMessage()` returns outermost message, not actionable leaf | `Message` only at origin of user-visible condition |

---

## 5. Implementation Boundary Translation

At the boundary between a domain interface and its implementation, foreign error
types must be translated into domain error codes before crossing back into the
domain.

---

§5.1  **[MUST][ARCH]**  At every domain-interface implementation boundary,
      translate all implementation-specific errors to `myapp.Error` with an
      appropriate `Code` before returning.
      Domain callers have no knowledge of implementation-specific error types
      (e.g. `sql.ErrNoRows`, `*pq.Error`). An untranslated implementation error
      forces callers to import the implementing package to inspect the error,
      coupling the domain to the implementation and breaking when the
      implementation changes.
      ```go
      // ✗  Leaks sql.ErrNoRows into the domain
      if err := row.Scan(&user.ID, &user.Username); err != nil {
          return nil, err
      }

      // ✓  Translated at the implementation boundary
      if err := row.Scan(&user.ID, &user.Username); err == sql.ErrNoRows {
          return nil, &myapp.Error{Code: myapp.ENOTFOUND}
      } else if err != nil {
          return nil, &myapp.Error{Code: myapp.EINTERNAL, Err: err}
      }
      ```

§5.2  **[MUST][CODE]**  When an implementation encounters an error with no
      specific domain mapping, wrap it as `EINTERNAL` with an `Op` set and return
      it; do not return the raw error directly.
      A raw implementation error returned through a domain interface is undefined
      to the caller: `ErrorCode()` classifies it as `EINTERNAL` anyway, but the
      operator loses the logical `Op` context that locates the failure.
      ```go
      // ✗  Raw error, no Op context for operator
      if _, err := s.db.Exec(query); err != nil {
          return err
      }

      // ✓  Wrapped with Op so operator can locate the failing operation
      if _, err := s.db.Exec(query); err != nil {
          return &myapp.Error{Op: op, Err: err}
      }
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Returning `sql.ErrNoRows` from a repository method | ARCH | §5.1 | Domain callers must import `database/sql` to check the error | `&myapp.Error{Code: myapp.ENOTFOUND}` |
| Returning raw `*pq.Error` across the domain boundary | ARCH | §5.1 | Exposes implementation details; breaks on DB swap | Translate to `EINTERNAL` or specific code at boundary |
| `return nil, err` in implementation when `err` is a library type | CODE | §5.2 | `ErrorCode()` returns `EINTERNAL` but operator has no `Op` context | `return nil, &myapp.Error{Op: op, Err: err}` |

---

## 6. Logical Stack Traces via Op Wrapping

The operator role requires a logical stack trace — only the application-level
operations that matter, on a single greppable line.

---

§6.1  **[MUST][CODE]**  In every function that wraps an error, set `Op` to the
      logical operation name before returning; the convention is
      `"TypeName.MethodName"` for methods and `"functionName"` for unexported
      functions.
      Without `Op` at each layer, the trace has gaps; the operator must fall back
      to runtime stack dumps that include library internals to find the failing
      operation.
      ```go
      // ✗  No Op — operator cannot locate the failure
      return &myapp.Error{Err: err}

      // ✓  Op at every wrapping layer
      const op = "UserService.CreateUser"
      return &myapp.Error{Op: op, Err: err}
      ```

§6.2  **[MUST][CODE]**  Implement `Error() string` to emit the `Op` chain first
      (colon-separated), followed by the leaf `Code` and `Message`, on a single
      line.
      Printing the op chain first allows log lines to be sorted and grouped by
      operation path in any log aggregation tool. A multi-line format prevents
      grepping; operator-role log lines must be greppable.
      ```go
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
      // Output: "UserService.CreateUser: attachRole: <internal> syntax error at or near INSERT"
      ```

§6.3  **[SHOULD][CODE]**  Do not embed a full runtime stack trace in application
      `Error` values; use `Op`-based logical traces instead.
      Runtime traces expose library internals and framework frames irrelevant to
      application flow; in high-traffic services they add significant allocation
      pressure and log volume. The logical trace contains only the operations the
      developer considers meaningful.

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `&myapp.Error{Err: err}` — wrapping without `Op` | CODE | §6.1 | Trace has gaps; operator cannot locate the failing function | Always set `Op` when wrapping |
| Multi-line `Error()` output | CODE | §6.2 | Log lines cannot be grepped as a unit | Single-line format with op-chain prefix |
| `runtime.Stack()` captured into error value | CODE | §6.3 | Library frames obscure application flow; high allocation cost | `Op`-chain logical trace |
| Setting `Op` only at the outermost layer | CODE | §6.1 | Intermediate operations invisible; operator cannot isolate the failing statement | `Op` at every function that wraps |

---

## Master Anti-patterns Table

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `package errors; type Error struct { ... }` | ARCH | §1.1 | Stutter; error type leaves domain | `package myapp; type Error struct { ... }` |
| `pkg/errors.Wrap(err, "msg")` as primary domain error | ARCH | §1.2 | Domain coupled to external library | Domain `myapp.Error` |
| Typed `Op string` / `Kind string` wrapper fields | CODE | §2.1 | Unnecessary indirection | Plain `string` fields |
| `&myapp.Error{Op: op, Code: EINTERNAL, Err: err}` | CODE | §2.2 | Chain traversal returns wrong code/message | Separate wrapping node from leaf node |
| Functions with undocumented `Code` return values | CODE | §2.3 | Callers cannot handle specific codes | Document every returned `Code` in function comment |
| Direct type assertion `err.(*myapp.Error).Code` | CODE | §3.2 | Panics if not `*myapp.Error`; doesn't traverse chain | `myapp.ErrorCode(err)` |
| Premature fine-grained error codes | CODE | §3.1 | Proliferates error vocabulary before need is proven | Start with 4 generic codes; specialize on demand |
| `ErrorCode(nil) → EINTERNAL` | CODE | §3.3 | Spurious internal-error branches | Return `""` for nil |
| `err.(*pq.Error)` in service or handler | CODE | §4.1 | Couples domain to implementation | `myapp.ErrorCode(err)` |
| Raw library error returned to end user | CODE | §4.2 | May leak schema/query internals | Wrap as `EINTERNAL` |
| `Message` on every wrapping node | CODE | §4.4 | Returns outermost (least specific) message | `Message` only at origin of user-visible condition |
| `return nil, sql.ErrNoRows` from repository | ARCH | §5.1 | Leaks implementation type into domain | `&myapp.Error{Code: myapp.ENOTFOUND}` |
| `return nil, err` with no `Op` for a library error | CODE | §5.2 | Operator loses location context | `&myapp.Error{Op: op, Err: err}` |
| `&myapp.Error{Err: err}` — wrapping without `Op` | CODE | §6.1 | Logical trace has gaps | Always set `Op` when wrapping |
| Multi-line `Error()` output | CODE | §6.2 | Ungreppable in log aggregation | Single-line with op-chain prefix |
| `runtime.Stack()` in error values | CODE | §6.3 | Library frames obscure flow; high allocation | `Op`-chain logical trace |
| Setting `Op` only at outermost layer | CODE | §6.1 | Intermediate operations invisible | `Op` at every wrapping layer |
