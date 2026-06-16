# w7panel-server/app/application API 文档

本文档只维护 `w7panel-server/app/application` 模块中尚未归入其他 API 专题的接口。该模块历史上承载了大量面板核心能力，集群、文件、镜像、Longhorn、微应用等接口已经拆到独立专题，本文不再重复维护。

## 整体使用方式

`application.md` 不是 application 模块完整路由清单。开发时如果接口已经能归入现有专题，只维护对应专题文档；只有 OpenAPI、验证码、公开站点这类轻量接口继续放在本文。

### 基本流程

1. 新增或修改 OpenAPI、验证码、公开站点接口时，维护本文。
2. 新增或修改集群、Helm、YAML、终端、代理、DNS、GPU、Longhorn、容器文件、容器镜像、微应用静态资源等接口时，维护对应专题文档。
3. 新增 application 模块路由时，先判断是否能归入现有专题；能归入专题的，不在本文重复记录。

### 场景选择

| 场景 | 维护位置 | 说明 |
|------|----------|------|
| 集群运维 | [cluster-ops.md](./cluster-ops.md) | K8s、Helm、YAML、终端、代理、DNS、GPU |
| 存储管理 | [longhorn.md](./longhorn.md) | Longhorn 安装、卷状态和卷操作 |
| 容器文件 | [container-files.md](./container-files.md) | PID、WebDAV、压缩、权限、上传 |
| 容器镜像 | [container-images.md](./container-images.md) | 镜像导出推送等 |
| 微应用静态资源 | [microapp-static.md](./microapp-static.md) | 微应用代理、静态下载和回源 |
| OpenAPI、验证码、公开站点 | 本文 | 未单独成篇的轻量接口 |

### 使用边界

- 本文只维护未拆分接口，不维护已归入专题的路由索引和参数表。
- 如果本文接口后续拆成独立专题，需要从本文移除接口明细，并同步更新 [index.md](./) 的入口索引。
- 公开接口必须只返回业务字段，不允许透出完整 Secret、token、密码或敏感 metadata。

## 通用说明

### 鉴权

OpenAPI、验证码和公开站点接口均无需用户 token。公开接口只能返回允许公开的业务字段。

### 响应格式

OpenAPI 页面返回 HTML，OpenAPI spec 返回 JSON。验证码和公开站点接口按 Controller 实际返回业务对象或字符串，不额外包裹统一 `data` 字段。

### 参数位置

本文接口参数位置较少：OpenAPI 页面入口的 `url` 使用 query；验证码校验的 `point`、`key` 使用 form；公开 ConfigMap 读取接口的 `name` 使用 path。其它接口无请求参数。

### 安全边界

- 公开站点接口不得返回完整 K8s Secret、ConfigMap metadata、managedFields、token、密码或密钥。
- 验证码接口只负责生成和校验验证码，登录接口是否强制验证码由 `captcha.enabled` 控制。

## 能力概览

| 能力 | 说明 |
|------|------|
| OpenAPI 文档 | 提供 OpenAPI 页面和 JSON spec 兼容入口 |
| 验证码 | 生成和校验滑块验证码，登录开启验证码时使用 |
| 公开站点接口 | 备案、初始化状态、联系信息和允许公开的 ConfigMap 业务字段 |

## OpenAPI 文档

### GET `/docs/openapi`

功能：打开 OpenAPI 文档页面。

认证：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `url` | query | 否 | string | OpenAPI spec 地址，默认 `/docs/openapi/spec` |

响应参数：HTML 页面。

### GET `/docs/openapi/spec`

功能：返回 OpenAPI 规格内容。

认证：无需用户 token。

请求参数：无。

响应参数：OpenAPI JSON。

### GET `/openapi.json`

功能：OpenAPI JSON 兼容入口。

认证：无需用户 token。

请求参数：无。

响应参数：OpenAPI JSON。

## 验证码接口

### GET `/panel-api/v1/captcha`

功能：生成滑块验证码。

认证：无需用户 token。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | number | 固定为 `0` |
| `captcha_key` | string | 加密后的验证码 key |
| `image_base64` | string | 背景图 Base64 |
| `tile_base64` | string | 滑块图 Base64 |
| `tile_width` | number | 滑块宽度 |
| `tile_height` | number | 滑块高度 |
| `tile_x` | number | 滑块目标 X 坐标 |
| `tile_y` | number | 滑块目标 Y 坐标 |

### POST `/panel-api/v1/verify-captcha`

功能：验证滑块验证码。

认证：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `point` | form | 是 | string | 前端提交的滑块坐标 |
| `key` | form | 是 | string | `/captcha` 返回的 `captcha_key` |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `ok` | bool | 是否验证通过 |
| `msg` | string | 错误信息，验证失败时返回 |

登录接口在 `captcha.enabled=true` 时会要求携带 `point` 和 `key`，详见 [credentials.md](./credentials.md)。

## 公开站点接口

公开站点接口无需用户 token，但必须只返回业务允许公开的字段，不允许返回完整 Secret、token、密码或带敏感 metadata 的 K8s 对象。

### GET `/panel-api/v1/noauth/site/beian`

功能：从 `default/beian` ConfigMap 获取备案信息。

认证：无需用户 token。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `icpnumber` | string | ICP 备案号 |
| `number` | string | 公安备案号 |
| `location` | string | 备案地区 |

如果 ConfigMap 不存在，返回 `"success"`。

### GET `/panel-api/v1/noauth/site/beian2`

功能：使用内部 Kubernetes client 获取备案信息。

认证：无需用户 token。

请求参数：无。

响应参数：同 `/panel-api/v1/noauth/site/beian`。

### GET `/panel-api/v1/noauth/site/k3k-config`

功能：获取 K3k 公开配置。

认证：无需用户 token。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `indexpage` | string | 首页类型或登录页配置 |

如果 ConfigMap 不存在，返回 `"success"`。

### GET `/panel-api/v1/noauth/site/init-user`

功能：获取初始化用户、Console 注册和验证码开关状态。

认证：无需用户 token。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `canInitUser` | string | 是否允许初始化用户，字符串布尔值 |
| `allowConsoleRegister` | string | 是否允许 Console 注册，字符串布尔值 |
| `captchaEnabled` | string | 是否开启验证码，字符串布尔值 |

### GET `/panel-api/v1/noauth/site/lianxi`

功能：获取联系信息 ConfigMap 列表。

认证：无需用户 token。

请求参数：无。

响应参数：Kubernetes `corev1.ConfigMapList`。查询失败返回空 `ConfigMapList`。

### GET `/panel-api/v1/noauth/site/{name}/configmap`

功能：获取允许公开访问的 ConfigMap。仅当 ConfigMap label `w7.cc/noauth=true` 时返回真实对象。

认证：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | ConfigMap 名称 |

响应参数：Kubernetes `corev1.ConfigMap`；不存在或未授权公开时返回空 ConfigMap。

## 维护检查

- 新增 application 模块接口时，先判断是否属于现有专题文档；属于专题时只维护专题文档，不在本文补充。
- 修改 Helm、终端、代理、DNS、GPU 时维护 [cluster-ops.md](./cluster-ops.md)。
- 修改 Longhorn 安装、卷状态或卷操作时维护 [longhorn.md](./longhorn.md)。
- 修改文件、WebDAV、压缩、权限、分片上传或挂载文件时维护 [container-files.md](./container-files.md)。
- 修改微应用或静态资源时维护 [microapp-static.md](./microapp-static.md)。
- 修改容器镜像导出推送时维护 [container-images.md](./container-images.md)。
- 修改公开接口时检查 [credentials.md](./credentials.md) 的公开接口说明是否需要同步。
