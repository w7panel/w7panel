package console

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DomainParseConfig struct {
	console2.Abstract
}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443

func (c DomainParseConfig) GetName() string {
	return "domain-config"
}

func (c DomainParseConfig) Configure(cmd *cobra.Command) {
	// username password register

}

func (c DomainParseConfig) GetDescription() string {
	return "fix domain config"
}

// Handle 处理域名解析配置命令，获取当前IP地址并创建或更新default命名空间下的domain-parse ConfigMap
// 如果ConfigMap不存在，则创建包含当前IP地址的新ConfigMap
// 如果ConfigMap已存在，则不执行任何操作
func (c DomainParseConfig) Handle(cmd *cobra.Command, args []string) {
	ip, err := helper.MyIp()
	if err != nil {
		slog.Error("Failed to get ip", "error", err)
		return
	}
	sdk := k8s.NewK8sClient()
	_, err = sdk.ClientSet.CoreV1().ConfigMaps("default").Get(context.Background(), "domain-parse", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			configmap := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "domain-parse",
				},
				Data: map[string]string{
					"cname": "",
					"ips":   ip,
					"type":  "A",
				},
			}
			sdk.ClientSet.CoreV1().ConfigMaps("default").Create(context.Background(), configmap, metav1.CreateOptions{})
		}
	}
}
