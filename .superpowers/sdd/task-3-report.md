# Task 3 Report: Supabase local 与初始迁移

## 状态

已完成。

## 实现内容

- 复用已有 Supabase local 初始化结果：`supabase/config.toml`
- 配置 Supabase Auth 本地回跳地址为 Vite dev server：`http://127.0.0.1:5173`
- 添加初始迁移：`supabase/migrations/202606230001_initial_schema.sql`
- 添加环境变量样例：`.env.example`

## 迁移内容

- 启用 PostgreSQL 扩展：`pgcrypto`
- 创建迁移标记表：`public.schema_migrations_marker`
- 写入初始迁移标记：`202606230001_initial_schema`

## 验证

```bash
rg -n "127\\.0\\.0\\.1:3000|localhost:3000|127\\.0\\.0\\.1:5173|localhost:5173|site_url|additional_redirect_urls" supabase .env.example apps/web/package.json
supabase stop && supabase start
supabase db reset
docker exec supabase_db_calandar psql -U postgres -d postgres -Atc "select exists (select 1 from pg_extension where extname = 'pgcrypto') as has_pgcrypto, exists (select 1 from information_schema.tables where table_schema='public' and table_name='schema_migrations_marker') as has_marker; select name from public.schema_migrations_marker order by name;"
go test ./apps/api/...
cd apps/web && npm run build
```

结果：全部通过。

- `rg`：相关文件中 `site_url`、`additional_redirect_urls` 和 WebAuthn 注释示例均指向 `http://127.0.0.1:5173`，未命中旧回跳端口
- `supabase stop && supabase start`：本地 Supabase 使用更新后的配置重启成功
- `supabase db reset`：成功应用 `202606230001_initial_schema.sql`
- 容器内 `psql` 查询：返回 `t|t` 和 `202606230001_initial_schema`
- `go test ./apps/api/...`：API 所有 package 测试通过
- `npm run build`：TypeScript 与 Vite 生产构建通过

## 备注

- 本机未安装宿主机 `psql`，已使用 `supabase_db_calandar` 容器内置 `psql` 完成数据库结果校验。
- `supabase db reset` 报告 `WARN: no files matched pattern: supabase/seed.sql`，这是当前 Supabase 默认 seed 配置下的 warning，不影响迁移执行结果。
