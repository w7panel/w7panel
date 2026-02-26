package webhook

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/longhorn"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Deployment 资源

// 处理 Ingress 资源
func (m *ResourceMutator) handleConfigmap(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 Ingress admission 请求")
	// 解码请求中的 Ingress 资源

	// if req.Operation == "DELETE" {
	// 	return admission.Allowed("")
	// }

	configmap := &v1.ConfigMap{}
	if err := (m.decoder).Decode(req, configmap); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	defer checkLonghorn(configmap)
	defer checkLogo(configmap)

	return admission.Allowed("")
}

func checkLonghorn(configMap *v1.ConfigMap) {
	time.AfterFunc(1*time.Second, func() {
		if configMap.Namespace == "kube-system" && configMap.Name == "k3k.overselling.config" {
			longhorn.LonghorStoragePercentage(configMap.Data["storage"])
		}
	})
}

func checkLogo(configMap *v1.ConfigMap) {
	time.AfterFunc(1*time.Second, func() {
		k8s.WriteLogo(configMap)
	})
}
