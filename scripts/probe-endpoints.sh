#!/usr/bin/env bash
# Probe upstream HTTP endpoints for Agent protocol compatibility.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SITE="${1:-}"

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  cat <<'EOF'
Usage: scripts/probe-endpoints.sh [SITE]

Probe GET /v1/models and POST on chat/completions, messages, responses.
SITE defaults to sites.json → default_site.

Examples:
  ./scripts/probe-endpoints.sh
  ./scripts/probe-endpoints.sh b.ai
EOF
  exit 0
fi

if [[ -f "${ROOT}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${ROOT}/.env"
  set +a
fi

args=(probe-endpoints)
if [[ -n "$SITE" ]]; then
  args+=(--site "$SITE")
fi

exec python3 "${ROOT}/lib/maas.py" "${args[@]}"
