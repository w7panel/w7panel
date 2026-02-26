package appgroup

import (
	"os"
	"testing"
)

func TestDownStatic(t *testing.T) {

	os.Setenv("MICROAPP_PATH", "/home/workspace/k8s-offline/kodata/microapp")
	fetchWebZipAndDownload("https://zpk.w7.cc/zpk/respo/info/w7_cdncache", "w7-cdncache-xraavren")
}
