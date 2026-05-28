# 通用约定

## 鉴权

除明确标注“无需鉴权”的接口外，请求头必须携带：

```http
Authorization: Bearer <token>
```

## 成功响应

`w7panel-server/main.go` 覆盖了框架默认成功响应处理器，成功响应会直接返回业务数据。

普通成功：

```json
"success"
```

对象成功：

```json
{
  "field": "value"
}
```

## 错误响应

错误响应仍使用框架默认格式：

```json
{
  "error": "错误信息",
  "code": 500
}
```

## 参数位置

`form` 表示 Controller 使用 `form` tag 绑定，通常支持 query、form-urlencoded 或 multipart form，具体取决于 HTTP 方法和 Content-Type。`json` 表示 JSON body。
