package controller

import (
	"testing"

	"github.com/w7panel/w7panel/common/service/console"
)

func TestConsole_Redirect(t *testing.T) {

	client := console.DefaultClient(false)
	redirectUrl, err := client.OauthService.GetLoginUrl("https://www.k8s-offline.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(redirectUrl)
}
