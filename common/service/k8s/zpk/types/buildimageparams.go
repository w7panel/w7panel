package types

import corev1 "k8s.io/api/core/v1"

type BuildImageParams struct {
	DockerRegistry           DockerRegistry     `json:"dockerRegistry" `
	DockerfilePath           string             `json:"dockerfilePath"`
	BuildContext             string             `json:"buildContext"`
	PushImage                string             `json:"pushImage"`
	ZipUrl                   string             `json:"zipUrl"`
	Identifie                string             `json:"identifie"`
	NotifyCompletionUrl      string             `json:"notifyCompletionUrl"`
	NotifyFailedUrl          string             `json:"notifyFailedUrl"`
	HostNetwork              bool               `json:"hostNetwork"`
	HostAliases              []corev1.HostAlias `json:"hostAliases"`
	Title                    string             `json:"title"`
	Labels                   map[string]string  `json:"labels"`
	BuildJobName             string             `json:"buildJobName"`
	Schedule                 string             `json:"schedule"`
	DockerRegistrySecretName string             `json:"dockerRegistrySecretName"`
	// PushImage                string             `json:"pushImage"`
}

func NewBuildImageParams(dockerRegistry DockerRegistry, dockerfilePath string, buildContext string, pushImage string, zipUrl string, identifie string, notifyCompletionUrl string, hostNetwork bool, hostAliases []corev1.HostAlias, title string, labels map[string]string, buildJobName string, notifyFailUrl string) *BuildImageParams {
	return &BuildImageParams{
		DockerRegistry:      dockerRegistry,
		DockerfilePath:      dockerfilePath,
		BuildContext:        buildContext,
		PushImage:           pushImage,
		ZipUrl:              zipUrl,
		Identifie:           identifie,
		NotifyCompletionUrl: notifyCompletionUrl,
		NotifyFailedUrl:     notifyFailUrl,
		HostNetwork:         hostNetwork,
		HostAliases:         hostAliases,
		Title:               title,
		Labels:              labels,
		BuildJobName:        buildJobName,
	}
}

func (b *BuildImageParams) GetDockerRegisty() DockerRegistry {
	return b.DockerRegistry
}

func (m *BuildImageParams) GetBuildContext() string {
	return m.BuildContext
}

func (m *BuildImageParams) GetDockerfile() string {
	return m.DockerfilePath
}

// GetDockerfilePath 返回 Dockerfile 的完整路径，路径由构建上下文路径和 "/Dockerfile" 拼接而成。
// 返回值是字符串类型，表示 Dockerfile 的绝对路径。
func (m *BuildImageParams) GetDockerfilePath() string {
	return m.GetBuildContext() + m.GetDockerfile()
}

func (b *BuildImageParams) GetPushImage() string {
	return b.PushImage
}

func (b *BuildImageParams) GetZipUrl() string {
	return b.ZipUrl
}

func (b *BuildImageParams) GetIdentifie() string {
	return b.Identifie
}

func (b *BuildImageParams) GetNotifyCompletionUrl() string {
	return b.NotifyCompletionUrl
}

func (b *BuildImageParams) GetNotifyFailedUrl() string {
	return b.NotifyFailedUrl
}

func (b *BuildImageParams) GetHostNetwork() bool {
	return b.HostNetwork
}

func (b *BuildImageParams) GetHostAliases() []corev1.HostAlias {
	return b.HostAliases
}

func (b *BuildImageParams) GetTitle() string {
	return b.Title
}
func (b *BuildImageParams) GetLabels() map[string]string {
	return b.Labels
}
func (b *BuildImageParams) GetBuildJobName() string {
	return b.BuildJobName
}
