# Hangout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Hangout 从空仓库实现为可内测的移动端优先社交日历与活动协调 Web 应用。

**Architecture:** 采用前后端同仓库：`apps/api` 为 Go + Gin + GORM 后端，`apps/web` 为 React + Vite 前端，`supabase` 存放本地 Supabase、迁移和 Auth profile trigger。前端只用 `supabase-js` 做 Auth/session，所有业务数据读写统一经 Gin `/api/*`；生产部署为香港 ECS 单机托管前端静态文件并反代 Gin。

**Tech Stack:** Go 1.23+, Gin, GORM, PostgreSQL/Supabase local, React 18, TypeScript, Vite, Tailwind CSS, shadcn/ui, supabase-js, Vitest, Testing Library, Go `testing` + `httptest`。

## Global Constraints

- Web 应用移动端优先，UI 风格对标 Luma。
- P0 核心闭环为：建活动 -> 分享链接 -> 群成员点进来报名 -> 自动生成个人日程。
- 认证由 Supabase Auth 托管，P0 只做邮箱+密码注册/登录/重置密码。
- Gin 后端只验签 Supabase JWT，不自行签发业务 token。
- 前端不得使用 `supabase.from(...)` 或 `supabase.rpc(...)` 直接读写业务表/业务逻辑。
- 业务数据统一通过 Gin 暴露的 `/api/*` 访问。
- 用户 ID 使用 uuid，与 Supabase `auth.users.id` 对齐。
- 开发、测试、生产统一使用 PostgreSQL 语义；不引入 SQLite。
- 时间统一 UTC 存储，前端按 Asia/Shanghai 展示。
- RSVP 状态为 `invited` / `going` / `not_going` / `maybe`。
- 日程可见性为 `public` / `busy_only` / `private`。
- 日程冲突只软提示，不阻止操作；判定为 `existing.start < new.end && new.start < existing.end`。
- 微信 OAuth、微信 JS-SDK 自定义分享卡片、模板消息不纳入当前阶段。
- 部署基线为香港阿里云 ECS 单机，Cloudflare 默认 DNS only。

---

## 阶段文件

1. [Phase 1 - 项目骨架与数据库基线](file:///Users/bytedance/projects/calandar/docs/superpowers/plans/2026-06-23-hangout/01-foundation.md)
2. [Phase 2 - Supabase Auth 与用户档案](file:///Users/bytedance/projects/calandar/docs/superpowers/plans/2026-06-23-hangout/02-auth-users.md)
3. [Phase 3 - 活动、分享页、RSVP 与自动日程](file:///Users/bytedance/projects/calandar/docs/superpowers/plans/2026-06-23-hangout/03-events-rsvp.md)
4. [Phase 4 - 个人日程与日历聚合](file:///Users/bytedance/projects/calandar/docs/superpowers/plans/2026-06-23-hangout/04-schedules-calendar.md)
5. [Phase 5 - 好友与群组](file:///Users/bytedance/projects/calandar/docs/superpowers/plans/2026-06-23-hangout/05-friends-groups.md)
6. [Phase 6 - 收件箱与通知](file:///Users/bytedance/projects/calandar/docs/superpowers/plans/2026-06-23-hangout/06-inbox-notifications.md)
7. [Phase 7 - 前端页面闭环与视觉打磨](file:///Users/bytedance/projects/calandar/docs/superpowers/plans/2026-06-23-hangout/07-frontend-experience.md)
8. [Phase 8 - 部署、内测与上线切换](file:///Users/bytedance/projects/calandar/docs/superpowers/plans/2026-06-23-hangout/08-deployment.md)

## 交付节奏

- 每个阶段结束都应有一个可运行、可测试、可 review 的增量。
- 每个阶段单独提交，提交信息采用 `feat: phase N ...` 或 `chore: phase N ...`。
- Phase 1 到 Phase 4 完成后，核心闭环应能通过本地环境跑通。
- Phase 5 到 Phase 6 完成后，P0 社交协同能力应能跑通。
- Phase 7 到 Phase 8 完成后，可以进入小范围 IP 内测和正式域名切换。

## 跨阶段接口约定

- 后端统一响应：成功 `{ "data": ..., "error": null }`，失败 `{ "data": null, "error": { "code": "...", "message": "..." } }`。
- 后端鉴权 Header：`Authorization: Bearer <supabase_access_token>`。
- 前端 API 客户端统一从 Supabase session 读取 access token，并附加到 Gin 请求。
- 数据库迁移统一放在 `supabase/migrations/*.sql`，GORM model 只描述应用层结构，不依赖 AutoMigrate 修改生产表。
- 后端测试使用 PostgreSQL test database 或 Supabase local，不使用 SQLite。

## 自检清单

- [ ] 已覆盖 spec 中 P0 认证、活动、RSVP、个人日程、日历视图、好友、群组、收件箱、通知。
- [ ] 已覆盖匿名只读活动详情与登录后 RSVP。
- [ ] 已覆盖容量已满阻止 going。
- [ ] 已覆盖 RSVP 与 event schedule 的 upsert/delete 联动。
- [ ] 已覆盖冲突检测软提示。
- [ ] 已覆盖 OG/Twitter meta 的 `/e/:slug`。
- [ ] 已覆盖香港 ECS 单机部署路径。
