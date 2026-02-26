package metrics

import (
	dto "github.com/prometheus/client_model/go"
	v1 "k8s.io/api/core/v1"
)

const HAMIPORT = "31992"

const (
	// CPU, in cores. (500m = .5 cores)
	HostCoreUtilizationName                  v1.ResourceName = "HostCoreUtilization"
	HostGPUMemoryUsageName                   v1.ResourceName = "HostGPUMemoryUsage"
	Device_memory_desc_of_containerName      v1.ResourceName = "Device_memory_desc_of_container"
	Device_utilization_desc_of_containerName v1.ResourceName = "Device_utilization_desc_of_container"
	vGPU_device_memory_limit_in_bytesName    v1.ResourceName = "vGPU_device_memory_limit_in_bytes"
	vGPU_device_memory_usage_in_bytesName    v1.ResourceName = "vGPU_device_memory_usage_in_bytes"
	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	// ResourceMemory ResourceName = "memory"
	// // Volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)
	// ResourceStorage ResourceName = "storage"
	// // Local ephemeral storage, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	// ResourceEphemeralStorage ResourceName = "ephemeral-storage"
)

type Deveiceuuid struct {
	Deviceuuid string `json:"deviceuuid"`
}

type HostCoreUtilization struct {
	NodeId string `json:"nodeid"`
	Value  float64
	Deveiceuuid
}

type HostGPUMemoryUsage struct {
	NodeId string `json:"nodeid"`
	Value  float64
	Deveiceuuid
}

// Device_memory_desc_of_container	容器设备内存实时使用情况	{context="0",ctrname="2-1-3-pod-1",data="0",deviceuuid="GPU-00552014-5c87-89ac-b1a6-7b53aa24b0ec",module="0",offset="0",podname="2-1-3-pod-1",podnamespace="default",vdeviceid="0",zone="vGPU"} 0
type Device_memory_desc_of_container struct {
	GPUMetadata
	Value float64
}

// Device_utilization_desc_of_container	容器设备实时利用率	{ctrname="2-1-3-pod-1",deviceuuid="GPU-00552014-5c87-89ac-b1a6-7b53aa24b0ec",podname="2-1-3-pod-1",podnamespace="default",vdeviceid="0",zone="vGPU"} 0
type Device_utilization_desc_of_container struct {
	GPUMetadata
	Value float64
}

// 某个容器的设备限制
// {ctrname="2-1-3-pod-1",deviceuuid="GPU-00552014-5c87-89ac-b1a6-7b53aa24b0ec",podname="2-1-3-pod-1",podnamespace="default",vdeviceid="0",zone="vGPU"} 2.62144e+09
type vGPU_device_memory_limit_in_bytes struct {
	GPUMetadata
	Value float64
}

//vGPU_device_memory_usage_in_bytes
//某个容器的设备使用情况	{ctrname="2-1-3-pod-1",deviceuuid="GPU-00552014-5c87-89ac-b1a6-7b53aa24b0ec",podname="2-1-3-pod-1",podnamespace="default",vdeviceid="0",zone="vGPU"} 0

type vGPU_device_memory_usage_in_bytes struct {
	GPUMetadata
	Value float64
}

type GPUMetadata struct {
	PodName       string `json:"podname"`
	PodNamespace  string `json:"podnamespace"`
	ContainerName string `json:"ctrname"`
	DeviceUUID    string `json:"deviceuuid"`
	VDeviceID     string `json:"vdeviceid"`
	Zone          string `json:"zone"`
}

type VgpuMetrics struct {
	DeviceMemoryUsage           []*vGPU_device_memory_usage_in_bytes
	DeviceMemoryLimit           []*vGPU_device_memory_limit_in_bytes
	DeviceUtilization           []*Device_utilization_desc_of_container
	DeviceMemoryDescOfContainer []*Device_memory_desc_of_container
	HostCoreUtilization         []*HostCoreUtilization
	HostGPUMemoryUsage          []*HostGPUMemoryUsage
	GPUDeviceCoreLimit          []*GPUDeviceCoreLimit
	GPUDeviceMemoryLimit        []*GPUDeviceMemoryLimit
	GPUDeviceCoreAllocated      []*GPUDeviceCoreAllocated
	GPUDeviceMemoryAllocated    []*GPUDeviceMemoryAllocated
	GPUDeviceSharedNum          []*GPUDeviceSharedNum
	VGPUPodsDeviceAllocated     []*VGPUPodsDeviceAllocated
}

// 31993
// GPU 设备核心限制 {deviceidx="0",deviceuuid="GPU-00552014-5c87-89ac-b1a6-7b53aa24b0ec",nodeid="aio-node67",zone="vGPU"} 100
type GPUDeviceCoreLimit struct {
	Deveiceuuid
	NodeId string
	Value  float64
}

// GPU 设备内存限制 {deviceidx="0",deviceuuid="GPU-00552014-5c87-89ac-b1a6-7b53aa24b0ec",nodeid="aio-node67",zone="vGPU"} 3.4359738368e+10
type GPUDeviceMemoryLimit struct {
	Deveiceuuid
	NodeId string
	Value  float64
}

// /分配给某个 GPU 的设备核心 {deviceidx="0",deviceuuid="GPU-00552014-5c87-89ac-b1a6-7b53aa24b0ec",nodeid="aio-node67",zone="vGPU"} 45
type GPUDeviceCoreAllocated struct {
	Deveiceuuid
	NodeId string
	Value  float64
}

// 分配给某个 GPU 的设备内存{devicecores="0",deviceidx="0",deviceuuid="aio-node74-arm-Ascend310P-0",nodeid="aio-node74-arm",zone="vGPU"} 3.221225472e+09
type GPUDeviceMemoryAllocated struct {
	Deveiceuuid
	NodeId string
	Value  float64
}

// 共享此 GPU 的容器数量 {deviceidx="0",deviceuuid="GPU-00552014-5c87-89ac-b1a6-7b53aa24b0ec",nodeid="aio-node67",zone="vGPU"} 1
type GPUDeviceSharedNum struct {
	Deveiceuuid
	NodeId string
	Value  float64
}

// 从 pod 分配的 vGPU {containeridx="Ascend310P",deviceusedcore="0",deviceuuid="aio-node74-arm-Ascend310P-0",nodename="aio-node74-arm",podname="ascend310p-pod",podnamespace="default",zone="vGPU"} 3.221225472e+09
type VGPUPodsDeviceAllocated struct {
	Deveiceuuid
	NodeName       string `json:"nodename"`
	PodName        string `json:"podname"`
	ContainerIdx   string `json:"containeridx"`
	DeviceUsedCore string `json:"deviceusedcore"`
	PodNamespace   string `json:"podnamespace"`
	Value          float64
}

func labelsToMap(labels []*dto.LabelPair) map[string]string {
	result := make(map[string]string)
	for _, label := range labels {
		result[label.GetName()] = label.GetValue()
	}
	return result
}

// DeviceMetrics represents real-time device usage metrics

// MemoryStats contains memory utilization metrics
func ParseMapFamily(data map[string]*dto.MetricFamily) *VgpuMetrics {
	hosts := []*HostCoreUtilization{}
	hostusages := []*HostGPUMemoryUsage{}
	vgpumemorys := []*vGPU_device_memory_usage_in_bytes{}
	vgpulimits := []*vGPU_device_memory_limit_in_bytes{}
	deviceutils := []*Device_utilization_desc_of_container{}
	devicememorys := []*Device_memory_desc_of_container{}

	GPUDeviceCoreLimitArr := []*GPUDeviceCoreLimit{}
	GPUDeviceMemoryLimitArr := []*GPUDeviceMemoryLimit{}
	GPUDeviceCoreAllocatedArr := []*GPUDeviceCoreAllocated{}
	GPUDeviceMemoryAllocatedArr := []*GPUDeviceMemoryAllocated{}
	GPUDeviceSharedNumArr := []*GPUDeviceSharedNum{}
	VGPUPodsDeviceAllocatedArr := []*VGPUPodsDeviceAllocated{}

	for name, family := range data {
		metrics := family.GetMetric()

		switch name {
		case "Device_memory_desc_of_container":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				devicememory := Device_memory_desc_of_container{
					GPUMetadata: GPUMetadata{
						PodName:       labels["podname"],
						PodNamespace:  labels["podnamespace"],
						ContainerName: labels["ctrname"],
					},
				}
				devicememorys = append(devicememorys, &devicememory)
			}
		case "Device_utilization_desc_of_container":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				deviceutil := Device_utilization_desc_of_container{
					GPUMetadata: GPUMetadata{
						PodName:       labels["podname"],
						PodNamespace:  labels["podnamespace"],
						ContainerName: labels["ctrname"],
					},
					Value: metric.GetGauge().GetValue(),
				}
				deviceutils = append(deviceutils, &deviceutil)
			}

		case "vGPU_device_memory_limit_in_bytes":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				vgpulimit := vGPU_device_memory_limit_in_bytes{
					GPUMetadata: GPUMetadata{
						PodName:       labels["podname"],
						PodNamespace:  labels["podnamespace"],
						ContainerName: labels["ctrname"],
						DeviceUUID:    labels["deviceuuid"],
					},
					Value: metric.GetGauge().GetValue(),
				}
				vgpulimits = append(vgpulimits, &vgpulimit)
			}
		case "vGPU_device_memory_usage_in_bytes":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				vgpumemory := vGPU_device_memory_usage_in_bytes{
					GPUMetadata: GPUMetadata{
						PodName:       labels["podname"],
						PodNamespace:  labels["podnamespace"],
						ContainerName: labels["ctrname"],
						DeviceUUID:    labels["deviceuuid"],
					},
					Value: metric.GetGauge().GetValue(),
				}
				vgpumemorys = append(vgpumemorys, &vgpumemory)
			}
		case "HostCoreUtilization":

			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				host := HostCoreUtilization{
					NodeId: labels["nodeid"],
					Value:  metric.GetGauge().GetValue(),
					Deveiceuuid: Deveiceuuid{
						Deviceuuid: labels["deviceuuid"],
					},
				}
				hosts = append(hosts, &host)

			}
		case "HostGPUMemoryUsage":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				hostusage := HostGPUMemoryUsage{
					NodeId: labels["nodeid"],
					Value:  metric.GetGauge().GetValue(),
					Deveiceuuid: Deveiceuuid{
						Deviceuuid: labels["deviceuuid"],
					},
				}
				hostusages = append(hostusages, &hostusage)

			}
		case "GPUDeviceCoreLimit":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				corelimit := GPUDeviceCoreLimit{
					NodeId: labels["nodeid"],
					Deveiceuuid: Deveiceuuid{
						Deviceuuid: labels["deviceuuid"],
					},
					Value: metric.GetGauge().GetValue(),
				}
				GPUDeviceCoreLimitArr = append(GPUDeviceCoreLimitArr, &corelimit)
			}
		case "GPUDeviceMemoryLimit":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				memorylimit := GPUDeviceMemoryLimit{
					NodeId: labels["nodeid"],
					Deveiceuuid: Deveiceuuid{
						Deviceuuid: labels["deviceuuid"],
					},
					Value: metric.GetGauge().GetValue(),
				}
				GPUDeviceMemoryLimitArr = append(GPUDeviceMemoryLimitArr, &memorylimit)
			}
		case "GPUDeviceCoreAllocated":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				coreallocated := GPUDeviceCoreAllocated{
					NodeId: labels["nodeid"],
					Deveiceuuid: Deveiceuuid{
						Deviceuuid: labels["deviceuuid"],
					},
					Value: metric.GetGauge().GetValue(),
				}
				GPUDeviceCoreAllocatedArr = append(GPUDeviceCoreAllocatedArr, &coreallocated)
			}
		case "GPUDeviceMemoryAllocated":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				memoryallocated := GPUDeviceMemoryAllocated{
					NodeId: labels["nodeid"],
					Deveiceuuid: Deveiceuuid{
						Deviceuuid: labels["deviceuuid"],
					},
					Value: metric.GetGauge().GetValue(),
				}
				GPUDeviceMemoryAllocatedArr = append(GPUDeviceMemoryAllocatedArr, &memoryallocated)
			}
		case "GPUDeviceSharedNum":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				sharednum := GPUDeviceSharedNum{
					NodeId: labels["nodeid"],
					Deveiceuuid: Deveiceuuid{
						Deviceuuid: labels["deviceuuid"],
					},
					Value: metric.GetGauge().GetValue(),
				}
				GPUDeviceSharedNumArr = append(GPUDeviceSharedNumArr, &sharednum)
			}
		case "vGPUPodsDeviceAllocated":
			for _, metric := range metrics {
				labels := labelsToMap(metric.GetLabel())
				vgpupodsdeviceallocated := VGPUPodsDeviceAllocated{
					NodeName:       labels["nodename"],
					PodName:        labels["podname"],
					PodNamespace:   labels["podnamespace"],
					DeviceUsedCore: labels["containeridx"],
					Value:          metric.GetGauge().GetValue(),
				}
				VGPUPodsDeviceAllocatedArr = append(VGPUPodsDeviceAllocatedArr, &vgpupodsdeviceallocated)
			}
		}

	}
	return &VgpuMetrics{
		DeviceMemoryUsage:           vgpumemorys,
		DeviceMemoryLimit:           vgpulimits,
		DeviceUtilization:           deviceutils,
		DeviceMemoryDescOfContainer: devicememorys,
		HostCoreUtilization:         hosts,
		HostGPUMemoryUsage:          hostusages,
		GPUDeviceCoreLimit:          GPUDeviceCoreLimitArr,
		GPUDeviceMemoryLimit:        GPUDeviceMemoryLimitArr,
		GPUDeviceCoreAllocated:      GPUDeviceCoreAllocatedArr,
		GPUDeviceMemoryAllocated:    GPUDeviceMemoryAllocatedArr,
		GPUDeviceSharedNum:          GPUDeviceSharedNumArr,
		VGPUPodsDeviceAllocated:     VGPUPodsDeviceAllocatedArr,
	}
}

func GetNodeInnertIp(node *v1.Node) (string, error) {
	if true {
		return "218.23.2.55", nil
	}
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address, nil
		}
	}
	return "", nil
}
