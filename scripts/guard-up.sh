#!/usr/bin/env bash
set -euo pipefail

project_name="${COMPOSE_PROJECT_NAME:-$(basename "$PWD")}"
legacy_services="$(docker ps --filter "label=com.docker.compose.project=${project_name}" --format '{{.Label "com.docker.compose.service"}}' | grep -E '^(web|web-legacy)$' || true)"

if [[ -n "${legacy_services}" ]]; then
  echo "❌ Legacy UI container(s) detected. Startup aborted to avoid running old pre-React frontend."
  echo
  echo "Detected:"
  echo "${legacy_services}" | sort -u | sed 's/^/  - /'
  echo
  echo "Stop/remove them first:"
  echo "  docker compose --profile legacy-ui down"
  echo "  docker rm -f link-manager-web-1 2>/dev/null || true"
  exit 1
fi

echo "✅ No legacy UI containers detected. Starting stack..."
exec docker compose up -d --build "$@"