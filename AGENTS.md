# Kratos Template Instructions

## 项目快照

这是一个基于最新 `kratos new` 官方 layout 扩展出来的 Kratos v3 后端模板。模板自身保留 `cmd/server`，因为 `kratos new <name> -r <repo>` 会把它重命名成 `cmd/<name>` 并替换 `go.mod` module path。

模板默认包含：

- Kratos v3 HTTP + gRPC
- Buf protobuf 生成链
- OpenAPI 3.1 文档发布器，包含 SSE overlay 与 enum 注释增强
- Google Wire
- Ent ORM 基础 schema 与生成链
- Redis 客户端
- JWT Ed25519 token 管理
- Casbin 鉴权中间件
- Asynq 后台任务运行时
- Redsync 分布式锁依赖
- zap + slog 日志与日志轮转
- 常用 helper：`typecatch`、`pagination`、`id`、`decimalx`
- Dockerfile 与本地 Postgres/Redis docker-compose

## 常用 helper 用法

- `typecatch.CopyTo[SRC, DST](&src)`：只用于相邻层之间同名字段、同语义的结构体复制，例如 Ent entity 转 biz 模块结果类型。字段名、单位、权限过滤或业务语义不同就显式映射。

## Project Skills

任务匹配时优先读取 `.agents/skills/` 下的项目技能：

- `kratos-architecture-navigation`：非平凡改动、跨层变更、模块边界不清。
- `kratos-backend-api`：proto/API/service/biz/data 接口开发。
- `kratos-data-ent`：Ent schema、repo、事务、Redis store。
- `kratos-auth-permission`：JWT、Casbin、匿名白名单、安全敏感接口。
- `kratos-job-runtime`：Asynq、Redsync、后台任务、队列与锁。
- `kratos-integrations`：外部 HTTP 客户端、Webhook、对象存储、Redis Streams/SSE、分布式锁。
- `kratos-devops`：Makefile、工具链、配置、Docker、部署。
- `kratos-reuse-map`：新增 helper 前查现有复用点。
- `kratos-test-quality`：测试、lint、生成代码、验证策略。

## 常用命令

```bash
make init      # 安装 wire/buf/ent/golangci-lint
make api       # 生成 api proto
make openapi   # 生成 OpenAPI 3.1 发布文档
make config    # 生成 internal/conf proto
make ent       # 生成 Ent 代码
make generate  # 生成 Wire 并 go mod tidy
make all       # api + config + ent + generate
make build     # 编译 cmd 下唯一入口
make run       # 本地运行 cmd 下唯一入口
make test      # go test ./...
make lint      # golangci-lint run
```

## 架构规则

默认层次是 `service -> biz -> data`：

- `api/`：proto 是接口契约源头。
- `internal/service/`：只做 proto 和 biz 对象转换、当前用户提取、错误透传。
- `internal/biz/`：业务编排、校验、事务边界，Repo interface 放这里。
- `internal/data/`：Repo 实现、Ent/Redis 访问、系统错误翻译。
- `internal/dep/`：外部依赖初始化，如 DB、Redis、Casbin、Asynq、Redsync。
- `internal/server/`：HTTP/gRPC、中间件、路由注册。
- `internal/job/`：后台任务运行时。
- `internal/pkg/`：跨模块技术 helper。
- `docs/openapi/`：正式 OpenAPI 发布产物、SSE overlay 与聚合 bundle。

不要让 biz import Ent、SQL、Redis 这类具体存储实现。不要在 service 写业务规则。

### 模块内分层

简单模块先平铺，文件少时不要提前建目录：

```text
internal/biz/todo.go
internal/data/todo.go
internal/service/todo.go
```

模块变复杂后拆成同名目录：

```text
internal/biz/order/
  use_case.go   # UseCase、Repo interface、跨模块 Provider interface、构造函数
  types.go      # 模块共享请求/结果类型、状态常量、值对象；不要另建 internal/dto
  errors.go     # 模块私有 sentinel 或错误 helper；公开 API 错误仍从 proto error reason 生成
  command.go    # 写操作编排
  query.go      # 读操作编排
  validate.go   # 复杂校验
```

结构体组织规则：

- 按业务模块包组织结构体，不按 DTO/VO/BO 这类类型种类建立中央包。
- API 边界类型以 proto 生成类型为准；模块内跨 service/biz/data 共享的请求、结果、值对象放在 `internal/biz/<module>/types.go`。
- `types.go` 明显变大后，在同一模块目录按用途拆成 `request.go`、`result.go`、`model.go`、`status.go` 等；不要迁到 `internal/dto`。
- data 层不要把 Ent entity 传出层边界，先映射成对应 biz 模块类型。

`internal/service/<module>/` 和 `internal/data/<module>/` 跟随同一模块名；service 只做 proto 与 biz 模块类型转换，data 只实现 biz 的 Repo interface。跨模块调用用窄 Provider interface 放在消费方 biz 模块，不直接 import 对方 data。

错误处理规则：

- 公开 API 错误先在 `api/<domain>/<api>/v1/*_error.proto` 或当前模块 error proto 定义 reason，再由 biz/data 返回对应 Kratos error。
- data 层把 Ent/Redis/SQL 错误翻译成领域错误，不把底层错误直接抛给 service。
- biz 层做业务错误归一和冲突/幂等判断；service 层只透传错误，不重新包装业务错误。
- 包内非 API sentinel 用 `errors.Is` 可识别的 `var ErrX = errors.New(...)`，只在模块内部或明确的 Provider contract 中使用。

## 模板约束

- 模板仓库里入口目录保持 `cmd/server`。
- 不要在模板文件里硬编码新项目名；依赖 Kratos CLI 替换 module path。
- 示例配置不能提交真实 DSN、token、密钥、证书。
- 默认配置应能在无 DB/Redis 的情况下启动；需要外部依赖时通过 config/env 打开。
- 新增生成源后先改源文件，再运行生成命令，不手改生成物。
- OpenAPI 正式产物由 `scripts/openapi` 生成到 `docs/openapi`，不要再依赖根目录 `openapi.yaml`。
- SSE 文档用 `docs/openapi/overlays/<module>.yaml` 补 `text/event-stream`、响应头、examples。手写 SSE route 可以使用文档专用 proto 参与生成，但不要再注册对应 generated HTTP server，避免重复 path。
- enum 文档必须写在 proto enum 和 enum value 注释里，发布器会生成 `x-enum-varnames`、`x-enum-descriptions` 和 description 中的“枚举值”说明块。

## 生成文件

不要手改：

- `api/**/*.pb.go`
- `api/**/*_http.pb.go`
- `api/**/*_grpc.pb.go`
- `internal/conf/conf.pb.go`
- `internal/data/ent` 生成文件，除 `internal/data/ent/schema`
- `cmd/*/wire_gen.go`
- `docs/openapi/modules/*.openapi.json`
- `docs/openapi/bundles/*.json`

## 验证

改 Go 代码后至少跑：

```bash
go test ./...
make build
```

改 proto/config/schema/Wire 后先跑对应生成命令，再测试。
