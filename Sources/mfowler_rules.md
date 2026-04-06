# Refactoring Rules (Fowler)

Based on Martin Fowler's *Refactoring: Improving the Design of Existing Code*, 2nd edition (JavaScript examples).

---

## 1. Core Philosophy

**Refactoring is disciplined restructuring that preserves observable behavior while improving internal structure.** It is not the same as rewriting, fixing bugs, or adding features.

- **Keep the two hats separate.** When adding a feature, do not restructure existing code — only add new capabilities. Measure progress by adding tests and getting them to pass. When refactoring, do not add new behavior — only restructure. Do not add tests unless you find a case you missed earlier; only change tests when an interface changes. Switch hats deliberately and frequently, but never wear both at once.
- **Make the change easy, then make the easy change.** Before adding a feature, refactor so that the feature slots in naturally. As Jessica Kerr put it: "It's like I want to go 100 miles east but instead of just traipsing through the woods, I'm going to drive 20 miles north to the highway and then go 100 miles east at three times the speed." This is almost always faster than forcing the feature into ugly code.
- **Small steps, always.** Each refactoring step must leave the code in a working state. If the code is broken for more than a few minutes, you are not refactoring — you are restructuring. Commit after each successful small step so you can revert cheaply. The rhythm is: make a small change, compile, test, commit.
- **The purpose is economic, not aesthetic.** "The whole purpose of refactoring is to make us program faster, producing more value with less effort." Never justify refactoring on cleanliness or purity; justify it on development velocity. The most dangerous trap is trying to justify refactoring in terms of "clean code" or "good engineering practice" — the economic argument is the only argument managers will actually respond to.
- **Design Stamina Hypothesis.** Good internal design lets you add features quickly indefinitely. Bad internal design lets you sprint at first, then slows to a crawl as every new feature requires more and more time to understand. The code base starts looking like a series of patches covering patches. Continuous refactoring is how you maintain design stamina.
- **Code speaks to two audiences.** The compiler does not care about naming or structure, but the next programmer (often yourself) does. "Any fool can write code that a computer can understand. Good programmers write code that humans can understand."
- **Put understanding into the code.** As Ward Cunningham says: whenever you acquire an insight about what code does, move that understanding from your head back into the code immediately. Rename, extract, restructure so the insight is preserved for everyone who follows. Code that communicates its understanding requires less future archaeology.
- **The true test of good code is how easy it is to change.** Not aesthetics, not line count, not "purity" — changeability. A healthy code base maximizes productivity by making it easy to find where to change, make the change quickly, and introduce no errors.
- **Refactoring vs. restructuring.** "Restructuring" is the general term for any code reorganization. Refactoring is a specific kind: behavior-preserving transformations applied in small steps, each leaving the code in a working state. If someone says their code was broken for a couple of days while they were "refactoring," they were not refactoring.

---

## 2. When to Refactor

### Do refactor opportunistically, not in scheduled blocks

Refactoring is not an activity separated from programming — any more than you set aside time to write if-statements. Most refactoring happens while doing other things:

- **Preparatory refactoring**: Before adding a feature, refactor so the feature is easy to add. This is the best time to refactor. Look at the existing code and see whether it needs a different structure for your feature to slot in naturally.
- **Comprehension refactoring**: When reading code to understand it, refactor as you go. Rename variables now that you understand what they are. Chop a long function into smaller parts. As Ralph Johnson describes it: wiping the dirt off a window so you can see beyond. Those who dismiss comprehension refactoring as "useless fiddling" never see the opportunities hidden behind the confusion.
- **Litter-pickup refactoring**: When you spot something bad while doing something else, fix it immediately if small. Note it and fix it later if large. Always leave the camp site cleaner than when you found it. The nice thing about refactoring is you don't break the code with each small step — so it can take months to complete the job, yet the code is never broken while you're partway through it.

### The Rule of Three (Don Roberts)

- First time you do something: just do it.
- Second time you do something similar: wince, but do it anyway.
- Third time you do something similar: refactor.
- Or: three strikes, then you refactor.

### Long-term refactoring: do it gradually

For large refactorings (replacing a library, untangling a dependency web), do not dedicate a week to it. Instead, move the code in the right direction every time you touch it. Use **Branch By Abstraction** for library swaps: introduce an abstraction that can act as an interface to either library, then once calling code uses the abstraction, it's much easier to swap one library for another.

### Long-running features: use feature toggles

When Continuous Integration makes it impractical to have a partial feature on the mainline, use feature flags (feature toggles) to switch off any in-process features that can't be broken down into complete deliverable increments. This allows CI to work even for features that take weeks to build.

### Refactoring in code reviews

The best format for code review with respect to refactoring is sitting one-on-one with the original author, going through the code and refactoring as you go. The logical conclusion is pair programming: continuous code review embedded in programming.

### Do NOT refactor when:

- **The code works and you will never touch it again.** Ugly code that you treat as a black-box API need not be cleaned up. It's only when you need to understand how it works that refactoring gives benefit.
- **Rewriting is faster than refactoring.** When the existing code is so tangled that understanding it costs more than starting fresh. This requires judgment; try for a short time first to gauge difficulty.
- **The test suite is red.** Never refactor on a red bar. Restore green first, then refactor. "Never refactor on a red bar" — this phrase also means: if a test fails during refactoring, if you can't immediately see and fix the problem, revert to your last good commit and redo the steps with smaller increments.

### What to tell your manager

Refactoring is part of programming, not separate from it. Professionals refactor as part of delivering features and fixing bugs quickly. "If a manager would object, don't tell!" Subversive? Fowler says no: software developers are professionals whose job is to build effective software as rapidly as possible. Refactoring is the fastest way. A schedule-driven manager wants things done fastest; how you do it is your professional responsibility.

If the manager is genuinely tech-savvy and understands the Design Stamina Hypothesis, they should be actively encouraging refactoring and looking for signs a team isn't doing enough.

---

## 3. Code Smells and Remedies

Code smells are not precise rules but heuristics — signals that warrant a closer look. "No set of metrics rivals informed human intuition." You have to develop your own sense of how many instance variables or lines of code in a method are too many.

### Mysterious Name

A name that forces the reader to think or guess is a smell. Naming is one of the two hard things in programming. But more than aesthetics: "a good name can save hours of puzzled incomprehension in the future." People are often afraid to rename, thinking it's not worth the trouble — this is wrong.

- Signal: You have to read the function body to understand what a function does.
- Signal: You cannot think of a good name for something — this is often a sign of a deeper design problem. Puzzling over a tricky name often leads to significant simplifications.
- Apply **Change Function Declaration** to rename functions. Often, a good way to improve a name is to first write a comment to describe the function's purpose, then turn that comment into the name.
- Apply **Rename Variable** to rename variables.
- Apply **Rename Field** to rename fields and record keys.

### Duplicated Code

"If you see the same code structure in more than one place, you can be sure that your program will be better if you find a way to unify them. Duplication means that every time you read these copies, you need to read them carefully to see if there's any difference."

- Same expression in two methods of the same class: apply **Extract Function** and call from both.
- Similar but not identical code: use **Slide Statements** to arrange the similar items together for easy extraction.
- Duplicate fragments in subclasses of a common base class: use **Pull Up Method** to avoid calling one from another.

### Long Function

"In our experience, the programs that live best and longest are those with short functions." All payoffs of indirection — explanation, sharing, and choosing — are supported by small functions. The key to making small functions easy to understand is good naming.

- **The comment rule**: Whenever you feel the need to comment what a block of code does, write a function instead, named after the intention of the code rather than how it works. Even if the function call is longer than the code it replaces, if the name explains the purpose, extract it. "The key here is not function length but the semantic distance between what the method does and how it does it."
- Any function with more than half-a-dozen lines of code starts to smell; it's not unusual to have functions that are a single line of code.
- Signal: A block of code has a comment that tells you what it is doing — extract into a function named after the comment. Even a single line is worth extracting if it needs explanation.
- Signal: Conditionals and loops signal extraction. Use **Decompose Conditional** for conditional expressions. Extract switch/case legs into single function calls. If there's more than one switch on the same condition, apply **Replace Conditional with Polymorphism**.
- Signal: A loop that is hard to name — it may be doing two different things. Use **Split Loop**, then extract each loop.
- If temp variables and parameters block extraction: first use **Replace Temp with Query** to eliminate temps; use **Introduce Parameter Object** and **Preserve Whole Object** to reduce parameter lists.
- If parameter lists are still too long: use **Replace Function with Command** (heavy artillery — only if genuinely needed).

### Long Parameter List

"In our early programming days, we were taught to pass in as parameters everything needed by a function. This was understandable because the alternative was global data, and global data quickly becomes evil. But long parameter lists are often confusing in their own right."

- One parameter can be derived from another: apply **Replace Parameter with Query** to remove the redundant one.
- Many fields pulled from a structure: pass the structure with **Preserve Whole Object** instead.
- Several parameters always fit together: combine with **Introduce Parameter Object**.
- A parameter is a boolean flag dispatching different behavior: apply **Remove Flag Argument** and create two explicit functions. `setOn()` and `setOff()` are clearer than `set(true)` and `set(false)`.
- Multiple functions share the same set of parameters: use **Combine Functions into Class** to capture the common values as fields.

### Global Data

"The problem with global data is that it can be modified from anywhere in the code base, and there's no mechanism to discover which bit of code touched it. Time and again, this leads to bugs that breed from a form of spooky action from a distance."

- Signal: class variables, singletons, global variables that are modified from multiple locations.
- Immediately apply **Encapsulate Variable** to wrap any global in a function so access is trackable. "At least when you have it wrapped by a function, you can start seeing where it's modified and start to control its access."
- Move the variable into a class or module where only that module's code can see it.
- Global data is especially nasty when mutable. Global data that can be guaranteed never to change after program start is relatively safe — if the language can enforce that guarantee.
- "The dose makes the poison": You can get away with small doses of global data, but it gets exponentially harder to deal with the more you have.

### Mutable Data

"Changes to data can often lead to unexpected consequences and tricky bugs. I can update some data here, not realizing that another part of the software expects something different and now fails — a failure that's particularly hard to spot if it only happens under rare conditions."

- Encapsulate all writes via **Encapsulate Variable** so side effects are visible in one place.
- A variable updated to store different things: apply **Split Variable** to keep them separate.
- Separate queries from commands: use **Separate Query from Modifier** so callers don't need to call code with side effects unless they truly need to.
- Remove unnecessary setters with **Remove Setting Method** — just trying to find clients of a setter helps spot opportunities to reduce the scope of a variable.
- Mutable derived data is "particularly pungent": apply **Replace Derived Variable with Query** — it's a rich source of confusion, bugs, and inconsistency.
- When scope is broad: limit mutation with **Combine Functions into Class** or **Combine Functions into Transform**.
- When a data structure is modified in place: prefer replacing the whole structure with **Change Reference to Value**.

### Divergent Change

"Divergent change occurs when one module is often changed in different ways for different reasons." Example: "I will have to change these three functions every time I get a new database; I have to change these four functions every time there is a new financial instrument." These are separate contexts; move them into separate modules.

- Two concerns that form a natural sequence (fetch data then transform it): use **Split Phase** to separate them with a clear data structure between them.
- More back-and-forth in calls: use **Move Function** to divide code by context.
- Functions that mix both concerns: use **Extract Function** first, then move.
- Module is a class: use **Extract Class** to formalize the split.

### Shotgun Surgery

"Shotgun surgery is similar to divergent change but is the opposite. You whiff this when, every time you make a change, you have to make a lot of little edits to a lot of different classes. When the changes are all over the place, they are hard to find, and it's easy to miss an important change."

- Consolidate logic with **Move Function** and **Move Field** into a single module.
- A bunch of functions operating on similar data: use **Combine Functions into Class**.
- Functions transforming or enriching a data structure: use **Combine Functions into Transform**.
- **Useful tactic**: use inlining refactorings (**Inline Function** or **Inline Class**) to pull together poorly separated logic into a temporarily large intermediate form, then re-extract in a better shape. "Even though we are fond of small functions, we aren't afraid of creating something large as an intermediate step to reorganization."

### Feature Envy

"A classic case of Feature Envy occurs when a function in one module spends more time communicating with functions or data inside another module than it does within its own module." The function clearly wants to be with the data.

- Apply **Move Function** to relocate the function to the data it envies.
- Only part of a function envies another module: apply **Extract Function** on the envious part, then **Move Function** to give it a dream home.
- Function uses features of several modules: place it in the module with most of the data it uses. Use **Extract Function** to break the function into pieces that go in different places.
- Exception: patterns like Strategy and Visitor intentionally separate data from the behavior that varies. "The fundamental rule of thumb is to put things together that change together."

### Data Clumps

"Data items tend to be like children: they enjoy hanging around together." Groups of fields that appear together in multiple classes, or groups of parameters appearing in many method signatures, are an object trying to be born.

- Clump as fields: use **Extract Class** to turn them into an object.
- Clump in method signatures: use **Introduce Parameter Object** or **Preserve Whole Object**.
- Create a **class** (not just a plain record) so behavior can migrate to the new object over time. "We've often seen this as a powerful dynamic that creates useful classes."
- Test: consider deleting one of the data values. If the others make no sense without it, an object is needed.
- Do not worry about data clumps that use only some of the fields of the new object. As long as you are replacing two or more fields with the object, you will come out ahead.

### Primitive Obsession

"We find many programmers are curiously reluctant to create their own fundamental types which are useful for their domain — such as money, coordinates, or ranges." Strings are particularly common: "A telephone number is more than just a collection of characters." These are called "stringly typed" variables.

- Primitive needs behavior (formatting, validation, comparison): apply **Replace Primitive with Object**.
- Primitive is a type code controlling conditionals: apply **Replace Type Code with Subclasses** then **Replace Conditional with Polymorphism**.
- Groups of primitives that always appear together: apply **Extract Class** and **Introduce Parameter Object**.
- Examples of stringly typed variables: phone numbers, monetary amounts, geographic coordinates, physical quantities (do not add inches to millimeters), ranges.

### Repeated Switches

The first edition called this "Switch Statements," but the smell is not a switch per se — it is the same conditional switching logic (switch/case or cascading if/else) appearing in multiple places. "The problem with such duplicate switches is that, whenever you add a clause, you have to find all the switches and update them."

- Apply **Replace Conditional with Polymorphism** to centralize dispatch logic.
- A single, non-duplicated switch is not necessarily a smell. The problem is the repetition across multiple places.

### Loops

"Loops have been a core part of programming since the earliest languages. But we feel they are no more relevant today than bell-bottoms and flock wallpaper."

- Apply **Replace Loop with Pipeline** (map, filter, reduce) to express intent clearly. Pipeline operations make it easy to quickly see what elements are included and what is done with them.

### Lazy Element

A function, class, or module that no longer justifies its existence adds indirection without benefit. "It may be a function that's named the same as its body code reads, or a class that is essentially one simple function."

- Apply **Inline Function** when a function does nothing more than its name says.
- Apply **Inline Class** when a class has been reduced to triviality.
- Apply **Collapse Hierarchy** when a subclass or superclass no longer earns its place.

### Speculative Generality

"You get it when people say, 'Oh, I think we'll need the ability to do this kind of thing someday' and thus add all sorts of hooks and special cases to handle things that aren't required." YAGNI. "The machinery just gets in the way, so get rid of it."

- Signal: the only users of a function or class are test cases. Delete the test case and apply **Remove Dead Code**.
- Remove unused abstract classes with **Collapse Hierarchy**.
- Remove unnecessary delegation with **Inline Function** and **Inline Class**.
- Remove unused parameters with **Change Function Declaration**.

### Temporary Field

An object field that is only set in certain circumstances. "Such code is difficult to understand, because you expect an object to need all of its fields."

- Use **Extract Class** to create a home for the orphan variables.
- Use **Move Function** to put all the code that concerns those fields into the new class.
- Use **Introduce Special Case** to eliminate conditional code for when the variables aren't valid.

### Message Chains

`a.getB().getC().getD()` — the client is coupled to the entire navigation structure. Any change to the intermediate relationships forces the client to change.

- Apply **Hide Delegate** to add a method on the first object that performs the full navigation.
- Alternative: use **Extract Function** on the code that uses the final object, then **Move Function** to push it down the chain so the navigation stays local.
- Warning: doing Hide Delegate at every object in the chain often turns every intermediate object into a middle man. "Often, a better alternative is to see what the resulting object is used for."

### Middle Man

"One of the prime features of objects is encapsulation — hiding internal details from the rest of the world. Encapsulation often comes with delegation." But this can go too far: you look at a class's interface and find half the methods are delegating to another class.

- Apply **Remove Middle Man** to let callers call the delegated object directly.
- If only a few methods are pass-throughs: inline them with **Inline Function**.
- If there is additional behavior worth preserving: use **Replace Superclass with Delegate** or **Replace Subclass with Delegate** to fold the middle man into the real object.

### Insider Trading

Modules that pass excessive data back and forth are too tightly coupled. "Modules that whisper to each other by the coffee machine need to be separated."

- Reduce coupling with **Move Function** and **Move Field**.
- When two modules need to share data: create a third module as a shared intermediary, or use **Hide Delegate** to make another module act as intermediary.
- Inheritance is a common source of excessive coupling between subclass and parent: consider **Replace Subclass with Delegate**.

### Large Class

"When a class is trying to do too much, it often shows up as too many fields. When a class has too many fields, duplicated code cannot be far behind."

- Signal: common prefixes or suffixes for some subset of variables — for example, `depositAmount` and `depositCurrency` suggest an opportunity to extract them into a component.
- Apply **Extract Class** to split related groups of fields into cohesive components.
- Apply **Extract Superclass** when a natural generalization exists.
- Apply **Replace Type Code with Subclasses** when the class has type-specific behavior.
- Look at clients of the class: different subsets of clients using different subsets of the interface suggests separate classes.

### Alternative Classes with Different Interfaces

Two classes doing similar things but with different signatures cannot be substituted for each other.

- Use **Change Function Declaration** to align method signatures.
- Use **Move Function** to align behavior.
- If duplication results: use **Extract Superclass** to atone.

### Data Class

A class with only fields and getters/setters. "Such classes are dumb data holders and are often being manipulated in far too much detail by other classes." Behavior belongs with data.

- Apply **Encapsulate Record** to any public fields immediately.
- Apply **Remove Setting Method** on fields that should not change after construction.
- Look for where getting/setting methods are used by other classes. Use **Move Function** (or **Extract Function** + move) to bring that behavior into the data class.
- Exception: a data class used as a result record from a distinct operation — for example, the output of **Split Phase** — is acceptable, especially if immutable. "Immutable fields don't need to be encapsulated and information derived from immutable data can be represented as fields rather than getting methods."

### Refused Bequest

A subclass that inherits methods or data it does not want. The traditional advice is to push unused code down with **Push Down Method** and **Push Down Field**. But Fowler moderates this:

- "Nine times out of ten this smell is too faint to be worth cleaning." Subclassing to reuse implementation, even if the subclass ignores some inherited behavior, is often acceptable.
- The smell is much stronger when the subclass refuses the interface of the superclass, not just the implementation. This is the serious case: "Don't fiddle with the hierarchy; you want to gut it by applying **Replace Subclass with Delegate** or **Replace Superclass with Delegate**."

### Comments

"Comments are often used as a deodorant. It's surprising how often you look at thickly commented code and notice that the comments are there because the code is bad."

- Before writing a comment to explain what a block does: apply **Extract Function** and name it after the intent.
- Before writing a comment to explain what a function does: apply **Change Function Declaration** to rename it more clearly.
- Before writing a comment to explain a required system state: apply **Introduce Assertion** to make the assumption executable.
- When to use a comment: when you don't know what to do; when explaining *why* (not *what*); when flagging a known limitation or area of uncertainty; when describing algorithmic choices.

---

## 4. Refactoring Catalogue

### 4.1 Composing Functions (Chapter 6)

---

**Extract Function**

Motivation: "If you have to spend effort looking at a fragment of code and figuring out what it's doing, then you should extract it into a function and name the function after the 'what.'" The separation between intention and implementation is what matters, not length. A function whose body is as clear as its name is a good candidate for inlining; a function whose purpose is not obvious from its implementation is a candidate for extraction.

Detection signals:
- You feel the need to comment what a block does.
- A block's intent is not obvious from its implementation.
- A block is used more than once.
- A comment precedes a block ("// calculate outstanding").

Mechanics:
1. Create a new function. Name it by intent — what it does, not how it does it. If you can't come up with a better name than the code itself, don't extract yet. But it's OK to try, discover it isn't helping, and inline it back.
2. If the language supports nested functions, nest the extracted function inside the source function to reduce out-of-scope variable issues.
3. Copy the extracted code from the source function into the new function.
4. Scan the extracted code for references to variables local in scope to the source function. Pass them as parameters. Variables used but not assigned to can simply be passed in. If a variable is only used inside the extracted code but declared outside, move the declaration inside.
5. If an extracted variable is assigned to (and the assignment needs to be visible outside): treat the extracted code as a query and return the assigned value. If too many variables are assigned, abandon the extraction and apply **Split Variable** or **Replace Temp with Query** first.
6. Compile after all variables are dealt with.
7. Replace the extracted code in the source function with a call to the new function.
8. Test.
9. Look for other code that does the same or similar thing and consider **Replace Inline Code with Function Call**.

Nuance: Fowler's convention is to name parameters with an indefinite article (e.g., `aPerformance`, `aCustomer`) to signal their type role. The return value is always named `result` — "that way I always know its role."

---

**Inline Function**

Motivation: "Sometimes, I do come across a function in which the body is as clear as the name." Also use when you have a group of functions that seem badly factored and need reassembling into one big function before re-extracting in a better shape. Also use when there is too much indirection — when every function does simple delegation and you get lost in it all.

Mechanics:
1. Check that the function is not polymorphic (has subclasses that override it).
2. Find all callers of the function.
3. Replace each call with the function's body. Test after each replacement.
4. The entire inlining doesn't have to happen at once. If some parts are tricky, move one line at a time using **Move Statements to Callers** mechanics.
5. Remove the function definition.
6. If you encounter recursion, multiple return points, or inlining across objects without accessors — do not do this refactoring.

---

**Extract Variable** (formerly Introduce Explaining Variable)

Motivation: "Expressions can become very complex and hard to read. Local variables may help break the expression down into something more manageable. In particular, they give me an ability to name a part of a more complex piece of logic." Also handy for debugging.

Mechanics:
1. Ensure the expression has no side effects.
2. Declare an immutable variable. Set it to the expression.
3. Replace the original expression with the variable.
4. Test. If the expression appears more than once, replace each occurrence, testing after each.

Nuance: If the expression belongs in a broader class context (not just this function), extract it as a method instead of a local variable. In a class, Extract Variable naturally becomes an Extract Function — "this is one of the great benefits of objects, they give you a reasonable amount of context."

---

**Inline Variable**

Motivation: Sometimes the variable name adds nothing to understanding, or it impedes a further extraction.

Mechanics:
1. Check the right-hand side of the assignment is free of side effects.
2. If the variable isn't declared immutable, do so and test (confirms it's assigned only once).
3. Find the first reference and replace with the right-hand side.
4. Test. Repeat until all references are replaced.
5. Remove the declaration and assignment. Test.

---

**Change Function Declaration** (Rename Function, Add/Remove Parameter)

Motivation: "The most important element of such a joint is the name of the function. A good name allows me to understand what the function does when I see it called, without seeing the code that defines its implementation." Parameters set the context in which a function can be used; changing them can remove coupling or increase applicability.

Two paths:
- **Simple mechanics**: Use when all callers are visible and the change is straightforward. Change the declaration, find all references, update them all. Test. Best to separate name changes from parameter changes — do each as a separate step.
- **Migration mechanics**: Use when callers are spread across many places, the function is a published interface, or the change is complex. (1) Apply **Extract Function** to the function body to create the new function. If it will have the same name, give it a temporary name. (2) Add any needed parameters. Test. (3) Apply **Inline Function** to the old function. (4) If you used a temporary name, rename to the final name. Test. For published APIs: you can pause after creating the new function, deprecate the original, and wait for clients to migrate.

---

**Encapsulate Variable**

Motivation: Functions can be moved around more easily than data. Moving data means updating all references. Wrapping data in functions first means you can rename by changing the function name in one place, monitor all accesses, and control mutation. "The key defense here is Encapsulate Variable."

- Apply immediately to any global or wide-scope mutable variable.
- The wrapping function becomes the control point for monitoring and restricting access.

---

**Rename Variable**

Never fear renaming. A good name saves hours of puzzled reading. Renaming is not just aesthetic — when you can't think of a good name, it's often a sign of a design problem. Use automated rename where available.

---

**Introduce Parameter Object**

Motivation: Groups of data that always appear together as parameters suggest a natural abstraction. Benefits: shrinks parameter lists; simplifies calling; and — crucially — creates a place for behavior to migrate to over time.

- Prefer a class over a plain record so behavior can migrate to the object.

---

**Combine Functions into Class**

Motivation: When functions share a common set of parameters or a common data structure, grouping them into a class provides shared context. Individual functions become methods; shared state becomes fields.

- Use when multiple functions share several parameter values — the class is like a set of partially applied functions.
- Use when the mutable state makes **Combine Functions into Transform** inadvisable.

---

**Combine Functions into Transform**

Motivation: Useful in read-only pipelines; avoids scattered derivation logic. Takes input data and enriches it with all derived fields, returning the enriched structure.

- Use **Combine Functions into Class** instead when the data is mutable (transforms can cause consistency problems with mutation — the original data changes but the enriched copy doesn't reflect that).

---

**Split Phase**

Motivation: "When code deals with two different things, look for a way to split it into separate modules." The first phase calculates data required for the second; the second renders or processes it.

- Signal: a section of code handles two different concerns in sequence (e.g., parse input then process it; calculate data then format it).
- The intermediate data structure between phases makes the boundary explicit and testable. It also makes it easy to add a second consumer (e.g., an HTML renderer) that operates on the same intermediate structure without duplicating the calculation logic.

Mechanics:
1. Extract the second-phase code into its own function.
2. Create an intermediate data structure as the link between the two phases.
3. Pass this data structure as a parameter to the second-phase function.
4. Move data from the original parameters into the intermediate structure, eliminating those parameters from the second-phase function one at a time.
5. Extract the first phase into its own function. Test.

---

### 4.2 Moving Features (Chapter 8)

---

**Move Function**

Motivation: "When I modularize a program, I'm trying to separate the code into zones to maximize the interaction inside a zone and minimize interaction between zones." Move a function when it references more context from another module than its own.

- Trigger: the function uses more features from another context than its own.
- Trigger: the function is closely coupled to elements in another module.
- After moving: decide whether to keep the source as a delegating function or inline it at call sites.

Mechanics:
1. Examine all code the function uses. Consider whether those should also move.
2. Check if the function is a polymorphic method — extra care needed for inheritance hierarchies.
3. Copy the function into the target context. Adjust to fit (rename parameters, update internal calls).
4. Perform static analysis.
5. Turn the source function into a delegating function calling the target.
6. Test.
7. Consider whether to inline the source function (remove the delegation).

---

**Move Field**

Motivation: A field is used more by another class than its own. Data clumps suggest co-location. Encapsulation motivates co-locating data with the code that uses it.

- Trigger: a field is passed as a parameter to many functions in another class — often a sign it belongs there.

---

**Move Statements into Function / Move Statements to Callers**

Move into function when: code always appears before or after a call and is semantically part of the call.

Move to callers when: a function's behavior needs to diverge across call sites. Apply when a single statement begins to need to vary per call site.

---

**Slide Statements**

Purpose: Reorder statements so related code is adjacent. This is a prerequisite for most extractions — reducing the chance of needing to pass temporary variables.

---

**Split Loop**

Motivation: A loop that does two things at once is harder to understand and extract than two loops each doing one thing. Do not worry about the apparent performance cost — "most of the time, rerunning a loop like this has a negligible effect on performance."

---

**Replace Loop with Pipeline**

Apply map, filter, reduce (or similar) to express intent clearly. "Pipeline operations make it easy to quickly see the elements that are included in the processing and what is done with them."

---

**Remove Dead Code**

"When we put code in version control, every line of code that's been removed is easily retrieved. There's no longer any need to comment-out code. Dead code is pure noise." Delete it. Version control is the backup.

---

### 4.3 Organizing Data (Chapter 9)

---

**Split Variable**

Each variable should have exactly one responsibility. If a variable is assigned more than once for different purposes, it has too many responsibilities.

- Declare each new variable as `const` (or equivalent immutable declaration) where possible. This ensures each is only assigned once.
- Exception: loop variables and collecting variables (accumulators) are naturally assigned multiple times and are fine.

---

**Encapsulate Record**

Apply immediately to any public-field record or plain object that is accessed directly. Gives you a place to add invariants and behavior later.

---

**Encapsulate Collection**

Return copies of collection fields; provide add/remove methods instead of direct access. Prevents callers from modifying the collection directly.

---

**Replace Primitive with Object** (formerly Replace Data Value with Object)

"Even experienced developers find this one of the most valuable refactorings." Apply when you find yourself duplicating logic around a primitive — formatting, validation, comparison.

---

**Replace Derived Variable with Query**

"Mutable data that can be calculated elsewhere is particularly pungent. It's not just a rich source of confusion, bugs, and missed dinners at home — it's also unnecessary." Remove variables whose value is computed from other data; replace with a function that computes on demand.

---

**Change Reference to Value / Change Value to Reference**

- Change Reference to Value: when an object's identity doesn't matter (e.g., a Money amount), make it a value object — immutable, equal by content.
- Change Value to Reference: when multiple objects should share and update the same data, make them refer to a shared instance rather than hold separate copies.

---

### 4.4 Simplifying Conditional Logic (Chapter 10)

---

**Decompose Conditional**

This is **Extract Function** applied specifically to if/else logic. Extract the condition and each branch into named functions. The names explain *why* the code branches, not just *what* happens. "A big block of if-else that makes you stop and wonder what's going on can be replaced with self-explanatory function calls."

---

**Consolidate Conditional Expression**

Combine a series of conditionals that all lead to the same result into a single expression. Then extract the condition into a function whose name explains why the result follows.

- Do not consolidate genuinely independent checks that happen to share a result.

---

**Replace Nested Conditional with Guard Clauses**

Use guard clauses when one branch is the "normal" path and others are special cases or exceptional situations. Guard clauses signal "this is unusual; if it applies, exit early."

- Do not use when both branches are equally important — a flat if/else is clearer there.
- When reversing conditions to convert to guard clauses, it's fine to have multiple exits from a function. Fowler considers "a function with a single exit point" to be a useful rule only on rare occasions.

---

**Replace Conditional with Polymorphism**

Motivation: "Conditional logic like this tends to decay as further modifications are made unless it's reinforced by more structural elements." When the same switch on a type code appears in multiple functions, every new case requires updating every switch.

- Use when the same switch on a type code appears in multiple functions (the repeated switch smell).
- Use when the logic has a natural base case with clear variants that override it.
- Do not replace every conditional with polymorphism — only where it genuinely reduces repetition or clarifies extension. "Not every conditional is a good candidate for Replace Conditional with Polymorphism."

---

**Introduce Special Case** (formerly Introduce Null Object)

Handle a special value (commonly null) with a dedicated class or object that returns sensible defaults. Eliminates repeated null checks scattered through the codebase.

- Can be implemented as a class, an object literal, or a transform (a function that converts null to the special-case representation).

---

**Introduce Assertion**

Add an explicit assertion for an assumption the code makes. "Assertions are a form of communication: they tell the reader that the code assumes a particular state." They also help catch bugs during development.

- Do not use assertions for conditions that are likely to be false in production and require proper error handling there. Assertions are for conditions that should always be true.

---

### 4.5 Refactoring APIs (Chapter 11)

---

**Separate Query from Modifier**

"A good rule is: Any function that returns a value should not have observable side effects — the Command-Query Separation principle." If a function both returns a value and has observable side effects, split it into a pure query and a pure command.

- Queries (returning a value with no side effects) can be called freely, moved anywhere, tested easily, called in any order.

---

**Parameterize Function**

If two functions do the same logic with different literal values, merge into one with a parameter for the differing value. Also apply when a function is nearly what you need but has literal values that conflict with your needs (preparatory refactoring).

---

**Remove Flag Argument**

When a boolean parameter selects between two behaviors, replace with two explicit functions. "I dislike flag arguments because they complicate the process of understanding what function calls mean. When I see the call, I have no immediate sight of what true means."

- Exception: a flag that is passed through from a higher context without switching on it is less harmful.

---

**Preserve Whole Object**

Pass the original object rather than pulling out multiple individual values. Reduces parameter list length; makes the function resilient to adding new values it needs in the future.

- Do not apply if it would create an undesirable dependency on the whole object where the callee should not know about it.

---

**Replace Parameter with Query**

When a function can derive a parameter's value from other data it already has access to, remove the parameter and compute the value internally. Reduces the burden on the caller.

---

**Replace Query with Parameter**

Move a global or module-state query inside a function out as a parameter, making the function pure and its inputs explicit. Useful for improving testability or removing a hidden dependency.

---

**Remove Setting Method**

Remove a setter when a field should only be set at construction time. Makes the immutability intent explicit. Even just trying to find clients of a setter often helps spot opportunities to reduce the scope of a variable.

---

**Replace Constructor with Factory Function**

Use when you need more flexibility in naming, subclass selection, or creation logic. Constructors are limited: in JavaScript, they can't return subclasses — a factory function provides the needed flexibility. Also useful when you want to encapsulate the decision of which subclass to create.

---

**Replace Function with Command**

Wrap a function in a command object when you need: undo support; building parameters over time; fine-grained lifecycle control; or the ability to subclass to extend behavior.

- "This is the Command pattern. Do not apply unless you genuinely need these capabilities; functions are simpler."

---

**Replace Command with Function**

Inline a command object back into a plain function when the extra complexity is not needed.

---

### 4.6 Dealing with Inheritance (Chapter 12)

---

**Pull Up Method**

Motivation: "Eliminating duplicate code is important. Two duplicate methods may work fine as they are, but they are nothing but a breeding ground for bugs in the future. Whenever there is duplication, there is risk that an alteration to one copy will not be made to the other."

- First make the methods identical (using **Parameterize Function** or by hand). Often Pull Up Method comes after Parameterize Function.
- The most awkward complication: the method body refers to features on the subclass not on the superclass. Use **Pull Up Field** and **Pull Up Method** on those elements first.

Mechanics:
1. Inspect methods to ensure they are identical. If not, refactor them until their bodies are identical.
2. Check that all method calls and field references inside the method body refer to features accessible from the superclass.
3. If methods have different signatures, use **Change Function Declaration** to align them.
4. Create a new method in the superclass. Copy the body of one of the methods.
5. Run static checks.
6. Delete one subclass method. Test. Repeat until all are gone.
7. Consider adding a "subclass responsibility" trap method on the superclass for any method it calls that subclasses must implement.

---

**Pull Up Field**

Motivation: Reduces duplication in two ways: removes the duplicate data declaration, and creates a place to move behavior that uses the field.

Mechanics:
1. Inspect all users of the candidate field to ensure they're used the same way.
2. If fields have different names, use **Rename Field** to give them the same name.
3. Create a new field in the superclass (protected in common languages).
4. Delete the subclass fields. Test.

---

**Pull Up Constructor Body**

When subclasses have common constructor initialization, move the common statements into the superclass constructor.

Mechanics:
1. Define a superclass constructor if one doesn't exist. Ensure it's called by subclasses.
2. Use **Slide Statements** to move common statements to just after the super call.
3. Remove common code from each subclass, add it to superclass, pass referenced constructor parameters to super.
4. Test.
5. If common code cannot move to the constructor start: use **Extract Function** followed by **Pull Up Method**.

---

**Push Down Method**

Motivation: If a method is only relevant to one subclass (or a small proportion), removing it from the superclass makes that clearer.

- Can only be done if the caller knows it's working with a particular subclass — otherwise use **Replace Conditional with Polymorphism** with placebo behavior on the superclass.

Mechanics:
1. Copy the method into every subclass that needs it.
2. Remove the method from the superclass. Test.
3. Remove the method from each subclass that doesn't need it. Test.

---

**Push Down Field**

Move a field used only by some subclasses to those subclasses.

---

**Replace Type Code with Subclasses**

Motivation: When a type code controls conditional behavior in multiple places, subclasses allow polymorphism to centralize the dispatch. Also useful when fields or methods are only valid for particular type code values — a subclass makes that relationship explicit.

Two forms:
- **Direct inheritance**: Subclass the class itself. Simpler. Cannot use if the type changes at runtime or if the class is already subclassed for another reason.
- **Indirect inheritance**: Create a separate type class hierarchy, held as a field. Allows the type to change at runtime and combines with other subclass hierarchies.

Mechanics (direct):
1. Self-encapsulate the type code field.
2. Pick one type code value. Create a subclass. Override the type code getter to return the literal value.
3. Use **Replace Constructor with Factory Function** and put selector logic in the factory.
4. Test. Repeat for each type code value. Test after each.
5. Remove the type code field. Test.
6. Use **Push Down Method** and **Replace Conditional with Polymorphism** on any methods that use the type code accessors.

---

**Remove Subclass**

When a subclass no longer earns its keep — not enough special behavior remains — replace it with a field on the superclass.

Mechanics:
1. Use **Replace Constructor with Factory Function** on the subclass constructor.
2. Use **Extract Function** on any type tests and **Move Function** to move them to the superclass.
3. Create a field to represent the subclass type.
4. Change methods that refer to the subclass to use the new field.
5. Delete the subclass. Test.

---

**Extract Superclass**

When two classes have common features, create a superclass to house the shared data and behavior. "I can make the common behavior more explicit by extracting a common superclass."

Mechanics:
1. Create an empty superclass. Let the candidate classes extend from it.
2. Use **Pull Up Constructor Body**, **Pull Up Field**, **Pull Up Method** to move shared elements to the superclass.
3. Inspect remaining methods for common parts. Use **Extract Function** to split off the common part if needed, then **Pull Up Method**.
4. Check clients of the original classes to see if they can use the superclass interface.

---

**Collapse Hierarchy**

When a superclass and subclass are not different enough to justify a hierarchy, merge them. Apply when a class and its parent have converged through refactoring.

---

**Replace Subclass with Delegate**

Motivation: "Inheritance is a card that can only be played once. If I have more than one reason to vary something, I can only use inheritance for a single axis of variation." Also, "inheritance introduces a very close relationship between classes. Any change I want to make to the parent can easily break children."

Key distinctions vs. Replace Superclass with Delegate:
- Replace Subclass with Delegate: the relationship was IS-A and the subclass genuinely extends the superclass. The issue is either (a) needing multiple axes of variation, or (b) needing to change the type dynamically (e.g., `aBooking.bePremium()`).
- Replace Superclass with Delegate: the subclass should never have been a subclass in the first place (the IS-A relationship was wrong).

Fowler's guideline: "I use inheritance frequently, partly because I always know I can use Replace Subclass with Delegate should I need to change it later. Inheritance is a valuable mechanism that does the job most of the time without problems. So I reach for it first, and move onto delegation when it starts to rub badly."

Mechanics:
1. If many callers use constructors, apply **Replace Constructor with Factory Function**.
2. Create an empty delegate class. Its constructor should take any subclass-specific data plus, usually, a back-reference to the host.
3. Add a field to the superclass to hold the delegate.
4. Modify subclass creation to initialize the delegate field.
5. Choose a subclass method to move. Use **Move Function** to move it to the delegate. Don't remove the source's delegating code yet.
6. Add dispatch logic to the superclass method: check for presence of delegate, use it if present.
7. If the source method has callers outside the class, move the delegating code from the subclass to the superclass, guarded by the delegate check.
8. Test. Repeat until all subclass methods are moved.
9. Change all callers of the subclass constructor to use the superclass constructor. Test.
10. Remove the subclass with **Remove Dead Code**.

Handling `super` calls in the delegate: either extract a private base-price method on the superclass so the delegate can call it without recursion, or recast the delegate as an extension (pass the base result to the delegate to extend).

---

**Replace Superclass with Delegate**

Motivation: "One of the classic examples of mis-inheritance from the early days of objects was making a stack be a subclass of list." This exposes all list operations on the stack interface, most of which are invalid for a stack. "If functions of the superclass don't make sense on the subclass, that's a sign that I shouldn't be using inheritance to use the superclass's functionality."

Also: the type-instance homonym — using a CarModel class (with name and engine size) as a superclass for a PhysicalCar is a common and often subtle modeling mistake. A physical car is not a type of car model.

Rule: every instance of the subclass must be a valid instance of the superclass in every context where the superclass is used (Liskov Substitution Principle). If not, the inheritance is wrong.

Fowler's guidance: "So my advice is to (mostly) use inheritance first, and apply Replace Superclass with Delegate when (and if) it becomes a problem." Even though forwarding functions are boring to write, they are too simple to get wrong.

Mechanics:
1. Create a field in the subclass that refers to the superclass object. Initialize this delegate to a new instance.
2. For each element of the superclass, create a forwarding function in the subclass that forwards to the delegate. Test after each consistent group.
3. When all superclass elements have been overridden with forwarders, remove the inheritance link. Test.

---

**Substitute Algorithm**

Replace the body of a function with a cleaner algorithm that produces the same result. Do this when you discover a simpler way, or when you need to align with a library function.

---

## 5. Testing Strategy

### The absolute prerequisite: tests before refactoring

- "Before you start refactoring, make sure you have a solid suite of tests. These tests must be self-checking." If you do not have tests, write them first. Refactoring without tests is reckless.
- Tests must be **self-checking**: they pass or fail automatically. "If I don't, I'd end up spending time hand-checking values from the test against values on a desk pad, and that would slow me down."
- Run the test suite after every refactoring step. "If I try to do too much, making a mistake will force me into a tricky debugging episode. Small changes, enabling a tight feedback loop, are the key."

### Test design

- **Risk-driven testing**: "Testing should be risk-driven; I'm trying to find bugs, now or in the future. Therefore I don't test accessors that just read and write a field: They are so simple that I'm not likely to find a bug there."
- Test behavior, not implementation. Focus on calculated values, boundary conditions, error cases.
- **Never share mutable fixture between tests.** Use `beforeEach` to create a fresh fixture for each test. "Shared fixtures cause tests to interact, yielding different results depending on what order tests are run in."
- **Verify that each test can fail.** When writing a test against existing code, temporarily inject a fault to confirm the test detects it. "A test that never fails gives false confidence."
- "Never refactor on a red bar." The phrase also covers: if you can't immediately find a failing test's cause, revert to the last commit and redo with smaller steps.

### When to run tests

- "Run tests exercising the code you're working on at least every few minutes; run all tests at least daily."
- The tighter the feedback loop, the shorter the debugging episode when you make a mistake.

### Responding to bugs

- "When you get a bug report, start by writing a unit test that exposes the bug." Only then fix it. The test ensures the bug stays fixed. Ask: does this bug give clues to other gaps in the test suite?

### TDD

- "Writing the test first concentrates me on the interface rather than the implementation (always a good thing). It also means I have a clear point at which I'm done coding — when the test passes."
- The test-code-refactor cycle should occur many times per hour.

### Enough tests

- "The best measure for a good enough test suite is subjective: How confident are you that if someone introduces a defect into the code, some test will fail?"
- Coverage metrics reveal gaps but cannot measure quality.
- "Over-testing is vanishingly rare compared to under-testing. Write more tests."

---

## 6. Refactoring and Performance

"To make the software easier to understand, I often make changes that will cause the program to run slower. This is an important issue."

- **Do not optimize during refactoring.** Finish refactoring first, then do performance tuning afterwards. "Most of the time, rerunning a loop like this has a negligible effect on performance. If you timed the code before and after this refactoring, you would probably not notice any significant change."
- **Profile before optimizing.** "Most programmers, even experienced ones, are poor judges of how code actually performs. Many of our intuitions are broken by clever compilers, modern caching techniques, and the like." In Ron Jeffries's example, the team speculated extensively about a performance problem and were completely wrong; profiling revealed the real cause in five minutes.
- **The 90-percent rule.** The performance of software usually depends on just a few parts of the code. Optimizing everything equally wastes 90% of the effort on code that isn't causing problems.
- **Well-factored code is easier to tune.** "Having a well-factored program helps with optimization in two ways. First, it gives me time to spend on performance tuning. Second, with a well-factored program I have finer granularity for my performance analysis."
- **Three approaches to performance**: (1) Time budgeting — for hard real-time systems only. (2) Constant attention — intuitively attractive but does not work well; improvements are spread all around and made with narrow perspective. (3) Profile-guided optimization — build software well-factored first, then use a profiler to find hot spots, optimize only those.

---

## 7. Economics and Management Framing

- **Design Stamina Hypothesis**: "By putting our effort into a good internal design, we increase the stamina of the software effort, allowing us to go faster for longer." The teams that can add features faster are the ones with better internal quality, because good modularity means only understanding a small subset of the code base to make a change.
- **The economic argument is the only one that works**: "The point of refactoring isn't to show how sparkly a code base is — it is purely economic. We refactor because it makes us faster — faster to add features, faster to fix bugs. It's important to keep that in front of your mind and in front of communication with others."
- **What to tell managers**: If they are tech-savvy and understand Design Stamina, encourage them to require regular refactoring. If not: "Don't tell! ... Software developers are professionals. Our job is to build effective software as rapidly as we can. My experience is that refactoring is a big aid to building software quickly... The fastest way is by refactoring — therefore I refactor."
- **Planned vs. opportunistic**: Most refactoring should be "unremarkable, opportunistic." Dedicated refactoring sprints should be rare exceptions, not a regular practice. "A week spent refactoring now can repay itself over the next couple of months" — but only when a team has badly neglected it.
- **Feature branches vs. CI**: Long-running feature branches make refactoring harder because refactorings involving widespread changes (e.g., renaming a widely-used function) create semantic merge conflicts. "Many teams find refactorings so exacerbate merge problems that they stop refactoring." CI and refactoring work well together.

---

## 8. Anti-Patterns Table

| Bad Pattern | Smell(s) | Refactoring to Apply |
|---|---|---|
| A function with a comment before each block | Long Function, Comments | Extract Function (name after the comment) |
| `doXAndY()` — function that does two things | Long Function | Extract Function, Split Phase |
| `process(true)` — boolean flag argument | Long Parameter List | Remove Flag Argument → two explicit functions |
| `a.getB().getC().getD()` — message chain | Message Chains | Hide Delegate or Extract + Move Function |
| Half the methods on a class just delegate to another class | Middle Man | Remove Middle Man or Inline Function |
| Class with only getters/setters and no behavior | Data Class | Move Function (from clients into class); Encapsulate Record |
| Same switch statement in three places | Repeated Switches | Replace Conditional with Polymorphism |
| Null checks repeated everywhere | Temporary Field, Message Chains | Introduce Special Case (Null Object) |
| `if (type === 'A') ... else if (type === 'B') ...` in multiple functions | Repeated Switches, Primitive Obsession | Replace Type Code with Subclasses + Replace Conditional with Polymorphism |
| Variables named `temp`, `data`, `info`, `result2` | Mysterious Name | Rename Variable / Rename Function |
| A string holding a phone number, currency, or status code | Primitive Obsession | Replace Primitive with Object |
| Three or more parameters that always appear together | Data Clumps, Long Parameter List | Introduce Parameter Object |
| `Stack extends List` (superclass interface doesn't fit subclass) | Refused Bequest | Replace Superclass with Delegate |
| `calculateAndSendBill()` — query with side effect | Mutable Data | Separate Query from Modifier |
| A field computed from other fields but stored redundantly | Mutable Data, Divergent Change | Replace Derived Variable with Query |
| Global mutable state | Global Data | Encapsulate Variable |
| A variable reused for different purposes across a function | Mutable Data | Split Variable |
| `if (result !== null)` before every use of a computed value | Temporary Field | Introduce Special Case; Replace Temp with Query |
| Two functions identical except for a literal value | Duplicated Code | Parameterize Function |
| A subclass that inherits methods it does not want | Refused Bequest | Push Down Method + Push Down Field; Replace Subclass with Delegate |
| A class that changes for database reasons and for financial reasons | Divergent Change | Extract Class; Split Phase |
| Changing one feature requires editing 6 different files | Shotgun Surgery | Move Function + Move Field; Combine Functions into Class |
| Code has hooks for features that were never built | Speculative Generality | Collapse Hierarchy; Inline Function/Class; Remove Dead Code |
| A loop that counts items and sums their values simultaneously | Long Function, Loops | Split Loop (one loop per concern) |
| Complex nested if/else with a "happy path" buried in the middle | Long Function | Replace Nested Conditional with Guard Clauses |
| A condition whose purpose is not obvious | Mysterious Name | Decompose Conditional (extract condition into named function) |
| Subclass only differs from parent in one boolean-controlled branch | Primitive Obsession | Replace Type Code with Subclasses; simplify via Pull Up |
| Class hierarchy too deep with little difference between levels | Lazy Element, Speculative Generality | Collapse Hierarchy |
| Inheritance used for two independent axes of variation | Refused Bequest, Large Class | Replace Subclass with Delegate |
| Type code stored as a string or number controlling conditional logic | Primitive Obsession | Replace Primitive with Object, then Replace Type Code with Subclasses |
| Subclass needs to change type dynamically at runtime | Refused Bequest | Replace Subclass with Delegate (enables `bePremium()` style mutation) |
| CarModel used as superclass for PhysicalCar | Refused Bequest | Replace Superclass with Delegate (type-instance homonym) |
| Two subclasses with almost-identical methods differing only in literal values | Duplicated Code | Parameterize Function, then Pull Up Method |
