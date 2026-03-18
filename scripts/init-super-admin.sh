#!/usr/bin/env bash
set -euo pipefail

EMAIL="${1:-}"
PASSWORD="${2:-}"
BALANCE_CENTS="${3:-100000}"

if [ -z "$EMAIL" ] || [ -z "$PASSWORD" ]; then
  echo "用法: bash scripts/init-super-admin.sh <email> <password> [balance_cents]"
  exit 1
fi

HASH=$(docker compose exec -T api node -e "import bcrypt from 'bcryptjs'; console.log(await bcrypt.hash(process.argv[1], 10));" "$PASSWORD" | tail -n 1)

SQL=$(cat <<SQL
insert into users (id, email, password_hash, role, status, balance_cents)
values (gen_random_uuid(), '${EMAIL}', '${HASH}', 'super_admin', 'active', ${BALANCE_CENTS})
on conflict (email)
do update set role='super_admin', password_hash='${HASH}', balance_cents=${BALANCE_CENTS}, updated_at=now();
SQL
)

docker compose exec -T postgres psql -U "${POSTGRES_USER:-cliapi}" -d "${POSTGRES_DB:-cliapi}" -c "$SQL"
echo "super_admin 已初始化: ${EMAIL}"
