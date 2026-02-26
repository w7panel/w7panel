package middleware

import (
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type Metrics struct {
	middleware.Abstract
}

func (self Metrics) Process(gin *gin.Context) {
	token := gin.MustGet("k8s_token").(string)
	k3ktoken := k8s.NewK8sToken(token)
	if k3ktoken.IsShared() { //如果是子用户 直接转发agent pod
		client := k8s.NewK8sClient().Sdk

		proxyUrl := "/apis/metrics.k8s.io/v1beta1/namespaces/" + k3ktoken.GetNamespace() + "/pods" + gin.Param("path")
		gin.Request.URL.Path = proxyUrl
		err := client.Proxy(gin.Request, gin.Writer)
		if err != nil {
			self.JsonResponseWithServerError(gin, err)
			return
		}
		return
	}
	gin.Next()
}
