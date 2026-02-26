package overselling

import (
	"encoding/json"
	"fmt"

	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// 集群已分配资源
type Resource struct {
	CPU       resource.Quantity `json:"cpu"`
	Memory    resource.Quantity `json:"memory"`
	Storage   resource.Quantity `json:"storage"`
	BandWidth resource.Quantity `json:"bandwidth"`
}

func (a *Resource) Add(b Resource) *Resource {
	a.CPU.Add(b.CPU)
	a.Memory.Add(b.Memory)
	a.Storage.Add(b.Storage)
	a.BandWidth.Add(b.BandWidth)
	return a
}

func (a *Resource) Dayu(b Resource) bool {
	if a.CPU.Cmp(b.CPU) > 0 &&
		a.Memory.Cmp(b.Memory) > 0 &&
		a.Storage.Cmp(b.Storage) > 0 {
		return true
	}
	return false
}
func (a *Resource) Clone() *Resource {
	return &Resource{
		CPU:       a.CPU,
		Memory:    a.Memory,
		Storage:   a.Storage,
		BandWidth: a.BandWidth,
	}
}
func (a *Resource) JsonString() string {
	str, err := json.Marshal(a)
	if err != nil {
		return "{}"
	}
	return string(str)
}

func CreateFromString(str string) *Resource {
	a := &Resource{}
	err := json.Unmarshal([]byte(str), a)
	if err != nil {
		return &Resource{}
	}
	return a
}

func EmptyResource() *Resource {
	return &Resource{
		CPU:       resource.MustParse("0"),
		Memory:    resource.MustParse("0Gi"),
		Storage:   resource.MustParse("0Gi"),
		BandWidth: resource.MustParse("0M"),
	}
}

// 百分比超额配置
type OverSellingConfig struct {
	CPU          int32 `json:"cpu"`
	Memory       int32 `json:"memory"`
	Storage      int32 `json:"storage"`
	BandWidth    int32 `json:"bandwidth"`
	BandWidthNum int32 `json:"bandwidthNum"`
}

// 集群已分配资源与百分比超额配置
type OverSellingResource struct {
	Allocated   Resource
	OverSelling OverSellingConfig
}

func (a *OverSellingResource) OverSellingResource() Resource {
	if a.OverSelling.CPU > 0 {
		a.Allocated.CPU.Set(a.Allocated.CPU.Value() * int64(a.OverSelling.CPU) / 100)
	}
	if a.OverSelling.Memory > 0 {
		a.Allocated.Memory.Set(a.Allocated.Memory.Value() * int64(a.OverSelling.Memory) / 100)
	}
	if a.OverSelling.Storage > 0 {
		a.Allocated.Storage.Set(a.Allocated.Storage.Value() * int64(a.OverSelling.Storage) / 100)
	}
	if a.OverSelling.BandWidth > 0 {
		a.Allocated.BandWidth.Set(a.Allocated.BandWidth.Value() * int64(a.OverSelling.BandWidth) / 100)
	}
	return a.Allocated
}

type OverSellingCheck struct {
	CurrentAllocated Resource
}

func OrderInfoToResource(orderInfo *console.OrderInfo) *Resource {
	return &Resource{
		CPU:       resource.MustParse(fmt.Sprintf("%d", orderInfo.Cpu)),
		Memory:    resource.MustParse(fmt.Sprintf("%dGi", orderInfo.Memory)),
		Storage:   resource.MustParse(fmt.Sprintf("%dGi", orderInfo.Storage)),
		BandWidth: resource.MustParse(fmt.Sprintf("%dM", orderInfo.Bandwidth)),
		// Bandwidth: orderInfo.Bandwidth,
	}
}

// k3kuser 循环引用了
type oversellingUser struct {
	*corev1.ServiceAccount
}



func (a *oversellingUser) IsClusterUser() bool {
	return a.Labels["w7.cc/user-mode"] == "cluster"
}


