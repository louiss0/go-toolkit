# Changelog
All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to
Semantic Versioning.

## [Unreleased]
### Added
- CLI commands for add/remove/init/test/scaffold/search/config workflows.
- Module index search backed by index.golang.org.
- TOML config support with global and repo-local resolution.
- Provider mappings for reading user.name from gitconfig files.
- Scaffold helpers for package folders, README, and optional module init.

### Changed
- Default test command runs `go test ./...`.
- Module path parsing now validates 1-3 segment inputs and fills defaults.

### Fixed
- Invalid input errors now use the custom InvalidInput error type.
