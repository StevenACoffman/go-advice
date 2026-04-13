# Working Effectively with Legacy Go Code

Distilled from Michael C. Feathers' *Working Effectively with Legacy Code* (WELC), adapted for Go's idioms: implicit interfaces, no inheritance, goroutines, and the standard `testing` package.

---

## 1. What Legacy Code Is

**"Legacy code is simply code without tests."** — Feathers

Code without tests is bad code regardless of how clean or well-structured it looks. With tests, behavior can be changed quickly and verifiably. Without them, you don't know whether changes make things better or worse. *Clean code without tests is an aerial gymnastics act without a net.*

Four reasons to change software — they all require the same discipline:
1. Adding a feature
2. Fixing a bug
3. Improving the design (refactoring)
4. Optimizing resource usage

**Behavioral change is what matters**, not the label. Adding new behavior is safe; changing or deleting existing behavior without tests is not.

---

## 2. The Legacy Code Change Algorithm

When you must make a change in a legacy Go codebase:

1. **Identify change points** — where exactly must code change?
2. **Find test points** — where can you insert assertions to detect behavior change?
3. **Break dependencies** — extract interfaces, inject parameters, extract functions
4. **Write tests** — characterization tests first, then tests for new behavior
5. **Make changes and refactor** — with the tests as a safety net

The day-to-day goal is to make functional changes that deliver value *while bringing more code under test*. At the end of each episode, point to both the new feature and its tests. Over time, tested islands grow into continents.

**When breaking dependencies, suspend your sense of aesthetics.** The incision may look ugly; what lies beneath can get healthier. You can heal the scar once you have tests.

---

## 3. What Is (and Isn't) a Unit Test in Go

A good unit test:
- **Runs fast** — sub-millisecond is ideal; anything over 100ms is slow for a unit test
- **Localizes problems** — when it fails, the failure message points directly at the cause

A test is **not** a unit test if it:
- Talks to a real database or calls `sql.Open` with a real DSN
- Makes real HTTP requests or opens real sockets
- Touches the filesystem (reads/writes real files)
- Requires special environment setup (env vars, config files, running services)

These tests have value — write them — but keep them out of the fast feedback loop. Two idiomatic approaches:

- **`testing.Short()` guard** — callers run `go test -short` to skip slow tests (shown below)
- **Build tag** — put slow tests in a file with `//go:build integration` at the top; run with `go test -tags=integration ./...`

Go does not give any special treatment to filenames like `_integration_test.go` — a build tag or `t.Skip` is the actual enforcement mechanism. **If your test suite takes more than a few seconds to run without -short, nobody runs it before committing.**

```go
func TestFoo(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // ... slow test
}
```

---

## 4. The Legacy Code Dilemma

**"When we change code, we should have tests in place. To put tests in place, we often have to change code."**

This circularity is the central challenge. The resolution: make *conservative, safe* initial refactors to get code under test, even if the result looks odd. Initial refactorings that break dependencies leave seams — visible incision points that can be healed with more tests later.

**Do not try to make everything perfect before adding tests.** "Better" is the enemy of "let's actually test something."

---

## 5. Sensing and Separation

Two reasons to break dependencies:

1. **Sensing** — you can't observe what the code computes. The computation happens but no return value or observable side effect is accessible from a test.
2. **Separation** — you can't even construct or call the code in a test harness. It pulls in databases, network connections, global state, or long-running processes.

In Go, both manifest as:
- A function or method that creates its dependencies internally (no injection point)
- A `struct` whose fields are concrete types with side effects
- Package-level `var` state that accumulates across calls
- `init()` functions with side effects
- Calls to `os.Getenv`, `log.Fatal`, or `os.Exit` buried in business logic

The goal is always to find or create a **seam** — a place where behavior can be changed without editing the code under test.

---

## 6. Seams in Go

A seam is a place where you can alter program behavior without changing the code at that place. Every seam has an **enabling point** — where the decision about which behavior to use is made.

**Go's primary seam type is the interface.** Because Go interfaces are implicit and small, extracting them is cheap and requires no changes to the concrete type.

### Interface Seam (preferred)

```go
// Before: hard dependency on real HTTP client
type Fetcher struct{}
func (f *Fetcher) Fetch(url string) ([]byte, error) {
    resp, err := http.Get(url)
    // ...
}

// After: inject via interface
type HTTPClient interface {
    Get(url string) (*http.Response, error)
}

type Fetcher struct {
    client HTTPClient
}
// Enabling point: where Fetcher is constructed
```

The concrete `http.Client` already satisfies `HTTPClient` without any modification.

### Function-Field Seam

When a full interface is overkill for a single function dependency:

```go
type Processor struct {
    now func() time.Time  // seam: injectable in tests
}

// Production:  p := Processor{now: time.Now}
// Test:        p := Processor{now: func() time.Time { return fixedTime }}
```

### Package-Level Variable Seam (last resort)

```go
var timeNow = time.Now  // override in tests; restore in t.Cleanup

func (s *Service) isExpired(t time.Time) bool {
    return t.Before(timeNow())
}
```

This works but is not goroutine-safe across parallel tests. Prefer constructor injection. For environment variable seams specifically, use `t.Setenv` (see §19) which handles save/restore automatically — but only call it from the test goroutine itself, not from goroutines the test spawns.

### Enabling Points in Go

| Seam type | Enabling point |
|-----------|----------------|
| Interface field | `New(dep MyInterface)` constructor |
| Function field | struct literal or `New` func |
| Build tag | `//go:build integration` |
| Package-level var | reassignment in test setup + `t.Cleanup` |

---

## 7. Fake Objects (Test Doubles)

When you can't sense behavior through the real collaborator, use a fake.

In Go, a **fake** is a struct that implements the interface used by the code under test, with hand-written logic simple enough to understand at a glance:

```go
type FakeMailSender struct {
    Sent []Message
}

func (f *FakeMailSender) Send(m Message) error {
    f.Sent = append(f.Sent, m)
    return nil
}
```

Test code has different standards from production code: **breaking encapsulation for observability is fine in tests.** Public fields on fakes, `t.Log` output, and exported sensing variables are all acceptable. What matters is that the test code is *clean and readable* — duplication in tests is as harmful as in production code; extract common setup into helpers.

For more sophisticated behavior (call verification, canned return values), use `github.com/stretchr/testify/mock` or hand-roll; prefer hand-rolling unless the interface is large.

**Use `httptest.NewServer` and `httptest.NewRecorder` for HTTP dependencies** — these are Go's built-in test doubles for HTTP and are almost always better than mocking `http.Client`.

---

## 8. Characterization Tests

When you need to change code and don't know what it does, write **characterization tests**: tests that document the *actual current behavior*, not what you think it should do.

**Algorithm:**
1. Put the code in a test harness.
2. Write an assertion you know will fail.
3. Let the failure tell you what the behavior actually is.
4. Change the test to expect the actual output.
5. Repeat for other inputs and edge cases.

```go
func TestLegacyFormatter_characterize(t *testing.T) {
    f := NewLegacyFormatter()
    got := f.Format(testInput)
    // Step 2: this will fail — the output tells you the actual value
    t.Fatalf("got: %q", got)
    // Step 4: replace t.Fatalf with: if got != "<actual value>" { t.Errorf(...) }
}
```

These tests have no "moral authority" — they don't confirm correctness, they confirm *current behavior*. Their value is as a regression detector: if a future change breaks a characterization test, you know you changed behavior you didn't intend to.

**Where to focus characterization tests:**
- Code paths you're about to modify
- The boundary between changed and unchanged code
- Any output that another system or user depends on

Write them around the **pinch point** — the narrowest place in the call graph that an effect must pass through. A few pinch-point tests cover many lower-level behaviors.

---

## 9. Adding Features Safely Without Full Test Coverage

When you can't write tests around the old code *yet*, these techniques let you add new code that is itself testable:

### Sprout Function / Method

Write new behavior as a separate, pure function. Call it from the untested legacy code. Test the function directly.

```go
// Old (untested monster function):
func (g *Gateway) processEntries(entries []Entry) {
    // 200 lines of legacy code
    // ... we must add: skip duplicate entries
    // DON'T add the logic inline
    unique := uniqueEntries(entries, g.existing)  // sprout
    // use unique instead of entries
}

// New (testable sprout):
func uniqueEntries(entries []Entry, existing map[string]bool) []Entry {
    var result []Entry
    for _, e := range entries {
        if !existing[e.ID] {
            result = append(result, e)
        }
    }
    return result
}
```

The sprout is tested independently. The call site in legacy code is the seam — it stays untested for now but is visible and minimal.

### Sprout Type (Sprout Struct)

When the existing struct can't be instantiated in a test harness and you can't afford to fix that now:

```go
// Old (untestable):
type QuarterlyReport struct { /* massive dependencies */ }
func (r *QuarterlyReport) Generate() string { /* 400 lines */ }

// New (testable, sprouted):
type ReportHeader struct{}
func (h *ReportHeader) Render(cols []string) string {
    // small, testable, new behavior
}

// In legacy code, just instantiate and call ReportHeader
```

### Wrap Function

Rename the old function; write a new function with the original name that calls old behavior plus new behavior:

```go
// Before:
func (s *Service) Save(data []byte) error { /* legacy */ }

// After:
func (s *Service) Save(data []byte) error {
    if err := s.validate(data); err != nil {  // new, testable
        return err
    }
    return s.saveImpl(data)  // old, renamed
}
func (s *Service) saveImpl(data []byte) error { /* original body */ }
```

Test `validate` independently. Test `Save` with a characterization test.

When to prefer Wrap Function over Sprout Function: when the new feature is **as important** as the original behavior — both deserve equal visibility at the call site. After wrapping, you often end up with a clean high-level sequence: `validate → execute → notify`.

### Wrap Type

The struct-level companion to Wrap Function. Add behavior to an existing type by creating a new type that holds the original and delegates to it. In Go this replaces the OO "Wrap Class / Decorator" pattern:

```go
// Old type (can't or won't modify):
type LegacyPayroll struct{ /* complex */ }
func (p *LegacyPayroll) Pay(emp Employee) { /* side effects */ }

// Wrapper: same interface, extra behavior
type LoggingPayroll struct {
    inner *LegacyPayroll
    log   Logger
}
func (lp *LoggingPayroll) Pay(emp Employee) {
    lp.log.Printf("paying %s", emp.Name)
    lp.inner.Pay(emp)
}
```

Define a shared interface (`Payroll`) so callers don't need to know which implementation they have. If the new behavior must happen at **every** existing call site, use the decorator form above. If it's needed only in one new location, a thin coordinator struct (not a decorator) is simpler.

**When to use Wrap Type:**
1. The new behavior is completely independent and you don't want to pollute the existing type.
2. The existing type has so many responsibilities that you refuse to make it worse — the wrapper is a stake in the ground, a visible roadmap for future decomposition.

---

## 10. Programming by Difference in Go

**Programming by Difference** means: introduce new behavior through a new type (or function) rather than modifying the existing one, test the new type first, then refactor the original once you have a safety net.

In OO languages this uses subclassing. Go has no subclassing, but the idea translates directly:

1. **Create a new type** that holds the original as a field and implements the same interface.
2. **Override only the differing behavior** in the new type.
3. **Write tests against the new type** — they pass immediately.
4. **Use the tests as a safety net** to refactor the original toward the cleaner design.

```go
type MessageForwarder struct { listAddress string }
func (m *MessageForwarder) FromAddress(msg Message) string {
    if len(msg.From) > 0 { return msg.From[0] }
    return m.defaultFrom()
}

// New requirement: anonymous lists hide the sender
type AnonymousForwarder struct {
    *MessageForwarder   // embed for all other behavior
    domain string
}
func (a *AnonymousForwarder) FromAddress(_ Message) string {
    return "anon-members@" + a.domain
}

// Test passes immediately; no changes to MessageForwarder.
```

**After the test is green**, fold the variation back in — replace the subtype with a configuration option or a strategy injection:

```go
type MessageForwarder struct {
    listAddress string
    domain      string
    anonymous   bool   // or: fromFn func(Message) string
}
func (m *MessageForwarder) FromAddress(msg Message) string {
    if m.anonymous { return "anon-members@" + m.domain }
    if len(msg.From) > 0 { return msg.From[0] }
    return m.defaultFrom()
}
```

**Embedding is not polymorphism.** When `AnonymousForwarder` embeds `*MessageForwarder`, calling any method on `*MessageForwarder` that internally calls `FromAddress` will call `MessageForwarder.FromAddress`, NOT `AnonymousForwarder.FromAddress`. Go has no virtual dispatch — the receiver type is always resolved statically. This is fundamentally different from OO subclassing. It means the embedding approach is only safe when the embedded type's methods don't call each other through the method name being "overridden".

**The Liskov trap:** Watch for the case where the new type's behavior is inconsistent with the contract of the interface (e.g., ignores its argument when the interface promises to use it). That's a design smell. Normalize the hierarchy — or switch to strategy injection — before it proliferates.

**Key insight:** "Programming by Difference lets us introduce variations quickly. We use tests to pin down the new behavior and move to more appropriate structures when we need to."

---

## 11. Dependency-Breaking Techniques (Full Catalog)

These are the core moves for getting untestable Go code under test.

### Extract Interface

Define a minimal Go interface covering only the methods you actually call. The concrete type satisfies it automatically; no changes needed.

```go
// Define the interface in terms of what YOUR code needs — not what the library returns.
type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
}

// Thin adapter that makes *redis.Client satisfy Cache:
type redisCache struct{ c *redis.Client }

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
    return r.c.Get(ctx, key).Result()
}
func (r *redisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
    return r.c.Set(ctx, key, value, ttl).Err()
}
func (r *redisCache) Del(ctx context.Context, keys ...string) error {
    return r.c.Del(ctx, keys...).Err()
}
```

**Do not put library-specific types in your interface.** If you write `Get(...) *redis.StringCmd`, callers still depend on the redis package and test fakes must return redis objects. The adapter pattern above keeps the interface clean: callers use `Cache`, tests fake `Cache` with plain Go values, and only `redisCache` imports the redis library.

**Naming tension:** If you want the interface named `Store` but the concrete type is already called `Store`, prefix the concrete type: `PostgresStore`, `FileStore`, etc. Avoid `IStore` — it's noise.

### Parameterize Constructor (Dependency Injection via `New`)

Move dependency creation out of the constructor. Keep a convenience constructor for production use:

```go
// Before:
func NewProcessor() *Processor {
    db := openRealDB()  // untestable
    return &Processor{db: db}
}

// After:
func NewProcessor(db Database) *Processor {
    return &Processor{db: db}
}
// Convenience for production:
func NewDefaultProcessor() *Processor {
    return NewProcessor(openRealDB())
}
```

### Extract and Override via Function Field

Extract a hard-coded call to a function field on the struct. This is the Go equivalent of "extract and override a factory/getter method in a subclass":

```go
type Service struct {
    clock func() time.Time  // seam: swap in tests
}
func (s *Service) isExpired(ts time.Time) bool {
    return ts.Before(s.clock())
}

// Production:  &Service{clock: time.Now}
// Test:        &Service{clock: func() time.Time { return fixedTime }}
```

Go has no virtual dispatch, so there is no way to "override" a method via embedding the way you would in an OO language. The function-field seam above is the idiomatic substitute. Embedding is useful for *adding* methods or *promoting* existing ones, but it cannot intercept a method call that the embedded type makes to itself (see §10, "Embedding is not polymorphism").

### Replace Global with Parameter

Global state is a dependency seam that is invisible and untestable. Pass it instead:

```go
// Before:
var globalConfig Config
func Process(input string) string {
    return globalConfig.Format(input)
}

// After:
func Process(cfg Config, input string) string {
    return cfg.Format(input)
}
```

If you must keep the global for callers that can't change, add a wrapper:

```go
func ProcessDefault(input string) string {
    return Process(globalConfig, input)
}
```

### Skin and Wrap an External API

When the third-party library is hard to fake (final classes in Java → non-interface types in Go without exported fields):

1. Define a Go interface that covers what you need.
2. Write a thin struct that delegates to the real library.
3. Inject the interface in tests; use the thin wrapper in production.

```go
// Interface uses only domain types — no AWS SDK imports needed in callers or tests.
type ObjectStore interface {
    Put(ctx context.Context, bucket, key string, body io.Reader, size int64) error
}

// Adapter: the only place that imports the AWS SDK.
type s3Store struct{ real *s3.Client }

func (w *s3Store) Put(ctx context.Context, bucket, key string, body io.Reader, size int64) error {
    _, err := w.real.PutObject(ctx, &s3.PutObjectInput{
        Bucket:        &bucket,
        Key:           &key,
        Body:          body,
        ContentLength: &size,
    })
    return err
}

// Test fake: no AWS SDK dependency.
type fakeStore struct{ puts []string }
func (f *fakeStore) Put(_ context.Context, bucket, key string, _ io.Reader, _ int64) error {
    f.puts = append(f.puts, bucket+"/"+key)
    return nil
}
```

The adapter is the *only* file that imports the AWS SDK. Everything else in the codebase uses `ObjectStore`. If the SDK is upgraded or swapped, only `s3Store` changes.

### Parameterize io.Reader / io.Writer

Go's `io.Reader` and `io.Writer` are natural seams. Replace `os.Stdin`/`os.Stdout`/file paths with interfaces:

```go
// Before:
func process() { data, _ := os.ReadFile("input.txt"); ... }

// After:
func process(r io.Reader) { data, _ := io.ReadAll(r); ... }
// Test: pass strings.NewReader("test data")
// Production: pass os.Open("input.txt")
```

### Break Dependency on `os.Exit` and `log.Fatal`

`os.Exit` and `log.Fatal` in business logic make code untestable. Extract them:

```go
// Before:
func run() {
    if err := validate(); err != nil {
        log.Fatal(err)  // kills the process; can't test
    }
}

// After:
func run() error {
    if err := validate(); err != nil {
        return err
    }
    return nil
}
// main() calls run() and handles exit
```

### Adapt Parameter

When you can't extract an interface from a parameter's type (e.g., it's a large external API with 20+ methods you don't own), **wrap just the slice you use**:

```go
// External type you can't modify:
// r *http.Request has 30 methods, you use 1

// Define a narrow interface for what you actually need:
type ParameterSource interface {
    GetParam(name string) string
}

// Production adapter:
type requestAdapter struct{ r *http.Request }
func (a *requestAdapter) GetParam(name string) string {
    return a.r.FormValue(name)
}

// Test fake:
type fakeParams struct{ values map[string]string }
func (f *fakeParams) GetParam(name string) string { return f.values[name] }

// Function under test now depends on the narrow interface:
func populate(src ParameterSource) { ... }
```

**Move toward interfaces that communicate responsibilities rather than implementation details.** A narrow interface named `ParameterSource` tells readers what the code needs; `*http.Request` just tells them what infrastructure it happens to be using.

### Break Out Method Object

When a monster function uses too many local variables to extract neatly, move the function to a new struct. Each local variable becomes a field:

```go
// Before: huge method with many locals
func (g *GDIBrush) Draw(roots []Point, colors ColorMatrix, sel []Point) {
    // 80 lines using lots of intermediate state
}

// After: new struct captures all state
type Renderer struct {
    brush  *GDIBrush
    roots  []Point
    colors ColorMatrix
    sel    []Point
}
func (r *Renderer) Draw() {
    // same 80 lines, now with fields instead of locals
    // — and fields can be sensed in tests
}

// GDIBrush delegates:
func (g *GDIBrush) Draw(roots []Point, colors ColorMatrix, sel []Point) {
    (&Renderer{brush: g, roots: roots, colors: colors, sel: sel}).Draw()
}
```

Now you can test `Renderer.Draw` by controlling its fields, which was impossible when everything was a local.

### Encapsulate Global References

Group related globals into a struct so they can be injected:

```go
// Before: scattered globals
var activeFrame [256]bool
var suspendedFrame [256]bool

func suspendFrame() {
    copy(suspendedFrame[:], activeFrame[:])
    clearActive()
}

// After: bundle in a struct, inject it
type Frame struct {
    Active    [256]bool
    Suspended [256]bool
}

type AGGController struct{ frame *Frame }
func (c *AGGController) SuspendFrame() {
    copy(c.frame.Suspended[:], c.frame.Active[:])
    c.clearActive()
}
```

**Rule:** If several globals are always used together or modified near each other, they belong in the same struct.

### Expose as Package Function (Expose Static Method)

When a method doesn't use receiver state, make it a standalone package function so it can be tested without instantiating the struct:

```go
// Can't instantiate RSCWorkflow in tests (too many deps):
func (r *RSCWorkflow) Validate(p Packet) error { ... }

// Extract the stateless logic:
func ValidatePacket(p Packet) error { ... }

// Original delegates:
func (r *RSCWorkflow) Validate(p Packet) error { return ValidatePacket(p) }
```

Test `ValidatePacket` directly. Move it to the right type later when you have coverage.

### Extract and Override Call

When a single method call creates a hard dependency (e.g., a global function, a static call), extract the call to a function field or an interface method:

```go
// Before:
func (p *PageLayout) rebindStyles() {
    p.styles = StyleMaster.FormStyles(p.template, p.id)  // global dep
}

// After: extract the call to an injectable function
type PageLayout struct {
    template     string
    id           int
    formStylesFn func(tmpl string, id int) []Style  // override in tests
}
func (p *PageLayout) rebindStyles() {
    p.styles = p.formStylesFn(p.template, p.id)
}

// Production:   p.formStylesFn = StyleMaster.FormStyles
// Test:         p.formStylesFn = func(string, int) []Style { return nil }
```

Prefer this over Replace Global Reference with Getter when there is **one** call site. Use the getter approach when many methods call the same global.

### Extract and Override Factory Method (Parameterize Construction)

When a constructor hard-codes object creation, move the creation to a factory function that can be overridden:

```go
// Before: WorkflowEngine builds TransactionManager itself
func NewWorkflowEngine() *WorkflowEngine {
    reader := newModelReader(appConfig.GetDry())
    store  := newXMLStore(appConfig.GetDry())
    tm     := newTransactionManager(reader, store)
    return &WorkflowEngine{tm: tm}
}

// After: factory is injectable
type WorkflowEngine struct {
    makeTM func() TransactionManager  // seam; must not be nil
    once   sync.Once
    tm     TransactionManager
}

// Production constructor wires in the real factory.
func NewWorkflowEngine() *WorkflowEngine {
    return &WorkflowEngine{
        makeTM: func() TransactionManager {
            reader := newModelReader(appConfig.GetDry())
            store  := newXMLStore(appConfig.GetDry())
            return newTransactionManager(reader, store)
        },
    }
}

func (w *WorkflowEngine) transactionManager() TransactionManager {
    w.once.Do(func() { w.tm = w.makeTM() })
    return w.tm
}

// Test:
engine := &WorkflowEngine{makeTM: func() TransactionManager { return &fakeTM{} }}
// Note: use sync.Once for the lazy-init guard — a bare nil check is not goroutine-safe.
```

### Introduce Instance Delegator (for Package-Level Functions)

When a package-level function is hard to substitute in tests, add an instance method that delegates to it, then inject the instance:

```go
// Package-level function (hard to replace):
func updateAccountBalance(userID int, amount Money) { ... }

// Add an instance method to a struct:
type BankingServices struct{}
func (b *BankingServices) UpdateBalance(userID int, amount Money) {
    updateAccountBalance(userID, amount)
}

// Caller receives *BankingServices via interface:
type BalanceUpdater interface {
    UpdateBalance(int, Money)
}
```

Over time, move the body of `updateAccountBalance` into `UpdateBalance` and delete the package-level function.

### Introduce Static Setter (for Singletons / Package-Level State)

When a global or singleton can't be parameterized immediately, add a setter so tests can substitute a fake:

```go
var defaultRouter Router = &productionRouter{}  // package global

// Add a setter:
func SetRouter(r Router) { defaultRouter = r }

// In tests:
func TestX(t *testing.T) {
    orig := defaultRouter          // capture before overwriting
    SetRouter(&fakeRouter{})
    t.Cleanup(func() { SetRouter(orig) })  // restore original — not a new instance
    // ...
}
```

**Warning:** Always reset in `t.Cleanup`. Singletons with static setters create shared state between tests — do not call `t.Parallel()` in any test that mutates this global.

### Parameterize Method

When a method creates an object internally that you need to control, add it as a parameter:

```go
// Before:
func (tc *TestCase) Run() {
    result := newTestResult()  // hard-coded
    tc.runTest(result)
}

// After:
func (tc *TestCase) Run() { tc.RunWith(newTestResult()) }
func (tc *TestCase) RunWith(result TestResult) {
    tc.runTest(result)
}
// Tests call RunWith(fakeResult)
```

### Pull Up Feature

When the methods you want to test are entangled with dependencies you can't instantiate, move the testable methods to a separate type (in Go: extract to a new struct or package):

```go
// BEFORE: getDeadtime is testable, but it lives on Scheduler which hits the DB.
type Scheduler struct { db Database; items []Item }
func (s *Scheduler) getDeadtime() int { ... }          // pure logic — testable
func (s *Scheduler) validate(item Item) error { ... }  // needs s.db — not testable

// AFTER: extract the pure part into its own type.
type SchedulingServices struct {
    items []Item
}
func (ss *SchedulingServices) GetDeadtime() int { ... }

// Scheduler now embeds SchedulingServices and keeps the DB dependency.
type Scheduler struct {
    SchedulingServices
    db Database
}
func (s *Scheduler) validate(item Item) error { ... }  // DB dep stays here
```

Test `SchedulingServices.GetDeadtime` directly without touching the DB.

### Push Down Dependency

When most methods are testable but a few have nasty dependencies, make the original type an interface (or embed point) and push the problematic behavior to a concrete subtype:

```go
// OffMarketValidator: isValid() is testable, showDialog() needs the UI
type OffMarketValidator struct {
    trade Trade
    flag  bool
    ShowMessage func()  // pushed down: inject for production, no-op for tests
}
func (v *OffMarketValidator) IsValid() bool {
    if inRange(v.trade) && validDest(v.trade) && inHours(v.trade) {
        v.flag = true
    }
    v.ShowMessage()
    return v.flag
}

// Test: set ShowMessage to func(){}
// Production: set ShowMessage to func(){ slog.Warn("off-market trade", "trade", v.trade) }
```

### Replace Global Reference with Getter

Wrap access to a global in a method so it can be overridden in a testing subtype or intercepted via a function field:

```go
// Before:
func (r *RegisterSale) AddItem(code Barcode) {
    item := GlobalInventory.ItemForBarcode(code)  // global dep
    r.items = append(r.items, item)
}

// After:
type InventorySource interface {
    ItemForBarcode(Barcode) Item
}

type RegisterSale struct {
    inventory InventorySource  // injected
    items     []Item
}
func (r *RegisterSale) AddItem(code Barcode) {
    item := r.inventory.ItemForBarcode(code)
    r.items = append(r.items, item)
}
```

### Subclass and Override → Interface + Fake

Go's equivalent of "subclass and override method" is **define an interface for the behavior you want to control and implement it with a fake**:

```go
// The production code calls some external operation:
type Logger interface {
    Log(msg string)
}

// Production: inject real logger
// Test: inject fake that captures calls
type fakeLogger struct{ logged []string }
func (f *fakeLogger) Log(msg string) { f.logged = append(f.logged, msg) }
```

### Supersede Instance Variable

When a constructor builds an object you need to replace in tests but can't parameterize yet, add a setter method:

```go
type BlendingPen struct {
    param Parameter
}

func NewBlendingPen() *BlendingPen {
    p := parameterFactory.Create("cm", "Fade", "Aspect Alter")
    p.AddChoice("blend")
    return &BlendingPen{param: p}
}

// Supersede method for tests:
func (b *BlendingPen) SetParameter(p Parameter) {
    b.param = p
}

// Test:
pen := NewBlendingPen()
pen.SetParameter(&fakeParameter{})
```

Put the setter in a `_test.go` file — a `_test.go` file is compiled only during `go test`, so the setter literally cannot be called from production code:

```go
// blending_pen_test.go  (only compiled during testing)
func (b *BlendingPen) setParameter(p Parameter) {
    b.param = p
}
```

If the field cannot be unexported and you need the setter in a non-test file, name it `setXXXForTest` and document it as test-only. Either way, grep for the name to confirm it's never called from production paths.

### Break Dependency on `init()`

`init()` functions run automatically before `main()` and before any test in the package — you cannot call or skip them. Code buried in `init()` is untestable by definition.

**Technique:** Move `init()` body to a named function; have `init()` call it. Tests can then call the named function (or not call it at all):

```go
// Before: untestable side effect in init
func init() {
    globalCache = newRedisCache(os.Getenv("REDIS_URL"))
}

// After: extractable, testable
var globalCache Cache

func initCache(redisURL string) {
    globalCache = newRedisCache(redisURL)
}

func init() { initCache(os.Getenv("REDIS_URL")) }
```

Tests that don't need the real cache simply skip calling `initCache` and inject a fake through the normal constructor path. Tests that do need to exercise the init path call `initCache` with a test URL.

**Deeper fix:** If the `init()` is setting package-level state that multiple tests share, the real fix is to eliminate the global — parameterize it into the types that need it (see "Encapsulate Global References"). Treat `init()` bodies the same way you'd treat code in `main()`: push it out to the edges where side effects are expected, and keep the core testable.

### Build Tag Substitution (Link Substitution)

Go's build tag system lets you provide alternate implementations at build time:

```go
// db_real.go:
//go:build !test
package storage
func OpenDB() *sql.DB { return openRealConnection() }

// db_fake.go:
//go:build test
package storage
func OpenDB() *sql.DB { return openInMemoryDB() }
```

Run tests with `go test -tags test ./...`. Use sparingly — interface injection is almost always cleaner.

---

## 12. Understanding Unfamiliar Legacy Code

**Don't try to hold the whole thing in your head.** Use external aids:

### Sketch Call Graphs

Draw boxes and arrows for the functions/types you're touching. Doesn't have to be UML — informal blobs and lines work. The act of drawing externalizes your mental model and catches misconceptions.

### Effect Sketches

For a change you're about to make:
1. Mark the line(s) you're changing.
2. Mark every variable or return value that can be affected.
3. Mark everything affected by those, transitively.

This reveals what to test before touching anything.

### Listing Markup

Print or annotate long functions in your editor:
- Use regions/comments to group logically related lines
- Mark where temporary variables are declared vs. where they're last used
- Circle potential extract-function boundaries
- Note the **coupling count** (how many inputs/outputs cross the boundary) — lower is better

### Scratch Refactoring

When a function is too large to understand in a reading pass:
1. Check out a scratch branch.
2. Refactor without tests — rename, extract, restructure — just to understand.
3. **Throw the scratch away.** Don't commit it.
4. Now write tests and refactor properly on the real branch.

Scratch refactoring is for learning, not for shipping.

### Delete Unused Code

If you're not sure whether code is used, delete it. Compilation and tests will tell you. Legacy code accumulates dead weight; the compiler is your ally in Go because unused imports and variables are compile errors.

### Tell the Story of the System

When architecture has been lost to time and schedule pressure, recover it through narrative. **Tell the story with a partner:**

1. One person asks: "What is the architecture of this system?"
2. The other explains using **only 2–3 concepts** — pretend the listener knows nothing.
3. Say the simplest thing, even if it feels like a lie of omission.
4. Then add the next most important things, one layer at a time.

The feeling that you're "lying" by oversimplifying is the key signal: **that feeling reveals the gap between what the system does and what it ideally would do.** A simpler story describes a simpler, better architecture. Each simplification is a refactoring opportunity.

Tell the story often, in different ways. Rotate who tells it. The concepts people naturally use in conversation — the ones that make the system click — are the abstractions the code should contain. **If there isn't a strong overlap between conversation and code, something is wrong.**

"Design is design, regardless of when it happens in the development cycle. One of the worst mistakes a team can make is to feel that design is over."

### Naked CRC (Index Card Conversations)

To understand or communicate the structure of a system without UML, use blank index cards on a table:

- Each card represents a **runtime instance** (not a class/type — an instance)
- **Overlap cards** to represent a collection of them
- Lay cards down one by one as you describe the system; point at them, move them
- No writing on the cards — motion and position carry meaning

This technique makes design tangible. It forces you to think in terms of running objects interacting, not static type diagrams. Interactions described this way are more memorable.

**Listen to your design conversations.** The concepts people naturally use when talking about the system are usually the abstractions the code should contain. When you notice that you're saying "there's a locking policy" but there's no `LockingPolicy` type, that's a seam waiting to be created.

---

## 13. Monster Functions in Go

A monster function is one too large and complex to test or change confidently. Go functions should nearly always fit on a screen.

### Two Varieties

**Bulleted function**: Nearly no nesting — a sequence of steps. Looks like a to-do list. Temporary variables declared in one block are used in the next, making naive extraction fail.

**Snarled function**: Deep `if`/`switch`/`for` nesting. Variables at outer scope are used inside inner blocks, creating extraction coupling.

### Strategies

**For bulleted functions:**
Extract each logical chunk into its own named function. The temporary variables shared between chunks become either parameters or a small struct passed between them.

```go
// Before:
func processOrder(o Order) error {
    // block 1: validate
    // block 2: price
    // block 3: apply discounts
    // block 4: save
}

// After:
func processOrder(o Order) error {
    if err := validateOrder(o); err != nil {
        return err
    }
    price, err := calculatePrice(o)
    if err != nil {
        return err
    }
    discounted := applyDiscounts(o, price)
    return saveOrder(o, discounted)
}
```

**For snarled functions:**

Introduce **sensing variables** — local variables that capture intermediate values — to make the function's behavior observable without changing its structure:

```go
var (
    appliedDiscount bool
    finalPrice      float64
)
// ... existing logic that sets these ...
// Now test by exposing them (temporarily) or returning them
```

Then use these as targets for extraction.

**Coupling count**: Before extracting a block, count inputs (variables used by the block but declared outside it) plus outputs (variables assigned by the block and used after it). Blocks with coupling count 0 or 1 extract cleanly. High coupling count means the block isn't a natural unit — look for a better boundary.

### Skeletonize vs. Find Sequences

Two complementary extraction strategies:

**Skeletonize** — extract condition and body to *separate* functions, leaving only the control structure:
```go
// Before:
if marginalRate() > 2 && order.HasLimit() {
    order.Readjust(calculator.RateForToday())
    order.Recalculate()
}

// After (skeletonized — control structure stays, details go elsewhere):
if orderNeedsRecalculation(order) {
    recalculateOrder(order, calculator)
}
```

**Find sequences** — extract condition and body *together* to reveal a higher-level sequence:
```go
// After (sequence found):
recalculateOrder(order, calculator)  // condition is internal
```

Use skeletonizing when you expect the control structure to be refactored further. Use find-sequences when the method is really just a linear pipeline that will be clearer as a sequence of calls. **Bulleted methods lean toward finding sequences; snarled methods lean toward skeletonizing.**

Always: **extract to the current struct first.** Even if logic logically belongs on another type, use an awkward name in the current struct first, make it compile, then move it. Direct extraction to another type is riskier without tests.

### Gleaning Dependencies

When critical behavior is tangled with secondary behavior and you can only write tests for part of it:

1. Write tests covering the *critical* behavior (the logic you can't afford to break).
2. Extract and move the *secondary* behavior (the code not covered by those tests) into its own function.
3. You can be confident the extraction doesn't break the critical path because the tests tell you immediately if it does.

Not all behaviors are equal. Some are more critical, and recognizing that lets you refactor selectively even without full coverage.

---

## 14. Large Packages in Go

A Go package with 20+ files and 5000+ lines is the equivalent of a monster class. Apply the same analysis:

### Finding Responsibilities

For each exported type/function in a large package:
1. Write one sentence describing its responsibility.
2. If the sentence contains "and", it may have more than one responsibility.
3. Group types/functions that change together and for the same reasons.
4. The groups are candidate sub-packages.

### Feature Sketches for Packages

Draw a graph where nodes are types and functions, edges are calls or field access. Clusters with few edges crossing the cluster boundary are natural sub-packages.

### Pinch Points

A pinch point is where many effects funnel through a narrow interface. Find pinch points by tracing what must change if a specific internal type changes. Write tests at the pinch point to cover refactoring of the internals beneath it.

### Heuristics for When to Extract a Sub-package

- The type has dependencies (imports) that other types in the package don't need.
- The type has more than one test file because its tests are getting unwieldy.
- You can describe the type's responsibility without mentioning any other types in the package.
- A different team or future package will want to import just this part.

---

## 15. The Single Responsibility Principle in Go

**A type should have exactly one reason to change.** If you can't describe what a `struct` does in a single sentence, it has too many responsibilities.

SRP violations in Go:
- A struct that both fetches data *and* transforms it *and* sends it somewhere
- A package that contains types for three unrelated domain concepts
- A function that validates, transforms, logs, and persists in one body

**Two levels of violation:**
- **Interface level**: The struct's exported methods span wildly different concerns. Solution: split the type.
- **Implementation level**: The type *delegates* to smaller types that each have one job. Less harmful — the struct is then a facade.

Facade structures are acceptable in Go. If a large struct merely composes several smaller ones and delegates, that's fine. The test problem is that you still have to instantiate all the smaller ones. Use constructor injection so tests can pass fakes.

---

## 16. Removing Duplication — Designs Emerge

**Zealous duplication removal reveals design.** When behavior exists in exactly one place, changes become one-knob operations. Orthogonality — one knob per behavior — is the goal.

Process:
1. Find code that is identical or nearly identical in two places.
2. Parameterize the variation.
3. Extract the common form to a shared function.
4. Name the extracted function after what it does, not after the code it replaced.

The names you invent during extraction are design acts. They reveal concepts that were implicit in the original code.

In Go, prefer table-driven tests when duplication shows up in tests — they make the variation explicit and the structure clear.

```go
tests := []struct {
    name  string
    input string
    want  string
}{
    {"empty", "", ""},
    {"single word", "hello", "HELLO"},
    {"with spaces", "hello world", "HELLO WORLD"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        assert.Equal(t, tt.want, process(tt.input))
    })
}
```

---

## 17. Safe Editing Practices

### Hyperaware Editing

Every keystroke you make either changes the behavior of the software or it doesn't. Know which one you're doing.

- Typing in a comment: does not change behavior.
- Changing a string literal in executed code: changes behavior.
- Reformatting: does not change behavior (but is a refactoring in the microscopic sense).
- Changing a numeric literal: changes behavior.

**Tests make you hyperaware.** When you can run tests in under a second, you know at every step whether you've changed behavior. Without tests, you're running on memory and hope.

"Programming is the art of doing one thing at a time." When you find yourself wanting to clean up something adjacent to your change: write the name on a slip of paper, finish the change you started, then come back. Don't refactor and change behavior simultaneously.

Hyperaware editing can be a *flow state* — intense and refreshing, not exhausting. What's exhausting is editing without feedback, holding fragile state in your head, not knowing whether you've broken something.

### Preserve Signatures

When you're breaking dependencies without tests (the initial safety-breaking work), **copy and paste method signatures rather than retyping them**. The most common source of bugs during dependency-breaking is typos in parameter names, types, or order.

Steps for safe extraction:
1. Copy the entire parameter list from the original function.
2. Create the new function signature by pasting.
3. Create the call site by pasting again and deleting the type names.

This mechanical discipline keeps your hands from introducing errors while your mind is focused on structure.

### Lean on the Compiler

When renaming a type, changing an interface, or moving behavior, let compile errors guide you to all the places that need updating:

```go
// Change a struct field type or remove a method:
// The compiler will list every call site that breaks.
// Navigate to each error and fix it.
```

**Caveats in Go:**
- Interfaces are satisfied implicitly. Removing a method from an interface won't generate errors at the implementation — it may silently leave implementations with dead methods.
- Renaming an exported identifier with a refactoring tool is safer than Lean on the Compiler for this case.
- Lean on the Compiler works best for concrete type changes, not interface changes.

**Compile-time interface satisfaction check** — place this at the top of any implementation file to turn implicit satisfaction into an explicit compile error:

```go
// Asserts that *MyImpl satisfies MyInterface at compile time.
var _ MyInterface = (*MyImpl)(nil)
```

This costs nothing at runtime and catches the case where a refactor silently breaks interface satisfaction. Use it on any type you intend to pass through an interface seam.

---

## 18. When You Feel Overwhelmed

Working in legacy code is genuinely difficult. The antidote is not stoicism; it's finding what motivates you and connecting that to the work.

**On the green-field illusion.** Teams working on legacy systems often fantasize about a rewrite. The rewrite team starts fresh, knows the pitfalls, makes better decisions — until they realize the old system is still being maintained, still changing, and the rewrite has to track all of it. The "replacement system with a better architecture" frequently fails to ship. The legacy system you're maintaining *is* the investment everyone depends on.

**Practical strategies when morale is low:**

1. **Pick the ugliest, most obnoxious set of types in the project** and get them under test as a team. When you've tackled the worst problem, you feel in control.

2. **TDD something outside of work** — a small side project. Feel the difference between code that has a fast test harness and code that doesn't. Then bring that same discipline to the project at work.

3. **Make a list of the changes that matter** — don't try to understand the whole thing at once. Pick the one change with the most value or the highest risk. Address that first.

4. **Wrap before rewriting.** Write a new function/type alongside the legacy one. Let clients migrate incrementally. The legacy code dies of starvation as callers switch.

5. **Leave the campsite cleaner than you found it.** Every time you touch a function, add at least one test, rename one unclear variable, and extract one monster block.

As you start to take control, you'll develop **oases of good code**. Work in the oases is genuinely enjoyable. The oases grow.

**Code is your house and you have to live in it.** Decisions that save time today by skipping tests cost far more than that time over the next six months. The goal is to get to the point where change is *expected* to be easy, and when it isn't, you fix that, too.

---

## 19. Go-Specific Patterns and Tools

### Table-Driven Tests
Standard Go idiom for characterization and regression tests. Captures all the cases in one place.

### `testing/quick` for Simple Property Testing
For pure functions, `testing/quick.Check` generates random inputs to find unexpected edge cases:
```go
f := func(s string) bool {
    return normalize(normalize(s)) == normalize(s) // idempotent?
}
quick.Check(f, nil)
```
`testing/quick` is simple but has no failure shrinking — when it finds a counter-example you get the raw random input. Use it for quick sanity checks on pure functions.

### Native Fuzzing (`testing.F`) — Go 1.18+
For deeper exploration with automatic failure shrinking, use Go's built-in fuzzer:
```go
func FuzzNormalize(f *testing.F) {
    f.Add("hello")          // seed corpus
    f.Add("HELLO WORLD")
    f.Fuzz(func(t *testing.T, s string) {
        // Must not panic; idempotency holds for any input:
        once := normalize(s)
        twice := normalize(once)
        if once != twice {
            t.Errorf("normalize(%q) = %q, normalize again = %q", s, once, twice)
        }
    })
}
```
Run with `go test -fuzz=FuzzNormalize`. Go shrinks failing inputs automatically and saves them in `testdata/fuzz/`. Prefer fuzzing over `testing/quick` when you need corpus management and shrinking.

### `httptest` for HTTP Dependencies
```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, `{"status":"ok"}`)
}))
defer srv.Close()
client := NewClient(srv.URL)
```

### `go test -race`
Always run `go test -race ./...` on concurrent code. The race detector catches seams of shared state that you didn't know were shared.

### Golden File Tests
For complex outputs (generated code, templates, reports), compare against a file:
```go
var update = flag.Bool("update", false, "update golden files")

func TestRenderer(t *testing.T) {
    got := renderer.Render(input)
    golden := filepath.Join("testdata", "expected_output.golden")
    if *update {
        os.WriteFile(golden, []byte(got), 0644)
    }
    want, _ := os.ReadFile(golden)
    if string(want) != got {
        t.Errorf("output mismatch; run with -update to refresh")
    }
}
```
Run with `go test -run TestRenderer -update` when intentionally changing output. Commit the updated golden files alongside the code change.

### Testing Error Paths with `errors.Is` / `errors.As`
Legacy code often returns errors as bare strings or concrete types. When writing characterization tests for error paths, use the standard checking idioms — don't compare error strings:

```go
// Fragile — breaks when message text changes:
if err.Error() != "connection refused" { t.Fatal(...) }

// Correct — checks the error type or sentinel:
if !errors.Is(err, ErrConnectionRefused) { t.Fatalf("want ErrConnectionRefused, got %v", err) }

// For wrapped errors (fmt.Errorf("%w", cause)):
var netErr *net.OpError
if !errors.As(err, &netErr) { t.Fatalf("want *net.OpError, got %T", err) }
```

When adding sentinel errors to legacy code to make error paths testable:
```go
var ErrNotFound = errors.New("not found")  // add to the package

func (s *Store) Get(id string) (Item, error) {
    // ... existing code ...
    return Item{}, fmt.Errorf("get %q: %w", id, ErrNotFound)  // wrap, don't replace
}
```

Wrapping with `%w` preserves the error chain so callers using `errors.Is` still work, and characterization tests can assert on the sentinel rather than the message text.

### Build Tags for Integration Tests
```go
//go:build integration
package mypackage
```
Run with: `go test -tags=integration ./...`

### `t.Parallel()` for Speed
Mark independent tests parallel. But be careful with global state seams — they become race conditions under parallelism.

### `t.Cleanup` for Safe Global Seams
```go
func TestWithGlobal(t *testing.T) {
    orig := globalVar
    globalVar = testValue
    t.Cleanup(func() { globalVar = orig })
    // ... test
}
```

### `t.Setenv` for Environment Variable Seams
`t.Setenv` sets an env var and automatically restores the original value when the test ends — no manual `t.Cleanup` needed:
```go
func TestFeatureFlag(t *testing.T) {
    t.Setenv("FEATURE_X_ENABLED", "true")
    // original value is restored after the test
}
```
Use this instead of `os.Setenv` + `t.Cleanup`. Note: `t.Setenv` calls `t.Helper` and will fail the test if called in a goroutine spawned by the test.

### `t.TempDir` for Filesystem Seams
`t.TempDir()` creates a temporary directory that is automatically removed when the test ends:
```go
func TestWriteFile(t *testing.T) {
    dir := t.TempDir()  // cleaned up automatically
    path := filepath.Join(dir, "output.txt")
    // write to path, test what was written
}
```
Use this instead of `os.MkdirTemp` + `t.Cleanup`. Avoids leaving test artifacts on disk when tests fail mid-way.

### Goroutine Leak Detection
Legacy code often starts goroutines that are never stopped — background workers, timers, event loops. These goroutines survive between tests and cause false passes, data races, and flaky failures.

Use `go.uber.org/goleak` to detect leaked goroutines:

```go
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)  // fails the suite if any goroutine leaks
}
```

Or per-test:
```go
func TestWorker(t *testing.T) {
    defer goleak.VerifyNone(t)
    w := NewWorker()
    w.Start()
    // ... test ...
    w.Stop()  // if Stop() doesn't drain the goroutine, goleak will catch it
}
```

The race detector (`go test -race`) catches data races. `goleak` catches goroutines that aren't being waited for. Use both.

### `TestMain` for Package-Level Setup
When all tests in a package need shared, expensive setup (compiled binary, docker container, database schema):

```go
func TestMain(m *testing.M) {
    // setup — runs once before all tests in the package
    db := setupTestDB()

    code := m.Run()  // run all tests

    // teardown — runs once after all tests
    db.Close()
    os.Exit(code)
}
```

`TestMain` is also the right place to call `goleak.VerifyTestMain`, set global test flags, or seed randomness. Keep setup minimal — the more shared state, the harder tests are to isolate.

### `go test -count=1` to Bypass Test Cache
Go caches test results. When you want to force a re-run (e.g., after changing a config file or external state):
```
go test -count=1 ./...
```
The cache is keyed on source files and flags, not on external state — so stale results are a real risk for integration tests.

---

## Summary Checklist

Before changing legacy Go code:

- [ ] Can I identify exactly which behavior is expected to change? (Identify change points)
- [ ] Have I written characterization tests for the code paths I'm touching?
- [ ] Does the code under change have a seam — an interface, function parameter, or io.Reader — I can inject a fake through?
- [ ] Am I changing existing behavior or adding new behavior? (Adding is safer)
- [ ] If I can't test the old code yet, can I sprout a new function and test that?
- [ ] Is the function I'm modifying small enough to understand? If not, sketch its structure first.
- [ ] Am I putting business logic inside `main()`, `init()`, `log.Fatal()`, or `os.Exit()`? Move it out — extract the body of `init()` to a named function that tests can call or skip.
- [ ] Are my new tests actually unit tests (fast, isolated) or integration tests? Label them correctly.
- [ ] Have I run `go test -race ./...`?
- [ ] Does the code start goroutines? Am I using `goleak` or `defer goleak.VerifyNone(t)` to catch leaks?
- [ ] Am I using `errors.Is` / `errors.As` to test error paths, not `err.Error()` string comparison?
- [ ] Am I using `t.TempDir()` and `t.Setenv()` instead of manual os calls + Cleanup?
- [ ] If the test cache might be stale, did I run `go test -count=1 ./...`?
- [ ] Is there duplication in the code I just wrote? Remove it — a design will emerge.
- [ ] Can I describe the responsibility of each struct/function I touched in one sentence?
- [ ] Did I leave at least one test, one renamed variable, and one extracted function behind as a campsite improvement?

When adding tests to legacy Go code, ask specifically:

- [ ] What does this function actually do? (Write a characterization test to find out)
- [ ] What is the narrowest seam where I can observe the behavior I care about?
- [ ] What is the minimum dependency-breaking change that lets me write a test?
- [ ] After I break the dependency, is there a scar I should note for future cleanup?

When editing without a safety net (breaking dependencies before tests exist):

- [ ] Am I preserving signatures — copying/pasting rather than retyping parameter lists?
- [ ] Am I doing one thing at a time, not refactoring and adding behavior simultaneously?
- [ ] Am I leaning on the compiler to find all the places that need updating?
- [ ] Have I added `var _ MyInterface = (*MyImpl)(nil)` to catch interface satisfaction breakage at compile time?
- [ ] Am I working with a pair partner? (Dependency-breaking is surgery; doctors don't operate alone.)

When adding a new feature to legacy code:

- [ ] Can I use Programming by Difference — create a new type alongside the old one, test it, then refactor?
- [ ] Did I choose Sprout Function/Type when I can't test the old code yet?
- [ ] Did I choose Wrap Function/Type when the new feature is as important as the old behavior?
- [ ] If I used Programming by Difference (temporary embedding to isolate new behavior), have I scheduled time to fold the variation back into the original type?
