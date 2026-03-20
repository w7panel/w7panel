package metrics

import (
	"testing"
	"time"

	"github.com/w7panel/w7panel/common/service/k8s"
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
