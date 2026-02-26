package types

import (
	"errors"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

/*
*
订单计算器
*/
type K3kOrderCompute struct {
	BuyResource
	UnitQuantity
	cost *K3kCost
	*Coupon
}

func NewK3kOrderComputeWithCostLimitRange(limitRange *LimitRangeQuota, cost *K3kCost) *K3kOrderCompute {
	buyRs := limitRange.GetHardBuyResource()
	qq := limitRange.GetDefaultUnitQuantity()
	return NewK3kOrderCompute(buyRs, qq, cost, nil)
}
func NewK3kOrderCompute(buyResource BuyResource, unitQuantity UnitQuantity, cost *K3kCost, coupon *Coupon) *K3kOrderCompute {
	return &K3kOrderCompute{
		BuyResource:  buyResource,
		UnitQuantity: unitQuantity,
		cost:         cost,
		Coupon:       coupon,
	}
}
func NewK3kOrderComputeWithCost(cost *K3kCost) *K3kOrderCompute {
	return &K3kOrderCompute{
		cost: cost,
	}
}

// cpu价格计算
func (c *K3kOrderCompute) GetUnitPrice() decimal.Decimal {
	cpu := decimal.NewFromFloat(c.cost.Cpu).Mul(decimal.NewFromInt(c.BuyResource.Cpu))
	memory := decimal.NewFromFloat(c.cost.Memory).Mul(decimal.NewFromInt(c.BuyResource.Memory))
	disk := decimal.NewFromFloat(c.cost.Storage).Mul(decimal.NewFromInt(c.BuyResource.Storage))
	bandwidth := decimal.NewFromFloat(c.cost.Bandwidth).Mul(decimal.NewFromInt(c.BuyResource.Bandwidth))
	return cpu.Add(memory).Add(disk).Add(bandwidth)
}

// 原价
func (c *K3kOrderCompute) GetOriginPrice() decimal.Decimal {
	return c.GetUnitPrice().Mul(decimal.NewFromFloat(c.UnitQuantity.GetMonth()))
}

// 判断是否匹配 买资源 以及 单位数量
func (c *K3kOrderCompute) IsCouponMatch() bool {
	if c.Coupon == nil {
		return false
	}
	return c.Coupon.BuyResource.Eq(c.BuyResource) && c.Coupon.UnitQuantity.Eq(c.UnitQuantity)
}

// 打折后的价格
func (c *K3kOrderCompute) GetDiscountPrice(buyMode string) decimal.Decimal {
	if (buyMode == "base") && c.IsGiveInBaseBuyMode() {
		return decimal.Zero
	}
	if c.Coupon != nil {
		// 判断是否匹配 买资源 以及 单位数量
		if c.IsCouponMatch() {
			originPrice := c.GetOriginPrice()
			return originPrice.Mul(decimal.NewFromInt(c.Coupon.Discount)).Div(decimal.NewFromInt(100))
		}
	}
	return c.GetDiscountPriceNotGive(buyMode)
}

func (c *K3kOrderCompute) GetDiscountPriceNotGive(buyMode string) decimal.Decimal {
	discount := c.GetDiscount(buyMode)
	originPrice := c.GetOriginPrice()
	return originPrice.Mul(decimal.NewFromInt(discount)).Div(decimal.NewFromInt(100))
}

func (c *K3kOrderCompute) IsGiveInBaseBuyMode() bool {
	if c.cost.Packages == nil {
		return false
	}
	for _, pkg := range c.cost.Packages {
		if pkg.Quantity.Value() == c.UnitQuantity.Quantity && pkg.Unit == c.UnitQuantity.Unit {
			for _, item := range pkg.Items {
				ac := item.Eq(c.BuyResource)
				if ac {
					if item.IsGive {
						return true
					}
				}
			}
		}
	}
	return false
}

func (c *K3kOrderCompute) GetExpandPrice(time2 time.Time) (decimal.Decimal, error) {

	if time2.IsZero() {
		return decimal.Zero, errors.New("到期时间未设置")
	}
	if time2.Before(time.Now()) {
		return decimal.Zero, errors.New("账户已过期，无法扩容")
	}
	sub := time2.Sub(time.Now())
	hour := sub.Hours()
	months := decimal.NewFromFloat(hour).Div(decimal.NewFromInt32(30 * 24))
	return c.GetUnitPrice().Mul(months), nil
}

func (c *K3kOrderCompute) GetDiscount(buyMode string) int64 {
	if c.cost.Packages == nil {
		return 100
	}
	for _, pkg := range c.cost.Packages {
		if pkg.Quantity.Value() == c.UnitQuantity.Quantity && pkg.Unit == c.UnitQuantity.Unit {
			for _, item := range pkg.Items {
				ac := item.Eq(c.BuyResource)
				if ac {
					if buyMode == "base" {
						if item.DiscountNew.value == 100 || item.DiscountNew.value == 0 { //100 不打折 直接原价
							continue
						}
						return (item.DiscountNew.value)
					} else if buyMode == "renew" {
						if item.DiscountRenew.value == 100 || item.DiscountRenew.Value() == 0 { //100 不打折 直接原价
							continue
						}
						return (item.DiscountRenew.value)
					}
				}
			}
			if buyMode == "base" { //order.BASE_BUY {
				if pkg.DiscountNew.value == 100 || pkg.DiscountNew.Value() == 0 { //100 不打折 直接原价
					continue
				}
				return (pkg.DiscountNew.value)
			}
			if buyMode == "renew" {
				if pkg.DiscountRenew.value == 100 || pkg.DiscountRenew.Value() == 0 {
					continue
				}
				return (pkg.DiscountRenew.value)
			}
		}
	}
	return 100
}

func (b *K3kOrderCompute) WithResource(rs BuyResource) *K3kOrderCompute {
	return NewK3kOrderCompute(rs.Clone(), b.UnitQuantity.Clone(), b.cost, b.Coupon)
}

func (b *K3kOrderCompute) WithCoupon(coupon *Coupon) *K3kOrderCompute {
	return NewK3kOrderCompute(b.BuyResource.Clone(), b.UnitQuantity.Clone(), b.cost, coupon)
}

func (b *K3kOrderCompute) SubResource(rs BuyResource) *K3kOrderCompute {
	nrs := b.BuyResource.Sub(rs)
	return NewK3kOrderCompute(nrs, b.UnitQuantity.Clone(), b.cost, b.Coupon)
}

func (b *K3kOrderCompute) WithQuantity(uq UnitQuantity) *K3kOrderCompute {
	return NewK3kOrderCompute(b.BuyResource.Clone(), uq.Clone(), b.cost, b.Coupon)
}

func (b *K3kOrderCompute) ToReqParams() map[string]string {
	return map[string]string{
		"cpu":       strconv.FormatInt(b.Cpu, 10),
		"memory":    strconv.FormatInt(b.Memory, 10),
		"storage":   strconv.FormatInt(b.Storage, 10),
		"bandwidth": strconv.FormatInt(b.Bandwidth, 10),
	}
}
