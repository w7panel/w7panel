package user

import (
	"testing"

	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCertificateService(t *testing.T) {
	// 创建fake客户端
	client := fake.NewSimpleClientset()
	service := NewCertificateService(client)

	// 测试生成CSR
	config := CSRConfig{
		CommonName:   "test-user",
		Organization: []string{"test-org"},
		Usages:       []certificatesv1.KeyUsage{certificatesv1.UsageDigitalSignature},
	}

	privateKey, csrPEM, err := service.GenerateCSR(config)
	if err != nil {
		t.Fatalf("GenerateCSR failed: %v", err)
	}
	if privateKey == nil {
		t.Error("Expected private key, got nil")
	}
	if len(csrPEM) == 0 {
		t.Error("Expected CSR PEM, got empty")
	}

	// 测试创建K8s CSR
	csrName := "test-csr"
	err = service.CreateK8sCSR(csrName, csrPEM, config.Usages)
	if err != nil {
		t.Fatalf("CreateK8sCSR failed: %v", err)
	}

	// 测试审批CSR
	err = service.ApproveCSR(csrName)
	if err != nil {
		t.Fatalf("ApproveCSR failed: %v", err)
	}

	// 测试获取签名证书
	_, err = service.GetSignedCertificate(csrName)
	if err == nil {
		t.Error("Expected error for not yet issued certificate, got nil")
	}
}
