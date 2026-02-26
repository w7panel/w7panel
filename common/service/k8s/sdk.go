package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	higressextv1 "gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress/client/pkg/apis/extensions/v1alpha1"
	higressnetworkingv1 "gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress/client/pkg/apis/networking/v1"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/terminal"
	appgroupv1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	microapp "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/microapp/v1alpha1"
	"github.com/gin-gonic/gin"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"github.com/google/uuid"
	k3kv1alpha "github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"golang.org/x/crypto/bcrypt"
	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	apirbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	openapiclient "k8s.io/client-go/openapi"
	"k8s.io/client-go/openapi/cached"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

const namespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
const tokenFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
const K3K_MENU_FOUNDER = `
["cluster","cluster-panel","cluster-resource","app","app-apps","app-apps-add","app-apps-edit","app-apps-delete","app-cronjob","app-cronjob-add","app-cronjob-edit","app-cronjob-delete","app-rvproxy","app-rvproxy-add","app-rvproxy-edit","app-rvproxy-delete","app-dblist","app-dblist-add","app-dblist-delete","app-gpustack","storage","storage-node","storage-node-add","storage-node-edit","storage-node-delete","storage-zone","zpk","system","system-cloud","system-order-center","system-cost-center","cluster-nodes","cluster-nodes-add","cluster-nodes-registries","cluster-nodes-gpu","cluster-nodes-memory","system-whitelist","system-manage","system-user","system-usergroup","system-permission","system-quota"]
`

var (
	Namespace                 string
	DefaultServiceAccountName = "w7"
	scheme                    = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = k3kv1alpha.AddToScheme(scheme)
	// _ = higressscheme.AddToScheme(scheme)
	_ = higressnetworkingv1.AddToScheme(scheme)
	_ = higressextv1.AddToScheme(scheme)
	_ = apirbacv1.AddToScheme(scheme)
	_ = appgroupv1.AddToScheme(scheme)
	_ = microapp.AddToScheme(scheme)
}

func GetScheme() *runtime.Scheme {
	return scheme
}

// LoggingRoundTripper 是一个自定义的 http.RoundTripper，用于打印请求和响应的详细信息。
type LoggingRoundTripper struct {
	Proxied http.RoundTripper
}

// RoundTrip 实现 http.RoundTripper 接口。
func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// 打印请求信息
	fmt.Printf("Request URL: %s %s\n", req.Method, req.URL)
	fmt.Println("Request Headers:")
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		fmt.Println("Request Body:")
		fmt.Println(string(bodyBytes))
	}

	// 发送请求
	resp, err := lrt.Proxied.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// 打印响应信息
	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Println("Response Headers:")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return resp, nil
}

func init() {
	data, err := os.ReadFile(namespaceFilePath)
	if err != nil {
		ns := os.Getenv("POD_NAMESPACE")
		if ns != "" {
			Namespace = "default"
		}
		return
	}
	Namespace = string(data)
}

func GetTokenSaName(token string) (string, *jwtv5.NumericDate) {
	data := jwtv5.MapClaims{}
	jwtToken, _, err := jwtv5.NewParser().ParseUnverified(token, data)
	if err != nil {
		return "", nil
	}

	expireData, _ := jwtToken.Claims.GetExpirationTime()

	saName, ok := data["kubernetes.io/serviceaccount/service-account.name"]
	if ok {
		//获取过期时间

		return saName.(string), expireData
	}

	kubernetesIO, ok := data["kubernetes.io"].(map[string]interface{})
	if !ok {
		return "", expireData
	}

	serviceaccount, ok := kubernetesIO["serviceaccount"].(map[string]interface{})
	if !ok {
		return "", expireData
	}

	serviceaccountName, ok := serviceaccount["name"].(string)
	if !ok {
		return "", expireData
	}
	return serviceaccountName, expireData
}

type ApplyOptions struct {
	// 资源名称
	Namespace  string `json:"namespace"`
	ServerSide bool   `json:"serverSide"`
}

func NewApplyOptions(ns string) *ApplyOptions {
	return &ApplyOptions{
		Namespace: ns,
	}
}

func NewApplyOptionsServerSide(ns string, serverSide bool) *ApplyOptions {
	return &ApplyOptions{
		Namespace:  ns,
		ServerSide: serverSide,
	}
}

type Sdk struct {
	clientConfig       clientcmd.ClientConfig
	restConfig         *rest.Config
	ClientSet          *kubernetes.Clientset
	Ctx                context.Context
	namespace          string
	serviceAccountName string
	dynamicClient      *dynamic.DynamicClient
	restMapper         meta.RESTMapper
}

type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
	Context() context.Context
}

func newForClientConfig(clientConfig clientcmd.ClientConfig, namespace string) (*Sdk, error) {
	config, err := clientConfig.ClientConfig()
	if err != nil {
		slog.Error("error", "err", err)
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return NewForRestConfig(config, namespace)

	}
	return NewForRestConfig(config, namespace)
}

func NewForRestConfig(config *rest.Config, namespace string) (*Sdk, error) {
	startTime := time.Now()
	debug, ok := os.LookupEnv("SDK_DEBUG")
	if ok && debug == "true" {
		config.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			return &LoggingRoundTripper{Proxied: rt}
		}
	}

	newClientSetStart := time.Now()
	clientSet, err := kubernetes.NewForConfig(config)
	slog.Info("[PERF] NewForRestConfig - NewForConfig(clientSet) took %v", "duration", time.Since(newClientSetStart))
	if err != nil {
		return nil, err
	}

	newDynamicStart := time.Now()
	dynamicClient, err := dynamic.NewForConfig(config)
	slog.Info("[PERF] NewForRestConfig - NewForConfig(dynamic) took %v", "duration", time.Since(newDynamicStart))
	if err != nil {
		return nil, err
	}

	restmapCacheStart := time.Now()
	restmapCache := memory.NewMemCacheClient(clientSet.Discovery())
	restmap := restmapper.NewDeferredDiscoveryRESTMapper(restmapCache)
	slog.Info("[PERF] NewForRestConfig - REST mapper setup took %v", "duration", time.Since(restmapCacheStart))

	ctx := context.Background()
	sdk := &Sdk{
		restConfig:    config,
		ClientSet:     clientSet,
		Ctx:           ctx,
		namespace:     namespace,
		dynamicClient: dynamicClient,
		restMapper:    restmap,
	}
	slog.Info("[PERF] NewForRestConfig total time %v", "duration", time.Since(startTime))
	return sdk, nil
}

func NewK8sClientInner() *Sdk {

	kubePath, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubePath = "/home/workspace/.kube/config"
		home, err := os.UserHomeDir()
		if err == nil {
			kubePath = home + "/.kube/config"
		}

	}
	// kubePath := facade.GetConfig().GetString("k8s.config_path")
	// kubePath := "/home/workspace/.kube/config"
	// kubePath := "/tmp/abc"
	if strings.TrimSpace(kubePath) == "" {
		kubePath = ""
	}
	if kubePath != "" {
		if _, err := os.Stat(kubePath); os.IsNotExist(err) {
			// slog.Warn("k8s config path not exits", "path", kubePath, "err", err)
			kubePath = ""
		}
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		// &clientcmd.ClientConfigLoadingRules{ExplicitPath: facade.GetConfig().GetString("k8s.config_path")},
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubePath, WarnIfAllMissing: true},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}})

	kubeConfigNamespace, _, err := clientConfig.Namespace()
	if (err != nil) || (kubeConfigNamespace == "") {
		kubeConfigNamespace = Namespace
	}
	sdk, err := newForClientConfig(clientConfig, kubeConfigNamespace)
	if err != nil {
		fmt.Println("kubeconfig" + kubePath)
		// slog.Warn("new k8s client error", "err", err)
		panic(err)
	}
	sdk.clientConfig = clientConfig

	return sdk

}

func (self *Sdk) PodExecClient() *Sdk {

	self.restConfig = dynamic.ConfigFor(self.restConfig)
	self.restConfig.GroupVersion = &schema.GroupVersion{
		Group:   "",
		Version: "v1",
	}
	self.restConfig.APIPath = "/api"
	// self.DynamicClient = dynamic.ConfigFor(self.restConfig)
	return self
}

type PoxyOption struct {
	Auth string
	Path string
}

func (self *Sdk) GetNamespace() string {
	return self.namespace
}

func (self *Sdk) GetRestMapping(apiVersion string, kind string) (*meta.RESTMapping, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}
	// 解析资源的 API 版本和资源类型
	schema := schema.GroupKind{Group: gv.Group, Kind: kind}
	mapping, err := self.restMapper.RESTMapping(schema)
	if err != nil {
		// panic(err)
		return nil, err
	}

	return mapping, nil
}

// helm 需要
func (self *Sdk) ToRESTConfig() (*rest.Config, error) {
	return self.restConfig, nil
}

// helm 需要
func (self *Sdk) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return memory.NewMemCacheClient(self.ClientSet.Discovery()), nil
}

// helm 需要
func (self *Sdk) ToRESTMapper() (meta.RESTMapper, error) {
	if self.restMapper == nil {
		// restmapCacheStart := time.Now()
		restmapCache := memory.NewMemCacheClient(self.ClientSet.Discovery())
		self.restMapper = restmapper.NewDeferredDiscoveryRESTMapper(restmapCache)
	}
	return self.restMapper, nil
}

// helm 需要
func (self *Sdk) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	if self.clientConfig == nil {
		// 从restConfig创建新的clientConfig
		config := clientcmdapi.NewConfig()

		// 创建集群配置
		cluster := clientcmdapi.NewCluster()
		cluster.Server = self.restConfig.Host
		cluster.CertificateAuthorityData = self.restConfig.CAData
		cluster.InsecureSkipTLSVerify = self.restConfig.Insecure

		// 创建用户认证信息
		authInfo := clientcmdapi.NewAuthInfo()
		authInfo.Token = self.restConfig.BearerToken
		authInfo.ClientCertificateData = self.restConfig.CertData
		authInfo.ClientKeyData = self.restConfig.KeyData
		authInfo.Username = self.restConfig.Username
		authInfo.Password = self.restConfig.Password

		// 设置上下文
		context := clientcmdapi.NewContext()
		context.Cluster = "default"
		context.AuthInfo = "default"
		context.Namespace = self.namespace

		// 添加到配置中
		config.Clusters["default"] = cluster
		config.AuthInfos["default"] = authInfo
		config.Contexts["default"] = context
		config.CurrentContext = "default"

		// 创建并返回ClientConfig
		return clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	}
	return self.clientConfig
}

func (self *Sdk) ToMetricsClient() (*metricsclient.Clientset, error) {
	return metricsclient.NewForConfig(self.restConfig)
}

func (self *Sdk) ToSigClient() (sigclient.Client, error) {
	startTime := time.Now()

	toRestStart := time.Now()
	config, err := self.ToRESTConfig()
	slog.Info("[PERF] Sdk.ToSigClient - ToRESTConfig took %v", "duration", time.Since(toRestStart))
	if err != nil {
		return nil, err
	}

	newClientStart := time.Now()
	result, err := sigclient.New(config, sigclient.Options{Scheme: scheme})
	slog.Info("[PERF] Sdk.ToSigClient - sigclient.New took %v", "duration", time.Since(newClientStart))
	slog.Info("[PERF] Sdk.ToSigClient total time %v", "duration", time.Since(startTime))
	return result, err
}

func (self *Sdk) OpenAPIV3Client() (openapiclient.Client, error) {
	discovery, err := self.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	return cached.NewClient(discovery.OpenAPIV3()), nil
}

func (self *Sdk) OpenAPISchema() (*openapi_v2.Document, error) {
	return self.ClientSet.Discovery().OpenAPISchema()

}

func (self *Sdk) GetServiceAccountName() string {

	if self.restConfig.BearerToken != "" {
		//json webtoken方式 解析podToken中的serviceAccount
		// 获取serviceAccountName
		name, _ := GetTokenSaName(self.restConfig.BearerToken)
		return name
	}
	return "unknown"
}

func (self *Sdk) Channel(token string) (*Sdk, error) {
	config := rest.CopyConfig(self.restConfig)
	config.BearerToken = token
	config.BearerTokenFile = ""
	config.TLSClientConfig.CertData = nil
	config.TLSClientConfig.KeyData = nil

	sdk, err := NewForRestConfig(config, self.namespace)
	if err != nil {
		return nil, err
	}
	sdk.clientConfig = self.clientConfig
	return sdk, nil
}

func (self *Sdk) Proxy(request *http.Request, response gin.ResponseWriter) (err error) {
	result, err := url.Parse(self.restConfig.Host)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error building kubeAPIURL: %v", "err", err)
		return
	}

	// 3. 设置HTTP代理
	proxy := httputil.NewSingleHostReverseProxy(result)
	tr, err := rest.TransportFor(self.restConfig)
	if err != nil {
		slog.Error("Error building transport: %v", "err", err)
		return err
	}
	proxy.Transport = tr
	proxy.ServeHTTP(response, request)

	return
}

func (self *Sdk) RunExec(ptyHandler PtyHandler, namespace string, podName string, containerName string, cmd []string, tty bool) (err error) {
	ttystr := "false"
	if tty {
		ttystr = "true"
	}
	request := self.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName).
		Param("stdin", "true").
		Param("stdout", "true").
		// Param("stderr", "true"). //k3k 无法进入pod里
		Param("tty", ttystr)
	for _, c := range cmd {
		request = request.Param("command", c)
	}
	if !tty {
		request = request.Param("stderr", "true")
	}
	exec, err := remotecommand.NewSPDYExecutor(self.restConfig, "POST", request.URL())
	if err != nil {
		slog.Error("Error while creating executor: %v", "err", err)
		return err
	}
	err = exec.StreamWithContext(ptyHandler.Context(),
		remotecommand.StreamOptions{
			Stdin:             ptyHandler,
			Stdout:            ptyHandler,
			Stderr:            ptyHandler,
			Tty:               tty,
			TerminalSizeQueue: ptyHandler,
		})

	slog.Info("k8s exec done", "err", err)
	return err
}

func (self *Sdk) CreateTokenRequest(serviceAccount string, expireSeconds int64, audiences []string) (string, error) {
	startTime := time.Now()
	tokenReq := v1.TokenRequest{
		Spec: v1.TokenRequestSpec{
			ExpirationSeconds: &expireSeconds,
			Audiences:         audiences,
			// BoundObjectRef: &v1.BoundObjectReference{
			// 	Kind:       "Secret",
			// 	Name:       "sealos-token-mtmhshtr-c9bc",
			// 	APIVersion: "v1",
			// },
		},
	}

	createTokenStart := time.Now()
	result, err := self.ClientSet.CoreV1().ServiceAccounts(self.namespace).CreateToken(self.Ctx, serviceAccount, &tokenReq, metav1.CreateOptions{})
	createTokenTime := time.Since(createTokenStart)

	slog.Info("[PERF] Sdk.CreateTokenRequest - CreateToken took %v", "duration", createTokenTime)
	slog.Info("[PERF] Sdk.CreateTokenRequest total time %v", "duration", time.Since(startTime))

	if err != nil {
		return "", err
	}
	return result.Status.Token, nil
}

func (self *Sdk) TokenReview(requestToken string) error {
	// sdk, err := self.Channel(requestToken)
	// if err != nil {
	// 	return err
	// }
	sdk := self
	clientset := sdk.ClientSet

	review := &v1.TokenReview{
		Spec: v1.TokenReviewSpec{
			Token: requestToken,
		},
	}
	tokenreview, err := clientset.AuthenticationV1().TokenReviews().Create(context.TODO(), review, metav1.CreateOptions{})
	if tokenreview != nil {
		if (!tokenreview.Status.Authenticated) || (tokenreview.Status.User.Username == "") {
			// http.Error(wr, "token is not authenticated", http.StatusBadRequest)
			// 创建error 对象
			return fmt.Errorf("token is not authenticated")
		}
	}

	if err != nil {
		_, err = self.ClientSet.Discovery().ServerVersion()
		if err != nil {
			return err
		}
		return err
	}

	return nil
}

func (self *Sdk) ApplyJson(json []byte, options ApplyOptions) (*unstructured.Unstructured, error) {
	uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, json)
	if err != nil {
		return nil, err
	}
	// resource = uncastObj.(*unstructured.Unstructured)
	// 解析资源列表
	var res unstructured.Unstructured
	res = *uncastObj.(*unstructured.Unstructured)

	rm, err := self.GetRestMapping(res.GetAPIVersion(), res.GetKind())
	if err != nil {
		return nil, err
	}
	name := res.GetName()
	ns := res.GetNamespace()
	gvr := rm.Resource
	if ns == "" {
		ns = options.Namespace
	}
	var resourceInterface dynamic.ResourceInterface
	resourceInterface = self.dynamicClient.Resource(gvr)
	if rm.Scope.Name() == meta.RESTScopeNameNamespace {
		resourceInterface = self.dynamicClient.Resource(gvr).Namespace(ns)
	}
	// 检查资源是否存在
	r, err := resourceInterface.Get(self.Ctx, name, metav1.GetOptions{})
	if err == nil {
		// 如果资源存在，则进行更新
		// r2, err := resourceInterface.Apply(context.TODO(), name, &res, metav1.ApplyOptions{FieldManager: "k8s-offline"})
		// r2, err := resourceInterface.Patch(context.TODO(), name, types.StrategicMergePatchType, &res, metav1.PatchOptions{FieldManager: "k8s-offline"})
		r2, err := resourceInterface.Update(self.Ctx, &res, metav1.UpdateOptions{FieldManager: "k8s-offline"})
		if err != nil {
			return nil, err
		}
		return r2, err
	}

	if err != nil && errors.IsNotFound(err) {
		// 如果资源不存在，则进行创建
		r3, err := resourceInterface.Create(self.Ctx, &res, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		return r3, err
	}
	return r, err
}

/**
 * k8s yaml 增加删除修改
 */
func (self *Sdk) ApplyYaml(yamlbytes []byte, options ApplyOptions) (*unstructured.Unstructured, error) {
	json, err := yaml.YAMLToJSON(yamlbytes)
	if err != nil {
		return nil, err
	}
	return self.ApplyJson(json, options)
}
func (self *Sdk) ApplyFile(fileName string, options ApplyOptions) error {

	args := []string{"apply", "-f", fileName}
	if options.ServerSide {
		args = append(args, "--server-side")
	}
	return self.applyRaw(args, options)
}

func (self *Sdk) ApplyDir(dirName string, options ApplyOptions) error {

	args := []string{"apply", "-k", dirName}
	if options.ServerSide {
		args = append(args, "--server-side")
	}
	return self.applyRaw(args, options)
}

func (self *Sdk) applyRaw(args2 []string, options ApplyOptions) error {
	_, errstr, err := helper.Runsh("kubectl", args2...)
	if err != nil {
		slog.Error("applyRaw err", "errstr", errstr)
		return err
	}
	return nil
	// factory := cmdutil.NewFactory(self)
	// stream := genericiooptions.NewTestIOStreamsDiscard()
	// flags := cmdapply.NewApplyFlags(stream)
	// cmdapply.NewCmdApply("apply", factory, stream)
	// cmd := &cobra.Command{
	// 	Use:                   "apply (-f FILENAME | -k DIRECTORY)",
	// 	DisableFlagsInUseLine: true,
	// 	Short:                 i18n.T("Apply a configuration to a resource by file name or stdin"),
	// 	RunE: func(cmd *cobra.Command, args []string) error {
	// 		o, err := flags.ToOptions(factory, cmd, "kubectl", args)
	// 		if err != nil {
	// 			slog.Error("err", "err", err)
	// 			return err
	// 		}
	// 		err = o.Validate()
	// 		if err != nil {
	// 			slog.Error("err", "err", err)
	// 			return err
	// 		}
	// 		err = o.Run()
	// 		if err != nil {
	// 			slog.Error("err", "err", err)
	// 			return err
	// 		}
	// 		return nil
	// 	},
	// }
	// flags.AddFlags(cmd)
	// cmd.SetArgs(args2)
	// return cmd.Execute()
}
func (self *Sdk) ApplyBytes(data []byte, options ApplyOptions) error {

	// sigClient, err := self.ToSigClient()
	// if err != nil {
	// 	return err
	// }
	helmclient := kube.New(self)
	origin, err := helmclient.Build(bytes.NewReader(data), true)
	if err != nil {
		return err
	}
	var current kube.ResourceList
	origin.Visit(func(info *resource.Info, err error) error {
		helper := resource.NewHelper(info.Client, info.Mapping)
		obj, err := helper.Get(info.Namespace, info.Name)
		if err == nil {
			//移除obj中  resourceVersion uid generation creationTimestamp 字段
			info.Refresh(obj, true)
		}
		current.Append(info)
		return nil
	})

	target, err := helmclient.Build(bytes.NewReader(data), false)
	if err != nil {
		return err
	}

	target, err = existingResourceConflict(target)
	if err != nil {
		return err
	}

	target.Visit(func(info *resource.Info, err error) error {
		newInfo := current.Get(info)
		if newInfo != nil {
			metadata, err := meta.Accessor(info.Object)
			if err != nil {
				return nil
			}
			metadata.SetResourceVersion(newInfo.Object.(*unstructured.Unstructured).GetResourceVersion())
			metadata.SetUID(newInfo.Object.(*unstructured.Unstructured).GetUID())
			metadata.SetGeneration(newInfo.Object.(*unstructured.Unstructured).GetGeneration())
			metadata.SetCreationTimestamp(newInfo.Object.(*unstructured.Unstructured).GetCreationTimestamp())
		}
		return nil
	})

	if len(target) == 0 && len(current) > 0 {
		_, err = helmclient.Create(current)
	} else if len(current) > 0 {
		_, err = helmclient.Update(current, target, false)
	}
	if err != nil {
		return err
	}
	return nil

	// tmpfile, err := os.CreateTemp(os.TempDir(), "apply")
	// if err != nil {
	// 	return err
	// }
	// tmpfile.Write(data)
	// defer tmpfile.Close()

	// args := []string{"apply", "-f", tmpfile.Name()}
	// if options.ServerSide {
	// 	args = append(args, "--server-side")
	// }
	// return self.applyRaw(args, options)

}

func (self Sdk) Login2(username string, password string, checkPassword bool) (*corev1.ServiceAccount, error) {
	sa, err := self.ClientSet.CoreV1().ServiceAccounts(self.namespace).Get(self.Ctx, username, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	annotations := sa.GetAnnotations()
	labels := sa.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
		sa.Labels = labels
	}
	if annotations == nil {
		annotations = make(map[string]string)
		sa.Annotations = annotations
	}
	expiretime, ok1 := annotations["w7.cc/expiretime"]
	if ok1 && expiretime != "" {
		//
		timeobj, err := time.ParseInLocation("2006-01-02 15:04:05", expiretime, time.Local)
		if err != nil {
			return nil, err
		}
		if timeobj.Unix() < time.Now().Unix() {
			mode, ok := labels["w7.cc/user-mode"]
			if ok && mode != "cluster" && mode != "founder" {
				return nil, fmt.Errorf("用户已过期") //未配置费用购买权限，无法登录
			}
			// 过期后
			// return nil, fmt.Errorf("用户已过期")
		}
	}
	if !checkPassword {
		return sa, nil
	}
	passwd, ok := annotations["password"]
	if !ok {
		return nil, fmt.Errorf("用户名密码错误")
	}
	if ok {
		err = bcrypt.CompareHashAndPassword([]byte(passwd), []byte(password))
		if err != nil {
			return nil, fmt.Errorf("用户名密码错误")
		}
		// 2025-06-29 14:19:37
	}
	if sa.Labels == nil {
		sa.Labels = make(map[string]string)
	}
	_, hasMode := sa.Labels["w7.cc/user-mode"]
	if !hasMode {
		sa.Labels["w7.cc/user-mode"] = "founder"
		sa.Annotations["w7.cc/menu-name"] = "k3k.permission.founder"
		sa.Annotations["w7.cc/menu"] = K3K_MENU_FOUNDER
		sa.Annotations["w7.cc/debug"] = "true"
		sa.Annotations["w7.cc/file-editor"] = "true"
		sa.Annotations["w7.cc/web-shell"] = "true"
		_, err := self.ClientSet.CoreV1().ServiceAccounts(self.namespace).Update(self.Ctx, sa, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return sa, nil
}
func (self Sdk) Login(username string, password string, createToken bool, seconds int64) (string, error) {

	sa, err := self.ClientSet.CoreV1().ServiceAccounts(self.namespace).Get(self.Ctx, username, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	annotations := sa.GetAnnotations()
	passwd, ok := annotations["password"]
	if !ok {
		return "", fmt.Errorf("用户名密码错误")
	}
	if ok {
		err = bcrypt.CompareHashAndPassword([]byte(passwd), []byte(password))
		if err != nil {
			return "", err
		}
	}
	if !createToken {
		return "", nil
	}

	token, err := self.CreateTokenRequest(username, seconds, []string{})
	if err != nil {
		slog.Error("create token error", "err", err)
		// if k8s.DefaultK8sToken != "" {
		// 	return k8s.DefaultK8sToken, nil
		// }
		return "", err
	}
	self.CreateServiceAccountSecret(username) // 首次token 不存在bug
	return token, nil

}

func (self Sdk) ResetPassword(username string, password string, usermode string) error {
	bpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: username,
			// Annotations: map[string]string{"password": string(bpassword)},
			// Labels:      map[string]string{"w7.cc/usermode": usermode},
			Namespace: self.GetNamespace(),
		},
	}
	client, err := self.ToSigClient()
	if err != nil {
		return err
	}
	_, err = controllerutil.CreateOrPatch(self.Ctx, client, sa, func() error {
		if sa.Annotations == nil {
			sa.Annotations = make(map[string]string)
		}
		sa.Annotations["password"] = string(bpassword)
		if sa.Labels == nil {
			sa.Labels = make(map[string]string)
		}
		sa.Labels["w7.cc/user-mode"] = usermode
		if usermode == "founder" {
			sa.Annotations["w7.cc/menu-name"] = "k3k.permission.founder"
			sa.Annotations["w7.cc/menu"] = K3K_MENU_FOUNDER
			sa.Annotations["w7.cc/debug"] = "true"
			sa.Annotations["w7.cc/file-editor"] = "true"
			sa.Annotations["w7.cc/web-shell"] = "true"
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
	// applysa, err := applyv1.ExtractServiceAccount(sa, "Mozilla")
	// if err != nil {
	// 	return err
	// }
	// applysa.WithNamespace(self.GetNamespace()).WithLabels(map[string]string{"w7.cc/usermode": usermode}).
	// 	WithName(username).WithAnnotations(map[string]string{"password": string(bpassword)}).WithAutomountServiceAccountToken(true)

	// _, err = self.ClientSet.CoreV1().ServiceAccounts(self.namespace).Apply(self.Ctx, applysa, metav1.ApplyOptions{FieldManager: "Mozilla"})
	// if err != nil {
	// 	return err
	// }
	// return nil
}

func (self Sdk) Register(username string, password string, namespace, role string, isClusterRole bool, usermode string) error {
	// configName := "w7" //facade.Config.GetString("k8s.config_secret_name")
	err := self.ResetPassword(username, password, usermode)
	if err != nil {
		slog.Error("reset password error", "err", err)
		return err
	}
	if isClusterRole {
		err := self.CreateClusterRoleBinding(username, "cluster-admin")
		if err != nil {
			slog.Warn("create cluster role binding error", "err", err, "sa", username, "roleName", "cluster-admin")
		}
	} else {
		err := self.CreateRoleBinding(username, role, namespace)
		if err != nil {
			slog.Warn("create role binding error", "err", err, "sa", username, "roleName", "cluster-admin")
		}
	}
	slog.Info("username", "uname", username)
	//首次token 不存在bug
	self.CreateServiceAccountSecret(username)

	return nil

}

func (self Sdk) CreateClusterRoleBinding(name string, roleName string) error {
	clusterRoleBinding := &apirbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []apirbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: self.GetNamespace(),
			},
		},
		RoleRef: apirbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	_, err := self.ClientSet.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (self Sdk) CreateRoleBinding(name string, roleName string, namespace string) error {
	roleBindings := &apirbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []apirbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: self.GetNamespace(),
			},
		},
		RoleRef: apirbacv1.RoleRef{
			Kind:     "Role",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	_, err := self.ClientSet.RbacV1().RoleBindings(namespace).Create(context.TODO(), roleBindings, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (self Sdk) GetNamespaces() (*corev1.NamespaceList, error) {
	namespaces, err := self.ClientSet.CoreV1().Namespaces().List(self.Ctx, metav1.ListOptions{})
	if err != nil {
		defaultNamespace := self.GetNamespace() // facade.GetConfig().GetString("k8s.default_namespace")
		if defaultNamespace == "" {
			return nil, err
		}
		namespaces = &corev1.NamespaceList{
			Items: []corev1.Namespace{
				corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: defaultNamespace,
					},
				},
			},
		}
	}
	return namespaces, nil
}

/**
* 回滚deployment statefulset daemonset
 */
func (self Sdk) RollBack(obj runtime.Object, restmapping *meta.RESTMapping, toRevision int64) (string, error) {
	// map, self.RestMapper.RESTMapping(deployment.GetResource().GroupVersionKind())
	fc, err := polymorphichelpers.RollbackerFn(&self, restmapping)
	if err != nil {
		return "", err
	}
	return fc.Rollback(obj, nil, toRevision, 0)
}

func (self Sdk) GetK8sRawObject(name string, apiVersion string, kind string, namespace string) (*unstructured.Unstructured, error) {
	mapping, err := self.GetRestMapping(apiVersion, kind)
	if err != nil {
		return nil, err
	}
	var resource dynamic.ResourceInterface
	resource = self.dynamicClient.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resource = resource.(dynamic.NamespaceableResourceInterface).Namespace(namespace)
	}
	return resource.Get(context.TODO(), name, metav1.GetOptions{})
}

func (self *Sdk) CreateServiceAccountSecret(serviceAccount string) (*corev1.Secret, error) {

	saSecret, err := self.ClientSet.CoreV1().Secrets(self.namespace).Get(self.Ctx, serviceAccount, metav1.GetOptions{})
	if err != nil {
		slog.Warn("get service account secret error1", "err", err, "sa", serviceAccount)
		saSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceAccount,
				Annotations: map[string]string{
					"kubernetes.io/service-account.name": serviceAccount,
				},
			},
			Type: corev1.SecretTypeServiceAccountToken,
		}
		saSecret, err = self.ClientSet.CoreV1().Secrets(self.namespace).Create(self.Ctx, saSecret, metav1.CreateOptions{})
		if err != nil {
			slog.Warn("create service account secret error", "err", err, "sa", serviceAccount)
			return nil, err
		}
	}
	return saSecret, nil
}

func (self Sdk) ToKubeconfig(apiServerUrl string) (*clientcmdv1.Config, error) {

	restConfig := self.restConfig
	kubeConfig := &clientcmdv1.Config{}
	name := uuid.New().String()

	// 设置集群信息
	// cluster := clientcmdapi.NewCluster()
	cluster := clientcmdv1.Cluster{}
	cluster.Server = restConfig.Host
	if apiServerUrl != "" {
		cluster.Server = apiServerUrl
	}
	cluster.CertificateAuthorityData = restConfig.CAData
	cluster.InsecureSkipTLSVerify = restConfig.Insecure

	if len(cluster.CertificateAuthorityData) == 0 {
		caFile := restConfig.CAFile
		data, err := os.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		cluster.CertificateAuthorityData = data
	}
	saName := self.GetServiceAccountName()
	// saName := "w7panel"
	if helper.IsLocalMock() {
		saName = helper.ServiceAccountName()
	}
	if saName == "" {
		slog.Warn("saName is empty")
		saName = facade.GetConfig().GetString("app.helm_release_name")
	}
	slog.Info("to kubectl sa name", "saName", saName)
	secret, err := self.CreateServiceAccountSecret(saName)
	if err != nil {
		return nil, err
	}
	token := secret.Data["token"] //secret是admin
	if len(token) == 0 {
		secret, err = self.CreateServiceAccountSecret(saName)
		token := secret.Data["token"]
		if len(token) == 0 {
			slog.Warn("token is empty", "secretName", secret.Name)
			return nil, fmt.Errorf("token is empty 请重试")
		}
	}
	namedcluster := clientcmdv1.NamedCluster{Name: name, Cluster: cluster}
	// 设置用户信息
	authInfo := clientcmdv1.AuthInfo{}
	authInfo.ClientCertificateData = restConfig.CertData
	authInfo.ClientKeyData = restConfig.KeyData
	authInfo.Token = string(token) //restConfig.BearerToken
	authInfo.Username = restConfig.Username
	authInfo.Password = restConfig.Password

	namedauth := clientcmdv1.NamedAuthInfo{Name: name, AuthInfo: authInfo}

	// 设置上下文信息
	context := clientcmdv1.Context{}
	// context := &clinetcmdv1.NamedContext{}
	context.Cluster = name
	context.AuthInfo = name

	namedcontext := clientcmdv1.NamedContext{Name: name, Context: context}

	// 将集群、用户和上下文添加到 kubeconfig 中
	kubeConfig.Clusters = append(kubeConfig.Clusters, namedcluster)
	kubeConfig.AuthInfos = append(kubeConfig.AuthInfos, namedauth)
	kubeConfig.Contexts = append(kubeConfig.Contexts, namedcontext)
	kubeConfig.CurrentContext = name

	return kubeConfig, nil
}

// k8a  /api 请求获取serverAddressByClientCIDRs
func (sdk Sdk) GetApiServerUrl() (string, error) {

	var jsonmap map[string]interface{}
	bytes, err := sdk.ClientSet.RESTClient().Get().AbsPath("/api").DoRaw(context.TODO())
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bytes, &jsonmap)
	if err != nil {
		return "", err
	}
	return "https://" + jsonmap["serverAddressByClientCIDRs"].([]interface{})[0].(map[string]interface{})["serverAddress"].(string), nil
}

// 获取daemonset agent pod by hostIp
func (self Sdk) GetDaemonsetAgentPod(namespace, hostIp string) (*corev1.Pod, error) {
	daemonsetPods, err := self.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "w7.cc/daemonset=w7"})
	if err != nil {
		slog.Warn("get daemonset pods error", "err", err)
		return nil, err
	}
	for _, pod := range daemonsetPods.Items {
		if pod.Status.HostIP == hostIp {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("not found pod")
}

func (self Sdk) GetDaemonsetAgentPods(namespace string) (*corev1.PodList, error) {
	daemonsetPods, err := self.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "w7.cc/daemonset=w7"})
	if err != nil {
		slog.Warn("get daemonset pods error", "err", err)
		return nil, err
	}
	return daemonsetPods, nil
}

func (self Sdk) GetContainerPid(findPod *corev1.Pod, containerId string) (int, error) {

	session := terminal.NewTerminalSession(nil)
	defer session.Close()
	containerName := findPod.Spec.Containers[0].Name
	//containerd://d8805359e736a93e4022d941cc7e3989058680459dae93eb26af90b21f3fa874

	if strings.HasPrefix(containerId, "containerd://") {
		containerId = containerId[len("containerd://"):]
	}
	cmd := []string{"nsenter", "-t", "1", "--mount", "--pid", "--", "crictl", "inspect", "--output", "go-template", fmt.Sprintf("--template='{{.info.pid}}'"), containerId}
	//crictl inspect --output go-template --template '{{.info.pid}}' 74506f333e1c00de74551c68566c73e71aff9054ebeb21037d33667585a9a943

	err := self.RunExec(session, findPod.Namespace, findPod.Name, containerName, cmd, false)
	if err != nil {
		return 0, err
	}
	pid := string(session.GetWriterBytes())
	pid = strings.Replace(pid, "\n", "", -1)
	pid = strings.Replace(pid, "'", "", -1)
	//pid string to int
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return 0, err
	}
	return pidInt, nil
}

func (self Sdk) GetDeploymentAppByIdentifie(namespace, identifie string) (*appsv1.DeploymentList, error) {
	deploymentApps, err := self.ClientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "w7.cc/identifie=" + identifie})
	return deploymentApps, err
}

func (self Sdk) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	deployment, err := self.ClientSet.AppsV1().Deployments(namespace).Get(self.Ctx, name, metav1.GetOptions{})
	return deployment, err
}

func (self Sdk) UpdateDeployment(namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	deployment, err := self.ClientSet.AppsV1().Deployments(namespace).Update(self.Ctx, deployment, metav1.UpdateOptions{})
	return deployment, err
}

func (self Sdk) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	rs, err := self.ClientSet.AppsV1().StatefulSets(namespace).Get(self.Ctx, name, metav1.GetOptions{})
	return rs, err
}

func (self Sdk) UpdateStatefulSet(namespace string, rs *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	rs2, err := self.ClientSet.AppsV1().StatefulSets(namespace).Update(self.Ctx, rs, metav1.UpdateOptions{})
	return rs2, err
}

func (self Sdk) GetDaemonset(namespace, name string) (*appsv1.DaemonSet, error) {
	rs, err := self.ClientSet.AppsV1().DaemonSets(namespace).Get(self.Ctx, name, metav1.GetOptions{})
	return rs, err
}

func (self Sdk) UpdateDaemonset(namespace string, rs *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	rs2, err := self.ClientSet.AppsV1().DaemonSets(namespace).Update(self.Ctx, rs, metav1.UpdateOptions{})
	return rs2, err
}

func (self Sdk) GetServiceAccount(namespace, name string) (*corev1.ServiceAccount, error) {
	if namespace == "" {
		namespace = "default"
	}
	sa, err := self.ClientSet.CoreV1().ServiceAccounts(namespace).Get(self.Ctx, name, metav1.GetOptions{})
	return sa, err
}

func (self Sdk) PatchServiceAccount(namespace, name string, data []byte) (*corev1.ServiceAccount, error) {
	if namespace == "" {
		namespace = "default"
	}
	sa, err := self.ClientSet.CoreV1().ServiceAccounts(namespace).Patch(self.Ctx, name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
	return sa, err
}

func (self Sdk) GetServiceAccountByConsoleId(namespace, consoleId string) (*corev1.ServiceAccountList, error) {
	saList, err := self.ClientSet.CoreV1().ServiceAccounts(namespace).List(self.Ctx, metav1.ListOptions{LabelSelector: "w7.cc/identifie=" + consoleId})
	return saList, err
}

func (self Sdk) CreateNamespace(namespace string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	ns, err := self.ClientSet.CoreV1().Namespaces().Create(self.Ctx, ns, metav1.CreateOptions{})
	return ns, err
}

func (self Sdk) GetLicense() (*corev1.Secret, error) {
	ns, err := self.ClientSet.CoreV1().Secrets("kube-system").Get(self.Ctx, "license", metav1.GetOptions{})
	return ns, err
}

func existingResourceConflict(resources kube.ResourceList) (kube.ResourceList, error) {
	var requireUpdate kube.ResourceList

	err := resources.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		helper := resource.NewHelper(info.Client, info.Mapping)
		_, err = helper.Get(info.Namespace, info.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("failed to get existing %s: %w", info.Mapping.Resource, err)
		}

		requireUpdate.Append(info)
		return nil
	})

	return requireUpdate, err
}

func (self Sdk) GetClusterId() (string, error) {
	// 从kube-system命名空间中获取名为"server1.node-password.k3s"的Secret资源
	secret, err := self.ClientSet.CoreV1().Secrets("kube-system").Get(self.Ctx, "server1.node-password.k3s", metav1.GetOptions{})
	if err != nil {
		slog.Error("获取集群ID失败", "error", err)
		return "", nil
	}
	return helper.StringToMD5(string(secret.Data["hash"])), nil
}
