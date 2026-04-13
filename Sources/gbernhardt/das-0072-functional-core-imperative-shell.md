# Functional Core, Imperative Shell

**Gary Bernhardt — Destroy All Software, Screencasts for Serious Developers**
**Episode DAS-0072**

_Transcript generated from video. Gary Bernhardt demonstrates functional core / imperative shell using his Ruby Twitter client._

---

The topic this week is the functional core of an application versus the imperative shell. We're going to use my Twitter client once again as an example because it is constructed in this way.

---

## The Functional Core

To start us off, let's look at the simplest functional class in the system, the tweet. We saw this class last time. It has an ID, a username, and a text. And it has two very simple methods that just return values, no mutation going on.

Moving on to something a little bit more complex, there is a timeline class that stores an array of tweets. And it knows how to update itself given some new tweets coming in from the API. But of course, when I say "update itself," I don't mean that it changes its own state. I mean that it constructs a new timeline with the new tweets plus the old tweets.

Moving on to something with a little bit more complex behavior, there is a cursor class. And its job is to remember where the cursor is on the screen and move it up and down. And it acts a lot like the cursor in Vim moving up and down. This class is quite a bit larger — it's almost 100 lines. I think it's actually too big and its responsibilities are sort of muddy and I'm not happy with it. But I am happy with the fact that it is immutable. It has an array of tweets and then a selection in those tweets. And it has methods that construct new cursors when we move up or move down in the tweet list. And they do not change any state on this object or any other object.

Finally, as our last example of the functional core, there is a tweet renderer. And we built a very simple version of this in the previous screencast. But this one, as you can see, is quite a bit longer — about 50 lines of code. It takes in an array of tweets, a selected tweet, and a height of the screen. And it ultimately ends up returning an array of lines of text that the screen class should draw.

These are all examples of the functional core. We do things like storing tweets, storing timelines, merging timelines, moving the cursor up and down (or at least remembering where it is), and rendering tweets onto the screen. But all of those things are functional. The only mutation that's ever done in any of these is to rebind local variables. Like you see right here — we have a `lines` variable, and then we potentially rebind it later. But that's not visible from the outside. So it doesn't affect the fact that these objects are immutable from an outside perspective.

---

## The Imperative Shell

If we have all of this functional code and nothing is actually being mutated, then how do we do things like handle user input, talk to the network, update the local database, all that kind of stuff?

Well, that's all done by the imperative shell, which is encapsulated in a single file — mostly at least — about 160 lines long. And this is the file that does things like read standard in, or handle updates from the network, or update references to those immutable objects as the state of the world changes.

Let's go back to the example of moving the cursor up and down on the screen. You can see that happening right here. If I hit `j`, it moves down. If I hit `k`, it moves up. But in both cases, we're just calling a method on the cursor that constructs a new one, and then assigning it to this instance variable.

Right above this method, you see another example where we construct an array of lines by using the tweet renderer, and then we render them onto the screen, which is a destructive action. We covered this last time, so I'm not going to go into any detail about it.

But I will go into detail about a third example, which involves pulling down new tweets from the Twitter API. As the application is starting up, this `run` method is the first one that runs. It does some stuff we're going to gloss over, but one of the things it does is to start the timeline stream.

This is a background thread that pulls new tweets down from Twitter and feeds them into the main thread via a queue. If we jump down to the `start_timeline_stream` method, you can see that it constructs a timeline stream, and then it starts a thread that just runs an infinite loop inside of that object.

Jumping down to the class, that method is sort of unsurprising. It has an infinite loop. It pulls down new tweets and constructs a new timeline, and then it throws that timeline into the queue.

Now, this does mean that that timeline object is shared between two different threads, but I have no fear of a race condition or of mutating it in the wrong way because the object is immutable — so there's just no danger there.

Once we put that timeline into the queue, the main thread is going to wake up and pull it out. So let's find where that happens — and it is a call to the `select` system call on both standard in (so we might be reading in input from the user) or we might be reading in new stuff in the queue. This is an instance of a special queue class that I have that is usable in a `select` statement because it has a fake IO object inside.

Once we find an event with this `select`, if it is a timeline event, then we call `handle_timeline`, which is the interesting part of this. And scrolling down, we pull the new timeline object out of the queue — so this is a complete timeline fed to us by the background thread. We then write it to the database, and then we update the world. The world is really just a wrapper object for a timeline.

The first interesting part of this is that, as I mentioned before, the timeline is constructed in the background thread. It's read here, but there's no risk of race because it's immutable. The second interesting part is the database update. Just like the standard in manipulation, the standard out manipulation, the main thread is the only one who touches the database, so I have no fears about multiple threads racing on the database, or racing on standard in or on standard out. All of these things are effectively synchronized for free because only the main thread touches them.

So that is the imperative shell that marries the state of the world to the state of the program, as well as storing that state. And we also saw the functional core, which is the bulk of the program, containing all of the logic and the data manipulation.

---

## Testing Implications

Now let's talk about the implications for testing, because I care a lot about testing.

First of all, all of that functional core code is tested in isolation. So if we look at the timeline class — that's the simplest one — it has very simple tests that take a timeline, add a new tweet to it, and then assert that the right tweets came out. This is a single line of code that reads very easily. Whereas if this were doing mutation, then we would have to construct a timeline, destructively push new tweets into it, and then assert that its state changed. So that would probably be three different lines of code.

Of course, timeline is very simple, so let's look at something more complex, like cursor. The cursor tests are actually very similar. There are a lot more lines, and they're a little bit more dense, but if we look at this slicing behavior — this is where we want to give the cursor a given tweet and return everything starting at that tweet and later (this is how it finds out which tweets are visible) — if we have a cursor that has two tweets inside of it and we tell it to start at tweet one, then we should get tweet one, tweet two. If we tell it to start at tweet two, we should only get tweet two. And if we tell it to start at some third tweet that isn't actually in the list, then we should get an index error. So we have three different assertions here in this test. Each of them is a single line, and each of them is very simple — explicit input and output.

Finally, let's move on to the tweet renderer class. Last time we TDD'd this and our tests were very simple. The actual tests are a little bit more complex because it has more complex behavior — like, for example, it renders multiple tweets, and the thing that it renders is not plain text, but these text objects that have color information encoded inside of them. But even with those complexities, you can see this test is very simple. It constructs two tweets, it throws them into the tweet renderer, and it asserts that it gets the right lines out. There's very little complexity here. There's no need to track down other classes because it's all data in, data out, and there's no need to reason about mocking or stubbing because there just isn't any.

### No mocks, no stubs

Speaking of a lack of mocking and stubbing, if we search the source code for the word "mock," it doesn't show up. And if we search for "stub," it does show up in exactly one place, but that's because for some reason I stubbed instead of using the tweet value object. I have the feeling that this was the very first test or pair of tests that I wrote for this class, but I could easily replace those with an actual tweet value object. And then this entire system would be tested without a single test double, even though almost all of it is tested in isolation.

### Testing the imperative shell — or not

Of course, this raises a pretty obvious question. If I'm testing everything in isolation and I'm not using mocks or stubs, then how am I testing the imperative shell around the program, where all of these references are being updated, it's touching the network and standard in, standard out, all this stuff?

Well, the answer may surprise you, but this file is **not tested at all**. All 165 lines of this — the entire outer shell of the program — contains no tests.

Now, there are a couple of reasons that I do this.

The first is sort of technical in nature. There are very few conditionals in this code. If we start scrolling down, you see a lot of flat methods, flat methods. There is a loop here, but there are very few conditionals. Here's one, but both of these have to happen for any tweets to get into the system, so I'm not really afraid of those breaking. And if I keep scrolling down, we have once again a loop but no conditionals. Here's a conditional, but this is easy enough to test — I use all of these keys all the time. If this became more complex, then I would become afraid and I would really want to test this. But other than this block of code, there's nothing in here that I'm really afraid of going out of sync.

The second reason that I don't test this file is that I tend to just cowboy code into here — either to explore a new library that I don't know (I didn't know any of the Twitter libraries before I started this), or to explore just how the application should work. I often don't know what I want, so I write some pretty nasty code in here, figure it out, and once I know what it's going to do (or I at least have a decent design), then I will move the source code into unit-tested files, often by doing TDD.

---

## The Refactor Cadence

To make that a bit more concrete, let's look at a piece of the Git history. You can see right here: "add the TweetRenderer class." And if we look at this commit, it adds a new spec file, it adds a new production code file, but it changes nothing else. So this new code was not used by the actual system. When this code was written, the system did already have this behavior, and I was re-implementing it. I was doing TDD, but I also was referring to the existing code, so they did end up being very similar, despite having slightly different interfaces.

Now, if we look at the history again, after I add the TweetRenderer, there's a commit that says: "replace untested TweetDrawing with TweetRenderer." And if we look at that commit, it adds one little piece to the screen class (which is not very interesting), but scrolling down, you see a lot of red — and this is all in that main file in the imperative shell of the program. All of this code went away and was replaced with that unit-tested TweetRenderer class.

This is the cadence of the entire project so far. I tend to bloat up this outer file up to something like 250 lines in some cases, and then I extract behavior out, moving it into unit-tested classes, figuring out how to make it more functional. So this is not by any means a traditional TDD flow — although in most cases, when I pull those classes out, I do do TDD there.

---

## Summary

We've now seen the functional core, the imperative shell, and the testing implications.

Obviously, it is not going to be easy to transition any given program into this style, where you have a single functional core and a single thin imperative shell wrapped around it. I think that the more important thing is to think about **where the mutation is happening** and try to localize it as much as possible.

I think that the ideal program contains a whole lot of immutable code in sort of functional style, and then small pieces of mutable code doing imperative things, but having them be very localized — separate from the data, and separate from the core behavior of the system.

I feel like I'm about to start preaching, so I think that's probably a good place to stop.

That is going to be it for the functional core, the imperative shell, and the testing implications of them. And I will see you in two weeks.
