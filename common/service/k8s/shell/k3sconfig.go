package shell

import (
	"encoding/json"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type k3sConfig struct {
	k3sNodeArgs   []string
	k3sNodeEnv    map[string]string
	k3sConfigYaml map[string]interface{}
	k3sRawConfig  string
}

func NewK3sConfig(k3sNodeArgs string, k3sNodeEnv string, k3sConfigYaml []byte) *k3sConfig {
	instance := &k3sConfig{}
	json.Unmarshal([]byte(k3sNodeArgs), &instance.k3sNodeArgs)
	json.Unmarshal([]byte(k3sNodeEnv), &instance.k3sNodeEnv)
	err := yaml.Unmarshal((k3sConfigYaml), &instance.k3sConfigYaml)
	if err != nil {
		slog.Error("unmarshal k3s config error", "err", err)
		instance.k3sConfigYaml = make(map[string]interface{})
	}
	if (instance.k3sConfigYaml == nil) || len(instance.k3sConfigYaml) == 0 {
		instance.k3sConfigYaml = make(map[string]interface{})
	}
	instance.k3sRawConfig = string(k3sConfigYaml)
	return instance
}

func NewK3sConfigByNode(node *v1.Node) *k3sConfig {
	args := node.Annotations["k3s.io/node-args"]
	nodeEnv := node.Annotations["k3s.io/node-env"]
	config, err := helper.ReadK3sConfig()
	if err != nil {
		slog.Error("read k3s config error", "err", err)
		config = []byte{}
	}
	k3sConfig := NewK3sConfig(args, nodeEnv, config)
	return k3sConfig
}

func (self *k3sConfig) IsClusterInit() bool {
	slog.Debug("k3sNodeArgs", "args", self.k3sNodeArgs)
	for _, v := range self.k3sNodeArgs {
		if v == "--cluster-init" {
			return true
		}
	}
	return false
}

func (self *k3sConfig) IsNotFirstNode() bool {
	for k, _ := range self.k3sNodeEnv {
		if k == "K3S_URL" || k == "K3S_TOKEN" {
			return true
		}
	}
	for _, v := range self.k3sNodeArgs {
		if v == "--token" {
			return true
		}
	}
	return false
}

func (self *k3sConfig) IsOutDB() bool {

	for _, v := range self.k3sNodeArgs {
		if v == "--datastore-endpoint" {
			return true
		}
	}
	for k, v := range self.k3sNodeEnv {
		if k == "K3S_DATASTORE_ENDPOINT" && v != "" {
			return true
		}
	}
	return false
}

func (self *k3sConfig) IsOutDBString() string {
	if self.IsOutDB() {
		return "true"
	}
	return "false"
}

func (self *k3sConfig) DbUrl() string {
	if self.IsOutDB() {
		dburl, ok := self.k3sConfigYaml["datastore-endpoint"].(string)
		if ok {
			return dburl
		}
	}
	return ""
}

func (self *k3sConfig) IsClusterInitString() string {
	if self.IsClusterInit() {
		return "true"
	}
	return "false"
}

func (self *k3sConfig) IsOutDbInYaml() bool {
	dbConfig, ok := self.k3sConfigYaml["datastore-endpoint"]
	if ok {
		return dbConfig != ""
	}
	return false
}

func (self *k3sConfig) IsClusterInitInYaml() bool {
	clusterInit, ok := self.k3sConfigYaml["cluster-init"]
	if ok {
		return clusterInit == "true"
	}
	return false
}

func (self *k3sConfig) GetMode() string {
	if self.IsClusterInit() {
		return "2"
	}
	if self.IsOutDB() {
		return "3"
	}
	return "1"
}

func (self *k3sConfig) HasTlsSanIp(ip string) bool {
	for i := 0; i < len(self.k3sNodeArgs); i++ {
		if self.k3sNodeArgs[i] == "--tls-san" && i+1 < len(self.k3sNodeArgs) && self.k3sNodeArgs[i+1] == ip {
			return true
		}
	}
	return false
}

func (self *k3sConfig) GetConfigTlsSanIp() []string {

	data, ok := self.k3sConfigYaml["tls-san"]
	if ok {
		if slice, ok1 := data.([]interface{}); ok1 {
			var result []string
			for _, v := range slice {
				if str, ok2 := v.(string); ok2 {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

func (self *k3sConfig) GetTlsSanIp() []string {
	var ips []string
	for i := 0; i < len(self.k3sNodeArgs); i++ {
		if self.k3sNodeArgs[i] == "--tls-san" && i+1 < len(self.k3sNodeArgs) {
			ips = append(ips, self.k3sNodeArgs[i+1])
		}
	}
	return ips
}

// 默认的tls-san ip，即同时出现在k3s node args和config中的ip 说明是默认的启动参数的ip
func (self *k3sConfig) GetDefaultTlsSanIp() []string {
	full := self.GetTlsSanIp()
	configTls := self.GetConfigTlsSanIp()
	diff := helper.Difference(full, configTls)
	result := []string{}
	for _, v := range diff {
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}
