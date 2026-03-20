package webhook

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/w7panel/w7panel/common/service/k8s/longhorn"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Service 资源
func (m *ResourceMutator) handlePvc(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 Service admission 请求")

	if req.Operation != "UPDATE" {
		return admission.Allowed("无需修改 PersistentVolumeClaim")
	}
	pvc := &v1.PersistentVolumeClaim{}
	if err := (m.decoder).Decode(req, pvc); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	oldPvc := &v1.PersistentVolumeClaim{}
	if err := (m.decoder).DecodeRaw(req.OldObject, oldPvc); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	sizeChange := false
	if (oldPvc.Spec.Resources.Requests.Storage().Value() != pvc.Spec.Resources.Requests.Storage().Value()) ||
		(oldPvc.Spec.StorageClassName != pvc.Spec.StorageClassName) {
		sizeChange = true

	}
	if sizeChange {
		longhorn.WebHookOnPvcSizeChange(oldPvc.DeepCopy())
	}

	return admission.Allowed("修改了 PVC 资源")
}
