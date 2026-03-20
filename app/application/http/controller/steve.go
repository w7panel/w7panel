package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/rancher/steve/pkg/server"
	"github.com/rancher/steve/pkg/sqlcache/informer/factory"
	"github.com/w7panel/w7panel/common/service/k8s"

	// "github.com/rancher/steve/pkg/ui"

	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Steve struct {
	controller.Abstract
}

func (self Steve) Handler(http *gin.Context) {

	sdk := k8s.NewK8sClient().Sdk
	restConfig, err := sdk.ToRESTConfig()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	server, err := server.New(http, restConfig, &server.Options{
		AuthMiddleware: nil,
		Next:           nil,
		SQLCache:       true,
		SQLCacheFactoryOptions: factory.CacheFactoryOptions{
			GCKeepCount: 1000,
		},
	})
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	server.ServeHTTP(http.Writer, http.Request)

}
