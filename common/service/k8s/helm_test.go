package k8s

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
)

func TestHelm_Exists(t *testing.T) {
	sdk := NewK8sClientInner()
	helm := NewHelm(sdk)
	namespace := "default"
	labelSelector := "w7.cc/identifie=nvidia-gpuoperator"
	exists := helm.Exists(namespace, labelSelector)
	if !exists {
		t.Errorf("Expected to find release with label selector %s in namespace %s", labelSelector, namespace)
	}
}

func TestHelm_ReUseValues(t *testing.T) {
	sdk := NewK8sClientInner()
	helm := NewHelm(sdk)
	namespace := "default"

	vals := map[string]interface{}{
		"devicePlugin.passDeviceSpecsEnabled": "true",
	}
	//"user suplied labels contains system reserved label name. System labels: [name owner status version createdAt modifiedAt]"
	ctx := context.Background()
	result, err := helm.ReUseValues(ctx, vals, "vgpu-hami-diqynppk", namespace)
	if err != nil {
		t.Errorf("Failed to reuse values: %v", err)
	}
	t.Logf("Reused values: %v", result)
}

// helm upgrade zpk1 --set global.cluster.storageClassName=disk1 --set global.DOMAIN=zpk1.fan.b2.sz.w7.com ./zpk-4.0.45.tgz
func Test_upgrade(t *testing.T) {
	sdk := NewK8sClientInner()
	helm := NewHelm(sdk)
	// namespace := "default"

	// vals1 := map[string]interface{}{
	// 	"global.cluster.storageClassName": "disk1",
	// 	"global.DOMAIN":                   "zpk1.fan.b2.sz.w7.com",
	// }
	vals1 := []string{
		"global.cluster.storageClassName=disk1",
		"global.DOMAIN=zpk1.fan.b2.sz.w7.com",
		"global.cluster.storageSize=10Gi",
		"global.cluster.storageRWmode='ReadWriteOnce",
	}
	settings := cli.New()
	optValues := &values.Options{}
	for _, val := range vals1 {
		optValues.StringValues = append(optValues.StringValues, val)
	}

	provider := getter.All(settings)
	vals, err := optValues.MergeValues(provider)
	if err != nil {
		slog.Error("merge values error", "err", err)
		os.Exit(1)
		return
	}
	chart, err := LocateChartByHelm("", "https://cdn.w7.cc/ued/zpk/zpk-4.0.45.tgz", "")
	if err != nil {
		t.Errorf("Failed to locate chart: %v", err)
		return
	}
	//"user suplied labels contains system reserved label name. System labels: [name owner status version createdAt modifiedAt]"
	ctx := context.Background()
	result, err := helm.Upgrade(ctx, chart, vals, "w7-zpkv2", "default", map[string]string{"a": "1"})
	if err != nil {
		t.Errorf("Failed to reuse values: %v", err)
		slog.Error("Failed to reuse values", "err", err)
	}
	t.Logf("Reused values: %v", result)
}
