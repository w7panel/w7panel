package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/cgroups"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/containerd/cgroups/v3/cgroup2/stats"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var lastMetrics *stats.Metrics
var lastTime time.Time

var cpuCount int

var storage *cgroupStorage

func init() {
	storage = &cgroupStorage{}
}
func StartCroupMetrics() {
	if helper.IsK3kVirtual() {
		sdk := k8s.NewK8sClient()

		metricsClient, err := sdk.ToSigClient()
		if err != nil {
			return
		}
		collectReport(metricsClient) // 首次执行
		// 启动定时任务
		ticker := time.NewTicker(10 * time.Second)
		for {
			if err := collectReport(metricsClient); err != nil {
				// 记录错误但继续运行
				continue
			}
			<-ticker.C
		}
	}
}
func collectReport(client sigclient.Client) error {
	slog.Error("collectReport start")
	rs, stat, err := collectMetrics()
	if err != nil {
		return err
	}
	data := map[string]string{
		"cpu":                    fmt.Sprintf("%d", rs.Cpu().MilliValue()),
		"memory":                 fmt.Sprintf("%d", rs.Memory().Value()),
		"memory.current":         strconv.FormatUint(stat.Memory.Usage, 10),
		"memory.inactiveFile":    strconv.FormatUint(stat.Memory.InactiveFile, 10),
		"memory.anno":            strconv.FormatUint(stat.Memory.Anon, 10),
		"memory.file":            strconv.FormatUint(stat.Memory.File, 10),
		"memory.kernel":          strconv.FormatUint(stat.Memory.GetKernelStack(), 10),
		"memory.mb":              fmt.Sprintf("%d", rs.Memory().Value()/1048576),
		"memory.current.mb":      strconv.FormatUint(stat.Memory.Usage/1048576, 10),
		"memory.inactiveFile.mb": strconv.FormatUint(stat.Memory.InactiveFile/1048576, 10),
		"memory.anno.mb":         strconv.FormatUint(stat.Memory.Anon/1048576, 10),
		"memory.file.mb":         strconv.FormatUint(stat.Memory.File/1048576, 10),
		"memory.kernel.mb":       strconv.FormatUint(stat.Memory.GetKernelStack()/1048576, 10),
	}
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metrics",
			Namespace: "default",
		},
	}
	_, err = controllerutil.CreateOrPatch(context.Background(), client, configmap, func() error {
		if configmap.Labels == nil {
			configmap.Labels = make(map[string]string)
		}
		configmap.Labels["w7.cc/sync"] = "true"
		configmap.Labels["w7.cc/sync-type"] = "metrics"
		configmap.Data = data
		return nil
	})
	return err
}

func collectMetrics() (corev1.ResourceList, *stats.Metrics, error) {
	stat, err := cgroups.CurrentStat()
	if err != nil {
		return corev1.ResourceList{}, nil, fmt.Errorf("current stat error")
	}
	usage := stat.Memory.GetAnon() + stat.Memory.GetFile() + stat.Memory.GetKernelStack()
	if usage < 0 {
		usage = 0
	}
	cpuUsage := stat.CPU.GetUsageUsec() * 1000 // opencontainerd fs2.go
	metricsPoint := &MetricsPoint{
		StartTime:         time.Now(),
		Timestamp:         time.Now(),
		CumulativeCpuUsed: cpuUsage,
		MemoryUsage:       usage,
	}

	if storage.prev == nil {
		storage.prev = metricsPoint
		return corev1.ResourceList{}, nil, fmt.Errorf("prev is nil error")
	}
	storage.last = metricsPoint

	rs, err := resourceUsage(*storage.last, *storage.prev)
	if err != nil {
		return rs, nil, err
	}
	storage.prev = storage.last
	return rs, stat, err
}

func collectCgroupMetrics(client sigclient.Client) error {
	return nil
}
