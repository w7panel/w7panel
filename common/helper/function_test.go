// nolint
package helper

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestRandomByte(t *testing.T) {
	length := 10
	bytes := RandomByte(length)
	if len(bytes) != length {
		t.Fatalf("Expected length %d, got %d", length, len(bytes))
	}
	for _, b := range bytes {
		if b != 'a' && b != 'b' && b != 'c' && b != 'd' && b != 'e' && b != 'f' {
			t.Fatalf("Expected byte value within 'a' to 'f', got %c", b)
		}
	}
}

func TestRandomByteCrypto(t *testing.T) {
	length := 10
	bytes := RandomByte(length)
	if len(bytes) != length {
		t.Fatalf("Expected length %d, got %d", length, len(bytes))
	}
	// allZeros := bytes.Repeat([]byte{0}, length)
	// if bytes == allZeros {
	// 	t.Fatalf("Expected at least one byte to be non-zero, got all zeros")
	// }
}

func TestLaravelAppKey(t *testing.T) {
	appKey := LaravelAppKey(32)
	if !strings.HasPrefix(appKey, "base64:") {
		t.Fatalf("Expected appKey to start with 'base64:', got %s", appKey)
	}
	appKey = LaravelAppKey(16)
	if !strings.HasPrefix(appKey, "base64:") {
		t.Fatalf("Expected appKey to start with 'base64:', got %s", appKey)
	}
}

func TestMyIp(t *testing.T) {
	// 模拟 http.Get 返回的响应
	// ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("127.0.0.1"))
	// }))
	// defer ts.Close()
	// http.DefaultClient.Transport = ts.Transport
	ip, err := MyIp()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if ip != "127.0.0.1" {
		t.Fatalf("Expected IP 127.0.0.1, got %s", ip)
	}
}

func TestNcenterShell(t *testing.T) {
	// 模拟 http.Get 返回的响应
	// ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("127.0.0.1"))
	// }))
	// defer ts.Close()
	// http.DefaultClient.Transport = ts.Transport
	// data, err := os.ReadFile("../../kodata/registries.yaml")
	// if err != nil {
	// 	t.Fatalf("Expected no error, got %v", err)
	// }

}

func TestExtractSingleFileFromTgz(t *testing.T) {
	// 模拟 http.Get 返回的响应
	data, err := ExtractSingleFileFromTgz("https://Project-HAMi.github.io/HAMi/charts/hami-2.5.0.tgz", "Chart.yaml")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	ddd := (string(data))
	print(ddd)
}

func TestCert(t *testing.T) {
	cert := `-----BEGIN CERTIFICATE-----
MIID/jCCAuagAwIBAgIBBzANBgkqhkiG9w0BAQsFADCBlzELMAkGA1UEBhMCQ04x
EDAOBgNVBAgMB2NvbXBhbnkxEDAOBgNVBAcMB0JlaWppbmcxCzAJBgNVBAoMAnc3
MRAwDgYDVQQLDAd4eHh4MTIzMR4wHAYDVQQDDBV4eHh4MTIzLmxpY2Vuc2Uudzcu
Y2MxJTAjBgkqhkiG9w0BCQEWFmFkbWluQHRlc3QuZXhhbXBsZS5jb20wHhcNMjUw
NjExMDIzMzI2WhcNMjUwOTExMDIzMzI2WjCBlzELMAkGA1UEBhMCQ04xEDAOBgNV
BAgMB2NvbXBhbnkxEDAOBgNVBAcMB0JlaWppbmcxCzAJBgNVBAoMAnc3MRAwDgYD
VQQLDAd4eHh4MTIzMR4wHAYDVQQDDBV4eHh4MTIzLmxpY2Vuc2UudzcuY2MxJTAj
BgkqhkiG9w0BCQEWFmFkbWluQHRlc3QuZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQC8jnGMsOkfcib4KJ0uSiol6NsI3UpHFcZF+TSc
GRFLTCvvC0c6MxXHks89i+Xh6HKZV2sceWi1rym40057QsZT8rpjmoxqrfgdZf5H
Y70lxiVx3i4ICwxWkRIbPcVjtAwHIzLZieprWrq+5kGkTtD9lwOzv0PKajPyTbeL
GKOMZRN93b7wvvsZHebdm6+trpWSdvEtfHjD5Ap0cgrQ+SC4vfmq/66GPX4qQbgG
cK+AEX++mYfSh4T6RCh6UDyFqMG1si9pfN6iLv6CTRpij0eEq32sWiAliTvUulN2
6b/69oKZqvGEvi7osqv8WV0+R41PWVY/VvbUzcrmxpTt2Dw1AgMBAAGjUzBRMB0G
A1UdDgQWBBQZP1MGR0D4lywPjYYkpA+M0yK5CzAfBgNVHSMEGDAWgBQZP1MGR0D4
lywPjYYkpA+M0yK5CzAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IB
AQBjLpEgP0F7+XGceYp2Ypr1PjbEqNfyPHKBU910DpuTI7z5LkEYB3GpeaO4v0KN
5Z07gRU9ug9GdmhlrPKdlEq+V8XNaipINFKtx+i3p3LRY2LYBMLv85U2dPOKMItM
U2PyAJDwnWIwhjW78WkG/Nvda57mrYunMKvea2TJ1h/eKel1R+o4iZe/+aSrVe9A
nHuaeG4+vh0Oh5m2wI7kbZWgKHQu3vsenRgGDKRzWNSxmj8AQ4JKKPmP9aGUf3jw
1oV0j1l+3Yc881l8hDm9Krmw0vlMUwsfj44O1lxALpSkJO0rqekcUot6eztTJwGG
TPZdYzQi8u0pZBe5xDNTpkT9
-----END CERTIFICATE-----`
	x509, err := ParseX509([]byte(cert))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	data, err := EncryptWithPublicKey(x509.PublicKey.(*rsa.PublicKey), []byte("hello"), true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	base64data := string(data)
	print(base64data)

	sign := `DuauPWuoED3i20UQMJiV2rnv5A+zxfwF6oVvwUKCh4UF5lVClIXPt6D+gdpOqZG8PvOK5SfTHtXAA8ZqmFJJynuHUNCtq+mWgR+P7TJb0joJrMdRAcNPrCvYS9XveKX8TsBXHfUcsyWMZz8uyRrrU7S7aLAChL2l4rRNPmie8ufovmWxRStWKtyspz7+zDebmUhfIhnLxaFdyhYH9gQcmRD/v99ZOi77arpu6pSQ7SO8WCTw71fhGTjWsfHb0hoUpvwbQL6EejDVKFAfOKIGTKOoCcZ90SaYWxJb3m0v8Kw5XoV1kVavg7aXaQkRhC18V1zFWu/vLUlstXoz3cYERQ==`
	sign2, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	err = VerifyDataWithPublicKey(x509.PublicKey.(*rsa.PublicKey), []byte("hello"), []byte(sign2))
	if err != nil {
		t.Fatalf("Expected error, got %v", err)
	}

}

func TestStringArr(t *testing.T) {
	k3kconfigYaml := map[string]interface{}{
		"tls-san": []string{"127.0.0.1"},
	}
	data, ok := k3kconfigYaml["tls-san"]
	if ok {
		if slice, ok1 := data.([]string); ok1 {
			fmt.Println(slice)
			// return slice
		}
	}
	fmt.Println(data)
}

func TestDifference(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{
		{
			name: "empty slices",
			a:    []string{},
			b:    []string{},
			want: []string{},
		},
		{
			name: "a empty",
			a:    []string{},
			b:    []string{"1", "2"},
			want: []string{},
		},
		{
			name: "b empty",
			a:    []string{"1", "2"},
			b:    []string{},
			want: []string{"1", "2"},
		},
		{
			name: "no difference",
			a:    []string{"1", "2"},
			b:    []string{"1", "2"},
			want: []string{},
		},
		{
			name: "partial difference",
			a:    []string{"1", "2", "3"},
			b:    []string{"1", "3"},
			want: []string{"2"},
		},
		{
			name: "complete difference",
			a:    []string{"1", "2", "3"},
			b:    []string{"4", "5"},
			want: []string{"1", "2", "3"},
		},
		{
			name: "duplicate in a",
			a:    []string{"1", "1", "2"},
			b:    []string{"1"},
			want: []string{"2"},
		},
		{
			name: "duplicate in a",
			a:    []string{"10.0.0.206", "218.23.2.55", "127.0.0.1"},
			b:    []string{"127.0.0.1"},
			want: []string{"127.0.0.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Difference(tt.a, tt.b)
			if len(got) != len(tt.want) {
				t.Fatalf("Difference() length mismatch, want %v, got %v", tt.want, got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("Difference() mismatch at index %d, want %v, got %v", i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestImageDigest(t *testing.T) {
	sha256, err := ImageDigest("ccr.ccs.tencentyun.com/onceyoungs/w7:prozt2.0.7")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	print(sha256)
}




func TestIpCity(t *testing.T) {
	os.Setenv("KO_DATA_PATH", "../../kodata")
	result, err := IpCity("120.209.216.232")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	print(result)
}
