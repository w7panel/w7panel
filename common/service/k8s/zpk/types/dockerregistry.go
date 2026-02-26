package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type DockerRegistry struct {
	Host      string `json:"host"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Namespace string `json:"namespace"`
}

func (d DockerRegistry) GetDockerImage() string {
	return fmt.Sprintf("%s/%s", d.Host, d.Namespace)
}
func (d DockerRegistry) GetDockerImageWithName(name string) string {
	return fmt.Sprintf("%s/%s/%s", d.Host, d.Namespace, name)
}
func (d DockerRegistry) GetDockerImageWithNameAndTag(name string, tag string) string {
	return fmt.Sprintf("%s/%s/%s:%s", d.Host, d.Namespace, name, tag)
}
func (d DockerRegistry) GetAuthJsonString() string {

	auth := fmt.Sprintf("%s:%s", d.Username, d.Password)
	base64_encode := func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	jsonVal := map[string]interface{}{
		"auths": map[string]interface{}{
			d.Host: map[string]interface{}{
				"auth": base64_encode(auth),
			},
		},
	}
	result, err := json.Marshal(jsonVal)
	if err != nil {
		return ""
	}
	return string(result)
}
