package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/w7panel/w7panel/common/service/registry"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Registry struct {
	controller.Abstract
}

var mmr = registry.CreateMicroRegistry()

var cdr = registry.ContainerDRegistryHandler

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(statusCode int) {
	sw.statusCode = statusCode
	sw.ResponseWriter.WriteHeader(statusCode)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if sw.statusCode == 0 {
		sw.statusCode = http.StatusOK
	}
	return sw.ResponseWriter.Write(b)
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
	w := &statusWriter{ResponseWriter: ctx.Writer}
	mmr.ServeHTTP(w, ctx.Request) //先从内存取镜像
	if w.statusCode != http.StatusOK {
		ctx.Writer.WriteHeader(w.statusCode)
	}
}

// Tags 返回镜像标签
func (c Registry) Post(ctx *gin.Context) {
	mmr.ServeHTTP(ctx.Writer, ctx.Request)
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

	mmr.ServeHTTP(ctx.Writer, ctx.Request)
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
	mmr.ServeHTTP(ctx.Writer, ctx.Request)
	// nerdctl commit image
}
