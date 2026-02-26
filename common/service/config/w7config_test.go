package config

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestW7ConfigRepository_Get(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	repo := NewW7ConfigRepository(sdk)
	config, err := repo.Get("console-75780")
	if err != nil {
		t.Errorf("Get() error = %v, wantErr %v", err, nil)
		return
	}
	// if config.ThirdpartyCDToken != "test-token" {
	// 	t.Errorf("Get() ThirdpartyCDToken = %v, want %v", config.ThirdpartyCDToken, "test-token")
	// }
	// if config.AccessToken != "test-access-token" {
	// 	t.Errorf("Get() AccessToken = %v, want %v", config.AccessToken, "test-access-token")
	// }
	// if config.ExpireTime != 1678909090 {
	// 	t.Errorf("Get() ExpireTime = %v, want %v", config.ExpireTime, 1678909090)
	// }
	if config.ClusterId != "test-cluster-id" {
		t.Errorf("Get() ClusterId = %v, want %v", config.ClusterId, "test-cluster-id")
	}
}
