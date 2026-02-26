package metrics

import (
	"strconv"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/longhorn"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K3kUsage struct {
	sdk *k8s.Sdk
}

func NewK3kUsage(sdk *k8s.Sdk) *K3kUsage {
	return &K3kUsage{
		sdk: sdk,
	}
}

// GetResourceUsage returns the CPU and memory usage for a user, along with the total percentage of allocated resources.
func (k *K3kUsage) GetResourceUsage(k3kuser *types.K3kUser) (cpuUsage, memoryUsage resource.Quantity, allocatedCPU, allocatedMemory resource.Quantity, err error) {
	if k3kuser.IsClusterUser() {
		if k3kuser.IsVirtual() {
			client, err := k8s.NewK8sClient().GetK3kClusterSdkByConfig(k3kuser.ToK3kConfig())
			if err != nil {
				return resource.Quantity{}, resource.Quantity{}, resource.Quantity{}, resource.Quantity{}, nil
			}
			configmap, err := client.ClientSet.CoreV1().ConfigMaps("default").Get(client.Ctx, "metrics", metav1.GetOptions{})
			if err != nil {
				return resource.Quantity{}, resource.Quantity{}, resource.Quantity{}, resource.Quantity{}, nil
			}
			cpuValue, _ := strconv.ParseInt(configmap.Data["cpu"], 10, 64)
			memoryValue, _ := strconv.ParseInt(configmap.Data["memory"], 10, 64)
			cpuUsage = *resource.NewMilliQuantity(cpuValue, resource.DecimalSI)
			memoryUsage = *resource.NewQuantity(memoryValue, resource.BinarySI)
		} else {
			podMetrics := PodMetrics.GetNamespaceMetrics(k3kuser.GetK3kNamespace())
			for _, metric := range podMetrics {
				cpuUsage.Add(*resource.NewMilliQuantity(metric.CPUUsage, resource.DecimalSI))
				memoryUsage.Add(*resource.NewQuantity(metric.MemoryUsage, resource.BinarySI))
			}
		}
		// Get all pod metrics in the user's namespace and sum them up

	} else {
		// Get node metrics
		nodeMetrics := NodeMetrics.GetLatestMetrics()
		for _, metric := range nodeMetrics {
			cpuUsage.Add(*resource.NewMilliQuantity(metric.CPUUsage, resource.DecimalSI))
			memoryUsage.Add(*resource.NewQuantity(metric.MemoryUsage, resource.BinarySI))
		}
	}

	// Get allocated resources
	// var allocatedCPU, allocatedMemory resource.Quantity
	if limitRange := k3kuser.GetLimitRange(); limitRange != nil && k3kuser.IsClusterUser() {
		allocatedCPU = limitRange.Hard["cpu"]
		allocatedMemory = limitRange.Hard["memory"]
		if allocatedCPU.IsZero() || allocatedMemory.IsZero() {
			allocatedCPU, allocatedMemory, _ = k.nodeAllocate(allocatedCPU, allocatedMemory)
		}
	} else {
		allocatedCPU, allocatedMemory, _ = k.nodeAllocate(allocatedCPU, allocatedMemory)
	}

	return cpuUsage, memoryUsage, allocatedCPU, allocatedMemory, nil
}

func (k *K3kUsage) nodeAllocate(allocatedCPU resource.Quantity, allocatedMemory resource.Quantity) (resource.Quantity, resource.Quantity, error) {
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

func (k *K3kUsage) GetResourceDiskUsage(k3kuser *types.K3kUser) (storageUsage int64, storageTotal int64, err error) {
	// scName := k3kuser.GetStorageClass()
	total := resource.MustParse(k3kuser.GetStorageRequestSize())
	longhornClient, err := longhorn.NewLonghornClient(k.sdk)
	if err != nil {
		return 0, 0, err
	}
	volumes, err := longhornClient.GetVolumeList()
	if err != nil {
		return 0, 0, err
	}
	pvcsizeMap := make(map[string]int64)
	for _, volume := range volumes.Items {
		if volume.Status.KubernetesStatus.PVCName == "" {
			continue
		}
		pvcsizeMap[volume.Status.KubernetesStatus.PVCName+":"+volume.Status.KubernetesStatus.Namespace] = volume.Status.ActualSize
	}

	if k3kuser.IsClusterUser() {
		pvcs, err := k.sdk.ClientSet.CoreV1().PersistentVolumeClaims(k3kuser.GetK3kNamespace()).List(k.sdk.Ctx, metav1.ListOptions{})
		if err != nil {
			return 0, 0, err
		}
		usage := int64(0)
		for _, pvc := range pvcs.Items {
			size, ok := pvcsizeMap[pvc.GetName()+":"+pvc.GetNamespace()]
			if ok {
				usage += size
			}
		}
		return usage, total.Value(), nil
	} else {
		nodes, err := longhornClient.GetNodeList()
		if err != nil {
			return 0, 0, err
		}
		usage := int64(0)
		total := int64(0)
		for _, node := range nodes.Items {
			for _, storageNode := range node.Status.DiskStatus {
				total += storageNode.StorageMaximum
				usage += storageNode.StorageMaximum - storageNode.StorageAvailable
			}
		}
		return usage, total, nil
	}
}
