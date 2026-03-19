package pid

import "testing"

func TestHandle(t *testing.T) {

	containerId := "containerd://7cf41ad10677cdd49ef421acef26eda6bd9e2b17caedb0b8d1f3bfecdf90e1bb"
	pidParams := PidParam{
		Namespace:            "default",
		HostIp:               "10.42.0.211",
		ContainerId:          containerId,
		FromPodName:          "w7-sitemanager-fwxjipou-site-manager-585499569d-8fkvv",
		FromPodContainerName: "site-manager",
	}

	pidObj, err := NewPidTest("console-164315")
	if err != nil {
		t.Error(err)
		return
	}
	result, err := pidObj.Handle(pidParams)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(result)

}
