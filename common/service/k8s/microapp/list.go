package microapp

import (
	"errors"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	microapp "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/microapp/v1alpha1"
	"github.com/samber/lo"
	sig "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// 头部显示
func ListTop(t string) (*microapp.MicroAppList, error) {
	token := k8s.NewK8sToken(t)
	role := token.GetRole()
	if role == "" {
		return nil, errors.New("role is empty")
	}

	clientSdk, err := k8s.NewK8sClient().Channel(t)
	if err != nil {
		return nil, err
	}
	newList := &microapp.MicroAppList{}
	currentList, err := loadMicroAppList(clientSdk)
	if err != nil {
		return nil, err
	}

	lo.ForEach(currentList.Items, func(item microapp.MicroApp, index int) {
		if item.Labels != nil {
			if item.RoleCount() > 1 || item.Labels["microapp.w7.cc/from"] == "root" {
				newList.Items = append(newList.Items, item)
			}

		}
	})

	return newList, nil
}

func loadMicroAppList(sdk *k8s.Sdk) (*microapp.MicroAppList, error) {
	list := &microapp.MicroAppList{}
	sigClient, err := sdk.ToSigClient()
	if err != nil {

		return nil, err
	}
	err = sigClient.List(sdk.Ctx, list, &sig.ListOptions{})
	if err != nil {
		slog.Error("loadMicroAppList", "err", err)
		return nil, err
	}
	return list, nil
}

func patchRootMicroApp(sdk *k8s.Sdk, origin *microapp.MicroApp, role string) error {
	sigclient, err := sdk.ToSigClient()
	if err != nil {
		slog.Error("createMicroApp", "err", err)
		return err
	}
	item := origin.DeepCopy()
	// itemCopy := item.DeepCopy()
	_, err = controllerutil.CreateOrUpdate(sdk.Ctx, sigclient, item, func() error {
		item.Labels["microapp.w7.cc/from"] = "root"
		item.SetResourceVersion("")
		item.SetUID("")
		// 移除不属于当前角色的权限配置信息
		item.Spec.Bindings = lo.Filter(item.Spec.Bindings, func(bindings microapp.Bindings, index int) bool {
			return bindings.Name == role
		})
		newRole := item.Spec.ConfigV2.Props.RoleConfig[role]
		item.Spec.ConfigV2.Props.RoleConfig = map[string]microapp.Role{}
		item.Spec.ConfigV2.Props.RoleConfig[role] = newRole
		return nil
	})
	return err
}

func delMicroApp(sdk *k8s.Sdk, item *microapp.MicroApp) error {
	sigclient, err := sdk.ToSigClient()
	if err != nil {
		slog.Error("createMicroApp", "err", err)
		return err
	}
	return sigclient.Delete(sdk.Ctx, item)
}
