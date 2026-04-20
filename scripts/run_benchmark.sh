#!/usr/bin/env bash
set -euo pipefail

SCENARIO="${1:-smoke}"
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
TS=$(date -u +%Y%m%dT%H%M%SZ)
OUT_DIR="$ROOT_DIR/artifacts/bench/$TS/$SCENARIO"
mkdir -p "$OUT_DIR"

POST_IDS_FILE="${POST_IDS_FILE:-$ROOT_DIR/artifacts/seed/post_ids.txt}"
POST_IDS=""
if [[ -f "$POST_IDS_FILE" ]]; then
  POST_IDS=$(head -n "${POST_ID_SAMPLE_SIZE:-1000}" "$POST_IDS_FILE" | paste -sd, -)
fi

set +e
docker run --rm --network host \
  -v "$ROOT_DIR/bench/k6:/scripts" \
  -v "$OUT_DIR:/out" \
  -e BASE_URL="${BASE_URL:-http://localhost:8088}" \
  -e SCENARIO="$SCENARIO" \
  -e TARGET_QPS="${TARGET_QPS:-200}" \
  -e TEST_DURATION="${TEST_DURATION:-30s}" \
  -e VUS_MAX="${VUS_MAX:-400}" \
  -e JWT_SECRET="${JWT_SECRET:-dev-secret}" \
  -e HOT_FEED_LIMIT="${HOT_FEED_LIMIT:-50}" \
  -e POST_IDS="$POST_IDS" \
  -e FAIL_ERROR_RATE="${FAIL_ERROR_RATE:-0.01}" \
  -e FAIL_P95_MS="${FAIL_P95_MS:-500}" \
  -e FAIL_P99_MS="${FAIL_P99_MS:-1000}" \
  grafana/k6:0.51.0 run /scripts/main.js --summary-export=/out/summary.json > "$OUT_DIR/stdout.txt" 2>&1
RC=$?
set -e

cat > "$OUT_DIR/report.md" <<RPT
# Benchmark Report
- Scenario: $SCENARIO
- Timestamp(UTC): $TS
- Target QPS: ${TARGET_QPS:-200}
- Duration: ${TEST_DURATION:-30s}
- Exit Code: $RC
- Summary JSON: summary.json
- Raw output: stdout.txt
RPT

if [[ $RC -ne 0 ]]; then
  echo "benchmark failed (rc=$RC). see $OUT_DIR/stdout.txt" >&2
  exit $RC
fi

echo "$OUT_DIR"
