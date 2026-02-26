// nolint
package console

import (
	"os"
	"testing"
)

func TestCreateProductOrder(t *testing.T) {
	os.Setenv("USER_AGENT", "we7test-beta")
	sdkClient, err := NewDefaultSdkClient()
	if err != nil {
		t.Fatal(err)
	}
	info, err := sdkClient.CreateDefaultProductOrder("10665")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestCreateProductOrder1(t *testing.T) {
	os.Setenv("USER_AGENT", "we7test-beta")
	sdkClient, err := NewDefaultSdkClient()
	if err != nil {
		t.Fatal(err)
	}
	info, err := sdkClient.CreateDefaultProductOrder("10665")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestPrepareProduct(t *testing.T) {
	os.Setenv("USER_AGENT", "we7test-beta")
	sdkClient, err := NewDefaultSdkClient()
	if err != nil {
		t.Fatal(err)
	}
	info, err := sdkClient.PrepareProduct2()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestShowCoupon(t *testing.T) {
	os.Setenv("USER_AGENT", "we7test-beta")
	// os.Setenv("LOCAL_MOCK", "true")
	// os.Setenv("")
	sdkClient, err := NewDefaultSdkClient()
	if err != nil {
		t.Fatal(err)
	}
	info, err := sdkClient.GetCoupon("20251212184530-4H2P3ZVD")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestUpdateCoupon(t *testing.T) {
	os.Setenv("USER_AGENT", "we7test-beta")
	// os.Setenv("LOCAL_MOCK", "true")
	// os.Setenv("")
	sdkClient, err := NewDefaultSdkClient()
	if err != nil {
		t.Fatal(err)
	}

	err = sdkClient.UpdateCoupon("20251229154927-Z2LTJTVR", "lock", "xxxddd")
	if err != nil {
		t.Fatal(err)
	}

}

func TestFindLastReturnOrder(t *testing.T) {
	os.Setenv("USER_AGENT", "we7test-beta")
	// os.Setenv("LOCAL_MOCK", "true")
	// os.Setenv("")
	// sdkClient, err := NewDefaultSdkClient()
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// err = sdkClient.FindLastReturnOrder("")
	// if err != nil {
	// 	t.Fatal(err)
	// }

}
