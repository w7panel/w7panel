package overselling

import (
	"errors"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	corev1 "k8s.io/api/core/v1"
)

// CanAddResource 检查是否可以添加指定资源
// 通过比较当前超售资源和已使用资源+新增资源的和来判断
func CanAddResource(rs *Resource, callback func(*corev1.ServiceAccount) *Resource) error {
	client, err := NewResourceClient(k8s.NewK8sClient().Sdk)
	if err != nil {
		return err
	}
	// 获取当前超售资源配额
	current, err := client.GetOverlingResource()
	if err != nil {
		return err
	}
	// 获取已使用的资源
	used, err := client.GetUsed(callback)
	if err != nil {
		return err
	}
	// 克隆已使用资源并添加新资源
	clone := used.Clone()
	usedStr := used.JsonString()
	//打印 clone.String()
	clone.Add(*rs)
	currentStr := current.JsonString()
	cloneStr := clone.JsonString()
	slog.Info("debug can add resource", "usedStr", usedStr, "currentStr", currentStr, "cloneStr", cloneStr)
	// 检查当前资源是否大于等于（包含）新增后的总使用量
	if current.Dayu(*clone) {
		return nil
	}
	// return nil
	return errors.New("resource not enough")
}
