// nolint
package sa

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestDoRegisterLink(t *testing.T) {

	sdk := k8s.NewK8sClient().Sdk
	sa, err := DoRegisterLink(sdk, "test1", "test1", "yibqvzoz")
	if err != nil {
		t.Error(err)
	}
	t.Log(sa)
}
