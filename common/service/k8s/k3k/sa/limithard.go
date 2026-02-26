package sa

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Limitrangeclient struct {
	client.Client
}

type LimitHardWrapper struct {
	*v1.ServiceAccount
	lrq *k3ktypes.LimitRangeQuota
	sync.Once
}

func NewLimitRangeWrapper(sa *v1.ServiceAccount) LimitHardWrapper {
	return LimitHardWrapper{ServiceAccount: sa}
}

// func (l *LimitHardWrapper) GetNamespace() string {
// 	return l.Namespace
// }

func (l *LimitHardWrapper) GetName() string {
	return l.Name
}

func (l *LimitHardWrapper) IsShared() bool {
	return l.Annotations[k3ktypes.K3K_CLUSTER_MODE] == "shared"
}

func (l *LimitHardWrapper) decodeToLrq(jstr string) (*k3ktypes.LimitRangeQuota, error) {
	lrq, err := k3ktypes.NewLimitRangeQuata(jstr)
	if err != nil {
		slog.Error("decodeToLrq", "err", err)
		return nil, err
	}
	return lrq, nil
}

func (l *LimitHardWrapper) decodeOnce() {
	l.Once.Do(func() {
		jstr, ok := l.Annotations[k3ktypes.W7_QUOTA_LIMIT]
		if !ok {
			return
		}
		slog.Debug("decodeOnce", "jstr", jstr)
		lrq, err := l.decodeToLrq(jstr)
		if err != nil {
			return
		}
		l.lrq = lrq
	})
}

func (l *LimitHardWrapper) GetResourceQuota() (*v1.ResourceQuota, error) {
	l.decodeOnce()
	if l.lrq == nil {
		return nil, errors.New("no limit found")
	}
	hard := l.lrq.Hard
	newHard := make(v1.ResourceList)
	for k, v := range hard {
		nk := k
		if k == "bandwidth" {
			continue
		}
		if k == k3ktypes.DataStorageSize || k == k3ktypes.SysStorageSize {
			continue
		}
		if v.IsZero() {
			myk := k
			slog.Error("zero k", "myk", myk, "v", v.String())
			continue
		}
		if !l.IsShared() { // 如果不是共享模式，则不限制requests limit virtual模式https pod 会无法启动 (自动证书pod 会无法启动)
			if (nk == v1.ResourceRequestsCPU) || (nk == v1.ResourceRequestsMemory) || k == v1.ResourceCPU || k == v1.ResourceMemory {
				continue
			}
		}

		if k == v1.ResourceCPU {
			nk = v1.ResourceRequestsCPU
			newHard[v1.ResourceLimitsCPU] = v
		}
		if k == v1.ResourceMemory {
			nk = v1.ResourceRequestsMemory
			newHard[v1.ResourceLimitsMemory] = v
		}

		newHard[nk] = v
	}

	originCount := len(newHard)
	zeroCount := 0
	for _, v := range newHard {
		if v.IsZero() {
			zeroCount++
		}
	}
	if originCount == zeroCount {
		return nil, errors.New("no hard foundx")
	}
	q := &v1.ResourceQuota{
		Spec: v1.ResourceQuotaSpec{
			Hard: newHard,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      l.Name,
			Namespace: "k3k-" + l.Name,
		},
	}
	if q.Spec.Hard == nil || len(q.Spec.Hard) == 0 {
		return nil, errors.New("no hard  found")
	}
	return q, nil
}

func (l *LimitHardWrapper) GetLimitRange() (*v1.LimitRange, error) {
	l.decodeOnce()
	if l.lrq == nil {
		return nil, errors.New("no limit found")
	}
	if !(l.IsShared()) {
		return nil, errors.New("not shared")
	}

	originCount := len(l.lrq.Limit)
	zeroCount := 0
	for _, v := range l.lrq.Limit {
		if v.IsZero() {
			zeroCount++
		}
	}
	if originCount == zeroCount {
		return nil, errors.New("no limit found")
	}
	limitrange := &v1.LimitRange{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LimitRange",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      l.Name,
			Namespace: "k3k-" + l.Name,
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Type:           v1.LimitTypeContainer,
					Default:        l.lrq.Limit,
					DefaultRequest: l.lrq.Limit,
				},
			},
		},
	}
	return limitrange, nil
}

func NewLimitRangeClient(c client.Client) *Limitrangeclient {
	return &Limitrangeclient{c}
}

func (lclient *Limitrangeclient) Handle(ctx context.Context, sa *v1.ServiceAccount) error {
	// 使用defer-recover来捕获可能的panic
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic in Handle", "panic", r)
		}
	}()
	// if (sa.Labels[k3ktypes.W7_QUOTA_LIMIT] != "") {
	// 	slog.Info("Skipping LimitRange processing for service account with existing annotation")
	// 	if sa.Labels[k3ktypes.W7_BASE_ORDER_STATUS] != "paid" {
	// 		return nil
	// 	}
	// }
	warpper := &LimitHardWrapper{ServiceAccount: sa}

	// 处理LimitRange
	slog.Info("Processing LimitRange", "serviceAccount", sa.Name)
	limit, err := warpper.GetLimitRange()
	if limit != nil && err == nil {
		limitCopy := limit.DeepCopy()
		result, err := controllerutil.CreateOrPatch(ctx, lclient.Client, limitCopy, func() error {
			limitCopy.Spec.Limits = limit.Spec.Limits
			return nil
		})
		if err != nil {
			slog.Error("Failed to create or patch LimitRange", "error", err, "result", result)
			return err
		}
		slog.Info("Successfully processed LimitRange")
	} else {
		slog.Info("Deleting LimitRange", "error", err)
		err = lclient.DeleteLimitRange(ctx, sa)
		if err != nil {
			slog.Error("Delete limit range error", "error", err)
		}
	}

	// 处理ResourceQuota
	slog.Info("Processing ResourceQuota", "serviceAccount", sa.Name)
	resourceQa, err := warpper.GetResourceQuota()
	if resourceQa != nil && err == nil {
		rcopy := resourceQa.DeepCopy()
		result, err := controllerutil.CreateOrPatch(ctx, lclient.Client, rcopy, func() error {
			rcopy.Spec.Hard = resourceQa.Spec.Hard
			return nil
		})
		slog.Info("CreateOrPatch ResourceQuota result", "result", result, "error", err)
		if err != nil {
			slog.Error("Failed to create or patch ResourceQuota", "error", err)
			return err
		}
		slog.Info("Successfully processed ResourceQuota")
	} else {
		slog.Info("Deleting ResourceQuota", "error", err)
		err = lclient.DeleteResourceQuota(ctx, sa)
		if err != nil {
			slog.Error("Delete resource quota error", "error", err)
		}
	}

	slog.Info("Handle completed successfully")
	return nil
}

func (c *Limitrangeclient) GetName(saName string) string {
	return saName
}
func (c *Limitrangeclient) GetNamespace(saName string) string {
	return "k3k-" + saName
}

func (c *Limitrangeclient) Delete(ctx context.Context, sa *v1.ServiceAccount) error {
	err := c.DeleteLimitRange(ctx, sa)
	if err != nil {
		slog.Error("delete limit range error", "error", err)
		// return err
	}
	return c.DeleteResourceQuota(ctx, sa)
}

func (c *Limitrangeclient) DeleteLimitRange(ctx context.Context, sa *v1.ServiceAccount) error {
	err := c.Client.Delete(ctx, &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetName(sa.Name),
			Namespace: c.GetNamespace(sa.Name),
		},
	})
	return err
}

func (c *Limitrangeclient) DeleteResourceQuota(ctx context.Context, sa *v1.ServiceAccount) error {

	return c.Client.Delete(ctx, &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetName(sa.Name),
			Namespace: c.GetNamespace(sa.Name),
		},
	})
}
