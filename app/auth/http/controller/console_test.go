package controller

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/console"
)

func TestConsole_Redirect(t *testing.T) {

	client := console.DefaultClient(false)
	redirectUrl, err := client.OauthService.GetLoginUrl("https://www.k8s-offline.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(redirectUrl)
}
