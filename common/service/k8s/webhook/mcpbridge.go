package webhook

import (
	"context"
	"log/slog"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Deployment 资源

// 处理 Ingress 资源
func (m *ResourceMutator) handleMcpBridge(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 mcpBridge admission 请求")
	// 解码请求中的 Ingress 资源

	defer time.AfterFunc(time.Second*5, func() {
		// 不能用ctx 会被取消掉
		go higress.WebhookMcpbridge(context.TODO(), m.client, req.Name, req.Namespace)
	})

	return admission.Allowed("")
}
