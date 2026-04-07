# W7Panel 路由重构后端开发详细指南 v4.2

**版本**: 4.2.0  
**创建日期**: 2026-02-21  
**状态**: 待开发

---

## 一、概述

本文档是后端开发的详细实施步骤，基于 v4.2 方案：
- 面板业务接口 → `/panel-api/v1/*`
- K8s 代理接口 → `/k8s-proxy/*`

---

## 二、前置条件

### 2.1 备份代码

```bash
export BACKUP_DIR=/home/wwwroot/w7panel-dev/backup_repo
mkdir -p $BACKUP_DIR
cp -r /home/wwwroot/w7panel-dev/w7panel $BACKUP_DIR/w7panel_backend_$(date +%Y%m%d%H%M%S)
echo "备份完成: $BACKUP_DIR/w7panel_backend_$(date +%Y%m%d%H%M%S)"
```

### 2.2 当前路由结构

| Provider | 路由前缀 | 路由数量 |
|----------|---------|---------|
| auth | `/k8s/*` | ~20 |
| application | `/k8s/*`, `/api/v1/*` | ~50 |
| k3k | `/k8s/k3k/*`, `/k8s/*` | ~20 |
| zpk | `/api/v1/zpk/*` | ~15 |
| metrics | `/k8s/metrics/*` | ~10 |

---

## 三、开发步骤

### Step 1: 创建 K8sResponseFilter 中间件

**文件**: `w7panel/common/middleware/k8s_response_filter.go`

```go
package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type K8sResponseFilter struct {
	middleware.Abstract
	RemoveFields []string
}

func NewK8sResponseFilter() K8sResponseFilter {
	return K8sResponseFilter{
		RemoveFields: []string{
			"managedFields",
		},
	}
}

func (f K8sResponseFilter) Process(ctx *gin.Context) {
	if ctx.Request.Method != "GET" {
		ctx.Next()
		return
	}

	if !strings.Contains(ctx.GetHeader("Accept"), "application/json") {
		ctx.Next()
		return
	}

	blw := &bodyLogWriter{body: bytes.Buffer{}, ResponseWriter: ctx.Writer}
	ctx.Writer = blw
	ctx.Next()

	if blw.body.Len() > 0 {
		content := blw.body.Bytes()
		filtered := f.filterManagedFields(content)
		ctx.Data(blw.statusCode, "application/json", filtered)
	}
}

func (f K8sResponseFilter) filterManagedFields(content []byte) []byte {
	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		slog.Debug("K8sResponseFilter: failed to parse JSON", "error", err)
		return content
	}

	f.removeFieldRecursive(data, "managedFields")

	filtered, err := json.Marshal(data)
	if err != nil {
		slog.Debug("K8sResponseFilter: failed to marshal JSON", "error", err)
		return content
	}

	return filtered
}

func (f K8sResponseFilter) removeFieldRecursive(data interface{}, fieldName string) {
	switch v := data.(type) {
	case map[string]interface{}:
		delete(v, fieldName)
		for key := range v {
			f.removeFieldRecursive(v[key], fieldName)
		}
	case []interface{}:
		for i := range v {
			f.removeFieldRecursive(v[i], fieldName)
		}
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body       bytes.Buffer
	statusCode int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
```

---

### Step 2: 修改 main.go

**文件**: `w7panel/main.go`

#### 2.1 添加 K8s 代理路由组（找到约第 180 行附近）

在 `httpServer.RegisterRouters` 函数中添加：

```go
// 在现有的 RegisterRouters 后面添加新的路由组
httpServer.RegisterRouters(func(engine *gin.Engine) {
	// K8s 代理 API - /k8s-proxy
	k8sProxy := engine.Group("/k8s-proxy")
	k8sProxy.Use(middleware2.Cors{}.Process)
	k8sProxy.Use(middleware2.Auth{}.Process)
	k8sProxy.Use(middleware2.K8sResponseFilter{}.Process)
	
	// 注册 K8s 代理路由
	k8sProxy.GET("/metrics/node", middleware2.Auth{}.Process, middleware2.Proxy{}.Process, metrics2.Metrics{}.NodeHandler)
	k8sProxy.GET("/metrics/pod", middleware2.Auth{}.Process, middleware2.Proxy{}.Process, metrics2.Metrics{}.PodHandler)
	k8sProxy.GET("/metrics/top/node", middleware2.Auth{}.Process, middleware2.Proxy{}.Process, metrics2.Metrics{}.TopNodeHandler)
	k8sProxy.GET("/metrics/namespace/:namespace/pod", middleware2.Auth{}.Process, middleware2.Proxy{}.Process, metrics2.Metrics{}.NamespacePodHandler)
	k8sProxy.GET("/metrics/namespace/:namespace/resource", middleware2.Auth{}.Process, middleware2.Proxy{}.Process, metrics2.Metrics{}.NamespaceResourceHandler)
	
	// K8s 原生 API 透传
	k8sProxy.NoRoute(controller2.Proxy{}.ProxyK8s)
})
```

---

### Step 3: 修改 auth/provider.go

**文件**: `w7panel/app/auth/provider.go`

#### 3.1 修改路由分组（第 36 行）

```go
// 修改前
localApiGroup := engine.Group("/k8s").Use(middleware.Cors{}.Process)

// 修改后
localApiGroup := engine.Group("/panel-api/v1/auth").Use(middleware.Cors{}.Process)
```

#### 3.2 完整路由修改后

```go
func (p Provider) RegisterHttpRoutes(server *httpserver.Server) {
	server.RegisterRouters(func(engine *gin.Engine) {
		localApiGroup := engine.Group("/panel-api/v1/auth").Use(middleware.Cors{}.Process)
		{
			localApiGroup.POST("/login", controller2.Auth{}.Login)
			localApiGroup.POST("/register", controller2.Auth{}.Register)
			localApiGroup.POST("/refresh-token2", controller2.Auth{}.RefreshToken2)
			localApiGroup.POST("/init-user", controller2.Auth{}.InitUser)
			localApiGroup.GET("/init-user", controller2.Auth{}.GetInitUser)
			localApiGroup.POST("/reset-password", middleware.Auth{}.Process, controller2.Auth{}.ResetPassword)
			localApiGroup.POST("/reset-password-current", middleware.Auth{}.Process, controller2.Auth{}.ResetPasswordCurrent)

			localApiGroup.GET("/console/oauth", controller2.Console{}.Redirect)
			localApiGroup.GET("/console/login", controller2.Auth{}.ConsoleLogin)
			localApiGroup.GET("/console/bind", middleware.Auth{}.Process, controller2.Console{}.BindConsole)
			localApiGroup.GET("/console/info", middleware.Auth{}.Process, controller2.Console{}.Info)
			localApiGroup.GET("/console/code/:code", middleware.Auth{}.Process, controller2.Console{}.ProxyCouponCode)
			localApiGroup.Any("/console/proxy/*path", middleware.NewAuth("founder").Process, controller2.Console{}.Proxy)
			localApiGroup.POST("/console/register-to-console", middleware.Auth{}.Process, controller2.Console{}.RegisterToConsole)
			localApiGroup.POST("/console/thirdparty-cd-token", middleware.Auth{}.Process, controller2.Console{}.ThirdPartyCDToken)
			localApiGroup.POST("/console/import-cert", middleware.Auth{}.Process, controller2.Console{}.ImportCert)
			localApiGroup.POST("/console/verify-cert", middleware.Auth{}.Process, controller2.Console{}.VerifyCert)
			localApiGroup.POST("/console/import-cert-console", middleware.Auth{}.Process, controller2.Console{}.ImportCertConsole)
		}
	})
}
```

---

### Step 4: 修改 application/provider.go

**文件**: `w7panel/app/application/provider.go`

这是最复杂的修改，有多个路由组需要修改。

#### 4.1 Helm API 路由（第 126 行）

```go
// 修改前
apiGroup := engine.Group("/api/v1")

// 修改后
apiGroup := engine.Group("/panel-api/v1")
```

#### 4.2 面板业务路由（第 142 行）

```go
// 修改前
localApiGroup := engine.Group("/k8s")

// 修改后
localApiGroup := engine.Group("/panel-api/v1")
```

#### 4.3 GPU 路由（第 190 行）

```go
// 修改前
gpuGroup := engine.Group("/k8s/gpu")

// 修改后
gpuGroup := engine.Group("/panel-api/v1/gpu")
```

#### 4.4 WebDAV 路由（第 206-213 行）

```go
// 修改前
for _, method := range webdavMethods {
	engine.Handle(method, "/k8s/webdav-agent/:pid/agent/*path", ...)
	engine.Handle(method, "/k8s/webdav-agent/:pid/subagent/:subpid/agent/*path", ...)
	engine.Handle(method, "noauth/v1/:name/proxy/*path", ...)
}

// 修改后
for _, method := range webdavMethods {
	engine.Handle(method, "/panel-api/v1/files/webdav-agent/:pid/agent/*path", middleware.Auth{}.Process, controller2.Webdav{}.HandlePid)
	engine.Handle(method, "/panel-api/v1/files/webdav-agent/:pid/subagent/:subpid/agent/*path", middleware.Auth{}.Process, controller2.Webdav{}.HandlePidSubPid)
	engine.Handle(method, "/panel-api/v1/files/webdav/*path", middleware.Auth{}.Process, controller2.Webdav{}.Handle)
	engine.Handle(method, "/k8s-proxy/noauth/v1/:name/proxy/*path", controller2.Proxy{}.ProxyCommon)
}
```

#### 4.5 Compress 路由（第 215-218 行）

```go
// 修改前
engine.POST("/k8s/compress-agent/:pid/compress", ...)
engine.POST("/k8s/compress-agent/:pid/extract", ...)

// 修改后
engine.POST("/panel-api/v1/files/compress-agent/:pid/compress", middleware.Auth{}.Process, controller2.CompressAgent{}.Compress)
engine.POST("/panel-api/v1/files/compress-agent/:pid/extract", middleware.Auth{}.Process, controller2.CompressAgent{}.Extract)
engine.POST("/panel-api/v1/files/compress-agent/:pid/subagent/:subpid/compress", middleware.Auth{}.Process, controller2.CompressAgent{}.Compress)
engine.POST("/panel-api/v1/files/compress-agent/:pid/subagent/:subpid/extract", middleware.Auth{}.Process, controller2.CompressAgent{}.Extract)
```

#### 4.6 Permission 路由（第 220-223 行）

```go
// 修改前
engine.POST("/k8s/permission-agent/:pid/chmod", ...)

// 修改后
engine.POST("/panel-api/v1/files/permission-agent/:pid/chmod", middleware.Auth{}.Process, controller2.PermissionAgent{}.Chmod)
engine.POST("/panel-api/v1/files/permission-agent/:pid/chown", middleware.Auth{}.Process, controller2.PermissionAgent{}.Chown)
engine.POST("/panel-api/v1/files/permission-agent/:pid/subagent/:subpid/chmod", middleware.Auth{}.Process, controller2.PermissionAgent{}.Chmod)
engine.POST("/panel-api/v1/files/permission-agent/:pid/subagent/:subpid/chown", middleware.Auth{}.Process, controller2.PermissionAgent{}.Chown)
```

#### 4.7 其他路由修改

```go
// Kubeconfig
// 修改前: engine.GET("k8s/kubeconfig", ...)
// 修改后:
engine.GET("/panel-api/v1/kubeconfig", middleware.Auth{}.Process, middleware.Proxy{}.Process, controller2.Proxy{}.Kubeconfig)

// S3
// 修改前: engine.Any("/s3bucket", ...)
// 修改后:
engine.Any("/panel-api/v1/s3bucket", middleware.Auth{}.Process, controller2.File{}.Upload).Use(middleware.Cors{}.Process)

// NoAuth ConfigMap
// 修改前: engine.GET("/k8s/noauth/api/v1/namespaces/:namespace/configmaps/:name", ...)
// 修改后:
engine.GET("/panel-api/v1/noauth/namespaces/:namespace/configmaps/:name", controller2.NoAuth{}.GetConfigMap)

// Microapp
// 修改前: engine.GET("/k8s/microapp/top", ...)
// 修改后:
engine.GET("/panel-api/v1/microapp/top", middleware.Auth{}.Process, controller2.MicroApp{}.List)
engine.GET("/panel-api/v1/microapp/:name/proxy/*path", middleware.Auth{}.Process, controller2.Proxy{}.ProxyMicroApp)
```

#### 4.8 添加 K8s 代理路由

在 `RegisterHttpRoutes` 函数末尾添加：

```go
// K8s 代理路由
for _, method := range webdavMethods {
	engine.Handle(method, "/k8s-proxy/v1/namespaces/:namespace/services/:name/proxy-root/*path", middleware.Auth{}.Process, controller2.Proxy{}.ProxyService)
	engine.Handle(method, "/k8s-proxy/v1/namespaces/:namespace/services/:name/proxy/*path", middleware.Auth{}.Process, middleware.Proxy{}.Process, controller2.Proxy{}.ProxyService)
	engine.Handle(method, "/k8s-proxy/v1/namespaces/:namespace/pods/:name/proxy/*path", middleware.Auth{}.Process, middleware.Proxy{}.Process, controller2.Proxy{}.ProxyPod)
	engine.Handle(method, "/k8s-proxy/v1/:name/proxy/*path", middleware.Auth{}.Process, controller2.Proxy{}.ProxyCommon)
}

engine.Any("/k8s-proxy/v1/namespaces/:namespace/services/:name/proxy-no/*path", middleware.ProxyNoAuth{}.Process, controller2.Proxy{}.ProxyNoAuthService)
```

---

### Step 5: 修改 k3k/provider.go

**文件**: `w7panel/app/k3k/provider.go`

#### 5.1 K3K 路由组（第 35 行）

```go
// 修改前
k3kGroup := engine.Group("/k8s/k3k")

// 修改后
k3kGroup := engine.Group("/panel-api/v1/k3k")
```

#### 5.2 超卖路由组（第 65 行）

```go
// 修改前
k3kGroup1 := engine.Group("/k8s/k3k/overselling")

// 修改后
k3kGroup1 := engine.Group("/panel-api/v1/k3k/overselling")
```

#### 5.3 用户信息路由组（第 72 行）

```go
// 修改前
k8kGroup := engine.Group("/k8s")

// 修改后
k8kGroup := engine.Group("/panel-api/v1")
```

---

### Step 6: 修改 zpk/provider.go

**文件**: `w7panel/app/zpk/provider.go`

#### 6.1 ZPK 路由组（第 61 行）

```go
// 修改前
localApiGroup := engine.Group("/api/v1/zpk")

// 修改后
localApiGroup := engine.Group("/panel-api/v1/zpk")
```

---

### Step 7: 修改 metrics/provider.go

**文件**: `w7panel/app/metrics/provider.go`

#### 7.1 面板自采集 Metrics

```go
// 修改前
engine.GET("/k8s/metrics/usage/normal", ...)
engine.GET("/k8s/metrics/usage/disk", ...)

// 修改后
engine.GET("/panel-api/v1/metrics/usage/normal", middleware.Auth{}.Process, controller2.Metrics{}.Usage)
engine.GET("/panel-api/v1/metrics/usage/disk", middleware.Auth{}.Process, controller2.Metrics{}.UsageDisk)
engine.GET("/panel-api/v1/metrics/installed", middleware.Auth{}.Process, controller2.Metrics{}.VmOperatorInstalled)
engine.GET("/panel-api/v1/metrics/state", middleware.Auth{}.Process, controller2.Metrics{}.MetricsState)
```

#### 7.2 K8s 代理 Metrics（这些移到 main.go 或 application/provider.go）

这些路由需要删除，因为在 main.go 中统一处理：

```go
// 删除以下路由
// engine.GET("/k8s/metrics/node", ...)
// engine.GET("/k8s/metrics/pod", ...)
// engine.GET("/k8s/metrics/top/node", ...)
// engine.GET("/k8s/metrics/namespace/:namespace/pod", ...)
// engine.GET("/k8s/metrics/namespace/:namespace/resource", ...)
```

---

## 四、路由对照表汇总

### 面板业务接口 `/panel-api/v1/*`

| 原路径 | 新路径 |
|--------|--------|
| `/k8s/login` | `/panel-api/v1/auth/login` |
| `/k8s/register` | `/panel-api/v1/auth/register` |
| `/k8s/refresh-token2` | `/panel-api/v1/auth/refresh-token2` |
| `/k8s/init-user` | `/panel-api/v1/auth/init-user` |
| `/k8s/userinfo` | `/panel-api/v1/auth/userinfo` |
| `/k8s/console/*` | `/panel-api/v1/console/*` |
| `/k8s/pid` | `/panel-api/v1/pid` |
| `/k8s/tty` | `/panel-api/v1/tty` |
| `/k8s/exec` | `/panel-api/v1/exec` |
| `/k8s/captcha` | `/panel-api/v1/captcha` |
| `/k8s/webdav-agent/*` | `/panel-api/v1/files/webdav-agent/*` |
| `/k8s/compress-agent/*` | `/panel-api/v1/files/compress-agent/*` |
| `/k8s/permission-agent/*` | `/panel-api/v1/files/permission-agent/*` |
| `/api/v1/helm/*` | `/panel-api/v1/helm/*` |
| `/api/v1/zpk/*` | `/panel-api/v1/zpk/*` |
| `/k8s/k3k/*` | `/panel-api/v1/k3k/*` |
| `/k8s/gpu/*` | `/panel-api/v1/gpu/*` |
| `/k8s/metrics/usage/*` | `/panel-api/v1/metrics/usage/*` |

### K8s 代理接口 `/k8s-proxy/*`

| 原路径 | 新路径 |
|--------|--------|
| `/api/v1/*` (NoRoute) | `/k8s-proxy/api/v1/*` |
| `/apis/*` (NoRoute) | `/k8s-proxy/apis/*` |
| `/k8s/metrics/node` | `/k8s-proxy/metrics/node` |
| `/k8s/metrics/pod` | `/k8s-proxy/metrics/pod` |
| `/k8s/v1/namespaces/*/proxy/*` | `/k8s-proxy/v1/namespaces/*/proxy/*` |

---

## 五、编译与验证

### 5.1 编译后端

```bash
cd /home/wwwroot/w7panel-dev/w7panel
go build -o ../dist/w7panel .

if [ $? -eq 0 ]; then
    echo "✅ 后端编译成功"
else
    echo "❌ 后端编译失败"
    exit 1
fi
```

### 5.2 API 测试

```bash
# 启动服务
cd /home/wwwroot/w7panel-dev/dist
export CAPTCHA_ENABLED=false
export LOCAL_MOCK=true
export KO_DATA_PATH=/home/wwwroot/w7panel-dev/dist/kodata
export KUBECONFIG=/home/wwwroot/w7panel-dev/kubeconfig.yaml
./w7panel server:start &

sleep 5

# 测试面板业务 API
curl -X POST "http://localhost:8080/panel-api/v1/auth/login" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin&password=123456"

# 测试 K8s 代理 API
export TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
curl "http://localhost:8080/k8s-proxy/api/v1/namespaces" \
  -H "Authorization: Bearer $TOKEN"

# 测试 managedFields 过滤
curl -s "http://localhost:8080/k8s-proxy/api/v1/namespaces/default/pods" \
  -H "Authorization: Bearer $TOKEN" | jq '.items[0].metadata | has("managedFields")'
# 预期: false
```

---

## 六、文件修改清单

| 文件 | 操作 | 修改内容 |
|------|------|---------|
| `w7panel/common/middleware/k8s_response_filter.go` | 新建 | K8s 响应过滤器 |
| `w7panel/main.go` | 修改 | 添加 `/k8s-proxy` 路由组 |
| `w7panel/app/auth/provider.go` | 修改 | `/k8s` → `/panel-api/v1/auth` |
| `w7panel/app/application/provider.go` | 修改 | 多处路由前缀修改 |
| `w7panel/app/k3k/provider.go` | 修改 | `/k8s/k3k` → `/panel-api/v1/k3k` |
| `w7panel/app/zpk/provider.go` | 修改 | `/api/v1/zpk` → `/panel-api/v1/zpk` |
| `w7panel/app/metrics/provider.go` | 修改 | `/k8s/metrics` → `/panel-api/v1/metrics` |

---

## 七、注意事项

1. **路由优先级**: 具体路由优先于 NoRoute
2. **中间件顺序**: Cors → Auth → K8sResponseFilter
3. **K8s 代理**: 使用 NoRoute 透传，Proxy middleware 处理条件转发
4. **managedFields 过滤**: 仅对 GET 请求和 JSON 响应生效

---

*本文档是后端开发详细步骤，请按顺序执行。完成后请参考前端开发文档进行前端修改。*
