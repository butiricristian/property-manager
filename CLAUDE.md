# Property Manager v2

## Tech Stack
- Backend: Go (chi router, pgx, sqlc, golang-migrate)
- Frontend: React + TypeScript + Vite + shadcn/ui (v4, base-ui) + Recharts
- Database: PostgreSQL

## Common Commands
- `make run` — start Go server on :8080
- `make sqlc` — regenerate Go code from SQL queries
- `make migrate-up` — run database migrations
- `cd web && npm run dev` — start Vite dev server (port 5173, proxies /api to :8080)
- `cd web && npm run build` — build frontend to web/dist/
- `make build` — full production build (frontend + Go binary)
- `go build ./...` — check Go compilation

## Project Structure
- `cmd/server/main.go` — entry point
- `internal/config/` — env-based config
- `internal/database/migrations/` — SQL migration files
- `internal/database/queries/` — sqlc query files
- `internal/database/sqlc/` — generated code (DO NOT EDIT)
- `internal/models/` — domain types
- `internal/handlers/` — HTTP handlers
- `internal/services/` — business logic (dashboard, CSV import)
- `internal/middleware/` — logging, CORS
- `web/` — React frontend

## Notes
- shadcn/ui v4 uses @base-ui/react (NOT @radix-ui). No `asChild` prop — use `render={<Component />}` instead.
- Select `onValueChange` receives `(value: string | null)` — always handle null case.
- sqlc generates pgtype.UUID, pgtype.Numeric, pgtype.Date — use helpers in handlers/helpers.go for conversion.
- DATABASE_URL defaults to `postgres://postgres:postgres@localhost:5432/property_manager?sslmode=disable`
