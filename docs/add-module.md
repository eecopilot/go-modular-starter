# Add A Business Module

业务模块放在 starter 里，不放在 `userkit` 里。`userkit` 只负责用户、登录、JWT 和 RBAC；订单、项目、账单、报表这类业务能力都应该作为 `internal/modules/<name>` 下的模块接入。

## 1. 创建目录

以 `orders` 为例：

```text
internal/modules/orders/
├── handler.go
└── handler_test.go
```

复杂业务可以继续拆成：

```text
internal/modules/orders/
├── handler.go
├── service.go
├── repository.go
├── model.go
└── handler_test.go
```

## 2. 写公开接口

不需要登录的模块可以直接注册路由：

```go
package orders

import (
	"net/http"

	"github.com/eecopilot/go-modular-starter/internal/httpserver"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/orders", h.List)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{"orders": []any{}})
}
```

然后在 `internal/bootstrap/bootstrap.go` 注册：

```go
orders.New().RegisterRoutes(mux)
```

## 3. 写受保护接口

需要登录的业务模块接收一个 auth middleware，并在 handler 里读取当前用户：

```go
package orders

import (
	"net/http"

	"github.com/eecopilot/go-modular-starter/internal/httpserver"
	"github.com/eecopilot/go-modular-starter/internal/modules/auth"
)

type AuthMiddleware func(http.Handler) http.Handler

type Handler struct {
	auth AuthMiddleware
}

func New(auth AuthMiddleware) *Handler {
	return &Handler{auth: auth}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/orders", h.auth(http.HandlerFunc(h.List)))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	current, ok := auth.CurrentUser(r)
	if !ok {
		httpserver.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, map[string]any{
		"owner_id": current.ID,
		"orders":   []any{},
	})
}
```

在 `internal/bootstrap/bootstrap.go` 里，放到 `USERKIT_ENABLED=true` 分支内注册：

```go
authModule := auth.New(service)
authModule.RegisterRoutes(mux)
orders.New(authModule.RequireUser).RegisterRoutes(mux)
```

`internal/modules/protectedexample` 是完整可运行参考。

## 4. 加测试

受保护模块测试不需要真的签 JWT。直接把用户放进 request context：

```go
func withUser(user userkit.User) AuthMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(auth.ContextWithUser(r.Context(), user)))
		})
	}
}
```

测试至少覆盖：

- 未登录返回 `401`。
- 登录后能读取当前用户 ID。
- 业务数据只返回当前用户有权访问的资源。

## 5. 接数据库

需要表结构时，在 `migrations/` 增加 SQL 文件：

```text
migrations/003_create_orders.sql
```

本地执行：

```bash
make migrate-up
```

迁移脚本应该尽量可重复执行，例如使用 `CREATE TABLE IF NOT EXISTS` 和 `ON CONFLICT DO NOTHING`。

## 6. 补文档和 smoke test

新增模块后同步更新：

- `README.md`：列出主要路由和最短 curl 示例。
- `scripts/smoke-http.sh`：如果是核心流程，加入 smoke test。
- `docs/IMPLEMENTATION_PLAN.md`：如果属于 starter 基础能力，更新状态。

最后执行：

```bash
go test ./...
make docker-app-up
make smoke-docker
make docker-down
```
