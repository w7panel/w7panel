package types

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
)

type param map[string]string
type Params []param

type K3kClusterPolicy struct {
	*v1alpha1.VirtualClusterPolicy
	cost       *K3kCost
	limitRange *LimitRangeQuota
	limitOnce  sync.Once
	onceCost   sync.Once
}

func NewK3kClusterPolicy(v *v1alpha1.VirtualClusterPolicy) *K3kClusterPolicy {
	return &K3kClusterPolicy{
		VirtualClusterPolicy: v,
	}
}

func (v *K3kClusterPolicy) GetCost() *K3kCost {
	// 返回K3kGroup的名称
	v.onceCost.Do(func() {
		v.cost = v.getCost()
	})
	return v.cost
}
func (v *K3kClusterPolicy) GetLimitRange() *LimitRangeQuota {
	v.limitOnce.Do(func() {
		v.limitRange = v.getLimitRange()
	})
	return v.limitRange
}

func (v *K3kClusterPolicy) getCost() *K3kCost {
	costConfig := v.Annotations[W7_COST]
	if costConfig == "" {
		return nil
	}
	cost, err := CreateCostFromString(costConfig)
	if err != nil {
		slog.Error("parse cost config error", "error", err)
		return nil
	}
	return cost
}

func (v *K3kClusterPolicy) getLimitRange() *LimitRangeQuota {
	jstr, ok := v.Annotations[W7_QUOTA_LIMIT]
	if ok {
		lqr2, err := NewLimitRangeQuata(jstr)
		if err != nil {
			slog.Error("parse quota limit error", "error", err)
			return nil
		}
		return lqr2
	}
	return nil
}

func (v *K3kClusterPolicy) CanPublish() bool {
	return v.cost != nil && v.limitRange != nil
}

func (u *K3kClusterPolicy) GetOrderCompute() (*K3kOrderCompute, error) {
	if u.getCost() == nil {
		return nil, errors.New("当前用户未配置费用套餐，无法购买")
	}

	return NewK3kOrderComputeWithCost(u.getCost()), nil
}

func (b *K3kClusterPolicy) ToPublishShopParams(name string) map[string]string {

	compute, err := b.GetOrderCompute()
	if err != nil {
		return nil
	}
	price := compute.GetOriginPrice()

	quantity := b.GetLimitRange().Quantity
	quantityStr := strconv.Itoa(int(quantity))

	return map[string]string{
		"cpu":       strconv.FormatInt(compute.Cpu, 10),
		"memory":    strconv.FormatInt(compute.Memory, 10),
		"storage":   strconv.FormatInt(compute.Storage, 10),
		"bandwidth": strconv.FormatInt(compute.Bandwidth, 10),
		// "buymode":   b.GetCost().BuyMode,
		"unit":      b.GetLimitRange().Unit,
		"quantity":  quantityStr,
		"price":     price.String(),
		"groupName": name,
		// "groupTitle":
	}
}

func (b *K3kClusterPolicy) ToPublishShopParams2(name string) (map[string]interface{}, error) {
	items, err := b.ToPackageItemsParams(true)
	if err != nil {
		return nil, err
	}
	title := b.Annotations["title"]
	pubTitle, ok := b.Annotations["publish-title"]
	if ok {
		title = pubTitle
	}

	return map[string]interface{}{
		"items":     items,
		"groupname": name,
		"title":     title,
		"city":      b.Annotations["city"],
	}, nil
}

func (b *K3kClusterPolicy) ToPackageItemsParams(onlyOnline bool) (Params, error) {

	params := Params{}
	city := b.Annotations["city"]
	title := b.Annotations["title"]
	pubTitle, ok := b.Annotations["publish-title"]
	demo := "1" // 默认非演示环境
	if (b.Labels != nil) && (b.Labels["w7.cc/demo-user"] == "true") {
		demo = "2" // 演示环境
	}
	if ok {
		title = pubTitle
	}
	if b.GetCost() != nil {
		compute, err := b.GetOrderCompute()
		if err != nil {
			return nil, err
		}

		packages := b.GetCost().Packages
		for _, pkg := range packages {
			discountNew := pkg.DiscountNew
			if discountNew.value <= 0 {
				discountNew.value = 100
			}
			items := pkg.Items
			if onlyOnline {
				items = pkg.OnLineItems()
			}
			for _, item := range items {
				buymode := "buy"
				if item.IsGive {
					buymode = "give"
				}
				itemDiscountNew := item.DiscountNew
				if itemDiscountNew.value <= 0 {
					itemDiscountNew.value = 100
				}
				if itemDiscountNew.Value() == 100 || itemDiscountNew.Value() == 0 {
					itemDiscountNew = discountNew
				}
				param := param{
					"cpu":         item.Cpu.String(),
					"memory":      item.Memory.String(),
					"storage":     item.Storage.String(),
					"bandwidth":   item.Bandwidth.String(),
					"discountnew": itemDiscountNew.String(),
					// "discountNew": item.DiscountNew.String(),
					// "discountRenew": item.DiscountRenew.String(),
					"groupname":   b.Name,
					"city":        city,
					"title":       title,
					"buymode":     buymode,
					"quantity":    pkg.Quantity.String(),
					"unit":        pkg.Unit,
					"label":       item.Label,
					"description": strings.Join(item.Description, "|"),
					"demo":        demo,
				}
				computeItem := compute.WithResource(item.ToBuyResource()).WithQuantity(pkg.ToUnitQuantity())
				param["price"] = computeItem.GetOriginPrice().String()
				param["discountprice"] = computeItem.GetDiscountPriceNotGive("base").String()
				params = append(params, param)
			}
		}
	}
	return params, nil
}
