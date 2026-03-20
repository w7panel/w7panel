package pid

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

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
	Pid      int    `json:"pid"`
	SubPid   int    `json:"subPid"`
	ProxyIp  string `json:"proxyIp"`
	AgentPod *corev1.Pod
	Pwd      string `json:"pwd"`
}

type PidCacheItem struct {
	podName     string
	namespace   string
	containerId string
	pid         int
}

func (p *PidResult) ToArray() map[string]string {
	podIp := p.ProxyIp
	pidstr := strconv.Itoa(p.Pid)
	subpidstr := strconv.Itoa(p.SubPid)
	webdavUrl := "/panel-api/v1/" + podIp + ":8000/proxy/panel-api/v1/files/webdav-agent/" + pidstr + "/agent"
	webdavBasePath := "panel-api/v1/files/webdav-agent/" + pidstr + "/agent" //前端根据这个过滤掉 当前目录?
	compressUrl := "/panel-api/v1/" + podIp + ":8000/proxy/panel-api/v1/files/compress-agent/" + pidstr
	permissionUrl := "/panel-api/v1/" + podIp + ":8000/proxy/panel-api/v1/files/permission-agent/" + pidstr
	if p.SubPid > 0 {
		webdavUrl = "/panel-api/v1/" + podIp + ":8000/proxy/panel-api/v1/files/webdav-agent/" + pidstr + "/subagent/" + subpidstr + "/agent"
		webdavBasePath = "panel-api/v1/files/webdav-agent/" + pidstr + "/subagent/" + subpidstr + "/agent"
		compressUrl = "/panel-api/v1/" + podIp + ":8000/proxy/panel-api/v1/files/compress-agent/" + pidstr + "/subagent/" + subpidstr
		permissionUrl = "/panel-api/v1/" + podIp + ":8000/proxy/panel-api/v1/files/permission-agent/" + pidstr + "/subagent/" + subpidstr
	}
	pod := p.AgentPod
	return map[string]string{
		"podName":       pod.Name,
		"pid":           pidstr,
		"subPid":        subpidstr,
		"namespace":     pod.Namespace,
		"containerName": pod.Spec.Containers[0].Name,
		"podIp":         podIp,
		// "pwd":            pwd,
		"webdavUrl": webdavUrl,
		// "webdavToken":    token,
		"webdavBasePath": webdavBasePath,
		"compressUrl":    compressUrl,
		"permissionUrl":  permissionUrl,
		"pwd":            p.Pwd,
		// "users":          users,
	}

}
