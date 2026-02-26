package shell

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLoadPid(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk

	pod, err := sdk.ClientSet.CoreV1().Pods("default").Get(sdk.Ctx, "nginx-synnjcft-77db6c45d6-g9kb9", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error getting pod: %v", err)
	}
	err = LoadPid(pod)
	// We can't fully test LoadPid without extensive mocking of helper functions
	// This is a basic test to check the function exists and signature
	if err != nil {
		t.Errorf("LoadPid() error = %v", err)
	}
}
