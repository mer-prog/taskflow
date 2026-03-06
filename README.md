# TaskFlow

Real-time team project management tool with Kanban boards.

## Tech Stack

- **Backend:** Go 1.22+, Echo v4, sqlc, PostgreSQL 16
- **Frontend:** Next.js 15 (App Router), TypeScript, Tailwind CSS, shadcn/ui
- **Auth:** JWT (access + refresh tokens), bcrypt
- **Infra:** Docker, AWS ECS Fargate, RDS, ALB

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- [sqlc](https://sqlc.dev/) (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### Setup

1. **Clone the repository**

```bash
git clone https://github.com/mer-prog/taskflow.git
cd taskflow
```

2. **Start PostgreSQL**

```bash
docker compose up db -d
```

3. **Run migrations**

```bash
psql "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable" \
  -f backend/db/migrations/001_init.up.sql
```

4. **Generate sqlc code**

```bash
cd backend
sqlc generate
```

5. **Set environment variables**

```bash
cp backend/.env.example backend/.env
```

6. **Run the API server**

```bash
cd backend
go run ./cmd/server
```

The server starts at `http://localhost:8080`.

### Using Docker Compose

```bash
docker compose up --build
```

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

#### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","display_name":"User"}'
```

#### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

## Project Structure

```
taskflow/
├── backend/
│   ├── cmd/server/main.go         # Entry point
│   ├── internal/
│   │   ├── config/                # Environment config
│   │   ├── handler/               # HTTP handlers
│   │   ├── middleware/            # JWT auth middleware
│   │   ├── model/                 # Request/response types
│   │   ├── service/               # Business logic
│   │   └── repository/           # sqlc generated DB layer
│   └── db/
│       ├── migrations/            # SQL migrations
│       └── queries/               # sqlc query definitions
└── frontend/                      # Next.js app (Phase 5+)
```

## License

MIT
