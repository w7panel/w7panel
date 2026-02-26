package console

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	"github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// username password register
type ResetPassword struct {
	console.Abstract
}

type resetOption struct {
	Username      string
	Password      string
	RoleName      string
	Namespace     string
	IsClusterRole bool
}

// ./runtime/main auth:register --username=hello --password=world
var ro3 = registerOption{}

func (c ResetPassword) GetName() string {
	return "auth:reset-password"
}

func (c ResetPassword) Configure(cmd *cobra.Command) {
	// username password register
	//
	cmd.Flags().StringVar(&ro3.Username, "username", "", "username")
	cmd.Flags().StringVar(&ro3.Password, "password", "", "password")
}

func (c ResetPassword) GetDescription() string {
	return "重置用户密码"
}

func (c ResetPassword) Handle(cmd *cobra.Command, args []string) {
	if len(ro3.Username) == 0 || len(ro3.Password) == 0 {
		slog.Error("username or password is empty")
		return
	}

	k8sAuth := k8s.NewK8sClient()
	if ro3.Namespace == "" {
		ro3.Namespace = k8sAuth.GetNamespace()
	}
	sa, err := k8sAuth.GetServiceAccount(ro3.Namespace, ro3.Username)
	if err != nil {
		slog.Error("err get service account", "err", err)
		return
	}
	err = k8sAuth.ResetPassword(ro3.Username, ro3.Password, sa.Labels["w7.cc/user-mode"])
	if err != nil {
		slog.Error("err register", "err", err)
	}
}
