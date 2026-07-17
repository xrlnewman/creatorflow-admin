# CreatorFlow Admin

CreatorFlow 是免费开源的内容排期与创作者协同运营后台，覆盖选题队列、编辑排班、创作者档案和复盘任务。它只做运营流程，不做诊断、处方、支付或真实创作者数据。

## 闭环流程

1. 前台创建选题（`待签到`）。
2. 创作者签到后进入 `已签到`，运营人员推进到 `候诊中`、`制作中`。
3. 服务完成后进入 `已完成`；每次状态变化写入 `appointment_events`，可从详情页回放。
4. 运营人员创建复盘任务，执行完成后由 `待完成` 变为 `已完成`。
5. 所有写请求要求 `Idempotency-Key`，Redis 负责幂等结果和并发锁，MySQL 8.4 负责持久化。

```bash
# 一键启动 API + MySQL 8.4 + Redis 8（会自动加载合成演示数据）
docker compose -f deploy/docker-compose.yml up --build

# 或仅使用无外部依赖的内存模式运行 API
go run ./server

# 管理后台
cd web && npm install && npm run dev
```

前端默认请求 `/api/v1`，Vite 开发服务器会把 `/api` 和 `/healthz` 代理到 `http://localhost:8080`。部署到独立域名时，可在构建时设置 `VITE_API_BASE_URL=https://api.example.com`；客户端会自动补齐 `/api/v1`，所有创建、签到、状态推进和复盘完成请求都会自动生成 `Idempotency-Key`。

后台的“选题队列”和“复盘任务”按钮会优先调用真实 API；API 暂不可用时保留内置演示数据并提示当前数据来源。侧栏“移动端体验”提供同一闭环的窄屏创作者视图，支持创建演示选题、签到、候诊、制作完成和复盘完成，便于用手机浏览器联调。

## 闭环 API 示例

```bash
# 创建选题（重复发送相同 Idempotency-Key 只会创建一次）
curl -X POST http://localhost:8080/api/v1/appointments \
  -H 'Content-Type: application/json' -H 'Idempotency-Key: demo-create-001' \
  -d '{"patient":"演示创作者","department":"全科门诊","doctor":"林编辑","scheduledAt":"2026-07-16T09:00:00+08:00"}'

# 推进状态：待签到 -> 已签到 -> 候诊中 -> 制作中 -> 已完成（将 AP-1001 替换为上一步返回的 id）
curl -X POST http://localhost:8080/api/v1/appointments/AP-1001/checkin -H 'Idempotency-Key: demo-checkin-001'
curl -X POST http://localhost:8080/api/v1/appointments/AP-1001/status \
  -H 'Content-Type: application/json' -H 'Idempotency-Key: demo-waiting-001' -d '{"status":"候诊中"}'

# 查看审计事件
curl http://localhost:8080/api/v1/appointments/AP-1001/events

# 完成复盘
curl -X POST http://localhost:8080/api/v1/followups/FW-0716-001/complete -H 'Idempotency-Key: demo-followup-001'
```

演示数据均为虚构数据；项目不得用于真实医疗诊断、处方、支付或创作者隐私存储。
