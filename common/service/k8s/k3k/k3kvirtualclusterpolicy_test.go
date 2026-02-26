// nolint
package k3k

import (
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCheckPublish(t *testing.T) {
	os.Setenv("USER_AGENT", "we7test-beta")
	sdk := k8s.NewK8sClient()
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	cfg := &v1alpha1.VirtualClusterPolicy{}
	if err := client.Get(sdk.Ctx, types.NamespacedName{Namespace: "default", Name: "dxlxupwa"}, cfg); err != nil {
		t.Error(err)
		return
	}
	// if cfg.Annotations["w7.cc/cost-name"] != "" {
	// 	configmap := &corev1.ConfigMap{}
	// 	err := client.Get(sdk.Ctx, types.NamespacedName{Namespace: "default", Name: cfg.Annotations["w7.cc/cost-name"]}, configmap)
	// 	if err != nil {
	// 		t.Error(err)
	// 		return
	// 	}
	// 	cost, err := k3ktypes.ConfigMapToCost(configmap)
	// 	if err != nil {
	// 		t.Error(err)
	// 		return
	// 	}
	// 	json, err := cost.ToJsonString()
	// 	if err != nil {
	// 		t.Error(err)
	// 		return
	// 	}
	// 	cfg.Annotations["w7.cc/cost"] = json
	// }
	// group := k3ktypes.NewK3kClusterPolicy(cfg)
	// CheckPublish(sdk.Ctx, client, cfg)
	// packages := group.GetCost().Packages
	// for _, pkg := range packages {
	// 	t.Log(pkg)
	// }

	CheckPublish(sdk.Ctx, client, cfg)
}
