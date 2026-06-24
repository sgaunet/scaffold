# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Operating Guidelines

**Read `docs/operating-guidelines.md` at the start of every session.** It
defines how to plan, verify, and iterate in this repository: plan mode,
subagent strategy, verification gates, self-improvement loop, and the
communication contract. Treat it as load-bearing context.

## Repository Overview

`scaffold` (`github.com/sgaunet/scaffold`) is a single static Go 1.25 CLI that
generates a Go project's tooling, CI, and config files from templates embedded
in the binary — no network at generation time. Built on cobra with stdlib
`text/template` + `embed` and `golang.org/x/mod/modfile`. The design is
constitution-driven (`.specify/memory/constitution.md`) and originated from the
spec under `specs/001-scaffold-generator/`.

## Architecture

- **CLI-thin / CLI-free core split**: `cmd/scaffold/main.go` only wires signals
  and maps errors to exit codes; `internal/cli` is the sole cobra consumer;
  `internal/scaffold` holds all logic and imports no CLI packages (Constitution II).
- **Predicate-based template registry**: templates live under
  `internal/scaffold/templates/`, embedded via one `//go:embed all:templates`.
  Each descriptor carries an `Applies(profile)` predicate, so `generate` and
  `list` share one source of truth.
- **Custom `[[ ]]` delimiters**: `text/template` uses `[[ ]]` instead of `{{ }}`
  so GoReleaser / GitHub Actions expressions pass through verbatim.
- **Plan-before-write**: `BuildPlan` classifies each file Create/Skip/Overwrite;
  skip-existing is the default, `--force` overwrites, conflicts return
  `ErrConflict` → exit code 10 (non-conflicting files are still written).
- **Process-free auto-detection**: `internal/detect` parses `go.mod` (modfile)
  and `.git/config` (INI) directly — no `go list`, no `git` subprocess.
- **Atomic writes + cancellation**: temp-file + rename per file; the write loop
  checks `ctx.Err()` so SIGINT/SIGTERM leaves no partial output.

See docs/architecture.md for detailed design decisions.

## Development Commands

```bash
# Build (static binary): CGO_ENABLED=0 go build -trimpath -o scaffold ./cmd/scaffold
task build

# Test (go test ./...)
task test

# Lint (golangci-lint run ./...)
task lint

# Local gate — run before committing (lint + test + build)
task check-before-commit

# Release / snapshot (goreleaser)
task release
task snapshot
```

## Code Quality Standards

**Linters configured** (do not duplicate rules):
- golangci-lint: see `.golangci.yml` (`default: all`, 27 linters disabled)
- GoReleaser: see `.goreleaser.yaml`
- pre-commit hooks (test / lint / build): see `.pre-commit-config.yaml`
- Dependabot (gomod + github-actions): see `.github/dependabot.yml`

## File Locations

- **Source**: `cmd/scaffold/` (entry), `internal/cli/`, `internal/scaffold/`, `internal/detect/`
- **Templates**: `internal/scaffold/templates/` (embedded in the binary)
- **Tests**: `*_test.go` colocated in `internal/`, plus `test/integration/`
- **Specs / design**: `specs/001-scaffold-generator/` (spec, plan, data-model, contracts)
- **Docs**: `docs/`
- **Config**: `.specify/`, root dotfiles

## Documentation

- docs/architecture.md: System design and component overview
- docs/workflows.md: Development processes and release flow
- docs/patterns.md: Code patterns and conventions
- docs/operating-guidelines.md: How to plan, verify, and iterate here

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:
`specs/002-config-interactive-prompts/plan.md`

Active feature: **Config-File Defaults & Interactive Setup** (`002-config-interactive-prompts`) —
adds two new ways to *supply* the existing generation inputs; generated files are unchanged.
- New CLI-free package `internal/config` (YAML schema + loader + precedence overlay; imports only
  stdlib + `gopkg.in/yaml.v3`). Extends `internal/cli/profilebuild.go` to insert the config tier:
  `env > config file > auto-detection > built-in defaults` (the per-input CLI-flag tier was
  removed in the 2026-06-24 amendment: `generate` is interactive-only).
- `generate` is interactive-only: no profile flags, no `-i`. It always runs the
  `charmbracelet/huh` form (renders on stderr, platform-first, conditional fields, pre-filled
  from resolved defaults, confirm-before-write) and requires a terminal (TTY-gated with
  `golang.org/x/term`; piped/`--quiet`/no-TTY → usage error, exit 2). `scaffold list` is the
  non-interactive preview. `generate`/`list` keep only `--config`/`--no-config` + globals.
- Constitution II boundary: `huh`/`cobra` only in `internal/cli`; `internal/scaffold` and
  `internal/config` stay CLI-free (assert via a consistency test).
- Config is read-only (FR-022); default path `$XDG_CONFIG_HOME/scaffold/config.yml` →
  `$HOME/.config/scaffold/config.yml`; `--config`/`SCAFFOLD_CONFIG`/`--no-config` control it.
- Design docs: `specs/002-config-interactive-prompts/{research,data-model,quickstart}.md`,
  `contracts/{cli.md,config.schema.md,config.example.yml}`.
- Prior feature (still in force): **Go Project Scaffolder** (`001-scaffold-generator`) — the core
  CLI-thin `internal/cli` over CLI-free `internal/scaffold`; `[[ ]]` template delimiters;
  skip-existing default (`--force`, exit 10 on conflict); detection in `internal/detect`.
<!-- SPECKIT END -->
