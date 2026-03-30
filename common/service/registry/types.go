package registry

import (
	containerd "github.com/containerd/containerd/v2/client"
	"github.com/w7panel/w7panel/common/helper"
)

const (
	registryNamespace = "k8s.io"
	// containerdRoot    = "/run/k3s/containerd"
	// containerdAddr    = "/run/k3s/containerd/containerd.sock"
	defaultDamain       = "registry.local.w7.cc"
	debugcontainerdRoot = "/var/lib/containerd"
	debugcontainerdAddr = "/run/containerd/containerd.sock"
	k3sContainerAddr    = "/var/run/k3s/containerd/containerd.sock"
	k3sContainerRoot    = "/var/lib/rancher/k3s/agent/containerd"
)

func CreateClient() (*containerd.Client, error) {
	client, err := containerd.New(containerAddr(), containerd.WithDefaultNamespace(registryNamespace))
	if err != nil {
		return nil, err
	}
	return client, err
}

func containerRoot() string {
	if helper.IsChildAgent() || helper.IsAgent() {
		return k3sContainerRoot
	}
	if helper.IsLocalMock() || helper.IsDebug() {
		return debugcontainerdRoot
	}
	return ""
}

func containerAddr() string {
	if helper.IsLocalMock() || helper.IsDebug() {
		return debugcontainerdAddr
	}
	if helper.IsChildAgent() || helper.IsAgent() {
		return k3sContainerAddr
	}

	return ""

}
