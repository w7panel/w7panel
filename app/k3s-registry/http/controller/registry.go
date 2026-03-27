package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/w7panel/w7panel/app/k3s-registry/logic"
	"github.com/w7panel/w7panel/common/service/registry"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Registry struct {
	controller.Abstract
}

var memoryRegistry = registry.CreateMicroRegistry()

var cdr = registry.ContainerDRegistryHandler

var registryLogic = logic.NewRegistryLogic()

// Version 返回 Registry API 版本
func (c Registry) Version(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"schemas": []string{"https://docs.docker.com/spec/api/v2/"},
	})
}

// Catalog 返回镜像列表
func (c Registry) Get(ctx *gin.Context) {
	values := ctx.Request.URL.Query()
	values.Set("ns", "ccr.ccs.tencentyun.com")
	ctx.Request.URL.RawQuery = values.Encode()
	cdr.ServeHTTP(ctx.Writer, ctx.Request)
}

// Tags 返回镜像标签
func (c Registry) Header(ctx *gin.Context) {
	cdr.ServeHTTP(ctx.Writer, ctx.Request)
}

// InitUpload 初始化 blob 上传
func (c Registry) InitUpload(ctx *gin.Context) {
	// name := ctx.Param("name")

	// uuid, err := registryLogic.InitUpload(ctx, name)
	// if err != nil {
	// 	c.JsonResponseWithServerError(ctx, err)
	// 	return
	// }

	// ctx.Header("Location", "/v2/"+name+"/blobs/uploads/"+uuid)
	// ctx.Header("Range", "bytes=0-0")
	// ctx.JSON(http.StatusAccepted, gin.H{})

	memoryRegistry.ServeHTTP(ctx.Writer, ctx.Request)
}

// CompleteUpload 完成 blob 上传
func (c Registry) CompleteUpload(ctx *gin.Context) {
	// name := ctx.Param("name")
	// uuid := ctx.Param("uuid")
	// digest := ctx.Query("digest")

	// body, _ := ctx.GetRawData()

	// if err := registryLogic.CompleteUpload(ctx, name, uuid, digest, body); err != nil {
	// 	c.JsonResponseWithServerError(ctx, err)
	// 	return
	// }

	// ctx.Header("Location", "/v2/"+name+"/blobs/"+digest)
	// ctx.JSON(http.StatusCreated, gin.H{})
	memoryRegistry.ServeHTTP(ctx.Writer, ctx.Request)
	// nerdctl commit image
}
