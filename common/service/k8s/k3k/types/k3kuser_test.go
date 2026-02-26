// nolint
package types

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetClusterStorageRequestSize(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}
	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, types.NamespacedName{Namespace: "default", Name: "g1"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := NewK3kUser(sa)
	price, err := k3kUser.GetBasePrice()
	if err != nil {
		t.Error(err)
		return
	}
	pstr := price.String()
	t.Log(pstr)
	buy := k3kUser.NeedCreateOrder()
	t.Log(buy)
	// size := k3kUser.GetClusterSysStorageRequestSize()
	// defaultSize := k3kUser.GetStorageRequestSize()
	t1 := k3kUser.GetLimitRange().GetHardRequestStorage().String()
	t2 := k3kUser.GetLimitRange().GetHardRequestStorage().ScaledValue(resource.Kilo)
	t.Log(t1, t2)
	scName := k3kUser.GetStorageClass()
	t.Log(scName)
}
