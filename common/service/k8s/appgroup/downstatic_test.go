package appgroup

import (
	"os"
	"testing"

	"github.com/w7panel/w7panel/common/service/k8s"
)

func TestDownStatic(t *testing.T) {

	os.Setenv("MICROAPP_PATH", "/home/workspace/w7panel/kodata/microapp")
	fetchWebZipAndDownload("http://zpk.w7.cc/zpk/respo/info/w7_zpkv2", "w7-zpkv2", "1.0.25")
	// kName = "w7_sitemanager"
	// cacheKey := staticDownloadCacheKey + kName + "" + version
}

func TestDownGroup(t *testing.T) {
	os.Setenv("STATIC_DOWN_ENABLED", "true")
	os.Setenv("MICROAPP_PATH", "/home/workspace/w7panel/kodata/microapp")
	appgroupObj, err := GetAppgroupUseSdk("w7-sitemanager-ipjjizit", "default", k8s.NewK8sClient().Sdk)
	if err != nil {
		t.Error(err)
	}
	DownStatic(appgroupObj)
	DownStaticStatus("w7-sitemanager", "1.0.25", "w7-sitemanager-ipjjizit")
}
