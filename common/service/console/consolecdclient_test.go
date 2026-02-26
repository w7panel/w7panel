package console

import (
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	// "gitee.com/we7coreteam/k8s-offline/common/service/config"
)

func TestRefreshUseCdToken(t *testing.T) {

	repository := config.NewW7ConfigRepository(k8s.NewK8sClientInner())

	w7config, err := repository.Get("admin")
	if err != nil {
		t.Error(err)
	}
	token := w7config.ThirdpartyCDToken

	cdclient := NewConsoleCdClient(token)

	_, err = cdclient.RefreshToken()
	if err != nil {
		t.Error(err)
	}
}

func TestPreInstall(t *testing.T) {
	SetConsoleApi("http://172.16.1.126:9004")
	consoleCdClient := NewConsoleCdClient("xxx")
	p2, err := consoleCdClient.PreInstall("test", "1")
	if err != nil {
		t.Error(err)
	}
	print(p2)
}

func TestVerifyCert(t *testing.T) {
	cert := `-----BEGIN CERTIFICATE-----
MIIEKDCCAxCgAwIBAgIBAjANBgkqhkiG9w0BAQsFADCBrDELMAkGA1UEBhMCQ04x
DTALBgNVBAgMBHRlYW0xDjAMBgNVBAcMBTc2MDUyMQswCQYDVQQKDAJ3NzEdMBsG
A1UECwwUMjAyNTA2MTIxNDMyMTZMRDFISFgxKzApBgNVBAMMIjIwMjUwNjEyMTQz
MjE2TEQxSEhYLmxpY2Vuc2UudzcuY2MxJTAjBgkqhkiG9w0BCQEWFmFkbWluQHRl
c3QuZXhhbXBsZS5jb20wHhcNMjUwNjEyMDYzMjQwWhcNMjUwOTExMDYzMjQwWjCB
rDELMAkGA1UEBhMCQ04xDTALBgNVBAgMBHRlYW0xDjAMBgNVBAcMBTc2MDUyMQsw
CQYDVQQKDAJ3NzEdMBsGA1UECwwUMjAyNTA2MTIxNDMyMTZMRDFISFgxKzApBgNV
BAMMIjIwMjUwNjEyMTQzMjE2TEQxSEhYLmxpY2Vuc2UudzcuY2MxJTAjBgkqhkiG
9w0BCQEWFmFkbWluQHRlc3QuZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUA
A4IBDwAwggEKAoIBAQCeO2nFYwy5tJYkAj/MHhDtfWJyBAIS5DX/HMVBqt4aoRau
bCpavNyqCIbmkxReTc9BgsAmRumyNs7BI6p+4nUK1FeHPs1b+2DvejzvK2e5nAL8
dOmUn9L5LbQsMJmPH2qkv7t/0XvEsW/B9rZtqY0Y5rbYxzQfsfHS6SGHCCVCjfUW
LbENV1unAyG1WOW2DcM2NlFXITdvGNm98L4Xu/dP2li8QdkIBTfDoBOj2za0dFTo
WzdhDxKKIYmQmmZrTNgh2VSzZrA0JAJ/l4q3zHCk9Sdvy8K9S4XNqf3vjct393sK
zmUxSMlTisTCVJXQLy9NuDrXjsMZUwsksd03n9ODAgMBAAGjUzBRMB0GA1UdDgQW
BBRJJOZfxYgoj6J3U8+VZ2ThwJAIgDAfBgNVHSMEGDAWgBRJJOZfxYgoj6J3U8+V
Z2ThwJAIgDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQA8t7gK
9zQi9PD459xneSVlseiOIOSXmw2Bt8Qt5jk+QWUkyGk+Sf/4Xc4eSAEjBN73cS4E
czky8AsslgfU/m5kNIRJYKMJouwO/E6prKF6soBLe/nAgjndW6VfLeDVp38xBh29
8a0N/bktEYKU8e/ThpsdN6zRZlQlbAfUv6i7iVptUZI1QUIi0cai/M2W3hq6tMgQ
32USxh3yYlwh3N3HnS4JrFQmYQehW747KWJV7AvRrcfcoUHsQAaaxTgTEEC6mXAf
o/Cibki8ZP6aYdu3ZSVFWJlTH2Rs2sAg0A3SzT7rAox0PETvLctCA0aIfNIqYnrD
nNvPNp8iHEnLR5Vq
-----END CERTIFICATE-----
`
	os.Setenv("USER_AGENT", "we7test-beta")
	parsedCert, err := helper.ParseX509([]byte(cert))
	if err != nil {
		t.Error("xxxx", "err", err)
	}
	verify, err := VerifyCert(parsedCert)
	if err != nil {
		t.Error("xxxx", "err", err)
	}
	print(verify)
}
