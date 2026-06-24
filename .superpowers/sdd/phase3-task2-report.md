# Phase 3 Task 2 Report

## Scope

执行 `.superpowers/sdd/phase3-task2-brief.md`：新增活动创建、匿名活动详情和我的活动 API。

## Changes

- 新增 `apps/api/internal/events/service_test.go`。
  - 覆盖活动创建生成 10 位 URL-safe `share_slug`。
  - 覆盖 `end_at <= start_at` 校验拒绝。
  - 覆盖匿名详情返回 going 参与人数。
  - 覆盖基于 `coalesce(end_at,start_at) < now` 的过期状态派生。
- 新增 `apps/api/internal/events/model.go`、`repository.go`、`service.go`。
  - 定义 `Event`、`Participant`、`EventDetail`、`CreateEventInput`。
  - 实现 `Create`、`FindBySlug`、`FindByID`、`CountGoing`、`ListMine`。
  - 使用 `crypto/rand` 生成 10 位 URL-safe 分享 slug。
  - 实现活动字段校验和只在响应层派生的 expired 状态。
- 新增 `apps/api/internal/events/handler.go`。
  - 实现 authenticated `POST /api/events`。
  - 实现 public `GET /api/events/:slug`。
  - 实现 authenticated `GET /api/events/mine`。
- 更新 `apps/api/internal/http/router.go` 注册 events 路由。
- 新增 `apps/api/internal/http/events_router_test.go`，验证匿名详情路由无需 Authorization header。

## Verification

TDD red check:

```bash
go test ./apps/api/internal/events -run 'TestCreate|TestGetDetail' -v
```

Result:

- Exit code: `1`
- 预期失败，原因是 `NewService`、`CreateEventInput`、`Event` 等 events 实现尚不存在。

Events package:

```bash
go test ./apps/api/internal/events -v
```

Result:

- Exit code: `0`
- `TestCreateGeneratesUniqueShareSlug`: PASS
- `TestCreateRejectsEndAtBeforeOrEqualStartAt`: PASS
- `TestGetDetailReturnsParticipantCountsWithoutAuth`: PASS
- `TestGetDetailDerivesExpiredStatusFromEndOrStartAt`: PASS

Public route regression:

```bash
go test ./apps/api/internal/http -run TestEventDetailRouteIsPublic -v
```

Result:

- Exit code: `0`
- `TestEventDetailRouteIsPublic`: PASS

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
- `internal/users`: ok

## Notes

- `GET /api/events/:slug` 挂在未加 auth middleware 的 `/api` group 下；`POST /api/events` 和 `GET /api/events/mine` 挂在 authenticated group 下。
- 过期状态是读取时派生，不回写数据库，避免详情接口产生副作用。
