// nolint
package k3k

import (
	"context"
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSyncIngress(t *testing.T) {
	// os.Setenv("LOCAL_MOCK", "true")
	ing := &K3kSync{
		VirtualName:      "ing-zqyhtpkg",
		VirtualNamespace: "default",
		K3kName:          "console-75780",
		K3kNamespace:     "k3k-console-75780",
		K3kMode:          "virtual",
	}
	err := SyncIngress(ing)
	if err != nil {

		t.Error(err)
	}
}

func TestSyncIngressHttps(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk
	ing, err := sdk.ClientSet.NetworkingV1().Ingresses("default").Get(sdk.Ctx, "ing-jkfsrckbxs", metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}
	err = SyncIngressHttp(ing)
	if err != nil {
		t.Error(err)
	}
}

func TestSyncIngressHttp(t *testing.T) {
	os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("K3K_NAME", "s25")
	os.Setenv("K3K_NAMESPACE", "k3k-s25")
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ing-rhdjzknn",
			Namespace: "default",
		},
	}
	err := SyncIngressHttp(ing)
	if err != nil {
		t.Error(err)
	}
}

func TestSyncConfigmap(t *testing.T) {
	os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("K3K_NAME", "v56")
	os.Setenv("K3K_NAMESPACE", "k3k-v56")
	k3ksync := &K3kSync{
		VirtualName:      "registries",
		VirtualNamespace: "default",
		K3kName:          "v56",
		K3kNamespace:     "k3k-v56",
	}
	err := SyncConfigmap(k3ksync)
	if err != nil {
		t.Error(err)
	}
}

func TestSyncChild(t *testing.T) {
	secret, err := k8s.NewK8sClient().Sdk.ClientSet.CoreV1().Secrets("k3k-console-75780").Get(context.TODO(), "who8-fan-b2-sz-w7-com-tls-secret", metav1.GetOptions{})
	if err != nil {
		t.Log(err)
		return
	}
	err = SyncToChildSecret(secret)
	if err != nil {
		t.Error(err)
	}
}

//who8-fan-b2-sz-w7-com-tls-secret

// func TestCnf(t *testing.T) {
// 	map1 := make(map[string]interface{})
// 	err := yaml.Unmarshal([]byte(k3kregCnf), map1)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	t.Log(map1)
// }
