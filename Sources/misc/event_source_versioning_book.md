# Versioning in an Event Sourced System

**Gregory Young**

---

## Table of Contents


1.  Status
    1.  Goals
    2.  Feedback
    3.  Status
2.  Why Version?
3.  Why can’t I update an event?
    1.  Immutability
    2.  Consumers
    3.  Audit
    4.  Worms
    5.  Crime
4.  Basic Type Based Versioning
    1.  Getting Started
    2.  Define a Version of an Event
    3.  Double Write
    4.  But…
5.  Weak Schema
    1.  Mapping
    2.  Wrapper
    3.  Overall Considerations
6.  Negotiation
    1.  Atom
    2.  Move Negotiation Forward
    3.  How to Translate
    4.  Use with Structural Data
7.  General Versioning Concerns
    1.  Versioning of Behaviour
    2.  Changing Semantic Meaning
    3.  Snapshots
    4.  Avoid ‘And’
8.  Whoops, I Did It Again
    1.  Accountants Use Pens
    2.  Types of Compensating Actions
    3.  How Do I Find What Needs Fixing?
    4.  More Complex Example
    5.  But I Really Screwed Up
9.  Copy and Replace
    1.  Simple Copy-Replace
    2.  Simple\>
    3.  With a System Live
    4.  Stream Boundaries are Wrong
    5.  Changing Stream Boundaries on a Running System
10. Cheating
    1.  Two Aggregates, One Stream
    2.  One Aggregate, Two Streams
    3.  Copy-Transform
    4.  Versioning Bankrupty
11. Internal vs External Models
    1.  External Integrations
    2.  Granularity
    3.  Rate of Change
    4.  How to Implement
    5.  Summary
12. Versioning Process Managers
    1.  Procees Managers
    2.  Basic Versioning
    3.  Upcasting State
    4.  Take Over
    5.  Warning

## Guide

1.  <a href="chap00.xhtml" class="bodymatter" epub:type="bodymatter">Begin Reading</a>


---


## Status

This is not actually part of the text. It is a place where a reader can track the status of the text being completed inside the text itself. This section will not be included in the final text.

Note there are some images used in the text that are not yet licensed. This is due to them possibly changing and will be licensed before the final edition.

- v10 5/4/17 Finish Process Managers. Various editing
- v8 14/2/17 Finish negotiation chapter. Various editing
- v7 9/2/17 Break apart general versioning. Add Behaviour and outside integration versioning, overall considerations, immutable intro
- v6 7/2/17 Add event bankruptcy, atom, other edits
- v5 24/1/17 Add cheating and bunches of editing
- v4 20/1/17 Weak Schema an Copy/Replace started. Some other areas added to + editing
- v3 18/1/17 small changes and release Whoops I did it again
- v2 17/1/17 some small changes and release Simple Type Versioning
- v1 17/1/17 release initial version including Why? and Why not update an event

### Goals

The goal of this text is not to discuss Event Sourcing but to discuss how to version Event Sourced systems. The main areas for discussion of versioning are:

- Why?
- Why Not Update and Event?
- Simple Type Based Versioning
- Weak Shema/Hybrid Schema
- Negotiation
- What to do with Mistaken Events
- Copy and Replace models
- Cheating
- Internal vs External
- Versioning of Process Managers

The length of the book is expected to be between 70-100 pages on this subject.

### Feedback

Feedback and/or questions on content you feel may not be covered well enough is highly appreciated! If a text has not gone through editing process yet grammar/spelling changes will probably be picked up in the editing process.

### Status

| Title                          | start | wr  | rel | img | edit |
|--------------------------------|-------|-----|-----|-----|------|
| Why?                           | x     | x   | x   |     | x    |
| Why Not Update an Event?       | x     | x   | x   |     | x    |
| Simple Type Based Versioning   | x     | x   | x   |     | x    |
| Weak Shema/Hybrid Schema       | x     | x   | x   |     |      |
| General Versioning             | x     | x   | x   |     |      |
| What to do with Mistaken Event | x     | x   | x   |     | x    |
| Copy an Replace Model          | x     | x   | x   |     |      |
| Negotiation                    | x     | x   | x   |     |      |
| Internal vs External           | x     | x   | x   |     |      |
| Process Managers               | x     | x   | x   |     |      |


---


## Why Version?

Have you ever done a big bang release? You schedule a maintenance window, carefully take down the system, release the new software, and… pray. I will never do a big bang release again. What scares me is not what happens if the release does not work, but what happens if it does not work and yet still appears to.

Most developers, even those working through big bang releases, have dealt with the versioning of software before. For example, they write SQL migration scripts to convert the old schema to the new one. But have you ever written SQL migration scripts to convert from the new schema with data added to it back to the old schema without losing information? What happens if, later, you get a catastrophic failure on the new release?

On teams I have worked with in the past, you got your choice: the fireman helmet or the cowboy hat. In other words, the fireman helmet means if you are working on a production issue, people know not bother you unless they have something material to your issue. I highly recommend this, as at least you get to keep a smile on your face while you are working live in production. To be fair, who wouldn’t want to wear a fireman helmet at work?

In modern systems, however, this is not acceptable. The days of maintenance windows and big bang releases are over. Instead, we now focus on releasing parts of software, and often end up wanting to run multiple versions of software side by side. We can no longer update every consumer when a producer is released. We are forced to deal with versioning issues. We are forced to deal with 24/7/365 operations.

There are many forces pushing us in this direction. We could talk about the move towards SOA/MicroServices, Continuous Deployment, core ideas from the business side, lossed opportunities and reputational risk.

Many of the strategies discussed apply to messaging systems in general. Event Sourced systems also face these kinds of versioning issues in a unique way, as they represent an append-only model. A projection, for instance, needs to be able to read an event that was written two years ago, even though use-cases may have changed.

What happens when you have mistakenly written an event in production? If we are to say an event is immutable and the model is append-only, many complications arise.

Over the years, I have met many developers who run into issues dealing with versioning, particularly in Event Sourced systems. This seems odd to me. As we will discuss, Event Sourced systems are in fact easier to version than structural data in most instances, as long as you know the patterns for how to version, where they apply, and the trade-offs between the options.


---


## Why can’t I update an event?

It continues to amaze me that, with every architectural or data style, there always comes a central question that defines it. When looking at document databases, the question that defines them is, “How do I write these two documents in a transaction?”

This question is actually quite reasonable from the perspective of someone whose career has been spent working with SQL databases, but, at the same time, it is completely unreasonable from the perspective of document databases. If you are trying to update multiple documents in a transaction, it most likely means that your model is wrong. Once you have a sharded system, attempting to update two documents in a transaction has many trade-offs in terms of transaction coordination. It affects everything.

Similarly, there is also such a question for those coming into Event Sourcing and dealing with an Event Store: “Why can’t I update an event?”

In dealing with support for EventStore (note in this text I will use Event Store as a general term and EventStore to discuss [GetEventStore](http://geteventstore.com)), we get this question often. The notion of not editing data is confusing for many people. Such systems are not, generally speaking, harder or easier to work with than other systems, but they are different.

Yet, however many times this question comes up, it is in fact the wrong one. While at face value it may seem obvious that you would want to be able to edit an existing event in your system, in reality, there are many circumstances where you should avoid it at all costs. “I put some bad data in this event, so I’m going to edit it quickly” may seem like a simple way to solve the issue. In many systems people have dealt with previously, an admin may simply log into Toad and issue an update statement against a table. In certain systems, however, this is a no-no.

### Immutability

Much is gained in terms of simplicity as well as scalability in log-based systems because the data is immutable. Many developers find the same thing when moving to functional programming; it takes a while to learn how to work in an immutable way because it is different. However, many problems go away with immutability.

It is an old joke that there are only two hard problems left in Computer Science

- Naming Things
- Cache Invalidation
- Off-by-one errors

How many times have you had bad data in a cache? Ever asked a user to delete their local cache? This issue can be a difficult one, often with complex hacky cache invalidation schemes put into place to try to mitigate the risk of cache invalidation issues.

Cache invalidation as an issue goes away when data is immutable. It is, by definition, infinitely cacheable. This allows systems that focus on immutable data to scale in many situations much more simply than other systems. It can allow for the use of off-the-shelf commoditized tools such as reverse proxies and cdns for scaleability, as well as geographic distribution.

And if you allow a single update… Well, your data is now definitely maybe immutable. Also known as mutable. A single update results in all data needing to be treated as mutable as there’s no knowing if or what may have been updated.

How do caches get invalidated on an update? In fact, all consumers, including, say, an in-memory domain object or a projection off to a SQL database, face this same issue.

### Consumers

If you edit an event, how will the consumers of that event be notified that the event has changed? As a concrete example, what if you had a projection that was interested in this event? It had received it previously, and had updated a column in a SQL database because of the event. If and when you changed the event, how would it be notified of the change that had occurred?

Other examples can be seen in other types of consumers. It is quite common to have a consumer/consumers listening to a stream that then takes action based upon the events. For instance, a consumer might decide to email a customer a welcome message based on a UserCreated event. Such an email would read something like:

> Dear Greg,

> Thank you for registering an account with our site. We are grateful for your business. If you need to manage your account, please visit http://oursite.com/accounts/gregyoung/manage

> Thanks,

> Your beloved business

If we were to go back and edit the username, should an additional email be sent? Is the URI in the original email still valid? How will we manage the process of generating a follow-up email to the user to notify them of the edited information?

Beyond all of this, what happens if you now replay a projection? It will come up with a different answer than the one it had previously, what everything that made a decision based on the original data, and while this can be tracked, it is not a trivial problem. This leads us to the next major issue.

### Audit

Your audit log is now suspect. You have no idea at any given point whether your projections to read models actually match up to what your events are.

A large number of Event Sourced systems are Event Sourced specifically because they need an audit trail. If you can edit your audit trail, is it actually an audit trail? Event Sourced systems address this need.

Over the years, I have dealt with many companies that, for legal reasons, had this requirement. Most, however, did not actually meet the legal requirements of being able to submit their audit log as evidence in a court case. The bar for being able to submit your audit log to a court of law is that you must be able to rebuild your current state from your audit log.

Consider for a moment that you disagree with your bank on your balance and they put forward a ledger that does not match the balance they say you should have nor the one you say you have. Would a juror accept this log as having anything to do with fact?

The moment you allow a single edit of an event, maintaining a proper audit log becomes impossible. Immutability is immutable. The moment you allow a single edit, everything becomes suspect.


> *Not this kind of worm*


### Worms

The gold standard for maintaining an audit log is to run on disks that do not allow editing, commonly known as WORM (Write Once Read Many) drives. When running on such a device, it is not possible to edit/rewrite a piece of information. This is by design.

The use of WORM drives is not only focused on the auditability of a system but also on its security. They are particularly useful in dealing with an attack vector known as a “super-user attack”, which is where we assume a rogue developer or system administrator is attacking our system and they have root access to the system.

Years ago, I worked at a company called Autotote (currently called Scientific Games) where such an attack happened. At the time, the company handled about 70% of the world’s paramutual betting market and made slot machines. This was obviously a system that would be of interest to an attacker, as it controlled all the money involved in wagers.

In the next cubicle over from me sat a gentleman named Chris Harn (Chris, if you are reading this, please drop me an email). Chris decided that the system was ripe for hacking. We ran a pool known as a pick-6, which is a generalization of a pick-N. The concept is that you pick the winner of N races, with a pick-6 requiring the winner to pick the winner of 6 consecutive races.

It is a very difficult bet to win and often punters would bet (and lose) hundreds of tickets on it. Given the volume of wagers on this type of pool, we would leave the wager at the remote track without shipping all of the details to the host track. At the end of the fourth race in the pick-6 example, the remote system would scan the tickets and then ship over the ones that could still possibly win. By the scan at the end of the fourth race, over 95% of the tickets had been removed.

Chris, however, had a different idea. What he would do is place a bet 1/2/3/4/ALL/ALL. He would then jump on the developer maintenance line, watch the first four races, and edit the bet on disk with a hex editor to be 3/7/9/2/ALL/ALL, or whatever happened to come in. The scan would occur at the end of the fourth race and would pick up his bet as a “possible winner” (while it is actually an assured winner). The system would then ship his bet over to the host system.

Chris got caught. I have always joked that, when dealing with super-user attacks, the culprit gets caught not because they are stupid but because they are unlucky. Chris was very unlucky.

This all happened during the 2002 Breeder’s Cup. For those unfamiliar with American horse race gambling, the Breeder’s Cup is the second largest racing day in America, and this was the largest scandal in American betting in the last century. They put through a bet 1/2/3/4/ALL/ALL (or similar likely with the favourites in each leg), then edited it as discussed. Smooth enough, but there was a problem. The Classic, a 43-1 horse, won the race, which made their tickets the only winning tickets, valued at a \$3,100,000 level. Normally, there would be 20 or 30 winners at \$100,000 each; instead, there was one winner at \$3,100,000. At that level, you can be sure the people running the pool will look into the winning ticket!

Needless to say, Chris was not in the office on Monday morning, though we did meet lots of new friends from the FBI. Meanwhile, there is a [Criminal Masterminds](http://www.imdb.com/title/tt0787557/) episode documenting the whole scandal. You can also read more about it on [Wikipedia](https://en.wikipedia.org/wiki/2002_Breeders'_Cup_betting_scandal).

This entire situation could have been avoided by using WORM drives. The ability to rewrite an event often opens up massive holes in security.

### Crime


Chris certainly placed himself on the wrong side of the law by editing an event. There are many other cases where editing an event will likely land you in prison.

In police information systems, financial/accounting systems, and even such simple things as border control, the editing of an event is not only considered morally wrong, it is likely a federal crime that comes with a prison sentence greater than five years for anyone found guilty of it. Whether we talk about fraud or embezzlement or discuss the possibility of moving a rogue agent into a state, there are a vast number of existing systems where the concept of editing an event is an impossibility.

Given all the reasons why we may not want to be able to edit an event, the question before us is: how do we handle an Event Sourced system, given such constraints? How can we run on a WORM drive while keeping a proper audit log and retain the ability to handle changes over time? How can we avoid “editing an event”?


---


## Basic Type Based Versioning

Throughout this chapter, we will look at how to version events in an Event Sourced system. We will, however, take on a few constraints that make the problem both more difficult and more unrealistic. We will assume that we are forced to use a serializer that has no support for versioning (such as a .NET or Java binary serializer, for example).

Even in such an unusual case, we can still handle versioning, but it may result in unusual conditions, depending on the circumstances.

In order to look through a concrete example, I will use the [SimpleCQRS example](http://github.com/gregoryyoung/m-r). In particular, we will focus on how to version a change to the Deactivate behavior.

### Getting Started

Most developers have dealt with versioning at some level before. For instance, when releasing a new version of software, it is common to upgrade a SQL database from one schema to another. When dealing with such changes, developers need to consider the differences between breaking and non-breaking changes.

- Adding a table
- Renaming a column
- Deleting a column
- Adding a view
- Deleting a stored procedure
- Adding a stored procedure

Normally, additive changes are non-breaking, though I once heard a funny story about a table named “foo” that put all the stored procedures into a special “debug mode”. We should strive, however, to make additive changes non-breaking.

As such, when dealing with versioning, we will insist on only having additive changes as opposed to destructive ones. We will always add a new version of an event. Accordingly, both destructive changes, such as the renaming of a property, as well as additive changes can be made when adding a new version of an event, since the addition of that new event still represents a change that is additive overall.

In the code example, we start with an event that looks like:

    public class InventoryItemDeactivated : Event {
        public readonly Guid Id;

        public InventoryItemDeactivated(Guid id)
        {
            Id = id;
        }
    }

The software change requested is our inclusion of a reason associated with why the user was deactivating the inventory item, and we will also rename the ID property “ItemId”. This is allowed as it is still an additive change, since it introduces a new version of the event and does not change the existing one.

To start, let’s add a new version of the event. Note that if we were to have started the system with this type of versioning in mind, we would have used the versioned naming format from the beginning.

    public class InventoryItemDeactivated_v1 : Event {
        public readonly Guid Id;

        public InventoryItemDeactivated_v1(Guid id)
        {
            Id = id;
        }
    }

    public class InventoryItemDeactivated_v2 : Event {
        public readonly Guid ItemId;
        public readonly string Reason;

        public InventoryItemDeactivated_v2(Guid id, string reason)
        {
            ItemId = id;
            Reason = reason;
        }
    }

There are now two versions of the event. Version 1, the original version, contains just an ID, while version 2 contains an ID (renamed as ItemId) and a reason why the user deactivated this item.

Next to edit is the domain model, which will change from something like:

    public void Deactivate()
    {
        if(!_activated) throw new InvalidOperationException("already deactivated");
        ApplyChange(new InventoryItemDeactivated_v1(_id));
    }

to

    public void Deactivate(string reason)
    {
        if(!_activated) throw new InvalidOperationException("already deactivated");
        ApplyChange(new InventoryItemDeactivated_v2(_id, reason));
    }

We will also add a new Apply method to the InventoryItem class to support the second version of the event.

    private void Apply(InventoryItemDeactivated_v2 e)
    {
        _activated = false;
    }

And we’re done. Our [aggregate](http://domainlanguage.com/ddd/reference/) will write out version 2 of the event in the future and supports replaying of version 1 and version 2 events.

However, there are still a few issues here. What happens when get to InventoryItemDeactivated_v17? What if there are five behaviors and we have now versioned all of them?

Method explosion. I have always joked in class that this is why the concept of code folding was added to editors.

### Define a Version of an Event

In our description of a new version of an event above, we left out an important detail.

A new version of an event must be convertible from the old version of the event. If not, it is not a new version of the event but rather a new event.

This rule is quite important. It also implies that any version of any event should be convertible from any version of a given event. If you find yourself trying to figure out how to convert your old event (Evil Kinevil jumping over a school bus on a motorcycle) to your new version of the event (a monkey eating a banana) and you can’t, this is because you have a new event and not a new version of it.

Provided we can always convert the old version of the event to the new version, we could upcast the version of the event as we read it from the Event Store. The event could be passed through some converters that move it forward in terms of version.

    InventoryItemDeactivated_v2 ConvertFrom(InventoryItemDeactivated_v1 e) {
        return new InventoryItemDeactivated_v2(e.Id, "Before we cared about reason.");
    }

The code will now upgrade the event to version 2 of InventoryItemDeactivated. As such, the code will no longer see an InventoryItemDeactivated_v1 event, and the handler in the domain object can be removed. You may notice we have to pass a default for a reason; this is the same decision that gets made when adding a non-nullable column to a table in a SQL database where there are existing rows.

There are, however, additional and more sinister problems with this style of versioning.


> *Consumer Problem*


What happens when we have a [Blue-Green](https://martinfowler.com/bliki/BlueGreenDeployment.html) or similar deployment where there are multiple concurrent versions of the software running side by side? If we upgrade node A to be version 2 here and node B is still on version 1, can node B read from the event stream that node A has written a version 2 event? Unfortunately, using this style of serialization requires every node to actually have the type for the event, as the type describes the schema of the event. If a consumer does not have the type, it will not be able to even deserialize that event. One might try to move around types of events, or we could move away from using this type system as our schema.

This applies to all consumers: projections, other version nodes, and downstream consumers.


> *Schema Server*


There are some implementations that exist that try to move the schema resolution to a central service. This can have pros and cons. On the one hand, it can remove some of the updating of consumer versioning issues, as they will request the schema on demand from a centralized service and apply it dynamically. On the other hand, it can introduce significant complexity, need for a framework, and a possible availability problem to the system. I have not worked enough with production systems doing this to give a proper weighing of these concerns, though, and tend to lean towards other mechanisms at the time of writing.

### Double Write

There is a common way in which people try to avoid the issues that arise from having old subscribers that do not understand the new version of the event. The idea is to have the new version of the producer write both the \_v1 and the \_v2 versions of the event when it writes. This is also known as Double Publish.


> *Double Write*


As an example, when releasing version 2 in this scenario, Version 2 will write both version 1 and version 2 of the deactivated event to the stream. When the other node running version 1 reads from the stream, it will ignore the version 2 instance of the event, which it does not understand, and will process version 1 of the event, which it does.

A rule must be followed in order to make this work. At any given point in time, you must only handle the version of the event you understand and ignore all others.

Over time, the old version of the event will be deprecated. This gives subscribers time to catch up.

This methodology is not unusual in a distributed system, where works reasonably well. It does not, however, work that well in an Event Sourced system. In an Event Sourced system, it will work fine on the stable situation, but will fail when you need to replay a projection.

When replaying a projection that currently supports version 3 of the event, what do we do when we receive a version 2 of the event that was written before the version 3 existed? A matching version 3 of the event will not be received if we ignore the version 2. To complicate matters further, there are some places where a version 2 will be written with a matching version 3 event and others where it will not be. This is due to the running of version 2 and version 3 side by side. These same issues exist when trying to hydrate state off of the event stream.

Remember that state hydrated from a stream to validate a command is a projection.

There are some ways to work around these issues. For instance, when receiving a version 2, you could look ahead to see if the next event is a version 3. If so, ignore it or upconvert it to a version 3. These mechanisms, however, quickly become quite complex.

Another option here is to make absolutely sure that everything is idempotent (including your domain objects, which can be a bit weird) and write the multiple versions with the same ID. This is feasible, but can be very costly. If you insist on going down this road, the easiest solution is to make sure the producer puts out all the versions with the same ID. Reading forward, look to the next event and to see if it has the same id, and processing if not. This is the least hacky way of handling it.

Although this mechanism works well under some circumstances for event-based systems, it is not recommended for Event Sourced systems due to the issues it introduces when replaying a projection.

### But…

You should generally avoid versioning your system via types in this way. Using types to specify schema leads to the inevitable conclusion that all consumers must be updated to understand the schema before a producer is. While this may seem reasonable when you have three consumers, this completely falls apart when you have 300.

Accepting the constraint that our serializer has absolutely no versioning support seems an odd place to start when almost all serializers have some level of versioning support (json, xml, protobufs, etc.). Accordingly, let’s try removing this constraint.


---


## Weak Schema

The serialization formats discussed up until now operate based on what is known as strong schema. In the type example for the .NET and binary serializer, the schema is stored in the type. This is also known as out-of-band schema, meaning that the schema is not included with the message but is held outside it.

The problem with strong schema, especially when using out-of-band schema such as types, is that without the schema you will not be able to deserialize a given message. This leads to the previously described problem of needing to update a consumer before updating a producer, which is unacceptable in many situations, as the consumer will not be able to deserialize the message otherwise.

### Mapping

Most systems today do not use this method of serialization for exactly these reasons. Instead, they will use something like json or xml, combined with what is known as weak-schema or hybrid-schema, to serialize their messages. While this entails more rules that must be followed, it also offers more flexibility, providing the rules are followed.

Some incorrectly call it self-describing. In fact, it is [semi-structured](https://en.wikipedia.org/wiki/Semi-structured_data)

    {
        foo : 'hello',
        bar : 3.9
    }

When handed this, without being told anything more than it is json, it can be parsed and seen as containing a key, foo, with a string value of hello and a key, bar, with a numeric value of “3.9”. We can use this property to our advantage in terms of dealing with the versioning of messages.

What if instead of deserializing to a type it was instead mapped to? The rules for mapping are simple. When mapping, you look at the json and at the instance.

- Exists on json and instance -\> value from json
- Exists on json but not on instance -\> NOP
- Exists on instance but not in json -\> default value

So, given json

    {
        number : 3.9,
        str : 'hello',
        other : 15
    }

when mapped to a type

    class Foo {
        public decimal Number;
        public string Str;
    }

it would produce an output of Foo { Number=3.9, str=”hello”}. If, however, it were mapped to a type such as

    class Foo {
        public decimal Foo;
        public string Str;
    }

it would produce an output of Foo {Foo=0.0, Str=”hello”}.

Such a mapping system is no more than an hour of work, even with a nice fluent interface for handling defaults. However, this concept is especially powerful when dealing with versioning of messages.

In the previous chapter, adding or modifying something on an event required the addition of a new version of the event via adding a new type.

        public class InventoryItemDeactivated_v1 : Event {
            public readonly Guid Id;

            public InventoryItemDeactivated_v1(Guid id)
            {
                Id = id;
            }
        }

        public class InventoryItemDeactivated_v2 : Event {
            public readonly Guid ItemId;
            public readonly string Reason;

            public InventoryItemDeactivated_v2(Guid id, string reason)
            {
                ItemId = id;
                Reason = reason;
            }
        }

When using mapping, there is no longer an addition of a new version of the event. Instead, you just edit the event already in place.

        public class InventoryItemDeactivated : Event {
            public readonly Guid Id;
            public readonly sting Reason;

            public InventoryItemDeactivated(Guid id, string reason)
            {
                Id = id;
            }
        }

The mapping handles the rest. If you have an InventoryItemDeactivated in the first version and you map it to one expecting the second version, it will still work, but Reason will be set to a default value.

If a producer were to begin putting out the second version, a consumer that understood the first version would continue working, seeing as it could map it, and Reason would just be ignored. There are, however, two factors that must be remembered here.

The first is that you are no longer allowed to rename something. If the change were from Id to ItemId, it would not work, as the first version would no longer receive Id. You can get around this by supporting both Id and ItemId, but this can quickly become annoying, especially with an Event Sourced system, where you cannot just deprecate it but must carry it forward into the future.

You can deprecate it, but it is annoying operation as well (see the Copy and Replace chapter for more on this).

The second is that there will often be programmatic checks to ensure what you expect to be in the message after the mapping is in fact present. While it may be expected that Id is still there, this may not be a given. As opposed to specifying each of these checks, it is common to support in the mapper.

    Map.To()
        .WhenAbsent("Reason", "no reason at all")
        .WhenAbsentError("Id")

While these types of validations can be defined in code, it might also be worth looking at defining them either through json schema or xsd.

Another option is to use hybrid-schema, where some things are required and some things are not, the latter being treated as being mapped. Protobufs is an example of a format that supports this. The general rule of thumb when working with hybrid-schema is to make the things without which the event would make absolutely no sense a requirement while leaving everything else as optional. In the example of InventoryItemDeactivated, it would make no sense if the Id of the item were not included (It was deactivated, but I won’t tell you which one!). The Reason, however, would remain optional. If no Id was present, a deserialization error would occur, while Reason being absent would just result in a null as the value (and it being set as not present).

### Wrapper

Less common than mapping to a type, another option with events is to write/generate a wrapper over the serialized form. While this tends to be more work, and often requires some additional rules to be followed to deal with schema evolution, it can also have some significant advantages.

For example, for the InventoryItemDeactivated event, instead of mapping the data to a new object, a wrapper could be created to wrap the underlying json.

    public class InventoryItemDeactivatedWrapper {
        public Guid Id {
            get {return Guid.Parse(_json["id"]);}
            set {_json["id"] = value.ToString();}
        }

        public string Reason {
            get {return _json["reason"];}
            set {_json["reason"] = value;}
        }

        public InventoryItemDeactivatedWrapper(string json) {
            _json = JObject.Parse(json);
        }
    }

The class is then not directly mapping to and from the json; it is instead wrapping it and providing typed access to it. Certain serializers, flatbuffers, for instance, will even generate the wrappers from schema definitions to avoid the overhead of manually creating this type of code.

One of the main advantages of writing a wrapper as opposed to a mapping is that, if you are enriching something, you can move it forward without understanding everything in the message. This “ility” is normally more important in a document-style messaging, but it can also be useful in some circumstances with events (for an example, see internal vs external models). Consider the InventoryItemDeactivated event, but where another consumer had a different version that included a userId.

    {
        id : "24570fb7-a26c-42e8-a9cd-f6bf8060790e",
        reason : "none at all",
        userId : 42

    }

If this json were mapped to a Deactivated event containing Id, Reason, and AuthCode and then serialized again, the userId would be lost in the mapping process. By using a wrapper, the userId can be preserved even if it is not understood.

Although the pseudo-code above is not memory-friendly (creating, as it does, an entire tree of JObjects representing the json), other serializers can generate wrappers that operate in place over a byte \[\] or other such structure without requiring any object creation. Some can even parse on demand as opposed to deserializing the entire buffer. This can become very important in some high-performance systems, as they can operate with zero allocations.

It is important to note that with many formats, including, for example, flatbuffers, you can only evolve schema by appending to the end of the schema, which introduces a whole new set of possible versioning issues.

### Overall Considerations

For most systems, a simple human-readable format such as json is fine for handling messages. Most production systems use mapping with either XML or json. Mapping with weak-schema will also remove many of the versioning problems associated with type-based strong schema discussed in the previous chapter.

There are scenarios where, either for performance reasons or in order to enable data to “pass through” without being understood, a wrapper may be a better solution than a mapping, but it will require more code. This methodology is, however, more common with messages, where an enrichment process is occurring, such as in a document-based system.

Overall though, there are very few reasons why it may be considered for the type-based mechanism. Once brought out of your immediate system and put into a serialized form, the type system does very little. The upcasting system is often difficult to maintain amongst consumers; even with a centralized schema repository, it requires another service. In most scenarios where messages are being passed to/from disparate services, late-binding tends to be a better option.


---


## Negotiation

All methods for handling versioning covered thus far have assumed single direction traffic from a producer to N consumers. This is a good model to consider for messaging systems, but it also comes with some limitations, such as the need for every consumer to understand what the producer publishes.

In the weak-schema and basic versioning chapters, we looked at strategies for making changes to messages in a versioning-friendly way. In particular, with weak and hybrid schema, the focus was on making changes such that they were version-insensitive for consumers.

There are other models available that do not make this assumption. Some transports support bi-directional communications, which can be used to negotiate what format a message will arrive in. The most common of such transports in use is HTTP.

### Atom

[Atom feeds](https://tools.ietf.org/html/rfc4287), though designed for blogs, are a common distribution method for event streams. The Atom protocol has been around for a long time and is supported by a huge number of things on the Internet; Chrome even has a custom Atom viewer built in. Overall, the protocol is relatively simple, but it can be a great tool when dealing with versioning.

When going to an Atom feed, a document is returned with the latest events, as well as rel links that allow paging through the feed.

      "links": [
        {
          "uri": "http://127.0.0.1:2113/streams/newstream",
          "relation": "self"
        },
        {
          "uri": "http://127.0.0.1:2113/streams/newstream/head/backward/20",
          "relation": "first"
        },
        {
          "uri": "http://127.0.0.1:2113/streams/newstream/3/forward/20",
          "relation": "previous"
        },
        {
          "uri": "http://127.0.0.1:2113/streams/newstream/metadata",
          "relation": "metadata"
        }
      ]

> Note: The json version of a feed being used here, as it is easier on the eyes than an XML feed.

These links are defined in [RFC 5005](https://tools.ietf.org/html/rfc5005#page-4) and allow for the paging of a larger feed via a set of smaller documents. As an example, to read the entire feed, use the last link then follow the previous link until the previous link disappears (at the head of the feed). This allows a subscriber to navigate through the overall feed.

Each page of the feed is then a document that contains links back to documents that describe the individual events.

        {
          "title": "1@newstream",
          "id": "http://127.0.0.1:2113/streams/newstream/1",
          "updated": "2017-02-06T15:23:24.30135Z",
          "author": {
            "name": "EventStore"
          },
          "summary": "SomeEvent",
          "links": [
            {
              "uri": "http://127.0.0.1:2113/streams/newstream/1",
              "relation": "edit"
            },
            {
              "uri": "http://127.0.0.1:2113/streams/newstream/1",
              "relation": "alternate"
            }
          ]
        },
        {
          "title": "0@newstream",
          "id": "http://127.0.0.1:2113/streams/newstream/0",
          "updated": "2017-02-06T15:23:20.986492Z",
          "author": {
            "name": "EventStore"
          },
          "summary": "SomeEvent",
          "links": [
            {
              "uri": "http://127.0.0.1:2113/streams/newstream/0",
              "relation": "edit"
            },
            {
              "uri": "http://127.0.0.1:2113/streams/newstream/0",
              "relation": "alternate"
            }
          ]

The alternate rel link points back to an individual event. A client would then get http://127.0.0.1:2113/streams/newstream/0 and http://127.0.0.1:2113/streams/newstream/1 in order to get the events that make up the feed. As new events arrive in the feed, new pages and URIs for events will be created. This allows a client to follow the feed and continually receive events as they occur in the system.

An interesting side benefit worth noting here is that the URIs and the pages are both immutable. As such, they can be infinitely cached, allowing systems to benefit from heavy use of off-the-shelf caching to enable better scalability and geographic distribution.

From a versioning perspective, the most important feature is that the client is requesting the event from the URI. As the client is making this request, it is also free to include an “Accept: something” header with it. This can tell the backend what format the client understands so it can provide it in an appropriate format. Client A may specify it understands vnd.company.com+event_v3 and Client B may specify it understands vnd.company.com+event_v5. Each client could then receive the event in the format it understands.

While HTTP in general and Atom in particular are not the only protocols to support such an interaction, they are, by far, the most popular. A similar protocol could, however, be implemented over nearly any transport, even message queues such as [rabbitmq](https://www.rabbitmq.com/).

In such a case, the key to the protocol is that the payload is not sent via the main feed; a URI or some other accessor is sent instead, following which the client determines how it would like to receive the payload. As part of this process, the client can determine which version of the payload it can understand and ensure it gets that one. This is implemented by sending a message over the transport such as:

    MessageAvailable {
        id : '0dd5976d-ad99-4333-b4e1-6b3b87af2f8f',
        type : 'fooOccurred'
        uri : 'http://mycompany.com/messages/0dd5976d-ad99-4333-b4e1-6b3b87af2f8f'
    }

Upon receipt, the client could then do a HTTP GET to the provided URI and content-type negotiate the message. Note that there is a unique URI for each message. Often, this is implemented by then converting the message; however, it could also just be that the producer has prepopulated a few versions into a key/value store and provides multiple URIs as rel links in the message to let the clients know where the messages are. For example:





Custom media types can also be used.

An obvious issue that arises with this style of interchanging messages is in how the flow of a message then works. An announcement message is sent over the queue or previously over the Atom feed. Upon receipt of the announcement, the client must then fetch the message in the format they prefer. While quite flexible, this method is less than optimal in terms of performance, as it requires a request per message.

### Move Negotiation Forward

An astute reader will notice that, for most systems, a consumer that wants version 2 of a message will for a long period of time continue to want version 2 of the message. Handling the content-type negotiation on a per-message basis will be wasteful, as it will want the same version every time for a long period of time.

Some message transports and middleware support the concept of moving content-type negotations forward. A consumer will register the content types it is interested in ahead of time and the publisher will then push the events in one of the pre-determined formats, as opposed to sending an announcement and requiring the consumer to negotiate per message.

While many transports support this, it can also be implemented relatively easily on top of any existing transport, providing a control channel is also provided (assuming an unidirectional transport, such as a message queue). In such a case, the consumer sends all of the content types it is capable of receiving beforehand. The publisher then converts a given message to the best match content type on its end before sending it to the consumer.

There are a few benefits here. Obviously, not needing to negotiate per message when changes only happen rarely, often on the order of months, has some performance advantages. Another less looked at benefit is that, if using type-based schema, only the translation code needs to be updated. Consumers can be updated as needed and can still get the old version of the message they understand. This is particularly valuable when there are a large number of consumers often managed by a large number of teams. If a new field is to be added to an event, only the teams that need to use it need support for it.

In practice, for all content-type negotiation there is also often a depreciation period, where very old versions of messages gradually cease to be supported.

Although it requires more work and is slightly less flexible than negotiation per request, moving negotiation forward can be useful in a large number of scenarios. The main circumstance in which you might consider using it is if you need to push more than roughly 100 messages per second to your consumers. When pushing larger quantities of messages, the ability to not negotiate on a per-message basis significantly improves performance.

### How to Translate

For both per message content-type negotiation and moving the negotiation forward, the producer will need to translate between versions of the message. The translations can happen either using type-based schema or weak schema.

The code will receive a message of version Vx. It will then need to translate that message to the version that the consumer wants. This process is very similar to the upcasting of events using type-based schema that was looked at in the basic versioning chapter.

    InventoryItemDeactivated_v2 ConvertFrom(InventoryItemDeactivated_v1 e) {
        return new InventoryItemDeactivated_v2(e.Id, "Before we cared about reason.");
    }

One major difference between the upcasting seen with type-based schema is it is often also needed to downcast. Given an event on version 12, it is common to need to convert it to version 9 for a given consumer. In such a case, there are two methods: to convert from v1 to v2, and to convert from v2 to v1.

    InventoryItemDeactivated_v2 ConvertFrom(InventoryItemDeactivated_v1 e) {
        return new InventoryItemDeactivated_v2(e.Id, "Before we cared about reason.");
    }


    InventoryItemDeactivated_v1 ConvertFrom(InventoryItemDeactivated_v2 e) {
        return new InventoryItemDeactivated_v1(e.Id);
    }

These methods then get registered, allowing the conversion from one version to any other. Occasionally, it can be preferred if there are many versions of events to allow skipping methods â€” for example, converting from v1 to v9 directly as opposed to performing nine conversions.

Weak schema follows a similar pattern, though it can often do it in a single step as opposed to converting each version, as more compatibility rules are to be adhered to with weak schema.

Overall, the conversion between versions of events is relatively straightforward code, though it can quickly become large with a large number of events and a large number of versions. If you find you are encountering a bit of a code explosion, it might be worth considering depreciating old versions of events.

### Use with Structural Data

Thus far, the discussion has revolved around the versioning of messages. The same pattern, however, can be used for structural data. It is quite common in integration scenarios to have a large number of consumers that are interested in getting an updated structural view of data as it changes.

A perfect example of this came up a few years ago when I was looking at an integration scenario at a financial institution. The system in question exists in almost every financial organization and is known as “security master”. Security master holds all the metadata associated with [CUSIPs](https://www.sec.gov/answers/cusip.htm) (unique identifiers of securities). Almost everything at the institution needs access to this metadata. For each security, it contains information such as market cap, links to other securities, and whether or not it is tradeable. It is common for a financial institution to have hundreds or even thousands of integrations to security master.

Before discussing how a similar pattern to the message versioning was used here, it is worth discussing how the existing system was implemented. The company was largely focused on Oracle databases. The system for managing the data was a relatively simple CRUD application that updated the data in a few tables in the Oracle database. Consumers, for the most part, were interested in one of three views. They either wanted summary information, a detailed view, or raw data. The tables in the master system were then replicated to all other databases, with the master system being the [single source](https://docs.oracle.com/cd/B28359_01/server.111/b28322/repsimpdemo.htm#CJFGJAIH). As such, these tables would live in the database of every consumer. If you wanted access to the information, you would open a ticket with IT and they would set up a new replication to your database, with the tables then updating accordingly.

While this replication scheme seems reasonable with a small-scale intergration, remember there were more than 1000 integrations done this way! A new design was thought up that operated very similarly to the message negotiation system above. Changes to security master were written as events. Consumers all received an Atom feed representing the changes that had occurred. As in the message system, these links pointed back to a small service that would handle the requests.

//TODO insert diagram

Instead of converting the messages to different views, this service would instead replay all events up to a given event. An example URI might be http://bank.com/securities/AAPL/5, in which AAPL is the security and 5 is the event up to which you’ve requested information. This code looked very similar to any other projection and would just return the state as of event 5.

One added benefit of this URI scheme is that, given the immutability of URIs, AAPL at event 5 will always be the same, allowing for caching. Another is that consumers will get a state per change, thereby not always getting the latest value, which can miss a change if these are multiple.

It is important to note that the service responsible for generating the varying structural views of the state at a given point is at most a few hundred lines of code. It basically just has a few small in-memory projections to build the varying structural views and return them.

Upon receiving a new link via the Atom feed, consumers could then content-type negotiate what they wanted to receive. Most clients just needed summary information and could use vnd.bank.com.security_summary+json. Others who needed the detailed view could use vnd.bank.com.security_detail+json. On top of this, tickets were no longer required in order to get access to the information. The security metadata was not secured information, and so URIs containing the information could be published. You could get per security, per market, or all changes via Atom feeds. If you needed information on a new staging server, you could just point to the URI that matches what you need and you were off.

Where this method really shines, though, is when a change eventually happens. The United States government, for example, might decide to track new terrorism information. Very few systems will actually care about this, but some may. By introducing a new vnd.bank.com.security_detail_2+json, the systems that need access to this information can be updated while all of the other systems will continue working with the first version of the security detail without needing an update.

While this pattern is not directly related to versioning of Event Sourced systems, it is a useful application of the same pattern. It is also another style of integration for Event Sourced systems where structure is used instead of shipping events. It is a good tool to have in your toolbox.


---


## General Versioning Concerns

There are lots of general concerns around creating an managing events that are orthogonal to any form of serialization. Over the years I have seen many teams get into trouble as although they had a good grasp on the issues that can come up with varying forms of serialization and interchange, they made mistakes in other more general areas.

This chapter will focus on the other more general areas of versioning in an Event Sourced system. Many apply to messaging systems in general but some become more important when faced with the need to be able to replay the history of a system by replaying events. These areas apply whichever serialization mechanisms are chosen though many can be fixed after the fact using patterns found later in the book.

### Versioning of Behaviour

The versioning of behaviour is probably the most common question people new to Event Sourcing run into. Often teams will push forward on a project without having understood an appropriate answer to this question which can lead to interesting places. The question is typically phrased in similar terms to “How do I version behaviour of my system as my domain logic changes?”

The thought here is when hydrating state to validate a command logic in the domain can change over time. If replaying events to get to current state how do you ensure that the same logic that was used when producing the event is also used when applying the event to hydrate a piece of state?

The answer is: you don’t.

The basis of the question misses a lot. The quintessential example of such a case would be a tax calculation. You could when hydrating your piece of state use something like the following using a standard OO way of structuring the domain object.

    public void Apply(ItemSold e) {
        _tax = e.Subtotal * .18;
    }

This will work fine for handling a tax of 18%. This however is a time bomb just waiting to go off. What will happen when they decide that as of Jan 1 2017 the tax will be 19%? The easy answer would be to go back and change your method.

    public void Apply(ItemSold e) {
        _tax = e.Subtotal * .19;
    }

When a projection is replayed what will happen? Remember that the domain object is a projection the same as a read model it just happens to be building a piece of state used in the domain model. This issue affects all projections. The next time an event gets replayed through this it will take a transaction that occurred on June 17, 2016 and say the tax is 19% of the subtotal. Do you call up all your customers from previous years and ask them to send the difference that occurs in tax? The rule starts from Jan 1 2017, it does not apply to previous events.

What many people fall into the trap of is putting logic for this into their handler. To try to versionise busines logic.

    public void Apply(ItemSold e) {
        if(e.Date < Convert.ToDateTime("1/1/2017")) {
            _tax = e.Subtotal * .18;
        } else {
            _tax = e.Subtotal * .19;
        }
    }

The logic is now looking at the datetime of the event and deciding which business logic should be used. What about after five years? What about for seven rules? This code very quickly spirals out of control. It is important to remember as well that this code needs to be applied to *every* projection that wants to look at tax on this particular event.

In single-temporal systems this is usually a bad idea. In multi-temporal systems such as accounting which is bi-temporal (transactions do not necessarily apply at the time they are accepted) there are times where it is required to implement such things.

If you find yourself putting branching logic or calculation logic in a projection, especially if it is based on time, you are probably missing logic in the creation of that event.

Instead of doing this what should be done is to enrich the event with the result of the tax calculation at the time that it is being created. In other words put a property Tax on the event and when creating the event (in the command context, or input event context if events only).

    public void SellItem(...) {
        //snip
        var tax = subtotal * .18;
        Apply(new ItemSold(_id,....,tax));
    }

Later when the tax changes a new version of software can be released to support the new tax rate.

    public void SellItem(...) {
        //snip
        var tax = subtotal * .19;
        Apply(new ItemSold(_id,....,tax));
    }

Anything that wants to process the event now just looks at the Tax property of the event and will receive the appropriate value for tax without the need to recalculate it.

It may be worth including a description field even if only a string that describes the type of calculation made, see semantic meaning below.

Note that in multi-temporal systems it is sometimes a necessity to versionize domain logic and may be required to make calls to external systems as events are processed. This is due to the nature of an event not necessarily applying to now when it is created. I cannot as example always determine what a lookup value will be in the future and may need to ask in the future. When integrating with such systems it is best if those systems are temporal in nature as well.

#### Exterior Calls

Very closely related to versioning of domain logic is dealing with calls to external systems. Most developers figure out really quickly that calling out actions on external services in the context of any projection is a really bad idea. Try to imagine charging somebody’s credit card, now let’s replay that projection! External calls requesting information have a tendency of slipping through though.

If you have an external call getting information from another service and you later do a replay, will the projection result in the same output after? What if the data in the external service has changed?

Handling external calls is quite similar to the versioning of domain logic. Make the call at the time of the event creation and enrich the information returned from the call onto the event. This allows for deterministic replays as the projection (whether a domain object or a read model) self-contained again. All the information needed to build the state is located in the event stream. Every replay will come out the same.

    public void SellItem(...) {
        //snip
        var customerScore = _externalService.GetCustomerScoreFor(customer);
        Apply(new ItemSold(_id,....,customerScore));
    }

This is the ideal world, but you can’t always do this …

There are times with external integrations where the cost of storing may be high or due to legal restrictions the information is not allowed to be stored with the event. I have seen a few such cases over the years. The best example was customer credit information, it was contractually disallowed to store the information that came back from a credit request with the event. As such there were places that when viewed would need to call out to get the credit score. The projection itself was replayable but from a user’s perspective the infomation was non-deterministic. This was understood due to the contractual obligation.

A related problem to this happens when projections start doing lookups to other projections. As example there could be a customer table managed by a CustomerProjection and an orders table managed by an OrdersProjection. Instead of having the OrdersProjection listen to customer events the decision was made to have it lookup in the CustomerProjection what the current Customer Name is to copy into the orders table. This can save time and duplication in the OrdersProjection code.

On a replay of only the Orders projection all of the orders will end up with the current Customer Name not the Customer Name from the time that the original order was placed. A replay will result in different results that the original. This is a tradeoff to the reduced code that was originally gained, the two projections now have a coupling.

In the case of sharing information the two projections can no longer be replayed independently, they instead need to be replayed as a group. It is quite easy to end up with an unmanageable web of dependencies like this though it can save a lot of code in some circumstances.

A common pattern to get around this is to consider the entire read model a single projection and replay the read model as a whole. Replaying all the projections for that given database together will return this situation to being deterministic. While you may be rebuilding things you don’t need to in most systems the performace hit taken is far less costly than the conceptual one to understand the web of dependencies between things. Remember as well that the replay is asynchronous, even it it takes 12 hours, it is not that big of a deal.

In a very large system replaying the entire model might be too costly, imagine a 10TB db. In this case the choice is forced between adding to the complexity of each projection to remove all dependencies between them or managing all the dependencies. The latter is normally done by placing projections into Projection Groups for replay but this can easily be overlooked and should be looked at as a last resort. It is easier and cleaner in most cases to remove the dependencies between the projections.

If you find a projection making calls to external services or to other projections it will be a problem to replay later.

As mentioned in dealing with tax in the versioning of behaviour, a place to watch out for when changing the results of calculations is that doing so may also be changing the semantic meaning of a given property of an event.

### Changing Semantic Meaning

One important aspect of versioning is that semantic meaning cannot change between versions of software. There is no good way for a downstream consumer to understand a semantic meaning change.

The quintessential example of a semantic meaning change is that of a thermostat sensor.

- 27 degrees.
- 27 degrees.
- 27 degrees.
- 80.6 degrees.
- 80.6 degrees.


Did it just get very hot around the thermostat or did the thermostat switch its measurements between Celcius and Fahrenheit? How would a downstream consumer be able to know the difference?

When a producer changes the semantic meaning of a property of an event, all downstream consumers must be updated to support the new semantic meaning or to be able to differentiate between the multiple meanings. This becomes far worse in Event Sourced systems as in the future replays need to happen.

In the temperature example, how could a future projection replay that series of events to derive some new information or a new read model? In order to do this it would to understand that at this point in time the value temperature changed from celcius to fahrenheit. It would need to have an if statement similar to:

    if(event.Date < DateItChanged) {
        realTemp = event.Temperature;
    } else {
        realTemp = (event.Temperature - 32) * 5/9;
    }

To stress the point, this would need to be in *every* consumer of the thermostat events. When its only one it does not seem that horrific, what happens when there are 5? How easy would it be to subtley forget one of these? Very quickly this becomes completely unmanageable and a more extreme measure will need to be taken such as the Copy-Replace pattern discussed later in the text.

### Snapshots

One area where developers tend to run into versioning issues is dealing with the versioning of snapshots. In Event Sourced systems it is common to think of all state as being transient. Domain objects or state in a functional system can be rebuilt by replaying a stream of events.

When the stream of events gets large, say 1000 events it is beneficial to begin to save the state of a replay at a given point. This allows for faster replays. As example a snapshot could be taken at event 990 of the stream and a replay to get the current state would only involve loading the sanpshot at event 990 then playing forward 10 events as opposed to needing to replay 1000 events.

It is worth noting that due to the conceptual and operational costs associated with persisted snapshots they are often not worth implementing. Even replaying 1000 events can be done within an acceptable response time for many systems.

Snapshots however have all the typical versioning problems found when versioning structural data. If a snapshot changes in requirements over time it is likely that it will not be able to be upgraded but will instead need to be rebuilt. If previously a state used by the domain contained

    {
       id : 848484,
       activated : false
    }

A change in logic required that the state must now also contain the name of the inventory item.

    {
       id : 848484,
       activated : false,
       name : "Gran Canaria Shirt"
    }

The information about the name is in the events that have been previously written to the stream. The state as of event 990 would be different at event 990 providing the name was being tracked previously. There is no way to “upgrade” the state as of version 990, instead the entire stream must be replayed to build the new state that also tracks the name.

The normal way of handling this is to rebuild the entire snapshot and to later delete the previously generated snapshot. On a system where down time is acceptable this is relatively easy to do. Bring the system down, recreate the snapshots, bring the system back up. This can also be done via a Big Flip.

When running side by side versions the versioning of snapshots can be a bit more complex as it is required to maintain both version 1 and version 2 snapshots at the same time. When running side by side the general strategy is first to build the new versions of the snapshots next to the old version of the snapshots. Once the snapshots are built the new version of software is released. The new version of software will run next to the old version of software and each will use its own version of snapshots.

At some point later the old version of software and its associated snapshots will be depreceated. The software version dependent on the snapshots is removed first and once there is no longer anything using the snapshots, the snapshots can be deleted as well.

One thing to watch out for is how quickly to depreceate the snapshots. Once the snapshot data is deleted it can become a very expensive operation to rebuild them. The rebuilding is also a prerequisite for attempting to downgrade to an older version as the older version is dependent on the snapshots it uses. It is often a good idea to keep older snapshots around for a while if only to allow the ability to downgrade to an older version of software in case of bugs etc.

Another issue worth considering is that multiple versions of snapshots need to be maintained. This obviously has some complexity associated with it but also requires storing multiple copies of the information. While most systems will not run into a storage problem due to the keeping of multiple snapshot versions it may become an issue in some systems.

### Avoid ‘And’

For people who are new to building Event Sourced or event-based systems in general, it is common to want to use the word “And” in an event name. However, this is an anti-pattern that leads to temporal coupling and is likely to become a versioning issue in the future.

A common example of this would be “TicketPaidForAndIssued”. Today the system will handle the payment settlement and the issuing of a ticket at the same time. But will it tomorrow? What would happen if these two things were to become temporally separated?

A less harmful issue to look out for is an event “type field”, such as IsOverride.

Developers new to Event Sourced systems have a tendency to think about these systems as being input-\>output. Under most circumstances, this is actually true. The system receives the command “PlaceOrder” and will then produce an event of “OrderPlaced”. This is not, however, always the case. Event Sourced systems have two sets of Use Cases: those you can tell the system to do, and those the system has said it has done. They do not always align.

A common example of this is looking at a stock market (or any other type of system where orders are matched). We can tell the system that we would like to “PlaceOrder”. The system also has a “TradeOccurred” event that occurs when orders match up based on price. There is, however, no “PlaceTrade” command we can send to the system.

For instance, let’s say we’re given a certain number of people who want to buy something:

| Volume |  Price | Instrument |
|-------:|-------:|:-----------|
|    100 | \$1.00 | something  |
|    500 | \$0.99 | something  |
|    500 | \$0.99 | something  |
|    800 | \$0.98 | something  |

In the stock market there are existing orders waiting to be filled. When you place an order the system checks orders that exist within the market to see if there are any orders that match based upon the price. If there are existing orders that match based upon price trades will occur. As example if there is an order to sell 100@\$1.00 and an order is placed to buy 100@\$1.00 a trade would occur for 100@1.00.

If we were to execute a command to try to sell 400@\$0.99 of something with the book above, the system would first cross the order with the 100@\$1.00 (at a price of \$0.99 depending on the market). After crossing 100 of the 400 we wanted to sell, it would then cross 300@\$0.99 with the second order. As such, from an outside perspective, it would be viewed as:

- PlaceOrder 400@\$0.99 something
- TradeOccurred 100@\$1.00 something
- TradeOccurred 300@\$0.99 something

We could also end up in a slightly different scenario:

- PlaceOrder 400@\$0.99 something
- TradeOccurred 100@\$1.00 something
- OrderPlaced 300@\$0.99 something

If less shares were available at 0.99:

- PlaceOrder 400@\$0.99 something
- TradeOccurred 100@\$1.00 something
- OrderCancelled

It is important to remember that stimulus and output are not necessarily the same set of use cases. There exist many cases where inputs and outputs of the system may differ both in number and name. Under the usual circumstances, they tend to be aligned, but this is not a given.

Coming back to the use of “And” in an event name: it forces a temporal coupling between the two things that are happening. If, later, these two concepts are split apart, there will be a messy versioning issue, likely requiring a Copy-Replace (to be discussed later) to be made. In short, do not use the word “And” in an event name; use two events instead.


---


## Whoops, I Did It Again

Over the years, one of the largest questions that comes up while teaching about Event Sourcing is, “What happens if I make a mistake?” This is a sensible enough question in an append-only model. Students can normally figure out relatively quickly how to version events forward, but that does not really help with them with a bug or a mistake. This is also one place where Event Sourced systems can actually run into additional costs compared to a typical CRUD-based system.

In a typical SQL database-type system, the solution to having a bug in the software would be to open SQL Server Manager or Toad and edit the data. It may be that it is just one row in a funny state, in which case you may just open that table and edit the data. Some cases may be more sinister; perhaps there are 900 customers who have been affected by this bug. No problem! Write an update statement that selects them!

//TODO DBA clipart.

Alas, this runs into many of the problems discussed previously in “Why can’t I update an event?” Though it can work on a single database system (often a monolith), it starts to fall over on itself when applied to a many-service environment.

What about the audit log of the system? Hopefully, there are triggers on those tables that will have captured this change, who made it, and, most importantly, why. We can quickly imagine a court case where we find out that the DBA issued an update that affected a user that it should not have and which the software would not allow based on some invariants (say, the terrorist flag in a banking system).

These types of edits petrify auditors. If the DBA has permission levels high enough to do this blanket update, how do we ensure he doesn’t have permissions to do lots of other insidious things (or undo them)?

What happens when there’s a data warehouse that consumes and aggregates our data with data from other services? How will the warehouse be alerted to these changes?

What happens if our admin made a mistake when they were building their update statement to update the hundreds of customers that had this problem? What if some customers were included that should not have been, or vice versa?

What if, as a business owner, I decide six months later that this was actually a major issue and I want to know all the customers who were affected by it? Where do I find this information?

Sometimes, these types of issues are important, other times they are not. In the types of systems where Event Sourced is applied, they tend to be important. You can get typical CRUD-type systems to handle many of these cases, but the simplicity of “we are just doing a quick edit” then goes completely out the window.

These ideas are by no means new. Pat Helland has also put forward [similar ideas](http://cidrdb.org/cidr2015/Papers/CIDR15_Paper16.pdf) and [this blog](https://blogs.msdn.microsoft.com/pathelland/2007/06/14/accountants-dont-use-erasers/) post of his is worth reading, but the concepts it outlines are much older. Mainframes were commonly built this way in the 70s and 80s. The smalltalk image format used the same [principles](http://wiki.squeak.org/squeak/49). Ultimately, these ideas go back centuries. This is not something new to learn. It is something old to research.


### Accountants Use Pens

In 2006, I gave the first talk on Event Sourcing and CQRS at QCon San Francisco. At the time, I was not a regular conference speaker, though I had spoken at a few user group meetups. My talk was about extending [Domain Driven Design](https://www.amazon.com/Domain-Driven-Design-Tackling-Complexity-Software/dp/0321125215) with [messaging patterns](http://www.enterpriseintegrationpatterns.com/) for high throughput systems (algorithmic trading in our case). My front row was Martin Fowler, Eric Evans, and Gregor Hohpe. I was terrified. I went through all my slides in about twenty minutes for a sixty minute talk.

I cannot tell you the feeling of watching Martin Fowler give you a red card. After, Eric told me, “That was a bad talk. I understood maybe 20% of what you were saying. I can only guess the rest.” For those who have met Eric Evans, to hear a negative word come out of his mouth is rare.

Luckily, I was invited back the following year, following which Eric said, “that was a very good talk.”


I have been using a slide similar to the image above since 2007 for this. What is so terribly wrong with the slide is that accountants don’t erase things in the middle of their ledgers. Event Sourced systems work in the exact same way. I have always joked that if you can’t figure out how to do something in an Event Sourced system, go ask your company’s CPA why and they will likely be able to tell you. This is one situation accountants understand well.

Let’s imagine that the account makes a fat-finger mistake and accidentally transfers \$10,000 instead of \$1,000 from Account A to Account B. Will the account just erase the zero? No, they will employ a Compensating Action. In other words, they will add new transactions to compensate for the one in error. There are two options here.

#### Partial Reversal

The accountant decides to balance the books by adding a [journal entry](http://www.accountingcoach.com/blog/what-is-a-journal-entry) that removes \$9,000 from Account B to Account A. They include a note with the journal entry that this is due to a mistake with the original transaction.

| ID  | From      | To        |  Amount | Reason   |
|-----|:----------|:----------|--------:|:---------|
| 13  | Account A | Account B | \$10000 | NONE     |
| 27  | Account B | Account A |  \$9000 | ERROR 13 |

Note that, for brevity’s sake, only the ID of the transaction in error is provided. Most actual accounting systems would also allow text to be introduced here.

Accountants, however, prefer not to do this. If the numbers are perfect numbers like \$10,000 and \$9,000, you can probably calculate in your head that the accountant originally intended to transfer \$1,000. But what if the numbers were \$6,984.82 and \$4,119.14? What if there were six accounts involved?

#### Full Reversal

In such a case, accountants tend to use what is known as a full reversal instead. With a full reversal, the account will correct the books by adding a new transaction that moves the \$10,000 from Account B back to Account A, noting in the journal entry that the entire transaction was a mistake. The accountant will then add another transaction tranferring \$1,000 from Account A to Account B as they originally intended.

| ID  | From      | To        |  Amount | Reason |
|-----|:----------|:----------|--------:|:-------|
| 13  | Account A | Account B | \$10000 | NONE   |
| 27  | Account B | Account A | \$10000 | REV 13 |
| 29  | Account A | Account B |  \$9000 | NONE   |

Accountants generally prefer full reversals since they make it much easier to figure out what was originally intended than partial reversals do. It is easy to forget about thinking of things from the perspective of someone who is coming to look back at the sitation and trying to figure out what happened after the fact.

Always consider the auditor’s perspective.

Similar examples of simple and full reversals in accounting can be found on Accounting Coach website under [Adjusting Entries](http://www.accountingcoach.com/adjusting-entries/explanation) and [Reversing Entries](http://www.accountingcoach.com/blog/reversing-entries), respectively.

### Types of Compensating Actions

Dealing with Event Sourced systems, things can often be done just like an accountant would do them. The example I have used in the past is a shopping cart. You can mistakenly add an item to a shopping cart and then remove it to end up with an equivalent (from the user perspective) shopping cart afterward. This is the same as dealing with the ledger example. From an audit perspective, the removal could be flagged as correcting an error.

This does not work in all cases, though. Accounting ledger systems are lucky in that they have natural compensating actions for all of their transactions. If you put in a credit of \$400, you can do an opposing debit of \$400. If you add an item, the item can be removed. While it is necessary to think about things like the ability to mark the debit as actually being due to an error, handling the case does not require the introduction of new events.

However, not all Event Sourced systems have natural compensating actions. When there is no use-case representing a credit, how do you reverse a debit? In systems where there are no natural compensating actions that provide a way to easily back out of something, either they need to be added or they can be added on the fly. Both of these actions come with a cost, the difference being when you pay it.

Introducing compensating actions can be a long but valuable exercise in better understanding the domain you are trying to model. “What happens if somebody makes a mistake here?” is a question that can lead towards deep knowledge discoveries. Quite often, when a SQL DBA is handling risk mitigation of these types of issues, the knowledge sits on the DBA team and the domain is not exposed to it until the task is complete. Whether with manual tasks or “clean up” batch jobs, there is often significant risk mitigation handled outside of the model.

For instance, it is common to have a system that has come up with a use-case of “TruckLeft”, but nothing to handle the situation that the person at the gate scanned the wrong truck. The introduction of analysis to discuss all of these edge conditions might also be quite costly. A significant portion of domains exist where it is not worth discussing these edge conditions, as they either happen rarely or the affects of them are small. If you have a significant number of these types of conditions, this is something you likely want to consider.

Another option is to create these events ad-hoc when/if they occur. Many Event Stores support the ability to write an event to a stream ad-hoc, say, from a json file with curl or from a small script. The problem with using these ad-hoc events is that there are consumers, most commonly projections, that do not yet understand this event. In such a case, you would need to update that consumer and then put in the ad-hoc correction.

For smaller systems, ad-hoc compensations work very well. As your number of consumers goes up and/or the difficultly of changing your consumers increases, it often becomes a bad idea to use this type of ad-hoc correction. This is true not only for Event Sourced systems but also for systems that have other push-based integration systems.

A third option, a hybrid of the previous two, is to introduce a special type of event into your system known as a Corrected event or a Cancelled event. In the case of a Cancelled event, for instance, it would contain a link to the original event that it was cancelling.

    Cancelled {
        event : '54@mystream'
    }

or somethng like:

    Cancelled {
        eventData : {
               id : '54',
               stream : 'mystream',
               type : 'AccountDebit',
               body : {
                     account : 'account-64748484',
                     amount : 65.78
               }
        }
    }

Depending on how you handle subscriptions, you can either send over a Cancelled event with a link back to the original event or you can include the body of the original event in the Cancelled event. These are just semantic differences, as the consumer can get the event data for 54@mystream if it wants to anyway. Both are valid implementations.

This is advertising to the consumers that any event can possibly be cancelled and if they really care about the cancellation of a previous event beyond notifying someone, they really should handle the Cancelled event.

### How Do I Find What Needs Fixing?

This is one of the biggest struggles for people in Event Sourced systems, especially those who may be new to Event Sourcing. We can send a compensating action, but how do we figure out which streams need it sent?

A commonly heard approach is to bring up a one-off instance of the domain model that will iterate through all of the possibly affected aggregates one by one, emitting the compensation as it finds domain objects that may be affected. This strategy is not a terrible one, but it can run into a few issues.

How, for one, does your code know what the IDs of all of the streams of that type are? Assuming that you found a problem in accounts, how do you know all the streams that are accounts? There are ways of working around this, such as using an Announcement Stream that tracks all of the account streams, but this must be in place already. If not, it can be a hurdle.

Does your domain object have enough data in it at the moment to actually identify if it is currently having a problem? It is common in Event Sourced systems to have the domain state/aggregate contain only the things requires to maintain the invariants it protects. Quite often, the domain object in such a scenario does not include the state it needs to be able to detect that this domain object needs the change.

As a concrete example, you realize the customers from Florida have had the wrong sales tax rate associated with them. The domain object however does not include a state field on it as it is not used in any of the invariants for the customer object. It is indeed possible at this point to go change the representation of the domain object to include this information, but there can be other complications with changing domain objects.

It is important here to remember that the state your domain model uses, whether in a functional or object-oriented style, is itself a projection. It is a projection that happens to be used for validating incoming messages. Why not just use another projection? If, for instance, you are using Oracle as a read model, why not just make another temporary projection in oracle?

Every projection is ephemeral. There is nothing wrong with bringing one up solely to determine what things have been affected.

What’s more, bringing up a temporary read model has some other benefits. It is easier to just throw it away after the procedure than to make changes to the domain objects. Those changes tend to end up sticking. Also, once you have a read model you can use that read model’s full power to analyze and simulate what the effect of your operation will be *before* you do it. If the streams involved happen to be related in a complex graph, such as in a financial portfolio, you can use a graph database to analyze the possible repercussions.

Another option is the [projections library](http://docs.geteventstore.com/projections/4.0.0/getting-started/) available in EventStore. The idea is here similar, but its implementation is outside the scope of this text.

While it may often seem like the right way to get things done, using of the existing domain model is often not the cleanest solution. Instead, write a one-off projection, analyze your data, and determine what the scope of your change will be. Understand it, discuss its consequences, and finally, if you determine it’s still the right move, run the compensating actions. Trust me, you don’t want to do it twice.

### More Complex Example

Cases dealing with invalidating a previous event are not always as simple. There are many cases where invalidating a previous event can actually cause a cascade through the system of invalidations of things that were intrinsically related. A good real world example of handling facts (events) that have been written and later are found out not to be true then cascading can be found in the trading domain.

A common use case in the trading domain is to have a small service that listens to executions (trades occuring in the market) and produces position updates. As example:

| Op  | Symbol | Volume |  Amount |
|-----|:-------|-------:|--------:|
| B   | SYM    |    100 | \$88.90 |
| B   | SYM    |    100 | \$88.95 |
| S   | SYM    |    100 | \$88.98 |

Here the sequence of executions states that there is a buy of 100 at \$88.90, then another buy of 100 at \$88.98, then a sale of 100 at \$88.98. Our service would like to calculate out position change events. A position is how much I own and at what price. There are however multiple ways of calculating a position.

The two most commonly used are LIFO (Last In First Out) and FIFO (First In First Out). If calculating based on LIFO the calculation of profit made and over all position would respond with:

PositionUpdated 100@88.90 PosiitonUpdated 200@88.925 //note this the price is calculated across the whole position ProfitTaken 3.00 //the trade made \$3.00 PositionUpdated 100@88.90

If however this were to be done with a FIFO strategy the calculation would come out differently.

PositionUpdated 100@88.90 PosiitonUpdated 200@88.925 //note this the price is calculated across the whole position ProfitTaken 8.00 //the trade made \$8.00 PositionUpdated 100@88.95

How you match up buys and shares depends heavily on the order that they come in. Where this can become a serious issue is what happens if a previous trade is invalidated? Every position update and profit taken that has occurred since the invalidated trade is now wrong.

If in the LIFO example the first trade at \$88.90 were later invalidated the ProfitTaken should have been \$8.00 not \$3.00. It is also easy to imagine that there are not only 3 of these events but 300 and now they have all been invalidated. The position updates are also what feeds the PnL (Profit and Loss) report, get this wrong and I can assure you angry traders will be showing up at your desk.

Not only is this situation able to be handled in an Event Sourced system, the Event Sourced system will also likely handle it in a better way than what would be found in a system that used updates.

When the invalidation comes in, the system can look up the exact trade that is being invalidated. It can then read backwards in the position stream until it comes to a 0 cross (position hits zero or switches sign).

The reading back to the 0 cross allows the system to not maintain intermediate snapshots of what was in the position at given points as it can reclaulate it from that point. Unless the number of trades between 0 crosses is very high this is generally preferable.

The system now knows that the trade is invalidated and plays forward from the last 0 cross. It operates as normal producing results from its calculation. When it hits the trade that is invalidated it ignores it. If supporting multiple invalidations it checks its invalidations first then ignores all the invalidated trades.

In the scenario above for LIFO this process would look like this:

| Op  | Symbol | Volume |  Amount |
|-----|:-------|-------:|--------:|
| B   | SYM    |    100 | \$88.90 |
| B   | SYM    |    100 | \$88.95 |
| S   | SYM    |    100 | \$88.98 |
| I   | SYM    |    100 | \$88.95 |

The invalidation is removing the 100 at \$88.95. As such the \$3.00 profit generated before is wrong based upon the LIFO strategy, it should have been \$8.00 with 0 shares remaining in the position at the end. The position stream would look like this:

PositionUpdated 100@88.90 PosiitonUpdated 200@88.925 //note this the price is calculated across the whole position ProfitTaken 3.00 //the trade made \$3.00 PositionUpdated 100@88.90 PositionInvalidated PositionUpdated 100@88.90 ProfitTaken 8.00 //the trade made \$8.00 PositionUpdated 0

This has a few benefits associated with it. Most noteably all downstream consumers of the position such and the PnL will automatically receive the position updates. They also do not need to have any special reclalculation logic as they just listen to the Position events.

While this could also be implemented using an update statement across the original positions, what happens when the trader wants to know why his position changed? Implementing this through a mechanism such as above provides a full audit history of what occured and more importantly why it occured. It can be seen via the causation id pattern that the change in the ProfitTaken to \$8.00 and the final PositionUpdated to 0 happened *because* of the Invalidation of the trade at \$88.95. While with such a simple action this might be easy to work out imagine there are 100 executions after and thus 100 position updates that follow. An alternative might be to add complexity by making it 5 invalidations, complexity can quickly spiral out of control in this problem.

This same methodology can also be used to back date a change to how the position is calculated if wanted. As example the trader decides they want to switch from LIFO to FIFO.

### But I Really Screwed Up

Sometimes things are just so screwed up it’s not worth trying to bring them forward. Nor is it worthwhile in such cases to maintain any kind of auditing functionality over the mistake, as you may not have a regulatory requirement to keep it and it may just confuse the auditors anyway.

A common example of this would be, “I just found out my import of 27gb of events on a new database is broken. Do I need to preserve the streams?” The short answer: no. Unless you have some odd data retention and auditibility rules associated with your system, just delete the database and start over.

If this were later in the life of the system, many Event Stores (EventStore included) do offer varying ways of deleting streams. Deleting a whole stream is a safe operation, whereas deleting a single event or editing an event is not. Deleting from the beginning of a stream is also a safe operation, TruncateBefore, in EventStore. You can utilize these mechanisms to delete full streams from an existing system and still be in the clear, provided it is supported in the given Event Store. Note that this will not work if you are running on a WORM drive.

A versioning-related comment I hear a lot here is, “We have data retention legislation that says we can only keep X information for T time”. This can also be handled with stream deletion, deleting the events specifically past a certain age.

Another possibility is to keep two streams of information, user-555-private and user-555-public, for example, and delete the user-555-private stream, if you must cease to retain private data, yet still need to maintain the events that were in the public stream. Note that if you allow consumers to retain such information, you should also put a RemovedPrivateData event so that any consumers can also be notified that they should no longer retain the information. If working from a WORM drive, another alternative allowable under many rule sets is to encrypt the event data and forget the key.

This ability to delete will also be useful in other patterns.

In this chapter, the discussion has centered around mistakes dealing with a single event. Other types of versioning issues, however, can be less obvious to fix, particularly those arising when entire or multiple streams have a versioning issue.


---


## Copy and Replace

Copy and replace (or Copy-Replace) is a pattern that can solve many types of versioning problems in an Event Sourced system. Have two events that need to be merged into one, or one event that needs to be split into two? This will do. You can even use it to update an event.

Travelling around talking with varying teams, I have found a huge number of them have given up on most of the other versioning strategies and just use Copy-Replace. If nothing else, it is clear the pattern has become quite popular over the last few years.

### Simple Copy-Replace

Copy and replace is a relatively easy pattern to get started with. It can technically fix any issue you could ever encounter in a stream. Have 27 events and want 3? No problem. If using the versioned typed strategies, you can even use a Copy-Replace to upgrade all of those events to the latest version while avoiding the need to upcast events to the latest version.

The general idea is to read the events out of the old stream and write them to a new one. You can forgo reading them all at once into memory (though for some types of transformations you may want to) and, instead, go one or a few at a time with a buffer. This can be seen in Figure 1.


> *Figure 1*


Once you have this operation in place, it is possible to add a transformation step in the middle. For example, to get rid of events, they would just not be copied. To transform events from one version to another, they could be upcast. Any merging and splitting of events can be done easily. Any transformation can sit in the middle.

In Figure 1, the Before stream contains 5 events, A,B,C,D,E, which are read up and written back into the After stream as need be. When the transformation is complete, the After stream contains events A,C, and E’, which is a transformation of the E event. The B and D events were not important, so they were not copied.

After the entire stream has been read through and the new stream created, the original stream is then deleted.

#### In Place Copy-Replace

Some Event Stores also support a “truncate before” operation, which deletes all events prior to event X. If this operation is supported, there is a slight variation of this pattern shown in Figure 2.

A pointer to the first event in a stream can be used in place of “truncate before”. The only difference is “truncate before” is a pointer that understands that things below it can be removed.


> *Figure2*


The only difference between this pattern and the original copy and replace one is that it is being done on the same stream. Instead of writing to a new stream, the events are appended back to the end of the same stream. All that’s necessary to make this work is to remember the first event that was written as part of the “new” stream and to ensure reading stops at that point. In Figure 2, for instance, the first event written is event 6, so you would stop reading at event 5. The delete operation is then either to use a “truncate before” or to update a pointer telling you where the beginning of the stream is.

### Simple\>

Both of these patterns are quite easy to implement â€” dangerously so. It is also quite easy to become dependent on them. As these operations can handle any possible case, it may start to seem best to just use them all the time and forget about all the other versioning approaches discussed. Why fuss over compensating actions when you can just remove the original offending event from existence?

*It is not so simple.*

While it may seem very simple, there are many ramifications to such types of operations.

*Copy-Replace is the nuclear-option of versioning.*

#### Consumers

What happens to consumers when When this operation is performed? Whether projections or other consumers, these may, for example, charge a credit card. Imagine that event A in Figure 2 is OrderPlaced, which should trigger an email to be sent.

Admiteddly, I am likely to blame for the current popularity of Copy-Replace. In my blog post, [Why can’t I update an event](http://codebetter.com/gregyoung/2013/05/28/why-cant-i-update-an-event/), I wrote:

> Occasionally there might be a valid reason why an event actually needs to be updated, I am not sure what they are but I imagine there must be some that exist. There are a few ways this can actually be done.

> The generally accepted way of handling the scenario while staying within the no-update constraints to create an entire new stream, copy all the events from the old stream (manipulating as they are copied). Then delete the old stream. This may seem like a PITA but remember all of the discussion above (especially about updating subscribers!).

In retrospect, I did not speak enough to the practical downsides of the approach.

One option is to make sure you include the same identifier on the new event as the one you are copying. This works well for some cases, but what happens if there is an event split? Certainly you can’t assign the same identifier to both of the split events. Falling back on idempotency is good in most cases, but not all.

Consider, for one, the transformation that occurred in Figure 1. The Before stream contained 5 events, A,B,C,D,E, and the After stream contained 3 events, A,C,E’. If a consumer is idempotent, it will receive E’, but what if the data in E’ is materially different to the data in E? Can it really be considered idempotency when two completely different events have the same identifier but potentially drastically different information in them? Perhaps the team handling the matter is disciplined enough to decide whether or not to use the same identifier in varying cases. But quite possibly they are not.

The worst issues around Copy-Replace are still being avoided by assuming the Copy-Replace is executed in an offline manner. An administrator stops incoming transactions, runs the Copy-Replace, re-runs all projections for read models, and finally brings the system back up. As discussed in the introduction, however, this is unacceptable for most modern systems. It might be fine for a five-minute migration, but what happens when dealing with a three-hour one?

### With a System Live

Once you decide to perform any form of Copy-Replace on a running system, things really start to get fun. When a Copy-Replace is run under such circumstances, the old events are transformed and appended to the log as new events. This implies that things such as projections will see them as new events.

How will projections to varying read models be notified that events B and D no longer exist? If they are idempotent and realize that event E’ is the same as E, what will they do about the different data in E’? How will they update their models to match the changes that occurred within the event store?

Next month, a decision is made to replay a read model. As a test, we capture the current read model, R1, at position P in the log. We then replay the projection for the same read model, R2, to position P. Does R1 = R2? R1 and R2 have seen different versions of history and are quite likely to be different, seeing as, after the Copy-Replace is applied, R1 no longer represents the events that are in the Event Store. Can we allow anyone to query R1 to make decisions after a Copy-Replace has been done? Provided the system is down, you could replay all projections then bring the system back up, which would remove this issue.

What happens if, in the old version of the software, a command is being processed, reads from the stream, attempts to write to the stream and is notified it is deleted? In other words, in the time it took to process the command, the Copy-Replace occurred, and now the stream is gone.

*Suddenly, Copy-Replace doesn’t seem so simple.*

There are some strategies that can be employed here. But not all work in all scenarios. And not all scenarios are solvable. Again, Copy-Replace is the nuclear option of versioning.

The first requirement is that both the old and new versions of the software be able to understand both the new and old versions of the stream. This sometimes happens automatically, depending on the type of transformation being applied via the Copy-Replace. If it does not, then an intermediate version must be released so the old version will understand the replacement stream. If not, when switched to the replacement stream, the old version of software will no longer work.

A small change will then be made to the original Copy-Replace instead of deleting the stream after. Write an event at the end of the stream saying this stream has been migrated. While seemingly a small detail, this event removes the need for the old version of the software to understand the new naming scheme that the streams will be migrated to (or perhaps they are just GUID’s). Essentially, this event will act as a pointer. A single generic event can be used for this StreamMovedTo { where : ‘…’}.

For auditing purposes, it can be worthwhile to put a pointer event pointing back to the original stream instead of deleting it.


> *Figure 3*


Figure 3 shows the general process discussed. The Copy-Replace will run asynchronously while the system is running. During the migration process, the producer will, when loading/writing to the stream, first check the last event in it. If it is a StreamMovedTo event, it will follow that link and use the new stream. If it is not, it will remember the version number of the event it read and use it as ExpectedVersion (even if you are using ExpectedVersion.Any normally on this write!). Using the ExpectedVersion here prevents the race condition of the stream from being migrated during the processing time.

Note that this can also work in the same-stream case.

This process will handle problems arising on the producer/write side, but, as discussed, the producer problems are not the big problems during this process. Since the events are being re-appended to the log, projections / read models are receiving these events as if they are new events.

Meanwhile, varying types of transformations may be occuring. Some transformations are reasonably safe (Upgrade Version of Event, Add New Event), some are slightly dangerous (Split Event, Merge Events) but can be worked around, and some are very dangerous (Delete Event, Update event).

To be fair, what many companies do here is just let the read models go out of sync briefly and then rebuild/replace asynchronously. If this is done a short period of time after the problem arises, the damage can often be limited and/or turned into a customer service issue. This is easier than trying to deal with the technical problem.

If you must deal with this, however, you need a way to signal all projections that the old stream is now garbage and is going to be rebuilt. This is done by writing another event to the Before stream â€” an Invalidated event.


> *Figure 4*


The pointer could be used as the Invalidated event, but it is better to make it a separate one, as there are other uses for this event.

A projection will now first receive an Invalidated for the stream. It is now the projection’s job to do any amount of clean up required to invalidate any information it has previously received for this stream. It is important to note that, although this is being brushed off, the logic for invalidation is often non-trivial and may even require storing more data than would otherwise be stored with the projection (no one said this would be easy).

After receiving the Invalidated event, the projection will then receive the entire new stream and simply process it same as it would any other new data. While seemingly simple and hassle-free, do not underestimate the amount of work and testing that goes into properly supporting an Invalidated event. For some projections, such as a list of customers, it is easy; just delete by ID. For others, it can become quite difficult, and you may be better off eating the inconsistency as discussed.

Still, Copy-Replace can be a powerful tool. It allows for many things you couldn’t do otherwise, including updating an event, but its power has a real cost in terms of complexity. It and its relatives are like the chemicals that go into cancer treatment: they may save you from one thing, but kill you in the process.

Thus far, though, we’ve only considered what happens if you completely, irreparably screw up within the boundaries of a stream. But what happens when you get your stream boundaries completely wrong and need to either split or merge multiple streams?

### Stream Boundaries are Wrong

One of the worst situations people eventually run into with an Event Sourced system is that they modeled their streams incorrectly. This can happen for a variety of reasons. Among the most common is the requirements of the system having changed over time. While developers like to think of the system they are working on and the system they started with as one and the same, the required changes a system undergoes over time eventually make it an entirely different one at the end.

Other times, developers make mistakes in analysis. Something that seems one way at the beginning becomes different as deeper knowledge is gained. A quintessential example of this is in the logistics domain. The system manages the maintenance of trucks. Originally, Engine was modeled as part of a Truck. Later, as the system matures and brings in more use cases, the realization dawns that Engines are taken out of trucks and put into other trucks. Trucks and engines do not share a life cycle and should be in different streams, yet somehow find themselves in the same one.

I am sure those with a DDD background are thinking Aggregate. I am deliberately not using the term, as Stream is more generic than Aggregate.

Needing to split is not unique to Event Sourced systems. Document-based systems [have an almost identical problem](https://arxiv.org/pdf/1506.08800.pdf) with this scenario. The existing documents need to be split to produce two documents.


> *Split Stream*


Much like dealing with Copy-Replace, Split-Stream involves reading up from the first stream and writing down to two new ones. For example, you could read up from vehicle-1 and then write back to two new streams, vehicle-1-new and engine-7.

You could even do a transformation in the middle, just like with Copy-Replace, but do not do this. Applying a transformation comes with all the same problems as doing a transformation with Copy-Replace. How does a consumer handle the new events it will see? If you can simply avoid all transformations, then you can completely rely on idempotency in this case. If all you have done is move the events to different streams, they will remain the same events.

One place to be careful with a Split-Stream, though, is if the two new streams will share some events. In such a case, the same event gets written to both. This may or may not cause problems with idempotency, depending on the Event Store.

There is also a closely related problem, known as Join-Stream. It is the exact opposite of Split-Stream. Essentially, there are two current streams where one is desired. Obviously, it will be handled in an almost identical way.


> *Join Stream*


Both of these are relatively easy to implement when stopping the system while they run. As they are only copying the same events and not translating, consumers simply need to be idempotent and everything will work without an issue. Alas, the luxury of being able to take the system down for a data migration is a rare one.

### Changing Stream Boundaries on a Running System

Both Split-Stream and Join-Stream have similar problems of being run on a live system as Copy-Replace. They are simpler providing you do *not* introduce a transformation as part of the operation. If you want to transform, do it in two stages and your Split-Stream and Join-Stream will remain simple. That is, be sure to keep composition in mind.

In the exact same manner that the Copy-Replace required that both versions of software support both models, both Split-Stream and Join-Stream require both versions of software to support them too. If the current version does not support both models, you will need to do an intermediate release to ensure it does.

The strategy for determining which bit of logic to use is also similar to the Copy-Replace. For Split-Stream, first try to read the last event from the single stream. If it is not a StreamMovedTo event, then continue to read the stream as normal; otherwise, follow the appropriate link in the StreamMovedTo (note there are multiple ones here). For Join-Stream, the operation is almost identical (see Figure 5). Go to the original stream, check the last event if it is a StreamMovedTo, then follow the link; otherwise, continue with that stream.


> *Figure 5*


As is also the case with Copy-Replace, it is very important to use ExpectedVersion when writing back to the stream. This is because the stream you are working on may get migrated while you are working on it. ExpectedVersion should be set to the last event read from the stream. If this step is forgotten, you may write an event to a stream after it is migrated and lose it in the process.

When dealing with Split-Stream and Join-Stream, it is very common to leave the original stream at least for a period of time. This is in case there was a problem with the operation or an auditor might want to make sure events were not added or lost along the way. It is also common to add a MigratedFrom {stream} event at the beginning of the migrated streams.

As long as the rule above about *not* performing a tranform while you are doing a split or join operation is respected, projections and other consumers will only need to be idempotent. Idempotency will catch all of the repeated events properly and you won’t need to worry about other edge cases, such as with Copy-Replace.

GetEventStore supports a slightly modified version of these patterns through its use of links. A link is a special type of event that acts as a pointer to another event. Using links, the process is a bit different, as well as better suited to auditing. Instead of copying the information on the Split-Stream or Join-Stream, you use links.

To start, block access to the original stream(s). This can be done by taking away write privileges. Next, write links to the stream(s) just as you would with a Copy-Replace. The new streams point back to the events in the original stream(s), but when you read from them they show up as if they were in the new stream(s). When an auditor looks at the system, they can see the original data and that the Split-Stream or Join-Stream occured. Also, since links are being appended, consumers will see them as links and not take action based upon them.

This chapter contains answers for what are probably the most complex problems in this book. Using these patterns, you can migrate 1M streams, including a selective split operation, on a 5TB database while your system is processing 2000 requests/second. This is likely not, however, your use case. In those others cases, there are other situational strategies for allowing the system to avoid downtime, not to mention these often horrifically complex scenarios.


---


## Cheating

The patterns Copy-Replace, Stream-Split, and Join-Stream covered in the previous chapter are quite intense. Getting them to work on a running system is non-trivial. There will be times that they might need to be used. However, most of the time they can be avoided.

The patterns will work in any scenario. It does not matter whether it is a system that is doing 5000 transactions per second with TB of data or one doing seconds per transaction with 1 GB of data. They work in the same way.

There are, however, “cheats” that exist and which can be used in place of the more complicated patterns. Depending on the scenario, they may or may not apply, and the trade-offs associated with them may be better or worse than the those associated with the complexity of other patterns.

In this chapter, we will look at some of these “cheats” and consider where and when they might be applicable. Some relate to adding code complexity to avoid the need for versioning. Others look at different deployment options. If possible, you should generally favor a pattern from this chapter unless you are in a situation where you cannot use it for some reason.

### Two Aggregates, One Stream

One of the most underrated patterns to avoid the need for a Stream-Split is to have two aggregates (states) being derived off of a single stream in the write model. It is strange that such a simple concept is often overlooked by teams who instead decide to go for the far more complex method of a Stream-Split. To be fair, the reason it gets overlooked is because many consider it to be hacky; it is breaking isolation, it is improper. At the end of the day, the job of a developer is to build working systems while mitgating risk. Code and architectural aesthetics lay a distant second.

There is *nothing* wrong with having two aggregates built off of the same stream.

Consider the previous example of the logistics company that has Engine currently modeled inside of Truck but then realizes that engines are actually taken out of trucks and placed into other trucks. Sometimes the truck is then discarded. This is a clear case of having two different life-cycles where a developer would really want to Stream-Split the Truck stream into a Truck stream and an Engine stream.

Another option is to not touch storage at all. Instead, only change the application model above it. In the application model, make Truck and Engine two separate things but allow them to be backed by the same stream. When loading an aggregate normally, events it does not understand are just ignored. So when Truck is replayed, Truck will listen to the events that it is interested in. When Engine is replayed, it will receive the events that it cares about.

Both instances will be able to be rebuilt from storage, and both instances can write to the stream. Ideally, they do not share any events in the stream but there are cases where they do. Try to keep things so that they do not share events. This will make it easier to do a Split-Stream operation later on, if wanted.

When writing, it is imperative that both use ExpectedVersion set to the last event they read. If they do not set ExpectedVersion, then there could be the other instance writing in an event that would disallow the event that they are currently writing. This is a classic example of operating on stale data. However, if they both use ExpectedVersion they will be assured the other instance has not updated the stream while a decision was made.

The setting of ExpectedVersion and having two aggregates can lead to more optimistic concurrency problems at runtime. If there is a lot of contention between the two aggregates, it may be best to Split-Stream. Under most circumstances, though, there is little contention either due to their separate life-cycles or due to a load overall system load.

Essentially what this pattern is doing is the exact same thing as a Split-Stream. It is just doing it dynamically at the application level as opposed to doing it at the storage level.

### One Aggregate, Two Streams

A similar process can be applied to the problem of Join-Streams. When loading the aggregate, a join operation is used dynamically to combine two streams worth of events into a single stream that is then used for loading.

The join operation reads from both streams and orders events between them.

    while(!stream1.IsEmpty() && !stream2.IsEmpty()) {
        if(stream1.IsEmpty()) yield return stream2.Take();
        if(stream2.IsEmpty()) yield return stream1.Take();
        yield return Min(stream1,stream2);
    }

This pseudocode will read from the two streams and provide a single stream with perfect ordering. It handles two streams but extrapolating this code to support N streams is not a difficult task. The aggregate or state is then built by consuming the single ordered stream generated.

The other side of the issue is handling the writing of an event. Which stream should it be written to? Most frameworks do not support the ability to specify a stream to write to, they are centered around stream per aggregate. If working without an existing framework or if you are writing your own it is a trivial issue. Instead of tracking Event, track something that wraps Event and also contains the stream that the Event applies to. Under most circumstances it will be stream per aggregate, but it supports the ability to have multiple streams backing an aggregate.

Another use case for having one aggregate backed by two streams is when faced with data privacy or retention policies. It is quite common that a system need to store information though some of it must be deleted after either a time period or a user request. This can be handled by creating two streams and having one aggregate running off of the join of the two streams. There could be User-000001-Public and User-000001-Private. User-000001 when loaded represents a join of these two streams. The User-000001-Private stream could have a policy associated to lose information or could be deleted leaving the public information in User-000001-Public.

Having multiple streams backing an aggregate is used much less often than multiple aggregates on the same stream. It feels more hacky and no frameworks that I know of support it. That said it can be a useful pattern when considered that the alternative would be to do a Join-Stream at the storage level.

### Copy-Transform

Every pattern that has been looked at up until this point has focused on making a versioned change in the same environment. What if intead of working the same environment there were to be a new environment every time a change was made? This pattern has been heavily recommended in the great paper [The Dark Side of Event Sourcing: Data Management](http://files.movereem.nl/2017saner-eventsourcing.pdf).

One nice aspect of Event Sourced systems is that they are centered around the concept of a log. An Event Store is a distributed log. Because all operations are appended to this log many things become simple such as replication. To get all of the data out of an Event Store all that is needed is to start at the beginning of the log and read to the end of the log. To get near real-time data after this point just continue to follow the log as writes are appended to it.


> *Big Flip*


Figure 1 certainly seems much more complicated than the other patterns thus far. It is however in many ways simpler than them. Instead of migrating data in the Event Store for the new version of the application, the old system will be migrated to the new system. This migration process is the same process you would use if you were migrating a legacy system to a new system. Instead of versioning storage on each release treat each release as an entire new piece of software and migrate the old system to it.

A nice side benefit is if you treat every release as a migration, you should be pretty good at migrations when it is time to actually migrate off the system!

Most developers have done a migration at some point in their career. One of the great things about doing a migration is that you don’t have to keep things the way those awful incompetent rascals before you did things; *even if they were you*. When doing a migration you can change however the old system worked into your new thing which is of course unicorns and rainbows, well until next year when you migrate off it again.

Instead of migrating system to system in Copy-Transform you migrate version to version. When doing this as in doing a Copy-Replace there is an opportunity to put a transformation in the process. Another way of thinking about Copy-Transform is Copy-Replace but on the entire Event Store not just on a stream.

In this transformation step, just like in Copy-Replace, anything can be done. Want to rename and event? Split one event into two? Join streams? The world is your oyster as you are writing a migration. The migration starts on the first event of the old system and goes forward until it is caught up. Only the old system is accepting writes so the migrated system can continue to follow it.

The new system not only has its own Event Store, it also has all of its own projections to read models hooked to it. It is an entire new system. As the data enters the Event Store it is then pushed out to the projections that update their read models. The Event Store may catch up before the projections. It is very important to monitor the projections so you know when they are caught up. Once they are caught up the system as a whole can do a [BigFlip](https://people.eecs.berkeley.edu/~brewer/Giant.pdf).

To do a BigFlip the new system will tell the Event Store of the old system that it should stop accepting writes, drain all the writes that happened before. Wait a short period of time. Then direct all traffic to the new system. This process generally takes only a few hundred milliseconds. After the new system has all load pointed to it and the old system is discardable.

A common practice is to issue a special command to the old system at this point that will write a marker event. This marker event can be used to identify when the new system has caught up with the old system and the load balancer, etc., can be updated to point all traffic at the new system. This is a better mechanism than waiting for a few seconds, as it signals exactly when the new system is caught up and avoids the possibility of losing events during the migration due to a race condition.


> *Big Flip*


It is important to not that because this is a full migration all of the issues dealing with projections from Copy-Replace go away. Every projection is rebuilt from scratch as part of the migration.

Overall this is a nice way of handling an upgrade. It does however have a few limitations. If the dataset size is 50 GB the full Copy-Transform is not an issue. It might take an hour or three but this is reasonable. I personally know of an EventStore instance in production that is 10TB, “We released this morning it should be ready in a week or two.”. Another place it may not work well is when you have a scaled or distributed read side. If I have 100 geographically distributed read models it might be quite difficult to make this type of migration work seemlessly.

Another disadvantage of Copy-Transform is that it requires double the hardware. The new system is a full new system. This is however much less of an issue with modern cloud based architectures.

If you are not facing these limitations a Copy-Transform is a reasonable strategy. It removes most of the pain of trying to version a live running system by falling back to a BigFlip. Conceptually this is simpler that trying to upgrade a live running system. It is also a good option to consider

### Versioning Bankrupty

Versioning Bankrupty is a specialized form of Copy-Transform. It was mentioned in the Basic Versioning chapter that everything you need to know about Event Sourcing you can learn by asking an accountant.

> So how do accountants handle data over long periods of time?

*They don’t.*

Most accountants deal with a period called [year end](https://debitoor.com/dictionary/year-end). At this time the books are closed, the accounts are all made to balance a new year is started. When starting a new year the accountant does not bring over all of the transactions from the previous year, they bring over the initial balances of the accounts for the new year.


One of the most common questions in regard to Event Sourced systems is “How do I migrate off of my old RDBMS system to an Event Sourced system?”. There are two common ways. The first is to try to reverse engineer the entire transaction history of given concepts from the old system to recreate a full event history in the new system. This can work quite well but runs into issues in that it can be a lot of work and that it may be impossible.

The second way of migrating is to bring over an “Initialized” event that initializing the aggregate/concept to a known starting point. As an example in the accounting example an AccountInitialized event would be brought over that included the initial balance of the account as well as any associated information like the name of the account holder.

In doing a system migration it is common to use both mechanisms chosen on an aggregate by aggregate basis. For some an “Initialized” event may be brought over and for others a history can be brought over. The choice between them is largely how much effort do I want to put in compared to what value is there in having the history.

*Remember that a Copy-Transform is a migration! Just from one version of your system to the next*

When migrating from your own system that is already Event Sourced it should be relatively easy to reverse engineer your history since it is well your history. In this circumstance there may be another reason to bring over “Initialized” events. They allow you to truncate your history at a given point! Imagine there is a mess that has been caused on a set of streams over time. You can summarize that mess and replace it with an “Initialized” event as part of your migration in Copy-Transform.

Businesses also often have time periods such as year end built into their domain. If you want to change the accounting system in a company it will almost certainly happen at year end, not at any other time during the year. This can also be a good time to execute a bankruptcy. To execute the bankrupty, run the migration then instead of deleting your old store and read models such as in Copy-Transform, keep them. Label them for that time period. They are now there for reporting purposes if needed.

I once worked at a company that did this to their accounting system every year regardless of if they were changing or not. In the accounting office there were 10 accounting systems. Nine were marked by their year and were read-only as only the current one could be changed. They even ran them on separate labeled machines such as “ACCT1994”. They declared bankruptcy every year even if they were not changing software and migrated into the new accounting system.

Versioning Bankruptcy can also be used as We-Have-Too-Much-Data Bankruptcy. If large amounts of data are coming into a system a yearly or shorter time frame where the old data is migrated to the new with “Initialized” events can allow for the archiving of the old data allowing for the current system to run with much smaller datasets.

On an algorithmic trading system I worked on we had a daily timeframe. The market open and closed daily. We would at the beginning of the day bring over all of our open positions, account information, and look up data. Data would come from the market for the day as well as our own data on positions etc. At the end we would go through an end of day process. One great feature of this is that a day’s log was totally self-contained and can be archived. Later to understand June 2, 2006 you can load only the log of June 2, 2006 and everything is needed.

There is also a downside to this style of archiving for either versioning or space needs. It can become quite difficult to analyze historical data outside of a single time period. Some systems tend to be based on periods, some based on continuous information. This pattern does not work well in continuous systems. Running a projection that covered two years worth of information from the accounting system was a real pain. With the algorithmic trading system, running a projection that spanned two weeks was a pain. As we did it on a regular basis, a solution was built for it. However, it was complex and took a good amount of time to implement.


---


## Internal vs External Models

Up to this point in the book, the focus has been on handling versioning issues of internal models. In practice however it is common to have both internal and external models for a system. The different models serve different needs. The versioning patterns can be used for both but external models tend to have some material differences compared with how to handle internal models.

External models do not necesarrily imply dealing with say a third party. It is quite common in an organization for a group of services to interact with each other through an internal model but to interact with the rest of integrations through a less granular external model.

An example of this can be seen in a medical system. Pharmacy, Medical Record Data, and Patient Data may all be implemented as individual services internally but the rest of the system interacts with them as a group through an external model.

The other systems may even be developed by the same teams. Often the choice between an internal vs. an external model is not team based but focused on attributes such as stability of the contract and the concept of knowledge hiding. Though there are other reasons that an external model may be chosen.

### External Integrations

One of the most common reasons to choose to add an external integration model is conforming the system to an existing standardized integration model. It is not always a possibility to be able to choose the integration model that others will consume.

At the same time when you integrate to an external model it is not always a good idea to use the external model as system’s internal model. Often there are things that the system will need to support that are not in the shared integration model. It will also be shown later in the chapter that often the external model will want things at a different granularity than what the internal system wants to view them as.

Examples of external integration models could include the [Global Justice XML](https://en.wikipedia.org/wiki/GJXDM) model, [HL7v2](http://www.hl7.org/implement/standards/product_brief.cfm?product_id=185) for medical data in the UK, or even [FIX](https://en.wikipedia.org/wiki/Financial_Information_eXchange) which is one protocol amongst many in financial systems.

These protocols tend to be created by committee. They are often massively complex and try to handle every possible situation regardless of how unlikely it is. The complexity of trying to normalize all of the medical coding in disparate health care systems is easy to imagine.

//TODO add example

Supporting these protocols can offer a lot of benefit on the large. They can allow for modularity between systems. They can allow other systems to easily integrate with your system. Often times there are even off the shelf tools that support such protocols that can be leveraged if you support the canonical protocol that it supports.

Often times these standardization efforts fail spectacularly as well and the systems are not interchangable at all.

Not all external models come from external integrations however. It is quite common to treat integrations with other services outside of a tight service group as an external integration due to some of the differences between how internal and external models are treated.

### Granularity

One of the largest differences between internal and external models is the level of granularity of the events that are provided for the model. Internal and external models have drastically different needs and when looking through the API published via each there will be large differences in the granularity of information provided on an event.

Internal models tend to have very fine-grained events. This is largely due to the separation between services and the focus on a use-case being represented by an event.

    OrderPlaced : {
        orderId : 8282,
        customerId : 4432,
        products : [5757, 3321],
        total : 17.95
    }

In the simplified example of an internal OrderPlaced event above much of the information provided is normalized. The information is represented by identifiers that point to the customer and product ids. In order for a consumer to interact with this event it must already know who customer 4432 is or to not be interested in that information.

When dealing with internal models the service boundaries chosen for the system tend to be visible within the events in the model. Generally when dealing with service boundaries identifiers are used as opposed to denormalization. The inverse is also true which is why using an external model as your internal model has a tenency to lead towards a more monolithic system.

A consumer of this model would need to listen to multiple events such as ProductCreated and CustomerCreated. It would then likely require some form of intermediate storage to remember who customer 4432 was if it wanted to also look into customer information associated with the order. A common place in systems to see such a behaviour off an internal model is a reporting database.

External models generally prefer more coarse-grained events. An external model will generally denormalize relevant information onto the event. The reasoning behind this is simple, the external consumer only wants to listen to a single event and to be able to act upon it. An external consumer prefers to not have to listen to multiple events with storing intermediary information.

    OrderPlaced : {
        orderId : 8282,
        customer : {
                       customerId : 4432,
                       customerName : 'John Doe',
                       postalCode : 'EC1-001',
                       customerAge : 35,
                       customerStatus : 'Gold'
                   }
        products : [
                        {name : 'razor blade ice cream', productId : 6565},
                        {name : 'thing you should never eat', productId 4242}
                   ],
        total : 17.95
    }

In the example of the external model event instead of only providing ids, the information has been denormalized into the event itself. There is not only a customerId but the information this system deems to be relevant about the customer. Note that this is just an example and I hope such a system would not be as naive as this one.

A consumer however can subscribe to this event and act directly upon it without the need to subscribe to other events. For an outside consumer this is far simpler than needing to subscribe to and understand the interactions of multiple events.

The external model is providing a level of indirection from the internal model. It is also hiding the details of how the internal model works. Much like with any other form of abstraction it allows the internal details of the system to change without breaking the consumer. If the internal model wanted to add a new use case for a ManagerialOverridedProductName, the external consumers of the API would not need to be aware of it as they would still receive just the OrderPlaced with the product information associated.

At the same time the internal system is free to change it’s service boundaries or how it handles it’s internal eventing without affecting the external consumers. Said simply the external model provides a level of indirection from the internal model.

### Rate of Change

Levels of indirection are often useful, they are used regularly in code. Whether we discuss an interface, a function being passed, or a facade all represent methods of indirection. All can provide reusability as well as information hiding.

As with most cases of indirection the contract associated generally has more design work associated with it. Changes to a contract associated with indirection tend to be expensive in comparison to other types of changes in software. The reason for this is that changes to a contract require a change in every user of the contract as well. When all of the consumers are in the same codebase this can be automated with tools such as intellij but when they are not expense can escalate quickly.

External models have the same tradeoffs associated with them. Changes to external models will possibly affect all consumers of the external model. For this reason external models generally are more formalized than internal models and changes to external models generally go through further diligence than changes to internal models.

External models will generally be designed up front with the idea of extensibility. They will normally be formalized to media types or described in a schema language. Interactions will generally be well documented for consumers.

Changes to external models often happen at a glacial pace unless the model has been specifically designed for extensibility. Breaking changes to external models rarely if ever happen. I have seen numerous projects utilizing an external model that had embarassing spelling mistakes in them that the team will keep forever because they loath the idea of introducing a possibly breaking change into their model. The costs of a breaking change far outweigh the benefits of removing the embarassing spelling mistakes.

Due to the worry of breaking changes many external models are built using weak-schema. Adding things to the external model will not break things. External models are almost never built using the type based versioning discussed though many frameworks push developers to do this.

When dealing with external models and versioning if pushing an event through a stage that then produces a similar event to the next stage or enriches the same event. Deserialization to a type is generally avoided. Instead a rule is put in place that “things that you do not understand” must be copied to your output. This avoids a change being required when an additive change is made.

It is common to use a negotiation model such as in the Hypermedia chapter in association with an external model to further isolate external consumers from changes to the externa model.

Internal models are the opposite of this. If you have a small cluster of services that are under control of the same group, changes are not nearly as dire as when dealing with a large number of external consumers.

The introduction of an external model to the system can prune many of the consumers. This pruning of consumers can allow more agility in dealing with the internal model as there are less coupling points to consider. A closely related set of services and a read model or two can evolve quite quickly together without having drastic effects on the majority of the downstream consumers associated.

### How to Implement

When faced with an internal vs an external model decision, many developers want to stick to a single model. This is a natural tendency to consider that having one model will be simpler than maintaining two. The focus is generally to either make the external model the internal model or vice-versa.

This in practice can introduce a large amount of complexity to the system. The two models have different purposes and have quite different levels of granularity. Another common issue to arise here is what if the system needs to support multiple external models say [HL7v2](http://www.hl7.org/implement/standards/product_brief.cfm?product_id=185) and [HL7v3](http://www.hl7.org/implement/standards/product_brief.cfm?product_id=163) or some other protocol for medical information?

A common implementation to introduce an external model is to introduce a service that acts as a [Message Translator](http://www.enterpriseintegrationpatterns.com/patterns/messaging/MessageTranslator.html). This service listens to events happening on the internal model. It may or may not have intermediary storage in order to denormalize information to the external model.


> *Translator Service*


There are a few benefits to using this strategy. The first is that all of the code associated with the translation to the external model is located in a single service. The rest of the services know nothing of the external model. A secondary benefit is that it is quite reasonable to introduce multiple of these services to support multiple external models.


> *Multiple Translators*


Multiple transalators are just an addition of an additional translator. For systems that are deployed on a per client basis, the translator services often become optional components of the system where zero or many may be deployed for a given client.

Another consideration to keep in mind is whether to buy or build the translator. For widespread protocols there are often existing adapters built into existing messaging products. These existing adapters often handle more tricky aspects of the external model such as validation and normalization of data within the messages. I know of a client who uses biztalk solely for the [HL7 integration](https://msdn.microsoft.com/en-us/library/ee409342.aspx), they don’t actually use it for handling their own messaging, only as a translator.

### Summary

There are many differences between internal and external models. Smaller systems can usually get away with only using an internal model. As the system size and complexity grows however it is often needed to introduce levels of indirection via the introduction of an external model. External models are often used even to other services maintained by the same team, the distintion is how the model is used not who uses it.

Internal and external models operate at quite different levels of granularity. Internal models tend to be very fine-grained while external models are more coarse. This difference is due to how consumers of the model view them and how they wish to use the model. Attempting to make a more coarse-grained internal model to also be used as an external model will likely start to blur your service boundaries.

Rate of change and overall risk of change are quite different between the two models as well. External models tend to be much more conservative in how they approach change compared to internal models. A common pattern is to introduce an external model to allow the internal model to be more agile in regards to change.


---


## Versioning Process Managers

Process Managers are a highly useful pattern in Event Sourced systems. There is however almost no information available via books, blogs, videos, etc on how to handle the versioning of Process Managers in a production system. This chapter will be focused on exploring Process Managers and in particular looking at how to handle multiple versions of processes in a production system.

### Procees Managers

[Process Managers](http://www.enterpriseintegrationpatterns.com/patterns/messaging/ProcessManager.html) though known by other names including Orchestrations represent a business process that spans more than a single transaction. They have a responsibility to either get the overall business process completed or to leave the system in a known termination state. If an error happens the Process Manager may or may not handle a rollback procedure.

Process Managers represent the handling of the overall business process. As example when an OrderPlaced event is received the system could start an OrderFulfillment Process Manager that will handle the overall process of fulfilling the order. If along the way there are issues that occur within the the system the Process Manager will either handle those errors automatically or detect that they have occurred and notify a customer service representative or system administrator who will manually deal with the errors.

These business processes are often quite valuable to the business overall. Oddly in many systems they are not actually modelled in any way but are instead in the minds of the users. While modelling explicit Process Managers is often a good idea, the versioning of a processes in an organization can at times be tricky.

### Basic Versioning

The most simple and generally preferred method of handling versioning of business processes is to handle it in the same way that the organization itself tends to handle it.

When introducing a change to a business process in an organization it is rare to introduce it as a change. Instead organizations will tend to introduce a new business process. This new business process will affect all instances of the business process that occur after a given point in time which is used as a cut off date.

If for example a bank were to introduce a new process for the approving of mortgages they would introduce it as a new business process. An email would come from the group managing the process that as of March 13th all new mortgage applications would go through the new process. There will be mortgage applications that were received on March 12th that will continue to go through the old application process.

There is a reason that organizations tend to introduce changes to a process as a new process. In particular it can be difficult to change mortgage applications that were running on the old process to instead run on the new process. The old process and the new process may or may not be compatible and changing a mortgage running through the old application process may not be possible.

As a rule you never change a Process Manager. You make a new process.

Once a Process Manager is running in production there will no longer be changes to that business process. The code of the Process Manager should be read-only. Any changes to the business process will result in a new Process Manager being shipped to production. This new Process Manager will also usually have a start time associated with it. All processes that start before noon on Tuesday will use the old Process Manager any that come after this point in time will use the new Process Manager.

There are many advantages to using this style of release. One of the largest gains is in terms of predictability. Changing processes as they are running is a dangerous operation. It is quite likely that the change will leave some instances of the Process Manager waiting on things that will never occur. Processes that end up waiting indefinitely on something that will never occur are also fondly called “zombies”.

Though not about versioning it is important when writing Process Managers to never wait infinitely on something. You can wait for Int64.MaxValue 3,372,036,854,775,807 seconds which is roughly 292,471,208,678 years but it is still not infinite. Ideally your timeouts should be much lower (days at most). This will allow an instance of a Process Manager that has gotten into an odd state to timeout and hopefully recover or notify somebody about the failure that has occurred.

If you can get to the point of having all changes to processes being released as new processes, versioning will not be a major issue. There is always a cutoff point and things before that will run on the old process, things after on the new process. There is no versioning needed in this case. Process Managers enable the ability to have multiple concurrent business processes as there is an instance of the Process Manager per thing that is being managed. It is a common pattern as well to have multiple business processes based on some attribute of the initial event (is this an order less than \$5 or an order \> \$1m, each could have its own process).

While bringing in new Process Managers at a point in time and leaving the old processes running to conclusion is the ideal, it is not always possible. Most Process Managers run over a time span of seconds to days. For these types of processes you can usually use the cutoff strategy described. Organizations tend to use the same versioning structure discussed at an organizational level anyways so it is a good fit. What if the process was a multiple year process?

### Upcasting State

When dealing with say a multiple year long process there are times where it is necessary to change not just what the future instances of the process will do but also the ones that are currently running. As the process duration becomes longer versioning becomes more important. This is a niche use case but one that it worth covering.

An example of this can be seen in a Process Manager that manages the balancing of funds in a trading plan. It could run for years, and there will be times when changes need to be made to existing instances of the process while they are running.

This changing of the process while it is running is a dangerous operation. As mentioned previously, a change can leave a running process in a state where it will never receive what it is waiting for. Another possibility is that due to the change the process may end up duplicating operations.

The duplication of operations and possible complications that can arise from it is easy to see with a simple example process. In a restaurant the business process could be either to have the customers pay before receiving their food or they could pay after receiving their food. If there was an order that was set to pay before it had its food, and the business process was changed to make the customer pay after they received their food it is quite possible that the process would attempt to have the customer pay twice.

While it is a niche situation that a Process Manager would be upgraded in place it is important to cover how to handle this case.

#### Direct to Storage

The simplest and most common method of changing the version of a Process Manager in place is to upgrade the state associated with the Process Manager instance. There are multiple ways of achieving this.

Many frameworks offer no explicit concept of changing the version of a running Process Manager. Instead the recommendation is to go to the database that is backing the Process Managers and update the state in the data store. This is a highly dangerous operation and should be avoided if at all possible.

At first it can seem to be a relatively simple thing to do. Issue an update statement against the backing storage and then all of the states are upgraded. How will this work when the system is running at the same time that the upgrade is taking place? What if a mistake is made? Ideally this backing storage should be a private detail of the framework not something you directly interact with.

#### New Version Migrates

In order to avoid the direct updating of underlying storage some frameworks offer an explicit method of handling an upgrade. The framework itself understands the concept of multiple versions of a Process Manager and can also understand when what was a v1 is now being upgrade to a v2. The frameworks generally offer a hook to the Process Manager at this point where the new version of the Process Manager can decide how to upgrade the old version of the state.

    public class MyProcessManager_v1 : WithState {
        private State_v1 state;
    }

    public class MyProcessManager_v2 : WithState, Upgrades {
        private State_v2 state;

        public State_v2 MigrateState(State_v1 state) {
            return new State_v2();
        }
    }

The code example above is a common implementation of a state migration as supported by the underlying framework. When the new version is used for the first time, the framework will see that the old version (v1) used a State_v1 state. The framework will then call into the MigrateState method passing the old state to the new version to allow for a migration before it passes the message into the new version of the Process Manager.

It is important to note that since MigrateState is just a method the new version of the Process Manager may even decide to interact with other systems in order to migrate the state. A common example of this would be the old version only kept accountId but the new version requires accountNumber as well that it under normal running saves off of messages.

Another benefit is that since the migration is happening as part of the delivery of the message this mechanism can also be run while continuing to process messages. The system does not need to be shutdown during the migration process though it may operate in a degraded way depending upon the complexity and interactions of the migration method.

The migration method at least removes the problem of directly interacting with the private Process Manager state storage. Not all frameworks support this type of migration. In general even if your framework supports this type of migration you should probably avoid it as there is another more idiomatic way of handling the migration.

### Take Over

As stated at the beginning of the chapter, it is best to let a Process Manager run to its end under the same version that it was started on. Changing a Process Manager while it is running is a tricky and dangerous operation. Explicit migration support from whatever framework you are using can help to mitigate some of the issues associated with the switching of versions but not all.

Of particular concern with the in place migration is that it can become very difficult to debug. When exactly did the migration occur, on which message? Some frameworks offer logging capabilities to show when it occured but not all. It is common to run into situations looking back where it is unclear which version of the Process Manager was running at which times. This is especially if you have messages that arrive close temporally to the release of the new version or if you have multiple servers and roll out the new version incrementally.

There is another option that can be used known as a Takeover or a Handoff that can help with both of these issues. Instead of upgrading in place the previous version of the Process Manager is told to end itself, it will then with its last dying breath raise a message “TakeoverRequested” that will start off the new version of the Process Manager.

Upon handling the “EndYourselfDueToTakeover” message the previous version of the Process Manager can do any clean up work that it wants based on its state. As example if it had previously taken a hold on a credit card as part of a sales process, it can release the hold. It leaves the process in a known end status. Once it is in a known end status it will then raise the “TakeoverRequested” message on the same [Correlation Id](http://www.enterpriseintegrationpatterns.com/patterns/messaging/CorrelationIdentifier.html). In raising this message it can put any relevant state that it has into this message. Once this message is raised, it is considered terminated.

The new version of the Process Manager will be associated to run on the “TakeoverRequested” message. It will then start up and enter into its workflow based upon the state in the take over message as opposed to its normal entry point.

The message names here are manufactured and likely would be domain relevant wordings.


> *Take Over*


This strategy keeps with the original goal of letting the original Process Manager run to the end of its life cycle. It will simply pass control to the new version. Another large benefit is that since the handover is done via a message on the same Correlation Id there is a record of exactly when the take over occured contained in the actual message stream and indexed by Correlation Id for debugging purposes.

Of all the choices looked at thus far the Takeover is the cleanest strategy. What is also nice about the Takeover strategy is that it does not require any explicit framework support to work. As such it is a framework agnostic strategy which can also help with removing coupling surface area from a framework.

Another point to consider especially for longer running processes is to break them apart into multiple Process Managers. If as example you have a single Process Manager that is responsible for everything from the time a mortgage application comes into the bank until it has been securitized and sold, versioning will be extremely difficult. Breaking this large process up into multiple related sub-processes will make versioning the sub-processes more simple as each is a smaller set of operations resulting in a smaller number of possible takeover scenarios that would be involved.

There are however another strategy that in some scenarios can make state management easier.

#### Event Sourced Process Managers

It being the book is about Versioning of Event Sourced systems, it is only natural to bring in the concept of Event Sourced Process Managers. Many frameworks such as [Akka.Persistence](http://getakka.net/docs/Persistence) use this mechanism by default. Many other systems are unable to support it as it requires the ability to replay previous messages.

As discussed previously when discussing models and state in general, one advantage of storing events as opposed to storing state is that state is transient

One major problem with any form of state migration shown is that the previous version of the Process Manager may have not saved enough information for the new version to be able to build up the state that it needs. An example of this could be that the previous version of the Process Manager never saved what the customer this order was for as it did not need it. The new version requires a credit/terrorism check on the customer before the order can be completed. When attempting to migrate the state this information does not exist.

There could be some process where the new Process Manager might go and talk to other services to try to determine this information but this can quickly become complex, and there may not be a service that has the information that is required. The information may only exist on the previous messages. A good example of this is an indentifier returned earlier from a third party that now needs to be sent back on a message in the future.

Event Sourced Process Managers can help in this situation. Instead of keeping state off in a state storage the Process Manager is rebuilt by replaying the messages it has previously seen. This gives all of the versioning benefits of working with transient state seen earlier in the book.

Note that the current state may still be stored in a persistent fashion but is considered transient, it can always be rebuilt.

Interesting how the framework maintains Event Sourced Process Managers and how they are tested is very closely related. Tests for Event Sourced Process Managers are generally all structured in the same format.

    public void TestSomethingOnMyProcessManager() {
        //arrange
        var fakePublisher = new FakePublisher();
        var process = new MyProcessManager(fakePublisher);
        process.Handle(new MyMessage1(...));
        process.Handle(new MyMessage2(...));
        process.Handle(new MyMessage2(...));
        fakePublisher.Clear();
        //act
        process.Handle(new MyMessage4(...));
        //assert
        Assert.That(fakePublisher.OnlyContains(x => x.Id == 17));
    }

    A> Note that Event Sourced Process Managers tend to only receive and raise events, they will not as example open up a HTTP connection and directly interact with something directly.

This test written in [AAA style](http://wiki.c2.com/?ArrangeActAssert) shows the pattern that is typically used to test an Event Sourced Process Manager. A Fake Publisher is used for the test. The Fake Publisher will basically take any messages that are published to it and put them into a list in memory that can later be asserted off of.

During the Arrange phase the Process Manager is built and messages are passed into it. It will publish any messages it wishes to the Fake Publisher. At the end of the Arrange phase the last line of code is always fakePublisher.Clear which will clear the Publisher of any messages generated during the Arrange phase. It is only wanted to Assert off the messages that occur during the Act phase not during the Arrange, these would have other tests associated with them.

The Act phase pushes in a single message and only ever a single message. There should only ever be a single line of code in the Act phase and it should be pushing a message into the Process Manager.

The Assert phase will then assert off of the messages that are in the Publisher. As the Publisher was cleared before the Act phase there will only be messages associated with the operation done in the Act phase. All assertions will be done on messages, there should not be an assert off the state of the Process Manager etc.

This same process can be applied to a Process Manager framework. Instead of loading up state then pushing a message into the Process Manager the framework builds the Process Manager and pipes its output to /dev/null replays all of the messages the Process Manager has seen previously into the Process Manager. Once the replay has occured the framework then connects the Process Manager to a real publisher (normally this is just handled by an if statement on the publisher). When the actual message is then pushed into the Process Manager any publishes that it does will be considered real publishes.

In other words the Process Manager is rebuilding its state on every message that it receives, much the same as an aggregate would rebuild its state off of an event stream in the domain model. For performance reasons it is not common to actually rebuild the full state on every message it receives. Instead the current state is cached (memoized) either in memory or in a persistent manner. The main difference is that the cached version is transient and only a performance optimization.

The state of the Process Manager can at any point be deleted and recreated. This is an especially useful attribute when discussing a new version of a Process Manager that needs to replace a currently running Process Manager. All that is required is to delete the currently cached version of state and let the new version of the Process Manager replay through the history to come to its concept of what the current state is.

This also handles the case of the previous version not having in its state things that will be needed by the new version of the process. In the previous case there was an identifier on a message that the old version did not care about but the new one does. When the new version gets replayed, it will see that message and be able to take whatever it wants out of it to apply to its own state.

Unfortunately everything is not puppies and rainbows. There is inherent complexity in bringing a new version of a Process Manager to continue from an old version of the Process Manager. The new version must be able to understand all the possible event streams that result from the previous version, or at least what it cares about. Often times this is not a huge problem but it depends how different the processes are between the two versions. This can result in a large amount of conditional logic for how to handle continuing on the new process where the old process had left off.

This versioning of Event Sourced Process Managers is often used in conjunction with the Takeover pattern. The old version is signaled that it is to terminate. It can still do anything that it wants as part of its termination and it then raises the “TakeoverRequested” message that will start the new version. The main difference with Event Sourced Process Managers and the pattern is that they do not send state when they ask the new version to Takeover.

### Warning

When versioning Process Managers there are many options. The majority of this chapter however is focused on difficult edge conditions. Situations where a running process is changed while it is running should be avoided. All of the cases where a running process is changed while running are niche scenarios, they are included as the topic of the book is versioning. In most circumstances if you are trying to version running Process Managers, *you are doing it wrong*. Stop and think why it is needed, likely you have a modelling issue. Why does the business want to upgrade the processes in place? There are valid reasons but it is worth looking deeper.

Instead focus on releasing new processes in the same way the business does. When releasing a new version the new version takes *future* versions of the process. The currenty running versions stay on the old version. It is not a “new version” of the Process Manager but a new Process Manager. This method of versioning is far simpler, your hair and your sanity will thank you.


---

