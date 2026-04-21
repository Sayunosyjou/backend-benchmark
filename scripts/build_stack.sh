#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

SERVICES=(core-service web-node gateway consumer seeder)

if [[ "${DISABLE_BUILDKIT:-0}" == "1" ]]; then
  export DOCKER_BUILDKIT=0
  export COMPOSE_DOCKER_CLI_BUILD=0
fi

for svc in "${SERVICES[@]}"; do
  echo "[build] $svc"
  docker compose build "$svc"
done

echo "[up] docker compose up -d"
docker compose up -d
