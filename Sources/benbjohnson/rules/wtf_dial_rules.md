# Ruleset: WTF Dial — Introduction

```
Source: "Introducing WTF Dial (again)" — Ben Johnson (gobeyond.dev, Jan 2021)
Scope:  Go; application deployment; development methodology;
        single-process web applications; small-scale infrastructure
```

---

## 1. Deployment Model

The Go static-compilation model is not incidental — it is a design constraint that
shapes decisions from dependency management to release automation.

---

§1.1  **[SHOULD][ARCH]**  Design every Go application to be distributed as a single
      statically-compiled binary with no external runtime, shared-library, or
      interpreter dependency.
      A dependency on a shared library or runtime that must be installed on the
      target system introduces a deployment failure mode that is entirely orthogonal
      to the correctness of the application code; CI passes, the binary is present,
      but the application fails to start because the host is missing a version of
      `libsomething.so`.
      ```
      ✗  Binary that dlopens a .so at startup, requires JVM/Python runtime, or calls
         out to a system-installed interpreter
      ✓  go build -o myapp ./cmd/myapp  — uploads and runs without installation steps
      ```

§1.2  **[CONSIDER][ARCH]**  For applications serving on the order of hundreds of
      users, prefer a single-process deployment on a small VPS over a distributed
      microservices architecture.
      A distributed deployment adds operational concerns — service discovery,
      inter-service latency, distributed tracing, independent failure domains — that
      provide team and scaling benefits only beyond a threshold the application has
      not yet reached; a single Go process with a small memory footprint can serve
      this load on minimal hardware while keeping operational complexity near zero.
      *(Rationale is stated quantitatively in the source ("hundreds of users",
      "$5/month VPS") — applied as a [CONSIDER] because the threshold is
      application-dependent.)*

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Binary requires a system-installed runtime (JVM, Python, Node) | ARCH | §1.1 | Deployment fails when target host lacks the runtime version; CI cannot validate host state | Statically compile to a self-contained binary |
| Shared-library dependency loaded at startup (`cgo` + system .so) | ARCH | §1.1 | Binary runs on build host but fails on clean deploy target | Link statically or eliminate the dependency |
| Microservices deployment for an application with a small user base | ARCH | §1.2 | Operational overhead dwarfs engineering output; distributed-systems failure modes appear before scaling benefits | Single-process deployment until team or load justifies the split |

---

## 2. Development Methodology

---

§2.1  **[SHOULD][METHOD]**  Build the complete working application before writing
      structured explanations of its architecture or publishing incremental design
      breakdowns.
      Documenting or teaching an architecture while simultaneously constructing it
      produces explanations tied to an incomplete and potentially wrong design;
      once a full working system exists, the actual trade-offs, the parts that
      required non-linear iteration, and the patterns that survived contact with
      reality can be described accurately.
      *(Derived from the author's direct account of abandoning an earlier incremental
      series because application development is non-linear — applied as [SHOULD]
      because the failure mode is documentation quality, not application correctness.)*

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Publishing architectural documentation in lockstep with an in-progress build | METHOD | §2.1 | Docs reflect a design that will change; readers follow a path the author later abandoned | Complete the application; document the final, working design |

---

## Master Anti-patterns Table

| Anti-pattern | Level | Rule violated | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Binary requires a system-installed runtime or shared library | ARCH | §1.1 | Deployment fails on clean hosts; CI does not validate host state | Statically compiled self-contained binary |
| Microservices for a small-scale application | ARCH | §1.2 | Distributed-systems operational overhead before any scaling benefit | Single-process deployment at small scale |
| Documenting architecture while the application is still being built | METHOD | §2.1 | Documentation describes an abandoned or incomplete design | Build completely; then document accurately |
