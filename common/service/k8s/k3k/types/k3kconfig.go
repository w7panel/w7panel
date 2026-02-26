package types

import corev1 "k8s.io/api/core/v1"

type K3kConfig struct {
	AllowConsoleRegister bool   `json:"allowConsoleRegister"`
	DefaultPolicyName    string `json:"defaultPolicyName"`
}

func NewK3kConfig(allowConsoleRegister bool, defaultPolicyName string) *K3kConfig {
	return &K3kConfig{
		AllowConsoleRegister: allowConsoleRegister,
		DefaultPolicyName:    defaultPolicyName,
	}
}

func NewK3kConfigBySecret(secret *corev1.Secret) *K3kConfig {
	return &K3kConfig{
		AllowConsoleRegister: string(secret.Data["allowConsoleRegister"]) == "true",
		DefaultPolicyName:    string(secret.Data["defaultPolicyName"]),
	}
}
func NewK3kConfigByConfigmap(cm *corev1.ConfigMap) *K3kConfig {
	return &K3kConfig{
		AllowConsoleRegister: string(cm.Data["allowConsoleRegister"]) == "true",
		DefaultPolicyName:    string(cm.Data["defaultPolicyName"]),
	}
}
