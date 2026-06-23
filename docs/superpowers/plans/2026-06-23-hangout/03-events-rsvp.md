# Hangout Phase 3 Events RSVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现活动创建、匿名详情、标准 OG 分享页、RSVP 三态、容量限制和 RSVP 联动自动日程。

**Architecture:** `events` 负责活动与参与记录，`schedules` 先提供 event-source 自动日程写入能力，完整手动日程 CRUD 放到 Phase 4。RSVP service 在一个数据库事务内完成容量检查、participant upsert、event schedule upsert/delete 和冲突查询。

**Tech Stack:** Gin, GORM, PostgreSQL partial unique index, React Router, Testing Library。

## Global Constraints

- 未登录可只读查看活动详情，报名需登录。
- 分享 P0 提供标准 OG / Twitter meta，不做微信 JS-SDK 自定义卡片。
- RSVP 状态为 `invited` / `going` / `not_going` / `maybe`。
- 报名 `going`/`maybe` 时默认勾选加入日历，默认可见性为 `busy_only`。
- 容量已满阻止新的 `going`，返回 409。
- 取消报名 / 改 `not_going` / 取消勾选加入日历时删除对应 event schedule。

---

## File Structure

- Create: `supabase/migrations/202606230003_events_rsvp.sql`
- Create: `apps/api/internal/events/model.go`
- Create: `apps/api/internal/events/repository.go`
- Create: `apps/api/internal/events/service.go`
- Create: `apps/api/internal/events/handler.go`
- Create: `apps/api/internal/events/service_test.go`
- Create: `apps/api/internal/events/handler_test.go`
- Create: `apps/api/internal/schedules/model.go`
- Create: `apps/api/internal/schedules/event_linker.go`
- Create: `apps/api/internal/schedules/conflicts.go`
- Modify: `apps/api/internal/http/router.go`
- Create: `apps/web/src/routes/EventDetailPage.tsx`
- Create: `apps/web/src/routes/NewEventPage.tsx`
- Create: `apps/web/src/lib/events.ts`

## Backend Interfaces

- `POST /api/events`
- `GET /api/events/:slug`
- `POST /api/events/:id/rsvp`
- `DELETE /api/events/:id/rsvp`
- `POST /api/events/:id/cancel`
- `GET /api/events/mine`
- `GET /e/:slug`
- `events.Service.Create(ctx, userID, CreateEventInput) (EventDetail, error)`
- `events.Service.RSVP(ctx, userID, eventID, RSVPInput) (RSVPResult, error)`
- `schedules.UpsertEventSchedule(ctx, tx, userID, event, visibility) error`
- `schedules.DeleteEventSchedule(ctx, tx, userID, eventID) error`

### Task 1: 数据库表与约束

- [ ] **Step 1: 创建迁移**

Create `supabase/migrations/202606230003_events_rsvp.sql`:

```sql
create table public.events (
  id bigserial primary key,
  owner_id uuid not null references public.users(id),
  scope text not null check (scope in ('personal', 'group')),
  group_id bigint,
  title text not null check (char_length(title) between 1 and 80),
  type text not null check (type in ('climbing', 'dining', 'travel', 'gaming', 'other')),
  start_at timestamptz not null,
  end_at timestamptz,
  location text,
  capacity integer check (capacity is null or capacity > 0),
  status text not null default 'active' check (status in ('active', 'cancelled', 'expired')),
  share_slug text not null unique,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  check (end_at is null or end_at > start_at)
);

create table public.event_participants (
  id bigserial primary key,
  event_id bigint not null references public.events(id) on delete cascade,
  user_id uuid not null references public.users(id) on delete cascade,
  rsvp text not null check (rsvp in ('invited', 'going', 'not_going', 'maybe')),
  source text not null check (source in ('self', 'invited')),
  add_to_calendar boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  unique (event_id, user_id)
);

create table public.schedules (
  id bigserial primary key,
  user_id uuid not null references public.users(id) on delete cascade,
  title text not null,
  start_at timestamptz not null,
  end_at timestamptz,
  location text,
  visibility text not null check (visibility in ('public', 'busy_only', 'private')),
  source text not null check (source in ('manual', 'event')),
  event_id bigint references public.events(id) on delete cascade,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  check (end_at is null or end_at > start_at)
);

create unique index schedules_user_event_unique
on public.schedules (user_id, event_id)
where source = 'event' and deleted_at is null;
```

- [ ] **Step 2: 验证迁移**

Run:

```bash
supabase db reset
```

Expected: events, event_participants, schedules 三张表创建成功。

### Task 2: 活动创建与匿名详情 API

- [ ] **Step 1: 写 service 测试**

Create `apps/api/internal/events/service_test.go` covering:

- create event generates unique `share_slug`
- reject `end_at <= start_at`
- anonymous detail returns participant counts and does not require auth
- expired event is derived when `coalesce(end_at,start_at) < now`

Run:

```bash
go test ./apps/api/internal/events -run 'TestCreate|TestGetDetail' -v
```

Expected: FAIL until service exists.

- [ ] **Step 2: 实现 model/repository/service**

Create:

- `model.go`: `Event`, `Participant`, `EventDetail`, `CreateEventInput`
- `repository.go`: `Create`, `FindBySlug`, `FindByID`, `CountGoing`, `ListMine`
- `service.go`: validation, slug generation using 10 URL-safe chars from crypto random, derived expired status

- [ ] **Step 3: 实现 handler**

Create `handler.go` with request DTOs and register:

- authenticated `POST /api/events`
- public `GET /api/events/:slug`
- authenticated `GET /api/events/mine`

Run:

```bash
go test ./apps/api/internal/events -v
go test ./apps/api/...
```

Expected: PASS.

### Task 3: RSVP 与自动日程联动

- [ ] **Step 1: 写 RSVP service 测试**

Cover:

- going inserts participant and event schedule
- maybe inserts participant and event schedule
- not_going deletes existing event schedule
- `add_to_calendar=false` stores participant but deletes event schedule
- capacity full returns domain error mapped to 409
- updating existing participant is idempotent
- invited participant changing to going keeps one row

Run:

```bash
go test ./apps/api/internal/events -run TestRSVP -v
```

Expected: FAIL.

- [ ] **Step 2: 实现 schedules event linker**

Create `apps/api/internal/schedules/event_linker.go`:

- `UpsertEventSchedule` inserts `source='event'`, title/time/location from event, default visibility from input
- on conflict `(user_id,event_id)` updates title/start/end/location but keeps existing visibility
- `DeleteEventSchedule` soft deletes or deletes the event schedule row consistently with repository pattern

- [ ] **Step 3: 实现 RSVP 事务**

In `events.Service.RSVP`:

- load event by id
- reject cancelled/expired events
- if new RSVP is `going`, lock participant rows for event and count current going excluding current user
- if capacity reached, return `ErrCapacityFull`
- upsert participant with `source='self'`
- if rsvp in `going`,`maybe` and `add_to_calendar=true`, call `UpsertEventSchedule`
- otherwise call `DeleteEventSchedule`
- return conflicts from `schedules.FindConflicts`

- [ ] **Step 4: 注册 RSVP endpoints**

Register:

- `POST /api/events/:id/rsvp`
- `DELETE /api/events/:id/rsvp`
- `POST /api/events/:id/cancel`

Run:

```bash
go test ./apps/api/...
```

Expected: PASS.

### Task 4: OG 分享页

- [ ] **Step 1: 写 handler 测试**

Create test for `GET /e/:slug` that asserts HTML contains:

- `<meta property="og:title"`
- `<meta property="og:type" content="website"`
- `<meta name="twitter:card" content="summary_large_image"`
- event title, time, and location escaped safely

- [ ] **Step 2: 实现 HTML handler**

Register `GET /e/:slug` outside `/api`. Handler returns a minimal HTML shell with OG/Twitter meta and a root div for frontend hydration.

Run:

```bash
go test ./apps/api/internal/events -run TestOGPage -v
```

Expected: PASS.

### Task 5: 前端活动创建和详情

- [ ] **Step 1: 创建 API 封装**

Create `apps/web/src/lib/events.ts` with `createEvent`, `getEventBySlug`, `rsvpEvent`, `cancelEvent`, `listMine`.

- [ ] **Step 2: 创建页面**

Create:

- `NewEventPage.tsx`: scope, title, type, start/end, location, capacity form
- `EventDetailPage.tsx`: anonymous read-only detail, RSVP buttons, add-to-calendar checkbox, visibility selector, capacity disabled state, conflict soft warning

- [ ] **Step 3: 验证构建**

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
git commit -m "feat: add events rsvp and share pages"
```
