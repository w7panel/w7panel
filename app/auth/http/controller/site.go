package controller

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/common/helper"
	"github.com/w7panel/w7panel/common/service/console"
	"github.com/w7panel/w7panel/common/service/k8s"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Site struct {
	controller.Abstract
}

func (self Site) RegisterZpkSite(http *gin.Context) {
	type ParamsValidate struct {
		Host          string `form:"host" binding:"required"`
		SiteIdentifie string `form:"siteIdentifie" binding:"required"`
		InstallId     string `form:"installId" binding:"required"`
		AppName       string `form:"appName" binding:"required"`
		ContainerName string `form:"containerName" binding:"required"`
		Namespace     string `form:"namespace" binding:"required"`
		ReleaseName   string `form:"releaseName" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	// if helper.IsLocalMock() {
	helper.Set(params.ReleaseName, params.InstallId, time.Minute)
	// }
	ok := helper.Check(params.ReleaseName, params.InstallId)
	if !ok {
		self.JsonResponseWithServerError(http, errors.New("权限错误"))
		return
	}
	secret, err := console.RegisterSiteZpk(params.Host, params.SiteIdentifie)
	if err != nil {
		slog.Error("注册站点失败", "err", err)
		self.JsonResponseWithServerError(http, err)
		return
	}

	slog.Info("注册站点成功", "secret", secret)
	sdk := k8s.NewK8sClientInner()
	err = console.PatchAppId(sdk, secret, params.AppName, params.Namespace, params.ContainerName)
	if err != nil {
		slog.Error("更新appid失败", "err", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
}
