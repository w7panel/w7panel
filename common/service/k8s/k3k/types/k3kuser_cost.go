package types

import (
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

// {"c":"2","m":"2","bw":"100","s":"50","total":1848,"dc":"100","unit":"1年"}
type K3kConstPackageItem struct {
	Cpu           Int64OrString `json:"c"`
	Storage       Int64OrString `json:"s"`
	Memory        Int64OrString `json:"m"`
	Bandwidth     Int64OrString `json:"bw"`
	DiscountNew   Int64OrString `json:"dc_new"`   //折扣
	DiscountRenew Int64OrString `json:"dc_renew"` //续费折扣
	Online        bool          `json:"online"`
	IsGive        bool          `json:"give"`
	Label         string        `json:"label"`
	Description   []string      `json:"description"`
}

func (p *K3kConstPackageItem) ToBuyResource() BuyResource {
	return BuyResource{
		Cpu:       p.Cpu.value,
		Memory:    p.Memory.value,
		Storage:   p.Storage.value,
		Bandwidth: p.Bandwidth.value,
	}
}

func (p *K3kConstPackageItem) Eq(rs BuyResource) bool {
	return p.Cpu.value == rs.Cpu &&
		p.Memory.value == rs.Memory &&
		p.Storage.value == rs.Storage &&
		p.Bandwidth.value == rs.Bandwidth
}

// [{"time":1,"timeUnit":"year","discount_all":100,"discount_new":100,"discount_renew":100,"config":[{"c":"2","m":"2","bw":"100","s":"50","total":1848,"dc":"100","unit":"1年"}],"cpu":12,"memory":12,"bandwidth":12,"storage":12,"total":"48.00"}]
type K3kConstPackage struct {
	Quantity      Int64OrString         `json:"time"`
	Unit          string                `json:"timeUnit"`
	DiscountNew   Int64OrString         `json:"discount_new"`   //新购折扣
	DiscountRenew Int64OrString         `json:"discount_renew"` // 续费
	Items         []K3kConstPackageItem `json:"config"`
}

func (p K3kConstPackage) OnLineItems() []K3kConstPackageItem {
	var result []K3kConstPackageItem
	for _, item := range p.Items {
		if item.Online {
			result = append(result, item)
		}
	}
	return result
}
func (p K3kConstPackage) ToUnitQuantity() UnitQuantity {
	return UnitQuantity{
		Unit:     p.Unit,
		Quantity: p.Quantity.Value(),
	}
}

// 单价
type K3kCost struct {
	// BuyMode   string            `json:"buymode"`
	Cpu       float64           `json:"cpu"`
	Memory    float64           `json:"memory"`
	Storage   float64           `json:"storage"`
	Bandwidth float64           `json:"bandwidth"`
	Packages  []K3kConstPackage `json:"packageConfig"`
}

func (c *K3kCost) ToString() string {
	return fmt.Sprintf("cpu:%v, memory:%v, storage:%v, bandwidth:%v", c.Cpu, c.Memory, c.Storage, c.Bandwidth)
}

func (c *K3kCost) ToJsonString() (string, error) {
	result, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// func (c *K3kCost) NeedBuy() bool {
// 	return c.BuyMode == "buy"
// }

// func (c *K3kCost) IsGive() bool {
// 	return c.BuyMode == "give"
// }

type innercost struct {
	// BuyMode   string            `json:"buymode"`
	Cpu       Float64OrString   `json:"cpu"`
	Memory    Float64OrString   `json:"memory"`
	Storage   Float64OrString   `json:"storage"`
	Bandwidth Float64OrString   `json:"bandwidth"`
	Packages  []K3kConstPackage `json:"packageConfig"`
}

func ConfigMapToCost(config *v1.ConfigMap) (*K3kCost, error) {
	cost := &K3kCost{
		Cpu:       helper.ParseFloat64(config.Data["cpu"]),
		Memory:    helper.ParseFloat64(config.Data["memory"]),
		Storage:   helper.ParseFloat64(config.Data["storage"]),
		Bandwidth: helper.ParseFloat64(config.Data["bandwidth"]),
	}
	costPackages := &[]K3kConstPackage{}
	err := json.Unmarshal([]byte(config.Data["packageConfig"]), costPackages)
	if err != nil {
		slog.Error("parse cost package error", "error", err)
		return nil, err
	}
	cost.Packages = *costPackages
	return cost, nil
}

func ConfigMapToCostString(config *v1.ConfigMap) (string, error) {
	cost, err := ConfigMapToCost(config)
	if err != nil {
		return "", err
	}
	return cost.ToString(), nil
}

func CreateCostFromString(str string) (*K3kCost, error) {
	cost := &innercost{}
	err := json.Unmarshal([]byte(str), cost)
	if err != nil {
		return nil, fmt.Errorf("parse cost error: %v", err)
	}
	cost2 := &K3kCost{
		// BuyMode:   cost.BuyMode,
		Cpu:       cost.Cpu.value,
		Memory:    cost.Memory.value,
		Storage:   cost.Storage.value,
		Bandwidth: cost.Bandwidth.value,
		Packages:  cost.Packages,
	}
	// packages := []K3kConstPackage{}
	// err = json.Unmarshal([]byte(packstr), &packages)
	// if err != nil {
	// 	slog.Error("parse package error", "err", err)
	// }
	// cost2.Packages = packages

	return cost2, nil
}

type k3kUserCost struct {
	cost *K3kCost
}

func Newk3kUserCost(cost *K3kCost) *k3kUserCost {
	return &k3kUserCost{cost}
}

func (u *k3kUserCost) NeedBuyResource() bool {
	if u.cost != nil {
		return true
	}
	return false
}
