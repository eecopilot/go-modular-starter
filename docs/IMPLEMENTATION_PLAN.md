# Go Modular Starter 实施规划

## 目标定位

`Go-Modular-Starter` 是一个面向任意业务的 Go Web Server starter。它不内置具体业务逻辑，而是提供一套稳定、可扩展、可替换的工程骨架，让新业务可以快速落地 HTTP API、认证授权、用户体系、数据库、缓存、配置、日志、测试和部署。

核心原则：

- 业务无关：starter 只提供通用工程能力和示例模块，不绑定具体行业模型。
- 模块化：每个业务能力以 module 方式注册路由、服务、仓储和生命周期钩子。
- 稳定优先：默认具备超时控制、优雅关闭、panic recovery、健康检查、依赖生命周期管理。
- 可替换：HTTP router、数据库、Redis、认证、日志实现都通过接口或边界层隔离。
- 可上线：本地开发、测试、迁移、Docker、环境变量和 CI 命令从第一版就可用。

## 当前依赖现状

当前已有两个独立模块：

- `github.com/eecopilot/userkit`
  - 用户、认证、RBAC、PostgreSQL repository、HTTP/Gin adapter。
  - starter 当前依赖版本：`v0.1.0`。
- `github.com/eecopilot/oauth`
  - OAuth 2.0 server、middleware、memory/Redis/PostgreSQL storage。

当前 starter 的默认集成优先级：

- 第一优先级：`userkit`。大多数业务 starter 首先需要注册、登录、JWT、RBAC、用户管理和 PostgreSQL 持久化。
- 第二优先级：稳定 Web Server 基础设施。认证之外，starter 必须先是一个能长期运行、能优雅关闭、能观测、能测试的服务。
- 可选扩展：`oauth`。当业务确实需要 OAuth2 授权服务器、第三方 client、token revocation/introspection 时再启用，不作为第一轮硬依赖。

需要先处理一个前置一致性问题：

- 本地 `oauth/go.mod` 当前 module path 仍是 `github.com/eep/oauth`，但 GitHub 仓库是 `github.com/eecopilot/oauth`。
- 如果后续在 starter 中启用 `oauth`，应先把 `oauth` module path 统一成 `github.com/eecopilot/oauth`，并打 tag。

## 当前实现状态

已完成：

- 稳定 `net/http` server 基线。
- HTTP timeout、request id、access log、panic recovery、安全 header、CORS、body limit。
- `SIGINT` / `SIGTERM` graceful shutdown。
- `/healthz`、`/readyz`、`/version`。
- 环境变量配置和 `.env.example`。
- PostgreSQL 本地 Docker Compose。
- `userkit` 可选集成，启用后提供注册、登录、`/me`、用户管理和 RBAC 路由。
- `userkit` 迁移 SQL 复制到 starter `migrations/`。
- 示例业务模块 `/api/v1/examples`。
- 受保护业务示例 `/api/v1/protected/example`，演示 Bearer Token 校验和当前用户读取。
- Dockerfile 和完整 app + migration + PostgreSQL compose profile。
- GitHub Actions CI：`go test ./...`、compose config、Docker build。

已验证：

- `go test ./...` 通过。
- `docker compose -f deployments/docker-compose.yml --profile app config` 通过。
- `make docker-app-up` 可以构建并启动完整栈。
- Docker 环境下 `/healthz`、`/readyz`、`/version` 返回 200。
- Docker 环境下 `userkit` 注册、登录、`/me` smoke test 通过。
- Docker 环境下受保护业务示例 smoke test 通过。
- Docker 环境下 example CRUD smoke test 通过。

## 目标目录结构

```text
Go-Modular-Starter/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── lifecycle.go
│   │   └── module.go
│   ├── bootstrap/
│   │   ├── bootstrap.go
│   │   └── dependencies.go
│   ├── config/
│   │   └── config.go
│   ├── httpserver/
│   │   ├── server.go
│   │   ├── middleware.go
│   │   ├── response.go
│   │   └── routes.go
│   ├── modules/
│   │   ├── health/
│   │   ├── auth/
│   │   └── example/
│   └── platform/
│       ├── logger/
│       ├── postgres/
│       ├── redis/
│       └── migration/
├── migrations/
├── configs/
├── deployments/
│   ├── Dockerfile
│   └── docker-compose.yml
├── scripts/
├── docs/
├── .env.example
├── Makefile
├── go.mod
└── README.md
```

## 分阶段实施

### Phase 0：仓库和依赖基线

目标：让 starter 作为独立业务模板仓库存在，并能稳定引用 `oauth` 和 `userkit`。

任务：

1. 初始化 `Go-Modular-Starter` Git 仓库。
2. 创建 `go.mod`，module 建议先用 `github.com/eecopilot/go-modular-starter`。
3. 给 `userkit` 打 `v0.1.0` tag，starter 第一轮先依赖它。
4. `oauth` 暂不进入 starter 第一轮依赖；后续需要时再修正 module path 并打 tag。
5. 在 workspace 根目录创建本地 `go.work`：

```bash
cd ~/workspace
go work init ./Go-Modular-Starter ./userkit
```

验收：

- `go test ./...` 在 starter 和 userkit 两个模块里都通过。
- starter 可以通过 `go.work` 使用本地 `userkit`。

### Phase 1：最小可运行 Web Server

目标：先跑起来一个稳定 HTTP 服务，而不是只搭目录。

能力：

- `/healthz`：进程活着即可返回 200。
- `/readyz`：依赖 PostgreSQL/Redis 可用时返回 200。
- `/version`：返回版本、commit、build time。
- 统一 JSON 响应格式。
- 统一错误响应格式。

Web Server 必备稳定性：

- `http.Server` 显式设置：
  - `ReadHeaderTimeout`
  - `ReadTimeout`
  - `WriteTimeout`
  - `IdleTimeout`
  - `MaxHeaderBytes`
- `context.Context` 贯穿启动和关闭。
- 监听 `SIGINT` / `SIGTERM`。
- 优雅关闭：
  - 停止接收新请求。
  - 给已有请求固定 shutdown timeout。
  - 关闭数据库、Redis、后台任务。
  - 超时后返回错误退出。
- middleware：
  - request id
  - access log
  - panic recovery
  - CORS
  - secure headers
  - body size limit

验收：

- `go run ./cmd/api` 可以启动。
- `curl localhost:8080/healthz` 返回 200。
- `Ctrl+C` 或 `SIGTERM` 能看到有序 shutdown 日志。

### Phase 2：配置系统

目标：所有上线相关参数都从配置进入，不把环境差异写死在代码里。

配置来源：

- 环境变量为主。
- `.env` 仅用于本地开发。
- `configs/config.example.yaml` 可选，用作说明，不作为强依赖。

配置项：

- app name / env / version
- HTTP address 和各类 timeout
- PostgreSQL DSN / pool 配置
- Redis address / password / DB / timeout
- OAuth token TTL / refresh rotation
- CORS allow origins
- log level / format
- shutdown timeout

验收：

- `.env.example` 覆盖所有必填配置。
- 缺少关键配置时启动失败，并输出清晰错误。

### Phase 3：平台层

目标：把基础设施接入集中管理，业务模块不直接散落初始化逻辑。

PostgreSQL：

- 使用 `database/sql` + `pgx` stdlib driver，方便直接接入 `userkit/postgres` 的 `*sql.DB` repository。
- 配置 max conns、min conns、max lifetime、idle time。
- 提供 `Ping(ctx)`。
- 关闭时调用 `Close()`。

Redis：

- 使用 `redis/go-redis/v9`。
- 配置 dial/read/write timeout。
- 提供 `Ping(ctx)`。
- 关闭时调用 `Close()`。

Logger：

- 默认使用标准库 `log/slog`。
- 支持 JSON/text 格式。
- request id 进入日志字段。

Migration：

- 提供 `make migrate-up`。
- 首版可用 SQL 文件和简单脚本，后续可切到 goose/atlas。

验收：

- 依赖不可用时 `/readyz` 返回非 200。
- 进程退出时平台资源全部关闭。

### Phase 4：模块系统

目标：业务能力用统一接口挂载，starter 能承载任意业务模块。

模块接口建议：

```go
type Module interface {
    Name() string
    RegisterRoutes(r Router)
}

type Lifecycle interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

模块约定：

- 每个 module 拥有自己的 handler/service/repository。
- module 不能直接读取全局变量。
- module 通过 dependencies 获取 logger、db、redis、auth。
- module 的路由统一挂载到 `/api/v1/...`。

首批模块：

- `health`：健康检查。
- `auth`：集成 `oauth` 和 `userkit`。
- `example`：一个最小 CRUD 示例，演示业务模块写法。

验收：

- 新增一个模块只需要实现接口并在 bootstrap 注册。
- example 模块有 handler/service/repository 分层示例。

### Phase 5：认证授权集成

目标：starter 默认带可用的用户体系和 JWT/RBAC 能力；OAuth2 授权服务器作为可选模块保留扩展位。

集成方式：

- `userkit` 负责用户、密码、角色、权限。
- starter 的 `auth` module 默认使用 `userkit` 的 JWT Bearer Token 和 RBAC。
- `oauth` 只有在业务需要 OAuth2 client、授权码、refresh token rotation、revocation/introspection 时接入。

首版 API：

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/me`
- `GET /api/v1/users`
- `GET /api/v1/roles`
- `GET /api/v1/permissions`

权限模型：

- middleware 读取 bearer token。
- RBAC 检查用于后台管理类业务。
- scope 检查留给 OAuth 可选模块。

验收：

- 可以注册用户。
- 可以登录并拿到 token。
- 可以访问受保护接口。
- token 过期、scope 不足、无权限时返回稳定错误格式。

### Phase 6：测试体系

目标：starter 不是 demo，而是后续业务可长期维护的基线。

测试层级：

- unit test：service、config、middleware。
- handler test：使用 `httptest`。
- integration test：PostgreSQL/Redis 可用时运行。
- smoke test：启动 server 后请求 health/login/protected endpoint。

命令：

```bash
make test
make test-integration
make lint
make e2e
```

验收：

- `go test ./...` 通过。
- 没有数据库时普通单测仍可跑。
- 集成测试通过环境变量显式开启。

### Phase 7：开发体验

目标：任何新业务复制这个 starter 后，10 分钟内能跑起来。

交付物：

- `README.md`：快速开始、目录说明、模块开发指南。
- `.env.example`
- `Makefile`
- `docker-compose.yml`：PostgreSQL + Redis。
- `Dockerfile`
- `scripts/dev.sh`
- `scripts/migrate.sh`

常用命令：

```bash
make dev
make test
make migrate-up
make docker-up
make docker-down
```

验收：

- 新机器按 README 可以启动服务。
- Docker Compose 可以拉起依赖。

### Phase 8：生产化收口

目标：作为稳定 starter，默认行为接近生产可用。

能力：

- graceful shutdown 完整验证。
- readiness 区分依赖故障。
- request timeout。
- slow request log。
- panic recovery 不泄漏堆栈给客户端。
- JSON access log。
- build info 注入。
- 基础安全 header。
- CORS 默认保守。
- Docker 镜像使用非 root 用户。
- CI 运行 `go test ./...`。

验收：

- 发送 `SIGTERM` 后，服务在 shutdown timeout 内退出。
- 未完成请求能在上下文取消时结束。
- Docker 镜像可运行。

## 推荐实施顺序

1. 初始化 starter 的 `go.mod`、README、Makefile、`.env.example`。
2. 实现稳定 HTTP server 和 graceful shutdown。
3. 接入配置、logger、health/readiness。
4. 接入 PostgreSQL 平台层。
5. 接入 `userkit` 的 auth/user/RBAC module。
6. 实现 module 注册机制。
7. 增加 example CRUD module。
8. 补齐测试、Docker、迁移和 CI。
9. 视业务需要再接入 `oauth` 可选模块。

## 第一轮可交付范围

第一轮先做到“能稳定启动和关闭”：

- `cmd/api/main.go`
- `internal/config`
- `internal/httpserver`
- `internal/platform/logger`
- `internal/modules/health`
- `.env.example`
- `Makefile`
- `README.md`
- `go test ./...`

完成第一轮后，再进入认证、数据库和业务模块。
