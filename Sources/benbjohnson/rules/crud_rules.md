# CRUD Rules

```
Source: "Common CRUD Design in Go" — Ben Johnson (gobeyond.dev, Jan 2021)
Scope:  Go; application-layer CRUD services backed by relational databases;
        domain-driven package layout with an interface in the root package
        and separate implementation packages (e.g., sqlite)
```

---

## 1. Service Interface Design

§1.1  `[MUST][ARCH]`  Define each CRUD service as an interface in the root domain package, not inside any implementation package.
      When the interface lives in an implementation package, callers import that package and are coupled to the technology. The root package owns the contract; implementations satisfy it from the outside.
      ✗  `sqlite.DialService` interface defined and owned by the sqlite package
      ✓  `wtf.DialService` defined in the root package; `sqlite.DialService` struct implements it

§1.2  `[MUST][ARCH]`  Treat each service as a black box: never expose transactions, connections, or other implementation details across the service boundary.
      Leaking transactions allows callers to compose cross-service calls that look atomic but are not; it also makes the underlying technology visible to every caller, preventing substitution.
      ✗  `BeginTx() *sql.Tx` method on the service interface
      ✓  Each service method manages its own transaction internally

§1.3  `[MUST][ARCH]`  Propagate the authenticated user through `context.Context`; enforce authorization inside the service implementation, not at a higher layer.
      A check placed only in an HTTP handler or middleware is bypassed by any other caller of the same service. Embedding the check in the service (or its SQL query) is the lowest and most reliable enforcement point.
      ✗  Middleware sets a header; service reads and trusts the raw header
      ✓  `wtf.NewContextWithUser(ctx, user)` stores the user; service reads it via `wtf.UserIDFromContext(ctx)` and applies the restriction in every query

§1.4  `[SHOULD][ARCH]`  Push authorization restrictions into the data query rather than filtering in application code after retrieval.
      Post-retrieval filtering fetches rows the caller is not allowed to see, wastes I/O, and can leak existence information through timing differences.
      ✗  Fetch all dials, then filter in a loop by membership
      ✓  `WHERE id IN (SELECT dial_id FROM dial_memberships WHERE user_id = ?)` applied in the SQL query

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Interface defined in implementation package | ARCH | §1.1 | Callers couple to the technology; substitution requires changing all import paths | Define interface in the root domain package |
| Exposing `BeginTx` or raw connection on service interface | ARCH | §1.2 | Callers compose cross-service "transactions" that have no real atomicity guarantees | Each service method is its own transactional unit |
| Authorization check only in HTTP middleware | ARCH | §1.3 | Direct calls to the service bypass the check | Enforce inside the service implementation |
| Fetching then filtering in application code | ARCH | §1.4 | Unnecessary data transfer; existence leakage via timing | Restrict in the SQL WHERE clause |

---

## 2. Single-Object Lookup

§2.1  `[MUST][CODE]`  When a caller requests a specific object by primary key, return either the object or an error — never `(nil, nil)`.
      A `nil` object with a `nil` error compiles without forcing the caller to check for the nil object. In practice callers do `if err != nil` and then dereference, causing a nil-pointer panic for any not-found ID.
      ✗
      ```go
      // dial is nil; next line panics
      dial, err := FindDialByID(ctx, 100)
      if err != nil { return err }
      fmt.Printf("WTF Level: %d", dial.Value)
      ```
      ✓  Return a domain-typed not-found error (`wtf.ENOTFOUND`) when the row does not exist

§2.2  `[SHOULD][CODE]`  Always populate parent-relationship objects on the returned entity; include child collections only when their size is bounded and they are nearly always needed alongside the parent.
      Parent objects (e.g., the `User` who owns a `Dial`) are almost always required for display or authorization — fetching them inline avoids a second round-trip. Unbounded child collections can blow up payload size and query cost and should be fetched separately.
      ✗  Return `Dial` with no `User` field — every caller must make a second fetch
      ✓  Return `Dial` with `User` populated; include `Memberships` (small, bounded) but not open-ended lists

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Returning `(nil, nil)` when a keyed lookup finds no row | CODE | §2.1 | Callers that check only `err != nil` panic on nil dereference | Return a typed not-found error |
| Returning a bare entity with no parent fields populated | CODE | §2.2 | N+1 fetches in every caller | Populate parent relationships in the same query |
| Including unbounded child collections inline | CODE | §2.2 | Payload explosion; slow queries for deeply nested graphs | Fetch large or unbounded collections separately |

---

## 3. Collection Search

§3.1  `[MUST][CODE]`  When searching a collection, returning an empty or nil slice with a nil error is correct — not finding matching records is not an error.
      The caller does not know a priori whether matches exist; that is the purpose of the search. A nil slice is safe: `len(nil) == 0` and `for range nil` iterates zero times.
      ✗  `return nil, ErrNotFound` when no rows match the filter
      ✓  `return nil, 0, nil` (or `[]*Dial{}, 0, nil`) on an empty result set

§3.2  `[MUST][CODE]`  Represent all search parameters in a filter struct; do not use positional arguments for filtering.
      Each new filter criterion added to a positional signature is a breaking API change. A struct allows future fields to be added without breaking existing callers.
      ✗  `FindDials(ctx, userID int, name string, offset, limit int) ([]*Dial, int, error)`
      ✓  `FindDials(ctx context.Context, filter DialFilter) ([]*Dial, int, error)`

§3.3  `[MUST][CODE]`  Declare optional filter fields as pointers; treat a nil field as "apply no restriction on this dimension."
      Non-pointer fields cannot distinguish "not supplied" from "zero value" — `ID = 0` means either "filter by ID zero" or "no ID filter," which are different.
      ✗  `type DialFilter struct { ID int }` — zero value is ambiguous
      ✓  `type DialFilter struct { ID *int }` — nil means "don't filter by ID"

§3.4  `[MUST][CODE]`  Return the total match count alongside the paged result slice.
      Pagination UI ("page 3 of 12") requires the total count independent of the page size. Omitting it forces a second count query on every page render.
      ✓  `FindDials(ctx, filter) ([]*Dial, int, error)` — the `int` is the total count across all pages; compute it in the same SQL query with `COUNT(*) OVER()`

§3.5  `[MUST][CODE]`  Map sort criteria to a fixed enumeration of SQL expressions; never interpolate an unvalidated sort string into a query.
      Passing raw user-supplied sort strings into SQL is an injection vector. Unindexed columns produce full-table scans.
      ✗  `` "ORDER BY " + filter.SortBy `` interpolated directly into SQL
      ✓
      ```go
      switch filter.SortBy {
      case "updated_at_desc":
          sortBy = "dm.updated_at DESC"
      default:
          sortBy = "dm.id ASC"
      }
      ```

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `ErrNotFound` when a search returns no rows | CODE | §3.1 | Callers must special-case a non-error condition | Return empty slice + nil error |
| Positional arguments for each filter field | CODE | §3.2 | Each new filter is a breaking signature change | Collect all filter fields into a struct |
| Non-pointer optional filter fields | CODE | §3.3 | Zero value is ambiguous | Use pointer fields; nil = "not filtered" |
| Returning only the current page without a total count | CODE | §3.4 | Pagination is impossible without a second round-trip | Return count alongside results |
| Passing raw sort string into SQL | CODE | §3.5 | SQL injection; full-table scan on unindexed columns | Switch over a predefined enumeration of sort options |

---

## 4. Create

§4.1  `[SHOULD][CODE]`  Accept a pointer to the new object; write generated fields (primary key, timestamps) back onto the pointed-to struct rather than returning a second object.
      Returning a second object requires the caller to discard the original and reference the returned copy to access the generated ID. Mutating in place is less cumbersome and avoids an extra allocation.
      ✗  `CreateDial(ctx, dial Dial) (Dial, error)` — caller must capture and switch to the returned copy
      ✓  `CreateDial(ctx context.Context, dial *Dial) error` — `dial.ID` is populated by the function

§4.2  `[SHOULD][CODE]`  Accept nested child objects on the new parent and create the entire graph in a single service call and transaction.
      Requiring the caller to make separate calls for each child forces them to manage partial failure and re-implement transaction semantics outside the service boundary.
      ✗  Caller calls `CreateDial`, then calls `CreateDialMembership` for each member in a separate request
      ✓  `CreateDial` accepts `dial.Memberships` and persists all records atomically

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Returning a new object instead of mutating the pointer | CODE | §4.1 | Caller confusion; extra allocation | Accept `*Dial`; populate generated fields in place |
| Multi-step creation exposed to callers | CODE | §4.2 | Partial failure leaves the graph inconsistent | Create parent + children in one transactional service call |

---

## 5. Update

§5.1  `[MUST][CODE]`  Use a dedicated Update struct with pointer fields for each updatable attribute; pass the target ID as a separate argument.
      Pointer fields distinguish "set this field to zero/empty" from "leave this field unchanged." Separating the ID from the update struct allows the same struct to be reused for bulk updates without a signature change.
      ✗  `UpdateDial(ctx, dial *Dial) error` — all fields are always overwritten; ID embedded in struct cannot be reused for bulk
      ✓  `UpdateDial(ctx, id int, upd DialUpdate) (*Dial, error)` with `type DialUpdate struct { Name *string }`

§5.2  `[SHOULD][CODE]`  Return the updated (or attempted) object even when an error occurs.
      On a validation failure the caller — particularly a stateless HTTP handler — needs the attempted state to re-render the form. Returning `(nil, err)` forces a second fetch or loses the user's input.
      ✗  `return nil, validationErr` — caller cannot re-display the submitted values
      ✓  `return attemptedDial, validationErr` — caller sees exactly what was attempted

§5.3  `[CONSIDER][CODE]`  Design the update signature with a separate ID argument so that bulk updates can be added later without a new function.
      Because `id` is already a separate argument, changing to `ids []int` (and `[]*Dial` return) requires no change to `DialUpdate`.
      ✓  `UpdateDials(ctx, ids []int, upd DialUpdate) ([]*Dial, error)` — the `DialUpdate` struct is reused unchanged

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Passing the full entity struct for update | CODE | §5.1 | All fields unconditionally overwritten; nil and zero values are ambiguous | Use a dedicated Update struct with pointer fields |
| Embedding the target ID in the Update struct | CODE | §5.1 | Prevents reuse for bulk updates | Keep ID as a separate argument |
| Returning `(nil, err)` on validation failure | CODE | §5.2 | Stateless callers lose the submitted state | Return the attempted object alongside the error |

---

## 6. Delete

§6.1  `[MUST][ARCH]`  Enforce ownership or authorization in every delete operation.
      A delete function that accepts only an ID and applies no ownership check allows any authenticated user to delete any record by guessing or iterating IDs.
      ✗  `DELETE FROM dials WHERE id = ?` with no user restriction
      ✓  `DELETE FROM dials WHERE id = ? AND user_id = ?` (or equivalent service-layer ownership check)

§6.2  `[CONSIDER][CODE]`  Accept a slice of IDs rather than a single ID so that bulk deletion requires no future API change.
      Changing from `id int` to `ids []int` is a breaking change; designing for slices from the start keeps the option open at no cost.
      ✓  `DeleteDials(ctx context.Context, ids []int) error` — single deletion passes `[]int{id}`

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Delete without ownership or authorization check | ARCH | §6.1 | Any user can delete any record by supplying an arbitrary ID | Restrict by `user_id` or equivalent in the query or service layer |
| Single-ID delete signature when bulk deletion is plausible | CODE | §6.2 | Bulk support later requires a breaking API change | Accept `[]int` from the start |

---

## Master Anti-Patterns

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Service interface defined inside an implementation package | ARCH | §1.1 | Callers couple to the technology; substitution requires changing all import paths | Define the interface in the root domain package |
| Exposing transactions or raw connections on the service boundary | ARCH | §1.2 | Callers compose pseudo-transactions with no real atomicity guarantees | Each service method owns its own transaction |
| Authorization enforced only in HTTP middleware | ARCH | §1.3 | Direct service callers bypass the check | Enforce inside the service implementation, ideally in the SQL WHERE clause |
| Post-retrieval authorization filtering in application code | ARCH | §1.4 | Fetches unauthorised rows; leaks existence via timing | Push restrictions into the data query |
| Returning `(nil, nil)` from a primary-key lookup | CODE | §2.1 | Callers that check only `err != nil` panic on nil dereference | Return a typed not-found error |
| Bare entity returned with no parent fields | CODE | §2.2 | N+1 fetches in every caller | Populate parent relationships inline |
| Unbounded child collections returned inline | CODE | §2.2 | Payload explosion; slow queries | Fetch large collections separately |
| Not-found error returned from a collection search | CODE | §3.1 | Callers must special-case a non-error condition | Return empty slice + nil error |
| Positional arguments for each filter field | CODE | §3.2 | Each new filter is a breaking signature change | Collect all filter fields in a struct |
| Non-pointer optional filter fields | CODE | §3.3 | Zero value is ambiguous; cannot express "unset" | Use pointer fields; nil = no restriction |
| Returning only the current page without a total count | CODE | §3.4 | Pagination impossible without a second query | Return count alongside results using `COUNT(*) OVER()` |
| Passing raw sort string into SQL | CODE | §3.5 | SQL injection; full-table scan on unindexed columns | Map sort criteria to a fixed enumeration of SQL expressions |
| Returning a new object instead of mutating the input pointer on Create | CODE | §4.1 | Caller confusion; extra allocation | Accept `*Dial`; write generated fields back in place |
| Multi-step creation exposed to callers | CODE | §4.2 | Partial failure leaves the graph inconsistent | Create parent + children in one transactional service call |
| Full entity struct passed to Update (no Update type) | CODE | §5.1 | All fields unconditionally overwritten; zero and nil ambiguous | Dedicated Update struct with pointer fields |
| Target ID embedded in the Update struct | CODE | §5.1 | Prevents reuse for bulk updates | Keep ID as a separate argument |
| Returning `(nil, err)` on validation failure in Update | CODE | §5.2 | Stateless callers lose the submitted state | Return the attempted object alongside the error |
| Delete with no ownership check | ARCH | §6.1 | Any user can delete any record by ID | Restrict by ownership in query or service layer |
