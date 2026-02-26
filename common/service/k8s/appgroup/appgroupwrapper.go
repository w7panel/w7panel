package appgroup

import (
	v1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"

	"github.com/samber/lo"
)

type appgroupWrapper struct {
	*v1alpha1.AppGroup
	changed       bool
	exists        bool
	deleteEnabled bool
	parent        *appgroupWrapper
}

func NewAppGroupWrapper(ag *v1alpha1.AppGroup, exists bool) *appgroupWrapper {
	return &appgroupWrapper{
		AppGroup: ag,
		exists:   exists,
	}
}

func (g *appgroupWrapper) SetParent(parent *appgroupWrapper) {
	g.parent = parent
}

func (g *appgroupWrapper) AddStatusItem(item v1alpha1.AppGroupItemStatus) {
	if g.Status.Items == nil {
		g.Status.Items = []v1alpha1.AppGroupItemStatus{}
	}
	if item.Kind != "Job" {
		contains := lo.ContainsBy(g.Status.Items, func(i v1alpha1.AppGroupItemStatus) bool {
			return i.Name == item.Name && i.Kind == item.Kind && i.ApiVersion == item.ApiVersion
		})
		if contains {
			// g.Status.Items = lo.Reject(g.Status.Items, func(i v1alpha1.AppGroupItemStatus, _ int) bool {
			// 	return i.Name == item.Name && i.Kind == item.Kind && i.ApiVersion == item.ApiVersion
			// })
			for i, v := range g.Status.Items {
				if v.Name == item.Name && v.Kind == item.Kind && v.ApiVersion == item.ApiVersion {
					g.Status.Items[i] = item
				}
				if i == 0 && v.Title != "" && v.Kind != "Job" {
					g.Spec.Title = v.Title
				}
			}
		} else {
			g.Status.Items = append(g.Status.Items, item)
		}
		g.changed = true

	}
	for i, v := range g.Status.Items {
		if i == 0 && v.Title != "" && v.Kind != "Job" {
			g.Spec.Title = v.Title
		}
	}
	g.FixDeployItem(item)
	if g.parent != nil {
		g.parent.AddStatusItem(item)
	}
	if g.Status.Ready {
		NotifyInstalled(g.AppGroup)
	}
}

func (g *appgroupWrapper) FixDeployItem(item v1alpha1.AppGroupItemStatus) {
	for in, v1 := range g.Status.DeployItems {
		for i, v := range v1.ResourceList {
			if v.Name == item.Name && v.Kind == item.Kind && v.ApiVersion == item.ApiVersion {
				g.changed = true
				if v.DeployStatus != item.DeployStatus && v.DeployStatus != v1alpha1.StatusDeployed && item.DeployStatus != v1alpha1.StatusUnknown {
					g.Status.DeployItems[in].ResourceList[i].DeployStatus = item.DeployStatus
				}
			}
		}
	}
	g.changed = true
	g.AppGroup.ComputeStatus()
}

func (g *appgroupWrapper) RemoveStatusItem(item v1alpha1.AppGroupItemStatus) {
	if g.Status.Items == nil {
		return
	}
	g.changed = true
	g.Status.Items = lo.Reject(g.Status.Items, func(i v1alpha1.AppGroupItemStatus, _ int) bool {
		return i.Name == item.Name && i.Kind == item.Kind && i.ApiVersion == item.ApiVersion
	})
	// 安装记录也移除
	for in, v1 := range g.Status.DeployItems {
		contains := lo.ContainsBy(v1.ResourceList, func(i v1alpha1.ResourceInfo) bool {
			return i.Name == item.Name && i.Kind == item.Kind && i.ApiVersion == item.ApiVersion
		})
		if contains {
			g.Status.DeployItems[in].ResourceList = lo.Reject(v1.ResourceList, func(i v1alpha1.ResourceInfo, _ int) bool {
				return i.Name == item.Name && i.Kind == item.Kind && i.ApiVersion == item.ApiVersion
			})
		}

	}
	g.deleteEnabled = true
	g.AppGroup.ComputeStatus()
	if g.parent != nil {
		g.parent.RemoveStatusItem(item)
	}

}

func (g *appgroupWrapper) DelHelmItem() {
	g.changed = true
	g.Status.Items = lo.Reject(g.Status.Items, func(i v1alpha1.AppGroupItemStatus, _ int) bool {
		return i.IsHelmWorkLoad
	})
}

func (g *appgroupWrapper) HasItem(item v1alpha1.AppGroupItemStatus) bool {
	contains := lo.ContainsBy(g.Status.Items, func(i v1alpha1.AppGroupItemStatus) bool {
		return i.Name == item.Name && i.Kind == item.Kind && i.ApiVersion == item.ApiVersion
	})
	return contains
}

func (g *appgroupWrapper) IsEmpty() bool {
	return g.Status.Items == nil || len(g.Status.Items) == 0

}

func (g *appgroupWrapper) IsChange() bool {
	return g.changed
}

func (g *appgroupWrapper) IsExists() bool {
	return g.exists

}

func (g *appgroupWrapper) IsDeleteEnabled() bool {
	return g.deleteEnabled

}
