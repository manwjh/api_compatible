# AGENTS.md — user-side 协作规则

配置分工见 **[CONFIG.md](./CONFIG.md)**（`sites.json` = 描述站点 · `assess-plan.json` = 测试计划）。

## 三层源评估

| 层 | 脚本 | 拓扑 |
|----|------|------|
| 1 | `scripts/assess-platform.sh` | 直打源 |
| 2 | `scripts/assess-protocol.sh` | 直打源 |
| 3 | `scripts/run-source-agent-test.sh` | 源 → LiteLLM → Agent |

一键：`scripts/assess-source.sh --site ID --agent NAME [--smoke]`

设计稿：[EC2-用户侧隔离实验点设计 §2.1](../../docs/experiment/EC2-用户侧隔离实验点设计.md#21-三层评估法)

## 配置文件

| 路径 | 职责 |
|------|------|
| `sites.json` | 站点：URL、`protocol`、`supported_models`（文档摘录） |
| `assess-plan.json` | 测试：`layer2.targets`、`layer3.models` / `opencode` |
| `.env` | 密钥（Git 忽略） |

## Git

禁止提交 `.env`、`.runtime/`。结论写 `docs/reports/`（命名见 CONFIG.md）。
