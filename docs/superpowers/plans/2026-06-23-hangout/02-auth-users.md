# Hangout Phase 2 Auth And Users Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 接入 Supabase Auth，建立 `public.users` 档案表、Auth insert trigger、Gin JWT 验签和前端登录注册流程。

**Architecture:** Supabase 负责邮箱注册、登录、登出和重置密码；Gin 中间件用 `SUPABASE_JWT_SECRET` 验 HS256 token，并把 `sub` 注入 request context。业务用户档案只读写 `public.users`，`auth.users` 仅通过 Supabase Auth/Admin 能力管理。

**Tech Stack:** Supabase Auth, Go JWT middleware, GORM, React Router, supabase-js。

## Global Constraints

- 认证由 Supabase Auth 托管，P0 只做邮箱+密码注册/登录/重置密码。
- Gin 后端只验签 Supabase JWT，不自行签发业务 token。
- `public.users.id` 为 uuid，等于 Supabase `auth.users.id`。
- 前端只允许使用 `supabase-js` 的 Auth/session 能力。
- 业务数据统一通过 Gin `/api/*` 访问。

---

## File Structure

- Create: `supabase/migrations/202606230002_auth_users.sql`
- Create: `apps/api/internal/auth/middleware.go`
- Create: `apps/api/internal/auth/middleware_test.go`
- Create: `apps/api/internal/users/model.go`
- Create: `apps/api/internal/users/repository.go`
- Create: `apps/api/internal/users/service.go`
- Create: `apps/api/internal/users/handler.go`
- Create: `apps/api/internal/users/handler_test.go`
- Modify: `apps/api/internal/http/router.go`
- Create: `apps/web/src/lib/supabase.ts`
- Modify: `apps/web/src/lib/api.ts`
- Create: `apps/web/src/routes/LoginPage.tsx`
- Create: `apps/web/src/routes/ProfilePage.tsx`
- Modify: `apps/web/src/App.tsx`

## Backend Interfaces

- `auth.UserIDFromContext(ctx context.Context) (uuid.UUID, bool)`
- `auth.RequireUser() gin.HandlerFunc`
- `users.Service.GetMe(ctx context.Context, userID uuid.UUID) (users.Profile, error)`
- `GET /api/auth/me`
- `GET /api/users/me`
- `PATCH /api/users/me`
- `GET /api/users/search?q=`

### Task 1: Supabase profile 表和 trigger

- [ ] **Step 1: 写迁移**

Create `supabase/migrations/202606230002_auth_users.sql`:

```sql
create table public.users (
  id uuid primary key references auth.users(id) on delete cascade,
  display_name text not null default '',
  avatar_url text,
  email text,
  status text not null default 'active' check (status in ('active', 'deactivated')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz
);

create index users_email_idx on public.users (lower(email)) where deleted_at is null;
create index users_display_name_idx on public.users (lower(display_name)) where deleted_at is null;

create or replace function public.handle_new_auth_user()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  insert into public.users (id, display_name, email)
  values (
    new.id,
    coalesce(new.raw_user_meta_data ->> 'display_name', split_part(new.email, '@', 1), '新用户'),
    new.email
  )
  on conflict (id) do update
    set email = excluded.email,
        updated_at = now();
  return new;
end;
$$;

drop trigger if exists on_auth_user_created on auth.users;
create trigger on_auth_user_created
after insert on auth.users
for each row execute procedure public.handle_new_auth_user();
```

- [ ] **Step 2: 验证迁移**

Run:

```bash
supabase db reset
```

Expected: `public.users` 存在，trigger 创建成功。

### Task 2: JWT 鉴权中间件

- [ ] **Step 1: 写失败测试**

Create `apps/api/internal/auth/middleware_test.go` with tests for missing token, invalid token, and valid token. Valid token claims must include `sub`, `aud=authenticated`, and future `exp`.

Run:

```bash
go test ./apps/api/internal/auth -v
```

Expected: FAIL because middleware package is not implemented.

- [ ] **Step 2: 实现中间件**

Create `apps/api/internal/auth/middleware.go`:

```go
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "user_id"

func RequireUser(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"data": nil, "error": gin.H{"code": "unauthorized", "message": "missing bearer token"}})
			c.Abort()
			return
		}
		tokenText := strings.TrimPrefix(header, "Bearer ")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"data": nil, "error": gin.H{"code": "unauthorized", "message": "invalid bearer token"}})
			c.Abort()
			return
		}
		sub, ok := claims["sub"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"data": nil, "error": gin.H{"code": "unauthorized", "message": "token subject is missing"}})
			c.Abort()
			return
		}
		userID, err := uuid.Parse(sub)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"data": nil, "error": gin.H{"code": "unauthorized", "message": "token subject is invalid"}})
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), userIDKey, userID))
		c.Next()
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
```

- [ ] **Step 3: 加依赖并验证**

Run:

```bash
cd apps/api
go get github.com/golang-jwt/jwt/v5 github.com/google/uuid
go test ./internal/auth -v
```

Expected: PASS.

### Task 3: 用户档案 API

- [ ] **Step 1: 定义 model/repository/service**

Create:

- `apps/api/internal/users/model.go` with `Profile{ID uuid.UUID, DisplayName string, AvatarURL *string, Email *string, Status string}`
- `apps/api/internal/users/repository.go` with `FindByID`, `UpdateMe`, `Search`
- `apps/api/internal/users/service.go` with validation: display name 1-40 chars, avatar URL optional, search query min 2 chars

- [ ] **Step 2: 写 handler 测试**

Create `apps/api/internal/users/handler_test.go` covering:

- `GET /api/auth/me` returns the profile for token `sub`
- `PATCH /api/users/me` updates `display_name` and `avatar_url`
- `GET /api/users/search?q=li` excludes deactivated users

Run:

```bash
go test ./apps/api/internal/users -v
```

Expected: FAIL until handlers and repository test database setup exist.

- [ ] **Step 3: 实现 handler 并注册路由**

Create `apps/api/internal/users/handler.go` with:

- `HandleAuthMe`
- `HandleGetMe`
- `HandlePatchMe`
- `HandleSearch`

Modify `apps/api/internal/http/router.go` so authenticated routes use `auth.RequireUser(cfg.SupabaseJWTSecret)`.

- [ ] **Step 4: 验证**

Run:

```bash
go test ./apps/api/...
```

Expected: PASS.

### Task 4: 前端 Auth 流程

- [ ] **Step 1: 安装依赖**

Run:

```bash
cd apps/web
npm install @supabase/supabase-js react-router-dom
```

- [ ] **Step 2: 创建 Supabase Auth 客户端**

Create `apps/web/src/lib/supabase.ts`:

```ts
import { createClient } from '@supabase/supabase-js'

export const supabase = createClient(
  import.meta.env.VITE_SUPABASE_URL as string,
  import.meta.env.VITE_SUPABASE_ANON_KEY as string,
)
```

- [ ] **Step 3: 修改 API 客户端附加 token**

Modify `apps/web/src/lib/api.ts` so each request reads `supabase.auth.getSession()` and adds `Authorization` when session exists. Keep all business requests going through `fetch`.

- [ ] **Step 4: 创建登录页和个人页**

Create `LoginPage.tsx` with email/password login, register, reset password, and sign out actions through `supabase.auth.*`. Create `ProfilePage.tsx` that calls `/api/users/me`.

- [ ] **Step 5: 验证前端构建**

Run:

```bash
cd apps/web
npm run build
```

Expected: PASS.

### Commit

- [ ] **Step 1: 阶段验证**

Run:

```bash
supabase db reset
go test ./apps/api/...
cd apps/web && npm run build
```

Expected: 全部 PASS。

- [ ] **Step 2: 提交**

Run:

```bash
git add apps/api apps/web supabase
git commit -m "feat: add supabase auth and user profiles"
```
