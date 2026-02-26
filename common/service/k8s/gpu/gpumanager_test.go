package gpu

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestInstallGpuOperator(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk

	_, err := NewGpuManager(sdk, "", "")
	if err != nil {
		t.Errorf("NewGpuManager failed: %v", err)
	}
	// res := g.InstallGpuOperator()
}

func TestGpuResponse(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk

	g, err := NewGpuManager(sdk, "", "")
	if err != nil {
		t.Errorf("NewGpuManager failed: %v", err)
	}
	res := g.ToJsonStruct()
	t.Log(res)
}

func TestGpuEnabled(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk

	g, err := NewGpuManager(sdk, "", "")
	if err != nil {
		t.Errorf("NewGpuManager failed: %v", err)
	}
	res := g.GpuEnabled(true)
	t.Log(res)
}

func TestGpuOpInstall(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk

	g, err := NewGpuManager(sdk, "", "")
	if err != nil {
		t.Errorf("NewGpuManager failed: %v", err)
	}
	res := g.InstallGpuOperator(true, "")
	t.Log(res)
}

func TestHamiInstall(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk

	g, err := NewGpuManager(sdk, "", "")
	if err != nil {
		t.Errorf("NewGpuManager failed: %v", err)
	}
	res := g.InstallHami("")
	t.Log(res)
}
