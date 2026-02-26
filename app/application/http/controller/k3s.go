package controller

import (
	// "github.com/go-openapi/spec"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K3s struct {
	controller.Abstract
}

func (self K3s) GogcInfo(http *gin.Context) {

}

func (self K3s) GoGcToggle(http *gin.Context) {
	type ParamsValidate struct {
		GogcEnabled string `form:"gogcEnabled" binding:"required,omitempty"`
		Gogcval     string `form:"gogcVal" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	token := http.MustGet("k8s_token").(string)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	sec, err := client.ClientSet.CoreV1().Secrets("kube-system").List(client.Ctx, metav1.ListOptions{LabelSelector: "k3s-config-type=env"})
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	if params.GogcEnabled == "true" {
		for _, v := range sec.Items {
			// _, _ := v.Data["GOGC"]
			v.Data["GOGC"] = []byte(params.Gogcval)
			v.Labels["last-modify"] = "k8s-offline"
			_, err := client.ClientSet.CoreV1().Secrets("kube-system").Update(client.Ctx, &v, metav1.UpdateOptions{})
			if err != nil {
				self.JsonResponseWithServerError(http, err)
				return
			}
		}
		return
	} else {
		for _, v := range sec.Items {
			_, ok := v.Data["GOGC"]
			if ok {
				delete(v.Data, "GOGC")
				v.Labels["last-modify"] = "k8s-offline"
				_, err := client.ClientSet.CoreV1().Secrets("kube-system").Update(client.Ctx, &v, metav1.UpdateOptions{})
				if err != nil {
					self.JsonResponseWithServerError(http, err)
					return
				}
			}
		}
		return
	}
	self.JsonSuccessResponse(http)

}

func (self K3s) GoGc(http *gin.Context) {

	token := http.MustGet("k8s_token").(string)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	sec, err := client.ClientSet.CoreV1().Secrets("kube-system").List(client.Ctx, metav1.ListOptions{LabelSelector: "k3s-config-type=env"})
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	gogcEnabled := false
	for _, v := range sec.Items {
		_, ok := v.Data["GOGC"]
		if ok {
			gogcEnabled = true
			break
		}
	}
	self.JsonResponseWithoutError(http, gin.H{"gogcEnabled": gogcEnabled})
}
