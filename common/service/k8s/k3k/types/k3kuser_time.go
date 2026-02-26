package types

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	corev1 "k8s.io/api/core/v1"
)

type k3kUserTime struct {
	*corev1.ServiceAccount
}

func Newk3kUserTime(sa *corev1.ServiceAccount) *k3kUserTime {
	return &k3kUserTime{sa}
}

func (u *k3kUserTime) changeExpireTime(hour int) {
	expireTime, err := u.GetExpireTime()
	if err != nil {
		expireTime = time.Now().Add(time.Hour * time.Duration(hour))
	}
	if err == nil {
		if expireTime.Before(time.Now()) {
			expireTime = time.Now()
		}
		expireTime = expireTime.Add(time.Hour * time.Duration(hour))
	}
	u.Annotations[K3K_EXPIRE_TIME] = expireTime.Format("2006-01-02 15:04:05")
}

func (u *k3kUserTime) GetExpireTime() (time.Time, error) {
	expireTimeStr, ok := u.Annotations[K3K_EXPIRE_TIME]
	if !ok {
		return time.Time{}, fmt.Errorf("expire time not set")
	}
	return time.Parse("2006-01-02 15:04:05", expireTimeStr)
}

func (u *k3kUserTime) IsExpired() bool {
	expireTime, err := u.GetExpireTime()
	if err != nil {
		return false // 如果没有设置过期时间，默认不过期
	}
	return time.Now().After(expireTime)
}

func (u *k3kUserTime) HasExpireTime() bool {
	_, ok := u.Annotations[K3K_EXPIRE_TIME]
	return ok
}

func (u *k3kUserTime) HasPendingRecycleTime() bool {
	_, ok := u.Annotations[K3K_PENDING_RECYCLE_TIME]
	return ok
}

// 获取待回收时间
func (u *k3kUserTime) GetPendingRecycleTime() (time.Time, error) {
	recycleTime, ok := u.Annotations[K3K_PENDING_RECYCLE_TIME]
	if !ok {
		expireTime, err := u.GetExpireTime()
		if err != nil {
			return time.Time{}, fmt.Errorf("pending recycle time not set")
		}
		return expireTime.Add(3 * 24 * time.Hour), nil
	}
	return time.Parse("2006-01-02 15:04:05", recycleTime)
}
func (u *k3kUserTime) SetPendingRecycleTime() {
	if u.HasPendingRecycleTime() {
		return
	}
	defaultTime := time.Now().Add(3 * 24 * time.Hour)
	if u.HasExpireTime() {
		expireTime, err := u.GetExpireTime()
		if err == nil {
			defaultTime = expireTime.Add(72 * time.Hour)
		}
	}
	u.Annotations[K3K_PENDING_RECYCLE_TIME] = defaultTime.Format("2006-01-02 15:04:05")
}

func (u *k3kUserTime) DelPendingRecycleTime() {
	delete(u.Annotations, K3K_PENDING_RECYCLE_TIME)
}

// 检查待回收是否超过3天
func (u *k3kUserTime) IsPendingRecycleExpired() bool {
	// _, ok := u.Annotations[K3K_EXPIRE_TIME]
	// if !ok {
	// 	return false //不存在就不计算过期
	// }
	pendingTime, err := u.GetPendingRecycleTime()
	if err != nil {
		return false
	}
	slog.Info("pending recycle time", "pendingTime", pendingTime, "now", time.Now())
	return time.Now().After(pendingTime)
}

func (u *k3kUser) GetDiffDays() int64 {
	expireTime, err := u.GetExpireTime()
	if err != nil {
		return 0
	}
	diffDays := expireTime.Sub(time.Now()).Hours() / 24 / 30
	return int64(diffDays)
}
func (u *k3kUser) GetDiffMonths() decimal.Decimal {
	expireTime, err := u.GetExpireTime()
	if err != nil {
		return decimal.Zero
	}
	if expireTime.Before(time.Now()) {
		return decimal.Zero
	}
	hours := expireTime.Sub(time.Now()).Hours()
	diffMonths := decimal.NewFromFloat(hours).Div(decimal.NewFromFloat(24 * 30))
	return (diffMonths)
}
