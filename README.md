# scaffold

> **Note:** this is a personal tool, built around my own workflow and
> conventions (my `mise`/`Taskfile`/`goreleaser` setup, my forge choices, my
> defaults). It's public and MIT-licensed, so feel free to fork it or use it
> as-is — but it isn't a community-driven project, and PRs bending it toward
> other workflows are unlikely to be merged.

A single static Go CLI that generates a Go project's tooling, CI, and config
files from templates embedded in the binary. Pick at most one forge platform and
an optional container toggle; existing files are skipped, and `generate` prompts
before overwriting any of them.

Every generated file ships inside the binary (no network at generation time) and
is a reviewable template in this repo under
[`internal/scaffold/templates/`](internal/scaffold/templates/).

## Install / build

```bash
CGO_ENABLED=0 go build -trimpath -o scaffold ./cmd/scaffold
# or, with the project's own tooling:
task build
```

## Commands

| Command | Purpose |
|---------|---------|
| `scaffold generate` (alias `gen`) | Guided setup form, then generate the files (requires a terminal) |
| `scaffold version` | Print version, commit, and build date |

## Usage

`scaffold generate` runs a short, guided form — **platform first**, then only the
questions relevant to your answers — and writes the files after you confirm.
There are no per-input flags: every prompt is pre-filled from your config file,
environment, and what the tool auto-detects, so usually you just press Enter
through it. Because it is interactive, `generate` **requires a terminal**.

```bash
# Inside an existing Go repo: walk through setup, accepting the detected defaults
scaffold generate

# Drive the proposed defaults from a specific config file
scaffold generate --config ./ci/scaffold.yml
```

The form starts with the platform choice, including **none** for the
platform-independent baseline only — in which case no CI or platform extras are
generated.

In a pipe, under `--quiet`, or with no terminal, `generate` exits with a usage
error (exit 2) instead of hanging. Set the defaults in the config file and run
`generate` in an interactive shell.

### Defaults come from a config file

There are no profile flags. Instead, a config file supplies the values proposed
in the form. Create
`~/.config/scaffold/config.yml` (honors `$XDG_CONFIG_HOME`):

```yaml
owner: sgaunet
platform: github      # github | gitlab | forgejo | none
docker: true
# Other recognized keys: name, binary, module, registry, main, homebrew,
# homebrew-tap, go-version, task-version, golangci-version, goreleaser-version
```

| Flag | Default | Meaning |
|------|---------|---------|
| `--config <path>` | `$XDG_CONFIG_HOME/scaffold/config.yml`, else `$HOME/.config/scaffold/config.yml` | config file to load (an explicit path that is missing → exit 2) |
| `--no-config` | false | ignore any config file (reproducible/CI runs) |

Global: `--output text\|json`, `--quiet`/`-q`, `--verbose`/`-v`, and `NO_COLOR`.
Precedence is **env (`SCAFFOLD_*`) > config file > auto-detection (`go.mod` /
`.git/config`) > built-in defaults**. `SCAFFOLD_CONFIG` sets the config path
(same as `--config`).

### What gets generated

- **Always**: `.goreleaser.yaml`, `mise.toml`, `.golangci.yml`,
  `.pre-commit-config.yaml`, `Taskfile.yml`, `Taskfile_dev.yml`
- **GitHub**: `.github/workflows/{linter,test,snapshot,release}.yml`,
  `.github/dependabot.yml`, `.github/FUNDING.yml`
- **Forgejo**: `.forgejo/workflows/{lint,test,snapshot,release}.yml`
- **GitLab**: `.gitlab-ci.yml`
- **Container support** (`docker: true`): `Dockerfile` plus
  `dockers:`/`docker_manifests:` blocks and CI image build/publish steps
- **Homebrew** (`homebrew: true`, GitHub only): a `homebrew_casks:` block in
  `.goreleaser.yaml` that publishes a Homebrew cask to `<owner>/<tap>` on
  release, plus the `HOMEBREW_TAP_TOKEN` secret wired into the release workflow

Every generated CI job provisions its toolchain through
[`mise`](https://mise.jdx.dev) rather than ad-hoc installs.

### Homebrew

With `homebrew: true` (requires `platform: github`), releases publish a Homebrew
**cask** (the modern replacement for deprecated formula `brews`) to a tap repo
you own. Before your first release:

1. Create the tap repo `<owner>/homebrew-tools` (or set `homebrew-tap: <name>`
   in the config).
2. Add a `HOMEBREW_TAP_TOKEN` repository secret — a token with write access to
   the tap repo (a classic PAT with `repo` scope, or a fine-grained token
   scoped to the tap).

The cask installs the binary, generates bash/zsh/fish completions from the
binary's `completion` subcommand, and strips the macOS quarantine attribute for
unsigned binaries.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Generic failure |
| `2` | Usage error (no terminal for `generate`, bad flag, invalid/missing config value) |
| `10` | Conflict — one or more existing files were skipped; re-run and confirm the overwrite prompt |

## Develop

```bash
task test          # unit + integration tests (black-box _test packages)
task lint          # golangci-lint
go test ./...      # equivalent to task test
```
