// nolint
package longhorn

import (
	"testing"

	"github.com/w7panel/w7panel/common/service/k8s"
)

func TestInitLonghornVolumesConfig(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	err := initLonghornVolumesConfig(sdk)
	if err != nil {
		t.Error(err)
	}
}
