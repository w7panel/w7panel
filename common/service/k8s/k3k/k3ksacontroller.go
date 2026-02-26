package k3k

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/sa"
	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"github.com/rancher/k3k/k3k-kubelet/translate"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func setupServiceAccountController(mgr ctrl.Manager, sdk *k8s.Sdk) error {

	client := mgr.GetClient()
	k3kClient := k3ktypes.NewK3kClient(client)
	limitClient := sa.NewLimitRangeClient(client)
	storage := sa.NewStorage(client)
	r := &K3kServiceAccountController{
		Client:      client,
		Scheme:      mgr.GetScheme(),
		k3kClient:   k3kClient,
		rolebinding: sa.NewRoleBinding(client),
		deleteRc:    sa.NewDeleteResource(client, k3kClient, limitClient),
		limitClient: limitClient,
		storage:     storage,
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ServiceAccount{}).
		Complete(r)
}

// K3kServiceAccountController reconciles ServiceAccount objects
type K3kServiceAccountController struct {
	client.Client
	Scheme      *runtime.Scheme
	k3kClient   *k3ktypes.K3kClient
	rolebinding *sa.RoleBinding
	deleteRc    *sa.DeleteResource
	limitClient *sa.Limitrangeclient
	storage     *sa.Storage
}

func (r *K3kServiceAccountController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	result, err := r.reconcile0(ctx, req)
	if err != nil {
		stack := debug.Stack()

		// 结构化日志记录
		slog.Error("详细错误信息",
			"error_message", err.Error(),
			"stack_trace", string(stack),
			"error_type", fmt.Sprintf("%T", err))
		// slog.Error("result", "err", err, "result", result)
	}
	return result, err

}

// Reconcile for ServiceAccount controller
func (r *K3kServiceAccountController) reconcile0(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic in Handle", "panic", r)
		}
	}()
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ServiceAccount", "namespace", req.Namespace, "name", req.Name)
	slog.Error("start sa", "uname", req.Name)
	// Fetch the ServiceAccount instance
	sa := &corev1.ServiceAccount{}
	if err := r.Get(ctx, req.NamespacedName, sa); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to get ServiceAccount")
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
		// ServiceAccount was deleted and we don't have finalizer, nothing to do
		return ctrl.Result{}, nil
	}
	// if sa.Name != "hello" {
	// 	return ctrl.Result{}, nil
	// }
	k3kUser := k3ktypes.NewK3kUser(sa)
	// Check if the ServiceAccount is being deleted
	if !sa.DeletionTimestamp.IsZero() {
		logger.Info("ServiceAccount is being deleted", "namespace", req.Namespace, "name", req.Name)
		slog.Debug("sa is deleting", "uname", req.Name)
		return r.deleteRc.HandleDeletion(ctx, sa, k3kUser)
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(sa, k3ktypes.K3kFinalizerName) {
		logger.Info("Adding finalizer", "namespace", req.Namespace, "name", req.Name)
		controllerutil.AddFinalizer(sa, k3ktypes.K3kFinalizerName)
		if err := r.Update(ctx, sa); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
		// Requeue to continue processing after finalizer is added
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}
	k3ktypes.SetSaVersion(sa.Name, sa.Annotations[k3ktypes.K3K_LOCK_VERSION])

	if k3kUser.IsNormalUser() {
		err := r.rolebinding.CreateNormalUserRoleBinding(ctx, sa, helper.ServiceAccountName())
		if err != nil {
			logger.Error(err, "Failed to create normal user role binding")
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
		return ctrl.Result{}, nil
	}
	if !k3kUser.IsClusterUser() {
		// Not our ServiceAccount, ignore it
		return ctrl.Result{}, nil
	}
	// 创建角色 需要job 查看权限

	// 处理资源回收阶段
	if err := r.deleteRc.HandleResourceRecycleStatus(ctx, sa, k3kUser); err != nil {
		logger.Error(err, "Failed to handle resource recycle status")
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: k3kUser.GetK3kNamespace(),
			Labels: map[string]string{
				"policy.k3k.io/policy-name": k3kUser.GetClusterPolicy(),
			},
		},
	}
	_, err := controllerutil.CreateOrPatch(ctx, r.Client, namespace, func() error {
		namespace.Labels = map[string]string{
			"policy.k3k.io/policy-name": k3kUser.GetClusterPolicy(),
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to create namespace")
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	err = r.rolebinding.CreateRole(ctx, sa, k3kUser.GetK3kNamespace())
	if err != nil {
		logger.Error(err, "Failed to create role")
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	// 创建 registries ConfigMap
	// err = r.createRegistriesConfigMap(ctx, k3kUser)
	// if err != nil {
	// 	logger.Error(err, "Failed to create registries ConfigMap")
	// 	return ctrl.Result{RequeueAfter: time.Second * 10}, err
	// }\
	err = r.limitClient.Delete(ctx, sa)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to handle limit range")
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
		// return ctrl.Result{}, nil
	}

	if !k3kUser.IsClusterReady() {
		slog.Error("cluster not ready", "uname", k3kUser.GetName())
		return ctrl.Result{}, nil
	}

	err = r.limitClient.Handle(ctx, sa)
	if err != nil {
		logger.Error(err, "Failed to handle limit range")
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	err = r.storage.Handle(ctx, k3kUser)
	if err != nil {
		logger.Error(err, "Failed to handle storage")
		// return ctrl.Result{}, err
	}

	// err = r.createKubeconfig(ctx, k3kUser)
	// if err != nil {
	// 	requeue := true
	// 	if errors.IsNotFound(err) {
	// 		requeue = false
	// 	}
	// 	slog.Error("cr kubeconfig error", "error", err.Error(), "uname", k3kUser.GetName(), "requeue", requeue)
	// 	if !requeue {
	// 		return ctrl.Result{}, err
	// 	}
	// 	return ctrl.Result{RequeueAfter: time.Second * 30}, err
	// }

	err = r.createAgent(ctx, k3kUser)
	if err != nil {
		slog.Error("cr agent error", "err", err, "uname", k3kUser.GetName())
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}

	if k3kUser.HasExpireTime() {
		return ctrl.Result{RequeueAfter: time.Minute * 30}, nil
	}

	return ctrl.Result{}, nil
}

// createRegistriesConfigMap 创建包含镜像仓库配置的 ConfigMap
func (r *K3kServiceAccountController) createRegistriesConfigMap(ctx context.Context, k3kUser *k3ktypes.K3kUser) error {
	logger := log.FromContext(ctx)
	defaultCnf := k3ktypes.K3kregCnf
	trans := translate.ToHostTranslator{
		ClusterName:      k3kUser.GetK3kName(),
		ClusterNamespace: k3kUser.GetK3kNamespace(),
	}
	configMapName := trans.TranslateName("default", "registries")
	// configMapName := "registries"
	// 检查 ConfigMap 是否已存在
	configMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Namespace: k3kUser.GetK3kNamespace(), Name: configMapName}, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			// 创建新的 ConfigMap，只包含 default.cnf
			newConfigMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configMapName,
					Namespace: k3kUser.GetK3kNamespace(),
				},
				Data: map[string]string{
					"default.cnf": defaultCnf,
				},
			}

			if err := r.Create(ctx, newConfigMap); err != nil {
				logger.Error(err, "Failed to create registries ConfigMap")
				return err
			}
			logger.Info("Created registries ConfigMap", "namespace", k3kUser.Namespace)
			return nil
		}
		logger.Error(err, "Failed to get registries ConfigMap")
		return err
	}

	return nil
}

func (r *K3kServiceAccountController) createKubeconfig(ctx context.Context, k3kUser *k3ktypes.K3kUser) error {
	configmap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Namespace: k3kUser.Namespace, Name: k3kUser.GetKubeconfigMapName()}, configmap)
	if err != nil {
		if errors.IsNotFound(err) {

			yaml, err := r.k3kClient.GetKubeConfigYaml(k3kUser)
			if err != nil {
				slog.Warn("failed to get kubeconfig", "err", err)
				return err
			}

			configmap = &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      k3kUser.GetKubeconfigMapName(),
					Namespace: k3kUser.Namespace,
				},
				Data: map[string]string{
					"kubeconfig": string(yaml),
				},
			}
			if err := r.Create(ctx, configmap); err != nil {
				slog.Error("failed to create configmap", "err", err)
				return err
			}
			return nil
		}
		return err
	}

	yaml, err := r.k3kClient.GetKubeConfigYaml(k3kUser)
	if err != nil {
		slog.Warn("failed to get kubeconfig", "err", err)
		return err
	}
	configmap.Data["kubeconfig"] = string(yaml)
	if err := r.Update(ctx, configmap); err != nil {
		slog.Error("failed to update configmap", "err", err)
		return err
	}

	return nil
}

func (r *K3kServiceAccountController) createAgent(ctx context.Context, k3kUser *k3ktypes.K3kUser) error {

	root := k8s.NewK8sClient()
	clientSdk, err := root.GetK3kClusterSdkByConfig(k3kUser.ToK3kConfig())
	if err != nil {
		slog.Warn("failed to get sdk", "err", err)
		return err
	}
	clientSigClient, err := clientSdk.ToSigClient()
	if err != nil {
		slog.Warn("failed to get sigclient", "err", err)
		return err
	}

	agentService := k3ktypes.ToK3kAgentService(k3kUser)
	_, err = controllerutil.CreateOrUpdate(ctx, clientSigClient, agentService, func() error { return nil })
	if err != nil {
		slog.Warn("failed to create agentService", "err", err)
		return err
	}

	ingService := k3ktypes.ToVirtualIngressService(k3kUser)
	clone := ingService.DeepCopy()
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, clone, func() error {
		clone.Spec = ingService.Spec
		return nil
	})
	if err != nil {
		slog.Warn("failed to create ingService", "err", err)
		return err
	}

	// Check if pod exists
	// ds2 := &appsv1.DaemonSet{}
	// err = clientSigClient.Get(ctx, types.NamespacedName{Namespace: k3kUser.Namespace, Name: k3kUser.GetAgentName()}, ds2)
	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		return r.createPod(ctx, clientSigClient, k3kUser)
	// 	}
	// 	return err
	// }

	ds := k3ktypes.ToK3kDaemonSet(k3kUser)
	copy := ds.DeepCopy()
	_, err = controllerutil.CreateOrPatch(ctx, clientSigClient, copy, func() error {
		copy.Annotations = ds.Annotations
		copy.Annotations["root-pod-ip"] = os.Getenv("POD_IP")
		// copy.Labels["d"]
		copy.Spec = ds.Spec
		return nil
	})
	if err != nil {
		slog.Warn("failed to create daemonSet", "err", err)
		return err
	}

	// helmVersion := os.Getenv("HELM_VERSION") //pod.Annotations["helm-version"]
	// podVersion := pod.Annotations["helm-version"]
	// rootPodIp := pod.Annotations["root-pod-ip"]
	// needReCreate := helmVersion != podVersion && helmVersion != "" || rootPodIp != os.Getenv("ROOT_POD_IP")
	// // If pod is in failed state, delete and recreate it
	// if pod.Status.Phase == corev1.PodUnknown || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || needReCreate {
	// 	if err := clientSigClient.Delete(ctx, pod); err != nil {
	// 		slog.Warn("failed to delete pod", "err", err)
	// 		return err
	// 	}
	// 	return r.createPod(ctx, clientSigClient, k3kUser)
	// }
	return nil
}

func (r *K3kServiceAccountController) createPod(ctx context.Context, client client.Client, k3kUser *k3ktypes.K3kUser) error {
	pod := k3ktypes.ToK3kPod(k3kUser)
	if err := client.Create(ctx, pod); err != nil {
		slog.Warn("failed to create agent", "err", err)
		return err
	}
	return nil
}
