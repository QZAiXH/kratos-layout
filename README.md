# kratos-layout

基于最新 `kratos new` 官方 layout 扩展的 Kratos v3 后端模板。

## 包含内容

- Kratos v3 HTTP/gRPC
- Buf + protobuf 生成链
- OpenAPI 3.1 文档发布器，包含 SSE overlay 与 enum 注释增强
- Wire 依赖注入
- Ent ORM 基础 schema
- Redis
- JWT Ed25519
- Casbin
- Asynq 后台任务运行时
- Redsync 分布式锁依赖
- zap/slog 日志与轮转
- 常用 helper：`typecatch`、`pagination`、`id`、`decimalx`
- Dockerfile 与 Postgres/Redis docker-compose
- `.agents` 项目技能

## 生成新项目

先把本目录提交到 Git 仓库，然后：

```bash
kratos new demo -r <your-git-repo-url> --timeout 120s
cd demo
make all
go test ./...
make build
```

本模板仓库内的入口目录必须保持 `cmd/server`。Kratos CLI 会在生成项目时改成 `cmd/<project>`。

## 本地开发

```bash
make init
make all
make openapi
go test ./...
make run
```

## GoLand Proto 报红

最新 Kratos v3 模板通过 Buf 远程依赖解析 `google/api/*.proto`，不会提交旧版 `third_party/`。

推荐在 GoLand 安装 `Buf for Protocol Buffers` 插件，并确保 `make init` 安装的 `buf` 在 `PATH`，插件会通过 `buf lsp serve` 读取 `buf.yaml` 和 `buf.lock`。

如果 IDE 仍然无法解析 import，执行：

```bash
make proto-ide
```

然后在 GoLand 的 `Settings | Languages & Frameworks | Protocol Buffers` 里，把项目根目录下的 `.proto-deps` 加到 Import Paths。`.proto-deps/` 只给 IDE 使用，已加入 `.gitignore`，生成代码仍然使用 Buf。

默认配置不强制连接 DB/Redis。需要本地依赖时：

```bash
docker compose -f deploy/docker-compose.yml up -d
```

然后按需修改 `configs/config.yaml` 或通过 `configs/config.env.example` 中的环境变量覆盖敏感值。

## 注意

- 不要提交真实 DSN、token、密钥、证书。
- 不要手改 generated files，改源文件后重新运行 `make api`、`make openapi`、`make config`、`make ent`、`make generate`。
- proto `go_package` 使用相对路径，避免 `kratos new` 替换 module path 时破坏 protobuf raw descriptor。
- OpenAPI 正式产物在 `docs/openapi`。SSE 响应用 `docs/openapi/overlays/*.yaml` 补 `text/event-stream`、headers、examples；enum 文档来自 proto enum/value 注释。
