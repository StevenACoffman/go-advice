Sources:          ff/v4 Command Pattern Spec (climax-generated applications, peterbourgon/ff/v4)
                  Cobra Command Pattern Spec (spf13/cobra)
                  Stdlib Command Pattern Spec (Go toolchain idiomatic pattern, flag package)
Scope:            Go CLI applications (Go only). §1–§2 and §3.1, §4.1, §5.1, §5.6 apply to
                  all three patterns. §2.2, §2.3, §3.2–§3.4, §4.2, §5.2, §5.5, and §6.2–§6.4
                  carry per-pattern scope annotations where they diverge.
Synthesised from: 3 rulesets

---

## §1  Command Architecture

§1.1  [MUST][ARCH]   Do not use `init()` to register commands; register in the
      constructor or the dispatcher's explicit list.
      `init()` execution order is implicit and untraced; readers cannot determine
      which commands are registered without scanning every file in the package,
      and the sequence cannot be controlled or tested.
      ✗  `func init() { root.AddCommand(serveCmd) }`
      ✓  ff/v4: `parent.Command.Subcommands = append(...)` inside `New()`
         Cobra: `root.AddCommand(serve.NewCommand())` inside `newRootCommand`
         stdlib: explicit slice in `cmd.Run`

§1.2  [MUST][ARCH]   Do not declare package-level mutable variables that hold
      command instances, shared flag destinations, or shared state initialised at
      program startup.
      Package-level mutable state is shared across all test invocations; parallel
      tests corrupt each other's flag values, causing non-deterministic failures
      that are impossible to reproduce in isolation.
      ✗  `var rootCmd = &cobra.Command{...}` or `var verbose bool` at package level
      ✓  Construct command instances and declare flag-destination variables inside
         the constructor function that accepts I/O writers
      *(Exception: `var Version = "dev"` — see Permitted Exceptions)*

§1.3  [MUST][ARCH]   Do not define a `Commander` interface for command
      polymorphism; use the framework's struct type (or the project's `Command`
      struct) directly with a function-pointer field.
      Every implementation of such an interface would be the same concrete struct
      type; the interface adds no polymorphism. When the framework struct gains a
      new hook or field, every adapter type must be updated in lockstep, producing
      compiler errors scattered across command packages rather than at the interface
      definition, making the source of the breakage invisible.
      ✗  `type Commander interface { Execute(ctx context.Context, args []string) error }`
         `type serveCmd struct{}; func (s *serveCmd) Execute(...) error { ... }`
      ✓  ff/v4: `cfg.Command = &ff.Command{Exec: cfg.exec}`
         Cobra: `cmd.RunE = func(...) error { return doWork() }`
         stdlib: `return &command.Command{Run: serveCmd}`

§1.4  [MUST][ARCH]   Export exactly one symbol from each command package: the
      constructor function.
      Exposing additional symbols — a second constructor, a `Run` function, the
      command struct itself — creates implicit coupling that prevents independent
      evolution; callers may depend on the extra symbol and resist refactoring.
      ✗  Exporting `NewCommand`, a `Run(args []string) error`, and an `Options` type
         from the same command package
      ✓  ff/v4: export `New` (constructor) and `Config` (required for parent
         embedding — this is the sole permitted second export in ff/v4 commands)
         Cobra: export `NewCommand() *cobra.Command`
         stdlib: export `<Name>Command() *command.Command`
      *(Exception: `var Version = "dev"` in the version package — see Permitted Exceptions)*

§1.5  [MUST][ARCH]   Declare `package cmd` as a dispatcher package imported by
      `main`; do not give it a `main` function or call `os.Exit` from it.
      A `package cmd` with its own `main` function cannot be imported without
      executing the program; calling `os.Exit` inside the dispatcher bypasses the
      returned error and prevents `main` from controlling the exit code.
      Separating dispatch from the entry point allows tests to call the dispatcher
      with controlled I/O, observe the returned error, and assert exit behaviour.
      ✗  `package main` in `cmd/cmd.go`, or `os.Exit` inside `Execute`
      ✓  `package cmd` in `cmd/cmd.go`; `main.go` calls `cmd.Execute(...)` and
         handles the returned error and exit code

| Anti-pattern | Level | Rule | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `init()` registration | ARCH | §1.1 | Implicit ordering causes non-deterministic registration; untraceable in large repos | Explicit `AddCommand` / `append` in constructor |
| `var rootCmd = &cobra.Command{...}` | ARCH | §1.2 | Shared state; parallel-test data races; flag values persist across test invocations | Construct command inside constructor accepting I/O writers |
| Package-level flag variables (`var verbose bool`) | ARCH | §1.2 | Values from one test invocation bleed into the next; parallel tests corrupt each other | Declare inside constructor; close over in `RunE`/`exec` |
| Custom `Commander` interface | ARCH | §1.3 | Framework struct changes require lockstep updates to all adapter types; compiler errors appear in command packages, not at the interface | Use framework struct with `Exec`/`RunE`/`Run` function-pointer field |
| Multiple exported symbols per command package | ARCH | §1.4 | Callers couple to internal helpers, preventing refactoring | One exported constructor; all other symbols unexported |
| `os.Exit` inside `cmd` package | ARCH | §1.5 | Dispatcher bypasses returned error; `main` loses exit-code control; test binary terminates mid-run | Return `error`; `main` owns `os.Exit` |

---

## §2  I/O and Signal Handling

§2.1  [MUST][CODE]   Accept `stdin io.Reader`, `stdout io.Writer`, and
      `stderr io.Writer` as constructor parameters; write output only through
      those writers; never read from `os.Stdin` or write to `os.Stdout` /
      `os.Stderr` directly inside a command.
      Hard-coded OS streams make command output uncapturable in tests; every test
      that exercises such a command must redirect at the OS level, serialising
      parallel tests and making output assertions unreliable.
      ✗  `fmt.Fprintln(os.Stdout, result)`
      ✓  ff/v4: `fmt.Fprintln(cfg.Stdout, result)` via embedded `root.Config`
         Cobra: `fmt.Fprintln(cmd.OutOrStdout(), result)`
         stdlib: pass writers into the factory or use a shared struct

§2.2  [MUST][CODE]   Use `signal.NotifyContext` to create the root context and
      call the returned stop function before `os.Exit`.
      `signal.NotifyContext` starts an internal goroutine to forward signals; the
      stdlib documentation requires calling `stop()` to release resources when the
      context is no longer needed. Skipping it leaks the goroutine until process
      exit and suppresses subsequent signal delivery to other listeners.
      ✗  `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)`
         `exitCode := run(ctx, ...); os.Exit(exitCode)` — stop never called; goroutine leaks
      ✓  `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)`
         `exitCode := run(ctx, ...); stop(); os.Exit(exitCode)`

§2.3  [MUST][CODE]   Separate the `run()` helper (returns `int`) from `main()`
      (calls `stop()` then `os.Exit(run(...))`).
      Deferred cleanup functions registered with `defer` do not execute when
      `os.Exit` is called; wrapping all program logic in `run()` ensures deferred
      calls — file closes, temporary directory removal, connection drains — execute
      before the process exits.
      ✗  `func main() { defer cleanup(); ...; os.Exit(1) }` — defer skipped
      ✓  `func run(ctx context.Context, args []string, stdout, stderr io.Writer) int { defer cleanup(); ... }`
         `func main() { ctx, stop := signal.NotifyContext(...); code := run(ctx, os.Args[1:], os.Stdout, os.Stderr); stop(); os.Exit(code) }`

| Anti-pattern | Level | Rule | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `fmt.Fprintln(os.Stdout, ...)` in command | CODE | §2.1 | Output uncapturable in tests; parallel tests must serialise to avoid interleaving | Accept `stdout io.Writer` in constructor; write through it |
| `os.Exit` before `stop()` | CODE | §2.2 | Signal-forwarding goroutine leaks; subsequent signals undelivered | Call `stop()` before `os.Exit` |
| `defer cleanup()` in `main()` with `os.Exit` | CODE | §2.3 | Deferred functions never run; resources leak on every non-zero exit | Move logic to `run()` which returns `int`; `main()` only calls `os.Exit` |

---

## §3  Flag Binding

§3.1  [MUST][CODE]   Bind flag values in the constructor (`New`, `NewCommand`,
      or `<Name>Command`), not inside the execution function (`exec`, `RunE`,
      or `Run`).
      Flag parsing completes before the execution function is called; binding
      inside the execution function is too late — the flag set has already been
      parsed and the variable will always hold its zero value.
      ✗  Cobra: `cmd.RunE = func(cmd *cobra.Command, args []string) error {`
         `    cmd.Flags().IntVar(&timeout, "timeout", 30, "...") // too late; already parsed`
         `    return doWork(timeout)`
         `}`
      ✓  Cobra: `cmd.Flags().IntVar(&timeout, "timeout", 30, "...")`  // in NewCommand, before RunE
         `cmd.RunE = func(cmd *cobra.Command, args []string) error { return doWork(timeout) }`
         ff/v4: `cfg.Flags.IntVar(&cfg.Timeout, 0, "timeout", 30, "...")` in `New()`, before exec

§3.2  [MUST][CODE]   Call `.SetParent(parent.Flags)` on every subcommand flag
      set so that parent flags are accepted at any nesting depth. *(ff/v4 only)*
      Without `SetParent`, a parent flag passed on a subcommand invocation is
      treated as an unrecognised flag; the parse fails with
      `flag provided but not defined: -<flag>`. The user made no mistake and the
      error message gives no hint that a missing `SetParent` call is the cause.
      Flag inheritance silently breaks at the first subcommand that omits the call.
      ✗  `cfg.Flags = ff.NewFlagSet("sub")` — parent flags rejected
      ✓  `cfg.Flags = ff.NewFlagSet("sub").SetParent(parent.Flags)`

§3.3  [MUST][CODE]   Use `flag.NewFlagSet` scoped to the command function; do
      not call `flag.Parse` or bind to `flag.CommandLine` directly. *(stdlib only)*
      The global `flag.Parse()` parses `os.Args[1:]` and cannot be called more
      than once safely; per-command `FlagSet` instances parse their own slice and
      produce independent state for each test invocation.
      ✗  `flag.BoolVar(&noNewline, "n", false, "suppress newline")`
      ✓  `fs := flag.NewFlagSet("shout", flag.ContinueOnError); noNewline := fs.Bool("n", false, "suppress newline")`

§3.4  [MUST][CODE]   Register every behavioral knob as a flag on the command's
      flag set; do not hide configuration behind hard-coded values, package-level
      variables, or `os.Getenv` calls outside the flag set. *(ff/v4 only)*
      Settings outside the flag set are invisible to `-h`; users cannot discover
      them and `ff.WithEnvVarPrefix` cannot automatically map them from environment
      variables. The invariant to preserve: running any command with `-h` reveals
      its complete configuration surface.
      ✗  `timeout := os.Getenv("MYAPP_TIMEOUT")` inside `exec`
      ✓  `cfg.Flags.DurationVar(&cfg.Timeout, 0, "timeout", 30*time.Second, "request timeout")`
         — automatically mapped from `MYAPP_TIMEOUT` via `ff.WithEnvVarPrefix`

| Anti-pattern | Level | Rule | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Binding flags inside `RunE`/`exec` | CODE | §3.1 | Variable always holds zero value; flag has no effect | Bind in constructor before assigning `RunE`/`Exec` |
| `ff.NewFlagSet` without `SetParent` *(ff/v4)* | CODE | §3.2 | Parse fails with "flag provided but not defined"; user made no mistake; cause invisible | `.SetParent(parent.Flags)` on every subcommand `FlagSet` |
| `flag.Parse()` (global) in command *(stdlib)* | CODE | §3.3 | Cannot be called twice; pollutes global state; breaks test isolation | `flag.NewFlagSet` inside the command function |
| `os.Getenv` outside flag set *(ff/v4)* | CODE | §3.4 | Setting invisible to `-h`; not mapped by env-var prefix; not overridable by flag | Declare as a registered flag |

---

## §4  Dispatch

§4.1  [MUST][CODE]   Give every command a unique dispatch key; do not register
      two commands under the same key.
      In slice-based dispatchers (ff/v4, Cobra), the first matching entry wins and
      later duplicates are permanently unreachable. In map-based dispatchers
      (stdlib), the later entry silently overwrites the earlier one. Neither
      produces an error at registration or at startup.
      ✗  Two commands both named `"version"` in the slice / subcommand list
      ✓  Unique `Name` (ff/v4), unique first word of `Use` (Cobra), unique
         `UsageLine` (stdlib) across all registered commands

§4.2  [MUST][CODE]   Dispatch via a map keyed on the dispatch string; do not
      use `switch`/`case` over the command name. After lookup, check the `ok`
      value and return `fmt.Errorf("%s: unknown command", args[0])` when the key
      is absent. *(stdlib only)*
      A `switch`/`case` dispatcher requires modifying the dispatcher file whenever
      a command is added; the change is easy to forget and creates a merge-conflict
      hotspot. An unguarded map lookup silently returns nil and produces a nil
      pointer dereference rather than a useful error message.
      ✗  `switch args[0] { case "greet": greet.Run(args[1:]); case "version": ... }`
      ✗  `cmd := m[args[0]]; cmd.Run(cmd, args[1:])` — panics on unknown command
      ✓  `cmd, ok := m[args[0]]; if !ok { return fmt.Errorf("%s: unknown command", args[0]) }`

| Anti-pattern | Level | Rule | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Duplicate dispatch key | CODE | §4.1 | One entry permanently unreachable (ff/v4/Cobra: shadowed; stdlib: overwritten); no error at registration or startup | Unique `Name` / `Use` / `UsageLine` per command |
| `switch`/`case` dispatch | CODE | §4.2 *(stdlib only)* | Dispatcher file edited on every add; merge-conflict hotspot | Map keyed on `UsageLine` |
| Unguarded map lookup on dispatch | CODE | §4.2 *(stdlib only)* | Nil pointer dereference instead of a useful error on unknown command | Check `ok`; return `fmt.Errorf("%s: unknown command", name)` |

---

## §5  Error Handling

§5.1  [MUST][CODE]   Return `error` from every command execution function; never
      call `os.Exit` inside a command.
      `os.Exit` bypasses all deferred cleanup, prevents `main` from controlling
      the exit code, and terminates the test binary mid-run, making the execution
      path untestable.
      ✗  `if err != nil { os.Exit(1) }` inside `exec` / `RunE` / `Run`
      ✓  `return fmt.Errorf("serve: listen %s: %w", addr, err)`

§5.2  [MUST][CODE]   Return `root.ExitError(n)` when a command must exit
      non-zero without printing an `"error: ..."` line. *(ff/v4 only)*
      Returning any other non-nil error causes the dispatcher to print
      `"error: <message>"`; when the command has already written its own
      diagnostic output, this produces a redundant and confusing second message.
      `root.ExitError` signals failure to the shell while suppressing the
      dispatcher's error line.
      ✗  `return fmt.Errorf("lint: found %d issues", n)` after already printing issue details
      ✓  `return root.ExitError(1)` after writing diagnostics to `cfg.Stderr`

§5.3  [MUST][CODE]   Use `RunE`, not `Run`, on every `cobra.Command`. *(Cobra only)*
      `Run` has signature `func(*Command, []string)` with no return value; errors
      that occur inside it cannot be propagated to the caller, forcing either
      `os.Exit` (violating §5.1) or silently swallowing the failure.
      ✗  `cmd.Run = func(cmd *cobra.Command, args []string) { doWork() }`
      ✓  `cmd.RunE = func(cmd *cobra.Command, args []string) error { return doWork() }`

§5.4  [MUST][CODE]   Set `SilenceErrors: true` and `SilenceUsage: true` on the
      root `cobra.Command`; call `cmd.Usage()` explicitly only before errors that
      are genuine usage mistakes (wrong argument count, invalid flag value).
      *(Cobra only)*
      With the defaults (`false`), Cobra prints the error both before the caller
      and after; the user sees each error twice. `SilenceUsage: false` appends the
      full usage template to every runtime error (network failure, DB timeout),
      adding irrelevant noise. Usage text is only instructive when the user made a
      usage mistake — not a runtime failure.
      ✗  Root `cobra.Command{}` with no Silence fields set; runtime DB error prints
         "Error: connection refused\n<entire usage template>"
      ✓  Root with `SilenceErrors: true, SilenceUsage: true`; `_ = cmd.Usage()` before
         returning an argument-validation error

§5.5  [MUST][CODE]   In the ff/v4 dispatcher, treat `ff.ErrHelp` and
      `ff.ErrNoExec` as normal termination; return exit code 0, print no error
      message. *(ff/v4 only)*
      `ff.ErrHelp` means the user passed `-h`/`--help`; ff already printed the
      help text before returning. `ff.ErrNoExec` means a group command was invoked
      without a subcommand; ff already printed the subcommand list. Printing
      `"error: help requested"` or `"error: ErrNoExec"` is user-hostile and
      incorrect — no error occurred.
      ✗  `if err != nil { fmt.Fprintln(os.Stderr, "error:", err); return 1 }`
      ✓  `if errors.Is(err, ff.ErrHelp) || errors.Is(err, ff.ErrNoExec) { return 0 }`

§5.6  [MUST][CODE]   Format error strings as lowercase with no trailing
      punctuation, in the form `<command>: <reason>`.
      Go errors are typically wrapped with `fmt.Errorf("context: %w", err)`;
      a string that starts with a capital letter or ends with `.` or `!` produces
      malformed wrapped messages: `"Serve: Listen: connection refused."`.
      ✗  `"Failed to listen on port"` — capitalised first word
         `"connection error."` — trailing period
         `"no arguments"` — missing `<command>:` prefix; wraps as `"context: no arguments"` with no origin
      ✓  `"serve: listen :8080: address in use"`
         `"echo: no arguments provided"`

| Anti-pattern | Level | Rule | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `os.Exit` inside a command | CODE | §5.1 | Deferred cleanup skipped; test binary terminated; exit code uncontrollable | Return `error`; `main` calls `os.Exit` |
| Non-`ExitError` error after printing diagnostics *(ff/v4)* | CODE | §5.2 | Dispatcher prints redundant `"error: ..."` after command already explained the failure | `return root.ExitError(1)` |
| `cobra.Command.Run` instead of `RunE` *(Cobra)* | CODE | §5.3 | Errors cannot be returned; forces `os.Exit` or silent discard | Use `RunE` exclusively |
| Missing `SilenceErrors`/`SilenceUsage` on root *(Cobra)* | CODE | §5.4 | Error printed twice; usage template appended to every runtime failure | Both `true` on root; `cmd.Usage()` only for genuine usage mistakes |
| Printing error on `ff.ErrHelp`/`ff.ErrNoExec` *(ff/v4)* | CODE | §5.5 | User sees `"error: help requested"` — hostile UX | `errors.Is` check; return exit code 0 for both |
| Capitalised or punctuated error strings | CODE | §5.6 | Wrapped errors malformed: `"Failed to connect: EOF."` | Lowercase, no trailing period: `"connect: EOF"` |

---

## §6  Post-Parse Initialization and Argument Validation

§6.1  [MUST][CODE]   Construct dependencies that require parsed flag values
      (API clients, loggers, DB connections) after all flag parsing completes and
      before any command execution begins.
      Constructing dependencies before parse means they receive every flag's zero
      value; constructing them inside the execution function mixes initialization
      with business logic, duplicates the construction on every invocation, and
      prevents the initialization from being tested in isolation.
      ff/v4: perform initialization between `Parse()` and `Run()` in the
      dispatcher (`cmd/cmd.go`).
      Cobra: perform initialization in `PersistentPreRunE` on the root command,
      which runs after all persistent flags are parsed and before any `RunE`.
      Note: Cobra's `PersistentPreRunE` does not execute when a subcommand is
      tested in isolation without a root parent; provide dependencies directly in
      unit tests or use integration tests through `cmd.Execute`.
      ✗  Cobra: `cmd.RunE = func(cmd *cobra.Command, args []string) error {`
         `    client, err := api.NewClient(token) // token is "" — flags not parsed yet`
         `    if err != nil { return err }`
         `    return client.DoWork(cmd.Context())`
         `}`
      ✓  Cobra: `root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {`
         `    var err error`
         `    sharedClient, err = api.NewClient(token) // token is populated by now`
         `    return err`
         `}`

§6.2  [MUST][CODE]   Set `Args` explicitly on every `cobra.Command`; do not
      rely on the nil (zero-value) default. *(Cobra only)*
      The nil default is equivalent to `ArbitraryArgs`: it silently accepts any
      positional arguments regardless of count or value. A typo that adds an extra
      argument produces no error; the extra argument is silently passed to `RunE`
      where it may be ignored or misinterpreted.
      ✗  `cmd := &cobra.Command{Use: "serve [FLAGS]"}` — nil Args
      ✓  `cmd := &cobra.Command{Use: "serve [FLAGS]", Args: cobra.NoArgs}`

§6.3  [SHOULD][CODE]  Define `PersistentPreRunE` only on the root command; do
      not define it on subcommands unless `cobra.EnableTraverseRunHooks = true`
      is set before `ExecuteContext`. *(Cobra only)*
      Cobra's default hook traversal stops at the first ancestor that defines
      `PersistentPreRunE`; when a subcommand also defines it, the root's hook is
      silently skipped for that subcommand. Shared state initialised in the root
      hook (API clients, DB connections) remains at its zero value, causing
      nil-pointer panics at runtime.
      ✗  Root and a subcommand both define `PersistentPreRunE` without
         `cobra.EnableTraverseRunHooks = true`
      ✓  Root defines `PersistentPreRunE`; subcommands use `PreRunE` for local
         pre-execution logic; or set `cobra.EnableTraverseRunHooks = true` once
         at startup if subcommand hooks are genuinely required

§6.4  [SHOULD][CODE]  Pass shared dependencies to separate-package subcommands
      via a getter function, not a direct pointer or value captured at construction
      time. *(Cobra only)*
      Subcommands are constructed before `PersistentPreRunE` runs; a direct
      pointer or value captured at construction time holds the zero value for the
      entire lifetime of the command. A getter closure reads the value that
      `PersistentPreRunE` set — always correctly, with no pointer indirection at
      the call site.
      ✗  `root.AddCommand(sub.NewCommand(sharedClient))` — `sharedClient` is nil
         at construction time; `RunE` sees nil
      ✓  `root.AddCommand(sub.NewCommand(func() *api.Client { return sharedClient }))`
         — getter reads the post-init value at call time

| Anti-pattern | Level | Rule | Failure mode | Preferred alternative |
|---|---|---|---|---|
| Client construction inside `RunE`/`exec` | CODE | §6.1 | Construction duplicated on every invocation; receives zero flag values | Construct in dispatcher (ff/v4) or `PersistentPreRunE` (Cobra) |
| Missing `Args` on Cobra command *(Cobra)* | CODE | §6.2 | Extra positional args silently accepted; typos produce no error | `cobra.NoArgs` / `cobra.ExactArgs(n)` on every command |
| Subcommand `PersistentPreRunE` without traverse hooks *(Cobra)* | CODE | §6.3 | Root init hook silently skipped; shared state remains nil at runtime | Reserve `PersistentPreRunE` for root; set `cobra.EnableTraverseRunHooks` if subcommand hooks needed |
| Direct pointer to shared state at construction *(Cobra)* | CODE | §6.4 | Dependency nil at construction; zero value used even after `PersistentPreRunE` runs | Getter closure `func() *T` reads the post-init value |

---

## Permitted Exceptions

**`var Version = "dev"` in the version command package** is the sole permitted
package-level mutable variable in a command package. The Go linker's
`-ldflags "-X <pkg>.Version=<val>"` mechanism can only override a package-level
`var`, not a constant or a local variable. Do not use `init()` to populate it;
read build info (VCS commit, Go version, module sum) inside the execution function.

---

## Master Anti-Patterns Table

| Anti-pattern | Level | Rule | Failure mode | Preferred alternative |
|---|---|---|---|---|
| `init()` registration | ARCH | §1.1 | Implicit ordering causes non-deterministic registration; untraceable in large repos | Explicit `AddCommand` / `append` in constructor |
| `var rootCmd = &cobra.Command{...}` | ARCH | §1.2 | Shared state; parallel-test data races; flag values persist across test invocations | Construct command inside constructor accepting I/O writers |
| Package-level flag variables (`var verbose bool`) | ARCH | §1.2 | Values from one test invocation bleed into the next; parallel tests corrupt each other | Declare inside constructor; close over in `RunE`/`exec` |
| Custom `Commander` interface | ARCH | §1.3 | Framework struct changes require lockstep updates to all adapter types; compiler errors appear in command packages, not at the interface | Function-pointer field on framework struct |
| Multiple exported symbols per command package | ARCH | §1.4 | Callers couple to internals; blocks refactoring | One exported constructor (ff/v4: `New` + `Config`; Cobra: `NewCommand`; stdlib: `<Name>Command`) |
| `os.Exit` inside `cmd` package | ARCH | §1.5 | Dispatcher bypasses returned error; `main` loses exit-code control; test binary terminates mid-run | Return `error`; `main` owns `os.Exit` |
| `fmt.Fprintln(os.Stdout, ...)` in command | CODE | §2.1 | Output uncapturable in tests; parallel tests must serialise | Write through injected `stdout io.Writer` |
| `os.Exit` before `stop()` | CODE | §2.2 | Signal-forwarding goroutine leaks; subsequent signals undelivered | Call `stop()` before `os.Exit` |
| `defer cleanup()` in `main()` with `os.Exit` | CODE | §2.3 | Deferred functions never run; resources leak on every non-zero exit | Move logic to `run()` returning `int`; `main()` only calls `os.Exit` |
| Binding flags inside `RunE`/`exec` | CODE | §3.1 | Variable always holds zero value; flag has no effect | Bind in constructor before assigning `RunE`/`Exec` |
| `ff.NewFlagSet` without `SetParent` *(ff/v4)* | CODE | §3.2 | Parse fails with "flag provided but not defined"; user made no mistake; cause invisible | `.SetParent(parent.Flags)` on every subcommand `FlagSet` |
| `flag.Parse()` (global) in command *(stdlib)* | CODE | §3.3 | Cannot be called twice; pollutes global state; breaks test isolation | `flag.NewFlagSet` inside command function |
| `os.Getenv` outside flag set *(ff/v4)* | CODE | §3.4 | Setting invisible to `-h`; not mapped by env-var prefix; not overridable by flag | Declare as a registered flag |
| Duplicate dispatch key | CODE | §4.1 | One entry permanently unreachable (ff/v4/Cobra: shadowed; stdlib: overwritten); no error at registration or startup | Unique `Name`/`Use`/`UsageLine` across all registered commands |
| `switch`/`case` dispatch *(stdlib)* | CODE | §4.2 | Dispatcher file edited on every add; merge-conflict hotspot | Map keyed on `UsageLine` |
| Unguarded map lookup on dispatch *(stdlib)* | CODE | §4.2 | Nil pointer dereference instead of a useful error on unknown command | Check `ok`; return `fmt.Errorf("%s: unknown command", name)` |
| `os.Exit` inside a command | CODE | §5.1 | Deferred cleanup skipped; test binary terminated; exit code uncontrollable | Return `error`; `main` calls `os.Exit` |
| Non-`ExitError` error after printing diagnostics *(ff/v4)* | CODE | §5.2 | Dispatcher prints redundant `"error: ..."` after command already explained the failure | `return root.ExitError(1)` |
| `cobra.Command.Run` instead of `RunE` *(Cobra)* | CODE | §5.3 | Errors cannot be returned; forces `os.Exit` or silent discard | Use `RunE` exclusively |
| Missing `SilenceErrors`/`SilenceUsage` on root *(Cobra)* | CODE | §5.4 | Error printed twice; usage template appended to every runtime failure | Both `true` on root; `cmd.Usage()` only for genuine usage mistakes |
| Printing error on `ff.ErrHelp`/`ff.ErrNoExec` *(ff/v4)* | CODE | §5.5 | User sees `"error: help requested"` — hostile UX | `errors.Is` check; return exit code 0 for both |
| Capitalised or punctuated error strings | CODE | §5.6 | Wrapped errors malformed: `"Failed to connect: EOF."` | Lowercase, no trailing period: `"connect: EOF"` |
| Client construction inside `RunE`/`exec` | CODE | §6.1 | Construction duplicated on every invocation; receives zero flag values | Construct in dispatcher (ff/v4) or `PersistentPreRunE` (Cobra) |
| Missing `Args` on Cobra command *(Cobra)* | CODE | §6.2 | Extra positional args silently accepted; typos produce no error | `cobra.NoArgs` / `cobra.ExactArgs(n)` on every command |
| Subcommand `PersistentPreRunE` without traverse hooks *(Cobra)* | CODE | §6.3 | Root init hook silently skipped; shared state remains nil at runtime | Reserve `PersistentPreRunE` for root; set `cobra.EnableTraverseRunHooks` if subcommand hooks needed |
| Direct pointer to shared state at construction *(Cobra)* | CODE | §6.4 | Dependency nil at construction; zero value used even after `PersistentPreRunE` runs | Getter closure `func() *T` reads the post-init value |

