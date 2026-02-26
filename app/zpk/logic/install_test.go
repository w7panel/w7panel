package logic

import (
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	helmtypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
)

func getApps() types.Package {
	uri := "https://zpk.w7.cc/respo/info/longflow_ai"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
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
		Identifie:   "longflow_ai",
		Namespace:   "default",
		PvcName:     "longflow-ai",
		IngressHost: "test.w7.cc",
		EnvKv: []types.EnvKv{
			{Name: "DOMAIN_URL", Value: "https://test.w7.cc"},
			{Name: "LANGFLOW_DATABASE_URL", Value: "postgresql://langflow:langflow@%HOST%:5432/langflow"},
		},
	}
	optionPg := types.InstallOption{
		Identifie: "longflow_pgsql",
		Namespace: "default",
		PvcName:   "longflow-ai",
		EnvKv: []types.EnvKv{
			{Name: "POSTGRES_USER", Value: "username2"},
		},
	}
	options := []types.InstallOption{optionMain, optionPg}

	var apps = types.NewPackage(manifestPackage, options, "install2", "install-id", "default", "", "", "")
	return apps
}

func getMysqlApps() types.Package {
	uri := "https://zpk.w7.cc/respo/info/w7_mysql"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}

	optionMain := types.InstallOption{
		Identifie: "w7_mysql",
		Namespace: "default",
		PvcName:   "local",
		// IngressHost: "test.w7.cc",
		EnvKv: []types.EnvKv{
			{Name: "MYSQL_ROOT_USERNAME", Value: "root"},
			{Name: "MYSQL_ROOT_PASSWORD", Value: "mypass234567"},
		},
	}

	options := []types.InstallOption{optionMain}
	var apps = types.NewPackage(manifestPackage, options, "mysql", "mysql-id", "default", "", "", "")
	return apps
}
func getRedisApps() types.Package {
	uri := "https://zpk.w7.cc/respo/info/w7_redis"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}

	optionMain := types.InstallOption{
		Identifie: "w7_redis",
		Namespace: "default",
		PvcName:   "local",
		// IngressHost: "test.w7.cc",
		EnvKv: []types.EnvKv{
			{Name: "REDIS_PASSWORD", Value: "mypass234567"},
		},
	}

	options := []types.InstallOption{optionMain}
	var apps = types.NewPackage(manifestPackage, options, "redis", "redis-id", "default", "", "", "")
	return apps
}

func getBuildApps() types.Package {
	//playedu
	uri := "https://zpk.w7.cc/respo/info/w7_answer"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	optionMain := types.InstallOption{
		Identifie: "w7_answer",
		Namespace: "default",
		DockerRegistry: helmtypes.DockerRegistry{
			Host:      "registry.local.w7.cc",
			Username:  "admin",
			Password:  "w7-secret",
			Namespace: "default",
		},
		ReleaseName: "test-build",
		PvcName:     "longflow-ai",
		IngressHost: "test.w7.cc",
		EnvKv:       []types.EnvKv{},
	}
	options := []types.InstallOption{optionMain}

	var apps = types.NewPackage(manifestPackage, options, "test-build", "test-build-id", "default", "", "", "")
	return apps

}

func getApps2() types.Package {
	//playedu
	uri := "https://zpk.w7.cc/respo/info/hightman_xunsearch"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	optionMain := types.InstallOption{
		Identifie: "hightman_xunsearch",
		DockerRegistry: helmtypes.DockerRegistry{
			Host:      "ccr.ccs.tencentyun.com",
			Username:  "446897682",
			Password:  "xxxxxxxx",
			Namespace: "afan-public",
		},
		PvcName:     "longflow-ai",
		IngressHost: "test.w7.cc",
		EnvKv:       []types.EnvKv{},
	}
	options := []types.InstallOption{optionMain}

	var apps = types.NewPackage(manifestPackage, options, "xunsearch", "install-id", "default", "", "", "")
	return apps

}

func getHelmApp() types.Package {
	//playedu
	uri := "https://zpk.w7.cc/respo/info/helm_dashboard"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	optionMain := types.InstallOption{
		Identifie: "helm_dashboard",
		EnvKv:     []types.EnvKv{},
	}
	options := []types.InstallOption{optionMain}

	var apps = types.NewPackage(manifestPackage, options, "helm-dashboard", "install-id", "default", "", "", "")
	return apps

}

func getHelmZipApp() types.Package {
	//playedu
	uri := "https://zpk.w7.cc/respo/info/helm_nginx"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	optionMain := types.InstallOption{
		Identifie: "helm_nginx",
		EnvKv: []types.EnvKv{
			{Name: "image.tag", Value: "v2.1.0"},
		},
	}
	options := []types.InstallOption{optionMain}

	var apps = types.NewPackage(manifestPackage, options, "helm-nginx", "install-id", "default", "", "", "")
	return apps

}

func getIngressesApp() types.Package {
	//playedu
	uri := "https://zpk.w7.cc/respo/info/ai_ollamaui"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	optionMain := types.InstallOption{
		Identifie: "ai_ollamaui",
		PvcName:   "longflow-ai",
		EnvKv: []types.EnvKv{
			{Name: "image.tag", Value: "v2.1.0"},
		},
	}
	option2 := types.InstallOption{
		Identifie: "ai_ollamaapi",
		PvcName:   "longflow-ai",
		EnvKv: []types.EnvKv{
			{Name: "image.tag", Value: "v2.1.0"},
		},
	}
	options := []types.InstallOption{optionMain, option2}

	var apps = types.NewPackage(manifestPackage, options, "ai-ollamaui", "install-id", "default", "ollama.cc", "ollama", "traefik")
	return apps

}

func TestInstall_Install(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	apps := getApps()
	install := NewInstall(sdk, apps)
	err := install.Install("install2", "default")
	if err != nil {
		t.Error(err)
	}

}

func TestInstall_Upgrade(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	apps := getApps()
	install := NewInstall(sdk, apps)
	err := install.Upgrade("install2", "default")

	if err != nil {
		t.Error(err)
	}

}

func TestInstall_Build(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	apps := getBuildApps()
	install := NewInstall(sdk, apps)
	err := install.InstallOrUpgrade("test-build", "default")
	if err != nil {
		t.Error(err)
	}

}

func TestInstall2(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	apps := getApps2()
	install := NewInstall(sdk, apps)

	err := install.InstallOrUpgrade("xunsearch", "default")
	if err != nil {
		t.Error(err)
	}
}

func TestInstallMysql(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	apps := getMysqlApps()
	install := NewInstall(sdk, apps)

	err := install.InstallOrUpgrade("mysql", "default")
	if err != nil {
		t.Error(err)
	}
}

func TestInstallRedis(t *testing.T) {
	sdk := k8s.NewK8sClientInner()

	apps := getRedisApps()
	install := NewInstall(sdk, apps)

	err := install.InstallOrUpgrade("redis", "default")
	if err != nil {
		t.Error(err)
	}
}

func TestInstallHelm(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	apps := getHelmApp()

	install := NewInstall(sdk, apps)

	err := install.InstallOrUpgrade("helm-dashboard", "default")
	if err != nil {
		t.Error(err)
	}
}

func TestInstallHelmZip(t *testing.T) {
	sdk := k8s.NewK8sClientInner()
	apps := getHelmZipApp()

	install := NewInstall(sdk, apps)

	err := install.InstallOrUpgrade("helm-nginx", "default")
	if err != nil {
		t.Error(err)
	}
}

func TestIngressesApps(t *testing.T) {
	// os.Setenv("KUBERNETES_MASTER", "https://172.16.1.13:6443")
	// os.Setenv("KUBERNETES_SERVICE_HOST", "172.16.1.13")
	// os.Setenv("KUBERNETES_SERVICE_PORT", "6443")
	// token := "eyJhbGciOiJSUzI1NiIsImtpZCI6Iml3WktfcGw1VkMzdHdMei1SMjNpMzhWelQ2V0Iyb3FLSUd4RlFaWlMxbncifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzI2ODI0NjYwLCJpYXQiOjE3MjY4MjEwNjAsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiZGFiMjYxY2ItNGNhMS00YTM1LWIzYzItNjUxZTJmNGMxNjYwIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImFkbWluIiwidWlkIjoiNWJlYmIzYjUtOGVjNS00NWJhLTgwYmMtOTg4OWRjZmU0ZGYwIn19LCJuYmYiOjE3MjY4MjEwNjAsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmFkbWluIn0.TpwQWWQ7OFKmTtD7coxWWjbgpUmv4GCIE4CoRsz2bFd7yuZ5ABJ_gS2BETwyQeoXot8oaXpJpPhGbCgNXReNYir9iWJpEUPytMOr3mgJf_W0nccs6HE_KGMQ5hXqi_ZEEJ1puCQXgU4CtB1n_TlYenhzEDOWk9T9yx_P9cQm2eCqHra4_Gd9NZu5xOInJyWdn-RDH2ABuM6v1emieztcyKacT4kyTMaCB_RiRS3F4uKzKL6y9rVRv92Gs1p4mZywwKlUNr9WoqPBE07cNbxhOx9FJIC8PbJshh9oH2BP0il04jG2reDcGwX6mxEbr_DkdF6cHZhdn8uPMadeEzbXWQ"
	// sdk, err := k8s.NewK8sClient().Channel(token)
	// if err != nil {
	// 	t.Error(err)
	// }

	sdk := k8s.NewK8sClientInner()

	// helmApi := k8s.NewHelm(sdk)

	apps := getIngressesApp()
	install := NewInstall(sdk, apps)
	err := install.InstallOrUpgrade("ai-ollamaui", "default")
	if err != nil {
		t.Error(err)
	}
}

func getConsoleApps() *types.Package {
	//playedu
	uri := "deploy://console/4727/"
	// uri := "deploy://console/97912"
	token := "qEINzTKqtPUYKi7f"
	repo := NewRepo(uri, token, "")
	manifestPackage, err := repo.Load()
	if err != nil {
		panic(err)
	}
	// print(manifestPackage.Manifest.RequireSite())
	// return nil

	preInstall, err := repo.PreInstall("5723")
	if err != nil {
		panic(err)
	}
	// manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	optionMain := types.InstallOption{
		Identifie: "w7_pros_28692",
		Namespace: "default",
		DockerRegistry: helmtypes.DockerRegistry{
			Host:      "ccr.ccs.tencentyun.com",
			Username:  "446897682",
			Password:  "Fykdmx521",
			Namespace: "afan-public",
		},
		DockerRegistrySecretName: "ccr.ccs.tencentyun.com.default",
		ReleaseName:              preInstall.ReleaseName,
		PvcName:                  "default-volume",
		IngressHost:              "w7job4.test.w7.com",
		EnvKv:                    []types.EnvKv{},
	}
	options := []types.InstallOption{optionMain}

	var apps = types.NewPackage(manifestPackage, options, preInstall.ReleaseName, "install2", "default", "w7job5.test.w7.com", "", "")
	apps.ReleaseName = preInstall.ReleaseName
	apps.Root.ZipUrl = preInstall.ZipURL
	apps.Root.ServiceAccountName = "w7"
	apps.Root.ThirdpartyCDToken = "qEINzTKqtPUYKi7f"
	// cdClient := console.NewConsoleCdClient(token)
	// appSecret, err := cdClient.CreateSite("hellosssddxxx.cc", preInstall.ReleaseName)
	// if err != nil {
	// 	slog.Warn("create site error may not need secret", "err", err)
	// }

	// if appSecret != nil {
	// 	appId := types.Env{}
	// 	appId.Name = "APP_ID"
	// 	appId.Value = appSecret.AppId
	// 	secret := types.Env{}
	// 	secret.Name = "APP_SECRET"
	// 	secret.Value = appSecret.AppSecret
	// 	apps.Root.Manifest.Platform.Container.Env = append(apps.Root.Manifest.Platform.Container.Env, appId, secret)
	// }
	return &apps

}

func TestConsoleApp(t *testing.T) {
	// os.Setenv("KUBERNETES_MASTER", "https://172.16.1.13:6443")
	// os.Setenv("KUBERNETES_SERVICE_HOST", "172.16.1.13")
	// os.Setenv("KUBERNETES_SERVICE_PORT", "6443")
	// token := "eyJhbGciOiJSUzI1NiIsImtpZCI6Iml3WktfcGw1VkMzdHdMei1SMjNpMzhWelQ2V0Iyb3FLSUd4RlFaWlMxbncifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzI2ODI0NjYwLCJpYXQiOjE3MjY4MjEwNjAsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiZGFiMjYxY2ItNGNhMS00YTM1LWIzYzItNjUxZTJmNGMxNjYwIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImFkbWluIiwidWlkIjoiNWJlYmIzYjUtOGVjNS00NWJhLTgwYmMtOTg4OWRjZmU0ZGYwIn19LCJuYmYiOjE3MjY4MjEwNjAsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmFkbWluIn0.TpwQWWQ7OFKmTtD7coxWWjbgpUmv4GCIE4CoRsz2bFd7yuZ5ABJ_gS2BETwyQeoXot8oaXpJpPhGbCgNXReNYir9iWJpEUPytMOr3mgJf_W0nccs6HE_KGMQ5hXqi_ZEEJ1puCQXgU4CtB1n_TlYenhzEDOWk9T9yx_P9cQm2eCqHra4_Gd9NZu5xOInJyWdn-RDH2ABuM6v1emieztcyKacT4kyTMaCB_RiRS3F4uKzKL6y9rVRv92Gs1p4mZywwKlUNr9WoqPBE07cNbxhOx9FJIC8PbJshh9oH2BP0il04jG2reDcGwX6mxEbr_DkdF6cHZhdn8uPMadeEzbXWQ"
	// sdk, err := k8s.NewK8sClient().Channel(token)
	// if err != nil {
	// 	t.Error(err)
	// }
	// os.Setenv("USER_AGENT", "we7test-beta")
	sdk := k8s.NewK8sClientInner()

	// helmApi := k8s.NewHelm(sdk)

	apps := getConsoleApps()
	// getConsoleApps()

	// completeUrl := "http://127.0.0.1:9007/api/v1/zpk/?namespace=" + params.Namespace + "&releaseName=" + releaseName +
	// 	"&domainUrl=" + params.IngressHost + "&deploymentName=" + packageApps.Root.GetName() + "&thirdpartyCDToken=" + params.ThirdpartyCDToken + "&api-token=" + token

	// packageApps.Root.InstallOption.BuildImageSuccessUrl = completeUrl

	install := NewInstall(sdk, *apps)
	err := install.InstallOrUpgrade(apps.ReleaseName, "default")
	if err != nil {
		t.Error(err)
	}
}

func getServiceLbApp() types.Package {
	//playedu
	uri := "https://zpk.w7.cc/respo/info/k8s_offline"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	optionMain := types.InstallOption{
		Identifie: "k8s_offline",
		PvcName:   "longflow-ai",
		EnvKv: []types.EnvKv{
			{Name: "image.tag", Value: "v2.1.0"},
		},
	}
	// option2 := types.InstallOption{
	// 	Identifie: "ai_ollamaapi",
	// 	PvcName:   "longflow-ai",
	// 	EnvKv: []types.EnvKv{
	// 		{Name: "image.tag", Value: "v2.1.0"},
	// 	},
	// }
	options := []types.InstallOption{optionMain}

	var apps = types.NewPackage(manifestPackage, options, "ai-ollamaui", "install-id", "default", "ollama.cc", "ollama", "traefik")
	return apps

}
func getWordpress() *types.Package {
	//playedu
	uri := "https://zpk.w7.cc/respo/info/w7_wordpress"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	optionMain := types.InstallOption{
		Identifie: "w7_wordpress",
		DockerRegistry: helmtypes.DockerRegistry{
			Host:      "ccr.ccs.tencentyun.com",
			Username:  "446897682",
			Password:  "xxxxxxxx",
			Namespace: "afan-public",
		},
		PvcName:     "default-volume",
		IngressHost: "wordpress-test.ollama.cc",
		EnvKv:       []types.EnvKv{},
	}
	options := []types.InstallOption{optionMain}

	var apps = types.NewPackage(manifestPackage, options, "wd", "install-id2", "default", "", "", "")
	return &apps

}
func TestWordpress(t *testing.T) {
	apps := getWordpress()
	sdk := k8s.NewK8sClientInner()
	install := NewInstall(sdk, *apps)
	err := install.InstallOrUpgrade(apps.ReleaseName, "default")
	if err != nil {
		t.Error(err)
	}
}

func getReuireInstallApp() *types.Package {
	//playedu
	uri := "https://9871zpk.test.w7.com/respo/info/mysql_test"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	optionMain := types.InstallOption{
		Identifie: "require_test",
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
	options := []types.InstallOption{optionMain}

	var apps = types.NewPackage(manifestPackage, options, "wd", "install-id2", "default", "", "", "")
	return &apps

}
func TestRequireInstallApp1(t *testing.T) {
	// apps := getReuireInstallApp()
	sdk := k8s.NewK8sClientInner()
	// install := NewInstall(sdk, *apps)
	// err := install.InstallOrUpgrade(apps.ReleaseName, "default")
	// if err != nil {
	// 	t.Error(err)
	// }
	optionMain := types.InstallOption{
		Identifie: "mysql_test",
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
	uri := "https://9871zpk.test.w7.com/respo/info/mysql_test"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		panic(err)
	}
	packageApp := types.NewPackageApp(manifestPackage, &optionMain)
	rapp := NewRequireInstallJob(packageApp, sdk)
	rapp.Run()
}

func TestLoadConsole(t *testing.T) {
	os.Setenv("USER_AGENT", "we7test-beta")
	uri := "deploy://console/133802/"
	// uri := "deploy://console/97912"
	token := "4JeklLWUzPl0w6MI"
	repo := NewRepo(uri, token, "")
	_, err := repo.Load()
	if err != nil {
		panic(err)
	}
}
