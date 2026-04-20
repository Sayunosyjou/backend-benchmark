#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
mkdir -p "$ROOT_DIR/artifacts/seed"

POSTS="${SEED_POST_COUNT:-5000}"
USERS="${SEED_USER_COUNT:-200}"
OUT_FILE="/out/post_ids.txt"


docker compose run --rm \
  -e SEED_POST_COUNT="$POSTS" \
  -e SEED_USER_COUNT="$USERS" \
  -e OUTPUT_FILE="$OUT_FILE" \
  -v "$ROOT_DIR/artifacts/seed:/out" \
  seeder

echo "seed complete: $ROOT_DIR/artifacts/seed/post_ids.txt"
