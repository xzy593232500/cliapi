create extension if not exists pgcrypto;

create table if not exists users (
  id uuid primary key,
  email text unique not null,
  password_hash text not null,
  role text not null default 'user',
  status text not null default 'active',
  balance_cents bigint not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists api_keys (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  name text,
  key_prefix text not null,
  key_hash text unique not null,
  status text not null default 'active',
  last_used_at timestamptz,
  created_at timestamptz not null default now()
);

create table if not exists token_usage_logs (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  api_key_id uuid references api_keys(id) on delete set null,
  model text,
  input_tokens integer not null default 0,
  output_tokens integer not null default 0,
  total_tokens integer not null default 0,
  cost_cents bigint not null default 0,
  request_id text,
  created_at timestamptz not null default now()
);

create table if not exists balance_ledger (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  type text not null,
  amount_cents bigint not null,
  balance_after_cents bigint not null,
  reference_type text,
  reference_id text,
  note text,
  created_at timestamptz not null default now()
);

create table if not exists redemption_codes (
  id uuid primary key,
  code text unique not null,
  amount_cents bigint not null,
  status text not null default 'unused',
  created_by uuid references users(id) on delete set null,
  used_by uuid references users(id) on delete set null,
  used_at timestamptz,
  expires_at timestamptz,
  max_uses integer not null default 1,
  used_count integer not null default 0,
  created_at timestamptz not null default now()
);
