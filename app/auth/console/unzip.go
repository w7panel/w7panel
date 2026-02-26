package console

import (
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// username password register
type Unzip struct {
	console2.Abstract
}

type unzipOption struct {
	zipPath    string
	targetPath string
	decodeGBk  bool
}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443
var unzipOp = unzipOption{}

func (c Unzip) GetName() string {
	return "unzip"
}

func (c Unzip) Configure(cmd *cobra.Command) {
	// username password register
	cmd.Flags().StringVar(&unzipOp.zipPath, "zipPath", "", "zip文件路径")
	cmd.Flags().StringVar(&unzipOp.targetPath, "targetPath", "", "目标路径")
	cmd.Flags().BoolVar(&unzipOp.decodeGBk, "decodeGBK", false, "是否解码GBK编码")
}

func (c Unzip) GetDescription() string {
	return "解压文件"
}

func (c Unzip) Handle(cmd *cobra.Command, args []string) {
	helper.Unzip(unzipOp.zipPath, unzipOp.targetPath, unzipOp.decodeGBk)
}
