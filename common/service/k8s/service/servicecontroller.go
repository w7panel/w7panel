package service

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceController reconciles Service objects
type ServiceController struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewServiceController creates a new ServiceController
func NewServiceController(mgr ctrl.Manager) error {
	r := &ServiceController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}

// Reconcile handles Service events
func (r *ServiceController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// logger := log.FromContext(ctx)
	// logger.Info("Reconciling Service", "namespace", req.Namespace, "name", req.Name)

	// Fetch the Service instance
	svc := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {
		// Handle deletion
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	isAllowProxy, ok := svc.Labels["allow-proxy"]
	if ok && isAllowProxy == "true" {
		allowSvcProxyMap.appendAllowProxy(req.Namespace, req.Name)
	} else {
		allowSvcProxyMap.removeAllowProxy(req.Namespace, req.Name)
	}

	// Requeue after 1 minute to continue monitoring
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager
func SvcSetupManager(mgr ctrl.Manager) error {
	return NewServiceController(mgr)
}
