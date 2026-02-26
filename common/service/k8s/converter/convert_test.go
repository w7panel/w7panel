package converter

import (
	"os"
	"testing"
)

func TestConvertOpenApiToSchema(t *testing.T) {
	// 设置环境变量 KO_DATA_PATH
	os.Setenv("KO_DATA_PATH", "/home/workspace/k8s-offline/ko-data/")
	defer os.Unsetenv("KO_DATA_PATH")
	// 调用被测方法
	err := ConvertOpenApiToSchema()
	if err != nil {
		t.Errorf("ConvertOpenApiToSchema failed: %v", err)
	}
}
