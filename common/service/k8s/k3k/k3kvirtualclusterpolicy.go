package k3k

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K3kVirtualClusterPolicyController struct {
	client.Client
	Scheme *runtime.Scheme
	Sdk    *k8s.Sdk
}

func setupPolicyController(mgr ctrl.Manager, sdk *k8s.Sdk) error {
	r := &K3kVirtualClusterPolicyController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Sdk:    sdk,
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VirtualClusterPolicy{}).
		Complete(r)
}

// 使用webhook 处理
// Reconcile for Job controller
func (r *K3kVirtualClusterPolicyController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func PublishToShop(ctx context.Context, client client.Client, k3kpolicy *v1alpha1.VirtualClusterPolicy) error {
	if console.GetCurrentLicense() == nil {
		slog.Error("no license to publish")
		return nil
	}
	slog.Error("publish to shop")

	group := k3ktypes.NewK3kClusterPolicy(k3kpolicy)
	if !group.CanPublish() {
		// return nil
	}
	if k3kpolicy.Annotations["w7.cc/cost-name"] != "" {
		configmap := &corev1.ConfigMap{}
		err := client.Get(ctx, types.NamespacedName{Namespace: "default", Name: k3kpolicy.Annotations["w7.cc/cost-name"]}, configmap)
		if err != nil {
			return err
		}
		cost, err := k3ktypes.ConfigMapToCost(configmap)
		if err != nil {
			return err
		}
		json, err := cost.ToJsonString()
		if err != nil {
			return err
		}
		k3kpolicy.Annotations["w7.cc/cost"] = json
	}

	urlValues, err := group.ToPublishShopParams2(k3kpolicy.Name)
	if err != nil {
		return err
	}

	if config.MainW7Config == nil {
		slog.Error("no main config")
		return nil
	}
	city := config.CurrentCity
	if ncity, ok := k3kpolicy.Annotations["city"]; ok {
		city = ncity
	}
	// urlValues["description"] = k3kpolicy.Annotations["description"]
	urlValues["clusterid"] = config.MainW7Config.ClusterId
	urlValues["city"] = city
	urlValues["clusterurl"] = config.MainW7Config.OfflineUrl

	sdkClient, err := console.NewSdkClient(console.GetCurrentLicense())
	if err != nil {
		return err
	}
	return sdkClient.PublishPanelResource2(urlValues)

	// consolecdClient := console.NewConsoleCdClient(config.MainW7Config.ThirdpartyCDToken)
	// return consolecdClient.PublishPanelResource(urlValues)

}

func DeleteFromShop(k3kpolicy *v1alpha1.VirtualClusterPolicy) error {
	if console.GetCurrentLicense() == nil {
		slog.Error("no license")
		return nil
	}
	slog.Error("delete from shop")
	if config.MainW7Config == nil {
		slog.Error("no main config")
		return nil
	}
	// consolecdClient := console.NewConsoleCdClient(config.MainW7Config.ThirdpartyCDToken)
	data := map[string]string{
		"groupname": k3kpolicy.Name,
	}
	// return consolecdClient.DeletePanelResource(data)
	sdkClient, err := console.NewSdkClient(console.GetCurrentLicense())
	if err != nil {
		return err
	}
	return sdkClient.DeletePanelResource2(data)

}

func CheckPublish(ctx context.Context, r client.Client, k3kpolicy *v1alpha1.VirtualClusterPolicy) error {

	// cfg := &corev1.ConfigMap{}
	// if err := r.Get(ctx, types.NamespacedName{Namespace: "kube-system", Name: "k3k.config"}, cfg); err != nil {
	// 	if errors.IsNotFound(err) {
	// 		slog.Error("configmap not found")
	// 		return err
	// 	}
	// 	slog.Error("failed to get configmap")
	// 	return err
	// }
	// if cfg.Data["showInShop"] == "true" && k3kpolicy.Labels["w7.cc/showInShop"] == "true" {
	if k3kpolicy.Labels["w7.cc/showInShop"] == "true" {
		err := PublishToShop(ctx, r, k3kpolicy)
		if err != nil {
			slog.Error("failed to publish to shop ")
		}
		return nil
	}
	if k3kpolicy.Labels["w7.cc/showInShop"] != "true" {
		err := DeleteFromShop(k3kpolicy)
		if err != nil {
			slog.Error("failed to delete ")
		}
	}
	return nil
}
