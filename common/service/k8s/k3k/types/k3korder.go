package types

import (
	"errors"
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/overselling"
	"k8s.io/apimachinery/pkg/api/resource"
)

type PriceResource struct {
	BuyResource
	UnitQuantity
}

type Coupon struct {
	PriceResource
	Discount  int64
	GroupName string
	CanUse    bool
}

type LockReturnK3kOrder struct {
	CurrentTime int64 `json: "currentTime"`
	BuyResource
	OrderSn string `json: "orderSn"`
	BuyMode string `json: "buyMode"`
}

func ApiCouponToCoupon(apicoupon *console.Coupon) *Coupon {
	if apicoupon.Discount > 100 {
		apicoupon.Discount = 100
	}
	return &Coupon{
		PriceResource: PriceResource{
			BuyResource: BuyResource{
				Cpu:       apicoupon.Cpu,
				Memory:    apicoupon.Memory,
				Storage:   apicoupon.Storage,
				Bandwidth: apicoupon.Bandwidth,
			},
			UnitQuantity: UnitQuantity{
				Unit:     apicoupon.TimeUnit,
				Quantity: apicoupon.TimeQuantity,
			},
		},
		GroupName: apicoupon.Groupname,
		Discount:  apicoupon.Discount,
		CanUse:    apicoupon.CanUse,
	}
}

func K3kOrderToBuyResource(order *console.K3kOrder) BuyResource {
	return BuyResource{
		Cpu:       order.Cpu,
		Memory:    order.Memory,
		Storage:   order.Storage,
		Bandwidth: order.Bandwidth,
	}
}

var ZeroBuyResource = BuyResource{}

type BuyResource struct {
	Cpu       int64 `form:"cpu"`
	Storage   int64 `form:"storage"`
	Memory    int64 `form:"memory"`
	Bandwidth int64 `form:"bandwidth"`
}

func (b BuyResource) Diff(b2 BuyResource) bool {
	return b.Cpu < b2.Cpu ||
		b.Storage < b2.Storage ||
		b.Memory < b2.Memory ||
		b.Bandwidth < b2.Bandwidth
}
func (b BuyResource) Less(b2 BuyResource) bool {
	return b.Cpu < b2.Cpu ||
		b.Storage < b2.Storage ||
		b.Memory < b2.Memory ||
		b.Bandwidth < b2.Bandwidth
}

func (b BuyResource) Eq(b2 BuyResource) bool {
	return b.Cpu == b2.Cpu &&
		b.Storage == b2.Storage &&
		b.Memory == b2.Memory &&
		b.Bandwidth == b2.Bandwidth
}

func (b BuyResource) Sub(b2 BuyResource) BuyResource {
	return BuyResource{
		Cpu:       b.Cpu - b2.Cpu,
		Storage:   b.Storage - b2.Storage,
		Memory:    b.Memory - b2.Memory,
		Bandwidth: b.Bandwidth - b2.Bandwidth,
	}
}

func (b BuyResource) Valid() error {
	if
	// b.Cpu == 0 && b.Storage == 0 && b.Bandwidth == 0 && b.Memory == 0) ||
	b.Cpu < 0 || b.Storage < 0 || b.Bandwidth < 0 || b.Memory < 0 {
		slog.Error("buy resource is invalid", "cpu", b.Cpu, "storage", b.Storage, "memory", b.Memory, "bandwidth", b.Bandwidth)
		return errors.New("buy resource is invalid")
	}
	return nil
}
func (b BuyResource) IsEmpty() bool {
	if b.Cpu == 0 && b.Storage == 0 && b.Bandwidth == 0 && b.Memory == 0 {
		return true
	}
	return false

}

func (b BuyResource) Clone() BuyResource {
	return BuyResource{
		Cpu:       b.Cpu,
		Storage:   b.Storage,
		Memory:    b.Memory,
		Bandwidth: b.Bandwidth,
	}
}

func (b BuyResource) ToOverSellingResource() *overselling.Resource {
	return &overselling.Resource{
		CPU:       resource.MustParse(fmt.Sprintf("%d", b.Cpu)),
		Memory:    resource.MustParse(fmt.Sprintf("%dGi", b.Memory)),
		Storage:   resource.MustParse(fmt.Sprintf("%dGi", b.Storage)),
		BandWidth: resource.MustParse(fmt.Sprintf("%dM", b.Bandwidth)),
	}
}

type UnitQuantity struct {
	Quantity int64  `form:"quantity"`
	Unit     string `form:"unit"`
}

func (u UnitQuantity) IsEmpty() bool {
	return u.Quantity <= 0
}

func (u UnitQuantity) Eq(uq UnitQuantity) bool {
	return u.Quantity == uq.Quantity && u.Unit == uq.Unit
}

func (u UnitQuantity) Clone() UnitQuantity {
	return UnitQuantity{
		Quantity: u.Quantity,
		Unit:     u.Unit,
	}
}

func (lr UnitQuantity) GetMonth() float64 {
	quantity := lr.Quantity
	switch lr.Unit {
	case "hour":
		return float64(quantity) / 30 / 24
	case "day":
		return float64(quantity) / 30
	case "month":
		return float64(quantity)
	case "year":
		return float64(quantity) * 12

	default:
		return 0
	}
}

func (lr UnitQuantity) GetHours() float64 {
	quantity := lr.Quantity
	switch lr.Unit {
	case "hour":
		return float64(quantity)
	case "day":
		return float64(quantity) * 24
	case "month":
		return float64(quantity) * 30 * 24
	case "year":
		return float64(quantity) * 12 * 30 * 24

	default:
		return 0
	}
}

type BuyBaseResource struct {
	BaseConfigName string `json:"baseOAuthConfigName"`
	UnitQuantity
	BuyResource
	CouponCode string `form:"couponCode"`
}

// 续费
type BuyRenewResource struct {
	BaseConfigName string `form:"baseOAuthConfigName"`
	UnitQuantity
	CouponCode string `form:"couponCode"`
}

type BuyExpandResource struct {
	// BuyResource
	BaseConfigName string `form:"baseOAuthConfigName"`
	BuyResource
}

func (b *BuyExpandResource) Valid() error {
	if b.Cpu == 0 && b.Storage == 0 && b.Bandwidth == 0 && b.Memory == 0 {
		return errors.New("至少购买一个资源")
	}
	return nil
}

type BuyExpandResourceParams struct {
	BaseConfigName string `form:"baseOAuthConfigName"`
	Cpu            string `form:"cpu"`
	Storage        string `form:"storage"`
	Memory         string `form:"memory"`
	Bandwidth      string `form:"bandwidth"`
}

type BuyNotify struct {
	K3kName string `form:"k3kName" binding:"required"`
	OrderSn string `form:"orderSn" binding:"required"`
}
