package appgroup

import (
	"gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateAppGroup(name string, namespace string) *v1alpha1.AppGroup {
	// 创建 AppGroup 对象
	appGroup := &v1alpha1.AppGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppGroup",
			APIVersion: "appgroup.w7.cc/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Finalizers: []string{"appgroup.w7.cc/finalizer"},
		},
	}
	return appGroup
}
