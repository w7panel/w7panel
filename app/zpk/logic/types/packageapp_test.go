package types_test

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic"
	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
)

func TestNewPackageApps(t *testing.T) {
	uri := "https://zpk.w7.cc/zpk/respo/info/w7_zpkv2"
	manifestPackage, err := logic.LoadPackage(uri)
	if err != nil {
		t.Error(err)
	}
	/**
	Identifie                string         `json:"identifie"`
	PvcName                  string         `json:"pvc_name"`
	DockerRegistry           DockerRegistry `json:"registry"`
	DockerRegistrySecretName string         `json:"docker_registry_secret_name"`
	Namespace                string         `json:"namespace"`
	InstallId                string         `json:"installId"` //安装的id
	EnvKv                    `json:"env_kv"`
	*/
	/*

	 */
	optionMain := types.InstallOption{
		Identifie: "longflow_ai",
		Namespace: "default",
		InstallId: "install-id",
		EnvKv: []types.EnvKv{
			{Name: "DOMAIN_URL", Value: "https://test.w7.cc"},
			{Name: "LANGFLOW_DATABASE_URL", Value: "postgresql://langflow:langflow@%HOST%:5432/langflow111"},
		},
	}
	optionPg := types.InstallOption{
		Identifie: "longflow_pgsql",
		Namespace: "default",
		InstallId: "install-id",
		EnvKv: []types.EnvKv{
			{Name: "POSTGRES_USER", Value: "username2"},
		},
	}
	options := []types.InstallOption{optionMain, optionPg}

	var app = types.NewPackage(manifestPackage, options, "releasename", "install-id", "default", "", "", "")
	apps := app.Children
	if len(apps) == 0 {
		t.Error("NewPackageApps should return not empty")
	}
	for _, app := range apps {
		t.Log(app.Manifest.Platform.Container.StartParams)
	}

}
