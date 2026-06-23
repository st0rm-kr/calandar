# Hangout Phase 6 Inbox Notifications Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现统一收件箱、站内通知列表、未读计数，以及邀请/申请产生通知的领域事件写入。

**Architecture:** 收件箱是动态聚合视图，不新增写入表；通知是独立提醒流，写入 `notifications` 表。各领域 service 在成功创建好友申请、活动邀请、群邀请、活动取消时调用 notification service。

**Tech Stack:** Gin, GORM, PostgreSQL jsonb, React。

## Global Constraints

- 收件箱聚合 `event_participants(rsvp=invited)`、`friendships(status=pending, addressee=本人)`、`group_invites(status=pending, invitee=本人)`。
- 收件箱不新增写操作，批阅复用各域已有 endpoint。
- notifications 是提醒流，不作为收件箱待办的权威来源。
- 活动邀请行内批阅需要返回冲突提示。

---

## File Structure

- Create: `supabase/migrations/202606230005_notifications.sql`
- Create: `apps/api/internal/notifications/model.go`
- Create: `apps/api/internal/notifications/repository.go`
- Create: `apps/api/internal/notifications/service.go`
- Create: `apps/api/internal/notifications/handler.go`
- Create: `apps/api/internal/notifications/service_test.go`
- Create: `apps/api/internal/inbox/service.go`
- Create: `apps/api/internal/inbox/handler.go`
- Create: `apps/api/internal/inbox/service_test.go`
- Modify: `apps/api/internal/events/service.go`
- Modify: `apps/api/internal/friends/service.go`
- Modify: `apps/api/internal/groups/service.go`
- Modify: `apps/api/internal/http/router.go`
- Create: `apps/web/src/lib/inbox.ts`
- Create: `apps/web/src/lib/notifications.ts`
- Create: `apps/web/src/routes/InboxPage.tsx`
- Create: `apps/web/src/routes/NotificationsPage.tsx`

## Backend Interfaces

- `GET /api/inbox`
- `GET /api/notifications`
- `POST /api/notifications/:id/read`
- `POST /api/notifications/read-all`
- `notifications.Service.Create(ctx, userID, typ, payload) error`
- `inbox.Service.List(ctx, userID) (InboxResult, error)`

### Task 1: notifications 表与服务

- [ ] **Step 1: 创建迁移**

Create `supabase/migrations/202606230005_notifications.sql`:

```sql
create table public.notifications (
  id bigserial primary key,
  user_id uuid not null references public.users(id) on delete cascade,
  type text not null check (type in ('friend_request', 'friend_accepted', 'event_invite', 'event_changed', 'event_cancelled', 'group_invite')),
  payload jsonb not null default '{}'::jsonb,
  read_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz
);

create index notifications_user_unread_idx
on public.notifications (user_id, created_at desc)
where read_at is null and deleted_at is null;
```

- [ ] **Step 2: 写 service 测试**

Cover:

- create notification stores json payload
- list returns unread count and latest-first items
- read marks only current user's notification
- read-all marks all current user's unread notifications

Run:

```bash
go test ./apps/api/internal/notifications -v
```

Expected: FAIL.

- [ ] **Step 3: 实现 notifications service/handler**

Register:

- `GET /api/notifications`
- `POST /api/notifications/:id/read`
- `POST /api/notifications/read-all`

Run:

```bash
go test ./apps/api/internal/notifications -v
go test ./apps/api/...
```

Expected: PASS.

### Task 2: 收件箱聚合

- [ ] **Step 1: 写 inbox service 测试**

Cover:

- invited event appears in inbox with event title/time/location and conflict count
- pending friend request appears only for addressee
- pending group invite appears only for invitee
- handled items are excluded
- cancelled or expired event invite appears disabled with reason

Run:

```bash
go test ./apps/api/internal/inbox -v
```

Expected: FAIL.

- [ ] **Step 2: 实现 inbox service**

Return shape:

```go
type InboxResult struct {
	Counts InboxCounts `json:"counts"`
	Items  []InboxItem `json:"items"`
}

type InboxItem struct {
	ID            string    `json:"id"`
	Kind          string    `json:"kind"`
	SourceID      uint      `json:"source_id"`
	Title         string    `json:"title"`
	Subtitle      string    `json:"subtitle"`
	CreatedAt     time.Time `json:"created_at"`
	Disabled      bool      `json:"disabled"`
	DisabledReason string   `json:"disabled_reason,omitempty"`
	ConflictCount int       `json:"conflict_count,omitempty"`
}
```

Kind values: `event_invite`, `friend_request`, `group_invite`.

- [ ] **Step 3: 注册 `GET /api/inbox`**

Run:

```bash
go test ./apps/api/internal/inbox -v
go test ./apps/api/...
```

Expected: PASS.

### Task 3: 领域事件写通知

- [ ] **Step 1: 写集成测试**

Cover:

- sending friend request creates `friend_request` notification
- accepting friend request creates `friend_accepted` notification
- inviting event creates `event_invite` notification
- cancelling event creates `event_cancelled` notification for participants
- inviting group creates `group_invite` notification

Run:

```bash
go test ./apps/api/internal/... -run TestNotificationsFromDomainActions -v
```

Expected: FAIL until services call notifications.

- [ ] **Step 2: 注入 notification service**

Modify events/friends/groups service constructors to accept `notifications.Service`. For tests that do not need notifications, pass a no-op implementation.

- [ ] **Step 3: 验证**

Run:

```bash
go test ./apps/api/...
```

Expected: PASS.

### Task 4: 前端收件箱和通知页

- [ ] **Step 1: 创建 API 封装**

Create `inbox.ts` and `notifications.ts` with typed methods.

- [ ] **Step 2: 创建收件箱页**

`InboxPage.tsx` displays:

- activity invite actions: going, maybe, not going, add-to-calendar checkbox, visibility select
- friend request actions: accept, reject
- group invite actions: accept, reject
- conflict rows in red
- disabled cancelled/expired activity invites as gray rows

- [ ] **Step 3: 创建通知页**

`NotificationsPage.tsx` displays notification stream, unread styling, read and read-all actions, click target routing based on payload entity ids.

- [ ] **Step 4: 验证**

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
git commit -m "feat: add inbox and notifications"
```
