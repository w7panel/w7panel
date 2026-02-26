package appgroup

import (
	"testing"
)

type mockWorkloadWrapper struct {
	isHelm bool
	name   string
}

func TestWorkloadManager_Handle(t *testing.T) {

	// manager := NewWorkLoadTestManager()
	// deployment, err := manager.sdk.ClientSet.AppsV1().Deployments("default").Get(manager.sdk.Ctx, "test1-ypyvijxs", metav1.GetOptions{})
	// if err != nil {
	// 	t.Errorf("Error getting deployment: %v", err)
	// }
	// wrapper := NewWorkloadWrapper(deployment)
	// manager.Handle(wrapper, false)

	// 验证AppGroupApi的Persist方法是否被调用
	// 验证AppGroupApi的Persist方法是否返回错误
}

func TestWorkloadManager_HandleJob(t *testing.T) {

	// manager := NewWorkLoadTestManager()
	// deployment, err := manager.sdk.ClientSet.BatchV1().Jobs("default").Get(manager.sdk.Ctx, "w7-surveyking-ciobztnc-build-ouctj", metav1.GetOptions{})
	// if err != nil {
	// 	t.Errorf("Error getting deployment: %v", err)
	// }
	// wrapper := NewWorkloadWrapper(deployment)
	// manager.Handle(wrapper, false)

	// 验证AppGroupApi的Persist方法是否被调用
	// 验证AppGroupApi的Persist方法是否返回错误
}
