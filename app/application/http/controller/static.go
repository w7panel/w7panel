package controller

import (
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Static struct {
	controller.Abstract
}

func (self Static) StaticInfo(http *gin.Context) {
	identifie := http.Param("identifie")
	version := http.Query("version")
	releaseName := http.Query("releaseName")
	status := appgroup.DownStaticStatus(identifie, version, releaseName)
	self.JsonResponseWithoutError(http, gin.H{
		"status": status,
	})
}

func (self Static) Download(http *gin.Context) {
	name := http.Param("name")
	namespace := http.Param("namespace")
	token := http.MustGet("k8s_token").(string)
	sdk, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	appgroupObj, err := appgroup.GetAppgroupUseSdk(name, namespace, sdk)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	appgroup.DownStatic(appgroupObj)

}
