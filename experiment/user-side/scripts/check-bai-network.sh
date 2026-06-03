#!/usr/bin/env bash
# Quick connectivity check for b.ai (direct vs local proxy).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT}/lib/maas.sh"
maas_load_env

PROXY="${HTTPS_PROXY:-${MAAS_PROXY:-socks5h://127.0.0.1:10808}}"

echo "DNS (system): $(dig +short b.ai A 2>/dev/null | head -3 | tr '\n' ' ')"
echo "Proxy env:    ${HTTPS_PROXY:-<unset>}"
echo

check() {
  local label="$1"
  shift
  if curl -sS -o /dev/null -w "%{http_code}" --connect-timeout 8 "$@" https://b.ai/ >/tmp/bai_check_code 2>/tmp/bai_check_err; then
    echo "${label}: HTTP $(cat /tmp/bai_check_code) OK"
  else
    echo "${label}: FAIL ($(tr '\n' ' ' </tmp/bai_check_err | head -c 120))"
  fi
}

check "Direct (no proxy env)" env -u ALL_PROXY -u HTTPS_PROXY -u HTTP_PROXY
check "Via proxy" -x "$PROXY"

if [[ -f "${ROOT}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${ROOT}/.env"
  set +a
  if [[ -n "${BAI_API_KEY:-}" ]]; then
    if curl -sS -o /dev/null -w "%{http_code}" --connect-timeout 15 -x "$PROXY" \
      -H "Authorization: Bearer ${BAI_API_KEY}" https://api.b.ai/v1/models >/tmp/bai_api_code 2>/dev/null; then
      echo "api.b.ai /v1/models (with key): HTTP $(cat /tmp/bai_api_code)"
    fi
  fi
fi

rm -f /tmp/bai_check_code /tmp/bai_check_err /tmp/bai_api_code 2>/dev/null || true
