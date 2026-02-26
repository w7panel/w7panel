// nolint
package console

import (
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

type mockW7ConfigRepository struct {
	config.W7ConfigRepositoryInterface
	getFunc func(name string) (*config.W7Config, error)
}

type mockConsoleCdClient struct {
	createLicenseSiteFunc func(data map[string]string) (*License, error)
}

func TestLicenseClient_CreateLicenseSite(t *testing.T) {
	os.Setenv("USER_AGENT", "we7test-beta")
	SetConsoleApi("http://172.16.1.150:9004")
	sdk := k8s.NewK8sClientInner()
	repo := config.NewW7ConfigRepository(sdk)
	client := NewLicenseClient(repo, sdk)
	client.SetLicense(&License{AppId: "500475", AppSecret: "3c08d42f2ff07cd420c9f1d5d1a56cc0", FounderSaName: "admin"})
	// license, err := client.CreateLicenseSite("admin", true)
	// if err != nil {
	// t.Error(err)
	// }
	// t.Log(license)
}
