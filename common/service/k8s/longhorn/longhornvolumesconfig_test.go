// nolint
package longhorn

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestInitLonghornVolumesConfig(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	err := initLonghornVolumesConfig(sdk)
	if err != nil {
		t.Error(err)
	}
}
