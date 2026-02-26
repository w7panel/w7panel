package gpu

const NVIDIA_IDENTIFIE = "nvidia-gpuoperator"

const HAMI_IDENTIFIE = "vgpu-hami"

const NVIDIA_CLASS_NAME = "nvidia"

type GpuInstallMode string

const (
	HAMI_STAT_PORT     string = "31993"
	HAMI_REALTIME_PORT string = "31992"
)

const (
	UNINSTALL     GpuInstallMode = "0"
	INSTALLING    GpuInstallMode = "1"
	INSTALLED     GpuInstallMode = "2"
	CONNOTINSTALL GpuInstallMode = "3"
)

type DeviceInfo struct {
	ID       string `json:"id"`
	AliasId  string `json:"aliasId"`
	Index    uint
	Count    int32  `json:"count"`
	Devmem   int32  `json:"devmem"`
	Devcore  int32  `json:"devcore"`
	Type     string `json:"type"`
	Numa     int    `json:"numa"`
	Mode     string `json:"mode"`
	Health   bool   `json:"health"`
	Driver   string `json:"driver"`
	NodeName string `json:"nodeName"`
}
