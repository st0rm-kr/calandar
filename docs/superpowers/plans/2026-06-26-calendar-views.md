# 首页日历视图切换 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Do not use sub coding agents for this repository session. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 将首页 Calendar 右上角从来源筛选改成 `月视图 / 周视图 / 列表 / 空闲` 视图切换，并保持订阅源列表负责来源选择。

**Architecture:** 前端复用现有 `/api/calendar?...&sources=...` 数据和 `CalendarItem` source metadata，在 `CalendarPage.tsx` 中增加 `CalendarView` 状态、URL query 同步和四个视图渲染分支。后端不新增接口，测试集中在 `App.test.tsx` 行为覆盖。

**Tech Stack:** React 19, React Router DOM, TypeScript, Tailwind CSS, Vitest + Testing Library.

## Global Constraints

- 不使用子 coding agents。
- 不触碰无关 `.gitignore` / `apps/web/.gitignore` 改动。
- 默认视图为 `month`，非法 `view` 回退到 `month`。
- 颜色不能作为唯一状态表达；视图按钮必须有文本和 `aria-pressed`。
- 好友 `busy_only` 只展示忙碌，不泄露标题和地点。

---

### Task 1: 视图切换状态与红测试

**Files:**
- Modify: `apps/web/src/App.test.tsx`
- Modify: `apps/web/src/CalendarPage.tsx`

**Interfaces:**
- Produces: `CalendarView = 'month' | 'week' | 'list' | 'availability'`
- Produces: view buttons named `月视图`, `周视图`, `列表`, `空闲`

- [x] **Step 1: Write failing tests**
  - Add tests that assert the old `全部 / 我的群 / 好友` controls are gone.
  - Add tests that assert the new four view buttons render with selected state.
  - Add tests for direct URLs `/?view=week`, `/?view=list`, `/?view=availability`, and invalid `/?view=bad`.

- [x] **Step 2: Run red tests**
  - Run: `cd apps/web && npm test -- App.test.tsx -t "calendar view"`
  - Expected: fail because view buttons and URL handling do not exist.

- [x] **Step 3: Implement minimal state**
  - Replace `CalendarFilter` UI usage in `CalendarPage.tsx` with `CalendarView`.
  - Always call `getCalendar(..., 'all', sources)` because subscriptions now control sources.
  - Read and write `view` from URL search params.

- [x] **Step 4: Run green tests**
  - Run: `cd apps/web && npm test -- App.test.tsx -t "calendar view"`
  - Expected: pass.

### Task 2: 周视图

**Files:**
- Modify: `apps/web/src/App.test.tsx`
- Modify: `apps/web/src/CalendarPage.tsx`

**Interfaces:**
- Consumes: `CalendarView`
- Produces: week layout with label `周视图时间表`

- [x] **Step 1: Write failing test**
  - Assert clicking `周视图` shows `周视图时间表`, weekday columns, and hour labels.

- [x] **Step 2: Implement week rendering**
  - Render 7 days from current selected week.
  - Render hour rail from `07:00` to `24:00`.
  - Render calendar items inside each day column with source label and time.

- [x] **Step 3: Verify**
  - Run: `cd apps/web && npm test -- App.test.tsx -t "calendar view"`

### Task 3: 列表视图

**Files:**
- Modify: `apps/web/src/App.test.tsx`
- Modify: `apps/web/src/CalendarPage.tsx`

**Interfaces:**
- Consumes: `CalendarView`
- Produces: agenda layout with label `接下来 14 天`

- [x] **Step 1: Write failing test**
  - Assert clicking `列表` groups items by natural date and shows source labels.

- [x] **Step 2: Implement agenda rendering**
  - Sort items by `start_at`.
  - Group by day heading.
  - Show item time, title, source label, and RSVP/visibility context.

- [x] **Step 3: Verify**
  - Run: `cd apps/web && npm test -- App.test.tsx -t "calendar view"`

### Task 4: 空闲视图

**Files:**
- Modify: `apps/web/src/App.test.tsx`
- Modify: `apps/web/src/CalendarPage.tsx`

**Interfaces:**
- Consumes: `CalendarView`
- Produces: availability layout with label `可约时间`

- [x] **Step 1: Write failing test**
  - Assert clicking `空闲` shows busy blocks and at least one available window.

- [x] **Step 2: Implement availability rendering**
  - Convert visible items to busy blocks.
  - Use default available window `18:00-23:00`.
  - Subtract busy blocks from the default window per day.
  - Show busy block source labels and available windows.

- [x] **Step 3: Verify**
  - Run: `cd apps/web && npm test -- App.test.tsx -t "calendar view"`

### Task 5: 全量验证

**Files:**
- Test: `apps/web/src/App.test.tsx`

- [x] **Step 1: Run frontend tests**
  - Run: `cd apps/web && npm test`

- [x] **Step 2: Run lint**
  - Run: `cd apps/web && npm run lint`

- [x] **Step 3: Run build**
  - Run: `cd apps/web && npm run build`

- [x] **Step 4: Scoped diff check**
  - Run: `git diff --check -- apps/web/src/App.test.tsx apps/web/src/CalendarPage.tsx docs/superpowers/plans/2026-06-26-calendar-views.md`
