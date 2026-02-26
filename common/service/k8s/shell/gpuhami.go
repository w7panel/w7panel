package shell

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/gpu"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types2 "k8s.io/apimachinery/pkg/types"
)

func (s *k3sConfigController) StartGpuTimer() {
	// go func() {
	// 	for {
	// 		select {
	// 		case <-time.After(30 * time.Second):
	// 			slog.Info("check gpu operator after 10s")

	// 			nodes, err := s.nodeLister.List(labels.NewSelector())
	// 			if err != nil {
	// 				continue
	// 			}
	// 			for _, node := range nodes {
	// 				s.checkGpuOperator(node)
	// 			}
	// 		}
	// 	}
	// }()
}

func (s *k3sConfigController) HandleGpuDaemonset(new *appsv1.DaemonSet, isDelete bool) {
	// slog.Debug("handle gpu daemonset", "name", new.Name)
	if true {
		return
	}
	if new.Name == "nvidia-operator-validator" {
		// if new.Name == "nvidia-driver-daemonset" {
		if isDelete {
			s.gpuOperatorIsDeployed = false
			s.gopMode(gpu.UNINSTALL)
			return
		}
		if new.Status.CurrentNumberScheduled > 0 {
			s.gpuOperatorIsDeployed = true
		}
	}
	if new.Name == "hami-device-plugin" {
		if isDelete {
			// s.gpuOperatorIsDeployed = false
			s.hamiMode(gpu.UNINSTALL)
			return
		}
		if new.Status.CurrentNumberScheduled > 0 {
			s.hamiMode(gpu.INSTALLED)
		}
	}
}

func (s *k3sConfigController) hamiMode(mode gpu.GpuInstallMode) {
	data := map[string]string{
		"hami-mode": string(mode),
	}
	if mode == gpu.UNINSTALL {
		data["enabled"] = "false"
	}
	gpu.PatchK3sGpu(s.Sdk, data)
}

func (s *k3sConfigController) gopMode(mode gpu.GpuInstallMode) {
	data := map[string]string{
		"gpu-operator-mode": string(mode),
	}
	if mode == gpu.UNINSTALL {
		data["enabled"] = "false"
	}
	gpu.PatchK3sGpu(s.Sdk, data)
}

func (s *k3sConfigController) checkGpuOperator(node *v1.Node) error {
	//Driver file installation is complete
	slog.Debug("check gpu operator", "node", node.Name)
	if !isCurrentDaemonsetNode(node) {
		return nil
	}
	// if !s.gpuOperatorIsDeployed {
	// 	return nil
	// }

	gpuOk := helper.NvidiaReadyFileExites()
	if gpuOk {
		s.gopMode(gpu.INSTALLED)
		val, ok := node.Labels["gpu"]
		if ok && val == "on" {
			return nil
		}
		patchData := []byte(`{"metadata":{"labels":{"gpu":"on"}}}`)
		//删除标签
		_, err := s.Sdk.ClientSet.CoreV1().Nodes().Patch(s.Ctx, node.Name, types2.MergePatchType, patchData, metav1.PatchOptions{})
		if err != nil {
			slog.Error("patch node error", "err", err)
			return err
		}
	} else {
		//gpu 标签改为off
		val, ok := node.Labels["gpu"]
		if !ok {
			return nil
		}
		if ok && val == "off" {
			return nil
		}
		if ok && val == "on" {
			patchData := []byte(`{"metadata":{"labels":{"gpu":"off"}}}`)
			_, err := s.Sdk.ClientSet.CoreV1().Nodes().Patch(s.Ctx, node.Name, types2.MergePatchType, patchData, metav1.PatchOptions{})
			if err != nil {
				slog.Error("patch node error", "err", err)
			}
		}
	}
	return nil
}
