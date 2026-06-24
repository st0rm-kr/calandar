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

alter table public.notifications enable row level security;
