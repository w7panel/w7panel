// nolint
package metrics

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
)

func TestGetResourceDiskUsage(t *testing.T) {

	usge := NewK3kUsage(k8s.NewK8sClient().Sdk)
	sa, err := k8s.NewK8sClient().GetServiceAccount("default", "k3k-s84")
	if err != nil {
		t.Error(err)
	}
	a, b, err := usge.GetResourceDiskUsage(types.NewK3kUser(sa))
	if err != nil {
		t.Error(err)
	}
	t.Log(a, b)

	// a1, b1, err := usge.GetResourceDiskUsage(types.NewK3kUser(sa))
}
