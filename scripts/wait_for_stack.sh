#!/usr/bin/env bash
set -euo pipefail

TIMEOUT_SEC="${WAIT_TIMEOUT_SEC:-180}"
START_TS=$(date +%s)

check() {
  local url="$1"
  curl -fsS "$url" >/dev/null
}

while true; do
  if check "http://localhost:8088/healthz" && check "http://localhost:8088/readyz"; then
    echo "stack ready"
    exit 0
  fi
  now=$(date +%s)
  if (( now - START_TS > TIMEOUT_SEC )); then
    echo "timeout waiting stack" >&2
    exit 1
  fi
  sleep 2
done
