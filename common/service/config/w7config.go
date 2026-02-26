package config

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	service "github.com/w7corp/sdk-open-cloud-go/service"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	CurrentCity, _ = helper.MyCity() // 当前服务器所在的城市
	// CurrentCity = "北京"
}

type licenseVerify struct {
	sync.Map
}

func (l *licenseVerify) IsCompany(name string) bool {
	value, ok := l.Load(name)
	if ok {
		return value.(string) == "company"
	}
	return false
}

func (l *licenseVerify) IsTeam(name string) bool {
	value, ok := l.Load(name)
	if ok {
		return value.(string) == "team"
	}
	return false
}

// var LicenseVerify sync.Map

var LicenseVerify = licenseVerify{}

// company team类型 取ServiceAccount 其中一个
var MainW7Config *W7Config

var CurrentCity string

// func IsLicenseVerify(name string) bool {
// 	return LicenseVerify.IsCompany(name) || LicenseVerify.IsTeam(name)
// }
// func IsCompany(name string) bool {
// 	return LicenseVerify.IsCompany(name)
// }
// func IsTeam(name string) bool {
// 	return LicenseVerify.IsTeam(name)
// }

func SetVerifyType(name, verifyType string) {
	LicenseVerify.Store(name, verifyType)
}

type W7Config struct {
	Name              string                  `json:"name"`
	ThirdpartyCDToken string                  `json:"thirdparty_cd_token"`
	CDTokenExpireTime int                     `json:"cd_token_expire_time"`
	ClusterId         string                  `json:"cluster_id"`
	OfflineUrl        string                  `json:"offline_url"`
	AccessToken       string                  `json:"access_token"`
	ExpireTime        int                     `json:"expire_time"`
	ApiServerUrl      string                  `json:"api_server_url"`
	UserInfo          *service.ResultUserinfo `json:"user_info"`
	License           *x509.Certificate       `json:"license"`
	DebugValue        string                  `json:"debug_value"`
}

func NewEmptyConfig() *W7Config {
	return &W7Config{
		ThirdpartyCDToken: "",
		ClusterId:         "",
		AccessToken:       "",
		ExpireTime:        0,
		ApiServerUrl:      "",
	}
}

func (c *W7Config) Clone() *W7Config {
	return &W7Config{
		Name:              c.Name,
		ThirdpartyCDToken: c.ThirdpartyCDToken,
		CDTokenExpireTime: c.CDTokenExpireTime,
		ClusterId:         c.ClusterId,
		OfflineUrl:        c.OfflineUrl,
		AccessToken:       c.AccessToken,
		ExpireTime:        c.ExpireTime,
		ApiServerUrl:      c.ApiServerUrl,
		UserInfo:          c.UserInfo,
		License:           c.License,
		DebugValue:        c.DebugValue,
	}
}

func (c *W7Config) IsWillExpired() bool {
	return int64(c.ExpireTime) < time.Now().Unix()+300
}

func (c *W7Config) IsExpired() bool {
	return int64(c.ExpireTime) < time.Now().Unix()
}
func (c *W7Config) IsCDTokenWillExpired() bool {
	return int64(c.CDTokenExpireTime)-600 < time.Now().Unix()
}

func (c *W7Config) GetLicenseType() string {
	license := c.License
	licenseType := "free"
	isExpired := false
	if license != nil && len(license.Subject.Province) > 0 {
		licenseType = license.Subject.Province[0]
	}
	if license != nil {
		// endTime = license.NotAfter
		if license.NotAfter.Before(time.Now()) {
			isExpired = true
		}
		if isExpired {
			licenseType = "free"
		}
	}
	return licenseType
}

func (c *W7Config) NotFree() bool {
	licenseType := c.GetLicenseType()
	return licenseType == "team" || licenseType == "company"
}

func (c *W7Config) ToArray() map[string]interface{} {
	license := c.License
	licenseType := "free"
	licenseId := "0"
	if license != nil && len(license.Subject.Province) > 0 {
		licenseType = license.Subject.Province[0]
	}
	endTime := time.Now()
	isExpired := false
	if license != nil {
		licenseId = license.SerialNumber.String()
		endTime = license.NotAfter
		if license.NotAfter.Before(time.Now()) {
			isExpired = true
		}
		if isExpired {
			licenseType = "free"
		}
	}

	return map[string]interface{}{
		"thirdparty_cd_token": c.ThirdpartyCDToken,
		"cluster_id":          c.ClusterId,
		"offline_url":         c.OfflineUrl,
		"access_token":        c.AccessToken,
		"expire_time":         c.ExpireTime,
		"api_server_url":      c.ApiServerUrl,
		"require_oauth":       c.UserInfo == nil,
		"is_register":         c.ClusterId != "",
		"userinfo":            c.UserInfo,
		"license_type":        licenseType,
		"license_id":          licenseId,
		"license_end_time":    endTime.Format("2006-01-02 15:04:05"),
		"license_is_expired":  isExpired,
		"debug_value":         c.DebugValue,
		// "license":             c.License.Raw,
	}
}

type W7ConfigRepositoryInterface interface {
	Get(name string) (*W7Config, error)
	Set(w7config *W7Config) error
	GetByConsoleId(consoleId string) (*W7Config, error)

	List() ([]*W7Config, error)
}

type w7ConfigRepository struct {
	*k8s.Sdk
}

func NewW7ConfigRepository(sdk *k8s.Sdk) *w7ConfigRepository {
	return &w7ConfigRepository{sdk}
}

func secretName(name string) string {
	return name + ".w7-config"
}

func (c *w7ConfigRepository) secretToW7config(secret *v1.Secret, name string) *W7Config {
	expireTimeStr := string(secret.Data["expire_time"])
	expireTime, err := strconv.Atoi(expireTimeStr)
	if err != nil {
		expireTime = 0
		//return W7Config{}, fmt.Errorf("failed to convert expire_time to int: %w", err)
	}
	cdExpireTime, err := strconv.Atoi(string(secret.Data["cd_token_expire_time"]))
	if err != nil {
		cdExpireTime = 0
	}
	userInfoStr, ok := secret.Data["userinfo"]
	userInfo := service.ResultUserinfo{}
	if ok {
		err = json.Unmarshal(userInfoStr, &userInfo)
		if err != nil {
			slog.Warn("failed to unmarshal userinfo: %w", "err", err)
		}
	}
	var x509Certificate *x509.Certificate = nil
	if len(secret.Data["license"]) > 0 {
		cert, err := helper.ParseX509(secret.Data["license"])
		if err == nil {
			x509Certificate = cert
		}
	}

	return &W7Config{
		Name:              name,
		ThirdpartyCDToken: string(secret.Data["thirdparty_cd_token"]),
		CDTokenExpireTime: cdExpireTime,
		AccessToken:       string(secret.Data["access_token"]),
		ExpireTime:        expireTime,
		ClusterId:         string(secret.Data["cluster_id"]),
		OfflineUrl:        string(secret.Data["offline_url"]),
		UserInfo:          &userInfo,
		License:           x509Certificate,
		DebugValue:        string(secret.Labels["w7.cc/test"]),
	}
}

func (c *w7ConfigRepository) Get(name string) (*W7Config, error) {
	secretName := secretName(name)
	secret, err1 := c.ClientSet.CoreV1().Secrets(c.GetNamespace()).Get(c.Ctx, secretName, metav1.GetOptions{})
	if err1 != nil {
		slog.Warn("failed to get w7-config: ")
		return &W7Config{Name: name}, fmt.Errorf("failed to get w7-config: %w", err1)
	}
	return c.secretToW7config(secret, name), nil

}

func (c *w7ConfigRepository) List() ([]*W7Config, error) {
	secrets, err := c.ClientSet.CoreV1().Secrets(c.GetNamespace()).List(c.Ctx, metav1.ListOptions{
		LabelSelector: "w7.cc/oauth-config=true",
	})
	if err != nil {
		return []*W7Config{}, fmt.Errorf("failed to list secrets: %w", err)
	}
	configs := []*W7Config{}
	for _, secret := range secrets.Items {
		if strings.HasSuffix(secret.Name, ".w7-config") {
			configs = append(configs, c.secretToW7config(&secret, strings.TrimSuffix(secret.Name, ".w7-config")))
		}
	}
	return configs, nil
}

func (c *w7ConfigRepository) GetByConsoleId(consoleId string) (*W7Config, error) {
	secrets, err := c.ClientSet.CoreV1().Secrets(c.GetNamespace()).List(c.Ctx, metav1.ListOptions{
		LabelSelector: "w7.cc/console-uid=" + consoleId,
	})
	if err != nil {
		return &W7Config{}, fmt.Errorf("failed to list secrets: %w", err)
	}
	for _, secret := range secrets.Items {
		if strings.HasSuffix(secret.Name, ".w7-config") {
			return c.secretToW7config(&secret, strings.TrimSuffix(secret.Name, ".w7-config")), nil
		}
	}
	return &W7Config{}, fmt.Errorf("not found w7 config for console id: %s", consoleId)
}

func (c *w7ConfigRepository) Set(config *W7Config) error {
	secretName := secretName(config.Name)
	secret, err := c.ClientSet.CoreV1().Secrets(c.GetNamespace()).Get(c.Ctx, secretName, metav1.GetOptions{})
	isUpdate := true
	if err != nil {
		secret = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: c.GetNamespace(),
			},
			Data: map[string][]byte{},
			Type: v1.SecretTypeOpaque,
		}
		isUpdate = false
	}
	if secret.Labels == nil {
		secret.Labels = make(map[string]string)
	}
	secret.Labels["w7.cc/oauth-config"] = "true"
	secret.Labels["w7.cc/test"] = helper.RandomString(5)
	secret.Data["thirdparty_cd_token"] = []byte(config.ThirdpartyCDToken)
	secret.Data["access_token"] = []byte(config.AccessToken)
	if config.ExpireTime == 0 {
		secret.Data["expire_time"] = []byte("0")
	}
	if config.CDTokenExpireTime > 0 {
		secret.Data["cd_token_expire_time"] = []byte(strconv.Itoa(config.CDTokenExpireTime))
	}
	secret.Data["expire_time"] = []byte(strconv.Itoa(config.ExpireTime))
	secret.Data["cluster_id"] = []byte(config.ClusterId)
	secret.Data["offline_url"] = []byte(config.OfflineUrl)
	if config.UserInfo != nil {
		userInfoBytes, err := json.Marshal(config.UserInfo)
		secret.Labels["w7.cc/console-uid"] = strconv.Itoa(config.UserInfo.UserId)
		if err == nil {
			secret.Data["userinfo"] = userInfoBytes
		}
	}
	if (config.License != nil) && len(config.License.Raw) > 0 {
		secret.Data["license"] = config.License.Raw
	}
	if !isUpdate {
		_, err = c.ClientSet.CoreV1().Secrets(c.GetNamespace()).Create(c.Ctx, secret, metav1.CreateOptions{})
	} else {
		_, err = c.ClientSet.CoreV1().Secrets(c.GetNamespace()).Update(c.Ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			slog.Warn("failed to update w7-config: %w", "err", err)
		}
	}
	return err
}
