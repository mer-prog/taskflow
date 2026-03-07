# TaskFlow - Real-Time Kanban Project Management Tool

A multi-tenant team project management tool featuring Kanban boards with real-time synchronization via WebSocket. Built as a fullstack application with Go and Next.js.

---

## Table of Contents

1. [Tech Stack](#tech-stack)
2. [Architecture](#architecture)
3. [Features](#features)
4. [Project Structure](#project-structure)
5. [Database Design](#database-design)
6. [API Design](#api-design)
7. [WebSocket Design](#websocket-design)
8. [Security Design](#security-design)
9. [Screen Specifications](#screen-specifications)
10. [State Management](#state-management)
11. [Internationalization (i18n)](#internationalization-i18n)
12. [CI/CD Pipeline](#cicd-pipeline)
13. [Setup Guide](#setup-guide)
14. [Design Decisions](#design-decisions)
15. [Running Costs](#running-costs)
16. [Author](#author)

---

## Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Frontend | Next.js (App Router) | 16.1.6 |
| | React | 19.2.3 |
| | TypeScript | 5.x |
| | Tailwind CSS | 4.x |
| | shadcn/ui (Radix UI) | 1.4.3 |
| | dnd-kit (drag and drop) | core 6.3.1 / sortable 10.0.0 |
| | next-intl (internationalization) | 4.8.3 |
| | Zustand (state management) | 5.0.11 |
| | Lucide React (icons) | 0.577.0 |
| | Sonner (toast notifications) | 2.0.7 |
| Backend | Go | 1.24.0 |
| | Echo v4 (HTTP framework) | 4.15.1 |
| | pgx/v5 (PostgreSQL driver) | 5.8.0 |
| | gorilla/websocket | 1.5.3 |
| | golang-jwt/jwt/v5 | 5.3.1 |
| | golang.org/x/crypto (bcrypt) | 0.48.0 |
| | sqlc (SQL code generation) | - |
| Database | PostgreSQL | 16 |
| Infrastructure | Docker (multi-stage build) | golang:1.24-alpine to distroless |
| | Docker Compose | - |
| | Terraform (AWS IaC) | >= 1.5 |
| Deployment (demo) | Render (backend) | Free tier |
| | Vercel (frontend) | Free tier |
| Deployment (production IaC) | AWS ECS Fargate | - |
| | AWS RDS PostgreSQL 16 | - |
| | AWS ALB | - |
| CI/CD | GitHub Actions | - |

---

## Architecture

### System Overview

```
                          ┌──────────────────────────────────────────────┐
                          │              Client (Browser)                │
                          └──────────┬─────────────────┬────────────────┘
                                     │                 │
                              HTML/JS│                 │REST API / WebSocket
                                     │                 │
                          ┌──────────▼──────┐  ┌──────▼──────────────────┐
                          │  Next.js App    │  │     Go API Server       │
                          │  (Vercel)       │  │  (Render / AWS ECS)     │
                          │                 │  │                         │
                          │ - App Router    │  │ - Echo v4 (HTTP)        │
                          │ - Zustand       │  │ - JWT Authentication    │
                          │ - dnd-kit       │  │ - WebSocket Hub         │
                          │ - next-intl     │  │ - sqlc (type-safe SQL)  │
                          └─────────────────┘  └───────────┬─────────────┘
                                                           │
                                                           │ pgx/v5
                                                           │
                                                 ┌─────────▼───────────┐
                                                 │   PostgreSQL 16     │
                                                 │ (Render / AWS RDS)  │
                                                 │                     │
                                                 │ 12 tables           │
                                                 │ Multi-tenant        │
                                                 └─────────────────────┘
```

### Backend Layer Architecture

```
Handler (HTTP layer)
    │  Request validation and response formatting
    ▼
Service (Business logic layer)
    │  Domain logic and transaction control
    ▼
Adapter (Repository implementation)
    │  Implements Service interfaces
    ▼
Repository (sqlc generated code)
    │  Type-safe SQL execution
    ▼
PostgreSQL
```

Handlers depend only on Service interfaces and never access the database directly. The Adapter pattern bridges sqlc-generated code to Service-layer interfaces.

### Deployment Architecture

#### Demo Environment (Render + Vercel)

| Component | Service | Plan |
|-----------|---------|------|
| Frontend | Vercel | Free |
| Backend | Render (Docker) | Free |
| Database | Render PostgreSQL | Free |

#### Production Environment (AWS IaC / Terraform)

| Component | AWS Service | Configuration |
|-----------|-------------|---------------|
| Network | VPC | 2 public + 2 private subnets |
| Load Balancer | ALB | HTTP to HTTPS redirect |
| Application | ECS Fargate | 256 CPU / 512 MB x 2 tasks |
| Database | RDS PostgreSQL 16 | db.t3.micro / gp3 20GB |
| Logging | CloudWatch Logs | 30-day retention |
| NAT | NAT Gateway | Single AZ (cost optimization) |

---

## Features

### Authentication and Authorization

- User registration and login with email and password
- JWT access tokens (15-minute expiry) stored in memory
- Refresh tokens (7-day expiry) managed via httpOnly cookies
- Token rotation (previous refresh token invalidated on refresh)
- Refresh token deletion on logout

### Workspace (Tenant) Management

- Create, list, view, and update workspaces
- Invite members by email address
- Update member roles (Owner / Admin / Member / Viewer)
- Remove members
- Automatic default workspace creation on user registration

### Project Management

- Create, list, view, update, and archive projects
- Project member management
- Per-project board management

### Kanban Board

- Create, view, update, and delete boards
- Create, update, delete, and reorder columns
- Column color and WIP limit settings
- Create, view, update, and delete tasks
- Drag-and-drop task movement (within and across columns)
- Optimistic updates (instant UI feedback with rollback on failure)

### Task Features

- Task priority levels (urgent / high / medium / low)
- Assignee management
- Due date tracking
- Label management (create, assign, remove; color-coded)
- Comments

### Real-Time Sync (WebSocket)

- Per-board WebSocket connections
- Live updates for task create, update, delete, and move events
- Live updates for column create, update, delete, and reorder events
- Self-action filtering (prevents duplicate application of own actions)
- Automatic reconnection with exponential backoff

### Dashboard

- Task summary (total, completed, progress percentage)
- Task counts by priority level
- Overdue task list
- Personal assigned tasks list

### Internationalization

- Two-language support: English and Japanese
- Routing-based locale switching with next-intl
- Approximately 100 translation keys

### Responsive Design

- Mobile-first approach (375px and above)
- Hamburger menu for sidebar toggle
- Tailwind CSS responsive prefixes for adaptive layouts

---

## Project Structure

```
taskflow/
├── backend/
│   ├── cmd/server/
│   │   └── main.go                  # Entry point, DI, and routing
│   ├── Dockerfile                    # Multi-stage build (alpine to distroless)
│   ├── internal/
│   │   ├── adapter/                  # Repository interface implementations
│   │   │   ├── auth_repository.go
│   │   │   ├── board_repository.go
│   │   │   ├── dashboard_repository.go
│   │   │   ├── project_repository.go
│   │   │   ├── task_repository.go
│   │   │   └── tenant_repository.go
│   │   ├── config/
│   │   │   └── config.go            # Environment variable loading
│   │   ├── handler/                  # HTTP handlers (9 files)
│   │   │   ├── auth.go
│   │   │   ├── board.go
│   │   │   ├── column.go
│   │   │   ├── dashboard.go
│   │   │   ├── label.go
│   │   │   ├── project.go
│   │   │   ├── task.go
│   │   │   ├── tenant.go
│   │   │   └── ws.go
│   │   ├── middleware/
│   │   │   ├── auth.go              # JWT authentication middleware
│   │   │   └── tenant.go            # Tenant scope and RBAC
│   │   ├── model/
│   │   │   ├── models.go            # JWT claims, request/response types
│   │   │   └── board.go             # Board response types
│   │   ├── repository/              # sqlc auto-generated code
│   │   │   ├── db.go
│   │   │   ├── models.go
│   │   │   └── *.sql.go
│   │   ├── service/                  # Business logic
│   │   │   ├── auth.go
│   │   │   ├── board.go
│   │   │   ├── dashboard.go
│   │   │   ├── errors.go
│   │   │   ├── project.go
│   │   │   ├── task.go
│   │   │   ├── tenant.go
│   │   │   ├── board_test.go        # Board service unit tests
│   │   │   └── task_test.go         # Task service unit tests
│   │   └── ws/                       # WebSocket
│   │       ├── hub.go               # Hub (broadcast management)
│   │       ├── hub_manager.go       # HubManager (per-board Hub)
│   │       └── client.go            # Client (read/write pumps)
│   ├── db/
│   │   ├── migrations/
│   │   │   ├── 001_init.up.sql              # Initial schema (12 tables)
│   │   │   ├── 001_init.down.sql
│   │   │   ├── 002_add_project_archived_at.up.sql
│   │   │   ├── 002_add_project_archived_at.down.sql
│   │   │   ├── 003_add_column_color_task_priority.up.sql
│   │   │   └── 003_add_column_color_task_priority.down.sql
│   │   ├── queries/                  # sqlc query definitions (9 files)
│   │   │   ├── auth.sql
│   │   │   ├── boards.sql
│   │   │   ├── columns.sql
│   │   │   ├── comments.sql
│   │   │   ├── labels.sql
│   │   │   ├── projects.sql
│   │   │   ├── tasks.sql
│   │   │   ├── tenants.sql
│   │   │   └── users.sql
│   │   └── seed.sql                 # Demo data
│   └── sqlc.yaml                     # sqlc configuration
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx                         # Root layout
│   │   │   └── [locale]/
│   │   │       ├── layout.tsx                     # Locale layout (next-intl)
│   │   │       ├── page.tsx                       # Landing page
│   │   │       ├── login/page.tsx                 # Login
│   │   │       ├── register/page.tsx              # User registration
│   │   │       └── ws/[slug]/
│   │   │           ├── layout.tsx                 # Workspace layout
│   │   │           ├── page.tsx                   # Dashboard
│   │   │           ├── projects/page.tsx           # Project list
│   │   │           ├── settings/page.tsx           # Workspace settings
│   │   │           ├── members/page.tsx            # Member management
│   │   │           └── p/[id]/board/page.tsx       # Kanban board
│   │   ├── components/
│   │   │   ├── board/                 # Board components (6 files)
│   │   │   │   ├── KanbanBoard.tsx    # Main board (DndContext)
│   │   │   │   ├── KanbanColumn.tsx   # Column (SortableContext)
│   │   │   │   ├── TaskCard.tsx       # Task card (useSortable)
│   │   │   │   ├── TaskDetailModal.tsx # Task detail modal
│   │   │   │   ├── AddTaskForm.tsx    # Task creation form
│   │   │   │   └── ColumnHeader.tsx   # Column header
│   │   │   ├── layout/                # Layout components (3 files)
│   │   │   │   ├── Header.tsx
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   └── LocaleSwitcher.tsx
│   │   │   └── ui/                    # shadcn/ui components (17 files)
│   │   ├── hooks/
│   │   │   ├── useAuth.ts            # Authentication guard hook
│   │   │   └── useWebSocket.ts       # WebSocket connection hook
│   │   ├── lib/
│   │   │   ├── api.ts                # API client (automatic token refresh)
│   │   │   └── utils.ts              # Utilities (cn function)
│   │   ├── stores/
│   │   │   ├── authStore.ts          # Authentication state
│   │   │   ├── boardStore.ts         # Board, column, and task state
│   │   │   ├── dashboardStore.ts     # Dashboard state
│   │   │   └── workspaceStore.ts     # Tenant, project, and member state
│   │   └── i18n/
│   │       ├── routing.ts            # Locale routing configuration
│   │       ├── request.ts            # Server-side translation retrieval
│   │       └── navigation.ts         # Locale-aware navigation
│   ├── messages/
│   │   ├── en.json                   # English translations (~100 keys)
│   │   └── ja.json                   # Japanese translations (~100 keys)
│   └── vercel.json                   # Vercel deployment configuration
├── infra/
│   ├── terraform/
│   │   ├── main.tf                   # VPC / ALB / ECS / RDS definitions
│   │   ├── variables.tf              # Variable definitions
│   │   └── outputs.tf                # Output definitions
│   ├── ecs-task-definition.json
│   └── architecture.md
├── .github/workflows/
│   ├── backend-ci.yml                # Go vet / build / test (PostgreSQL service container)
│   └── frontend-ci.yml              # tsc / eslint / next build
├── docker-compose.yml                # Local development (PostgreSQL + API)
├── render.yaml                       # Render Blueprint
└── README.md                         # Project overview
```

---

## Database Design

### Table Overview

12 tables with 19 indexes. All primary keys are UUID v4. All timestamps are TIMESTAMPTZ.

```
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│   tenants     │       │    users      │       │refresh_tokens │
│───────────────│       │───────────────│       │───────────────│
│ id (PK)       │       │ id (PK)       │       │ id (PK)       │
│ name          │       │ email (UQ)    │       │ user_id (FK)  │
│ slug (UQ)     │       │ password_hash │       │ token_hash(UQ)│
│ created_at    │       │ display_name  │       │ expires_at    │
│ updated_at    │       │ avatar_url    │       │ created_at    │
└───────┬───────┘       │ created_at    │       └───────────────┘
        │               │ updated_at    │
        │               └───────┬───────┘
        │                       │
        ▼                       ▼
┌───────────────────────────────────────┐
│         tenant_members                │
│───────────────────────────────────────│
│ id (PK)                              │
│ tenant_id (FK) + user_id (FK) = UQ   │
│ role (owner/admin/member/viewer)     │
│ created_at                           │
└───────────────────────────────────────┘
        │
        ▼
┌───────────────┐       ┌───────────────────────────────────────┐
│   projects    │       │         project_members               │
│───────────────│       │───────────────────────────────────────│
│ id (PK)       │──────►│ id (PK)                              │
│ tenant_id(FK) │       │ project_id (FK) + user_id (FK) = UQ  │
│ name          │       │ role                                 │
│ description   │       │ created_at                           │
│ archived_at   │       └───────────────────────────────────────┘
│ created_at    │
│ updated_at    │
└───────┬───────┘
        │
        ▼
┌───────────────┐
│    boards     │
│───────────────│
│ id (PK)       │
│ tenant_id(FK) │
│ project_id(FK)│
│ name          │
│ created_at    │
│ updated_at    │
└───────┬───────┘
        │
        ▼
┌───────────────┐       ┌───────────────┐
│   columns     │       │    tasks      │
│───────────────│       │───────────────│
│ id (PK)       │──────►│ id (PK)       │
│ tenant_id(FK) │       │ tenant_id(FK) │
│ board_id (FK) │       │ column_id(FK) │
│ name          │       │ title         │
│ position      │       │ description   │
│ color         │       │ position      │
│ wip_limit     │       │ priority      │
│ created_at    │       │ assignee_id   │
│ updated_at    │       │ due_date      │
└───────────────┘       │ created_at    │
                        │ updated_at    │
                        └───────┬───────┘
                                │
                ┌───────────────┼───────────────┐
                ▼               ▼               ▼
        ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
        │ task_labels   │ │task_comments │ │   labels     │
        │──────────────│ │──────────────│ │──────────────│
        │task_id (PK,FK)│ │ id (PK)      │ │ id (PK)      │
        │label_id(PK,FK)│ │ tenant_id(FK)│ │ tenant_id(FK)│
        └──────────────┘ │ task_id (FK) │ │ name         │
                         │ user_id (FK) │ │ color        │
                         │ content      │ │ created_at   │
                         │ created_at   │ └──────────────┘
                         │ updated_at   │
                         └──────────────┘
```

### Migrations

| File | Description |
|------|-------------|
| `001_init.up.sql` | Initial schema: 12 tables and 19 indexes |
| `001_init.down.sql` | Drop all tables and indexes |
| `002_add_project_archived_at.up.sql` | Add `archived_at` column to projects table |
| `002_add_project_archived_at.down.sql` | Remove `archived_at` column |
| `003_add_column_color_task_priority.up.sql` | Add `color` and `wip_limit` to columns, `priority` to tasks |
| `003_add_column_color_task_priority.down.sql` | Remove added columns |

### Multi-Tenant Data Isolation

All tables (except authentication tables) include a `tenant_id` column, and every SQL query filters by `tenant_id`. The `X-Tenant-ID` header value sent by the client is validated against the JWT claims by the tenant scope middleware.

---

## API Design

### Base URL

```
/api/v1
```

### Authentication Endpoints (Public)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Register a new user |
| POST | `/auth/login` | Log in |
| POST | `/auth/refresh` | Reissue access token |
| POST | `/auth/logout` | Log out |

### Health Check

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Server health check |

### WebSocket

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ws?board_id={id}` | Per-board real-time connection |

### Tenant Endpoints (JWT Required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/tenants` | Create tenant |
| GET | `/tenants` | List tenants |
| GET | `/tenants/:id` | Get tenant details |
| PATCH | `/tenants/:id` | Update tenant |
| GET | `/tenants/:id/members` | List members |
| POST | `/tenants/:id/members` | Add member |
| PATCH | `/tenants/:id/members/:uid` | Update member role |
| DELETE | `/tenants/:id/members/:uid` | Remove member |

### Project Endpoints (JWT + Tenant Scope)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/projects` | List projects |
| POST | `/projects` | Create project |
| GET | `/projects/:id` | Get project details |
| PATCH | `/projects/:id` | Update project |
| DELETE | `/projects/:id` | Archive project |
| GET | `/projects/:id/members` | List project members |
| POST | `/projects/:id/members` | Add project member |
| DELETE | `/projects/:id/members/:uid` | Remove project member |
| GET | `/projects/:id/boards` | List boards for project |

### Board Endpoints (JWT + Tenant Scope)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/boards` | Create board |
| GET | `/boards/:id` | Get board details (includes columns and tasks) |
| PATCH | `/boards/:id` | Update board |
| DELETE | `/boards/:id` | Delete board |

### Column Endpoints (JWT + Tenant Scope)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/columns` | Create column |
| PATCH | `/columns/reorder` | Reorder columns |
| PATCH | `/columns/:id` | Update column |
| DELETE | `/columns/:id` | Delete column |

### Task Endpoints (JWT + Tenant Scope)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/tasks` | Create task |
| PATCH | `/tasks/move` | Move task (across columns) |
| GET | `/tasks/:id` | Get task details |
| PATCH | `/tasks/:id` | Update task |
| DELETE | `/tasks/:id` | Delete task |
| POST | `/tasks/:id/labels` | Add label to task |
| DELETE | `/tasks/:id/labels/:lid` | Remove label from task |
| POST | `/tasks/:id/comments` | Add comment |
| GET | `/tasks/:id/comments` | List comments |

### Label Endpoints (JWT + Tenant Scope)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/projects/:id/labels` | List labels for project |
| POST | `/projects/:id/labels` | Create label |
| DELETE | `/labels/:id` | Delete label |

### Dashboard Endpoints (JWT + Tenant Scope)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/dashboard/summary` | Task summary statistics |
| GET | `/dashboard/overdue` | Overdue tasks list |
| GET | `/dashboard/my-tasks` | Current user's assigned tasks |

**Total: 46 endpoints**

### Authentication Headers

```
Authorization: Bearer <access_token>
X-Tenant-ID: <tenant_uuid>
```

Tenant-scoped endpoints require the `X-Tenant-ID` header. The middleware validates this header against the JWT `tenant_id` claim and verifies tenant membership.

---

## WebSocket Design

### Connection Protocol

```
ws(s)://<host>/api/v1/ws?board_id=<uuid>
```

The authentication token is sent via the `Sec-WebSocket-Protocol` header:

```
Sec-WebSocket-Protocol: access_token.<jwt_token>
```

This approach avoids exposing the token in the URL, which could appear in server logs and browser history.

### Hub Pattern

```
HubManager
    │
    ├── Hub (board_id=aaa)
    │     ├── Client A (user_1)
    │     ├── Client B (user_2)
    │     └── Client C (user_3)
    │
    └── Hub (board_id=bbb)
          ├── Client D (user_1)
          └── Client E (user_4)
```

- **HubManager**: Manages one `Hub` per board ID
- **Hub**: Broadcasts messages to all clients connected to the same board
- **Client**: Implements goroutine-based read and write pumps

### Event Types

| Event | Trigger | Payload |
|-------|---------|---------|
| `task:created` | Task creation | TaskDetail object |
| `task:updated` | Task update | TaskDetail object |
| `task:deleted` | Task deletion | `{ id: string }` |
| `task:moved` | Task movement | Move information |
| `column:created` | Column creation | Column object |
| `column:updated` | Column update | Column object |
| `column:deleted` | Column deletion | `{ id: string }` |
| `column:reordered` | Column reorder | Order information |

### Client-Side Handling

- REST handlers update the database, then broadcast via `HubManager`
- Receivers update their state through `handleWSMessage`
- Actions from the same user (`user_id` match) are skipped to prevent double application
- `task:moved` and `column:*` events trigger a full board refetch
- Disconnections are handled with exponential backoff (starting at 1 second, capped at 30 seconds)

---

## Security Design

### Authentication Flow

```
1. Login
   POST /auth/login { email, password }
       ↓
   bcrypt verification (cost factor 12)
       ↓
   Generate JWT access token (HS256, 15 minutes)
   Generate refresh token (random 32 bytes, SHA-256 hash stored in DB, 7 days)
       ↓
   Response: { access_token, user } + Set-Cookie (refresh_token, httpOnly)

2. API Request
   Authorization: Bearer <access_token>
       ↓
   JWTAuth middleware validates token and sets user_id and tenant_id in Context

3. Token Refresh
   POST /auth/refresh (refresh_token from cookie)
       ↓
   Invalidate previous refresh token (rotation)
       ↓
   Issue new access token and refresh token

4. Frontend Auto-Refresh
   apiFetch() receives 401 → automatic refresh → retry original request
```

### Role-Based Access Control (RBAC)

| Role | Permissions |
|------|------------|
| Owner | All operations + tenant settings + member management |
| Admin | All operations + member management |
| Member | CRUD on tasks, columns, and boards |
| Viewer | Read-only access |

The `TenantScope` middleware validates the `X-Tenant-ID` header against JWT claims and verifies tenant membership. Write operations enforce role-based access control.

### Additional Security Measures

- Passwords: bcrypt with cost factor 12
- CORS: Only allows whitelisted frontend origins (comma-separated)
- Internal error details hidden from clients (generic error messages returned)
- WebSocket authentication: Via protocol header (token not in URL)
- Tenant isolation: All queries include `tenant_id` filter
- Production Docker image: distroless (minimal attack surface)
- Refresh tokens: SHA-256 hash stored in database (raw tokens are never persisted)

---

## Screen Specifications

### Page List

| Route | Page | Auth | Description |
|-------|------|------|-------------|
| `/{locale}` | Landing | No | Application introduction with login link |
| `/{locale}/login` | Login | No | Email and password input form |
| `/{locale}/register` | Registration | No | Display name, email, and password input form |
| `/{locale}/ws/{slug}` | Dashboard | Yes | Task overview, priority breakdown, project list |
| `/{locale}/ws/{slug}/projects` | Projects | Yes | Create, edit, and delete projects |
| `/{locale}/ws/{slug}/settings` | Settings | Yes | Workspace name configuration |
| `/{locale}/ws/{slug}/members` | Members | Yes | Member list, role management, invitations, removal |
| `/{locale}/ws/{slug}/p/{id}/board` | Kanban Board | Yes | Drag-and-drop Kanban board |

### Layout Hierarchy

- **Root Layout**: Font and metadata configuration
- **Locale Layout**: `NextIntlClientProvider` and `Toaster` setup
- **Workspace Layout**: `Header` + `Sidebar`, authentication guard, tenant context management

### Key Components

| Component | Description |
|-----------|-------------|
| `KanbanBoard` | Manages `DndContext` with horizontal column layout |
| `KanbanColumn` | Manages task sorting via `SortableContext` |
| `TaskCard` | Draggable task card using `useSortable` |
| `TaskDetailModal` | Modal for viewing and editing task details |
| `AddTaskForm` | Task creation form at column bottom |
| `ColumnHeader` | Column name, color, and WIP limit display and editing |
| `Header` | Logo, hamburger menu, and user dropdown |
| `Sidebar` | Navigation links and project list |
| `LocaleSwitcher` | Language toggle button |

---

## State Management

Client state is managed through four Zustand stores.

| Store | Responsibility | Key State |
|-------|---------------|-----------|
| `authStore` | Authentication | `user`, `accessToken`, `tenantId`, `isAuthenticated` |
| `boardStore` | Board state | `columns`, `selectedTask`, `comments`, `boardId` |
| `workspaceStore` | Workspace state | `tenant`, `projects`, `members`, `currentUserRole` |
| `dashboardStore` | Dashboard state | `summary`, `overdueTasks`, `myTasks` |

### Optimistic Update Pattern

Drag-and-drop and similar operations update the store immediately, with rollback on API failure:

```
1. Update store immediately (optimistic)
2. Call REST API
3. On success: keep current state
4. On failure: rollback to previous state
```

Applied in `createTask`, `moveTask`, `deleteTask`, and `deleteColumn`.

---

## Internationalization (i18n)

### Supported Languages

| Code | Language |
|------|----------|
| `ja` | Japanese |
| `en` | English |

### Implementation

- Routing-based locale management using `next-intl`
- URL path prefix determines locale (`/ja/...`, `/en/...`)
- Translations used in both Server Components and Client Components
- Approximately 100 translation keys covering navigation, forms, error messages, and board operations

### Translation File Structure

```
messages/
├── en.json    # English
└── ja.json    # Japanese
```

---

## CI/CD Pipeline

### GitHub Actions Workflows

Triggered on push and pull request to the `main` branch. Path filters ensure only the affected component runs.

#### Backend CI (`backend-ci.yml`)

```
Trigger: Changes in backend/**
Runner: ubuntu-latest
Service container: postgres:16-alpine

Steps:
1. Set up Go 1.22 (with dependency caching)
2. go mod download
3. go vet ./...
4. go build ./...
5. Run SQL migrations
6. go test ./... -v -count=1
```

#### Frontend CI (`frontend-ci.yml`)

```
Trigger: Changes in frontend/**
Runner: ubuntu-latest

Steps:
1. Set up Node.js 20 (with npm caching)
2. npm ci
3. npx tsc --noEmit (type checking)
4. npm run lint (ESLint)
5. npm run build (build verification)
```

---

## Setup Guide

### Prerequisites

- Go 1.22 or later
- Node.js 20 or later
- Docker and Docker Compose
- sqlc (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### Local Development

```bash
# 1. Clone the repository and start PostgreSQL
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

# 5. Start the frontend (in a separate terminal)
cd frontend
cp .env.example .env.local
npm install && npm run dev
```

### Demo Accounts

Available after running `seed.sql`:

| Email | Password | Role |
|-------|----------|------|
| `demo@taskflow.app` | `demo1234` | Owner |
| `alice@taskflow.app` | `demo1234` | Admin |
| `bob@taskflow.app` | `demo1234` | Member |

### Docker Compose (Full Stack)

```bash
docker compose up --build
```

This starts PostgreSQL (port 5432) and the Go API (port 8080).

### Environment Variables

#### Backend

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENV` | Runtime environment (`development` / `production`) | `development` |
| `DATABASE_URL` | PostgreSQL connection string (overrides DB_* variables) | -- |
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

#### Frontend

| Variable | Description |
|----------|-------------|
| `NEXT_PUBLIC_API_URL` | Backend API base URL |
| `NEXT_PUBLIC_WS_URL` | WebSocket endpoint URL |

### Deployment (Render + Vercel)

1. Connect the repository to [Render](https://render.com). The `render.yaml` Blueprint auto-provisions the Web Service and PostgreSQL database.
2. Set `DATABASE_URL` in the `taskflow-env` environment variable group.
3. Run migrations against the Render database.
4. Import `frontend/` into [Vercel](https://vercel.com) with environment variables:
   - `NEXT_PUBLIC_API_URL` = `https://taskflow-api.onrender.com/api/v1`
   - `NEXT_PUBLIC_WS_URL` = `wss://taskflow-api.onrender.com/api/v1/ws`

### Deployment (AWS / IaC Only)

Terraform code is in `infra/terraform/`. Not actively deployed.

```bash
cd infra/terraform
terraform init
terraform plan -var="db_password=..." -var="container_image=..."
```

---

## Design Decisions

### Adapter Pattern

sqlc-generated code lives in the `repository` package. The Service layer defines its own interfaces, and Adapters implement those interfaces by wrapping sqlc calls. This decouples business logic from generated code and makes it straightforward to swap in mock repositories for testing.

### Optimistic Updates

For drag-and-drop operations, user experience is the top priority. The UI updates immediately, the REST API call runs in the background, and changes are rolled back only on failure.

### WebSocket Hub Pattern

Each board ID gets its own independent Hub, preventing unnecessary message delivery to unrelated boards. Goroutine-based read and write pumps handle concurrent connections efficiently.

### sqlc for Type-Safe SQL

Instead of an ORM, raw SQL queries are written directly and sqlc generates type-safe Go code. This makes SQL optimization straightforward while maintaining compile-time type safety.

### Shared-Schema Multi-Tenancy

All tenants share a single database and schema. Every table includes a `tenant_id` column, and the middleware validates JWT claims against the header value. This approach keeps infrastructure costs low compared to schema-per-tenant isolation.

### JWT with Refresh Token Rotation

Access tokens are stored in memory (not httpOnly cookies) to give the frontend flexibility. Refresh tokens are protected via httpOnly cookies and rotated on each refresh, invalidating the previous token to mitigate replay attacks.

### Distroless Docker Image

The production build uses `gcr.io/distroless/static-debian12:nonroot`. This minimal image contains no shell or package manager, reducing the attack surface.

### Single NAT Gateway

The AWS configuration uses a single-AZ NAT Gateway to minimize costs. For higher availability requirements, NAT Gateways can be added to each AZ.

---

## Running Costs

### Demo Environment (Render + Vercel)

| Service | Monthly Cost |
|---------|-------------|
| Render Web Service | Free |
| Render PostgreSQL | Free |
| Vercel (frontend) | Free |
| **Total** | **$0/month** |

Note: Render free tier services sleep after 15 minutes of inactivity.

### Production Environment (AWS / Estimated)

| Service | Configuration | Estimated Monthly Cost |
|---------|---------------|----------------------|
| ECS Fargate | 256 CPU / 512 MB x 2 tasks | ~$18 |
| RDS PostgreSQL | db.t3.micro / gp3 20GB / single AZ | ~$15 |
| NAT Gateway | Single AZ | ~$32 |
| ALB | 1 instance | ~$16 |
| CloudWatch Logs | 30-day retention | ~$3 |
| Data Transfer | Estimated | ~$3 |
| **Total** | | **~$87/month** |

Based on the us-west-2 region. Tokyo region pricing would be slightly higher.
Vercel frontend assumed to be on the free tier.

---

## Author

- GitHub: [mer-prog](https://github.com/mer-prog)
- Repository: [mer-prog/taskflow](https://github.com/mer-prog/taskflow)

---

## License

MIT
