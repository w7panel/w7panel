// nolint
package console

import (
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestConsoleClient_RegisterUseCdToken(t *testing.T) {

	os.Setenv("LOCAL_MOCK", "true")
	SetConsoleApi("http://172.16.1.116:9004")
	sdk := k8s.NewK8sClientInner()
	config2, _ := sdk.ToKubeconfig("https://218.23.2.55:6443")
	consoleClient := NewClusterClient(config.NewW7ConfigRepository(sdk), sdk, config2)
	// consoleClient.thirdPartyCDToken = "test-token"
	consoleClient.SetOfflineUrl("http://218.23.2.55:9090")

	err := consoleClient.RegisterUseCdToken(true, "s61")
	if err != nil {
		t.Errorf("RegisterUseCdToken() error = %v, wantErr %v", err, nil)
	}
}

func TestRefreshToken(t *testing.T) {
	// os.Setenv("HELM_VERSION", "1.0.63.1")
	// SetConsoleApi("http://172.16.1.116:9004")
	RefreshCDToken()
}

func TestRefreshCDToken(t *testing.T) {
	// Setup
	os.Setenv("HELM_VERSION", "1.0.63.1")
	// SetConsoleApi("http://172.16.1.116:9004")
	RefreshCDToken()

}
