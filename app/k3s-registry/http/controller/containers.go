package controller

import (
	"gitee.com/we7coreteam/k8s-offline/app/k3s-registry/logic"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Containers struct {
	controller.Abstract
}

var containersLogic = logic.NewContainersLogic()

// List 获取容器列表
func (c Containers) List(ctx *gin.Context) {
	containers, err := containersLogic.List(ctx)
	if err != nil {
		c.JsonResponseWithServerError(ctx, err)
		return
	}
	c.JsonResponseWithoutError(ctx, containers)
}

// Get 获取容器详情
func (c Containers) Get(ctx *gin.Context) {
	id := ctx.Param("id")

	container, err := containersLogic.Get(ctx, id)
	if err != nil {
		c.JsonResponseWithError(ctx, err, 404)
		return
	}
	c.JsonResponseWithoutError(ctx, container)
}

// Layers 获取容器镜像层
func (c Containers) Layers(ctx *gin.Context) {
	id := ctx.Param("id")

	layers, err := containersLogic.GetLayers(ctx, id)
	if err != nil {
		c.JsonResponseWithServerError(ctx, err)
		return
	}
	c.JsonResponseWithoutError(ctx, layers)
}
