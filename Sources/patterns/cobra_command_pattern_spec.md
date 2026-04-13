# Cobra CLI Command Pattern Spec

## Overview

Use this spec when `github.com/spf13/cobra` is a dependency of the project.
Cobra provides subcommand dispatch, persistent flags, argument validation, shell
completions, and help generation. It uses `github.com/spf13/pflag` for flag
parsing.

Replace `<org>/<repo>` throughout with the Go module path from `go.mod`.

---

## Architecture Rules

| Rule | Rationale |
|---|---|
| No `init()` for registration | `AddCommand` calls in the root constructor are explicit and traceable; `init()` execution order is implicit |
| No package-level `var rootCmd` | Globals prevent isolated testing; pass I/O writers into the root constructor |
| Use `RunE`, not `Run` | `Run` cannot return errors; `RunE` returns `error` which Cobra propagates to the caller |
| `SilenceErrors: true` on root | Prevents Cobra from printing the error before returning it; `main` prints and controls exit |
| `SilenceUsage: true` on root | Prevents Cobra from printing the full usage template on every `RunE` error; runtime failures (network, DB) are not usage mistakes. Call `cmd.Usage()` manually before returning errors that *are* usage mistakes |
| Write to `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` | Respects writers set by the constructor; avoids hardcoded `os.Stdout` / `os.Stderr` |
| Use `ExecuteContext` with `root.SetArgs` | Enables context threading and deterministic arg injection for tests |
| Bind flag values in `NewCommand`, not in `RunE` | Flags are parsed before `RunE` is called; binding inside `RunE` is too late and has no effect |
| Return `error`; never call `os.Exit` in a command | Only `main` controls exit codes |
| Error strings: lowercase, no trailing punctuation | Format: `<command>: <reason>` |

---

## Directory Structure

Flat layout (simple CLIs):

```
/
├── main.go                        # Entry point only — no logic
├── go.mod
└── cmd/
    ├── cmd.go                     # Execute function (package cmd)
    ├── root.go                    # newRootCommand constructor (package cmd)
    ├── version.go                 # version subcommand (package cmd)
    └── <name>.go                  # one file per subcommand (package cmd)
```

Per-package layout (larger CLIs with substantial per-command logic or tests):

```
/
├── main.go
├── go.mod
└── cmd/
    ├── cmd.go                     # Execute function (package cmd)
    ├── root.go                    # newRootCommand constructor (package cmd)
    ├── version/
    │   └── version.go             # package version
    └── <name>/
        └── <name>.go              # package <name>
```

Both layouts work. Use the flat layout for simple CLIs; switch to per-package when
a subcommand has its own test file or significant implementation.

---

## `cobra.Command` Fields

The relevant exported fields. The full struct is in the Cobra source; only the
fields used in practice are listed here.

```go
type Command struct {
    // Use is the one-line usage message. The first word is the dispatch name.
    // Syntax: "serve [FLAGS]" — dispatch name is "serve".
    Use string

    // Aliases are alternative dispatch names accepted in place of Use.
    Aliases []string

    // Short is shown next to the command name in parent help output.
    Short string

    // Long is shown in the command's own '--help' output.
    Long string

    // Example is shown below Long in the command's '--help' output.
    Example string

    // Args is the positional argument validator. Default (nil): ArbitraryArgs.
    // Always set this explicitly.
    Args PositionalArgs

    // ValidArgs is the list of valid positional argument values for shell completion.
    ValidArgs []string

    // RunE is the terminal function called when this command is selected.
    // args are positional arguments after flag parsing.
    // Use RunE, not Run, so errors can be returned.
    RunE func(cmd *Command, args []string) error

    // PersistentPreRunE runs before RunE for this command and all descendants.
    // Use for initialization that requires parsed persistent flag values.
    PersistentPreRunE func(cmd *Command, args []string) error

    // SilenceErrors suppresses automatic error printing. Set true on root.
    SilenceErrors bool

    // SilenceUsage suppresses automatic usage printing on error. Set true on root.
    SilenceUsage bool
}
```

Do not call `Execute` or `ExecuteContext` from subcommand packages. Those are
called only from `cmd/cmd.go`.

---

## Root Command — `cmd/root.go`

`newRootCommand` accepts I/O writers, sets them on the command tree, and registers
all subcommands. It is unexported; only `Execute` in `cmd/cmd.go` calls it.

```go
// cmd/root.go
package cmd

import (
    "io"

    "github.com/spf13/cobra"
    "github.com/<org>/<repo>/cmd/version"
    // add subcommand imports here
)

// verbose is a package-level variable so that PersistentPreRunE and any code
// in the cmd package (e.g. shared initialization) can read it without needing
// to traverse the flag set. Subcommands in separate packages access it via
// cmd.Root().PersistentFlags().Lookup("verbose") or through a shared struct
// set up in PersistentPreRunE.
var verbose bool

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
    root := &cobra.Command{
        Use:           "<cli-name> [FLAGS] <COMMAND>",
        Short:         "<one-line description of the program>",
        SilenceErrors: true,
        SilenceUsage:  true,
    }
    root.SetOut(stdout)
    root.SetErr(stderr)

    // Persistent flags are inherited by all subcommands.
    root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

    // Register subcommands here. Order controls help output order.
    root.AddCommand(version.NewCommand())
    // root.AddCommand(<name>.NewCommand())

    return root
}
```

If the application has no shared persistent flags, omit the `PersistentFlags()`
block entirely.

---

## Command Template

Each subcommand exports exactly one function: `NewCommand`. Flag variables are
declared inside `NewCommand` (not at package level) so that each call produces
independent state — this prevents data races when tests run commands in parallel.
`RunE` is a closure over those variables.

```go
// cmd/<name>/<name>.go
package <name>

import (
    "github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
    // Declare flag-destination variables here, inside NewCommand.
    // Each call gets its own copy — no shared state between commands or tests.

    cmd := &cobra.Command{
        Use:   "<name> [FLAGS]",
        Short: "<one-line description>",
        Long:  "<paragraph description>",
        Args:  cobra.NoArgs,
    }
    // Bind local flags to the variables declared above.
    // cmd.Flags().StringVarP(&someField, "some-flag", "s", "", "description")

    cmd.RunE = func(cmd *cobra.Command, args []string) error {
        // cmd.OutOrStdout() and cmd.ErrOrStderr() for output.
        // cmd.Context() for the request context.
        return nil
    }
    return cmd
}
```

### Rules

- `NewCommand` is the **only exported symbol** in the package.
- Declare flag-destination variables **inside `NewCommand`** (not at package level).
  Closures capture them; each `NewCommand()` call gets independent state.
- Bind flags **in `NewCommand`**, not inside `RunE`. Flags are parsed before `RunE`
  runs.
- Use `RunE`, not `Run`. Commands that cannot fail should still return `nil`.
- Write to `cmd.OutOrStdout()` and `cmd.ErrOrStderr()`. Never use `os.Stdout` /
  `os.Stderr` directly.
- Set `Args` explicitly. The default (`nil`) silently accepts any arguments.
- Obtain a context via `cmd.Context()` inside `RunE`. This is set by
  `ExecuteContext` in the dispatcher.
- Do not call `os.Exit` inside a command.

---

## Concrete Examples

### Command with no flags

```go
// cmd/version/version.go
package version

import (
    "fmt"

    "github.com/spf13/cobra"
)

// Version is set at build time:
// go build -ldflags "-X 'github.com/<org>/<repo>/cmd/version.Version=1.2.3'"
var Version = "dev"

func NewCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "version",
        Short: "print version information",
        Long:  "Prints version information for the application.",
        Args:  cobra.NoArgs,
    }
    cmd.RunE = func(cmd *cobra.Command, _ []string) error {
        _, err := fmt.Fprintln(cmd.OutOrStdout(), "version "+Version)
        return err
    }
    return cmd
}
```

### Command that reads positional arguments

```go
// cmd/echo/echo.go
package echo

import (
    "fmt"
    "strings"

    "github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "echo <ARG>...",
        Short: "echo arguments",
        Long:  "Prints all provided arguments joined by spaces.",
        Args:  cobra.MinimumNArgs(1),
    }
    cmd.RunE = func(cmd *cobra.Command, args []string) error {
        _, err := fmt.Fprintln(cmd.OutOrStdout(), strings.Join(args, " "))
        return err
    }
    return cmd
}
```

### Command with flags

Flag-destination variables are declared inside `NewCommand` and captured by the
`RunE` closure. Each call to `NewCommand` produces an independent set of
variables — safe for parallel tests.

```go
// cmd/shout/shout.go
package shout

import (
    "fmt"
    "strings"

    "github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
    var noNewline bool

    cmd := &cobra.Command{
        Use:   "shout [FLAGS] <ARG>...",
        Short: "print arguments in uppercase",
        Long:  "Prints arguments in uppercase. Use -n to suppress newline.",
        Args:  cobra.MinimumNArgs(1),
    }
    cmd.Flags().BoolVarP(&noNewline, "no-newline", "n", false, "suppress trailing newline")

    cmd.RunE = func(cmd *cobra.Command, args []string) error {
        out := strings.ToUpper(strings.Join(args, " "))
        if noNewline {
            _, err := fmt.Fprint(cmd.OutOrStdout(), out)
            return err
        }
        _, err := fmt.Fprintln(cmd.OutOrStdout(), out)
        return err
    }
    return cmd
}
```

### Group parent (no `RunE`)

A command that exists only to group subcommands should have no `RunE`. When
invoked without a subcommand, Cobra prints the command's help and returns nil.

```go
// cmd/config/config.go
package config

import (
    "github.com/spf13/cobra"
    "github.com/<org>/<repo>/cmd/config/get"
    "github.com/<org>/<repo>/cmd/config/set"
)

func NewCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config <SUBCOMMAND>",
        Short: "manage configuration",
        // No RunE — invoking 'config' alone prints help.
    }
    cmd.AddCommand(get.NewCommand())
    cmd.AddCommand(set.NewCommand())
    return cmd
}
```

---

## Execute Function — `cmd/cmd.go`

`cmd/cmd.go` is the **only place `newRootCommand` is called**.

```go
// cmd/cmd.go
package cmd

import (
    "context"
    "io"
)

// Execute constructs the command tree and runs the selected command.
// args must not include the executable name (pass os.Args[1:]).
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
    root := newRootCommand(stdout, stderr)
    root.SetArgs(args)
    return root.ExecuteContext(ctx)
}
```

`SetArgs` overrides Cobra's default of reading `os.Args[1:]` directly. Without
it, tests would pick up the test binary's own args.

`ExecuteContext` propagates `ctx` to all `RunE` functions via `cmd.Context()`.

---

## Hook Lifecycle

Hooks execute in this order for every selected command:

```
PersistentPreRunE  (closest ancestor that defines it)
PreRunE            (current command only)
RunE               (current command only)
PostRunE           (current command only)
PersistentPostRunE (closest ancestor that defines it)
```

**The "closest ancestor" rule is critical**: by default Cobra walks up the command
tree and calls only the *first* `PersistentPreRunE` it finds — it stops there. If
the root defines `PersistentPreRunE` for initialization and a subcommand also
defines its own `PersistentPreRunE`, the root's hook will **not** run for that
subcommand. This silently skips initialization, potentially causing nil-pointer
panics at runtime.

To run all ancestors' persistent hooks, set the global option once at startup:

```go
// main.go or cmd/cmd.go — before ExecuteContext
cobra.EnableTraverseRunHooks = true
```

As a simpler rule: **avoid defining `PersistentPreRunE` on subcommands** when the
root uses it for initialization. Reserve `PersistentPreRunE` for the root only.

`PreRunE` and `PostRunE` are not inherited by children — they apply only to the
command that defines them.

---

## Post-Parse Initialization

Dependencies that require parsed persistent flag values (API clients, DB
connections, loggers) should be constructed in `PersistentPreRunE` on the root
command. This runs after all flag parsing, before any subcommand's `RunE`.

Store shared state in `cmd` package-level variables so that subcommands in the
same package can read them directly. For subcommands in separate packages, pass
a pointer to a shared struct at construction time:

```go
// cmd/root.go

// sharedClient is populated by PersistentPreRunE before any RunE runs.
// Subcommands in package cmd read it directly.
// Subcommands in separate packages receive it via their NewCommand constructor.
var sharedClient *api.Client

var token string

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
    root := &cobra.Command{...}
    root.PersistentFlags().StringVar(&token, "token", "", "API token")

    root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
        client, err := api.NewClient(token) // token is parsed by this point
        if err != nil {
            return fmt.Errorf("construct client: %w", err)
        }
        sharedClient = client
        return nil
    }

    root.AddCommand(version.NewCommand())
    // Subcommands in separate packages: pass a getter function so RunE reads
    // the value PersistentPreRunE set, not the zero value at construction time.
    // root.AddCommand(mysubcmd.NewCommand(func() *api.Client { return sharedClient }))
    return root
}
```

In the subcommand, pass a getter function instead of a double pointer. The getter
is a closure over the package-level variable, so it always reads the value that
`PersistentPreRunE` set — no pointer indirection required at the call site:

```go
// cmd/root.go
root.AddCommand(mysubcmd.NewCommand(func() *api.Client { return sharedClient }))
```

```go
// cmd/mysubcmd/mysubcmd.go
func NewCommand(getClient func() *api.Client) *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.RunE = func(cmd *cobra.Command, args []string) error {
        // getClient() is non-nil here because PersistentPreRunE ran first.
        return getClient().DoSomething(cmd.Context())
    }
    return cmd
}
```

---

## Argument Validation

Set `Args` on every command. The default (`nil`) is equivalent to `ArbitraryArgs`
— it silently accepts any arguments.

| Validator | Behaviour |
|---|---|
| `cobra.NoArgs` | Error if any positional args are provided |
| `cobra.ArbitraryArgs` | Accept any number of args (explicit no-op) |
| `cobra.ExactArgs(n)` | Error if not exactly n args |
| `cobra.MinimumNArgs(n)` | Error if fewer than n args |
| `cobra.MaximumNArgs(n)` | Error if more than n args |
| `cobra.RangeArgs(min, max)` | Error if count is outside [min, max] |
| `cobra.OnlyValidArgs` | Error if any arg is not in `ValidArgs` |
| `cobra.MatchAll(a, b, ...)` | Combines multiple validators; all must pass |

For domain-specific validation, implement `func(cmd *cobra.Command, args []string) error` directly:

```go
Args: func(cmd *cobra.Command, args []string) error {
    if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
        return err
    }
    if !isValidColor(args[0]) {
        return fmt.Errorf("shout: invalid color: %q", args[0])
    }
    return nil
},
```

### Showing usage on genuine usage errors

With `SilenceUsage: true` on the root, Cobra never prints usage automatically.
For errors that *are* usage mistakes (wrong argument count, invalid flag value),
print usage explicitly before returning:

```go
func NewCommand() *cobra.Command {
    var mode string

    cmd := &cobra.Command{...}
    cmd.Flags().StringVar(&mode, "mode", "fast", `processing mode: "fast" or "slow"`)

    cmd.RunE = func(cmd *cobra.Command, args []string) error {
        if mode != "fast" && mode != "slow" {
            _ = cmd.Usage() // prints usage to cmd.ErrOrStderr()
            return fmt.Errorf("shout: invalid mode %q: must be fast or slow", mode)
        }
        // ... runtime work; do NOT call cmd.Usage() on network/IO errors
        return nil
    }
    return cmd
}
```

Do not call `cmd.Usage()` for runtime errors (network failures, DB errors,
permission denied). Usage text is irrelevant to those failures and adds noise.

---

## Flag Reference

### Local flags

Available only on the command they are defined on. Bind to variables declared
inside `NewCommand`:

```go
cmd.Flags().StringVarP(&field, "long-name", "s", "", "description")
cmd.Flags().BoolVarP(&field, "long-name", "b", false, "description")
cmd.Flags().IntVarP(&field, "long-name", "i", 0, "description")
```

### Persistent flags

Inherited by all subcommands. Define on the root or on a group parent:

```go
root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
```

### Required flags

```go
cmd.Flags().StringVarP(&region, "region", "r", "", "AWS region (required)")
cmd.MarkFlagRequired("region")

// Persistent variant:
root.PersistentFlags().StringVarP(&token, "token", "t", "", "API token (required)")
root.MarkPersistentFlagRequired("token")
```

### Flag groups

```go
// Both flags must be provided together, or neither:
cmd.MarkFlagsRequiredTogether("username", "password")

// At most one of these flags may be set:
cmd.MarkFlagsMutuallyExclusive("json", "yaml")

// At least one of these flags must be set:
cmd.MarkFlagsOneRequired("json", "yaml")

// Exactly one (combine both):
cmd.MarkFlagsOneRequired("json", "yaml")
cmd.MarkFlagsMutuallyExclusive("json", "yaml")
```

---

## Entry Point — `main.go`

`SilenceErrors: true` on the root means Cobra does not automatically print the
error. `main` prints it and controls the exit code.

```go
// main.go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/<org>/<repo>/cmd"
)

const (
    exitFail    = 1
    exitSuccess = 0
)

func main() {
    ctx := context.Background()
    if err := cmd.Execute(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(exitFail)
    }
}
```

---

## Testing

Because `Execute` accepts `io.Writer` parameters and `root.SetArgs` overrides the
OS args, all output is capturable. Because flag variables are declared inside
`NewCommand`, parallel tests have no shared state.

### Integration test via `cmd.Execute`

Prefer integration tests — they exercise the full parse → init → exec path:

```go
// cmd/cmd_test.go
package cmd_test

import (
    "bytes"
    "context"
    "io"
    "testing"

    "github.com/<org>/<repo>/cmd"
)

func TestShout(t *testing.T) {
    var stdout bytes.Buffer
    err := cmd.Execute(context.Background(), []string{"shout", "hello", "world"}, &stdout, io.Discard)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got, want := stdout.String(), "HELLO WORLD\n"; got != want {
        t.Errorf("got %q, want %q", got, want)
    }
}

func TestNoSubcommand(t *testing.T) {
    // Root with no subcommand and no RunE: Cobra prints help and returns nil.
    err := cmd.Execute(context.Background(), []string{}, io.Discard, io.Discard)
    if err != nil {
        t.Fatalf("expected nil, got %v", err)
    }
}
```

### Unit test of a single subcommand

Construct the command directly, set I/O writers on it, then call `Execute`.
Because flag variables are local to `NewCommand`, each test call is isolated.

**`PersistentPreRunE` does not run in subcommand unit tests.** When a subcommand
is constructed and executed without a root parent, there is no root
`PersistentPreRunE` in the chain. Any shared dependencies normally initialized
by that hook (API clients, DB connections) will be at their zero value. Either
provide them explicitly in the test (e.g. inject a test client via the
constructor) or use integration tests through `cmd.Execute` when the full
initialization chain matters.

```go
// cmd/shout/shout_test.go
package shout_test

import (
    "bytes"
    "context"
    "io"
    "testing"

    "github.com/<org>/<repo>/cmd/shout"
)

func TestShout_basic(t *testing.T) {
    var stdout bytes.Buffer
    cmd := shout.NewCommand()
    cmd.SetOut(&stdout)
    cmd.SetErr(io.Discard)
    cmd.SetArgs([]string{"hello", "world"})
    if err := cmd.ExecuteContext(context.Background()); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got, want := stdout.String(), "HELLO WORLD\n"; got != want {
        t.Errorf("got %q, want %q", got, want)
    }
}

func TestShout_noArgs(t *testing.T) {
    cmd := shout.NewCommand()
    cmd.SetOut(io.Discard)
    cmd.SetErr(io.Discard)
    cmd.SetArgs([]string{})
    if err := cmd.ExecuteContext(context.Background()); err == nil {
        t.Fatal("expected error for missing args, got nil")
    }
}
```

---

## Checklist: Adding a New Command

- [ ] Create `cmd/<name>/` package (or `cmd/<name>.go` for flat layout)
- [ ] Write exported `NewCommand() *cobra.Command`
- [ ] Set `Use`, `Short`, and `Args` on the returned command
- [ ] Use `RunE` (not `Run`) as a closure assigned after construction
- [ ] Declare flag-destination variables **inside `NewCommand`** (not at package level)
- [ ] Bind flag values in `NewCommand` before assigning `RunE`
- [ ] Write to `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`; never `os.Stdout` / `os.Stderr`
- [ ] Call `root.AddCommand(<name>.NewCommand())` in `newRootCommand` in `cmd/root.go`
- [ ] Add the import for the new package in `cmd/root.go`

---

## Pattern Comparison

| Concern | Cobra | ff/v4 |
|---|---|---|
| External dependency | `github.com/spf13/cobra` + `github.com/spf13/pflag` | `github.com/peterbourgon/ff/v4` |
| Command representation | `*cobra.Command` (library struct) | `*ff.Command` (library struct) |
| Flag parsing | pflag via `cmd.Flags()` / `cmd.PersistentFlags()` | `ff.NewFlagSet` with `SetParent` |
| Flag inheritance | `PersistentFlags()` on a parent command | `SetParent(parent.Flags)` on every subcommand flag set |
| Flag-value state | Variables closed over inside `NewCommand` | Fields on `Config` struct |
| Context | `cmd.Context()` inside `RunE` | `ctx` argument to `exec` |
| Output | `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` | `cfg.Stdout` / `cfg.Stderr` via embedded `root.Config` |
| Registration | `root.AddCommand(...)` in root constructor | `append` to `Subcommands` inside `New()` |
| Post-parse init | `PersistentPreRunE` on root | Between `Parse()` and `Run()` in `cmd/cmd.go` |
| Help | Automatic; `--help` / `-h` at any level | `ff.ErrHelp`; `--help` / `-h` at any level |
| No subcommand given (root has no `RunE`) | Cobra prints root help; returns nil | `ff.ErrNoExec`; dispatcher prints help and returns nil |
| Group parent commands | Omit `RunE`; Cobra prints help on bare invocation | Set `Exec: nil`; returns `ff.ErrNoExec` |
| Shell completions | Built-in (bash, zsh, fish, powershell) | Not built-in |
| Doc generation | Built-in (markdown, man, yaml) | Not built-in |
| Dispatch key | First word of `Use` (case-sensitive) | `Name` field (case-insensitive) |
| Testability | `Execute(ctx, args, stdout, stderr)` + per-command `cmd.ExecuteContext` | `Run(ctx, args, stdout, stderr)` |
