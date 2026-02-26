package shell

import (
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *k3sConfigController) loadPublicIp(node *v1.Node) error {
	loadIp, ok := node.Labels["w7.cc/load-public-ip"]
	if !ok || loadIp != "true" {
		slog.Debug("not need load public ip")
		return nil
	}
	if isCurrentDaemonsetNode(node) {
		ip, err := helper.MyIp()
		if err != nil {
			slog.Error("failed to get my ip", "err", err)
			node.Labels["w7.cc/load-public-ip"] = "false"
			_, err := s.Sdk.ClientSet.CoreV1().Nodes().Update(s.Ctx, node, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to get my ip update node")
			}
			return fmt.Errorf("failed to get my ip")
		}
		node.Labels["w7.public-ip"] = ip
		node.Labels["w7.cc/load-public-ip"] = "false"
		_, err = s.Sdk.ClientSet.CoreV1().Nodes().Update(s.Ctx, node, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to get my ip update node")
		}
	}
	return nil

}
