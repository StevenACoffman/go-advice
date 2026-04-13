Find my posts on IT strategy, enterprise architecture, and digital transformation at [ArchitectElevator.com](https://architectelevator.com/blog).

It's hard to imagine an asynchronous system without queues aka [Message Channels](/patterns/messaging/MessageChannel.html). Queues are the core element that provides [location and temporal decoupling](/ramblings/coupling_facets.html) between sender and receiver. But there is quite a bit more behind the little barrel icons: control flow, traffic shaping, back pressure, and [dead letters](/patterns/messaging/DeadLetterChannel.html).

Data Flow is only half the story
--------------------------------

When Bobby and I wrote _Enterprise Integration Patterns_, we naturally spent most of the time on the _data flow_ of messaging systems: how messages are generated, transformed, and routed. Some patterns, especially [Endpoint Patterns](/patterns/messaging/MessagingEndpointsIntro.html) such as [Polling Consumer](/patterns/messaging/PollingConsumer.html) or [Event-Driven Consumer](/patterns/messaging/EventDrivenConsumer.html) clearly have a ring of control flow to it, but we never actually mention the term in the entire book (funnily there is one occurrence in [Recipient List](/patterns/messaging/RecipientList.html), but it's actually a verb and object: "In order for the Recipient List to control flow of information"). The term gets a few mentions in the blog, although [we conclude](/ramblings/74_messaging.html) that "Message-oriented API's steer towards a data flow architecture". So, it's time to make control flow a first-class citizen.

Like so many terms, _Control Flow_ is also an overloaded term. In most imperative programming languages, control flow relates to branching and loop statements. That control flow also exists in distributed systems, where multiple entities execute local control flows. For example, a recipient may be in a loop waiting for data whereas a sender may send a message whenever a user clicks a button. Those control flows assume linearity, meaning one thing happens after another: an `if` statement determines which branch will be executed, but it's only one or the other (let's ignore [speculative execution](https://en.wikipedia.org/wiki/Speculative_execution) at the CPU level).

When designing asynchronous systems, we are interested in how these local control flows relate to each other. For example, when the user pushes a button and the sender sends a message, where does that message stay until the receiver's loop gets around to asking for a message? So, we are looking at the coordination across these local entities that determines the flow of messages:

> In distributed systems, control flow indicates which element actively drives the interaction, for example by sending a message or calling a function.

[Claude](https://claude.ai/) describes the differences between data flow and control flow as the "the what vs. the how": data describes what is being exchanged (and perhaps transformed along the way) whereas control describes how it gets there. A [paper by old friend Schahram Dustdar](https://dsg.tuwien.ac.at/team/sd/papers/Zeitschriftenartikel_2021_SD_Control.pdf) makes the case for emphasizing data flow, ending up with a visual notation where "control flow decisions are encapsulated within the brick implementations", similar to _Enterprise Integration Patterns_ (I forgive Schahram for not citing our book as he already cites Cesare, Olaf, Marlon, and James Lewis).

> Control flow defines operational characteristics of a distributed system like scalability, robustness, or latency.

Although making the control flow implicit (as the above paper suggests) provides clean modeling and evocative diagrams, the run-time characteristics of such systems like scalability, robustness, or latency very much depend on control flow. So, when we build serverless IoT systems, for example, it behooves architects to make control flow a first-class consideration.

Once Again: Drawing the Line
----------------------------

I often introduce distributed system ideas by starting with "2 boxes and a line". What starts so simple, turns out to be quite a bit more interesting once we zoom into the details. That casually drawn line could mean multiple things. Most diagrams show the flow of data from sender to receiver (almost all EIP diagrams do this without explicitly stating it). But the line could also indicate control flow, for example an RPC call. The following diagram illustrates common combinations of data and control flow:

![](/img/ControlFlow_boxes.png "Data flow or control flow?")

Things get interesting with polling: the control flow is from right to left as the receiver actively calls the sender to fetch data.

> Data and control flow may well point in opposite directions. When drawing an arrow, you should specify which flow it indicates.

The picture above is actually somewhat simplified for Messaging, as it doesn't account for [Polling Consumers](/patterns/messaging/PollingConsumer.html) or [Event-Driven Consumers](/patterns/messaging/EventDrivenConsumer.html). Once you place a queue in between sender and received (as most messaging systems do), the control flow becomes yet more interesting:

![](/img/ControlFlow_queue.png "A queue inverts control flow.")

The sender can push messages at its preferred rate into the queue, meaning the control flow runs from the sender to the queue. Likewise, most queues support (or require) [Polling Consumers](/patterns/messaging/PollingConsumer.html), meaning the receiver's control flow points from right to left. Thus, we can express the magical property of queues by means of control flow:

> A queue inverts control flow.

Control Flow: Push and Pull
---------------------------

The absence of control flow from _Enterprise Integration Patterns_ didn't escape some very smart people. Back in 2015 I wrote a blog post titled [Sync or Swim](/ramblings/80_syncorswim.html), which was inspired by the work from old Google colleague, Ivan Gewirtz. Ivan worked in media processing, meaning high volume data processing pipelines. As compute capacities back then weren't what they are today, throughput and efficiency were important considerations. That's why his designs had to consider control flow along with data flow. Being a fan of the integration pattern icons, he realized that they don't express this aspect of system design:

> _Enterprise Integration Patterns_ only shows one half of distributed system design: the data flow.

Fortunately, it's relatively easy to extend the pattern icons to also show control flow by adjusting their shape. If each interaction can be actively pushing or polling data, or passively waiting to be called, we end up with four variations.

![](/img/ControlFlow_pushpull.png "Augmenting EIP icons with control flow")

*   An element can actively push data to a recipient, meaning data and control flow align. We call this element a _Sender_.
*   An element can passively receive data, e.g. from a _Sender_. We call this element a _Sink_.
*   An element can passively provide data, meaning other elements can fetch data from it. We call this a _Source._
*   A _Fetcher_ actively requests (fetches) data from a _Source_. Data and control flow face in opposite directions in this scenario.

Although pushing and pulling are complementary to each other, they aren't the only control flow styles. Many cloud tools also support batch operations, adding cardinality in the mix (you could push or pull single elements or batches). Google Cloud additionally defines an [Export Subscription](https://cloud.google.com/pubsub/docs/subscriber) for Pub/Sub in addition to push and pull. We'll come to these special cases later.

We can now combine the little noses and notches into four common combinations for elements in a [Pipes and Filters](/patterns/messaging/PipesAndFilters.html) architecture:

![](/img/ControlFlow_pushpullcombos.png "Augmenting EIP icons with control flow")

*   A _Puller_ consumes messages actively, as indicated by the little nose pointing left. It "pulls" messages whenever it is ready to do so. However, it doesn't publish messages that it processed until a subsequent message consumer requests a message. The control flow runs from right to left, as indicated by the noses: when a subsequent element requests an element, the Puller in turn fetches a new source message. This makes a _Puller_ synchronous in the sense that message requests that it receives and sends are time-correlated.
*   A _Pusher_ receives messages from an active sender and pushes messages that it processed to the next element. The _Pusher_ is also synchronous in a sense that each message sent coincides with a message received. Control flow (the direction of the noses) matches data flow: from left to right.
*   As we saw, a _Queue_ connects two active elements, one that sends messages to the queue and another one that fetches them from the queue. Our notation makes this role as a control-flow inverter rather obvious. It is also apparent that a _Queue_ behaves asynchronously: the arrival and departure rate are independent. That's why we often equate asynchronous messaging with the use of queues.
*   A _Driver_ is active on both ends, meaning it actively fetches messages and actively sends them to the recipient. Like a queue, it also inverts control flow (the noses are pointing in both directions), but unlike the queue it has full control over the cadence of message fetching and sending. A Driver is therefore helpful as a rate limiter, which we will discuss in more detail later.

The element's role also determines what parameters a user has to specify to define the element's behavior. _Pushers_ and _Pullers_ typically don't take parameters related to control flow as they follow the control flow of the related _Sender_ or _Fetcher_. _Queues_ are passive, so they also don't take control flow parameters (they do take parameters related to flow control, such as retry counts or Time-to-Live / TTL; more on that in a subsequent post). A _Driver_ is active, so it's usually configured via a polling interval (the maximum time between two subsequent message fetches) or batch sizes if batching is supported, as is the case in most high-throughput systems. Batching allows one fetch operation to retrieve multiple messages to reduce the polling overhead. However, large batches may exceed maximum payload sizes and cause inefficient error handling (if processing one message of the batch fails, either the entire batch or a subset have to be returned for retry). As we will soon see, the presence of such parameters often gives a hint at how cloud services are constructed.

Affordances
-----------

The visual language of the little nooks and noses provides an [affordance](https://en.wikipedia.org/wiki/The_Design_of_Everyday_Things#Contents), an element that tells you how you can interact with an object. A toggle button on a light switch is an affordance, and much of Dan Norman's brilliant [The Design Of Everyday Things](http://www.amazon.com/exec/obidos/ASIN/0465050654/enterpriseint-20 "Amazon") discusses how poor affordances make us feel dumb, e.g. a door with a handle that screams "pull" combined with a small sign saying "push".

The _Push_ and _Pull_ affordances help us understand the control flow in asynchronous systems in an almost playful manner. If you want to connect a _Sender_ to a _Fetcher_, you are faced with two noses pointing at each other. The only matching element to connect them is a _Queue_. Once connected, you can insert an arbitrary number of _Pushers_ left of the _Queue_ and _Pullers_ to the right of it.

![](/img/ControlFlow_driver_queue.png "Visual affordances make asynchronous system design simple like assembling Lego blocks")

Likewise, it's easy to see that you need a _Driver_ to connect a _Source_ with a _Sink_. And once again, you can insert _Pushers_ and _Pullers_ al gusto.

Using the icons, finding the matching component and having a suitable name for it becomes so simple that we might feel it's not even worth discussing. In fact, that's exactly the point: finally, distributed system design feels as easy as assembling Lego bricks. Unsurprisingly, most cloud services include these concepts, but often without making them explicit.

Pushing and Pulling in the Cloud
--------------------------------

Serverless integration services offered by cloud providers have to be scalable and robust, so it comes as no surprise that they are particular about push vs. pull control flow and batching of requests. Let's have a look at popular messaging services across the three major cloud providers.

![](/img/ControlFlow_cloud.png "Push and pul in popular cloud services")

Google Cloud leads the pack for supporting and clearly documenting [Push and Pull subscriptions](https://cloud.google.com/pubsub/docs/subscriber) for their Pub/Sub service. And for each kind of access, they document how [flow control](ramblings/queues_flow_control.html) is implemented:

![](/img/ControlFlow_gcppubsub.png)

GCP's Export subscriptions deliver data directly to another service like BigQuery or block storage, where the user does not have to specify a push or pull model.

Azure Event Grid also supports [push](https://learn.microsoft.com/en-us/azure/event-grid/namespace-delivery-retry) and [pull](https://learn.microsoft.com/en-us/azure/event-grid/pull-delivery-overview) delivery.

AWS is less explicit about push or pull, implying the model based on type of service as opposed to giving you a choice. SQS (queue) uses a pull model as we would expect whereas SNS (Pub-Sub) uses a push model. EventBridge Buses also use a push model whereas EventBridge Pipes pulls messages and pushes them, acting as a _Driver_.

Queues Are Hidden in Plain Sight
--------------------------------

Cloud services that support push subscriptions aren't a direct pass through. Instead, the services buffer events (usually after filtering) before passing them to downstream services via active _Drivers_. The drivers work in parallel, arranged in worker pools, to achieve sufficient throughput (diagram on the left):

![](/img/ControlFlow_push_delivery.png)

The parallel processing explains why these services don't maintain delivery order. All major providers make a disclaimer similar to Azure:

> Event Grid doesn't guarantee order for event delivery, so subscribers may receive them out of order.

So, even services that look like a _Pusher_ from the outside (because you feed messages in, and they in turn push messages out), have queues inside. Most documents don't state this explicitly, but once you read the documents (and have a good language to describe control flow), it becomes apparent.

These queues bring much needed stability, but also increase latency. Amazon EventBridge (which I happen to be most familiar with) reduced latency in 2023, but [P90 latencies still hover around 250ms](https://lucvandonkersgoed.com/2022/09/06/serverless-messaging-latency-compared/) according to one heavy user. The [docs even state "about half a second"](https://aws.amazon.com/eventbridge/faqs/#Limits_and_performance), which may render the service not a great fit for latency-sensitive operations. Also in 2023, [Azure published (experimental) stats for Event Grid](https://techcommunity.microsoft.com/t5/messaging-on-azure-blog/azure-event-grid-s-pull-delivery-performance/ba-p/4003820) (which focus more on high throughput than latency) that results in an _average_ latency of 11ms in _Pull_ mode. It would be interesting to see P90/P95 values, as the graphs show some pronounced spikes. Overall, the service design is in line with the fundamentals of asynchronous system design:

> Many serverless event routers optimize for high throughput and operational stability, not low latency.

Asynchronous, message-oriented services perform well under heavy load (more on that in a subsequent post) at the expense of latency. For a cloud provider, it's easy to see that the former trumps the latter as a design goal.

If you're looking to reduce latency, pull delivery can be the better option. Although polling is generally inefficient, both Azure and AWS users reported that pull delivery can yield lower latency than push, largely because the recipient controls the polling rate and batch size. The low latency may come at the expense of higher overhead (and perhaps hitting quota limits) when you use smaller batch sizes for polling, so it's worthwhile considering whether your application can live with the higher latency.

Drivers in the wild
-------------------

Some serverless event services get away without queues because they act as a _Driver_, meaning they have control over the rate at which they fetch messages. [Amazon EventBridge Pipes](https://docs.aws.amazon.com/eventbridge/latest/userguide/pipes-concepts.html) fetches data from sources such as SQS queues or DynamoDB streams (as expected, the diagram on that page shows data flow, not control flow). If you look carefully at the [supported sources](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes-event-source.html) (and you've read the AWS documentation carefully), you'll notice that all sources are _Sources_ in our language, meaning they need an active _Fetcher_ to retrieve data. This explains, why EventBridge Pipes cannot process events from SNS or S3 because they are active _Senders_. Two noses don't connect, however, you could connect them by inserting a Queue, as we saw above.

![](/img/ControlFlow_pipes.png "Amazon EventBridge Pipes has a built-in driver")

EventBridge Pipes having its own driver gives it some advantages. For one, it doesn't require a queue. Queues tend to be expensive and require flow control, as we will see in a future post. When Pipes sends message to a rate-limited target, it can simply slow down the fetch rate of its driver without having to buffer messages. The last advantage is likely the most desirable, in-order delivery:

> If a source enforces order to the events sent to pipes, that order is maintained throughout the entire process to the target.

We quickly notice that the language could be more precise: The events aren't really _sent_, Pipes _fetches_ them. Also, they don't need to be events but can be any form of message. Still, the result is laudable. As we saw, most event routers based on a queue and competing consumers do nor preserve order. Understanding the internal architecture of such services makes the design trade-offs apparent and easier to follow.

Simple Pictures Deepen Our Understanding
----------------------------------------

Assembling the icons with little noses and notches may seem playful, but also trivial in a way. I remember quite some time ago when I shared my messaging toolkit with a customer who for a course in integration patterns. When they looked at the icons and the simple domain language, their feedback was that this is beneath their audience. They somehow wanted things to feel more complicated.

The role of the pattern icons (for data flow and control flow) is to take the noise out and let you focus on the critical design decisions for your asynchronous systems. Once you do that, aspects like rate limiting or out-of-order delivery become apparent and no longer require you to read reams of documentation. That's the power of simple models:

> If your decision seems trivial, you are using the right model.