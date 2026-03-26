# Testing Strategy Rules (Feathers)

Rules synthesized from Michael Feathers' "Testing Patience" talks (GeekFest and YOW! 2016).
The single organizing principle: **quality comes from deliberate thought, not from tests catching
bugs.** Every rule below follows from that.

---

## The meta-rule: the goal is never to write tests

Before writing any test, ask: what are you actually trying to do? Are you refactoring a specific
piece of code? Adding functionality here? The answer determines what tests you need — and
probably means you don't need to test everything, just enough to support the specific change you
are making with confidence. Engineering judgment, not mechanical coverage targets, is the guide.

When doing TDD, test failures are construction feedback, not bug detection. You write a failing
test first, then write code to satisfy it. That is not bug detection — that is construction. The
test failure is feedback as you move along a path toward working software, not a signal that
something went wrong. The quality you achieve through TDD does not come from the tests catching
errors. It comes from the concentration the process demands. Testing functions almost as a
*ritual* — a structured practice that forces deliberate thought. The tests themselves are almost
beside the point.

- **Do not** write tests to satisfy a coverage metric, demonstrate diligence, or fulfill a
  process ritual. Write them because a specific goal — quality, maintenance, or validation —
  requires them.
- **Do not** begin by asking "what tests should I write?" Begin by asking "what am I trying to
  accomplish?" The tests follow from the answer.

---

## 1. Know which goal a test serves before writing it

Testing serves three distinct purposes. Conflating them leads to the wrong tool for the job.

**Quality** — forcing deliberate thought before and during writing code. TDD, design by contract,
property-based testing, and clean room intended functions all work by this mechanism. The test
failures are almost beside the point; what matters is the concentrated thinking the process
demands. Any practice that forces you to specify intent before writing code tends to produce
similar quality results — because the mechanism is the same.

**Maintenance** — establishing behavioral invariants so code can be changed safely. Regression
suites, characterization tests, and golden masters all serve this goal.

**Validation** — confirming the software is acceptable to users. This is distinct from
*verification* (does the code meet its specification?). Verification requires tests that assert
against a formal spec. Validation asks whether users find the result acceptable — and for a large
and growing share of software, that question is better answered in production than in a test
suite. Progressive rollout and A/B testing are legitimate validation strategies, not shortcuts.

- **Do** identify which goal a test serves before writing it. A test for quality applies design
  pressure during construction. A test for maintenance preserves behavior across refactoring. A
  test for validation answers whether users accept what was built.
- **Do** choose the tool that matches the goal. If the domain has a hard formal specification
  (financial systems, medical devices, regulated software), write verification tests that assert
  against that spec. If the operative question is user acceptance, automated tests may not be the
  right tool — production observation may be.
- **Do not** assume that writing more tests serves all three goals equally. A dense regression
  suite serves maintenance. It does not substitute for deliberate design thinking (quality) or
  user feedback (validation).
- **Do not** treat the goal as "find bugs." The conventional model — tests catch bugs, caught bugs
  mean quality — is incomplete. Teams with rigorous deliberate thought and minimal test suites
  routinely ship higher quality than teams with broad coverage and shallow thinking.

---

## 2. Specify intent before writing code

The mechanism behind TDD, design by contract, and clean room programming is the same: writing
down what the code must do *before* writing the code forces precision that eliminates errors at
the source. The act of articulation is where quality is produced.

Stavely's clean room work revealed something: experienced programmers write the intended function
first, then write the code to satisfy it. That is test-driven development before TDD had a name.
The technique is different; the mechanism — deliberate thought preceding code — is identical.

Feathers found this directly in his own design-by-contract work: every time his preconditions or
postconditions grew massive and complicated, it was a signal — *"I'm doing something wrong. I need
to simplify my design so that my reasoning about it can be easier."* Complex specifications mean
complex code. The pressure toward simpler specifications is pressure toward simpler design.

- **Do** write a contract, specification, or test for a function before implementing it. The form
  can be a failing test (TDD), a structured doc comment stating preconditions and postconditions
  (design by contract), or a formal predicate comment in the clean room style. All three work by
  the same mechanism.
- **Do** use the following contract comment pattern before implementing any non-trivial function:

  ```
  // Requires: [what must be true when this function is called]
  // Ensures:  [what is guaranteed to be true when it returns]
  // Example:  swap(x=3, y=5) → x=5, y=3 (only when y > x)
  func swap(x, y *int) { ... }
  ```

  If the `Requires` clause needs more than one sentence, the function has too many entry
  conditions. Stop. Split the function before writing it. If writing the `Ensures` clause
  requires enumerating many different cases, the function is doing too much. Decompose it until
  each piece has a one-line postcondition.

- **Do** treat a test that is hard to write as a design signal — not a testing problem. Difficulty
  writing the specification means the code's responsibilities are not clearly defined. Clarify the
  design, then write the test.
- **Do not** write the implementation first and tests second as a confirmation ritual. That order
  discards the primary benefit: the thinking that precedes code is where quality is produced.
- **Do not** treat the test suite as the only acceptable form of this discipline. Code review
  against explicit intended functions and design by contract achieve the same quality mechanism
  without a test runner.

> **The complexity signal**: if specifying what a function does requires more than a few words —
> in a contract comment, a property, or a test description — the function is doing too much. The
> act of specification is a design tool. When the specification grows, simplify the design until
> the specification shrinks. This signal appears consistently across TDD, design by contract,
> property-based testing, and clean room intended functions. It is the same message every time.

---

## 3. Use tests as design pressure — redesign first

Tests that are painful to write are reporting a design problem. The correct response is to fix
the design, not to work around the test difficulty.

When you find yourself fighting a test — the setup is elaborate, collaborators need to be mocked
five levels deep, or internal state must be exposed to get coverage — stop. Redesign the interface
until the test is easy to write. Then write the test.

Mock-intensive isolation testing, done with the right intent, applies the same "deliberate
thought" mechanism as TDD. Steve Freeman's team tested each object in isolation against mocks,
with no integration testing, and achieved very low bug counts including zero integration bugs at
assembly time. The quality came from thinking carefully about every interaction before
implementing it — not from tests catching mistakes. The mock forced the specification of the
interface before the implementation existed.

- **Do** treat every test that requires elaborate setup or deep mocking as a refactoring trigger.
  The code under test has poor boundaries. Redesign before testing.
- **Do** use mocks at the interaction boundary — the point where one object hands off to another —
  to force explicit specification of that interface before implementing it. The mock is a design
  tool; it documents the contract the collaborator must satisfy.
- **Do** use property-based tests to apply stronger design pressure than example-based tests. If
  the properties of a function are hard to articulate or require many exceptions, the function is
  too complex. Decompose it until each piece has clean, independently verifiable properties.
- **Do** make failures immediately visible rather than hiding them. A function that silently
  returns a zero value on failure, or that catches errors and swallows them, removes the pressure
  that would otherwise force callers to think carefully about the failure case. Return errors
  explicitly. Let failures propagate to where they can be meaningfully handled. Hidden failures
  are hidden complexity.
- **Do not** work around testability problems by adding test-only seams, making fields package-
  visible, or restructuring production code solely to accommodate the test framework. Fix the
  design instead.
- **Do not** use mocks to achieve coverage of code that is already written. Mocking as coverage
  scaffolding — adding mocks to make existing untestable code testable — is a design-smell
  response, not a design-pressure response.
- **Do not** equate high test coverage with good design. A codebase can have 90% coverage and
  still be deeply coupled, hard to change, and poorly specified.
- **Do not** add catch-all error handlers at internal boundaries. Defensive catch-all handling
  hides failures from the callers who need to know about them.

> **Red flag — Hard-to-test code**: writing a test requires mocking five collaborators or
> accessing private state. The module's boundaries are wrong.
>
> **Red flag — Unmockable design**: a function cannot be called in a test without triggering
> database connections, network calls, or file I/O. The function is doing too much.
>
> **Red flag — Test fighting back**: every attempt to write the test reveals a new dependency
> that must be stubbed. Stop testing and redesign.
>
> **Red flag — Mocking implementation details**: the mock must know the internal sequence of
> calls inside the function under test, not just the interface it calls out through. The
> abstraction boundary is in the wrong place.
>
> **Red flag — Tests that get turned off**: a test suite that teams routinely skip or disable is
> reporting that the feedback cost exceeds the feedback value. The tests are too slow, too
> brittle, or covering things that don't warrant coverage. The correct response is not to enforce
> the discipline of running them — it is to fix the tests or the system they are testing.

---

## 4. Push side effects to the boundary, keep the core pure

Property-based testing and characterization testing both require functions that can be called
with arbitrary inputs and observed for their outputs. This is only possible when logic is
separated from side effects. If business logic is tangled with I/O, time, mutation, or external
state throughout, neither property tests nor characterization tests can reach it cleanly.

The pattern: partition the code into a **pure core** (deterministic functions with no side
effects — all inputs arrive as parameters, all outputs are return values) and an **imperative
shell** (where I/O, network calls, time, and mutation live). The core is testable with
properties. The shell is thin and testable with integration or acceptance tests. Rules 5 and 7
below both assume this separation exists. If it does not, create it before writing tests.

- **Do** identify the boundary between pure logic and effectful operations before writing any
  function. Logic that can be expressed as a pure function should be. Side effects belong at the
  edges of the system, not scattered through the middle.
- **Do** express domain computations as functions from input to output with no hidden state. A
  function that computes a discount given a price and a customer tier is pure. A function that
  reads the customer tier from a database and then computes a discount is not. Split them.
- **Do not** mix database reads, network calls, or time access into the same function as business
  logic. The mixing makes both the logic and the I/O harder to test.
- **Do not** treat this as an architectural constraint that requires a large upfront design
  decision. Apply it incrementally: when writing a new function, ask whether it could be pure.
  If it can be made pure by accepting its dependencies as parameters instead of fetching them
  internally, make it pure.

> **Red flag — Logic buried in I/O**: a function's test requires a database fixture or a network
> stub just to exercise an arithmetic or transformation operation. Extract the pure logic into a
> separate function, test that function directly, and test the I/O boundary separately.

---

## 5. Write property-based tests for pure functions

Property-based testing — articulating invariants that must hold for all valid inputs, then letting
a framework generate inputs to falsify them — is the natural testing style for pure and functional
code. It forces a more precise understanding of what a function actually guarantees.

In Go, use `testing/quick` from the standard library for simple cases or
`pgregory.net/rapid` for richer generators, shrinking, and stateful model testing.
The F# QuickCheck example from Feathers' talks translates directly: write the property as a
function that accepts generated inputs and returns a bool, then hand it to the framework.

**Common invariant categories and what they look like:**

- **Monotonic/ordered** — output elements satisfy an ordering relation for all adjacent pairs.
  A sort function: every element at index *i* is ≤ the element at index *i+1*.
- **Commutative** — reordering inputs produces the same result. A sort function: prepending the
  minimum value and sorting produces the same result as sorting and then prepending.
- **Idempotent** — applying the operation twice produces the same result as applying it once.
  A deduplication function: deduplicate(deduplicate(x)) == deduplicate(x).
- **Boundary-preserving** — aggregate properties of inputs are preserved in outputs. An interval
  merge function: the minimum and maximum values across all input intervals are preserved in the
  output; output count ≤ input count; at least one interval in means at least one interval out.
- **Round-trip** — encoding followed by decoding returns the original. Serialize then deserialize
  recovers the original value.

- **Do** identify which invariant category applies to the function before writing the property.
  Most functions exhibit one or two of the patterns above.
- **Do** look for reflexive, transitive, and symmetric properties in functions that relate values
  to each other. These algebraic structures characterize correctness more completely than any
  finite set of examples.
- **Do** treat difficulty articulating properties as a decomposition signal. If a property
  requires a long list of exceptions, the function is doing too much.
- **Do not** use property-based testing as a substitute for a clear function contract. Write the
  properties explicitly first; they are the contract.
- **Do not** expect to know which generated inputs will be tested. A framework like QuickCheck
  generates hundreds of inputs automatically; you will only see them if one causes a failure.
  That is a feature, not a problem.

---

## 6. Let the type system do what it can

A static type system is the set of tests the language designer could write without ever seeing
your program. In a statically typed language, some tests do not need to be written — the compiler
already checks those properties.

- **Do** rely on the compiler to enforce structural invariants. A value that cannot be constructed
  in an invalid state does not need a test asserting it is never constructed that way.
- **Do** model domain constraints in types rather than in validation tests where possible. A
  newtype wrapper with a validated constructor that can only hold values in `[1, 12]` eliminates
  the need to test that an integer is in range wherever it is used.
- **Do not** write tests that merely verify that the compiler's type system is working. A test
  asserting that a function returns a string when a string is expected adds no value in a
  statically typed language.
- **Do not** treat static typing as a silver bullet. It covers what the language designer could
  anticipate. It does not cover behavioral correctness, domain invariants the type system cannot
  express, or properties that emerge from component interactions.

---

## 7. Make time and external state injectable

Code that accesses `time.Now()`, random number generators, or other external state directly is
untestable because the test cannot control the inputs. The fix is always the same: inject the
dependency.

Time is tractable — it is monotone. If you turn it into data (a parameter, an interface, a
sequence of events), you can fabricate test inputs that behave identically to real time.
Asynchrony is harder: events arriving out of order or delayed cannot be handled by controlling a
clock. Async code that is difficult to test probably lacks explicit enumeration of its valid event
orderings. Making those orderings explicit in the design also makes them testable.

- **Do** inject time as a parameter or interface rather than calling `time.Now()` directly inside
  functions. A `clock Clock` interface with a `Now() time.Time` method makes time controllable
  in tests.
- **Do** inject random sources (`rand.Source`, `io.Reader` for crypto randomness) rather than
  accessing global generators. The function signature documents the dependency; the test
  controls it.
- **Do** inject I/O as interfaces (`io.Reader`, `io.Writer`, `fs.FS`) rather than accessing the
  filesystem or network directly. Functions that accept interfaces are testable without touching
  the real filesystem.
- **Do** treat async code that is hard to test as a signal that the valid event orderings are
  implicit in the code rather than explicit in the design. Enumerate them. Each ordering becomes
  a test case.
- **Do not** access `time.Now()`, `os.Getenv`, or `rand.Float64()` inside domain logic. Push
  those calls to the boundary (main, setup, or dependency wiring) and inject the results.

> **Red flag — Hidden time dependency**: a test that passes sometimes and fails others, or that
> requires `time.Sleep`, almost always contains a hidden call to `time.Now()` or equivalent.

---

## 8. Calibrate test investment to code lifetime and system type

The assumption that testing effort pays off over time is only valid when code lives long enough
to warrant the investment.

**Determine whether the system is tended or untended first.** This is the primary decision. If
someone can step in and roll back, push a patch, or intervene when something goes wrong, the
system is tended and the tolerance for discovering issues in production is higher. If the code
cannot be changed once deployed — medical device firmware, satellite software, systems with no
update path — the system is untended and must be verified rigorously before any release.

- **Do** assess whether the code is expected to be temporary before deciding how many tests to
  write. Code intended to be replaced in a month warrants less investment than code expected to
  run for five years.
- **Do** assess whether the system is **tended** or **untended** before making any other testing
  decision. The entire calculus depends on it.
- **Do** assess domain risk. Medical record systems and financial systems have hard verification
  requirements that cannot be traded away. A UI feature that controls icon placement on a social
  network does not.
- **Do not** apply a uniform testing standard across all code regardless of lifetime, criticality,
  or recoverability. The right level of investment is contextual.
- **Do not** maintain a test suite for code that no longer serves a purpose. When a feature is
  retired, delete the code and the tests that covered it.

> **Decision matrix — how much to test:**
>
> | System type | Code lifetime | Risk profile          | Suggested posture                           |
> |-------------|---------------|-----------------------|---------------------------------------------|
> | Untended    | Long          | High (health/finance) | Full verification before any release        |
> | Untended    | Long          | Low                   | Strong test coverage; no prod shortcuts     |
> | Untended    | Short         | Any                   | Write intended functions first (clean room); get it right before release; no regression suite |
> | Tended      | Long          | High                  | Strong coverage + production monitoring     |
> | Tended      | Long          | Low                   | TDD for design; regression for changes      |
> | Tended      | Short         | High                  | Strong coverage on core paths; skip UI      |
> | Tended      | Short         | Low                   | Deliberate design; minimal formal tests     |

---

## 9. Treat production rollout as a valid validation strategy — with conditions

For tended systems with appropriate risk profiles, progressive production rollout is a legitimate
substitute for some pre-production testing. The accountability condition is what makes it work:
the developer who ships the code must be the person who handles the consequences. That closed
feedback loop — the same person writes the code and sees the effects in production — changes
behavior in ways that a QA gate does not.

- **Do** use progressive rollout (deploy to tolerant users, observe, expand) when the system is
  tended, the domain risk is low, and the team is directly accountable for what ships.
- **Do** treat the slow-test problem honestly. If tests are slow enough that teams turn them off,
  those tests are not providing value. Either speed them up, cut them, or replace them with a
  different feedback mechanism. A test suite that teams routinely skip or disable has already
  been rejected as a feedback mechanism — acknowledge that and act on it.
- **Do** use QA specialists as consultants who ask hard questions *during development* — "Did you
  think about this case? What about that edge condition?" — rather than as a final gate who
  catches what developers missed. When developers internalize those questions, they stop injecting
  the bugs in the first place.
- **Do not** use production-as-test-environment to avoid accountability. "Eat your own dog food"
  means the person who wrote the code sees the consequences. That accountability is what makes
  the strategy work — and what separates it from carelessness.
- **Do not** apply this strategy to untended systems, regulated domains, or cases where failures
  cannot be reversed quickly.
- **Do not** treat "QA failed to catch it" as an explanation for a production defect. QA is a
  mirror of development quality. Developers put bugs in; QA reflects that back. It is never QA's
  fault.

---

## 10. Use characterization tests to establish behavioral baselines in legacy code

The paradox of legacy code: you want tests before refactoring, but the code is too coupled to
test without refactoring first. The resolution is to apply only the minimum restructuring needed
to open a test point, characterize current behavior, then change.

**Sequence for working safely in untested legacy code:**

1. Identify the specific change being made. Do not set out to test everything — test only what
   supports this change.
2. Apply dependency-breaking techniques (*Working Effectively with Legacy Code*, appendix, 24
   techniques) — minimal, visually inspectable refactorings that open test points without
   changing behavior. Apply just enough to get a seam around the thing being changed.
3. Write a characterization test: call the function with an arbitrary input, run it, take the
   actual output, put that output in the test as the expected value. The test now documents
   what the code currently does.
4. Make the targeted change.
5. Run the test. Any behavioral difference is now visible.

When dependency-breaking is insufficient — the function under change is too entangled to open a
seam without significant restructuring — use the **sprout method**: write the new behavior as a
new, standalone function with clean boundaries (testable in isolation, with properties if it is
pure), characterize the existing function separately with a golden master or characterization
test, then wire the new function into the existing call site. This avoids changing the hard-to-
test code while still keeping the new code properly tested and specified.

- **Do** treat code released into production as its own specification. Bugs users have worked
  around have become features they depend on. Characterization tests preserve those de-facto
  contracts across refactoring.
- **Do** consider the **golden master** pattern for large codebases: run the code against known
  data sets, capture output, compare future runs against that baseline. It is a valid stepping-
  stone. At large scale its failures are hard to diagnose (you didn't create the inputs); at
  targeted scale it is practical. Expect to see more of it in the industry.
- **Do not** require characterization tests to assert correct behavior — they assert *current*
  behavior. The question is not "is this right?" but "does this still do what it did before?"

---

## 11. Manage code lifetime actively: delete and rewrite

Entropy accumulates in software. The active responses are deletion (when features are retired)
and rewriting (when modules become too crufty to safely change). Both require the same
precondition: boundaries clean enough that one piece can be removed or replaced without
cascading through adjacent pieces.

**The dead code axiom**: every line of code in a repository is implicitly a claim that it is in
active use. Dead code makes that claim falsely. It consumes maintenance attention, confuses
readers, inflates apparent system complexity, and hides the true carrying cost of the system.
Delete it. Do not leave stub implementations, commented-out code, or unreachable branches.
If the code was valuable, recover it from version control or rewrite it — the rewrite will be
better because the context is now clearer.

- **Do** delete code when a feature is retired. Deciding to retire a feature without deleting its
  code is half the job.
- **Do** have the conversation about carrying cost. When a stakeholder requests a new feature,
  surface any existing feature whose presence makes the new one significantly harder to add. "How
  much revenue does that feature actually produce?" is a legitimate engineering question and often
  one that has never been asked.
- **Do** keep modules small enough that rewriting one is a credible option when it becomes too
  crufty to refactor cleanly. When a module is easier to replace than to fix, replace it.
- **Do** give explicit standing permission for targeted rewrites: if the module is too hard to
  refactor, the functionality is replaceable, and the interface is clear — rewrite it without
  requiring approval. Entropy is a thing. Systemic rewrites at bounded scale are how you stay
  ahead of it.
- **Do not** let dead code accumulate on the grounds that it might be useful later. Recover it
  from version control or rewrite it — the rewrite will be better.
- **Do not** invest heavily in tests for code that is expected to be thrown away. The test
  investment assumes the code will live long enough to justify it.
- **Do not** pursue large-scale rewrites that must simultaneously match every feature of the
  existing system. That doubles the work and commonly fails. Rewrite at bounded, module-sized
  scale.
