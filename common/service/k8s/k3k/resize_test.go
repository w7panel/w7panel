// nolint
package k3k

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	ktypes "k8s.io/apimachinery/pkg/types"
)

func TestResize(t *testing.T) {
	// Setup test environment
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "s100"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	err = Resize(k3kUser, resource.MustParse("8Gi"))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(k3kUser)

}
