package webhook

import (
	"context"
	"log/slog"
	"net/http"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress/client/pkg/apis/extensions/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Deployment 资源

// 处理 Ingress 资源
func (m *ResourceMutator) handleWasmPlugin(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 mcpBridge admission 请求")
	// 解码请求中的 Ingress 资源
	wasmPlugin := &v1alpha1.WasmPlugin{}
	if err := (m.decoder).Decode(req, wasmPlugin); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	higress.WebhookWasmPlugin(wasmPlugin)
	return admission.Allowed("")
}
