// nolint
package types

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	ktypes "k8s.io/apimachinery/pkg/types"
)

type mockCostLimitRange struct {
	cost       *K3kCost
	limitRange *LimitRangeQuota
}

func TestGetBasePriceTotal(t *testing.T) {

	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	k3kpolicy := &v1alpha1.VirtualClusterPolicy{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "wxdhqgcy"}, k3kpolicy)
	if err != nil {
		t.Error(err)
		return
	}
	// rs := NewBaseResource(NewK3kGroup(k3kpolicy))
	// total, err := rs.GetBasePriceTotal()
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	// tstr := total.String()
	// t.Logf("total: %v", tstr)
}
