#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
mkdir -p "$ROOT_DIR/artifacts/seed"

POSTS="${SEED_POST_COUNT:-5000}"
USERS="${SEED_USER_COUNT:-200}"
OUT_FILE="/out/post_ids.txt"
SEED_MODE="${SEED_MODE:-compose}" # compose | remote
GATEWAY_BASE="${GATEWAY_BASE:-${TARGET_BASE_URL:-http://localhost:8088}}"

if [[ "$SEED_MODE" == "compose" ]]; then
  docker compose run --rm \
    -e SEED_POST_COUNT="$POSTS" \
    -e SEED_USER_COUNT="$USERS" \
    -e OUTPUT_FILE="$OUT_FILE" \
    -v "$ROOT_DIR/artifacts/seed:/out" \
    seeder
else
  docker build -t local/seeder-go:latest "$ROOT_DIR/seeder-go"
  docker run --rm \
    -e GATEWAY_BASE="$GATEWAY_BASE" \
    -e JWT_SECRET="${JWT_SECRET:-dev-secret}" \
    -e SEED_POST_COUNT="$POSTS" \
    -e SEED_USER_COUNT="$USERS" \
    -e OUTPUT_FILE="$OUT_FILE" \
    -v "$ROOT_DIR/artifacts/seed:/out" \
    local/seeder-go:latest
fi

echo "seed complete: $ROOT_DIR/artifacts/seed/post_ids.txt"
