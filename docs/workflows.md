# Development Workflows

## Feature Development
1. Create a feature branch from `main` (this repo uses speckit feature branches,
   e.g. `001-scaffold-generator`).
2. Implement changes with colocated `*_test.go` tests.
3. Run the local gate: `task check-before-commit` (lint + test + build).
4. Push and open a PR.
5. Merge after review and green CI.

## Code Review Process
- No `CONTRIBUTING.md` yet; follow `docs/operating-guidelines.md`.
- All PRs should pass the automated `test` and `linter` workflows before merge.
- Apply the staff-engineer test before requesting review.

## Testing Strategy
- Unit tests: `*_test.go` colocated in `internal/` packages.
- Integration tests: `test/integration/` — black-box, end-to-end CLI runs.
- Run everything with `task test` (`go test ./...`).

## Release Process
- Automated via GoReleaser (`.goreleaser.yaml`) + GitHub Actions.
- `release.yml` triggers on tags; `snapshot.yml` builds unreleased snapshots;
  `test.yml` and `linter.yml` run on push / PR.
- Local equivalents: `task release`, `task snapshot`.
- Dependencies kept current by Dependabot (gomod + github-actions).
