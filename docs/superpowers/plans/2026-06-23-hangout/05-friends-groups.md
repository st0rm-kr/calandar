# Hangout Phase 5 Friends Groups Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现好友申请、好友列表、点对点活动邀请、群组创建、群邀请、入群、退群和群内活动。

**Architecture:** 好友和群组各自拥有独立 service；活动邀请复用 `event_participants(rsvp='invited')`。群活动创建时校验发起人是群成员，群邀请进入 `group_invites(status='pending')`，收件箱聚合放到 Phase 6。

**Tech Stack:** Gin, GORM, PostgreSQL unique constraints, React。

## Global Constraints

- 好友申请双向确认。
- P0 只支持邀请站内好友或群成员；外部用户仅通过分享链接进入。
- 群成员均可在群内发起活动。
- 群活动可见性为链接/群内可见，P0 以链接访问详情为主。
- 已 RSVP 的用户再被邀请不回退为 `invited`。

---

## File Structure

- Create: `supabase/migrations/202606230004_friends_groups.sql`
- Create: `apps/api/internal/friends/model.go`
- Create: `apps/api/internal/friends/repository.go`
- Create: `apps/api/internal/friends/service.go`
- Create: `apps/api/internal/friends/handler.go`
- Create: `apps/api/internal/friends/service_test.go`
- Create: `apps/api/internal/groups/model.go`
- Create: `apps/api/internal/groups/repository.go`
- Create: `apps/api/internal/groups/service.go`
- Create: `apps/api/internal/groups/handler.go`
- Create: `apps/api/internal/groups/service_test.go`
- Modify: `apps/api/internal/events/service.go`
- Modify: `apps/api/internal/events/handler.go`
- Modify: `apps/api/internal/http/router.go`
- Create: `apps/web/src/lib/friends.ts`
- Create: `apps/web/src/lib/groups.ts`
- Create: `apps/web/src/routes/FriendsPage.tsx`
- Create: `apps/web/src/routes/GroupsPage.tsx`
- Create: `apps/web/src/routes/GroupDetailPage.tsx`

## Backend Interfaces

- Friends: `GET /api/friends`, `GET /api/friends/requests`, `POST /api/friends/requests`, `POST /api/friends/requests/:id/accept`, `POST /api/friends/requests/:id/reject`, `DELETE /api/friends/:userId`
- Groups: `GET /api/groups`, `POST /api/groups`, `GET /api/groups/:id`, `PATCH /api/groups/:id`, `DELETE /api/groups/:id`, `POST /api/groups/:id/members`, `POST /api/groups/:id/invites`, `POST /api/groups/invites/:id/accept`, `POST /api/groups/invites/:id/reject`, `DELETE /api/groups/:id/members/:userId`, `GET /api/groups/:id/events`
- Event invites: `POST /api/events/:id/invite`

### Task 1: 好友数据模型与 API

- [ ] **Step 1: 创建迁移**

Create `supabase/migrations/202606230004_friends_groups.sql`:

```sql
create table public.friendships (
  id bigserial primary key,
  requester_id uuid not null references public.users(id) on delete cascade,
  addressee_id uuid not null references public.users(id) on delete cascade,
  status text not null check (status in ('pending', 'accepted', 'rejected')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  unique (requester_id, addressee_id),
  check (requester_id <> addressee_id)
);

create index friendships_addressee_pending_idx
on public.friendships (addressee_id)
where status = 'pending' and deleted_at is null;
```

- [ ] **Step 2: 写 service 测试**

Cover:

- user cannot request self
- duplicate request is idempotent
- reverse pending request can be accepted by creating one accepted relationship
- accepted friendship appears in both users' lists
- delete friend removes accepted relationship for both users

Run:

```bash
go test ./apps/api/internal/friends -v
```

Expected: FAIL.

- [ ] **Step 3: 实现 friends service/handler**

Implement exact status transitions:

- `pending -> accepted`
- `pending -> rejected`
- rejected can be requested again by same requester and becomes `pending`
- accepted cannot be requested again

Run:

```bash
go test ./apps/api/internal/friends -v
go test ./apps/api/...
```

Expected: PASS.

### Task 2: 群组数据模型与 API

- [ ] **Step 1: 扩展迁移**

Append to `202606230004_friends_groups.sql`:

```sql
create table public.groups (
  id bigserial primary key,
  name text not null check (char_length(name) between 1 and 60),
  description text,
  owner_id uuid not null references public.users(id),
  invite_code text not null unique,
  status text not null default 'active' check (status in ('active', 'dissolved')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz
);

create table public.group_members (
  id bigserial primary key,
  group_id bigint not null references public.groups(id) on delete cascade,
  user_id uuid not null references public.users(id) on delete cascade,
  role text not null check (role in ('owner', 'member')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  unique (group_id, user_id)
);

create table public.group_invites (
  id bigserial primary key,
  group_id bigint not null references public.groups(id) on delete cascade,
  inviter_id uuid not null references public.users(id),
  invitee_id uuid not null references public.users(id),
  status text not null default 'pending' check (status in ('pending', 'accepted', 'rejected')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  unique (group_id, invitee_id)
);
```

- [ ] **Step 2: 写 service 测试**

Cover:

- creating group inserts owner member row
- invite code join inserts member once
- only group owner can patch/dissolve group
- invite creates pending group invite
- accepting group invite inserts member and marks invite accepted
- owner leaving last group dissolves group
- owner leaving non-empty group returns conflict requiring transfer

Run:

```bash
go test ./apps/api/internal/groups -v
```

Expected: FAIL.

- [ ] **Step 3: 实现 groups service/handler**

Implement invite code generation with crypto random 12 URL-safe chars.

Run:

```bash
go test ./apps/api/internal/groups -v
go test ./apps/api/...
```

Expected: PASS.

### Task 3: 点对点与群活动邀请

- [ ] **Step 1: 写 events invite 测试**

Cover:

- event owner can invite accepted friends
- group event creator can invite group members
- invite inserts participant `rsvp='invited'`, `source='invited'`
- existing `going/maybe/not_going` participant is unchanged
- non-friend invite for personal event returns 403

Run:

```bash
go test ./apps/api/internal/events -run TestInvite -v
```

Expected: FAIL.

- [ ] **Step 2: 实现 `POST /api/events/:id/invite`**

Add service method `Invite(ctx, actorID, eventID, inviteeIDs)` that calls friends/groups authorization helpers.

Run:

```bash
go test ./apps/api/internal/events -run TestInvite -v
go test ./apps/api/...
```

Expected: PASS.

### Task 4: 前端好友和群组页面

- [ ] **Step 1: 创建 API 封装**

Create `friends.ts` and `groups.ts` with typed methods for all endpoints listed above.

- [ ] **Step 2: 创建好友页**

`FriendsPage.tsx` includes search by nickname/email, send request button, friends list, delete friend action.

- [ ] **Step 3: 创建群组页和详情页**

`GroupsPage.tsx` includes my groups, create group form, invite code join form. `GroupDetailPage.tsx` includes members, group events, create group event link, invite user form, leave/dissolve actions.

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
git commit -m "feat: add friends and groups"
```
