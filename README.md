# Social Media Architecture Benchmark MVP

面向架构主链路 QPS 验证的最小可运行实现（单机 Docker Compose）。

## 架构链路

Client -> OpenResty(Lua JWT+黑名单) -> Java Spring Boot Web Node -> gRPC -> Go Core Service -> Valkey/MongoDB/Redpanda(兼容 Kafka) -> Go Consumer 批量异步落库。

## 目录

- `gateway-openresty/`: 网关 + Lua 鉴权
- `web-node-java/`: Java 21 + Spring Boot 3 Web Node（只走 gRPC 下游）
- `core-service-go/`: Go gRPC 核心服务（缓存/持久化/事件生产）
- `consumer-go/`: 事件消费者（200ms 或满批次 flush）
- `seeder-go/`: 自动 seed 数据
- `proto/`: gRPC 协议
- `bench/k6/`: k6 压测脚本
- `scripts/`: 一键脚本
- `artifacts/`: 压测输出

## 快速开始（测试机可直接复制）

```bash
cp .env.example .env

docker compose up -d --build
./scripts/wait_for_stack.sh
./scripts/seed_data.sh

./scripts/run_benchmark.sh smoke
./scripts/run_benchmark.sh mixed
./scripts/find_max_qps.sh

ls -R artifacts/bench
```

## 接口

- `POST /api/v1/posts`（JWT）
- `GET /api/v1/posts/{postId}`（cache->db->cache）
- `GET /api/v1/feed/hot?limit=50`
- `POST /api/v1/posts/{postId}/like`（JWT，like 先写 Valkey，再异步事件落库）
- `GET /healthz`
- `GET /readyz`

## 压测场景

通过 `./scripts/run_benchmark.sh <scenario>`：

- `smoke`: 链路联通 + 鉴权 + 读写冒烟
- `read-heavy`: 读为主（hot feed + post detail）
- `mixed`: 默认 75% hot, 15% detail, 5% create, 5% like

参数由环境变量控制：

- `TARGET_QPS`
- `TEST_DURATION`
- `VUS_MAX`
- `HOT_FEED_LIMIT`
- `FAIL_ERROR_RATE`
- `FAIL_P95_MS`
- `FAIL_P99_MS`

## 自动找上限

`./scripts/find_max_qps.sh` 默认逐轮提升 QPS，若触发阈值（k6 thresholds）则停止。

可调参数：

- `START_QPS`
- `STEP_QPS`
- `MAX_ROUNDS`
- `TEST_DURATION`

## 输出结果

每轮压测输出到：

- `artifacts/bench/<timestamp>/<scenario>/summary.json`
- `artifacts/bench/<timestamp>/<scenario>/stdout.txt`
- `artifacts/bench/<timestamp>/<scenario>/report.md`

seed 输出：

- `artifacts/seed/post_ids.txt`

## MVP 简化说明

- JWT 仅 HS256。
- 黑名单仅 Valkey key 检查（`jwt:blacklist:<token>`）。
- Hot feed 使用 Valkey ZSET 简化热度分数。
- 无推荐系统、无行为分析、无 ClickHouse、无外部云依赖。
- “发帖不可立即读”语义通过 `PENDING -> READY`（Consumer 异步确认）实现。

## 常用命令

```bash
# 启停

docker compose up -d --build
docker compose ps
docker compose logs -f gateway web-node core-service consumer

docker compose down -v

# 手工 smoke
curl -sS http://localhost:8088/healthz
curl -sS http://localhost:8088/readyz
```
