# Repository Guidelines

## Project Structure & Module Organization

This repository contains a Go CLI application for reviewing git branches with local AI provider CLIs. The executable entrypoint is `cmd/uatiari/main.go`. Core application code is organized by responsibility under `internal/`: `app` handles argument parsing and orchestration, `config` manages configuration, `git` gathers repository context, `provider` and `provider/clirunner` run external AI CLIs, `review` builds review requests, `report` formats results, `skills` handles framework-specific review guidance, and `version` stores release metadata. Tests live next to the code they cover as `*_test.go`. Build and packaging helpers live in `scripts/`.

## Build, Test, and Development Commands

- `go test ./...`: runs the full Go test suite.
- `go build -o uatiari ./cmd/uatiari`: builds a local development binary at the repository root.
- `go run ./cmd/uatiari <branch> --provider=codex`: runs the CLI directly from source.
- `./scripts/package.sh`: builds a trimmed release binary and packages it as `uatiari-<os>-<arch>.tar.gz`.

The project targets Go `1.26.4` as declared in `go.mod`.

## Coding Style & Naming Conventions

Use standard Go formatting: run `gofmt` on changed Go files before submitting. Keep package names short, lowercase, and aligned with directory names. Prefer small structs and interfaces at package boundaries, as seen in `internal/provider`. Test functions should use Go’s `TestNameScenario` style, for example `TestParseUnknownOptionFails`.

## Testing Guidelines

Use the standard `testing` package. Add or update adjacent `*_test.go` files for changes in parsing, configuration, provider execution, review generation, reporting, or skills. Prefer table-driven tests when adding multiple cases, and use simple fakes for provider behavior instead of invoking real AI CLIs. Run `go test ./...` before opening a pull request.

## Commit & Pull Request Guidelines

Recent history uses conventional commit prefixes such as `feat:`, `refactor:`, and `chore:`. Keep commit subjects imperative and scoped to one logical change, for example `feat: add python skill detection`.

Pull requests should include a short description, the reason for the change, and the verification performed. Link related issues when available. For CLI behavior changes, include example commands or before/after output. For user-facing documentation changes, update `README.md` in the same PR when relevant.

## Security & Configuration Tips

Do not commit local provider credentials, machine-specific config, or generated release archives. Runtime provider selection can come from `--provider`, `~/.config/uatiari/config.toml`, or `UATIARI_PROVIDER`; keep tests deterministic by setting inputs explicitly.
