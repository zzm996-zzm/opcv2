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

前端默认通过同源 `/api` 访问后端。需要让前端直接访问另一个 API 域名时，在构建前设置：

```bash
VITE_API_BASE_URL=https://api.example.com
```

本地 Vite dev server 仍可继续用 `VITE_API_PROXY_TARGET` 配置 `/api` 代理目标。

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
> 竞品扫描 worker 默认不会生成开发样例；本地联调如需生成演示结果，可显式设置 `OPCV2_COMPETITOR_SCANNER_PROVIDER=development`。该配置在 production 环境会被拒绝。

### 国内服务器构建慢

服务器 compose 默认使用国内可访问性更好的镜像源和包源：

- Docker 基础镜像：`docker.m.daocloud.io/library/...`
- Alpine apk：`https://mirrors.aliyun.com/alpine`
- Go modules：`https://goproxy.cn,direct`
- npm：`https://registry.npmmirror.com`

如果你的服务器已经配置了 Docker Hub 加速器，或部署在海外，可以在 `.env.server` 中把这些变量改回官方源：

```bash
OPCV2_POSTGRES_IMAGE=postgres:16-alpine
OPCV2_REDIS_IMAGE=redis:7-alpine
OPCV2_GO_BUILDER_IMAGE=golang:1.22.3-alpine
OPCV2_RUNTIME_IMAGE=alpine:3.20
OPCV2_NODE_BUILDER_IMAGE=node:22-alpine
OPCV2_NGINX_IMAGE=nginx:1.27-alpine
OPCV2_ALPINE_REPOSITORY=
OPCV2_GOPROXY=https://proxy.golang.org,direct
OPCV2_NPM_REGISTRY=https://registry.npmjs.org/
```

## 已完成

- 手机验证码登录/注册、Refresh Token 轮换、退出登录
- 会员套餐、积分账户、积分账本、兑换码幂等兑换
- 首页、登录页、会员中心、基础受保护路由
- PostgreSQL/Redis/Worker/API/Web 的 Compose 骨架
