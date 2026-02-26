package gpustack

import (
	"testing"
)

func TestGetGpuStackWorkers(t *testing.T) {
	// Setup test server
	api := NewGpuStackApi("http://gstabc.b2.sz.w7.com", "admin", "tyvrmptvhv")
	// result, err := api.GetGpuStackWorkers()
	err := api.DeleteGpuStackWorkerByName("gpustack-backend-tzjdfcyo-qshtsseybg-0")
	if err != nil {
		t.Error(err)
	}
	// t.Log(result)
}
