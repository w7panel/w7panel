package higress

import (
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestInitW7ProxyPlugin(t *testing.T) {
	os.Setenv("KO_DATA_PATH", "/home/workspace/k8s-offline/kodata")
	client, err := k8s.NewK8sClient().Sdk.ToSigClient()
	if err != nil {
		t.Error(err)
	}
	err = InitW7ProxyPlugin(t.Context(), k8s.GetScheme(), client)
	if err != nil {
		t.Error(err)
	}
}
