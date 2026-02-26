// nolint
package sa

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestReal(t *testing.T) {

	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}
	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, types.NamespacedName{Namespace: "default", Name: "s83"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	limiiClient := NewLimitRangeClient(client)
	err = limiiClient.Handle(sdk.Ctx, sa)
	if err != nil {
		t.Error(err)
		return
	}
}
