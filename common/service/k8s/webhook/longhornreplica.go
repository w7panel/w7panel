package webhook

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/longhorn"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Deployment 资源

// 处理 Ingress 资源
func (m *ResourceMutator) handleLonghornReplica(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 longhornV1beta2 node admission 请求")
	// 解码请求中的 Ingress 资源

	// if req.Operation == "DELETE" {
	// 	return admission.Allowed("")
	// }

	// rp := &longhornV1beta2.Replica{}
	// if err := (m.decoder).Decode(req, node); err != nil {
	// 	return admission.Errored(http.StatusBadRequest, err)
	// }

	defer longhorn.WebHookLonghornReplica()

	return admission.Allowed("")
}
