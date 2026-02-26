package types

import (
	"errors"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/overselling"
	corev1 "k8s.io/api/core/v1"
)

type k3kUserOrder struct {
	*corev1.ServiceAccount
	*k3kUserOverSelling
	*k3kUserTime
	cost *K3kCost
}

func Newk3kUserOrder(sa *corev1.ServiceAccount, overUser *k3kUserOverSelling, userTime *k3kUserTime, cost *K3kCost) *k3kUserOrder {
	return &k3kUserOrder{sa, overUser, userTime, cost}
}

func (u *k3kUserOrder) SetBaseOrder(orderSn string) {
	u.Labels[W7_BASE_ORDER_SN] = orderSn
}

func (u *k3kUserOrder) SetRenewOrder(orderSn string) {
	u.Labels[W7_RENEW_ORDER_SN] = orderSn
	u.Labels[W7_RENEW_ORDER_STATUS] = W7_ORDER_WAIT //因为支付成功后会判断是否PAID，所以这里先设置为等待支付
}

func (u *k3kUserOrder) SetExpandOrder(orderSn string) {
	u.Labels[W7_EXPAND_ORDER_SN] = orderSn
	u.Labels[W7_EXPAND_ORDER_STATUS] = W7_ORDER_WAIT
}

func (u *k3kUserOrder) GetBaseOrderSn() string {
	return u.Labels[W7_BASE_ORDER_SN]
}

func (u *k3kUserOrder) GetRenewOrderSn() string {
	return u.Labels[W7_RENEW_ORDER_SN]
}

func (u *k3kUserOrder) GetExpandOrderSn() string {
	return u.Labels[W7_EXPAND_ORDER_SN]
}

func (u *k3kUserOrder) SetBaseOrderPaid(info *console.OrderInfo) {
	if u.Labels[W7_BASE_ORDER_SN] == info.OrderSn && u.Labels[W7_BASE_ORDER_STATUS] != W7_ORDER_PAID {
		// u.Labels[K3K_BUY_MODE] = "buy"
		u.Labels[W7_BASE_ORDER_STATUS] = W7_ORDER_PAID
		u.changeExpireTime(int(info.GetHour()))
		rs := overselling.OrderInfoToResource(info)
		u.Annotations[W7_OVER_BASE_RESOURCE] = rs.JsonString()
		u.Annotations[W7_QUOTA_LIMIT_LOCK] = "true" //锁定配额，防止配额被费用套餐覆盖
		u.Labels[W7_OVER_MODE] = "wait"

	}
}

func (u *k3kUserOrder) SetRenewOrderPaid(info *console.OrderInfo) {
	if u.Labels[W7_RENEW_ORDER_SN] == info.OrderSn && u.Labels[W7_RENEW_ORDER_STATUS] != W7_ORDER_PAID {
		// u.Labels[K3K_BUY_MODE] = "renew"
		u.Labels[W7_RENEW_ORDER_STATUS] = W7_ORDER_PAID
		u.changeExpireTime(int(info.GetHour()))
	}
}

func (u *k3kUserOrder) SetExpandOrderPaid(info *console.OrderInfo) {
	if u.Labels[W7_EXPAND_ORDER_SN] == info.OrderSn && u.Labels[W7_EXPAND_ORDER_STATUS] != W7_ORDER_PAID {
		// u.Labels[K3K_BUY_MODE] = "renew"
		u.Labels[W7_EXPAND_ORDER_STATUS] = W7_ORDER_PAID
		// u.changeExpireTime(hour)
		u.Annotations[W7_OVER_RESOURCE] = overselling.OrderInfoToResource(info).JsonString()
		u.Labels[W7_OVER_MODE] = "wait"
	}
}

func (u *k3kUserOrder) SetOrderStatus(info *console.OrderInfo) {
	if info.OrderStatus != "paid" {
		return
	}
	switch info.BuyMode {
	case "base":
		u.SetBaseOrderPaid(info)
		break
	case "renew":
		u.SetRenewOrderPaid(info)
		break
	case "expand":
		u.SetExpandOrderPaid(info)
		break
	}
}

func (u *k3kUserOrder) NeedCreateOrder() bool {
	pass, ok := u.Labels[W7_BASE_ORDER_PASS]
	if ok {
		return pass == "false"
	}
	if u.NeedBuyResource() {
		status, ok := u.Labels[W7_BASE_ORDER_STATUS]
		if !ok {
			return true
		}
		if status == W7_ORDER_PAID {
			return false
		}
		return true
	}
	return false
}

// 是否可以续费

// 到期后必须续费，否则无法使用
func (u *k3kUserOrder) NeedRenew() bool {
	if err := u.CanRenewError(); err != nil {
		expireTime, err := u.GetExpireTime()
		if err != nil {
			return false
		}
		if expireTime.Before(time.Now()) { //3天内可以续费
			// if expireTime.Before(time.Now().Add(-time.Hour * 72)) { //3天内可以续费
			return true
		}
	}
	return false
}

func (u *k3kUserOrder) CanCreateBaseOrderError() error {
	pass, ok := u.Labels[W7_BASE_ORDER_PASS]
	if ok {
		if pass == "false" {
			return nil
		}
	}
	if u.NeedBuyResource() {
		status, ok := u.Labels[W7_BASE_ORDER_STATUS]
		if !ok {
			return nil
		}
		if status == W7_ORDER_PAID {
			return errors.New("已经购买基础资源，无法重复购买")
		}
		return nil
	}
	return errors.New("当前用户未配置费用套餐，无法购买")
}

func (u *k3kUserOrder) CanRenewError() error {
	if u.NeedBuyResource() {
		_, err := u.GetExpireTime() // 如果没有过期时间，则不需要续费
		if err != nil {
			return errors.New("未购买基础资源，无需购买")
		}
		return nil
	}
	return errors.New("当前用户未配置费用套餐，无法购买")
}

func (u *k3kUserOrder) CanExpandError() error {
	if !u.IsOverSellingSuccess() {
		return errors.New("超额检查失败，无法扩容")
	}
	if u.NeedBuyResource() {
		extime, err := u.GetExpireTime() // 如果没有过期时间，则不需要续费
		if err != nil {
			return errors.New("未购买基础资源，无法扩容")
		}
		ok := extime.After(time.Now())
		if ok {
			return nil
		}
		return errors.New("基础资源已过期，无法扩容")
	}
	return errors.New("当前用户未配置费用套餐，无法购买")
}

func (u *k3kUserOrder) NeedBuyResource() bool {
	if u.cost != nil {
		return true
	}
	return false
}

func (u *k3kUserOrder) HasProcessReturnOrder() bool {
	data, ok := u.Annotations[W7_RETURN_ORDER_INFO]
	if ok && data != "" {
		return true
	}
	return false
}
