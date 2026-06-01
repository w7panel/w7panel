# 面板内应用开发示例

本文档说明如何开发一个可以在 W7Panel 面板中安装和使用的独立应用。这里的“应用”不是给 `w7panel-server` 新增一个业务模块，也不是直接改 `w7panel-ui` 页面，而是类似 [w7panel-sitemanager](https://github.com/w7panel/w7panel-sitemanager) 的独立应用包：应用有自己的后端、前端和镜像，并发布到官方制品库后在面板内运行。复杂应用可以额外提供 Helm Chart；简单应用可以只提供镜像。

## 应用形态

面板内应用通常包含以下部分：

| 部分 | 作用 | 示例 |
|------|------|------|
| Go 后端 | 提供应用自己的 API、任务、数据库逻辑 | `main.go`、`app/application/` |
| 前端 UI | 嵌入面板打开，通常作为微应用页面运行 | `ui/site/`、`ui/environment/` |
| Dockerfile | 构建应用服务镜像，简单应用可只发布镜像 | `Dockerfile` |
| Helm Chart | 可选，用于复杂 Kubernetes 部署编排 | `charts/` |
| 配置文件 | 服务端口、数据库、运行环境变量 | `config.yaml` |
| 发布流程 | 构建镜像、上传官方制品库，按需附加 Helm 和前端包 | `.github/workflows/release.yml` |

以 `w7panel-sitemanager` 为例，它的主要结构如下：

```text
w7panel-sitemanager/
├── main.go                         # 应用入口，加载配置并注册服务
├── config.yaml                     # 应用配置，支持环境变量替换
├── go.mod                          # Go 模块和依赖
├── Dockerfile                      # 应用镜像构建
├── Makefile                        # build、dockerbuild、helm-package、publish
├── app/application/
│   ├── provider.go                 # 注册路由、命令、数据库初始化
│   ├── http/controller/            # HTTP Controller
│   ├── http/middleware/            # 应用鉴权、CORS 等中间件
│   ├── logic/                      # 业务逻辑
│   └── command/                    # 应用命令
├── common/
│   ├── dao/                        # GORM gen DAO
│   ├── entity/                     # 数据实体
│   ├── helper/                     # 工具函数
│   └── service/                    # 外部服务封装
├── database/table.yaml             # 数据表定义
├── ui/
│   ├── site/                       # 站点管理 UI
│   └── environment/                # 环境管理 UI
└── charts/                         # 可选，Helm Chart 和子 Chart
```

## 发布形态选择

应用不强制提供 Helm Chart，应根据复杂度选择发布形态：

| 形态 | 适用场景 | 需要提供 |
|------|----------|----------|
| 镜像应用 | 单容器、少量环境变量、无复杂依赖的简单应用 | Docker 镜像、启动端口、环境变量说明 |
| 镜像 + 前端 | 有面板内 UI，但后端部署仍是单容器 | Docker 镜像、`frontend.zip`、前端入口说明 |
| Helm 应用 | 多容器、PVC、ServiceAccount、RBAC、子 Chart、复杂 values | Docker 镜像、Helm Chart、可选 `frontend.zip` |

建议优先从“镜像应用”开始，只有在以下情况再引入 Helm Chart：

- 需要 PVC、ConfigMap、Secret、ServiceAccount 或 RBAC。
- 需要部署多个 Deployment、Service、Job 或子应用。
- 需要暴露大量可配置 values。
- 需要和其它组件做复杂的 Kubernetes 调度或亲和性配置。

## 开发目标

下面用一个最小的 `demo-manager` 应用说明开发流程。目标是实现：

- 应用后端提供 `/api/project/list` 接口。
- 应用前端在面板内读取 `window.$wujie.props` 传入的面板参数。
- 前端请求应用自身 API 时携带应用 token。
- 前端请求面板或 Kubernetes 资源时使用面板 token。
- 应用构建 Docker 镜像，并发布到官方制品库。
- 简单应用只发布镜像；有 UI 时附加前端资源；复杂应用再附加 Helm Chart。

## 创建仓库

推荐从一个独立仓库开始，而不是在 `w7panel` 主仓库中新增目录：

```bash
mkdir w7panel-demo-manager
cd w7panel-demo-manager
go mod init github.com/your-org/w7panel-demo-manager
```

建议目录：

```text
w7panel-demo-manager/
├── main.go
├── config.yaml
├── Dockerfile
├── Makefile
├── app/application/
│   ├── provider.go
│   ├── http/controller/project.go
│   ├── http/middleware/auth.go
│   └── logic/project.go
├── common/
│   ├── entity/
│   ├── dao/
│   └── service/
├── ui/site/
└── charts/                          # 可选，复杂部署再添加
```

命名建议：

| 项目 | 建议 |
|------|------|
| 仓库名 | `w7panel-{app}` |
| Go module | `github.com/{org}/w7panel-{app}` |
| Chart name | `{app}` 或 `{app}-manager`，仅 Helm 应用需要 |
| 镜像名 | `zpk.w7.cc/public/{app}` 或官方制品库分配的地址 |
| Service 端口 | 默认 `8000`，镜像应用和 Helm 应用都要保持一致 |

## 后端入口

应用后端可以使用 `w7-rangine-go` 启动 HTTP 服务，并在启动时注册应用自己的 Provider。

```go
// main.go
package main

import (
	"bytes"
	_ "embed"

	appProvider "github.com/your-org/w7panel-demo-manager/app/application"
	"github.com/spf13/viper"
	app "github.com/we7coreteam/w7-rangine-go/v2/src"
	"github.com/we7coreteam/w7-rangine-go/v2/src/core/helper"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

//go:embed config.yaml
var configFileContent []byte

func main() {
	newApp := app.NewApp(app.Option{
		DefaultConfigLoader: func(config *viper.Viper) {
			config.SetConfigType("yaml")
			if err := config.MergeConfig(bytes.NewReader(helper.ParseConfigContentEnv(configFileContent))); err != nil {
				panic(err)
			}
		},
	})

	httpServer := new(http.Provider).
		Register(newApp.GetConfig(), newApp.GetConsole(), newApp.GetServerManager()).
		Export()
	httpServer.Use(middleware.GetPanicHandlerMiddleware())

	new(appProvider.Provider).Register(httpServer, newApp.GetConsole())
	newApp.RunConsole()
}
```

配置文件示例：

```yaml
app:
  name: w7-demo-manager
  env: ${APP_ENV-release}
  server: http
server:
  http:
    host: ${W7_DEMO_MANAGER_SERVER_HOST-0.0.0.0}
    port: ${W7_DEMO_MANAGER_SERVER_PORT-8000}
log:
  default:
    driver: stack
    channels:
      - file
      - console
  file:
    driver: file
    path: demo-manager.log
    level: error
  console:
    driver: console
    level: info
setting:
  oauth_token: ${OAUTH_TOKEN-change-me}
```

注意事项：

- `config.yaml` 中的敏感值必须支持通过环境变量覆盖。
- 生产环境不要把默认 token、数据库密码、外部 API 密钥写死。
- 服务端口需要与镜像暴露端口、应用配置端口保持一致；使用 Helm 时还要与 `service.port` 一致。

## 注册应用路由

应用路由通常放在 `/api` 下，这是应用内部 API，不等同于 W7Panel 主面板的 `/panel-api/v1/`。

```go
// app/application/provider.go
package application

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/w7panel-demo-manager/app/application/http/controller"
	"github.com/your-org/w7panel-demo-manager/app/application/http/middleware"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/console"
	httpServer "github.com/we7coreteam/w7-rangine-go/v2/src/http/server"
)

type Provider struct{}

func (p Provider) Register(server *httpServer.Server, consoleManager console.Console) {
	p.RegisterHttpRoutes(server)
}

func (p Provider) RegisterHttpRoutes(server *httpServer.Server) {
	server.RegisterRouters(func(engine *gin.Engine) {
		engine.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{})
		})

		root := engine.Group("/api", middleware.Cors{}.Process)
		root.Match([]string{"POST", "OPTIONS"}, "/project/list", middleware.Auth{}.Process, controller.Project{}.List)
	})
}
```

Controller 示例：

```go
// app/application/http/controller/project.go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Project struct {
	controller.Abstract
}

func (c Project) List(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"items": []gin.H{
			{
				"name":   "demo-project",
				"status": "Running",
			},
		},
	})
}
```

鉴权中间件示例：

```go
// app/application/http/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type Auth struct {
	middleware.Abstract
}

func (a Auth) Process(ctx *gin.Context) {
	expected := facade.GetConfig().GetString("setting.oauth_token")
	token := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	if expected == "" || token != expected {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	ctx.Next()
}
```

约定：

- 应用自身 API 使用应用自己的鉴权方式，例如 `OAUTH_TOKEN`。
- 应用前端调用面板/K8s API 时，使用面板传入的 `paneltoken`。
- 不要把面板用户 token 当作应用内部永久凭证保存。
- CORS、OPTIONS 和错误响应要覆盖面板内 iframe 或微应用调用场景。

## 前端接入面板

面板内应用通常作为微应用打开，前端需要读取 `window.$wujie.props`。`w7panel-sitemanager` 中常见的传入值包括：

| 参数 | 用途 |
|------|------|
| `paneltoken` | 调用面板 API 或 K8s 代理时使用 |
| `OAUTH_TOKEN` | 调用应用自身 API 时使用 |
| `releaseName` | 当前应用 Helm release 名称 |
| `group` / `appgroup` | 应用分组或关联资源前缀 |

应用自身 API 请求封装示例：

```javascript
// ui/site/src/utils/app-api.js
import axios from 'axios'

const appApi = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

appApi.interceptors.request.use((config) => {
  config.headers = config.headers || {}
  config.headers.Authorization = `Bearer ${window.$wujie?.props?.OAUTH_TOKEN || ''}`
  return config
})

export default appApi
```

面板/K8s API 请求封装示例：

```javascript
// ui/site/src/utils/panel-api.js
import axios from 'axios'

const panelApi = axios.create({
  timeout: 10000,
})

panelApi.interceptors.request.use((config) => {
  config.headers = config.headers || {}
  config.headers.Authorization = `Bearer ${window.$wujie?.props?.paneltoken || ''}`

  if (
    config.url?.startsWith('/api/v1') ||
    config.url?.startsWith('/apis/')
  ) {
    config.url = `/k8s-proxy${config.url}`
  }

  return config
})

export default panelApi
```

页面示例：

```vue
<script setup>
import { onMounted, ref } from 'vue'
import appApi from '@/utils/app-api'
import panelApi from '@/utils/panel-api'

const loading = ref(false)
const projects = ref([])
const releaseName = window.$wujie?.props?.releaseName || ''

async function loadProjects() {
  loading.value = true
  try {
    const { data } = await appApi.post('/project/list')
    projects.value = data.items || []
  } finally {
    loading.value = false
  }
}

async function loadCurrentPods() {
  return panelApi.get('/api/v1/namespaces/default/pods', {
    params: {
      labelSelector: `app.kubernetes.io/instance=${releaseName}`,
    },
  })
}

onMounted(loadProjects)
</script>

<template>
  <section>
    <el-button :loading="loading" @click="loadProjects">刷新</el-button>
    <el-table :data="projects">
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="status" label="状态" />
    </el-table>
  </section>
</template>
```

前端注意事项：

- 应用自身 API 和面板/K8s API 要分开封装。
- 请求 K8s 原生资源时通过 `/k8s-proxy/api/v1/*` 或 `/k8s-proxy/apis/*`。
- 不要把 `paneltoken` 写入 localStorage 或 URL。
- 长任务、WebSocket、终端和轮询在页面卸载时必须清理。
- 兼容独立开发模式时，为 `window.$wujie` 缺失提供 mock props。

本地开发 mock 示例：

```javascript
if (!window.$wujie) {
  window.$wujie = {
    props: {
      paneltoken: import.meta.env?.VITE_PANEL_TOKEN || '',
      OAUTH_TOKEN: 'change-me',
      releaseName: 'demo-manager',
      group: 'demo-manager',
      appgroup: 'demo-manager',
    },
  }
}
```

## 使用面板默认 Wujie 事件

面板已通过 Wujie bus 为微应用提供常用事件，完整事件、参数和回调说明见 [frontend/wujie-events.md](../frontend/wujie-events.md)。应用前端只需要在微应用环境中调用 `window.$wujie.bus.$emit()`，不需要自己实现面板弹窗、文件管理、日志弹窗等能力。

建议先封装一个小工具，避免页面中到处直接访问全局对象：

```javascript
// ui/site/src/utils/wujie-events.js
export function emitPanelEvent(eventName, payload, callback, rejectCallback) {
  const bus = window.$wujie?.bus
  if (!bus) {
    console.warn(`[wujie] bus not found: ${eventName}`)
    return
  }

  bus.$emit(eventName, payload, callback, rejectCallback)
}

export function onPanelEvent(eventName, handler) {
  const bus = window.$wujie?.bus
  if (!bus) {
    return () => {}
  }

  bus.$on(eventName, handler)
  return () => bus.$off?.(eventName, handler)
}
```

打开面板弹窗页面：

```javascript
import { emitPanelEvent } from '@/utils/wujie-events'

export function openHelpPage() {
  emitPanelEvent('openPage', {
    title: '帮助文档',
    src: 'https://example.com/help',
  })
}
```

打开当前应用关联容器的文件管理：

```javascript
import { emitPanelEvent } from '@/utils/wujie-events'

export function openAppFiles() {
  emitPanelEvent('openFile', {
    kind: 'deployments',
    appname: window.$wujie?.props?.releaseName || 'demo-manager',
    path: '/home',
  })
}
```

打开 Pod 日志弹窗：

```javascript
import { emitPanelEvent } from '@/utils/wujie-events'

export function openPodLog(pod) {
  emitPanelEvent('podLog', {
    namespace: pod.metadata.namespace,
    name: pod.metadata.name,
    container: pod.spec?.containers?.[0]?.name,
  })
}
```

调用带回调的压缩下载事件：

```javascript
import { emitPanelEvent } from '@/utils/wujie-events'

export function downloadLogs(pidInfo) {
  emitPanelEvent('zip', {
    pid: pidInfo,
    input: ['/var/log/app'],
    output: '/tmp/app-logs.tar.gz',
  }, (result) => {
    if (result?.link) {
      window.open(result.link)
    }
  })
}
```

检查面板会话并接收回调：

```javascript
import { emitPanelEvent } from '@/utils/wujie-events'

export function refreshPanelToken() {
  emitPanelEvent('checkSession', (token) => {
    if (!token) {
      return
    }
    // 只在内存中使用 token，不写入 localStorage 或 URL。
  })
}
```

监听面板下发给微应用的事件：

```javascript
import { onMounted, onUnmounted } from 'vue'
import { onPanelEvent } from '@/utils/wujie-events'

let offRouteChange = null

onMounted(() => {
  offRouteChange = onPanelEvent('routeChange', (path) => {
    // 根据面板菜单或容器路由变化更新微应用内部路由。
  })
})

onUnmounted(() => {
  offRouteChange?.()
})
```

使用约定：

- 优先使用面板已有事件，例如 `openPage`、`openApp`、`openFile`、`toFile`、`podLog`、`zip`、`uploadFile`、`checkSession`。
- 入参使用对象，只有 `toStoreInstall`、`changeMenu` 等简单事件才传字符串。
- 多数带回调事件不是 Promise，按事件说明通过第二个或第三个参数接收结果。
- 不要把 `paneltoken`、`OAUTH_TOKEN` 或事件回调返回的 token 持久化到 localStorage、sessionStorage、URL。
- 页面卸载时清理 `$on` 注册的监听，避免微应用切换后重复响应。

## 构建前端资源

如果应用有独立前端目录：

```bash
cd ui/site
npm install
npm run build
```

发布时通常将构建结果打包成 `frontend.zip`：

```bash
cd ui/site/dist
zip -r ../../../frontend.zip .
```

说明：

- `frontend.zip` 是面板/ZPK 可识别的前端附件之一。
- 如果应用有多个前端入口，可以分别构建，也可以按应用约定打包到一个前端附件中。
- 前端构建产物不要依赖本机绝对路径。

## Docker 镜像

最小 Dockerfile 示例：

```dockerfile
FROM alpine

ENV TZ=Asia/Shanghai

COPY ./builder/server /home/demo-manager
COPY ./config.yaml /home/config.yaml

RUN mkdir -p /home/demo-manager-data

EXPOSE 8000

CMD ["/home/demo-manager", "server:start", "-f", "/home/config.yaml"]
```

Makefile 示例：

```makefile
IMAGE_REPOSITORY ?= zpk.w7.cc/public/demo-manager
IMAGE_TAG ?= v0.1.0
IMAGE_TARGET ?= $(IMAGE_REPOSITORY):$(IMAGE_TAG)

.PHONY: tidy build dockerbuild dev

tidy:
	go mod tidy

build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-w -s" -o builder/server .

dockerbuild:
	docker build -t $(IMAGE_TARGET) .

dev:
	go run . server:start
```

如果使用 SQLite、CGO 或 musl 静态编译，需要按实际依赖调整 `CGO_ENABLED`、`CC`、基础镜像和运行时库。

## 发布到官方制品库

简单应用可以只构建镜像并发布到官方制品库，不需要维护 `charts/`。

基本流程：

1. 构建 Go 二进制。
2. 构建 Docker 镜像。
3. 使用版本号给镜像打 tag。
4. 登录官方制品库。
5. 推送镜像。
6. 在应用制品信息中声明镜像地址、启动端口和必要环境变量。

命令示例：

```bash
make build
docker build -t zpk.w7.cc/public/demo-manager:v0.1.0 .
docker login zpk.w7.cc
docker push zpk.w7.cc/public/demo-manager:v0.1.0
```

镜像应用需要明确以下信息：

| 项目 | 示例 | 说明 |
|------|------|------|
| image | `zpk.w7.cc/public/demo-manager:v0.1.0` | 应用镜像 |
| port | `8000` | 应用 HTTP 端口 |
| health | `/health` | 健康检查路径 |
| env | `OAUTH_TOKEN` | 必需环境变量 |
| frontend | `frontend.zip` | 有面板 UI 时提供 |

注意事项：

- 镜像 tag 必须稳定，不要使用会被覆盖的临时 tag 作为正式版本。
- 官方制品库中的应用版本应能追溯到 Git tag 或 commit。
- 环境变量默认值只能用于开发，生产敏感值必须由安装时配置。
- 如果应用需要 PVC、RBAC、多 Service 等复杂资源，改用 Helm 应用形态。

## Helm Chart（可选）

Helm Chart 不是所有应用都必须提供。只有当镜像应用无法表达部署需求时，再提供 `charts/`，让面板按 Chart 安装复杂资源。

最小 `charts/Chart.yaml`：

```yaml
apiVersion: v2
name: demo-manager
description: Demo manager for W7Panel
type: application
version: 0.1.0
appVersion: "v0.1.0"
```

最小 `charts/values.yaml`：

```yaml
replicaCount: 1

image:
  repository: zpk.w7.cc/public/demo-manager
  pullPolicy: IfNotPresent
  tag: v0.1.0

service:
  type: ClusterIP
  port: 8000

env:
  APP_ENV: release
  W7_DEMO_MANAGER_SERVER_PORT: "8000"
  OAUTH_TOKEN: change-me

resources: {}
```

Deployment 模板要点：

```yaml
containers:
  - name: demo-manager
    image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
    imagePullPolicy: {{ .Values.image.pullPolicy }}
    ports:
      - name: http
        containerPort: {{ .Values.service.port }}
    env:
      - name: APP_ENV
        value: {{ .Values.env.APP_ENV | quote }}
      - name: OAUTH_TOKEN
        value: {{ .Values.env.OAUTH_TOKEN | quote }}
```

Chart 检查：

```bash
helm lint charts
helm template demo-manager charts
helm package charts --destination charts
```

注意事项：

- `Chart.yaml` 的 `version` 和镜像 tag 发布时要同步。
- Service 端口、容器端口、应用配置端口必须一致。
- 需要持久化数据时增加 PVC，并在 `values.yaml` 暴露 storageClass、size、accessMode。
- 需要调用 Kubernetes API 时配置 ServiceAccount 和 RBAC，不要默认给过大权限。

发布检查：

- tag、镜像 tag、应用制品版本是否一致。
- 使用 Helm 时，Chart `version`、Chart `appVersion` 是否与镜像 tag 同步。
- Docker registry 凭证是否可用。
- ZPK 账号是否有目标 artifact 权限。
- 镜像制品是否已推送到官方制品库。
- 有 UI 时，前端包是否已附加。
- 使用 Helm 时，Helm 包是否已附加。
- 发布后在面板应用市场或安装入口能看到新版本。

## 面板内联调

应用安装到面板后，按以下路径排查：

| 检查项 | 方法 |
|--------|------|
| Pod 是否启动 | `kubectl get pods -n default -l app.kubernetes.io/instance=<release>` |
| Service 是否正常 | `kubectl get svc -n default` |
| 后端健康检查 | 请求应用 `/health` |
| 应用 API 鉴权 | 请求 `/api/project/list`，检查 `OAUTH_TOKEN` |
| 前端是否加载 | 面板中打开应用页面，检查 Network |
| 面板 token 是否传入 | 浏览器 console 检查 `window.$wujie.props.paneltoken` 是否存在 |
| K8s 代理是否正确 | Network 请求应走 `/k8s-proxy/api/v1/*` 或 `/k8s-proxy/apis/*` |

常用命令：

```bash
kubectl logs -n default deploy/<release-name>
kubectl describe pod -n default <pod-name>
helm get values <release-name> -n default       # Helm 应用使用
helm get manifest <release-name> -n default    # Helm 应用使用
```

## 安全规范

- `OAUTH_TOKEN`、面板 token、registry 密码、数据库密码不得写入前端仓库、日志或 URL。
- 应用内部 API 必须做鉴权，不能只依赖“只能从面板打开”。
- 前端调用 K8s 代理必须使用面板传入的 `paneltoken`，不要复用应用内部 token。
- 后端代理外部 URL 时必须限制目标，避免 SSRF。
- 文件、压缩、命令执行、终端类能力必须限制路径、命令和权限范围。
- 使用 Helm 时，Chart 中默认 RBAC 权限按最小权限配置。

## 常见问题

| 问题 | 常见原因 | 处理 |
|------|----------|------|
| 面板打开空白 | 前端包路径不对或 `frontend.zip` 内容层级错误 | 确认 zip 根目录包含 `index.html` 和静态资源 |
| 应用 API 401 | `OAUTH_TOKEN` 未传入或前后端值不一致 | 检查应用环境变量、`window.$wujie.props.OAUTH_TOKEN` 和请求头 |
| K8s API 401/403 | `paneltoken` 缺失或权限不足 | 检查 Wujie props、用户权限、目标资源 RBAC |
| Pod 启动失败 | 镜像 tag 不存在、配置缺失、端口不一致 | 查看 `kubectl describe pod` 和容器日志 |
| Helm 安装失败 | Chart 模板错误或 values 缺字段 | 执行 `helm lint` 和 `helm template` |
| 发布后版本没变 | 制品版本、镜像 tag 或 Chart 版本未同步 | 检查 Makefile、release workflow 和 ZPK 附件 |

## 提交前清单

```text
□ 应用是独立仓库，不直接修改 w7panel-server/w7panel-ui 主工程
□ 后端 `/health` 和应用 API 可用
□ 应用 API 已实现鉴权，默认 token 可通过环境变量覆盖
□ 前端区分应用自身 API 和面板/K8s API
□ 前端没有持久化 paneltoken、OAUTH_TOKEN 等敏感值
□ Docker 镜像可以构建并启动
□ 镜像已发布到官方制品库，tag 与应用版本一致
□ frontend.zip 根目录结构正确
□ 简单应用不强制提供 Helm Chart
□ 使用 Helm 时，Chart 通过 helm lint 和 helm template
□ ZPK 发布包含所需附件：镜像信息、可选 frontend、可选 helm
□ README 或开发文档记录安装、配置、联调和发布方式
```

## 参考实现

- [w7panel-sitemanager](https://github.com/w7panel/w7panel-sitemanager)：包含 Go 后端、前端 UI、Dockerfile、Helm Chart、ZPK 发布流程，是复杂面板内应用开发的完整参考。简单应用可只参考其中的后端、前端和镜像发布部分。
