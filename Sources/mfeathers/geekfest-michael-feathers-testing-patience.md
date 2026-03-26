# Testing Patience: Rethinking What Tests Are For

*Michael Feathers — GeekFest (Groupon Berlin)*

---

Testing is one of those topics every developer talks about, but rarely examines deeply. We do it. We advocate for it. We measure it with coverage metrics. But what is testing actually *for*?

That question is harder than it looks — in part because the terminology is a mess. Black box testing, white box testing, user testing, unit testing: everyone uses different words and means different things. Feathers recounts once asking a group of expert testers — Cem Kaner, James Bach, and Brett Pettichord, the authors of *Lessons Learned in Software Testing* — to draw a Venn diagram of all the testing types. The flip chart that resulted was chaotic. There are genuine taxonomies of testing in the literature, and they're all different. Once you cut through that confusion and ask what testing is actually *for*, you start to see places where you can vary your approach and get advantages you hadn't considered.

## The Dominant Goal: Fewer Errors in Production

Most teams come to testing with a single goal: fewer production errors. That's a reasonable starting point. Errors in production are costly — they interrupt the work you actually want to be doing, and they erode trust.

But the path to that goal is less obvious than it first appears.

## Testing as a Ritual, Not a Detector

Here's a counterintuitive observation: when you're doing TDD, test failures are not primarily about *finding bugs*. You write a failing test first, then write code to satisfy it. That's not bug detection — that's construction. The test failure is feedback as you move along a path toward working software, not a signal that something went wrong.

What this means is that the quality you're achieving through TDD doesn't come from the tests catching errors. It comes from the concentration the process demands. You think carefully about what each object must do, how it interacts with others, what edge cases matter. Because you've thought through those cases explicitly, you simply write code that isn't wrong.

Testing functions almost as a *ritual* — a structured practice that tricks us into quality by forcing deliberate thought. The tests themselves are almost beside the point.

Steve Freeman's team in London demonstrated this vividly. They were doing mock-intensive TDD — testing each object in isolation against stubs and mocks, not testing three or four objects working together. No automated acceptance testing either. And yet they had a remarkably low bug count, with no integration bugs when components were assembled. The quality came from the discipline of thinking carefully about every interaction, not from the tests catching mistakes.

## Clean Room Programming: Quality Without Tests

The history of software engineering offers a more extreme example. In the *clean room software design* process described by Alan Stavely, teams achieved very high quality with *no unit testing at all*.

The method: every developer wrote code and simultaneously wrote a formal comment — called an **intended function** — specifying exactly what that piece of code was supposed to do. Teams then gathered for intensive code review sessions, using those intentions as the basis for verification. A simple swap operation, for instance, would be headed by a formal predicate: "if Y is greater than X, apply this function" — and then the code below it.

The book notes something revealing about how experienced programmers approached this: they would typically write the intended function *first*, then write the code to satisfy it. Sound familiar? This is test-driven development before TDD had a name.

Applied to a 40,000-line Fortran flight control and satellite system for NASA, the project achieved a bug density of 4.5 per thousand lines — already a very low number — and then the clean room process brought it down to near-zero from there. No unit testing. Just rigorous, deliberate specification of intent before and during coding.

**Mob programming** reflects a similar principle. It's essentially pair programming on steroids: a whole team in a room with a projector, all coding together on the same thing at once. It looks like a huge waste of effort. Some companies that do it are very happy with the results.

## Design Pressure

This pattern — writing the spec before the code — consistently applies *design pressure* that simplifies software. When Feathers worked on medical device software using **design by contract** (Bertrand Meyer's formalization of preconditions and postconditions), he found that the discipline didn't just improve correctness. It forced simplification. If your preconditions had multiple complex clauses, that was a signal to break the function into smaller pieces. The specification became a design tool.

TDD works the same way. Writing a test before code that's hard to test is feedback that the design is too complex. The test isn't just verifying behavior — it's revealing structure.

### Property-Based Testing

**Property-based testing** takes this a step further. Instead of specifying individual cases, you articulate mathematical properties that should hold for any valid input — then a framework generates random inputs and tries to find a violation.

For a sorting routine, for example: if you prepend an element smaller than all others in the list, it should still be the first element after sorting. For an interval-coalescing function (where `[0,3]` and `[1,2]` merge to `[0,3]`, and `[4,5]` and `[5,6]` merge to `[4,6]`), you might specify: the output count is always ≤ the input count; if there's at least one interval in, there's at least one out; the minimum and maximum values across all intervals are preserved.

Writing these properties requires thinking in terms of *what is always true* rather than *what happens in this case*. Like design-by-contract and TDD, this applies design pressure — if your properties are complex and hard to articulate, that's usually a sign the function can be decomposed into smaller pieces with simpler, independently verifiable properties.

## Introducing Hazard

There's a counterintuitive principle at work in all of this: a small amount of difficulty or ambiguity can make you more careful.

Researchers discovered — in Germany or England, Feathers couldn't recall which — that they could reduce accidents in a small town by making the boundaries between pedestrian areas and streets *ambiguous*. Drivers approaching the town center suddenly weren't sure if they were on a street or not. They slowed down. The hazard induced care.

The parallel in software: learning C as a first language, where an out-of-bounds array access causes serious undefined behavior, creates programmers who never go out of bounds. The hazard builds good habits. Learning in Pascal, with runtime bounds checking catching every mistake, may produce programmers who are less careful — because the safety net is always there.

GPS navigation is another example: the part of your mind that processes GPS directions is the same part needed to remember a route. Use GPS everywhere, and your internal map degrades.

This doesn't mean static typing is bad. Static type systems are genuinely powerful — they represent, as one way of putting it, the set of tests the language designer could think up without ever seeing your program. In a statically typed language, you may not need to write certain tests at all because the compiler already checks those properties.

But the pendulum swings. Industries cycle between dynamic and static typing, each time believing they've found the answer. As Brooks put it, there's no silver bullet. Static typing isn't a silver bullet; neither is dynamic. The more honest view is that both have contexts where they apply well.

## Putting Things Into Production

Another tactic for quality: treat production itself as a testing environment — for some things.

Some teams skip heavy pre-production testing for small changes. They deploy to a small group of tolerant users, observe, then expand rollout gradually. For non-critical code in attended systems (where someone can respond if something goes wrong), this can be entirely reasonable.

This view is contested. Feathers recounts a recent Twitter exchange where someone called him a "horrible person" for advocating it, arguing that you're forcing work onto prod support teams and making users pay for developer mistakes. His response: aren't we one team? It's a choice an organization can make. For some things, and with some risk profiles, it's the right choice.

This is fundamentally about choosing the right feedback loop. The cost of a feedback loop matters. If you can get high-quality signal quickly by deploying and watching, that may be more efficient than maintaining an elaborate pre-production test suite.

The key qualifier is **tended vs. untended systems**. If someone can step in and fix problems, the tolerance for discovering issues in production is higher. If you're shipping firmware to a spacecraft near Pluto (where a firmware update takes hours at light speed, even if it's technically possible), or code that controls an elevator, that calculus changes entirely.

## Moral Hazard

Testing can also create problems. The economic concept of **moral hazard** applies: when you remove consequences from risk, people take more risks. Safety features in cars — airbags, better crumple zones — are associated with people driving more carelessly than they did in earlier decades.

The software equivalent: a separate QA department can create a culture where developers write sloppy code because QA will catch it. When QA doesn't catch a bug, the response becomes "QA failed us" rather than "we put a bug in." Feathers recalls telling a CTO: QA is a big mirror on development. It's never their fault. Developers put the bugs in.

This isn't an argument against QA — it's an argument for integrating QA thinking into development from the start. QA specialists are most valuable as consultants who ask hard questions early: *Did you think about this case? What about that edge condition?* When developers internalize those questions, they stop injecting the bugs in the first place.

## Three Goals for Testing

Looking across all these approaches, testing serves three distinct purposes:

### 1. Quality
Testing forces deliberate thought. The value isn't in the tests catching bugs — it's in the thinking the tests require. This is why any practice that demands you specify intent before writing code (TDD, design-by-contract, clean room intended functions) tends to produce similar quality results.

### 2. Maintenance
Tests establish **behavioral invariants** — a record of what the code actually does. This is particularly valuable when working with legacy code.

When approaching unfamiliar legacy code, a useful technique is **characterization testing**: rather than reasoning about what the code should do, you ask it what it *does*. Write a test with a placeholder expected value — say, the empty string. Run it. It fails. Take the actual output and put it in the test. Now the test passes — and you've documented current behavior.

This matters because code released into production becomes its own specification. Bugs that users have worked around become features they depend on. Fix one and you'll hear about it. Characterization tests let you refactor without accidentally breaking those de-facto contracts.

The **golden master** pattern is related — run your code, capture output, compare future runs against it, typically with random inputs at scale. It's often written off as a cheap, lazy alternative to real unit tests. It does work. The downside at large scale is that failures are hard to diagnose because you didn't create the inputs. Applied at a smaller, more targeted scale, it's a practical way to establish behavioral baselines.

### 3. Validation
There's a distinction between **verification** (does the code meet its specification?) and **validation** (is this acceptable to the user?). Many systems have moved almost entirely toward validation.

For domains with hard specifications — financial systems, biomedical software — verification remains essential. But for much of what we build today, the question is just: does the user want this? Testing in production via incremental rollout is a validation-centric activity. These are legitimate engineering choices, not shortcuts.

## Transience and Rewrite

There's one more underappreciated option: throwing code away.

Chad Fowler — who worked in Berlin for a period — ran a team with an explicit discipline: if a microservice gets too complicated, throw it away and rewrite it. The risk of being held captive by your own code is real. Feathers has visited companies in China with 30-year-old applications built on antiquated architecture that nobody wants to work on but nobody can replace either, because they've grown too large, too muddled, and too interconnected. Keeping things small enough to be disposable is one way to avoid that fate.

Fred George's "programmer anarchy" experiments demonstrated a more extreme version. His team was doing online arbitrage: spot a supplier with 1,000 birdcages, put up a webpage to test demand, buy the lot, resell elsewhere. The business model itself was transient — so the software was too. No managers (developers found and evaluated the deals themselves), no refactoring (because the code wouldn't live long enough to need it), any programming language, minimal testing. They were calling these microservices before that term had wide currency.

This doesn't generalize. But it illustrates that many constraints we treat as universal — refactoring discipline, test coverage, architectural consistency — are actually a response to specific contextual forces, chiefly that code is expected to live for a long time. When it isn't, different rules apply.

The same principle shows up in how software products now handle user interfaces. Retailers discovered they could drive sales by making inventory feel scarce — if you don't buy this shirt today, it might not be here tomorrow. Software products increasingly treat their own features the same way: Twitter has overhauled its UI many times; almost every major app changes things in subtle ways on a continuous basis. Behavior preservation is less important than it once seemed, at least in those areas. Factor that in when choosing how much to invest in tests that lock down UI behavior.

## Summary

Testing is not primarily about finding bugs. It's a tool — or a set of tools — for different purposes:

- **Quality** comes from deliberate thought. Any practice that forces you to specify intent before writing code tends to produce quality. Testing is one such practice, not the only one.
- **Maintenance** is about establishing behavioral invariants. Characterization tests and golden masters serve this even when they don't verify against a specification.
- **Validation** is about whether users accept what you've built. Some of the most effective validation happens in production.

The right testing approach is contextual. It depends on whether the system is tended or untended, how critical the code is, how transient the codebase is intended to be, and what feedback loops are available. Testing is a powerful tool — but it's one of many, and knowing what you actually want from it makes it far more useful.

---

## Q&A Highlights

**On static typing as a form of testing:**
The type system is the set of tests the language designer could think up without ever seeing your program. In a statically typed language, some tests don't need to be written — the compiler already checks those properties. But static and dynamic typing both have legitimate uses. Much of the current enthusiasm for static typing everywhere is a reaction to bad experiences with JavaScript, not a settled truth about language design. That reaction is understandable but probably overcorrects.

**On legacy code and where to start:**
The paradox of legacy code is that you want tests before refactoring, but the code is too coupled to test without refactoring first. The answer is dependency-breaking techniques — *Working Effectively with Legacy Code* covers 24 of them in the appendix. These are minimal, visually inspectable refactorings designed to open test points without changing behavior. You apply just enough of them to get tests around the thing you actually want to change, then refactor freely underneath. Writing those tests is easier than it sounds: you're just asking the code what it does, not verifying it against a spec. Run it, capture the output, make that the expectation.

**On reactive programming and time:**
Time in reactive systems is tractable — it's monotone, and if you can turn it into data (a sequence of events), you can fabricate test inputs that behave the same as real time. What's actually harder is *asynchrony*: events arriving out of order, events being delayed. Those require carefully enumerating the relevant orderings rather than just controlling a clock. How much support your framework gives you for abstracting time determines how hard the testing problem really is in practice. (An audience member mentioned a forthcoming book on Rx Java testing with a chapter specifically on this.)

**On what test to write first in a legacy codebase:**
The goal is never to write tests. The goal is to solve a specific problem. Before writing any test, ask: what am I actually trying to do? Am I refactoring this particular piece? Adding functionality here? The answer determines what tests you need — and probably means you don't need to test everything, just enough to support the specific change you're making with confidence. Engineering judgment, not mechanical coverage, is the right guide.
