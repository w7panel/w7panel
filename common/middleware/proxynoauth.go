package middleware

import (
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

// proxy-no路由 占用了header Authorization
type ProxyNoAuth struct {
	middleware.Abstract
}

func (self ProxyNoAuth) Process(gin *gin.Context) {
	iftoken := helper.GetToken(gin)
	if iftoken == "" {
		gin.Next()
		return
	}

	token := iftoken
	k3ktoken := k8s.NewK8sToken(token)
	if k3ktoken.IsK3kCluster() && !helper.IsChildAgent() { //如果是子用户 直接转发agent pod
		config, err := k3ktoken.GetK3kConfig()
		if err != nil {
			self.JsonResponseWithServerError(gin, err)
			return
		}
		path := gin.Request.URL.String()
		// agentHost := config.GetK3kAgentLbHost()
		proxyUrl := "http://" + config.GetK3kAgentLbHost()
		saName, err := k3ktoken.GetSaName()

		headers := map[string]string{
			"PANEL_ROLE":     k3ktoken.GetRole(), //当前角色
			"PANEL_USERNAME": saName,             //当前用户
		}
		proxy, err := helper.ProxyUrl(proxyUrl, path, "", headers, nil)
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
