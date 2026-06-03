# Codex 兼容性评估报告

| 项目 | 内容 |
|------|------|
| **评估对象** | OpenAI Codex CLI / Codex.app |
| **目标 API** | `https://api.b.ai/v1` |
| **评估环境** | macOS (darwin 25.0.0, aarch64) |
| **评估日期** | 2026-06-01 |
| **评估结论** | **不兼容** — API 可用，但协议层与 Codex 要求不匹配 |

---

## 1. 执行摘要

本次评估针对本机 Codex 安装状态及第三方 API 提供商 **b.ai**（`https://api.b.ai/v1`）的兼容性进行验证。

**核心结论：**

- Codex 客户端已正确安装，本地配置与状态数据库健康。
- b.ai API Key 有效，`/v1/models` 与 `/v1/chat/completions` 接口工作正常。
- **Codex 无法直接使用 b.ai API**：Codex 0.133+ 强制依赖 OpenAI **Responses API**（`/v1/responses`），而 b.ai 仅开放 Chat Completions、Anthropic Messages 及 Models 三类端点。
- 在未引入 Responses 桥接层或更换 API 提供商前，Codex 无法完成推理请求。

**综合兼容性评级：🔴 不兼容（Blocked）**

---

## 2. 评估范围

| 维度 | 是否覆盖 |
|------|----------|
| Codex 客户端安装与版本 | ✅ |
| 本地配置（`~/.codex/config.toml`） | ✅ |
| 认证状态 | ✅ |
| CLI 可用性与依赖项 | ✅ |
| API 连通性与鉴权 | ✅ |
| API 协议兼容性（Responses vs Chat） | ✅ |
| 端到端 Codex 推理 | ⚠️ 因协议不兼容未通过 |
| ripgrep 代码搜索 | ✅（补充验证，2026-06-01） |
| 插件 / MCP / Computer Use | ℹ️ 未深入测试（非阻塞项） |

---

## 3. 本地 Codex 环境评估

### 3.1 安装信息

| 项目 | 结果 | 状态 |
|------|------|------|
| 桌面应用 | `/Applications/Codex.app` | ✅ 已安装 |
| CLI 可执行文件 | `/Applications/Codex.app/Contents/Resources/codex` | ✅ 存在 |
| CLI 版本 | `0.133.0` | ⚠️ 有更新（最新 `0.135.0`） |
| 安装方式 | Codex.app 内置 | ✅ |
| PATH 集成 | `codex` 不在系统 PATH | ⚠️ 需手动调用或创建 symlink |
| CODEX_HOME | `~/.codex` | ✅ |

### 3.2 认证状态

| 项目 | 结果 | 状态 |
|------|------|------|
| `codex login status` | Not logged in | ❌ 未登录 |
| `~/.codex/auth.json` | 不存在 | ❌ 无持久化凭据 |
| 环境变量认证 | 未配置 | ❌ |
| `config.toml` 中的 `openai_base_url` | 未配置 | — |

> 注：设置 `OPENAI_API_KEY` 环境变量后，`codex doctor` 可识别为有效认证来源，但仍无法解决 API 协议不兼容问题。

### 3.3 本地配置

当前 `~/.codex/config.toml` 主要包含：

- 插件市场与 MCP 服务器（`node_repl`）配置
- Browser / Documents / Spreadsheets / Presentations 插件已启用
- **未配置** `openai_base_url`、`model_provider` 或自定义 `[model_providers.*]`

配置文件语法解析正常，SQLite 状态库（`state_5.sqlite`、`logs_2.sqlite`、`goals_1.sqlite`）完整性检查通过。

### 3.4 依赖项

| 依赖 | 状态 | 影响 |
|------|------|------|
| **ripgrep (`rg`)** | ✅ 已安装 `15.1.0`（`/opt/homebrew/bin/rg`） | 代码搜索依赖已满足 |
| WebSocket 连接 | 对 b.ai 端点失败 | ⚠️ 预期行为（b.ai 不支持 Responses WS） |
| 终端类型 | 非交互环境 `TERM=dumb` | ℹ️ 仅影响 CI/脚本场景 |

> **补充（2026-06-01）**：初评时 `rg` 未安装，`codex doctor` 报 search 警告；经 `brew install ripgrep` 后复测通过。

### 3.5 `codex doctor` 摘要

**初评（无 ripgrep）：**

```
9 ok · 1 idle · 5 notes · 2 warn · 2 fail
```

**复测（ripgrep 15.1.0 已安装）：**

```
10 ok · 1 idle · 4 notes · 1 warn · 2 fail
```

主要失败项（两次一致）：

1. **auth** — 无 Codex 凭据
2. **terminal** — 非真实终端（脚本环境特有）

主要警告项：

| 警告 | 初评 | 复测 |
|------|------|------|
| **search** — 缺少 ripgrep | ⚠️ | ✅ 已消除 |
| **websocket** — Responses WebSocket 403 | ⚠️ | ⚠️ 仍存在（b.ai 不支持） |

复测 search 检查详情：

```
✓ search  ripgrep 15.1.0 (system, `rg`)
    search command           rg
    search provider          system
    search command readiness ripgrep 15.1.0
```

---

## 4. 目标 API（b.ai）评估

### 4.1 基本信息

| 项目 | 值 |
|------|-----|
| Base URL | `https://api.b.ai/v1` |
| 认证方式 | Bearer Token（`Authorization: Bearer sk-...`） |
| API 风格 | OpenAI 兼容 + Anthropic Messages |

### 4.2 端点兼容性矩阵

| 端点 | Codex 是否需要 | b.ai 是否支持 | 测试结果 |
|------|----------------|---------------|----------|
| `GET /v1/models` | 可选 | ✅ | HTTP 200，返回 27+ 模型 |
| `POST /v1/chat/completions` | ❌（Codex 0.133+ 已弃用） | ✅ | HTTP 200，`gpt-5-mini` 正常返回 |
| `POST /v1/messages` | ❌ | ✅（Anthropic 格式） | HTTP 403（部分模型需充值解锁） |
| `POST /v1/responses` | **✅ 必需** | ❌ | HTTP 403，明确拒绝 |
| `WSS /v1/responses` | 可选（流式） | ❌ | HTTP 403 |

### 4.3 b.ai 拒绝 Responses API 的响应

```json
{
  "message": "HTTP node only allows access to inference API paths (/v1/chat/completions, /v1/messages, /v1/models)",
  "success": false
}
```

### 4.4 可用模型（节选）

b.ai `/v1/models` 返回的部分模型：

| 厂商 | 模型 ID |
|------|---------|
| OpenAI | `gpt-5.4`, `gpt-5.2`, `gpt-5-mini`, `gpt-5-nano`, `gpt-5.5` 等 |
| Anthropic | `claude-opus-4.6`, `claude-sonnet-4.6`, `claude-haiku-4.5` 等 |
| Google | `gemini-3.1-pro`, `gemini-3-flash` 等 |
| 其他 | `deepseek-v3.2`, `kimi-k2.5`, `glm-5`, `minimax-m3` 等 |

> Chat Completions 实测：`gpt-5-mini` 在 `max_tokens=50` 下正常返回 `"OK"`，API Key 鉴权有效。

---

## 5. 协议层兼容性分析

### 5.1 Codex API 演进背景

自 2025 年底起，OpenAI Codex 逐步弃用 Chat Completions API，转向 **Responses API**：

| 时间节点 | 变化 |
|----------|------|
| 2025-12 | 官方宣布弃用 `wire_api = "chat"` |
| 2026-02 | 完全移除 Chat Completions 支持 |
| 当前（0.133.0） | 配置 `wire_api = "chat"` 将直接报错 |

Codex 0.133.0 实测报错：

```
`wire_api = "chat"` is no longer supported.
How to fix: set `wire_api = "responses"` in your provider config.
```

### 5.2 协议差异对比

| 特性 | Chat Completions | Responses API |
|------|------------------|---------------|
| 请求端点 | `/v1/chat/completions` | `/v1/responses` |
| 数据结构 | 消息列表（message-centric） | 项目列表（item-centric） |
| 工具调用 | 基础 function calling | 完整 agentic 工具链 |
| 多轮状态 | 客户端维护 | 支持 `previous_response_id` |
| Codex 支持 | ❌ 已移除 | ✅ 唯一支持 |

### 5.3 不兼容根因

```
┌─────────────┐     Responses API      ┌─────────────┐
│   Codex     │ ────── /v1/responses ──▶│   b.ai API  │
│  CLI 0.133  │                         │             │
│             │ ◀──── HTTP 403 ────────│  仅支持:     │
└─────────────┘                         │  chat       │
                                        │  messages   │
                                        │  models     │
                                        └─────────────┘
```

**根因**：b.ai 作为 OpenAI 兼容代理，仅实现了 Chat Completions 层，未实现 Codex 所需的 Responses API 层。这不是配置问题，而是**协议能力缺失**。

---

## 6. 端到端测试记录

### 6.1 curl 直连测试

| 测试用例 | 命令/方法 | 结果 |
|----------|-----------|------|
| 模型列表 | `GET /v1/models` | ✅ 通过 |
| Chat 推理 | `POST /v1/chat/completions` + `gpt-5-mini` | ✅ 通过 |
| Responses 推理 | `POST /v1/responses` + `gpt-5-mini` | ❌ HTTP 403 |
| 错误模型 | `POST /v1/chat/completions` + `gpt-4o-mini` | ❌ model_not_found |

### 6.2 Codex CLI 测试

| 测试用例 | 配置 | 结果 |
|----------|------|------|
| `codex doctor`（无认证） | 默认 | ❌ auth 失败 |
| `codex doctor`（有 API Key） | `OPENAI_API_KEY` 环境变量 | ✅ auth 通过 |
| `codex doctor`（b.ai base URL） | `openai_base_url = "https://api.b.ai/v1"` | ❌ WebSocket/Responses 403 |
| `codex exec`（b.ai + chat wire） | `wire_api = "chat"` | ❌ 版本不支持 chat |
| `codex exec`（b.ai + responses） | `openai_base_url` 指向 b.ai | ❌ 超时/无法完成推理 |

### 6.3 ripgrep 补充验证（2026-06-01）

初评因缺少 `rg`，Codex 代码搜索能力无法验证。安装 ripgrep 后补充测试如下：

| 测试用例 | 命令/方法 | 结果 |
|----------|-----------|------|
| ripgrep 安装 | `brew install ripgrep` | ✅ `15.1.0` → `/opt/homebrew/bin/rg` |
| 系统 PATH | `which rg` | ✅ 可解析 |
| Codex 搜索依赖 | `codex doctor` → search 项 | ✅ `ripgrep 15.1.0 (system, rg)` |
| 直接文本搜索 | `rg "兼容性评估" ~/playcode/mymacos` | ✅ 命中 2 个文件 |

**说明**：ripgrep 修复的是 **Codex 本地代码搜索** 能力，**不改变** b.ai API 协议不兼容的结论。待 API 问题解决后，搜索功能即可正常配合 Agent 使用。

---

## 7. 兼容性评级

| 评估维度 | 评级 | 说明 |
|----------|------|------|
| **API 鉴权** | 🟢 兼容 | Key 有效，Bearer 认证正常 |
| **API 连通性** | 🟢 兼容 | DNS、TLS、HTTP 均正常 |
| **Chat Completions** | 🟢 兼容 | 可用于 curl / 其他客户端 |
| **Responses API** | 🔴 不兼容 | b.ai 明确不支持 |
| **Codex 推理** | 🔴 不兼容 | 无法完成 agent 任务 |
| **Codex 本地安装** | 🟡 部分就绪 | 已安装；ripgrep 已补齐；仍缺 API 认证 |
| **ripgrep 代码搜索** | 🟢 就绪 | `codex doctor` search 检查通过 |
| **综合结论** | 🔴 **不兼容** | 需架构变更方可使用 |

---

## 8. 风险与影响

| 风险 | 级别 | 影响 |
|------|------|------|
| Codex 完全无法调用 b.ai 模型 | **高** | 当前配置下无法使用 Codex 进行代码代理 |
| API Key 已在对话中暴露 | **高** | 建议立即轮换 Key |
| Codex 版本滞后（0.133 vs 0.135） | 低 | 不影响本次兼容性结论 |
| ~~缺少 ripgrep~~ | ~~中~~ → **已解决** | 2026-06-01 安装后 `codex doctor` search ✅ |
| b.ai 部分模型需充值 | 中 | 影响特定模型可用性，与 Codex 兼容性无关 |

---

## 9. 建议方案

### 方案 A：更换 API 提供商（推荐）

选择原生支持 OpenAI Responses API 的提供商：

- OpenAI 官方 API
- 支持 Responses 的企业 LLM 网关 / LiteLLM（需验证 `/v1/responses` 路由）

配置示例：

```toml
# ~/.codex/config.toml
openai_base_url = "https://your-provider.com/v1"
```

```bash
echo "$OPENAI_API_KEY" | codex login --with-api-key
```

### 方案 B：部署 Responses → Chat 桥接层

在本地或内网部署中间层，将 Codex 的 Responses 请求转换为 Chat Completions：

```
Codex ──▶ 本地 Bridge (/v1/responses) ──▶ b.ai (/v1/chat/completions)
```

参考社区方案：

- [Codex Discussion #7782](https://github.com/openai/codex/discussions/7782) — 弃用说明与迁移指南
- VibeAround API Bridge（`va-ai-api-bridge`）— Responses ↔ Chat 双向转换

**注意事项**：桥接层需处理工具调用、流式 SSE、多轮上下文重建，实现复杂度较高。

### 方案 C：向 b.ai 反馈功能需求

请求 b.ai 增加 `/v1/responses` 端点支持。这是从根本上解决兼容性的路径，但取决于供应商 roadmap。

### 方案 D：本地环境修复（与 API 无关，建议同步执行）

```bash
# 1. 将 codex 加入 PATH
sudo ln -sf /Applications/Codex.app/Contents/Resources/codex /usr/local/bin/codex

# 2. 安装 ripgrep（✅ 已于 2026-06-01 完成，rg 15.1.0）
brew install ripgrep

# 3. 更新 Codex
codex update

# 4. 配置认证（待 API 兼容后）
echo "$OPENAI_API_KEY" | codex login --with-api-key
```

---

## 10. 决策建议

| 场景 | 建议 |
|------|------|
| 必须使用 Codex + b.ai | 采用**方案 B**（桥接层），评估维护成本 |
| 可更换 API 提供商 | 采用**方案 A**，选择支持 Responses 的服务 |
| 仅需 API 调用（非 Codex） | b.ai 可直接用于 Chat Completions 客户端 |
| 短期无法解决 | 继续使用 Cursor / 其他支持 Chat API 的工具 |

**当前不建议**在 b.ai 上投入 Codex 集成工作，除非确认 b.ai 将支持 Responses API 或团队有能力维护桥接层。

---

## 11. 附录

### A. 测试环境

```
OS:              macOS 25.0.0 (aarch64)
Codex CLI:       0.133.0
Codex App:       /Applications/Codex.app
CODEX_HOME:      ~/.codex
API Endpoint:    https://api.b.ai/v1
测试日期:        2026-06-01
```

### B. 参考文档

- [Codex Advanced Configuration](https://developers.openai.com/codex/config-advanced)
- [Codex Configuration Reference](https://developers.openai.com/codex/config-reference)
- [Deprecating chat/completions support (#7782)](https://github.com/openai/codex/discussions/7782)
- [openai_base_url config override (#12031)](https://github.com/openai/codex/pull/12031)

### C. 术语表

| 术语 | 说明 |
|------|------|
| Responses API | OpenAI 新一代推理接口，Codex 0.133+ 唯一支持的 wire 协议 |
| Chat Completions | OpenAI 传统对话接口，Codex 已弃用 |
| wire_api | Codex 配置项，指定与模型提供商通信的协议类型 |
| model_provider | Codex 配置项，指定使用的模型提供商 |
| ripgrep (`rg`) | 高性能文本搜索工具，Codex 代码搜索依赖 |

---

*报告由兼容性评估脚本自动生成，基于 2026-06-01 实测数据；ripgrep 补充验证于同日完成。*
