package types

import (
	corev1 "k8s.io/api/core/v1"
)

type k3kUserBase struct {
	*corev1.ServiceAccount
}

func Newk3kUserBase(sa *corev1.ServiceAccount) *k3kUserBase {
	return &k3kUserBase{sa}
}

func (u *k3kUserBase) IsExpand() bool {
	return u.Labels[W7_EXPAND_ORDER_STATUS] == W7_ORDER_PAID
}
