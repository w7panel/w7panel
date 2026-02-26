package k8s

import (
	"testing"
)

func TestCRD_AppGpuClass(t *testing.T) {
	sdk := NewK8sClient().Sdk
	crd := NewCRD(sdk)
	err := crd.AppGpuClass()
	if err != nil {
		t.Error(err)
	}
}

func TestCRD_ApplyCRDS(t *testing.T) {
	sdk := NewK8sClient().Sdk
	crd := NewCRD(sdk)
	err := crd.ApplyCrds()
	if err != nil {
		t.Error(err)
	}
}
