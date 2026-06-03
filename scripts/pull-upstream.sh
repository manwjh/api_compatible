#!/usr/bin/env bash
# Clone/update optional upstream reference trees (gitignored). See AGENTS.md.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

pull_repo() {
  local name="$1"
  local url="$2"
  local branch="$3"
  local dir="$ROOT/$name"

  if [[ -d "$dir/.git" ]]; then
    echo "→ Updating $name ($branch)..."
    git -C "$dir" fetch --depth 1 origin "$branch"
    git -C "$dir" checkout "$branch"
    git -C "$dir" reset --hard "origin/$branch"
  elif [[ -e "$dir" ]]; then
    echo "Error: $dir exists but is not a git repo. Remove it and retry." >&2
    return 1
  else
    echo "→ Cloning $name ($branch)..."
    git clone --depth 1 --branch "$branch" "$url" "$dir"
  fi
  echo "→ Done: $dir"
}

usage() {
  cat <<'EOF'
Usage: scripts/pull-upstream.sh [opencode|newapi|codex|all]

Pull optional upstream reference code into the project root.
These directories are gitignored and not needed to run ./t_claude / t_codex / t_opencode.

  opencode   anomalyco/opencode (branch dev)
  newapi     QuantumNous/new-api (branch main)
  codex      openai/codex (branch main)
  all        opencode + newapi + codex (default)
EOF
}

target="${1:-all}"
case "$target" in
  -h|--help)
    usage
    exit 0
    ;;
  opencode)
    pull_repo opencode https://github.com/anomalyco/opencode dev
    ;;
  newapi)
    pull_repo newapi https://github.com/QuantumNous/new-api main
    ;;
  codex)
    pull_repo codex https://github.com/openai/codex main
    ;;
  all)
    pull_repo opencode https://github.com/anomalyco/opencode dev
    pull_repo newapi https://github.com/QuantumNous/new-api main
    pull_repo codex https://github.com/openai/codex main
    ;;
  *)
    echo "Unknown target: $target" >&2
    usage >&2
    exit 1
    ;;
esac
