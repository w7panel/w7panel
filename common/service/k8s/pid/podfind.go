package pid

import (
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type podFind struct {
	root   *kubernetes.Clientset
	client *kubernetes.Clientset
}

func (f *podFind) GetVirtualClusterNodePod(namespace, hostIp string) (*corev1.Pod, error) {
	nodes, err := f.client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	podName := ""
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Address == hostIp {
				podName = node.Name
				break
			}
		}
	}
	if podName == "" {
		return nil, fmt.Errorf("not found node")
	}
	pod, err := f.root.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func (f *podFind) GetPanelAgentPod(hostIp string) (*corev1.Pod, error) {
	dsPods, err := f.getRootClusterDaemonPodList()
	if err != nil {
		slog.Error("not find dspods", "err", err)
		return nil, err
	}
	for _, pod := range dsPods.Items {
		if pod.Status.HostIP == hostIp {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("not found pod")
}

// huoq
func (f *podFind) getRootClusterDaemonPodList() (*corev1.PodList, error) {
	daemonsetPods, err := f.root.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{LabelSelector: "w7.cc/daemonset=w7"})
	if err != nil {
		slog.Warn("get daemonset pods error", "err", err)
		return nil, err
	}
	return daemonsetPods, nil
}
