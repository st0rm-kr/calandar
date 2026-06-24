# Phase 3 Task 5 Report

## Scope

执行 `.superpowers/sdd/phase3-task5-brief.md`：实现前端活动创建、活动详情分享页、RSVP 入口和 events API 封装。

## Changes

- 新增 `apps/web/src/lib/events.ts`。
  - 提供 `createEvent`、`getEventBySlug`、`rsvpEvent`、`cancelEvent`、`listMine`。
  - 对齐后端 envelope 响应、Bearer token 鉴权和 events/RSVP/cancel 路由。
- 新增 `apps/web/src/NewEventPage.tsx`。
  - 支持 scope、title、type、start/end、location、capacity 表单。
  - 创建成功后跳转到活动分享详情页。
- 新增 `apps/web/src/EventDetailPage.tsx`。
  - 支持匿名读取活动详情。
  - 展示时间、地点、状态、名额和分享 slug。
  - 提供 RSVP 按钮、加入日历 checkbox、可见性 selector。
  - 名额满时禁用 going RSVP；RSVP 成功后展示冲突软提示。
- 扩展 `apps/web/src/App.tsx`。
  - 注册 `/events/new` 和 `/events/:slug` 路由。
  - 首页增加创建活动入口。
- 扩展前端测试。
  - `apps/web/src/lib/events.test.ts` 覆盖 events API 封装。
  - `apps/web/src/App.test.tsx` 覆盖新建活动路由和详情页 slug 加载。

## Verification

TDD red checks:

```bash
cd apps/web && npm test -- src/lib/events.test.ts
```

Result:

- Exit code: `1`
- 预期失败，原因是 `src/lib/events.ts` 尚未存在。

```bash
cd apps/web && npm test -- src/App.test.tsx
```

Result:

- Exit code: `1`
- 预期失败，原因是 `/events/new` 和 `/events/:slug` 路由尚未注册。

Focused tests:

```bash
cd apps/web && npm test -- src/App.test.tsx src/lib/events.test.ts
```

Result:

- Exit code: `0`
- Test files: `2 passed`
- Tests: `10 passed`

Task build:

```bash
cd apps/web && npm run build
```

Result:

- Exit code: `0`
- `tsc -b && vite build`: PASS

Phase verification:

```bash
supabase db reset
```

Result:

- Exit code: `0`
- Local database reset and migrations applied.

```bash
go test ./apps/api/...
```

Result:

- Exit code: `0`
- API packages passed; no-test packages reported as expected.

```bash
cd apps/web && npm run build
```

Result:

- Exit code: `0`
- `tsc -b && vite build`: PASS

Additional web test suite:

```bash
cd apps/web && npm test
```

Result:

- Exit code: `0`
- Test files: `3 passed`
- Tests: `14 passed`

## Notes

- `cancelEvent` maps to the backend activity cancel endpoint `POST /api/events/:id/cancel`.
- RSVP removal is not exposed as a separate wrapper in this task because the brief requested only the five named events API functions.
