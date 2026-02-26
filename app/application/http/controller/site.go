package controller

import (
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Site struct {
	controller.Abstract
}

func (self Site) Beian(http *gin.Context) {
	sdk := k8s.NewK8sClient()
	configmap, err := sdk.ClientSet.CoreV1().ConfigMaps("default").Get(http, "beian", metav1.GetOptions{})
	if err != nil {
		self.JsonSuccessResponse(http)
		return
	}

	response := gin.H{}
	if data, ok := configmap.Data["icpnumber"]; ok {
		response["icpnumber"] = data
	}
	if data, ok := configmap.Data["number"]; ok {
		response["number"] = data
	}
	if data, ok := configmap.Data["location"]; ok {
		response["location"] = data
	}

	self.JsonResponseWithoutError(http, response)
}

func (self Site) K3kConfig(http *gin.Context) {
	sdk := k8s.NewK8sClient()
	configmap, err := sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Get(http, "k3k.config", metav1.GetOptions{})
	if err != nil {
		self.JsonSuccessResponse(http)
		return
	}

	response := gin.H{}
	if data, ok := configmap.Data["indexpage"]; ok {
		response["indexpage"] = data
	}

	self.JsonResponseWithoutError(http, response)
}

func (self Site) InitUser(http *gin.Context) {
	releaseName := facade.Config.GetString("app.helm_release_name")
	sdk := k8s.NewK8sClient()

	response := gin.H{
		"canInitUser":          "false",
		"allowConsoleRegister": "false",
		"captchaEnabled":       "false",
	}

	_, err := sdk.ClientSet.CoreV1().ConfigMaps(sdk.GetNamespace()).Get(http, releaseName+"-init-user", metav1.GetOptions{})
	if err == nil {
		response["canInitUser"] = "true"
	}

	if facade.Config.GetBool("captcha.enabled") {
		response["captchaEnabled"] = "true"
	}

	k3kconfig, err := sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Get(http, "k3k.config", metav1.GetOptions{})
	if err == nil {
		if data, ok := k3kconfig.Data["allowConsoleRegister"]; ok {
			response["allowConsoleRegister"] = data
		}
	}

	self.JsonResponseWithoutError(http, response)
}
