# Phase 2 Task 3 Report

## Scope

执行 `.superpowers/sdd/phase2-task3-brief.md`：新增用户档案 API，包括 profile model/repository/service、认证 handler、路由注册与测试。

## Changes

- 新增 `apps/api/internal/users/model.go`。
  - 定义 `Profile{ID, DisplayName, AvatarURL, Email, Status}`，映射 `public.users`。
- 新增 `apps/api/internal/users/repository.go`。
  - 实现 `FindByID`、`UpdateMe`、`Search`。
  - 查询排除 `deleted_at IS NOT NULL` 记录，搜索仅返回 `status = active` 用户。
- 新增 `apps/api/internal/users/service.go`。
  - 校验 display name 为 1-40 个字符。
  - 校验可选 avatar URL。
  - 校验 search query 最少 2 个字符。
- 新增 `apps/api/internal/users/handler_test.go`。
  - 覆盖 `GET /api/auth/me` 根据 token `sub` 返回当前档案。
  - 覆盖 `PATCH /api/users/me` 更新 `display_name` 和 `avatar_url`。
  - 覆盖 `GET /api/users/search?q=li` 排除 deactivated 用户。
- 新增 `apps/api/internal/users/handler.go`。
  - 实现 `HandleAuthMe`、`HandleGetMe`、`HandlePatchMe`、`HandleSearch`。
- 更新 `apps/api/internal/http/router.go`。
  - 注册认证用户路由，并使用 `auth.RequireUser(cfg.SupabaseJWTSecret)`。
- 更新 `apps/api/internal/config/config.go` 和 `apps/api/cmd/api/main.go`。
  - 加载 `SUPABASE_JWT_SECRET`。
  - 将 config 和数据库连接传入 router dependencies。

## Verification

Baseline:

```bash
cd apps/api
go test ./...
```

Result:

- Exit code: `0`
- `internal/auth`: PASS
- `internal/http`: PASS

RED:

```bash
cd apps/api
go test ./internal/users -v
```

Result:

- Exit code: `1`
- 失败原因：`undefined: NewHandler`。

GREEN:

```bash
cd apps/api
go test ./internal/users -v
```

Result:

- Exit code: `0`
- `TestHandleAuthMeReturnsProfileForTokenSubject`: PASS
- `TestHandlePatchMeUpdatesDisplayNameAndAvatarURL`: PASS
- `TestHandleSearchExcludesDeactivatedUsers`: PASS

Final:

```bash
go test ./apps/api/...
```

Result:

- Exit code: `0`
- `cmd/api`, `internal/config`, `internal/db`: no test files.
- `internal/auth`: PASS.
- `internal/http`: PASS.
- `internal/users`: PASS.

## Notes

- 曾尝试安装 `gorm.io/driver/sqlite` 做 test database setup，但 `proxy.golang.org` 超时；最终 handler 测试改用 in-memory fake repository，避免新增网络依赖。
- 当前 repository 仍基于 GORM 实现，生产路径使用 PostgreSQL 连接。
