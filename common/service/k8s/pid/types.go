package pid

var pidCache = make(map[string]string)

func cachePutPid(containerId string, pid string) {
	pidCache[containerId] = pid
}

func cacheGetPid(containerId string) string {
	return pidCache[containerId]
}

type PidParam struct {
	Namespace            string `form:"namespace" binding:"required"`
	HostIp               string `form:"HostIp" binding:"required"`
	ContainerId          string `form:"containerId"`
	FromPodName          string `form:"podName"`       //原始pod名
	FromPodContainerName string `form:"containerName"` //原始pod container名
}

type PidResult struct {
	Pid    int `json:"pid"`
	SubPid int `json:"subPid"`
}
