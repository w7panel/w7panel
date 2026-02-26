package coredns

import (
	"bytes"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/coredns/caddy"
	"github.com/coredns/caddy/caddyfile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const filename = "Caddyfile"

type CoreDnsController struct {
	*caddy.Controller
}

func NewTestController(serverType, input string) *CoreDnsController {
	c := caddy.NewTestController(serverType, input)
	return &CoreDnsController{
		c,
	}
}

func ParseConfig() ([]caddyfile.ServerBlock, error) {
	sdk := k8s.NewK8sClient()
	cfg, err := sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Get(sdk.Ctx, "coredns-custom", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	data := cfg.Data["demo.server"]
	serverBlocks, err := caddyfile.Parse(filename, bytes.NewReader([]byte(data)), nil)
	if err != nil {
		return nil, err
	}
	return serverBlocks, nil
}

func ParseToJsonConfig() ([]byte, error) {
	sdk := k8s.NewK8sClient()
	cfg, err := sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Get(sdk.Ctx, "coredns-custom", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	data := cfg.Data["demo.server"]
	serverBlocks, err := caddyfile.ToJSON([]byte(data))
	if err != nil {
		return nil, err
	}
	return serverBlocks, nil
}

func DnsConfigToCorefile(config *DnsConfig) {
	return
}
