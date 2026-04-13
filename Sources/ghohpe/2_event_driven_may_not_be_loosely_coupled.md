Find my posts on IT strategy, enterprise architecture, and digital transformation at [ArchitectElevator.com](https://architectelevator.com/blog).

Event-driven architectures (EDA) are seeing a marked renaissance, especially in the context of serverless applications. That's great, but it's also not a new topic by any means. Besides my [EDA presentation at Distributed Event-Based Systems (DEBS) from 2007](https://2007.debs.org/hohpe.pdf), you might recall my ambitious plans to write an [EIP volume on Event Patterns](/ramblings/72_eipvolumes.html) about a decade ago, right after [Conversation Patterns](/patterns/conversation), ahem.

As EDA isn't a new topic, it behooves us to not be caught celebrating our new discovery without acknowledging past challenges and understanding architecture trade-offs. One such trade-off is coupling. "Decoupled" is a frequently cited benefit of EDAs, routinely showing up in [many definitions of EDA](https://aws.amazon.com/event-driven-architecture/): "Decoupled systems that run in response to events". The assumption behind such statements is that "decoupled" is a simple and universally positive thing. But architects know that life isn't that simple. [Coupling is a sliding scale and has many dimensions](/ramblings/coupling_facets.html), so perhaps there's more behind events and coupling!

Messages vs. Events
-------------------

Before we can discuss the magical properties of events, we first want to gain clarity on whether events are something entirely unique, or whether the benefits commonly associated with events exist more broadly. "My chicken laid an egg" is mighty interesting for a person who's only ever seen a chicken but less so for people who know about birds in general (apologies for my poor attempt at making a non-car analogy).

Events are commonly described as a notification of a change in state, or something that took place. Such notifications must be conveyed to other systems, and the natural way to do so is via a message. Whereas some narratives contrast events with messages, _EIP_ takes a clear stance on the relationship between events and messages (Akka and the [reactive manifesto appear to agree](https://doc.akka.io/guide/concepts/message-driven-event-driven.html#_events_are_messages)):

> Events are Messages

Denoted in a UML Class Diagram, the relationship looks as follows:

![](/img/messages_events.png "Events are messages")

EIP underlines this _is-a_ relationship by calling events [Event Message](/patterns/messaging/EventMessage.html). In hindsight, this naming sets an appropriate context. As Eric Evans of DDD fame [pointed out in a recent conversation](https://thenewstack.io/celebrating-20-years-of-domain-driven-design-ddd-and-eip/), an event could be something different depending on the domain. In the domain of integration, it denotes a message. The [Message Construction Introduction](/patterns/messaging/MessageConstructionIntro.html) clarifies the distinction between the interaction style and message semantics (there are actually quite a few hidden gems in these introductions):

> Messaging is an interaction style whereas "Event" describes the semantics of a message.

Therefore:

> When assessing the properties of event-based solutions, the discussion splits into two parts: 1) properties of messaging in general and 2) those of events in particular (contrasted with commands, for example).

To stretch our poor analogy, events are chickens and messages are birds. They both lay eggs but the chickens may lay the most tasty ones.

Messages vs. Channels
---------------------

Event-based architectures are associated with [Publish-Subscribe Channels](/patterns/messaging/PublishSubscribeChannel.html) because multiple recipients might be interested in the occurrence of an event. This stands in contrast to the [Point-to-Point Channels](/patterns/messaging/PointToPointChannel.html) commonly used for commands or document messages.

[![](/img/PublishSubscribeSolution.gif)](/patterns/messaging/PublishSubscribeChannel.html)

Again, we must delineate which characteristics of an event-driven system derive from the properties of the channel rather than the message intent (such as being an Event). For example, if I place a _Command Message_ on a [Publish-Subscribe Channel](/patterns/messaging/PublishSubscribeChannel.html), would I gain the same benefits?

Now some people might think that placing a command message on a [Publish-Subscribe Channel](/patterns/messaging/PublishSubscribeChannel.html) is non-sensical, but life isn't that simple (crocodiles also lay eggs--I promise I'll stop pushing the analogy). Consider the following example:

![](/img/eda_coupling_command_event.png)

Does naming a message `OrderPlaced` (indicating an event has occurred) or `ProcessOrder` (instructing recipients to process the order) make a system more or less coupled? If an e-mail should be sent to a customer as part of processing an order, a recipient could be added to perform this function regardless of the event semantics. We commonly think that the schema of a command is defined by the recipient of the command, e.g. the e-mail sender has an interface for sending e-mails, but to me that is just common usage and not an inherent characteristic. So, it may be less about the semantics of the message than which entity defines the message schema. Hold that thought--we will come back to it later.

If events on a queue still seem odd to you, consider that [Publish-Subscribe Channels](/patterns/messaging/PublishSubscribeChannel.html) are commonly coupled with [Point-to-Point Channels](/patterns/messaging/PointToPointChannel.html) to scale out event consumption via a [Competing Consumers](/patterns/messaging/CompetingConsumers.html) approach. This combination is so common that GCP's Pub/Sub [bundles it into a single service](/patterns/messaging/PublishSubscribeChannel.html#example4):

![](/img/google_pub_sub.png)

> Frequently Publish-Subscribe Channels define the application message flow, whereas Point-to-Point Channels are used for operational aspects.

So, you'll find events being transported via a Point-to-Point Channel more frequently than you might expect. And if you use an event-router like AWS EventBridge, you are already sending Events via a Point-to-Point Channel because the service uses several internally.

Events and Coupling
-------------------

Now that we better understand what events are and how they're related to channels, we can have a much richer conversation about coupling. My [previous blog post laid useful groundwork](/ramblings/coupling_facets.html) by enumerating the facets of coupling:

[![](/img/coupling_dimensions_all.png)](/img/coupling_dimensions_all_large.png)

Along these dimensions, we can now compare the coupling of event-based communication with other types of integration. Based on the discussion above, we don't separate by message semantics but by interaction style:

*   [RPC (Remote Procedure Calls)](/patterns/messaging/EncapsulatedSynchronousIntegration.html), which use a synchronous request-response style interaction.
*   Point-to-Point [Messaging](https://www.enterpriseintegrationpatterns.com/patterns/messaging/Messaging.html) as a one-way, asynchronous interaction style.
*   [Publish-Subscribe](/patterns/messaging/PublishSubscribeChannel.html) messaging, commonly used for [Event Messages](/patterns/messaging/EventMessage.html).

![](/img/eda_coupling_table.png "Evaluating coupling dimensions across integration styles")

Recalling that coupling indicates change propagation, the table includes typical changes that are relevant for each dimension of coupling. A quick scan reveals that the level of coupling for general message-oriented solutions, which may send [Command Messages](/patterns/messaging/CommandMessage.html), is identical for most dimensions.

*   _Temporal Coupling_ propagates changes in latency or availability. Synchronous communication styles are inherently coupled this way (if a provider is slow, the consumer will also be until it times out), whereas one-way, asynchronous ones (Point-to-Point and Pub-Sub alike) are not.
*   _Location Coupling_ propagates a recipient moving to a different location or scaling out to more instances. Message channels decouple location changes because they are (mostly) logical constructs: adding more consumers to a channel (or relocating them) doesn't affect the sender. Some channel names include physical locations (such as cloud availability zones), so they can decouple location changes within a zone but not across. RPC specify a recipient address but can achieve some decoupling through [lookups](/patterns/conversation/ConsultDirectory.html) (like DNS or logical names like [BNS](https://sre.google/sre-book/production-environment/#managing-machines-XQsKIBTq)) or inserting a load balancer.
*   _Space Coupling_ - for this discussion I separate _Location_ from _Space_ and _Topology_ decoupling, with Space decoupling denoting the ability to insert an intermediary without changes to the sender. Message channels also provide this form of decoupling: the sender continues sending messages to the same channel, unaware that it is no longer talking directly to the recipient.
*   _Topology Coupling_ - is used more narrowly here to specifically call out adding a recipient. Publish-Subscribe channels do this free of side-effects, unlike any of the other options.
*   Regardless of the interaction style, sender and receiver have to agree on a _data format and associated semantics_, e.g. a person's age is stored in an integer field called "age", the unit is years, and it counts the years passed since birth, unlike the [traditional Korean system](https://www.bbc.com/news/world-asia-66028606). Most changes (except perhaps adding optional fields) will propagate, meaning sender and receiver are coupled.

Thanks to breaking down the problem and the solution space into multiple dimensions, we come to our main insights. First, while there is a big difference between synchronous and asynchronous interaction styles, the difference between messaging in general and event-driven communication appears relatively minor:

> Event-Driven Architectures are loosely coupled, but very similar to any asynchronous, message-oriented interactions.

Second, the additional topology decoupling for adding recipients is due to the use of a [Publish-Subscribe](/patterns/messaging/PublishSubscribeChannel.html), not from using event semantics.

> Most of the decoupling properties of EDAs derive from the use of Publish-Subscribe Channels, not event semantics.

The key difference in coupling lies in being able to add recipients without affecting existing recipients nor the sender. Two main scenarios make this ability appealing:

*   During the construction or growth phase of the system.
*   If you don't have control over the sender.

Let's look at each one.

Easy Adding Can Make Modifying Harder
-------------------------------------

Being able to add message recipients without side effects is great for a system that is growing, for example, because it is still under construction or because adding recipients is a natural occurrence over the life of the system. But that convenience comes at a price:

*   First, you may not be aware of the evolving system structure (I hinted at this almost 20 years ago in my chapter in [97 Things Every Software Architect Should Know](http://www.amazon.com/exec/obidos/ASIN/059652269X/enterpriseint-20 "Amazon"): [Don't Control, but Observe](https://learning.oreilly.com/library/view/97-things-every/9780596800611/ch48.html "O'Reilly Media")).
*   Second, the more you take advantage of your ability to add recipients, the harder it will become to make a change. Thus, taking advantage of one dimension of decoupling exposes you more to another dimension of coupling.

![](/img/eda_coupling_table_issue.png "Benefiting from one dimension of decoupling can make you suffer from another one.")

The risk is that EDAs demo very well, but become messy as they grow over time. The second risk is related:

> Folks pitching EDA as decoupled may be the ones early in the lifecycle (like vendors) and won't have to live with the consequences of their (coupling) decisions.

Having event hierarchies and distinguishing between internal and external events can help place some sanity back into the event cloud, and some cohesion back into the application.

> Having dependencies between components all over the system is likely a poor design choice and should not be considered "loose coupling".

Inversion of (Coupling) Control
-------------------------------

Architects are known to [zoom in](https://architectelevator.com/architecture/architects-zoom/), so if we look yet more closely at coupling, we find that it not only has dimensions but also a direction. Coupling can be asymmetric, meaning change propagation has a direction. Here we find the real magic of Publish-Subscribe channels: adding subscribers doesn't affect the sender:

![](/img/eda_coupling_pubsub.png "Publish-Subscribe Channels decouple adding recipients from the sender")

In contrast, a [Recipient List](/patterns/messaging/RecipientList.html), which also sends messages to multiple recipients, requires a change to the source to add a recipient. This makes the source subject to change propagation.

[![](/img/eda_coupling_recipientlist.png)](/patterns/messaging/RecipientList.html)

In both cases, sending a message (and likewise changing its format or semantics) travels from left to right, meaning a change in the source very much impacts the recipients (sender to recipient coupling). For stable sources (like system services) that's less of a concern, but it can put a major monkey wrench into your "loosely coupled" applications. In short, a Recipient List propagates changes both ways, whereas a Publish-Subscribe Channel does so only in one direction (from sender to receiver).

We do encounter [Recipient Lists](/patterns/messaging/RecipientList.html) rather frequently in EDAs. For example, AWS EventBridge rules and targets act as recipient list. So, in a way, using an event broker like EventBridge negates some of the "right-to-left" decoupling benefits of Publish-Subscribe channels. You won't have to modify the sender, but you have to modify a central element nonetheless.

We conclude that:

> If you don't have control over the source, Publish-Subscribe channels are ideal.

This property also explains the origin of services like AWS EventBridge. EventBridge evolved from CloudWatch Events, which made system events available for rule-based alerting. Naturally, the sources being AWS services, adding listeners had to free of side effects to the producers.

> In your own application (where you have control over sender and receiver), adding recipients free of side effects is a less-pronounced benefit.

One last observation before we wrap up: decisions about topology and interaction styles are closely related to an application's architecture. Message formats and semantics are associated with application and domain-driven design. So could it be that we're back at [Architect's Dream, Developer's Nightmare??](https://2007.debs.org/hohpe.pdf)