// nolint
package sa

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteAssociatedResources(t *testing.T) {

	sdk := k8s.NewK8sClient()
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Fatal(err)
	}
	k3kClient := k3ktypes.NewK3kClient(client)
	limitClient := NewLimitRangeClient(client)
	deleteRc := NewDeleteResource(client, k3kClient, limitClient)
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "console-98655",
			Namespace: "k3k-console-98655",
		},
	}
	user := k3ktypes.NewK3kUser(sa)
	deleteRc.deleteAssociatedResources(sdk.Ctx, user)
}
