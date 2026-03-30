package controller

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/common/service/registry"

	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Registry struct {
	controller.Abstract
}

var regisry *registry.RegistryHandler

func init() {
	reg, err := registry.InitReigstry(context.Background())
	if err != nil {
		slog.Error("init registry err", "err", err)
		return
	}
	regisry = reg
}

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
	if regisry != nil {
		regisry.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func (c Registry) Post(ctx *gin.Context) {

	values := ctx.Request.URL.Query()
	values.Set("ns", "ccr.ccs.tencentyun.com")
	ctx.Request.URL.RawQuery = values.Encode()
	if regisry != nil {
		regisry.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

// func (c Registry) Finish(ctx *gin.Context) {
// 	name := ctx.Param("name")
// 	ref := ctx.Param("reference")
// 	values := ctx.Request.URL.Query()
// 	values.Set("ns", "ccr.ccs.tencentyun.com")
// 	ctx.Request.URL.RawQuery = values.Encode()
// 	if regisry != nil {
// 		regisry.ServeHTTP(ctx.Writer, ctx.Request)
// 	}
// }

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
	// nerdctl commit image
}
