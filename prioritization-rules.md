# Prioritization Rules for Code Review, Codebase Review, and Code Generation

Derived from Alex Ewerlöf's pmresearcher Substack series. All conviction arguments draw
directly from source text; direct quotes are in quotation marks. Claims marked [synthesized]
are reasonable inferences from source narrative rather than verbatim content.

Rules are weighted by source tier:
**Tier 1** (core frameworks) > **Tier 2** (quantified inputs) > **Tier 3** (anti-patterns) > **Tier 4** (governance).

---

## How to use this document

**In a code review:** run P1 first to classify maturity stage, then check all R rules
against the diff. Quote the conviction argument when raising a finding.

**In a codebase review:** open with the C1 diagnosis. Recommendations without a stage
classification will be treated as a shopping list and the foundational items will be skipped.

**In code generation:** apply G1 (stage-appropriate defaults) before writing anything.
When the prerequisite for the requested step is absent, generate the prerequisite instead.

**When challenged:** conviction arguments are written to be quoted directly. Name the source
document. Source documents can be shared with the audience if the argument needs to go
further.

**When rules conflict:** see the Conflict Resolution section at the end of Part 1.

---

## Part 1 — Universal Rules (apply in all three contexts)

### Rule P1: Classify maturity stage before recommending anything

**The rule:** Before making any prioritization recommendation, identify the current maturity
stage of the *environment* (the company and its systems), not the *people*. Point-a-d.md
uses A/B/C/D as conceptual levels. The stage characteristics and observable signals below
are synthesized from the article's narrative — the article provides the framework; the
specifics of each stage are inferred from context.

| Stage | What the source says | Observable signals in a codebase |
|---|---|---|
| **A** | The company "doesn't know how to get from A to B"; broken ownership; heroic culture; "poor understanding of software products" | No alerting configuration, no runbooks, error handling absent in non-trivial paths, no integration tests, ownership undefined in incident |
| **B** | The stable foundation that must exist before stage C work is meaningful [synthesized] | Basic alerting present, deployment doesn't regularly break production, error paths return structured errors, on-call rotation exists |
| **C** | Where "people at C have never been part of taking a company from A to B — they showed up when everything was ready"; the environment people are "spoiled" by | SLOs defined, load-tested, typed failure classification, runbooks for known failure modes, circuit breakers anchored to measurements |
| **D** | Implied endpoint: the fully optimized, differentiated state [synthesized] | Error budget culture, self-healing automation, cost dashboards per service, redundancy patterns proven in production |

**Conviction argument:** *"Point-a-d.md: 'Focus on a few problems and conclude them instead
of attacking multiple problems at the same time. Don't act like a child in the candy shop.
The problems may look small and obvious, but there are deep roots in the organization. It's
a marathon, not a sprint.' The maturity level describes the environment, not the engineers:
'It's a friction problem not an IQ problem.' Stage C people showing up in a stage A
environment with stage C solutions delay the foundational work — they 'showed up when
everything was ready' and may not know how to build the foundation that was already there
when they arrived."*

**Key corollary:** Poor tech health is a symptom. Source-identified root causes include:
lack of budget, feature-factory pressure, broken ownership, poor organizational architecture,
old-school mindset, management failing to act as role model. Recommending technical
sophistication without addressing root causes treats the symptom.

**Anti-pattern:** Treating every system as stage C because that is where most published
best-practices are written for.

---

### Rule P2: Apply the 3 T's test to every prioritization decision

**The rule:** Any proposed work must satisfy all three conditions. A decision is premature
if *any one* of the three T's is wrong. Apply the FAB check within each T.

**FAB check** (from tech-bet.md and premature-optimization.md): Before accepting any claim
as justification, classify it: *Fact* (measured, observed), *Assumption* (inferred,
plausible), or *Belief* (opinion, interpretation). "Most of what flies around as facts is
just someone's assumption or a perspective of the data that supports their belief."

**Thing** — Is this the right problem?
- Do we have data that supports it, or is the urgency a belief?
- Are we optimizing the right variable, or just the most visible one?
- Sub-questions from premature-optimization.md: "Are we optimizing the right thing? Do we
  have data that supports it? Have we done the due diligence to avoid data bias?"

**Time** — Is now the right time?
- Has prerequisite work been completed? (See Rule P3.)
- "Is this the right time to invest in optimization? Is it too soon/late?"
- "What's the time span till the investment pays off?" If unknown, that is a bet (Rule R5).

**Trade-offs** — Are the costs accepted?
- "Engineering is the art of trade-offs. Are we aware of the metrics that will be hurt by
  the optimization?"
- "What is the list of things that will surprise us when the optimization is done?"
- Is there a contingency plan if the trade-offs are larger than expected?

**Conviction argument:** *"Premature-optimization.md: 'Good optimization improves the right
things, at the right time, and with reasonable trade-offs.' Getting any one wrong makes the
entire investment premature. A technically correct reliability improvement that ships before
the feature is stable wastes the investment: the measurement baseline doesn't exist, the
feature may change, and team attention is split across two competing concerns."*

---

### Rule P3: Functional correctness before non-functional improvement

**The rule:** Non-functional requirements (reliability, performance, observability, security
hardening, scalability) must not be prioritized above a feature that does not yet work
correctly for its primary use case.

> *"First make it work, then make it better."* — nfr.md

**In code review:** Reject NFR additions to code paths with known functional bugs or
missing test coverage of the happy path.

**In code generation:** Do not add retry logic, circuit breakers, connection pool tuning,
or distributed tracing to code that is not yet correct for the simple case.

**Why NFRs go wrong early:** Estimating NFR scope is hard. Nfr.md names the failure mode
explicitly: "it's tempting to overestimate its usage in the future and design something
capable of handling 100x or even more load or usage. This is a clear case of tech bet."
NFRs also have slow feedback loops — their value is hard to see short-term, which makes
them easy to gold-plate without accountability.

**Conviction argument:** *"Nfr.md: 'functional requirements are more important than
non-functional requirements.' Reliability scaffolding built around incorrect behavior embeds
the bug. When the underlying behavior is later corrected, the scaffolding must be unwound or
worked around, often at higher cost than if the prerequisite had been delivered first."*

---

### Rule P4: Never own reliability of what the team cannot control — and take control of what the team is responsible for

**The rule has two directions:**

**Direction 1 — scope down:** When a team is held accountable for a metric that includes
variables it cannot influence (third-party API errors, shared platform outages, network
latency), redirect effort toward limiting exposure: timeouts, fallbacks, SLI exclusions for
those events. Do not invest in improving what cannot be fixed.

**Direction 2 — scope up:** When a team is responsible for an outcome but does not control
all the variables that produce it, the correct long-term response is to fix the ownership
boundary, not to live with the misalignment indefinitely. Options from responsible-for-
control.md: move the service to the responsible team, merge the functionality, or create a
new implementation that the responsible team owns.

**Conviction argument:** *"Responsible-for-control.md: 'You should never be responsible for
what you don't control. The reverse is also true: you should take control of what you are
responsible for.' A team that owns an SLO covering a dependency it cannot fix will burn
error budget it cannot recover, page on incidents it cannot resolve, and generate
postmortems with no actionable outcomes. Short term: exclude uncontrollable failures from
the SLI denominator. Long term: align control and responsibility by changing the ownership
boundary."*

---

### Rule P5: Treat reliability as a feature competing on the backlog — and make its cost visible to those who control the budget

**The rule:** Reliability investment — SLO definition, alerting, runbook creation,
resilience patterns, failure testing — must be prioritized against feature work using the
same criteria: measurable user impact and opportunity cost. It must not be treated as a
separate category that accumulates silently or gets addressed only after incidents.

**The accountability gap:** The cost of unreliability (support volume, SLA penalties, lost
renewals, engineering incident time) is real but hidden, while the cost of reliability work
is visible and easy to cut. Mapping-reliability-to-accountability.md: "Leadership can afford
not to care because it's not *their* problem. We need to make reliability everyone's
concern, particularly the upper management who formally have control over budgeting, hiring,
deadlines and other factors that directly impact system health."

**How to make the cost visible:**
- Calculate what a 1% increase in failure rate costs in support tickets, customer trust,
  and engineering incident time
- Map SLO breaches to the org chart node that controls the budget to fix them — "show them
  the gauge and let them drive"
- Prefer real-time dashboards over monthly reports: "lagging metrics lead to even more
  lagging actions"

**Warning — Goodhart's Law:** Mapping-reliability-to-accountability.md: "Once a metric
becomes a goal, it stops being a good measure." Do not tie team performance reviews directly
to SLI numbers. Find metrics that, if gamed, still improve outcomes for all parties.

**Conviction argument:** *"Mapping-reliability-to-accountability.md: 'reliability should be
treated as a feature and prioritized against other features in the backlog. The rule of
10x/9 states that reliability has a cost.' Engineering leaders who cannot secure investment
to improve reliability are facing a leadership accountability gap, not an engineering
capability gap. 'You get what you pay for. Tech health has a cost.'"*

---

### Conflict Resolution

**P3 vs P5:** P3 (functional before NFR) and P5 (reliability on the backlog) are not in
conflict if sequenced correctly. P5 says reliability *work* should be on the backlog — it
does not say it should be prioritized above basic correctness. The correct sequence: (1)
functional correctness, (2) consumer-facing SLI that measures correctness, (3) reliability
investment guided by that SLI.

**P1 vs team self-assessment:** When a team claims a higher stage than observable signals
support, apply the C1 diagnosis (observable evidence, not assertion). If the evidence
contradicts the claim, name the specific gap: "The stage C pattern this PR introduces
requires [X prerequisite] to be operable. I don't see [X] in the codebase. Which PR added
it?" This is not an accusation — it is a question about prerequisites.

**R5 vs P3:** When a change is simultaneously new functionality (P3 applies: deliver
working behavior first) and an unvalidated abstraction (R5 applies: label as tech bet),
both apply. Generate the functional behavior and label the abstraction layer as a tech bet
with a validation plan.

---

## Part 2 — Code Review Rules

### Rule R1: Flag complexity that exceeds the current maturity stage

**When to apply:** Any PR that introduces an advanced pattern (distributed tracing, circuit
breakers, multi-level caching, saga orchestration, chaos testing, automated rollbacks,
multi-AZ failover, Kubernetes migration) without evidence that simpler prerequisites exist.

**What to look for:** Advanced reliability pattern present, but no consumer-facing SLI, no
alerting, no runbook, no basic integration test coverage, or no operational history with
this pattern in this team.

**How to raise it:** *"This change adds [X] but the service does not yet have [simpler
prerequisite]. Point-a-d.md: introducing [X] at this stage delays the foundational work
that would make [X] meaningful and measurable. The people who designed [X] may 'have never
been part of taking a company from point A to B' and are assuming prerequisites that don't
exist here. Recommend deferring [X] until [prerequisite] exists."*

---

### Rule R2: Flag reliability code without a measuring SLI — and flag internal-vitals metrics misrepresented as SLIs

**When to apply:**
- **(a)** Any PR that adds retry logic, fallback behavior, timeout tuning, connection pool
  limits, or graceful degradation without a corresponding metric that measures consumer
  outcomes.
- **(b)** Any PR that adds or references metrics measuring internal system vitals (CPU
  usage, memory, internal queue depth, error log count) as the reliability signal, without
  a consumer-facing SLI.

**What to look for:**
- (a) Resilience code present; no counter or histogram distinguishing good vs failed vs
  degraded outcomes at the service boundary.
- (b) Alerts on internal metrics only; no measurement of whether the consumer experienced a
  good or bad event.

**How to raise it (a):** *"This change improves resilience but adds no measurement.
Service-level-adoption-obstacles.md: 'a good SLI is connected to how consumers perceive
reliability. Traditional metrics often measure internal vitals of the system and have little
to no relation with how consumers perceive service reliability.' Reliability investment
without consumer-facing measurement is advocacy, not engineering. Recommend adding the
metric before or alongside this change."*

**How to raise it (b):** *"This metric measures an internal vital, not a consumer outcome.
Service-level-adoption-obstacles.md identifies this as the defining failure mode of
traditional monitoring. 'Vanity metrics are now called Service Levels. Problem solved!'
Recommend replacing or supplementing with an event-based SLI measuring good/valid outcomes
at the service boundary."*

---

### Rule R3: Flag when an SLO target is raised without costing the change

**When to apply:** Any PR or architectural decision that tightens a latency threshold, adds
a retry that implicitly commits to faster recovery, or raises an availability expectation —
whether in config, documentation, or SLA language.

**What each nine actually costs** (enumerated directly from 10x9.md):
- Refactoring or rewriting critical code paths
- Organizational restructuring to reduce handover risk
- Re-architecting for redundancy: fallbacks, failover
- Increasing team bandwidth: headcount, reduced feature pressure
- Slowing shipping cadence: "Change is the number one enemy of reliability"
- On-call rotation: typically requires 5–8 people; in many jurisdictions on-call carries
  statutory pay premiums (~40% extra, per the article's Sweden example)
- Automation of detection and recovery: mandatory at 5-nines — "5-nines allows only 26s
  of downtime in a month. You cannot afford to have a human in the loop"
- Vendor changes, infrastructure upgrades, multi-region deployments
- Education: "bringing leadership on the same page as the developers to speak the same
  language and prioritize reliability as a feature in the backlog"

**How to raise it:** *"Tightening [threshold/target] raises the implied reliability target.
10x9.md: 'For every 9 you add to SLO, you're making the system 10x more reliable but also
10x more expensive.' This is not a rule of thumb — it reflects real costs in refactoring,
on-call compensation, automation, and infrastructure. This change should be accompanied by
explicit acceptance of those costs, or a decision not to proceed."*

---

### Rule R4: Reject cargo-culted patterns — diagnose using VSI

**When to apply:** Any PR that imports a pattern, library, or architectural approach
justified by association ("this is how [large company] does it", "industry standard", "the
SRE book recommends it") without evidence the pattern fits current constraints.

**VSI diagnosis** (from cargo-culting.md). Cargo cults share three elements — identify
which are present:
1. **Value dominance** — the idea is accepted because of *who* advocates it (Google,
   Netflix, ex-big-tech hire), not because of demonstrated fit; halo effect
2. **Shallow understanding** — the WHAT is known but not the HOW or WHY; the nuances that
   make it work in its original context are missing
3. **Imitation** — superficial rituals adopted (dashboards, process ceremonies, tech radar,
   maturity scores) without the underlying cultural or structural changes

**What to look for:** SOLID rewrite in a CRUD-heavy app; Kubernetes for a team with no
problem EC2 couldn't solve; microservices before monolith ownership is stable; DORA
dashboards before anyone agreed what action to take on the data.

**How to raise it:** *"Cargo-culting.md: 'Understanding WHAT (they do) is important but
don't stop there. Dig deeper to understand HOW (it works) and WHY (they do that).' This
pattern was designed for [original context]. The relevant questions are: what problem does
it solve at that scale, does that problem exist here, and what does it cost to operate with
our team size and operational experience? Recommend justifying by fit (best-practice.md),
not by association."*

---

### Rule R5: Distinguish tech debt from tech bet — apply the delay test

**When to apply:** Any PR described as "paying down tech debt" that actually introduces a
new abstraction, replaces a working system with an unvalidated one, or implements a pattern
the team has not used in production before.

**Tech debt vs tech bet** (directly from tech-bet.md):

| | Tech debt | Tech bet |
|---|---|---|
| Payment direction | Cost paid now for past corner-cutting | Cost paid now, hoping for future return |
| Justification | "We cut corners to ship" | "This makes us future-proof" |
| Symptom | Capabilities lag product requirements | Product requirements lag the tech |
| Root cause | Ambitious product, poor engineering | Ambitious engineering, poor product management |
| Mitigation | Start a payment plan | Start simplification and remove cruft |

**The delay test** (from tech-bet.md): "If someone is pushing for technical investment, ask
them how long this can be postponed and what are the costs to doing it later. If the problem
is too far into the future, it's most probably a bet."

**The production test** (from tech-bet.md): "A greenfield team should have a demo in less
than a month to show that they have a good grasp on the problem. They should have something
in production in less than 3 months. Anything that takes longer than 3 months is too high
of a bet. Break the problem into smaller deliverable chunks."

**How to raise it:** *"This change is a tech bet (tech-bet.md), not tech debt repayment.
'Tech debt is reactive, whereas tech bet is proactive. Tech debt gives the engineers a bad
conscience, whereas tech bet is a sign of naïve ambition or malice.' The distinction matters
for prioritization: tech debt repayment has a clear payback model. A tech bet needs a
validation plan — what will the team measure in 30 days to decide if this paid off, and
what is the simplification plan if it doesn't?"*

---

### Rule R6: Flag changes that add organizational handover risk

**When to apply:** Any PR that adds a new dependency on a service owned by a different
team, introduces a new shared library with no clear ownership, or routes a critical path
through a team boundary without a defined SLA.

**How to raise it:** *"Organization-architecture.md: 'Every team dependency increases the
risk of misunderstanding. Every point of handover opens a miscommunication crack for things
to fall into. Every unclear or broken ownership is a vulnerability for the blame game. Every
incident starts a novel maze to figure out which team is behind what component.' This change
adds [handover point]. Before merging: who owns [dependency] during an incident? What is
their response SLA? Is that SLA acceptable for our SLO? Ambiguous ownership is the
reliability risk, not the code."*

---

### Rule R7: Tech debt must be paired with a repayment commitment at point of creation

**When to apply:** Any PR that deliberately cuts a corner, skips a test, uses a known
suboptimal implementation, or introduces a TODO with no associated tracking.

**The pairing rule** (from tech-debt-day.md): "The PR that creates tech debt should come
paired with the issue to deal with it." The first rule from that team's policy: "The first
rule is not to create tech debt in the first place." When it is created deliberately, it
must be labeled and tracked immediately — not left to accumulate.

**Signals that tech debt is accumulating without payment plan** (from tech-debt-day.md's
software rot symptoms):
- Increasing time to implement features of similar scope (Lead Time growing)
- Longer time to fix defects (TTR growing)
- Longer onboarding time for new contributors (TTFC growing)
- More frequent incidents (MTBF decreasing)

**How to raise it:** *"Tech-debt-day.md: 'The PR that creates tech debt should come paired
with the issue to deal with it.' This change introduces [specific shortcut] without a linked
debt-payback issue. Deliberate debt is acceptable when it's conscious, labeled, and
trackable. Unlabeled debt compounds — 'the more tech debt there is, the less care is given
to new development' (broken window theory). Please add a linked issue before merging."*

---

## Part 3 — Codebase Review Rules

### Rule C1: Produce a maturity stage diagnosis before listing recommendations

**The rule:** Begin every codebase review with a stage classification (A/B/C/D per Rule P1,
using observable signals, not assertions). All subsequent recommendations must be
appropriate to that stage and sequenced — the next stage only, not stages beyond it.

**Structure of diagnosis:**
1. What stage is this, and what observable evidence (alerting presence, SLI existence,
   ownership clarity, incident history, test coverage) supports that classification?
2. What is the one highest-leverage gap preventing advancement to the next stage?
3. What work is currently in progress that is inappropriate for this stage and should be
   deferred?

**Conviction argument:** *"Point-a-d.md: 'Focus on a few problems and conclude them instead
of attacking multiple problems at the same time.' A codebase review that lists 20
recommendations without stage context will be treated as a shopping list, with the most
interesting items picked and the foundational items skipped. The diagnosis produces a ranked
sequence, not a backlog dump."*

---

### Rule C2: Treat missing consumer-facing SLIs as the highest-priority reliability gap

**The rule:** If the codebase has no production SLIs measured at the service boundary —
no event counters distinguishing good from bad consumer outcomes — classify this as a Stage
A→B gap that blocks all other reliability prioritization.

**What is not a consumer-facing SLI:**
- CPU/memory/disk alerts — internal vitals
- Deployment frequency or PR merge rate — process metrics
- Error log count — may double-count or miss client-side failures
- Dashboard presence — does not equal measurement

**What is a consumer-facing SLI:** A ratio of good events to valid events, measured at the
boundary of the service, defined in terms the consumer would recognize as "working" or
"not working."

**Conviction argument:** *"Service-level-adoption-obstacles.md: 'Service levels give a
direction to optimization efforts. Premature optimization is when the wrong thing is
optimized, or at the wrong time. Service levels provide leverage to motivate optimization
and data to prevent premature optimization.' Without consumer-facing SLIs, every reliability
recommendation is a guess. The measurement infrastructure is the precondition; everything
else is downstream of it."*

---

### Rule C3: Identify control boundary violations as structural reliability risk

**The rule:** Map every reliability-critical code path to its dependency tree. Flag any
path where the owning team cannot intervene when a dependency fails — no timeout, no
fallback, no feature flag, no SLI exclusion for that dependency.

**Control boundary violations appear as:**
- HTTP clients with no timeout configured
- SLIs that include error rates from dependencies the team did not build and cannot deploy
- No circuit breaker or fallback for a dependency with a known reliability history
- Retry logic retrying against a dependency with no escalation path

**Long-term resolution:** Responsible-for-control.md gives concrete options where a service
has a single consumer and ownership is misaligned due to Conway's law: move ownership to
the responsible team, create a new in-house service, or merge the functionality — bringing
control and responsibility into alignment.

**Conviction argument:** *"Responsible-for-control.md: 'You should never be responsible for
what you don't control.' Control boundary violations are not minor code smells — they are
structural bets that reliability will be externally determined. A team that pages on
incidents it cannot resolve generates postmortems with no actionable outcomes."*

---

### Rule C4: Identify maturity-model compliance work and derank it — frame maturity as tech, not people

**The rule:** When a codebase shows evidence of work driven by a scoring system (internal
maturity grades, compliance checklists, audit requirements unconnected to user impact),
flag the pattern and recommend reanchoring to consumer-facing SLIs.

**Why grading systems fail** (from engineering-grading-systems.md):
- **One-size-fits-all:** GOOD cannot be defined across a diverse range of services without
  cargo-culting best practices from one context into another
- **Hands-off:** grading systems come from architects detached from the work ("Ivory Tower
  Architect")
- **Presumptuous:** assumes engineers don't know what good looks like — they do; they lack
  budget, ownership, and time
- **Distract accountability:** frames quality as engineer failure while ignoring that "tech
  health has a cost" and leadership controls the investment

**The correct framing:** "If you must create a maturity model, make damn sure to frame it
around the tech, not people. Don't make it personal. People lift the tech only when
technical maturity is detached from their own identity."

**The alternative:** Work with each team individually. Define SLI/SLO per service type.
Give teams full ownership to improve their own service. Measure outcomes, not process scores.

**Conviction argument:** *"Engineering-grading-systems.md: 'It is naïve to think that
emitting a manifesto or chasing people or EMs will lead to meaningful change. I haven't met
a single engineer that I can call immature. The root causes are systemic.' The correct
prioritization anchor is: what does the consumer experience, and is it improving? Not: what
score does the team have on a maturity rubric."*

---

### Rule C5: Flag org seams embedded in the code as reliability risk

**The rule:** Identify architectural patterns that encode organizational boundaries: shared
databases owned by another team, API clients with no error budget agreement, event consumers
with no dead-letter queue ownership defined. These are team handover points embedded in the
architecture.

**Conviction argument:** *"Organization-architecture.md: 'Every point of handover opens a
miscommunication crack for things to fall into. Every unclear or broken ownership is a
vulnerability for the blame game. Every incident starts a novel maze to figure out which
team is behind what component.' When a shared database or event bus crosses a team boundary
without explicit ownership and SLA, incidents at that boundary generate extended MTTR because
no single team has both the context and the authority to fix it."*

---

### Rule C6: When a prioritization initiative spans teams, diagnose the governance structure

**The rule:** When a codebase review reveals a cross-team technical initiative underway
(standardization efforts, platform migrations, observability rollouts, compliance
implementation), check whether the governance structure fits the nature of the work. The
two failure modes are opposite: standing committees that persist past their mandate, and
ad-hoc collaboration that lacks mandate at all.

**Technical Committee failure mode** (from technical-committee-lifecycle.md): Committees
that don't dismantle after their original delivery enter a predictable lifecycle —
Inertia → Expansion → Dominance → Collapse. "To keep the committee alive, it expands scope,
recruits more people, and assimilates their mandate to the hive." It then "starts to demand
important decisions to run through it" — becoming a bottleneck rather than an accelerant.
Identify committees that have entered this phase by their output: manifestos and playbooks
with no connection to teams' day-to-day work.

**Ephemeral Taskforce (ETF)** as the alternative (from ephemeral-taskforce.md): "A selected
group of people with cross-functional knowledge, mandate, and responsibility who are
assembled for a specific delivery with a clear end in mind. The group is dismantled after
the objective is accomplished."

ETF is appropriate when:
- The initiative spans the org and requires more bandwidth than one person
- A clear delivery and end-date can be defined

ETF is *not* appropriate when:
- One person can deliver it (ETF is overkill)
- The work is ongoing without an end-date (reorganization is the better tool)

**How to raise it:** *"Technical-committee-lifecycle.md identifies the failure pattern:
committees that persist past their delivery mandate enter inertia, expand scope, and become
blocking rather than enabling — 'what started as a tool to collaborate is effectively abused
to block other forms of collaboration.' Ephemeral-taskforce.md provides the alternative:
a time-bounded group with formal mandate, the right domain expertise, and a clear
disbandment condition. For [this initiative], the question is: is there a defined delivery
end-point, and is there a plan to disband when it is reached?"*

---

## Part 4 — Code Generation Rules

### Rule G1: Default to stage-appropriate complexity

**The rule:** Infer the maturity stage from context (test coverage, existing observability,
code age, stated constraints). Generate code appropriate to that stage. Do not generate
advanced patterns unless the context shows the foundation exists to operate them.

| Stage | Generate | Withhold until next stage |
|---|---|---|
| A→B | Working implementation with clear error paths and basic test coverage | Retry logic, circuit breakers, caching layers |
| B→C | Consumer-facing SLI metrics, structured logging, basic timeouts | Multi-level fallbacks, adaptive throttling, distributed tracing |
| C→D | Adaptive resilience, cost-optimized patterns, multi-region | Global consistency mechanisms, automated rollback pipelines |

**Conviction argument:** *"Generating a circuit breaker for a service with no SLI embeds a
dependency that will be operated blind. The team cannot tell if the circuit is opening
correctly, cannot set appropriate thresholds, and cannot know if it is causing more harm
than good. Stage-appropriate generation produces code the team can actually operate with
their current capabilities and instrumentation."*

---

### Rule G2: Label generated abstractions as tech bets when unvalidated

**The rule:** When generating a novel abstraction, shared utility, or pattern the team has
not used in production before, label it explicitly:

```
// Tech bet: this abstraction assumes [X].
// Validate by measuring [Y] after [timeframe].
// Simplification plan if [Y] does not improve: [simpler alternative].
```

**The three-month rule** (from tech-bet.md): Any abstraction that cannot ship to production
in its initial form within three months should be broken into smaller deliverable steps.
"Avoid big bang releases like plague. Rather have a broken product now than the promise of a
solid one in the future."

**Conviction argument:** *"Tech-bet.md: 'The defining feature of tech bet is lack of
information or wrong information which leads to inaccurate predictions and costly wrong
decisions.' Generated abstractions that seem obviously correct are frequently tech bets in
disguise — they assume usage patterns, team familiarity, and operational context that may
not exist."*

---

### Rule G3: Generate reliability code in the correct sequence

**The rule:** Respect the ordering. When the calling context does not include step N-1,
generate step N-1 instead of step N.

1. **Consumer-facing metric** measuring the failure being addressed (must exist first)
2. **Alert** on that metric using burn rate, not threshold (must exist before resilience
   code)
3. **Resilience pattern** (retry, timeout, fallback) anchored to the alert thresholds
4. **Automated response** (circuit breaker, adaptive throttling) only after step 3 has been
   in production long enough to calibrate thresholds

**Conviction argument:** *"NFR sequencing (nfr.md) and service-level-adoption-obstacles.md
establish the ordering constraint. A circuit breaker with no measurement is a black box —
it will open at the wrong time, with the wrong threshold, for the wrong reason, and no one
will know. The metric is not overhead on top of the reliability feature; it is the
precondition that makes the reliability feature operable."*

---

### Rule G4: When generating SLO-related code, generate the SLI alongside it

**The rule:** Any generated code that implements a reliability commitment (SLA language,
uptime guarantee, retry-within-N-seconds behavior) must be accompanied by the SLI that
measures whether that commitment is being met. Do not generate the commitment without the
measurement.

**Canonical event-based SLI pattern (Go):**

```go
// SLI definition: good_events / valid_events
// good  = job completed without a system-caused error
// valid = job attempted (excludes caller-cancelled jobs — outside engineering control)
reportJobsTotal.WithLabelValues("attempted").Inc()

// ... perform work ...

if errors.Is(err, context.Canceled) {
    // Caller cancelled before completion: excluded from valid denominator
    // per responsible-for-control.md — the team does not control caller cancellation
    reportJobsTotal.WithLabelValues("cancelled").Inc()
    return err
}
if err != nil {
    reportJobsTotal.WithLabelValues("failed").Inc()
    return err
}
reportJobsTotal.WithLabelValues("succeeded").Inc()
```

**Goodhart's Law guard:** The SLI must measure something the consumer would recognize as the
outcome they care about. If the metric is easy to make green without improving consumer
experience, it is a vanity metric.

**Conviction argument:** *"Service-level-adoption-obstacles.md: 'Service levels provide
leverage to motivate optimization and data to prevent premature optimization.' Generating a
reliability commitment without its measurement produces a promise with no feedback loop. The
SLI is not documentation — it is the mechanism by which the team knows whether the code is
working."*

---

### Rule G5: Apply fit-over-best when selecting patterns

**The rule:** When multiple patterns could solve a problem, evaluate by fit, not prestige.
The fit evaluation questions (from best-practice.md):

- Where does this practice come from, and what constraints made it the right choice there?
- What are the nuances required for it to work? (These are usually omitted in the writeup.)
- What problem does it solve, and does that problem exist at current scale and stage?
- Where does it *not* apply? (Best-practice advocates rarely publish this part.)
- Can this be justified by logical reasoning, or does it require name-dropping to convince?

**The expertise gradient** (directly from best-practice.md): "Junior follows the rules.
Senior makes up the rules. Expert knows when to break the rules." Generate code at the
appropriate level of rule-following for the maturity stage.

**Conviction argument:** *"Best-practice.md: 'Best practices are often someone's
interpretation of why something worked at a certain time and environment. This interpretation
is often exaggerated to make a point. If you follow best practices to the dot, at best you
end up where another company was in the past. At worst, you'll ruin your unique edge.' The
correct standard is: fit for the current team, current scale, and current stage."*

---

## Quick-Reference Anti-Pattern Table

| Symptom observed | Rule | Diagnosis |
|---|---|---|
| Advanced pattern, no basic observability | P1, R1, G1 | Stage mismatch: stage C work in stage A environment |
| Reliability code, no consumer-facing SLI | R2, C2, G3 | Measurement precondition missing |
| Internal-vitals metrics labeled as SLIs | R2, C2 | Vanity metrics: measures system, not consumer experience |
| SLO target raised, no cost discussion | R3 | Implied nine added without accepting its 10× cost |
| "Industry standard" without fit analysis | R4, G5 | VSI pattern: value dominance + shallow understanding + imitation |
| "Tech debt" that introduces new abstraction | R5, G2 | Mislabeled tech bet: no delay-cost estimate, no validation plan |
| NFR work on incorrectly behaving feature | P3, G3 | Sequence violation: correct before improving |
| SLI includes errors from uncontrollable dependency | P4, C3 | Control boundary violation |
| PR creates shortcut with no linked payback issue | R7 | Tech debt without payment commitment |
| Reliability work absent from feature backlog | P5 | Hidden cost asymmetry: unreliability cost invisible to budget holders |
| New team dependency with no SLA | R6, C5 | Org seam without accountability |
| Maturity checklist driving team priorities | C4 | Grade optimization replacing outcome optimization |
| 20 recommendations with no sequencing | C1 | Shopping list without stage diagnosis |
| Same metric used for grading and optimization | P5, C4 | Goodhart's Law: metric became goal, stopped being measure |
| Greenfield with no demo after 1+ months | R5, G2 | Tech bet: big-bang risk, no iterative validation |
| Standing committee blocking technical decision | C6 | TCLC inertia/dominance phase: mandate has outlived delivery |
| Cross-team initiative with no end-date or disbandment plan | C6 | ETF needed, or reorg if ongoing |

---

## Source Documents

All source documents are in `/Users/steve/Documents/agent-orange/substack/pmresearcher/`:

| File | Rules | What it actually contributes |
|---|---|---|
| `20230804_100650_point-a-d.md` | P1, R1, C1 | A/B/C/D conceptual framework; friction vs IQ; "focus on a few problems"; stage C people in stage A environments |
| `20250224_181737_premature-optimization.md` | P2, C1 | 3 T's (Thing/Time/Trade-offs); sub-questions per T; FAB check |
| `20240422_081030_nfr.md` | P3, G3 | "First make it work, then make it better"; NFR overestimation as tech bet |
| `20240220_185447_responsible-for-control.md` | P4, C3 | Both directions: never own what you can't fix; take control of what you're responsible for; options for fixing ownership boundaries |
| `20241206_163121_mapping-reliability-to-accountability.md` | P5 | Reliability as backlog feature; leadership accountability gap; Goodhart's Law; "show them the gauge and let them drive" |
| `20240524_130358_service-level-adoption-obstacles.md` | R2, C2, G4 | Consumer-facing SLI vs internal vitals; "vanity metrics" anti-pattern; SLI as precondition for all optimization |
| `20231201_053017_10x9.md` | R3 | "10x more reliable, 10x more expensive" per nine; enumerated cost categories including on-call, automation, cadence slowdown |
| `20240125_221037_tech-bet.md` | R5, G2 | Debt vs bet distinction; delay test; three-month production rule; FAB framework |
| `20241110_160839_cargo-culting.md` | R4, G5 | VSI framework: value dominance, shallow understanding, imitation; historical origin and corporate examples |
| `20241204_163905_best-practice.md` | R4, G5 | "Best" is absolute; "fit" is context-relative; junior/senior/expert gradient; exaggerated interpretation |
| `20240726_212043_organization-architecture.md` | R6, C5 | Handover seams as reliability risk; Conway's Law; DORA metric degradation at org boundaries |
| `20241204_174705_engineering-grading-systems.md` | C4 | Four failure modes of maturity models; frame maturity as tech not people; SLI/SLO as the correct alternative |
| `20230427_200126_technical-committee-lifecycle.md` | C6 | TCLC seven-stage lifecycle; expansion and dominance failure modes; Porsche quote on committee timidity |
| `20250531_080056_ephemeral-taskforce.md` | C6 | ETF definition; comparison to TC on knowledge/mandate/responsibility/productivity; when ETF is and is not appropriate |
| `20230109_150339_tech-debt-day.md` | R7 | "PR that creates debt paired with payback issue"; software rot symptoms (LT, TTR, MTBF, TTFC); 10% as one team's negotiated outcome; "first rule: don't create debt in the first place" |
