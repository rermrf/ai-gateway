# AI Gateway - Agent 指南 (AGENTS.md)

本文件面向在本仓库中自动编写/修改代码的智能体（agentic coding agents）。目标：
- 让你能快速跑起来（build/lint/test/dev）。
- 让你写出的代码符合本项目既有风格与约定。
仓库概览：
- Go 后端：`cmd/server/main.go` + `internal/` 分层（api/http -> service -> repository -> dao）。
- 管理后台：`web/admin`（React + TS + Vite + Tailwind）。
备注（Cursor/Copilot 规则）：
- 未发现 `.cursor/rules/`、`.cursorrules` 或 `.github/copilot-instructions.md`；如后续新增，请把关键约束同步到本文。

---

## 1) 常用命令 (Build/Lint/Test)

### 后端 (Go)
依赖与初始化：
- `make setup`（`go mod download` + `go mod tidy`）

格式化：
- `make fmt`（`gofmt -w .` + `go mod tidy`）
- 只格式化单个文件：`gofmt -w path/to/file.go`

Lint：
- `make lint`（`golangci-lint run ./...`）
- 需要本机安装 `golangci-lint`（未看到仓库内的 `.golangci.yml`，因此使用默认规则集）。

测试：
- `make test`（等价于 `go test -race -shuffle=on -short -failfast ./...`）

运行单个包测试：
- `go test ./internal/service/user -race -count=1`

运行单个测试函数（推荐写法：正则加锚点 + 禁用缓存）：
- `go test ./internal/service/user -run '^TestRegister$' -count=1 -race`

运行单个测试用例/子测试：
- `go test ./internal/service/user -run '^TestRegister$/^case_name$' -count=1`

运行服务：
- `make run`（默认配置）
- `make run-config`（指定 `config/config.yaml`）
- `./scripts/start-with-env.sh`（读取 `.env`，强校验 `DB_PASSWORD` / `JWT_SECRET`）

构建：
- `make build`（输出：`bin/ai-gateway`）

代码生成：
- `make gen`（`go generate ./...`）
- 项目内有 `//go:generate mockgen ...`（例如 `internal/service/user/user.go`），改动接口后记得重新生成 mocks。

### 前端 (web/admin)
安装依赖：
- `cd web/admin && npm ci`

开发：
- `cd web/admin && npm run dev`

构建：
- `cd web/admin && npm run build`

Lint：
- `cd web/admin && npm run lint`
- 修复（eslint 默认脚本不带 `--fix`）：`cd web/admin && npm run lint -- --fix`
（前端当前未看到测试脚本；如需新增，优先考虑 `vitest` + `@testing-library/react`。）

---

## 2) 代码风格与约定 (Go)
### 2.1 工程结构与职责边界

- `internal/api/http/...`：Gin handler + middleware（只做鉴权、参数校验、HTTP 映射；不要塞业务逻辑）。
- `internal/service/...`：业务逻辑（组合 repository、做领域校验、写日志）。
- `internal/repository/...`：数据访问抽象（可组合 cache + dao；对外返回 `domain`）。
- `internal/repository/dao/...`：GORM 模型与 CRUD（对外返回 dao struct；对“未找到”通常返回 `(nil, nil)`）。
- `internal/domain/...`：协议无关的领域模型。
- `internal/errs`：统一错误码与 `AppError`（新增错误优先在此集中管理）。

### 2.2 imports 与格式化

- 使用 `gofmt`（仓库已通过 `make fmt` 强制）。
- import 分组保持三段（标准库 / 第三方 / 本项目），组间空行；本仓库大量使用 `ai-gateway/internal/...` 作为本项目导入前缀。

### 2.3 命名规范

- 包名：全小写、短、单数（如 `user`, `apikey`, `middleware`）。
- 接口：通常命名为 `Service`、`Repository`、`DAO`（本项目大量使用 `type Service interface { ... }`）。
- 实现结构体：小写 `service`（避免导出实现，利于替换），构造函数 `NewService`。
- context 变量统一叫 `ctx`，并作为函数第一个参数（除 receiver 外）。

### 2.4 错误处理

- 对外（HTTP）不要直接暴露底层错误细节；对内要记录原始错误（logger fields）。
- 业务错误尽量复用 `internal/errs` 里的预定义错误（例如 `errs.ErrUserNotFound`）。
- 新增错误时：
  - 能用已有错误码/错误实例就不要新增。
  - 需要新语义时在 `internal/errs/error.go` 增加 `ErrorCode` + 对应 `ErrXXX`（保持错误码分段含义）。
- 判断错误：优先 `errors.Is/As`（例如 `errors.Is(err, errs.ErrUserNotFound)`）。

### 2.5 日志

- 统一使用 `internal/pkg/logger` 的接口（底层是 zap）。
- 建议在构造函数里用 `l.With(...)` 固定组件字段（例：`logger.String("service", "user")`）。
- 避免在热路径打印大对象；流式/大响应只打 request_id、模型名、耗时、状态码等关键信息。

### 2.6 数据库与事务

- DAO 层使用 GORM：`db.WithContext(ctx)`；查询不到记录一般返回 `(nil, nil)`（见 `internal/repository/dao/user.go`）。
- 上层 service/repository 再把 `(nil, nil)` 映射成领域错误（例如 `errs.ErrUserNotFound`）。
- 新增表/字段：优先修改 `internal/repository/dao` 模型 + `internal/ioc/db.go` 的 `AutoMigrate(...)` 列表，并补充 `scripts/mysql/*.sql`（如项目需要兼容手工迁移）。

### 2.7 HTTP Handler 约定 (Gin)

- 入参校验：`c.ShouldBindJSON(&req)` 失败返回 `400`。
- 中间件：`internal/api/http/middleware` 已提供 `Recovery` / `Logger` / `Cors` / `RequestID` / `JWTAuth` / `APIKeyAuth` / `RateLimiter`。
- 鉴权：
  - 用户 API：`JWTAuth`；管理员 API：额外 `RequireAdmin()`。
  - OpenAI/Anthropic 兼容接口：`APIKeyAuth`。

---

## 3) 代码风格与约定 (前端 web/admin)
### 3.1 技术栈与约束

- Vite + React + TypeScript（`strict: true`，并启用 `noUnusedLocals/noUnusedParameters`）。
- ESLint：`eslint.config.js` 使用 recommended + typescript-eslint + react-hooks + react-refresh。
- 路径别名：`@/*` -> `web/admin/src/*`（见 `vite.config.ts` / `tsconfig.app.json`）。

### 3.2 imports

- 外部依赖优先，其次别名 `@/...`，最后相对路径。
- 类型导入使用 `import type { ... } from '...'`（仓库已有示例：`web/admin/src/contexts/AuthContext.tsx`）。

### 3.3 命名与组件组织

- React 组件：`PascalCase` 文件名与导出名一致（例：`pages/Login.tsx` 导出 `Login`）。
- hooks：`useXxx`。
- 类型：可用 `interface`（当前仓库 `types/index.ts` 以 `interface` 为主）。

### 3.4 API 调用与错误处理

- 统一走 `web/admin/src/api/client.ts`（axios instance + token 注入 + 401 自动跳转）。
- 页面层捕获错误时，显示对用户友好的文案；调试信息用 `console.error` 即可（必要时再引入更系统的上报）。

---

## 4) 智能体工作方式建议

- 优先跑 `make test` / `make lint` 验证后端变更；前端跑 `npm run lint` + `npm run build`。
- 不要提交或打印真实密钥：`.env`、provider apiKey、JWT secret 等属于敏感信息。
- 生成代码（mocks 等）属于仓库的一部分：如果接口变了导致编译失败，记得 `make gen`。
