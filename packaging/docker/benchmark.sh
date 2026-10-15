#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
compose=(docker compose -f "$root/packaging/docker/docker-compose.yml" --env-file "$root/packaging/docker/.env")
container="$("${compose[@]}" ps -q surf)"
if [ -z "$container" ]; then
  echo "Surf is not running; run 'make up' first." >&2
  exit 1
fi

before="$(docker stats --no-stream --format '{{json .}}' "$container")"
session="$("${compose[@]}" --profile tools run --rm benchmark)"
after="$(docker stats --no-stream --format '{{json .}}' "$container")"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
output="$root/.local/benchmarks/surf-$timestamp.json"
mkdir -p "$(dirname "$output")"
printf '{"timestamp":"%s","host_before":%s,"session":%s,"host_after":%s}\n' \
  "$timestamp" "$before" "$session" "$after" | tee "$output"
printf 'Saved benchmark: %s\n' "$output"
