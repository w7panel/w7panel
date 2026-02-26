package appgroup

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestFixDeployItem(t *testing.T) {

	sdk := k8s.NewK8sClient().Sdk
	groupApi, err := NewAppGroupApi(sdk)
	if err != nil {
		t.Error(err)
	}
	group, err := groupApi.GetAppGroup("default", "noco-base-dksyfldn")
	if err != nil {
		t.Error(err)
	}
	group.ComputeStatus()
	// job, err := sdk.ClientSet.BatchV1().Jobs("default").Get(sdk.Ctx, "noco-base-sfpcwsli-build-install-xspca", metav1.GetOptions{})
	// if err != nil {
	// 	t.Error(err)
	// }

}
