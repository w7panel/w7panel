package pid

import "testing"

func TestHandle(t *testing.T) {

	containerId := "containerd://1db837210dfcfe3f069c5ca3ac44b26f90a7a704a8571ef69dc666332d68cc1e"
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
