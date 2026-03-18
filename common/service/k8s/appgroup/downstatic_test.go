package appgroup

import (
	"os"
	"testing"
)

func TestDownStatic(t *testing.T) {

	os.Setenv("MICROAPP_PATH", "/home/workspace/w7panel/kodata/microapp")
	fetchWebZipAndDownload("http://zpk.w7.cc/zpk/respo/info/w7_zpkv2", "w7-zpkv2", "1.0.25")
}
