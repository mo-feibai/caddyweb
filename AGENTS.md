# AGENTS.md

## Project Structure

Monorepo with two projects:
- **CaddyAPI** (Go/Gin) - port 8081, or set `PORT` env var
- **CaddyWeb** (Vue 3 + Vite + TypeScript) - port 8080

## Commands

### Backend
```bash
cd CaddyAPI && go run main.go    # Run server
cd CaddyAPI && go build          # Build binary
```

### Frontend
```bash
cd CaddyWeb && pnpm dev         # Dev server (--host for network access)
cd CaddyWeb && pnpm build       # Production build
cd CaddyWeb && pnpm preview    # Preview build
```

### Type-checking
```bash
npx vue-tsc --noEmit           # TypeScript check (frontend)
go build ./...                 # Go build (backend, no separate typecheck)
```

## Development Workflow

1. Start backend first: `cd CaddyAPI && go run main.go`
2. Then frontend: `cd CaddyWeb && pnpm dev`
3. Frontend proxies `/api` to `http://localhost:8081`

## Notes

- No test suites configured
- No pre-commit hooks
- Frontend uses Element Plus UI + Pinia state
- Backend uses gorilla/websocket for real-time