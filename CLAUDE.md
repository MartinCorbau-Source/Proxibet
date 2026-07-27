# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Proxibet ("Pari entre pro") is a monorepo with three projects:

- **proxiback/** — Go 1.25 REST API (module `github.com/MartinCorbau-Source/proxibet/proxiback`).
- **proxifront/** — Angular 22 standalone frontend (Angular Material, Vitest).
- **ProxiBetApp/** — .NET MAUI 10 mobile app. See [ProxiBetApp/CLAUDE.md](ProxiBetApp/CLAUDE.md) for its design system (tokens under `Resources/Styles/`, reusable component layer under `Components/`).

They are wired together via `docker-compose.yml` at the repo root (services `database` = Postgres 17, `api` = the Go backend built from `proxiback/Dockerfile`). The frontend is not part of docker-compose and is run separately with the Angular CLI. `ProxiBetApp/` is fully independent (no docker-compose wiring), run via Visual Studio / `dotnet build`.

## Commands

### Backend (`proxiback/`)

Run via `make` from the repo root, or `cd proxiback` and use `go` directly.

- `make run` — run the API locally with Go (sets `APP_ENV`, `HTTP_PORT=8080`, `DATABASE_URL` for local Postgres).
- `make build` — compile a static binary to `proxiback/bin/proxibet-api`.
- `make test` — `go test ./...`
- `make test-race` — tests with the race detector.
- `make test-cover` / `make coverage-html` — coverage report (text / HTML).
- Single test: `go test ./tests/internal/user/... -run TestName` (or target any package path directly).
- `make format` — `go fmt ./...`
- `make vet` — `go vet ./...`
- `make check` — format + vet + test, the standard pre-commit check.
- `make docker-up` / `make docker-down` / `make docker-restart` — run API + Postgres via docker compose.
- `make database-up` — start only Postgres.
- `make database-shell` — open `psql` in the running Postgres container.

Local dev DB URL (used by `make run`): `postgres://proxibet:proxibet@localhost:5432/proxibet?sslmode=disable`.

### Frontend (`proxifront/`)

Run from inside `proxifront/`:

- `npm start` (or `ng serve`) — dev server at `http://localhost:4200/`.
- `npm run build` — production build to `proxifront/dist/`.
- `npm test` (or `ng test`) — unit tests via Vitest.
- `npm run lint` / `npm run lint:fix` — ESLint (`angular-eslint` + `typescript-eslint`).
- `npm run format` / `npm run format:check` — Prettier.

## Architecture

### Backend

- Standard library `net/http` with `http.ServeMux` (Go 1.22+ method+path patterns, e.g. `"POST /api/v1/auth/register"`) — no web framework.
- Feature-based package layout under `internal/`: `internal/auth` (handler/service/DTOs) and `internal/user` (model, repository interface + Postgres implementation, validation, normalization, domain errors).
- Layering: HTTP `handler` → `Service` (business logic) → `user.Repository` interface, backed by `user.NewPostgresRepository` (pgx/v5 pool). Handlers translate domain errors (e.g. `user.ErrEmailAlreadyInUse`, `auth.ErrInvalidEmailFormat`) into HTTP status codes + JSON error bodies (`{"error": CODE, "message": "..."}`) — see the `switch`/`errors.Is` dispatch in `internal/auth/handler.go`.
- `cmd/api/main.go` wires everything: reads `DATABASE_URL` from env, builds the pgx pool, constructs repository → service → handler, registers routes on the mux, and starts the server on `:8080`. There's a `GET /health` endpoint.
- Passwords are hashed with bcrypt (`golang.org/x/crypto/bcrypt`).
- DB schema managed via plain SQL migrations in `migrations/` (`NNNNNN_description.up.sql` / `.down.sql`), currently just `000001_create_users`.
- JWT-related env vars are already defined in `docker-compose.yml` (`JWT_SECRET`, `ACCESS_TOKEN_DURATION`, `REFRESH_TOKEN_DURATION`) for the login/JWT auth flow, but no login/JWT issuance code exists yet — only `POST /api/v1/auth/register` is implemented so far.
- Go tests live in a separate `tests/` tree mirroring `internal/` (e.g. `tests/internal/user/validate_test.go`), not colocated `_test.go` files.

### Frontend

- Angular 22, **fully standalone** (no NgModules anywhere) — components declare their own `imports: []`, bootstrapped via `bootstrapApplication(App, appConfig)` in `main.ts`.
- Routing (`src/app/app.routes.ts`) uses lazy `loadComponent` for every route; no guards exist yet.
- Component convention: `src/app/pages/<name>/<name>.component.{ts,html,scss,spec.ts}`, class name `<Name>Component` (the root component is the sole exception: class `App`, selector `app-root`).
- No `services/`, `guards/`, or `environments/` folders exist yet — there's no HTTP client wrapper, no auth service/guard, and no environment-based API URL configuration. These are all greenfield additions when frontend/backend integration work starts.
- Test runner is **Vitest** (`@angular/build:unit-test` builder), not Jasmine/Karma — specs use Vitest-style `describe/it/expect` globals.
