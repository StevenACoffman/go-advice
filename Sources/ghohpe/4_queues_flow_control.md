Find my posts on IT strategy, enterprise architecture, and digital transformation at [ArchitectElevator.com](https://architectelevator.com/blog).

Queues are key elements of any asynchronous system because they can [invert control flow](/ramblings/queues_control_flow.html). Because they can re-shape traffic, they enable us to build high-throughput systems that behave gracefully under heavy load. But all that magic doesn't come for free: to function well, queues require flow control. Time for another look at the venerable message queue.

Traffic Shaping
---------------

After developing [a notation and vocabulary to express control flow](/ramblings/queues_control_flow.html), we can describe queues from a new dimension: time. We mentioned arrival and departure rates in the previous post: rates represent the first derivative of a variable over time. Hence, a message arrival rate describes how many messages arrive in a specified interval, e.g., per second. That arrival rate, in turn, varies over time and helps describe the dynamic behavior of any system. For example, peaks in the arrival rate are the typical cause of system overload. Also, the rate at which the arrival rate increases (the second derivative) defines how quickly a system has to scale up to respond to sudden load spikes (to underline its relevance, [AWS Lambda just improved that metric by up to 12x](https://aws.amazon.com/about-aws/whats-new/2023/12/aws-lambda-functions-scale-up/)).

Plotting arrival rates over time provides a great way to depict the power of queues. Queues can flatten that message rate curve, also known as [Traffic Shaping](https://en.wikipedia.org/wiki/Traffic_shaping "Wikipedia") in networking circles:

![](/img/ControlFlow_trafficshaping.png "Queues are traffic shapers")

We see an uneven arrival rate of messages on the left, which turns into a nice and steady processing rate at the right. Naturally, when the arrival drops below the processing rate, the queue shrinks and once it is empty, the processing rate tracks the arrival rate as long as the latter is lower than the former. If the arrival exceeds the processing rate, the queue springs back into action. Many real-life queues allow receivers to process messages in batches instead of one-by-one, but we'll park that aspect for later.

The great advantage of traffic shaping is that it's much easier to build and tune a system that handles the traffic pattern on the right than one that handles the traffic on the left. Specifically, systems that queue requests and process them using a [worker pool](/patterns/messaging/CompetingConsumers.html) behave more gracefully under heavy load: response times will increase, but the system is less likely to collapse under its own weight thanks to constant processing rates. In contrast, synchronous systems can get so busy accepting new requests that they no longer have resources available to process existing requests. The sad result is that as load increases beyond a certain threshold, system throughput actually decreases, as shown in the diagram. It's easy to see that this isn't a good thing.

![](/img/ControlFlow_load.png "system throughput under heavy load")

That's why queues are widely revered as a "buffer" to protect a system from the noisy and spiky world of flash sales, marketing campaigns, or denial of service attacks. Jamming messages into a queue at a high arrival rate is unlikely to overload the system. Meanwhile, the workers are chugging along at the optimal processing rate.

Something's gotta give
----------------------

Now, as architects we know that there are no miracles in system architecture, and the same is true for queues. As any element, a queue has finite capacity. Usually that capacity is high in comparison to processing capacity (like worker nodes) because storing messages is cheaper than running threads. Nevertheless, it is finite. Interestingly, both [Amazon SQS](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/quotas-queues.html) and [GCP Pub/Sub](https://cloud.google.com/pubsub/quotas) state that there is no limit on the number of messages. "Unlimited" is a word that providers rarely use, but if we read carefully, "unlimited" just means that no explicit limit is set, so it doesn't equate to "infinite". We get a hint of that nuance when reading up on [Lambda function limits](https://docs.aws.amazon.com/lambda/latest/dg/gettingstarted-limits.html):

> Each instance of your execution environment can serve an unlimited number of requests. In other words, the total invocation limit is based only on concurrency available to your function.

Similarly, GCP's Pub/Sub doesn't limit the number of messages that are retained, but it does limit the duration:

> By default, a Subscription unacknowledged messages in persistent storage for 7 days from the time of publication. There is no limit on the number of retained messages.

Even in the absence of a hard limit, letting a queue grow arbitrarily may not be a good idea. [Little's Result](https://en.wikipedia.org/wiki/Little%27s_law), one of the fundamental equations of queuing theory, reminds us that the average wait time equals the number of customers in the system divided by the arrival rate. So, wait times goes up with queue length (Little's result assumes a stable system, one where the arrival rate doesn't consistently exceed the processing rate).

Long wait times will impact the user experience. Messages, such as product orders, may no longer be relevant if it takes a year until they are processed. So, your system may technically function, but it no longer addresses your users' expectations, a great case of "all lights are green, the system is down".

Flow Control
------------

For a queued system to perform adequately under consistently high load, it requires some form of [Flow Control](https://en.wikipedia.org/wiki/Flow_control_(data) "Wikipedia").

Wikipedia describes the role of flow control as:

> Managing the rate of data transmission between two nodes to prevent a fast sender from overwhelming a slow receiver.

As we saw above, a queue provides flow control by decoupling the control flow between sender and receiver. But to do so effectively, it ultimately requires additional flow control.

Three main mechanisms provide flow control to avoid filling a queue beyond its manageable or useful size and to keep the system stable:

![](/img/ControlFlow_flowcontrol.png)

*   _Time-to-live (TTL)_: A limited time to live drops old messages from the queue to make room for new messages. This approach works well for messages whose value decreases over time, such as data streams (the CPU utilization from 2 weeks ago may no longer be relevant for an alert) or customer orders (customers may be surprised to have an order processed that they placed weeks ago).
*   A _Tail drop_ (see [Wikipedia](https://en.wikipedia.org/wiki/Tail_drop)) does the opposite by dropping new messages that arrive. This may be appropriate if old messages are too valuable to be dropped or if senders have a feedback mechanism that allows them to realize that messages are being dropped, allowing them to retry later.
*   _Backpressure_ informs upstream systems that the queue isn't able to handle incoming messages so that those systems can reduce the arrival rate, for example by showing an error message to the user.

Although neither option appears particularly appealing, implementing explicit flow control is much better than letting excess traffic take its course. For example, when the AWS Serverless DA team built their [Serverlesspresso app](https://workshop.serverlesscoffee.com/), they found that their backend consisting of two baristas is easily overloaded by demand spikes for free coffee. They could queue requests, but learned that folks would not wait longer than a minute or two for their order to be completed. This led to wasted coffee plus a poor user experience as subsequent users would wait for the baristas to prepare coffees that were never picked up. So, the team resorted to back pressure to not allow new orders when the queue reached a specified size.

Dropping newly arriving messages may seem like a poor business choice—what if one of those messages includes a large order? In reality, though, insufficient flow control will lead to exactly the same effect: if an ecommerce website becomes overloaded, it will try to process orders already in the queue. Meanwhile, new customers can't place an order because the see an error message. Depending on the usefulness of the error message and the UI flow, such an error message could equate to backpressure, e.g., if the shopping basket is maintained for later order placement. Throwing a 500 error when users try to place the order is the very definition of tail dropping.

I am firm believer in [learning about architecture from real life situations](/ramblings/18_starbucks.html), so you will find all forms of flow control in many lines, e.g., those outside a popular club (we once waited over an hour to get into the then-famous Ghost bar on top of the Palms Casino in Las vegas). Customers will likely institute their own Time-to-live policy by eventually giving up. The bouncers may straight out-turn away arrivals in a tail-drop fashion. Or they may set up a secondary queue or holding place of sorts as a form of backpressure.

You'd expect the folks at the large cloud providers to know a thing or two about handling heavy load. Unsurprisingly, we can learn how AWS uses these mechanisms from an [article authored by one of their senior principal engineers](https://aws.amazon.com/builders-library/using-load-shedding-to-avoid-overload/), even though they don't use the common names:

*   "_We've found it's extremely important to place an upper bound on the amount of time that an incoming request sits on a queue, and we throw it out if it's too old_." ≅ **Time-to-Live**
*   "_The Application Load Balancer rejects excess traffic_" ≅ **Tail Drop**

Popular message queue systems like [RabbitMQ have built-in flow control](https://www.rabbitmq.com/docs/flow-control) to protect their queues:

> A back pressure mechanism is applied by RabbitMQ nodes to publishing connections in order to avoid runaway memory usage growth. RabbitMQ will reduce the speed of connections which are publishing too quickly for queues to keep up.

Rate Limiting: Proactive Traffic Shaping
----------------------------------------

Backpressure, TTL, and Tail Drop are the equivalent of a pressure relieve valve in that they engage once things are about to boil over. When you are designing a system that receives messages, you may know its limits upfront and avoid having to back off. That's why many cloud systems have explicit options to limit the delivery rate for push delivery. In a way, you are exerting constant back pressure to limit the rate of message arrival.

EventBridge Pipes has an explicit (fixed) rate limit for external APIs (HTTP calls), through the [Invocation Rate Setting](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-api-destinations.html#eb-api-destination-create) for API Destinations. It invokes Lambda functions directly through an asynchronous invocation without any explicit limit (we assume the built-in drivers have an algorithm similar to GCP). Other event sources like SNS use so-called [Event Source Mapping](https://docs.aws.amazon.com/lambda/latest/dg/invocation-eventsourcemapping.html), which uses its own poller (i.e. _Driver_), configured via maximum batch size (counted in messages) and maximum time (counted in seconds).

GCP Pub/Sub doesn't have an explicit setting, but relies on a [Slow Start Algorithm](https://cloud.google.com/pubsub/docs/push#delivery_rate) that increases delivery speed as long as the downstream system handles the load (indicated by a 99% acknowledgment rate and less than one second push request latency). If the receiver starts to struggle, the sender eases off.

I found a quota of 5000 messages / second for Azure Event Grid push delivery, but I wasn't able to find how it backs off in case that it overloads downstream services. An article published on the [Messaging on Azure Blog](https://techcommunity.microsoft.com/t5/messaging-on-azure-blog/azure-event-grid-s-pull-delivery-performance/ba-p/4003820) shows dynamic traffic ramp-up, but it's for pull delivery:

![](https://techcommunity.microsoft.com/t5/image/serverpage/image-id/532973iD93855C71280BA19/)

Queues require flow control
---------------------------

Understanding flow control allows us to build robust systems that include queues. We can summarize our key learning as:

> Queues invert control flow but require flow control.

A vocabulary to describe the dynamic behavior of asynchronous messaging systems allows us to build reliable systems and makes a useful addition to the existing [messaging patterns](/patterns/messaging) that mainly describe the data flow of messages.

Related posts
-------------

You may enjoy the other posts in this series:

*   [The Many Facets of Coupling](/ramblings/coupling_facets.html)
*   [Event-driven = Loosely coupled? Not so fast!](/ramblings/eventdriven_coupling.html)
*   [Control Flow—The Other Half of Integration Patterns](/ramblings/queues_control_flow.html)
*   [Starbucks Does not Use Two-Phase Commit](/ramblings/18_starbucks.html)