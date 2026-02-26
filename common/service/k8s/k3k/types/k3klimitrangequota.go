package types

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/overselling"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/json"
)

type Hard struct {
	Cpu       Int64OrString `json:"cpu,omitempty"`
	Memory    Int64OrString `json:"memory,omitempty"`
	Storage   Int64OrString `json:"requests.storage,omitempty"`
	Bandwidth Int64OrString `json:"bandwidth,omitempty"`
}
type limit struct {
	Cpu    Int64OrString `json:"cpu,omitempty"`
	Memory Int64OrString `json:"memory,omitempty"`
}
type Rangehard struct {
	StorageClass string `json:"storageclass,omitempty" protobuf:"bytes,1,opt,name=storageclass"`
	Hard         Hard   `json:"hard,omitempty"`
}

// "{\"storageclass\":\"disk1\",\"hard\":{\"cpu\":2,\"memory\":4,\"requests.storage":\"10\",\"bandwidth\":100}}"
// 1.0.101 版本后，str字段中 不再有单位
func NewLimitRangeQuata(str string) (*LimitRangeQuota, error) {
	lr := &Rangehard{}
	str = strings.ReplaceAll(str, "Mbps", "M")
	str = strings.ReplaceAll(str, "M", "")
	str = strings.ReplaceAll(str, "Gi", "")
	str = strings.ReplaceAll(str, "G", "")
	err := json.Unmarshal([]byte(str), lr)
	if err != nil {
		return nil, err
	}
	result := &LimitRangeQuota{
		Hard: v1.ResourceList{
			v1.ResourceCPU:             resource.MustParse(lr.Hard.Cpu.String()),
			v1.ResourceMemory:          resource.MustParse(lr.Hard.Memory.String() + "Gi"),
			v1.ResourceRequestsStorage: resource.MustParse(lr.Hard.Storage.String() + "Gi"),
			Bandwidth:                  resource.MustParse(lr.Hard.Bandwidth.String() + "M"),
		},
		StorageClass: lr.StorageClass,
	}
	return result, nil
}

// declare
type LimitRangeQuota struct {
	StorageClass string          `json:"storageclass,omitempty" protobuf:"bytes,1,opt,name=storageclass"`
	Hard         v1.ResourceList `json:"hard,omitempty" protobuf:"bytes,3,rep,name=min,casttype=ResourceList,castkey=ResourceName"`
	Limit        v1.ResourceList `json:"limit,omitempty" protobuf:"bytes,3,rep,name=min,casttype=ResourceList,castkey=ResourceName"`
	Unit         string          `json:"unit,omitempty" protobuf:"bytes,2,opt,name=unit"`
	Quantity     int32           `json:"quantity,omitempty" protobuf:"bytes,4,opt,name=quantity"`
}

func (lr *LimitRangeQuota) IsCpuMemoryBandWidthChange(lr2 *LimitRangeQuota) bool {
	if lr2.Hard == nil {
		return false
	}
	if lr.Hard != nil {
		lrHardCpu := lr.GetHardCpu()
		lrHardMemory := lr.GetHardMemory()
		lrHardBd := lr.GetBandWidth()
		lr2HardCpu := lr2.GetHardCpu()
		lr2HardMemory := lr2.GetHardMemory()
		lr2HardBd := lr2.GetBandWidth()
		if lrHardCpu.Cmp(*lr2HardCpu) != 0 || lrHardMemory.Cmp(*lr2HardMemory) != 0 || lrHardBd.Cmp(lr2HardBd) != 0 {
			return true
		}
	}
	return false
}

func (lr *LimitRangeQuota) GetHardRequestStorage() *resource.Quantity {
	if lr.Hard != nil {
		if v, ok := lr.Hard[v1.ResourceRequestsStorage]; ok {
			return &v
		}
	}
	ret := resource.MustParse("0")
	return &ret
}

func (lr *LimitRangeQuota) GetHardSysRequestStorage() *resource.Quantity {
	if lr.Hard != nil {
		if v, ok := lr.Hard[SysStorageSize]; ok {
			return &v
		}
	}
	ret := resource.MustParse("0")
	return &ret
}

//	func (lr *LimitRangeQuota) GetExpandStorage() *resource.Quantity {
//		if lr.Hard != nil {
//			if v, ok := lr.Hard[ExpandStorageSize]; ok {
//				return &v
//			}
//		}
//		ret := resource.MustParse("0")
//		return &ret
//	}
func (lr *LimitRangeQuota) CanResizeSysStorage(unUsed resource.Quantity, resizeTo resource.Quantity) bool {
	currentSysStorage := lr.GetHardSysRequestStorage()
	cloneResizeTo := resizeTo.DeepCopy()
	cloneResizeTo.Sub(*currentSysStorage)

	currentSysStorageStr := currentSysStorage.String()
	cloneResizeToStr := cloneResizeTo.String()
	slog.Info("unUsed", "unUsed", unUsed.String(), "cloneResizeTo", cloneResizeToStr, "currentSysStorage", currentSysStorageStr)
	if unUsed.Cmp(cloneResizeTo) < 0 {
		return false
	}
	return true
}
func (lr *LimitRangeQuota) ResizeSysStorageInt(storageSize int) {
	lr.Hard[SysStorageSize] = resource.MustParse(fmt.Sprintf("%dGi", storageSize))
}

func (lr *LimitRangeQuota) ResizeSysStorage(storageSize resource.Quantity) {
	lr.Hard[SysStorageSize] = storageSize
}

func (lr *LimitRangeQuota) GetHardCpu() *resource.Quantity {
	if lr.Hard != nil {
		return lr.Hard.Cpu()
	}
	ret := resource.MustParse("0")
	return &ret
}

func (lr *LimitRangeQuota) GetHardMemory() *resource.Quantity {
	if lr.Hard != nil {
		return lr.Hard.Memory()
	}
	ret := resource.MustParse("0")
	return &ret
}

func (lr *LimitRangeQuota) GetHardBandWidth() *resource.Quantity {
	if lr.Hard != nil {
		if v, ok := lr.Hard[Bandwidth]; ok {
			return &v
		}
	}
	ret := resource.MustParse("0M")
	return &ret
}

func (lr *LimitRangeQuota) Expand(rs *overselling.Resource) {
	hardRs := lr.GetHardResource()
	newRs := hardRs.Add(*rs)
	lr.ResetHard(newRs)
}

func (lr *LimitRangeQuota) ResetHard(rs *overselling.Resource) {
	lr.Hard[v1.ResourceCPU] = rs.CPU
	lr.Hard[v1.ResourceMemory] = rs.Memory
	lr.Hard[v1.ResourceRequestsStorage] = rs.Storage
	lr.Hard[Bandwidth] = rs.BandWidth
}

func (lr *LimitRangeQuota) GetHardResource() *overselling.Resource {
	return &overselling.Resource{
		CPU:       *lr.GetHardCpu(),
		Memory:    *lr.GetHardMemory(),
		Storage:   *lr.GetHardRequestStorage(),
		BandWidth: *lr.GetHardBandWidth(),
	}
}

func (b *LimitRangeQuota) getBandwidthInt64() int64 {
	ret := b.GetHardBandWidth().String()
	ret = strings.ReplaceAll(ret, "Mbps", "")
	ret = strings.ReplaceAll(ret, "Mi", "")
	ret = strings.ReplaceAll(ret, "M", "")
	s, err := strconv.ParseInt(ret, 10, 64)
	if err != nil {
		return 0
	}
	return s
}

func (lr *LimitRangeQuota) GetHardBuyResource() BuyResource {
	return BuyResource{
		Cpu:       lr.GetHardCpu().Value(),
		Memory:    lr.GetHardMemory().Value() / 1024 / 1024 / 1024,
		Storage:   lr.GetHardRequestStorage().Value() / 1024 / 1024 / 1024,
		Bandwidth: lr.getBandwidthInt64(),
	}
}

// func (lr *LimitRangeQuota) GetHour() int64 {
// 	quantity := lr.Quantity
// 	switch lr.Unit {
// 	case "hour":
// 		return int64(quantity)
// 	case "day":
// 		return int64(quantity) * 24
// 	case "week":
// 		return int64(quantity) * 24 * 7
// 	case "year":
// 		return int64(quantity) * 30 * 12 *24
// 	case "month":
// 		return int64(quantity) * 24 * 30

// 	default:
// 		return 0
// 	}
// }

func (lr *LimitRangeQuota) GetDays() float64 {
	quantity := lr.Quantity
	switch lr.Unit {
	case "hour":
		return float64(quantity) / 24
	case "day":
		return float64(quantity)
	case "week":
		return float64(quantity) * 7
	case "month":
		return float64(quantity) * 30
	case "year":
		return float64(quantity) * 12 * 30

	default:
		return 0
	}
}

func (lr *LimitRangeQuota) GetDefaultUnitQuantity() UnitQuantity {
	return UnitQuantity{Quantity: int64(lr.Quantity), Unit: lr.Unit}
}

/*
*

	带宽限制
*/
func (lr *LimitRangeQuota) GetBandWidth() resource.Quantity {
	if lr.Hard != nil {
		if v, ok := lr.Hard[Bandwidth]; ok {
			return v
		}
	}
	return resource.MustParse("0M")
}

func (lr *LimitRangeQuota) ToString() string {
	result, err := json.Marshal(lr)
	if err != nil {
		return ""
	}
	return string(result)
}
