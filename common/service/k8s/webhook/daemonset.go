package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 StatefulSet 资源
func (m *ResourceMutator) handleDaemonset(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 Daemonset admission 请求")

	modified := false
	// 解码请求中的 StatefulSet 资源
	ds := &appsv1.DaemonSet{}
	if err := (m.decoder).Decode(req, ds); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	ResetImage(ds.Namespace, ds.Name, "DaemonSet", ds.Annotations)

	clusterName, ok := ds.Labels["cluster"]
	if !ok {
		return admission.Allowed("无需修改 statefulset")
	}

	rs, err := getResourceLimit(m.client, m.sdk, clusterName, ds.Labels["type"])
	if err != nil {
		slog.Error("not found resource limit")
	}
	for i := range ds.Spec.Template.Spec.Containers {
		container := &ds.Spec.Template.Spec.Containers[i]
		if rs != nil {
			container.Resources.Limits = rs
			container.Resources.Requests = rs
			modified = true
		}
	}

	// 如果没有修改，直接返回允许
	if !modified {
		return admission.Allowed("StatefulSet 已经挂载了 registries.yaml 或不需要修改")
	}

	// 序列化修改后的 StatefulSet
	mds, err := json.Marshal(ds)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// 返回修改后的资源
	return admission.PatchResponseFromRaw(req.Object.Raw, mds)
}
