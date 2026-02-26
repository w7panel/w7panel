package shell

import (
	"encoding/json"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/gpu"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types2 "k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"
)

func Test_patchK3sGpu(t *testing.T) {
	// 创建一个假的 k8s.Sdk 实例
	sdk := k8s.NewK8sClientInner()
	// 创建一个假的 data 参数
	data := map[string]string{
		// "key1": "value1",
		"hami-mode": "1",
	}
	// 将 data 转换为 JSON
	dataBytes, err := json.Marshal(data)
	assert.NoError(t, err)
	// 创建一个假的 patchData 参数
	patchData := map[string]interface{}{
		"data": dataBytes,
	}
	// 将 patchData 转换为 JSON
	patchDataBytes, err := json.Marshal(patchData)
	assert.NoError(t, err)
	// 模拟 Patch 方法
	sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Patch(
		sdk.Ctx,
		"k3s.gpu",
		types2.MergePatchType,
		patchDataBytes,
		metav1.PatchOptions{},
	)
	// 调用 patchK3sGpu 方法
	err = gpu.PatchK3sGpu(sdk, data)
	assert.NoError(t, err)
}
