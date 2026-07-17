# 容器文件管理 API

本文档说明容器文件管理相关 API 的职责、调用凭据、访问模式和请求/响应参数。文件管理能力主要服务于应用详情文件页、微应用文件弹窗、压缩解压、权限修改、挂载文件编辑和分片上传。

## 整体使用方式

容器文件接口用于把面板文件管理能力映射到目标容器 rootfs。开发时先定位目标容器 PID，拿到后端返回的 WebDAV、压缩、权限 URL 和 `webdavToken`，再按文件操作类型调用对应接口。

### 基本流程

1. 使用用户 token 调用 `/panel-api/v1/pid`，传入 namespace、节点 IP、Pod/容器信息。
2. 后端返回 `pid`、`subPid`、`webdavUrl`、`webdavBasePath`、`compressUrl`、`permissionUrl`、`webdavToken`。
3. WebDAV 读写、压缩解压、权限修改使用返回的 URL 和 `Authorization: Bearer {webdavToken}`。
4. 挂载文件、分片上传、下载复制等场景按各自接口处理，不要手动拼生产代理路径。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| 定位容器 rootfs | `/panel-api/v1/pid` | 返回文件访问所需 URL 和 token |
| 目录和文件读写 | WebDAV `PROPFIND`、`GET`、`PUT`、`DELETE` 等 | 保持 WebDAV 协议响应 |
| 压缩解压 | `compressUrl` 下 `/compress`、`/extract` | 对容器内路径操作 |
| 修改权限 | `permissionUrl` 下 `/chmod`、`/chown` | 修改容器内文件权限 |
| 大文件上传 | 分片上传接口 | 上传、检查、合并分片 |
| 挂载文件编辑 | mount file 接口 | 管理 ConfigMap/Secret 挂载内容 |

### 使用边界

- `webdavToken` 当前来自 `/pid` 返回字段，本质上是当前请求上下文里的 token。
- WebDAV 标准方法必须保留 XML、文件流等协议响应，不要包装为普通 JSON。
- `LOCAL_MOCK=true` 只改变 rootfs 定位路径，不代表跳过用户认证。
- `/proc`、`/sys`、`/dev`、socket、fifo、device 等特殊文件需要按文档限制处理。
- 生产模式 URL 可能带 Agent 代理前缀，前端应使用后端返回的 URL。

## 通用说明

### 鉴权

本文接口均需要用户 token：

```http
Authorization: Bearer <token>
```

`/panel-api/v1/pid` 返回的 `webdavToken` 当前就是用户 token。前端访问 WebDAV、压缩、权限接口时，应使用：

```http
Authorization: Bearer <webdavToken>
```

`LOCAL_MOCK=true` 只改变文件系统定位方式，不跳过用户认证。

### 响应格式

当前后端成功响应处理器直接返回业务对象，不额外包裹 `code`、`data`、`message`。WebDAV 接口必须保持 WebDAV 标准响应，`PROPFIND` 返回 XML，`GET` 返回文件流，不能改成普通 JSON。

错误响应通常由框架统一处理；显式校验错误常见 HTTP 状态码为 `400`，服务端错误为 `500`，下载文件不存在为 `404`。

### 参数位置

`/panel-api/v1/pid` 主要使用 query/form 参数。WebDAV 接口通过 path 表示目标 PID 和容器内路径，Header 承载 `Depth`、`Destination`、`Overwrite` 等协议参数，`PUT` body 为文件内容。压缩、权限、分片上传和挂载文件接口按各自参数表使用 JSON body、form 或 multipart。

## 能力概览

| 能力 | 说明 |
|------|------|
| 定位容器文件系统 | 通过 `/panel-api/v1/pid` 获取目标容器的 `pid`、`subPid`、`webdavUrl`、`compressUrl`、`permissionUrl` 和 `webdavToken` |
| WebDAV 文件访问 | 使用标准 WebDAV 方法读取、写入、删除、移动、复制和列目录 |
| 压缩解压 | 对容器内路径执行 zip、tar、tar.gz、tar.xz 等压缩，支持多种格式解压 |
| 权限修改 | 对容器内文件执行 chmod、chown |
| 文件下载和复制 | 通过 `/download/*path`、`/cp`、`/cppid`、`/mvpid` 复制或下载文件 |
| 分片上传 | 面板侧分片上传、检查和合并，可直接合并到目标容器 |
| 挂载文件 | 查询、创建、更新、删除和修改 ConfigMap/Secret 挂载文件 |

## 调用流程

1. 前端使用用户 token 调用 `/panel-api/v1/pid`，传入 `namespace`、`HostIp`、`podName`、`containerName` 或 `containerId`。
2. 后端返回 `webdavUrl`、`webdavBasePath`、`compressUrl`、`permissionUrl`、`webdavToken`。
3. 前端使用 `webdavToken` 访问 WebDAV、压缩和权限接口。
4. 需要生产 Agent 代理时，前端使用后端返回的 URL，不自行拼接生产代理路径。

## LOCAL_MOCK 路径差异

文件接口最终都会映射到目标进程 rootfs：

| 模式 | rootfs 基础路径 |
|------|-----------------|
| `LOCAL_MOCK=true` | `/host/proc/{pid}/root` |
| 生产模式 | `/proc/{pid}/root` |
| 子进程 rootfs | `{rootfs}/proc/{subPid}/root` |

`subPid` 为空或 `0` 时表示访问主进程 rootfs。

## 容器定位接口

### GET `/panel-api/v1/pid`

功能：获取目标容器的文件访问信息。

鉴权：需要用户 token。该接口带 1 分钟响应缓存。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form | 是 | string | 原始 Pod 所在命名空间 |
| `HostIp` | query/form | 是 | string | 目标 Pod 所在节点 IP；注意参数名首字母大写 |
| `containerId` | query/form | 否 | string | 容器 ID；有值时可用于定位 PID |
| `podName` | query/form | 否 | string | 原始 Pod 名称 |
| `containerName` | query/form | 否 | string | 原始容器名 |

响应字段均为字符串：

| 字段 | 类型 | 说明 |
|------|------|------|
| `podName` | string | Agent Pod 名称，不一定是原始业务 Pod 名称 |
| `pid` | string | 主进程 PID |
| `subPid` | string | 子进程 PID；无子进程时为空字符串 |
| `namespace` | string | Agent Pod namespace |
| `containerName` | string | Agent Pod 第一个容器名 |
| `podIp` | string | Agent Pod IP，也用于生产代理地址 |
| `webdavUrl` | string | WebDAV 访问基础路径 |
| `webdavBasePath` | string | 编辑器用于截断/展示路径的基础路径，不带开头 `/` |
| `compressUrl` | string | 压缩/解压 API 基础路径 |
| `permissionUrl` | string | 权限修改 API 基础路径 |
| `pwd` | string | 初始工作目录 |
| `agentUrl` | string | Agent 代理基础路径 |
| `webdavToken` | string | 文件访问 token，当前等于用户 token |

生产模式响应示例：

```json
{
  "podName": "w7panel-agent-abcde",
  "pid": "12345",
  "subPid": "",
  "namespace": "default",
  "containerName": "agent",
  "podIp": "10.0.0.10",
  "webdavUrl": "/panel-api/v1/10.0.0.10:8000/proxy/panel-api/v1/files/webdav-agent/12345/agent",
  "webdavBasePath": "panel-api/v1/files/webdav-agent/12345/agent",
  "compressUrl": "/panel-api/v1/10.0.0.10:8000/proxy/panel-api/v1/files/compress-agent/12345",
  "permissionUrl": "/panel-api/v1/10.0.0.10:8000/proxy/panel-api/v1/files/permission-agent/12345",
  "pwd": "/",
  "agentUrl": "/panel-api/v1/10.0.0.10:8000/proxy",
  "webdavToken": "eyJ..."
}
```

当前 `/pid` 不生成 `/subagent/{subPid}` 形式的访问 URL，`subPid` 响应字段通常为空。

## WebDAV 接口

### 路径

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| 多方法 | `/panel-api/v1/files/webdav-agent/:pid/agent/*path` | 需要 token | 访问主进程 rootfs |

注册的多方法包括：`PROPFIND`、`PROPPATCH`、`MKCOL`、`COPY`、`MOVE`、`LOCK`、`UNLOCK`、`LINK`、`UNLINK`、`GET`、`PUT`、`DELETE`、`HEAD`、`OPTIONS`、`PATCH`、`POST`。

### 请求参数

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `pid` | path | 是 | string | 主进程 PID |
| `path` | path | 是 | string | 容器内路径；空路径按 `/` 处理 |
| body | body | `PUT` 时必填 | bytes | 写入文件内容 |

常用 Header：

| Header | 必填 | 说明 |
|--------|------|------|
| `Authorization: Bearer {webdavToken}` | 是 | 文件访问 token |
| `Depth` | `PROPFIND` 时常用 | `0` 或 `1` |
| `Destination` | `MOVE`/`COPY` 时必填 | 目标 URL 或目标路径；代理层会改写 WebDAV Destination 路径 |
| `Overwrite` | `MOVE`/`COPY` 时可选 | 是否覆盖目标 |
| `Content-Type` | `PUT` 时建议 | 文件 MIME 或 `application/octet-stream` |

### 响应

| 方法 | 成功响应 | 说明 |
|------|----------|------|
| `GET` | bytes | 文件内容；目录或不可读特殊文件会按 WebDAV 错误处理 |
| `PROPFIND` | XML | WebDAV `multistatus` |
| `PUT` | WebDAV 状态码 | 写入成功 |
| `DELETE` | WebDAV 状态码 | 删除成功 |
| `MKCOL` | WebDAV 状态码 | 创建目录成功 |
| `MOVE`/`COPY` | WebDAV 状态码 | 移动或复制成功 |

`PROPFIND` 会额外返回 `w7panel` 命名空间属性：

| 属性 | 类型 | 说明 |
|------|------|------|
| `uid` | string | 文件 uid |
| `gid` | string | 文件 gid |
| `mode` | string | 八进制权限，例如 `644` |
| `type` | string | `file`、`directory`、`symlink`、`device`、`fifo`、`socket`、`unknown` |
| `editable` | bool string | 是否可读写编辑；特殊文件和越界 symlink 为 `false` |

限制和行为：

| 项 | 说明 |
|----|------|
| 目录最大条目 | 单次 `Readdir` 最多 5000 项 |
| 特殊文件 | device/fifo/socket/unknown 不允许 `Read`/`Seek` |
| symlink | 指向容器 rootfs 外部的 symlink 会标记为不可编辑 |
| Content-Type | 普通文件按扩展名推断，未知为 `application/octet-stream`；非普通文件为 `application/linux-file` |

请求示例：

```http
PROPFIND /panel-api/v1/files/webdav-agent/12345/agent/etc
Authorization: Bearer <webdavToken>
Depth: 1
```

```http
PUT /panel-api/v1/files/webdav-agent/12345/agent/tmp/hello.txt
Authorization: Bearer <webdavToken>
Content-Type: text/plain

hello
```

## 压缩解压

### POST `/compress`

请求体可用 JSON 或 form：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `sources` | 是 | array[string] | 要压缩的容器内路径列表 |
| `output` | 是 | string | 输出压缩包容器内路径 |

格式由 `output` 扩展名决定：

| 扩展名 | 行为 |
|--------|------|
| `.zip` 或未知扩展名 | ZIP |
| `.tar` | tar |
| `.tar.gz` / `.tgz` | tar + gzip |
| `.tar.xz` / `.txz` | tar + xz |
| `.tar.bz2` / `.tbz2` | 当前代码会进入 bzip2 压缩分支，但返回 `bzip2 compression not supported` |

请求示例：

```http
POST /panel-api/v1/files/compress-agent/12345/compress
Authorization: Bearer <webdavToken>
Content-Type: application/json

{
  "sources": ["/etc/nginx", "/var/log/nginx/access.log"],
  "output": "/tmp/nginx.tar.gz"
}
```

成功响应：JSON 字符串 `"success"`。

### POST `/extract`

请求体可用 JSON 或 form：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `source` | 是 | string | 压缩包容器内路径 |
| `target` | 是 | string | 解压目标目录，目录不存在会自动创建 |

支持解压格式：

| 扩展名 | 说明 |
|--------|------|
| `.zip` | ZIP |
| `.tar` | tar；也会尝试自动识别 gzip/xz/bzip2 tar 流 |
| `.tar.gz` / `.tgz` | tar + gzip |
| `.tar.bz2` / `.tbz2` | tar + bzip2 |
| `.tar.xz` / `.txz` | tar + xz |
| `.7z` | 7z |
| 未知扩展名 | 默认按 zip/自动检测逻辑处理 |

安全限制：解压 ZIP/TAR/7z 时会检查目标路径，阻止 `../` 这类路径穿越写出目标目录。

请求示例：

```http
POST /panel-api/v1/files/compress-agent/12345/extract
Authorization: Bearer <webdavToken>
Content-Type: application/json

{
  "source": "/tmp/site.zip",
  "target": "/var/www/html"
}
```

成功响应：JSON 字符串 `"success"`。

## 权限修改

### POST `/chmod`

请求体可用 JSON 或 form：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `path` | 是 | string | 容器内文件路径 |
| `mode` | 是 | string | 八进制权限字符串，例如 `755`、`0644` |
| `recursive` | 否 | bool | 是否递归；递归时对目录树内每个路径执行 `os.Chmod` |

请求示例：

```json
{
  "path": "/var/www/html/index.php",
  "mode": "755",
  "recursive": false
}
```

成功响应：JSON 字符串 `"success"`。

### POST `/chown`

请求体可用 JSON 或 form：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `path` | 是 | string | 容器内文件路径 |
| `owner` | 是 | string | owner 表达式 |
| `recursive` | 否 | bool | 是否递归；递归时对目录树内每个路径执行 `os.Chown` |

`owner` 解析规则：

| 写法 | 行为 |
|------|------|
| `1000` | uid 和 gid 都设为 `1000` |
| `root` | uid 设为 `0`，gid 保持默认 `0` |
| `root:root` / `root.root` | uid/gid 都设为 `0` |
| 其它用户名/组名 | 当前仅内置识别 `root`；无法识别时为 `-1`，由 `os.Chown` 按系统语义处理 |

请求示例：

```json
{
  "path": "/var/www/html",
  "owner": "root:root",
  "recursive": true
}
```

成功响应：JSON 字符串 `"success"`。

## 下载、复制和移动

### GET `/panel-api/v1/download/*path`

功能：从 `s3.base_dir` 下载文件。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `path` | path | 是 | string | 相对 `s3.base_dir` 的文件路径 |

成功响应：文件流。

响应 Header：

| Header | 说明 |
|--------|------|
| `Content-Type` | `application/octet-stream` |
| `Content-Disposition` | `attachment; filename={服务端文件完整路径}` |

错误：文件不存在返回 `404`。

### POST `/panel-api/v1/cppid`

功能：在 `s3.base_dir` 与容器 rootfs 路径之间复制文件或目录。

`POST /panel-api/v1/mvpid` 当前绑定到同一个 `CpPidFile` 方法，行为与 `/cppid` 相同，并不是 rename/move。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `from` | query/form/body | 是 | string | 源路径 |
| `to` | query/form/body | 是 | string | 目标路径 |
| `upload` | query/form/body | 是 | string | `1` 表示从 `s3.base_dir/from` 复制到容器路径 `to`；其它值表示从容器路径 `from` 复制到 `s3.base_dir/to` |
| `pid` | query/form/body | 是 | string | 目标 PID；当前实现只校验必填，路径转换由 `procpath.ConvertToLocalPath` 处理 |

成功响应：JSON 字符串 `"success"`。

### POST `/panel-api/v1/cp`

功能：使用 kubectl cp 在 `s3.base_dir` 与 Pod 之间复制文件。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `from` | query/form/body | 是 | string | 源路径 |
| `to` | query/form/body | 是 | string | 目标路径 |
| `namespace` | query/form/body | 是 | string | Pod namespace |
| `upload` | query/form/body | 是 | string | `1` 表示从 `s3.base_dir/from` 上传到 `podName:to`；其它值表示从 `podName:from` 下载到 `s3.base_dir/to` |
| `podName` | query/form/body | 是 | string | Pod 名称 |

成功响应：JSON 字符串 `"success"`。

## 分片上传

分片上传先写入 `s3.base_dir/.chunks/{identifier}`。合并时：

| 场景 | 合并目标 |
|------|----------|
| 未传 `pid` 或 `pid=0` | `s3.base_dir/{fileName}` |
| 传入 `pid` | `{rootfs}/{fileName}`，`subPid` 非空时使用子进程 rootfs |

### POST `/panel-api/v1/files/upload-chunk`

请求类型：`multipart/form-data`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `file` | form file | 是 | file | 当前分片文件 |
| `chunkIndex` | form | 是 | string/int | 分片序号，从 0 开始 |
| `chunkTotal` | form | 是 | string/int | 总分片数 |
| `identifier` | form | 是 | string | 文件唯一标识，通常为文件 hash |
| `fileName` | form | 否 | string | 原始文件名，仅用于日志 |
| `relativePath` | form | 否 | string | 预留字段，当前未使用 |
| `fileSize` | form | 否 | string/int64 | 文件总大小，当前仅解析预留 |

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `chunkExists` | bool | 分片是否已存在 |
| `chunkIndex` | int | 分片序号 |
| `chunkTotal` | int | 总分片数；分片已存在时不返回 |
| `written` | int64 | 本次写入字节数；分片已存在时不返回 |

上传成功示例：

```json
{
  "chunkExists": false,
  "chunkIndex": 0,
  "chunkTotal": 3,
  "written": 1048576
}
```

分片已存在示例：

```json
{
  "chunkExists": true,
  "chunkIndex": 0
}
```

### GET `/panel-api/v1/files/check-chunk`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `identifier` | query | 是 | string | 文件唯一标识 |
| `chunkIndex` | query | 是 | string/int | 分片序号 |
| `chunkTotal` | query | 是 | string/int | 总分片数 |
| `fileName` | query | 否 | string | 文件名 |
| `relativePath` | query | 否 | string | 相对路径 |

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `chunkExists` | bool | 分片是否存在 |

### POST `/panel-api/v1/files/merge-chunks`

请求类型：`application/json`

请求参数：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `identifier` | 是 | string | 文件唯一标识 |
| `fileName` | 是 | string | 合并后的文件名或相对路径 |
| `totalChunks` | 是 | int | 总分片数 |
| `fileSize` | 否 | int64 | 文件总大小，当前预留 |
| `pid` | 否 | string | 目标容器 PID |
| `subPid` | 否 | string | 子进程 PID |

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `fileUrl` | string | 下载 URL；直接合并到容器 rootfs 时该 URL 可能不适合下载 |
| `fileName` | string | 文件名 |
| `fileSize` | int64 | 合并后写入字节数 |

响应示例：

```json
{
  "fileUrl": "/panel-api/v1/download/upload/demo.zip",
  "fileName": "upload/demo.zip",
  "fileSize": 3145728
}
```

错误：

| 场景 | 状态码 | 说明 |
|------|--------|------|
| 分片目录不存在 | `400` | `chunk directory not found` |
| 缺少指定分片 | `400` | `missing chunk {index}` |
| 创建目录/文件失败 | `500` | 返回底层错误 |

## 挂载文件接口

挂载文件接口用于读取和修改工作负载 PodSpec 中通过 ConfigMap、Secret、Projected 挂载出来的文件。支持的工作负载必须能从对象中解析 `spec.template.spec`。

### GET `/panel-api/v1/mountfiles`

功能：获取工作负载中的挂载文件列表。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form | 否 | string | 命名空间，空时使用当前 SDK 默认 namespace |
| `apiVersion` | query/form | 是 | string | 工作负载 apiVersion，例如 `apps/v1` |
| `kind` | query/form | 是 | string | 工作负载 kind，例如 `Deployment` |
| `name` | query/form | 是 | string | 工作负载名称 |
| `includeContent` | query/form | 否 | bool | 是否读取 ConfigMap/Secret 内容 |

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `namespace` | string | 命名空间 |
| `apiVersion` | string | 工作负载 apiVersion |
| `kind` | string | 工作负载 kind |
| `name` | string | 工作负载名称 |
| `mounts` | array[object] | 挂载说明列表 |

`mounts[]` 字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `containerName` | string | 容器名 |
| `containerType` | string | `container` 或 `initContainer` |
| `volumeName` | string | Volume 名称 |
| `mountPath` | string | 容器挂载路径 |
| `subPath` | string | volumeMount subPath，可能为空 |
| `readOnly` | bool | 是否只读 |
| `sourceType` | string | 来源类型：`configMap`、`secret`、`serviceAccountToken` |
| `sourceName` | string | ConfigMap/Secret 名称 |
| `files` | array[object] | 文件列表 |

`mounts[].files[]` 字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `path` | string | 容器内绝对路径 |
| `relativePath` | string | 相对挂载路径 |
| `sourceType` | string | 来源类型 |
| `sourceName` | string | ConfigMap/Secret 名称 |
| `key` | string | ConfigMap/Secret key |
| `optional` | bool | 是否 optional |
| `mode` | int32 | 文件 mode 数值 |
| `modeOctal` | string | 八进制 mode，例如 `0644` |
| `content` | string | 文本内容，仅 `includeContent=true` 且内容为文本时返回 |
| `contentBase64` | string | 二进制内容 base64，仅 `includeContent=true` 且内容为二进制时返回 |
| `binary` | bool | 是否二进制 |

响应示例：

```json
{
  "namespace": "default",
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "name": "demo",
  "mounts": [
    {
      "containerName": "web",
      "containerType": "container",
      "volumeName": "config",
      "mountPath": "/etc/nginx/conf.d",
      "readOnly": true,
      "sourceType": "configMap",
      "sourceName": "demo-config",
      "files": [
        {
          "path": "/etc/nginx/conf.d/default.conf",
          "relativePath": "default.conf",
          "sourceType": "configMap",
          "sourceName": "demo-config",
          "key": "default.conf",
          "modeOctal": "0644",
          "content": "server { listen 80; }"
        }
      ]
    }
  ]
}
```

### POST `/panel-api/v1/mountfiles`

功能：创建挂载文件。当前实现会新建 ConfigMap，并把文件挂载到匹配容器的 PodSpec。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form/body | 否 | string | 命名空间 |
| `apiVersion` | query/form/body | 是 | string | 工作负载 apiVersion |
| `kind` | query/form/body | 是 | string | 工作负载 kind |
| `name` | query/form/body | 是 | string | 工作负载名称 |
| `path` | query/form/body | 是 | string | 要创建的容器内绝对路径 |
| `containerName` | query/form/body | 否 | string | 容器名；用于选择挂载目标 |
| `content` | query/form/body | 否 | string | 文件内容 |
| `mode` | query/form/body | 否 | string | 文件权限，例如 `0644` |

行为：

| 项 | 说明 |
|----|------|
| ConfigMap 名称 | `w7-mf-{workloadName}-{fileName}` 形式，具体以 `buildMountFileConfigMapName` 为准 |
| Volume 名称 | 基于文件名生成，具体以 `buildMountFileVolumeName` 为准 |
| 文件 key | `path.Base(path)` |
| 默认 mode | Kubernetes ConfigMap volume 默认 mode |

成功响应：JSON 字符串 `"success"`。

### PUT `/panel-api/v1/mountfiles`

功能：更新已有挂载文件内容。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form/body | 否 | string | 命名空间 |
| `apiVersion` | query/form/body | 是 | string | 工作负载 apiVersion |
| `kind` | query/form/body | 是 | string | 工作负载 kind |
| `name` | query/form/body | 是 | string | 工作负载名称 |
| `path` | query/form/body | 是 | string | 已挂载文件的容器内绝对路径 |
| `content` | query/form/body | 否 | string | 新内容 |

行为：

| 来源 | 更新方式 |
|------|----------|
| `configMap` | 更新 `ConfigMap.data[key]`；若原 key 在 `binaryData` 中，则写入 `binaryData[key]` |
| `secret` | 更新 `Secret.data[key]`，内容按字节写入 |
| 其它来源 | 返回不支持错误 |

成功响应：JSON 字符串 `"success"`。

### DELETE `/panel-api/v1/mountfiles`

功能：删除已有挂载文件。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form/body | 否 | string | 命名空间 |
| `apiVersion` | query/form/body | 是 | string | 工作负载 apiVersion |
| `kind` | query/form/body | 是 | string | 工作负载 kind |
| `name` | query/form/body | 是 | string | 工作负载名称 |
| `path` | query/form/body | 是 | string | 已挂载文件的容器内绝对路径 |

行为：

| 来源 | 删除方式 |
|------|----------|
| `configMap` | 删除对应 key；若 ConfigMap 只有一个条目，会移除 PodSpec 中相关 volume 并删除 ConfigMap |
| `secret` | 删除对应 key，并移除 PodSpec 中该文件引用 |

成功响应：JSON 字符串 `"success"`。

### PUT `/panel-api/v1/mountfiles/chmod`

功能：修改挂载文件在 PodSpec 中的 mode。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form/body | 否 | string | 命名空间 |
| `apiVersion` | query/form/body | 是 | string | 工作负载 apiVersion |
| `kind` | query/form/body | 是 | string | 工作负载 kind |
| `name` | query/form/body | 是 | string | 工作负载名称 |
| `path` | query/form/body | 是 | string | 已挂载文件的容器内绝对路径 |
| `mode` | query/form/body | 是 | string | 文件权限，例如 `0644` |

成功响应：JSON 字符串 `"success"`。若路径没有挂载，返回错误：`path {path} 没有挂载`。

## 开发检查

- WebDAV 必须保持标准协议响应，尤其 `PROPFIND` 必须返回 XML。
- 后端修改 `webdavUrl`、`compressUrl`、`permissionUrl`、`webdavBasePath` 时，必须同步检查前端文件管理和微应用文件弹窗。
- `LOCAL_MOCK=true` 只影响 `/host/proc` 路径选择，不跳过 Auth。
- 文件路径要避免让用户自行拼接生产 Agent 代理路径，优先使用 `/pid` 返回值。
- 新增压缩格式、WebDAV 属性、分片字段或挂载文件字段时，需要同步本文档、前端类型和相关测试。
