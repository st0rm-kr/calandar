create table public.events (
  id bigserial primary key,
  owner_id uuid not null references public.users(id),
  scope text not null check (scope in ('personal', 'group')),
  group_id bigint,
  title text not null check (char_length(title) between 1 and 80),
  type text not null check (type in ('climbing', 'dining', 'travel', 'gaming', 'other')),
  start_at timestamptz not null,
  end_at timestamptz,
  location text,
  capacity integer check (capacity is null or capacity > 0),
  status text not null default 'active' check (status in ('active', 'cancelled', 'expired')),
  share_slug text not null unique,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  check (end_at is null or end_at > start_at)
);

create table public.event_participants (
  id bigserial primary key,
  event_id bigint not null references public.events(id) on delete cascade,
  user_id uuid not null references public.users(id) on delete cascade,
  rsvp text not null check (rsvp in ('invited', 'going', 'not_going', 'maybe')),
  source text not null check (source in ('self', 'invited')),
  add_to_calendar boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  unique (event_id, user_id)
);

create table public.schedules (
  id bigserial primary key,
  user_id uuid not null references public.users(id) on delete cascade,
  title text not null,
  start_at timestamptz not null,
  end_at timestamptz,
  location text,
  visibility text not null check (visibility in ('public', 'busy_only', 'private')),
  source text not null check (source in ('manual', 'event')),
  event_id bigint references public.events(id) on delete cascade,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  check (end_at is null or end_at > start_at)
);

create unique index schedules_user_event_unique
on public.schedules (user_id, event_id)
where source = 'event' and deleted_at is null;
