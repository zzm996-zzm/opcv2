# Loot Drop 数据同步

Loot Drop MySQL 只是抓取源，不是 OPC 的业务数据库。`apps/lootdrop-import` 会把源库完整镜像到 OPC PostgreSQL 的 `lootdrop_*` 独立表，并保留每行 `source_payload`、来源 hash 和同步运行记录。

## 同步流程

1. 在 Loot Drop crawler 中完成 `crawl --full` 或 `crawl --incremental`。
2. 执行导入器。每次导入会对 `startups`、`rebuild_plans`、`ideas` 和其他数据集做幂等 upsert；源库已经删除的记录也会从镜像表删除。
3. `startups` 同步为项目超市真实案例库的只读投影，原始事实仍以 `lootdrop_startups` 为准。
4. 翻译器只读取 `lootdrop_translations` 中 pending 的记录。源 hash 变化时状态自动回到 pending，中文结果写入独立 JSON，不覆盖英文原文。

## 本地命令

```bash
GOPROXY=https://goproxy.cn,direct go run ./apps/lootdrop-import \
  -source-dsn 'lootdrop:lootdrop@tcp(127.0.0.1:3307)/lootdrop?parseTime=true&charset=utf8mb4' \
  -database-url 'postgres://opcv2:opcv2@127.0.0.1:5432/opcv2?sslmode=disable'
```

翻译密钥只通过环境变量传入，不要写入代码、`.env` 或提交记录：

```bash
OPCV2_DEEPSEEK_API_KEY='你的密钥' \
GOPROXY=https://goproxy.cn,direct \
go run ./apps/lootdrop-translate -batch-size 10
```

大字段案例可以使用 `-dataset startup -batch-size 2`；网络或模型返回异常时用 `-retry-failed` 重试。可用下面的查询查看同步和翻译状态：

```sql
SELECT id, source_run_id, status, imported_at
FROM lootdrop_sync_runs
ORDER BY id DESC LIMIT 10;

SELECT dataset, status, COUNT(*)
FROM lootdrop_translations
GROUP BY dataset, status
ORDER BY dataset, status;
```

线上部署只负责迁移和应用代码；线上导入应在能访问 Loot Drop MySQL 的环境执行同一个导入器，避免让线上 OPC 服务直接依赖源库连接。
