package controller

import (
	"context"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubeBlocks struct {
	controller.Abstract
}

func (self KubeBlocks) InstallJobYaml(http *gin.Context) {

	// run cmd
	//kubectl apply -f ./crds-kubeblocks/1.0.1/ --server-side
	//helm install kubeblocks ./kubeblocks-1.0.1.tgz -n kb-system
	token := http.MustGet("k8s_token").(string)
	// client, err := k8s.NewK8sClient().Channel(token)
	// if err != nil {
	// 	self.JsonResponseWithServerError(http, err)
	// 	return
	// }
	user, err := k3k.TokenToK3kUser(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	saName := helper.ServiceAccountName()
	if user.IsClusterUser() {
		saName = user.Name
	}
	job := zpk.ToKubeblockInstallJob(saName)
	self.JsonResponseWithoutError(http, job)
	return
}

func (self KubeBlocks) Install(http *gin.Context) {

	// run cmd
	//kubectl apply -f ./crds-kubeblocks/1.0.1/ --server-side
	//helm install kubeblocks ./kubeblocks-1.0.1.tgz -n kb-system
	token := http.MustGet("k8s_token").(string)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	user, err := k3k.TokenToK3kUser(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	saName := helper.ServiceAccountName()
	if user.IsClusterUser() {
		saName = user.Name
	}
	job := zpk.ToKubeblockInstallJob(saName)
	job, err = client.ClientSet.BatchV1().Jobs("default").Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, job)
	return
}
