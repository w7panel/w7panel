package controller

import (
	"context"
	"errors"
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
		// slog.Error("init registry err", "err", err)
		return
	}
	regisry = reg
}

// Version 返回 Registry API 版本
func (self Registry) Version(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"schemas": []string{"https://docs.docker.com/spec/api/v2/"},
	})
}

// Catalog 返回镜像列表
func (self Registry) Handler(ctx *gin.Context) {
	if regisry != nil {
		regisry.ServeHTTP(ctx.Writer, ctx.Request)
		return
	}
	err := errors.New("not support")
	self.JsonResponseWithServerError(ctx, err)
}
