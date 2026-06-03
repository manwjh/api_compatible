#!/usr/bin/env python3
"""Site registry and config helpers for t_* launchers (sites.json)."""

from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any


def repo_root() -> Path:
    return Path(__file__).resolve().parent.parent


def load_sites(root: Path | None = None) -> dict[str, Any]:
    path = (root or repo_root()) / "sites.json"
    with path.open(encoding="utf-8") as fh:
        return json.load(fh)


def load_dotenv(root: Path | None = None) -> None:
    path = (root or repo_root()) / ".env"
    if not path.is_file():
        return
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, value = line.partition("=")
        key = key.strip()
        value = value.strip().strip('"').strip("'")
        os.environ.setdefault(key, value)


def apply_proxy_env() -> None:
    if os.environ.get("MAAS_PROXY_SKIP"):
        return
    proxy = (
        os.environ.get("MAAS_PROXY")
        or os.environ.get("HTTPS_PROXY")
        or os.environ.get("ALL_PROXY")
    )
    if not proxy:
        import socket

        for port in (10808, 10809, 7890, 1087):
            try:
                with socket.create_connection(("127.0.0.1", port), timeout=0.3):
                    proxy = f"socks5h://127.0.0.1:{port}"
                    break
            except OSError:
                continue
    if not proxy:
        return
    os.environ.setdefault("MAAS_PROXY", proxy)
    os.environ.setdefault("ALL_PROXY", proxy)
    os.environ.setdefault("HTTPS_PROXY", proxy)
    os.environ.setdefault("HTTP_PROXY", proxy)


def site_entry(sites: dict[str, Any], site_id: str | None) -> tuple[str, dict[str, Any]]:
    sid = site_id or sites.get("default_site")
    if not sid:
        raise SystemExit("No site specified and sites.json has no default_site")
    entry = sites.get("sites", {}).get(sid)
    if not entry:
        raise SystemExit(f"Unknown site: {sid}")
    return sid, entry


def api_key_for(site: dict[str, Any]) -> str:
    env_name = site["api_key_env"]
    key = os.environ.get(env_name, "").strip()
    if not key:
        raise SystemExit(
            f"Missing API key: set {env_name} in .env (see .env.example)"
        )
    return key


def _http_opener() -> urllib.request.OpenerDirector:
    proxy = (
        os.environ.get("HTTPS_PROXY")
        or os.environ.get("ALL_PROXY")
        or os.environ.get("HTTP_PROXY")
        or os.environ.get("MAAS_PROXY")
    )
    if proxy:
        return urllib.request.build_opener(
            urllib.request.ProxyHandler({"http": proxy, "https": proxy})
        )
    return urllib.request.build_opener()


def http_json(
    method: str,
    url: str,
    api_key: str,
    body: dict[str, Any] | None = None,
    extra_headers: dict[str, str] | None = None,
    timeout: float = 60,
) -> tuple[int, Any]:
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
    }
    if extra_headers:
        headers.update(extra_headers)
    data = None
    if body is not None:
        data = json.dumps(body).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with _http_opener().open(req, timeout=timeout) as resp:
            raw = resp.read().decode("utf-8")
            return resp.status, json.loads(raw) if raw else None
    except urllib.error.HTTPError as exc:
        raw = exc.read().decode("utf-8", errors="replace")
        try:
            payload: Any = json.loads(raw) if raw else None
        except json.JSONDecodeError:
            payload = raw
        return exc.code, payload
    except urllib.error.URLError as exc:
        raise SystemExit(f"Network error for {url}: {exc.reason}") from exc


def cmd_list_sites(sites: dict[str, Any]) -> None:
    default = sites.get("default_site")
    for sid, entry in sites.get("sites", {}).items():
        mark = " (default)" if sid == default else ""
        name = entry.get("name", sid)
        base = entry.get("base_url", "")
        print(f"{sid}{mark}\t{name}\t{base}")


def cmd_get(args: argparse.Namespace, sites: dict[str, Any]) -> None:
    if args.field == "default_site":
        print(sites.get("default_site", ""))
        return
    _, site = site_entry(sites, args.site)
    field = args.field
    if field == "json":
        print(json.dumps(site, ensure_ascii=False, indent=2))
    elif field == "api_key_env":
        print(site["api_key_env"])
    elif field == "base_url":
        print(site["base_url"])
    elif field == "anthropic_base_url":
        print(
            site.get(
                "anthropic_base_url",
                site["base_url"].rstrip("/").removesuffix("/v1"),
            )
        )
    elif field == "default_model":
        agent = args.agent
        if not agent:
            raise SystemExit("--agent required for default_model")
        print(site.get("default_models", {}).get(agent, ""))
    elif field == "notes":
        print(site.get("notes", ""))
    else:
        raise SystemExit(f"Unknown field: {field}")


def cmd_list_models(args: argparse.Namespace, sites: dict[str, Any]) -> None:
    load_dotenv()
    sid, site = site_entry(sites, args.site)
    key = api_key_for(site)
    base = site["base_url"].rstrip("/")
    status, payload = http_json("GET", f"{base}/models", key)
    if status != 200:
        raise SystemExit(f"GET /v1/models failed ({status}): {payload}")
    models = payload.get("data", payload) if isinstance(payload, dict) else payload
    if not isinstance(models, list):
        raise SystemExit(f"Unexpected /v1/models payload: {payload!r}")
    for item in models:
        if isinstance(item, dict):
            print(item.get("id", item))
        else:
            print(item)
    if args.limit and len(models) > args.limit:
        return


def write_claude_config(
    site: dict[str, Any], api_key: str, out: Path
) -> None:
    anthropic_base = site.get(
        "anthropic_base_url",
        site["base_url"].rstrip("/").removesuffix("/v1"),
    )
    payload = {
        "env": {
            "ANTHROPIC_BASE_URL": anthropic_base,
            "ANTHROPIC_AUTH_TOKEN": api_key,
            "ANTHROPIC_API_KEY": api_key,
            "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
        }
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")


def write_codex_config(
    site: dict[str, Any], model: str, out: Path
) -> None:
    lines = [
        f'model = "{model}"',
        f'openai_base_url = "{site["base_url"]}"',
        "",
    ]
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text("\n".join(lines), encoding="utf-8")


def write_opencode_config(
    site: dict[str, Any], api_key: str, model: str, out: Path
) -> None:
    oc = site.get("opencode", {})
    provider_id = oc.get("provider_id", site.get("name", "custom").lower())
    provider_name = oc.get("provider_name", site.get("name", provider_id))
    npm = oc.get("npm", "@ai-sdk/openai-compatible")
    payload = {
        "$schema": "https://opencode.ai/config.json",
        "model": f"{provider_id}/{model}",
        "provider": {
            provider_id: {
                "npm": npm,
                "name": provider_name,
                "options": {
                    "baseURL": site["base_url"],
                    "apiKey": api_key,
                },
                "models": {
                    model: {"name": model},
                },
            }
        },
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")


def cmd_write_config(args: argparse.Namespace, sites: dict[str, Any]) -> None:
    load_dotenv()
    sid, site = site_entry(sites, args.site)
    key = api_key_for(site)
    out = Path(args.out)
    if args.kind == "claude":
        write_claude_config(site, key, out)
    elif args.kind == "codex":
        model = args.model or site.get("default_models", {}).get("codex", "")
        if not model:
            raise SystemExit("No model specified and site has no default codex model")
        write_codex_config(site, model, out)
    elif args.kind == "opencode":
        model = args.model or site.get("default_models", {}).get("opencode", "")
        if not model:
            raise SystemExit("No model specified and site has no default opencode model")
        write_opencode_config(site, key, model, out)
    print(out)


def probe_label(status: int, ok_codes: set[int]) -> str:
    if status in ok_codes:
        return "OK"
    if status in {401, 403}:
        return f"HTTP {status}"
    if status == 404:
        return "HTTP 404"
    return f"HTTP {status}"


def cmd_probe(args: argparse.Namespace, sites: dict[str, Any]) -> None:
    load_dotenv()
    sid, site = site_entry(sites, args.site)
    key = api_key_for(site)
    base = site["base_url"].rstrip("/")
    model_chat = site.get("default_models", {}).get("codex", "gpt-5-mini")
    model_msg = site.get("default_models", {}).get("claude", "claude-haiku-4.5")

    results: list[tuple[str, str, str]] = []

    st, _ = http_json("GET", f"{base}/models", key)
    results.append(("GET", "/v1/models", probe_label(st, {200})))

    st, _ = http_json(
        "POST",
        f"{base}/chat/completions",
        key,
        {
            "model": model_chat,
            "messages": [{"role": "user", "content": "Reply OK"}],
            "max_tokens": 16,
        },
    )
    results.append(("POST", "/v1/chat/completions", probe_label(st, {200})))

    st, _ = http_json(
        "POST",
        f"{base}/messages",
        key,
        {
            "model": model_msg,
            "max_tokens": 16,
            "messages": [{"role": "user", "content": "Reply OK"}],
        },
        extra_headers={"anthropic-version": "2023-06-01"},
    )
    results.append(("POST", "/v1/messages", probe_label(st, {200})))

    st, _ = http_json(
        "POST",
        f"{base}/responses",
        key,
        {"model": model_chat, "input": "Reply OK"},
    )
    results.append(("POST", "/v1/responses", probe_label(st, {200})))

    print(f"Site: {sid} ({site.get('name', sid)})")
    print(f"Base: {site['base_url']}")
    print("")
    print(f"{'Method':<6} {'Endpoint':<24} Result")
    print("-" * 44)
    for method, endpoint, result in results:
        print(f"{method:<6} {endpoint:<24} {result}")

    agents = {
        "OpenCode": results[1][2] == "OK",
        "Claude Code": results[2][2] == "OK",
        "Codex": results[3][2] == "OK",
    }
    print("")
    print("Agent readiness (protocol probe):")
    for agent, ready in agents.items():
        mark = "yes" if ready else "no"
        print(f"  {agent}: {mark}")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)

    p = sub.add_parser("list-sites", help="List registered upstream sites")
    p.set_defaults(func=lambda a, s: cmd_list_sites(s))

    p = sub.add_parser("get", help="Print a site field")
    p.add_argument("field", choices=[
        "default_site",
        "json",
        "api_key_env",
        "base_url",
        "anthropic_base_url",
        "default_model",
        "notes",
    ])
    p.add_argument("--site")
    p.add_argument("--agent", choices=["claude", "codex", "opencode"])
    p.set_defaults(func=cmd_get)

    p = sub.add_parser("list-models", help="List models from GET /v1/models")
    p.add_argument("--site")
    p.add_argument("--limit", type=int, default=0)
    p.set_defaults(func=cmd_list_models)

    for kind in ("claude", "codex", "opencode"):
        p = sub.add_parser(f"write-{kind}-config", help=f"Write {kind} temp config")
        p.add_argument("--site")
        p.add_argument("--out", required=True)
        p.add_argument("--model")
        p.add_argument("--kind", default=kind)
        p.set_defaults(func=cmd_write_config)

    p = sub.add_parser("probe-endpoints", help="Probe key HTTP endpoints for a site")
    p.add_argument("--site")
    p.set_defaults(func=cmd_probe)

    return parser


def main() -> None:
    root = repo_root()
    sites = load_sites(root)
    parser = build_parser()
    args = parser.parse_args()
    if args.command != "list-sites" and args.command != "get":
        load_dotenv(root)
        apply_proxy_env()
    args.func(args, sites)


if __name__ == "__main__":
    main()
