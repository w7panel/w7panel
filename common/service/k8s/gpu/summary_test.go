package gpu

import (
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"

	"github.com/stretchr/testify/assert"
)

func TestClusterSummary_Summary(t *testing.T) {
	// client := fake.NewSimpleClientset()
	os.Setenv("LOCAL_MOCK", "true")
	sdk := k8s.NewK8sClientInner()
	clusterSummary := NewClusterSummary(sdk)

	summary, err := clusterSummary.Summary()
	if err != nil {
		t.Error(err)
	}
	print(summary)
	assert.NoError(t, err)
	assert.Equal(t, int32(5), summary.GPUDeviceSharedNum)
	assert.Equal(t, int32(1024), summary.GPUDeviceMemoryAllocated)
}
