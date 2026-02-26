package helper

import (
	"context"
	"fmt"
	"net"
	"time"
)

// DNSRecord 表示 DNS 记录
type DNSRecord struct {
	Type     string
	Name     string
	Value    string
	TTL      uint32
	Priority uint16 // 用于 MX 记录
}

// DNSGetRecord 查询 DNS 记录
func DNSGetRecord(domain string, recordType string) ([]DNSRecord, error) {
	resolver := &net.Resolver{
		PreferGo: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch recordType {
	case "A":
		return queryARecords(ctx, resolver, domain)
	case "AAAA":
		return queryAAAARecords(ctx, resolver, domain)
	case "CNAME":
		return queryCNAMERecord(ctx, resolver, domain)
	case "MX":
		return queryMXRecords(ctx, resolver, domain)
	case "TXT":
		return queryTXTRecords(ctx, resolver, domain)
	case "NS":
		return queryNSRecords(ctx, resolver, domain)
	default:
		return nil, fmt.Errorf("unsupported record type: %s", recordType)
	}
}

// queryARecords 查询域名的A记录
// 参数:
//
//	ctx context.Context - 上下文
//	resolver *net.Resolver - DNS解析器
//	domain string - 要查询的域名
//
// 返回:
//
//	[]DNSRecord - 查询到的A记录列表
//	error - 错误信息
func queryARecords(ctx context.Context, resolver *net.Resolver, domain string) ([]DNSRecord, error) {
	var records []DNSRecord
	ips, err := resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		if ip.IP.To4() != nil {
			records = append(records, DNSRecord{
				Type:  "A",
				Name:  domain,
				Value: ip.IP.String(),
			})
		}
	}
	return records, nil
}

// queryAAAARecords 查询域名的AAAA记录
// 参数:
//
//	ctx context.Context - 上下文
//	resolver *net.Resolver - DNS解析器
//	domain string - 要查询的域名
//
// 返回:
//
//	[]DNSRecord - 查询到的AAAA记录列表
//	error - 错误信息
func queryAAAARecords(ctx context.Context, resolver *net.Resolver, domain string) ([]DNSRecord, error) {
	var records []DNSRecord
	ips, err := resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		if ip.IP.To4() == nil && ip.IP.To16() != nil {
			records = append(records, DNSRecord{
				Type:  "AAAA",
				Name:  domain,
				Value: ip.IP.String(),
			})
		}
	}
	return records, nil
}

// queryCNAMERecord 查询域名的CNAME记录
// 参数:
//
//	ctx context.Context - 上下文
//	resolver *net.Resolver - DNS解析器
//	domain string - 要查询的域名
//
// 返回:
//
//	[]DNSRecord - 查询到的CNAME记录
//	error - 错误信息
func queryCNAMERecord(ctx context.Context, resolver *net.Resolver, domain string) ([]DNSRecord, error) {
	cname, err := resolver.LookupCNAME(ctx, domain)
	if err != nil {
		return nil, err
	}
	return []DNSRecord{{
		Type:  "CNAME",
		Name:  domain,
		Value: cname,
	}}, nil
}

// queryMXRecords 查询域名的MX记录
// 参数:
//
//	ctx context.Context - 上下文
//	resolver *net.Resolver - DNS解析器
//	domain string - 要查询的域名
//
// 返回:
//
//	[]DNSRecord - 查询到的MX记录列表
//	error - 错误信息
func queryMXRecords(ctx context.Context, resolver *net.Resolver, domain string) ([]DNSRecord, error) {
	var records []DNSRecord
	mxs, err := resolver.LookupMX(ctx, domain)
	if err != nil {
		return nil, err
	}
	for _, mx := range mxs {
		records = append(records, DNSRecord{
			Type:     "MX",
			Name:     domain,
			Value:    mx.Host,
			Priority: mx.Pref,
		})
	}
	return records, nil
}

// queryTXTRecords 查询域名的TXT记录
// 参数:
//
//	ctx context.Context - 上下文
//	resolver *net.Resolver - DNS解析器
//	domain string - 要查询的域名
//
// 返回:
//
//	[]DNSRecord - 查询到的TXT记录列表
//	error - 错误信息
func queryTXTRecords(ctx context.Context, resolver *net.Resolver, domain string) ([]DNSRecord, error) {
	var records []DNSRecord
	txts, err := resolver.LookupTXT(ctx, domain)
	if err != nil {
		return nil, err
	}
	for _, txt := range txts {
		records = append(records, DNSRecord{
			Type:  "TXT",
			Name:  domain,
			Value: txt,
		})
	}
	return records, nil
}

// queryNSRecords 查询域名的NS记录
// 参数:
//
//	ctx context.Context - 上下文
//	resolver *net.Resolver - DNS解析器
//	domain string - 要查询的域名
//
// 返回:
//
//	[]DNSRecord - 查询到的NS记录列表
//	error - 错误信息
func queryNSRecords(ctx context.Context, resolver *net.Resolver, domain string) ([]DNSRecord, error) {
	var records []DNSRecord
	nss, err := resolver.LookupNS(ctx, domain)
	if err != nil {
		return nil, err
	}
	for _, ns := range nss {
		records = append(records, DNSRecord{
			Type:  "NS",
			Name:  domain,
			Value: ns.Host,
		})
	}
	return records, nil
}
