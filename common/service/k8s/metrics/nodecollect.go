package metrics

import (
	"log/slog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type NodeCollect struct {
	// 定义指标的字段
	nodes           []*v1.Node
	nodeMetricsList *v1beta1.NodeMetricsList
	podMetricsList  *v1beta1.PodMetricsList
}

func NewNodeCollect(nodes []*v1.Node, node *v1beta1.NodeMetricsList, pod *v1beta1.PodMetricsList) *NodeCollect {
	return &NodeCollect{
		nodes:           nodes,
		nodeMetricsList: node,
		podMetricsList:  pod,
	}
}

func (s *NodeCollect) Collect() {
	// 实现收集逻辑

	nodeNamedMetrics := make(map[string]*v1beta1.NodeMetrics)
	for _, nodeM := range s.nodeMetricsList.Items {
		// ...
		nodeNamedMetrics[nodeM.Name] = &nodeM
	}

	for _, node := range s.nodes {
		// ...
		ip, err := s.GetNodeInnertIp(node)
		if err != nil {
			slog.Error("GetNodeInnertIp", slog.String("error", err.Error()))
			continue
		}
		nodescrape := NewNodeScrape(ip, HAMIPORT)
		vgpu, err := nodescrape.Scrape()
		if err != nil {
			slog.Error("Scrape", slog.String("error", err.Error()))
			continue
		}
		nodeM, ok := nodeNamedMetrics[node.Name]
		if !ok {
			continue
		}
		tr := NewNodeMetricsTransform(vgpu, nodeM, s.podMetricsList.Items)
		tr.Transfom()
		slog.Info("Scrape", slog.String("ip", ip), slog.Any("vgpu", vgpu))
	}
}

func (s *NodeCollect) CollectNode(node *v1.Node) error {
	return nil
}

func (s *NodeCollect) GetNodeInnertIp(node *v1.Node) (string, error) {
	// if true {
	// 	return "218.23.2.55", nil
	// }
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address, nil
		}
	}
	return "", nil
}
