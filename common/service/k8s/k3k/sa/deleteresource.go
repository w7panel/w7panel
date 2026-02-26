package sa

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// const k3kFinalizerName = "k3k.sa/finalizer"

type DeleteResource struct {
	client.Client
	k3kClient        *k3ktypes.K3kClient
	limitrangeclient *Limitrangeclient
}

func NewDeleteResource(client client.Client, k3kClient *k3ktypes.K3kClient, limitrangeclient *Limitrangeclient) *DeleteResource {
	return &DeleteResource{
		Client:           client,
		k3kClient:        k3kClient,
		limitrangeclient: limitrangeclient,
	}
}

func (r *DeleteResource) HandleDeletion(ctx context.Context, sa *corev1.ServiceAccount, k3kUser *k3ktypes.K3kUser) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// If finalizer is present, delete external resources
	if controllerutil.ContainsFinalizer(sa, k3ktypes.K3kFinalizerName) {
		// Delete associated resources
		k3ktypes.DelSaVersion(sa.Name)
		if err := r.deleteAssociatedResources(ctx, k3kUser); err != nil {
			// logger.Error(err, "Failed to delete associated resources")
			slog.Error("Failed to delete associated resources", "error", err)
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}

		// Remove finalizer
		logger.Info("Removing finalizer", "namespace", sa.Namespace, "name", sa.Name)
		controllerutil.RemoveFinalizer(sa, k3ktypes.K3kFinalizerName)
		if err := r.Update(ctx, sa); err != nil {
			slog.Error("Failed to update ServiceAccount", "error", err)
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
	}

	return ctrl.Result{}, nil
}

// deleteAssociatedResources deletes all resources associated with the ServiceAccount
func (r *DeleteResource) deleteAssociatedResources(ctx context.Context, k3kUser *k3ktypes.K3kUser) error {
	logger := log.FromContext(ctx)

	// Delete Pod
	// err := r.Sdk.ClientSet.CoreV1().Pods(k3kUser.Namespace).Delete(ctx, k3kUser.GetAgentName(), metav1.DeleteOptions{})
	err := r.Client.Delete(ctx, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: k3kUser.GetAgentName(), Namespace: k3kUser.Namespace}})
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to delete pod")
		return err
	}

	// Delete Service
	// err = r.Sdk.ClientSet.CoreV1().Services(k3kUser.Namespace).Delete(ctx, k3kUser.GetAgentName(), metav1.DeleteOptions{})
	err = r.Client.Delete(ctx, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: k3kUser.GetAgentName(), Namespace: k3kUser.Namespace}})
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to delete service")
		return err
	}

	err = r.k3kClient.Delete(k3kUser)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to delete k3kUser")
		return err
	}
	err = r.limitrangeclient.Delete(ctx, k3kUser.ServiceAccount)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to delete limitrange")
		return err
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k3kUser.GetClusterServer0PvcName(),
			Namespace: k3kUser.GetK3kNamespace(),
		},
	}

	err = r.Client.Delete(ctx, pvc)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to delete pvc")
		return err
	}

	// consoleId := k3kUser.GetConsoleId()
	// if consoleId != "" {
	// 	err = r.Client.Delete(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "console-" + consoleId + ".w7-config", Namespace: k3kUser.Namespace}})
	// 	if err != nil && !errors.IsNotFound(err) {
	// 		slog.Error("Failed to delete console config secret", "err", err)
	// 		logger.Error(err, "Failed to delete console config secret")
	// 		return err
	// 	}
	// }
	err = r.Client.Delete(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: k3kUser.Name + ".w7-config", Namespace: k3kUser.Namespace}})
	if err != nil && !errors.IsNotFound(err) {
		slog.Error("Failed to delete config secret", "err", err)
		logger.Error(err, "Failed to delete config secret")
		return err
	}

	err = r.k3kClient.DeleteNamespace(k3kUser)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Failed to delete namespace")
			return nil
		}
		//namespace 删除失败，则不重复执行
		return nil
	}

	return nil
}

func (r *DeleteResource) HandleResourceRecycleStatus(ctx context.Context, sa *corev1.ServiceAccount, k3kUser *k3ktypes.K3kUser) error {
	logger := log.FromContext(ctx)
	currentStatus := k3kUser.GetResourceStatus()

	// 检查是否需要更新状态
	switch currentStatus {
	case k3ktypes.K3K_STATUS_USER_CREATING, k3ktypes.K3K_STATUS_USER_READY:
		// 如果用户已过期，更新为待回收状态
		_, err := controllerutil.CreateOrPatch(ctx, r.Client, sa, func() error {
			if k3kUser.IsExpired() {
				k3kUser.SetResourceStatus(k3ktypes.K3K_STATUS_USER_WAIT)
			}
			if k3kUser.IsPause() {
				k3kUser.UnPause()
			}
			if k3kUser.HasPendingRecycleTime() {
				k3kUser.DelPendingRecycleTime()
			}
			return nil
		})

		if err != nil {
			slog.Error("K3K_STATUS_USER_READY Failed to update ServiceAccount status", "error", err)
		}

	case k3ktypes.K3K_STATUS_USER_WAIT:

		_, err := controllerutil.CreateOrPatch(ctx, r.Client, sa, func() error {
			if !k3kUser.HasPendingRecycleTime() {
				k3kUser.SetPendingRecycleTime()
			}
			if !k3kUser.IsPause() {
				k3kUser.Pause()
			}
			if k3kUser.IsPendingRecycleExpired() {
				logger.Info("Pending recycle period expired, starting resource recycling", "name", sa.Name)
				k3kUser.SetResourceStatus(k3ktypes.K3K_STATUS_USER_RECYCLE)
			}
			return nil
		})

		if err != nil {
			slog.Error("K3K_STATUS_USER_WAIT Failed to update ServiceAccount status", "error", err)
			return err
		}
		if k3kUser.IsPendingRecycleExpired() {
			if err := r.deleteAssociatedResources(ctx, k3kUser); err != nil {
				slog.Error("Failed to delete associated resources", "error", err)
				return err
			}
		}

	case k3ktypes.K3K_STATUS_USER_RECYCLE:
		if err := r.deleteAssociatedResources(ctx, k3kUser); err != nil {
			slog.Error("Failed to delete associated resources", "error", err)
			return err
		}
		// 检查资源回收是否完成
		if r.isResourceRecyclingComplete(ctx, k3kUser) {
			logger.Info("Resource recycling complete", "name", sa.Name)
			k3kUser.SetResourceStatus(k3ktypes.K3K_STATUS_USER_NEW)
			k3kUser.ReNew()
			if err := r.Update(ctx, sa); err != nil {
				return fmt.Errorf("failed to update ServiceAccount status: %v", err)
			}
		}
	}

	return nil
}

func (r *DeleteResource) isResourceRecyclingComplete(ctx context.Context, k3kUser *k3ktypes.K3kUser) bool {
	// 检查相关资源是否已被删除
	// 例如检查Pod、Service、k3k集群等是否还存在

	// 检查Pod是否存在
	pod := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{Namespace: k3kUser.Namespace, Name: k3kUser.GetAgentName()}, pod)
	if err == nil {
		// Pod仍然存在，回收未完成
		return false
	}

	if !errors.IsNotFound(err) {
		// 发生了其他错误，保守起见认为回收未完成
		return false
	}

	// 检查Service是否存在
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Namespace: k3kUser.Namespace, Name: k3kUser.GetAgentName()}, service)
	if err == nil || !errors.IsNotFound(err) {
		return false
	}

	// 检查k3k集群是否存在的逻辑可能需要根据实际情况调整
	// 这里假设如果Delete操作返回NotFound错误，则表示集群已不存在
	err = r.k3kClient.Delete(k3kUser)
	if err == nil || !errors.IsNotFound(err) {
		return false
	}

	// 所有资源都已删除，回收完成
	return true
}
