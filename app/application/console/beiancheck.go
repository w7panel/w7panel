package console

import (
	"log/slog"
	"os"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

type BeianCheck struct {
	console2.Abstract
}
type hostOption struct {
	host string
}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443
var hostOp = hostOption{}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443

func (c BeianCheck) GetName() string {
	return "beian-check"
}

func (c BeianCheck) Configure(cmd *cobra.Command) {
	// username password register
	cmd.Flags().StringVar(&hostOp.host, "host", "", "域名")

}

func (c BeianCheck) GetDescription() string {
	return "benan check"
}

func (c BeianCheck) Handle(cmd *cobra.Command, args []string) {
	if hostOp.host == "" {
		slog.Error("host is empty")
		os.Exit(1)
		return
	}
	err := higress.CheckHost(hostOp.host)
	if err != nil {
		slog.Error("域名未备案", "err", err)
		os.Exit(1)
		return
	}
	slog.Info("域名已备案")
	os.Exit(0)
}
