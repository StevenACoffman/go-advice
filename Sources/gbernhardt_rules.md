# Functional Core / Imperative Shell Rules (Bernhardt)

Rules distilled from Gary Bernhardt's "Functional Core, Imperative Shell" (Destroy All Software,
episode DAS-0072). The single organizing principle: **the amount of code that mixes computation
with state change determines how hard the system is to test and reason about.** Push computation
into an immutable core; confine all mutation and side effects to a thin outer shell.

---

## 1. Make the default return a new value, not a mutation

When writing a function that updates a data structure, return a new value rather than modifying
the receiver. A function that takes a `Timeline` and some new tweets should return a new
`Timeline` — not append to an internal slice. The caller holds the old value; it receives the new
one and decides what to do with it.

The only mutation that is acceptable inside a core object is the rebinding of a local variable
within a function body. That is invisible from outside and does not affect immutability at the
boundary other code crosses.

- **Do** write domain functions as transformations: `func (t Timeline) WithNewTweets(tweets []Tweet) Timeline`.
- **Do** return new structs rather than mutating pointer receivers in domain logic.
- **Do** use local variable rebinding (`result = compute(result)`) freely inside a function body —
  it is invisible from outside and does not break immutability.
- **Do not** add methods to domain types that mutate any field on the receiver.
- **Do not** mutate a slice or map that was passed in as a parameter — return a new one.

---

## 2. Separate pure logic from side effects before writing either

Before implementing a function that combines computation with IO, network calls, or database
access, split it into two pieces: a pure function that does the computation, and a thin
wrapper (or caller) that handles the IO. Write the pure piece first.

A function that computes which restaurants are top-rated is pure: it takes ratings, returns an
ordered list. A function that reads ratings from the database and returns a list is not — but it
can call the pure one. Keep them separate so each can be changed and tested independently.

- **Do** identify whether a function requires external state (IO, time, network, database) before
  writing it. If it does not, write it as a pure function with no external calls.
- **Do** accept dependencies as parameters rather than fetching them internally. A function that
  accepts a `[]Rating` is testable; a function that calls `db.Query(...)` inside is not.
- **Do** split a function that computes a result from one that reads inputs, even when they are
  always called together. The split has essentially no cost and makes both pieces testable.
- **Do not** put a database call, network request, or `time.Now()` inside a function that also
  contains business logic. Move the IO to the caller and pass the result in.
- **Do not** design a domain function so that testing it requires a database fixture or network
  stub. If it does, the function is not pure — split it until it is.

> **Red flag — logic buried in I/O**: a test for a domain function requires setting up a database,
> starting a server, or mocking a network client. The function is doing two jobs. Extract the pure
> computation and test that directly; test the I/O wrapper separately.

---

## 3. Confine mutable state to a thin shell at the program boundary

The shell is the narrow zone where values change: it receives new values from the core, assigns
them to variables, and coordinates with the outside world. It does not contain logic. The pattern
is: call a pure function → receive a new value → assign the new value to the mutable variable.

In the Twitter client, pressing `j` calls `cursor.MoveDown()`, receives a new `Cursor`, and
assigns it to `@cursor`. The assignment is in the shell; the logic of what "down" means is in the
core. The shell never knows what "down" means — it only knows it received a new cursor.

Mutable references (the variables that track current state) live in the shell. The values they
point to are immutable. The shell updates the reference; the core produces the value.

- **Do** keep a clear mental boundary between code that computes (core) and code that changes
  state or calls external systems (shell). When writing a function, know which zone it is in.
- **Do** write shell code as flat sequences of assignments and IO calls with minimal branching.
  If the shell code develops complex conditionals, that is a signal to extract logic into the core.
- **Do** keep the shell as the sole point of contact with each external resource. If one goroutine
  or one file owns all writes to the database, all writes to stdout, and all reads from stdin,
  those resources are serialized for free — no locks required.
- **Do not** put computed logic (sorting, filtering, ranking, formatting) in the shell. Pass values
  to a pure function and use the result.
- **Do not** update state from multiple places in the shell. One mutable reference, one assignment
  site.

> **Red flag — fat shell**: the shell file has grown past ~200 lines and contains conditionals,
> loops with complex logic, or functions that return non-trivial computed values. The shell has
> absorbed logic that belongs in the core. Extract it.

---

## 4. Immutable values make concurrency coordination structural

An immutable value can cross a thread (or goroutine) boundary without any locking. There is
nothing to race on: neither thread can change the value, so it does not matter which one reads it
or when.

The pattern: a background goroutine constructs a new value (a complete `Timeline` built from
fetched tweets) and sends it over a channel to the main goroutine. The main goroutine receives
the value, writes it to the database, and updates the mutable reference. The background goroutine
never holds a mutable reference; the main goroutine is the sole writer to all IO. Both properties
hold by structure, not by locking.

- **Do** pass immutable values between goroutines via channels without synchronization. The
  immutability makes the channel the only coordination point needed.
- **Do** designate a single goroutine (typically the main goroutine or a dedicated coordinator)
  as the sole owner of each external resource (database, stdout, stdin). Let other goroutines
  produce values and send them; the owner goroutine does the writing.
- **Do** use `errgroup` or a fan-in channel pattern when multiple goroutines need to compute
  values in parallel — collect immutable results, then process them in the owner goroutine.
- **Do not** share mutable state across goroutines. If two goroutines need to update the same
  thing, one of them is in the wrong place. Restructure so only one goroutine owns that resource.
- **Do not** add a mutex to a shared data structure as the first response to a concurrency
  problem. Ask whether the data structure can be immutable and passed by value instead.

---

## 5. Core functions need no test doubles — use real values

Pure functions take values in and return values out. There are no collaborators to mock, no
shared state to reset between tests, no sequence of calls to verify. Tests have a single shape:
construct input values, call the function, assert on the output value.

A test for a `cursor.SliceFrom(tweet)` function constructs a cursor with two tweets and a tweet
value, calls the function, and asserts the returned slice. One line. No setup, no teardown, no
mocking framework. If the same test required mutation, it would need three steps: construct,
mutate, inspect state. Immutability reduces the test to one.

When mocks or stubs seem necessary, it is a signal that the code under test is not pure. Make it
pure — accept the dependency as a parameter rather than fetching it — and the test doubles
disappear.

- **Do** test core functions with real input values and assert on real output values. Do not reach
  for a mocking library.
- **Do** write test cases as single expressions when possible:
  `assert.Equal(t, expected, cursor.SliceFrom(tweet1))`.
- **Do** use table-driven tests when a function's behavior varies across input ranges. Each row is
  a pure input/output pair.
- **Do not** introduce a mock or stub for a pure function's dependency. If the function needs
  external data, it is not pure yet — make it accept the data as a parameter.
- **Do not** share state between test cases for pure functions. Each case should be independently
  constructable from its inputs.

> **Red flag — mocks in core tests**: a test for a domain function requires a mock database,
> mock HTTP client, or mock clock. The function has a hidden dependency on external state. Extract
> a pure inner function that accepts the data as a parameter; test that instead.

---

## 6. Test the shell only when it becomes complex enough to be afraid of

The shell is untested when it is a flat sequence of assignments and IO calls with few
conditionals. The judgment to not test the shell is contingent on that simplicity — it is not a
blanket rule. When the shell develops branches, accumulates logic, or reaches a size where
silent misbehavior becomes plausible, it needs tests.

Bernhardt's shell is ~160 lines with no tests. The key-dispatch block is the one place with
enough branching to worry about, and it is exercised every time the application runs. He says
directly: "if this became more complex, I would become afraid, and I would want tests." The
decision is calibrated to current complexity, not to a principle about shells.

- **Do** assess the shell's complexity honestly before deciding whether to test it. Count the
  conditionals. If the shell is flat, direct, and exercised constantly, tests add little.
- **Do** add shell tests when a conditional branch is complex, infrequently exercised, or not
  easy to verify through use. The threshold is fear: test what you are afraid of.
- **Do not** apply "the shell is untested" as an absolute rule. It is a judgment calibrated to
  simplicity. A complex shell requires tests; failing to test it is a liability, not a pattern.
- **Do not** let the shell absorb logic on the grounds that it will not be tested anyway. Logic
  in the shell that has not been extracted to the core is technical debt, not an architectural
  feature.

---

## 7. Use the shell as an exploration zone; extract when the design is clear

The shell is where new libraries get discovered and application structure gets figured out when
the design is unknown. Write rough code in the shell — exploratory, untested, possibly messy —
to understand the problem. Once a design emerges that you have confidence in, extract that
behavior into a unit-tested functional class via TDD, and delete the shell code it replaces.

The sequence matters: the new functional class is written and fully tested before it is wired
into the system. The commit that adds the class changes nothing else. The subsequent commit
does the swap — wires the new class in and removes the old shell code — producing a large
deletion. The shell shrinks; the core grows.

TDD applied during extraction produces slightly different interfaces than direct transcription of
existing code, even when the behavior is already known. That design pressure is intentional: the
act of writing tests for the extracted class reveals cleaner boundaries.

- **Do** write exploratory code in the shell without tests when the design is genuinely unknown.
  Uncertainty about design is a legitimate reason to defer tests.
- **Do** stop and extract once you know what the code is doing and have a design you trust.
  Exploration is a temporary state; accumulated shell code that has been figured out is debt.
- **Do** write the extracted functional class via TDD, with full test coverage, before wiring it
  into the system. Do not add tests to the old shell code — extract and replace.
- **Do** delete all the shell code that the new functional class replaces. The swap commit should
  show a large deletion. If it does not, the extraction is incomplete.
- **Do not** accumulate explored-but-not-extracted code in the shell indefinitely. The rhythm is:
  bloat → extract via TDD → delete → shrink. The shell is not a permanent home for any logic.
- **Do not** test-drive the shell during exploration. The design is unknown; TDD assumes enough
  clarity to specify behavior before writing it. Explore first, specify later.

---

## 8. Ask "where is the mutation happening?" and localize it

The pattern does not require a clean upfront architectural decision. It can be applied
incrementally to any codebase. The starting question is always: where is the mutation happening?
Then localize it.

Every function that does not need to mutate state can be made pure. Every pure function can be
extracted from wherever it lives and tested directly. Every test that becomes easier after
extraction is evidence of a better boundary. The shell shrinks one extraction at a time.

- **Do** ask of every new function: can this be pure? If the function's behavior depends only on
  its parameters, make it accept those parameters explicitly and return a value.
- **Do** treat test difficulty as a locating device. A test that is hard to write because of
  mutable state is pointing at a boundary that should move. Move the boundary; the test becomes
  easy.
- **Do** apply the pattern incrementally: one pure extraction at a time, each extraction followed
  by deletion of the mutable code it replaces.
- **Do not** require the entire program to be restructured before beginning. Extract the first
  function that could be pure, test it, and delete the impure version. Repeat.
