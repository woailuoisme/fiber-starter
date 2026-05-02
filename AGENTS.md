# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go 1.26 + Fiber v3 backend. HTTP entrypoints live in `cmd/server`, CLI entrypoints in `cmd/cli`. Application code follows a Laravel-style layout: `app/Http` for controllers, middleware, requests, and HTTP services; `app/Console` for commands and scheduling; `app/Models`, `app/Exceptions`, `app/Providers`, `app/Services`, and `app/Support` for domain, error handling, infrastructure, and shared helpers. Configuration lives in `config/`, database assets in `database/`, routes in `routes/`, generated OpenAPI output in `docs/`, and tests in `tests/`.

Versioned HTTP routes should go under `routes/v1/` and be mounted from `routes/api.go`. Keep response shaping in `app/Support/response.go` and middleware logic in `app/Http/Middleware/`.

## Build, Test, and Development Commands
- `make init` installs tools, syncs dependencies, and creates `.env` if missing.
- `make dev` starts Air when available, otherwise falls back to `go run ./cmd/server`.
- `make run` runs the server directly.
- `make test` runs `go test -v ./...`.
- `make check` runs `fmt`, `vet`, `lint`, and `test`.
- `make docs` regenerates OpenAPI 3.1 output for Scalar.
- `make atlas-diff NAME=<name>` and `make atlas-apply` manage migrations.
- `make k6-root` and `make k6-root-load` run k6 smoke/load tests for `/`.

## Coding Style & Naming Conventions
Use `gofmt` formatting, tabs, short package names, and explicit error handling with `%w` where wrapping is needed. Keep files focused on a single responsibility. Prefer descriptive names such as `auth_jwt.go`, `request_id.go`, and `response.go`. Do not edit generated outputs in `docs/`, `database/`, or vendored code unless regenerating them intentionally. Linting is handled by `.golangci.yml` with a low-noise set including `govet`, `staticcheck`, `errcheck`, `ineffassign`, `unused`, and `gosec`.

## Testing Guidelines
Tests use the `testing` package and Fiber’s `app.Test` helpers. Name tests `TestFeature_Behavior` or similar. Put regression tests in `tests/`. Run `go test ./...` before submitting changes, and cover middleware, controllers, queue behavior, storage, and response formatting as needed.

## Commit & Pull Request Guidelines
Keep commits short and focused, using lower-case messages such as `fix response format` or `add k6 smoke test`. Pull requests should include a concise summary, test results, and notes for config, migration, or API changes. Attach examples or screenshots when HTTP responses or docs change.

## Security & Configuration Tips
Never hardcode secrets. Use `.env` and keep `.env.example` synchronized. Validate all HTTP and CLI inputs, and review `make lint` output before merging. When adding new environment variables or external services, document the defaults and failure mode in the relevant config or README section.


<claude-mem-context>
# Memory Context

# [fiber-starter] recent context, 2026-05-03 12:11am GMT+8

No previous sessions found.
</claude-mem-context>