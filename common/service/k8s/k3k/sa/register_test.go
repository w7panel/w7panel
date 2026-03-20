// nolint
package sa

import (
	"testing"

	"github.com/w7panel/w7panel/common/service/k8s"
)

func TestDoRegisterLink(t *testing.T) {

	sdk := k8s.NewK8sClient().Sdk
	sa, err := DoRegisterLink(sdk, "test1", "test1", "yibqvzoz")
	if err != nil {
		t.Error(err)
	}
	t.Log(sa)
}
