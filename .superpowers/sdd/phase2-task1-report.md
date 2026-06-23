# Phase 2 Task 1 Report

## Scope

Executed `.superpowers/sdd/phase2-task1-brief.md`: add the Supabase profile table and auth user creation trigger.

## Changes

- Added `supabase/migrations/202606230002_auth_users.sql`.
- Created `public.users` with auth user cascade delete, profile fields, status check, timestamps, and soft-delete marker.
- Enabled row level security on `public.users`.
- Added partial indexes for active `email` and `display_name` lookups.
- Added `public.handle_new_auth_user()` as a `security definer` trigger function.
- Added `on_auth_user_created` trigger on `auth.users` after insert.

## Verification

Ran:

```bash
supabase db reset
```

Result:

- Exit code: `0`
- Applied `202606230001_initial_schema.sql`
- Applied `202606230002_auth_users.sql`
- Finished reset on branch `feat/hangout-implementation`

Ran object existence check:

```sql
select
  to_regclass('public.users')::text as users_table,
  to_regclass('public.users_email_idx')::text as email_idx,
  to_regclass('public.users_display_name_idx')::text as display_name_idx,
  exists (
    select 1
    from pg_proc p
    join pg_namespace n on n.oid = p.pronamespace
    where n.nspname = 'public'
      and p.proname = 'handle_new_auth_user'
  ) as has_function,
  exists (
    select 1
    from pg_trigger t
    join pg_class c on c.oid = t.tgrelid
    join pg_namespace n on n.oid = c.relnamespace
    where n.nspname = 'auth'
      and c.relname = 'users'
      and t.tgname = 'on_auth_user_created'
      and not t.tgisinternal
  ) as has_trigger;
```

Result row:

```json
{
  "users_table": "users",
  "email_idx": "users_email_idx",
  "display_name_idx": "users_display_name_idx",
  "has_function": true,
  "has_trigger": true
}
```

Ran RLS and policy check:

```sql
select
  c.relrowsecurity as users_rls_enabled,
  count(p.polname)::int as users_policy_count
from pg_class c
join pg_namespace n on n.oid = c.relnamespace
left join pg_policy p on p.polrelid = c.oid
where n.nspname = 'public'
  and c.relname = 'users'
group by c.relrowsecurity;
```

Result row:

```json
{
  "users_rls_enabled": true,
  "users_policy_count": 0
}
```

## Notes

- Supabase reported `WARN: no files matched pattern: supabase/seed.sql`; this did not fail the reset.
- No business policy was added for `public.users`, so client access remains denied by default until explicit policies are designed.
- Supabase advisory still reports RLS disabled for `public.schema_migrations_marker`; this task only changed `public.users`.
