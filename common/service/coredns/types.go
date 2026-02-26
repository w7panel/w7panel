package coredns

import (
	"encoding/json"
	"fmt"

	"github.com/coredns/caddy/caddyfile"
)

type DnsConfig struct {
	Dns []Dns
}

func (self DnsConfig) ToCaddyfile() ([]byte, error) {
	file := self.ToEncodedCaddyfile()
	data, err := json.Marshal(file)
	if err != nil {
		return nil, err
	}
	return caddyfile.FromJSON(data)
}

func (self DnsConfig) ToEncodedCaddyfile() caddyfile.EncodedCaddyfile {
	result := caddyfile.EncodedCaddyfile{}
	for _, dns := range self.Dns {
		block := caddyfile.EncodedServerBlock{}
		block.Keys = []string{dns.Domain}
		for _, record := range dns.DnsRecords {
			ndomain := dns.ToDomain(record)
			answerRes := fmt.Sprintf("%s %d In %s %s", ndomain, record.TTL, record.Type, record.Content)
			template := []interface{}{"template", "IN", record.Type, record.Content, []string{"answer", answerRes}}
			block.Body = append(block.Body, template)
		}
		result = append(result, block)
	}
	return result
}

type DnsRecord struct {
	Name    string
	Type    string
	TTL     int
	Content string
}

type Dns struct {
	Domain     string
	DnsRecords []DnsRecord
}

func (self Dns) ToDomain(record DnsRecord) string {
	return record.Name + "." + self.Domain + "."
}
