package k3k

import (
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetUnUsedStorageSize(user *types.K3kUser) (*resource.Quantity, error) {
	sdk := k8s.NewK8sClient().Sdk
	res, err := sdk.ClientSet.CoreV1().ResourceQuotas(user.GetK3kNamespace()).Get(sdk.Ctx, user.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	total := res.Status.Hard[v1.ResourceRequestsStorage]
	used := res.Status.Used[v1.ResourceRequestsStorage]

	totalClone := total.DeepCopy()
	totalClonePointer := &totalClone
	totalClonePointer.Sub(used)
	return totalClonePointer, nil

}

func Resize(user *types.K3kUser, resizeTo resource.Quantity) error {
	if !user.IsShared() {
		// 非共享用户不支持扩容系统存储
		return nil
	}
	sdk := k8s.NewK8sClient().Sdk
	sigClient, err := sdk.ToSigClient()
	if err != nil {
		return err
	}
	unused, err := GetUnUsedStorageSize(user)
	if err != nil {
		return err
	}
	unusedStr := unused.String()
	resizeToStr := resizeTo.String()
	slog.Info(fmt.Sprintf("unused: %v, resizeTo: %v", unusedStr, resizeToStr))
	if user.CanResizeSysStorage(*unused, resizeTo) {
		_, err = controllerutil.CreateOrPatch(sdk.Ctx, sigClient, user.ServiceAccount, func() error {
			user.ResizeSysStorage(resizeTo)
			return nil
		})
		return err
	}
	return nil
}
