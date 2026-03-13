package internal

import (
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
)

type Config struct {
	// containerd socket 路径
	ContainerdSocket string

	// 数据目录
	DataDir string

	// Runtime 目录
	RuntimeDir string

	// Agent 目录
	AgentDir string

	// 是否启用
	Enabled bool
}

func LoadConfig() *Config {
	return &Config{
		ContainerdSocket: facade.GetConfig().GetString("k3s-registry.containerd_socket"),
		DataDir:          facade.GetConfig().GetString("k3s-registry.data_dir"),
		RuntimeDir:       facade.GetConfig().GetString("k3s-registry.runtime_dir"),
		AgentDir:         facade.GetConfig().GetString("k3s-registry.agent_dir"),
		Enabled:          facade.GetConfig().GetBool("k3s-registry.enabled"),
	}
}

// DefaultConfig 默认配置
func DefaultConfig() map[string]any {
	return map[string]any{
		"k3s-registry.enabled":           "${K3S_REGISTRY_ENABLED-false}",
		"k3s-registry.containerd_socket": "/run/k3s/containerd/containerd.sock",
		"k3s-registry.data_dir":         "/var/lib/rancher/k3s/agent/containerd",
		"k3s-registry.runtime_dir":      "/run/k3s/containerd/io.containerd.runtime.v2.task/k8s.io",
		"k3s-registry.agent_dir":        "/var/lib/rancher/k3s/agent",
	}
}
