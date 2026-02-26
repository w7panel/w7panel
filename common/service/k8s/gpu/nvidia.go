package gpu

import (
	"log/slog"

	corev1 "k8s.io/api/core/v1"
)

const (
	HandshakeAnnos      = "hami.io/node-handshake"
	RegisterAnnos       = "hami.io/node-nvidia-register"
	NvidiaGPUDevice     = "NVIDIA"
	NvidiaGPUCommonWord = "GPU"
	GPUInUse            = "nvidia.com/use-gputype"
	GPUNoUse            = "nvidia.com/nouse-gputype"
	NumaBind            = "nvidia.com/numa-bind"
	NodeLockNvidia      = "hami.io/mutex.lock"
	// GPUUseUUID is user can use specify GPU device for set GPU UUID.
	GPUUseUUID = "nvidia.com/use-gpuuuid"
	// GPUNoUseUUID is user can not use specify GPU device for set GPU UUID.
	GPUNoUseUUID = "nvidia.com/nouse-gpuuuid"

	InRequestDevicesAnnos = "hami.io/vgpu-devices-to-allocate"
	SupportDevicesAnnos   = "hami.io/vgpu-devices-allocated"
)

func FetchDevices(node *corev1.Node) ([]*DeviceInfo, error) {
	var err error
	var deviceInfos []*DeviceInfo

	deviceEncode, ok := node.Annotations[RegisterAnnos]
	if !ok {
		slog.Warn("%s node cloud not get hami.io/node-nvidia-register annotation", "node", node.Name)
		return deviceInfos, nil
	}
	deviceInfos, err = DecodeNodeDevices(deviceEncode)
	return deviceInfos, err
}
