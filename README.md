# 墨衡：杂志社投稿、审稿与刊发协作平台

墨衡把作者投稿、编辑分派、审稿意见、主编终审、刊期编排和公开刊发连接成一条可审计的工作流。后端是 Go 1.24+ 服务，默认可以用内存适配器离线演示，也可以切换到 PostgreSQL；前端是 Vue 3 + TypeScript + Vite + Pinia + Element Plus 的编辑部工作台。

## 架构与目录

- `cmd/server`：HTTP 服务进程、信号处理和优雅停机。
- `internal/domain`：identity、manuscript、review、publication、audit 四个业务边界；状态机、值对象和不变量都在领域层。
- `internal/application`：注册登录、投稿快照、分派回避、审稿、终审、刊期、撤回、报表和催办 worker；所有跨聚合写操作经过事务端口。
- `internal/repository/memory`：确定性的离线演示和单元测试存储。
- `internal/repository/postgres`：pgx 实现，使用 JSONB 聚合快照、唯一约束、外键、检查约束、索引和乐观版本。
- `internal/transport/http`、`internal/middleware`：`/api/v1` Gin API、request_id、错误码、权限、限流、安全响应头和超时。
- `internal/platform`：可替换时钟、ID、幂等、文件存储和 outbox。
- `migrations`：可重复执行的 PostgreSQL 迁移与演示种子。
- `api/openapi`：OpenAPI 3.0 接口文档。
- `deploy`：生产 Compose 叠加配置、滚动更新和日志保留策略。
- `web`：编辑工作台、公开刊发检索、状态筛选、分页、时区展示与响应式布局。
- `tests/integration`：可选 PostgreSQL 连接集成检查。

## 核心状态机

稿件从 `draft` 进入 `submitted`，经过 `under_initial_review`、`revision_needed`/`under_re_review` 和 `under_final_review`，再进入 `accepted`、`scheduled`、`published` 或带原因的 `rejected`/`withdrawn`。投稿会锁定正文、附件和声明快照；录用后的修改必须标记为勘误版本。主编决策始终携带当前稿件版本和已完成的审稿意见。

刊期从 `planning` 锁定到 `locked`，发布为 `released` 时，刊期内文章和稿件在同一事务内转为 `published`。撤回必须提供原因，并追加审计事件，公开检索只返回仍为 `published` 的文章。

## 本地运行

需要 Go 1.24+。默认 `STORAGE_MODE=memory`，不依赖数据库：

```bash
go run ./cmd/server
```

访问 `http://localhost:8080/healthz` 和 `http://localhost:8080/readyz`。前端开发服务器：

```bash
cd web
npm install
npm run dev
```

前端会把 `/api`、`/healthz` 和 `/readyz` 代理到 Go 服务。工作台使用 `X-Actor-ID` 作为本地演示身份，默认是 `demo-editor`；实际部署应接入短期访问令牌与可撤销刷新令牌。

## PostgreSQL 与容器

复制 `.env.example` 为 `.env`，设置 `STORAGE_MODE=postgres` 和 `DATABASE_URL`，再执行迁移文件。也可以直接运行：

```bash
docker compose up --build
```

Compose 会启动 PostgreSQL 和非 root 的应用容器，迁移目录作为初始化脚本挂载。镜像使用官方 Go 多架构基础镜像，支持 `linux/amd64` 和 `linux/arm64` 构建；本地验证以当前平台为准。

## 演示数据

`migrations/000002_seed.up.sql` 提供四个角色账号：

- `author@example.test`：作者
- `editor@example.test`：编辑
- `chief@example.test`：主编
- `section@example.test`：栏目管理员

种子密码仅用于本地演示，生产环境必须替换并通过配置管理系统注入。

## API 示例

完整 schema 见 `api/openapi/openapi.json`。典型调用：

```bash
curl -X POST http://localhost:8080/api/v1/manuscripts \
  -H 'Content-Type: application/json' -H 'X-Actor-ID: demo-author' \
  -d '{"section_id":"research","title":"一篇研究","abstract":"摘要","markdown":"正文内容……"}'
```

所有 API 错误都返回稳定 `code`、可读 `message`、可选字段错误和 `request_id`。列表接口提供分页、排序和白名单筛选；公开刊发检索不需要认证。

## 验证

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
cd web
npm test -- --run
npm run typecheck
npm run build
```

已在本项目基线执行上述 Go 与前端检查。PostgreSQL 集成检查只有在设置 `TEST_DATABASE_URL` 时运行；没有该变量时会明确跳过，不会影响离线演示。

## 限制

本项目提供本地文件存储和通知适配器端口，默认实现是内存适配器；生产部署需要接入对象存储、邮件或企业通知实现，并保持同样的 context、重试上限和死信语义。认证中间件保留了本地演示身份头，生产环境应由网关或 JWT 验证器提供 actor。
