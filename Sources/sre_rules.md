# Site Reliability Engineering Rules

Rules derived from *Site Reliability Engineering: How Google Runs Production Systems* edited
by Betsy Beyer, Chris Jones, Jennifer Petoff, and Niall Richard Murphy. The single organizing
principle: **reliability is a feature — design, measure, and budget for it explicitly, or you
will not have it**.

---

## 1. Cap operational work at 50% to protect engineering capacity

SRE teams that spend more than half their time on operational work — tickets, on-call
interrupts, manual toil — have no capacity left to eliminate the root causes of that work.
The 50% rule is not a guideline; it is a structural ceiling enforced by redirecting excess
load back to the development team.

- **Do** track the split between engineering work (projects, automation, capacity planning)
  and operational work (incidents, tickets, manual tasks) every week. If operational work
  exceeds 50% for more than two consecutive quarters, escalate to engineering leadership.
- **Do** redirect excess pages and tickets to the development team when the SRE team is over
  the cap. This creates the right incentive for developers to fix reliability at the source
  rather than externalizing it to SRE.
- **Do not** absorb unbounded operational load silently. A team that accepts every interrupt
  signals to developers that reliability problems have no cost, which guarantees more of them.
- **Do** treat the 50% cap as a system property, not a personal one. Individuals should not
  heroically absorb excess load; the system should reject it structurally.
- **Do** measure ticket volume, page rate, and time-to-resolve as lagging indicators of
  operational health. Rising trends predict future capacity crises before they arrive.
- **Do** invoke "give back the pager" when a service's operational load persistently exceeds
  the cap and structural fixes require multiple quarters to implement. SRE formally returns
  on-call responsibility to the development team while both teams work together to make the
  service operationally sustainable. Without the credible option to return the pager, the
  cap has no enforcement mechanism.

---

## 2. Treat simplicity as a reliability property

Complexity that was not required by the problem — accidental complexity — is a permanent
reliability liability. Every line of code that does not need to exist is a bug that has not
been written yet. "Boring" software is a virtue, not a failure of imagination.

- **Do** distinguish *essential complexity* (inherent in the problem domain) from *accidental
  complexity* (introduced by implementation choices, abstractions added prematurely, or
  features nobody asked for). Push back on accidental complexity at design review.
- **Do** use the *negative lines of code* metric: a change that deletes code while preserving
  behavior is almost always an improvement. A smaller codebase has fewer defects, is faster
  to onboard, and is easier to test.
- **Do not** treat surprise as acceptable in production systems. Unexpected behavior —
  however clever — is a reliability failure. Predictable, boring systems that stick to their
  documented behavior are easier to operate than systems that are interesting.
- **Do** have SRE teams push back when accidental complexity is introduced into systems they
  are responsible for. Accepting a service that is unnecessarily complex means accepting the
  operational burden of that complexity indefinitely.
- **Do** favor simple, explicit failure modes over sophisticated recovery logic. A service
  that fails loudly and obviously is faster to diagnose than one that degrades silently in a
  complex way.
- **Do not** add features, configuration options, or abstraction layers speculatively. The
  operational cost of each is paid every time the system is understood, debugged, or handed
  to a new engineer.

---

## 3. Define SLIs, SLOs, and error budgets before you ship

An SLO without a corresponding error budget is just a wish. An error budget converts an
abstract reliability target into a concrete quantity of unreliability you are allowed to
spend, which makes every architectural and release decision accountable.

- **Do** define a Service Level Indicator (SLI) as a specific, measurable ratio: successful
  requests divided by total requests, latency below a threshold, queue depth below a limit.
  SLIs must be measurable from the user's perspective, not from internal instrumentation.
- **Do** set a Service Level Objective (SLO) as a target percentage for that SLI over a
  rolling window (e.g., 99.9% of requests succeed over a 30-day rolling window). SLOs must
  be strict enough to matter and loose enough to allow shipping.
- **Do not** set SLOs at 100%. 100% is not achievable, not measurable with confidence, and
  eliminates all error budget, which halts all risk-taking including legitimate feature
  development.
- **Do** recognize that reliability has sharply diminishing returns past the point users can
  observe. The cost of each incremental improvement compounds nonlinearly — an additional
  nine of availability can cost 100× the previous one. A user on a 99% reliable smartphone
  cannot distinguish 99.99% from 99.999% service reliability. Optimizing beyond the point
  users notice wastes engineering capacity that could ship features or reduce technical debt.
  Error budgets make this trade-off explicit rather than leaving it to negotiation.
- **Do** compute the error budget as `(1 − SLO) × window`. 99.9% over 30 days = 43.8 minutes
  of budget. Track actual consumption against this budget in real time.
- **Do** use error budget depletion as the primary signal for release decisions: when the
  budget is healthy, release aggressively; when it is exhausted, freeze releases and focus on
  reliability work. This removes subjective negotiation from the reliability conversation.
- **Do not** let the development team and SRE team fight over the SLO number. Derive the SLO
  from user expectations and business requirements, then make both teams accountable to it.
- **Do** distinguish the SLO (internal target) from the SLA (contractual commitment to
  customers). The SLO should be tighter than the SLA so you have headroom before breaching
  contractual obligations.

---

## 4. Alert on symptoms, not causes

Alerting on causes (CPU usage, disk full, process restarts) generates noise and fatigue
without improving user outcomes. Alerting on symptoms (error rate elevated, latency
degraded) keeps on-call engineers focused on what actually affects users. Monitoring output
exists in exactly three forms — anything else is a design error.

- **Do** classify every monitoring output into exactly one of three tiers:
  - **Alerts** — a human must act immediately; page the on-call engineer.
  - **Tickets** — a human must act within days; no damage results from deferral.
  - **Logging** — no one needs to read this; recorded only for forensics and diagnostics.
  Any alert that does not require immediate action is not an alert — it is a ticket or a log
  entry that has been miscategorized.
- **Do** have software interpret monitoring signals and decide which tier applies. Monitoring
  should never require a human to interpret any part of the alerting domain; humans should
  only be notified when they need to take action.
- **Do** define alerts in terms of the four golden signals: latency (request duration),
  traffic (request rate), errors (failed requests), and saturation (resource utilization
  approaching limits). These four signals are sufficient to detect virtually every
  user-visible failure.
- **Do** alert on latency at the 99th percentile, not the mean. Averages mask tail latency.
  A service that serves 99% of requests in 1 ms and 1% in 30 s has an average of ~1.3 ms
  — the mean looks fine while 1% of users experience a 30-second wait.
- **Do** use multi-window, multi-burn-rate alerting for SLO violations: fire a page when the
  burn rate is high over a short window (fast burn — you will exhaust the error budget in
  hours) and a ticket when the burn rate is moderate over a long window (slow burn — you will
  exhaust it in days).
- **Do not** rely on whitebox (internal) metrics alone. Couple whitebox metrics with blackbox
  probing (synthetic requests from outside the system) to detect failures that internal
  instrumentation cannot see.
- **Do** periodically audit every alert: "Did this page require immediate action? Did it
  have a documented runbook? Was it a false positive?" Delete or demote alerts that fail
  this audit. A runbook that says "restart the service" without explaining why is not a
  runbook.

---

## 5. Measure and eliminate toil

Toil is manual, repetitive, automatable operational work that scales with service growth
and produces no enduring improvement. It is distinct from overhead (meetings, reviews) and
from engineering work that has lasting value. Eliminating toil is the primary productivity
lever for SRE teams.

- **Do** classify a task as toil if it is manual, repetitive, triggered by a service (not by
  a human decision), lacks permanent value, and scales O(n) with service traffic or size. If
  all five apply, automate or eliminate it.
- **Do not** confuse toil with necessary overhead. Writing postmortems, reviewing designs,
  mentoring — these are not toil even though they are not coding. They produce lasting value.
- **Do** keep toil below 50% of each SRE engineer's time (see Rule 1). When an individual
  engineer's toil exceeds this threshold, treat it as a signal that a service or process
  needs engineering attention.
- **Do** automate toil incrementally rather than waiting for a perfect system. A script that
  handles 80% of a class of tickets is worth shipping immediately; the remaining 20% can be
  handled manually until the next iteration.
- **Do not** automate toil that should be eliminated at the source. If a service generates
  hundreds of restart tickets because of a memory leak, the right fix is the memory leak —
  not a more sophisticated restart script.
- **Do** track toil volume over time as a team metric. Flat or rising toil despite automation
  investment indicates that service growth is outpacing automation, or that new sources of
  toil are not being identified.

---

## 6. Design for graceful degradation under load

A service that fails hard under load is harder to recover than one that sheds work cleanly.
Design explicit degradation paths: which features can be disabled, which responses can be
cached, and which requests can be rejected — before load reaches critical thresholds.

- **Do** implement load shedding at every entry point. When the system is saturated, return
  explicit errors (429 Too Many Requests) rather than silently queuing work that will time
  out anyway, compounding the overload.
- **Do** use adaptive throttling (client-side) to complement server-side load shedding.
  Clients track their own accept/reject ratio and throttle their own request rate before
  sending. This prevents thundering-herd retry storms from converting a partial outage into
  a total one.
- **Do not** let retry logic operate without backoff and jitter. Exponential backoff with
  random jitter is the minimum acceptable retry strategy. Linear retries without jitter cause
  synchronized storms.
- **Do** use request deadlines throughout the call graph. A 10-second deadline at the edge
  must be propagated as a shorter deadline to all downstream calls, so that no downstream
  service does work whose result will arrive too late to be used.
- **Do not** propagate deadlines naively by passing the original deadline unchanged. Subtract
  transmission overhead and local processing time so that deadlines shrink as they traverse
  the call graph, not remain flat.
- **Do** design explicit fallback responses for non-critical features: serve a cached result,
  return an empty list, or disable the feature entirely. A page that loads without
  personalized recommendations is better than a page that does not load.
- **Do** use circuit breakers to stop sending requests to a known-failing dependency. An open
  circuit allows the dependency time to recover without being hammered by retries from all
  upstream clients simultaneously.

---

## 7. Prevent cascading failures by controlling queue depth and concurrency

Cascading failures propagate because overloaded services accept work they cannot complete,
delaying responses, causing upstream timeouts, causing upstream retries, and increasing the
load further. The mechanism is a feedback loop; break it by bounding queues and rejecting
excess early.

- **Do** set explicit concurrency limits on all thread pools, connection pools, and request
  queues. Unbounded queues under load increase latency without increasing throughput and
  eventually cause out-of-memory failures.
- **Do not** assume that adding replicas solves a cascading failure in progress. Under load,
  new replicas often cannot initialize fast enough, and the load that caused the cascade is
  still arriving. Shed load first; add capacity second.
- **Do** implement health checks that reflect actual readiness, not just process liveness.
  A server that is alive but whose database connection pool is exhausted should return
  unhealthy from its load-balancer health check so it stops receiving traffic.
- **Do** use flow control at the RPC level: if a service's response queue is full, it should
  apply backpressure to the caller rather than buffering indefinitely.
- **Do** protect shared infrastructure (databases, caches, rate-limit stores) from cascading
  failure by adding a local cache layer. Cache failures should never be propagated as hard
  errors to the caller; serve stale data instead.
- **Do not** design services that fail open by default when a dependency is unavailable.
  Fail open is appropriate only when the missing signal is truly non-critical; otherwise,
  prefer degraded responses to silent correctness failures.

---

## 8. Load balance at every layer, not just at the edge

Edge load balancers distribute traffic across replicas but cannot see the state of replicas'
internal queues. Subsetting, power-of-two-choices, and least-loaded selection at the RPC
layer reduce hotspots that edge balancers are blind to.

- **Do** implement least-loaded (or power-of-two-choices) load balancing at the RPC client
  layer, not only at L4/L7 edge balancers. Edge balancers distribute connections; they do
  not balance work within a long-lived connection.
- **Do** use subsetting to limit the number of backends each client connects to. A client
  that maintains connections to every backend creates a fan-out problem when the backend
  count scales to thousands; subsetting bounds connection overhead on both sides.
- **Do not** use round-robin in isolation when request cost varies significantly. A slow
  request to one backend creates a queue while round-robin continues sending new requests to
  that same backend.
- **Do** weight backends by their current utilization when the client has that signal. A
  backend recovering from a GC pause should receive less traffic until its latency
  normalizes.
- **Do** co-locate replicas across failure domains (availability zones, power domains,
  network domains) so that a single infrastructure failure does not take out all replicas of
  a service simultaneously.

---

## 9. Never implement your own distributed coordination

Leader election, distributed locking, and critical shared state are problems with known,
subtle failure modes under network partition. Informal implementations — database row locks,
Redis SETNX, heartbeat-based leadership — have all been proven insufficient in production.
Use a system that has been formally proven and battle-tested.

- **Do** use a proven distributed consensus system (etcd, ZooKeeper, Consul, or Google's
  Chubby) for any problem that requires leader election, distributed locking, or agreement
  on shared state across processes.
- **Do not** implement leader election with a database row, a filesystem lock, or a cache
  entry. These mechanisms do not handle network partitions correctly and produce split-brain:
  two nodes simultaneously believe they hold the lock, silently corrupting state.
- **Do not** use heartbeat-based "leadership" without a fencing token. A leader that is slow
  or paused (GC, I/O stall) will miss heartbeats; if a new leader is elected while the old
  one is still writing, you have split-brain even with heartbeats.
- **Do** treat split-brain as the specific failure mode to design against. Any distributed
  coordination scheme that cannot answer "what happens if the network partitions for 30
  seconds?" is not safe for production.
- **Do** use the Paxos or Raft family of protocols (or a library that implements them) rather
  than designing a novel protocol. The published protocols have known safety proofs and
  known implementation pitfalls; novel protocols have neither.
- **Do** scope distributed locks to the minimum necessary duration and always associate them
  with a lease expiry. A lock held indefinitely by a crashed process must expire
  automatically, not require manual intervention.

---

## 10. Run structured, blameless postmortems within 48 hours

A postmortem that focuses on human error instead of system properties produces blame and
conceals the structural conditions that made the error possible. The goal is a written
record that makes the system safer for the next engineer, not a verdict on the last one.

- **Do** write a postmortem for every incident that violated an SLO or had significant user
  impact. The threshold should be explicit, not left to judgment of the on-call engineer who
  is usually exhausted after an incident.
- **Do** complete the postmortem within 24–48 hours while details are fresh. A postmortem
  written a week later is a reconstruction, not a record.
- **Do** structure every postmortem with: timeline (what happened and when), root cause
  analysis (proximate and distal causes), impact (duration, users affected, SLO budget
  consumed), and action items with owners and due dates.
- **Do not** write "human error" as a root cause. Human error is always a symptom of a
  system that made the error easy to make. Ask: what made this mistake possible? What would
  have caught it earlier? What would have reduced the blast radius?
- **Do** track action items from postmortems in a bug tracker with owners and deadlines, not
  in the postmortem document itself. Postmortems with untracked action items are theater.
- **Do** share postmortems across teams. The value of a postmortem is proportional to how
  many engineers read it. An incident in service A often contains lessons directly applicable
  to services B, C, and D.
- **Do not** let "we fixed the immediate problem" substitute for addressing structural causes.
  The most important action items are usually not the ones that resolve the current incident
  but the ones that prevent the next one.

---

## 11. Use the Incident Command System for large incidents

Incidents with multiple teams, many responders, and evolving scope fail under ad-hoc
coordination. Roles become ambiguous, information is lost, and the person doing the most
technical work is simultaneously expected to communicate status to executives. Separate
these functions explicitly.

- **Do** designate an Incident Commander (IC) who owns communication and coordination, not
  technical investigation. The IC does not need to be the most senior engineer; they need to
  be available to manage the room.
- **Do** designate separate Operations Lead (technical triage), Communications Lead (external
  and internal status updates), and Planning Lead (resource management, documentation) for
  large incidents. These are roles, not people — one person can hold multiple roles in a
  small incident.
- **Do not** let the IC become the primary debugger. If the IC is deep in a shell session,
  they are not monitoring the big picture, coordinating across teams, or communicating to
  stakeholders.
- **Do** maintain a running incident document in real time: hypotheses tested, commands run,
  results observed. This document survives engineer handoffs and feeds the postmortem.
- **Do** handoff command explicitly when an incident spans multiple shifts. "You now have
  command" spoken aloud, confirmed by the incoming IC, with a verbal and written summary of
  current status and next steps.
- **Do** maintain on-call playbooks and practice them. A documented playbook produces roughly
  a 3× improvement in mean time to repair compared to improvising. Run *Wheel of Misfortune*
  exercises: spin a random past incident scenario, have an on-call engineer walk through
  diagnosis and remediation live, then debrief. The engineer who has practiced the scenario
  once is far more effective than one encountering it for the first time under pressure.
- **Do** declare the incident over explicitly and conduct a brief debrief immediately after
  — while memory is fresh and before engineers disperse to other tasks.

---

## 12. Test reliability the same way you test functionality

Untested reliability properties decay invisibly. A service whose disaster recovery procedure
has not been executed in 18 months is a service whose disaster recovery procedure does not
work. Test reliability through the same engineering discipline applied to features.

- **Do** follow the testing hierarchy in order: unit tests (individual functions), integration
  tests (assembled components with dependencies injected), system tests (full-stack
  correctness including smoke, regression, and performance), stress tests (breaking-point
  behavior under load), canary tests (structured user acceptance: deploy to a small live
  traffic subset, bake, observe for unexpected variance before expanding). Each level catches
  a different class of failure; skipping levels creates blind spots that appear in production.
- **Do** write unit tests for alert rules, runbooks, and on-call decision trees, not just
  for application code. An untested alert that fires incorrectly during an incident is worse
  than no alert.
- **Do** run integration tests against realistic load profiles, not just functional
  correctness. A service that passes all unit tests but degrades at 10× baseline traffic has
  not been tested.
- **Do** conduct load tests and stress tests before each major launch. Identify the service's
  breaking point, confirm that load shedding fires at the right threshold, and verify
  graceful degradation behavior.
- **Do** run disaster recovery drills (game days) on a scheduled basis. Simulate the
  scenarios in your runbooks: zone failure, database failover, traffic spike. Discover that
  the runbook is wrong in a drill, not in an outage.
- **Do not** test in production exclusively. Production testing (canary releases, feature
  flags, chaos engineering) is valuable but requires safe fallback paths that can only be
  validated in a lower environment first.
- **Do** use canary deployments for every release: route a small fraction of traffic to the
  new version, compare error rate and latency against the baseline, and automate rollback if
  metrics diverge. Never release to 100% of traffic without a canary phase.
- **Do** treat rollback as a first-class operation. Rollback must be faster than forward fix.
  A release that cannot be rolled back in under five minutes should not be released without
  a dark launch or feature flag.

---

## 13. Gate SRE handoff on a Production Readiness Review

An SRE team that accepts pager responsibility for a service without a Production Readiness
Review (PRR) inherits undefined SLOs, missing dashboards, untested runbooks, and unknown
capacity headroom. The PRR is the enforcement mechanism for the SRE engagement model:
deficiencies discovered during review must be fixed by the development team before SRE
accepts the pager.

- **Do** treat a completed PRR as a prerequisite for SRE accepting on-call responsibility
  for any service. This is not a courtesy review; it is a gating criterion.
- **Do** use a PRR checklist that covers at minimum: SLOs defined and documented, dashboards
  cover all SLO metrics, alerting configured and tested, on-call runbooks exist and have been
  reviewed, rollback procedure has been executed at least once, and capacity headroom is
  confirmed for expected peak load plus safety margin.
- **Do not** accept a service whose dashboards do not cover the metrics defined in the SLO.
  A dashboard that does not show SLO compliance is not a dashboard for the on-call engineer;
  it is decoration.
- **Do** assign action items from the PRR to the development team with due dates. SRE does
  not take the pager until the action items are closed — this creates the right incentive for
  developers to invest in production readiness before handoff, not after.
- **Do** revisit the PRR checklist when a service undergoes significant architectural change
  or traffic growth. A service that passed a PRR two years ago may no longer meet the same
  criteria at 10× its original scale.
- **Do** use the PRR as a learning tool, not only as a gate. The checklist accumulated by
  the SRE team represents the collective memory of past failures; a development team that
  reads it learns from incidents they did not personally experience.

---

## 14. Manage on-call load to sustain engineer health and judgment

On-call is high-stakes work that degrades in quality when engineers are overloaded. A team
that is paged more than twice per on-call shift has insufficient time to investigate each
incident properly, write postmortems, and sleep. The system produces both bad outcomes and
burned-out engineers.

- **Do** target fewer than two significant events per on-call shift (8–12 hour period). More
  than this indicates either that the alerting threshold is too low or that there are
  unresolved reliability problems generating chronic noise. Each incident — root-cause
  analysis, remediation, postmortem, follow-up bug fixes — takes roughly six hours; two per
  shift is the arithmetic maximum for thorough handling.
- **Do** cap on-call time at 25% of each SRE engineer's total working time, not 50%. The
  target breakdown is 50% engineering work, 25% on-call, 25% other operational work (tickets,
  manual tasks, non-urgent toil). On-call is one component of operational work, not all of it.
- **Do** ensure every engineer carries the pager at least once or twice per quarter.
  Operational underload is as damaging as overload: an engineer who is never paged loses
  familiarity with the system's failure modes and cannot contribute to improving them. A
  team sized so that some engineers never carry the pager is too large for the service.
- **Do** structure on-call with a primary and a secondary. The secondary escalates when the
  primary is unreachable — not when the primary is overwhelmed. Volume overload is a system
  problem, not a reason to add more humans to the same alert storm.
- **Do** use follow-the-sun rotations for globally operated services: hand the pager to a
  team in a timezone where it is daytime. No engineer should routinely carry a pager through
  their sleeping hours. A dual-site team needs a minimum of six engineers per site (with
  week-long shifts) to honor the 25% on-call cap; a single-site team needs at least eight.
- **Do** give on-call engineers the authority to act: to escalate, to roll back a release, to
  shed load, to take a service offline. On-call engineers who cannot act without management
  approval will be slow to act when speed matters most.
- **Do** compensate on-call work explicitly and fairly. Engineers who carry pagers must see
  the pager as part of a reciprocal relationship with the organization, not an invisible tax.
- **Do not** require on-call engineers to be available during business hours the day after a
  significant overnight incident without rest. Cognitive performance after sleep deprivation
  is comparable to impairment from alcohol; decisions made in this state are not reliable.
- **Do** conduct weekly on-call reviews: examine every page, every ticket, every interrupt
  from the prior week. Categorize each as actionable (led to a fix), informational (useful
  but no action needed), or noise (should not have fired). Eliminate the noise category
  systematically.
- **Do** rotate on-call across the full team, not only among the engineers who know the most
  about the system. Engineers who never carry a pager do not internalize the operational
  consequences of the code they write.

---

## 15. Make troubleshooting a systematic process, not an art

Troubleshooting under pressure degrades when it is treated as intuition. Stress hormones
impair cognitive function and induce confirmation bias — the tendency to interpret new
evidence as confirming the first hypothesis rather than refuting it, even when the evidence
actually points elsewhere. A structured diagnostic method — hypothesis generation,
prediction, test, observation — counters this by externalizing reasoning and forcing
explicit hypothesis rejection before moving on.

- **Do** start with a clear problem statement: what is the observed behavior, what is the
  expected behavior, and what is the impact? Do not begin debugging until these three
  questions have written answers.
- **Do** form a hypothesis before running a command. "I expect that the error rate is
  elevated because of a bad canary deployment; I will check the canary's error rate." Then
  check. If the hypothesis is wrong, form a new one.
- **Do not** make changes to a production system during an incident without first recording
  the current state and being confident you can reverse the change. Changes that cannot be
  reversed convert a bad incident into a catastrophic one.
- **Do** use the simplest tool first. Check dashboards before running ad-hoc queries. Check
  logs before attaching a debugger. The most probable causes of production incidents are
  deployment changes, configuration changes, and traffic spikes — check these before looking
  for novel causes.
- **Do** divide the problem space by halving: identify the point in the call graph where
  behavior transitions from expected to unexpected. Binary search through the system is
  faster than tracing from either end.
- **Do** time-box each diagnostic hypothesis. If you have not confirmed or refuted a
  hypothesis in 10–15 minutes, abandon it and try another. Sunk-cost bias during incidents
  delays resolution.

---

## 16. Treat automation as a spectrum, not a binary

Automation is not simply "manual vs. automated." Intermediate stages — operator-assisted
automation, automation with human checkpoints, automation with manual override — are often
more appropriate and safer than fully autonomous systems, especially early in a system's
lifecycle.

- **Do** start automation at the lowest stage of autonomy appropriate for the risk: produce
  a recommendation and let a human execute it; then automate execution with human approval;
  then automate fully with monitoring and rollback. Move up the autonomy ladder as confidence
  in the automation grows.
- **Do not** automate before the manual process is well understood and documented. Automation
  encodes the understanding you have at the time of writing; if that understanding is wrong,
  the automation will be wrong faster than a human would be.
- **Do** value automation for consistency as well as speed. A human executing a procedure
  varies with fatigue, distraction, and experience; automation applies the same steps the
  same way every time. Consistency reduces the variance in outcomes, which is worth building
  even when the raw time savings are small.
- **Do** ensure that automated systems can be disabled quickly. Automation that cannot be
  turned off under adversarial conditions is a liability, not an asset.
- **Do** apply the value-of-automation calculation: (time saved per run × number of runs per
  year) must exceed (time to build + time to maintain). Many automation projects do not
  pay off because the system they automate changes faster than the automation can keep up.
- **Do** make automation failures loud and obvious. Silent automation that fails closed is
  dangerous; the operator does not know the work is not happening until the consequences
  arrive.
- **Do not** use automation to handle alerts that should be eliminated. Automated remediation
  of a recurring alert is a patch on top of a defect; the correct fix is the defect.

---

## 17. Enforce capacity planning as a regular engineering discipline

Capacity planning that happens only in response to latent crises is indistinguishable from
no capacity planning. Demand forecasting, provisioning lead times, and resource accounting
must be treated as engineering work with deadlines and owners.

- **Do** maintain a demand forecast that projects traffic growth 8–12 quarters out, updated
  quarterly. Provisioning lead times for hardware and network infrastructure often exceed 6
  months; a shorter forecast horizon guarantees you will be surprised.
- **Do** provision for N+2 redundancy for critical services: enough capacity to absorb one
  failure mode (N+1) with margin for a second simultaneous event and for planned maintenance.
  N+1 is not enough when a single failure triggers a cascading load shift.
- **Do not** treat peak traffic as the provisioning target. Provision for peak × safety
  margin. The appropriate margin depends on traffic predictability: services with smooth,
  predictable traffic can use a smaller margin; services with spiky or unpredictable traffic
  need more.
- **Do** account for resource efficiency losses at scale: GC pauses, connection overhead,
  log volume, and service mesh proxy CPU. Benchmarks run on small clusters do not predict
  behavior at production scale.
- **Do** plan for demand-side interventions as a complement to supply-side provisioning.
  Load shedding, feature flags to disable expensive features, and traffic shaping are cheaper
  than provisioning for every possible spike.

---

## 18. Require hermetic, reproducible builds and treat config as code

Roughly 70% of production outages are caused by changes to a live system — deployments,
configuration pushes, and dependency upgrades. The three practices that minimize this: progressive
rollouts (limit blast radius), fast automated detection (catch problems early), and automated
rollback (revert before impact compounds). A release that cannot be reproduced, rolled back, or
audited is a release that cannot be recovered from cleanly.

- **Do** use a build system that produces hermetic builds: all inputs (source, dependencies,
  toolchain) are pinned at specific versions and fetched from content-addressed storage.
  Given the same inputs, the build must produce the same binary on every machine.
- **Do** store build artifacts in a content-addressed artifact registry, not in a
  timestamp-named bucket. A binary identified by its hash can be verified; one identified by
  a timestamp cannot.
- **Do not** build directly from the main branch for releases. Build from a tagged commit with
  a verified pipeline. Ad-hoc builds from working directories are not reproducible and not
  auditable.
- **Do** use a staged release pipeline: build once, promote through dev → staging →
  production. Each stage validates the same artifact; never rebuild between stages.
- **Do** extend rollout duration in proportion to service risk. A routine service can be
  updated across all instances within hours; sensitive infrastructure should roll out over
  days, interleaved across geographic regions, so that a bad build's effects are visible in
  one region before it reaches the next.
- **Do** version and release configuration files through the same pipeline as binaries. The
  strongest form of this is packaging configuration as a separate versioned artifact promoted
  through the same staged pipeline — the hermetic principle applied to config. Binary
  configurations tightly bound to a specific binary version should be versioned and released
  together, ensuring that config skew between the checked-in version and the running version
  cannot occur.
- **Do not** allow configuration changes that bypass version control. "Just update the
  config" is the phrase that precedes a class of outages that cannot be reproduced,
  rolled back cleanly, or explained in the postmortem.
- **Do** include rollback artifacts in every release. Before releasing version N, confirm
  that version N−1 is available in the artifact registry and that its rollback procedure has
  been tested.

---

## Anti-patterns

| Anti-pattern | Why it fails |
|---|---|
| Paging on CPU > 80% | CPU is a cause, not a symptom; users are unaffected until latency or errors rise |
| SLO set to 100% | Eliminates all error budget, halts all risk-taking including legitimate releases |
| "Human error" as root cause | Conceals the system property that made the error easy; produces blame, not improvement |
| Unbounded retry loops without backoff | Converts partial outage into total collapse via synchronized retry storm |
| On-call absorbs all operational load silently | Signals to developers that reliability has no cost; guarantees growing toil |
| Automating before understanding the manual process | Encodes wrong assumptions at machine speed |
| Undifferentiated alerts (all pages equal) | Fatigue causes legitimate alerts to be treated as noise |
| Deploying to 100% without canary | First signal of a bad release is at full blast radius |
| Mean latency as the SLI | Hides tail latency; 99th percentile is the relevant metric for user experience |
| Queue depth unbounded | High load increases latency without bound; triggers OOM before load shedding |
| Postmortems without tracked action items | Produces documentation theater; systemic problems recur |
| Release freeze as default reliability response | Stops features but not background load growth; does not improve reliability |
| N+1 redundancy for critical services | One failure leaves no margin for a simultaneous second event |
| Configuration changes outside version control | Configuration causes as many outages as code; gap in audit trail |
| Incident Commander also debugging | IC loses situational awareness; coordination and communication degrade |
| On-call rotation only among senior engineers | Others never develop operational intuition; reliability knowledge concentrates dangerously |
| Runbooks that say "restart the service" | Symptom treatment without understanding; same incident recurs |
| Dashboards that display internal metrics only | Blackbox failures (network, load balancer, DNS) are invisible to whitebox metrics |
| Capacity plan updated only after a capacity event | Provisioning lead times exceed reaction time; surprise is guaranteed |
| Error budget shared between SRE and dev as a zero-sum resource | Produces conflict rather than shared incentive to maintain reliability |
| Rolling your own leader election or distributed lock | Informal approaches have known failure modes under partition; split-brain causes silent data corruption |
| Taking the pager before a PRR is complete | SRE inherits a service with undefined SLOs, missing dashboards, and untested runbooks |
| Treating accidental complexity as inevitable | Complexity that wasn't required by the problem accumulates, making every future change riskier and slower |
