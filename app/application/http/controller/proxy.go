package controller

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/microapp"
	microappType "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/microapp/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Proxy struct {
	controller.Abstract
}

func (self Proxy) ProxyK8s(http *gin.Context) {
	// 提取并归一化路径：移除 /k8s-proxy 或 /k8s 前缀
	path := http.Param("path")
	if path == "" {
		path = http.Request.URL.Path
	}
	// 移除前缀
	if strings.HasPrefix(path, "/k8s-proxy") {
		path = strings.TrimPrefix(path, "/k8s-proxy")
	} else if strings.HasPrefix(path, "/k8s") {
		path = strings.TrimPrefix(path, "/k8s")
	}
	if path == "" {
		path = "/"
	}

	// 修改请求路径
	http.Request.URL.Path = path
	http.Request.URL.RawPath = ""

	// 获取 token
	token := http.MustGet("k8s_token").(string)
	local := http.Query("local")
	forceLocal := false
	if local == "true" || local == "1" {
		forceLocal = true
	}
	useRootSdk := false
	if strings.HasPrefix(http.Request.URL.Path, "/api/v1/namespaces/default/configmaps/domain-parse") && http.Request.Method == "GET" {
		// forceLocal = true
		useRootSdk = true
	}
	// 创建 K8s 客户端
	client, err := k8s.NewK8sClient().ChannelLocal(token, forceLocal)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	if useRootSdk {
		client = k8s.NewK8sClient().Sdk
	}

	// 检查并修改 http.Request 中的 Authorization 头部
	auth := http.Request.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		// bearerToken := strings.TrimPrefix(auth, "Bearer ")
		restConfig, err := client.ToRESTConfig()
		if err != nil {
			self.JsonResponseWithServerError(http, err)
			return
		}
		http.Request.Header.Set("Authorization", "Bearer "+restConfig.BearerToken)
		// if bearerToken != token {
		// 	// 如果 Bearer 令牌与 token 不一致，修改 Authorization 头部
		// 	http.Request.Header.Set("Authorization", "Bearer "+token)
		// }
	} else if auth != "" {
		// 如果 Authorization 头部不是 Bearer 类型，考虑是否需要清除或处理
		// 根据实际需求决定，这里假设我们清除它并使用 token
		http.Request.Header.Set("Authorization", "Bearer "+token)
	}

	// 处理请求 - K8sResponseFilter middleware handles filtering
	err = client.Proxy(http.Request, http.Writer)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
}

func (self Proxy) ProxyDebug(gin *gin.Context) {
	result := make(map[string]interface{})
	result["fullPath"] = gin.FullPath()
	result["path"] = gin.Request.URL.Path
	result["query"] = gin.Request.URL.Query()
	result["header"] = gin.Request.Header
	result["host"] = gin.Request.Host
	result["joinPath"] = gin.Request.URL.JoinPath("/").RequestURI()
	gin.JSON(http.StatusOK, result)
}

/*
*

	@Description: 转发请求
	proxyUrl: 代理地址
*/
func (self Proxy) Proxy(gin *gin.Context) {
	proxyUrl := gin.GetHeader("proxy-url")
	if proxyUrl == "" {
		proxyUrl = "https://zpk.w7.cc"
		// self.JsonResponseWithServerError(gin, errors.New("proxy-url is required"))
		// return
	}

	var path = gin.Param("path")
	if path == "" {
		path = "/"
	}

	self.proxyUrl(gin, proxyUrl, path)
}
func (self Proxy) ProxyNoAuthService(gin *gin.Context) {
	self.ProxyService(gin)
}

// 转发k8s
func (self Proxy) ProxyService(gin *gin.Context) {
	ns := gin.Param("namespace")
	name := gin.Param("name")
	path := gin.Param("path")
	if path == "" {
		path = "/"
	}

	//name split : port
	schema, host, port := self.parseName(name)
	cdomain := helper.ClusterDomain(host, ns)
	proxyUrl := schema + "://" + cdomain + ":" + port

	if facade.GetConfig().GetBool("longhorn.mock") {
		// proxyUrl = "http://218.23.2.55:9090/"
	}
	// token := gin.GetString("k8s_token")
	// if token != "" { //proxy-no 不传递token
	// 	k8sToken := k8s.NewK8sToken(token)
	// 	if k8sToken.IsVirtual() {
	// 		client, err := k8s.NewK8sClient().Channel(token)
	// 		if err != nil {
	// 			self.JsonResponseWithServerError(gin, err)
	// 			return
	// 		}
	// 		client.Proxy(gin.Request, gin.Writer)
	// 		return
	// 	}
	// }

	self.proxyUrl(gin, proxyUrl, path)
}

func (self Proxy) ProxyPod(gin *gin.Context) {
	ns := gin.Param("namespace")
	name := gin.Param("name")
	path := gin.Param("path")
	if path == "" {
		path = "/"
	}
	//name split : port
	namePort := strings.Split(name, ":")
	podName := namePort[0]
	podPort := "80"
	if len(namePort) > 1 {
		podPort = namePort[1]
	}

	client, err := k8s.NewK8sClient().Channel(gin.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(gin, err)
		return
	}
	pod, err := client.ClientSet.CoreV1().Pods(ns).Get(client.Ctx, podName, v1.GetOptions{})
	if err != nil {
		self.JsonResponseWithServerError(gin, err)
		return
	}
	podId := pod.Status.PodIP
	if podId == "" {
		self.JsonResponseWithServerError(gin, errors.New("pod ip is empty"))
		return
	}

	proxyUrl := "http://" + podId + ":" + podPort

	self.proxyUrl(gin, proxyUrl, path)
}

func (self Proxy) ProxyCommon(gin *gin.Context) {

	name := gin.Param("name")
	path := gin.Param("path")
	if path == "" {
		path = "/"
	}
	schema, host, port := self.parseName(name)

	proxyUrl := schema + "://" + host + ":" + port
	slog.Info("proxyUrl:" + proxyUrl)

	self.proxyUrl(gin, proxyUrl, path)
}

func (self Proxy) proxyUrl(gin *gin.Context, proxyUrl string, path string) {
	remote, err := url.Parse(proxyUrl)
	if err != nil {
		self.JsonResponseWithServerError(gin, err)
		return
	}
	proxyPath := remote.Path
	if path == "" {
		path = proxyPath
	}

	// 标记已处理，阻止后续中间件
	gin.Abort()

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = path

		// 处理 WebDAV Destination 头
		if dest := req.Header.Get("Destination"); dest != "" {
			if destURL, err := url.Parse(dest); err == nil {
				destPath := destURL.Path
				if idx := strings.Index(destPath, "/k8s/webdav"); idx >= 0 {
					destPath = destPath[idx:]
				} else if idx := strings.Index(destPath, "/webdav"); idx >= 0 {
					destPath = destPath[idx:]
				}
				if destURL.RawQuery != "" {
					destPath += "?" + destURL.RawQuery
				}
				req.Header.Set("Destination", destPath)
				slog.Info("Rewrote Destination header", "original", dest, "new", destPath)
			}
		}
	}
	proxy.ModifyResponse = func(res *http.Response) error {
		res.Header.Del("Access-Control-Allow-Origin")
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("Proxy error", "error", err, "path", r.URL.Path)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprintf(`{"code":502,"error":"%s"}`, err.Error())))
	}

	// 使用 defer recover 捕获可能的 panic
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic in proxy", "error", r)
		}
	}()

	proxy.ServeHTTP(gin.Writer, gin.Request)
}

func (self Proxy) ProxyAddr(http *gin.Context) {
	type ParamsValidate struct {
		ProxyUrl string `form:"proxyUrl" binding:"required`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	res, err := helper.RetryHttpClient().R().Get(params.ProxyUrl)
	if err != nil {
		http.String(200, "")
		return
	}
	http.String(200, res.String())
	// res.Body().Close()
	// uri, err := url.Parse(params.ProxyUrl)
	// if err != nil {
	// 	self.JsonResponseWithServerError(http, err)
	// 	return
	// }

	// self.proxyUrl(http, params.ProxyUrl, "")
}

func (self Proxy) Kubeconfig(gin *gin.Context) {
	client, err := k8s.NewK8sClient().Channel(gin.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(gin, err)
		return
	}
	apiServerUrl := gin.Query("apiServerUrl")
	config, err := client.ToKubeconfig(apiServerUrl)
	if err != nil {
		self.JsonResponseWithServerError(gin, err)
		return
	}
	self.JsonResponseWithoutError(gin, config)

}

func (self Proxy) parseName(name string) (string, string, string) {
	namePort := strings.Split(name, ":")
	host := namePort[0]
	port := "80"
	schema := "http"
	if len(namePort) == 3 {
		schema = namePort[0]
		host = namePort[1]
		port = namePort[2]
	}
	if len(namePort) == 2 {
		host = namePort[0]
		port = namePort[1]
	}
	if len(namePort) == 1 {
		host = namePort[0]
	}
	return schema, host, port
}

func (self Proxy) ProxyMicroApp(gin *gin.Context) {

	name := gin.Param("name")
	path := gin.Param("path")

	token := gin.MustGet("k8s_token").(string)

	k3kuser, err := k3k.TokenToK3kUser(token)
	if err != nil {
		self.JsonResponseWithServerError(gin, err)
		return
	}
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(gin, err)
		return
	}
	microAppObj := &microappType.MicroApp{}
	sigclient, err := client.ToSigClient()
	if err != nil {
		self.JsonResponseWithServerError(gin, err)
		return
	}
	err = sigclient.Get(client.Ctx, types.NamespacedName{Name: name, Namespace: k3kuser.GetNamespace()}, microAppObj)
	if err != nil {
		self.JsonResponseWithServerError(gin, err)
		return
	}
	role := k3kuser.GetRole()
	// if role != "founder" && role != "admin" {
	// 	self.JsonResponseWithServerError(gin, errors.New("无权限访问"))
	// 	return
	// }
	if microAppObj.IsFromRoot() || !k3kuser.IsClusterUser() {
		if helper.IsK3kVirtual() {
			role = "founder"
		}
		proxy := microapp.NewMicroAppProxy(microAppObj, k3kuser.IsClusterUser(), role)
		revert, err := proxy.Proxy(path)
		if err != nil {
			self.JsonResponseWithServerError(gin, err)
			return
		}
		revert.ServeHTTP(gin.Writer, gin.Request)
		return
	}
	if k3kuser.IsClusterUser() {

		k8stoken := k8s.NewK8sToken(token)
		config, err := k8stoken.GetK3kConfig()
		if err != nil {
			self.JsonResponseWithServerError(gin, err)
			return
		}
		path := gin.Request.URL.String()
		// agentHost := config.GetK3kAgentLbHost()
		proxyUrl := "http://" + config.GetK3kAgentLbHost()
		restConfig, err := client.ToRESTConfig()
		if err != nil {
			self.JsonResponseWithServerError(gin, err)
			return
		}
		gin.Request.Header.Set("AuthorizationX", "Bearer "+restConfig.BearerToken)
		proxy, err := helper.ProxyUrl(proxyUrl, path, "", nil, nil)
		if err != nil {
			self.JsonResponseWithServerError(gin, err)
			return
		}
		proxy.ServeHTTP(gin.Writer, gin.Request)

	}
}
