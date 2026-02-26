package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/longhorn"
	longhornV1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Deployment 资源

// 处理 Ingress 资源
func (m *ResourceMutator) handleLonghornNode(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 longhornV1beta2 node admission 请求")
	// 解码请求中的 Ingress 资源

	// if req.Operation == "DELETE" {
	// 	return admission.Allowed("")
	// }

	node := &longhornV1beta2.Node{}
	if err := (m.decoder).Decode(req, node); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if (node).Spec.Disks == nil {
		return admission.Allowed("")
	}
	change := false
	for name, disk := range node.Spec.Disks {
		if disk.Tags == nil {
			disk.Tags = []string{}
		}
		if len(disk.Tags) == 0 {
			disk.Tags = append(disk.Tags, name)
			node.Spec.Disks[name] = disk
			change = true
		}
	}
	defer longhorn.WebhookLonghornNode(node)
	if !change {
		return admission.Allowed("")
	}
	mds, err := json.Marshal(node)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, mds)
}
