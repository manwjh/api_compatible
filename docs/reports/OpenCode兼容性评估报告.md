# OpenCode 兼容性评估报告

| 项目 | 内容 |
|------|------|
| **评估对象** | OpenCode CLI（开源 AI 编程 Agent） |
| **目标 API** | `https://api.b.ai/v1` |
| **评估环境** | macOS (darwin 25.0.0, aarch64) |
| **评估日期** | 2026-06-01 |
| **评估结论** | **兼容** — b.ai Chat Completions 与 OpenCode 协议匹配，端到端推理通过 |

---

## 1. 执行摘要

本次评估在本地安装 OpenCode，配置 b.ai 作为自定义 OpenAI 兼容 Provider，并完成 API 连通性与端到端 Agent 推理验证。

**核心结论：**

- OpenCode **1.15.13** 已通过 `npm install -g opencode-ai` 安装并可正常运行。
- b.ai API Key 有效，`/v1/models` 与 `/v1/chat/completions` 工作正常。
- **OpenCode 可直接使用 b.ai API**：OpenCode 通过 `@ai-sdk/openai-compatible` 调用 **Chat Completions**（`/v1/chat/completions`），与 b.ai 支持的端点完全一致。
- 与 Codex 不同，OpenCode **不依赖** OpenAI Responses API（`/v1/responses`），因此不受 b.ai 对该端点的 403 限制。
- 端到端测试：`opencode run -m bai/kimi-k2.5` 成功返回 `COMPATIBLE`，Agent 任务可完成。

**综合兼容性评级：🟢 兼容（Ready）**

> 对比参考：同目录下 [Codex兼容性评估报告.md](./Codex兼容性评估报告.md) 结论为 **不兼容**，因 Codex 0.133+ 强制 Responses API。

---

## 2. 评估范围

| 维度 | 是否覆盖 |
|------|----------|
| OpenCode 安装与版本 | ✅ |
| 自定义 Provider 配置（`opencode.json`） | ✅ |
| 认证状态（`auth.json`） | ✅ |
| CLI 可用性 | ✅ |
| API 连通性与鉴权 | ✅ |
| API 协议兼容性（Chat Completions） | ✅ |
| 端到端 OpenCode 推理（`opencode run`） | ✅ 通过 |
| 工具调用 / MCP / LSP | ℹ️ 未深入测试（非阻塞项） |
| 推理型模型（gpt-5-mini 等） | ⚠️ 部分模型需注意 token 分配 |

---

## 3. 本地 OpenCode 环境评估

### 3.1 安装信息

| 项目 | 结果 | 状态 |
|------|------|------|
| 安装方式 | `npm install -g opencode-ai@latest` | ✅ |
| CLI 路径 | `/opt/homebrew/bin/opencode` | ✅ |
| CLI 版本 | `1.15.13` | ✅ 最新稳定版 |
| 源码仓库 | `upstream/opencode/`（`upstream/pull.sh opencode`，供参考） | ✅ |
| 数据目录 | `~/.local/share/opencode` | ✅ |
| 配置目录 | `~/.config/opencode` | ✅ |

### 3.2 认证状态

| 项目 | 结果 | 状态 |
|------|------|------|
| Provider ID | `bai` | ✅ |
| 凭据存储 | `~/.local/share/opencode/auth.json` | ✅ |
| `opencode providers list` | `bai api` 已识别 | ✅ |

### 3.3 项目配置

项目根目录 `opencode.json` 已配置 b.ai 自定义 Provider：

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "bai/kimi-k2.5",
  "provider": {
    "bai": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "b.ai",
      "options": {
        "baseURL": "https://api.b.ai/v1"
      },
      "models": {
        "kimi-k2.5": { "name": "Kimi K2.5" },
        "deepseek-v3.2": { "name": "DeepSeek V3.2" },
        "gpt-5-mini": { "name": "GPT-5 Mini" },
        "claude-sonnet-4.6": { "name": "Claude Sonnet 4.6" }
      }
    }
  }
}
```

`opencode debug config` 解析正常；`opencode models bai` 列出 4 个已配置模型。

### 3.4 依赖项

| 依赖 | 状态 | 影响 |
|------|------|------|
| **ripgrep (`rg`)** | 首次运行自动下载至 `~/.cache/opencode/bin/rg` | ⚠️ 首次启动约 6 分钟（GitHub 下载）；建议 `brew install ripgrep` 提前安装 |
| Node.js | `/opt/homebrew/bin/node` 可用 | ✅ |
| Bun（源码开发） | 未安装 | ℹ️ 不影响 npm 全局安装使用 |

---

## 4. 目标 API（b.ai）评估

### 4.1 基本信息

| 项目 | 值 |
|------|-----|
| Base URL | `https://api.b.ai/v1` |
| 认证方式 | Bearer Token（`Authorization: Bearer sk-...`） |
| API 风格 | OpenAI 兼容 + Anthropic Messages |

### 4.2 端点兼容性矩阵

| 端点 | OpenCode 是否需要 | b.ai 是否支持 | 测试结果 |
|------|-------------------|---------------|----------|
| `GET /v1/models` | 可选（模型发现） | ✅ | HTTP 200，返回 27+ 模型 |
| `POST /v1/chat/completions` | **✅ 必需** | ✅ | HTTP 200，多模型正常 |
| `POST /v1/messages` | ❌（Anthropic 原生路径） | ✅ | 未用于 OpenCode 配置 |
| `POST /v1/responses` | ❌ | ❌ | HTTP 403（OpenCode 不使用） |

### 4.3 OpenCode 与 b.ai 协议匹配

```
┌─────────────┐   Chat Completions    ┌─────────────┐
│  OpenCode   │ ── /v1/chat/completions ──▶│   b.ai API  │
│  CLI 1.15   │                          │             │
│             │ ◀──── HTTP 200 ──────────│  支持:       │
└─────────────┘                          │  chat ✅     │
  @ai-sdk/openai-compatible              │  messages    │
                                         │  models      │
                                         └─────────────┘
```

**根因（与 Codex 对比）**：OpenCode 使用 AI SDK 的 OpenAI Compatible 适配器，走 Chat Completions 协议；b.ai 完整支持该端点，因此**无需桥接层**。

### 4.4 可用模型（节选）

b.ai `/v1/models` 返回的部分模型：

| 厂商 | 模型 ID | OpenCode 实测 |
|------|---------|---------------|
| Moonshot | `kimi-k2.5` | ✅ 端到端通过 |
| DeepSeek | `deepseek-v3.2` | ✅ curl 正常 |
| OpenAI | `gpt-5-mini` | ⚠️ 见 5.2 节 |
| Anthropic | `claude-sonnet-4.6` | ✅ 已配置（未端到端复测） |

---

## 5. 协议层兼容性分析

### 5.1 OpenCode vs Codex 协议差异

| 特性 | OpenCode | Codex 0.133+ |
|------|----------|--------------|
| 主要 API | Chat Completions | Responses API |
| AI SDK 包 | `@ai-sdk/openai-compatible` | OpenAI 原生 Responses |
| b.ai 兼容性 | 🟢 直接可用 | 🔴 403 阻断 |
| 自定义 Base URL | `provider.*.options.baseURL` | `openai_base_url`（仍受 Responses 限制） |

### 5.2 推理型模型注意事项

部分 OpenAI 推理模型（如 `gpt-5-mini`）在 b.ai 上会将 `max_tokens` 优先分配给 **reasoning tokens**，可能导致 `message.content` 为空：

| 测试 | 参数 | 结果 |
|------|------|------|
| `gpt-5-mini` | `max_tokens=50` | ❌ content 为空，20 reasoning tokens |
| `gpt-5-mini` | `max_tokens=200, reasoning_effort=none` | ✅ 正常返回文本 |
| `kimi-k2.5` | `max_tokens=50` | ✅ 返回 `OK` |

**建议**：在 OpenCode 中优先使用 `kimi-k2.5`、`deepseek-v3.2` 等非推理型或 content 输出稳定的模型；若使用 gpt-5 系列，需在 Provider/模型层配置足够的 output limit 或关闭 reasoning。

---

## 6. 端到端测试记录

### 6.1 curl 直连测试

| 测试用例 | 方法 | 结果 |
|----------|------|------|
| 模型列表 | `GET /v1/models` | ✅ 通过 |
| Chat 推理 | `POST /v1/chat/completions` + `kimi-k2.5` | ✅ 返回 `OK` |
| Chat 推理 | `POST /v1/chat/completions` + `deepseek-v3.2` | ✅ 返回 `OK` |
| Responses 推理 | `POST /v1/responses` | ❌ HTTP 403（OpenCode 不需要） |

### 6.2 OpenCode CLI 测试

| 测试用例 | 命令/配置 | 结果 |
|----------|-----------|------|
| 配置解析 | `opencode debug config` | ✅ b.ai Provider 正确加载 |
| 模型列表 | `opencode models bai` | ✅ 4 个模型 |
| 凭据识别 | `opencode providers list` | ✅ `bai api` |
| 端到端推理 | `opencode run -m bai/kimi-k2.5 "Reply with exactly one word: COMPATIBLE"` | ✅ 返回 **COMPATIBLE** |
| Session 导出 | `opencode export ses_17cf821deffej5z5XPppmuOEmd` | ✅ finish=stop, output=3 tokens |

**Session 实测数据（`ses_17cf821deffej5z5XPppmuOEmd`）：**

| 字段 | 值 |
|------|-----|
| Provider | `bai` |
| Model | `kimi-k2.5` |
| Agent | `build` |
| 用户输入 | Reply with exactly one word: COMPATIBLE |
| 助手回复 | **COMPATIBLE** |
| finish | `stop` |
| Tokens | input=6619, output=3 |

> 注：首次 `opencode run` 因自动下载 ripgrep 延迟约 6 分钟；ripgrep 缓存后后续启动正常。

---

## 7. 兼容性评级

| 评估维度 | 评级 | 说明 |
|----------|------|------|
| **API 鉴权** | 🟢 兼容 | Key 有效，Bearer 认证正常 |
| **API 连通性** | 🟢 兼容 | DNS、TLS、HTTP 均正常 |
| **Chat Completions** | 🟢 兼容 | OpenCode 唯一需要的推理端点 |
| **Responses API** | ⚪ 不适用 | OpenCode 不使用该端点 |
| **OpenCode 推理** | 🟢 兼容 | Agent 任务可完成 |
| **OpenCode 本地安装** | 🟢 就绪 | 已安装、已配置、已验证 |
| **综合结论** | 🟢 **兼容** | 可直接用于日常开发 |

---

## 8. 风险与影响

| 风险 | 级别 | 影响 |
|------|------|------|
| API Key 已在对话中暴露 | **高** | 建议立即轮换 Key |
| 首次启动 ripgrep 下载慢 | 中 | 首次约 6 分钟；可 `brew install ripgrep` 缓解 |
| gpt-5 系列 reasoning 占 token | 中 | 可能导致空回复；换模型或调 output limit |
| b.ai 部分模型需充值 | 中 | 影响特定模型，与 OpenCode 框架兼容性无关 |
| b.ai 不支持 Responses API | 低 | **不影响 OpenCode**；仅影响 Codex |

---

## 9. 建议方案

### 推荐配置（已验证）

1. **安装 OpenCode**

```bash
npm install -g opencode-ai@latest
# 或: brew install anomalyco/tap/opencode
```

2. **配置凭据**（交互式）

```bash
cd /path/to/your/project
opencode providers login
# 选择 Other → 输入 provider id: bai → 粘贴 API Key
```

3. **项目配置** — 使用本仓库 `opencode.json`（或复制到目标项目）

4. **运行**

```bash
# 交互式 TUI
opencode

# 非交互式单次任务
opencode run --dangerously-skip-permissions -m bai/kimi-k2.5 "你的任务描述"
```

### 环境优化

```bash
# 避免首次启动长时间等待
brew install ripgrep

# 将 OpenCode 自带 rg 加入 PATH（可选）
export PATH="$HOME/.cache/opencode/bin:$PATH"
```

### 模型选择建议

| 场景 | 推荐模型 |
|------|----------|
| 日常编码 Agent | `bai/kimi-k2.5` 或 `bai/deepseek-v3.2` |
| 复杂推理 | `bai/claude-sonnet-4.6`（需确认账户余额） |
| 慎用 | `bai/gpt-5-mini`（reasoning token 可能导致空 content） |

---

## 10. 决策建议

| 场景 | 建议 |
|------|------|
| 使用 b.ai + 开源 Agent | **采用 OpenCode**，本报告已验证兼容 |
| 必须使用 Codex | 需更换支持 Responses 的 API 或部署桥接层（见 Codex 报告） |
| 仅需 API 调用（非 Agent） | b.ai Chat Completions 可直接用于 curl / SDK |
| 团队统一工具链 | OpenCode + b.ai 为当前环境下的**可行组合** |

**当前建议**：在 b.ai API 环境下，优先使用 **OpenCode** 替代 Codex 进行 AI 编程辅助。

---

## 11. 附录

### A. 测试环境

```
OS:              macOS 25.0.0 (aarch64)
OpenCode CLI:    1.15.13
安装路径:        /opt/homebrew/bin/opencode
项目目录:        /Users/wangjunhui/playcode/mymacos
配置文件:        ./opencode.json
凭据文件:        ~/.local/share/opencode/auth.json
API Endpoint:    https://api.b.ai/v1
测试日期:        2026-06-01
```

### B. 参考文档

- [OpenCode Providers 文档](https://opencode.ai/docs/providers)
- [OpenCode 自定义 Provider（openai-compatible）](https://opencode.ai/docs/providers#other-providers)
- [OpenCode GitHub](https://github.com/anomalyco/opencode)
- [Codex 兼容性评估报告（对比）](./Codex兼容性评估报告.md)

### C. 术语表

| 术语 | 说明 |
|------|------|
| Chat Completions | OpenAI 传统对话接口，OpenCode 通过 `@ai-sdk/openai-compatible` 使用 |
| Responses API | OpenAI 新一代推理接口，Codex 必需，b.ai 不支持 |
| Provider | OpenCode 配置中的模型提供商条目 |
| `opencode run` | 非交互式单次 Agent 执行命令 |

---

*报告基于 2026-06-01 实测数据生成。*
