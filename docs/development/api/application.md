# w7panel-server/app/application API 文档

本文档是 `w7panel-server/app/application` 模块的路由归属索引。该模块历史上承载了大量面板核心能力，接口已经按业务专题分散到更适合维护的文档中；新增或修改接口时，应优先更新专题文档，避免在本文件重复维护完整参数表。

## 整体使用方式

`application.md` 不是完整接口明细主文档，而是 `app/application` 模块的路由归属索引。开发时先在这里确认接口属于哪个业务专题，再到对应专题文档维护请求参数、响应参数和调用示例。

### 基本流程

1. 遇到 `app/application` 下的路由，先在“路由归属”表中找到对应路径。
2. 集群、Helm、YAML、终端、代理、DNS、GPU 等接口跳转到 [cluster-ops.md](./cluster-ops.md)。
3. Longhorn、容器文件、容器镜像、微应用静态资源分别跳转到对应专题文档。
4. OpenAPI、验证码、公开站点接口这类轻量能力仍在本文维护。
5. 新增接口时优先判断是否能归入现有专题，避免继续扩大 application 模块文档。

### 场景选择

| 场景 | 维护位置 | 说明 |
|------|----------|------|
| 查路由归属 | 本文 | 快速定位专题文档 |
| 集群运维 | [cluster-ops.md](./cluster-ops.md) | K8s、Helm、YAML、终端、代理、DNS、GPU |
| 存储管理 | [longhorn.md](./longhorn.md) | Longhorn 安装、卷状态和卷操作 |
| 容器文件 | [container-files.md](./container-files.md) | PID、WebDAV、压缩、权限、上传 |
| 容器镜像 | [container-images.md](./container-images.md) | 镜像导出推送等 |
| 微应用静态资源 | [microapp-static.md](./microapp-static.md) | 微应用代理、静态下载和回源 |
| 公开轻量接口 | 本文 | OpenAPI、验证码、公开站点 |

### 使用边界

- 本文只维护归属索引和少量未拆分接口，不重复维护已归入专题的完整参数表。
- 修改路由归属时需要同步更新 [README.md](./README.md) 的入口索引。
- 公开接口必须只返回业务字段，不允许透出完整 Secret、token、密码或敏感 metadata。

## 拆分原则

| 接口类型 | 维护位置 | 说明 |
|----------|----------|------|
| 集群资源、Helm、YAML、终端、代理、DNS、GPU | [cluster-ops.md](./cluster-ops.md) | 集群操作和运维类能力 |
| Longhorn 安装、卷状态和卷操作 | [longhorn.md](./longhorn.md) | Longhorn 存储管理能力 |
| 容器 PID、WebDAV、压缩解压、权限、下载、分片上传、挂载文件 | [container-files.md](./container-files.md) | 容器文件系统和文件传输能力 |
| 微应用、微应用代理、静态资源状态/下载/回源 | [microapp-static.md](./microapp-static.md) | Wujie 微应用和前端静态资源 |
| 容器导出镜像、镜像推送 | [container-images.md](./container-images.md) | 容器镜像相关能力 |
| 公开站点接口、验证码、OpenAPI 文档 | 本文件 | application 模块中尚未单独成篇的轻量接口 |

## 路由归属

### OpenAPI

| 方法 | 路径 | 说明 | 文档 |
|------|------|------|------|
| `GET` | `/docs/openapi` | OpenAPI 文档页面 | 本文件 |
| `GET` | `/docs/openapi/spec` | OpenAPI JSON/YAML 内容 | 本文件 |
| `GET` | `/openapi.json` | OpenAPI JSON 兼容入口 | 本文件 |

### 集群操作

| 方法 | 路径 | 说明 | 文档 |
|------|------|------|------|
| `GET` | `/panel-api/v1/namespaces` | 命名空间列表 | [cluster-ops.md](./cluster-ops.md#命名空间) |
| 多方法 | `/panel-api/v1/helm/releases*` | Helm Release 列表、详情、安装、卸载、更新 | [cluster-ops.md](./cluster-ops.md#helm) |
| `GET` | `/panel-api/v1/app-info` | 面板自身 Helm/Deployment 信息 | [cluster-ops.md](./cluster-ops.md#get-panel-apiv1app-info) |
| `POST` | `/panel-api/v1/yaml` | 应用 Kubernetes YAML | [cluster-ops.md](./cluster-ops.md#yamlcompose-和回滚) |
| `POST` | `/panel-api/v1/kcompose` | Docker Compose 转 Kubernetes YAML | [cluster-ops.md](./cluster-ops.md#yamlcompose-和回滚) |
| `PUT` | `/panel-api/v1/rollback` | 工作负载回滚 | [cluster-ops.md](./cluster-ops.md#yamlcompose-和回滚) |

### 终端与代理

| 方法 | 路径 | 说明 | 文档 |
|------|------|------|------|
| `GET` | `/panel-api/v1/tty` | 面板本地终端 WebSocket | [cluster-ops.md](./cluster-ops.md#终端与执行) |
| `GET` | `/panel-api/v1/nodetty` | 节点终端 WebSocket | [cluster-ops.md](./cluster-ops.md#终端与执行) |
| `GET/POST` | `/panel-api/v1/exec`、`/panel-api/v1/exec2` | Pod 命令执行 | [cluster-ops.md](./cluster-ops.md#终端与执行) |
| `POST` | `/panel-api/v1/exec-all` | 批量命令执行 | [cluster-ops.md](./cluster-ops.md#终端与执行) |
| `GET` | `/panel-api/v1/nodepid` | 节点或 Agent PID 信息 | [cluster-ops.md](./cluster-ops.md#get-panel-apiv1nodepid) |
| 多方法 | `/panel-api/v1/namespaces/:namespace/services/:name/proxy*` | Service 代理 | [cluster-ops.md](./cluster-ops.md#代理接口) |
| 多方法 | `/panel-api/v1/namespaces/:namespace/pods/:name/proxy/*path` | Pod 代理 | [cluster-ops.md](./cluster-ops.md#代理接口) |
| 多方法 | `/panel-api/v1/:name/proxy/*path` | 通用代理 | [cluster-ops.md](./cluster-ops.md#代理接口) |
| `ANY` | `/panel-api/v1/proxy-url/` | 请求指定 URL 并返回文本 | [cluster-ops.md](./cluster-ops.md#get-panel-apiv1proxy-url) |
| `GET` | `/panel-api/v1/kubeconfig` | 生成 kubeconfig | [cluster-ops.md](./cluster-ops.md#get-panel-apiv1kubeconfig) |

### DNS、诊断、Longhorn、GPU

| 方法 | 路径 | 说明 | 文档 |
|------|------|------|------|
| `POST` | `/panel-api/v1/pinyin` | 中文转拼音 | [cluster-ops.md](./cluster-ops.md#dns-和诊断) |
| `GET` | `/panel-api/v1/dnsip` | 域名解析 IP | [cluster-ops.md](./cluster-ops.md#dns-和诊断) |
| `GET` | `/panel-api/v1/dns-cname` | 域名 CNAME | [cluster-ops.md](./cluster-ops.md#dns-和诊断) |
| `GET` | `/panel-api/v1/myip` | 获取出口公网 IP | [cluster-ops.md](./cluster-ops.md#dns-和诊断) |
| `POST` | `/panel-api/v1/db-conn-test` | 数据库连接测试 | [cluster-ops.md](./cluster-ops.md#dns-和诊断) |
| `POST` | `/panel-api/v1/ping-etcd` | etcd health 检查 | [cluster-ops.md](./cluster-ops.md#dns-和诊断) |
| 多方法 | `/panel-api/v1/dns/*` | DNS Zone、Record、Server | [cluster-ops.md](./cluster-ops.md#dns-和诊断) |
| 多方法 | `/panel-api/v1/longhorn/*` | Longhorn 安装、卷状态、卷操作 | [longhorn.md](./longhorn.md) |
| 多方法 | `/panel-api/v1/gpu/*` | GPU 配置、安装、指标、GPUStack worker | [cluster-ops.md](./cluster-ops.md#gpu) |

### 容器文件

| 方法 | 路径 | 说明 | 文档 |
|------|------|------|------|
| `GET` | `/panel-api/v1/pid` | 定位容器 PID，生成 WebDAV/压缩/权限 URL | [container-files.md](./container-files.md#容器定位接口) |
| 多方法 | `/panel-api/v1/files/webdav-agent/*` | WebDAV 文件操作 | [container-files.md](./container-files.md#webdav-接口) |
| `POST` | `/panel-api/v1/files/compress-agent/*` | 压缩和解压 | [container-files.md](./container-files.md#压缩解压) |
| `POST` | `/panel-api/v1/files/permission-agent/*` | chmod/chown | [container-files.md](./container-files.md#权限修改) |
| `GET` | `/panel-api/v1/download/*path` | 文件下载 | [container-files.md](./container-files.md#下载复制和移动) |
| `POST` | `/panel-api/v1/cp` | `kubectl cp` 文件复制 | [container-files.md](./container-files.md#post-panel-apiv1cp) |
| `POST` | `/panel-api/v1/cppid`、`/panel-api/v1/mvpid` | PID rootfs 文件复制 | [container-files.md](./container-files.md#post-panel-apiv1cppid) |
| `POST` | `/panel-api/v1/files/mvtopod` | 临时文件移动到容器 | [container-files.md](./container-files.md#post-panel-apiv1filesmvtopod) |
| 多方法 | `/panel-api/v1/files/*chunk*` | 分片上传、检查、合并 | [container-files.md](./container-files.md#分片上传) |
| 多方法 | `/panel-api/v1/mountfiles*` | ConfigMap/Secret 挂载文件管理 | [container-files.md](./container-files.md#挂载文件接口) |
| `ANY` | `/panel-api/v1/s3bucket`、`/s3bucket` | S3 兼容上传入口 | [container-files.md](./container-files.md#分片上传) |

### 微应用、静态资源和镜像

| 方法 | 路径 | 说明 | 文档 |
|------|------|------|------|
| `GET` | `/panel-api/v1/microapp/top` | 顶部微应用列表 | [microapp-static.md](./microapp-static.md#get-panel-apiv1microapptop) |
| `GET` | `/panel-api/v1/microapp/:name/info` | 微应用详情 | [microapp-static.md](./microapp-static.md#get-panel-apiv1microappnameinfo) |
| `ANY` | `/panel-api/v1/microapp/:name/proxy/*path` | 微应用后端代理 | [microapp-static.md](./microapp-static.md#any-panel-apiv1microappnameproxypath) |
| `GET` | `/panel-api/v1/static/:identifie/status` | 静态资源下载状态 | [microapp-static.md](./microapp-static.md#get-panel-apiv1staticidentifiestatus) |
| `POST` | `/panel-api/v1/static/:namespace/download/:name` | 触发静态资源下载 | [microapp-static.md](./microapp-static.md#post-panel-apiv1staticnamespacedownloadname) |
| `GET` | `/panel-api/v1/static/proxy/:zpkUrl/:identifie/:version/frontend/*path` | 静态资源回源代理 | [microapp-static.md](./microapp-static.md#get-panel-apiv1staticproxyzpkurlidentifieversionfrontendpath) |
| `POST` | `/panel-api/v1/containers/image/export-push` | 容器 rootfs 导出镜像并推送 | [container-images.md](./container-images.md#面板容器镜像导出推送) |

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

- 新增 application 模块接口时，先判断是否属于现有专题文档；属于专题时只在本文件补路由索引。
- 修改 Helm、终端、代理、DNS、GPU 时维护 [cluster-ops.md](./cluster-ops.md)。
- 修改 Longhorn 安装、卷状态或卷操作时维护 [longhorn.md](./longhorn.md)。
- 修改文件、WebDAV、压缩、权限、分片上传或挂载文件时维护 [container-files.md](./container-files.md)。
- 修改微应用或静态资源时维护 [microapp-static.md](./microapp-static.md)。
- 修改容器镜像导出推送时维护 [container-images.md](./container-images.md)。
- 修改公开接口时检查 [credentials.md](./credentials.md) 的公开接口清单是否需要同步。
