package webhook

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 secret 资源
func (m *ResourceMutator) handleSecret(ctx context.Context, req admission.Request) admission.Response {
	slog.Error("处理 secret admission 请求")

	// if req.Operation == "DELETE" {
	// 	return admission.Allowed("")
	// }
	secret := &v1.Secret{}
	// 判断是否Delete 请求
	delete := false
	if req.Operation == "DELETE" {
		delete = true
		if err := (m.decoder).DecodeRaw(req.OldObject, secret); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	} else {
		if err := (m.decoder).Decode(req, secret); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	}

	if helper.IsK3kVirtual() {
		defer k3k.SyncHttpAfter(secret, "sync-secret") // 同步到主集群
	}

	if !delete {
		if !helper.IsChildAgent() {
			time.AfterFunc(time.Second*10, func() {
				k3k.SyncToChildSecret(secret.DeepCopy()) // 同步到子集群
			})
		}
	}
	return admission.Allowed("处理 secret 请求")
}
