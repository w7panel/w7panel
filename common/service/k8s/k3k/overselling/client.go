package overselling

import (
	"fmt"
	"strconv"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/longhorn"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceClient struct {
	sdk    *k8s.Sdk
	client sigclient.Client
}

func NewResourceClient(sdk *k8s.Sdk) (*ResourceClient, error) {
	sigClient, err := sdk.ToSigClient()
	if err != nil {
		return nil, err
	}
	return &ResourceClient{
		sdk:    sdk,
		client: sigClient,
	}, nil
}

func (c *ResourceClient) GetUsedOld() (*Resource, error) {
	list := &corev1.ResourceQuotaList{}
	err := c.client.List(c.sdk.Ctx, list, &sigclient.ListOptions{})
	if err != nil {
		return nil, err
	}
	storage, err := c.storageScheduled()
	if err != nil {
		return nil, err
	}
	cpu := resource.MustParse("0")
	memory := resource.MustParse("0")
	cpup := &cpu
	memoryp := &memory
	for _, item := range list.Items {
		spec := item.Spec
		cpup.Add(spec.Hard["limits.cpu"])
		memoryp.Add(spec.Hard["limits.memory"])
	}
	return &Resource{
		CPU:     cpu,
		Memory:  memory,
		Storage: storage,
	}, nil
}

func (c *ResourceClient) GetUsed(callback func(*corev1.ServiceAccount) *Resource) (*Resource, error) {
	list := &corev1.ServiceAccountList{}
	err := c.client.List(c.sdk.Ctx, list, &sigclient.ListOptions{})
	if err != nil {
		return nil, err
	}
	cpu := resource.MustParse("0")
	memory := resource.MustParse("0")
	cpup := &cpu
	memoryp := &memory
	for _, item := range list.Items {
		if item.Labels == nil {
			continue
		}
		if item.Labels["w7.cc/user-mode"] != "cluster" {
			continue
		}
		if item.Labels["k3k.io/cluster-status"] != "ready" {
			continue
		}
		userRs := callback(&item)
		cpup.Add(userRs.CPU)
		memoryp.Add(userRs.Memory)
	}
	storage, err := c.storageScheduled()
	if err != nil {
		return nil, err
	}

	return &Resource{
		CPU:     cpu,
		Memory:  memory,
		Storage: storage,
	}, nil
}

func (c *ResourceClient) GetOverlingResource() (*Resource, error) {
	current, err := c.getResource()
	if err != nil {
		return nil, err
	}
	sellingConfig, err := c.GetSellingConfig()
	if err != nil {
		sellingConfig = OverSellingConfig{
			CPU:          int32(100),
			Memory:       int32(100),
			Storage:      int32(100),
			BandWidth:    int32(1000),
			BandWidthNum: int32(100),
		}
	}
	current.BandWidth = resource.MustParse(fmt.Sprintf("%dMi", sellingConfig.BandWidthNum))
	rs := &OverSellingResource{Allocated: *current, OverSelling: sellingConfig}
	result := rs.OverSellingResource()
	return &result, nil
}

func (c *ResourceClient) getResource() (*Resource, error) {
	allocatedCPU, allocatedMemory, err := c.nodeAllocate()
	if err != nil {
		return nil, err
	}
	allocatedStorage, err := c.storageAllocate()
	if err != nil {
		return nil, err
	}
	return &Resource{
		CPU:     allocatedCPU,
		Memory:  allocatedMemory,
		Storage: allocatedStorage,
	}, nil
}

func (c *ResourceClient) GetSellingConfig() (OverSellingConfig, error) {
	configmap, err := c.sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Get(c.sdk.Ctx, "k3k.overselling.config", metav1.GetOptions{})
	if err != nil {
		return OverSellingConfig{}, err
	}

	cpu, err := strconv.ParseInt(configmap.Data["cpu"], 10, 32)
	if err != nil {
		return OverSellingConfig{}, err
	}
	memory, err := strconv.ParseInt(configmap.Data["memory"], 10, 32)
	if err != nil {
		return OverSellingConfig{}, err
	}
	storage, err := strconv.ParseInt(configmap.Data["storage"], 10, 32)
	if err != nil {
		return OverSellingConfig{}, err
	}
	bandwidth, err := strconv.ParseInt(configmap.Data["bandwidth"], 10, 32)
	if err != nil {
		return OverSellingConfig{}, err
	}
	bandwidthNum, err := strconv.ParseInt(configmap.Data["bandwidthNum"], 10, 32)
	if err != nil {
		return OverSellingConfig{}, err
	}

	return OverSellingConfig{
		CPU:          int32(cpu),
		Memory:       int32(memory),
		Storage:      int32(storage),
		BandWidth:    int32(bandwidth),
		BandWidthNum: int32(bandwidthNum),
	}, nil
}

func (k *ResourceClient) nodeAllocate() (allocatedCPU, allocatedMemory resource.Quantity, err error) {
	nodes, err := k.sdk.ClientSet.CoreV1().Nodes().List(k.sdk.Ctx, metav1.ListOptions{})
	if err != nil {
		return resource.Quantity{}, resource.Quantity{}, err
	}

	for _, node := range nodes.Items {
		allocatedCPU.Add(*node.Status.Allocatable.Cpu())
		allocatedMemory.Add(*node.Status.Allocatable.Memory())
	}
	return allocatedCPU, allocatedMemory, nil
}

func (k *ResourceClient) storageAllocate() (allocatedStorage resource.Quantity, err error) {
	nodes, err := longhorn.GetLonghornNodeList()
	if err != nil {
		return resource.Quantity{}, err
	}

	for _, node := range nodes.Items {
		for _, disk := range node.Status.DiskStatus {
			rs := resource.MustParse("0")
			rs.Set(disk.StorageMaximum)
			allocatedStorage.Add(rs)
		}
	}
	return allocatedStorage, nil
}

func (k *ResourceClient) storageScheduled() (allocatedStorage resource.Quantity, err error) {
	nodes, err := longhorn.GetLonghornNodeList()
	if err != nil {
		return resource.Quantity{}, err
	}

	for _, node := range nodes.Items {
		for _, disk := range node.Status.DiskStatus {
			rs := resource.MustParse("0")
			rs.Set(disk.StorageScheduled)
			allocatedStorage.Add(rs)
		}
	}
	return allocatedStorage, nil
}
