# Hangout Phase 1 Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立可运行的 Go/Gin 后端、React/Vite 前端、Supabase local 与基础数据库迁移。

**Architecture:** 单仓库划分 `apps/api`、`apps/web`、`supabase`、`deploy`。后端暴露 `/api/health` 并连接 PostgreSQL；前端通过 Vite dev server 调用后端健康检查，先不接入认证。

**Tech Stack:** Go 1.23+, Gin, GORM, pgx driver, React 18, TypeScript, Vite, Tailwind CSS, shadcn/ui, Supabase CLI, PostgreSQL。

## Global Constraints

- Web 应用移动端优先，UI 风格对标 Luma。
- 认证由 Supabase Auth 托管，Gin 后端只验签 Supabase JWT。
- 前端不得使用 `supabase.from(...)` 或 `supabase.rpc(...)` 直接读写业务表/业务逻辑。
- 开发、测试、生产统一使用 PostgreSQL 语义；不引入 SQLite。
- 时间统一 UTC 存储，前端按 Asia/Shanghai 展示。

---

## File Structure

- Create: `go.work`，聚合 `apps/api` Go module。
- Create: `apps/api/go.mod`，后端依赖定义。
- Create: `apps/api/cmd/api/main.go`，Gin 启动入口。
- Create: `apps/api/internal/config/config.go`，读取环境变量。
- Create: `apps/api/internal/db/db.go`，初始化 GORM PostgreSQL 连接。
- Create: `apps/api/internal/http/router.go`，注册路由、中间件和统一响应。
- Create: `apps/api/internal/http/health_handler.go`，实现 `/api/health`。
- Create: `apps/api/internal/http/health_handler_test.go`，验证健康检查响应。
- Create: `apps/web/package.json`、`apps/web/vite.config.ts`、`apps/web/src/*`，前端骨架。
- Create: `apps/web/src/lib/api.ts`，封装 Gin API 请求。
- Create: `supabase/config.toml`，Supabase local 配置。
- Create: `supabase/migrations/202606230001_initial_schema.sql`，创建应用 schema 的扩展和基础枚举约束起点。
- Create: `.env.example`，列出本地和生产必须配置的环境变量。

### Task 1: 后端骨架与健康检查

**Files:**
- Create: `go.work`
- Create: `apps/api/go.mod`
- Create: `apps/api/cmd/api/main.go`
- Create: `apps/api/internal/config/config.go`
- Create: `apps/api/internal/db/db.go`
- Create: `apps/api/internal/http/router.go`
- Create: `apps/api/internal/http/response.go`
- Create: `apps/api/internal/http/health_handler.go`
- Create: `apps/api/internal/http/health_handler_test.go`

**Interfaces:**
- Produces: `http.NewRouter(deps Dependencies) *gin.Engine`
- Produces: `config.Load() (config.Config, error)`
- Produces: `db.Open(ctx context.Context, cfg config.Config) (*gorm.DB, error)`

- [ ] **Step 1: 初始化 Go module**

Run:

```bash
mkdir -p apps/api/cmd/api apps/api/internal/config apps/api/internal/db apps/api/internal/http
cd apps/api
go mod init github.com/bytedance/calandar/apps/api
go get github.com/gin-gonic/gin gorm.io/gorm gorm.io/driver/postgres github.com/jackc/pgx/v5
cd ../..
go work init ./apps/api
```

Expected: `apps/api/go.mod` 和 `go.work` 生成成功。

- [ ] **Step 2: 写健康检查失败测试**

Add to `apps/api/internal/http/health_handler_test.go`:

```go
package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	router := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	const want = `{"data":{"status":"ok"},"error":null}`
	if rec.Body.String() != want {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
```

Run:

```bash
go test ./apps/api/internal/http -run TestHealth -v
```

Expected: FAIL because `NewRouter` is undefined.

- [ ] **Step 3: 实现统一响应和健康检查**

Add to `apps/api/internal/http/response.go`:

```go
package http

import "github.com/gin-gonic/gin"

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

func respondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"data": nil, "error": ErrorBody{Code: code, Message: message}})
}
```

Add to `apps/api/internal/http/router.go`:

```go
package http

import "github.com/gin-gonic/gin"

type Dependencies struct{}

func NewRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	api.GET("/health", healthHandler)
	return router
}
```

Add to `apps/api/internal/http/health_handler.go`:

```go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func healthHandler(c *gin.Context) {
	respondOK(c, http.StatusOK, gin.H{"status": "ok"})
}
```

- [ ] **Step 4: 添加启动入口与配置读取**

Add to `apps/api/internal/config/config.go`:

```go
package config

import (
	"errors"
	"os"
)

type Config struct {
	Addr        string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:        getenv("API_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
```

Add to `apps/api/internal/db/db.go`:

```go
package db

import (
	"context"

	"github.com/bytedance/calandar/apps/api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(ctx context.Context, cfg config.Config) (*gorm.DB, error) {
	conn, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}
	return conn, nil
}
```

Add to `apps/api/cmd/api/main.go`:

```go
package main

import (
	"context"
	"log"

	"github.com/bytedance/calandar/apps/api/internal/config"
	"github.com/bytedance/calandar/apps/api/internal/db"
	apihttp "github.com/bytedance/calandar/apps/api/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Open(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
	router := apihttp.NewRouter(apihttp.Dependencies{})
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 5: 验证后端测试**

Run:

```bash
go test ./apps/api/...
```

Expected: PASS.

### Task 2: 前端骨架与 API 客户端

**Files:**
- Create: `apps/web/package.json`
- Create: `apps/web/index.html`
- Create: `apps/web/vite.config.ts`
- Create: `apps/web/tsconfig.json`
- Create: `apps/web/src/main.tsx`
- Create: `apps/web/src/App.tsx`
- Create: `apps/web/src/lib/api.ts`
- Create: `apps/web/src/index.css`

**Interfaces:**
- Consumes: `GET /api/health`
- Produces: `apiGet<T>(path: string): Promise<T>`

- [ ] **Step 1: 创建 Vite React 项目**

Run:

```bash
npm create vite@latest apps/web -- --template react-ts
cd apps/web
npm install
npm install -D tailwindcss postcss autoprefixer vitest @testing-library/react @testing-library/jest-dom jsdom
npx tailwindcss init -p
```

Expected: `apps/web` 可以执行 `npm run dev`。

- [ ] **Step 2: 添加 API 客户端**

Add to `apps/web/src/lib/api.ts`:

```ts
export type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(path)
  const body = (await response.json()) as ApiEnvelope<T>
  if (!response.ok || body.error) {
    throw new Error(body.error?.message ?? `Request failed: ${response.status}`)
  }
  if (body.data === null) {
    throw new Error('Response data is empty')
  }
  return body.data
}
```

- [ ] **Step 3: 显示健康检查结果**

Replace `apps/web/src/App.tsx` with:

```tsx
import { useEffect, useState } from 'react'
import { apiGet } from './lib/api'
import './index.css'

type Health = { status: string }

export default function App() {
  const [status, setStatus] = useState('checking')

  useEffect(() => {
    apiGet<Health>('/api/health')
      .then((data) => setStatus(data.status))
      .catch(() => setStatus('unavailable'))
  }, [])

  return (
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-md rounded-3xl bg-white/10 p-6 shadow-xl">
        <p className="text-sm text-white/60">Hangout</p>
        <h1 className="mt-3 text-3xl font-semibold">社交日历</h1>
        <p className="mt-4 text-white/70">API status: {status}</p>
      </section>
    </main>
  )
}
```

- [ ] **Step 4: 验证前端构建**

Run:

```bash
cd apps/web
npm run build
```

Expected: PASS.

### Task 3: Supabase local 与初始迁移

**Files:**
- Create: `supabase/config.toml`
- Create: `supabase/migrations/202606230001_initial_schema.sql`
- Create: `.env.example`

**Interfaces:**
- Produces: PostgreSQL extension `pgcrypto`
- Produces: `public.users` 以外的表在后续迁移中增量创建

- [ ] **Step 1: 初始化 Supabase**

Run:

```bash
supabase init
```

Expected: `supabase/config.toml` 存在。

- [ ] **Step 2: 添加初始迁移**

Add to `supabase/migrations/202606230001_initial_schema.sql`:

```sql
create extension if not exists pgcrypto;

create table if not exists public.schema_migrations_marker (
  id bigserial primary key,
  name text not null unique,
  applied_at timestamptz not null default now()
);

insert into public.schema_migrations_marker (name)
values ('202606230001_initial_schema')
on conflict (name) do nothing;
```

- [ ] **Step 3: 添加环境变量样例**

Add to `.env.example`:

```dotenv
API_ADDR=:8080
DATABASE_URL=postgres://postgres:postgres@127.0.0.1:54322/postgres?sslmode=disable
SUPABASE_URL=http://127.0.0.1:54321
SUPABASE_ANON_KEY=replace-with-local-anon-key
SUPABASE_JWT_SECRET=replace-with-local-jwt-secret
SUPABASE_SERVICE_ROLE_KEY=replace-with-local-service-role-key
VITE_SUPABASE_URL=http://127.0.0.1:54321
VITE_SUPABASE_ANON_KEY=replace-with-local-anon-key
VITE_API_BASE_URL=
```

- [ ] **Step 4: 验证迁移可执行**

Run:

```bash
supabase start
supabase db reset
```

Expected: reset 成功，PostgreSQL 可连接。

### Commit

- [ ] **Step 1: 运行阶段验证**

Run:

```bash
go test ./apps/api/...
cd apps/web && npm run build
```

Expected: 两条命令均 PASS。

- [ ] **Step 2: 提交**

Run:

```bash
git add go.work apps/api apps/web supabase .env.example
git commit -m "chore: add hangout foundation"
```
