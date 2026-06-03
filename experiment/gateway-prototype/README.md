# gateway-prototype — 中转站原型实验点

对应 [EC2-中转站原型实验点设计](../../docs/experiment/EC2-中转站原型实验点设计.md)：境外 EC2 上 New API + Channel + 平台 Token 交付用户侧。

## 状态

**占位目录** — 设计稿已完成，可复现资产（Compose、Channel 模板、部署脚本）待按设计 §9–§10 补充。

## 关联实现

| 组件 | 位置 |
|------|------|
| 语料采集插件（可选） | [../corpus-tap/](../corpus-tap/) |
| New API 参考源码 | [../../upstream/newapi/](../../upstream/newapi/)（`upstream/pull.sh newapi`） |
| 用户侧验证 | [../user-side/](../user-side/)（登记 `sites.json` 后 probe / `t_*`） |

## 预期内容（待添加）

- `docker-compose.yml` 或片段（New API + MySQL + Redis）
- Channel / 用户 / Token 初始化说明或脚本
- 与 [语料插件](../corpus-tap/deploy/docker-compose.snippet.yml) 的合并说明
