package console

import (
	"os"

	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	w7 "github.com/w7corp/sdk-open-cloud-go"
	"github.com/w7corp/sdk-open-cloud-go/service"
)

// BindConsoleUserUseAccessToken 绑定用户信息 k3ksacontroller.go 中使用
func BindConsoleUserUseAccessToken(saName string, sdk *k8s.Sdk, accessToken *service.ResultAccessToken) error {
	repository := config.NewW7ConfigRepository(sdk)
	client := DefaultClient(false)
	oclient := NewOauthClient(client, repository)
	_, err := oclient.BindUseAccessToken(saName, accessToken)
	return err

}

type OauthClient struct {
	client *w7.Client
	repo   config.W7ConfigRepositoryInterface
}

func NewOauthClient(client *w7.Client, repo config.W7ConfigRepositoryInterface) *OauthClient {
	// client.SetHttpClient(helper.RetryHttpClient())
	return &OauthClient{
		client: client,
		repo:   repo,
		// repo:   config.NewW7ConfigRepository(),
	}
}

func (c *OauthClient) Bind(code string, saName string) (*service.ResultUserinfo, error) {
	result, errn := c.client.OauthService.GetAccessTokenByCode(code)
	if errn.Errno != 0 {
		return nil, errn.ToError()
	}
	return c.BindUseAccessToken(saName, result)

}

func (c *OauthClient) BindUseAccessToken(saName string, result *service.ResultAccessToken) (*service.ResultUserinfo, error) {
	config, _ := c.repo.Get(saName)

	config.AccessToken = result.AccessToken
	config.ExpireTime = (result.ExpireTime)
	// return c.repo.Set(config)

	userInfo, errno := c.client.OauthService.GetUserInfo(result.AccessToken)
	if errno.Errno != 0 {
		return nil, c.repo.Set(config)
	}
	config.UserInfo = userInfo
	cdToken, err := AccessTokenToCDToken(result.AccessToken)
	if err != nil {
		return nil, c.repo.Set(config)
	}
	config.ThirdpartyCDToken = cdToken.Token
	config.CDTokenExpireTime = (result.ExpireTime)
	err = c.repo.Set(config)
	if err != nil {
		return nil, err
	}
	return userInfo, nil

	// return c.client.Oauth().GetToken(code, )
}

func (c *OauthClient) GetAccessToken(code string) (*service.ResultAccessToken, error) {
	return c.client.OauthService.GetAccessTokenByCode(code)
}

func (c *OauthClient) GetUserInfo(code string) (*service.ResultAccessToken, *service.ResultUserinfo, error) {
	result, errn := c.client.OauthService.GetAccessTokenByCode(code)
	if errn.Errno != 0 {
		return nil, nil, errn.ToError()
	}
	userInfo, errno := c.client.OauthService.GetUserInfo(result.AccessToken)
	if errno.Errno != 0 {
		return result, nil, errno.ToError()
	}
	return result, userInfo, nil
}

// func (c *OauthClient) UnBind() error {
// 	config, _ := c.repo.Get()
// 	config.AccessToken = ""
// 	config.ExpireTime = 0
// 	return c.repo.Set(config)
// }

func DefaultClient(debug bool) *w7.Client {
	// debug = true
	userAgent, ok := os.LookupEnv("USER_AGENT")
	apiUrl := "https://openapi.w7.cc"
	if ok && (userAgent == "we7test-develop" || userAgent == "we7test-beta") {
		debug = true
	}
	if debug {
		apiUrl = "https://api.w7.cc"
	}

	// const APPID = '499430';
	// const APPSECRET = '697a59ecbf0585f305e6c4e800464031';
	//
	// "wac1iw35fng1mtc23c",
	// "d40fCiDygouYBSEReBDu42L6I9CqptyPvAN/ucCpRQ3I0ZHoau1OP8HWCeiDsCE",
	client := w7.NewClient(
		"wac1iw35fng1mtc23c",
		"d40fCiDygouYBSEReBDu42L6I9CqptyPvAN/ucCpRQ3I0ZHoau1OP8HWCeiDsCE",

		w7.Option{
			ApiUrl: apiUrl,
			Debug:  debug,
		})

	if ok {
		client.GetHttpClient().SetHeader("User-Agent", userAgent).EnableTrace()
	}

	return client
}
