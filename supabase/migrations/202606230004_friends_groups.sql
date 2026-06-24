create table public.friendships (
  id bigserial primary key,
  requester_id uuid not null references public.users(id) on delete cascade,
  addressee_id uuid not null references public.users(id) on delete cascade,
  status text not null check (status in ('pending', 'accepted', 'rejected')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  unique (requester_id, addressee_id),
  check (requester_id <> addressee_id)
);

create index friendships_addressee_pending_idx
on public.friendships (addressee_id)
where status = 'pending' and deleted_at is null;

alter table public.friendships enable row level security;

create table public.groups (
  id bigserial primary key,
  name text not null check (char_length(name) between 1 and 60),
  description text,
  owner_id uuid not null references public.users(id),
  invite_code text not null unique,
  status text not null default 'active' check (status in ('active', 'dissolved')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz
);

create table public.group_members (
  id bigserial primary key,
  group_id bigint not null references public.groups(id) on delete cascade,
  user_id uuid not null references public.users(id) on delete cascade,
  role text not null check (role in ('owner', 'member')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  unique (group_id, user_id)
);

create table public.group_invites (
  id bigserial primary key,
  group_id bigint not null references public.groups(id) on delete cascade,
  inviter_id uuid not null references public.users(id),
  invitee_id uuid not null references public.users(id),
  status text not null default 'pending' check (status in ('pending', 'accepted', 'rejected')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  unique (group_id, invitee_id)
);

alter table public.groups enable row level security;
alter table public.group_members enable row level security;
alter table public.group_invites enable row level security;
