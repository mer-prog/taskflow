[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-15-000000?logo=next.js)](https://nextjs.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Backend CI](https://github.com/mer-prog/taskflow/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/mer-prog/taskflow/actions/workflows/backend-ci.yml)
[![Frontend CI](https://github.com/mer-prog/taskflow/actions/workflows/frontend-ci.yml/badge.svg)](https://github.com/mer-prog/taskflow/actions/workflows/frontend-ci.yml)

# TaskFlow

> Real-time project management tool with Kanban boards

## Screenshots

| Dashboard | Kanban Board | Mobile View |
|-----------|-------------|-------------|
| ![Dashboard](docs/screenshots/dashboard.png) | ![Kanban](docs/screenshots/kanban.png) | ![Mobile](docs/screenshots/mobile.png) |

## Features

- Multi-tenant workspace management
- Drag & drop Kanban boards with real-time sync
- Role-based access control (Owner / Admin / Member / Viewer)
- WebSocket-powered live collaboration
- i18n support (English / Japanese)
- Project dashboard with task analytics
- JWT authentication with refresh token rotation
- Responsive design (mobile-first)

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15 (App Router), TypeScript, Tailwind CSS, shadcn/ui, dnd-kit, next-intl, Zustand |
| Backend | Go 1.22+, Echo v4, gorilla/websocket, sqlc |
| Database | PostgreSQL 16 |
| Auth | JWT (access + refresh), bcrypt |
| Infra | Docker, Render, Vercel, Terraform (AWS IaC) |
| CI/CD | GitHub Actions |

## Architecture

```
[Browser] ---- [Vercel (Next.js)] ---- [Render / AWS (Go API)] ---- [PostgreSQL]
                                              |
                                         WebSocket
```

### Demo Environment

| Component | Service |
|-----------|---------|
| Frontend | Vercel |
| Backend | Render (Docker) |
| Database | Render PostgreSQL |

### Production Environment (AWS IaC)

| Component | Service |
|-----------|---------|
| Frontend | Vercel |
| Backend | ECS Fargate (ALB + 2 tasks) |
| Database | RDS PostgreSQL |
| Network | VPC + public/private subnets |

See [infra/architecture.md](./infra/architecture.md) for details.

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- [sqlc](https://sqlc.dev/) (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### Local Development

```bash
# 1. Clone and start PostgreSQL
git clone https://github.com/mer-prog/taskflow.git
cd taskflow
docker compose up db -d

# 2. Run migrations
psql "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable" \
  -f backend/db/migrations/001_init.up.sql \
  -f backend/db/migrations/002_add_project_archived_at.up.sql \
  -f backend/db/migrations/003_add_column_color_task_priority.up.sql

# 3. (Optional) Load demo data
psql "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable" \
  -f backend/db/seed.sql

# 4. Start the backend
cp backend/.env.example backend/.env
cd backend && go run ./cmd/server

# 5. Start the frontend (in another terminal)
cd frontend
cp .env.example .env.local
npm install && npm run dev
```

### Demo Account

After running `seed.sql`, you can log in with:

| Email | Password | Role |
|-------|----------|------|
| `demo@taskflow.app` | `demo1234` | Owner |
| `alice@taskflow.app` | `demo1234` | Admin |
| `bob@taskflow.app` | `demo1234` | Member |

## Deployment

### Render + Vercel

The project includes a `render.yaml` Blueprint for one-click backend setup:

1. Connect this repo to [Render](https://render.com) — auto-provisions Web Service + PostgreSQL
2. Set `DATABASE_URL` in the `taskflow-env` env group
3. Run migrations against the Render database
4. Import `frontend/` to [Vercel](https://vercel.com) with env vars:
   - `NEXT_PUBLIC_API_URL` = `https://taskflow-api.onrender.com/api/v1`
   - `NEXT_PUBLIC_WS_URL` = `wss://taskflow-api.onrender.com/api/v1/ws`

### AWS (IaC only)

Terraform code in `infra/terraform/` — not actively deployed. See [infra/README.md](./infra/README.md).

```bash
cd infra/terraform
terraform init
terraform plan -var="db_password=..." -var="container_image=..."
```

## Environment Variables

### Backend

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENV` | `development` / `production` | `development` |
| `DATABASE_URL` | PostgreSQL connection string (overrides DB_* vars) | -- |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `taskflow` |
| `DB_PASSWORD` | Database password | `taskflow` |
| `DB_NAME` | Database name | `taskflow` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `JWT_SECRET` | JWT signing key | (dev default) |
| `JWT_ACCESS_EXPIRY` | Access token TTL | `15m` |
| `JWT_REFRESH_EXPIRY` | Refresh token TTL | `168h` |
| `CORS_ORIGIN` | Allowed origins (comma-separated) | `http://localhost:3000` |

### Frontend

| Variable | Description |
|----------|-------------|
| `NEXT_PUBLIC_API_URL` | Backend API base URL |
| `NEXT_PUBLIC_WS_URL` | WebSocket endpoint URL |

## Project Structure

```
taskflow/
├── backend/
│   ├── cmd/server/          # Entry point
│   ├── Dockerfile           # Multi-stage production build
│   ├── internal/
│   │   ├── config/          # Environment config
│   │   ├── handler/         # HTTP handlers
│   │   ├── middleware/      # JWT auth, tenant scope
│   │   ├── model/           # Request/response types
│   │   ├── service/         # Business logic
│   │   ├── repository/      # sqlc generated DB layer
│   │   └── ws/              # WebSocket hub
│   └── db/
│       ├── migrations/      # SQL migrations
│       ├── queries/         # sqlc query definitions
│       └── seed.sql         # Demo data
├── frontend/
│   ├── src/app/[locale]/    # App Router pages
│   ├── src/components/      # UI components
│   ├── src/lib/             # API client, stores
│   ├── messages/            # i18n (en, ja)
│   └── vercel.json
├── infra/
│   ├── terraform/           # AWS IaC (VPC, ECS, RDS, ALB)
│   ├── ecs-task-definition.json
│   └── architecture.md
├── .github/workflows/       # CI pipelines
├── render.yaml              # Render Blueprint
└── docker-compose.yml
```

## CI/CD

GitHub Actions runs on every push and PR to `main`:

- **Backend CI** — `go vet`, `go build`, migrations, `go test` (with PostgreSQL service container)
- **Frontend CI** — `tsc --noEmit`, `eslint`, `next build`

## License

MIT
