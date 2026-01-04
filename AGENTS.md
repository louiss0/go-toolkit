# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` holds Cobra command implementations (one file per command).
- `internal/` contains app-specific packages such as config, scaffold, and runner helpers.
- `custom_errors/` and `custom_flags/` provide reusable error and flag types.
- Test suites live alongside packages as `*_suite_test.go` and `*_test.go`.
- Entry points: `main.go` and `cmd/root.go`.

## Build, Test, and Development Commands
- `go test ./...`: run the full Go test suite across all packages.
- `ginkgo run`: run Ginkgo v2 suites (preferred for CLI/TDD flow).
- `ginkgo watch`: watch files and rerun tests on change.

## Coding Style & Naming Conventions
- Go code is formatted with `gofmt` (tabs for indentation, K&R braces).
- Use descriptive, intent-based names (commands as verbs, packages as nouns).
- Files in `cmd/` follow `New<Command>NameCmd` patterns and are registered in
  `cmd/root.go`.
- Prefer small, focused packages; keep mutations close to where state is
  created.

## Testing Guidelines
- Test framework: Ginkgo v2 with `testify/assert`.
- Suite files use `*_suite_test.go` (e.g., `internal/config/config_suite_test.go`).
- Prefer behavior-focused tests over implementation details.
- Ginkgo runs tests in parallel by default; avoid shared mutable state or guard
  it carefully.

## Commit & Pull Request Guidelines
- Commit format is required:
  `type(scope)!: subject` (subject ≤ 64 chars, imperative mood).
  Example: `feat(cmd)!: add search command`.
- Scope is mandatory, lowercase, and hyphenated if needed (e.g., `cmd`,
  `internal-config`).
- Keep commits to a single logical change and ensure tests pass.
- PRs should include a concise description, test notes, and linked issues when
  applicable.

## Agent Notes
- This repository is a Cobra-based CLI template. When adding a command, create
  a file in `cmd/`, expose `New<Command>NameCmd`, register it in `cmd/root.go`,
  and add tests in `cmd/`.
