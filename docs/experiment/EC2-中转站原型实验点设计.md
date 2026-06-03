# EC2 中转站原型实验点设计

> **文档类型**：实验方法论 · **非** 兼容性认证报告（网关侧指标与抽样 probe 可记入 reports；**三 Agent L3–L5 主结论在用户侧 Runner 完成**）  
> **范围**：境外 AWS EC2 上部署 **投入运营评估用的中转站原型**（基线实现：[New API](https://github.com/QuantumNous/new-api)）：Channel 接 **Anthropic / OpenAI 官方** 与 **同区域 AWS Bedrock**；管理员建 **用户 / Access Token**；将 **平台 Token + API 入口** 交付 [用户侧实验点](./EC2-用户侧隔离实验点设计.md) 作为 `experiment/user-side/sites.json` 中转源之一  
> **与 [中转站主流技术栈调研](../research/中转站主流技术栈调研.md) 的关系**：该文为产品栈 E1；本文为 **可运营的境外原型机** 搭建与交付规范  
> **与 [EC2-用户侧隔离实验点设计](./EC2-用户侧隔离实验点设计.md) 的关系**：本文 **不** 承担 Runner 出站审计与三 Agent 主矩阵；用户侧使用本文下发的 Key 完成 E4 与 N1–N3  
> **可选插件**：[中转站语料采集插件设计](./中转站语料采集插件设计.md)（**Corpus Tap**：简单规则 + 全量收集，按 `user_id` 分区；清洗与微调格式在离线完成）

### 文档元信息

| 项 | 内容 |
|----|------|
| **编写日期** | 2026-06-03 |
| **状态** | 设计稿（基础设施待实施） |
| **调研基线** | New API `v1.0.0-rc.9`（实施时锁定镜像 tag / digest） |
| **复审触发** | New API 大版本、Channel 类型变更、Bedrock 区域/模型策略、原型机 SG 调整、Corpus Tap 存储或采集规则变更 |

---

## 目录

1. [实验点定位](#1-实验点定位)
2. [实验要回答的问题](#2-实验要回答的问题)
3. [逻辑架构](#3-逻辑架构)
4. [评估维度（网关侧）](#4-评估维度网关侧)
5. [AWS 部署建议](#5-aws-部署建议)
6. [New API 配置（Channel / 用户 / Token）](#6-new-api-配置channel--用户--token)
7. [出站审计（中转站侧）](#7-出站审计中转站侧)
8. [可选：语料采集插件（Corpus Tap）](#8-可选语料采集插件corpus-tap)
9. [分阶段实施](#9-分阶段实施)
10. [交付用户侧（sites.json）](#10-交付用户侧sitesjson)
11. [证据归档](#11-证据归档)
12. [风险与局限](#12-风险与局限)
13. [实施检查清单](#13-实施检查清单)

---

## 1. 实验点定位

### 1.1 本实验点是什么

| 是 | 不是 |
|----|------|
| **运营商侧** 中转站原型：New API + MySQL（+ Redis） | 用户/developer Runner（见 [用户侧稿](./EC2-用户侧隔离实验点设计.md)） |
| 持 **上游** 原厂 Key / Bedrock IAM（Channel） | 在 Agent 配置里填写上游 Key |
| 创建 **实验用户** 与 **平台 Access Token** | 用平台 Token 跑 Claude/Codex/OpenCode 的主流程（在用户侧做） |
| 评估「建站 + 发券 + 协议面是否对用户暴露」 | 替代 b.ai 等商业站的 E3 结论 |

### 1.2 与商业中转站的关系

本原型用于 **自托管、全功能可配** 的基线；[b.ai](../reports/README.md#ba2026-06-01) 等商业站可在用户侧 EC2 上 **并行** 登记为另一中转源，无需重复建站。

### 1.3 实施前必须固定

| 固定项 | 否则会出现 |
|--------|------------|
| New API **版本 / digest** | 无法与后续 L4 回归对比 |
| Channel 与 `model_mapping` 表 | 用户侧 model 名对不上 |
| 平台 Token 与 Channel Key **分离** | 交付混淆，用户侧 N2 失真 |
| 对用户暴露的 **base URL**（内网/反代/TLS） | `experiment/user-side/sites.json` 与 probe 失败 |

---

## 2. 实验要回答的问题

### 2.1 运营与协议面（网关侧）

| 问题 | 验证方式 |
|------|----------|
| Channel 能否稳定转发至 Bedrock / OpenAI / Anthropic？ | New API 日志、管理台渠道测试、可选 curl |
| 对用户 Token 是否暴露 **三主端点**（Chat / Messages / Responses）？ | 在用户侧 Runner 执行 `probe-endpoints.sh newapi-prototype`（推荐）或网关机 curl |
| `model_mapping` 是否正确？ | 用户侧 `experiment/user-side/t_*` L4 |

### 2.2 中转站侧出站（G1–G2）

在 **仅 New API 进程** 应访问上游的前提下，原型机出站是否仅含 §7.2 清单？  
（**不** 替代用户侧 N1–N3。）

---

## 3. 逻辑架构

```text
┌──────────────────────────────────────────────────────────────────┐
│  VPC（独立账号 · 区域 = Bedrock 已开通 · 境外）                    │
│  ┌─ EC2 中转站原型（本实验点）─────────────────────────────────┐  │
│  │  Corpus Tap :8443 ──► New API :3000  +  MySQL (+ Redis)      │  │
│  │       │（可选；对用户暴露 Tap，见 §8）    │ Channel              │  │
│  │       │                                  ├──► Bedrock …        │  │
│  │       ├── 语料 S3 + corpus-db（Tap 专用）  ├──► api.openai …    │  │
│  │  管理台：内网 / SSH 隧道 only                                 │  │
│  └───────────────────────────┬──────────────────────────────────┘  │
│                              │ HTTPS（平台 Token → Tap 或直连 API） │
│                              ▼                                      │
│  ┌─ EC2 用户侧 Runner（另一实验点 · 另文档）────────────────────┐  │
│  │  Claude Code / Codex / OpenCode                              │  │
│  └──────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

**阶段 0**：仅部署本实验点 EC2，用 curl / 管理台验证 Channel；用户侧 EC2 可未就绪。  
**阶段 1+**：交付 §10 → 用户侧登记站点 → 用户侧 probe 与 L4。

---

## 4. 评估维度（网关侧）

| 层级 | 内容 | 在哪里测 |
|------|------|----------|
| **P2** | 四端点 HTTP（probe） | 用户侧 Runner（推荐）或网关机 |
| **G1** | 原型机空闲出站 | 本 EC2，无用户 Agent |
| **G2** | 经平台 Token 触发一次 relay 期间出站 | 本 EC2 New API → 上游 |
| **L3–L5** | Agent E4 | **[用户侧稿](./EC2-用户侧隔离实验点设计.md)**，报告 scope「经 `newapi-prototype`」 |

---

## 5. AWS 部署建议

| 项 | 建议 |
|----|------|
| **实例** | `t3.large` 起；≥ 40 GiB（New API + MySQL） |
| **网络** | 业务端口 **不对公网**；对用户侧 Runner SG **放行入站** 至 Corpus Tap `:8443`（或直连 New API `:3000`，无插件时） |
| **Bedrock** | Instance Profile + **VPC Interface Endpoint** |
| **密钥** | 上游 Key、MySQL、管理员密码：**SSM**；不进 Git |
| **合规** | New API **AGPL-3.0**：内部原型 / 私有 VPC；对外 SaaS 须法务评审 |

### 5.1 Compose 参考（实施落地，不进 Git）

```text
corpus-tap:8443 ──► new-api:3000 ──► mysql:8
                              └──► redis（按文档）
corpus-tap ──► corpus-db (PostgreSQL) + S3/MinIO
```

未启用语料插件时，用户侧可直连 `new-api:3000`（与 [语料插件设计](./中转站语料采集插件设计.md) §3.1 二选一）。

`./upstream/pull.sh newapi` 拉取参考源码到 `upstream/newapi/`（gitignored）。

---

## 6. New API 配置（Channel / 用户 / Token）

### 6.1 Channel

每上游 **一条 Channel**；实验期关闭自动 failover。

| Channel ID | 类型 | 上游凭据 |
|------------|------|----------|
| `ch-bedrock` | AWS / Bedrock | Instance Profile 或 SSM AK/SK + 区域 |
| `ch-openai` | OpenAI | SSM `OPENAI_API_KEY` |
| `ch-anthropic` | Anthropic | SSM `ANTHROPIC_API_KEY` |

**model_mapping（示例）**

| 对外 model（用户侧 `--model`） | 实际上游 |
|-------------------------------|----------|
| `claude-exp-bedrock` | Bedrock 上 Claude model ID |
| `claude-exp-anthropic` | `claude-haiku-4-5` 等 |
| `gpt-exp-openai` | `gpt-4o-mini` 等 |

### 6.2 用户与 Access Token（交付物）

| 步骤 | 操作 |
|------|------|
| 1 | 用户 `exp-runner`（或按运营流程命名） |
| 2 | 生成 **Access Token** → 写入 SSM，交付用户侧 `.env` |
| 3 | Token 可见模型 = §6.1 对外名列表 |
| 4 | 记录 **对用户可达** 的 `base_url`、`anthropic_base_url`（§10；启用 Tap 时 host 指向 Tap） |

**禁止** 将 Channel 上游 Key 写入用户侧 Runner。

### 6.3 网关侧抽样 probe（可选）

在网关机或已打通网络的运维机：

```bash
curl -H "Authorization: Bearer <平台Token>" https://<relay-host>/v1/models
# 完整四条见用户侧 probe-endpoints.sh
```

---

## 7. 出站审计（中转站侧）

### 7.1 观测手段

SG egress、Flow Logs、New API access log（脱敏）。

### 7.2 预期允许出站

| 阶段 | 源 | 目标 |
|------|-----|------|
| 构建 | 主机 | 镜像仓库、apt |
| G1 | 主机 | SSM、NTP、CloudWatch |
| G2 | New API | `api.anthropic.com`、`api.openai.com`、Bedrock VPC EP |
| G2 | Corpus Tap（若启用） | S3 / MinIO、corpus-db；New API MySQL **只读**（Token 解析） |
| 内网 | New API / Tap | MySQL、Redis、PostgreSQL（语料）私有 IP |

**不应** 在本实验点 EC2 上长期运行 Coding Agent CLI（避免与用户侧审计混淆）。临时调试须在报告中标注。

---

## 8. 可选：语料采集插件（Corpus Tap）

> 完整契约见 **[中转站语料采集插件设计](./中转站语料采集插件设计.md)**。

| 项 | 说明 |
|----|------|
| **目的** | 按 **New API `user_id`** 全量落盘入站/出站 JSON，供离线清洗与领域微调；**不在网关做清洗** |
| **插上即用** | 用户侧 `base_url` 指向 **Tap** 而非 `:3000`；Compose 增服务；**不改** New API 源码 |
| **与计费** | New API `logs` 仍管配额；语料在 **独立 PG + S3** |

实验期建议：每领域/实验用户 **独立 New API 用户 + Token**，便于 S3 前缀 `user_id=<n>/` 直接对应领域桶。

---

## 9. 分阶段实施

| 阶段 | 内容 | 产出 |
|------|------|------|
| **0** | VPC + EC2 + New API + MySQL + `ch-bedrock` + 1 Token | Channel 通、G1–G2 样例 |
| **0b** |（可选）+ Corpus Tap + S3 + corpus-db | 单用户 POST 全量入库验收 |
| **1** | + `ch-openai`、`ch-anthropic`；§10 写入 `experiment/user-side/sites.json` 草稿 | **交付包**（URL + Token 占位说明） |
| **2** | 用户侧 EC2 登记并 probe | P2 四端点记录 |
| **3** | 用户侧 L4 矩阵 | reports：`newapi-prototype × Agent` |
| **4** | 高可用 / 分机 / 反代 TLS | 运营加固（可选） |

---

## 10. 交付用户侧（sites.json）

原型 URL 稳定后，维护者提交（**host 按实际修改**）：

```json
"newapi-prototype": {
  "name": "New API relay prototype (EC2)",
  "base_url": "https://<relay-host>/v1",
  "anthropic_base_url": "https://<relay-host>",
  "api_key_env": "NEWAPI_PROTOTYPE_TOKEN",
  "default_models": {
    "claude": "claude-exp-bedrock",
    "codex": "gpt-exp-openai",
    "opencode": "claude-exp-bedrock"
  },
  "notes": "Platform Access Token; base_url host = Corpus Tap when enabled; L3-L5 on user-side Runner"
}
```

`experiment/user-side/.env.example`：

```bash
# Platform token from docs/experiment/EC2-中转站原型实验点设计.md (not upstream provider key)
# NEWAPI_PROTOTYPE_TOKEN=sk-...
```

用户侧流程：[EC2-用户侧隔离实验点设计](./EC2-用户侧隔离实验点设计.md) 模式 B → `cd experiment/user-side && ./scripts/run-user-side-compat.sh --site newapi-prototype`（或 `probe` + `./t_* -y`）。

**分机部署**：`base_url` 使用用户侧 SG 可访问的 **内网 IP 或私有域名**，勿写仅本机 `127.0.0.1`（除非用户侧用 SSH 隧道刻意为之，须在报告注明）。

---

## 11. 证据归档

| 产物 | 存放 |
|------|------|
| New API 版本、Channel 表（脱敏）、G1–G2 | 可附于用户侧 report 附录或单独 `newapi-prototype-运营评估.md` |
| P2 / L3–L5 | 用户侧 reports，scope **经 newapi-prototype** |
| 管理员操作 | 截图/导出勿含上游 Key |

---

## 12. 风险与局限

| 风险 | 说明 |
|------|------|
| Responses × Bedrock | tool 映射须用户侧 L4 验证 |
| AGPL | 对外运营需合规 |
| 与用户侧混淆 | 同一 EC2 既跑 New API 又跑 Agent 时，报告须拆角色或分机 |
| ≠ b.ai | 端点裁剪策略可能不同 |
| 语料合规 | 全量采集须用户授权与 retention；见 [语料插件设计](./中转站语料采集插件设计.md) |

---

## 13. 实施检查清单

**原型机**

- [ ] VPC + EC2 + Bedrock VPC EP + Instance Profile  
- [ ] New API + MySQL；改默认管理员密码  
- [ ] 三 Channel（按阶段）+ `model_mapping`  
- [ ] 用户 + 平台 Token → SSM  
- [ ] G1–G2 出站归档  
- [ ] Admin 不对 0.0.0.0  
- [ ]（可选）Corpus Tap + S3；`base_url` 指向 Tap；插件 §13 验收  

**交付**

- [ ] §10 `experiment/user-side/sites.json` + `experiment/user-side/.env.example`  
- [ ] 用户侧 SG 可访问 relay host:port  
- [ ] 通知用户侧执行 probe + L4  

**文档**

- [ ] [reports/README.md](../reports/README.md) 增加「newapi-prototype」索引（有结论后）  

---

## 参考链接

- [New API](https://github.com/QuantumNous/new-api) · [文档](https://docs.newapi.pro/)  
- [AWS Bedrock 端点](https://docs.aws.amazon.com/bedrock/latest/userguide/endpoints.html)  
- [EC2-用户侧隔离实验点设计](./EC2-用户侧隔离实验点设计.md)  
- [中转站语料采集插件设计](./中转站语料采集插件设计.md)  
- [中转站主流技术栈调研](../research/中转站主流技术栈调研.md)
