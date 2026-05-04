# Go CLI Command Pattern Spec

## Overview

This project implements a CLI dispatcher using `github.com/peterbourgon/ff/v4` (`ff`).
Commands are represented as `ff.Command` values. Do not use other CLI frameworks
(cobra, urfave/cli, etc.). Do not use interfaces for command polymorphism.
Use the `ff.Command` struct with the Config struct pattern described below.

Note: Replace `<org>/<repo>` below with the Go module path from `go.mod`.

**Flags-first configuration is a core requirement.** Every knob that affects
behaviour must be a registered flag on an `ff.FlagSet`. `ff` then knows how to
parse that value from CLI args, environment variables, and config files, and how
to describe it in `-h` output. The invariant to preserve: running any command
with `-h` (at any depth) must reveal the complete configuration surface area of
that command. Never hide behaviour behind hard-coded values, package-level
variables, or `os.Getenv` calls that bypass the flag set.

______________________________________________________________________

## Directory Structure

```
/
├── main.go                        # Entry point only — no logic
├── go.mod
├── cmd/
│   ├── cmd.go                     # Dispatcher and command registration (package cmd)
│   ├── root/                      # Default name; configurable via climax init --root-pkg
│   │   └── root.go                # Config: shared I/O, flags, root ff.Command, ExitError
│   ├── version/
│   │   └── version.go             # Version command (omit with climax init --no-version)
│   └── <name>/
│       └── <name>.go              # One package per command
```

The dispatcher file may also be named `command.go`; climax's AST analysis accepts either.

______________________________________________________________________

## Command Type

All commands use `ff.Command` from `github.com/peterbourgon/ff/v4`. The relevant
exported fields are:

```go
type Command struct {
    Name        string                                // dispatch key; case-insensitive match
    Usage       string                                // one-line syntax, e.g. "<cli> <cmd> [FLAGS] <ARG>"
    ShortHelp   string                                // shown next to name in parent help
    LongHelp    string                                // shown in command's own help (optional)
    Flags       ff.Flags                              // nil → empty flag set; --help always works
    Subcommands []*Command                            // populated by each subcommand's New()
    Exec        func(context.Context, []string) error // nil → returns ff.ErrNoExec
}
```

Do not call `Parse`, `Run`, or any other method on `ff.Command` from command
packages. Those are called exclusively by the dispatcher in `cmd/cmd.go`.

______________________________________________________________________

## Root Config

`cmd/root/root.go` defines `Config`. It holds `stdin`/`stdout`/`stderr`, any flags
shared across all commands, and the root `ff.Command`. It also declares `ExitError`
(see below). Subcommand configs embed `*root.Config` to inherit these.

```go
// cmd/root/root.go
package root

import (
    "fmt"
    "io"

    "github.com/peterbourgon/ff/v4"
)

// ExitError lets a command exit with a specific code without printing "error: ...".
// Return it from exec; main.go handles it via errors.As.
type ExitError int

func (e ExitError) Error() string { return fmt.Sprintf("exit status %d", int(e)) }

type Config struct {
    Stdin   io.Reader
    Stdout  io.Writer
    Stderr  io.Writer
    Flags   *ff.FlagSet
    Command *ff.Command
}

func New(stdin io.Reader, stdout, stderr io.Writer) *Config {
    var cfg Config
    cfg.Stdin = stdin
    cfg.Stdout = stdout
    cfg.Stderr = stderr
    // No shared flags by default — cfg.Flags is nil; ff provides --help automatically.
    // To add shared flags, first declare fields on Config (e.g. Verbose bool), then:
    // cfg.Flags = ff.NewFlagSet("<cli-name>")
    // cfg.Flags.BoolVar(&cfg.Verbose, 'v', "verbose", "log verbose output")
    cfg.Command = &ff.Command{
        Name:      "<cli-name>",
        Usage:     "<cli-name> <SUBCOMMAND> ...",
        ShortHelp: "<one-line description>",
        Flags:     cfg.Flags,
    }
    return &cfg
}
```

______________________________________________________________________

## ExitError

`root.ExitError` lets a command exit with a specific non-zero code without
printing an `error: ...` line. The dispatcher suppresses help output for it,
and `run()` in `main.go` calls `os.Exit` directly via `errors.As`:

```go
// In any command's exec function:
if !ok {
    return root.ExitError(1) // exits 1; no "error:" printed
}
return nil // exits 0
```

Use `root.ExitError` whenever a command needs to signal failure to the shell
(e.g. `climax lint` finding issues, a check command finding violations) without
printing a redundant error message — the command has already written its own
output explaining the outcome.

______________________________________________________________________

## Adding a New Command

Each command lives in its own package under `cmd/`. The package contains a
`Config` struct, an exported `New` factory, and an unexported `exec` method.

### Command Template

```go
// cmd/<name>/<name>.go
package <name>

import (
    "context"
    "fmt"

    "github.com/peterbourgon/ff/v4"
    "<org>/<repo>/cmd/root"
)

type Config struct {
    *root.Config
    // command-local flag values declared here
    Flags   *ff.FlagSet
    Command *ff.Command
}

func New(parent *root.Config) *Config {
    var cfg Config
    cfg.Config = parent
    cfg.Flags = ff.NewFlagSet("<name>").SetParent(parent.Flags)
    // bind flags: cfg.Flags.StringVar(&cfg.SomeFlag, 0, "some-flag", "", "description")
    cfg.Command = &ff.Command{
        Name:      "<name>",
        Usage:     "<cli-name> <name> [FLAGS]",
        ShortHelp: "<one-line description>",
        LongHelp:  "<paragraph description>",
        Flags:     cfg.Flags,
        Exec:      cfg.exec,
    }
    parent.Command.Subcommands = append(parent.Command.Subcommands, cfg.Command)
    return &cfg
}

func (cfg *Config) exec(_ context.Context, _ []string) error {
    // cfg.Stdin, cfg.Stdout, cfg.Stderr available via embedded root.Config
    // flag values available as cfg fields — already parsed before exec is called
    _, _ = fmt.Fprintln(cfg.Stdout, "<name>: not yet implemented")
    return nil
}
```

### Rules

- `New` and `Config` are the only exported identifiers in the package (commands that need a user-visible package-level variable, such as `Version string`, may export it too).
- `New` appends to `parent.Command.Subcommands` — no other registration needed.
- Flag values are bound to `Config` fields in `New()`, not inside `exec`.
- Every behavioral knob must be a registered flag on `cfg.Flags`. Never use hard-coded values, package-level variables, or `os.Getenv` calls outside the flag set — they make settings invisible to `-h`.
- `SetParent(parent.Flags)` must be called on every subcommand flag set so that parent flags are accepted at any depth.
- Write to `cfg.Stdout` / `cfg.Stderr`. Never use `os.Stdout` / `os.Stderr` directly.
- Return `error` or `root.ExitError`. Do not call `os.Exit` inside a command.
- Error strings are lowercase, no trailing punctuation: `<command>: <reason>`.

### Nested Commands

A command nested under another non-root command embeds its parent's `Config`
instead of `*root.Config`, giving it access to both shared I/O and any flags
the parent defines:

```go
// cmd/create/create.go — nested under "config"
package create

import (
    "context"

    "github.com/peterbourgon/ff/v4"
    "<org>/<repo>/cmd/config"
)

type Config struct {
    *config.Config // embeds parent; root.Config accessible transitively
    Flags   *ff.FlagSet
    Command *ff.Command
}

func New(parent *config.Config) *Config {
    var cfg Config
    cfg.Config = parent
    cfg.Flags = ff.NewFlagSet("create").SetParent(parent.Flags)
    cfg.Command = &ff.Command{
        Name:      "create",
        Usage:     "<cli-name> config create [FLAGS]",
        ShortHelp: "create a new config entry",
        LongHelp:  "Creates a new configuration entry.",
        Flags:     cfg.Flags,
        Exec:      cfg.exec,
    }
    parent.Command.Subcommands = append(parent.Command.Subcommands, cfg.Command)
    return &cfg
}

func (cfg *Config) exec(_ context.Context, _ []string) error { return nil }
```

### Concrete Example — command with no flags

```go
// cmd/ping/ping.go
package ping

import (
    "context"
    "fmt"

    "github.com/peterbourgon/ff/v4"
    "<org>/<repo>/cmd/root"
)

type Config struct {
    *root.Config
    Flags   *ff.FlagSet
    Command *ff.Command
}

func New(parent *root.Config) *Config {
    var cfg Config
    cfg.Config = parent
    cfg.Flags = ff.NewFlagSet("ping").SetParent(parent.Flags)
    cfg.Command = &ff.Command{
        Name:      "ping",
        Usage:     "<cli-name> ping",
        ShortHelp: "check connectivity",
        LongHelp:  "Checks connectivity and prints a response.",
        Flags:     cfg.Flags,
        Exec:      cfg.exec,
    }
    parent.Command.Subcommands = append(parent.Command.Subcommands, cfg.Command)
    return &cfg
}

func (cfg *Config) exec(_ context.Context, _ []string) error {
    _, _ = fmt.Fprintln(cfg.Stdout, "pong")
    return nil
}
```

### Concrete Example — command that reads positional arguments

```go
// cmd/echo/echo.go
package echo

import (
    "context"
    "fmt"
    "strings"

    "<org>/<repo>/cmd/root"
    "github.com/peterbourgon/ff/v4"
)

type Config struct {
    *root.Config
    Flags   *ff.FlagSet
    Command *ff.Command
}

func New(parent *root.Config) *Config {
    var cfg Config
    cfg.Config = parent
    cfg.Flags = ff.NewFlagSet("echo").SetParent(parent.Flags)
    cfg.Command = &ff.Command{
        Name:      "echo",
        Usage:     "<cli-name> echo <ARG>...",
        ShortHelp: "echo arguments",
        LongHelp:  "Prints all provided arguments joined by spaces.",
        Flags:     cfg.Flags,
        Exec:      cfg.exec,
    }
    parent.Command.Subcommands = append(parent.Command.Subcommands, cfg.Command)
    return &cfg
}

func (cfg *Config) exec(_ context.Context, args []string) error {
    if len(args) == 0 {
        return fmt.Errorf("echo: no arguments provided")
    }
    _, _ = fmt.Fprintln(cfg.Stdout, strings.Join(args, " "))
    return nil
}
```

### Concrete Example — command with flags

Flag values are bound to `Config` fields in `New()`. By the time `exec` is
called, all flags are already parsed.

```go
// cmd/shout/shout.go
package shout

import (
    "context"
    "fmt"
    "strings"

    "<org>/<repo>/cmd/root"
    "github.com/peterbourgon/ff/v4"
)

type Config struct {
    *root.Config
    NoNewline bool
    Flags     *ff.FlagSet
    Command   *ff.Command
}

func New(parent *root.Config) *Config {
    var cfg Config
    cfg.Config = parent
    cfg.Flags = ff.NewFlagSet("shout").SetParent(parent.Flags)
    cfg.Flags.BoolVar(&cfg.NoNewline, 'n', "no-newline", "suppress trailing newline")
    cfg.Command = &ff.Command{
        Name:      "shout",
        Usage:     "<cli-name> shout [FLAGS] <ARG>...",
        ShortHelp: "print arguments in uppercase",
        LongHelp:  "Prints arguments in uppercase. Use -n/--no-newline to suppress newline.",
        Flags:     cfg.Flags,
        Exec:      cfg.exec,
    }
    parent.Command.Subcommands = append(parent.Command.Subcommands, cfg.Command)
    return &cfg
}

func (cfg *Config) exec(_ context.Context, args []string) error {
    out := strings.ToUpper(strings.Join(args, " "))
    if cfg.NoNewline {
        _, _ = fmt.Fprint(cfg.Stdout, out)
    } else {
        _, _ = fmt.Fprintln(cfg.Stdout, out)
    }
    return nil
}
```

### Rich version command

The generated version command prints rich build info (VCS commit, build date,
Go version, platform) with an optional `--json` flag for machine-readable output.

`var Version = "dev"` is a **deliberate exception** to the no-global-variables
rule: the Go linker's `-ldflags "-X <pkg>.Version=<val>"` mechanism can only
override a package-level `var`, not a constant or a local variable. This is the
only package-level mutable variable permitted in a climax command.

Do not use `init()` to populate it. Read build info inside `exec` instead —
the version string is only needed when the `version` subcommand actually runs.

Key structural elements of the full version command:

**`type Option func(i *Info)`** — functional-options type for post-gather customization.

**`type Info struct`** — exported struct with JSON tags. Fields: `GitVersion`, `ModuleSum`,
`GitCommit`, `GitTreeState`, `BuildDate`, `BuiltBy`, `GoVersion`, `Compiler`, `Platform`
(all JSON-serialized), plus `ASCIIName`, `Name`, `Description`, `URL` (JSON `-`, display only).

**`With*` constructors** — `WithAppDetails(name, description, url string) Option`,
`WithASCIIName(name string) Option`, `WithBuiltBy(name string) Option`.

**`GetVersionInfoFrom(bi *debug.BuildInfo, _ string, options ...Option) *Info`** — builds an
`Info` from an explicit `BuildInfo`, applying options. Accepts `nil` (all VCS fields become
`"unknown"`). Use this in tests to avoid touching global state.

**Pointer-receiver output methods on `*Info`**: `String() string` (tabwriter-aligned key:value
output) and `JSONString() (string, error)` (indented JSON).

```go
// cmd/version/version.go  (abbreviated)
package version

var Version = "dev"

type Option func(i *Info)

type Info struct {
    GitVersion   string `json:"gitVersion"`
    ModuleSum    string `json:"moduleChecksum"`
    GitCommit    string `json:"gitCommit"`
    GitTreeState string `json:"gitTreeState"`
    BuildDate    string `json:"buildDate"`
    BuiltBy      string `json:"builtBy"`
    GoVersion    string `json:"goVersion"`
    Compiler     string `json:"compiler"`
    Platform     string `json:"platform"`

    ASCIIName   string `json:"-"`
    Name        string `json:"-"`
    Description string `json:"-"`
    URL         string `json:"-"`
}

type Config struct {
    *root.Config
    JSON    bool
    Flags   *ff.FlagSet
    Command *ff.Command
}

func WithAppDetails(name, description, url string) Option { … }
func WithASCIIName(name string) Option                    { … }
func WithBuiltBy(name string) Option                      { … }

func New(parent *root.Config) *Config { … }

func GetVersionInfoFrom(bi *debug.BuildInfo, _ string, options ...Option) *Info { … }

func (i *Info) String() string             { … } // tabwriter key:value
func (i *Info) JSONString() (string, error) { … } // json.MarshalIndent

func (cfg *Config) exec(_ context.Context, _ []string) error { … }
```

______________________________________________________________________

## Registering Commands

`cmd/cmd.go` is the **only place `New()` is called**. It constructs the root
config, calls each subcommand's `New()` in order (which controls help output
order), and exposes `Run` for `main`.

Each `New()` call registers itself by appending to `parent.Command.Subcommands`.
Do not register commands anywhere else — no `init()`, no globals.

```go
// cmd/cmd.go
package cmd

// climax:name <cli-name>
// climax:root-pkg root
// climax:env-prefix <CLI_NAME>

import (
    "context"
    "errors"
    "fmt"
    "io"

    "github.com/peterbourgon/ff/v4"
    "github.com/peterbourgon/ff/v4/ffhelp"
    "<org>/<repo>/cmd/root"
    "<org>/<repo>/cmd/version"
    // climax:imports
)

// Run parses args and dispatches to the matching command.
// args must not include the executable name (pass os.Args[1:]).
//
// Every flag can be set via a <CLI_NAME>_-prefixed environment variable.
// The mapping rule: prepend <CLI_NAME>_, uppercase, replace dashes/dots with
// underscores. Flags on the command line always take precedence over env vars.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
    r := root.New(stdin, stdout, stderr)
    version.New(r)
    // register new commands here

    if err := r.Command.Parse(args, ff.WithEnvVarPrefix("<CLI_NAME>")); err != nil {
        _, _ = fmt.Fprintf(stderr, "\n%s\n", ffhelp.Command(r.Command))
        return fmt.Errorf("parse: %w", err)
    }

    if err := r.Command.Run(ctx); err != nil {
        // Suppress help output for ErrNoExec (no subcommand given) and
        // ExitError (command already reported its own outcome).
        var exitErr root.ExitError
        if !errors.Is(err, ff.ErrNoExec) && !errors.As(err, &exitErr) {
            _, _ = fmt.Fprintf(stderr, "\n%s\n", ffhelp.Command(r.Command.GetSelected()))
        }
        return err
    }

    return nil
}
```

### Dispatch Rules

- `Run` receives `os.Args[1:]` — executable name already removed by `main`.
- Subcommand selection is **case-insensitive** match on `Name`. No prefix matching.
- `-h` / `--help` at any level causes `Parse` to return `ff.ErrHelp`. `ff` prints the help text itself before returning; the caller does not print anything. Treated as success.
- A command with no `Exec` returns `ff.ErrNoExec`; treated as success.
- Unknown subcommand returns an error; `main` controls the exit code.

### Environment variable mapping

Two modes are available, chosen at `climax init` time:

**Prefixed mode** (default): `ff.WithEnvVarPrefix("MYAPP")` prepends the
application prefix. The derivation rule: uppercase the flag name, replace `-`
and `.` with `_`, prepend the prefix and `_`.

```
--port          → MYAPP_PORT
--log-level     → MYAPP_LOG_LEVEL
--db.host       → MYAPP_DB_HOST
```

**No-prefix mode** (`--no-env-prefix`): `ff.WithEnvVars()` enables env vars
without any prefix. The derivation rule is the same except no prefix is prepended.

```
--port          → PORT
--log-level     → LOG_LEVEL
--db.host       → DB_HOST
```

Use no-prefix mode for tools where the env var namespace is not shared with
other applications (e.g. a single-purpose container entrypoint), or where
operators expect unprefixed names by convention.

CLI args always win over env vars; env vars win over config files. Running
`<cli> <cmd> --help` shows flag names; the env var names are derived from them.

______________________________________________________________________

## Post-Parse Initialization

Because `ff` separates `Parse()` from `Run()`, dependencies that require parsed
flag values (API clients, DB connections, loggers) can be initialized in `cmd.go`
between the two calls and assigned to fields on `root.Config` that all subcommand
configs inherit.

To use this pattern, first add the shared fields and flags to `root.Config`:

```go
// cmd/root/root.go
type Config struct {
    Stdin   io.Reader
    Stdout  io.Writer
    Stderr  io.Writer
    Token   string         // added: shared API token flag value
    Client  *api.Client    // added: constructed between Parse and Run
    Flags   *ff.FlagSet
    Command *ff.Command
}

func New(stdin io.Reader, stdout, stderr io.Writer) *Config {
    var cfg Config
    // ... I/O assignment ...
    cfg.Flags = ff.NewFlagSet("<cli-name>")
    cfg.Flags.StringVar(&cfg.Token, 0, "token", "", "API token")
    // ... command construction ...
}
```

Then initialize the dependency in `cmd.go` between the two calls:

```go
// cmd/cmd.go Run(), between Parse and Run:
if err := r.Command.Parse(args, ff.WithEnvVarPrefix("<CLI_NAME>")); err != nil { ... }

client, err := api.NewClient(r.Token) // r.Token was populated during Parse
if err != nil {
    return fmt.Errorf("construct client: %w", err)
}
r.Client = client // now available to all exec functions via embedded root.Config

if err := r.Command.Run(ctx); err != nil { ... }
```

______________________________________________________________________

## Entry Point

`main.go` is intentionally thin. It sets up signal-safe shutdown via
`signal.NotifyContext` and delegates to a separate `run()` function that returns
an `int` exit code. `main()` calls `stop()` first, then `os.Exit`.

The `run()`/`main()` split is load-bearing: if `run()` called `os.Exit` directly,
`stop()` would never execute, leaving the signal-handler goroutine running until
the process terminated. `run()` must return an `int`, not call `os.Exit` itself.

`ff.ErrHelp` and `ff.ErrNoExec` are not failures. `root.ExitError` bypasses
the `"error: ..."` printer and maps to its own exit code.

```go
// main.go
package main

import (
    "context"
    "errors"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/peterbourgon/ff/v4"
    "<org>/<repo>/cmd"
    "<org>/<repo>/cmd/root"
)

const (
    exitFail    = 1
    exitSuccess = 0
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(),
        os.Interrupt,    // SIGINT = Ctrl+C
        syscall.SIGQUIT, // Ctrl-\
        syscall.SIGTERM, // polite termination request
    )
    code := run(ctx)
    stop()
    os.Exit(code)
}

// run is intentionally separated from main to improve testability. Please preserve this comment.
func run(ctx context.Context) int {
    err := cmd.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
    var exitErr root.ExitError
    switch {
    case err == nil, errors.Is(err, ff.ErrHelp), errors.Is(err, ff.ErrNoExec):
        return exitSuccess
    case errors.As(err, &exitErr):
        return int(exitErr)
    default:
        _, _ = fmt.Fprintf(os.Stderr, "error: %+v\n", err)
        return exitFail
    }
}
```

______________________________________________________________________

## Testing Commands

The injected `stdin`/`stdout`/`stderr` make commands testable without capturing
global state or spawning a subprocess. Pass `*bytes.Buffer` values instead of
the OS streams.

### Testing through `cmd.Run` (preferred)

```go
func TestEcho(t *testing.T) {
    var stdout, stderr bytes.Buffer

    err := cmd.Run(
        context.Background(),
        []string{"echo", "hello", "world"},
        strings.NewReader(""),
        &stdout,
        &stderr,
    )
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got := strings.TrimSpace(stdout.String()); got != "hello world" {
        t.Errorf("stdout = %q, want %q", got, "hello world")
    }
}
```

### Testing a command in isolation

When you want to test one command without constructing the full dispatcher:

```go
func TestPing(t *testing.T) {
    var stdout bytes.Buffer
    r := root.New(strings.NewReader(""), &stdout, io.Discard)
    ping.New(r)

    if err := r.Command.Parse([]string{"ping"}); err != nil {
        t.Fatalf("parse: %v", err)
    }
    if err := r.Command.Run(context.Background()); err != nil {
        t.Fatalf("run: %v", err)
    }
    if got := strings.TrimSpace(stdout.String()); got != "pong" {
        t.Errorf("stdout = %q, want %q", got, "pong")
    }
}
```

### Testing `root.ExitError`

```go
func TestLintFindsIssues(t *testing.T) {
    var stderr bytes.Buffer
    err := cmd.Run(context.Background(), []string{"lint", "/bad/path"},
        strings.NewReader(""), io.Discard, &stderr)

    var exitErr root.ExitError
    if !errors.As(err, &exitErr) {
        t.Fatalf("expected ExitError, got %v", err)
    }
    if int(exitErr) != 1 {
        t.Errorf("exit code = %d, want 1", int(exitErr))
    }
}
```

______________________________________________________________________

## No Package-Level Mutable State; No `init()` Functions

Do not use package-level mutable variables. Do not use `init()` functions.

Mutable globals make behaviour depend on which code ran first, cause tests to bleed state into one another, and introduce data races. `init()` adds a second hazard on top of that: it runs before any flag is parsed, its side effects cannot be suppressed in a test, and — because execution order is determined by the import graph — it is invisible to anyone reading `main`.

Both rules are mechanically enforced: `gochecknoglobals` flags unexpected package-level vars; `gochecknoinits` rejects any `init()` function.

### Where to initialize instead

Use the most deferred option that covers the dependency's scope:

1. **Inside `exec`** — resources needed by exactly one command. Runs only when that command is selected; never pays the cost otherwise.
2. **Between `Parse()` and `Run()` in `cmd/cmd.go`** — shared dependencies (API clients, loggers, DB connections) that require a parsed flag value to construct.
3. **In `New()` in `cmd/cmd.go`** — command registration and flag binding only. No I/O, no network, no filesystem.

### Permitted exceptions

The rule is not "no package-level vars" — it is "no *writeable* shared state." A package-level `var` is acceptable when the value is set once before the program's logic begins, is never reassigned, and cannot be expressed as a `const`, local variable, or constructor parameter. Five patterns meet this bar:

**1. Sentinel errors**

```go
var ErrNotFound = errors.New("not found")
```

`errors.Is` works by pointer identity — it walks the error chain comparing pointers, not values. A freshly allocated error would never match a stored sentinel, so the sentinel must be a stable package-level pointer. The `reassign` linter enforces that it is never reassigned. The global is load-bearing for the API contract, not a source of mutable state.

**2. Blank-identifier interface assertions**

```go
var _ SomeInterface = (*MyType)(nil)
```

`_` has no storage and cannot be read or written. This is a compile-time check: the build fails if `*MyType` ever drifts out of conformance with `SomeInterface`. `gochecknoglobals` correctly ignores it because it holds no runtime state.

**3. Link-time version injection**

```go
// Override at build time: go build -ldflags "-X <pkg>.Version=1.2.3"
var Version = "dev"
```

The linker's `-ldflags "-X <pkg>.Version=<val>"` writes directly into the symbol table. Neither a `const` nor a local variable has a symbol the linker can target. The variable is read-only at runtime — nothing in the program reassigns it — but the language has no way to express that as a `const`. This is the **one sanctioned mutable global in a climax command**.

**4. Pre-compiled regular expressions**

```go
var nameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
```

`*regexp.Regexp` is goroutine-safe after construction, so one instance can be shared across all callers without synchronization. Because compiling a pattern is expensive relative to matching, the standard Go idiom is to compile once at startup and reuse the result indefinitely. `MustCompile` rather than `Compile` is deliberate: a bad pattern panics immediately at startup rather than returning an error silently on the first production call.

**5. Embedded files**

```go
//go:embed template.tmpl
var tmpl string
```

The compiler resolves `//go:embed` at build time and writes file contents into the variable before `main` runs. The result is backed by static binary data and is never modified at runtime — effectively a read-only constant. The Go specification requires the directive to appear immediately above a package-level `var`; it cannot appear inside a function.

### The unifying principle

All five are cases where the language cannot express a constant as a `const`. Every one is set before the program's logic begins and never changes afterward. They are not loopholes — they are the precise boundary of "no *writeable* shared state."

______________________________________________________________________

## Code Generation

The `climax` tool scaffolds and extends applications that follow this spec.

### `climax init [FLAGS] [path]`

Creates a new application skeleton at `path` (default: `.`). The path must be
inside an existing Go module. Generated files, in order:

1. `main.go`
2. `cmd/cmd.go`
3. `cmd/<root-pkg>/<root-pkg>.go`
4. `cmd/version/version.go` (unless `--no-version`)

| Flag              | Default                                    | Description                                                                                              |
| ----------------- | ------------------------------------------ | -------------------------------------------------------------------------------------------------------- |
| `--name`          | last import path segment                   | `ff.Command.Name` for the root command (allows hyphens)                                                  |
| `--short`         | `"TODO: describe <name> here"`             | `ff.Command.ShortHelp` for the root command                                                              |
| `--long`          | _(omitted)_                                | `ff.Command.LongHelp` for the root command                                                               |
| `--root-pkg`      | `root`                                     | Go package name and file basename for the root config package                                            |
| `--env-prefix`    | `--name` uppercased, hyphens → underscores | Prefix for environment variable names, passed to `ff.WithEnvVarPrefix`. Mutually exclusive with `--no-env-prefix`. |
| `--no-env-prefix` | false                                      | Use `ff.WithEnvVars()` instead — env vars enabled without an application prefix. Mutually exclusive with `--env-prefix`. |
| `--no-version`    | false                                      | Skip generating `cmd/version/version.go`                                                                 |

### `climax add [FLAGS] <name> [path]`

Creates `cmd/<name>/<name>.go` and registers it in the dispatcher. `<name>`
must be a valid Go identifier (used as the package name). `[path]` must be the
root of an existing climax application.

| Flag             | Default                      | Description                                              |
| ---------------- | ---------------------------- | -------------------------------------------------------- |
| `--name`         | same as `<name>`             | `ff.Command.Name` for the new command (allows hyphens)   |
| `--short`        | `"<name> command"`           | `ff.Command.ShortHelp` for the new command               |
| `--long`         | `"<Name> is a new command."` | `ff.Command.LongHelp` for the new command                |
| `-p`, `--parent` | root package                 | Go package name of the parent command (for nesting)      |

The CLI name used in `Usage` strings is resolved in this order:
1. `// climax:name` marker in the dispatcher
2. `Name` field of the root `ff.Command`, extracted via AST from `cmd/<root-pkg>/<root-pkg>.go`
3. Last segment of the import path

Registration uses AST analysis to locate the correct insertion point, so it
works even when marker comments have been removed or the file is named `command.go`.

`climax add` inspects the root `Config` struct before generating the exec stub:

- If `Stdout` and/or `Stderr` fields are present → stub uses `fmt.Fprintln(cfg.Stdout, ...)` and imports `"fmt"`
- If a logger field is detected (field name contains `"log"`) → stub uses `cfg.<LoggerField>.Info(...)`
- Otherwise → stub returns `nil` with no imports

### Persistence markers

`climax init` writes three marker comments to the dispatcher that carry values
needed by subsequent tool runs:

```go
// climax:name <cli-name>      // ff.Command.Name used for the root command
// climax:root-pkg <pkg>       // root config package name (default: root)
// climax:env-prefix <PREFIX>  // env var prefix; absent when --no-env-prefix was used
```

`climax:env-prefix` is present and holds the prefix string when env vars are
prefixed (the default). It is **absent — not present at all** — when `--no-env-prefix`
was passed at init time, which causes the generated code to use `ff.WithEnvVars()`
instead. The marker records the env-var configuration for human reference and
future tooling; it is read by the dispatcher analysis layer but is not currently
acted on by `climax lint` or `climax update`.

When all three markers are present, `climax add` uses text insertion at the marker
positions. When one or more are absent, it falls back to full AST analysis of
the dispatcher to determine insertion points — which requires that the file still
has a parenthesized import block, a top-level `func Run`, a root assignment
(`r := root.New(...)`), and a `r.Command.Parse(...)` call.

### `climax lint [path]`

Checks a climax application for structural drift from the current scaffold
templates. Issues are grouped by file; at most three can be reported — one per
structural group:

| Group | File | Properties checked |
|---|---|---|
| 1 | `main.go` | `signal.NotifyContext`, separate `run()` function, `os.Stdin` passed to `cmd.Run` |
| 2 | `cmd/cmd.go` | `stdin io.Reader` parameter in `Run`, `stdin` forwarded to `root.New` |
| 3 | `cmd/<root>/<root>.go` | `Stdin io.Reader` field in `Config`, `stdin io.Reader` parameter in `New`, `cfg.Stdin = stdin` assignment |

All three properties in a group must be present to suppress that group's issue.
Each issue is shown as a focused unified diff. Exits with status 1 when any issues
are found.

Success output:

```
✓  No structural drift found.
```

Failure output:

```
⚠  2 structural issue(s) found in /path/to/app

── main.go: signal-safe shutdown, run() separation, os.Stdin

   --- a/main.go
   +++ b/main.go	(expected per climax template)
   @@ structural pattern @@
   ...
```

### `climax update [--apply] [path]`

A **climax development tool** — detects drift between climax's own source files
and the scaffold template files in `pkg/scaffold/templates/`. Only runs against
the climax module itself (`github.com/StevenACoffman/climax`). Run after changing
a structural pattern in `main.go`, `cmd/cmd.go`, or `cmd/root/root.go`.

Fifteen structural properties are checked independently:

| File | Properties |
|---|---|
| `main.go` | `signal.NotifyContext`, `run()` separation, `os.Stdin` passed to `cmd.Run` |
| `cmd/cmd.go` | `stdin io.Reader` parameter in `Run`, `stdin` forwarded to `root.New`, `ff.WithEnvVarPrefix` in Parse call |
| `cmd/root/root.go` | `Stdin io.Reader` field, `stdin io.Reader` parameter in `New`, `cfg.Stdin = stdin` assignment |
| `cmd/version/version.go` | `JSON` flag, tabwriter output, `GetVersionInfoFrom` function, pointer receivers on `Info` methods, `Option` type, `With*` constructors |
| `cmd/mango/mango.go` | `Section int` field in `Config` |

Each drift item is one of:

- **`✗` auto-fixable** — source has the property, template lacks it; `--apply` patches it.
- **`✗` manual-needed** — source has the property, template lacks it, but no string-replacement patch exists (e.g. tabwriter output requires a broader template rewrite).
- **`⚠` manual review** — template has the property, source does not. Removing properties from templates is a deliberate decision; never auto-patched.

Without `--apply`, prints a drift report and exits non-zero. With `--apply`,
patches only the auto-fixable items in the template files in place.

### `climax version [--json]`

Prints build and version information for the `climax` binary itself, read from
the embedded module build info. Use `--json` for machine-readable output.

______________________________________________________________________

## Constraints

| Rule | Rationale |
|---|---|
| Every configurable knob is a registered flag | Hard-coded values and `os.Getenv` calls bypass `-h` discoverability and make configuration invisible to operators |
| Use `ff`; no other CLI frameworks | `ff` provides flags, subcommand dispatch, and help with minimal surface area |
| No `Commander` interface | Go composition via `Exec` function pointer is sufficient |
| No `init()` functions | `init()` runs unconditionally before flag parsing; effects cannot be suppressed in tests; execution order across packages is implicit. Register commands via `New()` in `cmd/cmd.go`; initialize resources inside `exec` when actually needed. See "No Package-Level Mutable State" above |
| No package-level mutable variables (five narrow exceptions apply) | Writeable globals cause data races, test pollution, and hidden coupling. Permitted: sentinel errors, blank-identifier assertions, `var Version`, `regexp.MustCompile(...)`, `//go:embed`. See "No Package-Level Mutable State" above |
| Config struct per command | Carries parsed flag values and inherited I/O; avoids global state |
| Flag values bound in `New()`, not in `exec` | Flags are parsed before `exec` is called; binding in `exec` is too late |
| `SetParent` on every subcommand flag set | Allows parent flags to be accepted at any subcommand depth |
| Never use `os.Stdout`/`os.Stderr` directly | Write to `cfg.Stdout`/`cfg.Stderr` for testability |
| Return `root.ExitError`, not `os.Exit` | Commands don't control the process; only `main()` calls `os.Exit` — after `stop()` releases the signal context |
| `run()` returns `int`; `main()` calls `stop()` then `os.Exit` | Calling `os.Exit` inside `run()` bypasses `stop()`, leaving the signal-handler goroutine running until the process terminates |
| Errors bubble to `main` | `run()` is the single place that maps errors to exit codes |
| `ff.ErrHelp` and `ff.ErrNoExec` are success | Handle both in `run()`'s switch; do not propagate as failures |
| `Name` is the dispatch key | It must match what the user types (case-insensitive) |
| `cmd` package is not a binary | It is a dispatcher package imported by `main`, not a standalone binary |

______________________________________________________________________

## Checklist: Adding a New Command

- [ ] Create `cmd/<name>/` package
- [ ] Define `Config` struct embedding `*root.Config` (or `*<parent>.Config` for nesting), with command-local flag value fields
- [ ] Write `New(parent *root.Config) *Config` that:
  - creates `ff.NewFlagSet("<name>").SetParent(parent.Flags)`
  - binds flag values to `Config` fields
  - exposes every configurable behaviour as a flag (no `os.Getenv`, no hard-coded values)
  - constructs `ff.Command` with `Name`, `Usage`, `ShortHelp`, `Flags`, and `Exec`
  - appends to `parent.Command.Subcommands`
- [ ] Write `func (cfg *Config) exec(ctx context.Context, args []string) error`
- [ ] Call `<name>.New(r)` in the registration block in `cmd/cmd.go`
- [ ] Add the import for the new package in `cmd/cmd.go`
