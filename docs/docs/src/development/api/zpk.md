# 应用管理与 ZPK API

本文档说明 ZPK 应用配置、列表、安装、升级、构建镜像和应用管理相关 API。ZPK 是面板应用安装和交付的主要入口，底层通常会关联 Helm Release、Kubernetes 资源、容器镜像构建和微应用静态资源。

## 整体使用方式

ZPK 接口用于从应用仓库读取安装配置，并完成安装、升级、构建和应用状态管理。开发时先读取配置确认安装参数，再根据应用类型决定直接安装、先构建镜像，或联动 Helm、容器镜像和静态资源接口。

### 基本流程

1. 调用 `/panel-api/v1/zpk/config` 读取仓库配置，确认应用名称、命名空间、部署项、镜像构建、PVC、域名和依赖项。
2. 需要构建镜像时，调用 ZPK buildimage 接口创建构建任务，并结合 [container-images.md](./container-images.md) 查看镜像能力。
3. 安装或升级应用时，调用 `/panel-api/v1/zpk/install`，后端根据配置应用 Helm 或 Kubernetes 资源。
4. 安装后通过 ZPK 列表、详情、升级等接口管理应用。
5. 如果应用带微前端静态包，再结合 [microapp-static.md](./microapp-static.md) 查询静态资源状态或触发下载。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| 读取安装配置 | `/panel-api/v1/zpk/config` | 根据仓库地址返回安装表单和依赖 |
| 查看已安装应用 | `/panel-api/v1/zpk/` | 按 `source=zpk` 查询 Helm Release |
| 安装或升级 | `/panel-api/v1/zpk/install` | 提交安装参数 |
| 构建镜像 | `/panel-api/v1/zpk/buildimage/*` | 创建镜像构建 Job/CronJob |
| 应用生命周期 | ZPK 管理接口 | 列表、升级、删除或状态查询 |

### 使用边界

- ZPK 配置来自远程或本地仓库，调用前要明确 `repoUrl` 和第三方 CD token 来源。
- 安装和升级会修改 Kubernetes 资源，必须明确 namespace、releaseName 和依赖关系。
- 构建镜像涉及 Registry、containerd、Job/CronJob，失败排查需要联动容器镜像文档。
- 微应用前端包和回源代理不在本文重复维护，见 [microapp-static.md](./microapp-static.md)。

## 通用说明

### 鉴权

除明确标注为回调或公开兼容入口的接口外，ZPK 接口均需要用户 token：

```http
Authorization: Bearer <user-token>
```

### 响应格式

接口直接返回安装配置、Helm Release、环境变量、Job/CronJob 等业务对象，不额外包裹统一 `data` 字段。操作成功时部分接口返回 `"success"`。

### 参数位置

安装、传统应用安装和构建任务通常使用 JSON body 或 form/query 参数混合绑定；具体以各接口请求参数表为准。涉及 URL、OCI 路径和 Chart 名称时必须做好 URL 编码。

## 能力概览

| 能力 | 说明 |
|------|------|
| 安装配置 | 读取 ZPK 安装表单、依赖、域名、存储和镜像配置 |
| 应用列表与安装 | 查询已安装 ZPK 应用，提交安装或升级 |
| 传统应用和外部依赖 | 查询传统应用环境、安装传统应用、读取外部依赖环境 |
| Helm/OCI | 内存仓库、Chart 元信息和 OCI 文件下载 |
| 镜像构建 | 创建构建镜像 Job/CronJob，处理构建成功回调 |
| 本地地址 | 获取本地 ZPK 访问地址 |

## ZPK 配置

### GET `/panel-api/v1/zpk/config`

功能：读取指定 ZPK 仓库地址的应用安装配置，返回应用基础信息、安装依赖、部署参数、构建要求等。

请求参数：

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `repoUrl` | query | string | 是 | ZPK 仓库地址 |
| `thirdpartyCDToken` | query | string | 否 | 第三方持续交付 token |
| `releaseName` | query | string | 否 | 已安装 Release 名称，升级场景可传入 |

响应参数：返回 `PackageAddConfig[]`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `requireBuild` | bool | 是否需要构建镜像 |
| `namespace` | string | 推荐或目标命名空间 |
| `requirePvc` | bool | 是否需要 PVC |
| `requireDomain` | bool | 是否需要配置域名 |
| `requireDomainHttps` | bool | 是否需要 HTTPS |
| `requireDomainForce` | bool | 是否强制配置域名 |
| `startParams` | object | 启动参数配置 |
| `identifie` | string | 应用唯一标识 |
| `name` | string | 应用名称 |
| `ingress` | object | Ingress 配置 |
| `icon` | string | 应用图标地址 |
| `version` | string | 应用版本 |
| `requireInstall` | bool | 是否需要执行安装 |
| `releaseName` | string | Release 名称 |
| `outModuleNames` | array | 对外依赖模块名称列表 |
| `dependsOnes` | array | 依赖项配置 |
| `requireParentReleaseName` | bool | 是否需要父 Release 名称 |
| `parentTitle` | string | 父应用标题 |
| `parentIdentifie` | string | 父应用标识 |
| `deployName` | string | Deployment 名称 |
| `isHelm` | bool | 是否为 Helm 应用 |
| `deployItems` | array | 部署项列表 |
| `isConsole` | bool | 是否为控制台应用 |
| `zipURL` | string | ZIP 包地址 |
| `requireLimit` | bool | 是否需要资源限制配置 |
| `volumesMounts` | array | 容器挂载配置 |
| `volumes` | array | Volume 配置 |
| `isUpgrade` | bool | 是否为升级配置 |
| `installFormulas` | array | 安装公式配置 |

## ZPK 列表

### GET `/panel-api/v1/zpk/`

功能：获取指定命名空间下通过 ZPK 安装的 Helm Release 列表。

请求参数：

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `namespace` | query | string | 否 | 命名空间；未传时由服务端默认逻辑处理 |

响应参数：返回 Helm Release 列表，底层查询固定使用 `source=zpk` 标签过滤。

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | Release 名称 |
| `namespace` | string | Release 所在命名空间 |
| `version` | int | Release 版本号 |
| `chart` | object | Chart 信息 |
| `info` | object | Release 状态信息 |
| `manifest` | string | Release 渲染后的 Manifest |

说明：具体字段结构跟随 Helm Release 对象。

## 安装 ZPK 应用

### PUT `/panel-api/v1/zpk/install`

功能：安装或升级 ZPK 应用，支持普通 ZPK、Helm ZPK、传统应用等安装参数。

请求参数：JSON Body。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `namespace` | string | 是 | 安装命名空间 |
| `repoUrl` | string | 是 | ZPK 仓库地址 |
| `releaseName` | string | 是 | Release 名称 |
| `installOptions` | array | 是 | 安装项配置，见 `InstallOption` |
| `ingressHost` | string | 否 | Ingress 访问域名 |
| `ingressSeletorName` | string | 否 | Ingress selector 名称 |
| `ingressClass` | string | 否 | IngressClass 名称 |
| `ingressForceHttps` | bool | 否 | 是否强制 HTTPS |
| `thirdpartyCDToken` | string | 否 | 第三方持续交付 token |
| `clusterId` | string | 否 | 集群 ID |
| `isTrandition` | bool | 否 | 是否传统应用安装 |
| `zipUrl` | string | 否 | ZIP 包地址 |

`InstallOption` 参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `identifie` | string | 应用或模块标识 |
| `pvcname` | string | PVC 名称 |
| `registry` | string | 镜像仓库地址 |
| `dockerRegistrySecretName` | string | 镜像仓库 Secret 名称 |
| `namespace` | string | 命名空间 |
| `installId` | string | 安装 ID |
| `releaseName` | string | Release 名称 |
| `suffix` | string | 名称后缀 |
| `envkv` | object | 环境变量键值配置 |
| `ingressSelectorName` | string | Ingress selector 名称 |
| `ingressHost` | string | Ingress 访问域名 |
| `ingressClassName` | string | IngressClass 名称 |
| `ingressForceHttps` | bool | 是否强制 HTTPS |
| `replicas` | int | 副本数 |
| `isChild` | bool | 是否子应用 |
| `isUpgrade` | bool | 是否升级安装 |
| `annotations` | object | 注解配置 |
| `serviceAccountName` | string | ServiceAccount 名称 |
| `buildImageSuccessUrl` | string | 镜像构建成功回调地址 |
| `parentReleaseName` | string | 父 Release 名称 |
| `preSubPath` | string | 预置子路径 |
| `cpu` | string | CPU 限制或请求值 |
| `memory` | string | 内存限制或请求值 |
| `volumes` | array | Volume 配置 |
| `volumesMounts` | array | VolumeMount 配置 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `releaseName` | string | Release 名称 |
| `installId` | string | 安装 ID |
| `namespace` | string | 安装命名空间 |

请求示例：

```json
{
  "namespace": "default",
  "repoUrl": "oci://example.com/apps/demo",
  "releaseName": "demo",
  "installOptions": [
    {
      "identifie": "demo",
      "releaseName": "demo",
      "namespace": "default",
      "replicas": 1
    }
  ]
}
```

## 升级信息

### GET `/panel-api/v1/zpk/upgrade-info`

功能：检查指定 ZPK 应用是否存在可升级版本。

请求参数：

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `namespace` | query | string | 是 | Release 所在命名空间 |
| `releaseName` | query | string | 是 | Release 名称 |
| `thirdpartyCDToken` | query | string | 否 | 第三方持续交付 token |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `upgrade` | bool | 是否可升级 |
| `version` | string | 可升级版本 |
| `message` | string | 升级说明或不可升级原因 |
| `data` | object | 升级检查附加数据 |

说明：当配置 `app.upgrade_enable=false` 时，接口返回不可升级结果。

## 镜像构建回调

### ANY `/panel-api/v1/zpk/build-image-success`

功能：镜像构建成功后的回调入口，用于后续更新部署域名、Deployment 等安装状态。

请求参数：

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form | string | 是 | 命名空间 |
| `releaseName` | query/form | string | 是 | Release 名称 |
| `domainHost` | query/form | string | 是 | 访问域名 |
| `deploymentName` | query/form | string | 是 | Deployment 名称 |
| `thirdpartyCDToken` | query/form | string | 否 | 第三方持续交付 token |

响应参数：当前实现成功路径无显式业务响应。

## 传统应用环境

### GET `/panel-api/v1/zpk/trandition/env`

功能：获取传统应用安装环境模板。

请求参数：无。

响应参数：返回环境模板映射，key 为环境标识。

| 字段 | 类型 | 说明 |
|------|------|------|
| `zpkUrl` | string | 环境对应的 ZPK 地址 |
| `icon` | string | 环境图标 |
| `title` | string | 环境标题 |
| `detailUrl` | string | 环境详情地址 |

说明：当前包含 `php7.2` 等传统运行环境。

## 安装传统应用

### POST `/panel-api/v1/zpk/trandition/install`

功能：按传统应用模式安装 ZIP 应用包。

请求参数：Form。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `namespace` | string | 是 | 安装命名空间 |
| `releaseName` | string | 是 | Release 名称 |
| `zpkUrl` | string | 是 | 传统环境 ZPK 地址 |
| `zipUrl` | string | 是 | 应用 ZIP 包地址 |
| `pvcName` | string | 是 | PVC 名称 |
| `entryRoot` | string | 是 | 应用入口目录 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `releaseName` | string | Release 名称 |
| `installId` | string | 安装 ID |
| `namespace` | string | 安装命名空间 |

## 外部依赖环境

### GET `/panel-api/v1/zpk/out-depends/env`

功能：查询指定依赖标识在命名空间中的安装状态和可复用环境信息。

请求参数：

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `namespace` | query | string | 是 | 命名空间 |
| `identifie` | query | string | 是 | 依赖标识 |

响应参数：返回 `DependEnvResult`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `installed` | bool | 依赖是否已安装 |
| `name` | string | 依赖名称 |
| `staticUrl` | string | 静态访问地址 |
| `envs` | object | 环境变量配置 |
| `pvcName` | string | PVC 名称 |
| `preSubPath` | string | 预置子路径 |
| `preVolumeName` | string | 预置 Volume 名称 |
| `replicas` | int | 副本数 |
| `VolumesMounts` | array | VolumeMount 配置 |
| `Volumes` | array | Volume 配置 |

## Helm 内存仓库

### POST `/panel-api/v1/zpk/helm/memory`

功能：创建内存态 Helm ZPK 描述，返回 `memory://` 形式的 ZPK 地址。

请求参数：Form。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `identifie` | string | 否 | 应用标识；为空时由服务端生成 |
| `title` | string | 否 | 应用标题 |
| `icon` | string | 否 | 应用图标 |
| `description` | string | 否 | 应用描述 |
| `chartName` | string | 是 | Chart 名称或 tgz 路径 |
| `repository` | string | 否 | Helm 仓库地址 |
| `version` | string | 否 | Chart 版本 |
| `kv` | object/string | 否 | Helm values 键值配置 |

说明：当 `repository` 为空且 `chartName` 指向 tgz 包时，服务端会解析包内 `Chart.yaml` 补充元信息。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `identifie` | string | 应用标识 |
| `zpkUrl` | string | 内存态 ZPK 地址，格式为 `memory://{identifie}` |

## Chart 元信息

### GET `/panel-api/v1/zpk/helm/chart-yaml`

功能：解析指定 Helm tgz 包中的 `Chart.yaml`。

请求参数：

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `tgzPath` | query | string | 是 | Helm chart tgz 文件路径 |

响应参数：返回解析后的 `Chart.yaml` 对象。

| 字段 | 类型 | 说明 |
|------|------|------|
| `apiVersion` | string | Chart API 版本 |
| `name` | string | Chart 名称 |
| `version` | string | Chart 版本 |
| `appVersion` | string | 应用版本 |
| `description` | string | Chart 描述 |
| `type` | string | Chart 类型 |

说明：实际响应字段跟随 tgz 包内 `Chart.yaml` 内容。

## 最新版本环境

### GET `/panel-api/v1/zpk/last-version/env`

功能：查询指定应用最新版本对应的环境信息。

请求参数：

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `namespace` | query | string | 是 | 命名空间 |
| `name` | query | string | 是 | 应用名称 |

响应参数：同 `GET /panel-api/v1/zpk/out-depends/env` 的 `DependEnvResult`。

## OCI 文件下载

### GET `/panel-api/v1/zpk/oci/down/*oci`

功能：下载 OCI ZPK 中的文件资源，例如图标、`files.json`、代码 ZIP 包等。

鉴权：该接口未注册 Auth 中间件。

请求参数：

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `oci` | path | string | 是 | OCI 资源路径，必须包含 `oci://` |
| `mediaType` | query | string | 否 | 文件媒体类型，用于设置响应 `Content-Type` |

响应参数：文件流。

响应头说明：

| `mediaType` 场景 | `Content-Type` |
|------------------|----------------|
| icon | `image/png` |
| files.json | `application/json` |
| code zip / web-code zip | `application/zip` |

## 构建镜像 Job

### POST `/panel-api/v1/zpk/buildimage/job`

功能：创建一次性 Kubernetes Job，用于构建并推送应用镜像。

请求参数：JSON Body，类型为 `BuildImageParams`。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `dockerRegistry` | object | 否 | 镜像仓库账号信息 |
| `dockerRegistry.host` | string | 否 | 镜像仓库地址 |
| `dockerRegistry.username` | string | 否 | 镜像仓库用户名 |
| `dockerRegistry.password` | string | 否 | 镜像仓库密码 |
| `dockerRegistry.namespace` | string | 否 | 镜像仓库命名空间 |
| `dockerfilePath` | string | 否 | Dockerfile 路径 |
| `buildContext` | string | 否 | 构建上下文路径 |
| `pushImage` | string | 否 | 推送镜像地址 |
| `zipUrl` | string | 否 | 源码 ZIP 地址 |
| `identifie` | string | 否 | 应用标识 |
| `notifyCompletionUrl` | string | 否 | 构建成功通知地址 |
| `notifyFailedUrl` | string | 否 | 构建失败通知地址 |
| `hostNetwork` | bool | 否 | 是否使用宿主机网络 |
| `hostAliases` | array | 否 | Pod HostAliases 配置 |
| `title` | string | 否 | 构建任务标题 |
| `labels` | object | 否 | Job 标签 |
| `buildJobName` | string | 否 | Job 名称 |
| `schedule` | string | 否 | 定时表达式；Job 接口通常不使用 |
| `dockerRegistrySecretName` | string | 否 | 镜像仓库 Secret 名称 |
| `panelRegistryHost` | string | 否 | 面板镜像仓库地址 |

响应参数：返回 Kubernetes `batch/v1.Job` 对象。

| 字段 | 类型 | 说明 |
|------|------|------|
| `metadata` | object | Job 元数据 |
| `spec` | object | Job 期望状态 |
| `status` | object | Job 当前状态 |

## 构建镜像 CronJob

### POST `/panel-api/v1/zpk/buildimage/cronjob`

功能：创建 Kubernetes CronJob，用于定时构建并推送应用镜像。

请求参数：JSON Body，同 `POST /panel-api/v1/zpk/buildimage/job` 的 `BuildImageParams`，其中 `schedule` 用于 CronJob 定时表达式。

响应参数：返回 Kubernetes `batch/v1.CronJob` 对象。

| 字段 | 类型 | 说明 |
|------|------|------|
| `metadata` | object | CronJob 元数据 |
| `spec` | object | CronJob 期望状态 |
| `status` | object | CronJob 当前状态 |

## 本地访问地址

### GET `/panel-api/v1/zpk/local-url`

功能：获取指定 ZPK 实例的本地访问地址、Ingress 状态和 OAuth token。

请求参数：

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `instance` | query | string | 否 | 实例名称，默认 `w7-zpkv2` |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `host` | string | 访问域名；未找到 Ingress 时为空字符串 |
| `hasIngress` | bool/string | 是否存在 Ingress；无 Ingress 的失败路径返回字符串 `"false"` |
| `isHttps` | bool | 是否 HTTPS |
| `oauthToken` | string | OAuth token |

响应示例：

```json
{
  "host": "zpk.example.com",
  "hasIngress": true,
  "isHttps": true,
  "oauthToken": "token"
}
```
