# Code Design Rules (Ousterhout)

Rules derived from *A Philosophy of Software Design* by John Ousterhout.
The single organizing principle: **complexity is the enemy**. Every rule below
is a tactic for reducing it.

---

## 1. Make modules deep

A module's interface should be dramatically simpler than its implementation.
The best modules hide enormous complexity behind a small, clean surface.

- **Do** implement complex logic (buffering, encoding, format parsing, retries,
  locking) inside the module and expose none of it to callers.
- **Do** make the common case require the fewest possible calls and the least
  prior knowledge. Rare/advanced operations can be more verbose.
- **Do not** split a simple operation across multiple classes that callers must
  compose themselves (e.g., requiring three objects just to read a file).
- **Do not** create a class or function whose interface is nearly as complex as
  its body. A wrapper that does nothing except translate one call into an
  identical call adds interface cost with zero benefit.

> **Red flag — Shallow module**: the interface for a class or function is not
> much simpler than its implementation.

---

## 2. Hide information; do not leak it

Every module should fully own at least one design decision. That decision must
not appear in the module's interface or be duplicated in another module.

- **Do** encapsulate file formats, wire protocols, encoding schemes, data
  structures, and concurrency strategies inside the module that owns them.
- **Do** merge two classes when they share knowledge of the same design
  decision (e.g., a reader and a writer that both understand the same file
  format).
- **Do not** structure code around the *order* of operations (read → modify →
  write) when information hiding argues for a different decomposition. Choose
  decomposition by knowledge, not by time.
- **Do not** expose internal representation through getters/setters. A
  `getMap()` that returns the internal map is information leakage even though
  the field is private.
- **Do** provide defaults automatically ("do the right thing without being
  asked"). Callers should be unaware of options they do not need.

> **Red flag — Information leakage**: the same design decision is reflected in
> multiple modules. Changing it requires touching more than one place.
>
> **Red flag — Temporal decomposition**: code structure mirrors the time-order
> of operations rather than cohesion of knowledge.
>
> **Red flag — Overexposure**: the API for a commonly used feature forces
> callers to learn about rarely needed features.

---

## 3. Design general-purpose interfaces

A somewhat general-purpose API tends to be simpler and deeper than a
special-purpose one, even when only one use exists today.

- **Do** ask: "What is the simplest interface that covers all current needs?"
  Fewer methods with broader semantics beat many narrow ones.
- **Do** ask: "Is this method only useful for one particular situation?" If
  yes, it is probably too special-purpose.
- **Do not** mirror every UI operation with a matching back-end method
  (`backspace()`, `deleteSelection()`, …). One general `delete(start, end)`
  replaces them all.
- **Do not** create "false abstractions" — a method that purports to hide
  information but whose callers must understand the hidden details anyway just
  to call it correctly.
- **Do** keep general-purpose mechanism in lower layers; put code that
  specializes the mechanism for a particular application in upper layers.

> **Red flag — Special-general mixture**: a general-purpose mechanism contains
> code specialized for a particular use of that mechanism.

---

## 4. Keep each layer's abstraction distinct

Adjacent layers in a system must provide different abstractions. Same
abstraction in two adjacent layers is a signal that the decomposition is wrong.

- **Do not** write pass-through methods — methods whose entire body is a
  single forwarding call with the same or nearly the same signature.
- **Do** eliminate pass-through methods by: (a) exposing the lower layer
  directly, (b) redistributing responsibilities, or (c) merging the layers.
- **Do not** use the Decorator pattern unless the decorator adds substantial
  functionality. Decorators tend to be shallow and create an explosion of
  boilerplate.
- **Do not** pass a variable through a long chain of methods that do not use
  it. Use a context object stored in instance variables instead.

> **Red flag — Pass-through method**: a method does almost nothing except pass
> its arguments to another method with a similar signature.

---

## 5. Pull complexity downward

When a module and its callers share responsibility for handling complexity,
push it into the module. More callers exist than developers; it is better for
the module to bear the pain once than to make every caller deal with it.

- **Do** handle edge cases, retry logic, format normalization, and default
  selection inside the module.
- **Do not** expose configuration parameters for things the module can
  determine on its own. Before exporting a parameter ask: "Will callers
  actually know the right value better than I do here?"
- **Do** compute sensible defaults automatically; let callers override only in
  exceptional circumstances.
- **Do not** throw exceptions for conditions the caller almost certainly cannot
  handle; handle them internally or define them away.

---

## 6. Define errors and special cases out of existence

Exception handling is one of the worst sources of complexity. The best
approach is to redesign so the error condition cannot occur.

- **Do** redefine semantics so the error disappears: instead of "delete
  variable (error if absent)", define it as "ensure variable does not exist
  (always succeeds)".
- **Do** mask exceptions at low levels when higher levels have nothing useful
  to do with them (e.g., TCP masking packet loss through retransmission).
- **Do** aggregate exceptions: let them propagate to a single high-level
  handler rather than writing distinct handlers for each caller.
- **Do** crash on errors that are truly unrecoverable (out of memory,
  internal inconsistency) rather than littering the code with checks that
  cannot improve the outcome.
- **Do not** throw exceptions as a substitute for thinking about what clean
  handling would look like. If you cannot figure out what to do, callers
  probably cannot either.
- **Do** eliminate special-case boolean flags and null sentinels by designing
  the normal case to subsume the special case (e.g., an empty selection whose
  start == end, rather than a "selection exists" flag).

> **Red flag — Unnecessary special cases**: code that is riddled with `if`
> statements checking for conditions that could be designed away.

---

## 7. Split and join methods based on abstraction, not length

The right reason to split a method is that it produces a cleaner abstraction
— a child that is general-purpose and understood without reading the parent.
The wrong reason is that it is "too long".

- **Do** split when a subtask is independently reusable and can be fully
  understood without knowing the parent's context.
- **Do** split when a method has an overly complex interface because it does
  multiple unrelated things; each new method should have a simpler interface.
- **Do not** split just to enforce an arbitrary line count. A long method with
  a simple signature and clear flow is fine — it is deep.
- **Do** join two shallow methods into one deeper method when doing so
  simplifies the interface, eliminates an intermediate data structure, removes
  duplication, or enables better encapsulation.

> **Red flag — Conjoined methods**: you cannot understand one method without
> also reading the other.

---

## 8. Eliminate duplication

Repeated code means the right abstraction has not been found yet.

- **Do** factor repeated code into a shared method when the method can have a
  clean, simple signature.
- **Do** refactor so that a decision or computation exists in exactly one
  place.
- **Do not** copy-paste logic and adjust it slightly; this is where bugs hide
  and where maintenance costs compound.

> **Red flag — Repetition**: the same or nearly the same code appears in
> multiple places.

---

## 9. Choose precise, consistent names

Names are documentation. The best name creates an accurate image in the
reader's mind with no ambiguity.

- **Do** use names that are specific enough that someone reading a call site
  without documentation can correctly guess the meaning: `getActiveIndexlets()`
  not `getCount()`.
- **Do** use boolean names that are predicates: `cursorVisible` not
  `blinkStatus`.
- **Do not** reuse the same name for different things within the same scope or
  module. Use distinguishing prefixes when necessary (`srcFileBlock` vs.
  `dstFileBlock`).
- **Do** use a consistent name for a given concept everywhere it appears; do
  not invent synonyms.
- **Do** interpret difficulty naming something as a design smell: if you
  cannot find a precise, short name it often means the entity has unclear
  purpose or is doing too many things.
- **Do** match declaration types to allocation types; do not declare `List`
  and allocate `ArrayList` if the concrete type matters to readers.

> **Red flag — Vague name**: the name is broad enough to refer to many things
> and conveys little information.
>
> **Red flag — Hard to pick a name**: difficulty finding a precise, intuitive
> name for an entity suggests the entity may not have a clear definition.

---

## 10. Write comments that add information the code cannot express

Comments exist to capture design decisions, constraints, invariants, and
intent that cannot be inferred from reading the code.

- **Do** comment every class (overall abstraction and limitations), every
  non-trivial variable (units, invariants, null semantics, ownership), and
  every public method (behavior, arguments, return value, side effects,
  exceptions, preconditions).
- **Do** write interface comments *before* writing the body. The act of
  writing the comment is a design check: if the comment is long and tangled,
  the interface is probably wrong.
- **Do not** repeat the code in a comment. If someone could reconstruct the
  comment by reading the adjacent code, the comment adds nothing.
- **Do not** use the same words in the comment that appear in the name.
- **Do** write at the right level of abstraction: interface comments describe
  *what* and *why* at a high level; implementation comments explain *why* a
  non-obvious approach was chosen, not *what* the code is doing.
- **Do** place comments close to the code they describe. Farther away =
  higher abstraction level required to stay accurate through code churn.
- **Do not** put design rationale only in commit messages; put it in the code
  where developers will actually see it when making changes.
- **Do** document cross-module design decisions in a single authoritative
  location and reference that location from the affected code.

> **Red flag — Comment repeats code**: all information in the comment is
> immediately obvious from the adjacent code.
>
> **Red flag — Hard to describe**: the documentation for a variable or method
> must be very long to be complete, indicating a design problem.

---

## 11. Make code obvious

Obscurity is a direct cause of complexity. Code is obvious when a reader's
first guess about its behavior is correct.

- **Do** use white space to separate major phases of a method; pair each
  block with a short comment describing its purpose.
- **Do** define specialized structs/types for values returned together instead
  of using generic pairs or tuples whose element names convey nothing.
- **Do not** violate common reader expectations without documenting the
  deviation prominently (e.g., a constructor that spawns background threads).
- **Do** document event handlers and callbacks with a comment stating exactly
  when they are invoked, since the control flow cannot be traced statically.
- **Do** match declaration and allocation types when the concrete type matters.

> **Red flag — Nonobvious code**: the behavior or meaning of a piece of code
> cannot be understood quickly.

---

## 12. Be consistent

Consistency creates cognitive leverage: learn the pattern once, apply it
everywhere.

- **Do** follow naming, style, and structural conventions already present in
  the codebase. "When in Rome…"
- **Do not** introduce a variant spelling, casing, or pattern for your
  personal preference. The value of consistency outweighs the value of any
  one convention being slightly better.
- **Do** use automated checks (linters, pre-commit hooks) to enforce
  conventions, not just human vigilance.
- **Do not** use the same name or pattern for dissimilar things; consistency
  depends on readers trusting that similar appearance means similar behavior.

---

## 13. Invest in the design; do not cut corners

Tactical shortcuts ("just get it working") compound into systems that are
expensive to change and understand.

- **Do** take the time to find the simplest design even if it slows the
  current feature by 10-20%. The payback comes within months.
- **Do** consider at least two radically different designs for any non-trivial
  module before committing to one.
- **Do** improve the design whenever you touch code. After a change, the
  system should look as if it had been designed with that change in mind from
  the start.
- **Do not** let complexity accumulate. A single hack matters little; the
  accumulation of hundreds matters enormously.
- **Do** think of good comments and clean interfaces as design work, not
  busywork. They reveal design problems early when they are cheapest to fix.

---

## 14. Design for performance without adding complexity

Simplicity and performance are mostly aligned: simple code has fewer layers
to cross, fewer conditions to check, and better cache behavior.

- **Do** choose naturally efficient algorithms and data structures as a first
  choice when they are as simple as slower alternatives.
- **Do not** optimize prematurely. Measure first; programmers' intuitions
  about bottlenecks are unreliable.
- **Do** identify the critical path and minimize the work on it. Remove
  special-case checks from the hot path; handle edge cases with a single
  early test that branches away from the fast path.
- **Do** discard performance changes that do not produce a measurable
  improvement; they add complexity for no gain.

---

## Summary of Red Flags

| Red Flag | Symptom |
|---|---|
| Shallow Module | Interface not much simpler than implementation |
| Information Leakage | Same design decision in multiple modules |
| Temporal Decomposition | Code structure mirrors execution order, not knowledge |
| Overexposure | Common API forces learning rare features |
| Pass-Through Method | Method only forwards to another with same signature |
| Repetition | Same nontrivial code appears multiple times |
| Special-General Mixture | General mechanism contains special-purpose code |
| Conjoined Methods | Can't understand one method without reading the other |
| Comment Repeats Code | Comment says only what code already shows |
| Implementation Contaminates Interface | Interface doc describes internals |
| Vague Name | Name too broad to convey meaning |
| Hard to Pick Name | Difficulty naming reveals unclear purpose |
| Hard to Describe | Complete doc requires excessive length |
| Nonobvious Code | Behavior/meaning not clear from quick reading |
