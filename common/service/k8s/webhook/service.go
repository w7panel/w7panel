package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Service 资源
func (m *ResourceMutator) handleService(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 Service admission 请求")
	svc := &v1.Service{}
	if err := (m.decoder).Decode(req, svc); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// 检查是否需要修改
	modified := false

	pointString, ok := os.LookupEnv("SVC_LB_CLASS")
	if ok && svc.Spec.Type == v1.ServiceTypeLoadBalancer {
		if svc.Spec.LoadBalancerClass == nil {
			svc.Spec.LoadBalancerClass = &pointString
			modified = true
		}
	}

	// 如果没有修改，直接返回允许
	if !modified {
		return admission.Allowed("无需修改 Service")
	}

	// 序列化修改后的 Service
	marshaledSvc, err := json.Marshal(svc)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledSvc)
}
