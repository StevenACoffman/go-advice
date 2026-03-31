# Data Systems Design Rules (Kleppmann — 2nd Edition)

Rules derived from *Designing Data-Intensive Applications* 2nd edition by Martin Kleppmann
and Chris Riccomini. The single organizing principle: **every architectural decision is a
trade-off — choose based on your actual workload, not on what sounds impressive**.

---

## 1. Separate operational and analytical systems

OLTP and OLAP have different schemas, access patterns, and optimization goals. Conflating
them produces a system optimised for neither.

- **Do** use OLTP systems (relational databases, document stores) for transactional
  workloads: point queries by primary key, small frequent reads and writes, low latency.
- **Do** use OLAP systems (data warehouses: BigQuery, Snowflake, Redshift) for analytical
  workloads: complex aggregate queries over millions of rows, infrequent bulk writes.
- **Do** route analytical queries to a separate read replica, warehouse, or OLAP engine —
  never against the same database that serves live traffic.
- **Do not** run `SELECT COUNT(*) ... GROUP BY` or full-table scans against your
  production OLTP database; a single analytical query can saturate I/O for all users.
- **Do** distinguish *systems of record* (authoritative, normalised, written to first) from
  *derived data systems* (caches, indexes, ML models, materialized views). If derived data
  is lost it can be recreated; systems of record cannot be regenerated from downstream data.
- **Do** make the lineage of derived data explicit in code and documentation: which system
  of record feeds which derived dataset, via which pipeline.

---

## 2. Choose the right data model for the access patterns

The data model shapes every query, every join, and every schema change forever.

- **Do** use a relational model (PostgreSQL, MySQL) when data has many-to-many
  relationships, when joins are common, when normalization is needed for consistency, or
  when the schema is stable.
- **Do** use a document model (MongoDB, DynamoDB) when data is a tree of one-to-many
  relationships, when the application loads the whole document at once, and when schema
  flexibility or heterogeneity is important.
- **Do not** use a document model when your application frequently needs to reference data
  across documents (many-to-many relationships) — application-level joins are slower and
  harder to maintain than database joins.
- **Do** use a graph model (Neo4j, KùzuDB, Amazon Neptune) when anything can be related
  to anything: social graphs, fraud detection, knowledge graphs, recommendation engines.
  Graph query languages (Cypher, SPARQL) express multi-hop traversals in a few lines that
  require dozens of lines of recursive SQL.
- **Do** use **DataFrames** (Pandas, Spark DataFrames, DuckDB, Dask) as the bridge
  between relational data and the numerical arrays that ML algorithms require. DataFrames
  generalize relational tables to large numbers of columns and support transformations
  (one-hot encoding, matrix pivots) that SQL expresses awkwardly or not at all.
- **Do** use **GraphQL** when client code (mobile app, SPA) must determine the exact
  shape of the response, and when different UI screens need different subsets of the same
  data. GraphQL works on any backing store (relational, document, or graph).
- **Do not** use GraphQL for internal service-to-service APIs or for recursive/arbitrary
  queries — GraphQL intentionally prohibits recursion and open-ended search conditions to
  prevent DoS. Internal APIs should use gRPC or REST.
- **Do** prefer declarative query languages (SQL, Cypher, SPARQL) over imperative ones.
  Declarative queries let the database choose an execution plan; they parallelize more
  easily and survive optimizer improvements without code changes.
- **Do not** try to emulate graph traversal in a relational database with recursive CTEs
  when a proper graph database is available — the impedance mismatch produces fragile code.
- **Do** favour normalization in OLTP (single copy of each fact, updated once) and
  deliberate denormalization in OLAP (pre-joined, read-optimised). Treat denormalized
  analytical data as derived; maintain it through ETL or CDC pipelines.

---

## 3. Model data as events when state-change history matters

Event sourcing represents data as an append-only log of immutable events. It is a
first-class data model alongside relational, document, and graph — not just an
infrastructure pattern.

- **Do** use event sourcing when auditability, replayability, or complex business domain
  logic is needed: reservations, financial transactions, workflow orchestration, regulated
  industries. The event log is the ground truth; all other representations are derived.
- **Do** name events in the **past tense** (`SeatBooked`, `PaymentRefunded`) because they
  record facts that have already occurred. A consumer of the event log must not reject an
  event — validation happens in the command layer before the event is written.
- **Do** derive all read-optimised views (CQRS — Command Query Responsibility
  Segregation) from the event log deterministically. You must always be able to delete a
  materialized view and rebuild it by replaying the same events through the same code.
- **Do** keep all information needed to reproduce a computation inside the event itself.
  If a computation requires an exchange rate, include the rate in the event — do not fetch
  external data at replay time, or you will get different results on different days.
- **Do** use event sourcing as a natural complement to stream processing: Kafka stores the
  event log; stream processors (Flink, Kafka Streams) maintain the materialized views.
- **Do not** send external notifications (emails, webhooks, push notifications) directly
  from event handlers that run during view rebuilds. Separate side-effect actions from
  view-maintenance code, and guard them with an idempotency key.
- **Do** handle GDPR right-to-deletion in event logs via **crypto-shredding**: encrypt the
  user's personal data fields with a per-user key, then delete the key when the user
  requests deletion. The encrypted bytes remain in the log; they are unreadable without
  the key. If data is per-user only, deleting the user's entire partition is simpler.
- **Do** use EventStoreDB, MartenDB (PostgreSQL-backed), or Axon Framework for
  dedicated event sourcing storage; use Kafka for the event log when you also need
  stream processing consumers.

---

## 4. Match storage engine to workload

No single storage engine is best for all workloads. The two dominant families make
opposite trade-offs.

- **Do** use a **B-tree** engine (most PostgreSQL, MySQL InnoDB) when the workload is
  read-heavy, range queries are common, or predictable read latency matters most.
  B-trees offer O(log n) point lookups and efficient range scans.
- **Do** use an **LSM-tree** engine (RocksDB, Cassandra, LevelDB, ScyllaDB) when the
  workload is write-heavy and sequential write throughput is the bottleneck. LSM writes
  are always sequential (friendly to SSDs) and apply better compression.
- **Do** use **leveled compaction** (default in most LSM engines) for read-heavy or
  mixed workloads; use **size-tiered compaction** for write-maximizing workloads where
  reads can tolerate checking more SSTables.
- **Do** use **column-oriented storage** (Parquet, ORC, DuckDB, ClickHouse) for
  analytical queries. Queries that touch 3 of 100 columns read 97% less data than
  row-oriented storage. Compression ratios are also much higher (adjacent values compress
  well).
- **Do not** add an index speculatively. Every secondary index slows writes and uses disk.
  Choose indexes from actual query patterns; remove unused ones.
- **Do** use bloom filters for LSM-based systems to avoid reading all SSTables on a
  negative lookup.
- **Do** prefer **external sorting** (sort utility, database ORDER BY with spill-to-disk)
  over in-memory sorts for datasets larger than available RAM.

| Workload | Preferred engine | Why |
|---|---|---|
| High write throughput, SSD | LSM-tree (RocksDB, Cassandra) | Sequential writes, lower write amplification |
| Read-heavy, range scans | B-tree (PostgreSQL, InnoDB) | Predictable latency, efficient range scan |
| Analytical, few columns from many | Column-oriented (Parquet, ClickHouse) | Read only needed columns, vectorized processing |
| Full-text search | Inverted index (Elasticsearch, Lucene) | Term → posting list with TF-IDF scoring |
| Semantic / vector search | ANN index (pgvector, HNSW, IVF) | Nearest-neighbour in embedding space |
| ML training data, wide matrices | DataFrame store (DuckDB, ArcticDB) | Columnar + large numbers of columns |

---

## 5. Design for schema evolution from day one

Systems must change. Data written today will be read by code that did not exist when it
was written.

- **Do** ensure every change to a serialized schema is **backward compatible** (new code
  can read old data) and **forward compatible** (old code can read new data).
- **Do not** use language-specific serialization (Java `Serializable`, Python `pickle`)
  for anything stored to disk or sent over the network. It locks the format to a single
  language, enables arbitrary code execution on deserialization, and provides no version
  tolerance.
- **Do** use **Protocol Buffers or Avro** for binary operational data. Both enforce
  schema, produce compact wire formats, and have explicit evolution rules (field tags in
  Protobuf; name-matching with writer/reader schema resolution in Avro).
- **Do** use **JSON** for data exchanged between independent organizations or for
  human-readable APIs. Avoid JSON for high-throughput internal data pipelines.
- **Do** follow Protobuf evolution rules: never reuse a field tag number; only add new
  fields with new tag numbers; mark removed fields reserved so the tag is never recycled.
- **Do** follow Avro evolution rules: only add or remove fields that have default values.
  Fields without defaults cannot be added (new reader can't fill in missing values) or
  removed (old reader can't handle absent field).
- **Do** apply schema changes to databases using **online migration tools** (gh-ost,
  pt-online-schema-change) when the table is large. Never run `ALTER TABLE` that rewrites
  a multi-gigabyte table on a live OLTP system.
- **Do not** rely on schema-on-read as an excuse to skip schema design. Heterogeneous
  or dynamically structured data is a valid use case for document stores, but every
  reader must then handle all historical data shapes.
- **Do** design schemas with the principle that **data outlives code**. Data written
  today will likely still exist when the code that wrote it is long gone. Write schemas
  that can be read by code that didn't exist when the data was written, and that can read
  data written by code that no longer exists. This is the operational motivation for
  backward and forward compatibility — not just versioning hygiene.
- **Do** choose serialization format based on access pattern:
  - **Columnar** (Parquet, ORC, Arrow): analytical workloads, batch pipelines, reading
    a few columns from many — reads only the columns needed, best compression.
  - **Row-based** (Avro, Protobuf): event streaming, write-heavy pipelines, replication
    logs — entire records read at once, random access by key, lower write overhead.
  - Do not use Parquet in a Kafka topic or Avro as a data warehouse storage format;
    the formats are optimised for opposite access patterns.
- **Do** version APIs explicitly at the HTTP/RPC interface level, separately from the
  schema evolution handled by Protobuf or Avro:
  - **REST**: encode the version in the URL path (`/api/v2/users`) or the `Accept`
    header (`Accept: application/vnd.myapi.v2+json`). URL versioning is more visible
    but harder to route; header versioning is cleaner but invisible in logs and links.
  - **gRPC+Protobuf**: field-tag–based evolution handles most changes without a
    version bump. Breaking changes (field type changes, renamed services) require a new
    package name and parallel deployment.
  - Do not remove fields from a gRPC service in a single deploy — clients may still
    be sending them. Deprecate, drain, then remove in separate deploys.

---

## 6. Design replication topology deliberately

Replication provides durability, fault tolerance, and read scale — but each topology
has specific failure modes that must be handled explicitly in application code.

- **Do** use **single-leader replication** as the default for OLTP systems. It prevents
  write conflicts and provides a deterministic ordering of changes. Route all writes to
  the leader; distribute reads across followers.
- **Do not** use asynchronous replication for data you cannot afford to lose. If the
  leader fails after acknowledging a write but before it replicates, that write is lost.
  Use semi-synchronous replication (at least one follower acknowledged) for critical data.
- **Do** implement **read-after-write consistency** explicitly: after a user submits data,
  read their own writes from the leader (or a replica guaranteed to be caught up) for a
  short window. Don't assume a random follower has the user's just-written data.
- **Do** implement **monotonic reads**: always route the same user's reads to the same
  replica (hash by user ID), or require replicas to not return data older than a timestamp
  seen by that user. Without this, users can see data move backwards in time.
- **Do** implement **consistent prefix reads** for causally related writes: ensure an
  answer is never visible before its question. Use causal tokens or route causally
  related reads/writes to the same shard.
- **Do** use **logical (row-based) replication** (MySQL binlog, PostgreSQL logical
  decoding) rather than WAL-shipping. Logical logs decouple the format from the storage
  engine internals, allow independent version upgrades, and enable change data capture
  (CDC) for downstream consumers.
- **Do not** treat failover as automatic and safe. Always address these risks:
  - Asynchronous replication → unreplicated writes silently lost on leader failure.
  - External systems (Redis, Kafka) using primary keys from the old leader may produce
    conflicts with the new leader's sequence.
  - Split-brain (two nodes believe they are leader) requires a fencing mechanism that
    causes the old leader to stop accepting writes.
- **Do** use **CDC (change data capture)** via Debezium or similar to propagate changes
  from the system of record to caches, search indexes, and data warehouses, instead of
  dual-writing from application code. Dual-writes are not atomic and produce inconsistency
  on partial failure.
- **Do not** expose your internal database schema directly as the CDC event schema.
  CDC effectively turns the database schema into a public API: removing or renaming a
  column breaks every downstream consumer, potentially causing production outages.
  Use the **outbox pattern** instead: write changes to a dedicated outbox table (with
  its own stable schema) within the same transaction as the domain change; point CDC
  at the outbox table, not the internal tables. Internal schema changes no longer break
  external consumers.
- **Do not** enable **unclean leader election** (the Kafka `unclean.leader.election.enable`
  option, or the equivalent in other systems) unless you have explicitly decided to trade
  durability for availability. An unclean election promotes a replica that may be missing
  committed writes; those writes are permanently lost and the former leader's log will
  conflict if it rejoins. If you enable it, consumers must tolerate duplicate or missing
  messages across the leadership transition.
- **Do not** use **Last Write Wins (LWW)** with wall-clock timestamps for conflict
  resolution on mutable records. It has three distinct failure modes:
  1. **Silent data loss from clock skew**: a node with a lagging clock cannot overwrite
     values previously written by a node with a faster clock until the skew resolves —
     writes from the lagging node are silently dropped with no error to the application.
  2. **Inability to distinguish causally ordered from concurrent writes**: if client B
     increments a value written by client A but B's timestamp is earlier due to skew,
     LWW discards B's increment — the causally later write loses.
  3. **Timestamp collisions**: at millisecond clock resolution, two independent writers
     can produce the same timestamp; any tiebreaker (random number) may itself violate
     causality.
  LWW is only safe for insert-only or naturally idempotent data. For mutable shared data,
  use a merge strategy or route all writes for a record to a single leader (conflict avoidance).
- **Do** use **CRDTs (Conflict-free Replicated Data Types)** or **Operational
  Transformation** libraries (Automerge, Yjs) for collaborative real-time editing where
  concurrent writes must be automatically merged without losing any user's intent.
  CRDTs provide proven merge semantics for counters, sets, ordered lists, and text.
- **Do** consider a **sync engine** (PouchDB/CouchDB, Automerge, Yjs, Google Firestore)
  when building offline-capable collaborative apps where users must work without a network
  connection and reconcile changes asynchronously. A sync engine is multi-leader
  replication taken to the extreme of per-device leaders.
- **Do not** use a sync engine when the user's data set is too large to download
  entirely to the client (e.g., the full catalog of an e-commerce site). Sync engines
  require the client to hold the full working set locally.

---

## 7. Partition data to avoid hot spots

Partitioning (sharding) is required when a single node cannot hold the dataset or handle
the write throughput.

- **Do** use **range partitioning** when range queries are common (time-series, sorted
  data). Place adjacent keys on the same partition to enable efficient scans.
- **Do not** use a monotonically increasing key (timestamp, auto-increment ID) as the
  sole partition key with range partitioning — all writes concentrate on the last
  partition. Add a prefix or use hash partitioning.
- **Do** use **hash partitioning** when distributing writes evenly matters more than
  range queries. Hash the partition key to assign records to partitions uniformly.
- **Do** route requests to the correct node via a **routing tier** (or gossip-based
  node discovery) rather than broadcasting queries to all nodes.
- **Do** place data that is frequently joined together on the **same partition** (co-location).
  Cross-partition joins require scatter-gather over all nodes and are much more expensive.
- **Do not** conflate partitioning with replication: each partition normally has multiple
  replicas. Partitioning splits data; replication copies it.
- **Do** plan for rebalancing. Use a fixed number of partitions much larger than the
  number of nodes (Elasticsearch default: 1,000 shards across N nodes) so that adding a
  node only requires moving some existing partitions, not repartitioning all data.

---

## 8. Use the correct isolation level for the concurrency hazard

"ACID" means different things in different databases. Choose the isolation level that
prevents the specific anomaly you care about; stronger guarantees cost throughput.

- **Do** default to at least **snapshot isolation** (Oracle, PostgreSQL `REPEATABLE READ`,
  MySQL `REPEATABLE READ`) for general application code. It prevents dirty reads, dirty
  writes, and non-repeatable reads. It is safe for analytics queries running alongside
  OLTP writes.
- **Do** use **serializable isolation** for any operation where correctness depends on a
  "check then act" sequence across multiple records: booking systems (no double-booking),
  financial transfers (no overdraft), username registration (unique constraint), on-call
  scheduling. Snapshot isolation does NOT prevent write skew on multi-row checks.
- **Do** prefer **Serializable Snapshot Isolation (SSI)** (PostgreSQL `SERIALIZABLE`,
  CockroachDB, FoundationDB) over two-phase locking (2PL). SSI provides full
  serializability with a small performance overhead; 2PL blocks readers and causes
  cascading delays under high contention.
- **Do** use **atomic operations** (database `UPDATE counter = counter + 1`, Redis `INCR`)
  instead of read-modify-write cycles in application code. ORMs make it easy to
  accidentally produce unsafe read-modify-write cycles.
- **Do not** use `SELECT FOR UPDATE` as a substitute for thinking about whether
  serializability is needed. Lock the minimum number of rows for the minimum duration.
- **Do not** confuse **serializability** with **linearizability** — they are orthogonal
  guarantees:
  - **Serializability** (transaction isolation): multi-object transactions appear to
    have executed in some serial order. Prevents write skew and phantoms. SSI in
    PostgreSQL/CockroachDB is serializable.
  - **Linearizability** (consistency / recency): once a write completes, every
    subsequent read on any node sees that write. Single-object guarantee. etcd and
    Spanner are linearizable.
  - A database can be one, both, or neither. CockroachDB provides serializability (SSI)
    but not full linearizability on reads by default. Spanner provides strict
    serializability (both). Cassandra provides neither by default.
- **Do** keep **SSI transactions short**. Under Serializable Snapshot Isolation, a
  transaction that reads and writes over a long time window accumulates a large conflict
  footprint. Other transactions that touch the same records in the interim will force
  one side to abort. Long-running read-write transactions under SSI have a much higher
  abort rate than short ones. Structure workflows to commit quickly; do expensive
  computation outside the transaction and use it as input to a short transactional step.
- **Do** retry transactions on transient failures (deadlock, serialization failure,
  timeout) with **exponential backoff**. Distinguish transient errors (retry) from
  permanent errors (constraint violation — don't retry).
- **Do** generate a unique **idempotency key** per user-initiated operation and persist
  it in the transaction alongside the operation result. On retry, look up the key first
  and return the cached result rather than executing again.
- **Do not** rely on database transactions alone to guarantee exactly-once end-to-end
  semantics. A `COMMIT` succeeds but the HTTP response is lost; the client retries and
  the operation runs twice. Application-level deduplication is required.

| Hazard | Prevented by | Isolation level |
|---|---|---|
| Dirty reads | Read committed | Read committed (default in most DBs) |
| Dirty writes | Read committed | Read committed |
| Read skew (non-repeatable read) | Snapshot isolation | Repeatable read / Snapshot |
| Lost updates (concurrent increment) | Atomic ops / explicit lock | Repeatable read + atomic op |
| Write skew (multi-row "check then act") | Serializability | Serializable only |
| Phantom reads | Serializability | Serializable only |

---

## 9. Assume the network and clocks are unreliable

Distributed systems fail differently from single-node systems. Partial failures are
normal, not exceptional.

- **Do** design for unbounded network delays. A request can be lost, delayed by minutes,
  or delivered multiple times. A response may never arrive even if the request succeeded.
  You cannot distinguish a crashed node from a slow one with a timeout alone.
- **Do** set timeouts experimentally based on measured response time distributions,
  not by guessing. Use adaptive failure detectors (like Phi Accrual) for systems
  that can measure response time variance.
- **Do not** assume a short timeout is safe. Declaring a node dead when it is merely slow
  and transferring its load to other nodes can cause a **cascading failure** — now all
  nodes are overloaded.
- **Do** design for **metastable failures**: under severe overload, synchronized retries
  from many clients can recreate the spike that caused the overload, keeping the system
  stuck even after the original load is reduced. A system in this state requires manual
  intervention (restart or shedding load externally) to recover. Prevention: exponential
  backoff with jitter on all retries, circuit breakers that stop retrying a failing service,
  and load shedding before the queue depth makes recovery impossible.
- **Do** implement **circuit breakers**: when a downstream service has failed or timed out
  repeatedly, stop sending it requests for a cooldown period rather than continuing to
  hammer it. This gives the downstream service time to recover and prevents the caller's
  threads from piling up on a blocked dependency.
- **Do** implement **backpressure**: when a consumer cannot keep up with a producer's
  rate, the consumer signals the producer to slow down rather than buffering unboundedly.
  Unbounded queues mask overload; backpressure surfaces it and prevents OOM crashes.
- **Do** implement **load shedding**: when a server approaches its capacity limit, proactively reject low-priority requests with a `503` rather than queuing everything until
  latency explodes for all callers.
- **Do** use **exponential backoff with jitter** on all retries. Synchronized retries from
  many clients (all retrying at the same interval) recreate the spike that caused the
  overload. Jitter spreads the retry load.
- **Do** be aware of **head-of-line blocking**: a few slow requests ahead in a queue hold
  up fast requests behind them. Measure response times on the client side (which includes
  queue wait time), not just on the server side (which does not).
- **Do** account for **tail latency amplification** when a request fans out to multiple
  backends in parallel. If each backend has a p99 latency of 200ms independently, a
  request that calls 100 backends in parallel will hit 200ms on ~63% of requests — even
  though any single backend only shows that latency 1% of the time. Fan-out multiplies
  tail risk. Mitigations: issue a small number of parallel "hedge" requests and take the
  first response; bound fan-out by sharding rather than broadcasting; add per-request
  timeouts on individual backend calls so one slow node doesn't hold everything.
- **Do** be aware of **cross-channel timing races**: if you write data to storage and
  then publish a notification (message queue, email, push notification) via a different
  channel, the consumer of the notification may read from a lagging replica and not see
  the data the notification is about. Fix options: (a) use linearizable storage so any
  reader sees the latest write, (b) include the essential data directly in the
  notification payload, or (c) include a `min_version` / `etag` that the reader must
  wait for before considering its read complete.
- **Do not** use wall-clock time (`time.Now()`, `System.currentTimeMillis()`) to determine
  the order of events across machines. Clock skew means a message timestamped at
  `t=100ms` on machine A may arrive at machine B, which shows `t=99ms`. Use logical
  clocks (Lamport timestamps, vector clocks, Hybrid Logical Clocks) for event ordering.
- **Do** guard against **process pauses**: GC stop-the-world events, VM preemption,
  and OS scheduling can pause a process for seconds. Any lease or lock with a short
  timeout can expire during such a pause. Use fencing tokens (monotonically increasing
  integers from a consensus service) to prevent a "zombie" process from acting after
  its lease expires.
- **Do** treat network faults as a first-class engineering concern, not an edge case.
  Run regular fault injection (Chaos Engineering). One study found ~12 network faults
  per month in a medium datacenter.
- **Do not** assume TCP guarantees end-to-end correctness. TCP prevents corruption within
  a single connection; it does not prevent a server from processing a request and crashing
  before acknowledging it, or a client retrying and sending the request twice.
- **Do** use **deterministic simulation testing (DST)** as a complement to fault injection.
  DST runs your actual code (not a model) through a large number of randomized executions
  in which network delays, I/O, and clock timing are all controlled by the simulator. This
  lets the simulator explore far more scenarios than hand-written tests, including scenarios
  that rarely occur in practice. Crucially, when a failure is found, the exact sequence of
  operations can be replayed — unlike chaos engineering, which has no record of what
  triggered the failure. FoundationDB, TigerBeetle, and Antithesis use DST to find bugs
  that fault injection cannot reproducibly exercise.

---

## 10. Use consensus for anything that requires a single authoritative decision

When multiple nodes must agree on a value and that agreement must survive failures,
only a consensus algorithm provides safety guarantees.

- **Do** use a consensus-based coordination service (ZooKeeper, etcd) for leader election,
  distributed locks, and service discovery. Never implement leader election with timeouts
  and database flags — two nodes can both believe they are leader (split-brain).
- **Do** use **linearizable storage** (Spanner, CockroachDB, FoundationDB) when
  correctness requires that reads always see the most recent write, regardless of which
  node serves the read. Multi-leader and leaderless (Dynamo-style) replication are NOT
  linearizable.
- **Do not** assume Dynamo-style quorum reads (`r + w > n`) are linearizable. Race
  conditions are still possible. Cassandra LWW conflict resolution is nonlinearizable
  because it uses time-of-day clocks that can skew.
- **Do** use **Serializable Snapshot Isolation** (SSI) in a distributed database
  (CockroachDB, FoundationDB) rather than implementing your own distributed coordination.
  Building correct distributed transactions from scratch is extremely difficult.
- **Do** understand what your database actually guarantees. "ACID compliant" is a
  marketing term; verify the specific isolation level and consistency guarantee.
- **Do** route all writes that must enforce a uniqueness constraint to a **single
  partition or shard** (keyed by the unique value, e.g., username). A stream processor
  consuming that shard sequentially on a single thread can enforce the constraint
  deterministically and without cross-shard coordination: for each request, check a local
  database of claimed values; emit a success or rejection message to an output stream.
  This scales by increasing the number of shards — each shard is processed independently.
- **Do** understand the costs of consensus before reaching for it:
  - Requires a **strict majority** to operate: 3 nodes to tolerate 1 failure, 5 nodes
    to tolerate 2. You cannot run consensus with 2 nodes.
  - Every write touches a quorum — **adding more nodes makes writes slower**, not
    faster. Consensus does not scale horizontally on the write path.
  - A **network partition blocks the minority** partition from making progress. A
    consensus cluster split into two unequal halves will stop accepting writes on the
    smaller side rather than risk split-brain.
  - Cross-region consensus incurs multi-hundred-millisecond round trips. For
    globally distributed, high-write-throughput workloads, strict consensus may be
    unsuitable — consider eventual consistency with end-to-end integrity checks (§13,
    §14) instead.
- **Do not** implement consensus yourself using timeouts and database flags. Use a
  proven implementation: etcd (Raft), ZooKeeper (Zab), or a database that provides
  linearizable operations natively.

---

## 11. Exploit immutability in batch pipelines

Batch processing derives its fault-tolerance and correctness properties from treating
inputs as read-only and generating outputs from scratch.

- **Do** treat batch job inputs as immutable. Never modify the source data; always write
  output to a new location.
- **Do** design batch jobs to be **idempotent and re-runnable**. If the job produces
  wrong output, delete the output and rerun from the same input — do not attempt
  incremental repair.
- **Do** leverage **human fault tolerance**: immutable inputs mean you can roll back
  to a previous version of the job code and rerun it to fix buggy output. With a
  mutable OLTP database, rolling back the application code does nothing to fix bad
  data already written. The inability to recover from bad deploys is an architectural
  property of mutable state, not just an operational inconvenience.
- **Do** use open table formats (Delta Lake, Apache Iceberg, Apache Hudi) on object
  stores when you need **time travel** — the ability to query a snapshot of the data
  as it existed at a past point in time. This enables rollback to a pre-bug snapshot,
  audit queries, and reproducible ML training sets, all without extra infrastructure.
- **Do not** write to external databases or send emails/notifications from within a batch
  job's processing loop. Side effects that complete before the job finishes leave external
  systems inconsistent when the job fails and is rerun. Materialize all output to a file
  or staging table first; commit it atomically at the end.
- **Do** prefer modern batch frameworks (Apache Spark, Apache Flink, cloud data
  warehouses) over MapReduce. MapReduce is largely obsolete; Spark and Flink are faster
  and require far less boilerplate.
- **Do** use object stores (S3, GCS) as the durable storage layer for batch pipelines in
  cloud environments. Objects are immutable by design; pipelines become naturally
  checkpoint-able.
- **Do** choose an orchestration system (Airflow, Dagster, Prefect) to manage job
  dependencies, retries, and SLA monitoring for any multi-step pipeline. Do not chain
  cron jobs together with shell scripts.

---

## 12. Apply stream processing for low-latency derived data

Stream processing extends batch processing principles to unbounded datasets and provides
sub-second latency for maintaining derived views.

- **Do** think of all indexes, caches, and materialized views in terms of the **write
  path vs. read path trade-off**: every index is a decision to do more work at write
  time (update the index) to save work at read time (avoid a full scan). Moving work to
  the write path reduces read latency but increases write cost and storage. Tune this
  boundary deliberately, not accidentally.
- **Do** use **event time** (the timestamp embedded in the event at the source) for all
  windowed aggregations, not processing time (when the event arrived at the processor).
  Processing-time metrics show false spikes after restarts when a backlog is drained.
- **Do** account for **late-arriving events**. You can never be certain a window is
  complete. Either ignore stragglers (and track dropped event counts) or publish
  corrections as a follow-up event with the same key. Do not assume events arrive in
  order.
- **Do** choose the message broker style that matches your workload:
  - **Log-based** (Kafka, Kinesis, Pub/Sub): durable, ordered within a partition,
    replayable, fan-out to multiple consumer groups. Best when message ordering matters,
    processing is fast per message, and consumers need the ability to replay. Also the
    right choice for event sourcing, CDC, and stream-table joins.
  - **AMQP/JMS-style** (RabbitMQ, ActiveMQ, SQS): task-queue semantics, message
    deleted after acknowledgment, parallelism at the individual-message level. Best when
    messages are expensive to process (each message dispatched to any available worker),
    ordering is unimportant, and you do not need to replay past messages.
  - Do not treat log-based brokers as drop-in replacements for task queues without
    redesigning for partition-level parallelism.
- **Do** configure a **dead letter queue (DLQ)** on every production consumer. A
  malformed or unprocessable message causes the consumer to crash, be redelivered,
  crash again, and loop indefinitely — blocking all subsequent messages on an ordered
  partition. After a configurable number of retries (typically 3–5), route the message
  to a DLQ and continue processing. Monitor DLQs: any message in a DLQ is an
  operational incident requiring investigation. Do not silently drop messages.
- **Do not** query a remote database inside a stream processor's hot path. This is slow
  and risks overloading the database. Instead, subscribe to a CDC stream of that table
  and maintain a local read-only copy inside the stream processor.
- **Do** use **stream-table joins** via local state stores (Kafka Streams, Flink
  `KeyedBroadcastProcessFunction`) to enrich events with reference data without remote
  calls.
- **Do** choose the right windowing type:
  - **Tumbling**: non-overlapping fixed intervals (billing periods, hourly summaries).
  - **Hopping**: overlapping intervals for smoothed trends (5-min window every 1 min).
  - **Session**: activity bursts separated by inactivity gaps (user session analytics).
  - **Sliding**: all events within a fixed interval of each other (SLA violation detection).
- **Do** take periodic checkpoints of all stateful operators and commit consumer offsets
  only after the checkpoint succeeds. On failure, restore from the last checkpoint and
  replay messages from the matching broker offset.
- **Do** make every stream processor operation **idempotent**. At-least-once delivery
  (the default in Kafka) means messages can be redelivered on consumer restart. Use
  request IDs or event IDs to deduplicate in the state store.
- **Do** handle **slowly changing dimensions (SCDs)** explicitly in stream joins.
  When an event stream is joined to a dimension that changes over time (exchange rate,
  tax rate, product price, user profile), the join is nondeterministic unless you
  version the dimension: give each version a stable ID, include that ID in the event
  at write time, and join on the version ID rather than a "current" snapshot. This
  makes the join deterministic and reproducible when you reprocess historical data.
  Alternatively, denormalize: copy the dimension value directly into the event at the
  time it is produced, so no future join is needed.
- **Do not** use stream processing for queries with strict linearizability requirements.
  Stream processors provide eventual consistency; for strict consistency, synchronous
  coordination is required.

---

## 13. Build end-to-end correctness; don't rely on any single layer

No single layer of infrastructure (TCP, database transactions, at-least-once delivery)
provides complete end-to-end correctness. Each layer protects only within its scope.

- **Do** generate a unique **request ID** for every user-initiated operation in the
  client and propagate it through every service call, queue message, and database write.
  Use this ID to deduplicate retries at every layer.
- **Do** store the request ID alongside the operation result in the database within the
  same transaction. On retry, look up the ID and return the stored result rather than
  re-executing. A `UNIQUE` constraint on the `request_id` column is sufficient; a
  duplicate insert will fail and the transaction will abort — preventing double execution.
- **Do** implement **end-to-end checksums** for data pipelines that pass through multiple
  systems. A checksum at the source and a verification at the destination catches
  corruption introduced by any intermediate layer.
- **Do not** assume a completed database transaction means the operation succeeded
  end-to-end. The HTTP response carrying the commit acknowledgment may be lost; the
  client will retry; the operation will execute a second time unless application-level
  idempotence is implemented.
- **Do** separate the **notification of success** from the **execution of the operation**.
  Inform the user synchronously ("your payment is processing"); complete side effects
  (email, webhook, ledger update) asynchronously after all consistency checks pass.
- **Do** prefer **event log + idempotent consumers** over distributed two-phase commit
  (2PC) for coordinating writes across heterogeneous systems (database + search index +
  cache). 2PC amplifies failures (any participant failure aborts the transaction); a
  durable log decouples producers and consumers and tolerates slow or failed consumers.
- **Do not** implement business logic in stored procedures. They are hard to test, version,
  deploy, and debug. Keep stateless logic in application servers; keep only state management
  in the database.
- **Do** use **sharded logs + stream processors** to enforce constraints across multiple
  shards without distributed transactions (2PC). Example — a funds transfer that must
  check a source account balance and atomically debit and credit two accounts on
  different shards: (1) write the transfer request to a log shard; (2) a stream processor
  reads the shard, checks the constraint against local state, and emits events to per-account
  output shards; (3) per-account processors apply the debit/credit idempotently.
  Each step is a local operation; the log provides ordering; idempotency handles retries.
  This gives equivalent correctness to a cross-shard transaction at much higher throughput.

---

## 14. Distinguish timeliness from integrity

The word "consistency" conflates two distinct properties. Confusing them leads to
either over-engineering (paying for timeliness you don't need) or under-engineering
(missing integrity guarantees that matter).

- **Do** understand the difference:
  - **Timeliness**: users observe up-to-date state. Violations are temporary — waiting
    and retrying resolves them. The CAP theorem's "C" and linearizability are timeliness
    properties.
  - **Integrity**: no data is lost, doubled, or corrupted. Violations are permanent —
    they require explicit detection and repair. Atomicity and durability (the A and D
    in ACID) are integrity properties.
- **Do** prioritise integrity over timeliness. A credit card statement that takes 24 hours
  to reflect a transaction is annoying (timeliness violation). A credit card statement
  where the sum of transactions does not equal the balance is catastrophic (integrity
  violation). Design your system so integrity cannot be violated even if timeliness is
  temporarily lost.
- **Do** recognize that event-driven / stream-based systems naturally sacrifice
  timeliness (reads may be stale) while preserving integrity (exactly-once delivery,
  idempotent consumers). This is usually the right trade-off.
- **Do** consider **compensating transactions** ("apology workflows") as a valid
  engineering pattern for loosely interpreted constraints: if the cost of occasionally
  violating a constraint is low and recoverable (refund, upgrade, apology email), you
  can accept the write optimistically and check the constraint after the fact. This
  eliminates synchronous coordination overhead. Airlines, hotels, and banks all operate
  this way deliberately.
- **Do not** use strict synchronous coordination (2PC, distributed locking) for
  constraints that the business already handles with apologies anyway. The coordination
  overhead is real; the benefit is often illusory.
- **Do** still enforce integrity: exactly-once message delivery, idempotent operations,
  end-to-end request IDs, and audit logs are integrity mechanisms — use them even when
  you relax timeliness.

---

## 15. Use durable execution for multi-service exactly-once workflows

When a business process spans multiple services and each step must execute exactly once
despite failures, durable execution frameworks eliminate the need to hand-roll retry
and idempotency logic across the entire call graph.

- **Do** use a durable execution framework (Temporal, Restate) for workflows that involve
  multiple external service calls (payment gateways, email providers, third-party APIs)
  where partial execution — credit card charged but bank account not credited — is
  unacceptable.
- **Do** understand how durable execution works: the framework logs every RPC call and
  its result to durable storage. On failure and re-execution, it replays the log, skipping
  already-completed steps and returning their cached results.
- **Do** write workflow code to be **deterministic**: same inputs must produce the same
  sequence of RPC calls in the same order. Do not use `time.Now()`, `math.random()`, or
  any non-deterministic system call inside workflow code — use the framework's own
  deterministic wrappers.
- **Do not** reorder or add/remove function calls in an existing workflow definition that
  may have in-flight executions. The framework replays old executions using the current
  code; reordering calls will cause undefined behavior. Deploy new workflow logic as a
  new workflow version; let old executions drain on the old version.
- **Do** still require external services to expose **idempotent APIs** with unique request
  IDs. Durable execution re-executes tasks on failure; if the external service is not
  idempotent, a duplicate call will still cause a duplicate action.
- **Do** use orchestration-style workflow engines (Temporal, Airflow, Dagster, Prefect)
  for ETL pipelines and data workflows; use durable-execution-style frameworks (Temporal,
  Restate) for transactional business processes that involve external service calls.

---

## 16. Build for observability from day one

You cannot debug a distributed system by reading code alone. Observability — the ability
to understand a system's internal state from its external outputs — must be designed in,
not added as an afterthought.

- **Do** instrument every service with **distributed tracing** (OpenTelemetry) from the
  start. A trace ID generated at the request origin must be propagated through every
  downstream HTTP call, gRPC call, Kafka message header, and database query. Use Jaeger
  or Zipkin (or a cloud tracing service) to visualise traces.
- **Do** use **structured logging** (JSON, not free-form text). Every log line must carry
  the trace ID, span ID, service name, and any other fields needed to correlate it with
  related events. Free-form logs are difficult to query at scale.
- **Do** expose **metrics** (Prometheus, OpenTelemetry Metrics) for every service:
  request rate, error rate, and latency percentiles (p50, p95, p99) as a minimum. Add
  business-level metrics (orders per minute, payment failures) alongside infrastructure
  metrics.
- **Do** propagate the **request ID** (from §13) in trace context so that a single
  user-visible operation can be traced across all the services, queues, and databases
  that handled it.
- **Do not** rely on logs alone for distributed system debugging. Logs from independent
  services are difficult to correlate without a shared trace ID. Distributed tracing
  provides the causal chain that logs cannot reconstruct.
- **Do** alert on **symptoms** (user-visible latency and error rate from SLO budgets),
  not just causes (CPU usage, disk space). A service can be unhealthy without any
  infrastructure metric firing.
- **Do** distinguish **SLO** from **SLA**:
  - **SLO (Service Level Objective)**: an internal engineering target (e.g., p99 latency
    < 200ms for 99.9% of requests). Violations trigger alerts and engineering action.
  - **SLA (Service Level Agreement)**: a customer-facing contract with financial
    penalties for breach (e.g., uptime < 99.5% triggers a credit). Violation is
    customer-visible and has business consequences.
  - Set SLOs tighter than SLAs (e.g., SLO=99.9% uptime, SLA=99.5%) to give the
    engineering team time to detect and fix problems before the SLA is breached.
  - "Error budget" = the allowable fraction of time/requests that can fail within the
    SLO period. Burn through the error budget → freeze feature work to focus on
    reliability.
- **Do not** average percentile metrics across instances or time windows. The average
  of p99 values from 10 machines is not the p99 of the combined system — it is
  mathematically meaningless. To aggregate percentiles correctly, collect histogram
  buckets from each source and merge the histograms before computing percentiles.
  Prometheus histograms and OpenTelemetry histograms support this correctly;
  gauge-based percentile metrics (like StatsD timing) do not aggregate safely.

---

## 17. Build sync engines for offline-capable collaborative applications

A sync engine is the client-side counterpart to multi-leader replication: it gives users
a local copy of their data, lets them work offline, and reconciles changes when
connectivity is restored.

- **Do** use a sync engine when users must be able to continue editing while offline
  and changes must be merged when they reconnect. Examples: note-taking apps, document
  editors, calendar apps, project management tools.
- **Do** use an **open-protocol sync engine** (Automerge, Yjs, PouchDB/CouchDB) when
  users must own their data and the app must work if the developer shuts down their
  servers (local-first software). Use a proprietary backend (Google Firestore, Realm)
  when vendor lock-in is acceptable and you need a managed service.
- **Do** think of offline as "very high network latency" rather than a special mode.
  An app using a sync engine does not need a separate offline code path: all reads and
  writes go against local state; the sync engine handles propagation in the background.
- **Do not** use a sync engine when the user's total dataset is too large to store locally
  (e.g., an entire product catalog or content library). Sync engines require downloading
  all relevant data to the device upfront.
- **Do** use CRDT-based sync engines (Automerge, Yjs) for collaborative documents and
  structured data where concurrent edits must merge without losing any user's work.
  For simple append-only or per-user data, simpler last-writer-wins or operational
  transformation approaches may suffice.
- **Do** test sync engine conflict resolution explicitly with scenarios involving two
  clients editing the same field simultaneously, then reconnecting in different orders.
  The correctness of the merge depends on the CRDT implementation, not on your application
  code.

---

## 18. Comply with data regulations and minimise data liability

Legal obligations are architectural constraints, not afterthoughts.

- **Do** collect only the data your application actually requires right now (GDPR data
  minimization principle). Speculative data collection creates liability without benefit.
- **Do** implement **right-to-deletion**: when a user requests deletion, purge their data
  from the system of record and from all derived datasets (caches, search indexes, data
  warehouses, backups). Design pipelines to support this from the start; retrofitting
  deletion into an immutable log is extremely difficult. Use crypto-shredding for
  event-sourced systems (see §3).
- **Do** assess the full cost of a data decision, including reputational and legal risk,
  not just storage cost.
- **Do** verify that your use of cloud services complies with data residency requirements
  (GDPR, CCPA, sector-specific regulations) before storing data in a region.
- **Do not** log sensitive personal data (passwords, payment card numbers, health data)
  in application logs or analytics pipelines unless legally required and properly
  encrypted/masked.
- **Do** audit ML and predictive analytics systems for bias before deployment. If your
  training data reflects historical discrimination (biased hiring decisions, racially
  unequal lending), your model will learn and amplify that discrimination. Correlates of
  protected characteristics (zip code, IP address) can be as discriminatory as the
  protected characteristics themselves.
- **Do** design predictive systems with **explainability** and **appeal mechanisms**.
  If an algorithm makes a decision that affects a person's access to a loan, job, or
  service, there must be a human-reviewable explanation and a correction path when the
  model is wrong.
- **Do** monitor for **feedback loops** in ML systems: a model that predicts risk can
  cause exclusion, which worsens outcomes, which confirms the prediction. Explicitly
  model and test whether the system reinforces or counteracts existing inequalities.

---

## 19. Choose cloud-native architecture deliberately

Cloud services trade control and predictability for elasticity and reduced operational
burden. Know what you are gaining and losing.

- **Do** separate storage from compute in cloud architectures. Object stores (S3, GCS)
  scale independently of compute nodes; analytical queries can spin up ephemeral compute
  clusters against shared storage without needing persistent disks.
- **Do not** distribute prematurely. A single-machine PostgreSQL or SQLite handles
  millions of users for most applications. Distribution adds operational complexity,
  consistency hazards, and network costs. Start with a monolith; distribute when
  measurement proves you must.
- **Do** decompose a monolith into **microservices** primarily when **organizational
  scalability** demands it — different teams need to deploy independently without
  coordinating. Do not decompose for technical scalability alone; a well-indexed database
  with read replicas scales further than most applications will ever need.
- **Do** choose the right synchronous API style when you must use RPC:
  - **REST + JSON + OpenAPI**: use when crossing organizational boundaries, serving
    browser or mobile clients, or building public APIs. No code generation required;
    human-readable; every HTTP client speaks it; tooling (Swagger UI, curl) is universal.
  - **gRPC + Protobuf**: use for internal service-to-service communication where you
    control both ends. Smaller binary payload, HTTP/2 multiplexing, bidirectional
    streaming, schema enforced by the compiler, built-in evolution via field tags.
    Harder to debug without specialized tooling; browsers cannot call gRPC directly
    without a proxy (grpc-web).
  - Do not use REST for high-throughput internal APIs that transfer large volumes of
    structured data — the JSON serialization overhead is real.
- **Do** consider a **service mesh** (Istio, Linkerd) when a microservices deployment
  reaches the point where cross-cutting concerns (mTLS between services, distributed
  tracing, circuit breaking, retries, load balancing) are being reimplemented
  independently in each service. A service mesh deploys a sidecar proxy alongside each
  service container and handles these concerns at the network layer, transparently,
  without application code changes. Adopt a service mesh deliberately — it adds
  operational complexity and a learning curve. It is not the right first step for a
  small deployment.
- **Do** prefer **event-driven (dataflow) communication** over synchronous RPC between
  services when possible. Subscribing to a stream of changes (exchange rate updates,
  inventory changes) and maintaining a local copy eliminates a synchronous network call
  on every request — "the fastest and most reliable network request is no network request
  at all." Reserve synchronous RPC for operations that genuinely require an immediate
  response.
- **Do** use cloud-managed services for databases, queues, and object stores unless you
  have a specific, demonstrated reason to self-host (cost, latency, compliance, feature
  gap). Managed services eliminate an entire class of operational failures.
- **Do** design for **variable workloads** when using cloud services: autoscaling groups,
  serverless functions, and spot/preemptible instances only make financial sense when
  load is genuinely bursty. Steady, predictable workloads are cheaper on reserved or
  on-premise hardware.
- **Do** be aware of **serverless cold starts**: function-as-a-service (AWS Lambda, Google
  Cloud Functions) incurs initialization latency on first invocation. For
  latency-sensitive user-facing APIs, use pre-warmed containers or provisioned concurrency.
- **Do** consider vendor lock-in risk when building on proprietary cloud APIs. Use
  standard interfaces (S3-compatible object stores, Kafka-compatible brokers) where
  possible.
- **Do not** store sensitive customer data in vendor-managed cloud services without
  reviewing compliance posture, data processing agreements, and jurisdiction.

---

## Summary of Anti-patterns

| Anti-pattern | Consequence | Correct approach |
|---|---|---|
| Run analytical queries against OLTP database | User-facing latency spikes; DB saturation | Separate OLAP system (warehouse, read replica) |
| Dual-write to database + search index | Inconsistency on partial failure | CDC pipeline from database to all derived stores |
| Use wall-clock time to order distributed events | Wrong ordering; nonlinearizable LWW | Logical or hybrid logical clocks |
| Short timeout for failure detection under load | Cascading failure; healthy node declared dead | Adaptive timeouts; Phi Accrual detector |
| Language-native serialization (pickle, Java Serializable) | Language lock-in; RCE risk; no version tolerance | Protobuf or Avro with explicit schema evolution |
| Snapshot isolation for "check then act" multi-row logic | Write skew; double-booking; overdraft | Serializable isolation (SSI) |
| Read-modify-write via ORM | Lost updates under concurrency | Database atomic operations (`UPDATE x = x + 1`) |
| Rely on TCP/DB transactions for end-to-end exactly-once | Duplicate operations on client retry | Application-level idempotency key + dedup |
| Side effects inside batch job loop | Inconsistency when job fails mid-run | Stage output; commit atomically at end |
| Use processing time for stream windowing | False metric spikes after restart | Embed event time in source; use event-time windows |
| Query remote DB in hot path of stream processor | Latency spike; downstream overload | CDC to local state store; stream-table join |
| Distribute before single-node is insufficient | Unnecessary complexity, latency, cost | Measure first; distribute only when proved necessary |
| Implement business logic in stored procedures | Untestable, undeployable, unversionable | Stateless application servers + state in database |
| Collect data speculatively "in case it's useful" | GDPR liability; breach exposure surface | Data minimization; collect only what is used now |
| Assume Dynamo-style quorum reads are linearizable | Race conditions; stale reads | Linearizable storage (Spanner, CockroachDB) for strong consistency |
| Ignore replication lag in application code | Users see their own writes disappear | Read-after-write consistency; monotonic reads |
| Use auto-increment as hash partition key | All writes on last partition (hot spot) | Hash partition key or add randomized prefix |
| Declare leader by timeout without fencing | Split-brain; two leaders accept writes | Fencing tokens from consensus service (ZooKeeper, etcd) |
| Write to object store and assume instant global visibility | Race condition if another process reads before propagation | Use versioned object keys; check etag/generation before processing |
| LWW conflict resolution on mutable records | Silent data loss on concurrent writes | Conflict avoidance (single-leader per record) or CRDT merge |
| Reorder function calls in existing workflow definition | In-flight executions replay incorrectly | Deploy new workflow version; drain old executions first |
| Nondeterministic code in durable execution workflow | Replay diverges; undefined state | Use framework's deterministic time/random wrappers |
| No distributed tracing in multi-service system | Impossible to correlate events across services | OpenTelemetry trace IDs propagated through all calls and messages |
| ML training data reflects historical discrimination | Model learns and amplifies existing bias | Audit training data; test for disparate impact; add appeal mechanism |
| Synchronous RPC between microservices for every interaction | Tight coupling; cascading failure on dependency outage | Event-driven dataflow with local state store for reference data |
| Expose internal DB schema directly to CDC consumers | Schema change breaks downstream services | Outbox pattern: write stable outbox table; point CDC at outbox |
| Enable unclean leader election without understanding the tradeoff | Committed writes silently lost on leadership change | Disable unclean election; consumers must handle gaps if enabled |
| Write to storage then notify via separate channel | Consumer reads from lagging replica, misses the data | Linearizable storage, or embed data in notification, or pass min_version token |
| Fan out request to 100 backends in parallel | p99 of one backend becomes p63 of whole request | Bound fan-out; hedge requests; per-backend timeouts |
| No dead letter queue on stream consumer | Malformed message loops forever, blocking partition | Configure DLQ after N retries; alert on any DLQ message |
| Join stream to current dimension table snapshot | Nondeterministic on reprocessing; wrong historical results | Version dimension records; join on version ID at time of event |
| Average p99 latency values across instances | Mathematically invalid; hides true tail latency | Merge histograms; compute percentile from merged histogram |
| SLO = SLA (no buffer between internal target and customer contract) | No time to detect and fix breach before customer impact | Set SLO tighter than SLA; track error budget burn rate |
| Use REST+JSON for high-throughput internal service APIs | JSON serialization overhead; no schema enforcement | gRPC+Protobuf for internal; REST+JSON for external/browser |
| Long read-write transaction under SSI | High abort rate as conflict window grows | Compute outside transaction; commit in a short transactional step |
