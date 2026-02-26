package gpu

import (
	"errors"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/metrics"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// sum(GPUDeviceSharedNum) VGPU核数  vgpu分配率
// sum(GPUDeviceCoreAllocated) 算力分配 /GPUDeviceCoreLimit 算力总数
// sum(nodeGPUMemoryPercentage)  显存分配率 （GPUDeviceMemoryAllocated) / sum(GPUDeviceMemoryLimit) 显存总数

type ClusterSummary struct {
	sdk *k8s.Sdk
}

type Percentage struct {
	Total      float64 `json:"total"`
	Current    float64 `json:"current"`
	Percentage float64 `json:"percentage"`
	Type       string  `json:"type"`
}
type HamiNode struct {
	Name      string `json:"name"`
	Devmem    int32  `json:"devmem"`
	Devcore   int32  `json:"devcore"`
	VgpuTotal int32  `json:"vgpuTotal"`
}

type Summary struct {
	GPUDeviceSharedNum            int32 `json:"GPUDeviceSharedNum"`            //3
	GPUDeviceSharedTotal          int32 `json:"GPUDeviceSharedTotal"`          //10 vgpu分配率
	GPUDeviceCoreAllocated        int32 `json:"GPUDeviceCoreAllocated"`        //65
	GPUDeviceCoreLimit            int32 `json:"GPUDeviceCoreLimit"`            //100 算力分配率
	GPUDeviceMemoryAllocated      int64 `json:"GPUDeviceMemoryAllocated"`      // 显存分配数
	GPUDeviceMemoryAllocatedTotal int32 `json:"GPUDeviceMemoryAllocatedTotal"` // 显存总数
}

func NewSummary() *Summary {
	return &Summary{
		GPUDeviceSharedNum:            0,
		GPUDeviceSharedTotal:          0,
		GPUDeviceCoreAllocated:        0, // 算力分配数
		GPUDeviceCoreLimit:            0, // 算力总数
		GPUDeviceMemoryAllocated:      0,
		GPUDeviceMemoryAllocatedTotal: 0,
	}
}

func (c *ClusterSummary) sumCoreMemory(nodeList *v1.NodeList) (int32, int32, error) {
	nodes, err := c.TotalGPUCoreMem(nodeList)
	if err != nil {
		return 0, 0, err
	}
	var (
		vgpuTotal int32 = 0
		totalMem  int32 = 0
	)
	for _, node := range nodes {
		vgpuTotal += node.VgpuTotal
		totalMem += node.Devmem
	}
	return vgpuTotal, totalMem, nil
}
func (c *ClusterSummary) Summary() (*Summary, error) {

	nodes, err := c.sdk.ClientSet.CoreV1().Nodes().List(c.sdk.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	result := &Summary{
		GPUDeviceSharedNum:            0,
		GPUDeviceSharedTotal:          0,
		GPUDeviceCoreAllocated:        0, // 算力分配数
		GPUDeviceCoreLimit:            0, // 算力总数
		GPUDeviceMemoryAllocated:      0,
		GPUDeviceMemoryAllocatedTotal: 0,
	}
	totalCore, totalMem, err := c.sumCoreMemory(nodes)
	if err != nil {
		return nil, err
	}
	result.GPUDeviceMemoryAllocatedTotal = totalMem
	// result.GPUDeviceCoreLimit = totalCore
	result.GPUDeviceSharedTotal = totalCore

	vstat, err := c.vstat(nodes)
	if err != nil {
		slog.Warn("c.vstat", "err", err)
		return result, nil
	}
	// sum(GPUDeviceSharedNum) VGPU核数  vgpu分配率
	// sum(GPUDeviceCoreAllocated) 算力分配 /GPUDeviceCoreLimit 算力总数
	// sum(nodeGPUMemoryPercentage)  显存分配率 （GPUDeviceMemoryAllocated) / sum(GPUDeviceMemoryLimit) 显存总数
	result.GPUDeviceSharedNum = lo.SumBy(vstat.GPUDeviceSharedNum, func(item *metrics.GPUDeviceSharedNum) int32 {
		return int32(item.Value)
	})
	result.GPUDeviceMemoryAllocated = lo.SumBy(vstat.GPUDeviceMemoryAllocated, func(item *metrics.GPUDeviceMemoryAllocated) int64 {
		return int64(item.Value) / 1024 / 1024
	})

	result.GPUDeviceCoreAllocated = lo.SumBy(vstat.GPUDeviceCoreAllocated, func(item *metrics.GPUDeviceCoreAllocated) int32 {
		return int32(item.Value)
	})
	result.GPUDeviceCoreLimit = lo.SumBy(vstat.GPUDeviceCoreLimit, func(item *metrics.GPUDeviceCoreLimit) int32 {
		return int32(item.Value)
	})
	return result, nil
}

func NewClusterSummary(sdk *k8s.Sdk) *ClusterSummary {
	return &ClusterSummary{
		sdk: sdk,
	}
}

func (c *ClusterSummary) vstat(nodes *v1.NodeList) (*metrics.VgpuMetrics, error) {
	if len(nodes.Items) == 0 {
		return nil, errors.New("nodes is empty")
	}
	first := nodes.Items[0] //31993 31992 端口不区分node 根据deveice 区分
	nodeScape, err := metrics.NewNodeScrapeUseNode(&first, HAMI_STAT_PORT)
	if err != nil {
		slog.Warn("metrics.NewNodeScrapeUseNode", "err", err)
		return nil, errors.New("new nodeScape error")
	}
	vstat, err := nodeScape.Scrape()
	if err != nil {
		slog.Warn("nodeScape.Scrape", "err", err)
		return nil, errors.New("new nodeScape error")
	}
	return vstat, nil
}

func (c *ClusterSummary) TotalGPUCoreMem(nodeList *v1.NodeList) ([]*HamiNode, error) {
	var result []*HamiNode
	for _, node := range nodeList.Items {
		hnode := &HamiNode{
			Name:      node.Name,
			Devmem:    0,
			Devcore:   0,
			VgpuTotal: 0,
		}
		devices, err := FetchDevices(&node)
		if err != nil {
			continue
		}
		for _, device := range devices {
			// device.Devcore = device.Devcore
			hnode.Devmem += device.Devmem
			hnode.Devcore += device.Devcore
			hnode.VgpuTotal += device.Count
		}
		result = append(result, hnode)
	}
	return result, nil
}

func (c *ClusterSummary) GetNodesDevicesByNodes(nodeList *v1.NodeList) ([]*DeviceInfo, error) {
	devicesResult := []*DeviceInfo{}
	for _, node := range nodeList.Items {

		devices, err := FetchDevices(&node)
		if err != nil {
			continue
		}
		for _, device := range devices {
			device.NodeName = node.Name
			devicesResult = append(devicesResult, device)
		}

	}
	return devicesResult, nil
}

func (c *ClusterSummary) GetNodesDevices() ([]*DeviceInfo, error) {
	nodes, err := c.sdk.ClientSet.CoreV1().Nodes().List(c.sdk.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return c.GetNodesDevicesByNodes(nodes)
}
