# user-side 配置分工

评估链路：**Layer 1–2 直打源** · **Layer 3 源 → LiteLLM → Agent**。配置按「描述站点」与「描述测什么」拆分。

## 文件一览

| 文件 | 描述什么 | 不写什么 | 主要消费者 |
|------|----------|----------|------------|
| **`sites.json`** | 站点身份、URL、`protocol`、`supported_models`（文档摘录）、`api_key_env` | 测哪些 model、Agent 用哪条 wire | Layer 1；LiteLLM 出站 URL |
| **`assess-plan.json`** | 每层测什么：Layer 2 探测矩阵、Layer 3 模型与 OpenCode provider | 站点 URL、密钥 | `assess-protocol`；`write-litellm-config`；`t_*` |
| **`.env`** | API Key 值 | 站点结构 | 全部（Git 忽略） |
| **`.runtime/litellm.<site>.yaml`** | 生成：LiteLLM relay | — | Layer 3 |
| **`.runtime/*` Agent 配置** | 生成：`t_*` 临时配置 | — | Layer 3 smoke |

## 三种「模型列表」（勿混）

| 来源 | 含义 | 谁维护 | 用于 |
|------|------|--------|------|
| **`sites.json` → `supported_models`** | 运营商/文档宣称支持的 model id | 人工摘录 | Layer 1 与 catalog **对照**；不驱动探测 |
| **运行时 `GET /v1/models`** | 平台当前返回的 catalog | 上游 | Layer 1 **catalog 分支** |
| **`assess-plan.json` → `layer2.targets`** | 本实验要 probe 的 model × wire | 实验设计 | Layer 2 **实际探测** |
| **`assess-plan.json` → `smoke_scenarios`** | Layer 3 smoke 场景（prompt / 判定） | 实验设计 | `run-smoke` / `--smoke` |

## `sites.json` 字段

| 字段 | 说明 |
|------|------|
| `name` | 显示名 |
| `report_domain` | 报告文件名域名段（默认同站点 id） |
| `protocol` | 源原生协议 profile：`anthropic` \| `openai` \| `chat` |
| `base_url` / `anthropic_base_url` | 上游 OpenAI / Anthropic 前缀 |
| `api_key_env` | `.env` 中平台 Token 变量名 |
| `supported_models` | 文档模型列表（**≠** `/v1/models`） |
| `notes` | 自由备注 |

## `assess-plan.json` 字段

| 路径 | Layer | 说明 |
|------|-------|------|
| `sites.<id>.layer2.targets[]` | **2** | `{ "model", "wires": ["chat","messages","responses"] }`；wire 与 `sites.json` 的 `protocol` 取交集 |
| `sites.<id>.layer3.models` | **3** | 各 Agent 默认 model id（LiteLLM / `t_*`） |
| `sites.<id>.layer3.opencode` | **3** | OpenCode provider 块（`provider_id`、`npm` 等） |
| `smoke_scenarios[]` | **3 smoke** | 全局 smoke 场景；站点可 `layer3.scenarios` 覆盖 |
| `smoke_mode` | **3 smoke** | `relay`（默认，经 LiteLLM 直打）或 `agent`（`t_*` 全链路） |

Layer 1–2 记录 **`latency_ms`**（观测，不进 PASS）。Layer 2 额外记录协议面 facet：`shape` / `usage` / `stream`（soft）。

**Smoke 第 1 题 `model_probe`**：要求 JSON 自报 `model` / `release_date`，与 `layer3.models` 对照，用于发现中转站「挂名模型、实跑他模」。

全局 `protocol_profiles` / `wire_labels` 为只读参考，与 `maas.py` 内置 profile 一致。

## 配置 → 脚本映射

```text
sites.json
  └─ assess-platform.sh     Layer 1：catalog 分支 + supported_models ↔ catalog

sites.json + assess-plan.json
  └─ assess-protocol.sh     Layer 2：layer2.targets × protocol wires

sites.json + assess-plan.json + .env
  └─ litellm-proxy.sh       生成 .runtime/litellm.<site>.yaml
  └─ run-source-agent-test.sh   Layer 3：probe-relay + smoke
  └─ t_*                    Layer 3：Agent 配置（layer3.models / opencode）

assess-source.sh            Layer 1 + 2 + 3 一键
run-user-side-compat.sh     Layer 1–2 一次 + protocol scope 内多 Agent Layer 3
```

## 报告

路径：`docs/reports/{report_domain}-源评估报告-{YYYY-MM-DD}.md`

**推荐**：由结构化评估结果自动生成（勿手抄日志）：

```bash
cd experiment/user-side
source .env
./scripts/assess-source.sh --site <id> --agent <name> --write-report
# 或
python3 lib/maas.py assess-source --site <id> --agent <name> --write-report
```

机器可读结果：`.runtime/<site>-assess-<YYYYMMDD>.json`（与报告同日）。

查询报告路径：

```bash
python3 lib/maas.py report-path --site <id> --relative
```

设计稿：[EC2-用户侧隔离实验点设计 §2.1](../../docs/experiment/EC2-用户侧隔离实验点设计.md#21-三层评估法)
