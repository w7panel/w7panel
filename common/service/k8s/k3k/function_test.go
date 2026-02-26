// nolint
package k3k

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTokenToK3kUser(t *testing.T) {

}

func TestRefreshK3kUser(t *testing.T) {
	token := `eyJhbGciOiJSUzI1NiIsImtpZCI6IklZd3JvWW5mQWM5Uk1VaWlhVmVLOWQzcE1SZFoteloyUXhCOC10YlQxakEifQ.eyJhdWQiOlsiY29uc29sZS03NTc4MCIsImszay1jb25zb2xlLTc1NzgwIiwiazNrLWNvbnNvbGUtNzU3ODAtc2VydmljZS5rM2stY29uc29sZS03NTc4MCIsInNoYXJlZCIsImhzeWZibnFvIiwiMSIsIjEiLCJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbCIsImszcyJdLCJleHAiOjE3NTYzNzcyMTAsImlhdCI6MTc1NjM3MzYxMCwiaXNzIjoiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJqdGkiOiJlYWIxZTE1YS0zNzdiLTRmOTAtOWM3Yy01NmQxMWVmNGUyZTYiLCJrdWJlcm5ldGVzLmlvIjp7Im5hbWVzcGFjZSI6ImRlZmF1bHQiLCJzZXJ2aWNlYWNjb3VudCI6eyJuYW1lIjoiY29uc29sZS03NTc4MCIsInVpZCI6IjIwOWI5MWE5LTVlZDctNDRmOS05MDNmLWNlM2RmNGUyNzU2MCJ9fSwibmJmIjoxNzU2MzczNjEwLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpjb25zb2xlLTc1NzgwIn0.o2kEye4rp8rH8W9YtXtsZueHZoWoR8SKXoCmEljXrmg-2dSkrGrSZIT1ofCN16i8QbP2TdbeY4vVuBGumE2yIA78cb2hMyIV3VbFROGoPDIYZNewDbc7hi0_0UASzwGjdAIxCNCQJzJZ0SKM9lNyjijTcy9l0U6yrpqnE_qqbP2xFf4OrhdiDczzKlUh6_D2nVawlYUMzPSvRq9Kv7-mdMoevqZ-QLB48Ej_Je45ff4VwqH5Wx-gOF-VOy4IME4ZHz4LdMLnpqbTSapZPzjrzAlILAyVIJvQFZruQzzlPXVMqQGxRmPdGePUqTbFf4F0aWDxlKGrBVLH9KV0-5-sWw	
	`
	user, err := TokenToK3kUser(token)
	if err != nil {
		t.Error(err)
	}

	t.Log(user)
	t.Log(user.ToArray())
}

func TestTryOver(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk
	user, err := sdk.ClientSet.CoreV1().ServiceAccounts("default").Get(sdk.Ctx, "console-164315", metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}
	t.Log(user)

	kuser := types.NewK3kUser(user)
	err = TryCheckOverSellingResource(sdk, kuser)
	if err != nil {
		t.Error(err)
	}
}

func TestResourceVal(t *testing.T) {
	storage := resource.MustParse("100Mi")
	scaledValue := storage.Value()
	t.Log(scaledValue)

	// scaledValue2 := storage.CanonicalizeBytes()()
	// t.Log(scaledValue2)

}
