# Phase 3 Task 1 Report

## Scope

执行 `.superpowers/sdd/phase3-task1-brief.md`：新增活动、报名和日程相关数据库表与约束。

## Changes

- 新增 `supabase/migrations/202606230003_events_rsvp.sql`。
  - 创建 `public.events`。
  - 创建 `public.event_participants`。
  - 创建 `public.schedules`。
  - 创建 `schedules_user_event_unique` 条件唯一索引，约束 event 来源日程在未软删除状态下按 `(user_id, event_id)` 唯一。
- 表结构包含 brief 指定的主键、外键、枚举 check、时间范围 check、容量 check、唯一约束和软删除字段。

## Verification

Migration reset:

```bash
supabase db reset
```

Result:

- Exit code: `0`
- 已应用迁移：
  - `202606230001_initial_schema.sql`
  - `202606230002_auth_users.sql`
  - `202606230003_events_rsvp.sql`
- 提示 `WARN: no files matched pattern: supabase/seed.sql`，不影响本次迁移验证。

Database object check:

```bash
docker exec supabase_db_calandar psql -U postgres -d postgres -Atc "select table_name from information_schema.tables where table_schema = 'public' and table_name in ('events','event_participants','schedules') order by table_name; select indexname from pg_indexes where schemaname = 'public' and indexname = 'schedules_user_event_unique';"
```

Result:

- Exit code: `0`
- Found:
  - `event_participants`
  - `events`
  - `schedules`
  - `schedules_user_event_unique`

## Notes

- 宿主机未安装 `psql`，因此目录查询使用 Supabase Postgres 容器内的 `psql` 执行。
