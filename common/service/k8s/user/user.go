package user

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"

	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CSRConfig 配置CSR请求参数
type CSRConfig struct {
	CommonName   string
	Organization []string
	Usages       []certificatesv1.KeyUsage
}

// CertificateService 提供证书签发和管理功能
type CertificateService struct {
	client kubernetes.Interface
}

// NewCertificateService 创建新的证书服务
func NewCertificateService(client kubernetes.Interface) *CertificateService {
	return &CertificateService{client: client}
}

// GenerateCSR 生成证书签名请求
func (s *CertificateService) GenerateCSR(config CSRConfig) (*rsa.PrivateKey, []byte, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// 创建CSR模板
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   config.CommonName,
			Organization: config.Organization,
		},
	}

	// 生成CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CSR: %v", err)
	}

	// PEM编码CSR
	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	return privateKey, csrPEM, nil
}

// CreateK8sCSR 在Kubernetes中创建CertificateSigningRequest资源
func (s *CertificateService) CreateK8sCSR(name string, csrPEM []byte, usages []certificatesv1.KeyUsage) error {
	csr := &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request:    csrPEM,
			SignerName: "kubernetes.io/kube-apiserver-client",
			Usages:     usages,
		},
	}

	_, err := s.client.CertificatesV1().CertificateSigningRequests().Create(context.Background(), csr, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create CSR resource: %v", err)
	}

	return nil
}

// ApproveCSR 审批CertificateSigningRequest
func (s *CertificateService) ApproveCSR(name string) error {
	// 获取CSR
	csr, err := s.client.CertificatesV1().CertificateSigningRequests().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get CSR: %v", err)
	}

	// 添加审批条件
	csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
		Type:           certificatesv1.CertificateApproved,
		Status:         corev1.ConditionTrue,
		Reason:         "ApprovedByUserService",
		Message:        "Approved by user certificate service",
		LastUpdateTime: metav1.Now(),
	})

	// 更新审批状态
	_, err = s.client.CertificatesV1().CertificateSigningRequests().UpdateApproval(context.Background(), csr.Name, csr, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to approve CSR: %v", err)
	}

	return nil
}

// GetSignedCertificate 获取已签名的证书
func (s *CertificateService) GetSignedCertificate(name string) ([]byte, error) {
	csr, err := s.client.CertificatesV1().CertificateSigningRequests().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get CSR: %v", err)
	}

	if len(csr.Status.Certificate) == 0 {
		return nil, fmt.Errorf("certificate not yet issued")
	}

	return csr.Status.Certificate, nil
}
