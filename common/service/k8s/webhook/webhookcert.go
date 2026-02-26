package webhook

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CertDir  = "/tmp/k8s-webhook-server/serving-certs"
	certFile = "tls.crt"
	keyFile  = "tls.key"
)
const CERT_SECRET_NAME = "webhook-cert"

func ensureCertificates(namespace string) error {
	sdk := k8s.NewK8sClient().Sdk
	cert, err := NewCert(sdk).GetCert(getSvcHost(sdk.GetNamespace()))
	if err != nil {
		return err
	}
	return cert.WriteToFile()
}

func Cert(namespace string) error {
	return ensureCertificates(namespace)
}

type cert struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

func (c *cert) ToSecret() corev1.Secret {
	return corev1.Secret{
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte(c.Cert),
			"tls.key": []byte(c.Key),
		},
	}
}
func (c *cert) WriteToFile() error {
	// Create cert directory if not exists
	if err := os.MkdirAll(CertDir, 0755); err != nil {
		return err
	}

	// Write cert file
	certPath := filepath.Join(CertDir, certFile)
	if err := os.WriteFile(certPath, []byte(c.Cert), 0644); err != nil {
		return err
	}

	// Write key file
	keyPath := filepath.Join(CertDir, keyFile)
	if err := os.WriteFile(keyPath, []byte(c.Key), 0600); err != nil {
		return err
	}

	return nil
}

type webhookcert struct {
	sdk *k8s.Sdk
}

func NewCert(sdk *k8s.Sdk) *webhookcert {
	return &webhookcert{
		sdk: sdk,
	}
}

func (c *webhookcert) GetCert(host string) (*cert, error) {
	secret, err := c.sdk.ClientSet.CoreV1().Secrets(c.sdk.GetNamespace()).Get(c.sdk.Ctx, getSecret(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return c.CreateCert(host)
		}
		return nil, err
	}
	return &cert{
		Cert: string(secret.Data["tls.crt"]),
		Key:  string(secret.Data["tls.key"]),
	}, nil
}
func (c *webhookcert) CreateCert(host string) (*cert, error) {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: host,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
		DNSNames:              []string{host},
	}

	// Create self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	// Encode certificate and key to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	// Create secret with certificate
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: getSecret(),
		},
		Data: map[string][]byte{
			"tls.crt": certPEM,
			"tls.key": privPEM,
		},
	}

	// Create or update secret
	_, err = c.sdk.ClientSet.CoreV1().Secrets(c.sdk.GetNamespace()).Create(c.sdk.Ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			_, err = c.sdk.ClientSet.CoreV1().Secrets(c.sdk.GetNamespace()).Update(c.sdk.Ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &cert{
		Cert: string(certPEM),
		Key:  string(privPEM),
	}, nil
}
