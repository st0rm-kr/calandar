# Hangout Phase 4 Schedules Calendar Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现手动个人日程 CRUD、冲突检测、月视图聚合 API 和基础日历页面。

**Architecture:** `schedules` 拥有手动日程和 event 日程的统一查询能力；手动 CRUD 只允许修改 `source='manual'`。`calendar` 聚合我的 schedules 与我所属群活动的只读视图，P1 好友筛选先只保留接口参数兼容。

**Tech Stack:** Gin, GORM, PostgreSQL range overlap query, React, date-fns。

## Global Constraints

- 日程可见性为 `public` / `busy_only` / `private`。
- 活动自动日程默认 `busy_only`，用户可手动改。
- 日程冲突只软提示，不阻止创建、编辑或报名。
- 冲突判定为半开区间：`existing.start_at < new_end AND new_start < existing.end_at`。
- `exclude_id` 用于编辑时排除自身。

---

## File Structure

- Create: `apps/api/internal/schedules/repository.go`
- Create: `apps/api/internal/schedules/service.go`
- Create: `apps/api/internal/schedules/handler.go`
- Create: `apps/api/internal/schedules/service_test.go`
- Create: `apps/api/internal/calendar/service.go`
- Create: `apps/api/internal/calendar/handler.go`
- Create: `apps/api/internal/calendar/service_test.go`
- Modify: `apps/api/internal/http/router.go`
- Create: `apps/web/src/lib/schedules.ts`
- Create: `apps/web/src/lib/calendar.ts`
- Create: `apps/web/src/routes/CalendarPage.tsx`
- Create: `apps/web/src/routes/ScheduleEditorPage.tsx`

## Backend Interfaces

- `GET /api/schedules?from=&to=`
- `POST /api/schedules`
- `PATCH /api/schedules/:id`
- `DELETE /api/schedules/:id`
- `GET /api/schedules/conflicts?start=&end=&exclude_id=`
- `GET /api/calendar?from=&to=&filter=`
- `schedules.Service.CreateManual(ctx, userID, input) (ScheduleResult, error)`
- `schedules.Service.UpdateManual(ctx, userID, scheduleID, input) (ScheduleResult, error)`
- `schedules.Service.FindConflicts(ctx, userID, start, end, excludeID) ([]Conflict, error)`

### Task 1: 手动日程 CRUD

- [ ] **Step 1: 写 service 测试**

Create tests covering:

- create manual schedule accepts no overlap and returns empty conflicts
- create manual schedule accepts overlap and returns conflict list
- update manual schedule cannot update another user's schedule
- delete rejects `source='event'`
- list range returns schedules where `schedule.start_at < to AND from < coalesce(schedule.end_at, schedule.start_at + interval '1 minute')`

Run:

```bash
go test ./apps/api/internal/schedules -run TestManualSchedules -v
```

Expected: FAIL.

- [ ] **Step 2: 实现 repository 和 service**

Create repository functions:

- `ListByRange(ctx, userID, from, to time.Time) ([]Schedule, error)`
- `CreateManual(ctx, Schedule) (Schedule, error)`
- `UpdateManual(ctx, userID uuid.UUID, id uint, patch SchedulePatch) (Schedule, error)`
- `DeleteManual(ctx, userID uuid.UUID, id uint) error`
- `FindConflicts(ctx, userID uuid.UUID, start time.Time, end *time.Time, excludeID *uint) ([]Schedule, error)`

Service validation:

- title 1-80 chars
- end must be after start when present
- visibility must be one of exact enum values
- location optional

- [ ] **Step 3: 实现 handler 并注册路由**

Map domain errors:

- invalid input -> 400
- schedule not found -> 404
- event schedule delete attempt -> 409

Run:

```bash
go test ./apps/api/internal/schedules -v
go test ./apps/api/...
```

Expected: PASS.

### Task 2: 冲突检测 API

- [ ] **Step 1: 写 handler 测试**

Cover:

- `GET /api/schedules/conflicts?start=...&end=...` returns overlaps
- adjacent schedules are not conflicts
- `exclude_id` removes the edited schedule from result

- [ ] **Step 2: 实现 query parser**

Parse RFC3339 timestamps and convert to UTC before service call.

Run:

```bash
go test ./apps/api/internal/schedules -run TestConflictsHandler -v
```

Expected: PASS.

### Task 3: 日历聚合 API

- [ ] **Step 1: 写 calendar service 测试**

Cover:

- calendar includes user's manual schedules
- calendar includes user's event schedules
- calendar marks `busy_only` items as busy without exposing private detail fields to other users
- `filter=all` and `filter=groups` accepted

Run:

```bash
go test ./apps/api/internal/calendar -v
```

Expected: FAIL.

- [ ] **Step 2: 实现 calendar service**

Create `CalendarItem` fields:

- `id`
- `kind`: `schedule` or `event`
- `title`
- `start_at`
- `end_at`
- `location`
- `visibility`
- `color`: `blue`, `green`, or `gray`
- `has_conflict`

Conflict marker is computed by grouping user's items and checking pairwise overlap in the requested range.

- [ ] **Step 3: 注册 `GET /api/calendar`**

Run:

```bash
go test ./apps/api/internal/calendar -v
go test ./apps/api/...
```

Expected: PASS.

### Task 4: 前端日历与日程编辑

- [ ] **Step 1: 安装日期库**

Run:

```bash
cd apps/web
npm install date-fns
```

- [ ] **Step 2: 创建 API 封装**

Create:

- `apps/web/src/lib/schedules.ts`
- `apps/web/src/lib/calendar.ts`

Expose typed functions for list/create/update/delete/conflicts/calendar.

- [ ] **Step 3: 创建日历页**

Create `CalendarPage.tsx` with:

- month grid
- blue/green/gray marker dots
- click date to open daily list drawer
- conflict badge when `has_conflict=true`
- floating `+` action linking to event or schedule creation

- [ ] **Step 4: 创建日程编辑页**

Create `ScheduleEditorPage.tsx` with:

- title, start, end, location, visibility
- live conflict check after both start and end are valid
- save still enabled when conflicts exist

- [ ] **Step 5: 验证**

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
go test ./apps/api/...
cd apps/web && npm run build
```

Expected: 全部 PASS。

- [ ] **Step 2: 提交**

Run:

```bash
git add apps/api apps/web
git commit -m "feat: add schedules and calendar"
```
