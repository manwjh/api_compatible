# user-side — 用户侧隔离实验点

本目录是 **[EC2-用户侧隔离实验点设计](../../docs/experiment/EC2-用户侧隔离实验点设计.md)** 的可复现实现：`sites.json` 登记上游、`t_*` 驱动真实 Agent CLI、`scripts/` 做 L2 探测与 smoke。

索引：[experiment/README.md](../README.md) · 证据与方法论在 [`docs/`](../../docs/README.md)。

## 快速开始

在 **本目录**（`experiment/user-side/`）下操作：

```bash
cp .env.example .env   # 填入 sites.json 中各 api_key_env 对应的 Key
./t_claude --site b.ai --model claude-haiku-4.5 -y

# EC2 Runner 上一键 L2 + 可选 L3 smoke
./scripts/run-user-side-compat.sh --site b.ai --probe-only
./scripts/run-user-side-compat.sh --site newapi-prototype --smoke
```

从仓库根目录：`./experiment/user-side/t_claude --site b.ai -y`

## 目录

| 路径 | 说明 |
|------|------|
| `t_claude` / `t_codex` / `t_opencode` | Agent 启动器（共用 `lib/maas.*`） |
| `lib/maas.sh`, `lib/maas.py` | 站点解析、临时配置、L2 probe |
| `sites.json` | 上游站点注册表（无密钥） |
| `.env.example` | 密钥环境变量模板 |
| `opencode.json.example` | OpenCode 手工配置示例 |
| `scripts/probe-endpoints.sh` | L2 端点探测 |
| `scripts/run-user-side-compat.sh` | probe + 可选 `t_*` smoke |
| `scripts/check-bai-network.sh` | b.ai 网络/代理诊断 |
| `scripts/poc-litellm-bai-codex.sh` | LiteLLM PoC（探索性，非正式接口） |

协作规则：[AGENTS.md](./AGENTS.md)。
