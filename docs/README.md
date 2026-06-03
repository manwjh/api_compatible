# 文档索引 — 研究主产出

本目录是 **API Compatible** 的核心交付：`research` 建立判断基线，`experiment` 定义可复现实验，`reports` 归档站点 × Agent 的实证。根目录代码（`t_*`、`scripts/`、`corpus-tap/`）仅用于执行这里的实验设计，见 [根 README § 代码角色](../README.md#代码在仓库中的角色)。

---

## 三层结构

| 目录 | 角色 | 读者 |
|------|------|------|
| **[research/](./research/)** | 调研与参考：协议矩阵、中转站栈、网关转换地图 | 建立「什么算兼容」的理论基线 |
| **[experiment/](./experiment/)** | 云上实验点与插件契约：拓扑、变量、出站与语料 | 设计或复现实验 |
| **[reports/](./reports/)** | 站点 × Agent **实测结论**与复现步骤 | 查证据、写新报告 |

> Key 有效、`/v1/models` 可达 **≠** Agent 可用 — 主线判断见 [E2E 原生兼容性全景](./research/E2E原生兼容性全景.md)。

---

## 推荐阅读顺序

### 路径 A — 理解兼容性研究（通用）

1. [E2E 原生兼容性全景](./research/E2E原生兼容性全景.md)  
2. 评估 Token 站时：[中转站主流技术栈调研](./research/中转站主流技术栈调研.md)  
3. 需要协议桥接对照时：[编程 Agent 模型转换插件调研](./research/编程Agent模型转换插件调研.md)  
4. 查已有实证：[reports/](./reports/)（索引见 [reports/README.md](./reports/README.md)）

### 路径 B — 云上中转站 + 用户侧评估（完整实验）

1. [中转站主流技术栈调研](./research/中转站主流技术栈调研.md)（选型）  
2. [New API 技术栈全景](./research/NewAPI技术栈全景.md)（读源码 / 建站前）  
3. [EC2-中转站原型实验点设计](./experiment/EC2-中转站原型实验点设计.md)（建站、发 Token）  
4. [EC2-用户侧隔离实验点设计](./experiment/EC2-用户侧隔离实验点设计.md)（`run-user-side-compat.sh` / `t_*`）  
5. 可选语料：[中转站语料采集插件设计](./experiment/中转站语料采集插件设计.md) + [corpus-tap/](../corpus-tap/)  
6. 结论写入 [reports/](./reports/)

路径 B 的变量控制与 Runner 分工以 experiment 文档为准；**不在**中转站原型 EC2 上常规跑 `t_*`。

---

## research/

| 文档 | 说明 |
|------|------|
| [E2E原生兼容性全景.md](./research/E2E原生兼容性全景.md) | Agent 直连官方上游的兼容矩阵（研究基线） |
| [中转站主流技术栈调研.md](./research/中转站主流技术栈调研.md) | Token 中转站实现栈与 L2 探测 |
| [NewAPI技术栈全景.md](./research/NewAPI技术栈全景.md) | New API 单产品架构、Relay、Channel 与部署 |
| [编程Agent模型转换插件调研.md](./research/编程Agent模型转换插件调研.md) | 网关 / 协议桥接方案地图（非 E4 实证） |

---

## experiment/

| 文档 / 代码 | 说明 |
|-------------|------|
| [EC2-用户侧隔离实验点设计.md](./experiment/EC2-用户侧隔离实验点设计.md) | Runner + `t_*` + 出站审计；原厂或中转站 Token |
| [EC2-中转站原型实验点设计.md](./experiment/EC2-中转站原型实验点设计.md) | New API 运营原型；Token 交付用户侧 |
| [中转站语料采集插件设计.md](./experiment/中转站语料采集插件设计.md) | Corpus Tap：规则 + 全量采集（插上即用契约） |
| [corpus-tap/](../corpus-tap/) | 插件设计原型（Go 代理 + PG 迁移 + Compose 片段） |

---

## reports/

站点 × Agent 的 L2–L5 结论、阻塞原因与复现命令。本目录不写方法论，只写**证据**。

- 索引与样例结论：[reports/README.md](./reports/README.md)  
- 新增报告：完成实验点中的探针后，新增 `站点-Agent兼容性评估报告.md` 并更新索引

---

## 与根目录的关系

| 位置 | 内容 |
|------|------|
| [../README.md](../README.md) | 研究问题、方法概要、代码从属角色、复现验证入口 |
| `docs/`（本页） | 全部研究文档的导航与阅读顺序 |
| `t_*`、`scripts/` | 执行 experiment / reports 中的步骤，非独立产品 |

根目录 [README](../README.md) 不含具体测试结论；结论只在 [reports/](./reports/)。
