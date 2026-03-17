package console

import (
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestPath(t *testing.T) {
	os.Setenv("USER_AGFNT", "we7test-beta")
	os.Setenv("DEBUG", "true")
	err := PatchAppId(k8s.NewK8sClient().Sdk, &AppSecret{AppId: "1", AppSecret: "2"}, "cs-zyy-rxnnhxjl", "default", "cs-zyy")
	if err != nil {
		t.Errorf("PatchAppId() error = %v", err)
	}
}

// 修改原函数以接受接口作为参数进行测试
