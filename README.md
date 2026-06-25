# 智活AI OPC v2

OPC v2 是一个面向一人公司和小 B 团队的商业行动工作台。当前已完成项目骨架、账号登录、会员权益、积分账本和兑换码基础。

## 本地开发

启动完整本地环境：

```bash
docker compose up --build
```

默认服务：

- Web: `http://127.0.0.1:3000`
- API: `http://127.0.0.1:8080`
- PostgreSQL: `127.0.0.1:5432`
- Redis: `127.0.0.1:6390`

常用检查：

```bash
make test
make lint
make build
```

## 服务器部署

服务器部署走 Docker Compose，结构参考旧版 `~/data/www/opc`：

- `docker-compose.server.yml`
- `.env.server.example`
- `scripts/deploy_server.sh`

首次部署：

```bash
cp .env.server.example .env.server
# 修改 .env.server 里的密码、JWT secret、端口和 Provider key
./deploy.sh deploy
```

更新部署：

```bash
./deploy.sh
```

本地一键触发服务器更新：

```bash
./update.sh
```

默认会执行 `ssh prod`，进入服务器 `~/data/www/opcv2` 后运行 `./deploy.sh`。如需覆盖：

```bash
DEPLOY_HOST=prod DEPLOY_PATH=~/data/www/opcv2 ./update.sh
```

查看状态和日志：

```bash
./deploy.sh ps
./deploy.sh logs
```

默认服务器端口：

- Web: `http://服务器IP:8681`
- API: Web 容器内反代 `/api`
- PostgreSQL: 容器内 `postgres:5432`，宿主机 `5433`
- Redis: 容器内 `redis:6379`，宿主机 `6391`

> MVP 阶段还未接真实短信服务商，`.env.server.example` 默认 `OPCV2_ENV=development` 和固定验证码。正式上线前必须接入真实 SMS Provider，再切到 `OPCV2_ENV=production`。

## 已完成

- 手机验证码登录/注册、Refresh Token 轮换、退出登录
- 会员套餐、积分账户、积分账本、兑换码幂等兑换
- 首页、登录页、会员中心、基础受保护路由
- PostgreSQL/Redis/Worker/API/Web 的 Compose 骨架
