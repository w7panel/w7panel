package metrics

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	v1 "k8s.io/api/core/v1"
)

type NodeScrape struct {
	host string
	port string
}

func NewNodeScrapeUseNode(node *v1.Node, port string) (*NodeScrape, error) {
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeInternalIP {
			// return NewNodeScrape(address.Address, port), nil
			if helper.IsLocalMock() {
				return NewNodeScrape("218.23.2.55", port), nil
			}
			return NewNodeScrape(address.Address, port), nil
		}
	}
	return nil, fmt.Errorf("no internal ip found")
}

func NewNodeScrape(host string, port string) *NodeScrape {
	return &NodeScrape{
		host: host,
		port: port,
		// podMetrics:  podMetrics,
	}
}

func (n *NodeScrape) Scrape() (*VgpuMetrics, error) {
	// metrics, err := n.GetMetricsBytes()
	// if err != nil {
	// 	log.Fatalf("Error getting metrics: %v", err)
	// 	return err
	// }
	response, err := helper.RetryHttpClient().R().Get(fmt.Sprintf("http://%s:%s/metrics", n.host, n.port))
	if err != nil {
		slog.Error("Error getting metrics: %v", "err", err)
		return nil, err
	}
	bytedata := response.Body()
	mapdata, err := n.Parse(string(bytedata))
	// mapdata, err := n.Parse(n.GetTestData())
	if err != nil {
		slog.Error("Error parsing metrics: %v", "err", err)
		return nil, err
	}

	vgpuStat := ParseMapFamily(mapdata)
	return vgpuStat, nil

}
func (n *NodeScrape) GetTestData() string {
	str := `
# HELP Device_last_kernel_of_container Container device last kernel description
# TYPE Device_last_kernel_of_container gauge
Device_last_kernel_of_container{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 80450
# HELP Device_memory_desc_of_container Container device meory description
# TYPE Device_memory_desc_of_container counter
Device_memory_desc_of_container{context="0",ctrname="gpu-cxeyttpu",data="0",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",module="0",offset="0",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 0
# HELP Device_utilization_desc_of_container Container device utilization description
# TYPE Device_utilization_desc_of_container gauge
Device_utilization_desc_of_container{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 0
# HELP HostCoreUtilization GPU core utilization
# TYPE HostCoreUtilization gauge
HostCoreUtilization{deviceidx="0",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",zone="vGPU"} 0
# HELP HostGPUMemoryUsage GPU device memory usage
# TYPE HostGPUMemoryUsage gauge
HostGPUMemoryUsage{deviceidx="0",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",zone="vGPU"} 2.87244288e+08
# HELP vGPU_device_memory_limit_in_bytes vGPU device limit
# TYPE vGPU_device_memory_limit_in_bytes gauge
vGPU_device_memory_limit_in_bytes{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 3.145728e+08
# HELP vGPU_device_memory_usage_in_bytes vGPU device usage
# TYPE vGPU_device_memory_usage_in_bytes gauge
vGPU_device_memory_usage_in_bytes{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 0	
	`
	return str
}

func (n *NodeScrape) GetMetricsBytes() ([]byte, error) {
	data := `
# HELP Device_last_kernel_of_container Container device last kernel description
# TYPE Device_last_kernel_of_container gauge
Device_last_kernel_of_container{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 80450
# HELP Device_memory_desc_of_container Container device meory description
# TYPE Device_memory_desc_of_container counter
Device_memory_desc_of_container{context="0",ctrname="gpu-cxeyttpu",data="0",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",module="0",offset="0",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 0
# HELP Device_utilization_desc_of_container Container device utilization description
# TYPE Device_utilization_desc_of_container gauge
Device_utilization_desc_of_container{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 0
# HELP HostCoreUtilization GPU core utilization
# TYPE HostCoreUtilization gauge
HostCoreUtilization{deviceidx="0",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",zone="vGPU"} 0
# HELP HostGPUMemoryUsage GPU device memory usage
# TYPE HostGPUMemoryUsage gauge
HostGPUMemoryUsage{deviceidx="0",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",zone="vGPU"} 2.87244288e+08
# HELP vGPU_device_memory_limit_in_bytes vGPU device limit
# TYPE vGPU_device_memory_limit_in_bytes gauge
vGPU_device_memory_limit_in_bytes{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 3.145728e+08
# HELP vGPU_device_memory_usage_in_bytes vGPU device usage
# TYPE vGPU_device_memory_usage_in_bytes gauge
vGPU_device_memory_usage_in_bytes{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 0	
	`
	return []byte(data), nil
	// path, ok := os.LookupEnv("KO_DATA_PATH")
	// if !ok {
	// 	log.Fatalf("KO_DATA_PATH environment variable not set")
	// 	return nil, fmt.Errorf("KO_DATA_PATH environment variable not set")
	// }
	// return os.ReadFile(path + "/test/31992.txt")
}

func (n *NodeScrape) Parse(metricsData string) (map[string]*dto.MetricFamily, error) {
	// metricsData := `
	// # HELP Device_last_kernel_of_container Container device last kernel description
	// # TYPE Device_last_kernel_of_container gauge
	// Device_last_kernel_of_container{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 80450
	// # HELP Device_memory_desc_of_container Container device meory description
	// # TYPE Device_memory_desc_of_container counter
	// Device_memory_desc_of_container{context="0",ctrname="gpu-cxeyttpu",data="0",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",module="0",offset="0",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 0
	// # HELP Device_utilization_desc_of_container Container device utilization description
	// # TYPE Device_utilization_desc_of_container gauge
	// Device_utilization_desc_of_container{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 0
	// # HELP HostCoreUtilization GPU core utilization
	// # TYPE HostCoreUtilization gauge
	// HostCoreUtilization{deviceidx="0",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",zone="vGPU"} 0
	// # HELP HostGPUMemoryUsage GPU device memory usage
	// # TYPE HostGPUMemoryUsage gauge
	// HostGPUMemoryUsage{deviceidx="0",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",zone="vGPU"} 2.87244288e+08
	// # HELP vGPU_device_memory_limit_in_bytes vGPU device limit
	// # TYPE vGPU_device_memory_limit_in_bytes gauge
	// vGPU_device_memory_limit_in_bytes{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 3.145728e+08
	// # HELP vGPU_device_memory_usage_in_bytes vGPU device usage
	// # TYPE vGPU_device_memory_usage_in_bytes gauge
	// vGPU_device_memory_usage_in_bytes{ctrname="gpu-cxeyttpu",deviceuuid="GPU-59ab1128-3130-b95a-9870-4bd748926e97",podname="gpu-cxeyttpu-5bc96fc469-8q2zb",podnamespace="default",vdeviceid="0",zone="vGPU"} 0
	// `
	parser := expfmt.TextParser{}
	metrics, err := parser.TextToMetricFamilies(strings.NewReader(metricsData))
	if err != nil {
		slog.Error("Error parsing metrics: %v", "err", err)
	}
	return metrics, err
}
func parseMetrics(ip, port string) (map[string]*dto.MetricFamily, error) {
	http.DefaultClient.Timeout = 10 * time.Second
	response, err := http.Get(fmt.Sprintf("http://%s:%s/metrics", ip, port))
	if err != nil {
		log.Fatalf("Error fetching metrics: %v", err)
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Error fetching metrics: %v", err)
		return nil, err
	}
	parser := expfmt.TextParser{}
	metrics, err := parser.TextToMetricFamilies(response.Body)
	if err != nil {
		log.Fatalf("Error parsing metrics: %v", err)
	}
	return metrics, err
}
