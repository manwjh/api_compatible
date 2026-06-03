# E2E 原生兼容性全景

> **文档类型**：参考矩阵 · **范围**：原厂 / 云托管上游 × 三 Agent（Claude Code、Codex、OpenCode）· **不含** 中转站与自建网关  
> **厂商会持续改版**，以官方文档为准。表中 **●** 表示 L1+L2 原生对齐；L3–L5 实测见 [reports/](./reports/)。

### 文档元信息

| 项 | 内容 |
|----|------|
| **编写日期** | 2026-06-03 |
| **矩阵基线** | 依据各 Agent / 上游 **官方集成文档** 整理；L3–L5 以 [reports/](./reports/)（2026-06-01 实测）为参照 |
| **复审触发** | 任一下列 **评估标的** 大版本升级、官方新增/废弃 provider、或上游 API 改版 |

### 评估标的版本（矩阵所依）

| 标的 | 版本 / 门槛 | 来源 |
|------|-------------|------|
| **Claude Code CLI** | `2.1.159` | [ClaudeCode 兼容性评估报告](./reports/ClaudeCode兼容性评估报告.md)（2026-06-01） |
| **Codex CLI** | `0.133.0` 起（`wire_api = "responses"` 唯一）；报告测 `0.133.0`，当时最新已知 `0.135.0` | [Codex 兼容性评估报告](./reports/Codex兼容性评估报告.md) |
| **OpenCode CLI** | `1.15.13` | [OpenCode 兼容性评估报告](./reports/OpenCode兼容性评估报告.md) |

### 协议 / 上游参照版本（附录 curl 与集成说明）

| 参照项 | 版本或取值 |
|--------|------------|
| Anthropic Messages | Header `anthropic-version: 2023-06-01` |
| Azure OpenAI（Codex Responses 示例） | Query `api-version=2025-04-01-preview` |
| Codex 主 wire | `wire_api = "responses"`（**0.133 起**强制，Chat Completions 已移除） |

> 上游模型 ID、区域开通与 Bedrock mantle 能力以各云控制台为准，不单独锁定版本号。

---

## 目录

1. [定义与评估维度](#1-定义与评估维度)
2. [三 Agent 主 wire 要求](#2-三-agent-主-wire-要求)
3. [兼容性总矩阵](#3-兼容性总矩阵)
4. [Agent 官方集成补充](#4-agent-官方集成补充)
5. [上游补充说明](#5-上游补充说明)
6. [关系图](#6-关系图)
7. [附录：L2 连通性示例](#7-附录l2-连通性示例)

---

## 1. 定义与评估维度

### 1.1 E2E 原生兼容

同时满足下列三项，记为 **●（E2E 原生）**：

| 条件 | 含义 |
|------|------|
| **Agent 官方集成** | Agent 官方文档中的 provider、`config.toml` 块或环境变量（非自建 HTTP 代理） |
| **上游原生协议面** | 该集成所对接的上游 **官方** Base URL、端点与鉴权 |
| **E2E 能力闭环** | 除主推理 HTTP 200 外，Agent 默认工作流所需的 **流式、tool 多轮、鉴权链** 可成立 |

```text
E2E 原生兼容  =  Agent 官方集成  ×  上游官方协议面  ×  Agent 工作流能力闭环
```

### 1.2 图例与评估分层

| 符号 | 含义 |
|------|------|
| **●** | E2E 原生：官方集成 + 主 wire 对齐（L1+L2） |
| **◐** | 半原生：主 wire 协议族一致，路径 / 模型 ID / 鉴权 / 区域有差异 |
| **—** | 无官方 E2E 路径（协议族不同或无官方集成） |

| 层级 | 检查内容 | 通过含义 |
|------|----------|----------|
| L1 | Agent 是否有官方集成指向上游 | 选型可行 |
| L2 | 主 wire 端点 HTTP 可达 | 最小连通 |
| L3 | 流式（SSE / WebSocket / eventstream） | 交互不挂起 |
| L4 | tool / function 多轮回传 | Agent 能改代码、跑命令 |
| L5 | reasoning、thinking、vision 等 | 高级特性可用 |

**§3 矩阵仅覆盖 L1+L2。** L3–L5 因 Agent 版本、区域、模型而异，以 [reports/](./reports/) 为准。

---

## 2. 三 Agent 主 wire 要求

| Agent | 主推理协议 | 主端点 | L3–L4 典型依赖 |
|-------|------------|--------|----------------|
| **Claude Code** | Anthropic Messages | `POST /v1/messages` | SSE；`tool_use` / `tool_result`；`anthropic-version` |
| **Codex**（[§评估标的](#评估标的版本矩阵所依)：`0.133.0` 起） | OpenAI Responses | `POST /v1/responses` | SSE 或 WebSocket Responses；agentic tools |
| **OpenCode** | OpenAI Chat Completions | `POST /v1/chat/completions` | SSE；`tool_calls` 流式 |

三 Agent **各绑一种协议族**，不会跨族自动切换。Codex 自 **0.133.0** 起仅 `wire_api = "responses"`，Chat Completions 已移除。

---

## 3. 兼容性总矩阵

| 上游 | Claude Code | Codex | OpenCode | 官方集成要点 |
|------|:-----------:|:-----:|:--------:|--------------|
| **Anthropic API** | ● | — | — | 默认 · `/v1/messages` |
| **OpenAI API** | — | ● | ● | Codex：`openai`；OpenCode：Chat provider |
| **Azure OpenAI** | — | ● | ● | `model` = deployment 名；Codex 需 `wire_api = "responses"` |
| **Microsoft Foundry** | ● | — | — | `CLAUDE_CODE_USE_FOUNDRY=1` |
| **AWS Bedrock（runtime）** | ● | ● | — | `CLAUDE_CODE_USE_BEDROCK=1` · `amazon-bedrock` |
| **Bedrock mantle** | ● | ◐ | ◐ | `CLAUDE_CODE_USE_MANTLE=1`；Codex 无内置 mantle provider |
| **Vertex AI（Claude）** | ● | — | — | `CLAUDE_CODE_USE_VERTEX=1` |
| **Claude Platform on AWS** | ● | — | — | `CLAUDE_CODE_USE_ANTHROPIC_AWS=1` |
| **Ollama / LM Studio** | — | ● | — | `codex --oss` |
| **OpenAI 数据驻留** | — | ◐ | — | `[model_providers.openaidr]` |
| **DeepSeek / Moonshot 等** | — | — | ◐ | 原厂 OpenAI Chat；Codex 无 Responses |
| **Google Gemini** | — | — | — | `generateContent`；三 Agent 无官方集成 |

**常见结论**：Claude Code ↔ Anthropic / Bedrock / Vertex / Foundry / Claude AWS → **●**；Codex ↔ OpenAI / Azure Responses / Bedrock / Ollama → **●**；Codex ↔ 仅 Chat 的上游（含多数国产 API）→ **—**；任 Agent ↔ Gemini 原生 API → **—**。

---

## 4. Agent 官方集成补充

矩阵见 [§3](#3-兼容性总矩阵)。本节仅列 **配置入口与官方文档**。

### Claude Code

| 集成 | 配置 |
|------|------|
| Anthropic API | 默认 · `ANTHROPIC_API_KEY` 或 OAuth |
| Amazon Bedrock | `CLAUDE_CODE_USE_BEDROCK=1` + `AWS_REGION` |
| Bedrock Mantle | `CLAUDE_CODE_USE_MANTLE=1` |
| Vertex AI | `CLAUDE_CODE_USE_VERTEX=1` + `ANTHROPIC_VERTEX_PROJECT_ID` + `CLOUD_ML_REGION` |
| Microsoft Foundry | `CLAUDE_CODE_USE_FOUNDRY=1` |
| Claude Platform on AWS | `CLAUDE_CODE_USE_ANTHROPIC_AWS=1` + workspace |

文档：[Environment variables](https://code.claude.com/docs/en/env-vars) · [Bedrock](https://code.claude.com/docs/en/amazon-bedrock) · [Vertex](https://code.claude.com/docs/en/google-vertex-ai)

### Codex

| 集成 | 配置 |
|------|------|
| OpenAI | `model_provider = "openai"`（内置） |
| Amazon Bedrock | `model_provider = "amazon-bedrock"` + `[model_providers.amazon-bedrock.aws]` |
| Azure OpenAI | `[model_providers.azure]` · `wire_api = "responses"` · `api-version` |
| Ollama / LM Studio | `codex --oss` |
| 数据驻留 / mantle | 自定义 `[model_providers.*]` |

文档：[Advanced Configuration](https://developers.openai.com/codex/config-advanced) · [Codex on Azure](https://learn.microsoft.com/en-us/azure/foundry/openai/how-to/codex)

### OpenCode

通过 **OpenAI 兼容 provider 配置** 对接上游，无与 Claude Code / Codex 同级的云厂商 env 开关。

- **●**：OpenAI、Azure Chat 等原厂 Chat 面，字段满足 L3–L4 时记 ●。
- **◐**：DeepSeek、Moonshot、Bedrock mantle Chat 等 **原厂 OpenAI Chat 兼容**；能否 E2E 取决于字段完整度，非 OpenCode 单独保证。

文档：[OpenCode Providers](https://opencode.ai/docs/providers)

---

## 5. 上游补充说明

### Amazon Bedrock

| Runtime | 说明 |
|---------|------|
| **runtime** | Claude Code / Codex 经 **官方 SDK** 对接（Converse / Invoke 等），非裸 `/v1/*` HTTP 路径；与 mantle 是不同集成面。 |
| **mantle** | 暴露 `/anthropic/v1/messages`、`/openai/v1/chat/completions`、`/openai/v1/responses`；Claude Code 有 `USE_MANTLE`；Codex 需自定义 provider。 |

模型 ID 示例：`anthropic.claude-3-5-sonnet-20240620-v1:0` · `anthropic.claude-sonnet-4-6-v1`  
文档：[Bedrock 端点](https://docs.aws.amazon.com/bedrock/latest/userguide/endpoints.html)

### Microsoft Azure

- Azure OpenAI：`.../deployments/{deployment}/responses`（Codex）、`.../chat/completions`（OpenCode）。
- Foundry：Anthropic 面供 Claude Code（`CLAUDE_CODE_USE_FOUNDRY=1`）。
- 请求中 **`model` 为 deployment 名**，非 OpenAI 裸模型 ID。

### Google

- **Vertex AI（Claude）**：Claude Code ●。
- **Gemini（AI Studio / Vertex）**：协议族为 `generateContent`，三 Agent **—**。

### 国产 OpenAI Chat 兼容原厂

DeepSeek、Moonshot、智谱、通义、豆包等：OpenCode **◐**（Chat）；Codex **—**（无 Responses）；Claude Code **—**（非 Messages）。

---

## 6. 关系图

### 6.1 Agent → 主 wire → 上游

```mermaid
flowchart TB
  OC["OpenCode"] --> OCW["Chat Completions"]
  OCW --> OCU["● OpenAI · Azure Chat<br/>◐ Bedrock mantle Chat · 国产 Chat 兼容"]

  CX["Codex"] --> CXW["Responses"]
  CXW --> CXU["● OpenAI · Azure · amazon-bedrock · Ollama<br/>◐ mantle · 数据驻留"]

  CL["Claude Code"] --> CLW["Messages"]
  CLW --> CLU["● Anthropic · Bedrock · Mantle · Vertex · Foundry · Claude AWS"]
```

### 6.2 官方集成一览（三列，无跨列连线）

```mermaid
flowchart LR
  subgraph COL_CL["Claude Code"]
    direction TB
    cl1["● Anthropic API"]
    cl2["● AWS Bedrock / Mantle"]
    cl3["● Vertex · Foundry · Claude AWS"]
  end

  subgraph COL_CX["Codex"]
    direction TB
    cx1["● OpenAI · Bedrock · Azure"]
    cx2["● Ollama / LM Studio"]
    cx3["◐ 数据驻留 · mantle"]
  end

  subgraph COL_OC["OpenCode"]
    direction TB
    oc1["● OpenAI · Azure Chat"]
    oc2["◐ Chat 兼容原厂 · mantle Chat"]
  end
```

### 6.3 无官方 E2E 集成

```mermaid
flowchart TB
  NA["无官方 E2E 集成"]
  NA --> G["Gemini · generateContent"]
  NA --> C["Cohere 自有 API"]
  NA --> X["Codex × 仅 Chat 上游"]
  NA --> Y["Claude Code × OpenAI / Chat 系"]
```

---

## 7. 附录：L2 连通性示例

用于 **L2 主 wire** 核对；**不**代表 L3–L5 E2E 通过。

### 鉴权对照

| 上游 | 典型鉴权 |
|------|----------|
| OpenAI | `Authorization: Bearer` |
| Anthropic / mantle Messages | `x-api-key` + `anthropic-version: 2023-06-01` |
| Azure OpenAI | `api-key` Header |
| Bedrock mantle | Bedrock API Key（`x-api-key` 或 Bearer） |

### curl 示例

**OpenAI Responses（Codex）**

```bash
curl https://api.openai.com/v1/responses \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-5","input":"OK","max_output_tokens":16}'
```

**OpenAI Chat（OpenCode）**

```bash
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"OK"}],"max_tokens":16}'
```

**Anthropic Messages（Claude Code）**

```bash
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4-20250514","max_tokens":64,
       "messages":[{"role":"user","content":"OK"}]}'
```

**Azure Responses（Codex）**

```bash
curl "https://{resource}.openai.azure.com/openai/deployments/{deployment}/responses?api-version=2025-04-01-preview" \
  -H "api-key: $AZURE_OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"{deployment}","input":"OK"}'
```

**Bedrock mantle Messages（Claude Code）**

```bash
curl "https://bedrock-mantle.us-east-1.api.aws/anthropic/v1/messages" \
  -H "x-api-key: $BEDROCK_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{"model":"anthropic.claude-sonnet-4-6-v1","max_tokens":64,
       "messages":[{"role":"user","content":"OK"}]}'
```

---

[← 返回项目总览](../README.md)
