package types

import (
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/overselling"
	corev1 "k8s.io/api/core/v1"
)

type k3kUserOverSelling struct {
	*corev1.ServiceAccount
	*k3kUserBase
	overResource     *overselling.Resource
	overBaseResource *overselling.Resource
}

func Newk3kUserOverSelling(sa *corev1.ServiceAccount, base *k3kUserBase) *k3kUserOverSelling {
	u := &k3kUserOverSelling{ServiceAccount: sa, k3kUserBase: base}
	u.overResource = overselling.EmptyResource()
	overstr, ok2 := sa.Annotations[W7_OVER_RESOURCE]
	if ok2 {
		u.overResource = overselling.CreateFromString(overstr)
	}
	u.overBaseResource = overselling.EmptyResource()
	overBasestr, ok3 := sa.Annotations[W7_OVER_BASE_RESOURCE]
	if ok3 {
		u.overBaseResource = overselling.CreateFromString(overBasestr)
	}
	return u
}

func (k *k3kUserOverSelling) IsOverSellingWait() bool {
	return k.Labels[W7_OVER_MODE] == "wait"
}

func (k *k3kUserOverSelling) IsOverSellingSuccess() bool {
	return k.Labels[W7_OVER_MODE] == "success"
}

func (k *k3kUserOverSelling) IsOverSellingNoResource() bool {
	return k.Labels[W7_OVER_MODE] == "no-resource"
}

func (u *k3kUserOverSelling) NeedOverSellingCheck() bool {
	return u.CanOverSellingCheck() && !u.IsExpand()
}

// 是否可以超额检查
func (u *k3kUserOverSelling) CanOverSellingCheck() bool {
	return u.Labels[W7_OVER_MODE] == "wait" || u.Labels[W7_OVER_MODE] == "no-resource"
}

func (u *k3kUserOverSelling) GetOverResource() *overselling.Resource {
	if u.IsExpand() {
		return u.overResource
	}
	return u.overBaseResource
}

// wait no-resource success
