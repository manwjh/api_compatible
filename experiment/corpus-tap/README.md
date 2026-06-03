# Corpus Tap

New API 中转站 **语料全量采集**插件（透明反向代理）。属于 [`experiment/`](../README.md) 附件，设计契约见 [中转站语料采集插件设计](../../docs/experiment/中转站语料采集插件设计.md)。

## 快速开始（本地）

```bash
cd experiment/corpus-tap
cp .env.example .env
# 编辑 CORPUS_TAP_UPSTREAM 指向本地或远程 New API

export $(grep -v '^#' .env | xargs)
go run ./cmd/corpus-tap
```

另开终端，经 Tap 访问（需设置 `CORPUS_TAP_DEV_USER_ID` 对应实验用户）：

```bash
curl -sS -H "Authorization: Bearer $NEWAPI_PROTOTYPE_TOKEN" \
  -H "Content-Type: application/json" \
  http://127.0.0.1:8443/v1/models
```

采集 POST 推理请求后，在 `CORPUS_TAP_LOCAL_DATA_DIR` 下按 `user_id=<n>/dt=.../<exchange_id>/` 出现 `client_request.json.gz` 等文件。

## 环境变量

| 变量 | 说明 |
|------|------|
| `CORPUS_TAP_UPSTREAM` | **必填** New API 根 URL |
| `CORPUS_TAP_DEV_USER_ID` | 骨架期：将所有 Bearer 请求记到该 `user_id` |
| `CORPUS_TAP_DATABASE_URL` | PostgreSQL；不填则仅写本地文件 |
| `CORPUS_TAP_LOCAL_DATA_DIR` | 本地语料目录（未配 S3 时默认 `./data`） |
| `CORPUS_TAP_MODE=proxy-only` | 只转发，不落库 |

## Docker Compose

见 [deploy/docker-compose.snippet.yml](./deploy/docker-compose.snippet.yml)，合并到中转站原型 `docker compose` 后，用户侧 `base_url` 指向 **8443**。

## 骨架局限（待实现）

- [ ] New API MySQL 只读：`tokens` → `user_id` / `token_id`
- [ ] S3 上传（当前仅 `file://` 本地布局）
- [ ] 大体积 SSE 流式零拷贝 tee（当前整包缓冲）
- [ ] `readyz` 校验 DB + 存储

## 迁移

```bash
psql "$CORPUS_TAP_DATABASE_URL" -f migrations/001_init.sql
```

Compose 使用 `corpus-db` 服务时已自动执行 `001_init.sql`。
