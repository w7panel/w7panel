package coredns

import (
	"testing"
)

func TestConfig(t *testing.T) {
	dnsConfig := DnsConfig{
		Dns: []Dns{
			{
				Domain: "demo.com",
				DnsRecords: []DnsRecord{
					{
						Name:    "www",
						Type:    "A",
						TTL:     60,
						Content: "1.1.1.1",
					},
				},
			},
		},
	}
	data, err := dnsConfig.ToCaddyfile()
	if err != nil {
		t.Error(err)
	}
	strData := string(data)
	t.Log(strData)
}
