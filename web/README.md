# Web 前端目录

这个目录放 Go app 对外服务的前端静态资源。

当前约定：

- `dist/` 放最终静态产物，会通过 `go:embed` 打进 Go 二进制。
- `app/` 预留给可选的外部前端子项目，可以来自 GitHub 或其他 git remote。
- 默认 starter 不要求 Node；只有导入真实前端项目时才需要 Node 工具链。

## 导入前端子项目

选择一个能构建出 `index.html` 静态产物的开源前端项目，然后配置 `.env`：

```env
FRONTEND_GIT_URL=https://github.com/example/frontend.git
FRONTEND_GIT_REF=main
FRONTEND_APP_DIR=web/app
FRONTEND_DIST_DIR=
FRONTEND_OUTPUT_DIR=web/dist
```

执行：

```bash
make frontend-import
make dev
```

`make frontend-import` 会把项目克隆到 `web/app`，安装依赖，运行前端构建，自动识别 `dist`、`build`、`out` 这类常见产物目录，并复制到 `web/dist`。Go 服务仍然从 `/` 服务 `web/dist`。

常用覆盖配置：

- `FRONTEND_BUILD_COMMAND='VITE_API_BASE_URL=/api/v1 npm run build'`：给 Vite 这类项目传入后端 API 地址。
- `FRONTEND_DIST_DIR=custom-dist`：项目输出目录很特殊时手动指定。
- `FRONTEND_SKIP_INSTALL=true make frontend-build`：本地依赖已经安装好时，只重新构建并同步产物。

如果前端是 Next.js、Nuxt、SvelteKit 这类项目，需要先配置静态导出，再把 `FRONTEND_DIST_DIR` 指到导出的目录。
