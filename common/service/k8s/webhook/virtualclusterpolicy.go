package webhook

import (
	"context"
	"log/slog"
	"net/http"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (m *ResourceMutator) handleVirtualClusterPolicy(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 handleVirtualClusterPolicy admission xxxx 请求")

	// if true {
	// 	return admission.Allowed("VirtualClusterPolicy")
	// }
	// return admission.Denied("VirtualClusterPolicy")
	k3kpolicy := &v1alpha1.VirtualClusterPolicy{}
	if err := (m.decoder).Decode(req, k3kpolicy); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	defer func() {
		slog.Info("发布到商店", "name", k3kpolicy.Name)
		if err := k3k.CheckPublish(ctx, m.client, k3kpolicy); err != nil {
			slog.Error("发布到商店失败", "error", err)
		}
	}()

	return admission.Allowed("VirtualClusterPolicy")

	// userName := req.AdmissionRequest.UserInfo.Username
	// if !strings.HasPrefix(userName, "system:serviceaccount:") {
	// 	return admission.Denied("非系统服务账号，无法创建用户组")
	// }
	// spliteNames := strings.Split(userName, ":")
	// if len(spliteNames) < 4 {
	// 	return admission.Denied("非系统服务账号，无法创建用户组")
	// }
	// saName := spliteNames[3]
	// if !config.IsLicenseVerify(saName) {
	// 	return admission.Denied("未通过授权验证, 无法创建用户组")
	// }
	// k3kTypes.SetPolicyVersion(k3kpolicy.Name, k3kpolicy.Annotations[k3kTypes.K3K_LOCK_VERSION])

	// // 将 VirtualClusterPolicy 的 annotations 缓存到全局变量
	// policyCacheLock.Lock()
	// if k3kpolicy.Annotations != nil {
	// 	policyCache[k3kpolicy.Name] = k3kpolicy.Annotations
	// 	slog.Info("缓存 VirtualClusterPolicy 信息", slog.String("name", k3kpolicy.Name))
	// }
	// policyCacheLock.Unlock()

	// modified := false
	// // 检查是否需要修改
	// if k3kpolicy.Spec.AllowedMode == "virtual" && config.IsTeam(saName) {
	// 	k3kpolicy.Spec.AllowedMode = "shared"
	// 	modified = true
	// }
	// if !modified {
	// 	return admission.Allowed("VirtualClusterPolicy")
	// }
	// marshaled, err := json.Marshal(k3kpolicy)
	// if err != nil {
	// 	return admission.Errored(http.StatusInternalServerError, err)
	// }

	// return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
