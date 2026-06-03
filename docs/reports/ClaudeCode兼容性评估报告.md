# Claude Code 兼容性评估报告

| 项目 | 内容 |
|------|------|
| **评估对象** | Anthropic Claude Code CLI |
| **目标 API** | `https://api.b.ai/v1` |
| **评估环境** | macOS (darwin 25.0.0, aarch64) |
| **评估日期** | 2026-06-01 |
| **评估结论** | **基本兼容** — 协议层匹配，Haiku 模型端到端验证通过；高级模型受账户权限限制 |

---

## 1. 执行摘要

本次评估针对本机 Claude Code 安装状态及第三方 API 提供商 **b.ai**（`https://api.b.ai/v1`）的兼容性进行验证，评估方法与《Codex 兼容性评估报告》保持一致。

**核心结论：**

- Claude Code 已正确安装（评估期间由 `2.1.23` 更新至 `2.1.159`），CLI 在 PATH 中可用。
- b.ai API Key 有效，**原生支持 Anthropic Messages API**（`/v1/messages`），与 Claude Code 协议要求一致。
- **端到端验证通过**：配置 `ANTHROPIC_BASE_URL` 后，`claude-haiku-4.5` 成功返回 `API OK`。
- **工具调用完整验证通过**：`tool_use` 下发、`tool_result` 回传、多轮 Agent 循环、Bash/Read 内置工具及流式 `input_json_delta` 均正常。
- **部分模型不可用**：`claude-sonnet-4.6` 等高级模型返回 403，需账户充值解锁（非协议问题）。
- 与 Codex 不同，Claude Code **无需桥接层**即可对接 b.ai。

**综合兼容性评级：🟢 基本兼容（Compatible with Caveats）**

---

## 2. 评估范围

| 维度 | 是否覆盖 |
|------|----------|
| Claude Code 客户端安装与版本 | ✅ |
| 本地配置（`~/.claude.json` / settings） | ✅ |
| 认证状态 | ✅ |
| CLI 可用性与依赖项 | ✅ |
| API 连通性与鉴权 | ✅ |
| API 协议兼容性（Messages API） | ✅ |
| 流式 SSE 与工具参数 | ✅ |
| 端到端 Claude Code 推理 | ✅（Haiku） |
| 工具调用（Agent 模式） | ✅ 完整验证（Haiku） |
| ripgrep / Grep 工具 | ✅（补充验证，2026-06-01） |
| 插件 / MCP | ℹ️ 未深入测试（非阻塞项） |

---

## 3. 本地 Claude Code 环境评估

### 3.1 安装信息

| 项目 | 结果 | 状态 |
|------|------|------|
| CLI 可执行文件 | `~/.local/bin/claude` → `~/.local/share/claude/versions/2.1.159` | ✅ 存在 |
| CLI 版本（评估前） | `2.1.23` | ⚠️ 已过时 |
| CLI 版本（评估后） | `2.1.159` | ✅ 已更新 |
| 安装方式 | native（`installMethod: native`） | ✅ |
| PATH 集成 | `~/.local/bin/claude` 在 PATH 中 | ✅ |
| 桌面应用 | `/Applications/Claude.app` | ✅ 已安装（独立应用） |
| 数据目录 | `~/.claude/` | ✅ |

### 3.2 认证状态

| 项目 | 结果 | 状态 |
|------|------|------|
| Claude 订阅登录 | Not logged in | ❌ 未登录 |
| `~/.claude/settings.json` | 不存在 | — |
| API Key 环境变量 | 未持久化配置 | ⚠️ 需手动设置 |
| 第三方 API 认证 | 通过 `ANTHROPIC_AUTH_TOKEN` / `ANTHROPIC_API_KEY` | ✅ 实测有效 |

> 注：Claude Code 支持两种认证路径——Claude 订阅 OAuth（`/login`）或 API Key（`ANTHROPIC_AUTH_TOKEN`）。使用 b.ai 时走 API Key 路径即可，无需 Claude 订阅。

### 3.3 本地配置

当前 `~/.claude.json` 主要包含：

- 安装方式、功能开关缓存（GrowthBook features）
- 用户 ID、迁移标记
- **未配置** `env.ANTHROPIC_BASE_URL` 或 API Key

无项目级 `.claude/settings.json`。评估期间通过环境变量临时注入 API 配置。

### 3.4 依赖项

| 依赖 | 状态 | 影响 |
|------|------|------|
| **ripgrep (`rg`)** | ✅ 已安装 `15.1.0`（`/opt/homebrew/bin/rg`） | Grep 工具与内置搜索正常 |
| Node.js/Bun 运行时 | 内置于 native 二进制 | ✅ 无需外部 Node |
| 网络连通性 | b.ai HTTPS 正常 | ✅ |

> **补充（2026-06-01）**：初评时 `rg` 未安装，Claude Code 启动自检与 Grep 工具报错；经 `brew install ripgrep` 后复测通过。

### 3.5 启动健康检查

| 检查项 | 结果 |
|--------|------|
| `claude --version` | ✅ 正常 |
| `claude doctor` | ⚠️ 非交互终端报错（Ink raw mode），不影响 CLI 功能 |
| `claude login status` | ❌ Not logged in（预期，使用 API Key 模式） |
| 旧版 2.1.23 非交互启动 | ⚠️ 启动挂起 ~60s（claudeai-mcp 相关） |
| 新版 2.1.159 非交互启动 | ✅ ~3s 内就绪 |
| **ripgrep 自检（补充）** | ✅ `Ripgrep first use test: PASSED` |

---

## 4. 目标 API（b.ai）评估

### 4.1 基本信息

| 项目 | 值 |
|------|-----|
| Base URL | `https://api.b.ai/v1` |
| Claude Code 配置值 | `ANTHROPIC_BASE_URL=https://api.b.ai`（SDK 自动追加 `/v1/messages`） |
| 认证方式 | Bearer Token（`Authorization: Bearer sk-...`） |
| API 风格 | Anthropic Messages API 兼容 |

### 4.2 端点兼容性矩阵

| 端点 | Claude Code 是否需要 | b.ai 是否支持 | 测试结果 |
|------|---------------------|---------------|----------|
| `GET /v1/models` | 可选（Gateway 模型发现） | ✅ | HTTP 200 |
| `POST /v1/messages` | **✅ 必需** | ✅ | HTTP 200（Haiku） |
| `POST /v1/messages`（stream） | ✅ 流式推理 | ✅ | SSE 事件流正常 |
| `POST /v1/messages`（tools） | ✅ Agent 工具调用 | ✅ | 工具参数接受，流式返回正常 |
| `POST /v1/chat/completions` | ❌ 不使用 | ✅ | Claude Code 不调用 |
| `POST /v1/responses` | ❌ 不使用 | ❌ | Claude Code 不调用 |

### 4.3 Claude 模型可用性（b.ai 账户实测）

| 模型 ID | curl 测试 | Claude Code 测试 | 说明 |
|---------|-----------|-----------------|------|
| `claude-haiku-4.5` | ✅ HTTP 200，返回 "OK" | ✅ 返回 "API OK" | **完全可用** |
| `claude-sonnet-4.6` | ❌ HTTP 403 | ❌ HTTP 403 | 需充值解锁 |
| `claude-sonnet-4.5` | ❌ HTTP 403 | — | 需充值解锁 |
| `claude-opus-4.5` | ❌ HTTP 403 | — | 需充值解锁 |
| `claude-opus-4.6` | 未测 | — | 预计同类限制 |

b.ai 403 响应示例：

```json
{
  "error": {
    "code": "access_denied",
    "message": "Access restricted. Deposit required to unlock premium models."
  }
}
```

> 此为 **账户权限限制**，非协议不兼容。充值后 Sonnet/Opus 模型预计可正常使用。

### 4.4 流式与工具支持验证

**流式 SSE（纯文本）**（`stream: true`）：

```
event: message_start
event: content_block_start
event: content_block_delta  → "OK"
event: content_block_stop
event: message_stop
```

**流式 SSE（工具调用）**（`stream: true` + `tools`）：

```
event: content_block_start   → type: tool_use, name: calc
event: content_block_delta   → input_json_delta: {"expr":
event: content_block_delta   → input_json_delta:  "6*7"
event: content_block_delta   → input_json_delta: "}"
event: content_block_stop
event: message_delta         → stop_reason: tool_use
```

流式工具参数增量（`input_json_delta`）格式正确，与 Anthropic SDK 预期一致。

### 4.5 工具调用完整验证（2026-06-01 补充）

#### 4.5.1 API 层：单轮 tool_use

| 步骤 | 请求 | 响应 | 结果 |
|------|------|------|------|
| 用户提问 + tools 定义 | `get_weather(city)` 工具 | `stop_reason: tool_use` | ✅ |
| 模型输出 | — | `{"type":"tool_use","name":"get_weather","input":{"city":"Beijing"}}` | ✅ |

#### 4.5.2 API 层：多轮 tool_result 回环

| 轮次 | 消息结构 | 响应 | 结果 |
|------|----------|------|------|
| 第 1 轮 | user → 请求查天气 | assistant 返回 `tool_use` | ✅ |
| 第 2 轮 | user 携带 `tool_result`（"Beijing: 22C, sunny"） | assistant 文本："The weather in Beijing is currently **22°C and sunny**" | ✅ |

完整 Agent 工具循环（`tool_use` → 执行 → `tool_result` → 最终回复）在 API 层验证通过。

#### 4.5.3 Claude Code 层：内置工具 E2E

| 测试 | 命令要点 | 实测输出 | 耗时 | 结果 |
|------|----------|----------|------|------|
| **Bash 工具** | `--tools Bash --permission-mode bypassPermissions` | 模型调用 Bash → `echo TOOL_OK` → 回复 `DONE` | ~10s | ✅ |
| **Read 工具** | `--tools Read` | 正确读取 `ClaudeCode兼容性评估报告.md` 第 1 行标题 | ~10s | ✅ |
| **Grep 工具** | `--tools Grep` | 搜索 `兼容性评估`，返回 **2 files match** | ~14s | ✅ |
| **多工具串联** | `--tools Bash,Read` | Bash 输出 `STEP1_OK` + Read 返回 `# Codex 兼容性评估报告` | ~10s | ✅ |

Claude Code 调试日志摘录（Bash 工具）：

```
[INFO] [Stall] tool_dispatch_start tool=Bash toolUseId=toolu_01NjwWgYr8NL3Vwu4L14Jkzs
[INFO] [Stall] tool_dispatch_end   tool=Bash toolUseId=toolu_01NjwWgYr8NL3Vwu4L14Jkzs outcome=ok durationMs=107
[DEBUG] [API REQUEST] /v1/messages source=sdk   ← 工具结果回传后的第 2 次 API 调用
```

Claude Code 调试日志摘录（Grep 工具，ripgrep 安装后）：

```
[DEBUG] Ripgrep first use test: PASSED (mode=embedded, path=.../2.1.159)
[INFO] [Stall] tool_dispatch_start tool=Grep toolUseId=toolu_01LQbJbFzwKsMMV4T6NdGhXv
[INFO] [Stall] tool_dispatch_end   tool=Grep toolUseId=toolu_01LQbJbFzwKsMMV4T6NdGhXv outcome=ok durationMs=179
```

> 初评时无 `rg`，日志曾出现 `rg error ... No such file or directory`（指向 `~/.claude/plugins/cache` 缺失目录）。安装 ripgrep 后 Grep 工具正常，`outcome=ok`。

#### 4.5.4 工具调用兼容性矩阵

| 能力 | b.ai 支持 | Claude Code + b.ai 实测 |
|------|-----------|-------------------------|
| `tools` 参数传递 | ✅ | ✅ |
| `tool_use` 响应块 | ✅ | ✅ |
| `tool_result` 回传 | ✅ | ✅ |
| 多轮 Agent 循环 | ✅ | ✅ |
| 流式 `input_json_delta` | ✅ | ✅（API 层） |
| Bash 内置工具 | — | ✅ |
| Read 内置工具 | — | ✅ |
| Grep 内置工具（ripgrep） | — | ✅ |
| 多工具顺序调用 | — | ✅ |
| Tool Search（`tool_reference`） | 未测 | ❌ 禁用（Haiku + 第三方 URL） |
| 并行多工具（同轮） | 未测 | ℹ️ 未专项测试 |

#### 4.5.5 已知限制

1. **首次请求 system role 400**：Claude Code 发送 system prompt 时，b.ai 对 Haiku 返回 400，SDK 自动重试后工具调用仍正常完成。
2. **Tool Search 不可用**：Haiku 模型本身不支持 `tool_reference` 块；且第三方 `ANTHROPIC_BASE_URL` 会禁用 Tool Search 乐观模式。不影响标准 function calling。
3. **未测项**：Edit / Write / WebFetch 等工具未逐一验证；**Grep 已于补充测试中验证通过**。

---

## 5. 协议层兼容性分析

### 5.1 Claude Code API 要求

Claude Code 基于 Anthropic SDK，使用 **Messages API** 作为唯一推理协议：

| 特性 | 要求 | b.ai 支持 |
|------|------|-----------|
| 请求端点 | `POST /v1/messages` | ✅ |
| 认证 Header | `Authorization: Bearer` 或 `x-api-key` | ✅ |
| 流式 SSE | `stream: true` | ✅ |
| 工具调用 | `tools` + `tool_use` / `tool_result` | ✅ 完整验证 |
| 流式工具参数 | `input_json_delta` | ✅ |
| 多轮 Agent 循环 | assistant `tool_use` + user `tool_result` | ✅ |
| System Prompt | `system` 参数或 system message | ⚠️ 首次请求报 400，自动重试后成功 |
| Anthropic 版本头 | `anthropic-version` | ✅ SDK 自动添加 |

### 5.2 与 Codex 的关键差异

```
┌──────────────────┐     Messages API       ┌─────────────┐
│  Claude Code     │ ──── /v1/messages ───▶ │   b.ai API  │
│  CLI 2.1.159     │                         │             │
│                  │ ◀──── HTTP 200 ────────│  ✅ 支持     │
└──────────────────┘                         └─────────────┘

┌──────────────────┐     Responses API      ┌─────────────┐
│  Codex CLI       │ ──── /v1/responses ──▶ │   b.ai API  │
│  0.133.0         │                         │             │
│                  │ ◀──── HTTP 403 ────────│  ❌ 不支持   │
└──────────────────┘                         └─────────────┘
```

Claude Code 与 b.ai 在协议层**天然匹配**，无需桥接或格式转换。

### 5.3 首次请求 system role 警告

Claude Code 端到端测试中，首次 API 请求返回：

```json
{
  "error": {
    "message": "role 'system' is not supported on this model"
  }
}
```

SDK 自动重试后请求成功。此为 b.ai 对 Haiku 模型的 system role 限制，**不影响最终可用性**，但可能在日志中产生一次 400 错误记录。

---

## 6. 端到端测试记录

### 6.1 curl 直连测试

| 测试用例 | 方法 | 结果 |
|----------|------|------|
| 模型列表 | `GET /v1/models` | ✅ 通过 |
| Messages 推理 | `POST /v1/messages` + `claude-haiku-4.5` | ✅ 返回 "OK" |
| Messages 流式 | `POST /v1/messages` + `stream: true` | ✅ SSE 正常 |
| Messages 工具（单轮 tool_use） | `POST /v1/messages` + `tools` | ✅ `stop_reason: tool_use` |
| Messages 工具（多轮 tool_result） | 2 轮对话 + `tool_result` | ✅ 最终文本回复正确 |
| Messages 流式工具 | `stream: true` + `tools` | ✅ `input_json_delta` 正常 |
| Sonnet 推理 | `POST /v1/messages` + `claude-sonnet-4.6` | ❌ 403 需充值 |

### 6.2 Claude Code CLI 测试

| 测试用例 | 配置 | 结果 |
|----------|------|------|
| 版本检查 | 默认 | ✅ `2.1.159` |
| 登录状态 | 默认 | ❌ Not logged in（API Key 模式不需要） |
| E2E 推理（Haiku） | `ANTHROPIC_BASE_URL=https://api.b.ai` + API Key | ✅ 返回 `API OK` |
| E2E 推理（Sonnet） | 同上 + `--model claude-sonnet-4.6` | ❌ 403 需充值 |
| 旧版启动（2.1.23） | 非交互 `-p` 模式 | ⚠️ 挂起 ~60s |
| 新版启动（2.1.159） | 非交互 `-p` 模式 | ✅ ~3–9s 完成 |
| E2E 工具：Bash | `--tools Bash --permission-mode bypassPermissions` | ✅ 执行 `echo TOOL_OK`，回复 `DONE` |
| E2E 工具：Read | `--tools Read` | ✅ 正确读取本地文件第 1 行 |
| E2E 工具：Bash + Read 串联 | `--tools Bash,Read` | ✅ 两步工具均成功 |
| E2E 工具：Grep（ripgrep） | `--tools Grep` + b.ai | ✅ 返回 2 files match |
| ripgrep 安装复测 | `brew install ripgrep` + 启动自检 | ✅ `Ripgrep first use test: PASSED` |

**成功的 E2E 命令（纯推理）：**

```bash
export ANTHROPIC_BASE_URL="https://api.b.ai"
export ANTHROPIC_AUTH_TOKEN="sk-xxxxxxxx"
export ANTHROPIC_API_KEY="sk-xxxxxxxx"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1

claude --print --model claude-haiku-4.5 --max-budget-usd 1.00 \
  'Reply with exactly: API OK'
# 输出: API OK
```

**成功的 E2E 命令（工具调用）：**

```bash
export ANTHROPIC_BASE_URL="https://api.b.ai"
export ANTHROPIC_AUTH_TOKEN="sk-xxxxxxxx"
export ANTHROPIC_API_KEY="sk-xxxxxxxx"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1

# Bash 工具
claude --print --model claude-haiku-4.5 --max-budget-usd 2.00 \
  --tools Bash --allowedTools Bash --permission-mode bypassPermissions \
  'Use the Bash tool exactly once to run: echo TOOL_OK. Then reply with exactly: DONE'
# 输出: DONE

# Read 工具
claude --print --model claude-haiku-4.5 --max-budget-usd 2.00 \
  --tools Read --allowedTools Read --permission-mode bypassPermissions \
  'Use the Read tool to read ./ClaudeCode兼容性评估报告.md and tell me the exact title on line 1 only.'
# 输出: # Claude Code 兼容性评估报告

# Grep 工具（需 ripgrep）
claude --print --model claude-haiku-4.5 --max-budget-usd 2.00 \
  --tools Grep --allowedTools Grep --permission-mode bypassPermissions \
  'Use the Grep tool once to search for "兼容性评估" in the current directory. Reply with the count of matching files only.'
# 输出: 2 files match the search for "兼容性评估".
```

### 6.3 ripgrep 补充验证（2026-06-01）

初评因缺少 `rg`，Claude Code 内置搜索与 Grep 工具无法正常工作。安装 ripgrep 后补充测试如下：

| 阶段 | 测试 | 结果 |
|------|------|------|
| **初评（无 rg）** | 启动日志 | ❌ `rg error ... No such file or directory (os error 2)` |
| **安装** | `brew install ripgrep` | ✅ `15.1.0` → `/opt/homebrew/bin/rg` |
| **自检** | Claude Code 启动 | ✅ `Ripgrep first use test: PASSED` |
| **直接搜索** | `rg "兼容性评估" .` | ✅ 命中 2 个 `.md` 文件 |
| **Grep 工具 E2E** | `--tools Grep` + b.ai API | ✅ `2 files match`，`outcome=ok`（179ms） |

**初评 vs 复测对比：**

| 项目 | 初评（无 ripgrep） | 复测（ripgrep 15.1.0） |
|------|-------------------|------------------------|
| `which rg` | 未找到 | `/opt/homebrew/bin/rg` |
| Claude 启动自检 | 报错 / 降级 | `PASSED` |
| Grep 工具 | 未验证 / 预期失败 | ✅ E2E 通过 |
| Codex `doctor` search | ⚠️ 警告 | ✅ 通过（见 Codex 报告 §6.3） |

---

## 7. 功能限制与降级项

使用第三方 API 时，Claude Code 自动禁用或降级以下功能：

| 功能 | 状态 | 原因 |
|------|------|------|
| Tool Search（乐观模式） | ❌ 禁用 | `ANTHROPIC_BASE_URL` 非 Anthropic 官方域名 |
| Claude.ai MCP 连接器 | ❌ 禁用 | API Key 认证优先级高于 OAuth |
| Gateway 模型自动发现 | ❌ 未启用 | 需 `CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1`（v2.1.129+） |
| 内置 Slash Commands（部分） | ⚠️ 可能不可用 | 依赖 Anthropic 原生后端 |
| 高推理 Effort 模式 | ⚠️ 可能不可用 | 依赖 Anthropic 原生后端 |
| ripgrep / Grep 代码搜索 | ✅ 可用 | 2026-06-01 安装 `rg 15.1.0` 后验证通过 |

> Tool Search 可通过设置 `ENABLE_TOOL_SEARCH=true` 手动启用（需确认 b.ai 代理转发 `tool_reference` 块）。

---

## 8. 兼容性评级

| 评估维度 | 评级 | 说明 |
|----------|------|------|
| **API 鉴权** | 🟢 兼容 | Key 有效，Bearer 认证正常 |
| **API 连通性** | 🟢 兼容 | DNS、TLS、HTTP 均正常 |
| **Messages API 协议** | 🟢 兼容 | Claude Code 原生协议，完全匹配 |
| **流式 SSE** | 🟢 兼容 | SSE 事件流格式正确 |
| **工具调用（API 层）** | 🟢 兼容 | tool_use / tool_result / 多轮循环均通过 |
| **工具调用（Claude Code）** | 🟢 兼容 | Bash / Read / Grep / 多工具串联 E2E 通过 |
| **ripgrep / Grep 搜索** | 🟢 兼容 | 安装后自检 PASSED，Grep 工具 E2E 通过 |
| **流式工具参数** | 🟢 兼容 | input_json_delta SSE 格式正确 |
| **Haiku 模型推理** | 🟢 兼容 | 端到端验证通过 |
| **Sonnet/Opus 模型** | 🟡 受限 | 账户需充值，非协议问题 |
| **System Prompt** | 🟡 部分兼容 | 首次 400，SDK 自动重试成功 |
| **Claude Code 本地安装** | 🟢 就绪 | 已安装、已更新、PATH 可用 |
| **综合结论** | 🟢 **基本兼容** | 可用，有模型权限和功能降级限制 |

---

## 9. 风险与影响

| 风险 | 级别 | 影响 |
|------|------|------|
| 高级模型需充值才能使用 | **中** | Sonnet/Opus 不可用，仅 Haiku 可立即使用 |
| API Key 已在对话中暴露 | **高** | 建议立即轮换 Key |
| 旧版 2.1.23 非交互启动挂起 | 中 | 已更新至 2.1.159 解决 |
| ~~缺少 ripgrep~~ | ~~中~~ → **已解决** | 2026-06-01 安装后 Grep 工具与自检 ✅ |
| 第三方 API 功能降级 | 低 | Tool Search、MCP 等受限，核心推理不受影响 |
| b.ai 服务稳定性 | 中 | 依赖第三方 SLA，无 Anthropic 官方保障 |

---

## 10. 建议方案

### 方案 A：直接配置使用（推荐）

在 `~/.claude/settings.json` 中持久化配置：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.b.ai",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-api-key-here",
    "ANTHROPIC_API_KEY": "sk-your-api-key-here",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

或通过 shell profile（`~/.zshrc`）：

```bash
export ANTHROPIC_BASE_URL="https://api.b.ai"
export ANTHROPIC_AUTH_TOKEN="sk-your-api-key-here"
export ANTHROPIC_API_KEY="sk-your-api-key-here"
```

### 方案 B：解锁高级模型

在 b.ai 控制台充值账户，解锁 Sonnet/Opus 模型。充值后使用：

```bash
claude --print --model claude-sonnet-4.6 "your prompt"
```

### 方案 C：本地环境修复（建议同步执行）

```bash
# 1. 确保 Claude Code 为最新版
claude update

# 2. 安装 ripgrep（✅ 已于 2026-06-01 完成，rg 15.1.0）
brew install ripgrep

# 3. 轮换已暴露的 API Key
# 在 b.ai 控制台生成新 Key 并更新配置
```

### 方案 D：启用 Gateway 模型发现（可选）

Claude Code v2.1.129+ 支持从 Gateway 自动发现模型：

```bash
export CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY=1
```

启动后 `/model` 选择器会显示 b.ai 返回的模型列表（标注 "From gateway"）。

---

## 11. 与 Codex 对比

| 对比项 | Codex 0.133.0 | Claude Code 2.1.159 |
|--------|---------------|---------------------|
| 所需 API 协议 | Responses API (`/v1/responses`) | Messages API (`/v1/messages`) |
| b.ai 协议支持 | ❌ 不支持 | ✅ 支持 |
| 端到端可用性 | ❌ 不可用 | ✅ 可用（Haiku） |
| 需要桥接层 | ✅ 必须 | ❌ 不需要 |
| PATH 集成 | ❌ 不在 PATH | ✅ 在 PATH |
| 认证配置 | 未配置 | 未配置（均可通过 env 注入） |
| 综合兼容性 | 🔴 不兼容 | 🟢 基本兼容 |

---

## 12. 决策建议

| 场景 | 建议 |
|------|------|
| 使用 b.ai + Claude Code（轻量任务） | ✅ **立即可用**，配置 `ANTHROPIC_BASE_URL` + Haiku 模型 |
| 使用 b.ai + Claude Code（复杂编码） | ⚠️ 需充值解锁 Sonnet/Opus，或使用 Haiku 接受能力降级 |
| 使用 b.ai + Codex | ❌ 不可行，需换 provider 或部署桥接层 |
| 生产环境部署 | 建议持久化 `settings.json`、轮换 API Key（ripgrep ✅ 已安装） |

**当前建议：Claude Code + b.ai 组合可以投入使用**，优先使用 `claude-haiku-4.5` 进行验证，充值后切换至 Sonnet/Opus 获得更好编码能力。

---

## 13. 附录

### A. 测试环境

```
OS:              macOS 25.0.0 (aarch64)
Claude Code:     2.1.23 → 2.1.159（评估期间更新）
CLI Path:        ~/.local/bin/claude
Config:          ~/.claude.json
API Endpoint:    https://api.b.ai
测试日期:        2026-06-01
```

### B. 推荐配置模板

**~/.claude/settings.json（完整示例）：**

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.b.ai",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-api-key",
    "ANTHROPIC_API_KEY": "sk-your-api-key",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

### C. 参考文档

- [Claude Code LLM Gateway Configuration](https://code.claude.com/docs/en/llm-gateway)
- [Claude Code Third-Party Integrations](https://code.claude.com/docs/en/third-party-integrations)
- [ANTHROPIC_BASE_URL 配置指南](https://fazm.ai/blog/route-claude-api-through-custom-endpoint-anthropic-base-url)
- [Claude Code Environment Variables](https://code.claude.com/docs/en/settings#environment-variables)

### D. 术语表

| 术语 | 说明 |
|------|------|
| Messages API | Anthropic 对话推理接口，Claude Code 唯一使用的 wire 协议 |
| ANTHROPIC_BASE_URL | 覆盖 API 端点的环境变量，指向第三方 Gateway |
| ANTHROPIC_AUTH_TOKEN | API Key，以 Bearer Token 形式发送 |
| Tool Search | Claude Code 内置工具搜索优化，第三方 API 默认禁用 |
| Gateway Model Discovery | 从 `/v1/models` 自动发现可用模型并加入选择器 |

---

*报告基于 2026-06-01 实测数据。评估期间 Claude Code 由 v2.1.23 更新至 v2.1.159；ripgrep 与 Grep 工具补充验证于同日完成。*
