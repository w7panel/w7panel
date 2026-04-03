package containerd

import (
	"context"

	v2client "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/distribution/reference"
	"github.com/w7panel/w7panel/common/helper"
)

const (
	NS = "k8s.io"
	// containerdRoot    = "/run/k3s/containerd"
	// containerdAddr    = "/run/k3s/containerd/containerd.sock"
	REPO                = "registry.local.w7.cc"
	debugcontainerdRoot = "/var/lib/containerd"
	debugcontainerdAddr = "/run/containerd/containerd.sock"
	k3sContainerAddr    = "/var/run/k3s/containerd/containerd.sock"
	k3sContainerRoot    = "/var/lib/rancher/k3s/agent/containerd"
)

func withNamespace(ctx context.Context) context.Context {
	return namespaces.WithNamespace(ctx, NS)
}

func CreateClient() (*v2client.Client, error) {
	client, err := v2client.New(containerAddr(), v2client.WithDefaultNamespace(NS))
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

func parseRef(ref string) (reference.Reference, error) {
	return reference.Parse(ref)
}
