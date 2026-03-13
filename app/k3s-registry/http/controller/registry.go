package controller

import (
	"net/http"

	"gitee.com/we7coreteam/k8s-offline/app/k3s-registry/logic"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Registry struct {
	controller.Abstract
}

var registryLogic = logic.NewRegistryLogic()

// Version 返回 Registry API 版本
func (c Registry) Version(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"schemas": []string{"https://docs.docker.com/spec/api/v2/"},
	})
}

// Catalog 返回镜像列表
func (c Registry) Catalog(ctx *gin.Context) {
	catalog, err := registryLogic.GetCatalog(ctx)
	if err != nil {
		c.JsonResponseWithServerError(ctx, err)
		return
	}
	c.JsonResponseWithoutError(ctx, gin.H{"repositories": catalog})
}

// Tags 返回镜像标签
func (c Registry) Tags(ctx *gin.Context) {
	name := ctx.Param("name")
	tags, err := registryLogic.GetTags(ctx, name)
	if err != nil {
		c.JsonResponseWithServerError(ctx, err)
		return
	}
	c.JsonResponseWithoutError(ctx, gin.H{"name": name, "tags": tags})
}

// Manifest 获取镜像 manifest
func (c Registry) Manifest(ctx *gin.Context) {
	name := ctx.Param("name")
	reference := ctx.Param("reference")

	manifest, err := registryLogic.GetManifest(ctx, name, reference)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	ctx.Data(http.StatusOK, "application/vnd.docker.distribution.manifest.v2+json", []byte(manifest))
}

// PushManifest 推送镜像 manifest
func (c Registry) PushManifest(ctx *gin.Context) {
	name := ctx.Param("name")
	reference := ctx.Param("reference")

	var manifest string
	if err := ctx.ShouldBindJSON(&manifest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid manifest"})
		return
	}

	if err := registryLogic.PushManifest(ctx, name, reference, manifest); err != nil {
		c.JsonResponseWithServerError(ctx, err)
		return
	}

	ctx.Header("Location", "/v2/"+name+"/manifests/"+reference)
	ctx.JSON(http.StatusCreated, gin.H{})
}

// Blob 获取 blob
func (c Registry) Blob(ctx *gin.Context) {
	name := ctx.Param("name")
	digest := ctx.Param("digest")

	data, err := registryLogic.GetBlob(ctx, name, digest)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	ctx.Data(http.StatusOK, "application/octet-stream", data)
}

// BlobExists 检查 blob 是否存在
func (c Registry) BlobExists(ctx *gin.Context) {
	name := ctx.Param("name")
	digest := ctx.Param("digest")

	if registryLogic.BlobExists(ctx, name, digest) {
		ctx.JSON(http.StatusOK, gin.H{})
	} else {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	}
}

// InitUpload 初始化 blob 上传
func (c Registry) InitUpload(ctx *gin.Context) {
	name := ctx.Param("name")

	uuid, err := registryLogic.InitUpload(ctx, name)
	if err != nil {
		c.JsonResponseWithServerError(ctx, err)
		return
	}

	ctx.Header("Location", "/v2/"+name+"/blobs/uploads/"+uuid)
	ctx.Header("Range", "bytes=0-0")
	ctx.JSON(http.StatusAccepted, gin.H{})
}

// CompleteUpload 完成 blob 上传
func (c Registry) CompleteUpload(ctx *gin.Context) {
	name := ctx.Param("name")
	uuid := ctx.Param("uuid")
	digest := ctx.Query("digest")

	body, _ := ctx.GetRawData()

	if err := registryLogic.CompleteUpload(ctx, name, uuid, digest, body); err != nil {
		c.JsonResponseWithServerError(ctx, err)
		return
	}

	ctx.Header("Location", "/v2/"+name+"/blobs/"+digest)
	ctx.JSON(http.StatusCreated, gin.H{})
}
