package appgroup

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OldUpgrade struct {
	sdk      *k8s.Sdk
	groupApi *AppGroupApi
}

func NewOldUpgrade(sdk *k8s.Sdk) (*OldUpgrade, error) {
	groupApi, err := NewAppGroupApi(sdk)
	if err != nil {
		return nil, err
	}
	return &OldUpgrade{
		sdk:      sdk,
		groupApi: groupApi,
	}, nil
}

func (o *OldUpgrade) Upgrade() error {
	ingresses, err := o.sdk.ClientSet.NetworkingV1().Ingresses("default").List(o.sdk.Ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	slog.Info("ingress count", "count", len(ingresses.Items))
	for _, ingress := range ingresses.Items {
		appName, ok := ingress.Labels["app"]
		if !ok {
			slog.Info("ingress has not app label", "appName", appName)
			continue
		}
		_, ok2 := ingress.Labels["group"]
		if ok2 {
			slog.Info("ingress has group label", "appName", appName)
			continue
		}
		deployment, err := o.sdk.ClientSet.AppsV1().Deployments("default").Get(o.sdk.Ctx, appName, metav1.GetOptions{})
		if err != nil {
			slog.Error("get deployment error", "error", err, "deployment", appName)
			continue
		}
		groupName := o.GetAppGroupName(deployment)
		if groupName == "" {
			slog.Info("deployment has not release-name label", "appName", appName)
			continue
		}
		ingress.Labels["group"] = groupName
		//patch add group label
		slog.Info("patch ingress", "appName", appName)

		// patch := metav1.Patch{
		// 	Type:     metav1.PatchTypeJSONPatch,
		// 	Data:     []byte(`[{"op": "add", "path": "/metadata/labels/group", "value":"` + groupName + `"}]`),
		// }
		// _, err = o.sdk.ClientSet.NetworkingV1().Ingresses("default").Patch(o.sdk.Ctx, ingress.Name, metav1.ApplyPatchType, patch)

		_, err = o.sdk.ClientSet.NetworkingV1().Ingresses("default").Update(o.sdk.Ctx, &ingress, metav1.UpdateOptions{})
		if err != nil {
			slog.Error("update ingress error", "error", err, "appName", appName)
			continue
		}
		group, err := o.groupApi.GetAppGroup(deployment.Namespace, groupName)
		if err != nil {
			slog.Error("get appgroup error", "error", err, "appName", appName)
			continue
		}
		isHelmSync, ok := group.Annotations["w7.cc/helm-sync"] //同步旧版注解
		if ok && isHelmSync == "true" {
			title, ok := deployment.Annotations["w7.cc/title"]
			if ok {
				group.Spec.Title = title
			}
			icon, ok := deployment.Annotations["w7.cc/icon"]
			if ok {
				group.Spec.Logo = icon
			}
			group.Annotations = deployment.Annotations
			_, err = o.groupApi.UpdateAppGroup(deployment.Namespace, group)
			if err != nil {
				slog.Error("update appgroup error", "error", err, "appName", appName)
			}
		}

	}
	return nil
}

func (o *OldUpgrade) GetAppGroupName(d *appsv1.Deployment) string {
	labels := d.Labels
	if labels == nil {
		return ""
	}
	_, ok2 := labels["w7.cc/release-name"]
	if ok2 {
		return labels["w7.cc/release-name"]
	}
	_, ok1 := labels["app.kubernetes.io/instance"]
	if ok1 {
		return labels["app.kubernetes.io/instance"]
	}
	_, ok := labels["w7.cc/suffix"]
	if ok {
		return labels["w7.cc/suffix"]
	}

	return d.Name
	//isHelm
}
