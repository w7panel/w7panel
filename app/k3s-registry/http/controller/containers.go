package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/common/service/registry"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Containers struct {
	controller.Abstract
}

// List 获取容器列表
func (self Containers) List(ctx *gin.Context) {
	client, err := registry.CreateClient()
	if err != nil {
		self.JsonResponseWithServerError(ctx, err)
		return
	}
	list, err := client.ContainerService().List(ctx)
	if err != nil {
		self.JsonResponseWithServerError(ctx, err)
		return
	}
	defer client.Close()
	self.JsonResponseWithoutError(ctx, list)
}

// Get 获取容器详情
func (self Containers) Get(ctx *gin.Context) {
	id := ctx.Param("id")

	client, err := registry.CreateClient()
	if err != nil {
		self.JsonResponseWithServerError(ctx, err)
		return
	}
	data, err := client.LoadContainer(ctx, id)
	if err != nil {
		self.JsonResponseWithServerError(ctx, err)
		return
	}
	defer client.Close()
	self.JsonResponseWithoutError(ctx, data)
}

// Layers 获取容器镜像层
func (self Containers) Layers(ctx *gin.Context) {
	id := ctx.Param("id")

	client, err := registry.CreateClient()
	if err != nil {
		self.JsonResponseWithServerError(ctx, err)
		return
	}
	data, err := client.ImageService().Get(ctx, id)
	if err != nil {
		self.JsonResponseWithServerError(ctx, err)
		return
	}
	defer client.Close()
	self.JsonResponseWithoutError(ctx, data)
}
