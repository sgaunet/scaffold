# Architecture

## System Overview
`scaffold` is a single static Go CLI that generates a Go project's tooling, CI,
and config files from templates embedded in the binary. It detects project
context, builds a generation plan, then atomically writes the non-conflicting
files. The design follows a strict CLI-thin / CLI-free-core split mandated by
the project constitution (`.specify/memory/constitution.md`).

## Components
- **cmd/scaffold/main.go**: Entry point only — wires `signal.NotifyContext`
  (SIGINT/SIGTERM), runs the root command, maps the returned error to an exit code.
- **internal/cli**: The sole cobra consumer. `root`, `generate`, `list`,
  `version` commands; flag parsing, output formatting, exit-code mapping
  (`exit.go`), profile building from flags + detection.
- **internal/scaffold**: CLI-free core. `profile`, `registry`, `template`,
  `plan`, `render`, `writer`, `report`, `versions`, `embed`, `errors`.
- **internal/detect**: Process-free detection — `gomod.go` (module path via
  `golang.org/x/mod/modfile`), `gitremote.go` (origin URL + platform from `.git/config`).
- **internal/scaffold/templates/**: Embedded template tree (`//go:embed all:templates`).

## Design Decisions
1. **CLI-thin / CLI-free split (Constitution II)**: the core is fully testable
   without cobra; only `internal/cli` imports cobra.
2. **`[[ ]]` template delimiters**: avoid clashing with GoReleaser / GitHub
   Actions `{{ }}` expressions, which pass through verbatim.
3. **Predicate registry**: one `Applies(profile)` per template drives both
   `generate` and `list`, so new templates appear in both automatically.
4. **Skip-existing by default**: `--force` to overwrite; any skip → `ErrConflict`
   → exit code 10; non-conflicting files are still written.
5. **Atomic writes (Constitution VI)**: temp-file + `os.Rename`; `ctx.Err()` is
   checked between files so SIGINT cancels cleanly with no partial output.

## Integration Points
- No runtime network or subprocess calls. Reads `go.mod` and `.git/config` from disk.
- Output: tooling / CI / config files only (no application source).
- One forge platform per run (github / gitlab / forgejo), optional container toggle.

## Data Flow
flags + detection → `ProjectProfile` → `TemplateRegistry.Applicable(profile)` →
`BuildPlan` (Create / Skip / Overwrite) → `render` (`[[ ]]`) → `writer` (atomic
temp+rename) → `report`.
