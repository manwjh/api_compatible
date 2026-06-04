#!/usr/bin/env bash
# Layer 3: Source → LiteLLM → Agent (relay probe + optional smoke).
# Terminology: Layer 3 = assessment tier; L3/L4 = E2E depth inside smoke.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SITE=""
AGENT=""
PROBE_ONLY=0
SMOKE=0

usage() {
  cat <<'EOF'
Usage: experiment/user-side/scripts/run-source-agent-test.sh --site SITE --agent AGENT [options]

Layer 3 — upstream source → LiteLLM → specified Agent.

Required:
  --site ID       sites.json site id (upstream source)
  --agent NAME    claude | codex | opencode (must be in protocol scope)

Options:
  --probe-only    Layer 3 relay wire probe only (E2E L2 via LiteLLM)
  --smoke         relay probe + t_* non-interactive smoke (E2E L3+)
  -h, --help      Show this help

Examples:
  ./scripts/run-source-agent-test.sh --site b.ai --agent codex --probe-only
  ./scripts/run-source-agent-test.sh --site b.ai --agent claude --smoke
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --site)
      SITE="${2:-}"
      shift 2
      ;;
    --agent)
      AGENT="${2:-}"
      shift 2
      ;;
    --probe-only)
      PROBE_ONLY=1
      shift
      ;;
    --smoke)
      SMOKE=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -z "$SITE" || -z "$AGENT" ]]; then
  echo "Error: --site and --agent are required" >&2
  usage >&2
  exit 1
fi

case "$AGENT" in
  claude|codex|opencode) ;;
  *)
    echo "Error: --agent must be claude, codex, or opencode" >&2
    exit 1
    ;;
esac

if [[ "$PROBE_ONLY" -eq 0 && "$SMOKE" -eq 0 ]]; then
  echo "Error: specify --probe-only and/or --smoke" >&2
  usage >&2
  exit 1
fi

if [[ -f "${ROOT}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${ROOT}/.env"
  set +a
fi

cd "$ROOT"

IN_SCOPE="$(python3 "${ROOT}/lib/maas.py" get assess_agents --site "$SITE")"
if [[ ",${IN_SCOPE}," != *",${AGENT},"* ]]; then
  echo "Error: agent '${AGENT}' not in protocol scope for site '${SITE}' (in scope: ${IN_SCOPE})" >&2
  exit 1
fi

echo "==> Starting LiteLLM relay for site=${SITE}"
./scripts/litellm-proxy.sh start --site "$SITE"

echo "==> Layer 3 relay probe: site=${SITE} agent=${AGENT}"
python3 "${ROOT}/lib/maas.py" probe-relay --site "$SITE" --agent "$AGENT"

if [[ "$PROBE_ONLY" -eq 1 && "$SMOKE" -eq 0 ]]; then
  echo "==> Done (--probe-only)"
  exit 0
fi

if [[ "$SMOKE" -eq 0 ]]; then
  exit 0
fi

echo "==> Layer 3 smoke scenarios: site=${SITE} agent=${AGENT}"
python3 "${ROOT}/lib/maas.py" run-smoke --site "$SITE" --agent "$AGENT"

echo "==> Smoke finished: site=${SITE} agent=${AGENT}"
