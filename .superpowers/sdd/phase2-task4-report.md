# Phase 2 Task 4 Report

## Scope

执行 `.superpowers/sdd/phase2-task4-brief.md`：新增前端 Supabase Auth 流程，登录/个人页入口，以及 API 客户端 bearer token 透传。

## Changes

- 更新 `apps/web/package.json` 和 `apps/web/package-lock.json`。
  - 安装 `@supabase/supabase-js`。
  - 安装 `react-router-dom`。
- 新增 `apps/web/src/lib/supabase.ts`。
  - 使用 `VITE_SUPABASE_URL` 和 `VITE_SUPABASE_ANON_KEY` 创建 Supabase client。
- 更新 `apps/web/src/lib/api.ts`。
  - 每次业务请求前读取 `supabase.auth.getSession()`。
  - session 存在时设置 `Authorization: Bearer <access_token>`。
  - 所有业务请求仍通过 `fetch` 发出。
- 更新 `apps/web/src/lib/api.test.ts`。
  - 覆盖无 session 时空 headers 请求。
  - 覆盖有 session 时 bearer token 注入。
- 新增 `apps/web/src/LoginPage.tsx`。
  - 支持邮箱/密码登录、注册、重置密码和退出登录。
- 新增 `apps/web/src/ProfilePage.tsx`。
  - 调用 `/api/users/me` 加载当前用户档案。
- 更新 `apps/web/src/App.tsx`、`apps/web/src/main.tsx` 和 `apps/web/src/App.test.tsx`。
  - 接入 React Router。
  - 注册 `/`、`/login`、`/profile`。
  - 覆盖登录页路由和个人页加载行为。

## Verification

RED:

```bash
cd apps/web
npm test -- src/lib/api.test.ts
```

Result:

- Exit code: `1`
- 失败原因：`Failed to resolve import "./supabase"`，确认缺少 Supabase client。

GREEN:

```bash
cd apps/web
npm test -- src/lib/api.test.ts
```

Result:

- Exit code: `0`
- `src/lib/api.test.ts`: 4 tests PASS。

Route RED:

```bash
cd apps/web
npm test -- src/App.test.tsx
```

Result:

- Exit code: `1`
- 失败原因：`/login`、`/profile` 仍渲染旧健康页，确认路由未实现。

Route GREEN:

```bash
cd apps/web
npm test -- src/App.test.tsx
```

Result:

- Exit code: `0`
- `src/App.test.tsx`: 4 tests PASS。

Frontend tests:

```bash
cd apps/web
npm test
```

Result:

- Exit code: `0`
- Test files: 2 passed.
- Tests: 8 passed.

Brief build check:

```bash
cd apps/web
npm run build
```

Result:

- Exit code: `0`
- `tsc -b && vite build`: PASS。

Phase verification:

```bash
supabase db reset
go test ./apps/api/...
cd apps/web && npm run build
```

Result:

- `supabase db reset`: exit code `0`，local database reset 完成；提示 `WARN: no files matched pattern: supabase/seed.sql`。
- `go test ./apps/api/...`: exit code `0`，`internal/auth`、`internal/http`、`internal/users` PASS。
- `npm run build`: exit code `0`，前端生产构建 PASS。

## Notes

- `npm install` 期间出现 Node engine warning：当前 Node 为 `v23.11.0`，部分 ESLint 包要求 `^20.19.0 || ^22.13.0 || >=24`；安装命令最终 exit code 为 `0`。
- `npm install` 报告 1 个 low severity vulnerability；本任务未执行 `npm audit fix`，避免引入非 brief 范围变更。
