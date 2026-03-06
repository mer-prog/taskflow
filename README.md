# TaskFlow

Real-time team project management tool with Kanban boards.

## Architecture

```
[Browser] ──── [Vercel (Next.js)] ──── [Render / AWS (Go API)] ──── [PostgreSQL]
                                              │
                                         WebSocket
```

### Demo Environment

| Component | Service |
|-----------|---------|
| Frontend  | Vercel  |
| Backend   | Render (Docker) |
| Database  | Render PostgreSQL |

### Production Environment (AWS IaC)

| Component | Service |
|-----------|---------|
| Frontend  | Vercel |
| Backend   | ECS Fargate (ALB + 2 tasks) |
| Database  | RDS PostgreSQL |
| Network   | VPC + public/private subnets |

See [infra/architecture.md](./infra/architecture.md) for details.

## Tech Stack

- **Backend:** Go 1.22+, Echo v4, sqlc, PostgreSQL 16
- **Frontend:** Next.js 15 (App Router), TypeScript, Tailwind CSS, shadcn/ui
- **Auth:** JWT (access + refresh tokens), bcrypt
- **Real-time:** WebSocket (gorilla/websocket)
- **Infra:** Docker, Render, Vercel, Terraform (AWS IaC)

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- [sqlc](https://sqlc.dev/) (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### Local Development

1. **Clone and start PostgreSQL**

```bash
git clone https://github.com/mer-prog/taskflow.git
cd taskflow
docker compose up db -d
```

2. **Run migrations**

```bash
psql "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable" \
  -f backend/db/migrations/001_init.up.sql
```

3. **Start the backend**

```bash
cp backend/.env.example backend/.env
cd backend
go run ./cmd/server
```

4. **Start the frontend**

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

### Using Docker Compose

```bash
docker compose up --build
```

## Deployment

### Render (Backend)

The project includes a `render.yaml` Blueprint for one-click setup:

1. Connect this repo to [Render](https://render.com)
2. Render auto-detects `render.yaml` and provisions:
   - **Web Service**: Go API built from `backend/Dockerfile`
   - **PostgreSQL**: Free-tier database
3. Set `DATABASE_URL` in the `taskflow-env` env group (auto-provided by Render DB)
4. Run migrations against the Render database

### Vercel (Frontend)

1. Import the `frontend/` directory to [Vercel](https://vercel.com)
2. Set environment variables:
   - `NEXT_PUBLIC_API_URL` = `https://taskflow-api.onrender.com/api/v1`
   - `NEXT_PUBLIC_WS_URL` = `wss://taskflow-api.onrender.com/api/v1/ws`
3. Deploy

### AWS (IaC only)

Terraform code is in `infra/terraform/`. Not actively deployed — see [infra/README.md](./infra/README.md).

```bash
cd infra/terraform
terraform init
terraform plan -var="db_password=..." -var="container_image=..."
terraform apply
```

## Environment Variables

### Backend

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENV` | Environment (`development` / `production`) | `development` |
| `DATABASE_URL` | Full PostgreSQL connection string (overrides individual DB vars) | — |
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

## API Endpoints

### Health Check

```
GET /api/v1/health
```

### Authentication

```
POST /api/v1/auth/register   — Create account
POST /api/v1/auth/login      — Login
POST /api/v1/auth/refresh    — Refresh access token
POST /api/v1/auth/logout     — Logout
```

## Project Structure

```
taskflow/
├── backend/
│   ├── cmd/server/main.go         # Entry point
│   ├── Dockerfile                 # Multi-stage production build
│   ├── internal/
│   │   ├── config/                # Environment config
│   │   ├── handler/               # HTTP handlers
│   │   ├── middleware/            # JWT auth middleware
│   │   ├── model/                 # Request/response types
│   │   ├── service/               # Business logic
│   │   ├── repository/           # sqlc generated DB layer
│   │   └── ws/                    # WebSocket hub
│   └── db/
│       ├── migrations/            # SQL migrations
│       └── queries/               # sqlc query definitions
├── frontend/                      # Next.js app
│   ├── vercel.json                # Vercel config
│   └── src/
├── infra/                         # AWS IaC (Terraform)
│   ├── terraform/
│   ├── ecs-task-definition.json
│   └── architecture.md
└── render.yaml                    # Render Blueprint
```

## License

MIT
