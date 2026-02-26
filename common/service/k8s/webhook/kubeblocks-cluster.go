package webhook

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Deployment 资源

// 处理 Ingress 资源
func (m *ResourceMutator) handleKubeblocksCluster(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 handleKubeblocksCluster admission 请求")
	// 解码请求中的 Ingress 资源

	if req.Operation == "DELETE" {
		// 删除操作，允许通过
		defer deleteGroup(req.Namespace, req.Name)
		return admission.Allowed("")
	}
	return admission.Allowed("")
}

func deleteGroup(namespace, name string) error {
	groupApi, err := appgroup.NewAppGroupApi(k8s.NewK8sClient().Sdk)
	if err != nil {
		slog.Error("failed to get app api", slog.String("error", err.Error()))
		return err
	}
	err = groupApi.DeleteAppGroup(namespace, name)
	if err != nil {
		slog.Error("failed to delete app group", slog.String("error", err.Error()))
	}
	return err
}
