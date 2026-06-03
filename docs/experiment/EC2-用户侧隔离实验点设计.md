# EC2 用户侧隔离实验点设计

> **文档类型**：实验方法论 · **非** 兼容性认证报告（实测结论写入 [reports/](../reports/)）  
> **范围**：境外（或隔离）**EC2 Runner** 上模拟 **终端开发者** 行为：安装三 Agent、配置凭据、跑 L2–L5，并做 **出站审计（N1–N3）**  
> **凭据模式**：**原厂 Key 直连** 或 **中转站平台 Token**（含 [中转站原型实验点](./EC2-中转站原型实验点设计.md) 下发的 Key、b.ai 等 `experiment/user-side/sites.json` 站点）  
> **与 [E2E 原生兼容性全景](../research/E2E原生兼容性全景.md) 的关系**：直连原厂模式对齐全景 **L1+L2+**；中转站模式为 **经网关的 E4**  
> **与 [EC2-中转站原型实验点设计](./EC2-中转站原型实验点设计.md) 的关系**：中转站稿定义 **运营商侧 New API**；本文定义 **用户侧如何接入**；原型站点的 `base_url` + 平台 Token 登记到 `experiment/user-side/sites.json` 后在本实验点使用

### 文档元信息

| 项 | 内容 |
|----|------|
| **编写日期** | 2026-06-03 |
| **状态** | 设计稿（基础设施与报告待实施） |
| **复审触发** | 任-Agent 主 wire 变更、新增中转站源、用户侧 SG 白名单或 Runner 拓扑调整 |

---

## 目录

1. [实验点定位](#1-实验点定位)
2. [实验要回答的问题](#2-实验要回答的问题)
3. [逻辑架构](#3-逻辑架构)
4. [两种凭据模式](#4-两种凭据模式)
5. [评估维度](#5-评估维度)
6. [EC2 部署建议](#6-ec2-部署建议)
7. [出站审计（防火墙）](#7-出站审计防火墙)
8. [启动器与自动化（t_*）](#8-启动器与自动化t_)
9. [三 Agent 配置要点](#9-三-agent-配置要点)
10. [分阶段实施](#10-分阶段实施)
11. [证据归档与报告](#11-证据归档与报告)
12. [风险与局限](#12-风险与局限)
13. [实施检查清单](#13-实施检查清单)

---

## 1. 实验点定位

### 1.1 本实验点是什么

| 是 | 不是 |
|----|------|
| 模拟用户笔记本 / 开发机在 **隔离网络** 上跑 Coding Agent | 中转站运营商服务器（见 [中转站原型稿](./EC2-中转站原型实验点设计.md)） |
| 可切换 **直连原厂** 与 **连中转站** 两种配置做对照 | 在 Runner 上部署 New API / LiteLLM 作为常态（除非单独做调试） |
| 用 SG / Flow Log 证明 **用户进程** 的出站是否符合预期 | 在本机代替 EC2 做 N1–N3（结论 scope 须标注 Runner 环境） |
| **`t_claude` / `t_codex` / `t_opencode` 作为兼容性自动化入口** | 在中转站原型 EC2 上跑 Agent（见 [中转站原型稿](./EC2-中转站原型实验点设计.md)） |

### 1.2 与中转站原型的协作关系

```text
[ 中转站原型 EC2 ]  New API 建站、Channel、发平台 Token
        │
        │  交付：base_url、anthropic_base_url、Access Token、对外 model 名
        ▼
[ 用户侧 EC2 ]      clone 本仓库 → cd experiment/user-side → sites.json + .env → probe + ./t_*（自动化，§8）
```

商业站（如 **b.ai**）无需自建原型，也可直接在用户侧 EC2 上作为「中转站源」登记测试。

### 1.3 实施前必须固定

| 固定项 | 否则会出现 |
|--------|------------|
| **当前模式**（直连 / 中转） | 出站白名单与结论 scope 混乱 |
| **sites.json 站点 id**（`experiment/user-side/` 下） | 报告无法复现 |
| **Agent 版本与禁用项** | OAuth、`doctor` 污染 N2/N3 |
| **通过 / 失败判定** | 无法写入 [reports/](../reports/) |

---

## 2. 实验要回答的问题

### 2.1 兼容性（L2–L5 / E4）

在 **固定 Runner 环境** 下，对某一 **上游源**（原厂或某一中转站站点）：

- `./scripts/probe-endpoints.sh <site>` 是否满足该 Agent 主 wire？  
- `./t_* --site <site>` 是否 L3–L5 可复现？

| Agent | 主 wire |
|-------|---------|
| Claude Code | `POST /v1/messages` |
| Codex ≥ 0.133 | `POST /v1/responses` |
| OpenCode | `POST /v1/chat/completions` |

**报告 scope 必写**：

- **直连原厂**：对齐 E2E 全景，标注「Runner 直连 api.* / Bedrock」。  
- **经中转站**：标注站点 id（如 `newapi-prototype`、`b.ai`），**不等于** 其他中转站或全景 ●。

### 2.2 出站行为（N1–N3）

在用户侧 EC2 上，Agent 工作流是否出现 **未声明** 外连？  
模式不同，§7.2 允许目标不同（直连允许原厂 API；中转模式仅允许中转站 FQDN）。

---

## 3. 逻辑架构

```text
┌─────────────────────────────────────────────────────────────┐
│  VPC（建议独立实验账号）                                     │
│  ┌─ EC2 Runner（用户侧实验点）────────────────────────────┐  │
│  │  Claude Code / Codex / OpenCode                        │  │
│  │         │                                              │  │
│  │         ├── 模式 A：原厂 Key ──────────► api.* / Bedrock│  │
│  │         │                                              │  │
│  │         └── 模式 B：平台 Token ────────► 中转站 URL     │  │
│  │                      （原型 / b.ai / 其他 sites.json 站点） │  │
│  └───────────────────────────────────────────────────────┘  │
│  SG：按 §7 按模式切换白名单                                  │
└─────────────────────────────────────────────────────────────┘
```

**不推荐** 在用户侧 EC2 上长期同机部署 New API（与中转站原型角色混淆）。调试时可临时单机，报告须注明 **非标准拓扑**。

---

## 4. 两种凭据模式

### 4.1 模式 A：原厂直连

| 项 | 说明 |
|----|------|
| **凭据** | `ANTHROPIC_API_KEY`、`OPENAI_API_KEY`；Bedrock 用 Instance Profile 或专用 Key（**仅** 当测试 Claude/Codex 官方 Bedrock 集成时） |
| **BASE_URL** | 默认官方；或不设置 `*_BASE_URL` |
| **`experiment/user-side/sites.json`** | 可增 `official-openai` 等条目便于 `probe`；或手写 env |
| **用途** | Positive control、与 E2E 全景对照、排除 Runner/SG 误伤 |

### 4.2 模式 B：中转站

| 项 | 说明 |
|----|------|
| **凭据** | 中转站下发的 **平台 Access Token**（非 Channel 上游 Key） |
| **BASE_URL** | 站点 `base_url` / `anthropic_base_url`（来自 [中转站原型](./EC2-中转站原型实验点设计.md) §9 或商业站文档） |
| **`experiment/user-side/sites.json`** | 例：`newapi-prototype`、`b.ai` |
| **model** | 使用中转站对外映射名（如 `claude-exp-bedrock`） |
| **禁止** | Claude 内置 Bedrock、Codex `amazon-bedrock` provider（绕过中转 URL） |

### 4.3 模式切换记录

同一份报告若包含两种模式，须分表列出 L2–L5 与 N2，**不得** 混为一列 PASS。

---

## 5. 评估维度

与 [E2E 全景 §1.2](../research/E2E原生兼容性全景.md) 对齐，并增加 **N1–N3**：

| 层级 | 检查内容 |
|------|----------|
| L2 | `probe-endpoints.sh <site>` |
| L3 | 流式不挂起 |
| L4 | §9.2 最小任务 |
| L5 | 按模型选测 |
| **N1** | 空闲态出站 |
| **N2** | 单次推理出站（见 §7.2 当前模式） |
| **N3** | L4 期间非模型 API |

---

## 6. EC2 部署建议

| 项 | 建议 |
|----|------|
| **区域** | 与待测 Bedrock / 中转站原型 **同区域或低延迟**；直连原厂可不绑 Bedrock 区域 |
| **实例** | `t3.large` 起；仅 Runner ≥ 30 GiB |
| **IAM** | 模式 A 测 Bedrock 时：Instance Profile `bedrock:InvokeModel`；模式 B **通常不需要** Bedrock IAM（推理在中转站侧） |
| **密钥** | SSM / Secrets Manager；`.env` 不提交 Git |
| **启动器** | `./t_claude` / `./t_codex` / `./t_opencode --site <id> -y` |

---

## 7. 出站审计（防火墙）

### 7.1 观测手段

Security Group 白名单、`nftables` LOG、VPC Flow Logs（可选）。

### 7.2 预期允许出站（按模式选一）

**基础设施（两种模式共有，N1）**

| 目标 | 说明 |
|------|------|
| `amazonaws.com` 相关 | SSM、元数据、CloudWatch |
| NTP | 时间同步 |
| 构建期镜像源 / npm | 仅 **预装 CLI** 阶段；N1 窗口不应出现 |

**模式 A — N2（直连原厂）**

| 源 | 目标 |
|----|------|
| Agent | `api.anthropic.com`、`api.openai.com`、Bedrock endpoint（若测 Bedrock） |

**模式 B — N2（经中转站）**

| 源 | 目标 |
|----|------|
| Agent | 中转站 `base_url` 的 host（如原型机内网 IP、HTTPS 域名） |
| **不应出现** | `api.anthropic.com` 等（除非同时开了模式 A 配置残留） |

**须禁止或关闭（N2/N3 失真）**

Codex OAuth、`codex doctor`（主路径）、CLI 自动更新、未用 MCP / OpenCode 插件市场。

### 7.3 违规判定模板

```text
模式：[A 直连 / B 中转 <site-id>]
时间窗：[N1] [N2] [N3]
违规：目标 ∉ 当前模式 §7.2
```

---

## 8. 启动器与自动化（t_*）

### 8.1 定位

本仓库 **`experiment/user-side/t_claude`、`t_codex`、`t_opencode`** 的设计落点即 **用户侧 EC2 Runner** 上的 **兼容性自动化入口**（非中转站原型机上的常驻服务）。以下命令均在 **`experiment/user-side/` 目录** 下执行：

| 组件 | 作用 |
|------|------|
| `sites.json` + `.env` | 登记站点（原厂 / `newapi-prototype` / `b.ai` 等）与凭据 |
| `scripts/probe-endpoints.sh` | **L2** 四端点自动化探测 |
| `./t_<agent>` | 解析站点 → 写临时配置 → 调用底层 CLI（**L3–L5**） |
| `scripts/run-user-side-compat.sh` | Runner 上一键：**probe + 可选 smoke**（§8.3） |

开发者在 **本机** 也可 `cd experiment/user-side` 运行同一套 `t_*` 做快速调试，但 **带 N1–N3 的 E4 认证** 应在用户侧 EC2 上复跑；报告须写明 Runner 环境。

### 8.2 Runner 部署约定

```bash
# 示例：用户侧 EC2
git clone <api_compatible> && cd api_compatible/experiment/user-side
cp .env.example .env    # 由 SSM 注入 Key / 平台 Token
# 预装 Node（OpenCode）、可选禁止实验期 npm 全局升级

./scripts/run-user-side-compat.sh --site newapi-prototype
./scripts/run-user-side-compat.sh --site b.ai --probe-only
```

| 项 | 建议 |
|----|------|
| **仓库路径** | 固定如 `/opt/api_compatible/experiment/user-side`（SSM 下发 `.env` 不进 Git） |
| **代理** | 境外 Runner 通常 `MAAS_PROXY_SKIP=1`；大陆调试机再用 `MAAS_PROXY` |
| **非交互** | 自动化必须带 **`-y`**，并透过 `--` 传入 CLI 非交互子命令（§8.3） |
| **产物** | `experiment/user-side/.claude/settings.json`、`.runtime/codex.*`、`.runtime/opencode.*` 仅在 Runner 本地 |

### 8.3 自动化分层

| 层级 | 命令 | 自动化程度 |
|------|------|------------|
| L2 | `./scripts/probe-endpoints.sh <site>` | 全自动 |
| L3–L4 smoke | `run-user-side-compat.sh --smoke` | 调用 `t_*` + 非交互子命令（可扩展） |
| L4 完整 tool 链 | 单独编排 `t_* -- …`（Grep、`codex exec` 等） | 半自动；写入报告 |

**Smoke 示例（实现于 `run-user-side-compat.sh`）**：

```bash
./t_claude --site "$SITE" -y -- --print --max-budget-usd 1.00 "Reply with exactly: API OK"
./t_codex   --site "$SITE" -y -- exec "Reply with exactly: API OK"
./t_opencode --site "$SITE" -y -- run "Reply with exactly: API OK"
```

完整 L4（多轮 tool）在通过 smoke 后按 §10.2 单独执行，并采集 N3。

### 8.4 与中转站原型的接口

1. 原型机交付 URL + 平台 Token（[中转站原型 §9](./EC2-中转站原型实验点设计.md#9-交付用户侧sitesjson)）。  
2. 维护者合并 `experiment/user-side/sites.json` 条目 `newapi-prototype`。  
3. 用户侧 EC2 执行 §8.2 → 结论写入 `docs/reports/`。

---

## 9. 三 Agent 配置要点

统一流程：在 `experiment/user-side/sites.json` 登记站点 → `experiment/user-side/.env` 填 `api_key_env` → `cd experiment/user-side` 后 `./t_* --site <id> -y`（自动化）或交互调试时不加 `-y`。

### 9.1 模式 B 示例（中转站原型）

```bash
# Claude Code — 平台 Token，非 Anthropic 原厂 Key
export ANTHROPIC_BASE_URL="https://<relay-host>"    # 无尾斜杠
export ANTHROPIC_API_KEY="<平台 Access Token>"
./t_claude --site newapi-prototype --model claude-exp-bedrock -y
```

Codex / OpenCode：Key = 平台 Token；`OPENAI_BASE_URL` / Provider `baseURL` 指向中转站 `/v1` 前缀；Codex 保持 `wire_api=responses`，禁用 OAuth。

### 9.2 模式 A 示例（原厂）

```bash
export ANTHROPIC_API_KEY="<anthropic-official-key>"
# 不设置 ANTHROPIC_BASE_URL（或指向官方）
./t_claude --site <official-site-if-any> -y
```

---

## 10. 分阶段实施

| 阶段 | 内容 | 产出 |
|------|------|------|
| **0** | Runner EC2 + SG + 预装 CLI；模式 A 单 Agent probe + N1–N2 | 直连基线 |
| **1** | 模式 A Claude L4；可选 Positive 已含在 A | `reports/` 直连样例 |
| **2** | 中转站原型交付 Token；`experiment/user-side/sites.json` 登记；模式 B probe + N2 | 验证仅连中转 host |
| **3** | 模式 B 三 Agent L4（按 probe 结果裁剪 Codex） | `newapi-prototype × Agent` 报告 |
| **4** | 增加 b.ai 或其他商业站为第二中转源 | 多站点对比 |

**停止条件**：上一阶段 L4 或 N2 未通过，不扩。

### 10.2 L4 最小任务

| Agent | 任务 |
|-------|------|
| Claude Code | `--tools Grep` 搜固定字符串 |
| Codex | `codex exec` 单行非交互 |
| OpenCode | 单轮对话或一次读文件 |

---

## 11. 证据归档与报告

| 章节 | 内容 |
|------|------|
| 评估环境 | Runner 区域、实例、Agent 版本、**凭据模式**、站点 id |
| 网络策略 | SG 摘要、N1–N3、**模式 A/B** |
| 协议结论 | L2–L5（分模式分表） |
| 出站结论 | 预期 vs 实际 |
| 复现步骤 | §13 + `run-user-side-compat.sh` 参数 + 中转站原型 §9（若模式 B） |

命名示例：`EC2-Runner-newapi-prototype-ClaudeCode兼容性评估报告.md`、`EC2-Runner-OfficialAnthropic-ClaudeCode兼容性评估报告.md`。

---

## 12. 风险与局限

| 风险 | 说明 |
|------|------|
| 模式混用 | Agent env 残留导致 N2 假违规/假通过 |
| 中转站不可达 | 模式 B 失败需先查原型 EC2 / SG，勿误判 Agent |
| 本机 vs EC2 | 本机 `experiment/user-side/t_*` 不能替代 N1–N3；EC2 结论不自动适用于开发者笔记本 |
| 商业站裁剪 | b.ai 缺 Responses 与原型全功能不同，须分站点写 scope |

---

## 13. 实施检查清单

**Runner 基础设施**

- [ ] VPC + EC2 + SG（按模式维护两套 egress 规则或实验前切换）  
- [ ] 克隆本仓库至 Runner（如 `/opt/api_compatible`）  
- [ ] 预装固定版本 CLI；`MAAS_PROXY_SKIP=1`（境外）  
- [ ] SSM 注入 `.env` 凭据  

**自动化（t_*）**

- [ ] `./scripts/run-user-side-compat.sh --site <id>` L2 通过  
- [ ] 可选 `--smoke` 三 Agent 非交互通过  
- [ ] L4 / N3 按 §10.2 补跑并归档日志  

**模式 A**

- [ ] 原厂 Key / Bedrock IAM（若需要）  
- [ ] probe + N1–N2 + 至少 1 个 L4  

**模式 B**

- [ ] [中转站原型](./EC2-中转站原型实验点设计.md) 已交付 Token 与 URL  
- [ ] `experiment/user-side/sites.json` + `experiment/user-side/.env` 登记  
- [ ] probe 四条端点归档  
- [ ] N2 仅连中转 host + L4 + N3  

**文档**

- [ ] 更新 [reports/README.md](../reports/README.md)  
- [ ] 结论不写入根 README  

---

## 参考链接

- [中转站主流技术栈调研](../research/中转站主流技术栈调研.md)  
- [EC2-中转站原型实验点设计](./EC2-中转站原型实验点设计.md)  
- [编程 Agent 模型转换插件调研](../research/编程Agent模型转换插件调研.md)  
- [b.ai 报告索引](../reports/README.md#ba2026-06-01)
