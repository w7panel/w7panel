package metrics

import (
	"context"
	"sync"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// DiskMetric 存储单个磁盘挂载点的metrics数据
type DiskMetric struct {
	MountPoint           string  // 挂载点路径
	TotalBytes           int64   // 总容量（字节）
	UsedBytes            int64   // 已使用容量（字节）
	AvailableBytes       int64   // 可用容量（字节）
	UsagePercentage      float64 // 使用百分比
	TotalInodes          int64   // 总inode数量
	UsedInodes           int64   // 已使用inode数量
	FreeInodes           int64   // 可用inode数量
	InodeUsagePercentage float64 // inode使用百分比
	ReadBytesPerSecond   float64 // 读取速率（字节/秒）
	WriteBytesPerSecond  float64 // 写入速率（字节/秒）
	IoUtilization        float64 // IO使用率
}

// NodeMetric 存储单个节点的metrics数据
type NodeMetric struct {
	Name        string
	CPUUsage    int64 // CPU使用量（纳核）
	MemoryUsage int64 // 内存使用量（字节）
	Timestamp   time.Time
}

// NodeMetricsStorage 存储所有节点的历史metrics数据
type NodeMetricsStorage struct {
	metrics    []map[string]NodeMetric // 存储最近30次的所有节点数据
	mutex      sync.RWMutex
	maxHistory int
}

// 全局metrics存储实例
var NodeMetrics = &NodeMetricsStorage{
	maxHistory: 30,
	metrics:    make([]map[string]NodeMetric, 0, 30),
}

// AddMetrics 添加新的metrics数据
func (s *NodeMetricsStorage) AddMetrics(nodeMetrics map[string]NodeMetric) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 如果达到最大历史记录数，删除最旧的数据
	if len(s.metrics) >= s.maxHistory {
		s.metrics = s.metrics[1:]
	}

	// 添加新的metrics数据
	s.metrics = append(s.metrics, nodeMetrics)
}

// GetLatestMetrics 获取最新的metrics数据
func (s *NodeMetricsStorage) GetLatestMetrics() map[string]NodeMetric {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.metrics) == 0 {
		return make(map[string]NodeMetric)
	}

	return s.metrics[len(s.metrics)-1]
}

// GetAllMetrics 获取所有历史metrics数据
func (s *NodeMetricsStorage) GetAllMetrics() []map[string]NodeMetric {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 创建一个副本以避免并发访问问题
	result := make([]map[string]NodeMetric, len(s.metrics))
	copy(result, s.metrics)

	return result
}

// DiskUsage 磁盘使用情况
type DiskUsage struct {
	MountPoint           string  // 挂载点路径
	TotalBytes           float64 // 总容量（字节）
	UsedBytes            float64 // 已使用容量（字节）
	AvailableBytes       float64 // 可用容量（字节）
	UsagePercentage      float64 // 使用百分比
	TotalInodes          float64 // 总inode数量
	UsedInodes           float64 // 已使用inode数量
	FreeInodes           float64 // 可用inode数量
	InodeUsagePercentage float64 // inode使用百分比
	ReadBytesPerSecond   float64 // 读取速率（字节/秒）
	WriteBytesPerSecond  float64 // 写入速率（字节/秒）
	IoUtilization        float64 // IO使用率
}

// NodeUsage 节点资源使用情况
type NodeUsage struct {
	Name           string      // 节点名称
	CPUUsage       float64     // CPU使用量（核）
	CPUCapacity    float64     // CPU总量（核）
	CPUPercentage  float64     // CPU使用百分比
	MemoryUsage    float64     // 内存使用量（字节）
	MemoryCapacity float64     // 内存总量（字节）
	MemPercentage  float64     // 内存使用百分比
	DiskUsages     []DiskUsage // 磁盘使用情况
}

// ClusterDiskUsage 集群磁盘使用情况
type ClusterDiskUsage struct {
	TotalDiskBytes      float64 // 总磁盘容量（字节）
	UsedDiskBytes       float64 // 已使用磁盘容量（字节）
	AvailableDiskBytes  float64 // 可用磁盘容量（字节）
	TotalDiskPercentage float64 // 总磁盘使用百分比
	TotalInodes         float64 // 总inode数量
	UsedInodes          float64 // 已使用inode数量
	FreeInodes          float64 // 可用inode数量
	InodeUsagePercent   float64 // inode使用百分比
	TotalReadBytesPS    float64 // 总读取速率（字节/秒）
	TotalWriteBytesPS   float64 // 总写入速率（字节/秒）
	AvgIoUtilization    float64 // 平均IO使用率
}

// ClusterUsage 集群资源使用情况
type ClusterUsage struct {
	Nodes            []NodeUsage      // 各节点使用情况
	TotalCPUUsage    float64          // 总CPU使用量（核）
	TotalCPUCapacity float64          // 总CPU容量（核）
	TotalCPUPercent  float64          // 总CPU使用百分比
	TotalMemoryUsage float64          // 总内存使用量（字节）
	TotalMemCapacity float64          // 总内存容量（字节）
	TotalMemPercent  float64          // 总内存使用百分比
	DiskUsage        ClusterDiskUsage // 集群磁盘使用情况
}

// GetNodeUsage 获取节点资源使用情况
func GetNodeUsage(k8sClient *k8s.Sdk) (*ClusterUsage, error) {
	// 获取最新的metrics数据
	latestMetrics := NodeMetrics.GetLatestMetrics()

	// 获取节点信息以获取容量数据
	nodes, err := k8sClient.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	clusterUsage := &ClusterUsage{
		Nodes: make([]NodeUsage, 0, len(nodes.Items)),
	}

	// 处理每个节点的数据
	for _, node := range nodes.Items {
		nodeMetric, exists := latestMetrics[node.Name]
		if !exists {
			continue
		}

		cpuCapacity := float64(node.Status.Capacity.Cpu().MilliValue()) / 1000
		memoryCapacity := float64(node.Status.Capacity.Memory().Value())
		cpuUsage := float64(nodeMetric.CPUUsage) / 1000
		memoryUsage := float64(nodeMetric.MemoryUsage)

		// 处理磁盘使用情况

		nodeUsage := NodeUsage{
			Name:           node.Name,
			CPUUsage:       cpuUsage,
			CPUCapacity:    cpuCapacity,
			CPUPercentage:  (cpuUsage / cpuCapacity) * 100,
			MemoryUsage:    memoryUsage,
			MemoryCapacity: memoryCapacity,
			MemPercentage:  (memoryUsage / memoryCapacity) * 100,
		}

		clusterUsage.Nodes = append(clusterUsage.Nodes, nodeUsage)

		// 累加总量
		clusterUsage.TotalCPUUsage += cpuUsage
		clusterUsage.TotalCPUCapacity += cpuCapacity
		clusterUsage.TotalMemoryUsage += memoryUsage
		clusterUsage.TotalMemCapacity += memoryCapacity
	}

	// 计算总体使用百分比
	if clusterUsage.TotalCPUCapacity > 0 {
		clusterUsage.TotalCPUPercent = (clusterUsage.TotalCPUUsage / clusterUsage.TotalCPUCapacity) * 100
	}
	if clusterUsage.TotalMemCapacity > 0 {
		clusterUsage.TotalMemPercent = (clusterUsage.TotalMemoryUsage / clusterUsage.TotalMemCapacity) * 100
	}

	return clusterUsage, nil
}

// collectNodeMetrics 从K8s API获取节点metrics数据
func collectNodeMetrics(metricsClient *metrics.Clientset) error {
	// 获取所有节点的metrics
	nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// 创建新的metrics map
	newMetrics := make(map[string]NodeMetric)
	// 处理每个节点的metrics
	for _, node := range nodeMetrics.Items {
		// 获取基本metrics
		metric := NodeMetric{
			Name:        node.Name,
			CPUUsage:    node.Usage.Cpu().MilliValue(),
			MemoryUsage: node.Usage.Memory().Value(),
			// DiskMetrics: make(map[string]DiskMetric),
			Timestamp: time.Now(),
		}

		newMetrics[node.Name] = metric
	}

	// 将新的metrics添加到存储中
	NodeMetrics.AddMetrics(newMetrics)
	return nil
}

// Start 启动metrics采集任务
func StartNodeMetrics() {
	// 创建metrics客户端
	sdk := k8s.NewK8sClient()

	metricsClient, err := sdk.ToMetricsClient()
	if err != nil {
		return
	}

	// 启动定时任务
	ticker := time.NewTicker(20 * time.Second)
	for {
		if err := collectNodeMetrics(metricsClient); err != nil {
			// 记录错误但继续运行
			continue
		}
		<-ticker.C
	}
}
