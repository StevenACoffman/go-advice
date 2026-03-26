# Functional Core, Imperative Shell

**Gary Bernhardt** · Destroy All Software, Episode DAS-0072

* * *

The ideal program contains a large body of immutable code written in a functional style, and then small pieces of mutable code doing imperative things — but having those pieces be very localized, separate from the data, and separate from the core behavior of the system. The pattern that achieves this is **functional core, imperative shell**: push all the logic and data manipulation into immutable objects, and confine every side effect, every mutable reference, and every interaction with the outside world to a thin outer shell.

Bernhardt demonstrates this pattern through his Ruby Twitter client. The application is built in this way, so the codebase serves as a concrete, working illustration rather than a toy example.

* * *

## The Functional Core

The functional core is composed of immutable objects whose methods return new objects rather than changing themselves. The Twitter client has four.

**Tweet** is the simplest. It carries an ID, a username, and a text, and its methods only return values. No mutation at all.

**Timeline** stores an array of tweets and knows how to incorporate new tweets arriving from the API. When new tweets come in, it does not modify its own array; it constructs and returns a new `Timeline` that merges the new tweets with the old. The phrase "update itself" is a misleading description — it never changes state, it produces a successor.

**Cursor** tracks the selected position in the tweet list and moves it Vim-style. At almost 100 lines it is larger than Bernhardt would like — the responsibilities are muddier than they should be, and he is not happy with it — but it is immutable. It holds an array of tweets and a selection index. Every movement method returns a new cursor; nothing on this object or any other object is changed.

**TweetRenderer** is the most complex functional class, about 50 lines. It takes an array of tweets, a selected tweet, and a screen height, and returns an array of text lines for the screen class to draw. It is purely computational — data in, data out.

The only mutation in any of these classes is the rebinding of local variables: a `lines` variable reassigned partway through a method, for instance. Local rebinding is invisible from outside. These objects are immutable at the boundary that matters — the one other code crosses.

* * *

## The Imperative Shell

If nothing in the core is ever mutated, something still has to handle user input, talk to the network, write to the database, and hold the mutable references that connect the program's internal state to the state of the world. That is the imperative shell, and in this application it lives almost entirely in one file of about 160 lines.

### Cursor movement

The key-handling code is the clearest example. Pressing `j` calls `cursor.move_down`; pressing `k` calls `cursor.move_up`. Each returns a new cursor object. The shell assigns that new cursor to an instance variable. The mutation is the assignment — a single line, entirely in the shell. The logic of what "down" means lives in the functional core.

### Rendering

Just above the key handling, the shell calls `TweetRenderer` to produce an array of lines, then passes those lines to the screen to draw. Drawing is a destructive, side-effecting action. The renderer that decided what to draw is not — it computed an array and returned it.

### Background network I/O

Pulling new tweets from Twitter involves a background thread. On startup, `run` calls `start_timeline_stream`, which constructs a `TimelineStream` object and launches a thread running an infinite loop inside it. The loop fetches new tweets, constructs a new `Timeline` value, and pushes it into a queue.

The main thread blocks on a `select` system call watching two sources simultaneously: standard in for keyboard input, and the queue for incoming timelines. The queue is a special class that plugs into `select` by maintaining a fake IO object inside it — Ruby's `select` expects IO objects, so the queue wraps a pipe to make itself selectable.

When the queue has something, `handle_timeline` is called. It pulls the `Timeline` value out of the queue — a complete timeline constructed by the background thread — writes it to the database, and updates `@world`. The `@world` object is just a wrapper around the current timeline: it is the shell's mutable reference, the thing that changes as the world changes, holding a pointer to whichever immutable `Timeline` value is current.

Two properties of this design are worth noting explicitly.

First, the `Timeline` value crosses a thread boundary — constructed in the background thread, read in the main thread — without any locking. This is safe because the object is immutable. There is nothing to race on.

Second, the main thread is the sole writer to every external resource: the database, standard in, standard out. All IO is serialized for free. No mutex, no channel, no explicit synchronization — just a structural consequence of keeping side effects in one place.

* * *

## Testing

### Functional core tests need no test doubles

All functional core code is tested in isolation. Because every method is data in, data out, tests have a uniform shape: construct inputs, call the function, assert on the output.

A `Timeline` test takes a timeline, calls the update method with some new tweets, and asserts that the returned timeline contains the expected tweets — one assertion, one line. If the class were mutable, the test would need three steps: construct the object, call a mutating method, then separately inspect state. Immutability cuts the test to one.

Cursor tests follow the same pattern. The slicing behavior — which tweets are visible given a selected tweet — is tested with three assertions: starting at tweet one returns tweets one and two; starting at tweet two returns only tweet two; starting at a tweet not in the list returns an index error. Each is a single line. Explicit input, explicit expected output, no shared state between them.

`TweetRenderer` tests construct two tweets, pass them to the renderer, and assert on the returned lines. The renderer deals with colored text objects rather than plain strings — more complex than plain text — but the test structure is unchanged. There is no need to trace collaborators because there are none. There is no need to reason about mocking or stubbing because there just is not any.

### No mocks, no stubs

Searching the entire source for the word "mock" yields zero results. Searching for "stub" yields exactly one, in the first tests written for the cursor class — Bernhardt stubbed a tweet instead of using the real `Tweet` value object. Replacing that stub with a real `Tweet` would leave the entire system tested in isolation without a single test double. The functional core makes test doubles structurally unnecessary: pure values replace collaborators.

### The imperative shell is not tested at all

If everything is tested in isolation without mocks, what tests cover the imperative shell — the part that touches the network, the database, and the terminal?

None. The shell's 165 lines contain no tests whatsoever.

Two reasons:

**Few conditionals.** The shell is mostly flat sequences of method calls: event loops with no branching, method bodies that do one thing and return. The key-dispatch block is the one place with enough branching to be worth worrying about, and it is exercised every time the program is used. Critically, this is not a principle of never testing the shell — it is a contingent judgment. If that block became more complex, Bernhardt says he would become afraid, and he would want tests. The shell is untested because it is currently simple, not because it is untestable in principle.

**It is the exploration zone.** The shell is where new libraries get discovered and application behavior gets figured out. Bernhardt often does not know what he wants before writing the code. He writes rough code in the shell, figures it out, and once a decent design emerges, moves the code into unit-tested functional classes via TDD. The shell then has its old code deleted. Committing to tests before the design is known is wasteful; the solution is to treat the shell as a temporary holding area, not a place where finished code lives.

* * *

## The Refactor Cadence

The git history shows the cycle concretely.

A commit titled "add the TweetRenderer class" adds a new spec file and a new production file, but changes nothing else in the system. The system already had tweet-drawing behavior at the time — the new class was a TDD re-implementation of existing code, written with the old code as reference. Because TDD applies design pressure even when the behavior is already known, the new class ended up very similar to the old but with slightly different interfaces.

A subsequent commit, "replace untested TweetDrawing with TweetRenderer," makes the swap. It adds a small hook in the screen class, and then removes a large block from the shell — a wall of red in the diff. All of the old drawing code disappears. The unit-tested functional class takes its place.

This is the rhythm of the entire project. The shell bloats toward 250 lines as new behavior is explored and rough code accumulates. Then a functional class is extracted — usually via TDD — and everything it replaces in the shell is deleted. The shell shrinks back down. The functional core grows.

This is not traditional TDD: it does not test-drive the shell. But every extraction uses TDD, so the functional code that results from each cycle arrives with full test coverage.

* * *

## The Principle

Transitioning an existing program entirely into this architecture — one functional core, one thin imperative shell — is rarely straightforward. The more practical starting point is to identify **where the mutation is happening** and localize it. Push behavior that does not need to mutate state into immutable objects. Push mutation and side effects outward until they are as thin and structurally simple as possible.

When the separation holds, the core can be tested as pure functions with no test infrastructure. The shell stays thin enough that its few conditionals can be trusted through use. The system as a whole achieves high test confidence without mocks.
