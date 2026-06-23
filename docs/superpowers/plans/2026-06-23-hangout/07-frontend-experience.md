# Hangout Phase 7 Frontend Experience Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将前端页面串成完整移动端 P0 闭环，并完成 Luma 风格视觉、空状态、加载态和错误态打磨。

**Architecture:** 前端采用 route-level pages + shared components。`supabase-js` 只保留 Auth/session 入口，业务请求全部走 `src/lib/api.ts`；底部 Tab 管理主导航，活动分享详情页支持匿名只读和登录后回跳。

**Tech Stack:** React 18, TypeScript, Vite, Tailwind CSS, shadcn/ui, React Router, Testing Library, Vitest。

## Global Constraints

- 移动端优先，桌面端自适应增强。
- Luma 风格：大圆角卡片、克制留白、活动封面/色块、清晰层级、轻动效。
- 未登录访问活动详情为只读；点击 RSVP 引导登录，登录后回到原活动。
- 底部 Tab 包含日历、活动、收件箱、好友/群组、个人中心。
- 前端不得直接查询业务表或调用 Supabase RPC。

---

## File Structure

- Create: `apps/web/src/components/AppShell.tsx`
- Create: `apps/web/src/components/BottomTabs.tsx`
- Create: `apps/web/src/components/Card.tsx`
- Create: `apps/web/src/components/EmptyState.tsx`
- Create: `apps/web/src/components/LoadingState.tsx`
- Create: `apps/web/src/components/ErrorState.tsx`
- Create: `apps/web/src/components/EventTypeBadge.tsx`
- Create: `apps/web/src/components/ConflictBanner.tsx`
- Create: `apps/web/src/hooks/useSession.ts`
- Create: `apps/web/src/hooks/useAsync.ts`
- Modify: `apps/web/src/App.tsx`
- Modify: `apps/web/src/index.css`
- Create: `apps/web/src/routes/EventsPage.tsx`
- Create: `apps/web/src/routes/MePage.tsx`
- Create: `apps/web/src/routes/NotFoundPage.tsx`
- Create: `apps/web/src/routes/__tests__/auth-guard.test.tsx`
- Create: `apps/web/src/routes/__tests__/event-detail.test.tsx`

## Frontend Interfaces

- `useSession(): { session, profile, loading, signOut }`
- `AppShell` provides bottom tabs and safe-area padding.
- `RequireAuth` redirects unauthenticated users to `/login?redirect=<current>`.
- `LoginPage` consumes `redirect` query and navigates back after successful login.

### Task 1: 应用壳和路由守卫

- [ ] **Step 1: 写路由测试**

Create tests covering:

- unauthenticated `/` redirects to `/login?redirect=/`
- unauthenticated `/e/:slug` does not redirect
- successful login navigates to redirect target

Run:

```bash
cd apps/web
npm run test -- auth-guard.test.tsx
```

Expected: FAIL until route guard exists.

- [ ] **Step 2: 实现 `useSession`**

Use `supabase.auth.getSession()` and `supabase.auth.onAuthStateChange()`; after session exists, load `/api/users/me` through `apiGet`.

- [ ] **Step 3: 实现 AppShell 和 BottomTabs**

Bottom tabs:

- `/` 日历
- `/events` 活动
- `/inbox` 收件箱
- `/friends` 好友
- `/me` 我的

Inbox tab displays unread count from `/api/inbox`.

- [ ] **Step 4: 重构 `App.tsx` 路由**

Routes:

- `/login`
- `/`
- `/events`
- `/events/new`
- `/e/:slug`
- `/schedules/new`
- `/schedules/:id`
- `/inbox`
- `/friends`
- `/groups`
- `/groups/:id`
- `/notifications`
- `/me`
- `*`

Run:

```bash
cd apps/web
npm run build
```

Expected: PASS.

### Task 2: 共享视觉组件

- [ ] **Step 1: 添加 shadcn/ui 基础组件**

Run:

```bash
cd apps/web
npx shadcn@latest init
npx shadcn@latest add button input textarea select checkbox dialog drawer toast
```

Expected: `src/components/ui/*` 生成成功。

- [ ] **Step 2: 创建状态组件**

Create `Card`, `EmptyState`, `LoadingState`, `ErrorState`, `EventTypeBadge`, `ConflictBanner` with Tailwind classes:

- card: `rounded-3xl border border-black/5 bg-white p-5 shadow-sm`
- dark hero card: `rounded-[2rem] bg-neutral-950 p-6 text-white`
- primary button: large radius, high contrast

- [ ] **Step 3: 应用到主页面**

Update Calendar, EventDetail, Inbox, Friends, Groups, Me pages to use shared components and consistent spacing:

- page padding `px-4 pb-24 pt-4`
- max width `mx-auto max-w-md`
- title size `text-2xl font-semibold tracking-tight`

Run:

```bash
cd apps/web
npm run build
```

Expected: PASS.

### Task 3: 活动详情匿名与登录回跳

- [ ] **Step 1: 写活动详情测试**

Cover:

- anonymous user sees title/time/location and disabled RSVP prompt
- clicking RSVP when anonymous navigates to `/login?redirect=/e/<slug>`
- logged-in user can submit going/maybe/not_going
- capacity full disables going button
- conflict response displays `ConflictBanner`

Run:

```bash
cd apps/web
npm run test -- event-detail.test.tsx
```

Expected: FAIL.

- [ ] **Step 2: 实现匿名状态交互**

Modify `EventDetailPage.tsx`:

- always fetch public detail without token requirement
- if no session, render RSVP CTA that links to login redirect
- if session, render RSVP controls
- after RSVP, refresh detail and calendar state

Run:

```bash
cd apps/web
npm run test -- event-detail.test.tsx
npm run build
```

Expected: PASS.

### Task 4: 页面完成度检查

- [ ] **Step 1: 空状态和错误态覆盖**

Ensure these pages have non-empty empty/error/loading states:

- Calendar
- Events
- Inbox
- Friends
- Groups
- Notifications
- Me

- [ ] **Step 2: 手动冒烟脚本**

Run app locally and verify:

```bash
cd apps/web
npm run dev
```

Manual path:

- register/login
- create event
- open `/e/:slug` in private browser
- login from RSVP prompt
- RSVP going with add-to-calendar checked
- see event in calendar
- create friend request and process from inbox
- create group and group event

- [ ] **Step 3: 构建验证**

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
cd apps/web && npm run test -- --run && npm run build
```

Expected: 全部 PASS。

- [ ] **Step 2: 提交**

Run:

```bash
git add apps/web
git commit -m "feat: polish hangout frontend experience"
```
