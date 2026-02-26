package console

import (
	"log/slog"
	"os"
	"strconv"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// username password register
type CreateInnerDb struct {
	console2.Abstract
}

type dbOption struct {
	Database  string
	Namespace string
}

// ./runtime/main site:register --thirdPartyCDToken=qEINzTKqtPUYKi7f --host=w7job.test.w7.com --releaseName=app-nfohievs0w --deploymentName=w7-pros-28692-app-nfohievs0w --namespace=default
var dbro = dbOption{}

func (c CreateInnerDb) GetName() string {
	return "db:create-inner"
}

func (c CreateInnerDb) Configure(cmd *cobra.Command) {
	cmd.Flags().StringVar(&dbro.Database, "database", "", "数据库")
	cmd.Flags().StringVar(&dbro.Namespace, "namespace", "default", "命名空间")
}

func (c CreateInnerDb) GetDescription() string {
	return "创建数据库"
}

// go run main.go db:create-inner --database=xxx --namespace=default
func (c CreateInnerDb) Handle(cmd *cobra.Command, args []string) {
	sdk := k8s.NewK8sClientInner()
	// sdk.GetDeployment()
	list, err := sdk.GetDeploymentAppByIdentifie(dbro.Namespace, "w7-mysql")
	if err != nil {
		slog.Error("err find mysql deployment1", "err", err)
		list, err = sdk.GetDeploymentAppByIdentifie(dbro.Namespace, "w7-mysql5")
		if err != nil {
			slog.Error("err find mysql5 deployment", "err", err)
			return
		}
	}
	if len(list.Items) == 0 {
		slog.Error("err find mysql deployment")
		return
	}
	first := list.Items[0]
	container := first.Spec.Template.Spec.Containers[0]
	env := first.Spec.Template.Spec.Containers[0].Env
	userName := ""
	password := ""
	host := first.Name
	port := 3306
	for _, v1 := range container.Ports {
		port = int(v1.ContainerPort)
		break
	}
	for _, v := range env {
		if v.Name == "MYSQL_ROOT_USERNAME" {
			userName = v.Value
		}
		if v.Name == "MYSQL_ROOT_PASSWORD" {
			password = v.Value
		}
	}

	err = helper.CreateDatabase(host, strconv.Itoa(port), userName, password, dbro.Database)
	if err != nil {
		slog.Error("err create database", "err", err)
		os.Exit(0)
	}
	slog.Info("create database success")

}
