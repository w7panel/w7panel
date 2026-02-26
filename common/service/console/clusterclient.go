package console

import (
	"errors"
	"log/slog"

	"strconv"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

type clusterClient struct {
	w7config          config.W7ConfigRepositoryInterface
	sdk               *k8s.Sdk
	kubeConfig        *clientcmdv1.Config
	thirdPartyCDToken string
	offlineUrl        string
}

func NewClusterClient(api config.W7ConfigRepositoryInterface, sdk *k8s.Sdk, kubeconfig *clientcmdv1.Config) *clusterClient {
	return &clusterClient{w7config: api, sdk: sdk, kubeConfig: kubeconfig}
}

func (c *clusterClient) SetThirdPartyCdTDken(token string) {
	c.thirdPartyCDToken = token
}

func (c *clusterClient) SetOfflineUrl(url string) {
	c.offlineUrl = url
}

func (c *clusterClient) RegisterUseAccessToken(isK3k bool, k3kUserName string) error {
	config, err := c.w7config.Get(k3kUserName)
	if err != nil {
		slog.Error("获取配置失败", "err", err)
	}
	if config.AccessToken == "" {
		slog.Error("请先绑定微擎账户")
		return errors.New("请先绑定微擎账户")
	}
	token, err := AccessTokenToCDToken(config.AccessToken)
	if err != nil {
		return err
	}
	return c.RegisterUseCdTokenId(token.Token, config.ClusterId, isK3k, k3kUserName)
}

func (c *clusterClient) UnRegister(saName string) error {
	config, err := c.w7config.Get(saName)
	if err != nil {
		slog.Error("获取配置失败", "err", err)
	}
	if config.AccessToken == "" {
		slog.Error("请先绑定微擎账户")
		return errors.New("请先绑定微擎账户")
	}
	token, err := AccessTokenToCDToken(config.AccessToken)
	if err != nil {
		return err
	}
	id := config.ClusterId
	if (id == "") || (id == "0") {
		return errors.New("用户未注册到交付系统")
	}
	api := NewConsoleCdClient(token.Token)
	err = api.DeleteCluster(id)
	if err != nil {
		return err
	}
	config.ClusterId = ""
	return c.w7config.Set(config)

}

func (c *clusterClient) RegisterUseCdToken(isK3k bool, k3kUserName string) error {
	config, err := c.w7config.Get(k3kUserName)
	if err != nil {
		slog.Error("获取配置失败")
	}
	token := c.thirdPartyCDToken
	if token == "" {
		token = config.ThirdpartyCDToken
	}

	if token == "" {
		slog.Error("无法获取交付系统token")
		return errors.New("无法获取交付系统token")
	}
	return c.RegisterUseCdTokenId(token, config.ClusterId, isK3k, k3kUserName)
}

func (c *clusterClient) ImportCert(pemData []byte, saName string) error {
	w7config, err := c.w7config.Get(saName)
	if err != nil {
		slog.Error("获取配置失败")
	}
	cert, err := helper.ParseX509(pemData)
	if err != nil {
		return err
	}
	_, err = VerifyCert(cert)
	if err != nil {
		return err
	}
	w7config.License = cert
	// config.LicenseVerify = true
	if cert != nil && len(cert.Subject.Province) > 0 {
		// config.LicenseType = cert.Subject.Province[0]
		config.SetVerifyType(w7config.Name, cert.Subject.Province[0])

	}

	return c.w7config.Set(w7config)
}

func (c *clusterClient) RegisterUseCdTokenId(token string, id string, isK3k bool, k3kUserName string) error {
	config, _ := c.w7config.Get(k3kUserName)
	api := NewConsoleCdClient(token)
	kubeconfig, err := c.GetKubeConfig(k3kUserName)
	if err != nil {
		return err
	}
	// slog.Info(string(kubeconfig))
	cluster, err := api.CreateOrUpdateCluster(kubeconfig, id, c.offlineUrl, isK3k, k3kUserName)
	if err != nil {
		return err
	}
	if cluster.Id > 0 {
		config.ClusterId = strconv.Itoa(cluster.Id)
		config.OfflineUrl = c.offlineUrl
		return c.w7config.Set(config)
	}
	return nil

}

func (c *clusterClient) GetKubeConfig(saName string) ([]byte, error) {
	if c.kubeConfig == nil {
		w7config, err := c.w7config.Get(saName)
		if err != nil {
			w7config = config.NewEmptyConfig()
		}
		apiServerUrl := w7config.ApiServerUrl
		kubeconfig, err := c.sdk.ToKubeconfig(apiServerUrl)
		if err != nil {
			return nil, err
		}
		c.kubeConfig = kubeconfig
	}

	return helper.K8sObjToYaml(c.kubeConfig)
}
