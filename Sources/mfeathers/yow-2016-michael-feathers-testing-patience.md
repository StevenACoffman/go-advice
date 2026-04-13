# Testing Patience: Where Testing Has Been and Where It's Going

*Michael Feathers — YOW! 2016*

---

Testing keeps evolving in ways that are hard to get an overview of. Every few years a new type of testing arrives and gets absorbed into the process. Unit testing had a renaissance inside early Agile. That evolved into TDD. Then acceptance testing. Then continuous delivery pipelines. And every time you turn around, someone has a new testing technique they want you to adopt.

This talk is an attempt to step back and ask what testing is actually *for* — and to look at where the industry appears to be heading over the next five to ten years.

## The Taxonomy Problem

Years ago, Feathers was invited to a workshop with some of the most respected figures in the software testing community: Cem Kaner, James Bach, and Brett Pettichord — authors of *Lessons Learned in Software Testing*. Developers were brought in to show the testers some of the newer practices: refactoring, TDD, things like that. At some point, someone asked the experts to draw the taxonomy of software testing on a whiteboard — black box, white box, integration, unit, how do these all relate?

The result was the most spectacular Venn diagram imaginable, covered in dotted lines, because there are so many overlapping ways to classify these things and people use the terminology differently everywhere. The terminology wars are real and, in the end, probably not the most productive use of energy. The more important question is: what is testing actually *for*?

## The Traditional Answer: Fewer Errors in Production

The conventional answer is quality — fewer errors in production. Code that runs to completion, doesn't cause terrible side effects, behaves according to user intent. That's a reasonable goal.

But the mechanism we assume drives that — the idea that *tests catch bugs, and caught bugs mean quality* — turns out to be a flawed theory, or at least an incomplete one.

## The Flawed Theory of Testing

Steve Freeman told Feathers about a project his team had done in the context of extreme programming. They had a codebase with unit tests, and they radically restructured it — reshaping what the software did, retargeting it to a new domain — reusing extensively all the code they had. When Feathers asked whether they had acceptance testing in addition to unit tests, Freeman said no. Just unit tests. And yet they were able to recompose the code for the new domain rapidly and arrive in production with very few bugs.

Feathers was utterly amazed by this and ended up writing a blog post about what he was ascertaining from it.

This is the crack in the conventional model. If tests catch bugs, and fewer bugs mean quality, why did a team with minimal testing achieve high quality? The answer is this: the testing process itself, independent of whether any test *fails*, forces you to think very carefully about what you are building. That concentrated thought is what produces quality — not the catching of errors.

Testing, embedded inside development, is almost a trick. It's a trick to make you think more clearly about the problem. The test failures are not the point.

## Clean Room Programming: Quality Without Any Tests

This line of thinking leads to a fascinating historical example. Alan Stavely's book *Toward Zero Defect Programming* documents a movement from the 1980s and 1990s called clean room programming — not the modern usage of "clean room" which refers to hiring a separate group to create a knockoff of a product, but a rigorous discipline for producing high-quality software by eliminating testing entirely as a backstop.

The process worked like this: no tests were written. Forward progress required passing code review. And that code review was rigorous in a particular way — for every piece of code, the developer wrote a *predicate* (called an *intended function*) that specified the preconditions, postconditions, and semantic intent of what the code was supposed to do. These predicates were embedded directly in the code as comments, making the reasoning process visible to reviewers.

There was also some statistical/stochastic testing to verify certain properties held, but the driving mechanism was the review of those predicates, not test suites.

On a 40,000-line Fortran system, this process produced a bug density of 4.5 per thousand lines — very high quality. The reason it never became widespread is that it's an extremely labor-intensive process; getting teams into a room to review predicate specifications for every piece of code is hard to sustain at scale.

Feathers found a direct parallel in his own experience with design by contract — writing preconditions and postconditions for every significant piece of code. Every time his preconditions or postconditions grew massive and complicated, it was a signal: *I'm doing something wrong. I need to simplify my design so that my reasoning about it can be easier.* Complex predicates mean complex code. The process of articulating intended functions, like the process of articulating contracts, puts pressure toward simplicity.

But the pattern it reveals is the same one seen in TDD. From the book:

> *Experienced programmers would probably do things in the opposite order most of the time. They would probably write the intended function first and then write the code.*

Write your intention. Then write the code to satisfy it. This principle was in circulation before TDD had a name. The technique was different; the underlying mechanism — deliberate thought preceding code — was the same.

**Quality comes from deliberate thought.** Constraints that force us to think deeply about what we're producing are what generate quality. Testing is one way of introducing those constraints. It is not the only way.

## Property-Based Testing: The Next Evolution

As the industry moves further into functional programming, a new form of this same idea is gaining traction: property-based testing. It's rough when dealing with mutable code, but with functional code it's often quite a bit easier.

Traditional testing means thinking of individual cases and encoding them as examples. Property-based testing means thinking about *what is universally true* about the behavior you're trying to produce — the mathematical properties that should hold for any valid input — and then letting a framework generate randomized inputs to try to falsify those properties. A framework like QuickCheck can generate a hundred tests for you automatically. Crucially, you often don't know what those generated tests are unless one actually fails.

A sorting routine, for example, from an F# piece Feathers cites: one property is that after sorting every element should be monotonically ascending. Another is that appending a minimum value to a list and then sorting should produce the same result as prepending that minimum value — a form of algebraic commutativity. John Hughes, who created QuickCheck (the first major property-based testing tool, from the Haskell community), has shown it applies far beyond value correctness to performance characteristics, race conditions, and more.

The skill required is algebraic: you're looking for invariants — reflexive, transitive, symmetric properties — that characterize the behavior of the thing you're building. This is harder than writing examples, but it forces a more precise understanding. And that precision is exactly the mechanism that produces quality.

If you find yourself trying to articulate properties that require a huge number of exceptions and edge cases, that's a signal: your design is too complicated. Simplify it until the properties become clean. The constraints of property-based testing, like the constraints of TDD, apply design pressure.

Property-based testing represents a significant direction for the industry — a natural next step in the evolution from example-based thinking toward constraint-based thinking.

## Introducing Hazard to Make Systems Robust

There's a counterintuitive idea worth examining here.

Researchers in a German town discovered that they could *reduce* traffic accidents by making roads more ambiguous — removing lane separators, blurring the boundary between road and sidewalk, eliminating the clear demarcation that tells drivers exactly where they are. The result: drivers became more careful because the environment felt more dangerous.

At the same time, there's a run-on effect in automotive safety in the US: everything done to make cars safer has decreased fatalities, but people are actually more reckless driving now than they were maybe 34 years ago — because now it's safer, people can afford to be. Remove the net, and behavior changes.

Feathers chose C as his first programming language — he wanted to do something that felt very hard, to prove to himself he could be a programmer. He taught himself C, then changed his major to computer science. In his second programming course, sitting in a lab next to a classmate, he looked over her shoulder and saw an array subscript out of bounds exception on her screen. She was using Pascal. He'd never seen that exception before. He realized why: in C, an out-of-bounds array access sends pointers dancing over memory and causes unpredictable failures. Learning in that environment builds a native defense against it — fast. Pascal caught the error cleanly. C taught the lesson viscerally.

A certain amount of hazard makes people more careful. This applies to software development. It is ultimately the care that we place in our work that produces quality — not the tests themselves.

## Production as a Test Environment

This brings up a more provocative idea: just putting things into production.

There's a familiar scenario in continuous delivery: the tests are slow. What do teams actually do in response? They turn them off. Nobody wants to be the person who has to say "we did have a test for that, but we decided not to run it" — especially not when there's a separate QA department whose entire function is to act as a guardrail. So slow tests become a cultural liability: too politically expensive to remove, too slow to run faithfully.

Everything in software development is contextual; there is no universal right way. But for certain domains and risk profiles, continuous delivery can be taken further than conventional pipelines assume. The approach: write code, hand-check it, inspect it, run it through a handful of example cases, put it into version control, release immediately. If something goes wrong, something goes wrong. The progressive rollout variant — deploy to a small group of tolerant users first, observe, expand — formalizes this. The question is whether you *need* all the test infrastructure in the pipeline for every type of change.

The answer depends heavily on what kind of system you're building. A medical record system or a financial system has clear requirements to test rigorously before anything ships. A Facebook feature that affects which icon appears next to a status message has a different risk profile entirely. The domain determines what's appropriate.

Feathers acknowledges the tension: this is rough for him to talk about, because for years he's been the person running around telling teams to get more tests in place around legacy code. He still believes that. Those tests are often vital for enabling safe change. The point isn't to abandon testing — it's to recognize that production rollout is a legitimate option for certain constrained circumstances.

What enables this is the shift from **untended** to **tended** systems. Early in his career, Feathers worked at a medical device company. The conversations were: *if something's wrong with this firmware, we can't fix it in the field.* The system was untended — once deployed, it was done. More and more systems today are tended: if something goes wrong, someone can step in, roll back, push a patch. The entire cloud ecosystem has changed what's possible. Even space probes near Pluto are now firmware-patchable, even though the round-trip for the update takes hours at the speed of light.

As systems become more tended, the case for treating production as a feedback mechanism — rather than a place tests protect you from — gets stronger.

## Programmer Anarchy

Fred George (who had spoken at the same conference the day before) explored an extreme version of this. The business was online arbitrage: spot inventory at a low price, put up a page to test demand, buy and resell. The code that drove these operations was transient — it might be live for a month, generate revenue, then get thrown away when the opportunity closed.

In that context: no formal testing required. Any programming language. No managers. Developers made the product decisions. They were also early experimenters with what we now call microservices — small, independently deployable pieces whose short expected lifetimes made disciplined refactoring and test investment hard to justify.

This won't work for 90% of organizations. It required a very specific set of enabling conditions. But the lesson it demonstrates is that context shapes what practices make sense. You have to identify whether you're actually in that situation before adopting those practices.

Feathers draws a broader analogy here: if you look at things like eventual consistency and the breakdown of transactionality across distributed systems, the industry has already accepted that certain hard problems can be sidestepped in some domains and not others. The loosening of testing requirements follows the same logic — and will probably move further in that direction over time.

## Moral Hazard

The economic concept of moral hazard — when removing consequences from risk causes people to take more of it — applies directly to software development.

A separate QA department can function as moral hazard. If developers know QA will catch what they miss, there's an unconscious license to be less careful. The accountability loop is broken: the person writing the code is not the person who discovers whether it works in production.

QA has genuine value for performance testing, exploratory testing, and other specialized work. But for functional correctness, closing that feedback loop — making developers directly accountable for what ships — produces better code. It's the old saying: eat your own dog food. If you're the person ultimately responsible for what shipped, you're going to be far more careful about what you do. It's not a comfortable conversation to have, but the evidence of it is visible in teams that have adopted that model.

## The Three Goals of Testing

Pulling back, testing serves three distinct purposes:

**Quality** — forcing deliberate thought before and during writing code. The mechanism is constraint: TDD, design by contract, property-based testing, clean room intended functions. Any of these works. The test failures are almost beside the point. Quality comes from having thought carefully.

**Maintenance** — preserving behavioral invariance so that code can be changed safely. This is what TDD's regression suite does: it tells you immediately if a refactoring broke something. This is also what the *golden master* approach does, imperfectly. Golden master — running code against known data sets, capturing the output, and comparing future runs against that baseline — was dismissed early in Agile as a cheap substitute for real tests. But it's a valid strategy for establishing behavioral baselines on large codebases, particularly as a steppingstone toward something better. Expect to see more of it.

**Validation** — confirming that what you've built is acceptable to the people using it. This is different from verification (does the code meet the specification?) and is increasingly the operative question for most software. In regulated domains — banking, medical devices — you still need hard verification against formal specifications. But a large and growing portion of software development is in domains where the only real question is: do users find this acceptable? Production rollout is a validation strategy. A/B testing is a validation strategy. These are legitimate.

Testing supports all three of these goals. So do other practices. Knowing which goal you're serving helps you choose the right approach.

## Transience: Code That Was Never Meant to Last

There's a deep assumption embedded in most testing practice: that the effort invested in writing tests will pay off over time because the code will live for a long time. This assumption isn't always true.

This line of thinking started as a beer conversation. Feathers and a friend were talking and posed the question: what if every line of code you wrote disappeared three months after you wrote it? If something was valuable, you'd rewrite it — and the second or third version would be better. And with code constantly disappearing, you'd have a hard limit on feature count, which would force better prioritization. The industry builds too many features and carries too much code. We don't have good data on the carrying cost of a feature — the personnel, production support, maintenance burden — but if we did, we'd make different trade-offs.

The connection to testing is this: if code is transient — if it's going to be thrown away in a month anyway — the calculus around how much testing investment it warrants changes. Fred George's arbitrage business operated on that logic. The code was genuinely meant to be temporary.

There's a companion difficulty here that technology people encounter regularly: someone asks for a new feature, and you immediately see that an existing feature is in the way — if that other thing weren't there, adding the new one would be so much easier. But broaching that conversation is hard. "This would be easier if you didn't have that." "You want us to get rid of it? That's revenue." How much revenue? Is it really worth it? Having that conversation — about the carrying cost of features and the code behind them — is something the industry should get better at.

### Conditioning Users to Change

There's a related idea. When Feathers was young, he could go to a store, see a shirt he liked, decide to come back next week, and expect it to still be there. Retailers have since trained customers out of that expectation — scarcity and flux drive purchase decisions. The item might not be here next week.

Software is doing the same thing. Twitter has reshaped its user interface repeatedly without asking permission. Most large-scale applications make continuous small changes to features and UI. Users are increasingly conditioned to expect flux rather than stability.

This matters because it changes what *behavior preservation* means. If users already expect that features change, the requirement to preserve behavior across code changes is weaker — at least for the UI layer and certain feature areas. That changes the ROI calculation on the tests that enforce it.

The hard part is connecting the feature lifecycle to the code lifecycle. Business people can decide a feature is retired; the discipline to also delete the code that implemented it is much rarer. Dead code is a lie — the default assumption when reading a codebase is that everything visible is in use, and dead code corrupts that assumption. Getting better at knowing what to delete, and actually deleting it, is a meaningful form of software quality.

### Rewrites

Rewrites have a terrible reputation, mostly earned by large-scale failed rewrites where a team had to match every feature of an existing system while also keeping the old system running. The double work is brutal.

But at smaller scales, rewrites are powerful. Feathers isn't going to preach microservices — he thinks it's just a technology — but to the degree that code can be made more modular, targeted rewrites become conceivable in a way they aren't for monoliths. Chad Fowler, while at Wunderlist (later acquired by Microsoft), gave his developers a standing permission: if a microservice became too crufty, too hard to refactor, and they felt confident they could replace its functionality — just rewrite the damn thing. Don't even tell me. We know that entropy is a thing in software development. If you can do systemic rewrites for bounded pieces, you're better off than you were.

The industry has thirty-year-old code that nobody feels they can get rid of. The trajectory is toward more modularity, more willingness to rewrite bounded pieces, and more discipline about actually retiring code when features are retired.

## Where This Is Heading

The trend lines:

- **Property-based testing** will grow, driven by the move toward functional programming. Thinking in constraints rather than examples is a more rigorous and ultimately more powerful way to reason about software.
- **Tended systems** will become the norm. As more infrastructure moves to the cloud and more systems get continuous deployment capability, the case for production-as-feedback-environment gets stronger.
- **Transience** will be taken more seriously, both in terms of retiring features and their code, and in terms of calibrating test investment to expected code lifetime.
- **Unit testing** remains valuable and widely applicable — a technique for many circumstances.

The overall goal is less badness: quality, maintenance, and validation. Testing is one avenue for all three. It's not the only one. Knowing what goal you're pursuing at any given moment lets you choose the right tool — and sometimes choose not to test at all.
