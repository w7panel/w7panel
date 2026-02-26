package metrics

import (
	"fmt"
	"math"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

type cgroupStorage struct {
	// last stores node metric points from last scrape
	last *MetricsPoint
	// prev stores node metric points from scrape preceding the last one.
	// Points timestamp should proceed the corresponding points from last.
	prev *MetricsPoint
}

type MetricsPoint struct {
	// StartTime is the start time of container/node. Cumulative CPU usage at that moment should be equal zero.
	StartTime time.Time
	// Timestamp is the time when metric point was measured. If CPU and Memory was measured at different time it should equal CPU time to allow accurate CPU calculation.
	Timestamp time.Time
	// CumulativeCpuUsed is the cumulative cpu used at Timestamp from the StartTime of container/node. Unit: nano core * seconds.
	CumulativeCpuUsed uint64
	// MemoryUsage is the working set size. Unit: bytes.
	MemoryUsage uint64
}

func resourceUsage(last, prev MetricsPoint) (corev1.ResourceList, error) {
	if last.StartTime.Before(prev.StartTime) {
		return corev1.ResourceList{}, fmt.Errorf("unexpected decrease in startTime of node/container")
	}
	if last.CumulativeCpuUsed < prev.CumulativeCpuUsed {
		return corev1.ResourceList{}, fmt.Errorf("unexpected decrease in cumulative CPU usage value")
	}
	window := last.Timestamp.Sub(prev.Timestamp)
	cpuUsage := float64(last.CumulativeCpuUsed-prev.CumulativeCpuUsed) / window.Seconds()
	return corev1.ResourceList{
		corev1.ResourceCPU:    uint64Quantity(uint64(cpuUsage), resource.DecimalSI, -9),
		corev1.ResourceMemory: uint64Quantity(last.MemoryUsage, resource.BinarySI, 0),
	}, nil
}

// uint64Quantity converts a uint64 into a Quantity, which only has constructors
// that work with int64 (except for parse, which requires costly round-trips to string).
// We lose precision until we fit in an int64 if greater than the max int64 value.
func uint64Quantity(val uint64, format resource.Format, scale resource.Scale) resource.Quantity {
	q := *resource.NewScaledQuantity(int64(val), scale)
	if val > math.MaxInt64 {
		// lose an decimal order-of-magnitude precision,
		// so we can fit into a scaled quantity
		klog.V(2).InfoS("Found unexpectedly large resource value, losing precision to fit in scaled resource.Quantity", "value", val)
		q = *resource.NewScaledQuantity(int64(val/10), resource.Scale(1)+scale)
	}
	q.Format = format
	return q
}
