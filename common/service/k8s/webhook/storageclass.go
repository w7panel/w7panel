package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/longhorn"
	storagev1 "k8s.io/api/storage/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Deployment 资源

// 处理 Ingress 资源
func (m *ResourceMutator) handleStorageClass(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 StorageClass  admission 请求")
	// 解码请求中的 Ingress 资源

	// if req.Operation == "DELETE" {
	// 	return admission.Allowed("")
	// }

	sc := &storagev1.StorageClass{}
	if err := (m.decoder).Decode(req, sc); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	change := longhorn.WebHookStorageClass(sc)
	if !change {
		return admission.Allowed("未修改 StorageClass")
	}
	mds, err := json.Marshal(sc)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, mds)
}
