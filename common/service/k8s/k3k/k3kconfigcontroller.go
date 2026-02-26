package k3k

import (
	"context"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K3kConfigController struct {
	client.Client
	Scheme *runtime.Scheme
	Sdk    *k8s.Sdk
}

func setupConfigController(mgr ctrl.Manager, sdk *k8s.Sdk) error {
	r := &K3kConfigController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Sdk:    sdk,
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}

// Reconcile for Job controller
func (r *K3kConfigController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	if req.Namespace != "kube-system" && req.Name != "k3k.config" {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}
