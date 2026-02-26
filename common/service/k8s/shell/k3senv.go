package shell

import (
	"log/slog"
	"os"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// SecretTypeOpaque is the default. Arbitrary user-defined data
	SecretTypeK3sEnv v1.SecretType = "k3s.env"
)

func (s *k3sConfigController) initK3sEnvSecret(node *v1.Node) error {
	slog.Debug("init k3s env secret")
	if !isCurrentDaemonsetNode(node) {
		slog.Debug("not current daemonset node", "nodename", node.Name)
		return nil
	}
	// 读取配置文件内容
	configPath := AGENT_ENV_FILE
	if isControlNode(node) {
		configPath = SERVRE_ENV_FILE
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("read k3s env file error", "error", err)
		return err
	}

	// 解析配置文件内容为key-value格式
	envMap := make(map[string][]byte)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envMap[key] = []byte(value)
	}

	secretName := "k3s.env." + node.Name
	secret, err := s.Sdk.ClientSet.CoreV1().Secrets("kube-system").Get(s.Ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			slog.Error("get k3s env secret error", "error", err)
			return err
		}
		secretCreate := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: "kube-system",
				Labels: map[string]string{
					"node-name":       node.Name,
					"last-modify":     "k3s",
					"k3s-config-type": "env",
				},
			},
			Data: envMap,
			Type: SecretTypeK3sEnv,
		}
		_, err := s.Sdk.ClientSet.CoreV1().Secrets("kube-system").Create(s.Ctx, secretCreate, metav1.CreateOptions{})
		if err != nil {
			slog.Error("create k3s env secret error", "error", err)
			return err
		}
		return nil
	}

	// 更新已存在的 Secret
	secret.Data = envMap
	secret.Labels["last-modify"] = "k3s"
	secret.Labels["k3s-config-type"] = "env"
	secret.Labels["node-name"] = node.Name
	secret.Type = SecretTypeK3sEnv
	_, err = s.Sdk.ClientSet.CoreV1().Secrets("kube-system").Update(s.Ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		slog.Error("update k3s env secret error", "error", err)
		return err
	}
	return nil
}

func (s *k3sConfigController) handleK3sEnvSecret(secret *v1.Secret) error {

	if (secret.Type != SecretTypeK3sEnv) || (secret.Namespace != "kube-system") {
		slog.Debug("secret not k3s env", "name", secret.Name, "type", secret.Type, "namespace", secret.Namespace)
		return nil
	}
	modify, ok := secret.Labels["last-modify"]
	if !ok {
		slog.Debug("secret not found last-modify", "name", secret.Name)
		return nil
	}
	nodeName, ok := secret.Labels["node-name"]
	if !ok {
		slog.Debug("secret not found node name", "name", secret.Name)
		return nil
	}
	if modify == "k3s" {
		slog.Debug("secret is k3s modify", "name", secret.Name)
		return nil
	}
	node, err := s.nodeLister.Get(nodeName)
	if err != nil {
		slog.Debug("secret not found node", "name", secret.Name)
		return err
	}
	if node.Name == nodeName && isCurrentDaemonsetNode(node) {
		configPath := AGENT_ENV_FILE
		if isControlNode(node) {
			configPath = SERVRE_ENV_FILE
		}
		// if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		// 	slog.Debug("k3s env file exist", "path", configPath)
		// 	return nil
		// }
		filedata := ""
		for k, v := range secret.Data {
			filedata += k + "=" + string(v) + "\n"
		}
		err := helper.WriteFileAtomic(configPath, []byte(filedata))
		// err := os.WriteFile(configPath, []byte(filedata), 0644)
		if err != nil {
			return err
		}
		slog.Debug("restart node", "name", nodeName)
		return s.restartNode(node)
	}

	return nil

}
