package higress

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	higressapinetworkingv1 "gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress/api/networking/v1"
	higressnetworkingv1 "gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress/client/pkg/apis/networking/v1"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WebhookMcpbridge(ctx context.Context, r client.Client, reqName, reqNamespace string) error {
	slog.Info("reconciling McpBridge", "name", reqName, "namespace", reqNamespace)

	// 如果是default McpBridge，则跳过处理
	if reqName == "default" {
		return nil
	}

	if helper.IsK3kShared() {
		err := k3k.SyncMcpBridgeHttp(k3k.NewSyncObject(reqName, reqNamespace))
		if err != nil {
			slog.Error("failed to sync mcp bridge http", "err", err)
			return err
		}
		return nil
	}

	// 获取所有非default的McpBridge资源
	var mcpBridgeList higressnetworkingv1.McpBridgeList
	if err := r.List(ctx, &mcpBridgeList); err != nil {
		slog.Error("failed to list mcp bridges", "error", err)
		return err
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
			return err
		}
	} else {
		if err := r.Create(ctx, newMap); err != nil {
			slog.Error("failed to create mcp bridge", "error", err)
			return err
		}
	}

	return nil
}
