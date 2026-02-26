// nolint
package k8s

import (
	"testing"
)

func TestSingleton_GetK3kClusterSdkByConfig(t *testing.T) {
	// 创建测试用的K3kConfig
	k3kconfig := &K3kConfig{
		Name:      "console-75780",
		Namespace: "k3k-console-75780",
		ApiServer: "test-server",
	}

	sdk := NewK8sClient()
	client, err := sdk.GetK3kClusterSdkByConfig(k3kconfig)
	if err != nil {
		t.Errorf("GetK3kClusterSdkByConfig error: %v", err)
	}
	t.Logf("client: %v", client)
}
