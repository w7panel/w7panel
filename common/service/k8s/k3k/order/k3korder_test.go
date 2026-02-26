// nolint
package order

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	v1 "k8s.io/api/core/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
)

func TestF64(t *testing.T) {

	end, err := time.Parse("2006-01-02 15:04:05", "2025-10-22 02:21:57")
	if err != nil {
		t.Error(err)
		return
	}
	hours := end.Sub(time.Now()).Hours()
	t.Logf("%v", hours)

	test1 := hours / 24
	test2 := int32(test1)
	t.Logf("%v", test2)

	// t.Logf("%v", f)
}

func TestUser(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}
	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "console-164315"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	k3k.RefreshK3kUser(k3kUser, sdk, true)
	lr := k3kUser.HasProcessReturnOrder()
	t.Log(lr)
}

func TestCreateBaseResourceOrder(t *testing.T) {

	// os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("USER_AGENT", "we7test-beta")
	// os.Setenv("LOCAL_MOCK", "1")
	os.Setenv("DEBUG", "true")

	// console.SetConsoleApi("http://172.16.1.116:9004")
	// Setup mock
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "console-164315"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	k3k.RefreshK3kUser(k3kUser, sdk, true)
	bs := types.BuyResource{
		Cpu:       2,
		Memory:    4,
		Storage:   10,
		Bandwidth: 100,
	}
	pay, err := CreateBaseResourceOrder(&types.BuyBaseResource{CouponCode: "", BaseConfigName: "admin", UnitQuantity: types.UnitQuantity{Unit: "month", Quantity: 1}, BuyResource: bs}, k3kUser)
	if err != nil {
		t.Error(err)
		return
	}

	// Refresh(k3kUser)
	t.Log(pay)
}

func TestCreateRenewResourceOrder(t *testing.T) {

	os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("USER_AGENT", "we7test-beta")
	os.Setenv("DEBUG", "true")

	// console.SetConsoleApi("http://172.16.1.116:9004")
	// Setup mock
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "console-98655"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	needBuy := k3kUser.NeedBuyResource()
	t.Log(needBuy)
	pay, err := CreateRenewOrder(&types.BuyRenewResource{BaseConfigName: "admin", UnitQuantity: types.UnitQuantity{Unit: "month", Quantity: 3}}, k3kUser)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(pay)

	// NotifyPaid(k3kUser)
	// t.Log(pay)
}

func TestCreateExpandResourceOrder(t *testing.T) {

	os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("USER_AGENT", "we7test-beta")
	os.Setenv("DEBUG", "true")

	// console.SetConsoleApi("http://172.16.1.116:9004")
	// Setup mock
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "console-164315"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	bs := types.BuyResource{
		Cpu:       2,
		Memory:    4,
		Storage:   10,
		Bandwidth: 120,
	}
	pay, err := CreateExpandOrder(&types.BuyExpandResource{BaseConfigName: "admin", BuyResource: bs}, k3kUser)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(pay)

	// t.Log(pay)
}

func TestNotify(t *testing.T) {

	os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("USER_AGENT", "we7test-beta")
	os.Setenv("DEBUG", "true")

	// console.SetConsoleApi("http://172.16.1.116:9004")
	// Setup mock
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "console-98655"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	NotifyOrder(k3kUser, "20251118174758A7RFU3")

	// t.Log(pay)
}

func TestMockNotify(t *testing.T) {

	os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("USER_AGENT", "we7test-beta")
	os.Setenv("DEBUG", "true")

	// console.SetConsoleApi("http://172.16.1.116:9004")
	// Setup mock
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "console-164315"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	MockNotifyOrder(k3kUser, "20251124185846AY5HPU")

	// t.Log(pay)
}

func TestCheckCanBuy(t *testing.T) {

	os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("USER_AGENT", "we7test-beta")
	os.Setenv("DEBUG", "true")

	// console.SetConsoleApi("http://172.16.1.116:9004")
	// Setup mock
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "console-164315"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	err = CheckCanBuy(k3kUser)
	if err != nil {
		t.Log(err)
	}

	// t.Log(pay)
}

func TestLastPaidOrder(t *testing.T) {

	// os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("USER_AGENT", "we7test-beta")
	os.Setenv("DEBUG", "true")

	// console.SetConsoleApi("http://172.16.1.116:9004")
	// Setup mock
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "console-164315"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	k3kOrderApi, err := NewK3kOrderApi(sdk)
	if err != nil {
		t.Log(err)
		return
	}
	order, err := k3kOrderApi.FindLastReturnOrder(k3kUser)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(order)

	// t.Log(pay)
}

func TestLastProcessReturnOrder(t *testing.T) {

	os.Setenv("LOCAL_MOCK", "true")
	os.Setenv("USER_AGENT", "we7test-beta")
	os.Setenv("DEBUG", "true")

	// console.SetConsoleApi("http://172.16.1.116:9004")
	// Setup mock
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
		return
	}

	sa := &v1.ServiceAccount{}
	err = client.Get(sdk.Ctx, ktypes.NamespacedName{Namespace: "default", Name: "console-164315"}, sa)
	if err != nil {
		t.Error(err)
		return
	}
	k3kUser := types.NewK3kUser(sa)
	k3kOrderApi, err := NewK3kOrderApi(sdk)
	if err != nil {
		t.Log(err)
		return
	}

	if k3kUser.HasProcessReturnOrder() {
		slog.Info("has process return order", "name", sa.Name)
		err := k3kOrderApi.ProcessReturnOrder(k3kUser)
		if err != nil {
			slog.Warn("处理退款记录失败1", "name", sa.Name, "err", err)
			t.Log(err)
		}
	}
	slog.Info("check new return log")
	err = k3kOrderApi.ProcessReturnLastOrder(k3kUser, true)
	if err != nil {
		slog.Warn("处理退款记录失败2", "name", sa.Name, "err", err)
		t.Log(err)
	}

	// t.Log(pay)
}
