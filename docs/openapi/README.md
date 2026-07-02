# OpenAPI 文档生成说明

## 目录约定

- `docs/openapi/modules/*.openapi.json`：按 `api/<domain>/<version>` 或 `api/<domain>/<api>/<version>` 模块生成的正式发布产物。
- `docs/openapi/overlays/*.yaml`：补充上游生成器难以直接表达的文档细节，主要用于 SSE 的 media type、headers、examples。
- `docs/openapi/bundles/openapi.json`：全量模块聚合 bundle。

## 常用命令

```bash
make init
make api
make openapi
make all
```

## 生成约束

1. OpenAPI 发布产物来自 `protoc-gen-openapi v0.7.1` 基线，再由 `scripts/openapi` 规范化为 OpenAPI 3.1 JSON。
2. `docs/openapi/modules/*.openapi.json` 与 `docs/openapi/bundles/*.json` 为 generated 文件，不要手改；如需补充 SSE headers、examples 或修正文案，请改 `docs/openapi/overlays/*.yaml` 或 proto 注释。
3. `api/**/*.swagger.json` 已废弃，发布器会拒绝旧 Swagger 产物回归。
4. 模板同时支持 `api/todo/v1` 与 `api/user/admin/v1` 两种模块目录。
5. 枚举说明来自 proto enum 和 enum value 注释，发布器会补充 `x-enum-varnames`、`x-enum-descriptions`，并在 description 中追加“枚举值”说明块。

## SSE 文档方式

普通 Kratos server-streaming RPC 可以用 generated HTTP SSE handler，例如模板里的 `TodoService.WatchTodos`。如果某个 SSE 需要手写 HTTP route，则 proto 可以作为“文档专用 service”参与生成，但不要再注册对应 generated HTTP server，避免和手写 route 产生重复 path。

SSE overlay 至少补充：

- `responses.200.content.text/event-stream.schema.type: string`
- `Cache-Control`、`Connection`、`X-Accel-Buffering` 响应头
- 至少一个 `example` 或 `examples`
