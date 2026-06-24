# Phase 3 Task 3 Report

## Scope

执行 `.superpowers/sdd/phase3-task3-brief.md`：实现 RSVP 与自动日程联动，包括参与记录 upsert、容量限制、event schedule upsert/delete、冲突返回和 RSVP/cancel API 注册。

## Changes

- 扩展 `apps/api/internal/events/model.go`。
  - 新增 RSVP 状态常量、participant source 常量。
  - 新增 `RSVPInput`、`RSVPResult`、`ScheduleConflict`。
- 扩展 `apps/api/internal/events/service.go`。
  - 新增 `RSVP` 事务流程：加载活动、拒绝 cancelled/expired、容量检查、participant upsert、event schedule upsert/delete、返回冲突。
  - 用户 RSVP 只接受 `going` / `maybe` / `not_going`，拒绝外部提交 `invited`。
  - `going` 容量检查排除当前用户，保证重复 RSVP 幂等。
  - 过滤同一活动生成的 event schedule，避免把活动自己的日程作为冲突返回。
  - 新增 `Cancel` service 入口，并在 cancel 事务内删除该 event 的自动日程。
- 扩展 `apps/api/internal/events/repository.go`。
  - 新增事务包装、锁定参与记录后的 going 计数、participant upsert。
  - 通过 schedules linker 实现 event schedule upsert/delete。
  - 新增冲突查询适配和 cancel 状态更新。
- 新增 `apps/api/internal/schedules/model.go`、`event_linker.go`、`conflicts.go`。
  - 定义 `Schedule`、`EventScheduleInput`、`Conflict`。
  - `UpsertEventSchedule` 写入 `source='event'`，保留已有 visibility，只更新标题、时间、地点。
  - `DeleteEventSchedule` 按现有软删除模式设置 `deleted_at`。
  - `FindConflicts` 使用半开区间重叠条件查询当前用户日程。
- 扩展 `apps/api/internal/events/handler.go` 和 `apps/api/internal/http/router.go`。
  - 注册 `POST /api/events/:id/rsvp`。
  - 注册 `DELETE /api/events/:id/rsvp`。
  - 注册 `POST /api/events/:id/cancel`。
  - 容量满员映射为 `409 capacity_full`。
- 扩展测试。
  - RSVP service 覆盖 going/maybe 自动日程、not_going 删除、`add_to_calendar=false` 删除、容量满、幂等更新、invited 转 going。
  - RSVP service 覆盖用户提交 `invited` 被拒绝。
  - Cancel service 覆盖取消活动后删除 event schedules。
  - HTTP 路由测试覆盖容量满员映射为 409。

## Verification

TDD red check:

```bash
go test ./apps/api/internal/events -run TestRSVP -v
```

Result:

- Exit code: `1`
- 预期失败，原因是 `Service.RSVP`、`RSVPInput`、RSVP 常量和 `ScheduleConflict` 尚未实现。

RSVP service:

```bash
go test ./apps/api/internal/events -run TestRSVP -v
```

Result:

- Exit code: `0`
- `TestRSVPGoingInsertsParticipantAndEventSchedule`: PASS
- `TestRSVPMaybeInsertsParticipantAndEventSchedule`: PASS
- `TestRSVPNotGoingDeletesExistingEventSchedule`: PASS
- `TestRSVPAddToCalendarFalseStoresParticipantAndDeletesSchedule`: PASS
- `TestRSVPCapacityFullReturnsDomainError`: PASS
- `TestRSVPUpdatingExistingParticipantIsIdempotent`: PASS
- `TestRSVPInvitedParticipantChangingToGoingKeepsOneRow`: PASS
- `TestRSVPRejectsUserSubmittedInvited`: PASS
- `TestCancelDeletesEventSchedules`: PASS

Endpoint regression:

```bash
go test ./apps/api/internal/http -run 'TestRSVP|TestEventDetail' -v
```

Result:

- Exit code: `0`
- `TestEventDetailRouteIsPublic`: PASS
- `TestRSVPCapacityFullMapsToConflict`: PASS

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

- `schedules` 当前只实现 Phase 3 需要的 event-source linker 和冲突查询；手动日程 CRUD 仍留给 Phase 4。
- `Cancel` 只允许活动 owner 更新状态；非 owner 或不存在的 event 返回 not found。
