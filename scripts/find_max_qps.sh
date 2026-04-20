#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
START_QPS="${START_QPS:-200}"
STEP_QPS="${STEP_QPS:-200}"
MAX_ROUNDS="${MAX_ROUNDS:-10}"
TEST_DURATION="${TEST_DURATION:-30s}"
LAST_OK=0

for ((i=1; i<=MAX_ROUNDS; i++)); do
  QPS=$((START_QPS + (i-1)*STEP_QPS))
  echo "[find-max-qps] round=$i target_qps=$QPS"
  if TARGET_QPS="$QPS" TEST_DURATION="$TEST_DURATION" "$ROOT_DIR/scripts/run_benchmark.sh" mixed; then
    LAST_OK=$QPS
  else
    echo "[find-max-qps] threshold breached at $QPS"
    break
  fi
done

echo "max stable qps (approx): $LAST_OK"
