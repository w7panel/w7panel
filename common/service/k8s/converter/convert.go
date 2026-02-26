package converter

import (
	"fmt"
	"os"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

/**

vX.Y.Z - URL referenced based on the specified GitHub repository
vX.Y.Z - 根据指定的 GitHub 仓库引用的 URL
vX.Y.Z-standalone - de-referenced schemas, more useful as standalone documents
vX.Y.Z-standalone - 取消引用的架构，作为独立文档更有用
vX.Y.Z-local - relative references, useful to avoid the network dependency
vX.Y.Z-local - 相对引用，有助于避免网络依赖性
vX.Y.Z-strict - prohibits properties not defined in the schema
vX.Y.Z-strict - 禁止架构中未定义的属性

*/

func ConvertOpenApiToSchema() error {
	sdk := k8s.NewK8sClient()
	data, err := sdk.ClientSet.RESTClient().Get().AbsPath("/openapi/v2").SetHeader("Accept", "application/json").DoRaw(sdk.Ctx)
	if err != nil {
		return err
	}
	kodatapath, ok := os.LookupEnv("KO_DATA_PATH")
	if !ok {
		return fmt.Errorf("KO_DATA_PATH is not set")
	}
	converter := NewConverter(kodatapath+"/schema/", "", false, true, true, true)
	return converter.ConvertData(data)
}
