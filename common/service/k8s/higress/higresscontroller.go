package higress

import (
	"context"
	"log/slog"
	"os"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	higressapinetworkingv1 "gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress/api/networking/v1"
	higressextv1 "gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress/client/pkg/apis/extensions/v1alpha1"
	higressnetworkingv1 "gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress/client/pkg/apis/networking/v1"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const w7cdnproxy = "w7-cdn-proxy"

func SetupManager(mgr ctrl.Manager, sdk *k8s.Sdk) error {

	reconciler := &McpBridgeReconciler{
		Client: mgr.GetClient(),
		sdk:    sdk,
	}

	// 添加一个启动钩子，在 Manager 启动后执行
	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		// Manager 已启动，缓存已就绪，可以安全使用客户端
		if err := InitW7ProxyPlugin(ctx, mgr.GetScheme(), mgr.GetClient()); err != nil {
			slog.Warn("failed to init w7 proxy plugin", "error", err)
			// 这里我们只记录警告而不返回错误，保持与原代码行为一致
		}
		return nil
	}))

	if err := reconciler.SetupWithManager(mgr); err != nil {
		slog.Error("failed to setup controller", "error", err)
		return err
	}
	return nil
}

// McpBridgeReconciler 处理McpBridge资源的协调
type McpBridgeReconciler struct {
	client.Client
	sdk *k8s.Sdk
}

// Reconcile 处理McpBridge资源的变化
func (r *McpBridgeReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	slog.Info("reconciling McpBridge", "name", req.Name, "namespace", req.Namespace)

	// 如果是default McpBridge，则跳过处理
	if req.Name == "default" {
		return reconcile.Result{}, nil
	}

	if helper.IsK3kShared() {
		error := k3k.SyncMcpBridgeHttp(k3k.NewSyncObject(req.Name, req.Namespace))
		if error != nil {
			slog.Error("failed to sync mcp bridge http", "error", error)
			return reconcile.Result{RequeueAfter: 10 * time.Second}, error
		}
		return reconcile.Result{}, nil
	}

	// 获取所有非default的McpBridge资源
	var mcpBridgeList higressnetworkingv1.McpBridgeList
	if err := r.List(ctx, &mcpBridgeList); err != nil {
		slog.Error("failed to list mcp bridges", "error", err)
		return reconcile.Result{}, err
	}

	// 合并所有非default McpBridge的配置
	var config map[string]*higressapinetworkingv1.RegistryConfig = make(map[string]*higressapinetworkingv1.RegistryConfig)
	for _, bridge := range mcpBridgeList.Items {
		if bridge.Name == "default" {
			continue
		}

		for _, registry := range bridge.Spec.Registries {
			config[registry.Name] = registry
		}
	}
	defaultConfig := []*higressapinetworkingv1.RegistryConfig{}
	var defaultMcp = &higressnetworkingv1.McpBridge{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "higress-system",
		},
		Spec: higressapinetworkingv1.McpBridge{
			Registries: defaultConfig,
		},
	}
	err := r.Get(ctx, types.NamespacedName{Namespace: "higress-system", Name: "default"}, defaultMcp)
	update := true
	if err != nil {
		if apierrors.IsNotFound(err) {
			update = false
		}
	}

	// 将所有配置添加到default McpBridge
	for _, registry := range config {
		defaultConfig = append(defaultConfig, registry)
	}
	newMap := defaultMcp.DeepCopy()
	newMap.Spec.Registries = defaultConfig
	if update {
		if err := r.Update(ctx, newMap); err != nil {
			slog.Error("failed to update mcp bridge", "error", err)
			return reconcile.Result{RequeueAfter: time.Minute}, err
		}
	} else {
		if err := r.Create(ctx, newMap); err != nil {
			slog.Error("failed to create mcp bridge", "error", err)
			return reconcile.Result{RequeueAfter: time.Minute}, err
		}
	}

	return reconcile.Result{}, nil
}

// SetupWithManager 将控制器注册到Manager
func (r *McpBridgeReconciler) SetupWithManager(mgr manager.Manager) error {
	// 创建过滤器，排除名为"default"的McpBridge资源
	defaultFilter := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		return obj.GetName() != "default"
	})

	// 设置控制器
	return ctrl.NewControllerManagedBy(mgr).
		For(&higressnetworkingv1.McpBridge{}).
		WithEventFilter(defaultFilter).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}

// InitW7ProxyPlugin 初始化W7ProxyPlugin
func InitW7ProxyPlugin(ctx context.Context, sheame *runtime.Scheme, client client.Client) error {
	plugin := &higressextv1.WasmPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w7cdnproxy,
			Namespace: "higress-system",
		},
	}
	yamlDir := os.Getenv("KO_DATA_PATH") //facade.GetConfig().GetString("app.static_path")
	yaml, err := os.ReadFile(yamlDir + "/yaml/w7-cdn-proxy.yaml")
	if err != nil {
		slog.Error("failed to read yaml file", "error", err)
		return err
	}
	// // _, err := r.clientSet.ExtensionsV1alpha1().WasmPlugins("higress-system").Get(context.Background(), w7cdnproxy, metav1.GetOptions{})
	// err = r.Get(ctx, types.NamespacedName{Namespace: "higress-system", Name: w7cdnproxy}, plugin)
	// if err != nil && !apierrors.IsNotFound(err) {
	// 	slog.Error("failed to get wasm plugin", "error", err)
	// 	return err
	// }

	decoder := admission.NewDecoder(sheame)
	err = decoder.DecodeRaw(runtime.RawExtension{Raw: yaml, Object: plugin}, plugin)
	if err != nil {
		slog.Error("failed to decode yaml file", "error", err)
		return err
	}
	newPlugin := plugin.DeepCopy()
	_, err = controllerutil.CreateOrPatch(ctx, client, newPlugin, func() error {
		if newPlugin.Annotations == nil {
			newPlugin.Annotations = make(map[string]string)
		}
		if newPlugin.Labels == nil {
			newPlugin.Labels = make(map[string]string)
		}
		newPlugin.Annotations["w7.cc/plugin-url"] = "http://w7panel.default.svc:8000/ui/wasm/plugin-2.0.6.wasm"
		newPlugin.Labels["higress.io/wasm-plugin-name"] = "w7-cdn-proxy"
		newPlugin.Labels["higress.io/wasm-plugin-version"] = "2.0.6"
		newPlugin.Labels["higress.io/wasm-plugin-debug"] = "2.0.6"
		return nil
	})
	if err != nil {
		slog.Error("failed to create or patch wasm plugin", "error", err)
		return err
	}
	return nil

	// 判断是否已经存在，如果不存在则创建
	// if err != nil && strings.Contains(err.Error(), "not found") {

	// 	yamlDir := facade.GetConfig().GetString("app.static_path")
	// 	yaml, err := os.ReadFile(yamlDir + "/yaml/w7-cdn-proxy.yaml")
	// 	if err != nil {
	// 		slog.Error("failed to read yaml file", "error", err)
	// 		return err
	// 	}
	// 	option := k8s.NewApplyOptions(r.sdk.GetNamespace())
	// 	_, err = r.sdk.ApplyYaml(yaml, *option)
	// 	if err != nil {
	// 		slog.Error("failed to apply yaml file", "error", err)
	// 		return err
	// 	}
	// 	return nil
	// }

	// patchData := map[string]interface{}{
	// 	"met

	// return niladata": map[string]interface{}{
	// 		"annotations": map[string]string{
	// 			"w7.cc/plugin-url": "http://w7panel.default.svc:8000/ui/wasm/plugin-2.0.6.wasm",
	// 		},
	// 		"labels": map[string]string{
	// 			"higress.io/wasm-plugin-name":    "w7-cdn-proxy",
	// 			"higress.io/wasm-plugin-version": "2.0.6",
	// 			"higress.io/wasm-plugin-debug":   "2.0.6",
	// 		},
	// 	},
	// }

	// // 将 patchData 转换为 JSON
	// patchBytes, err := json.Marshal(patchData)
	// if err != nil {
	// 	return err
	// }

	// // r.Patch()
	// // 更新注解信息，防止被删除
	// // _, err = r.clientSet.ExtensionsV1alpha1().WasmPlugins("higress-system").Patch(context.Background(), w7cdnproxy, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	// // _, err = r.Patch(context.Background(), w7cdnproxy, types.MergePatchType, patchBytes, metav1.PatchOptions{})

	// err = r.Patch(ctx, plugin, client.RawPatch(types.MergePatchType, patchBytes))
	// if err != nil {
	// 	slog.Error("failed to patch wasm plugin", "error", err)
	// 	return err
	// }
}
