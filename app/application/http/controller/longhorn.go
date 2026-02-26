package controller

import (
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/longhorn"
	"github.com/gin-gonic/gin"

	// "github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"

	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Longhorn struct {
	controller.Abstract
}

func (self Longhorn) GetNeedDeleteReplicas(http *gin.Context) {
	type ParamsValidate struct {
		DiskSelector string `form:"diskselector" binding:"required"`
		NodeId       string `form:"nodeid" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	nodeIds := strings.Split(params.NodeId, ",")

	sdk, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	longhornclient, err := longhorn.NewLonghornClient(sdk)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	compose, err := longhornclient.GetVolumeReplicaCompose()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	replicas := compose.GetVolumeReplicas().GetNeedDeleteReplicas([]string{params.DiskSelector}, nodeIds)
	self.JsonResponseWithoutError(http, replicas)
}

func (self Longhorn) GetVolumesStatus(http *gin.Context) {
	type ParamsValidate struct {
		ConvertPvc string `form:"convertpvc"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	type VolumesStatus struct {
		NumberOfReplicas  int    `json:"numberOfReplicas"`
		Robustness        string `json:"robustness"`
		Size              int64  `json:"size,string"`
		ActualSize        int64  `json:"actualSize"`
		CreationTimestamp string `json:"creationTimestamp"`
		AccessMode        string `json:"accessMode"`
	}

	sdk := k8s.NewK8sClient().Sdk
	longhornclient, err := longhorn.NewLonghornClient(sdk)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	volumes, err := longhornclient.GetVolumeList()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	result := map[string]VolumesStatus{}
	for _, volume := range volumes.Items {
		if volume.Status.KubernetesStatus.PVCName == "" {
			continue
		}
		size := volume.Status.ActualSize
		vs := VolumesStatus{
			NumberOfReplicas:  volume.Spec.NumberOfReplicas,
			Robustness:        string(volume.Status.Robustness),
			Size:              volume.Spec.Size,
			ActualSize:        size,
			AccessMode:        string(volume.Spec.AccessMode),
			CreationTimestamp: volume.CreationTimestamp.Format("2006-01-02 15:04:05"),
			// CreatedAt:        volume.Status.KubernetesStatus.PVCName,
			// CreatedAt:        volume.Status.CreatedAt,
		}
		result[volume.Status.KubernetesStatus.PVCName+":"+volume.Status.KubernetesStatus.Namespace] = vs
	}

	self.JsonResponseWithoutError(http, result)
}

/*
*

	扩容卷
*/
func (self Longhorn) Expand(http *gin.Context) {

	token := http.MustGet("k8s_token").(string)
	user, err := k3k.TokenToK3kUser(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	if user.IsVirtual() {

	}
	// type ParamsValidate struct {
	// 	Namespace string `form:"namespace" binding:"required"`
	// 	Name	  string `form:"name" binding:"required"`
	// 	Size	  string `form:"size" binding:"required"`
	// }
	// params := ParamsValidate{}
	// if !self.Validate(http, &params) {
	// 	return
	// }
	// sdk, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	// if err != nil {
	// 	self.JsonResponseWithServerError(http, err)
	// 	return
	// }
	// longhornclient, err := longhorn.NewLonghornClient(sdk)
	// if err != nil {
	// 	self.JsonResponseWithServerError(http, err)
	// 	return
	// }
	// err = longhornclient.Expand(params.Namespace, params.Name, params.Size)
	// if err != nil {
	// 	self.JsonResponseWithServerError(http, err)
	// 	return
	// }
	// self.JsonResponseWithoutError(http, nil)
}
