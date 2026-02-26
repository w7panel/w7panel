package console

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	"github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// username password register
type Register struct {
	console.Abstract
}

type registerOption struct {
	Username      string
	Password      string
	RoleName      string
	Namespace     string
	IsClusterRole bool
	UserMode      string
}

// ./runtime/main auth:register --username=hello --password=world
var ro = registerOption{}

func (c Register) GetName() string {
	return "auth:register"
}

func (c Register) Configure(cmd *cobra.Command) {
	// username password register
	//
	cmd.Flags().StringVar(&ro.Username, "username", "", "username")
	cmd.Flags().StringVar(&ro.Password, "password", "", "password")
	cmd.Flags().BoolVar(&ro.IsClusterRole, "is-cluster-role", true, "是否集群角色")
	cmd.Flags().StringVar(&ro.RoleName, "rolename", "cluster-admin", "角色")
	cmd.Flags().StringVar(&ro.Namespace, "namespace", "", "命名空间")
	cmd.Flags().StringVar(&ro.UserMode, "usermode", "founder", "用户模式")
}

func (c Register) GetDescription() string {
	return "添加用户"
}

func (c Register) Handle(cmd *cobra.Command, args []string) {
	if len(ro.Username) == 0 || len(ro.Password) == 0 {
		slog.Error("username or password is empty")
		return
	}

	k8sAuth := k8s.NewK8sClient()
	if ro.Namespace == "" {
		ro.Namespace = k8sAuth.GetNamespace()
	}
	slog.Info("register", "username", ro.Username, "password", ro.Password)
	err := k8sAuth.Register(ro.Username, ro.Password, ro.Namespace, ro.RoleName, ro.IsClusterRole, ro.UserMode)
	if err != nil {
		slog.Error("err register", "err", err)
	}
}
