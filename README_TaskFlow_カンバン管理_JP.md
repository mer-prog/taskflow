# TaskFlow - リアルタイムカンバンプロジェクト管理ツール

マルチテナント対応のチーム向けプロジェクト管理ツール。
カンバンボードによるタスク管理と WebSocket によるリアルタイム同期を特徴とする、Go + Next.js のフルスタック構成。

---

## 目次

1. [技術スタック](#技術スタック)
2. [アーキテクチャ](#アーキテクチャ)
3. [機能一覧](#機能一覧)
4. [プロジェクト構成](#プロジェクト構成)
5. [データベース設計](#データベース設計)
6. [API 設計](#api-設計)
7. [WebSocket 設計](#websocket-設計)
8. [セキュリティ設計](#セキュリティ設計)
9. [画面仕様](#画面仕様)
10. [状態管理](#状態管理)
11. [国際化 (i18n)](#国際化-i18n)
12. [CI/CD パイプライン](#cicd-パイプライン)
13. [セットアップ手順](#セットアップ手順)
14. [設計上の意思決定](#設計上の意思決定)
15. [ランニングコスト](#ランニングコスト)
16. [著者](#著者)

---

## 技術スタック

| レイヤー | 技術 | バージョン |
|---------|------|-----------|
| フロントエンド | Next.js (App Router) | 16.1.6 |
| | React | 19.2.3 |
| | TypeScript | 5.x |
| | Tailwind CSS | 4.x |
| | shadcn/ui (Radix UI) | 1.4.3 |
| | dnd-kit (ドラッグ&ドロップ) | core 6.3.1 / sortable 10.0.0 |
| | next-intl (国際化) | 4.8.3 |
| | Zustand (状態管理) | 5.0.11 |
| | Lucide React (アイコン) | 0.577.0 |
| | Sonner (トースト通知) | 2.0.7 |
| バックエンド | Go | 1.24.0 |
| | Echo v4 (HTTP フレームワーク) | 4.15.1 |
| | pgx/v5 (PostgreSQL ドライバ) | 5.8.0 |
| | gorilla/websocket | 1.5.3 |
| | golang-jwt/jwt/v5 | 5.3.1 |
| | golang.org/x/crypto (bcrypt) | 0.48.0 |
| | sqlc (SQL コード生成) | - |
| データベース | PostgreSQL | 16 |
| インフラ | Docker (マルチステージビルド) | golang:1.24-alpine → distroless |
| | Docker Compose | - |
| | Terraform (AWS IaC) | >= 1.5 |
| デプロイ (デモ) | Render (バックエンド) | 無料プラン |
| | Vercel (フロントエンド) | 無料プラン |
| デプロイ (本番 IaC) | AWS ECS Fargate | - |
| | AWS RDS PostgreSQL 16 | - |
| | AWS ALB | - |
| CI/CD | GitHub Actions | - |

---

## アーキテクチャ

### システム全体構成図

```
                          ┌──────────────────────────────────────────────┐
                          │            クライアント (ブラウザ)              │
                          └──────────┬─────────────────┬────────────────┘
                                     │                 │
                              HTML/JS│                 │REST API / WebSocket
                                     │                 │
                          ┌──────────▼──────┐  ┌──────▼──────────────────┐
                          │  Next.js App    │  │     Go API サーバー       │
                          │  (Vercel)       │  │  (Render / AWS ECS)     │
                          │                 │  │                         │
                          │ ・App Router    │  │ ・Echo v4 (HTTP)        │
                          │ ・Zustand       │  │ ・JWT 認証              │
                          │ ・dnd-kit       │  │ ・WebSocket Hub         │
                          │ ・next-intl     │  │ ・sqlc (型安全 SQL)     │
                          └─────────────────┘  └───────────┬─────────────┘
                                                           │
                                                           │ pgx/v5
                                                           │
                                                 ┌─────────▼───────────┐
                                                 │   PostgreSQL 16     │
                                                 │ (Render / AWS RDS)  │
                                                 │                     │
                                                 │ 12 テーブル          │
                                                 │ マルチテナント分離    │
                                                 └─────────────────────┘
```

### バックエンドレイヤー構成

```
Handler (HTTP 層)
    │  リクエスト検証・レスポンス整形
    ▼
Service (ビジネスロジック層)
    │  ドメインロジック・トランザクション制御
    ▼
Adapter (リポジトリ実装)
    │  Service のインターフェースを実装
    ▼
Repository (sqlc 生成コード)
    │  型安全な SQL 呼び出し
    ▼
PostgreSQL
```

Handler は Service のインターフェースにのみ依存し、データベースに直接アクセスしない。
Adapter パターンにより、sqlc が生成したコードを Service 層のインターフェースに適合させている。

### デプロイ構成

#### デモ環境 (Render + Vercel)

| コンポーネント | サービス | プラン |
|--------------|---------|-------|
| フロントエンド | Vercel | 無料 |
| バックエンド | Render (Docker) | 無料 |
| データベース | Render PostgreSQL | 無料 |

#### 本番環境 (AWS IaC / Terraform)

| コンポーネント | AWS サービス | 構成 |
|--------------|-------------|------|
| ネットワーク | VPC | パブリック 2 + プライベート 2 サブネット |
| ロードバランサー | ALB | HTTP → HTTPS リダイレクト |
| アプリケーション | ECS Fargate | 256 CPU / 512 MB × 2 タスク |
| データベース | RDS PostgreSQL 16 | db.t3.micro / gp3 20GB |
| ログ | CloudWatch Logs | 30 日保持 |
| NAT | NAT Gateway | 単一 AZ (コスト削減) |

---

## 機能一覧

### 認証・認可

- メール + パスワードによるユーザー登録・ログイン
- JWT アクセストークン (有効期限 15 分) をメモリ保持
- リフレッシュトークン (有効期限 7 日) を httpOnly Cookie で管理
- トークンローテーション (リフレッシュ時に旧トークンを無効化)
- ログアウト時のリフレッシュトークン削除

### ワークスペース (テナント) 管理

- ワークスペースの作成・一覧・詳細・更新
- メンバー招待 (メールアドレス指定)
- メンバー権限変更 (Owner / Admin / Member / Viewer)
- メンバー削除
- ユーザー登録時にデフォルトワークスペースを自動作成

### プロジェクト管理

- プロジェクトの作成・一覧・詳細・更新・アーカイブ
- プロジェクトメンバー管理
- プロジェクト単位でのボード管理

### カンバンボード

- ボードの作成・詳細・更新・削除
- カラムの作成・更新・削除・並び替え
- カラムの色設定・WIP 制限
- タスクの作成・詳細・更新・削除
- ドラッグ&ドロップによるタスクの移動 (同一カラム内 + カラム間)
- 楽観的更新 (即座に UI 反映、失敗時ロールバック)

### タスク機能

- タスク優先度 (urgent / high / medium / low)
- 担当者アサイン
- 期限設定
- ラベル管理 (作成・割り当て・削除、色付き)
- コメント機能

### リアルタイム同期 (WebSocket)

- ボード単位での WebSocket 接続
- タスクの作成・更新・削除・移動をリアルタイムに反映
- カラムの作成・更新・削除・並び替えをリアルタイムに反映
- 自分自身の操作はスキップ (二重反映防止)
- 指数バックオフによる自動再接続

### ダッシュボード

- タスク概要 (総数・完了数・進捗率)
- 優先度別タスク集計
- 期限切れタスク一覧
- 自分の担当タスク一覧

### 国際化

- 日本語 / 英語の 2 言語対応
- next-intl によるルーティングベースのロケール切替
- 約 100 個の翻訳キー

### レスポンシブデザイン

- モバイルファースト設計 (375px 以上)
- ハンバーガーメニューによるサイドバー切替
- Tailwind CSS のレスポンシブプレフィックスによる適応

---

## プロジェクト構成

```
taskflow/
├── backend/
│   ├── cmd/server/
│   │   └── main.go                  # エントリーポイント・DI・ルーティング
│   ├── Dockerfile                    # マルチステージビルド (alpine → distroless)
│   ├── internal/
│   │   ├── adapter/                  # リポジトリインターフェース実装
│   │   │   ├── auth_repository.go
│   │   │   ├── board_repository.go
│   │   │   ├── dashboard_repository.go
│   │   │   ├── project_repository.go
│   │   │   ├── task_repository.go
│   │   │   └── tenant_repository.go
│   │   ├── config/
│   │   │   └── config.go            # 環境変数読み込み
│   │   ├── handler/                  # HTTP ハンドラー (9 ファイル)
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
│   │   │   ├── auth.go              # JWT 認証ミドルウェア
│   │   │   └── tenant.go            # テナントスコープ + RBAC
│   │   ├── model/
│   │   │   ├── models.go            # JWT クレーム・リクエスト/レスポンス型
│   │   │   └── board.go             # ボードレスポンス型
│   │   ├── repository/              # sqlc 自動生成コード
│   │   │   ├── db.go
│   │   │   ├── models.go
│   │   │   └── *.sql.go
│   │   ├── service/                  # ビジネスロジック
│   │   │   ├── auth.go
│   │   │   ├── board.go
│   │   │   ├── dashboard.go
│   │   │   ├── errors.go
│   │   │   ├── project.go
│   │   │   ├── task.go
│   │   │   ├── tenant.go
│   │   │   ├── board_test.go        # ボードサービスユニットテスト
│   │   │   └── task_test.go         # タスクサービスユニットテスト
│   │   └── ws/                       # WebSocket
│   │       ├── hub.go               # Hub (ブロードキャスト管理)
│   │       ├── hub_manager.go       # HubManager (ボード別 Hub)
│   │       └── client.go            # Client (読み書きポンプ)
│   ├── db/
│   │   ├── migrations/
│   │   │   ├── 001_init.up.sql              # 初期スキーマ (12 テーブル)
│   │   │   ├── 001_init.down.sql
│   │   │   ├── 002_add_project_archived_at.up.sql
│   │   │   ├── 002_add_project_archived_at.down.sql
│   │   │   ├── 003_add_column_color_task_priority.up.sql
│   │   │   └── 003_add_column_color_task_priority.down.sql
│   │   ├── queries/                  # sqlc クエリ定義 (9 ファイル)
│   │   │   ├── auth.sql
│   │   │   ├── boards.sql
│   │   │   ├── columns.sql
│   │   │   ├── comments.sql
│   │   │   ├── labels.sql
│   │   │   ├── projects.sql
│   │   │   ├── tasks.sql
│   │   │   ├── tenants.sql
│   │   │   └── users.sql
│   │   └── seed.sql                 # デモデータ
│   └── sqlc.yaml                     # sqlc 設定
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx                         # ルートレイアウト
│   │   │   └── [locale]/
│   │   │       ├── layout.tsx                     # ロケールレイアウト (next-intl)
│   │   │       ├── page.tsx                       # ランディングページ
│   │   │       ├── login/page.tsx                 # ログイン
│   │   │       ├── register/page.tsx              # ユーザー登録
│   │   │       └── ws/[slug]/
│   │   │           ├── layout.tsx                 # ワークスペースレイアウト
│   │   │           ├── page.tsx                   # ダッシュボード
│   │   │           ├── projects/page.tsx           # プロジェクト一覧
│   │   │           ├── settings/page.tsx           # ワークスペース設定
│   │   │           ├── members/page.tsx            # メンバー管理
│   │   │           └── p/[id]/board/page.tsx       # カンバンボード
│   │   ├── components/
│   │   │   ├── board/                 # ボードコンポーネント (6 ファイル)
│   │   │   │   ├── KanbanBoard.tsx    # メインボード (DndContext)
│   │   │   │   ├── KanbanColumn.tsx   # カラム (SortableContext)
│   │   │   │   ├── TaskCard.tsx       # タスクカード (useSortable)
│   │   │   │   ├── TaskDetailModal.tsx # タスク詳細モーダル
│   │   │   │   ├── AddTaskForm.tsx    # タスク追加フォーム
│   │   │   │   └── ColumnHeader.tsx   # カラムヘッダー
│   │   │   ├── layout/                # レイアウトコンポーネント (3 ファイル)
│   │   │   │   ├── Header.tsx
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   └── LocaleSwitcher.tsx
│   │   │   └── ui/                    # shadcn/ui コンポーネント (17 ファイル)
│   │   ├── hooks/
│   │   │   ├── useAuth.ts            # 認証ガードフック
│   │   │   └── useWebSocket.ts       # WebSocket 接続フック
│   │   ├── lib/
│   │   │   ├── api.ts                # API クライアント (自動トークンリフレッシュ)
│   │   │   └── utils.ts              # ユーティリティ (cn 関数)
│   │   ├── stores/
│   │   │   ├── authStore.ts          # 認証状態
│   │   │   ├── boardStore.ts         # ボード・カラム・タスク状態
│   │   │   ├── dashboardStore.ts     # ダッシュボード状態
│   │   │   └── workspaceStore.ts     # テナント・プロジェクト・メンバー状態
│   │   └── i18n/
│   │       ├── routing.ts            # ロケールルーティング設定
│   │       ├── request.ts            # サーバーサイド翻訳取得
│   │       └── navigation.ts         # ロケール対応ナビゲーション
│   ├── messages/
│   │   ├── en.json                   # 英語翻訳 (~100 キー)
│   │   └── ja.json                   # 日本語翻訳 (~100 キー)
│   └── vercel.json                   # Vercel デプロイ設定
├── infra/
│   ├── terraform/
│   │   ├── main.tf                   # VPC / ALB / ECS / RDS 定義
│   │   ├── variables.tf              # 変数定義
│   │   └── outputs.tf                # 出力定義
│   ├── ecs-task-definition.json
│   └── architecture.md
├── .github/workflows/
│   ├── backend-ci.yml                # Go vet / build / test (PostgreSQL サービスコンテナ)
│   └── frontend-ci.yml              # tsc / eslint / next build
├── docker-compose.yml                # ローカル開発用 (PostgreSQL + API)
├── render.yaml                       # Render Blueprint
└── README.md                         # プロジェクト概要
```

---

## データベース設計

### テーブル一覧

全 12 テーブル、19 インデックス。全主キーは UUID v4、全タイムスタンプは TIMESTAMPTZ。

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

### マイグレーション

| ファイル | 内容 |
|---------|------|
| `001_init.up.sql` | 初期スキーマ: 12 テーブル + 19 インデックス作成 |
| `001_init.down.sql` | 全テーブル・インデックス削除 |
| `002_add_project_archived_at.up.sql` | projects テーブルに `archived_at` カラム追加 |
| `002_add_project_archived_at.down.sql` | `archived_at` カラム削除 |
| `003_add_column_color_task_priority.up.sql` | columns に `color`・`wip_limit`、tasks に `priority` 追加 |
| `003_add_column_color_task_priority.down.sql` | 追加カラム削除 |

### マルチテナントデータ分離

全テーブル (認証系を除く) に `tenant_id` カラムを持ち、全 SQL クエリに `tenant_id` フィルターを適用。
`tenant_id` はクライアントから送信される `X-Tenant-ID` ヘッダーの値を、JWT のクレームと照合して検証する。

---

## API 設計

### ベース URL

```
/api/v1
```

### 認証エンドポイント (公開)

| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/auth/register` | ユーザー登録 |
| POST | `/auth/login` | ログイン |
| POST | `/auth/refresh` | アクセストークン再発行 |
| POST | `/auth/logout` | ログアウト |

### ヘルスチェック

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/health` | サーバー稼働確認 |

### WebSocket

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/ws?board_id={id}` | ボード別リアルタイム接続 |

### テナントエンドポイント (JWT 認証必須)

| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/tenants` | テナント作成 |
| GET | `/tenants` | テナント一覧 |
| GET | `/tenants/:id` | テナント詳細 |
| PATCH | `/tenants/:id` | テナント更新 |
| GET | `/tenants/:id/members` | メンバー一覧 |
| POST | `/tenants/:id/members` | メンバー追加 |
| PATCH | `/tenants/:id/members/:uid` | メンバー権限変更 |
| DELETE | `/tenants/:id/members/:uid` | メンバー削除 |

### プロジェクトエンドポイント (JWT + テナントスコープ)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/projects` | プロジェクト一覧 |
| POST | `/projects` | プロジェクト作成 |
| GET | `/projects/:id` | プロジェクト詳細 |
| PATCH | `/projects/:id` | プロジェクト更新 |
| DELETE | `/projects/:id` | プロジェクトアーカイブ |
| GET | `/projects/:id/members` | プロジェクトメンバー一覧 |
| POST | `/projects/:id/members` | プロジェクトメンバー追加 |
| DELETE | `/projects/:id/members/:uid` | プロジェクトメンバー削除 |
| GET | `/projects/:id/boards` | プロジェクトのボード一覧 |

### ボードエンドポイント (JWT + テナントスコープ)

| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/boards` | ボード作成 |
| GET | `/boards/:id` | ボード詳細 (カラム・タスク含む) |
| PATCH | `/boards/:id` | ボード更新 |
| DELETE | `/boards/:id` | ボード削除 |

### カラムエンドポイント (JWT + テナントスコープ)

| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/columns` | カラム作成 |
| PATCH | `/columns/reorder` | カラム並び替え |
| PATCH | `/columns/:id` | カラム更新 |
| DELETE | `/columns/:id` | カラム削除 |

### タスクエンドポイント (JWT + テナントスコープ)

| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/tasks` | タスク作成 |
| PATCH | `/tasks/move` | タスク移動 (カラム間) |
| GET | `/tasks/:id` | タスク詳細 |
| PATCH | `/tasks/:id` | タスク更新 |
| DELETE | `/tasks/:id` | タスク削除 |
| POST | `/tasks/:id/labels` | タスクにラベル追加 |
| DELETE | `/tasks/:id/labels/:lid` | タスクからラベル削除 |
| POST | `/tasks/:id/comments` | コメント追加 |
| GET | `/tasks/:id/comments` | コメント一覧 |

### ラベルエンドポイント (JWT + テナントスコープ)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/projects/:id/labels` | プロジェクトのラベル一覧 |
| POST | `/projects/:id/labels` | ラベル作成 |
| DELETE | `/labels/:id` | ラベル削除 |

### ダッシュボードエンドポイント (JWT + テナントスコープ)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/dashboard/summary` | タスク概要統計 |
| GET | `/dashboard/overdue` | 期限切れタスク一覧 |
| GET | `/dashboard/my-tasks` | 自分の担当タスク一覧 |

**合計: 46 エンドポイント**

### 認証ヘッダー

```
Authorization: Bearer <access_token>
X-Tenant-ID: <tenant_uuid>
```

テナントスコープ付きのエンドポイントでは `X-Tenant-ID` ヘッダーが必須。
ミドルウェアが JWT の `tenant_id` クレームとヘッダー値を照合し、テナントメンバーシップを検証する。

---

## WebSocket 設計

### 接続方式

```
ws(s)://<host>/api/v1/ws?board_id=<uuid>
```

認証トークンは `Sec-WebSocket-Protocol` ヘッダーで送信:

```
Sec-WebSocket-Protocol: access_token.<jwt_token>
```

URL にトークンを含めず、プロトコルヘッダー経由で渡すことでセキュリティを確保。

### Hub パターン

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

- `HubManager`: ボード ID ごとに `Hub` を管理
- `Hub`: 同一ボードに接続しているクライアントにメッセージをブロードキャスト
- `Client`: goroutine ベースの読み書きポンプを実装

### イベント一覧

| イベント | トリガー | ペイロード |
|---------|---------|-----------|
| `task:created` | タスク作成 | TaskDetail オブジェクト |
| `task:updated` | タスク更新 | TaskDetail オブジェクト |
| `task:deleted` | タスク削除 | `{ id: string }` |
| `task:moved` | タスク移動 | 移動情報 |
| `column:created` | カラム作成 | Column オブジェクト |
| `column:updated` | カラム更新 | Column オブジェクト |
| `column:deleted` | カラム削除 | `{ id: string }` |
| `column:reordered` | カラム並び替え | 並び順情報 |

### クライアント側の処理

- REST ハンドラーが DB を更新後、`HubManager` 経由でブロードキャスト
- 受信側は `handleWSMessage` で状態を更新
- 自分の操作 (`user_id` が一致) はスキップして二重反映を防止
- `task:moved` と `column:*` イベントではボード全体を再取得
- 接続断は指数バックオフ (初期 1 秒、最大 30 秒) で自動再接続

---

## セキュリティ設計

### 認証フロー

```
1. ログイン
   POST /auth/login { email, password }
       ↓
   bcrypt で検証 (コスト 12)
       ↓
   JWT アクセストークン生成 (HS256, 15分)
   リフレッシュトークン生成 (ランダム 32バイト, SHA-256 ハッシュで DB 保存, 7日)
       ↓
   レスポンス: { access_token, user } + Set-Cookie (refresh_token, httpOnly)

2. API リクエスト
   Authorization: Bearer <access_token>
       ↓
   JWTAuth ミドルウェアで検証 → user_id, tenant_id を Context に設定

3. トークンリフレッシュ
   POST /auth/refresh (Cookie から refresh_token)
       ↓
   旧リフレッシュトークンを無効化 (ローテーション)
       ↓
   新しいアクセストークン + リフレッシュトークンを発行

4. フロントエンド自動リフレッシュ
   apiFetch() で 401 受信時に自動リフレッシュ → リトライ
```

### ロールベースアクセス制御 (RBAC)

| ロール | 権限 |
|-------|------|
| Owner | 全操作 + テナント設定変更 + メンバー管理 |
| Admin | 全操作 + メンバー管理 |
| Member | タスク・カラム・ボードの CRUD |
| Viewer | 読み取りのみ |

ミドルウェア (`TenantScope`) が `X-Tenant-ID` と JWT クレームを照合し、テナントメンバーシップを検証。
書き込み操作時はロールに応じたアクセス制御を適用。

### その他のセキュリティ対策

- パスワード: bcrypt コスト 12
- CORS: フロントエンドオリジンのみ許可 (カンマ区切りで複数指定可)
- サーバーエラーの内部情報を隠蔽 (汎用メッセージを返却)
- WebSocket 認証: プロトコルヘッダー経由 (URL にトークンを含めない)
- テナント分離: 全クエリに `tenant_id` フィルター
- Docker 本番イメージ: distroless (最小攻撃面)
- リフレッシュトークン: SHA-256 ハッシュで DB 保存 (生トークンは保存しない)

---

## 画面仕様

### ページ一覧

| ルート | ページ名 | 認証 | 説明 |
|-------|---------|------|------|
| `/{locale}` | ランディングページ | 不要 | アプリ紹介、ログインリンク |
| `/{locale}/login` | ログイン | 不要 | メール + パスワード入力 |
| `/{locale}/register` | ユーザー登録 | 不要 | 表示名 + メール + パスワード入力 |
| `/{locale}/ws/{slug}` | ダッシュボード | 必須 | タスク概要・優先度集計・プロジェクト一覧 |
| `/{locale}/ws/{slug}/projects` | プロジェクト一覧 | 必須 | プロジェクトの作成・編集・削除 |
| `/{locale}/ws/{slug}/settings` | ワークスペース設定 | 必須 | ワークスペース名の変更 |
| `/{locale}/ws/{slug}/members` | メンバー管理 | 必須 | メンバー一覧・権限変更・招待・削除 |
| `/{locale}/ws/{slug}/p/{id}/board` | カンバンボード | 必須 | ドラッグ&ドロップ対応カンバンボード |

### レイアウト構成

- **ルートレイアウト**: フォント・メタデータ設定
- **ロケールレイアウト**: `NextIntlClientProvider`・`Toaster` 設定
- **ワークスペースレイアウト**: `Header` + `Sidebar`、認証ガード、テナントコンテキスト管理

### 主要コンポーネント

| コンポーネント | 説明 |
|-------------|------|
| `KanbanBoard` | `DndContext` を管理、カラムの横並びレイアウト |
| `KanbanColumn` | `SortableContext` でタスクのソートを管理 |
| `TaskCard` | `useSortable` でドラッグ可能なタスクカード |
| `TaskDetailModal` | タスク詳細の表示・編集モーダル |
| `AddTaskForm` | カラム末尾のタスク追加フォーム |
| `ColumnHeader` | カラム名・色・WIP 制限の表示・編集 |
| `Header` | ロゴ・ハンバーガーメニュー・ユーザーメニュー |
| `Sidebar` | ナビゲーション・プロジェクト一覧 |
| `LocaleSwitcher` | 言語切替ボタン |

---

## 状態管理

Zustand による 4 つのストアでクライアント状態を管理。

| ストア | 責務 | 主要な状態 |
|-------|------|-----------|
| `authStore` | 認証状態 | `user`, `accessToken`, `tenantId`, `isAuthenticated` |
| `boardStore` | ボード状態 | `columns`, `selectedTask`, `comments`, `boardId` |
| `workspaceStore` | ワークスペース状態 | `tenant`, `projects`, `members`, `currentUserRole` |
| `dashboardStore` | ダッシュボード状態 | `summary`, `overdueTasks`, `myTasks` |

### 楽観的更新パターン

ドラッグ&ドロップなどの操作では、即座にストアを更新し、REST API 呼び出し後に失敗した場合のみロールバック:

```
1. ストア即時更新 (楽観的)
2. REST API 呼び出し
3. 成功 → そのまま
4. 失敗 → 元の状態にロールバック
```

`createTask`、`moveTask`、`deleteTask`、`deleteColumn` で適用。

---

## 国際化 (i18n)

### 対応言語

| コード | 言語 |
|-------|------|
| `ja` | 日本語 |
| `en` | 英語 |

### 実装方式

- `next-intl` によるルーティングベースのロケール管理
- URL パスのプレフィックス (`/ja/...`, `/en/...`) でロケール決定
- サーバーコンポーネント・クライアントコンポーネント両方で翻訳を使用
- 約 100 個の翻訳キー (ナビゲーション、フォーム、エラーメッセージ、ボード操作など)

### 翻訳ファイル構成

```
messages/
├── en.json    # 英語
└── ja.json    # 日本語
```

---

## CI/CD パイプライン

### GitHub Actions ワークフロー

`main` ブランチへの push と Pull Request で自動実行。パス指定により変更のあったコンポーネントのみ実行。

#### バックエンド CI (`backend-ci.yml`)

```
トリガー: backend/** の変更
実行環境: ubuntu-latest
サービスコンテナ: postgres:16-alpine

ステップ:
1. Go 1.22 セットアップ (依存キャッシュ付き)
2. go mod download
3. go vet ./...
4. go build ./...
5. SQL マイグレーション実行
6. go test ./... -v -count=1
```

#### フロントエンド CI (`frontend-ci.yml`)

```
トリガー: frontend/** の変更
実行環境: ubuntu-latest

ステップ:
1. Node.js 20 セットアップ (npm キャッシュ付き)
2. npm ci
3. npx tsc --noEmit (型チェック)
4. npm run lint (ESLint)
5. npm run build (ビルド確認)
```

---

## セットアップ手順

### 前提条件

- Go 1.22 以上
- Node.js 20 以上
- Docker / Docker Compose
- sqlc (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### ローカル開発

```bash
# 1. リポジトリのクローンと PostgreSQL 起動
git clone https://github.com/mer-prog/taskflow.git
cd taskflow
docker compose up db -d

# 2. マイグレーション実行
psql "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable" \
  -f backend/db/migrations/001_init.up.sql \
  -f backend/db/migrations/002_add_project_archived_at.up.sql \
  -f backend/db/migrations/003_add_column_color_task_priority.up.sql

# 3. (任意) デモデータ投入
psql "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable" \
  -f backend/db/seed.sql

# 4. バックエンド起動
cp backend/.env.example backend/.env
cd backend && go run ./cmd/server

# 5. フロントエンド起動 (別ターミナル)
cd frontend
cp .env.example .env.local
npm install && npm run dev
```

### デモアカウント

`seed.sql` 実行後に使用可能:

| メールアドレス | パスワード | ロール |
|-------------|-----------|-------|
| `demo@taskflow.app` | `demo1234` | Owner |
| `alice@taskflow.app` | `demo1234` | Admin |
| `bob@taskflow.app` | `demo1234` | Member |

### Docker Compose (全体起動)

```bash
docker compose up --build
```

PostgreSQL (ポート 5432) と Go API (ポート 8080) が起動する。

### 環境変数

#### バックエンド

| 変数名 | 説明 | デフォルト |
|-------|------|-----------|
| `PORT` | サーバーポート | `8080` |
| `ENV` | 実行環境 (`development` / `production`) | `development` |
| `DATABASE_URL` | PostgreSQL 接続文字列 (DB_* 変数より優先) | -- |
| `DB_HOST` | データベースホスト | `localhost` |
| `DB_PORT` | データベースポート | `5432` |
| `DB_USER` | データベースユーザー | `taskflow` |
| `DB_PASSWORD` | データベースパスワード | `taskflow` |
| `DB_NAME` | データベース名 | `taskflow` |
| `DB_SSLMODE` | SSL モード | `disable` |
| `JWT_SECRET` | JWT 署名鍵 | (開発用デフォルト) |
| `JWT_ACCESS_EXPIRY` | アクセストークン有効期限 | `15m` |
| `JWT_REFRESH_EXPIRY` | リフレッシュトークン有効期限 | `168h` |
| `CORS_ORIGIN` | 許可オリジン (カンマ区切り) | `http://localhost:3000` |

#### フロントエンド

| 変数名 | 説明 |
|-------|------|
| `NEXT_PUBLIC_API_URL` | バックエンド API ベース URL |
| `NEXT_PUBLIC_WS_URL` | WebSocket エンドポイント URL |

### デプロイ (Render + Vercel)

1. リポジトリを [Render](https://render.com) に接続 → `render.yaml` Blueprint により Web Service + PostgreSQL が自動構築
2. `taskflow-env` 環境変数グループに `DATABASE_URL` を設定
3. Render データベースに対してマイグレーション実行
4. `frontend/` を [Vercel](https://vercel.com) にインポート、環境変数を設定:
   - `NEXT_PUBLIC_API_URL` = `https://taskflow-api.onrender.com/api/v1`
   - `NEXT_PUBLIC_WS_URL` = `wss://taskflow-api.onrender.com/api/v1/ws`

### デプロイ (AWS / IaC のみ)

Terraform コードは `infra/terraform/` に格納。実際のデプロイは未実施。

```bash
cd infra/terraform
terraform init
terraform plan -var="db_password=..." -var="container_image=..."
```

---

## 設計上の意思決定

### Adapter パターンの採用

sqlc が生成するコードは `repository` パッケージに閉じている。Service 層は独自のインターフェースを定義し、Adapter がそのインターフェースを実装することで、sqlc 生成コードへの依存を分離した。これにより、テスト時にモックリポジトリを容易に差し替えることができる。

### 楽観的更新の採用

ドラッグ&ドロップ操作ではユーザー体験が最重要。先に UI を更新し、バックグラウンドで API 呼び出しを行い、失敗時のみロールバックする方式を採用。

### WebSocket の Hub パターン

ボード ID ごとに独立した Hub を持つことで、無関係なボードへのメッセージ配信を防止。goroutine ベースの読み書きポンプにより、並行接続を効率的に処理。

### sqlc による型安全な SQL

ORM ではなく sqlc を採用し、生の SQL をそのまま記述。SQL の最適化が容易で、生成された Go コードにより型安全性を確保。

### マルチテナント分離方式

共有データベース・共有スキーマ方式を採用。全テーブルに `tenant_id` を持ち、ミドルウェアで JWT クレームとの整合性を検証。テナント単位のスキーマ分離と比較してインフラコストを抑制。

### JWT + リフレッシュトークンローテーション

アクセストークンはメモリ保持 (httpOnly Cookie ではない) でフロントエンドの柔軟性を確保。リフレッシュトークンは httpOnly Cookie で保護し、リフレッシュ時に旧トークンを無効化するローテーション方式でセキュリティを強化。

### distroless Docker イメージ

本番ビルドでは `gcr.io/distroless/static-debian12:nonroot` を使用。シェルやパッケージマネージャーを含まない最小イメージで攻撃面を削減。

### 単一 NAT Gateway

コスト削減のため、AWS 構成では単一 AZ に NAT Gateway を配置。高可用性が必要な場合は各 AZ に NAT Gateway を追加可能。

---

## ランニングコスト

### デモ環境 (Render + Vercel)

| サービス | 月額費用 |
|---------|---------|
| Render Web Service | 無料 |
| Render PostgreSQL | 無料 |
| Vercel (フロントエンド) | 無料 |
| **合計** | **$0/月** |

※ Render 無料プランは 15 分間アクセスがないとスリープする制約あり。

### 本番環境 (AWS / 推定)

| サービス | 構成 | 推定月額 |
|---------|------|---------|
| ECS Fargate | 256 CPU / 512 MB × 2 タスク | ~$18 |
| RDS PostgreSQL | db.t3.micro / gp3 20GB / 単一 AZ | ~$15 |
| NAT Gateway | 単一 AZ | ~$32 |
| ALB | 1 基 | ~$16 |
| CloudWatch Logs | 30 日保持 | ~$3 |
| データ転送 | 推定 | ~$3 |
| **合計** | | **~$87/月** |

※ us-west-2 リージョン、東京リージョンではやや高くなる。
※ Vercel (フロントエンド) は無料プランを想定。

---

## 著者

- GitHub: [mer-prog](https://github.com/mer-prog)
- リポジトリ: [mer-prog/taskflow](https://github.com/mer-prog/taskflow)

---

## ライセンス

MIT
