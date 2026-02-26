package metrics

import (
	"testing"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestCgroup(t *testing.T) {

	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}
	collectReport(client)
	time.Sleep(time.Second * 3)
	collectReport(client)
}
