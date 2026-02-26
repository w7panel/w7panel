package logic

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	helmtypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
)

func TestRequireInstall(t *testing.T) {

	// 创建一个测试用的secret
	// 调用RequireInstall方法
	err := RequireInstall("w7-mysql-fcsvhsvc", "default", "xxxx", "mydb", "u1", "p1")
	if err != nil {
		t.Fatalf("RequireInstall failed: %v", err)
	}

}

func TestRequireInstallApp(t *testing.T) {
	// apps := getReuireInstallApp()
	sdk := k8s.NewK8sClientInner()
	// install := NewInstall(sdk, *apps)
	// err := install.InstallOrUpgrade(apps.ReleaseName, "default")
	// if err != nil {
	// 	t.Error(err)
	// }

	optionMain := types.InstallOption{
		Identifie: "mysql_test1",
		Namespace: "default",
		DockerRegistry: helmtypes.DockerRegistry{
			Host:      "ccr.ccs.tencentyun.com",
			Username:  "446897682",
			Password:  "xxxxxxxx",
			Namespace: "afan-public",
		},
		PvcName:     "default-volume",
		IngressHost: "wordpress-test.ollama.cc",
		EnvKv: []types.EnvKv{
			{Name: "A", Value: "hello"},
			{Name: "B", Value: "world"},
			{Name: "MYSQL_DATABASE", Value: "aaabbb"},
		},
	}
	uri := "https://9871zpk.test.w7.com/respo/info/mysql_test1"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}

	packageApp := types.NewPackageApp(manifestPackage, &optionMain)
	rapp := NewRequireInstallJob(packageApp, sdk)
	rapp.Run()
}
