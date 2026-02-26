package longhorn

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"

	"github.com/stretchr/testify/assert"
)

var NewLonghornClientFunc = func(sdk k8s.Sdk) (*MockLonghornClient, error) {
	return nil, nil
}

type MockLonghornAPI struct {
	GetVolumeFunc     func(name string) (*Volume, error)
	GetDisksCountFunc func(selector []string) (int, error)
}

type MockLonghornClient struct {
	GetVolumeFunc     func(name string) (*Volume, error)
	GetDisksCountFunc func(selector []string) (int, error)
}

type Volume struct {
	Name string
	Spec VolumeSpec
}

type VolumeSpec struct {
	DiskSelector     []string
	NumberOfReplicas int
}

func TestVolumeUpdateReplicaCount(t *testing.T) {
	// 创建一个模拟的SDK实例
	sdk := k8s.NewK8sClientInner()

	// 调用被测试的方法
	err := VolumeUpdateReplicaCount(sdk, "disktestmany")

	// 断言结果
	assert.NoError(t, err)
	// assert.Equal(t, 3, 4)
}

func TestVolumeUpdateNodeLabels(t *testing.T) {
	// 创建一个模拟的SDK实例
	// sdk := k8s.NewK8sClientInner()

	WatchLonghorn()
	// assert.Equal(t, 3, 4)
}
