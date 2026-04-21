#!/usr/bin/env bash
set -euo pipefail

TIMEOUT_SEC="${WAIT_TIMEOUT_SEC:-180}"
STACK_BASE_URL="${STACK_BASE_URL:-http://localhost:8088}"
START_TS=$(date +%s)

check() {
  local url="$1"
  curl -fsS "$url" >/dev/null
}

while true; do
  if check "$STACK_BASE_URL/healthz" && check "$STACK_BASE_URL/readyz"; then
    echo "stack ready: $STACK_BASE_URL"
    exit 0
  fi
  now=$(date +%s)
  if (( now - START_TS > TIMEOUT_SEC )); then
    echo "timeout waiting stack: $STACK_BASE_URL" >&2
    exit 1
  fi
  sleep 2
done
