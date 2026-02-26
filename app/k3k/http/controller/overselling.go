package controller

import (
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/overselling"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type OverSelling struct {
	controller.Abstract
}

func (self OverSelling) OverSellingConfig(http *gin.Context) {
	sdk := k8s.NewK8sClient().Sdk
	client, err := overselling.NewResourceClient(sdk)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	config := overselling.OverSellingConfig{
		CPU:          100,
		Memory:       100,
		Storage:      100,
		BandWidth:    1000,
		BandWidthNum: 100,
	}
	config, err = client.GetSellingConfig()
	if err != nil {
		self.JsonResponseWithoutError(http, config)
		return
	}
	self.JsonResponseWithoutError(http, config)
	return
}

func (self OverSelling) CurrentResource(http *gin.Context) {
	sdk := k8s.NewK8sClient().Sdk
	client, err := overselling.NewResourceClient(sdk)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	rs, err := client.GetOverlingResource()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	type Result struct {
		CPU       int64 `json:"cpu"`
		Memory    int64 `json:"memory"`
		Storage   int64 `json:"storage"`
		BandWidth int64 `json:"bandwidth"`
	}
	result := Result{
		CPU:       rs.CPU.Value(),
		Memory:    rs.Memory.Value() / 1024 / 1024 / 1024,
		Storage:   rs.Storage.Value() / 1024 / 1024 / 1024,
		BandWidth: rs.BandWidth.Value() / 1024 / 1024,
		// BandWidth: 100,
	}
	self.JsonResponseWithoutError(http, result)

}

func (self OverSelling) CheckResource(http *gin.Context) {

	type Result struct {
		Pass bool `json:"pass"`
	}
	token := http.MustGet("k8s_token").(string)
	k3kUser, err := k3k.TokenToK3kUser(token)
	if err != nil {
		slog.Error("token解析失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	if !k3kUser.IsClusterUser() {
		self.JsonResponseWithServerError(http, fmt.Errorf("非集群用户,无法使用此接口"))
		return
	}
	result := Result{
		Pass: false,
		// Pass: false,
	}
	sdk := k8s.NewK8sClient().Sdk
	err = k3k.TryCheckOverSellingResource(sdk, k3kUser)
	if err != nil {
		slog.Error("集群资源不足", "error", err)
		self.JsonResponseWithoutError(http, result)
		return
	}
	result.Pass = true // 集群资源充足
	self.JsonResponseWithoutError(http, result)
	return

}
