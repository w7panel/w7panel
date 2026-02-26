package middleware

import (
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type K8sFilter struct {
	middleware.Abstract
}

func (self K8sFilter) Process(ctx *gin.Context) {
	// 路由归一化：仅对 JSON API 路径应用后续 K8s 相关过滤与同步
	routeType := ctx.GetString("route_type")
	if routeType != "json_api" {
		ctx.Next()
		return
	}
	token := ctx.MustGet("k8s_token").(string)
	k3ktoken := k8s.NewK8sToken(token)
	if !k3ktoken.IsK3kCluster() {
		ctx.Next()
		return
	}

	// k3kConfig, err := k3ktoken.GetK3kConfig()
	// if err != nil {
	// 	ctx.Next()
	// 	return
	// }

	path := ctx.Request.URL.Path
	namespace := ""
	resource := ""
	name := ""
	if strings.HasPrefix(path, "/apis") || strings.HasPrefix(path, "/api") {
		parts := strings.Split(path, "/")
		if len(parts) >= 4 {
			group := ""
			version := parts[2]

			if strings.HasPrefix(path, "/apis") {
				group = parts[2]
				version = parts[3]
				if len(parts) >= 6 && parts[4] == "namespaces" {
					namespace = parts[5]
					if len(parts) >= 7 {
						resource = parts[6]
						if len(parts) >= 8 {
							name = parts[7]
						}
					}
				}
			} else if strings.HasPrefix(path, "/api") {
				if len(parts) >= 5 && parts[3] == "namespaces" {
					namespace = parts[4]
					if len(parts) >= 6 {
						resource = parts[5]
						if len(parts) >= 7 {
							name = parts[6]
						}
					}
				}
			}

			// 存储到上下文
			ctx.Set("k8s_group", group)
			ctx.Set("k8s_version", version)
			ctx.Set("k8s_namespace", namespace)
			ctx.Set("k8s_resource", resource)
			ctx.Set("k8s_name", name)
		}
	}
	ctx.Next()
	// self.sync(namespace, resource, name, ctx, k3kConfig, k3ktoken) helm安装就取不到数据

}

func (K8sFilter) sync(namespace string, resource string, name string, ctx *gin.Context, k3kConfig *k8s.K3kConfig, k3ktoken *k8s.K8sToken) bool {
	if namespace == "" || resource == "" || name == "" {
		return true
	}

	go doSync(ctx, name, namespace, k3kConfig, k3ktoken, resource)
	return false
}

func doSync(ctx *gin.Context, name string, namespace string, k3kConfig *k8s.K3kConfig, k3ktoken *k8s.K8sToken, resource string) {
	if ctx.Request.Method == "POST" || ctx.Request.Method == "PUT" || ctx.Request.Method == "DELETE" || ctx.Request.Method == "PATCH" {
		k3kSync := &k3k.K3kSync{
			VirtualName:      name,
			VirtualNamespace: namespace,
			K3kName:          k3kConfig.Name,
			K3kNamespace:     k3kConfig.Namespace,
			K3kMode:          k3ktoken.K3kMode(),
		}
		if resource == "ingresses" {
			k3k.SyncIngress(k3kSync)
		}
		if resource == "configmaps" {
			k3k.SyncConfigmap(k3kSync)
		}
		if resource == "secret" {
			k3k.SyncSecret(k3kSync)
		}
		if resource == "mcpbridges" && k3ktoken.IsShared() {
			k3k.SyncMcpBridge(k3kSync)
		}
	}
}
