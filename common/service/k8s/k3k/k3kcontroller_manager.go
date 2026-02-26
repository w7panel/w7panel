package k3k

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// K3kServiceAccountController reconciles ServiceAccount objects

// SetupWithManager sets up the controllers with the Manager
func SetupK3kControllers(mgr ctrl.Manager) error {
	// mgr.GetControllerOptions().
	sdk := k8s.NewK8sClient().GetSdk()
	sigClient, err := sdk.ToSigClient()
	if err != nil {
		return err
	}
	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		// Manager 已启动，缓存已就绪，可以安全使用客户端
		var saList = &v1.ServiceAccountList{}
		err = sigClient.List(context.Background(), saList, client.InNamespace("default"))
		if err != nil {
			slog.Error("failed to list service account", "error", err)
			return err
		}
		for _, sa := range saList.Items {
			saVersion, ok := sa.Annotations[types.K3K_LOCK_VERSION]
			if !ok {
				saVersion = "1"
			}
			types.SetSaVersion(sa.Name, saVersion)
		}
		var policyList = &v1alpha1.VirtualClusterPolicyList{}
		err = sigClient.List(context.Background(), saList)
		if err != nil {
			slog.Error("failed to list virtual cluster policy", "error", err)
			return err
		}
		for _, policy := range policyList.Items {
			policyVersion, ok := policy.Annotations[types.K3K_LOCK_VERSION]
			if !ok {
				policyVersion = "1"
			}
			types.SetPolicyVersion(policy.Name, policyVersion)
		}
		return nil
	}))

	if err := setupServiceAccountController(mgr, sdk); err != nil {
		slog.Error("failed to setup service account controller", "error", err)
		return err
	}
	// if err := setupConfigController(mgr, sdk); err != nil {
	// 	slog.Error("failed to setup config controller", "error", err)
	// 	return err
	// }
	// if err := setupPolicyController(mgr, sdk); err != nil {
	// 	slog.Error("failed to setup virtual policy controller", "error", err)
	// 	return err
	// }

	// if err := setupPodController(mgr, sdk); err != nil {
	// 	return err
	// }

	if err := setupJobController(mgr, sdk); err != nil {
		slog.Error("failed to setup job controller", "error", err)
		return err
	}

	return nil
}

// Reconcile for ServiceAccount controller
