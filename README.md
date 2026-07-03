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
- 受保护业务示例 `/api/v1/protected/example`，演示登录后读取当前用户。
- `web/` 前端目录，当前用 `go:embed` 将 `web/dist` 静态资源嵌入 Go 二进制。
- Dockerfile 和 Docker Compose，支持依赖容器化或完整 app + PostgreSQL 启动。

## 5 分钟跑起来

```bash
git clone https://github.com/eecopilot/go-modular-starter.git
cd go-modular-starter
cp .env.example .env
make dev-up
```

`make dev-up` 会启动 PostgreSQL、执行迁移，并以前台方式启动 Go app。默认会启用 `userkit`，所以注册、登录和受保护业务接口都可以直接验证。

另开一个终端执行：

```bash
make smoke-local
```

<details>
<summary><strong>启动后首页</strong></summary>

![Go Modular Starter running at localhost:8080](docs/assets/starter.png)

</details>

停止 app 用 `Ctrl+C`，停止本地 PostgreSQL 用：

```bash
make dev-down
```

如果只想跑不带数据库和用户模块的空服务：

```bash
go test ./...
make dev
```

基础验证：

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
- 新增业务模块见 [docs/add-module.md](docs/add-module.md)。

## 前端

当前仓库默认只放 starter shell，但已经支持把任意可静态构建的开源前端作为子项目接进来。

```text
web/
├── app/           # 可选：外部前端子项目，默认不提交到 starter 仓库
├── embed.go       # go:embed 入口
├── dist/          # 当前被嵌入和服务的静态资源，导入前端后会被替换
└── README.md
```

后端启动后会把 `web/dist` 嵌入到 Go 二进制，并从 `/` 服务。API 路由仍然走 `/api/v1/...`、`/healthz`、`/readyz`、`/version`。

接入外部前端的最短路径：

```bash
cp .env.example .env
# 编辑 .env，通常只需要填 FRONTEND_GIT_URL / FRONTEND_GIT_REF
make frontend-import
make dev
```

常见配置：

```env
FRONTEND_GIT_URL=https://github.com/example/frontend.git
FRONTEND_GIT_REF=main
FRONTEND_APP_DIR=web/app
FRONTEND_DIST_DIR=
FRONTEND_OUTPUT_DIR=web/dist
```

`make frontend-import` 会克隆前端项目、安装依赖、运行构建，自动查找 `dist`、`build`、`out`、`.output/public` 这类常见产物目录，并把构建产物同步到 `web/dist`。如果某个项目输出目录特殊，再手动设置 `FRONTEND_DIST_DIR`。前端项目需要能输出包含 `index.html` 的静态目录；Vite/React/Vue 这类 SPA 通常可以直接用，Next.js/Nuxt/SvelteKit 这类项目需要先配置静态导出。前端调用后端 API 时建议使用同源路径 `/api/v1`。

<details>
<summary><strong>Userkit</strong></summary>

默认 `USERKIT_ENABLED=false`，starter 不要求本地必须有数据库。

开发时建议本地跑 app、Docker 跑 PostgreSQL。最短路径：

```bash
cp .env.example .env
make dev-up
```

手动分步也可以：

```bash
make docker-up
make migrate-up
make dev USERKIT_ENABLED=true
```

启用后的主要路由挂载在 `/api/v1`：

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/me
GET  /api/v1/users
GET  /api/v1/roles
GET  /api/v1/permissions
GET  /api/v1/protected/example
```

</details>

<details>
<summary><strong>示例业务模块</strong></summary>

`internal/modules/example` 是给业务模块开发用的参考实现：

- handler/store 分离。
- 标准库 `ServeMux` 路由。
- JSON 请求和响应。
- CRUD 测试。

真实业务可以复制这个模块结构，再替换为自己的 repository、service 和 handler。

`internal/modules/protectedexample` 演示登录后访问业务接口：

- 路由只在 `USERKIT_ENABLED=true` 时注册。
- 通过 `auth.RequireUser` 校验 `Authorization: Bearer <token>`。
- handler 通过 `auth.CurrentUser(r)` 读取当前用户，再用用户 ID 限定业务数据归属。

示例：

```bash
TOKEN="login-or-register-response-token"

curl http://localhost:8080/api/v1/protected/example \
  -H "Authorization: Bearer ${TOKEN}"
```

新增业务模块的详细步骤见 [docs/add-module.md](docs/add-module.md)。

</details>

## Docker

有两种 Docker 用法。

### 只启动 PostgreSQL

适合日常开发：数据库跑在 Docker，Go app 跑在本机。`make docker-up` 只启动 PostgreSQL；`make migrate-up` 会用一次性迁移容器执行 SQL，执行完就退出。

```bash
make docker-up
make migrate-up
```

端口：

- PostgreSQL：`localhost:55432`
- DSN：`postgres://starter:starter@localhost:55432/starter?sslmode=disable`

启用用户模块运行本地 app：

```bash
make dev USERKIT_ENABLED=true
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
make dev-up
make dev-down
make test
make build
make docker-build
make docker-up
make docker-app-up
make migrate-up
make smoke-local
make smoke-docker
```
