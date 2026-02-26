package metrics

import (
	"context"
	"sync"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// PodMetric 存储单个Pod的metrics数据
type PodMetric struct {
	Name        string    // Pod名称
	Namespace   string    // 命名空间
	CPUUsage    int64     // CPU使用量（纳核）
	MemoryUsage int64     // 内存使用量（字节）
	Timestamp   time.Time // 时间戳
}

// PodMetricsStorage 存储所有Pod的历史metrics数据
type PodMetricsStorage struct {
	// 存储结构: [历史记录][命名空间][pod名称]PodMetric
	metrics    []map[string]map[string]PodMetric
	mutex      sync.RWMutex
	maxHistory int
}

// 全局metrics存储实例
var PodMetrics = &PodMetricsStorage{
	maxHistory: 30,
	metrics:    make([]map[string]map[string]PodMetric, 0, 30),
}

// AddMetrics 添加新的metrics数据
func (s *PodMetricsStorage) AddMetrics(podMetrics map[string]map[string]PodMetric) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 如果达到最大历史记录数，删除最旧的数据
	if len(s.metrics) >= s.maxHistory {
		s.metrics = s.metrics[1:]
	}

	// 添加新的metrics数据
	s.metrics = append(s.metrics, podMetrics)
}

// GetLatestMetrics 获取最新的metrics数据
func (s *PodMetricsStorage) GetLatestMetrics() map[string]map[string]PodMetric {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.metrics) == 0 {
		return make(map[string]map[string]PodMetric)
	}

	return s.metrics[len(s.metrics)-1]
}

// GetAllMetrics 获取所有历史metrics数据
func (s *PodMetricsStorage) GetAllMetrics() []map[string]map[string]PodMetric {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 创建一个副本以避免并发访问问题
	result := make([]map[string]map[string]PodMetric, len(s.metrics))
	copy(result, s.metrics)

	return result
}

// GetNamespaceMetrics 获取特定命名空间的最新metrics数据
func (s *PodMetricsStorage) GetNamespaceMetrics(namespace string) map[string]PodMetric {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.metrics) == 0 {
		return make(map[string]PodMetric)
	}

	latestMetrics := s.metrics[len(s.metrics)-1]
	if nsMetrics, exists := latestMetrics[namespace]; exists {
		return nsMetrics
	}

	return make(map[string]PodMetric)
}

// GetNamespaceAllMetrics 获取特定命名空间的所有历史metrics数据
func (s *PodMetricsStorage) GetNamespaceAllMetrics(namespace string) []map[string]PodMetric {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]map[string]PodMetric, 0, len(s.metrics))

	for _, timePoint := range s.metrics {
		if nsMetrics, exists := timePoint[namespace]; exists {
			result = append(result, nsMetrics)
		} else {
			// 如果该时间点没有该命名空间的数据，添加一个空map保持时间序列的连续性
			result = append(result, make(map[string]PodMetric))
		}
	}

	return result
}

// collectPodMetrics 从K8s API获取Pod metrics数据
func collectPodMetrics(metricsClient *metrics.Clientset) error {
	// 获取所有命名空间的Pod metrics
	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// 创建新的metrics map，按命名空间组织
	newMetrics := make(map[string]map[string]PodMetric)

	// 处理每个Pod的metrics
	for _, pod := range podMetrics.Items {
		namespace := pod.Namespace
		podName := pod.Name

		// 确保命名空间map已初始化
		if _, exists := newMetrics[namespace]; !exists {
			newMetrics[namespace] = make(map[string]PodMetric)
		}

		// 计算Pod总的CPU和内存使用量
		var totalCPU, totalMemory int64
		for _, container := range pod.Containers {
			totalCPU += container.Usage.Cpu().MilliValue()
			totalMemory += container.Usage.Memory().Value()
		}

		// 创建Pod metric
		metric := PodMetric{
			Name:        podName,
			Namespace:   namespace,
			CPUUsage:    totalCPU,
			MemoryUsage: totalMemory,
			Timestamp:   time.Now(),
		}

		// 添加到对应命名空间的map中
		newMetrics[namespace][podName] = metric
	}

	// 将新的metrics添加到存储中
	PodMetrics.AddMetrics(newMetrics)
	return nil
}

// StartPodMetrics 启动Pod metrics采集任务
func StartPodMetrics() {
	// 创建metrics客户端
	sdk := k8s.NewK8sClient()

	metricsClient, err := sdk.ToMetricsClient()
	if err != nil {
		return
	}

	// 启动定时任务
	ticker := time.NewTicker(20 * time.Second)
	for {
		if err := collectPodMetrics(metricsClient); err != nil {
			// 记录错误但继续运行
			continue
		}
		<-ticker.C
	}
}
