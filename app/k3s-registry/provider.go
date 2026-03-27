package k3sregistry

import (
	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/app/k3s-registry/http/controller"
	"github.com/w7panel/w7panel/common/middleware"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/console"
	httpserver "github.com/we7coreteam/w7-rangine-go/v2/src/http/server"
)

type Provider struct{}

func (p Provider) Register(httpServer *httpserver.Server, console console.Console) {
	p.RegisterHttpRoutes(httpServer) //
}

func (p Provider) RegisterHttpRoutes(server *httpserver.Server) {
	server.RegisterRouters(func(engine *gin.Engine) {
		// Registry API - 镜像仓库
		registryGroup := engine.Group("")
		registryGroup.Use()
		{
			registryGroup.GET("/v2/*path", controller.Registry{}.Get)
			registryGroup.HEAD("/v2/*path", controller.Registry{}.Get)
			registryGroup.POST("/v2/*path", controller.Registry{}.Post)
			registryGroup.PUT("/v2/*path", controller.Registry{}.Post)
			registryGroup.PATCH("/v2/*path", controller.Registry{}.Post)
			// m.Handle("GET /v2/", r.registryHandler)
			// m.Handle("HEAD /v2/", r.registryHandler)
			// registryGroup.GET("/v2/", controller.Registry{}.Version)
			// registryGroup.GET("/v2/_catalog", controller.Registry{}.Catalog)
			// registryGroup.GET("/v2/:name/tags/list", controller.Registry{}.Tags)
			// registryGroup.GET("/v2/:name/manifests/*reference", controller.Registry{}.Manifest)
			// registryGroup.PUT("/v2/:name/manifests/*reference", controller.Registry{}.PushManifest)
			// registryGroup.GET("/v2/:name/blobs/*digest", controller.Registry{}.Blob)
			// registryGroup.HEAD("/v2/:name/blobs/*digest", controller.Registry{}.BlobExists)
			// registryGroup.POST("/v2/:name/blobs/uploads/", controller.Registry{}.InitUpload)
			// registryGroup.PUT("/v2/:name/blobs/uploads/:uuid", controller.Registry{}.CompleteUpload)
		}

		// Patch API - 容器操作
		patchGroup := engine.Group("/panel-api/v1/k3s-patch")
		patchGroup.Use(middleware.Auth{}.Process)
		{
			patchGroup.GET("/containers", controller.Containers{}.List)
			patchGroup.GET("/containers/:id", controller.Containers{}.Get)
			patchGroup.GET("/containers/:id/layers", controller.Containers{}.Layers)
			patchGroup.POST("/containers/:id/exec", controller.Exec{}.Run)
			patchGroup.POST("/containers/:id/commit", controller.Commit{}.Run)
		}
	})
}
