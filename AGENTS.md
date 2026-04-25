# Repository Guidelines

## Project Structure & Module Organization

This is a Go 1.26 Fiber v3 backend. Entrypoints live in `cmd/server` for the HTTP API and `cmd/cli` for Cobra commands. Application code is under `internal/`: `app` wires providers, `transport/http` contains routers, controllers, middleware, requests, and resources, `services` holds business integrations, `db` contains lazy SQL/GORM connections and seeders, and `domain` contains model and enum types. Config files are in `configs/`, migrations in `database/`, generated OpenAPI output in `docs/`, translations in `lang/`, static assets in `public/`, and repository-level tests in `tests/`.

## Build, Test, and Development Commands

- `make init` installs tools, syncs dependencies, and creates `.env` from `.env.example` when missing.
- `make dev` runs the API with Air if installed, otherwise `go run ./cmd/server`.
- `make run` starts the server directly.
- `make build` builds `build/fiber-starter`; `make build-cli` builds `build/fiber-starter-cli`.
- `make test` runs `go test -v ./...`.
- `make coverage` writes HTML and profile reports under `coverage/`.
- `make check` runs formatting, vet, lint, and tests.
- `make docs` regenerates OpenAPI/Swagger files in `docs/` for the Scalar UI.
- `make atlas-diff NAME=<name>` and `make atlas-apply` manage migrations; set `ENV=sqlite` as needed.

## Coding Style & Naming Conventions

Use idiomatic Go: tabs via `gofmt`, short package names, explicit error handling, and contextual error wrapping with `%w`. Keep generated files under `internal/db/gen` untouched unless regenerating. Prefer provider wiring through `internal/app/providers` over package globals. Run `make fmt` and `make lint`; lint includes `govet`, `staticcheck`, `errcheck`, `revive`, and `gosec`.

## Testing Guidelines

Tests use Go's standard `testing` package and Fiber's `app.Test` with `httptest` for HTTP behavior. Name tests `TestFeature_Behavior` or `TestFeatureDoesNotRegress`, and place cross-package regression tests in `tests/`. Add focused tests for middleware, controllers, CLI commands, database behavior, and security-sensitive changes.

## Commit & Pull Request Guidelines

Recent history uses short lowercase commits such as `fix`; keep commits concise and task-focused. Prefer clearer messages when possible, for example `fix request id logging` or `add sqlite migration`. Pull requests should include a summary, test results such as `make check`, linked issues, and screenshots or API examples when HTTP behavior or docs change. Mention migrations, new environment variables, and generated code.

## Security & Configuration Tips

Never hardcode secrets; use `.env` and keep `.env.example` updated. Validate HTTP and CLI inputs, avoid logging credentials or tokens, and run `make lint` before release changes because `gosec` is enabled. Review `git diff` before pushing, especially generated docs, migrations, and vendored dependency changes.
