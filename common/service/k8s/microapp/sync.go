package microapp

import (
	"errors"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	microapp "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/microapp/v1alpha1"
	"github.com/samber/lo"
)

func Sync(k3kName, k3kNs string) error {
	rootSdk := k8s.NewK8sClient().Sdk
	rootList, err := loadMicroAppList(rootSdk)
	sa, err := rootSdk.GetServiceAccount("default", k3kName)
	if err != nil {
		return err
	}
	currentRole, ok := sa.Annotations["w7.cc/role"]
	if !ok {
		return errors.New("sync microapp role is empty")
	}
	k3kConfig := k8s.NewK3kConfig(k3kName, k3kNs, helper.GetApiServerHost(k3kNs))
	root := k8s.NewK8sClient()
	clientsdk, err := root.GetK3kClusterSdkByConfig(k3kConfig)
	if err != nil {
		return err
	}
	clientList, err := loadMicroAppList(clientsdk)
	if err != nil {
		return err
	}
	rootItemsKeyBy := lo.KeyBy(rootList.Items, func(item microapp.MicroApp) string {
		return item.Name
	})
	// 已有的更新
	lo.ForEach(rootList.Items, func(item microapp.MicroApp, index int) {
		if item.Labels["role.w7.cc/"+currentRole] == "true" {
			err = patchRootMicroApp(clientsdk, &item, currentRole)

			if err != nil {
				slog.Error("patchMicroApp"+item.Name, "err", err)
			}
		}
	})
	// 删除多余的
	for _, item := range clientList.Items {
		if item.Labels["microapp.w7.cc/from"] == "root" {
			_, has := rootItemsKeyBy[item.Name]
			if !has {
				err = delMicroApp(clientsdk, &item)
				if err != nil {
					slog.Error("delMicroApp"+item.Name, "err", err)
				}
			}
		}
	}
	return nil
}
