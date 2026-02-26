package console

import (
	"testing"
)

type mockLicenseClient struct {
	cleanErr error
}

func (m *mockLicenseClient) CleanLicense() error {
	return m.cleanErr
}

func TestCleanLicenseCert(t *testing.T) {
	cleanLicenseCert(true)
}

func TestVerifyLicenseId(t *testing.T) {
	SetConsoleApi("http://10.0.2.15:9004")
	err := VerifyLicenseId("9", "admin")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

// func TestVerify(t *testing.T) {
// 	err := VerifyLicense()
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// }
