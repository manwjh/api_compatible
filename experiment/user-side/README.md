# user-side — 用户侧隔离实验点

本目录实现 **[EC2-用户侧隔离实验点设计](../../docs/experiment/EC2-用户侧隔离实验点设计.md)** 的 **三层源评估法**。

**配置分工**：[`CONFIG.md`](./CONFIG.md) — `sites.json` 描述站点 · `assess-plan.json` 描述测什么。

## 三层评估

| 层 | 脚本 | 配置 |
|----|------|------|
| 1 平台链接 | `assess-platform.sh` | `sites.json` + 运行时 catalog |
| 2 基础协议 | `assess-protocol.sh` | `assess-plan.json` → `layer2` |
| 3 指定 Agent | `run-source-agent-test.sh` | `assess-plan.json` → `layer3` + LiteLLM |

```bash
cp .env.example .env

./scripts/assess-platform.sh --site ai.oai.red
./scripts/assess-protocol.sh --site ai.oai.red
./scripts/assess-source.sh --site ai.oai.red --agent opencode --write-report
# 可选 L4 smoke：
./scripts/assess-source.sh --site ai.oai.red --agent opencode --smoke --write-report
```

报告由 `--write-report` 从结构化 JSON 自动生成（`docs/reports/`）；勿手抄日志。

协作规则：[AGENTS.md](./AGENTS.md)
