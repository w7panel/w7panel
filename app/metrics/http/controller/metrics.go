package controller

import (
	"fmt"
	"log/slog"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Metrics struct {
	controller.Abstract
}

func (self Metrics) Promhttp(http *gin.Context) {
	promhttp.Handler().ServeHTTP(http.Writer, http.Request)
}

func (self Metrics) PodHandler(http *gin.Context) {
	// 获取所有历史metrics数据
	allMetrics := metrics.PodMetrics.GetAllMetrics()

	// 构建符合Prometheus query_range API格式的响应
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "matrix",
			"result":     []interface{}{}, // 将在下面填充
		},
	}

	// 获取查询参数，确定返回哪种指标
	query := http.DefaultQuery("query", "pod_memory_working_set_bytes")

	// 解析查询参数，提取指标类型和标签选择器
	var queryType, podFilter string

	// 检查是否包含标签选择器 {pod=xxx}
	if idx := strings.Index(query, "{"); idx != -1 {
		if endIdx := strings.Index(query, "}"); endIdx != -1 && endIdx > idx {
			queryType = query[:idx]
			labelSelector := query[idx+1 : endIdx]

			// 解析标签选择器
			labels := strings.Split(labelSelector, ",")
			for _, label := range labels {
				parts := strings.Split(strings.TrimSpace(label), "=")
				if len(parts) == 2 && parts[0] == "pod" {
					// 去除可能的引号
					podFilter = strings.Trim(parts[1], "\"'")
					break
				}
			}
		}
	} else {
		queryType = query
	}

	// 根据查询类型处理不同的指标
	var result []interface{}
	if queryType == "pod_cpu_usage_seconds_total" || queryType == "rate(pod_cpu_usage_seconds_total)" {
		// 处理CPU指标
		podMetrics := make(map[string]map[string][]interface{}) // 命名空间 -> pod名称 -> 时间序列数据

		// 遍历所有历史数据点，组织CPU数据
		for _, timePoint := range allMetrics {
			for namespace, nsPods := range timePoint {
				if _, exists := podMetrics[namespace]; !exists {
					podMetrics[namespace] = make(map[string][]interface{})
				}

				for podName, metric := range nsPods {
					cpuValue := float64(metric.CPUUsage) / 1000.0 // 转换为核心数
					valueStr := fmt.Sprintf("%.4f", cpuValue)     // 格式化为字符串，保留4位小数

					// 创建Prometheus格式的数据点 [timestamp, "value"]
					dataPoint := []interface{}{
						float64(metric.Timestamp.Unix()),
						valueStr,
					}

					if _, exists := podMetrics[namespace][podName]; !exists {
						podMetrics[namespace][podName] = make([]interface{}, 0)
					}
					podMetrics[namespace][podName] = append(podMetrics[namespace][podName], dataPoint)
				}
			}
		}

		// 为每个Pod创建一个时间序列
		for namespace, nsPods := range podMetrics {
			for podName, values := range nsPods {
				// 如果设置了Pod过滤器，只返回匹配的Pod
				if podFilter != "" && podName != podFilter {
					continue
				}

				timeSeries := map[string]interface{}{
					"metric": map[string]string{
						"__name__":  "pod_cpu_usage_seconds_total",
						"pod":       podName,
						"namespace": namespace,
						"job":       "pod",
					},
					"values": values,
				}
				result = append(result, timeSeries)
			}
		}
	} else if queryType == "pod_memory_working_set_bytes" || queryType == "(pod_memory_working_set_bytes)" {
		// 处理内存指标
		podMetrics := make(map[string]map[string][]interface{}) // 命名空间 -> pod名称 -> 时间序列数据

		// 遍历所有历史数据点，组织内存数据
		for _, timePoint := range allMetrics {
			for namespace, nsPods := range timePoint {
				if _, exists := podMetrics[namespace]; !exists {
					podMetrics[namespace] = make(map[string][]interface{})
				}

				for podName, metric := range nsPods {
					memValue := float64(metric.MemoryUsage)
					valueStr := fmt.Sprintf("%.0f", memValue) // 内存通常显示为整数

					// 创建Prometheus格式的数据点 [timestamp, "value"]
					dataPoint := []interface{}{
						float64(metric.Timestamp.Unix()),
						valueStr,
					}

					if _, exists := podMetrics[namespace][podName]; !exists {
						podMetrics[namespace][podName] = make([]interface{}, 0)
					}
					podMetrics[namespace][podName] = append(podMetrics[namespace][podName], dataPoint)
				}
			}
		}

		// 为每个Pod创建一个时间序列
		for namespace, nsPods := range podMetrics {
			for podName, values := range nsPods {
				// 如果设置了Pod过滤器，只返回匹配的Pod
				if podFilter != "" && podName != podFilter {
					continue
				}

				timeSeries := map[string]interface{}{
					"metric": map[string]string{
						"__name__":  "pod_memory_working_set_bytes",
						"pod":       podName,
						"namespace": namespace,
						"job":       "pod",
					},
					"values": values,
				}
				result = append(result, timeSeries)
			}
		}
	}

	// 将结果添加到响应中
	response["data"].(map[string]interface{})["result"] = result

	self.JsonResponseWithoutError(http, response)
	return
}

func (self Metrics) NamespacePodHandler(http *gin.Context) {
	// 获取命名空间参数
	namespace := http.Param("namespace")
	if namespace == "" {
		http.JSON(400, gin.H{"error": "Namespace parameter is required"})
		return
	}

	// 获取特定命名空间的所有历史metrics数据
	allNsMetrics := metrics.PodMetrics.GetNamespaceAllMetrics(namespace)
	if len(allNsMetrics) == 0 {
		http.JSON(404, gin.H{"error": "No metrics found for namespace: " + namespace})
		return
	}

	// 构建符合Prometheus query_range API格式的响应
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "matrix",
			"result":     []interface{}{}, // 将在下面填充
		},
	}

	// 获取查询参数，确定返回哪种指标
	query := http.DefaultQuery("query", "pod_memory_working_set_bytes")

	// 解析查询参数，提取指标类型和标签选择器
	var queryType, podFilter string

	// 检查是否包含标签选择器 {pod=xxx}
	if idx := strings.Index(query, "{"); idx != -1 {
		if endIdx := strings.Index(query, "}"); endIdx != -1 && endIdx > idx {
			queryType = query[:idx]
			labelSelector := query[idx+1 : endIdx]

			// 解析标签选择器
			labels := strings.Split(labelSelector, ",")
			for _, label := range labels {
				parts := strings.Split(strings.TrimSpace(label), "=")
				if len(parts) == 2 && parts[0] == "pod" {
					// 去除可能的引号
					podFilter = strings.Trim(parts[1], "\"'")
					break
				}
			}
		}
	} else {
		queryType = query
	}

	// 根据查询类型处理不同的指标
	var result []interface{}
	if queryType == "pod_cpu_usage_seconds_total" || queryType == "rate(pod_cpu_usage_seconds_total)" {
		// 处理CPU指标
		podMetrics := make(map[string][]interface{}) // pod名称 -> 时间序列数据

		// 遍历所有历史数据点，组织CPU数据
		for _, timePoint := range allNsMetrics {
			for podName, metric := range timePoint {
				cpuValue := float64(metric.CPUUsage) / 1000.0 // 转换为核心数
				valueStr := fmt.Sprintf("%.4f", cpuValue)     // 格式化为字符串，保留4位小数

				// 创建Prometheus格式的数据点 [timestamp, "value"]
				dataPoint := []interface{}{
					float64(metric.Timestamp.Unix()),
					valueStr,
				}

				if _, exists := podMetrics[podName]; !exists {
					podMetrics[podName] = make([]interface{}, 0)
				}
				podMetrics[podName] = append(podMetrics[podName], dataPoint)
			}
		}

		// 为每个Pod创建一个时间序列
		for podName, values := range podMetrics {
			// 如果设置了Pod过滤器，只返回匹配的Pod
			if podFilter != "" && podName != podFilter {
				continue
			}

			timeSeries := map[string]interface{}{
				"metric": map[string]string{
					"__name__":  "pod_cpu_usage_seconds_total",
					"pod":       podName,
					"namespace": namespace,
					"job":       "pod",
				},
				"values": values,
			}
			result = append(result, timeSeries)
		}
	} else if queryType == "pod_memory_working_set_bytes" || queryType == "(pod_memory_working_set_bytes)" {
		// 处理内存指标
		podMetrics := make(map[string][]interface{}) // pod名称 -> 时间序列数据

		// 遍历所有历史数据点，组织内存数据
		for _, timePoint := range allNsMetrics {
			for podName, metric := range timePoint {
				memValue := float64(metric.MemoryUsage)
				valueStr := fmt.Sprintf("%.0f", memValue) // 内存通常显示为整数

				// 创建Prometheus格式的数据点 [timestamp, "value"]
				dataPoint := []interface{}{
					float64(metric.Timestamp.Unix()),
					valueStr,
				}

				if _, exists := podMetrics[podName]; !exists {
					podMetrics[podName] = make([]interface{}, 0)
				}
				podMetrics[podName] = append(podMetrics[podName], dataPoint)
			}
		}

		// 为每个Pod创建一个时间序列
		for podName, values := range podMetrics {
			// 如果设置了Pod过滤器，只返回匹配的Pod
			if podFilter != "" && podName != podFilter {
				continue
			}

			timeSeries := map[string]interface{}{
				"metric": map[string]string{
					"__name__":  "pod_memory_working_set_bytes",
					"pod":       podName,
					"namespace": namespace,
					"job":       "pod",
				},
				"values": values,
			}
			result = append(result, timeSeries)
		}
	}

	// 将结果添加到响应中
	response["data"].(map[string]interface{})["result"] = result

	self.JsonResponseWithoutError(http, response)
	return
}

func (self Metrics) TopNodeHandler(http *gin.Context) {
	// 创建K8s客户端
	k8sClient := k8s.NewK8sClient()

	// 获取节点资源使用情况
	clusterUsage, err := metrics.GetNodeUsage(k8sClient.Sdk)
	if err != nil {
		http.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get node usage: %v", err)})
		return
	}

	// 构建响应
	response := gin.H{
		"nodes": clusterUsage.Nodes,
		"cluster": gin.H{
			"cpu": gin.H{
				"usage":      clusterUsage.TotalCPUUsage,
				"capacity":   clusterUsage.TotalCPUCapacity,
				"percentage": clusterUsage.TotalCPUPercent,
			},
			"memory": gin.H{
				"usage":      clusterUsage.TotalMemoryUsage,
				"capacity":   clusterUsage.TotalMemCapacity,
				"percentage": clusterUsage.TotalMemPercent,
			},
		},
	}

	http.JSON(200, response)
}

func (self Metrics) NodeHandler(http *gin.Context) {
	// 获取所有历史metrics数据
	allMetrics := metrics.NodeMetrics.GetAllMetrics()

	// 构建符合Prometheus query_range API格式的响应
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "matrix",
			"result":     []interface{}{}, // 将在下面填充
		},
	}

	// 获取查询参数，确定返回哪种指标
	queryType := http.DefaultQuery("query", "node_memory_working_set_bytes")

	// 根据查询类型处理不同的指标
	var result []interface{}
	if queryType == "node_cpu_usage_seconds_total" || queryType == "rate(node_cpu_usage_seconds_total)" {
		// 处理CPU指标
		nodeMetrics := make(map[string][]interface{}) // 节点名称 -> 时间序列数据

		// 遍历所有历史数据点，组织CPU数据
		for _, timePoint := range allMetrics {
			for nodeName, metric := range timePoint {
				cpuValue := float64(metric.CPUUsage) / 1000.0 // 转换为核心数
				valueStr := fmt.Sprintf("%.4f", cpuValue)     // 格式化为字符串，保留4位小数

				// 创建Prometheus格式的数据点 [timestamp, "value"]
				dataPoint := []interface{}{
					float64(metric.Timestamp.Unix()),
					valueStr,
				}

				if _, exists := nodeMetrics[nodeName]; !exists {
					nodeMetrics[nodeName] = make([]interface{}, 0)
				}
				nodeMetrics[nodeName] = append(nodeMetrics[nodeName], dataPoint)
			}
		}

		// 为每个节点创建一个时间序列
		for nodeName, values := range nodeMetrics {
			timeSeries := map[string]interface{}{
				"metric": map[string]string{
					"__name__": "node_cpu_usage_seconds_total",
					"instance": nodeName,
					"job":      "node",
				},
				"values": values,
			}
			result = append(result, timeSeries)
		}
	} else if queryType == "node_memory_working_set_bytes" || queryType == "(node_memory_working_set_bytes)" {
		// 处理内存指标
		nodeMetrics := make(map[string][]interface{}) // 节点名称 -> 时间序列数据

		// 遍历所有历史数据点，组织内存数据
		for _, timePoint := range allMetrics {
			for nodeName, metric := range timePoint {
				memValue := float64(metric.MemoryUsage)
				valueStr := fmt.Sprintf("%.0f", memValue) // 内存通常显示为整数

				// 创建Prometheus格式的数据点 [timestamp, "value"]
				dataPoint := []interface{}{
					float64(metric.Timestamp.Unix()),
					valueStr,
				}

				if _, exists := nodeMetrics[nodeName]; !exists {
					nodeMetrics[nodeName] = make([]interface{}, 0)
				}
				nodeMetrics[nodeName] = append(nodeMetrics[nodeName], dataPoint)
			}
		}

		// 为每个节点创建一个时间序列
		for nodeName, values := range nodeMetrics {
			timeSeries := map[string]interface{}{
				"metric": map[string]string{
					"__name__": "node_memory_working_set_bytes",
					"instance": nodeName,
					"job":      "node",
				},
				"values": values,
			}
			result = append(result, timeSeries)
		}
	}

	// 将结果添加到响应中
	response["data"].(map[string]interface{})["result"] = result

	self.JsonResponseWithoutError(http, response)
	return
}

type MetricsInstall struct {
	Installed bool   `json:"installed"`
	BaseUrl   string `json:"baseUrl"`
	Namespace string `json:"namespace"`
}

func (self Metrics) NamespaceResourceHandler(http *gin.Context) {
	// 获取命名空间参数
	namespace := http.Param("namespace")
	if namespace == "" {
		http.JSON(400, gin.H{"error": "Namespace parameter is required"})
		return
	}

	// 获取特定命名空间的所有历史metrics数据
	allNsMetrics := metrics.PodMetrics.GetNamespaceAllMetrics(namespace)
	if len(allNsMetrics) == 0 {
		http.JSON(404, gin.H{"error": "No metrics found for namespace: " + namespace})
		return
	}

	// 计算CPU和内存的总使用量
	var totalCPUUsage, totalMemoryUsage float64
	for _, timePoint := range allNsMetrics {
		for _, metric := range timePoint {
			totalCPUUsage += float64(metric.CPUUsage) / 1000.0 // 转换为核心数
			totalMemoryUsage += float64(metric.MemoryUsage)    // 单位为字节
		}
	}

	// 构建响应
	response := gin.H{
		"namespace": namespace,
		"cpu": gin.H{
			"usage": totalCPUUsage,
			"unit":  "cores",
		},
		"memory": gin.H{
			"usage": totalMemoryUsage,
			"unit":  "bytes",
		},
	}

	http.JSON(200, response)
}

func (self Metrics) VmOperatorInstalled(http *gin.Context) {
	token := http.MustGet("k8s_token").(string)
	k8stoken := k8s.NewK8sToken(token)
	rootSdk := k8s.NewK8sClient() //不能.Sdk
	namespace := "w7-system"
	releaseName := "vm-operator"
	isVirtual := k8stoken.IsVirtual()
	sdk := rootSdk.Sdk
	result := &MetricsInstall{
		BaseUrl:   "/k8s/v1/namespaces/w7-system/services/vmsingle-vm-operator-k8s-offline-metrics-single:8429/proxy/",
		Installed: false,
		Namespace: namespace,
	}
	if isVirtual {
		result.BaseUrl = "/api/v1/namespaces/default/services/vmsingle-w7panel-metrics-k8s-offline-metrics-single:8429/proxy/"
		result.Namespace = "default"
		releaseName = "w7panel-metrics"
		client, err := rootSdk.Channel(token)
		if err != nil {
			slog.Error("channel error", "error", err)
			result.Installed = false
			self.JsonResponseWithoutError(http, result)
			return
		}
		sdk = client
	}
	helmApi := k8s.NewHelm(sdk)
	_, err := helmApi.Info(releaseName, result.Namespace)

	if err != nil {
		result.Installed = false
		self.JsonResponseWithoutError(http, result)
		return
	}
	result.Installed = true
	self.JsonResponseWithoutError(http, result)
}

func (self Metrics) MetricsState(http *gin.Context) {

	type MetricsState struct {
		CanShowClusterMetrics         bool `json:"canShowClusterMetrics"`
		CanShowNodeMetrics            bool `json:"canShowNodeMetrics"`
		CanShowPodMetrics             bool `json:"canShowPodMetrics"`
		NeedInstallMetricsInDashboard bool `json:"needInstallMetricsInDashboard"`
		NeedInstallMetricsInApp       bool `json:"needInstallMetricsInApp"`
	}
	token := http.MustGet("k8s_token").(string)
	k8stoken := k8s.NewK8sToken(token)
	rootSdk := k8s.NewK8sClient() //不能.Sdk

	state := &MetricsState{
		CanShowClusterMetrics:         false,
		CanShowNodeMetrics:            false,
		CanShowPodMetrics:             false,
		NeedInstallMetricsInDashboard: false,
		NeedInstallMetricsInApp:       false,
		// NeedInstallPodMetrics: false,
	}
	releaseName := "w7panel-metrics"
	sdk := rootSdk.Sdk
	helmApi := k8s.NewHelm(sdk)
	_, err := helmApi.Info(releaseName, "default")
	rootInstalled := err == nil
	state.CanShowClusterMetrics = rootInstalled
	if k8stoken.IsK3kCluster() {
		childSdk, err := k8s.NewK8sClient().Channel(token)
		if err != nil {
			slog.Error("channel error", "error", err)
			self.JsonResponseWithoutError(http, state)
			return
		}
		childHelmApi := k8s.NewHelm(childSdk)
		_, childerr := childHelmApi.Info(releaseName, "default")
		if childerr == nil {
			state.CanShowPodMetrics = true
		} else {
			state.NeedInstallMetricsInApp = false
		}
	} else {
		state.CanShowNodeMetrics = rootInstalled
		state.CanShowPodMetrics = rootInstalled
		state.NeedInstallMetricsInApp = !rootInstalled
		state.NeedInstallMetricsInDashboard = !rootInstalled
	}
	// childSdk, err := k8s.NewK8sClient().Channel(token)
	// if err != nil {
	// 	slog.Error("channel error", "error", err)
	// 	self.JsonResponseWithoutError(http, state)
	// 	return
	// }

	self.JsonResponseWithoutError(http, state)

}

func (self Metrics) Usage(http *gin.Context) {
	token := http.MustGet("k8s_token").(string)
	user, err := k3k.TokenToK3kUser(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	uage := metrics.NewK3kUsage(k8s.NewK8sClient().Sdk)
	cpu, memory, cputotal, memorytotal, err := uage.GetResourceUsage(user)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	response := gin.H{
		"cpu": gin.H{
			"usage": cpu.MilliValue(),
			"total": cputotal.MilliValue(),
		},
		"memory": gin.H{
			"usage": memory.Value(),
			"total": memorytotal.Value(),
		},
	}
	self.JsonResponseWithoutError(http, response)
}

func (self Metrics) UsageDisk(http *gin.Context) {
	token := http.MustGet("k8s_token").(string)
	user, err := k3k.TokenToK3kUser(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	uage := metrics.NewK3kUsage(k8s.NewK8sClient().Sdk)
	usage, total, err := uage.GetResourceDiskUsage(user)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	response := gin.H{
		"disk": gin.H{
			"usage": usage,
			"total": total,
		},
	}
	self.JsonResponseWithoutError(http, response)
}
