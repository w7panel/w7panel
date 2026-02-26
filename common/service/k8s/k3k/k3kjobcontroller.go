package k3k

import (
	"context"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type K3kJobController struct {
	client.Client
	Scheme *runtime.Scheme
	Sdk    *k8s.Sdk
}

func setupJobController(mgr ctrl.Manager, sdk *k8s.Sdk) error {
	r := &K3kJobController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Sdk:    sdk,
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Job{}).
		Complete(r)
}

// Reconcile for Job controller
func (r *K3kJobController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Job", "namespace", req.Namespace, "name", req.Name)

	// Fetch the Job instance
	job := &batchv1.Job{}
	if err := r.Get(ctx, req.NamespacedName, job); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to get Job")
			return ctrl.Result{}, err
		}
		// Job was deleted
		return ctrl.Result{}, nil
	}

	// Handle Job
	isK3kjob := job.Labels["k3k-job"] == "true"
	if isK3kjob {
		saName := job.Labels["k3k-sa"]
		if saName == "" {
			logger.Info("k3k-sa label is empty")
			return ctrl.Result{}, nil
		}
		sa := &corev1.ServiceAccount{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: job.Namespace, Name: saName}, sa); err != nil {
			logger.Error(err, "Failed to get ServiceAccount")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		if sa.Annotations[k3ktypes.K3K_JOB_NAME] != job.Name {
			return ctrl.Result{}, nil
		}
		// Update job status
		if job.Status.Succeeded > 0 {
			sa.Annotations[k3ktypes.K3K_JOB_STATUS] = k3ktypes.K3K_STATUS_COMPLETE
			// sa.Annotations[K3K_NAME] = saName
			// sa.Annotations[K3K_NAMESPACE] = job.Labels["k3k-cluster"]
			labelVal, ok := sa.Labels[k3ktypes.K3K_CLUSTER_STATUS]
			if ok && (labelVal == k3ktypes.K3K_STATUS_USER_NEW || labelVal == k3ktypes.K3K_STATUS_USER_CREATING) {
				sa.Labels[k3ktypes.K3K_CLUSTER_STATUS] = k3ktypes.K3K_STATUS_USER_READY
			}

		}
		if job.Status.Failed > 0 {
			sa.Annotations[k3ktypes.K3K_JOB_STATUS] = k3ktypes.K3K_STATUS_FAILED
		}
		if err := r.Update(ctx, sa); err != nil {
			logger.Error(err, "Failed to update ServiceAccount")
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
	}

	return ctrl.Result{}, nil
}
