#!/usr/bin/env sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required"
  exit 1
fi

# Very simple migration runner for MVP: apply all .sql files in backend/migrations in lexical order.
# Requires psql available in the environment where you run it.

DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

for f in "$DIR"/migrations/*.sql; do
  echo "Applying $f"
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
done


