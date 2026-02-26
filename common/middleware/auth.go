package middleware

import (
	"os"
	"path/filepath"
	"sync"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
	"gopkg.in/yaml.v3"
)

var (
	mockToken     string
	mockTokenOnce sync.Once
	mockTokenErr  error
)

type Auth struct {
	middleware.Abstract
	role string
}

func NewAuth(role string) Auth {
	auth := Auth{role: role}
	return auth
}

func (self Auth) Process(ctx *gin.Context) {

	// LOCAL_MOCK 模式下绕过认证（用于本地开发测试）
	if os.Getenv("LOCAL_MOCK") == "true" || os.Getenv("LOCAL_MOCK") == "1" {
		// 尝试从 kubeconfig 读取有效 token
		token := getMockToken()
		ctx.Set("k8s_token", token)
		ctx.Next()
		return
	}

	// 判断是否accept application/json
	// if !strings.Contains(ctx.Request.Header.Get("Accept"), "application/json") && ctx.Request.Method == "GET" && ctx.Request.URL.Path != "/" {
	// 	indexHtml, _ := Asset.ReadFile("asset/index.html")
	// 	ctx.Data(http2.StatusOK, "text/html; charset=UTF-8", indexHtml)
	// 	return
	// }
	// slog.Info("auth middleware request url: ", ctx.Request.URL)

	bearertoken := self.getToken(ctx)
	if bearertoken == "" {
		ctx.AbortWithStatusJSON(401, gin.H{
			"code": 401,
			"msg":  "请登录",
		})
		return
	}
	k8sToken := k8s.NewK8sToken(bearertoken)
	if self.role != "" {
		// if k8sToken.Role() != self.role {
		// 	//没有权限请求
		// 	ctx.AbortWithStatus(403)
		// 	return
		// }
	}
	if k8sToken.IsCacheToken() {
		ctx.Set("k8s_token", bearertoken)
		return
	}

	err := k8s.NewK8sClient().TokenReview(bearertoken)
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{
			"code": 401,
			"msg":  "请登录" + err.Error(),
		})
		return
	}
	k8sToken.Cache()
	if k3k.NeedRelogin(k8sToken) {

	}
	ctx.Set("k8s_token", bearertoken)
	// if facade.Config.GetBool("app.refresh_token_enable") {
	// 	if ctx.Writer.Status() >= http.StatusOK && ctx.Writer.Status() < 300 {
	// 		saName, date := k8s.GetTokenSaName(bearertoken)
	// 		if saName != "" && !date.IsZero() && date.After(time.Now().Add(-time.Minute*10)) {
	// 			token, err := k8s.NewK8sClient().CreateTokenRequest(saName, facade.Config.GetInt64("app.login_seconds"), []string{})
	// 			if err != nil {
	// 				slog.Info("refresh token err: ", "err", err)
	// 				return
	// 			}

	// 			ctx.Writer.Header().Set("access-token", token)
	// 		}
	// 	}
	// }
	ctx.Next()

	// ctx.Writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

func (self Auth) getToken(ctx *gin.Context) string {
	return helper.GetToken(ctx)
}

// getMockToken 从 kubeconfig 读取 token（仅用于 LOCAL_MOCK 模式）
// 使用 sync.Once 缓存结果，避免每次请求都读文件
func getMockToken() string {
	mockTokenOnce.Do(func() {
		kubeconfigPaths := []string{
			"./kubeconfig.yaml",
			"./config/kubeconfig.yaml",
			"./w7panel/kubeconfig.yaml",
		}

		// 优先使用环境变量指定的路径
		if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
			kubeconfigPaths = append([]string{envPath}, kubeconfigPaths...)
		}

		// 尝试基于 KO_DATA_PATH 推断路径
		if koDataPath := os.Getenv("KO_DATA_PATH"); koDataPath != "" {
			kubeconfigPaths = append(kubeconfigPaths, filepath.Join(filepath.Dir(koDataPath), "kubeconfig.yaml"))
		}

		for _, path := range kubeconfigPaths {
			if token := readTokenFromKubeconfig(path); token != "" {
				mockToken = token
				return
			}
		}

		mockToken = "local-mock-token"
	})
	return mockToken
}

// readTokenFromKubeconfig 从 kubeconfig 文件读取 token
func readTokenFromKubeconfig(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var config struct {
		Users []struct {
			User struct {
				Token string `yaml:"token"`
			} `yaml:"user"`
		} `yaml:"users"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return ""
	}

	for _, u := range config.Users {
		if u.User.Token != "" {
			return u.User.Token
		}
	}

	return ""
}
