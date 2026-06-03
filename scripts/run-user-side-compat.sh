#!/usr/bin/env bash
# User-side EC2 compatibility automation: L2 probe + optional L3 smoke via t_*.
# See docs/experiment/EC2-用户侧隔离实验点设计.md §8.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SITE=""
PROBE_ONLY=0
SMOKE=0
AGENTS=(claude codex opencode)

usage() {
  cat <<'EOF'
Usage: scripts/run-user-side-compat.sh --site SITE [options]

Run on the user-side EC2 Runner (see docs/experiment/EC2-用户侧隔离实验点设计.md).

Options:
  --site ID       Required. sites.json site id (e.g. b.ai, newapi-prototype)
  --probe-only    Only run probe-endpoints.sh (L2)
  --smoke         After probe, run non-interactive smoke via ./t_* -y
  --agents LIST   Comma-separated: claude,codex,opencode (default: all)
  -h, --help      Show this help

Examples:
  ./scripts/run-user-side-compat.sh --site b.ai --probe-only
  ./scripts/run-user-side-compat.sh --site newapi-prototype --smoke
  ./scripts/run-user-side-compat.sh --site b.ai --smoke --agents claude,opencode
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --site)
      SITE="${2:-}"
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
    --agents)
      IFS=',' read -r -a AGENTS <<< "${2:-}"
      shift 2
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

if [[ -z "$SITE" ]]; then
  echo "Error: --site is required" >&2
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

echo "==> L2 probe: site=${SITE}"
./scripts/probe-endpoints.sh "$SITE"

if [[ "$PROBE_ONLY" -eq 1 ]]; then
  echo "==> Done (--probe-only)"
  exit 0
fi

if [[ "$SMOKE" -eq 0 ]]; then
  echo "==> L2 complete. Use --smoke for L3 smoke, or run ./t_* manually for L4."
  exit 0
fi

smoke_claude() {
  echo "==> smoke: t_claude"
  ./t_claude --site "$SITE" -y -- --print --max-budget-usd 1.00 "Reply with exactly: API OK"
}

smoke_codex() {
  echo "==> smoke: t_codex exec"
  ./t_codex --site "$SITE" -y -- exec "Reply with exactly: API OK"
}

smoke_opencode() {
  echo "==> smoke: t_opencode run"
  ./t_opencode --site "$SITE" -y -- run "Reply with exactly: API OK"
}

for agent in "${AGENTS[@]}"; do
  case "$agent" in
    claude) smoke_claude ;;
    codex) smoke_codex ;;
    opencode) smoke_opencode ;;
    *)
      echo "Warning: unknown agent '${agent}', skip" >&2
      ;;
  esac
done

echo "==> Smoke finished for site=${SITE}"
