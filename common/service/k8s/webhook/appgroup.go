package webhook

import (
	"context"
	"log/slog"
	"net/http"

	v1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// not use in webhook
func (m *ResourceMutator) handleAppGroup(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 MicroApp admission 请求")

	microApp := &v1alpha1.AppGroup{}
	if err := (m.decoder).Decode(req, microApp); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.Allowed("")
}
