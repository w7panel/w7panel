package console

import (
	"io"
	"log"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/shell"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

type Goshell struct {
	console2.Abstract
}

type shellOption struct {
	cmd     string
	srcPath string
	toPath  string
	pid     string
	subPid  string
}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443
var shOp = shellOption{}

func (c Goshell) GetName() string {
	return "sh"
}

func (c Goshell) Configure(cmd *cobra.Command) {
	// username password register
	cmd.Flags().StringVar(&shOp.cmd, "cmd", "", "动作")
	cmd.Flags().StringVar(&shOp.srcPath, "srcPath", "", "原始路径")
	cmd.Flags().StringVar(&shOp.toPath, "toPath", "", "目标路径")
	cmd.Flags().StringVar(&shOp.pid, "pid", "", "pid")
	cmd.Flags().StringVar(&shOp.subPid, "subPid", "", "subPid")
}

func (c Goshell) GetDescription() string {
	return "执行shell命令"
}

func (c Goshell) Handle(cmd *cobra.Command, args []string) {
	// 禁用日志输出
	log.SetOutput(io.Discard)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	slog.SetDefault(logger)
	slog.Error("执行命令", "cmd", shOp.cmd, "srcPath", shOp.srcPath, "toPath", shOp.toPath, "pid", shOp.pid, "subPid", shOp.subPid)

	podShell := shell.NewPodShell(shOp.cmd, shOp.srcPath, shOp.toPath, args, shOp.pid, shOp.subPid)
	err := podShell.Run()
	if err != nil {
		slog.Error("执行失败", "error", err)
	}
	// slog.Info("执行完成" + args[0])
}
