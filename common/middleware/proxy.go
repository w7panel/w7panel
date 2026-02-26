package middleware

import (
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type Proxy struct {
	middleware.Abstract
}

func (self Proxy) Process(gin *gin.Context) {
	iftoken, exites := gin.Get("k8s_token")
	if !exites {
		gin.Next()
		return
	}
	token := iftoken.(string)
	k3ktoken := k8s.NewK8sToken(token)
	if k3ktoken.IsK3kCluster() && !helper.IsChildAgent() { //如果是子用户 直接转发agent pod
		config, err := k3ktoken.GetK3kConfig()
		if err != nil {
			self.JsonResponseWithServerError(gin, err)
			return
		}
		path := gin.Request.URL.String()
		agentHost := config.GetK3kAgentInnerIngressHost()
		proxyUrl := "http://" + config.GetVirtualIngressServiceName()
		auth := gin.Request.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			client, err := k8s.NewK8sClient().Channel(token)
			if err != nil {
				self.JsonResponseWithServerError(gin, err)
				return
			}
			restConfig, err := client.ToRESTConfig()
			if err != nil {
				self.JsonResponseWithServerError(gin, err)
				return
			}
			gin.Request.Header.Set("Authorization", "Bearer "+restConfig.BearerToken)
		}

		proxy, err := helper.ProxyUrl(proxyUrl, path, agentHost, nil, nil)
		if err != nil {
			self.JsonResponseWithServerError(gin, err)
			return
		}
		proxy.ServeHTTP(gin.Writer, gin.Request)
		gin.Abort()
		return
	}
	gin.Next()
}
