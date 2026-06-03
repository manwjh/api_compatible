# AGENTS.md — user-side 协作规则

面向维护者与 AI Agent：修改启动器、站点登记或 Runner 脚本前先读本文。研究仓库总规则见 [根 AGENTS.md](../../AGENTS.md)，实验目录索引见 [experiment/README.md](../README.md)。

## 定位

**user-side** 对应 [EC2-用户侧隔离实验点设计](../../docs/experiment/EC2-用户侧隔离实验点设计.md)。工作目录为 **`experiment/user-side/`**（`MAAS_ROOT` = 本目录）。

## 目录职责

| 路径 | 是否提交 Git | 说明 |
|------|--------------|------|
| `t_*` | ✅ | 兼容性自动化入口；逻辑在 `lib/maas.sh` |
| `scripts/run-user-side-compat.sh` | ✅ | L2 probe + 可选 `t_*` smoke |
| `lib/maas.sh`, `lib/maas.py` | ✅ | 共享启动、站点解析、临时配置生成 |
| `sites.json` | ✅ | 上游站点注册表（无密钥） |
| `.env.example` | ✅ | 密钥环境变量模板 |
| `opencode.json.example` | ✅ | OpenCode 手工配置示例（无密钥） |
| `.env` | ❌ | 本地 API Key |
| `.claude/` | ❌ | Claude Code 本地配置（含密钥） |
| `.runtime/` | ❌ | 启动器生成的临时 Agent 配置 |
| `opencode.json` | ❌ | 本地 OpenCode 配置（若含密钥） |

## 启动器约定

**部署位置**：E4 评估在 **用户侧 EC2 Runner** 上 `cd experiment/user-side` 后执行 `./scripts/run-user-side-compat.sh` 与 `./t_*`；[中转站原型 EC2](../../docs/experiment/EC2-中转站原型实验点设计.md) **不** 作为 `t_*` 的常规运行环境。

- **`./t_claude`**：Anthropic Messages；缺 CLI 时自动安装；写入 `.claude/settings.json`
- **`./t_codex`**：OpenAI Responses；缺 CLI 时自动安装；写入 `.runtime/codex.<site>.toml`
- **`./t_opencode`**：Chat Completions；缺 CLI 时自动安装；写入 `.runtime/opencode.<site>.json`

三者共用参数：`--site`、`--model`、`-y`、`--list-sites`、`--list-models`；其余参数透传给底层 CLI。

**不要**让启动器依赖 `upstream/` 参考源码（用 `../../upstream/pull.sh` 按需拉取）。

## 新增上游站点

1. 在 `sites.json` → `sites` 增加条目（`base_url`、`anthropic_base_url`、`api_key_env`、`default_models`、可选 `opencode` 块）
2. 在 `.env.example` 增加对应 `*_API_KEY` 占位
3. 用户本地 `.env` 填真实 Key
4. 用 `./t_*` 跑通后，在 `docs/reports/` 新增或更新评估报告

## 新增 Agent 启动器

1. 新增 `t_<name>`（source `lib/maas.sh`，调用 `maas_run_<name>`）
2. 在 `maas.sh` 实现：参数解析、CLI 检测/安装、环境变量与临时配置、`exec` 启动
3. 在 `maas.py` 按需增加 `get` / `write-*-config` 子命令
4. 在 `docs/reports/` 新增 `*兼容性评估报告.md`

## 安全与 Git

- **禁止**提交 API Key、`.env`、`.claude/`、`.runtime/`、含密钥的 `opencode.json`
- `sites.json` 只引用 `api_key_env` 名称，不写密钥值
- 评估报告中的 Key 应打码或提示轮换

## 文档同步

改布局或启动器行为时，同步更新：

- [user-side/README.md](./README.md)
- [experiment/README.md](../README.md)
- [根 README.md § 复现验证](../../README.md#复现验证从属章节)
- [docs/experiment/EC2-用户侧隔离实验点设计.md](../../docs/experiment/EC2-用户侧隔离实验点设计.md)
- [docs/reports/README.md](../../docs/reports/README.md)（若复现命令变更）
