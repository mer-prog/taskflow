# CLAUDE.md — TaskFlow Project Instructions

## Project Overview

TaskFlow is a real-time team project management tool with Kanban boards.
This is a portfolio project demonstrating fullstack skills: Go + Next.js + PostgreSQL + Docker + AWS.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15 (App Router), TypeScript, Tailwind CSS, shadcn/ui, dnd-kit, next-intl, Zustand |
| Backend | Go 1.22+, Echo v4, gorilla/websocket, sqlc |
| Database | PostgreSQL 16 |
| Auth | JWT (access + refresh), bcrypt |
| Infra | Docker, docker-compose, AWS ECS Fargate, RDS, ALB |
| CI/CD | GitHub Actions |

## Repository Structure

taskflow/
├── CLAUDE.md
├── docker-compose.yml
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── service/
│   │   ├── repository/ (sqlc generated)
│   │   └── ws/
│   ├── db/
│   │   ├── migrations/
│   │   └── queries/
│   └── sqlc.yaml
├── frontend/
│   ├── src/app/[locale]/
│   ├── messages/{ja,en}.json
│   └── ...
└── docs/design.md

## Architecture Rules

### Backend (Go)
1. Layer separation: handler → service → repository. Handlers never access DB directly.
2. Error handling: Always return structured errors with codes. Never expose internal errors to client.
3. Multi-tenancy: Every query MUST include tenant_id filter. Never trust client-provided tenant_id — extract from JWT context.
4. sqlc: Write raw SQL in db/queries/*.sql, then run sqlc generate. Do NOT write manual SQL in Go code.
5. WebSocket: Use Hub pattern. REST handlers trigger broadcasts via Hub after DB mutation.
6. Context propagation: Pass context.Context through all layers.

### Frontend (Next.js)
1. App Router: Use Server Components by default. Add "use client" only when needed.
2. i18n: All user-facing text via next-intl. Keys in messages/{ja,en}.json.
3. State: Zustand for client state. No Redux, no Context for global state.
4. API calls: Use lib/api.ts wrapper with automatic token refresh.
5. Optimistic updates: For D&D moves, update store immediately, REST call after. Rollback on error.
6. Mobile-first: Design for 375px+. Use Tailwind responsive prefixes.

### Database
1. All primary keys are UUID v4. Never use auto-increment.
2. Always TIMESTAMPTZ, never TIMESTAMP.
3. Migrations: Always provide both up.sql and down.sql.

### Security
1. Passwords: bcrypt with cost 12.
2. JWT: Access token (15min) in memory. Refresh token (7d) in httpOnly cookie.
3. CORS: Whitelist frontend origin only.
4. Input validation at handler layer before passing to service.

### Git
- Branch: feature/{phase}-{feature-name}
- Commit: conventional commits (feat:, fix:, refactor:, docs:)

## Environment Variables

### Backend (.env)
PORT=8080
ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=taskflow
DB_PASSWORD=taskflow
DB_NAME=taskflow
DB_SSLMODE=disable
JWT_SECRET=your-secret-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
CORS_ORIGIN=http://localhost:3000

## Implementation Phases
Phase 1: Go API Foundation + DB + Auth — ✅ DONE
Phase 2: Tenant & Project CRUD — ✅ DONE
Phase 3: Board, Column, Task CRUD + Dashboard API — ✅ DONE
Phase 4: WebSocket Real-time Sync — ✅ DONE
Phase 5: Next.js Frontend Foundation + Auth UI — ✅ DONE
Phase 6: Kanban UI + Drag & Drop — ✅ DONE
Phase 7: Dashboard + Settings + Members UI — ✅ DONE
Phase 8: Docker + Deploy (Render + Vercel, AWS IaC only) — ✅ DONE
Phase 9: CI/CD + Polish

## Coding Conventions
### Go
- File names: snake_case.go
- Package names: single lowercase word
- Error wrapping: fmt.Errorf("service.CreateTask: %w", err)
- Struct tags: json:"field_name" db:"field_name"

### TypeScript
- File names: PascalCase.tsx for components, camelCase.ts for utils
- No any — use unknown and narrow
