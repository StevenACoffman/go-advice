# DevOps Rules for Claude

Derived from *The DevOps Handbook* by Gene Kim, Jez Humble, Patrick Debois, and John Willis.

These rules describe the technical practices and cultural norms of high-performing technology organizations. Apply them when designing systems, reviewing code, structuring teams, or advising on delivery processes.

---

## Table of Contents

1. [The Three Ways: Core Principles](#1-the-three-ways-core-principles)
2. [Deployment Lead Time as the Primary Metric](#2-deployment-lead-time-as-the-primary-metric)
3. [Version Control Everything](#3-version-control-everything)
4. [Trunk-Based Development and Continuous Integration](#4-trunk-based-development-and-continuous-integration)
5. [The Deployment Pipeline](#5-the-deployment-pipeline)
6. [Automated Testing Hierarchy](#6-automated-testing-hierarchy)
7. [Infrastructure as Code and Environment Consistency](#7-infrastructure-as-code-and-environment-consistency)
8. [Production Telemetry and Observability](#8-production-telemetry-and-observability)
9. [Safe Deployments and Rapid Recovery](#9-safe-deployments-and-rapid-recovery)
10. [Feedback Loops and On-Call Culture](#10-feedback-loops-and-on-call-culture)
11. [Blameless Post-Mortems and Just Culture](#11-blameless-post-mortems-and-just-culture)
12. [Security and Compliance as Code](#12-security-and-compliance-as-code)
13. [Organizational Design and Conway's Law](#13-organizational-design-and-conways-law)
14. [Continual Learning and Improvement](#14-continual-learning-and-improvement)
15. [Anti-Patterns](#15-anti-patterns)

---

## 1. The Three Ways: Core Principles

The Three Ways are the underpinning principles from which all DevOps behaviors derive. Every technical and organizational practice in this document serves one or more of these ways.

### The First Way: Flow (Left to Right)

Optimize for fast, smooth flow from Development → Operations → Customer. Maximize global throughput, not local efficiency.

**Practices that enable Flow:**
- Make work visible at every stage — kanban boards, build dashboards, deployment status
- Limit Work in Progress (WIP) — context-switching destroys throughput; unblocking a stalled work item beats starting a new one
- Reduce batch sizes — a single commit deployed today beats ten commits deployed next month
- Reduce handoffs — each handoff introduces queue time and information loss
- Eliminate waste: partially done work, extra processes, extra features, task switching, waiting, defects, heroics

**Key metric:** *Deployment lead time* — time from code commit to running in production. Target: minutes to hours, never months.

### The Second Way: Feedback (Right to Left)

Create fast, constant feedback at every stage of the value stream. Amplify signals to prevent recurrence and embed knowledge where the work is done.

**Practices that enable Feedback:**
- Create telemetry at every layer (application, infrastructure, business)
- Surface problems at the earliest possible stage — a failing unit test costs 10× less than a failing integration test, 100× less than a production incident
- Swarm problems: when the deployment pipeline breaks, stop new work until it is fixed
- Make production metrics visible to everyone in the value stream, including developers

### The Third Way: Continual Learning and Experimentation

Build a generative, high-trust culture where learning from both successes and failures is institutionalized and multiplied across the organization.

**Practices that enable Learning:**
- Treat failures as learning opportunities, not occasions for blame
- Convert local discoveries into global improvements (shared code, runbooks, post-mortems)
- Deliberately inject failure to strengthen systems and organizational responses
- Reserve explicit time for improvement work — if improvement only happens when there is spare capacity, it never happens

---

## 2. Deployment Lead Time as the Primary Metric

Deployment lead time — from commit to production — is the single most important measure of flow. It predicts quality, reliability, and employee satisfaction.

**Target states:**
| Organization type | Deployment lead time | Deploy frequency |
|---|---|---|
| High performer | Minutes to hours | On demand (multiple/day) |
| Medium performer | Days to weeks | Weekly to monthly |
| Low performer | Weeks to months | Monthly to quarterly |

**What enables sub-hour lead time:**
- Automated build, test, and deployment pipeline
- Production-like environments available on demand
- Modular, loosely-coupled architecture so small teams deploy independently
- Trunk-based development with no long-lived feature branches
- Comprehensive automated test suite completing in < 10 minutes (commit stage)

**What %C/A reveals:** Measure *percent complete and accurate* at each stage — the fraction of work received that is usable without correction. A low %C/A reveals where rework is injected into the value stream. Fix the upstream process, not the downstream symptom.

---

## 3. Version Control Everything

Version control is not just for application source code. Every artifact needed to reproduce a production system must live in version control.

**What belongs in version control:**
- Application code and dependencies (vendored or locked)
- All environment configuration — dev, test, staging, production
- Infrastructure-as-code (Terraform, Pulumi, CloudFormation, Ansible playbooks)
- Container definitions (Dockerfiles, compose files, Kubernetes manifests)
- Database schema migrations
- Network configuration: DNS zone files, firewall rules, load balancer config
- Deployment scripts and pipeline definitions (Jenkinsfile, .github/workflows, etc.)
- Automated tests at all levels
- Operational runbooks and post-mortem reports

**Why this matters:** The ability to reproduce any environment from version control is the prerequisite for every other practice. Without it, "it worked in staging" is unfalsifiable. The Puppet Labs State of DevOps Report identified *version control for operations work* as one of the highest predictors of IT performance.

**Artifact repositories:** Compiled binaries and container images are stored in artifact repositories (Nexus, Artifactory, ECR), not re-built from source per environment. Build once, deploy everywhere.

---

## 4. Trunk-Based Development and Continuous Integration

Long-lived feature branches are incompatible with fast flow. The longer a branch lives, the more painful integration becomes.

**Rules:**
- All developers commit to trunk (main) at least once per day
- Feature branches, if used, live for hours, not days; never weeks
- Merge conflicts discovered at commit time cost 100× less to fix than at integration time
- Every commit triggers the full automated pipeline

**Techniques for committing incomplete work safely:**
- *Feature flags (feature toggles):* Deploy code dark; activate per-cohort or per-environment via config
- *Branch by abstraction:* Introduce an abstraction layer, move callers incrementally, remove the old implementation
- *Dark launch:* Execute new code paths in production alongside old ones, compare results, discard output

**HP LaserJet Firmware result:** Moving from 20 long-lived branches (6-week regression cycles) to trunk-based development produced: regression testing 6 weeks → 1 day, time spent on innovation 5% → 40%, overall cost reduction ~40%, programs under development +140%.

**Continuous Integration (CI) contract:**
1. The build must never be broken for more than a few minutes
2. If someone breaks the build, the team stops all new work until it is fixed
3. A broken build is treated with the urgency of a production incident
4. Everyone is responsible for keeping the build green — it is never "the build team's problem"

---

## 5. The Deployment Pipeline

The deployment pipeline is the automated path from commit to production. It is code, stored in version control, versioned, and reproducible.

**Pipeline stages (in order of execution speed):**

```
Commit stage       — compile, unit tests, static analysis          < 10 min
Acceptance stage   — acceptance tests, integration tests, security  10–30 min
Performance stage  — performance/load tests                         on demand
Production         — deploy to production (automated or gated)
```

**Commit stage rules:**
- Must complete in under 10 minutes — longer destroys the feedback loop
- Fast feedback is worth imperfect coverage at this stage
- Failure here blocks the entire pipeline and must be fixed immediately
- Build artefact produced here is the artefact deployed everywhere downstream

**Acceptance stage rules:**
- Tests the application as a whole, not units in isolation
- Uses the same artefact produced by the commit stage
- Environment is provisioned from code, identical to production
- Failure here prevents deployment to production

**Pipeline as code:**
- Pipeline definitions live in the same repository as the application
- Pipeline changes are reviewed, tested, and versioned like application code
- Infrastructure for the pipeline is itself reproducible from code

**What the pipeline gates:**
- No human deploys to production without the pipeline having run green
- The pipeline is the only path to production — "friends don't let friends deploy from laptops"

---

## 6. Automated Testing Hierarchy

Tests are organized as a pyramid: many fast, cheap tests at the base; fewer slow, expensive tests at the top. Run from bottom to top; fail fast.

**The test pyramid:**
```
        [Manual/exploratory]       ← few, for discovery
      [Integration tests]          ← fewer, real external services
    [Acceptance tests]             ← moderate, full-stack, application-level
  [Unit tests]                     ← many, fast, isolated, no I/O
```

**Unit tests:**
- No database, no network, no filesystem — mock all external dependencies at the boundary
- Complete in milliseconds; the full suite completes in seconds
- 80%+ class coverage enforced by pipeline (fail if it drops below threshold)
- When an acceptance or integration test fails, write a unit test that catches the same regression before fixing it — shift the detection left

**Acceptance tests:**
- Test the application from the outside in, through its public interface
- Run against a production-like environment provisioned from code
- Cover the highest-value user journeys and all critical failure scenarios
- Non-deterministic / flaky tests are fixed or deleted immediately — a test that fails randomly is worse than no test

**Test-Driven Development (TDD):**
1. Write a failing test for the desired behavior
2. Write the minimum production code to make it pass
3. Refactor for clarity and simplicity

TDD result (Microsoft Research): 60–90% fewer defects vs. non-TDD, at 15–35% longer initial development time. The defect reduction pays back the time cost within one release cycle.

**Non-functional requirement tests in the pipeline:**
- Performance: fail the pipeline if response time regresses > 2% vs. baseline
- Security: static analysis (SAST), dependency vulnerability scanning, at commit stage
- Configuration: lint and validate all infrastructure-as-code before applying
- Run under production-like load using production-like data volumes

**Exploratory testing:** Reserved for high-value activities that cannot be automated — discovering unknown unknowns, testing usability, investigating hypotheses. Not a substitute for automated regression coverage.

---

## 7. Infrastructure as Code and Environment Consistency

Environments that differ from production are a lying mirror. The only valid claim is "it works in an environment configured identically to production."

**Rules:**
- Every environment is created from code, on demand, in minutes
- Developers can create production-like environments on their workstations without opening a ticket
- Staging, QA, and production are configured by the same code; only parameters (endpoints, credentials, scale) differ
- No manual changes to any production environment — if a change cannot be expressed as code in version control, it does not happen
- Treat servers as *cattle, not pets*: when a server diverges from its desired state, replace it; do not SSH in and fix it

**Configuration management tools:** Ansible, Chef, Puppet, Salt for mutable infrastructure; Terraform/Pulumi + immutable images for cloud-native infrastructure; Kubernetes + Helm for container workloads.

**Immutable infrastructure:** Deploy new instances from a golden image; never modify running instances. When Netflix needs to update its fleet, it bakes a new AMI and replaces instances; it does not patch in place. Netflix average instance age: 24 days; 60% of instances less than one week old.

**"Definition of done" includes environment:**
- Code is not done until it has been tested in a production-like environment under production-like load
- Sprint reviews demonstrate features in a production-like environment, not on a laptop

**Database migrations:**
- Schema changes are code: versioned, tested, applied automatically by the pipeline
- Backward-compatible migrations run before deploying new code; cleanup migrations run after old code is retired (expand-contract pattern)
- Never apply schema migrations by hand

---

## 8. Production Telemetry and Observability

You cannot improve what you cannot measure. Telemetry is not an afterthought — it is a first-class feature of every service.

**The four signal layers (instrument all four):**
1. **Business metrics:** transactions, signups, conversion, revenue, churn — the signals that define whether the product is working
2. **Application metrics:** request rates, error rates, latency (p50/p95/p99), queue depths, cache hit rates
3. **Infrastructure metrics:** CPU, memory, disk I/O, network throughput, GC pause time
4. **Deployment pipeline metrics:** build duration, test pass rate, deploy frequency, change failure rate, MTTR

**Etsy principle:** Every feature important enough to implement is important enough to instrument. Adding a metric should require one line of code (StatsD, Prometheus client, OpenTelemetry). If it requires a ticket to Ops, the barrier is too high.

**Logging rules:**
- Use structured logging (JSON) — free-text logs cannot be queried at scale
- Log at appropriate levels: DEBUG (disabled in production), INFO (user and system actions), WARN (degraded state), ERROR (failed operation), FATAL (unrecoverable)
- Always log: authentication/authorization decisions, data mutations, privileged operations, resource exhaustion, circuit breaker trips, service startups/shutdowns
- Never log: passwords, secrets, PII in cleartext, full request/response bodies by default

**Alerting rules:**
- Alert only on conditions that require human action — alert fatigue leads to ignored alerts
- For Gaussian-distributed metrics: alert at 3σ from the mean (99.7% confidence)
- For non-Gaussian data: use moving averages, CUSUM, or anomaly detection — not static thresholds
- Stale metrics (stopped reporting) must alert as aggressively as bad values
- Every alert must have a clear owner and a documented response

**Distributed tracing:**
- Propagate a trace ID through every service call and log it
- Enables reconstruction of full request path across microservices
- Essential for diagnosing latency regressions in service graphs

**Dashboards:**
- Make telemetry an *information radiator* — visible to everyone, not locked behind access requests
- Overlay deployment events on metric graphs (vertical lines at deploy time)
- Self-service: engineers create their own dashboards without tickets

---

## 9. Safe Deployments and Rapid Recovery

Deployments should be boring, routine events — not high-risk operations requiring war-room ceremonies.

**What makes deployments safe:**
- Small, frequent deployments — lower risk per deployment, faster rollback surface
- Production telemetry monitoring actively during and after every deployment
- Automated health checks blocking traffic until new instances pass
- One-click (or automated) rollback capability

**Deployment strategies:**
- **Blue-green deployment:** Two identical production environments; switch traffic between them. Zero-downtime deployment; instant rollback by switching back.
- **Canary release:** Route a small percentage of traffic to the new version; observe; gradually increase. Roll back by routing 100% back to old version.
- **Dark launch / feature flags:** Deploy code to all users; activate feature for a subset via config. Decouple deployment from release. Instant rollback without redeployment.

**When things go wrong — in order of preference:**
1. **Feature flag off:** Disable the feature via config; no new deployment needed; seconds to execute
2. **Fix forward:** If the automated test suite is comprehensive and the fix is safe, deploy a fix immediately
3. **Rollback:** Redeploy the previous known-good artefact; applicable when fixing forward is risky

**"Stop the line" (Andon cord):**
- When the deployment pipeline turns red, no new work enters the pipeline
- The entire team swarms the breakage until it is fixed
- This is not optional — passing a broken pipeline to production is forbidden
- The person who broke the build is not blamed; they are the first resource on the fix

---

## 10. Feedback Loops and On-Call Culture

The people who write the code should experience the consequences of that code in production. Separating development from operational pain removes the feedback that drives quality.

**Developers on-call:**
- Developers participate in the on-call rotation alongside Operations engineers
- Engineers are woken at 2 a.m. for the problems created by their code
- A service is not "done" until it is running in production without excessive escalations
- Result: Developers design for operability — logging, health checks, graceful degradation — because they bear the operational cost

**Shared goals:**
- Dev and Ops share the same definition of "done": running reliably in production
- Dev and Ops share the same metric dashboard
- Dev and Ops share the same post-mortem process
- Ops writes user stories for operational work alongside development work — making Ops work visible in the backlog

**Feedback from operations to architecture:**
- On-call engineers have the authority to file "stop ship" bugs that block releases
- Mean time to recovery (MTTR) is a metric owned jointly by Dev and Ops
- Production incidents trigger post-mortems that create countermeasures in the next sprint

**High-performer MTTR benchmark:** High-performing organizations recover from production incidents 168× faster than low performers (median: minutes vs. days).

---

## 11. Blameless Post-Mortems and Just Culture

Complex systems fail. The question is not whether failure will occur but whether the organization learns from it or repeats it.

**Core principle:** "Human error is not the cause of system failures — it is a consequence of how the system was designed." Blame prevents learning; blame-free analysis enables it.

**Post-mortem process:**

*Timing:* Conduct the post-mortem as soon as possible after resolution — within 24–48 hours. Memory decays; chat logs and metrics do not.

*Participants:* Everyone who was involved in decisions, detection, diagnosis, or response. Include affected downstream teams.

*Content:*
1. Timeline — factual reconstruction using chat logs, metrics, deployment events
2. Contributing factors — what made the system vulnerable; what made detection slow
3. What went well — practices that limited impact
4. Countermeasures — specific, assigned, time-bounded actions that address root causes

*Rules:*
- Disallow counterfactual language: "should have," "would have," "could have"
- No "be more careful" countermeasures — if the fix is "pay attention," the system is not safe enough
- Every countermeasure has a named owner and a due date
- Incident is not closed until post-mortem is scheduled

*Distribution:* Publish post-mortems in a centralized, searchable location accessible to the entire organization. Engineers facing similar problems should be able to search for prior solutions.

**Decrease incident tolerance as reliability improves:**
As major incidents become rare, lower the threshold for what triggers a post-mortem. Track near-misses. Weak signals of latent problems are more valuable than learning only from catastrophes.

**Redefine failure:**
- High-performing organizations deploy 30× more frequently and have half the change failure rate
- More deployments = more learning opportunities — this is a feature, not a bug
- An engineer who caused two major incidents while enabling daily safe deployments is more valuable than one who caused zero incidents by deploying nothing

---

## 12. Security and Compliance as Code

Security controls that exist only at the end of the development lifecycle find problems too late and create bottlenecks. Integrate security into daily work at every stage.

**"Shift left" security practices:**

*At commit:*
- Static analysis security testing (SAST) runs in the commit stage — fail the build on critical findings
- Dependency vulnerability scanning (OWASP dependency-check, Snyk, Dependabot) — block on known exploits
- Secret detection — fail the build if credentials, API keys, or private keys appear in committed code
- Lint infrastructure-as-code with security rules (e.g., open security groups, public S3 buckets)

*In the pipeline:*
- Dynamic analysis (DAST) against a deployed instance in the acceptance environment
- Container image scanning before pushing to the registry
- Compliance policy checks (e.g., Open Policy Agent) as pipeline gates

*In production:*
- Runtime application self-protection (RASP) and WAF for boundary defense
- Enhanced telemetry for security events: authentication failures, privilege escalation, unusual access patterns
- Immutable infrastructure reduces the attack surface (no SSH, no manual changes)

**Change management without the bottleneck:**
- Low-risk standard changes (green pipeline, production-like test environment, feature flag, automated rollback) are pre-approved by policy — no human approval required per deployment
- High-risk changes require peer review and a change record, but the record is created by the pipeline automatically
- Separation of duties is achievable through code review + pipeline enforcement without manual CAB gates

**Compliance as code:**
- Encode compliance controls as automated tests and pipeline gates — not Word documents
- Automated evidence collection: the pipeline records what was deployed, when, by whom, after which tests passed
- Audit logs are immutable and centralized — not reconstructed manually before audits
- "Compliance is directly proportional to the degree to which policies are expressed as code"

---

## 13. Organizational Design and Conway's Law

"Organizations which design systems are constrained to produce designs which are copies of the communication structures of those organizations." — Conway's Law

**Implication:** If you want a loosely-coupled, independently-deployable microservices architecture, you must create loosely-coupled, independently-operable teams. Organizational structure and system architecture must be co-designed.

**Market-oriented teams:**
- Small teams (2-pizza rule: ≤ 10 people) that can independently design, build, test, and deploy their service end-to-end
- Each team owns the full lifecycle of its service — development, operations, and on-call
- Teams do not require coordination with other teams to deploy
- "You build it, you run it" — Werner Vogels, Amazon CTO

**Functional-oriented teams create bottlenecks:**
- A shared QA team is a handoff; a shared Ops team is a queue; a shared security review is a gate
- Each creates lead time variability and reduces flow
- Shared services (platform teams, SRE teams) reduce bottlenecks by building self-service capabilities that product teams consume, not by doing work for them

**Strangler fig pattern for legacy systems:**
- Never attempt a big-bang rewrite — incrementally replace functionality while keeping the old system running
- Introduce an abstraction layer (API gateway, adapter) in front of the legacy system
- Route new traffic to new implementations; leave legacy traffic unchanged
- Retire the legacy implementation once all traffic has migrated

**Team topologies for DevOps:**
- *Stream-aligned teams:* Own a value stream end-to-end; primary team type
- *Platform teams:* Build internal developer platforms (IDP) that stream-aligned teams self-serve; do not do Ops work for other teams
- *Enabling teams:* Temporary coaching teams that uplift stream-aligned team capability; dissolve or pivot once capability is embedded
- *Complex subsystem teams:* Own particularly complex technical domains (ML, cryptography, storage engines)

---

## 14. Continual Learning and Improvement

High-performing organizations improve their daily work as a matter of discipline, not as a consequence of having spare capacity.

**20% time for non-functional work:**
- Reserve 20% of every engineering cycle for reducing technical debt, improving automation, and addressing non-functional requirements
- If technical debt is severe, increase to 30%
- Without explicit reservation, technical debt compounds until all cycles are consumed by it

**Improvement blitz (kaizen blitz):**
- A concentrated period (days to a week) dedicated entirely to improvement
- No feature work during this period
- Teams self-organize around the problems they care most about
- Span the full value stream — Dev, Ops, and Security working together
- Demonstrate outcomes at the end; make wins visible

**Game days and chaos engineering:**
- Schedule a future catastrophic failure event; give teams time to prepare
- Execute without warning (cut power, terminate instances, drop a database)
- Discover latent defects: single points of failure, missing monitoring, cascading failures
- Increase intensity progressively
- Netflix Chaos Monkey: Randomly terminates production EC2 instances during business hours, forcing engineers to build services that survive node failures by design

**ChatOps and shared context:**
- All significant operations (deploys, incident response, configuration changes) executed through a shared chat room
- Benefits: New engineers learn how work is done by reading history; people ask for help publicly; institutional knowledge accumulates in searchable logs; remote teams share a "water cooler"
- Tools like Hubot integrate with Puppet, Jenkins, Capistrano — turning chat messages into operations

**Single shared source repository:**
- All teams share one source repository (Google: 1 billion files, 25,000 engineers, 2 billion LOC in 2015)
- Enables: write a tool once and use it everywhere; accurate dependency tracking; single version of each library in production
- Library owners are responsible for migrating all callers when APIs change — not callers doing so independently

**Technology standardization:**
- Identify the technologies that create the most operational toil and unplanned work
- Standardize on a small set of well-understood platforms (languages, databases, frameworks)
- Etsy: Consolidated from a sprawl of languages and databases to PHP/MySQL as the operational standard; every engineer — Dev and Ops — can read and modify the full stack
- "Buoys, not boundaries": Mark safe/supported territories clearly but allow exploration beyond them for innovation

**Teaching and internal conferences:**
- Allocate dedicated time for engineers to teach colleagues (e.g., 2 hours/week)
- Run internal technology conferences (Capital One: 13 tracks, 52 sessions, 1,200 attendees)
- Encourage external conference attendance and speaking — knowledge flows in both directions
- Pair programming, code reviews, and architecture reviews transfer tacit knowledge that documentation cannot

---

## 15. Anti-Patterns

These patterns are symptoms of low-performing organizations. When you observe them, identify which Way they violate and prescribe the countermeasure.

| Anti-pattern | Consequence | Countermeasure |
|---|---|---|
| Months-long deployment lead time | Competitive unresponsiveness, high-risk "big bang" releases | Decompose value stream; automate pipeline; trunk-based dev |
| Long-lived feature branches | Merge hell, delayed integration feedback, refactoring paralysis | Daily commits to trunk; feature flags for incomplete work |
| Manual environment provisioning | Configuration drift, "works on my machine," weeks of delay | Infrastructure-as-code; on-demand environment creation |
| Separate QA phase at project end | Problems discovered too late, firefighting before release | Automated testing pyramid in pipeline; TDD; shift left |
| Configuration drift (snowflake servers) | Non-reproducible environments, unpredictable failures | Immutable infrastructure; treat servers as cattle |
| Deployments requiring war-room ceremonies | Fear of change, infrequent releases, high blast radius | Blue-green/canary deployments; feature flags; automated rollback |
| Blame culture after incidents | Hidden problems, underreported near-misses, repeated failures | Blameless post-mortems; countermeasures in next sprint |
| Alert fatigue | Ignored pages, missed real incidents | Alert only on actionable conditions; 3σ thresholds; on-call rotation reviews |
| No production telemetry from application code | Blind to failures; MTTR in days | StatsD/Prometheus in every service; structured logging |
| Security review only at project end | Expensive late fixes; compliance bottlenecks | SAST/DAST in pipeline; secrets scanning; IaC policy gates |
| Ops team as ticket queue | Weeks of lead time for environment or config changes | Self-service platform; Ops writes automation, not tickets |
| Technical debt consuming all capacity | No time for features; reliability collapse | Explicit 20% reservation for non-functional work |
| Separate tools/practices for Dev vs. Ops | Invisible Ops work; different definitions of "done" | Shared backlog; Ops user stories; shared on-call rotation |
| Multiple library versions in production | Unpredictable behavior; security exposure | Monorepo or shared artefact registry; library owners drive migrations |
| Documentation-only standards | Standards unenforced; compliance theater | Encode standards as automated pipeline gates and tests |
| Big-bang rewrites of legacy systems | Multi-year risk; high failure rate | Strangler fig pattern; incremental replacement with abstraction layers |
| Optimization for local efficiency (team velocity) | Global throughput suffers; handoff queues build | Optimize for value stream lead time; limit WIP at stream level |

---

## Key Benchmarks

Reference these when evaluating delivery performance:

| Metric | High performer | Low performer |
|---|---|---|
| Deployment frequency | On demand (multiple/day) | Monthly or less |
| Lead time for changes | Minutes to hours | Weeks to months |
| MTTR (mean time to restore) | Minutes | Days |
| Change failure rate | 0–15% | 31–45% |
| Commit-stage build duration | < 10 minutes | Hours |
| Test coverage gate | ≥ 80% | Not enforced |
| Instance age (cloud) | Days (Netflix: avg 24 days) | Months to years |
| Production metric count (Etsy, 2014) | 800,000+ metrics | Few hundred |
| Alerting on non-actionable conditions | Rare | Constant |
