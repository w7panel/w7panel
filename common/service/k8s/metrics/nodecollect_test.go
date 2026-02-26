package metrics

import (
	"log/slog"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNodeCollect_Collect(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	ctx := sdk.Ctx
	metricsClient, err := sdk.ToMetricsClient()
	if err != nil {
		slog.Error("ToMetricsClient", slog.String("error", err.Error()))
		return
	}
	nodes, err := sdk.ClientSet.CoreV1().Nodes().List(sdk.Ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("List", slog.String("error", err.Error()))
		return
	}
	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("List", slog.String("error", err.Error()))
		return
	}
	nodeNamedMetrics := make(map[string]*v1beta1.NodeMetrics)
	for _, nodeM := range nodeMetricsList.Items {
		// ...
		nodeNamedMetrics[nodeM.Name] = &nodeM
	}
	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses("default").List(ctx, metav1.ListOptions{})

	nodesPoints := []*v1.Node{}
	for _, node := range nodes.Items {
		// ...
		nodesPoints = append(nodesPoints, &node)
	}

	collect := NewNodeCollect(nodesPoints, nodeMetricsList, podMetricsList)
	collect.Collect()
	t.Log(collect)
	t.Log(nodeNamedMetrics)
	t.Log(podMetricsList)
	t.Log(nodeMetricsList)
}
