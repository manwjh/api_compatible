#!/usr/bin/env bash
# PoC helper: LiteLLM /v1/responses → b.ai /v1/chat/completions for Codex.
# Usage:
#   ./scripts/poc-litellm-bai-codex.sh start    # start proxy in background
#   ./scripts/poc-litellm-bai-codex.sh probe    # L2 curl probes
#   ./scripts/poc-litellm-bai-codex.sh codex    # L3 non-stream Codex exec
#   ./scripts/poc-litellm-bai-codex.sh stop
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG="${ROOT}/.runtime/litellm.bai.yaml"
CODEX_CFG="${ROOT}/.runtime/codex.litellm-bai.toml"
PID_FILE="${ROOT}/.runtime/litellm.bai.pid"
LOG_FILE="${ROOT}/.runtime/litellm.bai.log"
PORT=4000
LITELLM_KEY="sk-litellm-local-poc"
CODEX_BIN="${CODEX_BIN:-/Applications/Codex.app/Contents/Resources/codex}"

cd "$ROOT"
mkdir -p .runtime

if [[ -f "${ROOT}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${ROOT}/.env"
  set +a
fi

if [[ -z "${BAI_API_KEY:-}" || "${BAI_API_KEY}" == "sk-your-bai-api-key" ]]; then
  echo "Missing BAI_API_KEY in .env" >&2
  exit 1
fi

start_proxy() {
  if [[ -f "$PID_FILE" ]] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
    echo "LiteLLM already running (pid $(cat "$PID_FILE"))"
    return 0
  fi
  export BAI_API_KEY
  nohup python3 -m litellm.proxy.proxy_cli --config "$CONFIG" --port "$PORT" \
    >"$LOG_FILE" 2>&1 &
  echo $! >"$PID_FILE"
  for _ in $(seq 1 30); do
    if curl -sf "http://127.0.0.1:${PORT}/health/liveliness" >/dev/null 2>&1; then
      echo "LiteLLM ready on http://127.0.0.1:${PORT}"
      return 0
    fi
    sleep 1
  done
  echo "LiteLLM failed to start; see $LOG_FILE" >&2
  tail -20 "$LOG_FILE" >&2 || true
  exit 1
}

stop_proxy() {
  if [[ -f "$PID_FILE" ]]; then
    kill "$(cat "$PID_FILE")" 2>/dev/null || true
    rm -f "$PID_FILE"
    echo "Stopped LiteLLM"
  fi
}

probe_responses() {
  start_proxy
  echo "=== L2: POST /v1/responses (non-stream) ==="
  curl -sS -w "\nHTTP %{http_code}\n" \
    "http://127.0.0.1:${PORT}/v1/responses" \
    -H "Authorization: Bearer ${LITELLM_KEY}" \
    -H "Content-Type: application/json" \
    -d '{"model":"gpt-5-mini","input":"Reply with exactly: LITELLM_OK"}'
  echo ""
  echo "=== L3: POST /v1/responses (stream) ==="
  curl -sS -N -w "\nHTTP %{http_code}\n" \
    "http://127.0.0.1:${PORT}/v1/responses" \
    -H "Authorization: Bearer ${LITELLM_KEY}" \
    -H "Content-Type: application/json" \
    -d '{"model":"gpt-5-mini","input":"Reply OK","stream":true}' | head -20
}

run_codex() {
  start_proxy
  export OPENAI_API_KEY="$LITELLM_KEY"
  export CODEX_HOME="${ROOT}/.runtime/codex-home-poc"
  mkdir -p "$CODEX_HOME"
  cp "$CODEX_CFG" "$CODEX_HOME/config.toml"
  echo "=== Codex exec (L3 text) ==="
  "$CODEX_BIN" -c "$CODEX_HOME/config.toml" exec --skip-git-repo-check \
    --model gpt-5-mini \
    "Reply with exactly one word: CODEX_OK" 2>&1 | tail -30
}

case "${1:-}" in
  start) start_proxy ;;
  stop) stop_proxy ;;
  probe) probe_responses ;;
  codex) run_codex ;;
  *)
    echo "Usage: $0 {start|stop|probe|codex}" >&2
    exit 1
    ;;
esac
