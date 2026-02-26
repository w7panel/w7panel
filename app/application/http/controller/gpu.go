package controller

import (
	// "github.com/go-openapi/spec"
	"errors"
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/gpu"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/gpu/gpustack"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/metrics"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Gpu struct {
	controller.Abstract
}

func (self Gpu) getManager(http *gin.Context) (*gpu.GpuManager, error) {
	type ParamsValidate struct {
		Namespace string `form:"namespace"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return nil, errors.New("参数错误")
	}
	token := http.MustGet("k8s_token").(string)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return nil, err
	}
	gpuManager, err := gpu.NewGpuManager(client, "", "")
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return nil, err
	}
	return gpuManager, nil
}

func (self Gpu) GetGpuConfig(http *gin.Context) {
	gpuManager, err := self.getManager(http)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	res := gpuManager.ToJsonStruct()
	self.JsonResponseWithoutError(http, res)
}

func (self Gpu) InstallGpuOperator(http *gin.Context) {
	type ParamsValidate struct {
		DriverVersion string `form:"driverVersion"`
		DriverEnabled bool   `form:"driverEnabled"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	gpuManager, err := self.getManager(http)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	err = gpuManager.InstallGpuOperator(params.DriverEnabled, params.DriverVersion)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonSuccessResponse(http)
}

func (self Gpu) InstallHami(http *gin.Context) {
	type ParamsValidate struct {
		RuntimeClassName string `form:"runtimeClassName"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	gpuManager, err := self.getManager(http)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	err = gpuManager.InstallHami(params.RuntimeClassName)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonSuccessResponse(http)
}

func (self Gpu) EnableGpu(http *gin.Context) {
	type ParamsValidate struct {
		Enabled bool `form:"enabled"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	gpuManager, err := self.getManager(http)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	err = gpuManager.GpuEnabled(params.Enabled)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonSuccessResponse(http)
}

func (self Gpu) HamiMetricsReal(http *gin.Context) {
	token := http.MustGet("k8s_token").(string)
	client, err := k8s.NewK8sClient().Channel(token)
	type NodeGpu struct {
		NodeName            string `form:"NodeName"`
		HostCoreUtilization string `form:"HostCoreUtilization"`
	}
	result := []NodeGpu{}
	if err != nil {
		self.JsonResponseWithoutError(http, result)
		return
	}
	nodes, err := client.ClientSet.CoreV1().Nodes().List(client.Ctx, metav1.ListOptions{})
	if err != nil {
		self.JsonResponseWithoutError(http, result)
		return
	}
	for _, node := range nodes.Items {
		nodeIp, err := metrics.GetNodeInnertIp(&node)
		if err != nil {
			slog.Error("获取节点内网IP失败", "err", err)
			continue
		}
		nodeScape := metrics.NewNodeScrape(nodeIp, metrics.HAMIPORT)
		vgpu, err := nodeScape.Scrape()
		if err != nil {
			slog.Error("获取节点vgpu信息失败", "err", err)
			continue
		}
		if (vgpu.HostCoreUtilization == nil) || len(vgpu.HostCoreUtilization) == 0 {
			continue
		}
		result = append(result, NodeGpu{
			NodeName:            node.GetName(),
			HostCoreUtilization: fmt.Sprintf("%f", vgpu.HostCoreUtilization[0].Value),
		})
	}
	self.JsonResponseWithoutError(http, result)
}

func (self Gpu) GpuSummary(http *gin.Context) {

	token := http.MustGet("k8s_token").(string)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	summary := gpu.NewClusterSummary(client)
	summaryData, err := summary.Summary()
	if err != nil {
		self.JsonResponseWithoutError(http, gpu.NewSummary())
		return
	}
	self.JsonResponseWithoutError(http, summaryData)
}

func (self Gpu) NodesDevices(http *gin.Context) {

	token := http.MustGet("k8s_token").(string)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	device := gpu.NewClusterSummary(client)
	deviceData, err := device.GetNodesDevices()
	if err != nil {
		self.JsonResponseWithoutError(http, []string{})
		return
	}
	self.JsonResponseWithoutError(http, deviceData)
}

func (self Gpu) CreateGpuStackWorker(http *gin.Context) {

	params := gpustack.WorkerConfig{}
	if !self.Validate(http, &params) {
		return
	}
	token := http.MustGet("k8s_token").(string)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	workerApi := gpustack.NewGpuStackWorker(client)
	_, err = workerApi.CreateWorker(&params)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, []string{})
}
