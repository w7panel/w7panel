package console

import (
	"crypto/x509"
	"errors"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type LicenseClient struct {
	w7config config.W7ConfigRepositoryInterface
	sdk      *k8s.Sdk
}

var licenseClient *LicenseClient

func init() {
	licenseClient, _ = NewDefaultLicenseClient()
}

func NewDefaultLicenseClient() (*LicenseClient, error) {
	sdk := k8s.NewK8sClient().Sdk
	if sdk == nil {
		return nil, errors.New("sdk is nil")
	}
	api := NewLicenseClient(config.NewW7ConfigRepository(sdk), sdk)
	return api, nil
}

func NewLicenseClient(api config.W7ConfigRepositoryInterface, sdk *k8s.Sdk) *LicenseClient {
	return &LicenseClient{w7config: api, sdk: sdk}
}

func (api *LicenseClient) GetLicense() (*License, error) {
	secret, err := api.sdk.GetLicense()
	if err != nil {
		return nil, err
	}
	var x509Certificate *x509.Certificate = nil
	if len(secret.Data["license"]) > 0 {
		cert, err := helper.ParseX509(secret.Data["license"])
		if err == nil {
			x509Certificate = cert
		}
	}
	return &License{
		AppId:         string(secret.Data["appId"]),
		AppSecret:     string(secret.Data["appSecret"]),
		License:       x509Certificate,
		FounderSaName: string(secret.Data["founderSaName"]),
	}, nil
}

func (api *LicenseClient) SetLicense(license *License) error {
	return api.setLicenseClean(license, false)
}

func (api *LicenseClient) CleanLicense() error {
	license, err := api.GetLicense()
	if err != nil {
		return err
	}
	return api.setLicenseClean(license, true)
}

func (api *LicenseClient) setLicenseClean(license *License, cleanCert bool) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "license",
			Namespace: "kube-system",
		},
		Data: map[string][]byte{},
	}

	sigCLient, err := api.sdk.ToSigClient()
	if err != nil {
		return err
	}
	_, err = controllerutil.CreateOrPatch(api.sdk.Ctx, sigCLient, secret, func() error {
		secret.Data["appId"] = []byte(license.AppId)
		secret.Data["appSecret"] = []byte(license.AppSecret)
		secret.Data["founderSaName"] = []byte(license.FounderSaName)
		if license.License != nil {
			secret.Data["license"] = license.License.Raw
		}
		if cleanCert {
			delete(secret.Data, "license")
		}
		return nil
	})
	return err
}

func (c *LicenseClient) CreateLicenseSite(saName string, ignoreExits bool) (*License, error) {
	if !ignoreExits {
		license, err := c.GetLicense()
		if err == nil {
			return license, err
		}
	}
	sa, err := c.sdk.GetServiceAccount(c.sdk.GetNamespace(), saName)
	if err != nil {
		return nil, err
	}
	if sa.Labels["w7.cc/user-mode"] != "founder" {
		return nil, errors.New("非创始人账号无法创建站点")
	}
	clusterUniquId, err := c.sdk.GetClusterId()
	if err != nil {
		return nil, err
	}

	config, err := c.w7config.Get(saName)
	if err != nil {
		slog.Error("获取配置失败")
		return nil, err
	}
	token := config.ThirdpartyCDToken
	api := NewConsoleCdClient(token)
	data := map[string]string{
		"offlineUrl": config.OfflineUrl,
		"sn":         clusterUniquId,
	}
	license, err := api.CreateLicenseSite(data)
	if err != nil {
		return nil, err
	}
	license.FounderSaName = saName //创始人账号
	SetCurrentLicense(license)
	// console.License = license
	err = c.SetLicense(license)
	if err != nil {
		return nil, err
	}
	return license, nil

}

func (c *LicenseClient) ImportCert(pemData []byte, saName string) error {
	license, err := c.GetLicense()
	if err != nil {
		return err
	}

	cert, err := helper.ParseX509(pemData)
	if err != nil {
		return err
	}
	// if cert.Issuer.Organization != nil && len(cert.Issuer.Organization) > 0 {
	// 	if cert.Issuer.Organization[0] != license.AppId {
	// 		return errors.New("证书颁发者与授权不符")
	// 	}
	// }
	_, err = VerifyCert(cert)
	if err != nil {
		return err
	}
	// 兼容w7config license先
	w7config, err := c.w7config.Get(saName)
	if err != nil {
		slog.Error("获取配置失败")
	}
	w7config.License = cert
	err = c.w7config.Set(w7config)
	if err != nil {
		return err
	}
	config.MainW7Config = w7config
	license.License = cert
	// config.LicenseVerify = true
	if cert != nil && len(cert.Subject.Province) > 0 {
		// config.LicenseType = cert.Subject.Province[0]
		// config.SetVerifyType(w7config.Name, cert.Subject.Province[0])

	}

	err = c.SetLicense(license)
	if err != nil {
		return err
	}
	SetCurrentLicense(license)
	return nil

}

func (c *LicenseClient) GetConfig(saName string) (*config.W7Config, error) {
	return c.w7config.Get(saName)
}
