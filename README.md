# Go Modular Starter

一个面向通用业务的 Go Web Server starter。当前第一轮重点是稳定的 HTTP 服务基线：配置、日志、健康检查、超时控制、中间件和优雅关闭。

## 当前能力

- 标准库 `net/http` Web Server。
- 显式 HTTP 超时配置。
- `SIGINT` / `SIGTERM` 优雅关闭。
- `/healthz`、`/readyz`、`/version`。
- request id、access log、panic recovery。
- CORS、安全响应头、请求体大小限制。
- 环境变量配置和 `.env.example`。
- 可选 `userkit` 集成，启用后提供 `/api/v1/auth/*`、`/api/v1/me`、用户管理和 RBAC 路由。
- 示例业务模块 `/api/v1/examples`，演示模块化 CRUD 写法。
- `web/` 前端目录，当前用 `go:embed` 将 `web/dist` 静态资源嵌入 Go 二进制。
- Dockerfile 和 Docker Compose，支持依赖容器化或完整 app + PostgreSQL 启动。

## 快速开始

```bash
go test ./...
go run ./cmd/api
```

验证：

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
curl http://localhost:8080/version
curl http://localhost:8080/
```

示例业务模块：

```bash
curl -X POST http://localhost:8080/api/v1/examples \
  -H 'Content-Type: application/json' \
  -d '{"name":"First item","description":"Demo item"}'

curl http://localhost:8080/api/v1/examples
```

## 规划

实施路线见 [docs/IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md)。

当前方向：

- 优先接入 `github.com/eecopilot/userkit`，提供注册、登录、JWT 和 RBAC。
- `github.com/eecopilot/oauth` 暂作为可选扩展，业务需要 OAuth2 授权服务器时再启用。

## 前端

当前前端只放 starter shell，不写业务页面。

```text
web/
├── embed.go       # go:embed 入口
├── dist/          # 当前被嵌入和服务的静态资源
└── README.md
```

后端启动后会把 `web/dist` 嵌入到 Go 二进制，并从 `/` 服务。API 路由仍然走 `/api/v1/...`、`/healthz`、`/readyz`、`/version`。

如果以后前端变成完整 SPA，可以加 `web/src` 和 Vite/React/Vue 等构建工具，仍然输出到 `web/dist`，继续由 Go 侧 `go:embed` 打包。

## Userkit

默认 `USERKIT_ENABLED=false`，starter 不要求本地必须有数据库。

开发时建议本地跑 app、Docker 跑 PostgreSQL。先启动数据库并执行迁移：

```bash
make docker-up
make migrate-up
```

然后启用 `userkit` 启动 app：

```bash
USERKIT_ENABLED=true \
USERKIT_DATABASE_URL='postgres://starter:starter@localhost:55432/starter?sslmode=disable' \
USERKIT_JWT_SECRET='change-me-to-a-long-random-secret' \
go run ./cmd/api
```

启用后的主要路由挂载在 `/api/v1`：

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/me
GET  /api/v1/users
GET  /api/v1/roles
GET  /api/v1/permissions
```

## 示例业务模块

`internal/modules/example` 是给业务模块开发用的参考实现：

- handler/store 分离。
- 标准库 `ServeMux` 路由。
- JSON 请求和响应。
- CRUD 测试。

真实业务可以复制这个模块结构，再替换为自己的 repository、service 和 handler。

## Docker

有两种 Docker 用法。

### 只跑依赖

适合日常开发：应用用 `go run` 跑在本机，PostgreSQL 跑在 Docker。

```bash
make docker-up
make migrate-up
```

端口：

- PostgreSQL：`localhost:55432`
- DSN：`postgres://starter:starter@localhost:55432/starter?sslmode=disable`

启用用户模块运行本地 app：

```bash
USERKIT_ENABLED=true \
USERKIT_DATABASE_URL='postgres://starter:starter@localhost:55432/starter?sslmode=disable' \
USERKIT_JWT_SECRET='change-me-to-a-long-random-secret' \
go run ./cmd/api
```

### 跑完整栈

适合验证容器化部署：Docker Compose 会启动 PostgreSQL、执行迁移、构建并启动 app。

```bash
make docker-app-up
```

端口：

- App：`http://localhost:8080`
- PostgreSQL：`localhost:55432`

验证基础服务：

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

一键 smoke test：

```bash
make smoke-docker
```

验证 `userkit` 注册和登录：

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","username":"demo","password":"password123","full_name":"Demo User"}'

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"demo@example.com","password":"password123"}'
```

查看日志：

```bash
docker compose -f deployments/docker-compose.yml --profile app logs -f app
```

停止服务：

```bash
make docker-down
```

重新构建 app 镜像：

```bash
make docker-build
```

## 常用命令

```bash
make dev
make test
make build
make docker-build
make docker-up
make docker-app-up
make migrate-up
make smoke-docker
```
