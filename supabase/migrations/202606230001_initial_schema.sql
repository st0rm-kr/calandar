create extension if not exists pgcrypto;

create table if not exists public.schema_migrations_marker (
  id bigserial primary key,
  name text not null unique,
  applied_at timestamptz not null default now()
);

insert into public.schema_migrations_marker (name)
values ('202606230001_initial_schema')
on conflict (name) do nothing;
