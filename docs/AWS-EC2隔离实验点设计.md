# AWS EC2 隔离实验点设计

> **文档类型**：实验方法论 · **非** 兼容性认证报告（实测结论写入 [reports/](./reports/)）  
> **范围**：官方上游（Anthropic、OpenAI、AWS Bedrock）→ LiteLLM → Claude Code / Codex / OpenCode，在 EC2 上完成 **E4 协议验证** 与 **出站行为审计**  
> **与 [E2E 原生兼容性全景](./E2E原生兼容性全景.md) 的关系**：全景矩阵为 **Agent 直连官方** 的 L1+L2；本文为 **统一网关 + 隔离网络** 的 E4 实验设计  
> **与 [编程 Agent 模型转换插件调研](./编程Agent模型转换插件调研.md) 的关系**：该文为 LiteLLM 等 L3 网关的 **方案地图（E0–E2）**；本文定义 **如何在云上可审计地跑 E4**

### 文档元信息

| 项 | 内容 |
|----|------|
| **编写日期** | 2026-06-03 |
| **状态** | 设计稿（基础设施与报告待实施） |
| **复审触发** | LiteLLM 大版本、任-Agent 主 wire 变更、AWS Bedrock OpenAI 兼容路径变更、实验点网络策略调整 |

---

## 目录

1. [为什么要先写本文](#1-为什么要先写本文)
2. [实验要回答的两个问题](#2-实验要回答的两个问题)
3. [逻辑架构](#3-逻辑架构)
4. [评估维度](#4-评估维度)
5. [AWS 部署建议](#5-aws-部署建议)
6. [出站审计（防火墙）](#6-出站审计防火墙)
7. [LiteLLM 与三 Agent 配置要点](#7-litellm-与三-agent-配置要点)
8. [分阶段实施](#8-分阶段实施)
9. [证据归档与报告](#9-证据归档与报告)
10. [风险与局限](#10-风险与局限)
11. [实施检查清单](#11-实施检查清单)

---

## 1. 为什么要先写本文

在开通 EC2、配密钥、装 Agent 之前，需要先固定：

| 固定项 | 否则会出现 |
|--------|------------|
| **允许出站清单** | 无法区分「违规外连」与「漏配的合法流量」 |
| **Agent 运行模式** | Codex OAuth、`doctor` 探针等混入结论 |
| **LiteLLM 版本与三端点** | 升级后 L4 tool 回归无法对比 |
| **通过 / 失败判定** | 只有主观「好像能聊」，无法写入 reports |

本文 **不写** 具体测试结果；跑通后按 [reports/README](./reports/README.md) 新增 `站点-Agent兼容性评估报告.md`，并在报告中引用本文章节号复现环境。

---

## 2. 实验要回答的两个问题

### 2.1 兼容性（E4）

在 **同一 LiteLLM 实例** 后挂 **Anthropic / OpenAI / Bedrock** 时，三类 Agent 是否满足各自 **主 wire** 下的 E2E 能力闭环（L2–L5）？

| Agent | 主 wire | LiteLLM 侧（须显式开启并实测） |
|-------|---------|--------------------------------|
| Claude Code | `POST /v1/messages` | Proxy `/v1/messages` |
| Codex ≥ 0.133 | `POST /v1/responses` | Proxy `/v1/responses`（版本 ≥ 1.66.3，以 LiteLLM 发行说明为准） |
| OpenCode | `POST /v1/chat/completions` 或官方 SDK 路径 | Proxy `/v1/chat/completions` 或 OpenAI-compatible Provider |

参照 [编程 Agent 模型转换插件调研](./编程Agent模型转换插件调研.md) §5.3、§7：**三端点须分别配置、分别探针**，不可假设「开了一个 `/v1/chat/completions` 即覆盖 Codex」。

### 2.2 出站行为（网络边界）

在 **仅允许** Agent → LiteLLM、LiteLLM → 上游（及下文 **预期基础设施出站**）的前提下，运行 Agent 默认工作流时：

- 是否仍存在 **未声明** 的 TCP/UDP 出站（官方 SaaS、OAuth、遥测、包管理器、MCP 默认远端等）？
- 若存在，**来源进程、目标 FQDN/IP、触发操作** 是否可记录？

该问题 **本机 `./t_*` 难以严格回答**（系统代理、浏览器、多网卡、后台服务混杂）。EC2 + 安全组 / 主机防火墙 +（可选）VPC Flow Logs 用于 **可举证** 的观测。

---

## 3. 逻辑架构

```text
┌─────────────────────────────────────────────────────────────┐
│  VPC（建议独立实验账号 / 独立 VPC）                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  EC2 实验主机（单角色或拆分为 Gateway + Runner）       │  │
│  │                                                       │  │
│  │   Claude Code ──┐                                     │  │
│  │   Codex CLI  ───┼──► LiteLLM Proxy (:4000 等)        │  │
│  │   OpenCode   ───┘         │                           │  │
│  │                           ▼                           │  │
│  │              ┌────────────┼────────────┐              │  │
│  │              ▼            ▼            ▼              │  │
│  │         Anthropic    OpenAI API    Bedrock           │  │
│  │         (官方)        (官方)      (同区域)            │  │
│  └───────────────────────────────────────────────────────┘  │
│  出站策略：默认拒绝，仅放行「预期清单」                      │
└─────────────────────────────────────────────────────────────┘
```

**推荐初始拓扑（阶段 0）**：LiteLLM 与 **一个** Agent 同机部署，先打通 **Claude Code → LiteLLM → Bedrock** 的出站审计流程，再扩三 Agent × 三上游。

**可选拆分（阶段 2+）**：

| 角色 | 职责 |
|------|------|
| `gateway` 实例 | 仅 LiteLLM + DB/Redis（若启用） |
| `runner` 实例 | 仅 Agent CLI；`ANTHROPIC_BASE_URL` / `OPENAI_BASE_URL` 指向 gateway 内网 IP |

拆分后可单独审计 **Runner 是否只连 Gateway**。

---

## 4. 评估维度

与 [E2E 原生兼容性全景](./E2E原生兼容性全景.md) §1.2 对齐，并 **增加网络层 N1–N3**：

| 层级 | 检查内容 | 本实验点 |
|------|----------|----------|
| L2 | 主 wire HTTP 可达 | `./scripts/probe-endpoints.sh` 指向 LiteLLM base URL |
| L3 | 流式不挂起 | Agent 交互或最小流式脚本 |
| L4 | tool / function 多轮 | 单次「读文件 + 写补丁」或 `codex exec` 最小任务 |
| L5 | reasoning / vision 等 | 按模型能力选测，不阻塞主线 |
| **N1** | 空闲态出站 | 安装后未启动 Agent：仅应有 apt/yum、SSM、NTP 等 **基础设施类** |
| **N2** | 单次推理出站 | 一轮对话期间：Runner → LiteLLM → 上游；无直连 api.anthropic.com 等（若策略禁止） |
| **N3** | 工具链出站 | L4 任务期间：是否出现 npm registry、MCP 远端、web search 等 **非模型 API** |

**判定原则**：

- **兼容性 FAIL**：L2 失败，或 L3/L4 在记录配置下不可复现。  
- **出站 FAIL**：出现 **未在 §6.2 预期清单** 中的目标，且不能归因于实验操作失误（如误开 OAuth）。  
- **出站 WARN**：清单外流量但已文档化（如 CLI 首次 `npm install`）— 须在报告中标注 **一次性 / 可关闭**。

---

## 5. AWS 部署建议

| 项 | 建议 |
|----|------|
| **区域** | 与 Bedrock 模型开通区域一致 |
| **实例** | `t3.large` 或同等（三 Agent 同机时略宽裕）；存储 ≥ 30 GiB |
| **IAM** | EC2 Instance Profile：`bedrock:InvokeModel`（及所用 API 的最小权限）；**不把** 云凭证写入 Agent 配置，由 LiteLLM 持 Bedrock 凭据 |
| **Bedrock 连通** | 优先 **VPC Interface Endpoint**（`com.amazonaws.<region>.bedrock-runtime`），减少公网出站条目 |
| **密钥** | 使用 SSM Parameter Store / Secrets Manager；`.env` 仅存在于实例且 **不** 进入 Git |
| **与本仓库启动器** | 可将 `sites.json` 扩展 LiteLLM 站点条目，或实验期在实例手写 `ANTHROPIC_BASE_URL=http://127.0.0.1:4000` 等；**不要求** 依赖 `opencode/`、`codex/` 源码目录 |

---

## 6. 出站审计（防火墙）

### 6.1 观测手段（可叠加）

| 手段 | 粒度 | 用途 |
|------|------|------|
| Security Group **egress 拒绝默认 + 白名单** | ENI | 粗粒度「能否出去」 |
| `nftables` / `iptables` LOG | 主机 | 看到被拒绝的连接尝试 |
| VPC Flow Logs | ENI | 事后审计、与 SG 对照 |
| 可选：透明 HTTP 代理 | 应用层 SNI/Host | 区分同一 443 上的不同服务 |

### 6.2 预期允许出站（示例，实施前按账号改 FQDN）

**基础设施（实例生命周期）**

| 目标 | 说明 |
|------|------|
| `amazonaws.com` 相关 | SSM、EC2 元数据、CloudWatch（若启用） |
| NTP | 时间同步 |
| 组织允许的镜像源 | 仅 **构建期**；测 N1 时应已安装完毕 |

**实验路径（N2）**

| 源 | 目标 | 说明 |
|----|------|------|
| Runner 进程 | LiteLLM 监听地址 | 内网 IP 或 `127.0.0.1` |
| LiteLLM | `api.anthropic.com`、`api.openai.com`、Bedrock endpoint / VPC endpoint | 仅网关进程需要 |

**明确禁止或须关闭后再测（否则 N2/N3 失真）**

| 类别 | 示例 | 处理 |
|------|------|------|
| Codex ChatGPT OAuth | `chatgpt.com` | 仅用 API Key + `OPENAI_BASE_URL` 指 LiteLLM |
| Codex `codex doctor` 默认探针 | 多域名 WS/HTTP | 实验脚本避免 `doctor`，或单独记 WARN |
| Claude / Codex 自动更新 | `registry.npmjs.org` 等 | 阶段 0 预装固定版本 CLI |
| MCP 默认远端 | 视配置 | `config.toml` / settings 清空未用 MCP |
| OpenCode 插件市场 | 视版本 | 最小配置，仅被测 Provider |

### 6.3 违规判定模板

```text
时间窗：[安装完成] [空闲 N1] [单次对话 N2] [L4 任务 N3]
记录：Flow Log / nftables LOG / LiteLLM access log（脱敏）
违规：目标 ∉ 预期清单 且 操作步骤未声明会触发
```

---

## 7. LiteLLM 与三 Agent 配置要点

### 7.1 LiteLLM

- **锁定版本**（容器 tag 或 `pip freeze`），写入未来 report 的「评估环境」表。  
- **独立路由**：为 Anthropic / OpenAI / Bedrock 各建 `model_name`，避免混用导致 tool 名映射错误（见插件调研 **E2** 条目：Codex + Bedrock + LiteLLM）。  
- **启用端点**：`/v1/messages`、`/v1/responses`、`/v1/chat/completions` 与 Admin 端口 **不要** 对公网开放；仅 SG 内 Runner 可访问。  
- **日志**：开启 access log（脱敏 Key），便于与 Flow Log 对照时间戳。

### 7.2 Claude Code

```bash
export ANTHROPIC_BASE_URL="http://<litellm-host>:4000"
export ANTHROPIC_API_KEY="<litellm-virtual-key-or-master>"
# 勿设置会绕过 BASE_URL 的官方直连项；Bedrock 路径由 LiteLLM 翻译，非 Claude 内置 Bedrock 模式
```

### 7.3 Codex

```toml
# ~/.codex/config.toml 或 .runtime/codex.<site>.toml
model_provider = "openai"   # 或自定义 provider 块
# OPENAI_BASE_URL → LiteLLM /v1/responses 前缀
# wire_api 必须为 responses（0.133+）
```

- **禁止** 实验主路径使用 `chatgpt.com/backend-api/codex` 登录。  
- Bedrock 经 LiteLLM 时，**不要** 与 Codex 内置 `amazon-bedrock` provider 混测（内置为 SigV4 直连 Bedrock，绕开 LiteLLM，破坏出站审计）。

### 7.4 OpenCode

- Provider 指向 LiteLLM 的 OpenAI-compatible `baseURL`（Chat Completions）。  
- 与 [OpenCode 兼容性评估报告](./reports/OpenCode兼容性评估报告.md) 相同：先 L2 `probe-endpoints`，再 L4。

---

## 8. 分阶段实施

| 阶段 | 内容 | 产出 |
|------|------|------|
| **0** | VPC + EC2 + SG 白名单 + LiteLLM + Bedrock；Claude Code 单 Agent | N1–N2 流程跑通；出站记录样例 |
| **1** | 同上 + L3–L4 最小任务；记录 LiteLLM 版本 | `reports/` 首份 **EC2-LiteLLM-Bedrock × Claude Code** 报告 |
| **2** | 增加 Codex、OpenCode；LiteLLM 三端点 | 三 Agent × Bedrock 矩阵（可先单上游） |
| **3** | LiteLLM 切换 OpenAI、Anthropic 官方路由 | 3×3 矩阵（按需裁剪 L5） |
| **4** | Gateway / Runner 拆分 | 强化 N2「Runner 仅连 Gateway」证据 |

**停止扩矩阵条件**：上一阶段 L4 或 N2 未通过，先修配置或升级 LiteLLM，再扩。

---

## 9. 证据归档与报告

每份 [reports/](./reports/) 报告建议包含：

| 章节 | 内容 |
|------|------|
| 评估环境 | 区域、实例类型、LiteLLM 版本、三 Agent 版本、本文档 **编写日期** 与 commit |
| 网络策略 | SG 规则摘要（无账号 ID）、N1–N3 时间窗 |
| 配置片段 | 脱敏后的 LiteLLM yaml、`config.toml`、`settings.json` 片段 |
| 协议结论 | L2–L5 表格 |
| 出站结论 | 预期 vs 实际表；违规/WARN 分项 |
| 复现步骤 | 引用本文 §11 检查清单 |

报告命名示例：`EC2-LiteLLM-Bedrock-ClaudeCode兼容性评估报告.md`（站点名与上游组合体现在文件名中）。

---

## 10. 风险与局限

| 风险 | 说明 |
|------|------|
| 白名单不完整 | 误把合法流量标为违规，或漏放导致假阴性「一切正常」 |
| TLS SNI 共用 | 多服务同 IP 时 Flow Log 只有 IP，需代理或 LiteLLM 日志补 Host |
| LiteLLM 与官方行为差 | E4 通过 ≠ 直连官方 E2E 原生 ●；结论 scope 须写清 **经 LiteLLM** |
| 成本与密钥 | Bedrock + 三上游并发 L4 会产生推理费用；密钥轮换后须重跑 L2 |
| 地域合规 | 数据是否经美国区 OpenAI/Anthropic API，由 LiteLLM 路由决定，与 Bedrock 区域策略分开评估 |

---

## 11. 实施检查清单

**设计（本文档阶段）**

- [ ] 确认实验账号 / VPC / 区域与 Bedrock 模型 ID  
- [ ] 写出 §6.2 预期出站清单（FQDN + 端口）  
- [ ] 选定 LiteLLM 版本与三端点配置草稿（密钥占位，不进 Git）

**基础设施**

- [ ] EC2 + Instance Profile（Bedrock）  
- [ ] SG egress 白名单 +（可选）Flow Logs  
- [ ] Bedrock VPC endpoint（推荐）  
- [ ] LiteLLM 部署且仅内网可达  

**基线探测**

- [ ] `probe-endpoints.sh` 对 LiteLLM base URL  
- [ ] N1 空闲出站记录归档  

**Agent（按阶段递增）**

- [ ] 预装固定版本 CLI，关闭实验期自动更新  
- [ ] Claude Code：`ANTHROPIC_BASE_URL` → LiteLLM  
- [ ] Codex：`OPENAI_BASE_URL` + `wire_api=responses`，无 OAuth  
- [ ] OpenCode：Provider → LiteLLM Chat  
- [ ] L4 最小任务 + N3 出站记录  

**文档**

- [ ] 新增/更新 [reports/README.md](./reports/README.md) 索引  
- [ ] 结论不写入 README 正文（保持仓库约定）

---

## 参考链接

- [LiteLLM Proxy 文档](https://docs.litellm.ai/docs/proxy/quick_start)  
- [Claude Code on Bedrock（官方直连，对照用）](https://code.claude.com/docs/en/amazon-bedrock)  
- [AWS Bedrock 端点](https://docs.aws.amazon.com/bedrock/latest/userguide/endpoints.html)  
- 本仓库：[中转站主流技术栈调研](./中转站主流技术栈调研.md)、[Codex 技术架构调研](./Codex技术架构调研.md)
