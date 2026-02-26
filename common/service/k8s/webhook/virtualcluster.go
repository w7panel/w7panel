package webhook

import (
	"context"
	"net/http"

	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (m *ResourceMutator) handleK3kCluster(ctx context.Context, req admission.Request) admission.Response {
	cluster := &v1alpha1.Cluster{}
	if err := (m.decoder).Decode(req, cluster); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if cluster.Spec.Addons == nil {
		cluster.Spec.Addons = []v1alpha1.Addon{}
	}
	// has := false
	// for i := range cluster.Spec.Addons {
	// 	addon := &cluster.Spec.Addons[i]
	// 	if addon.SecretRef == "k3k-virtual" {
	// 		has = true
	// 	}
	// }
	// if !has && req.Operation != "DELETE" {
	// 	cluster.Spec.Addons = append(cluster.Spec.Addons, v1alpha1.Addon{
	// 		SecretRef:       "k3k-virtual",
	// 		SecretNamespace: "k3k-system",
	// 	})
	// 	mds, err := json.Marshal(cluster)
	// 	if err != nil {
	// 		return admission.Errored(http.StatusInternalServerError, err)
	// 	}

	// 	return admission.PatchResponseFromRaw(req.Object.Raw, mds)
	// }

	return admission.Allowed("")
}
