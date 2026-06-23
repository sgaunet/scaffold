# scaffold

A single static Go CLI that generates a Go project's tooling, CI, and config
files from templates embedded in the binary. Pick at most one forge platform and
an optional container toggle; existing files are skipped unless you pass
`--force`.

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
| `scaffold generate` (alias `gen`) | Generate tooling/CI/config files for a project |
| `scaffold list` | List the files that would be generated for the given options |
| `scaffold version` | Print version, commit, and build date |

## Usage

```bash
# Inside an existing Go repo: detect name/module/platform, generate GitHub tooling
scaffold generate

# Preview a Forgejo project that ships a container image — writes nothing
scaffold generate --platform forgejo --docker --dry-run

# Fully explicit, scriptable, JSON result on stdout, overwrite existing files
scaffold generate --name semver --module github.com/sgaunet/semver \
  --platform github --force --output json --quiet

# See what a gitlab+docker run would write, as JSON, without touching disk
scaffold list --platform gitlab --docker --output json
```

`--platform` is **optional**: with no platform (and none detected) only the
platform-independent baseline is generated — no CI or platform extras.

### What gets generated

- **Always**: `.goreleaser.yaml`, `mise.toml`, `.golangci.yml`,
  `.pre-commit-config.yaml`, `Taskfile.yml`, `Taskfile_dev.yml`
- **GitHub**: `.github/workflows/{linter,test,snapshot,release}.yml`,
  `.github/dependabot.yml`, `.github/FUNDING.yml`
- **Forgejo**: `.forgejo/workflows/{lint,test,snapshot,release}.yml`
- **GitLab**: `.gitlab-ci.yml`
- **`--docker`**: `Dockerfile` plus `dockers:`/`docker_manifests:` blocks and CI
  image build/publish steps
- **`--homebrew`** (GitHub only): a `homebrew_casks:` block in `.goreleaser.yaml`
  that publishes a Homebrew cask to `<owner>/<tap>` on release, plus the
  `HOMEBREW_TAP_TOKEN` secret wired into the release workflow

Every generated CI job provisions its toolchain through
[`mise`](https://mise.jdx.dev) rather than ad-hoc installs.

### Homebrew

With `--homebrew` (requires `--platform github`), releases publish a Homebrew
**cask** (the modern replacement for deprecated formula `brews`) to a tap repo
you own. Before your first release:

1. Create the tap repo `<owner>/homebrew-tap` (or pass `--homebrew-tap <name>`).
2. Add a `HOMEBREW_TAP_TOKEN` repository secret — a token with write access to
   the tap repo (a classic PAT with `repo` scope, or a fine-grained token
   scoped to the tap).

The cask installs the binary, generates bash/zsh/fish completions from the
binary's `completion` subcommand, and strips the macOS quarantine attribute for
unsigned binaries.

## Flags (`generate`)

| Flag | Default | Meaning |
|------|---------|---------|
| `--name` | go.mod module base, else dir name | project name |
| `--binary` | `--name` | binary/image name |
| `--module` | parsed from `go.mod` | module path |
| `--platform` | detected from `.git/config` origin | `github\|gitlab\|forgejo` (optional) |
| `--docker` | false | enable container support |
| `--homebrew` | false | publish a Homebrew cask on release (github only) |
| `--homebrew-tap` | `homebrew-tap` | tap repo name (with `--homebrew`) |
| `--owner` | derived from remote/module | FUNDING + registry owner |
| `--registry` | per-platform default | image base (docker only) |
| `--main` | `./cmd/<binary>` | main package dir |
| `--dir` / `-C` | `.` | target project directory |
| `--force` / `-f` | false | overwrite existing files |
| `--dry-run` | false | compute + print the plan; write nothing |
| `--yes` / `-y` | false | assume yes to interactive overwrite prompts |

Global: `--output text|json`, `--quiet`/`-q`, `--verbose`/`-v`, and `NO_COLOR`.
Precedence is **flags > env (`SCAFFOLD_*`) > detected > defaults**.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Generic failure |
| `2` | Usage error (bad flag, invalid name, unknown platform) |
| `10` | Conflict — one or more existing files were skipped; re-run with `--force` |

## Develop

```bash
task test          # unit + integration tests (black-box _test packages)
task lint          # golangci-lint
go test ./...      # equivalent to task test
```
