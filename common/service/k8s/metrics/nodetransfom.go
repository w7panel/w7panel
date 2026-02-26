package metrics

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type NodeMetricsTransform struct {
	vgpuMetrics     *VgpuMetrics
	nodeMetrics     *v1beta1.NodeMetrics
	podMetricsItems []v1beta1.PodMetrics
}

func NewNodeMetricsTransform(vgpuMetrics *VgpuMetrics, nodeMetrics *v1beta1.NodeMetrics, podMetricsItems []v1beta1.PodMetrics) *NodeMetricsTransform {
	return &NodeMetricsTransform{
		vgpuMetrics:     vgpuMetrics,
		nodeMetrics:     nodeMetrics,
		podMetricsItems: podMetricsItems,
	}
}

func (n *NodeMetricsTransform) Transfom() {
	vgpuStat := n.vgpuMetrics
	if len(vgpuStat.HostCoreUtilization) > 0 {
		v := vgpuStat.HostCoreUtilization[0]
		value, err := resource.ParseQuantity(fmt.Sprintf("%f", v.Value))
		if err == nil {
			n.nodeMetrics.Usage[HostCoreUtilizationName] = value
		}

	}
	if len(vgpuStat.HostGPUMemoryUsage) > 0 {
		v := vgpuStat.HostGPUMemoryUsage[0]
		value, err := resource.ParseQuantity(fmt.Sprintf("%f", v.Value))
		if err == nil {
			n.nodeMetrics.Usage[HostGPUMemoryUsageName] = value
		}

	}
	for k, podMetrics := range n.podMetricsItems {
		n.podMetricsItems[k] = n.transformPod(&podMetrics)
	}
}

func (n *NodeMetricsTransform) transformPod(podMetrics *v1beta1.PodMetrics) v1beta1.PodMetrics {
	podName := podMetrics.Name

	for _, d1 := range n.vgpuMetrics.DeviceMemoryDescOfContainer {
		if d1.PodName == podName {
			for k, containerMetrics := range podMetrics.Containers {
				containerName := containerMetrics.Name
				if containerName == d1.ContainerName {
					val, err := resource.ParseQuantity(fmt.Sprintf("%f", d1.Value))
					if err != nil {
						continue
					}
					podMetrics.Containers[k].Usage[Device_memory_desc_of_containerName] = val
				}
			}
		}
	}
	for _, d2 := range n.vgpuMetrics.DeviceMemoryUsage {
		if d2.PodName == podName {
			for k, containerMetrics := range podMetrics.Containers {
				containerName := containerMetrics.Name
				if containerName == d2.ContainerName {
					val, err := resource.ParseQuantity(fmt.Sprintf("%f", d2.Value))
					if err != nil {
						continue
					}
					podMetrics.Containers[k].Usage[vGPU_device_memory_usage_in_bytesName] = val
				}
			}
		}
	}

	for _, d3 := range n.vgpuMetrics.DeviceUtilization {
		if d3.PodName == podName {
			for k, containerMetrics := range podMetrics.Containers {
				containerName := containerMetrics.Name
				if containerName == d3.ContainerName {
					val, err := resource.ParseQuantity(fmt.Sprintf("%f", d3.Value))
					if err != nil {
						continue
					}
					podMetrics.Containers[k].Usage[Device_utilization_desc_of_containerName] = val
				}
			}
		}
	}

	for _, d4 := range n.vgpuMetrics.DeviceMemoryLimit {
		if d4.PodName == podName {
			for k, containerMetrics := range podMetrics.Containers {
				containerName := containerMetrics.Name
				if containerName == d4.ContainerName {
					val, err := resource.ParseQuantity(fmt.Sprintf("%f", d4.Value))
					if err != nil {
						continue
					}
					podMetrics.Containers[k].Usage[vGPU_device_memory_limit_in_bytesName] = val
				}
			}
		}
	}
	return *podMetrics
}
