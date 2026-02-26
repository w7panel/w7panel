package types

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

/*
**

	$envs =  [
	          'USER_AGENT' => $buildJob->getUserAgent(),
	          'MODULE_NAME' => $buildJob->getApp()->getModuleName(),
	          'ATTACHMENT_TYPE' => $buildJob->getAttachment()->getType(),
	          'DOCKER_AUTH' => $buildJob->getDestinationImage()->getAuthJsonString(),
	          'PUSH_IMAGE' => (string)$buildJob->getDestinationImage()->getPushImage(),
	          'INSECURE' => $buildJob->getDestinationImage()->isSupportInnerHost() ? '--insecure' : '',
	          'DOWNLOAD_URL' => $this->getAttachment($buildJob)->getUrl(),
	          'NOTIFY_COMPLETION_URL' => $this->getCompletionUrl($buildJob),
	          'NOTIFY_FAILED_URL' => $this->getFailedUrl($buildJob),
	          'CURL_CA_BUNDLE' => '/kaniko/ssl/certs/ca-certificates.crt',
	          'DOCKER_FILE' => $buildJob->getDockerfile(),
	          'CONTEXT' => $buildJob->getContext()
	      ];
*/
type BuildImageInterface interface {
	GetDockerRegisty() DockerRegistry
	GetDockerfilePath() string
	GetBuildContext() string
	GetPushImage() string
	GetZipUrl() string
	GetIdentifie() string
	GetNotifyCompletionUrl() string
	GetNotifyFailedUrl() string
	GetHostNetwork() bool
	GetHostAliases() []corev1.HostAlias

	GetTitle() string

	GetBuildJobName() string

	GetLabels() map[string]string
}

type BuildInfo interface {
	GetUserAgent() string
	GetModuleName() string
	GetAttachmentType() string
	GetAuthJsonString() string
}

type BuildImageOption struct {
	BuildImageInterface
}

func NewBuildImageOption(packageApp BuildImageInterface) BuildImageOption {
	return BuildImageOption{packageApp}
}

func (d BuildImageOption) GetAttchmentType() string {
	return "zip"
}
func (d BuildImageOption) GetDockerRegisty() DockerRegistry {
	return d.BuildImageInterface.GetDockerRegisty()
}
func (d BuildImageOption) GetBuildContext() string {
	return d.BuildImageInterface.GetBuildContext()
}

func (d BuildImageOption) GetDockerfilePath() string {
	return d.BuildImageInterface.GetDockerfilePath()
}

func (d BuildImageOption) GetPushImage() string {
	return d.BuildImageInterface.GetPushImage()
}

func (d BuildImageOption) GetInsecure() string {
	if d.GetHostNetwork() {
		return "--insecure --insecure-pull"
	} else {
		return ""
	}
}

func (d BuildImageOption) GetDownloadUrl() string {
	return d.GetZipUrl()
}

func (d BuildImageOption) GetNotifyCompletionUrl() string {
	return d.BuildImageInterface.GetNotifyCompletionUrl()
}

func (d BuildImageOption) GetNotifyFailedUrl() string {
	return d.BuildImageInterface.GetNotifyFailedUrl()
}

func (d BuildImageOption) GetCurlCaBundle() string {
	return "/kaniko/ssl/certs/ca-certificates.crt"
}

func (d BuildImageOption) registryMap() string {
	return `index.docker.io=mirror.ccs.tencentyun.com;index.docker.io=registry.cn-hangzhou.aliyuncs.com;
	index.docker.io=docker.m.daocloud.io;index.docker.io=docker.1panel.live`
}

func (d BuildImageOption) ToMap() map[string]string {

	embed := "false"
	if d.IsInner() {
		embed = "true"
	}
	return map[string]string{
		"USER_AGENT":            "release",
		"MODULE_NAME":           strings.ReplaceAll(d.GetIdentifie(), "-", "_"),
		"ATTACHMENT_TYPE":       d.GetAttchmentType(),
		"DOCKER_AUTH":           d.GetDockerRegisty().GetAuthJsonString(),
		"PUSH_IMAGE":            d.GetPushImage(),
		"INSECURE":              d.GetInsecure(),
		"DOWNLOAD_URL":          d.GetDownloadUrl(),
		"NOTIFY_COMPLETION_URL": d.GetNotifyCompletionUrl(),
		"NOTIFY_FAILED_URL":     d.GetNotifyFailedUrl(),
		"CURL_CA_BUNDLE":        d.GetCurlCaBundle(),
		"CONTEXT":               d.GetBuildContext(),
		"DOCKER_FILE":           d.GetDockerfilePath(),
		"KANIKO_REGISTRY_MAP":   d.registryMap(),
		"EMBED":                 embed,
	}
}

func (d BuildImageOption) ToEnv() []corev1.EnvVar {
	var envs []corev1.EnvVar
	for k, v := range d.ToMap() {
		envVar := corev1.EnvVar{Name: k, Value: v}
		envs = append(envs, envVar)
	}
	return envs
}

func (d BuildImageOption) GetHostNetwork() bool {
	return d.BuildImageInterface.GetHostNetwork()
}

func (d BuildImageOption) GetHostPid() bool {
	return d.BuildImageInterface.GetHostNetwork()
}

func (d BuildImageOption) GetHostAliases() []corev1.HostAlias {
	return d.BuildImageInterface.GetHostAliases()
}
func (d BuildImageOption) IsInner() bool {
	return d.BuildImageInterface.GetHostNetwork()
}

func (d BuildImageOption) GetVolumes() []corev1.Volume {
	v := []corev1.Volume{}
	if !d.IsInner() {
		return v
	}
	v = append(v,
		corev1.Volume{Name: "my-host", VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{Path: "/"},
		}},
	)
	return v
}

func (d BuildImageOption) GetVolumeMounts() []corev1.VolumeMount {
	v := []corev1.VolumeMount{}
	if !d.IsInner() {
		return v
	}
	v = append(v, corev1.VolumeMount{Name: "my-host", MountPath: "/host"})
	return v
}
