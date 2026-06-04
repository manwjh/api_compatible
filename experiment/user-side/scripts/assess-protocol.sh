#!/usr/bin/env bash
# Layer 2: native protocol surface on source (direct upstream).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SITE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      cat <<'EOF'
Usage: experiment/user-side/scripts/assess-protocol.sh --site SITE

Layer 2 — Model × wire from assess-plan.json layer2 targets.
Catalog branch from Layer 1: listed (compare + test) | empty/unavailable (blind test).
EOF
      exit 0
      ;;
    --site)
      SITE="${2:-}"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

[[ -n "$SITE" ]] || { echo "Error: --site required" >&2; exit 1; }

if [[ -f "${ROOT}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${ROOT}/.env"
  set +a
fi

exec python3 "${ROOT}/lib/maas.py" assess-protocol --site "$SITE"
