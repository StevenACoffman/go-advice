# Testing Patience

**Michael Feathers**
R7K Research & Conveyance

---

## Reference: Lessons Learned in Software Testing

*Cem Kaner, James Bach, Bret Pettichord — A Context-Driven Approach*

---

## Test Types Taxonomy

```
Test Types
├── What Tests (What Gets Tested)
│   ├── Object Under Test (OUT)-Based Test Types
│   └── Domain-Based Test Types
├── When Tests (When Testing Occurs)
│   ├── Order-Based Test Types
│   ├── Lifecycle-Based Test Types
│   ├── Phase-Based Test Types
│   └── Built-In-Test (BIT) Types
├── Why Tests (Why Testing Occurs)
│   ├── Driver-Based Test Types
│   └── Reason-Based Test Types
├── Who Tests (Who Does Testing)
│   ├── Collaboration-Based Test Types
│   ├── Organization-Based Test Types
│   └── Role-Based Test Types
├── Where Tests (Why Testing Occurs)
│   ├── Organization-Location-Based Test Types
│   └── Physical-Location-Based Test Types
├── How Tests (How Testing is Performed)
│   ├── Automation-Based Test Types
│   ├── Level-of-Scripting-Based Test Types
│   └── Technique-Based Test Types
└── How Well Tests (Quality Verified)
    └── Quality-Characteristic-Based Test Types
```

---

## The Goal - Fewer Errors in Production

*Is that all?*

---

## Goal - Quality

### The Flawed Theory Behind Unit Testing

*June 12, 2008*

I have Google's blogsearch set to give me notifications about unit testing. On an average week, I read dozens of blogs and mailing list discussions about the topic. Occasionally, I read something new, but there's lot of repetition out there. The same arguments crop up often. Of all of them, though, there is one argument about unit testing which really bugs me because it rests on a flawed theory about testing and quality and sadly, it's an argument that I fell for a long time ago and I'd like to lay it to rest. Hopefully, this blog will help, but I have to relate a little history first.

Back in the very early 2000s, I had a conversation with Steve Freeman at a conference. We were talking about Test-Driven Development and Steve had the strong feeling that most of the people who were practicing TDD at the time were doing it wrong — they'd missed something.

Steve was and is part of a close-knit community in London who have been practicing XP and TDD from the very beginning. Among the fruits of their labor was the entire notion of mock objects. Steve Freeman and Tim MacKinnon wrote the paper that introduced the idea to the broader community. The rest is history. There are mock object frameworks out there for nearly every language in common use.

> *Quality is a function of precise thought and reflection*

---

### Reference: Toward Zero-Defect Programming

*Allan M. Stavely*

> "A NASA satellite-control system. This was 40,000 lines of FORTRAN. The bug density in testing was 4.5 per thousand lines. In the first three years of production use, only seven minor errors were found, which is less than 0.1 per thousand lines, and each of these was easy to fix."
>
> — Allan M. Stavely (*Toward Zero Defect Programming*)

---

### Intended Functions and Formal Specification

```
[y > x -> x, y := y, x | true -> l]
if (y > x)
  [x, y := y, x]
  {
      int temp;
      temp = x;
      x = y;
      y = temp;
  }
```

> "When are the intended functions written? They can be written after the code is written. If you did this, you would use the intended function for each construct to document what you had in mind when you wrote it .. Sometimes you might need to look at the construct itself, derive the function that it actually computes and use that as the 'intended' function."
>
> *"Experienced programmers would probably do things in the opposite order most of the time."*
>
> — Allan M. Stavely

**TDD?**

---

## Quality Comes From Deliberate Thought

*We can introduce constraints to force thought*

---

## Property-Based Testing

> "Writing tests first forces you to think about the problem you're solving. Writing property-based tests forces you to think way harder."
>
> — Jessica Kerr (@jessitron), 1:52 AM - 26 Apr 2013

### Example: F# Property-Based Test

```fsharp
let ``append minValue then sort should be same as sort then prepend minValue`` sort
Fn aList =
    let minValue = Int32.MinValue

    let appendThenSort = (aList @ [minValue]) |> sortFn
    let sortThenPrepend = minValue :: (aList |> sortFn)
    appendThenSort = sortThenPrepend

// test
Check.Quick (``append minValue then sort should be same as sort then prepend minVal
ue`` goodSort)
// Ok, passed 100 tests.
```

*(F# For Fun and Profit)*

### The Core Shift

```
examples  →  ∀
```

Property-based testing moves from specific examples to universal quantification over all possible inputs.

---

## Tactic 2: Introducing Hazard to Make Systems Robust

### Why making streets risky improves road safety

Mixing cars, cyclists and pedestrians may appear more dangerous, but 'Shared Space' approaches are reducing accidents across Europe.

A study of seven European Shared Space projects from 2004 to 2008, backed by the European Interreg IIIB North Sea Programme, supports Hamilton-Baillie's view. "There are fewer accidents," it concludes. "When a situation feels unsafe, people are more alert."

Unexpectedly, the study also found that despite motorists driving more slowly in Shared Space zones traffic delays were reduced — by up to 50% at the Laweiplein, which sees 22,000 vehicles pass through each day.

Traffic lights, road marking and pedestrian crossings have also been removed in Ashford in England — which has seen "a 60% drop in accidents in the first three years", says Hamilton-Baillie.

---

## Putting it into Production

### Continuous Delivery Pipeline

```
Delivery team  →  Version control  →  Build & unit tests  →  Automated acceptance tests  →  User acceptance tests  →  Release

Check in  →  Trigger  →  [FAIL: red]
            Feedback  ←

Check in  →  Trigger  →  [PASS: green]  →  Trigger  →  [FAIL: red]
            Feedback  ←                   Feedback  ←

Check in  →  Trigger  →  [PASS: green]  →  Trigger  →  [PASS: green]  →  Approval  →  [UAT]  →  Approval  →  [Release]
            Feedback  ←                                Feedback  ←
                                                       Feedback  ←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←←
```

**Question: What happens when the automated acceptance test stage is missing or broken?**

*(The middle stages — Build & unit tests and Automated acceptance tests — become a question mark if not reliably maintained.)*

---

## Tended Systems v. Untended Systems

---

### Programmer Anarchy

*"Developer driven development"* — Fred George, 2010

**Reduction in Roles:** Eliminates Project Manager, Business Analyst, QA/Test roles — only Customer and Developer remain.

**Missing Agile Practices (struck through = dropped):**
- Agile Manifesto & XP Values
- ~~Standups~~ Trust with co location
- ~~Story narratives~~
- ~~Retrospectives~~
- ~~Estimates~~
- ~~Iterations~~ Results, not blame
- ~~Mandatory~~ pairing
- ~~Unit tests, acceptance tests~~
- ~~Refactoring~~
- ~~Patterns~~ Small, short lived apps
- ~~Continuous integration~~ Continuous deployment

**Positives (+):** Lack of managers, Lightweight (beyond Agile), Works well for some companies

**Negatives (-):** Limited reference material, Requires highly skilled / disciplined team, No testing or planning

*If you set the context, you can get rid of many constraints*

---

### Moral Hazard and QA

> **Moral hazard** is a situation in which one party gets involved in a risky event knowing that it is protected against the risk and the other party will incur the cost.

*..having a QA Department*

---

## Death/Repurposing of QA

The QA role is being handed off — not eliminated, but transformed. As development practices mature, the responsibilities once held by a separate QA department transfer directly to developers.

---

## Goal - Maintenance

### The Golden Master Approach

Before making any change to the production code, do the following:

1. Create X number of random inputs, always using the same random seed, so you can generate always the same set over and over again. You will probably want a few thousand random inputs.
2. Bombard the class or system under test with these random inputs.
3. Capture the outputs for each individual random input.

*(from Sandro Mancuso — https://dzone.com/articles/testing-legacy-code-golden)*

---

## Goal - Validation

### Verification and Validation

The PMBOK guide, a standard adopted by IEEE, defines them as follows in its 4th edition:

- **Validation.** The assurance that a product, service, or system meets the needs of the customer and other identified stakeholders. It often involves acceptance and suitability with external customers. Contrast with *verification.*
- **Verification.** The evaluation of whether or not a product, service, or system complies with a regulation, requirement, specification, or imposed condition. It is often an internal process. Contrast with *validation.*

**Much development is moving toward a post-verification world**

---

### Testing as a thinking tool

> Testing as a thinking tool — property based behavioral invariant verification

---

## The Three Goals of Testing

| Mechanism | Goal |
|---|---|
| *Force Thinking* | Quality |
| *Behavioral Invariants* | Maintenance |
| *Acceptance* | Validation |

*Moving toward Transience*

---

## Transience in Practice

As software systems become more transient — short-lived services, microservices, disposable components:

- **Rewrite services when you want to**
- Behavioral invariants (property-based tests) ensure correctness across rewrites
- Acceptance tests validate business behavior regardless of implementation

### Example: Interval Merging (Ruby)

```ruby
intervals
  .sort
  .map { |interval| [interval.first, intervals.reject { |other| (other <=> interval) == 1 }.flatten.max] }
  .slice_when {|c,n| c.last < n.first }
  .map {|group| group.flatten.minmax }
```

---

## The Goal - Less Badness

The ultimate goal is not perfection — it is less badness. Testing serves multiple purposes:

1. **Quality** — Force deliberate thought through constraints (TDD, property-based tests)
2. **Maintenance** — Behavioral invariants allow confident refactoring and rewriting
3. **Validation** — Acceptance tests confirm you built the right thing

As systems become more transient, the emphasis shifts from verification (does it match the spec?) toward validation (does it serve the user?) and behavioral invariants (does it preserve correct properties through change?).

---

## Contact

michael.feathers@r7krecon.com
r7krecon.com
