package gpu

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
)

const (
	// OneContainerMultiDeviceSplitSymbol this is when one container use multi device, use : symbol to join device info.
	OneContainerMultiDeviceSplitSymbol = ":"

	// OnePodMultiContainerSplitSymbol this is when one pod having multi container and more than one container use device, use ; symbol to join device info.
	OnePodMultiContainerSplitSymbol = ";"

	// NvidiaGPUDevice     = "NVIDIA"
	AscendGPUDevice     = "Ascend"
	Ascend310PGPUDevice = "Ascend310P"
	HygonGPUDevice      = "DCU"
	CambriconGPUDevice  = "MLU"

	DsmluProfileAndInstance = "CAMBRICON_DSMLU_PROFILE_INSTANCE"

	NVIDIAPriority = "nvidia.com/priority"
)

func DecodeNodeDevices(str string) ([]*DeviceInfo, error) {
	if !strings.Contains(str, OneContainerMultiDeviceSplitSymbol) {
		slog.Warn("Node annotations not decode successfully")
		return []*DeviceInfo{}, errors.New("node annotations not decode successfully")
	}
	tmp := strings.Split(str, OneContainerMultiDeviceSplitSymbol)
	var retval []*DeviceInfo
	for _, val := range tmp {
		if strings.Contains(val, ",") {
			items := strings.Split(val, ",")
			if len(items) >= 7 || len(items) == 9 {
				count, _ := strconv.ParseInt(items[1], 10, 32)
				devmem, _ := strconv.ParseInt(items[2], 10, 32)
				devcore, _ := strconv.ParseInt(items[3], 10, 32)
				health, _ := strconv.ParseBool(items[6])
				numa, _ := strconv.Atoi(items[5])
				mode := "hami-core"
				index := 0
				if len(items) == 9 {
					index, _ = strconv.Atoi(items[7])
					mode = items[8]
				}
				i := DeviceInfo{
					ID:      items[0],
					AliasId: items[0],
					Count:   int32(count),
					Devmem:  int32(devmem),
					Devcore: int32(devcore),
					Type:    items[4],
					Numa:    numa,
					Health:  health,
					Mode:    mode,
					Index:   uint(index),
				}
				retval = append(retval, &i)
			} else {
				return []*DeviceInfo{}, errors.New("node annotations not decode successfully")
			}
		}
	}
	return retval, nil
}
