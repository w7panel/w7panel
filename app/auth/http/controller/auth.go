package controller

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service"
	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	saLogic "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/sa"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Auth struct {
	controller.Abstract
}

func (self Auth) Login(http *gin.Context) {
	captchaEnabled := facade.Config.GetBool("captcha.enabled")

	type ParamsValidate struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
		Point    string `form:"point"`
		Key      string `form:"key"`
	}

	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	if captchaEnabled {
		if params.Point == "" || params.Key == "" {
			self.JsonResponseWithError(http, errors.New("验证码参数缺失"), 500)
			return
		}
		err := helper.VerifyCaptcha(params.Point, params.Key, true)
		if err != nil {
			err2 := fmt.Errorf("验证码不正确")
			self.JsonResponseWithError(http, err2, 500)
			return
		}
	}

	client := k8s.NewK8sClient()
	sa, err := client.Login2(params.Username, params.Password, true)
	if err != nil {
		// err2 := fmt.Errorf("用户名密码不正确.")
		if k8serrors.IsNotFound(err) {
			self.JsonResponseWithError(http, errors.New("用户不存在"), 500)
			return
		}
		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.dologin(client.Sdk, sa, http, true)

}

func (self Auth) Register(http *gin.Context) {
	type ParamsValidate struct {
		Username   string `form:"username" binding:"required"`
		Password   string `form:"password" binding:"required"`
		PolicyName string `form:"policyName" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	sdk := k8s.NewK8sClient().Sdk
	_, err := saLogic.DoRegisterLink(sdk, params.Username, params.Password, params.PolicyName)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.JsonSuccessResponse(http)
}

// http://127.0.0.1:9007/k8s/console/oauth?redirect_uri=http://127.0.0.1:9007/k8s/console/login
func (self Auth) ConsoleLogin(http *gin.Context) {
	type ParamsValidate struct {
		Code       string `form:"code" binding:"required"`
		PolicyName string `form:"policyName"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	sdk := k8s.NewK8sClient().Sdk
	w7respo := config.NewW7ConfigRepository(sdk)
	client := console.DefaultClient(false)
	oclient := console.NewOauthClient(client, w7respo)
	accessToken, userInfo, err := oclient.GetUserInfo(params.Code)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}

	w7config, err := w7respo.GetByConsoleId(strconv.Itoa(userInfo.UserId))
	if err != nil {
		//尝试注册用户
		sa, err := saLogic.DoRegister(sdk, types.NewConsoleOAuthAccessToken2(accessToken), userInfo, params.PolicyName)
		if err != nil {
			if !k8serrors.IsAlreadyExists(err) {
				self.JsonResponseWithError(http, err, 500)
				return
			}
		}
		_, err = oclient.BindUseAccessToken(sa.Name, accessToken)
		if err != nil {
			err = sdk.ClientSet.CoreV1().ServiceAccounts(sa.Namespace).Delete(sdk.Ctx, sa.Name, metav1.DeleteOptions{})
			if err != nil {
				slog.Error("删除serviceaccount失败", "err", err)
			}
			self.JsonResponseWithError(http, err, 500)
			return
		}

		self.dologin(sdk, sa, http, false)
		return
	}
	saName := w7config.Name
	_, err = oclient.BindUseAccessToken(saName, accessToken)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	sa, err := sdk.GetServiceAccount(sdk.GetNamespace(), saName)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.dologin(sdk, sa, http, true)

}

func (self Auth) dologin(sdk *k8s.Sdk, sa *corev1.ServiceAccount, http *gin.Context, updateK3kUser bool) {
	seconds := facade.Config.GetInt64("app.login_seconds")
	token, isK3kUser, err := k3k.LoginByServiceAccount(sdk, sa, seconds, updateK3kUser)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			self.JsonResponseWithError(http, errors.New("用户不存在"), 500)
			return
		}
		self.JsonResponseWithError(http, err, 500)
		return
	}
	rs := service.GetRefreshToken(sa.Name)
	self.JsonResponseWithoutError(http, gin.H{
		"token":         token,
		"expire":        time.Now().Add(time.Duration(seconds) * time.Second).Unix(),
		"isK3kUser":     isK3kUser, //废弃		废弃字段，后续删除
		"isClusterUser": isK3kUser,
		"refreshToken":  rs.Token,
	})
	return
}

func (self Auth) RefreshToken(http *gin.Context) {
	type ParamsValidate struct {
		Namespace string `form:"namespace" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	oldToken := http.MustGet("k8s_token").(string)
	oldTokenObj := k8s.NewK8sToken(oldToken)
	saName, _ := k8s.GetTokenSaName(oldToken)
	if saName == "" {
		self.JsonResponseWithError(http, fmt.Errorf("not sa name"), 500)
		return
	}
	seconds := facade.Config.GetInt64("app.login_seconds")
	ans, err := oldTokenObj.GetAudience()
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	token, err := k8s.NewK8sClient().CreateTokenRequest(saName, seconds, ans)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.JsonResponseWithoutError(http, gin.H{
		"token":  token,
		"expire": time.Now().Add(time.Duration(seconds) * time.Second).Unix(),
	})
	return
}

func (self Auth) RefreshToken2(http *gin.Context) {
	type ParamsValidate struct {
		Token string `form:"refreshToken" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	userName, err := service.FindUsernameByToken(params.Token)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	sdk := k8s.NewK8sClient().Sdk
	sa, err := sdk.GetServiceAccount(sdk.GetNamespace(), userName)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.dologin(sdk, sa, http, true)
}

func (self Auth) InitUser(http *gin.Context) {

	releaseName := facade.Config.GetString("app.helm_release_name")
	type ParamsValidate struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	configMapName := releaseName + "-init-user"
	client := k8s.NewK8sClient()
	_, err := client.ClientSet.CoreV1().ConfigMaps(client.GetNamespace()).Get(http, configMapName, v1.GetOptions{})
	if err != nil {
		self.JsonResponseWithError(http, errors.New("已经初始化过用户"), 500)
		return
	}
	err = client.Register(params.Username, params.Password, client.GetNamespace(), "cluster-admin", true, "founder")
	if err != nil {
		self.JsonResponseWithError(http, errors.New("初始化用户失败"), 500)
		return
	}
	client.ClientSet.CoreV1().ConfigMaps(client.GetNamespace()).Delete(http, configMapName, metav1.DeleteOptions{})
	self.JsonSuccessResponse(http)

}

// kubectl create configmap offlineui-init-user --from-literal=a=b
func (self Auth) GetInitUser(http *gin.Context) {
	releaseName := facade.Config.GetString("app.helm_release_name")
	client := k8s.NewK8sClient()
	_, err := client.ClientSet.CoreV1().ConfigMaps(client.GetNamespace()).Get(http, releaseName+"-init-user", v1.GetOptions{})
	maps := make(map[string]string)
	maps["canInitUser"] = "true"
	maps["allowConsoleRegister"] = "false"
	maps["captchaEnabled"] = "false"
	if facade.Config.GetBool("captcha.enabled") {
		maps["captchaEnabled"] = "true"
	}
	if err != nil {
		maps["canInitUser"] = "false"
	}

	k3kconfig, err := client.ClientSet.CoreV1().ConfigMaps("kube-system").Get(http, "k3k.config", v1.GetOptions{})
	if err == nil {
		maps["allowConsoleRegister"] = k3kconfig.Data["allowConsoleRegister"]
	}
	self.JsonResponseWithoutError(http, maps)
	return
}

/*
*
获取用户信息
*/

/*
*
重置密码功能
*/
func (self Auth) ResetPassword(http *gin.Context) {

	type ParamsValidate struct {
		Username    string `form:"username" binding:"required"`
		Password    string `form:"password" binding:"required"`
		NewPassword string `form:"newPassword" binding:"required"`
	}

	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	sdk := k8s.NewK8sClient() //全局用户sdk

	// username := client.GetServiceAccountName()

	_, err := sdk.Login(params.Username, params.Password, false, facade.Config.GetInt64("app.login_seconds"))
	if err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("原始用户密码错误"), 500)
		return
	}
	sa, err := sdk.GetServiceAccount(sdk.GetNamespace(), params.Username)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	err = sdk.ResetPassword(params.Username, params.NewPassword, sa.Labels["w7.cc/user-mode"])
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.JsonResponseWithoutError(http, map[string]string{})
}

func (self Auth) ResetPasswordCurrent(http *gin.Context) {

	type ParamsValidate struct {
		Password    string `form:"password"`
		NewPassword string `form:"newPassword" binding:"required"`
	}

	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	sdk := k8s.NewK8sClient() //全局用户sdk
	token := http.MustGet("k8s_token").(string)
	k8sToken := k8s.NewK8sToken(token)
	userName, err := k8sToken.GetSaName()
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	// username := client.GetServiceAccountName()
	sa, err := sdk.GetServiceAccount("default", userName)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	if sa.Annotations != nil {
		sa.Annotations = map[string]string{}
	}
	if sa.Annotations["password"] != "" {
		_, err := sdk.Login(userName, params.Password, false, facade.Config.GetInt64("app.login_seconds"))
		if err != nil {
			self.JsonResponseWithError(http, fmt.Errorf("原始用户密码错误"), 500)
			return
		}
	}

	err = sdk.ResetPassword(userName, params.NewPassword, sa.Labels["w7.cc/user-mode"])
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.JsonResponseWithoutError(http, map[string]string{})
}
