package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	microapp "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/microapp/v1alpha1"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// not use in webhook
func (m *ResourceMutator) handleMicroApp(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 MicroApp admission 请求")

	modified := false
	microApp := &microapp.MicroApp{}
	if err := (m.decoder).Decode(req, microApp); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	modified = SetControllerReference(microApp, m.client)

	if !modified {
		return admission.Allowed("")
	}
	md, err := json.Marshal(microApp)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, md)
}

func SetControllerReference(microApp *microapp.MicroApp, client sigclient.Client) bool {
	if !controllerutil.HasControllerReference(microApp) {
		// return admission.Allowed("")
		group, err := appgroup.GetAppgroup(microApp.Name, microApp.Namespace, client)
		if err != nil {
			return false
		}
		err = controllerutil.SetControllerReference(group, microApp, k8s.GetScheme())
		if err != nil {
			slog.Error("SetControllerReference error", "error", err)
			return false
		}
		return true
	}
	return false
}
