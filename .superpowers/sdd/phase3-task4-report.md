# Phase 3 Task 4 Report

## Scope

执行 `.superpowers/sdd/phase3-task4-brief.md`：为活动分享页实现公开的 OG HTML 页面 `GET /e/:slug`，输出基础 OG/Twitter meta 和前端 hydration root。

## Changes

- 扩展 `apps/api/internal/events/handler.go`。
  - 新增 `HandleOGPage`，通过 slug 读取活动详情。
  - 使用 `html/template` 渲染最小 HTML shell，模板自动转义标题、时间描述和地点。
  - 输出 `og:title`、`og:type=website`、`og:description`、`twitter:card=summary_large_image`、Twitter 标题和描述。
  - 输出 `<div id="root" data-event-slug="..."></div>`，供前端后续 hydration 使用。
- 扩展 `apps/api/internal/http/router.go`。
  - 在 `/api` group 之外注册公开路由 `GET /e/:slug`。
- 新增 `apps/api/internal/events/og_handler_test.go`。
  - 覆盖 OG/Twitter meta。
  - 覆盖 event title、time、location 安全转义。
  - 覆盖 hydration root 存在。

## Verification

TDD red check:

```bash
go test ./apps/api/internal/events -run TestOGPage -v
```

Result:

- Exit code: `1`
- 预期失败，原因是 `HandleOGPage` 尚未实现。

Task verification:

```bash
go test ./apps/api/internal/events -run TestOGPage -v
```

Result:

- Exit code: `0`
- `TestOGPage`: PASS

API package:

```bash
go test ./apps/api/...
```

Result:

- Exit code: `0`
- `cmd/api`: no test files
- `internal/auth`: ok
- `internal/config`: no test files
- `internal/db`: no test files
- `internal/events`: ok
- `internal/http`: ok
- `internal/schedules`: no test files
- `internal/users`: ok

## Notes

- 分享页当前是最小 HTML shell；前端 bundle 注入、完整页面渲染和图片类 OG meta 未包含在本任务范围内。
