package higress

import (
	"errors"
	"log/slog"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress/client/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

var bfck = false

// func init() {
// 	go LoadBkConfig()
// }

// 是否需要备案检测
func NeedCheckBeian() bool {
	return bfck
}

func LoadBkConfig() (*v1alpha1.WasmPlugin, error) {
	sdk := k8s.NewK8sClient()
	client, err := sdk.ToSigClient()
	if err != nil {
		return nil, err
	}
	result := &v1alpha1.WasmPlugin{}
	err = client.Get(sdk.Ctx, types.NamespacedName{Name: "w7-white-domain", Namespace: "higress-system"}, result)
	if err != nil {
		return nil, err
	}
	WebhookWasmPlugin(result)
	return result, nil
}

func CheckHost(host string) error {
	plugin, err := LoadBkConfig()
	if err != nil {
		slog.Error("load bk config error", "error", err)
	}
	config := plugin.Spec.DefaultConfig
	val, ok := config.Fields["white_domains"]
	if !ok {
		return nil
	}
	vals := val.GetListValue().Values
	allowDomains := []string{}
	for _, v := range vals {
		structVal := v.GetStructValue()
		if structVal.Fields["enable"].GetBoolValue() {
			allowDomains = append(allowDomains, structVal.Fields["domain"].GetStringValue())
		}
	}
	//allowDomains w7.com w7.net 这种域名 fix containsAny
	// allowDomains = append(allowDomains, strings.Split(host, ".")...)

	for _, allowDomain := range allowDomains {
		if strings.Contains(host, allowDomain) {
			return nil
		}
	}
	return errors.New("host not in white list")
}

func WebhookWasmPlugin(wasmPlugin *v1alpha1.WasmPlugin) error {
	// wasmPlugin.Spec.DefaultConfig
	if wasmPlugin.Name == "w7-white-domain" && wasmPlugin.Namespace == "higress-system" {
		bfck = !wasmPlugin.Spec.DefaultConfigDisable
	}
	return nil

}
