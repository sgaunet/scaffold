# Code Patterns & Best Practices

## Error Handling
Sentinel errors in the core map to stable exit codes in the CLI layer.
```go
// internal/scaffold/scaffold.go — conflicts surface as a sentinel error
if plan.HasConflicts() {
    return fmt.Errorf("%w", ErrConflict)
}
// internal/cli/exit.go — ExitCode maps ErrConflict to exit code 10
```

## Testing Patterns
- Test file naming: `_test.go` colocated with the package under test.
- Prefer black-box tests (`package foo_test`) for the public surface; full
  end-to-end runs live in `test/integration/`.
- Tests exercise the CLI-free core directly, without cobra.

## Template Rendering
- `text/template` with custom delimiters: `template.New(name).Delims("[[", "]]")`.
- The same `renderString` renders both file bodies and destination paths (`DestTmpl`).
- Templates are embedded (`//go:embed all:templates`) — never read from disk at runtime.

## Go-Specific Patterns
- Static binaries: `CGO_ENABLED=0 go build -trimpath`.
- Stdlib-first: `text/template`, `embed`, `os.CreateTemp` + `os.Rename` for atomic writes.
- Detection without subprocesses: `golang.org/x/mod/modfile` for `go.mod`, a manual
  INI parse for `.git/config`.

## Common Utilities
- `internal/scaffold/registry.go`: `Applicable(profile)` — single source of truth
  for which templates apply.
- `internal/detect`: `FromGoMod`, `FromGitConfig`, `platformForHost`.
- `internal/cli/exit.go`: `ExitCode(err)` — error → exit-code mapping.
