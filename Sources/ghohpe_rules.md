# Enterprise Integration Patterns Rules (Hohpe & Woolf)

Rules derived from *Enterprise Integration Patterns* by Gregor Hohpe and Bobby Woolf.
The single organizing principle: **asynchronous messaging decouples systems across time,
but that decoupling is only real if you design explicitly for partial failure, duplication,
reordering, and the absence of shared state — and loose coupling is the architect's dream,
the developer's nightmare: you gain flexibility and lose debuggability in the same move**.

---

## 1. Choose the integration style before designing the interface

The four integration styles — File Transfer, Shared Database, Remote Procedure Invocation,
and Messaging — produce fundamentally different coupling profiles. Choosing late or by
default usually means choosing RPC, which looks simplest and ages worst.

- **Do** use **messaging** when callers and receivers must be available independently,
  when delivery must be retried until it succeeds, when traffic must be throttled, or
  when load must be distributed across consumers. Messaging is more immediate than File
  Transfer, better encapsulated than Shared Database, and more reliable than RPC. Two
  foundational properties make this work: **send and forget** (the sender hands the
  message to the channel and moves on without waiting for the receiver) and
  **store and forward** (the messaging system persists messages and delivers them when
  the receiver becomes available). Both properties are prerequisites for true
  decoupling — without them you have asynchronous-looking RPC.
- **Do** use **RPC** (gRPC, HTTP) only when you need a synchronous answer before the
  caller can proceed, and when you can tolerate hard coupling of caller and callee lifetimes.
- **Do** use a **shared database** only when multiple systems must see consistent data
  in the same transaction and the schema is owned by one team. Every consumer of the
  schema becomes a hidden dependent — schema changes break them all.
- **Do** use **file transfer** for bulk batch exchange where latency is not critical and
  a missed transfer can be replayed the next day.
- **Do not** default to RPC because it "looks like a function call". The caller blocks,
  the call fails atomically with the callee, and retry logic must be added by hand.
- **Do not** add messaging to a system that requires a synchronous result in the same
  request/response cycle without explicitly modelling the request-reply round-trip.

> **Red flag — Accidental synchrony**: an asynchronous system re-implemented as a
> sequence of blocking calls. You paid the cost of messaging and got the coupling of RPC.

### Coupling is multi-dimensional, not binary

The appropriate level of coupling depends on the level of control you have over the
endpoints. If you own all components, have full test coverage, and high automation,
some coupling is acceptable. If you do not control the other system, you must design
for independent variability.

Coupling has eight distinct dimensions. A system can be loosely coupled on one
dimension and tightly coupled on another — the combination determines actual fragility:

| Dimension | Change that propagates | Mitigation |
|-----------|----------------------|------------|
| Technology | Language / framework / library version | Standard wire formats (JSON, Protobuf) |
| Location | Recipient moves to a new address or scales out | Logical channel names instead of IP / hostname |
| Topology | An intermediary is inserted; a recipient is added or removed | Topology-decoupled channels; intermediary insertion without changing sender |
| Data format | A field is renamed, repositioned, added, or removed | Tagged formats (JSON, XML); Message Translator at the boundary |
| Semantic | Systems agree on a field name but not its meaning | Canonical Data Model; explicit documented contracts |
| Conversation | Message order, retry rules, or timeout behavior is assumed | Explicit protocol; Idempotent Receiver |
| Order | A receiver assumes messages arrive in a specific sequence | Design for out-of-order tolerance; use Resequencer only when required |
| Temporal | A slow or unavailable provider blocks the requester | Asynchronous messaging; graceful degradation |

- **Do** identify which coupling dimensions are most likely to cause a production
  incident before choosing the integration style. A coupling analysis is more useful
  than a blanket "loosely coupled" claim.
- **Do** prefer logical addressing over physical. The location-coupling spectrum,
  from tight to loose:
  `Hard-coded addresses → Host Names / URLs → Logical Names → Topics → Content filtering → Explicit Composition`
  The practical payoff of moving toward the loose end is **composability**: a component
  that makes no location assumptions about its peers allows an intermediary (translator,
  tap, load balancer) to be inserted without any change to the sender. However, the
  spectrum has a cost at the far end: Explicit Composition shifts all change dependency
  to the assembler, which becomes a god class and a maintenance bottleneck.
- **Warning — Hidden coupling**: a system can appear decoupled on one dimension while
  remaining tightly coupled on another. A topology-decoupled serverless application
  whose event payload is source-specific will force all consumers to change when the
  source changes — the topological decoupling is voided by data-format coupling. Name
  the dimension of remaining coupling explicitly; do not declare a system "loosely
  coupled" because one dimension improved.
- **Do** design conversation coupling explicitly. A retry is a conversation with rules:
  are retries allowed? Is the count bounded? Is the receiver idempotent? Does the
  receiver handle duplicates without visible side effects? Each assumption is a
  coupling point that must be documented and tested.

---

## 2. Keep application code unaware of messaging

The messaging system is infrastructure, not business logic. The more business code knows
about channels, headers, message formats, and broker APIs, the harder it becomes to test,
replace, or evolve the integration layer. Two complementary patterns enforce the boundary:
the Messaging Gateway (outbound) and the Service Activator (inbound).

### Messaging Gateway (outbound)

- **Do** wrap all outbound messaging calls in a **Messaging Gateway** — an interface
  whose methods express domain intent (`SubmitOrder(ctx, order Order) error`) not
  messaging primitives (`Publish(ctx, channel string, body []byte) error`).
- **Do** define the gateway as an interface. The real implementation uses the broker;
  the test implementation invokes callbacks synchronously in the same goroutine, with
  no broker infrastructure required.
- **Do** translate broker-specific errors inside the gateway into domain or application
  errors. Business code must not import broker packages.
- **Do** choose between a blocking gateway (simpler caller code, blocks a goroutine
  waiting for the reply) and an asynchronous gateway (returns immediately, delivers
  result via callback or channel). Use the blocking form unless concurrency demands otherwise.
- **Warning**: An asynchronous gateway that stores pending callback references will
  memory-leak if replies never arrive. Define a timeout and an expiry policy for every
  outstanding request entry. A reply that never arrives must eventually expire the
  entry — not hold it in a map forever.
- **Do** chain gateways when needed: a low-level gateway abstracts the broker API;
  a higher-level gateway exposes domain semantics. Each layer hides one concern.

### Service Activator (inbound)

- **Do** use a **Service Activator** to connect inbound messages to a service
  implemented as a plain synchronous function. The service should be written as if it
  will always be called directly — it accepts plain values and returns plain values,
  with no awareness of channels or message formats.
- **Do** design the Service Activator so the underlying service is callable both via
  messaging *and* via direct invocation. The service is written as a plain synchronous
  function; the activator wraps it. This dual-invocation property is the pattern's
  primary value: the service can be exercised in unit tests with no broker, yet is
  also reachable asynchronously from any messaging endpoint.
- **Do** design the activator to extract parameters from the message, invoke the
  service, and send any reply. If the message format is invalid, route to the Invalid
  Message Channel. If the service returns a business error, handle it at the application
  level — these are two distinct failure modes with different destinations.
- **Do** make the activator a Transactional Client when the service is transactional,
  so that message consumption and service invocation participate in the same transaction.
- **Do** run multiple Service Activators for the same service as Competing Consumers
  or coordinate them via a Message Dispatcher to increase throughput.

### Messaging Mapper

- **Do** use a **Messaging Mapper** to convert between domain structs and message
  payloads. Neither the domain struct nor the broker adapter should contain the other's
  field names. The mapper is invisible to both sides.
- **Do not** scatter `json.Marshal` / `proto.Marshal` + channel-name constants across
  business handlers. Centralize format decisions in the mapper.
- **Do not** use language-native serialization (Go `encoding/gob`, Java Serializable)
  as the wire format for inter-service messages. Native serialization mirrors the domain
  struct one-to-one; changing the struct breaks every consumer.

> **Red flag — Broker bleed**: business logic imports a Pub/Sub or AMQP package and
> constructs messages inline. The business code is now coupled to the broker.

---

## 3. Design messages to be self-contained and typed

A message travels through systems that may be geographically separated, independently
deployed, and incapable of making additional calls to retrieve context. It must carry
everything a receiver needs to act on it.

- **Do** separate header data (routing, correlation, expiration, type, message ID) from
  body data (application payload). Routers and brokers read headers; consumers read bodies.
- **Do not** put routing or infrastructure metadata in the message body. A router that
  must deserialize the body to decide where to send a message cannot work without
  understanding the business schema.
- **Do** assign a unique, system-generated message ID to every message. Use this ID for
  de-duplication and message tracking — not a business key.
- **Do** use separate message types for distinct intents:
  - **Command Message** — tells the receiver to do something specific; must contain all
    data necessary to perform the command. Send on a point-to-point channel so exactly
    one receiver acts.
  - **Event Message** — announces that something happened; the sender does not know or
    care who receives it. Send on a publish-subscribe channel so all interested parties
    are notified.
  - **Document Message** — transfers a data structure; the receiver decides what to do
    with it. Send on a point-to-point channel when exactly one receiver should process it.
- **Do** include a **Format Indicator** (version number or schema reference) in every
  message so receivers can handle format evolution without breaking.
- **Do** set **Message Expiration** on time-sensitive messages. An inventory count or
  stock price that is minutes old may be worse than no data — process or discard it,
  not both.
- **Do not** overload a single field with dual semantics (e.g., using the order ID as
  both the business key and the correlation identifier for request-reply tracking).

### Envelope Wrapper (adding system metadata without modifying the payload)

- **Do** use an **Envelope Wrapper** when you must add system-level metadata — routing
  headers, security tokens, audit fields — to a message whose payload you cannot or
  should not modify. The original message becomes the body of an outer envelope; the
  infrastructure reads the envelope and the consumer unwraps it to get the original.
- **Do** ensure that an unwrapper is deployed at every destination that will receive
  wrapped messages. A consumer that receives an envelope it does not unwrap will see
  the envelope as its payload, silently breaking its schema expectations.
- **Do not** use the Envelope Wrapper when a canonical header section in the message
  format is available. Adding envelope metadata directly to the header section is
  cleaner — the wrapper adds an extra serialization layer and an extra unwrapping step
  at every receiver.

### Message Sequence (large data split across messages)

- **Do** use a **Message Sequence** when data exceeds message-size limits. Every
  message in the sequence must carry: a sequence identifier shared by all messages in
  the sequence, a position indicator, and an end-of-sequence flag.
- **Do** send all messages in a sequence within a single transaction to guarantee
  they arrive in order. A partial send leaves the receiver in a state it cannot resolve.
- **Do** pair a Message Sequence with a Resequencer if parallel processing may reorder
  the messages before they reach the reassembler.

> **Red flag — Payload-dependent routing**: a router must deserialize the message body
> to determine the destination. Body and routing have become entangled.

---

## 4. Choose the right channel type

The channel type determines delivery semantics for every message on it. Choosing the
wrong type produces subtle bugs: messages silently processed twice, or processed by
zero consumers.

- **Do** use a **Point-to-Point Channel** for commands and tasks. Exactly one consumer
  receives each message. Multiple consumers on the same point-to-point channel compete
  for messages (load balancing), not duplicate them.
- **Do** use a **Publish-Subscribe Channel** for events. Every subscriber receives a
  copy. This is not load balancing — each subscriber receives the full event stream.
  A subscriber must be active when the message is published or it will miss the message
  (unless it is a Durable Subscriber).
- **Do not** use multiple consumers on a Pub/Sub channel to distribute load. Each
  consumer will receive every message; you will duplicate work, not share it.
- **Do** put only one message type on a single channel (**Datatype Channel**). Consumers
  on that channel know the type from the channel name alone.
- **Do** prefer combining related message subtypes on one channel with a Selective
  Consumer over creating a separate channel per subtype. Channels consume resources and
  add management overhead. The cut-off: create a separate channel when the subscriber
  populations for the subtypes are entirely disjoint (different teams receive different
  subtypes). Use a Selective Consumer when the same subscribers handle multiple subtypes
  with different logic.
- **Do not** create a channel per event subtype when a handful of related subtypes share
  the same subscriber population. A channel explosion (dozens of low-traffic channels)
  wastes resources and confuses operators about which channel to watch.

### Connecting external systems

- **Do** use a **Channel Adapter** when you cannot modify an external application to
  add messaging code directly. The adapter accesses the application's existing API or
  data and translates between the application's native format and the message format.
- **Do** use a **Messaging Bridge** when you must connect two different messaging
  systems. The bridge consumes messages from one system and republishes them to the
  other. A bridge introduces latency and the possibility of duplicate messages if the
  bridge restarts mid-delivery — design consumers downstream to be idempotent.

### Error channels

- A message the *receiver* cannot process (wrong format, missing required fields,
  failed business validation) is an **invalid message**. Route it to the **Invalid
  Message Channel** — an application-level error channel.
- A message the *messaging system* cannot deliver (expired, exceeded retry limit, no
  route exists) goes to the **Dead Letter Channel** — a system-level error channel.
  These are distinct responsibilities and must not be conflated.
- **Do** monitor the Dead Letter Channel in production. A growing DLC is always a
  symptom: consumers are failing, messages are expiring faster than they are consumed,
  or a poison-pill message is causing repeated crashes.
- **Warning**: A Dead Letter Channel that fills without being drained can crash the
  messaging system. Treat DLC depth as a critical metric.

> **Red flag — Channel overloading**: one channel carries two distinct message types
> that require separate consumer logic. The channel is doing two jobs.

---

## 5. Never drop a message silently

In a loosely coupled system a discarded message produces no error at the call site. The
failure is invisible until a downstream system notices missing data, which may be days later.

- **Do** route every message that cannot be processed to an explicit error channel —
  Invalid Message Channel for application-level failures, Dead Letter Channel for
  system-level failures. Never discard silently.
- **Do** use **Guaranteed Delivery** (broker-level persistence) when message loss is
  unacceptable. Persisted messages survive broker restarts; in-memory messages do not.
  Accept the throughput trade-off.
- **Do not** use in-memory-only channels for messages whose loss would require manual
  intervention to detect and repair.
- **Do** log the message ID and channel name whenever routing to an error channel so
  operators can diagnose and replay the message.

---

## 6. Design every receiver to be an Idempotent Receiver

Most messaging systems guarantee *at least once* delivery, not exactly once. Network
timeouts cause the sender to retransmit; broker restarts replay messages that were not
yet acknowledged. Duplicates are normal operating conditions, not bugs.

- **Do** design every receiver to produce the same outcome whether it processes a
  message once or ten times.
- **Do** prefer **semantic idempotency** where possible: `SET balance = 100` is
  idempotent; `ADD 10 TO balance` is not.
- **Do** implement **explicit de-duplication** when semantic idempotency is
  insufficient: record the message ID in a durable store (database unique index, Redis
  SET NX) and reject messages whose ID has already been processed.
- **Do** use a dedicated message identifier field for de-duplication, not a business
  key. De-duplication is an infrastructure concern; conflating it with business keys
  causes problems when business rules change.
- **Do** size the de-duplication window to match the sender's retry horizon: if a
  sender may retry for up to 24 hours, the de-duplication record must survive at least
  24 hours. Decide whether to persist the history across process restarts based on
  your reliability requirements — in-memory de-duplication is lost on crash.
- **Do not** use a database unique constraint as the de-duplication mechanism unless
  necessary. It works, but it creates dual semantics on the key field: the same column
  simultaneously enforces a business constraint and an infrastructure one. When business
  rules around that key change, the de-duplication invariant changes with them.
- **Do not** rely on the messaging system to guarantee exactly-once delivery even if
  the documentation claims it. Design for at-least-once and verify that extra deliveries
  cause no harm.

> **Red flag — Non-idempotent handler**: a handler that sends an email, charges a
> card, or increments a counter without a de-duplication guard. A single retry
> causes a duplicate charge or a duplicate notification.

---

## 7. Use transactions correctly — and know where they break down

Transactions can span a message receive and a database write, making them atomic. They
cannot span a network boundary invisibly. Using transactions incorrectly causes silent
message loss or duplicate processing.

### When transactions are the right tool

- **Do** use a **Transactional Client** when you need atomicity across messaging and
  other operations. Valid use cases:
  - *Receive + database write*: receive a message and update a database as a single
    unit — the canonical use case. If the write fails, the message returns to the queue.
  - *Send + receive pair*: receive a request and send a reply atomically.
  - *Message batch*: send or receive a group of related messages as a single unit.
  - *Workflow step*: send a work request and acquire a work item atomically.
- **Do** understand the mechanism: with a transactional receiver, the message is not
  removed from the queue until the transaction commits. On failure or rollback, the
  message is returned to the queue and will be redelivered.

### Where transactions break down

- **Do** call the equivalent of `tx.Rollback()` immediately after a successful
  `BeginTx` and ensure it is deferred. Without it, a panic or early return leaves the
  transaction open and holds queue locks.
- **Do not** use an Event-Driven Consumer (callback/`onMessage`) with a Transactional
  Client. If the callback panics or returns an error, the broker delivers the next
  message, silently losing the failing one. Use a Polling Consumer with a transactional
  loop instead.
- **Do not** attempt to send a request and wait for the reply within the same
  transaction. The request is not actually sent until the transaction commits; the reply
  will never arrive while the transaction is open.
- **Do not** hold a transaction open while calling an external HTTP service or waiting
  for user input. Transactions must be short; long-held transactions block queue
  resources and cause cascading failures.
- **Do not** use Competing Consumers with a Transactional Client on systems where
  multiple consumers can attempt the same message. The consumer that loses the race does
  all processing work and then must roll back — measure actual performance before
  committing to this design.
- **Do** give each performer in a Message Dispatcher its own session and transaction.
  A dispatcher that shares one session across performers will roll back all performers
  when any one fails.

> **Red flag — Long transaction**: a transaction that spans an outbound HTTP call,
> user interaction, or external acknowledgment. Queue locks accumulate and the
> system eventually deadlocks.

---

## 8. Model fan-out, ordering, and reassembly explicitly

When one message must be processed by multiple independent systems, or when a composite
message must be split and reassembled, the semantics of delivery and ordering matter.
Parallel processing destroys ordering; ordering must be explicitly restored.

- **Do** use a **Recipient List** when a fixed or computed set of receivers must each
  get a copy and you need to track who received it. Remove the recipient list from the
  message after routing so downstream components do not act on infrastructure metadata.
- **Do** use a **Splitter** to decompose composite messages into independent items, each
  sent as its own message. Inject a correlation ID and sequence position into every
  sub-message so an Aggregator can reassemble results.
- **Do** define the Aggregator's **completion condition** explicitly. The condition
  depends on the use case; the four main options are: (1) all N sub-messages received;
  (2) first N responses when only partial results are needed; (3) best result by a
  criterion (e.g. lowest price, fastest reply); (4) timeout — publish whatever arrived.
  Choose the condition at design time; an Aggregator that never satisfies its completion
  condition will accumulate partial aggregations indefinitely.
- **Do** define a policy for **late-arriving messages** after an aggregate has been
  published: route them to the Invalid Message Channel or discard with a log entry.
  Never silently drop them without a trace.
- **Do** name the **Composed Message Processor** pattern when you are combining a
  Splitter, per-item routing, and an Aggregator to process a composite message whose
  sub-items need different processing paths. It is the canonical solution for
  "split a message, route each piece differently, reassemble the results."
- **Do** use a **Scatter-Gather** when you need to broadcast a request to multiple
  recipients simultaneously and collect their replies. Set an explicit timeout. Apply
  the same completion-condition options as an Aggregator: all responses, first N,
  best by criteria, or timeout. Discard late replies that arrive after the aggregate
  has been published.
- **Do not** use a Scatter-Gather that waits indefinitely for all responses. One slow
  upstream stalls the entire aggregate.

### Restoring order after parallel processing

- **Do** use a **Resequencer** when parallel processing (Competing Consumers, Splitter
  fan-out) has destroyed a message ordering that downstream components require.
  Unlike an Aggregator (many messages → one), a Resequencer consumes and publishes the
  same number of messages — it only delays and reorders them.
- **Do** give every Resequencer an explicit timeout and a policy for what to do when
  the timeout expires (publish what is available, route the incomplete sequence to an
  error channel, or raise an alert). Without a timeout, a single missing or delayed
  message blocks the Resequencer permanently.
- **Do not** resequence unless order is genuinely required by downstream logic.
  Resequencing adds latency: earlier messages must wait for later ones to arrive before
  any can be forwarded.

> **Red flag — Accumulating aggregator**: an Aggregator or Resequencer whose completion
> condition can never be satisfied for certain inputs. Partial aggregations or buffers
> grow without bound until memory is exhausted.

---

## 9. Separate routing from processing; choose how routing state is carried

A component that routes messages should not also transform them. A component that
transforms messages should not also decide where they go. Beyond that separation, the
choice of where routing state lives — in the message itself or in an external process —
determines how the pipeline can evolve.

- **Do** put routing logic in a **Message Router**. Routers read message metadata and
  headers; they do not modify message content.
- **Do** put transformation logic in a **Message Translator**. Translators are
  stateless, have one input and one output, and know nothing about routing.
- **Do** handle the no-match case in every Content-Based Router: route to the Invalid
  Message Channel or include a default output channel. A router that silently drops
  unmatched messages is a bug.
- **Do not** embed routing decisions inside processing components ("if business field X,
  also publish to channel Y"). Routing is a separate concern.

Choose the right router for the cardinality of your routing decision:

| Pattern                    | Input → Output | Notes |
|----------------------------|---------------|-------|
| Content-Based Router       | 1 → 1         | One destination per message; mutually exclusive routes |
| Message Filter             | 1 → 0 or 1    | Infrastructure-level; drops non-matching messages before consumers see them |
| Recipient List             | 1 → N         | Explicit, computed set of destinations; sender tracks recipients |
| Splitter                   | 1 → N         | Decomposes composite into N independent messages |
| Aggregator                 | N → 1         | Combines N related messages into one |
| Resequencer                | N → N         | Same count; restores ordering only |

### Static vs dynamic routing

- **Do** use a **Content-Based Router** for routing rules that are known at deployment
  time and change infrequently. It is efficient (each message goes directly to the
  correct destination) but requires updating the router whenever a new recipient is
  added. A CBR that encodes many business rules inline can become a point of frequent
  maintenance — if the routing table changes often, externalize the rules into a
  configurable data structure (database, config file) rather than hardcoding them in
  the router.
- **Do** use a **Dynamic Router** when routing rules change frequently or when
  recipients should self-register their capabilities at runtime. The router maintains a
  routing table updated by control messages from recipients; no router redeployment is
  needed when a new consumer comes online. Accept the cost: the routing table itself
  becomes a source of errors and must be tested and monitored as rigorously as the
  routes it describes.
- **Do** consider using a Publish-Subscribe Channel with Message Filters as a reactive
  alternative to a Content-Based Router. The CBR has central control and predictive
  routing; the Pub/Sub + filter array has distributed control and reactive filtering.
  CBR is more efficient; Pub/Sub + filters requires no change when new consumers arrive.

### Routing state: in the message vs external

- **Do** use a **Routing Slip** (processing sequence embedded in the message header)
  when the sequence of processing steps varies per message but is known at message
  creation time. Each component pops the next step off the slip and forwards accordingly.
  Do not use a Routing Slip when the sequence is fixed — use a simpler static pipeline.
  Design the slip to be modifiable only by trusted components.
- **Do** use a **Process Manager** when the processing sequence is conditional,
  non-linear, or depends on intermediate results. The Process Manager externalizes
  routing state: it holds one stateful instance per in-flight request, decides the next
  step based on intermediate outcomes, and handles failures, timeouts, and compensation.
  That state must be persisted to survive restarts.

---

## 10. Decouple with Publish-Subscribe; re-couple intentionally with Request-Reply

Publish-Subscribe is the purest form of loose coupling: publishers do not know receivers
exist. Request-Reply is the controlled re-introduction of synchronous coupling for the
cases that require it.

- **Do** use a Publish-Subscribe Channel for event notifications where the publisher
  should remain unaware of who, if anyone, is listening.
- **Do** use **Durable Subscribers** when a consumer must not miss messages published
  while it is temporarily offline. Undrained durable subscriptions accumulate messages
  indefinitely — combine with Message Expiration.
- **Do** explicitly unsubscribe a Durable Subscriber when the consumer is permanently
  decommissioned. Failing to do so leaves the broker accumulating messages for a
  consumer that no longer exists.
- **Do** use **Request-Reply** when a caller genuinely needs a result before it can
  continue. Include a **Return Address** in the request (where to send the reply) and a
  **Correlation Identifier** (how to match the reply to the request). The receiver
  **must** read the Return Address from the request message and send the reply there —
  never hardcode the reply channel in the responder. A hardcoded reply destination
  makes the responder tightly coupled to a single caller and cannot work with
  per-caller reply channels.
- **Do** consider asynchronous request-reply when the caller does not need to block:
  send the request, proceed with other work, and handle the reply via a callback or
  a dedicated reply channel that the caller polls later.
- **Do** use per-caller reply channels, not a shared global reply channel. Multiple
  callers on a shared reply channel will consume each other's replies.
- **Do** define a timeout for every Request-Reply interaction. A reply that never
  arrives must produce an explicit error, not a blocked goroutine.
- **Do not** use a business key as the Correlation Identifier. Business keys change
  with business rules; infrastructure correlation identifiers must be stable.

> **Red flag — Shared reply channel**: two concurrent requestors use the same channel
> to receive replies. Each risks consuming the other's response.

---

## 11. Control consumer concurrency and throughput explicitly

In an asynchronous system, producers and consumers run at independent rates. An
unconstrained producer will eventually overwhelm any consumer. Both the lower bound
(throttling) and upper bound (scaling) of consumer throughput must be designed in.

- **Do** use a **Polling Consumer** when precise throughput control is required. The
  consumer pulls only when ready; it cannot be pushed into overload. To monitor multiple
  channels from one goroutine, use non-blocking receive calls cycling across channels.
- **Do** use **Competing Consumers** (multiple independent consumers on the same
  point-to-point channel) to scale throughput horizontally. Each competing consumer must
  run in its own goroutine with its own broker connection/session — not just its own
  subscription on a shared connection. Spreading competing consumers across multiple
  machines provides near-unlimited throughput scaling.
- **Do** use a **Message Dispatcher** instead of Competing Consumers when: (a) the
  broker handles multiple sessions poorly, (b) different message subtypes require
  different handlers, or (c) you need dispatch logic the broker does not provide.
  Performers in a dispatcher typically run in the same process; competing consumers can
  be distributed across machines.
- **Do** limit the number of concurrent consumers to a number your downstream
  dependencies (database, external API) can sustain — not to the number that drains the
  queue fastest.
- **Do not** use an Event-Driven Consumer (push callback) when the consumer's
  processing rate is significantly slower than the publisher's rate. The backlog of
  pending callbacks will grow unboundedly. Use a Polling Consumer with a bounded worker
  pool.
- **Do** distinguish **Message Filter** from **Selective Consumer**: a Message Filter
  is part of the messaging infrastructure — it removes non-matching messages before any
  consumer sees them and can be evaluated server-side to avoid network traffic entirely.
  A Selective Consumer is inside the receiving application — it receives messages and
  discards what it does not want. The filter is cheaper (messages never leave the
  broker) but changes require redeploying infrastructure; the selective consumer is more
  flexible but wastes network bandwidth on discarded messages.
- **Do** use a **Selective Consumer** to direct different message subtypes to different
  handlers on the same channel when the filtering decision cannot or should not be made
  at the broker level. Specify the selection predicate in the message header, not the
  body, so the broker *can* evaluate it without deserializing the payload if a server-
  side filter is added later.
- **Security note**: Selective Consumers do not provide security isolation. A consumer
  can change its selector to receive messages intended for another consumer. Use separate
  Datatype Channels to enforce message-type security boundaries.
- **Do** account for batch size tradeoffs when configuring consumers. Larger batches
  reduce per-message overhead but complicate error handling: if one message in a batch
  fails, the entire batch must be returned for retry or split into sub-batches, negating
  the efficiency gain. Choose batch sizes based on error handling cost, not just throughput.

---

## 12. Translate at the boundary; normalize in the middle

Every system has its own data model. The number of translators required to connect N
systems all-to-all grows as N×(N-1). A Canonical Data Model reduces this to 2×N, at
the cost of double translation on every message and a shared dependency on the model.

- **Do** translate between an external system's data format and the canonical format
  at the **Channel Adapter** — the boundary point. The adapter accesses the application's
  existing API or data feed and handles format conversion. Use a Channel Adapter when
  you cannot modify the external application to add messaging code directly.
- **Do** use a **Messaging Bridge** when integrating two separate messaging systems.
  The bridge consumes messages from one broker and republishes them to the other. A
  bridge introduces latency and the potential for duplicates on restart — design
  downstream consumers to be idempotent.
- **Do** define a **Canonical Data Model** that is independent of any single
  application's native format. The canonical model is a shared contract; no single
  application owns it.
- **Do not** use any one application's native format as the canonical format. The
  application that owns the format has zero translation cost; every other application
  pays double. The asymmetry breeds resentment and non-compliance.
- **Do** account for the full cost of the Canonical Data Model: every message requires
  two translations (source app → canonical, canonical → destination app), adding
  latency and serialization overhead. The canonical model also becomes a shared
  dependency — changing it requires updating all translators across all teams.
- **Do** use a **Normalizer** (Router + set of translators) when the same semantic
  message arrives in multiple incompatible formats from different sources. After
  normalization all downstream components see a single format.
- **Do** use a **Content Enricher** to add missing data from an external source when
  the originating system cannot provide it. The enricher adds fields; it does not
  modify existing fields. The external source becomes a dependency — if it is
  unavailable, enrichment fails and the pipeline stalls.
- **Do** use a **Claim Check** when large payloads (binary files, large JSON documents)
  would saturate the messaging infrastructure. Store the payload externally, pass a
  token in the message, and retrieve only when needed. Define a retention and cleanup
  policy; unclaimed data must not accumulate indefinitely.
- **Do not** over-filter with a **Content Filter**. Removing fields that one consumer
  does not need may remove fields that another consumer does need. Apply content
  filtering as close to the specific consumer as possible.

---

## 13. Build observability in from the start

Asynchronous messaging systems are hard to debug after the fact. A message can travel
through a dozen components before producing a wrong result; without instrumentation, the
path is invisible.

- **Do** attach a **Message History** header to every message: each component that
  handles the message appends its identifier before forwarding. Use this to trace paths
  and to detect routing loops in Publish-Subscribe topologies (a component can inspect
  its own ID in the history and discard a message it already forwarded).
- **Do** use a **Wire Tap** (a Recipient List with two outputs) to send copies of
  messages to a **Message Store** without disturbing the primary flow. Send in
  fire-and-forget mode; the tap must not slow the main pipeline.
- **Do not** correlate primary-flow and tap messages by message ID. Most brokers assign
  a new ID to the copy produced by a Wire Tap. Use a separate, preserved correlation
  field that is copied from input to output.
- **Do not** convert a Point-to-Point Channel to Publish-Subscribe to enable tapping.
  This changes delivery semantics and can cause consumers to each receive every message.
- **Do** implement a purge policy for the Message Store. Without it, the store grows
  indefinitely. Archive or delete records based on age or volume.
- **Do** use a **Smart Proxy** to instrument request-reply latency for a service that
  sends replies to a Return Address. The proxy intercepts the request, replaces the
  Return Address with its own, intercepts the reply, records the latency, and forwards
  the reply to the original requester.
- **Do** use a **Control Bus** on channels separate from application message flow to
  send configuration, receive heartbeats, activate detours, and collect statistics.
  Give control messages higher priority than application messages so they reach a
  congested component even when its queue is full.
- **Security**: Control Bus messages must be subject to stricter security policies than
  application messages. A malformed or adversarial control message can misconfigure or
  bring down a component. If the messaging infrastructure itself fails, the Control Bus
  fails with it — for critical operational commands, consider a separate out-of-band
  control channel that does not depend on the same broker.
- **Do** define component heartbeats at a known interval. Monitor for absence — a
  missing heartbeat is a signal; a delayed heartbeat is a warning.
- **Do** inject **Test Messages** into production pipelines at regular intervals to
  verify end-to-end correctness, not just component liveness. Tag test messages with a
  dedicated header field — never with a special value in a business field.
- **Warning**: Do not use Test Messages in stateful components that write to a database
  or send external notifications without filtering them out. Test orders will appear in
  revenue reports.
- **Do** implement a **Channel Purger** for test environments to drain all channels to a
  known empty state before each test run. In production, store purged messages before
  discarding so they can be replayed after fixing the root cause.

> **Red flag — Invisible path**: a message produces a wrong result but no component
> logged that it processed the message. There is no way to reconstruct what happened.

---

## 14. Keep the Pipes and Filters pipeline composable

The Pipes and Filters architectural style — independent processing steps connected by
channels — is the foundation of every integration pipeline. Its value comes from the
independence of its stages.

- **Do** make each filter independent: it reads from one input channel and writes to
  one output channel and knows nothing about the filters before or after it.
- **Do** exploit the parallel scalability of the pattern: because each filter is
  independent, multiple instances of the same filter can run concurrently on separate
  goroutines or machines, processing messages in parallel. Scale the bottleneck filter
  independently, without scaling the whole pipeline.
- **Do** design filters to be stateless where possible. Stateful filters (Aggregators,
  Resequencers, Process Managers) require persistent state to survive restarts — make
  that storage dependency explicit.
- **Do** give every stateful component (Aggregator, Resequencer) a timeout and a policy
  for what to do when the timeout expires. Without a timeout, a single missing message
  can block the component permanently.
- **Do** treat a pipeline as a unit of deployment. A filter that routes, transforms,
  enriches, and aggregates is not a pipeline — it is a monolith with channels on the sides.
- **Do** prefer fine-grained, recombinable filters over coarse-grained all-in-one
  processors when reuse and independent deployability matter. Accept the trade-off:
  more channels, higher per-message latency.
- **Do not** build transformation logic into a Message Router. A router that modifies
  message content is two components in a trench coat.
- **Do** treat **control flow** as a first-class design concern in every pipeline,
  alongside data flow. Control flow — which element actively drives the interaction —
  determines whether the pipeline degrades gracefully under load, whether order is
  preserved, and whether rate limits can be enforced. Data flow diagrams alone do not
  capture these properties. See §16 for the full control flow model (push/pull roles,
  queue as control-flow inverter, Driver pattern, traffic shaping, and flow control).

> **Red flag — God filter**: a single processing component that routes, transforms,
> enriches, and aggregates. It cannot be tested in isolation and cannot be reused.

---

## 15. Coupling has eight dimensions — "loosely coupled" without qualification is meaningless

Coupling is non-binary and multi-dimensional. A system can be loosely coupled along one
dimension and tightly coupled along another simultaneously — usually without the designer
realizing it. "Event-driven architectures are loosely coupled" is a statement that needs
qualifying: *with respect to which dimension? Compared to what?* The trade-off depends
on the answer.

**The basic theorem of coupling**: the appropriate level of coupling depends on the level
of control you have over the endpoints. If you own both systems and have good test coverage,
high coupling may be acceptable — a rename refactoring propagates instantly. If you control
neither endpoint, you need far more insulation.

The eight dimensions and their mitigations are tabulated in Rule 1. Two require
particular attention:

- **Semantic coupling** is the hardest to detect and the most expensive to fix.
  Format-oriented tools cannot resolve it: ZIP codes don't map cleanly to area codes;
  weekly reports don't align with 30-day intervals. Declaring fields `optional` in a
  schema shifts coupling from the IDL to the endpoint implementation — the contract
  looks flexible while the actual behavior is rigid.
- **Conversation coupling** is frequently overlooked. A retry is a conversation with
  protocol rules: are retries allowed? Is the count bounded? Is the receiver idempotent?
  Does the receiver deduplicate? Each assumption is a coupling point that must be
  documented and tested, not left implicit.

**Which integration style addresses which coupling dimension:**

The big leap in decoupling is from synchronous RPC to asynchronous messaging, not from
Point-to-Point messaging to Pub/Sub. The only dimension that distinguishes them is
topology decoupling for *adding recipients*:

| Dimension | RPC | Point-to-Point Messaging | Pub/Sub |
|---|---|---|---|
| Temporal | ✗ Blocked by slow/unavailable provider | ✓ Sender and receiver independent | ✓ Sender and receiver independent |
| Location | Partial (DNS/load balancer helps) | ✓ Logical channel name | ✓ Logical channel name |
| Space (insert intermediary) | ✗ Sender must change | ✓ Sender unaware | ✓ Sender unaware |
| Topology (add recipients) | ✗ Sender must change | ✗ Sender must change | ✓ Free of side effects |
| Data format | ✗ Both styles coupled equally | ✗ Both styles coupled equally | ✗ Both styles coupled equally |

**EDA and Pub/Sub coupling:**

- **Do** recognize that events are messages — "Event" describes the semantics of a message
  (something happened), not the interaction style. Most decoupling attributed to
  "event-driven" architectures derives from the channel type (Publish-Subscribe), not from
  the fact that the messages are called events.
- **Do not** conflate Publish-Subscribe decoupling with event semantics. A Command Message
  placed on a Publish-Subscribe Channel gains the same topology-decoupling benefits as an
  Event Message on the same channel.
- **Do** recognize that Pub/Sub coupling is *asymmetric*: adding subscribers does not
  affect the sender, but changing the message format or semantics propagates to all
  subscribers. This one-directional property — decoupling the right-to-left direction while
  remaining coupled left-to-right — is the real architectural advantage of Pub/Sub.
- **Do not** confuse a Recipient List with Publish-Subscribe. A Recipient List (used by
  AWS EventBridge rules and targets) sends to a computed set of recipients, but adding a
  recipient requires modifying the list — coupling propagates in both directions.
  Publish-Subscribe channels allow adding subscribers without affecting the sender or other
  subscribers.
- **Do** account for the cost of topology decoupling: the easier it is to add recipients,
  the harder it becomes to change the message format. As a Pub/Sub system grows and
  accumulates consumers, each format change requires coordinating more downstream teams.
  Topology decoupling and semantic coupling are inversely related as the system matures.
- **Do** use Pub/Sub most aggressively when you do not control the message source (e.g.,
  subscribing to cloud provider events). In your own application where you control both
  sender and receiver, the ability to add recipients without side effects is a smaller
  advantage — direct communication with clear contracts may be simpler.
- **Do not** assume a system is topology-decoupled because it uses logical identifiers.
  Serverless event-driven applications that route by ARN or topic name may appear
  composable, yet inserting an intermediary or changing the data source forces downstream
  consumers to change because message formats depend on the source — hidden coupling
  across dimensions.
- **Do not** mistake implicit dependencies for absent ones. A Pub/Sub system with many
  subscribers has dependencies between components — they are just undocumented. Scattered
  implicit dependencies are not loose coupling; they are tight coupling with worse
  observability. An integration with dependencies "all over the system" is a poor design
  regardless of whether those dependencies are explicit or hidden behind a topic.
- **Do** be skeptical of EDA decoupling claims relative to the speaker's stage in the
  lifecycle. A system that is still being built gains real benefit from easy recipient
  addition. A system that has accumulated dozens of consumers faces a coordination cost
  on every format change. People who pitch EDA as loosely coupled are often early in the
  lifecycle and will not live with the long-term consequences of those topology decisions.

**Broker and bus topology:**

- **Do** use a **Message Broker** (centralized routing) when you need a single place to
  control message flow, when routing rules change frequently, and when the broker's
  single-point-of-failure risk is acceptable. A broker is easier to operate than
  distributed routing but is a bottleneck and reduces the decoupling it appears to enable.
- **Do not** use a Message Broker as the default topology when direct point-to-point
  channels suffice. Every message through the broker is an extra hop.
- **Do not** overload the broker with transformation logic. A broker should route; keep
  translation in dedicated Message Translators.
- **Do** prefer a **Message Bus** (shared channel infrastructure with a common message
  format) over a mesh of point-to-point channels when many independent applications must
  communicate and when adding or removing participants must not require changes to
  existing ones.
- **Do not** adopt the Message Bus topology unless all participants can agree on the
  canonical message format. A bus with heterogeneous formats creates a translation problem
  at every receiver.
- **Do** use a **Detour** (a Control Bus-controlled bypass switch) to activate optional
  processing steps (validation, audit logging, debugging) at runtime without redeploying.
- **Do** recognize that Publish-Subscribe makes it harder, not easier, to answer "who
  consumed this message and what did they do with it?" Plan for that question at design
  time using Message History, Wire Tap, and Message Store.

> **Red flag — Invisible consumers**: a Pub/Sub topic with an unknown number of
> subscribers. When the message format must change, there is no way to know which
> downstream teams need to be notified.

---

## 16. Model control flow explicitly — queues invert it, Drivers shape it

Rules 1 and 15 establish the coupling-dimension vocabulary. This rule extends that
framework to the dimension EIP left implicit: control flow — which element actively
drives the interaction at each point in the pipeline — determines scalability, robustness,
and latency. In asynchronous systems, data flow and control flow often point in *opposite*
directions, and failing to model this produces systems whose run-time behaviour is opaque.

**Control flow fundamentals:**

- **Do** distinguish data flow (what is exchanged) from control flow (which element
  initiates the exchange). An arrow on an architecture diagram should state which it
  represents — they are not the same thing and regularly point in opposite directions.
  A Polling Consumer, for example, has data flowing left-to-right and control flow
  right-to-left.
- **Do** classify every element by its push/pull role — these are the atomic building
  blocks of the control flow vocabulary:
  - **Sender** — actively pushes data; data flow and control flow both run left-to-right.
  - **Sink** — passively receives pushed data; no control flow initiative.
  - **Source** — passively holds data waiting for a Fetcher; no control flow initiative.
  - **Fetcher** — actively pulls data from a Source; data and control flow point in
    opposite directions.
- **Do** identify the *pipeline combination* role of each composed component:
  - **Pusher** — receives from an active Sender and forwards actively to the next element.
    Control and data flow both run left-to-right. Synchronous in cadence: each message
    received triggers a message sent.
  - **Puller** — fetches actively from a Source and forwards passively when the downstream
    element requests. Control flow runs right-to-left. Synchronous in cadence: a
    downstream fetch triggers an upstream fetch.
  - **Queue** — connects a Sender (active push) to a Fetcher (active pull). Inverts control
    flow. Asynchronous: arrival and departure rates are fully independent.
  - **Driver** — fetches actively from a Source and pushes actively to a Sink. Also inverts
    control flow, but controls the fetch cadence directly without a buffer queue between
    the two active sides.
- **Do** use the affordance rule to determine which connecting element is required.
  Two active elements facing each other — a Sender feeding a Fetcher, or a Driver
  feeding another Driver — cannot connect without an intermediary: **insert a Queue**.
  Two passive elements facing each other — a Source and a Sink — cannot connect without
  an active intermediary: **insert a Driver**. The mismatch of push/pull roles is what
  forces the choice; the wrong connecting element breaks the control flow.
- **Do** apply the affordance rule to cloud service selection. EventBridge Pipes is a
  Driver — it actively fetches from Sources. Because it needs a passive Source, it cannot
  consume from SNS or S3 — both are active Senders, not Sources. Two active elements face
  each other and cannot connect; a Queue must be inserted between them first.
- **Do** recognize that a **Queue** connects a Sender to a Fetcher by inverting control
  flow: the sender pushes at its preferred rate; the receiver pulls at its preferred rate.
  The queue absorbs the mismatch. This is the formal mechanism behind temporal decoupling.
- **Do** use a **Driver** when you need control over message fetch cadence without a
  queue in between. A Driver rate-limits by adjusting its fetch cadence: when the target
  is rate-limited or slow, the Driver simply fetches less frequently — excess load never
  enters the pipeline as buffered messages. A queue cannot do this; it must accept before
  it can regulate, and must then apply TTL, tail drop, or backpressure to what it has
  already buffered. Because a Driver processes messages sequentially it also preserves
  in-order delivery, which competing consumers via a queue cannot.

**Traffic shaping:**

- **Do** use queues to flatten arrival-rate spikes before they reach worker pools. A queue
  absorbs a burst; the workers process at a steady rate regardless of arrival patterns.
  Systems designed around this model degrade gracefully under load: response times rise,
  but throughput remains stable.
- **Do not** rely on synchronous call stacks to absorb load spikes. Under heavy load,
  synchronous systems can reach a threshold beyond which throughput actually *decreases*
  as the system spends more time managing inflight requests than completing them. A
  queued, worker-pool design avoids this collapse.
- **Do** use a **Driver** rather than a queue when you need low latency and control over
  fetch rate simultaneously. EventBridge Pipes uses this model: it fetches from sources at
  a controlled rate without requiring a separate queue, reducing cost and preserving order.

**Flow control — preventing queue overflow:**

- **Do** implement explicit flow control rather than allowing queues to grow without bound.
  "Unlimited" queue depth from cloud providers means no hard limit is enforced, not that
  memory is infinite. Apply Little's Law: *average wait time = queue depth ÷ arrival rate*.
  Queue depth and wait time are proportional — a queue that grows indefinitely eventually
  makes messages irrelevant before they are processed.
- **Do** use **rate limiting** as the first and preferred form of flow control. Rate
  limiting is *proactive*: it constrains the arrival rate before queue pressure builds,
  by configuring delivery rate upfront to match the downstream system's known processing
  capacity. The three remaining mechanisms are *reactive* — they engage only after the
  queue is already under pressure. Design for rate limiting as the first line; treat
  TTL, tail drop, and backpressure as fallbacks for unexpected bursts.
- **Do** choose the appropriate reactive mechanism based on message value over time:
  - **Time-to-Live (TTL)**: drop messages queued longer than their useful life. Best for
    time-sensitive data (sensor readings, stock prices, UI events) where a stale message
    is worse than no message.
  - **Tail drop**: reject new arrivals when the queue is full. Best when old messages are
    more valuable than new ones and senders have a feedback path to retry later.
  - **Backpressure**: signal upstream that the queue cannot accept more messages so senders
    reduce their arrival rate. Best for ordered workflows where every message must be
    processed and senders can throttle.
- **Do not** treat a 500 error under overload as an unplanned event. An overloaded
  synchronous system returning errors to new callers is, functionally, an unintentional
  tail drop — without the feedback path that would allow callers to retry intelligently.

**Push vs. pull latency:**

- **Do not** assume push delivery is always lower-latency than pull. Push services at
  major cloud providers buffer events internally (in queues that are hidden from the API)
  before delivering — P90 latencies on the order of 250ms are common. Pull delivery, where
  the recipient controls polling rate and batch size, can achieve lower median latency
  despite the polling overhead. Choose based on measured latency requirements, not
  intuition.
- **Do** account for the fact that push-delivery event routers optimize for throughput and
  operational stability rather than latency. Services using parallel worker pools for
  delivery do not guarantee message order — understand this constraint before choosing a
  push-delivery architecture for order-sensitive workloads.

> **Red flag — "All lights green, system is down"**: queue depth is growing, all consumers
> are processing at full throughput, error rate is zero — yet users experience latency
> measured in minutes or hours because messages wait in queue long after they are relevant.
> Standard health checks that test liveness and error rate are blind to this condition.
> Monitor **queue age** — the time since the oldest unacknowledged message was enqueued —
> not just queue depth or throughput. Alert on age: a deep queue draining fast is healthy;
> a shallow queue whose oldest message is 10 minutes old is a user-visible outage.
