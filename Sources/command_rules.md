# Go CLI Command Pattern Spec

Three patterns implement CLI dispatch in this codebase. All share the same
architectural rules. Implementation details differ based on dependencies.

**How to choose:** check `go.mod`.

- `github.com/spf13/cobra` present → use **Pattern C: Cobra**
- `github.com/peterbourgon/ff/v4` present → use **Pattern B: ff/v4**
- Neither → use **Pattern A: stdlib**

Check for Cobra first — a project may have both `cobra` and `ff/v4`, and Cobra
takes precedence because it owns the command tree.

Replace `<org>/<repo>` throughout with the Go module path from `go.mod`.

---

## Shared Architecture Rules

These rules apply to **all three** patterns without exception.

| Rule | Rationale |
|---|---|
| Match the framework already in `go.mod`; do not introduce a new one | Three patterns exist because projects adopt different frameworks; the spec follows the project, not the other way around |
| No `init()` functions | `init()` runs at startup before flag parsing, cannot be suppressed in tests, and has implicit cross-package ordering; register in the dispatcher and initialize resources in `exec` when they are needed |
| One command per package (or file in the flat Cobra layout) | Keeps packages focused; factory function is the only public API |
| Return `error`; never call `os.Exit` in a command | Only `main` controls exit codes |
| `package cmd` is not a binary | It is a dispatcher package imported by `main` |
| Dispatch key must be unique across all commands | Duplicate keys silently shadow each other |
| Error strings: lowercase, no trailing punctuation | Format: `<command>: <reason>` |

---

## Shared Directory Skeleton

All three patterns share this top-level skeleton. Each pattern section below
describes what it adds to it.

```
/
├── main.go                        # Entry point only — no logic
├── go.mod
└── cmd/
    ├── cmd.go                     # Dispatcher / Execute function (package cmd)
    └── <name>/
        └── <name>.go              # One package per command
```

---

## Pattern A: stdlib

Use when neither `cobra` nor `ff/v4` is in `go.mod`.

### Additional directory

```
└── pkg/
    └── pattern/
        └── command/
            └── base.go            # Command struct — do not modify
```

### Base Type

```go
// pkg/pattern/command/base.go
package command

type Command struct {
    // Run holds the command implementation.
    // args are the arguments after the command name —
    // executable and command name are already stripped.
    Run func(cmd *Command, args []string) error

    // UsageLine is the one-line usage message and the dispatch key.
    // Must exactly match what the user types on the CLI.
    UsageLine string

    // Short is shown in help output next to the command name.
    Short string

    // Long is shown in 'help <command>' output.
    Long string
}
```

Do not modify this struct.

### Command Template

Each command package exports exactly one factory function and keeps the
implementation unexported.

```go
// cmd/<name>/<name>.go
package <name>

import (
    "github.com/<org>/<repo>/pkg/pattern/command"
)

// <Name>Command is the only exported symbol in this package.
func <Name>Command() *command.Command {
    return &command.Command{
        UsageLine: "<name>",
        Short:     "<one-line description>",
        Long:      "<paragraph description>",
        Run:       <name>Cmd,
    }
}

// <name>Cmd holds the implementation. It is unexported.
func <name>Cmd(_ *command.Command, args []string) error {
    // implementation
    return nil
}
```

### Rules

- The factory function is the **only exported symbol** in the package.
- Logic goes in the unexported `<name>Cmd` — not in the factory.
- `UsageLine` is the dispatch key. It must exactly match what the user types.
- Use `flag.NewFlagSet` scoped inside `<name>Cmd` for flags. Never use the
  global `flag` package or bind flags outside the command function.
- Always pass `flag.ContinueOnError` to `flag.NewFlagSet`. `flag.ExitOnError`
  calls `os.Exit` on parse failure, bypassing all error handling.
- When `fs.Parse` returns `flag.ErrHelp` (user passed `-h`), return it
  **unwrapped** so `main` can treat it as success. Wrap all other parse errors.
- With `flag.ContinueOnError`, the `flag` package automatically prints usage
  to `fs.Output()` (stderr by default) when `-h` is passed or an unknown flag
  is given. Do **not** add a redundant `fs.Usage()` call.
- The `cmd *command.Command` parameter in `Run` / `<name>Cmd` carries the
  command's own metadata (`UsageLine`, `Short`, `Long`). It is rarely needed
  in practice but is part of the signature — ignore it with `_` when unused.

### Concrete Examples

#### Command with no flags

```go
// cmd/version/version.go
package version

import (
    "fmt"

    "github.com/<org>/<repo>/pkg/pattern/command"
)

// Version is set at link time:
//
//	go build -ldflags "-X 'github.com/<org>/<repo>/cmd/version.Version=1.2.3'"
//
// var (not const): the Go linker can only override package-level variables via
// -ldflags -X. This is the one permitted mutable global in a command package.
var Version = "dev"

func VersionCommand() *command.Command {
    return &command.Command{
        UsageLine: "version",
        Short:     "print version information",
        Long:      "Prints version information for the application.",
        Run:       versionCmd,
    }
}

func versionCmd(_ *command.Command, args []string) error {
    fmt.Println("version " + Version)
    return nil
}
```

#### Command that reads positional arguments

```go
// cmd/echo/echo.go
package echo

import (
    "fmt"
    "strings"

    "github.com/<org>/<repo>/pkg/pattern/command"
)

func EchoCommand() *command.Command {
    return &command.Command{
        UsageLine: "echo",
        Short:     "echo arguments",
        Long:      "Prints all provided arguments joined by spaces.",
        Run:       echoCmd,
    }
}

func echoCmd(_ *command.Command, args []string) error {
    if len(args) == 0 {
        return fmt.Errorf("echo: no arguments provided")
    }
    fmt.Println(strings.Join(args, " "))
    return nil
}
```

#### Command with flags

Use `flag.NewFlagSet` scoped to the command function. Parse `args` with it;
use `fs.Args()` for remaining positional arguments after flag parsing.

```go
// cmd/shout/shout.go
package shout

import (
    "errors"
    "flag"
    "fmt"
    "strings"

    "github.com/<org>/<repo>/pkg/pattern/command"
)

func ShoutCommand() *command.Command {
    return &command.Command{
        UsageLine: "shout",
        Short:     "print arguments in uppercase",
        Long:      "Prints arguments in uppercase. Use -n to suppress newline.",
        Run:       shoutCmd,
    }
}

func shoutCmd(_ *command.Command, args []string) error {
    fs := flag.NewFlagSet("shout", flag.ContinueOnError)
    noNewline := fs.Bool("n", false, "suppress trailing newline")
    if err := fs.Parse(args); err != nil {
        if errors.Is(err, flag.ErrHelp) {
            return flag.ErrHelp // return unwrapped so main treats it as success
        }
        return fmt.Errorf("shout: %w", err)
    }
    out := strings.ToUpper(strings.Join(fs.Args(), " "))
    if *noNewline {
        fmt.Print(out)
    } else {
        fmt.Println(out)
    }
    return nil
}
```

### Dispatcher — `cmd/cmd.go`

`cmd/cmd.go` is the **only place commands are registered**. Add each new
command to the `commands` slice. Do not register commands anywhere else.

```go
// cmd/cmd.go
package cmd

import (
    "errors"
    "fmt"

    "github.com/<org>/<repo>/cmd/version"
    // add new command imports here
    "github.com/<org>/<repo>/pkg/pattern/command"
)

// Run dispatches args to the matching command.
// args must not include the executable name (pass os.Args[1:]).
func Run(args []string) error {
    commands := []*command.Command{
        version.VersionCommand(),
        // register new commands here
    }

    m := make(map[string]*command.Command)
    for i := range commands {
        m[commands[i].UsageLine] = commands[i]
    }

    if len(args) == 0 || (len(args) == 1 && args[0] == "help") {
        fmt.Println("Available Commands:")
        for i := range commands {
            fmt.Printf("%s - %s\n", commands[i].UsageLine, commands[i].Short)
        }
        return nil
    }

    cmd := m[args[0]]
    if cmd == nil {
        return errors.New(args[0] + ": invalid command")
    }
    // Strip the command name before dispatching.
    return cmd.Run(cmd, args[1:])
}
```

**Dispatch rules:**
- `Run` receives `os.Args[1:]` — executable name already removed by `main`.
- Lookup is by exact `UsageLine` match via map. No prefix matching, no fuzzy matching.
- Unknown command returns an error; `main` owns the exit code.

### Entry Point — `main.go`

```go
// main.go
package main

import (
    "errors"
    "flag"
    "fmt"
    "os"

    "github.com/<org>/<repo>/cmd"
)

const (
    exitFail    = 1
    exitSuccess = 0
)

func main() {
    err := cmd.Run(os.Args[1:])
    switch {
    case err == nil, errors.Is(err, flag.ErrHelp):
        os.Exit(exitSuccess)
    default:
        _, _ = fmt.Fprintf(os.Stderr, "error: %+v\n", err)
        os.Exit(exitFail)
    }
}
```

### Pattern A Checklist: Adding a New Command

- [ ] Create `cmd/<name>/` package
- [ ] Write exported `<Name>Command() *command.Command` factory function
- [ ] Write unexported `<name>Cmd(_ *command.Command, args []string) error`
- [ ] Set a unique `UsageLine` (this is the CLI verb, exact match required)
- [ ] If the command has flags: use `flag.NewFlagSet("name", flag.ContinueOnError)`
- [ ] If the command has flags: return `flag.ErrHelp` unwrapped; wrap all other parse errors
- [ ] Use `fs.Args()` for remaining positional arguments after flag parsing
- [ ] Add `<name>.<Name>Command()` to the `commands` slice in `cmd/cmd.go`
- [ ] Add the import for the new package in `cmd/cmd.go`

### Testing — Pattern A

Test the unexported command function directly from a same-package test file
(`package <name>`, not `package <name>_test`). Output goes to real stdout and
is not capturable without redirecting `os.Stdout`; focus tests on error
conditions and return values.

```go
// cmd/echo/echo_test.go
package echo

import "testing"

func TestEchoCmd_noArgs(t *testing.T) {
    cmd := EchoCommand()
    err := echoCmd(cmd, nil)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}
```

For commands with flags, test the flag-error path too:

```go
// cmd/shout/shout_test.go
package shout

import (
    "errors"
    "flag"
    "testing"
)

func TestShoutCmd_helpFlag(t *testing.T) {
    cmd := ShoutCommand()
    err := shoutCmd(cmd, []string{"-h"})
    if !errors.Is(err, flag.ErrHelp) {
        t.Fatalf("expected flag.ErrHelp, got %v", err)
    }
}
```

---

## Pattern B: ff/v4

Use when `github.com/peterbourgon/ff/v4` **is** in `go.mod` (and `cobra` is not).

This pattern adds:

- A `root.Config` that holds shared I/O writers and flags inherited by all commands
- `ff.FlagSet` with `SetParent` so parent flags (e.g. `--verbose`) are accepted at any level
- A `Config` struct per command that embeds `*root.Config`
- A post-parse initialization window in `cmd/cmd.go` for API clients, DB connections, loggers
- `context.Context` threading through dispatch to every command
- `ff.ErrHelp` treated as success by `run()`; `ff.ErrNoExec` absorbed by the dispatcher

### Additional directory structure

```
└── cmd/
    └── root/                      # Default name; configurable via climax init --root-pkg
        └── root.go                # root.Config: shared I/O, flags, root ff.Command, ExitError
```

### `ff.Command` Fields

All commands use `ff.Command` from `github.com/peterbourgon/ff/v4`.

```go
type Command struct {
    // Name is the dispatch key. Subcommand selection is case-insensitive.
    // Required.
    Name string

    // Usage is the one-line syntax string shown at the top of help output.
    // Example: "<cli-name> snapshot [FLAGS] <ARG>"
    // Recommended.
    Usage string

    // ShortHelp is shown next to the command name in parent help output.
    // Recommended.
    ShortHelp string

    // LongHelp is shown in the command's own help output.
    // Optional.
    LongHelp string

    // Flags is the ff.FlagSet for this command. Constructed and bound in New().
    // If nil, an empty flag set is used automatically (--help still works).
    // Optional.
    Flags ff.Flags

    // Subcommands lists commands available under this one.
    // Each subcommand's New() appends itself here.
    // Optional.
    Subcommands []*Command

    // Exec is the terminal function called when this command is selected.
    // args are the positional arguments left over after flag parsing.
    // If nil, running this command returns ff.ErrNoExec.
    // Optional.
    Exec func(ctx context.Context, args []string) error
}
```

Do not call `Parse`, `Run`, or any method on `ff.Command` directly from command
packages. Those are called by the dispatcher in `cmd/cmd.go`.

### Root Config — `cmd/root/root.go`

`root.Config` holds shared I/O and any flags shared across all subcommands.
Subcommand configs embed `*root.Config` to inherit these.

```go
// Package root defines the root configuration for the CLI.
package root

import (
    "fmt"
    "io"

    "github.com/peterbourgon/ff/v4"
)

// ExitError is returned by commands that want a specific non-zero exit code
// without printing an additional error message. run() in main.go checks for
// ExitError with errors.As and calls os.Exit(int(e)) directly, bypassing the
// default "error: ..." printer.
type ExitError int

func (e ExitError) Error() string { return fmt.Sprintf("exit status %d", int(e)) }

// Config holds shared I/O writers and the root ff.Command.
// All subcommand configs embed *Config to inherit these.
type Config struct {
    Stdin   io.Reader
    Stdout  io.Writer
    Stderr  io.Writer
    Flags   *ff.FlagSet
    Command *ff.Command
    // Add shared dependencies here (API clients, loggers, DB connections)
    // after Parse() and before Run() in cmd/cmd.go.
}

// New returns a new root Config with the given I/O writers.
func New(stdin io.Reader, stdout, stderr io.Writer) *Config {
    var cfg Config
    cfg.Stdin = stdin
    cfg.Stdout = stdout
    cfg.Stderr = stderr
    // No shared flags — cfg.Flags is nil; ff provides --help automatically.
    // Subcommands call SetParent(parent.Flags)
    // which is a no-op here; add shared flags (e.g. BoolVar) to activate.
    // To add shared flags, uncomment and bind before constructing the command:
    // cfg.Flags = ff.NewFlagSet("<cli-name>")
    // cfg.Flags.BoolVar(&cfg.MyFlag, 0, "my-flag", "", "description")
    cfg.Command = &ff.Command{
        Name:      "<cli-name>",
        Usage:     "<cli-name> <SUBCOMMAND> ...",
        ShortHelp: "<one-line description of the program>",
        Flags:     cfg.Flags,
    }
    return &cfg
}
```

If the application has no shared flags, omit `ff.NewFlagSet` and leave
`cfg.Flags` nil. `ff` creates an empty flag set automatically and `--help`
still works. This is the default generated by `climax init`; uncomment the
`ff.NewFlagSet` line when you add the first shared flag.

### ExitError

`root.ExitError` lets a command exit with a specific non-zero code without
printing an `error: ...` line to stderr. The dispatcher suppresses help output
for it, and `run()` in `main.go` returns the code to `main`, which then calls
`os.Exit`:

```go
// In any command's exec function:
if !ok {
    return root.ExitError(1) // exits 1; no "error:" line printed
}
return nil // exits 0
```

`run()` handles it via `errors.As` and returns the code as an int:

```go
case errors.As(err, &exitErr):
    return int(exitErr)
```

### Command Template

Each command lives in its own package. The package contains a `Config` struct,
an exported `New` factory, and an unexported `exec` method.

```go
// cmd/<name>/<name>.go
package <name>

import (
    "context"

    "github.com/peterbourgon/ff/v4"
    "<org>/<repo>/cmd/root"
)

type Config struct {
    *root.Config
    // command-local flag values declared here as fields
    Flags   *ff.FlagSet
    Command *ff.Command
}

func New(parent *root.Config) *Config {
    var cfg Config
    cfg.Config = parent
    cfg.Flags = ff.NewFlagSet("<name>").SetParent(parent.Flags)
    // bind flags: cfg.Flags.StringVar(&cfg.SomeField, 0, "some-flag", "", "description")
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
    // TODO: implement <name>.
    // Rename the second parameter from _ to args to access positional arguments.
    // cfg.Stdin, cfg.Stdout, cfg.Stderr available via embedded root.Config.
    // Flag values available as cfg fields — already parsed before exec is called.
    return nil
}
```

### Rules

- `New` and `Config` are the **only exported symbols** in the package.
- `New` appends the command to `parent.Command.Subcommands` — no other registration needed.
- `Name` is the string matched (case-insensitively) when the user types a subcommand.
- Flag values are bound to `Config` fields in `New()`, **not** inside `exec`.
- `exec` reads already-parsed values. Never call `Parse` or `fs.Parse` inside `exec`.
- `SetParent(parent.Flags)` **must** be called on every subcommand flag set so that
  parent flags (e.g. `--verbose`) are accepted at any subcommand level.
  `SetParent(nil)` is safe — it is a no-op and `--help` still works.
- Every behavioural knob must be a registered flag on `cfg.Flags`. Never use hard-coded values,
  package-level variables, or `os.Getenv` calls outside the flag set — they make settings invisible to `-h`.
- Write to `cfg.Stdout` / `cfg.Stderr`. **Never** use `os.Stdout` / `os.Stderr` directly.
- Return `error`. Do not call `os.Exit` inside a command; use `root.ExitError` instead.
- Set `Exec: nil` (omit the field) on commands that are **group parents** — commands
  whose only purpose is to collect subcommands. When invoked without a subcommand they
  return `ff.ErrNoExec`, which the dispatcher handles by printing help and returning nil.

### Concrete Examples

#### Command with no flags

```go
// cmd/version/version.go
package version

import (
    "context"
    "encoding/json"
    "fmt"
    "runtime"
    "runtime/debug"
    "strings"
    "text/tabwriter"
    "time"

    "github.com/peterbourgon/ff/v4"
    "<org>/<repo>/cmd/root"
)

// Version is the application version string. When built from a tagged release
// or installed via "go install", the Go toolchain embeds the module version
// automatically, and it is read from build info at startup. Override at link
// time only if the auto-detected value is incorrect:
//
//	go build -ldflags "-X '<org>/<repo>/cmd/version.Version=v1.2.3'"
//
// var (not const): the Go linker can only override package-level variables via
// -ldflags -X — const and local variables are link-time immutable. This is the
// one permitted mutable global in a command package. Do not use init() to
// populate it; read build info inside exec, where it is actually needed.
var Version = "dev"

// versionInfo holds build and VCS metadata for structured output.
type versionInfo struct {
    GitVersion   string `json:"gitVersion"`
    GitCommit    string `json:"gitCommit"`
    GitTreeState string `json:"gitTreeState"`
    BuildDate    string `json:"buildDate"`
    GoVersion    string `json:"goVersion"`
    Compiler     string `json:"compiler"`
    Platform     string `json:"platform"`
}

// Config holds the configuration for the version command.
type Config struct {
    *root.Config
    JSON    bool
    Flags   *ff.FlagSet
    Command *ff.Command
}

// New creates and registers the version command with the given parent config.
func New(parent *root.Config) *Config {
    var cfg Config
    cfg.Config = parent
    cfg.Flags = ff.NewFlagSet("version").SetParent(parent.Flags)
    cfg.Flags.BoolVar(&cfg.JSON, 0, "json", "output version information as JSON")
    cfg.Command = &ff.Command{
        Name:      "version",
        Usage:     "<cli-name> version [--json]",
        ShortHelp: "print version information",
        LongHelp: `Print build and version information for this <cli-name> binary.

Fields shown:

  GitVersion    module version tag (e.g. v0.3.1) or "devel" for local builds
  GitCommit     VCS commit hash
  GitTreeState  "clean" or "dirty" (whether the working tree had uncommitted changes)
  BuildDate     timestamp of the VCS commit used for the build
  GoVersion     Go toolchain version (e.g. go1.23.0)
  Compiler      Go compiler name (usually "gc")
  Platform      GOOS/GOARCH pair (e.g. darwin/arm64)

Use --json to get machine-readable output suitable for scripting.`,
        Flags: cfg.Flags,
        Exec:  cfg.exec,
    }
    parent.Command.Subcommands = append(parent.Command.Subcommands, cfg.Command)
    return &cfg
}

func gatherVersionInfo(bi *debug.BuildInfo) versionInfo {
    const unknown = "unknown"
    info := versionInfo{
        GitVersion:   Version,
        GitCommit:    unknown,
        GitTreeState: unknown,
        BuildDate:    unknown,
        GoVersion:    runtime.Version(),
        Compiler:     runtime.Compiler,
        Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
    }
    if bi == nil {
        return info
    }
    if (info.GitVersion == "dev" || info.GitVersion == "") &&
        bi.Main.Version != "" && bi.Main.Version != "(devel)" {
        info.GitVersion = bi.Main.Version
    }
    for _, s := range bi.Settings {
        switch s.Key {
        case "vcs.revision":
            info.GitCommit = s.Value
        case "vcs.modified":
            switch s.Value {
            case "true":
                info.GitTreeState = "dirty"
            case "false":
                info.GitTreeState = "clean"
            }
        case "vcs.time":
            if t, err := time.Parse("2006-01-02T15:04:05Z", s.Value); err == nil {
                info.BuildDate = t.Format("2006-01-02T15:04:05")
            }
        }
    }
    return info
}

func (cfg *Config) exec(_ context.Context, _ []string) error {
    bi, _ := debug.ReadBuildInfo()
    info := gatherVersionInfo(bi)
    if cfg.JSON {
        b, err := json.MarshalIndent(info, "", "  ")
        if err != nil {
            return fmt.Errorf("version: %w", err)
        }
        _, _ = fmt.Fprintln(cfg.Stdout, string(b))
        return nil
    }
    var b strings.Builder
    w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
    _, _ = fmt.Fprintf(w, "GitVersion:\t%s\n", info.GitVersion)
    _, _ = fmt.Fprintf(w, "GitCommit:\t%s\n", info.GitCommit)
    _, _ = fmt.Fprintf(w, "GitTreeState:\t%s\n", info.GitTreeState)
    _, _ = fmt.Fprintf(w, "BuildDate:\t%s\n", info.BuildDate)
    _, _ = fmt.Fprintf(w, "GoVersion:\t%s\n", info.GoVersion)
    _, _ = fmt.Fprintf(w, "Compiler:\t%s\n", info.Compiler)
    _, _ = fmt.Fprintf(w, "Platform:\t%s\n", info.Platform)
    _ = w.Flush()
    _, _ = fmt.Fprint(cfg.Stdout, b.String())
    return nil
}
```

#### Command that reads positional arguments

```go
// cmd/echo/echo.go
package echo

import (
    "context"
    "fmt"
    "strings"

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

#### Command with flags

Flag values are bound to `Config` fields in `New()`. By the time `exec` is
called all flags are already parsed — do not call `Parse` again.

```go
// cmd/shout/shout.go
package shout

import (
    "context"
    "fmt"
    "strings"

    "github.com/peterbourgon/ff/v4"
    "<org>/<repo>/cmd/root"
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

#### Group parent command (no `Exec`)

A group parent collects subcommands but does nothing itself. Omit `Exec`.
When invoked without a subcommand the dispatcher prints the group's help.

```go
// cmd/cloud/cloud.go
package cloud

import (
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
    cfg.Flags = ff.NewFlagSet("cloud").SetParent(parent.Flags)
    cfg.Command = &ff.Command{
        Name:      "cloud",
        Usage:     "<cli-name> cloud <SUBCOMMAND> ...",
        ShortHelp: "cloud resource management",
        LongHelp:  "Manage cloud resources. Run 'cloud --help' for available subcommands.",
        Flags:     cfg.Flags,
        // No Exec field — this command is a group parent only.
    }
    parent.Command.Subcommands = append(parent.Command.Subcommands, cfg.Command)
    return &cfg
}
```

Register subcommands by passing `cfg` as the parent in `cmd/cmd.go`:

```go
// cmd/cmd.go
cloud := cloud.New(r)
instances.New(cloud)   // registers as subcommand of cloud
regions.New(cloud)     // same
```

`instances` and `regions` use `*cloud.Config` as their parent type, not
`*root.Config`. Their `New` signature becomes:

```go
func New(parent *cloud.Config) *Config
```

and they call `SetParent(parent.Flags)` as normal.

### Dispatcher — `cmd/cmd.go`

`cmd/cmd.go` is the **only place `New()` is called**. It constructs the root
config, calls each subcommand's `New()` in registration order (which controls
help output order), and exposes `Run` for `main`.

Registration happens inside `New()` via `append` to `parent.Command.Subcommands`.
Do not register commands anywhere else — no `init()`, no globals.

```go
// Package cmd is the dispatcher; it routes CLI arguments to the matching command.
package cmd
// climax:name <cli-name>
// climax:root-pkg root

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
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
    r := root.New(stdin, stdout, stderr)
    version.New(r)
    // register new commands here

    if err := r.Command.Parse(args); err != nil {
        fmt.Fprintf(stderr, "\n%s\n", ffhelp.Command(r.Command))
        return fmt.Errorf("parse: %w", err)
    }

    // Post-parse initialization: construct dependencies that require parsed flags.
    // Assign to fields on r so all exec functions can access them via embedded root.Config.
    //
    // Example:
    //   client, err := api.NewClient(r.Token)
    //   if err != nil { return fmt.Errorf("construct client: %w", err) }
    //   r.Client = client

    if err := r.Command.Run(ctx); err != nil {
        // Don't print usage help for ErrNoExec (no subcommand given) or
        // ExitError (command already reported its own outcome).
        var exitErr root.ExitError
        if !errors.Is(err, ff.ErrNoExec) && !errors.As(err, &exitErr) {
            fmt.Fprintf(stderr, "\n%s\n", ffhelp.Command(r.Command.GetSelected()))
        }
        return err
    }

    return nil
}
```

**Dispatch rules:**
- `Run` receives `os.Args[1:]` — executable name already removed by `run()` in `main.go`.
- Subcommand selection is **case-insensitive** match on `Name`. No prefix or fuzzy matching.
- `-h` / `--help` at any level causes `Parse` to return `ff.ErrHelp`; `run()` treats it as success.
- A command with no `Exec` (e.g. root invoked without a subcommand) returns `ff.ErrNoExec`; the **dispatcher** returns it as-is and `run()` treats it as success.
- Unknown subcommand returns an error; `run()` owns the exit code.

### Entry Point — `main.go`

`main.go` is intentionally thin. It sets up signal-safe shutdown via
`signal.NotifyContext` and delegates to a separate `run()` function.
`ff.ErrHelp`, `ff.ErrNoExec`, and `root.ExitError` are all handled in `run()`.

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

    "<org>/<repo>/cmd"
    "<org>/<repo>/cmd/root"
    "github.com/peterbourgon/ff/v4"
)

const (
    exitFail    = 1
    exitSuccess = 0
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(),
        os.Interrupt,    // interrupt = SIGINT = Ctrl+C
        syscall.SIGQUIT, // Ctrl-\
        syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
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

### Code Generation — `climax`

The `climax` tool scaffolds and extends applications that follow this pattern.

#### `climax init [FLAGS] [path]`

Creates a new application at `path` (default: current directory). The path must
be inside a Go module.

| Flag | Default | Description |
|---|---|---|
| `--name` | last import path segment | `ff.Command.Name` for the root command (allows hyphens) |
| `--short` | `"TODO: describe <name> here"` | `ff.Command.ShortHelp` for the root command |
| `--long` | _(omitted)_ | `ff.Command.LongHelp` for the root command |
| `--root-pkg` | `root` | Go package name and file basename for the root config package |
| `--no-version` | false | Skip generating `cmd/version/version.go` |

#### `climax add [FLAGS] <name> [path]`

Creates `cmd/<name>/<name>.go` and registers it in `cmd/cmd.go`. `<name>` must
be a valid Go identifier (used as the package name).

| Flag | Default | Description |
|---|---|---|
| `--name` | same as `<name>` | `ff.Command.Name` for the new command (allows hyphens) |
| `--short` | `"<name> command"` | `ff.Command.ShortHelp` for the new command |
| `--long` | `"<Name> is a new command."` | `ff.Command.LongHelp` for the new command |

#### Persistence markers

`climax add` reads two marker comments from `cmd/cmd.go` to reconstruct values
set at init time. These are written by `climax init` and **must not be removed**:

```go
// climax:name <cli-name>    // ff.Command.Name used for the root command
// climax:root-pkg <pkg>     // root config package name (default: root)
```

`climax add` also requires `// climax:imports` and `// register new commands here`
to locate injection points in `cmd/cmd.go`.

### Pattern B Checklist: Adding a New Command

- [ ] Create `cmd/<name>/` package
- [ ] Define `Config` struct embedding `*root.Config`, with command-local flag value fields
- [ ] Write `New(parent *root.Config) *Config` that:
  - creates `ff.NewFlagSet("<name>").SetParent(parent.Flags)`
  - binds flag values to `Config` fields (not in `exec`)
  - constructs `ff.Command` with `Name`, `Usage`, `ShortHelp`, `Flags`, and `Exec: cfg.exec`
  - appends to `parent.Command.Subcommands`
- [ ] Omit `Exec` field (leave nil) if the command is a group parent with only subcommands
- [ ] Write `func (cfg *Config) exec(ctx context.Context, args []string) error`
- [ ] Write to `cfg.Stdout` / `cfg.Stderr`; never `os.Stdout` / `os.Stderr`
- [ ] Expose every configurable behaviour as a flag (no `os.Getenv`, no hard-coded values)
- [ ] Use `root.ExitError` for controlled non-zero exits; do not call `os.Exit` in commands
- [ ] Call `<name>.New(r)` in the registration block in `cmd/cmd.go`
- [ ] Add the import for the new package in `cmd/cmd.go`

### Testing — Pattern B

Because `cmd.Run` accepts `io.Writer` parameters, all output is capturable.
Prefer integration tests via `cmd.Run` — they exercise the full parse → init →
exec path and require no access to unexported symbols.

```go
// cmd/cmd_test.go
package cmd_test

import (
    "bytes"
    "context"
    "errors"
    "io"
    "strings"
    "testing"

    "github.com/<org>/<repo>/cmd"
    "github.com/peterbourgon/ff/v4"
)

func TestShout(t *testing.T) {
    var stdout bytes.Buffer
    err := cmd.Run(context.Background(), []string{"shout", "hello", "world"}, strings.NewReader(""), &stdout, io.Discard)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got, want := stdout.String(), "HELLO WORLD\n"; got != want {
        t.Errorf("got %q, want %q", got, want)
    }
}

func TestNoSubcommand(t *testing.T) {
    // Root with no subcommand: dispatcher returns ff.ErrNoExec; run() treats as success.
    err := cmd.Run(context.Background(), []string{}, strings.NewReader(""), io.Discard, io.Discard)
    if err != nil && !errors.Is(err, ff.ErrNoExec) {
        t.Fatalf("expected nil or ErrNoExec, got %v", err)
    }
}
```

For unit-testing a single command's `exec` method, construct `root.New` with
buffer writers and call `exec` directly from a same-package test file
(`package <name>`):

```go
// cmd/shout/shout_test.go
package shout

import (
    "bytes"
    "context"
    "io"
    "strings"
    "testing"

    "<org>/<repo>/cmd/root"
)

func TestExec(t *testing.T) {
    var stdout bytes.Buffer
    r := root.New(strings.NewReader(""), &stdout, io.Discard)
    cfg := New(r)
    err := cfg.exec(context.Background(), []string{"hello", "world"})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got, want := stdout.String(), "HELLO WORLD\n"; got != want {
        t.Errorf("got %q, want %q", got, want)
    }
}
```

---

## Pattern C: Cobra

Use when `github.com/spf13/cobra` **is** in `go.mod`.

Cobra provides subcommand dispatch, persistent flags, argument validation, shell
completions, and help generation. It uses `github.com/spf13/pflag` for flag
parsing.

### Directory Structure

**Flat layout** (simple CLIs — all commands in one package):

```
└── cmd/
    ├── cmd.go                     # Execute function (package cmd)
    ├── root.go                    # newRootCommand constructor (package cmd)
    ├── version.go                 # version subcommand (package cmd)
    └── <name>.go                  # one file per subcommand (package cmd)
```

**Per-package layout** (larger CLIs with substantial per-command logic or tests):

```
└── cmd/
    ├── cmd.go                     # Execute function (package cmd)
    ├── root.go                    # newRootCommand constructor (package cmd)
    ├── version/
    │   └── version.go             # package version
    └── <name>/
        └── <name>.go              # package <name>
```

Use the flat layout for simple CLIs. Switch to per-package when a subcommand
has its own test file or significant implementation.

### `cobra.Command` Fields

The fields used in practice (the full struct is in Cobra source):

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

### Root Command — `cmd/root.go`

`newRootCommand` accepts I/O writers, sets them on the command tree, and
registers all subcommands. It is unexported; only `Execute` in `cmd/cmd.go`
calls it.

```go
// cmd/root.go
package cmd

import (
    "io"

    "github.com/spf13/cobra"
    "github.com/<org>/<repo>/cmd/version"
    // add subcommand imports here
)

// verbose is package-level so that PersistentPreRunE and any initialization
// code in the cmd package can read it without traversing the flag set.
// Subcommands in separate packages receive it via a getter function passed
// through their NewCommand constructor.
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

**Why `SilenceErrors: true`:** Prevents Cobra from printing the error before
returning it. `main` prints the error and controls the exit code.

**Why `SilenceUsage: true`:** Prevents Cobra from printing the full usage
template on every `RunE` error. Runtime failures (network, DB, permission
denied) are not usage mistakes. Call `cmd.Usage()` manually only before
returning errors that *are* usage mistakes (wrong argument count, invalid flag
value).

If the application has no shared persistent flags, omit the
`PersistentFlags()` block entirely.

### Command Template

Each subcommand exports exactly one function: `NewCommand`. Flag variables are
declared inside `NewCommand` (not at package level) so that each call produces
independent state — this prevents data races when tests run commands in
parallel. `RunE` is a closure over those variables.

```go
// cmd/<name>/<name>.go
package <name>

import (
    "github.com/spf13/cobra"
)

// NewCommand is the only exported symbol in this package.
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
- Never declare `var rootCmd *cobra.Command` at package level. Package-level
  command variables prevent isolated testing and can cause data races in
  parallel tests. Construct the root command inside `newRootCommand` and expose
  it only through `Execute`.
- Declare flag-destination variables **inside `NewCommand`** (not at package level).
  Closures capture them; each `NewCommand()` call gets independent state.
- Bind flags **in `NewCommand`**, not inside `RunE`. Flags are parsed before `RunE` runs.
- Use `RunE`, not `Run`. Commands that cannot fail should still return `nil`.
- Write to `cmd.OutOrStdout()` and `cmd.ErrOrStderr()`. Never use `os.Stdout` /
  `os.Stderr` directly.
- Set `Args` explicitly on every command. The default (`nil`) silently accepts
  any arguments.
- Obtain a context via `cmd.Context()` inside `RunE`. This is set by
  `ExecuteContext` in the dispatcher.
- Do not call `os.Exit` inside a command.

### Concrete Examples

#### Command with no flags

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

#### Command that reads positional arguments

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

#### Command with flags

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

#### Group parent (no `RunE`)

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

### Execute Function — `cmd/cmd.go`

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

### Hook Lifecycle

Hooks execute in this order for every selected command:

```
PersistentPreRunE  (closest ancestor that defines it — see warning below)
PreRunE            (current command only)
RunE               (current command only)
PostRunE           (current command only)
PersistentPostRunE (closest ancestor that defines it)
```

**Critical: the "closest ancestor" trap.** By default Cobra walks up the
command tree and calls only the *first* `PersistentPreRunE` it finds — it stops
there. If the root defines `PersistentPreRunE` for initialization and a
subcommand also defines its own `PersistentPreRunE`, the root's hook will
**not** run for that subcommand. This silently skips initialization and can
cause nil-pointer panics at runtime.

To run all ancestors' persistent hooks, set this once before `ExecuteContext`:

```go
// cmd/cmd.go — before root.ExecuteContext(ctx)
cobra.EnableTraverseRunHooks = true
```

As a simpler rule: **avoid defining `PersistentPreRunE` on subcommands** when
the root uses it for initialization. Reserve `PersistentPreRunE` for the root.

`PreRunE` and `PostRunE` are not inherited by children — they apply only to the
command that defines them.

### Post-Parse Initialization

Dependencies that require parsed persistent flag values (API clients, DB
connections, loggers) should be constructed in `PersistentPreRunE` on the root
command. This runs after all flag parsing, before any subcommand's `RunE`.

Store shared state in `cmd` package-level variables so subcommands in the same
package can read them directly. For subcommands in separate packages, pass a
getter function at construction time so `RunE` reads the value `PersistentPreRunE`
set — not the zero value at construction time:

```go
// cmd/root.go

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
    // Pass a getter — not the pointer — so RunE reads the post-init value.
    root.AddCommand(mysubcmd.NewCommand(func() *api.Client { return sharedClient }))
    return root
}
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

### Argument Validation

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
| `cobra.MatchAll(a, b, ...)` | All validators must pass |

For domain-specific validation, implement the validator directly:

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

### Showing Usage on Genuine Usage Errors

With `SilenceUsage: true` on the root, Cobra never prints usage automatically.
For errors that *are* usage mistakes (wrong argument count, invalid flag value),
print usage explicitly before returning. Do **not** call `cmd.Usage()` for
runtime errors (network, DB, permission denied).

```go
cmd.RunE = func(cmd *cobra.Command, args []string) error {
    if mode != "fast" && mode != "slow" {
        _ = cmd.Usage() // prints usage to cmd.ErrOrStderr()
        return fmt.Errorf("shout: invalid mode %q: must be fast or slow", mode)
    }
    // runtime work below — do NOT call cmd.Usage() on network/IO errors
    return nil
}
```

### Flag Reference

#### Local flags

Available only on the command they are defined on. Bind to variables declared
inside `NewCommand`:

```go
cmd.Flags().StringVarP(&field, "long-name", "s", "", "description")
cmd.Flags().BoolVarP(&field, "long-name", "b", false, "description")
cmd.Flags().IntVarP(&field, "long-name", "i", 0, "description")
```

#### Persistent flags

Inherited by all subcommands. Define on the root or on a group parent:

```go
root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
```

#### Required flags

```go
cmd.Flags().StringVarP(&region, "region", "r", "", "AWS region (required)")
cmd.MarkFlagRequired("region")

// Persistent variant:
root.PersistentFlags().StringVarP(&token, "token", "t", "", "API token (required)")
root.MarkPersistentFlagRequired("token")
```

#### Flag groups

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

### Entry Point — `main.go`

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

### Pattern C Checklist: Adding a New Command

- [ ] Create `cmd/<name>/` package (or `cmd/<name>.go` for flat layout)
- [ ] Write exported `NewCommand() *cobra.Command`
- [ ] Set `Use`, `Short`, and `Args` on the returned command
- [ ] Use `RunE` (not `Run`) as a closure assigned after construction
- [ ] Declare flag-destination variables **inside `NewCommand`** (not at package level)
- [ ] Bind flag values in `NewCommand` before assigning `RunE`
- [ ] Write to `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`; never `os.Stdout` / `os.Stderr`
- [ ] If the command depends on shared state: pass a getter function via the constructor
- [ ] Call `root.AddCommand(<name>.NewCommand())` in `newRootCommand` in `cmd/root.go`
- [ ] Add the import for the new package in `cmd/root.go`

### Testing — Pattern C

Because `Execute` accepts `io.Writer` parameters and `root.SetArgs` overrides
the OS args, all output is capturable. Because flag variables are declared inside
`NewCommand`, parallel tests have no shared state.

#### Integration test via `cmd.Execute`

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
    // Root with no RunE: Cobra prints help and returns nil.
    err := cmd.Execute(context.Background(), []string{}, io.Discard, io.Discard)
    if err != nil {
        t.Fatalf("expected nil, got %v", err)
    }
}
```

#### Unit test of a single subcommand

Construct the command directly, set I/O writers, then call `ExecuteContext`.
Flag variables are local to `NewCommand`, so each test call is isolated.

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

## Pattern Comparison

| Concern | Pattern A: stdlib | Pattern B: ff/v4 | Pattern C: Cobra |
|---|---|---|---|
| External dependency | None | `github.com/peterbourgon/ff/v4` | `github.com/spf13/cobra` + `github.com/spf13/pflag` |
| Command representation | Custom `Command` struct | `*ff.Command` | `*cobra.Command` |
| Flag binding | `flag.NewFlagSet` inside exec function | `ff.FlagSet` bound in `New()` before exec | `pflag` via `cmd.Flags()` / `cmd.PersistentFlags()` |
| Flag-value state | Local vars in `<name>Cmd` | Fields on `Config` struct | Local vars closed over inside `NewCommand` |
| Flag inheritance | Not supported | `SetParent(parent.Flags)` propagates to any depth | `PersistentFlags()` on any ancestor |
| `-h` / `--help` | `flag.ErrHelp` returned from command; treated as success in `run()` | `ff.ErrHelp` returned from `Parse`; treated as success in `run()` | Handled by Cobra automatically at any level |
| Top-level help | `help` as special positional arg in dispatcher | `--help` / `-h` at any level via `ff` | `--help` / `-h` at any level |
| No subcommand given | Dispatcher prints command list; returns nil | Returns `ff.ErrNoExec`; `run()` treats as success | Cobra prints root help; returns nil (when root has no `RunE`) |
| Group (parent-only) commands | Not directly supported | Omit `Exec`; returns `ff.ErrNoExec` → dispatcher prints help | Omit `RunE`; Cobra prints help on bare invocation |
| I/O | `fmt.Println` to real stdout | `cfg.Stdout` / `cfg.Stderr` via `root.Config` | `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` |
| Shared dependencies | Not applicable | Fields on `root.Config`; initialized post-parse | Package-level vars or getter functions; initialized in `PersistentPreRunE` |
| Post-parse init | Not applicable | Between `Parse()` and `Run()` in `cmd/cmd.go` | `PersistentPreRunE` on root command |
| Context | Not threaded through | `context.Context` argument in every `exec` | `cmd.Context()` inside `RunE`; set by `ExecuteContext` |
| Registration | Explicit `commands` slice in dispatcher | `append` to `Subcommands` inside `New()` | `root.AddCommand(...)` in root constructor |
| Dispatch key | Exact `UsageLine` match (case-sensitive) | `Name` field (case-insensitive) | First word of `Use` (case-sensitive) |
| Argument validation | Manual in `<name>Cmd` | Manual in `exec` | Built-in validators (`NoArgs`, `MinimumNArgs`, etc.) |
| Shell completions | Not built-in | Not built-in | Built-in (bash, zsh, fish, PowerShell) |
| Man page / doc generation | Not built-in | Not built-in | Built-in (markdown, man, yaml) |
| Testability | Lower — output not capturable without `os.Pipe` | Higher — `cmd.Run` accepts `io.Reader`/`io.Writer` | Higher — `Execute(ctx, args, stdout, stderr)` + per-command `cmd.ExecuteContext` |
| Signal handling | Manual | `signal.NotifyContext` in `main`; `run()` handles exit | Manual |
| Controlled exit without error message | `os.Exit` in command (bad) | `root.ExitError` returned from exec | `os.Exit` in command (bad) |
| Hook lifecycle | None | None | `PersistentPreRunE` / `PreRunE` / `RunE` / `PostRunE` (with traversal caveat) |
| Code scaffolding | None | `climax` (`climax init`, `climax add`) | None |
