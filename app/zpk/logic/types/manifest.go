package types

import (
	"encoding/json"
	"strings"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	zpktype "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var Z = []string{"%STORAGE_RW_MODE%", "%STORAGE_SIZE%", "%STORAGE_CLASS_NAME%", "%DOMAIN_URL%", "%DOMAIN_SSL_URL%", "%DOMAIN_HOST%"}

type Manifest struct {
	Application Application        `json:"application"`
	Platform    Platform           `json:"platform"`
	V           intstr.IntOrString `json:"v"`
	Bindings    []Bindings         `json:"bindings"`
	WebApp      WebApp             `json:"webapp"`
}

type Menu struct {
	Displayorder intstr.IntOrString `json:"displayorder"`
	Do           string             `json:"do"`
	Icon         string             `json:"icon"`
	IsDefault    int                `json:"is_default"`
	Location     string             `json:"location"`
	Title        string             `json:"title"`
	Parent       string             `json:"parent"`
}

type Bindings struct {
	Framework         string `json:"framework"`
	IsDefaultRegister int    `json:"is_default_register"`
	Location          string `json:"location"`
	Support           string `json:"support"`
	Menu              []Menu `json:"menu"`
	Name              string `json:"name"`
	Status            int    `json:"status"`
	Title             string `json:"title"`
}

func (m *Manifest) ShowMenuTop() bool {
	return len(m.Bindings) >= 2
}

func (m *Manifest) MenuLabels() map[string]string {
	result := map[string]string{}
	for _, v := range m.Application.FrontType {
		if v == "thirdparty_cd" && m.ShowMenuTop() {
			result["w7.cc/menu-location"] = "top"
			for _, v1 := range m.Bindings {
				result["role.w7.cc/"+v1.Name] = "true"
			}
		}
	}
	return result
}
func (m *Manifest) requrirePvc() bool {
	volumes := m.Platform.Container.Volumes
	for _, volume := range volumes {
		if volume.Type == "diskStorage" {
			return true
		}
	}
	return false
}

func (m *Manifest) RequireSite() bool {
	types := m.Application.FrontType
	for _, s := range types {
		if s == "console" {
			return true
		}
	}
	return false
}

func (m *Manifest) GetRoutesByName(name string) []Routes {
	for _, ing := range m.Platform.Ingress {
		if ing.Name == name {
			return ing.Routes
		}
	}
	return []Routes{}
}

func (m *Manifest) GetBindsJson() []byte {
	binds := m.Bindings
	data, err := json.Marshal(binds)
	return lo.Ternary(err != nil, []byte{}, data)
}

// / 是否需要域名
func (m *Manifest) RequireDomain() bool {
	for _, item := range m.Application.FrontType {
		if item == "console" {
			return true
		}
	}
	for _, param := range m.Platform.Container.StartParams {
		if param.ValuesText == "%DOMAIN_URL%" || param.ValuesText == "%DOMAIN_SSL_URL%" || param.ValuesText == "%DOMAIN_HOST%" {
			return true
		}
	}
	return false
}

func (m *Manifest) RequireDomainForce() bool {
	for _, item := range m.Application.FrontType {
		if item == "console" {
			return true
		}
	}
	for _, param := range m.Platform.Container.StartParams {
		if (param.ValuesText == "%DOMAIN_URL%" || param.ValuesText == "%DOMAIN_SSL_URL%") && param.Required {
			return true
		}
	}
	return false
}

func (m *Manifest) GetShellByType(shellType string) zpktype.ManifestShellInterface {
	for _, param := range m.Platform.Container.Shells {
		if param.Type == shellType {
			return &param
		}
	}
	return nil
}

// 是否需要https
func (m *Manifest) RequireDomainHttps() bool {
	for _, item := range m.Application.FrontType {
		if item == "console" {
			return true
		}
	}
	for _, param := range m.Platform.Container.StartParams {
		if param.ValuesText == "%DOMAIN_SSL_URL%" {
			return true
		}
	}
	return false
}

func (m *Manifest) RequireBuild() bool {
	return m.Platform.Container.Image == "" && m.Application.Type != "helm"
}

func (m *Manifest) GetBuildContext() string {
	context := m.Platform.Container.Build.BuildContext
	if context == "." {
		context = ""
	}
	return "/workspace/" + context
}

func (m *Manifest) GetDockerfile() string {
	return "Dockerfile"
}

func (m *Manifest) SupportMicroApp() bool {
	return lo.Contains(m.Application.FrontType, "thirdparty_cd")
}

// GetDockerfilePath 返回 Dockerfile 的完整路径，路径由构建上下文路径和 "/Dockerfile" 拼接而成。
// 返回值是字符串类型，表示 Dockerfile 的绝对路径。
func (m *Manifest) GetDockerfilePath() string {
	return m.GetBuildContext() + "/Dockerfile"
}

// volumes 起个名字，防止重复
func (m *Manifest) GenVolumesName(preVolumesName map[string]string) {
	for key, volume := range m.Platform.Container.Volumes {
		if preVolumesName[volume.MountPath] != "" {
			volume.Name = preVolumesName[volume.MountPath]
			m.Platform.Container.Volumes[key].Name = volume.Name
		}
		if volume.Name == "" {
			volume.Name = strings.ToLower(volume.Type) + "-" + helper.RandomString(5)
			m.Platform.Container.Volumes[key].Name = volume.Name
		}
	}
}

func (m *Manifest) AppendDomainStartParams() {
	hasDomainConfig := false
	for _, val := range m.Platform.Container.StartParams {
		if val.ValuesText == "%DOMAIN_URL%" || val.ValuesText == "%DOMAIN_SSL_URL%" {
			hasDomainConfig = true
		}
	}
	if !hasDomainConfig {
		m.Platform.Container.StartParams = append(m.Platform.Container.StartParams, StartParams{
			Name:        "DOMAIN_URL",
			Title:       "域名地址",
			Description: "域名地址，例如：example.com",
			ValuesText:  "%DOMAIN_SSL_URL%",
			Required:    true,
		})
	}
}

func (m *Manifest) HasFront() bool {
	if len(m.Application.FrontType) > 0 {
		for _, v := range m.Application.FrontType {
			if v == "thirdparty_cd" {
				return true
			}
		}
	}
	return false
}

type Root struct {
	Manifest Manifest `json:"manifest"`
}

type Application struct {
	Author      string            `json:"author"`
	Description string            `json:"description"`
	FrontType   []string          `json:"front_type"`
	Identifie   string            `json:"identifie"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Icon        string            `json:"logo"`
	Once        bool              `json:"once"`
	Annotation  map[string]string `json:"annotation"`
}
type Build struct {
	BuildContext string `json:"build_context"`
}
type WebApp struct {
	Url string `json:"url"`
}
type Env struct {
	Name      string               `json:"name"`
	Value     string               `json:"value"`
	ValueFrom *corev1.EnvVarSource `json:"valueFrom,omitempty" protobuf:"bytes,3,opt,name=valueFrom"`
}
type LBPort struct {
	Value interface{}
}

type Ports struct {
	LbPort   interface{} `json:"lbPort"`
	Name     string      `json:"name"`
	Port     int         `json:"port"`
	Protocol string      `json:"protocol"`
}
type SecurityContext struct {
	FsGroup      int64 `json:"fsGroup"`
	RunAsGroup   int64 `json:"runAsGroup"`
	RunAsNonRoot bool  `json:"runAsNonRoot"`
	RunAsUser    int64 `json:"runAsUser"`
}

type StartParams struct {
	Description string `json:"description"`
	ModuleName  string `json:"module_name"`
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	ValuesText  string `json:"values_text"`
	Lock        bool   `json:"lock"` // 是否锁定 更新时候锁定
}
type Volumes struct {
	MountPath string `json:"mountPath"`
	SubPath   string `json:"subPath"`
	Type      string `json:"type"`
	Name      string `json:"name"` // 每个存储取名
	// HostPathType string `json:"hostPathType"`
	HostPath HostPath `json:"hostPath"`
}
type HostPath struct {
	Path string `json:"path"`
	Type string `json:"type"`
}
type BaseInfo struct {
	Identifie   string `json:"identifie"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type Container struct {
	BaseInfo            BaseInfo        `json:"baseInfo"`
	Build               Build           `json:"build"`
	Cmd                 []string        `json:"cmd"`
	ContainerPort       int             `json:"containerPort"`
	CPU                 float32         `json:"cpu"`
	CustomLogs          string          `json:"customLogs"`
	Env                 []Env           `json:"env"`
	Image               string          `json:"image"`
	InitialDelaySeconds int             `json:"initialDelaySeconds"`
	MaxNum              int             `json:"maxNum"`
	Mem                 float32         `json:"mem"`
	MinNum              int             `json:"minNum"`
	PolicyThreshold     int             `json:"policyThreshold"`
	PolicyType          string          `json:"policyType"`
	Ports               []Ports         `json:"ports"`
	Privileged          BoolOrString    `json:"privileged"`
	RuntimeClassName    string          `json:"runtimeClassName"`
	SecurityContext     SecurityContext `json:"securityContext"`
	StartParams         []StartParams   `json:"startParams"`
	Volumes             []Volumes       `json:"volumes"`
	Shells              []Shell         `json:"shells"`
}

// IsPrivileged 返回Container是否有特权
func (c *Container) IsPrivileged() bool {
	return c.Privileged.Bool()
}

type DependsOn struct {
	Name         string `json:"name"`
	Identifie    string `json:"identifie"`
	SubIdentifie string `json:"subidentifie"`
	SubName      string `json:"subname"`
	Required     bool   `json:"required"`
}

type Shell struct {
	Shell     string `json:"shell"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	SearchJob string `json:"searchJob"` //搜索job的名称
	Image     string `json:"image"`     //执行shell job的镜像 空使用默认当前镜像
}

func (s *Shell) GetShell() string {
	return s.Shell
}
func (s *Shell) GetTitle() string {
	if s.Title == "" {
		if s.Type == "install" {
			return "安装脚本"
		}
		if s.Type == "upgrade" {
			return "更新脚本"
		}
	}
	return s.Title
}

func (s *Shell) GetDisployTitle() string {
	if s.Type == "install" {
		return "[应用安装时触发]" + s.GetTitle()
	} else if s.Type == "upgrade" {
		return "[应用更新时触发]" + s.GetTitle()
	} else if s.Type == "uninstall" {
		return "[应用卸载时触发]" + s.GetTitle()
	} else if s.Type == "requireinstall" {
		return "[应用被安装时触发]" + s.GetTitle()
	}
	return "[自定义触发]" + s.GetTitle()
}
func (s *Shell) GetType() string {
	return s.Type
}

func (s *Shell) GetImage() string {
	return s.Image
}

type Depends struct {
	Identifie    string `json:"identifie"`
	Name         string `json:"name"`
	Required     bool   `json:"required"`
	From         string `json:"from"`
	Type         string `json:"type"`
	SubIdentifie string `json:"subidentifie"`
	SubName      string `json:"subname"`
}

type Helm struct {
	ChartName  string `json:"chartName"`
	Repository string `json:"repository"`
	Version    string `json:"version"`
}

func (m *Helm) GetChartName() string {
	return m.ChartName
}
func (m *Helm) GetRepository() string {
	return m.Repository
}
func (m *Helm) GetVersion() string {
	return m.Version
}

func (m *Depends) GetLoadUrl(p *ManifestPackage) string {
	from := m.From
	dependUrl := ""
	if "" != from {
		dependUrl = from + "/respo/info/" + m.Identifie
	}
	if dependUrl == "" {
		dependUrl = p.ZpkUrl + "/" + m.Identifie
	}
	return dependUrl
}

type Platform struct {
	Container Container   `json:"container"`
	Depends   []Depends   `json:"depends"`
	DependsOn []DependsOn `json:"dependsOn"`
	Helm      Helm        `json:"helm"`
	Supports  []string    `json:"supports"`
	Ingress   []Ingress   `json:"ingress"`
}

type ZpkInfo struct {
	Code int  `json:"code"`
	Data Data `json:"data"`
}
type Version struct {
	ID          int       `json:"id"`
	FormulaID   int       `json:"formula_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
type Data struct {
	Manifest    string            `json:"manifest"`
	Version     Version           `json:"version"`
	ZipURL      string            `json:"zip_url"`
	HelmUrl     string            `json:"helm_url"`
	OciURL      string            `json:"oci_url"`
	WebZipURL   map[string]string `json:"webzip_url"`
	ReleaseName string            `json:"app_name"` //控制台接口用这个字段
	DeployItems []DeployItem      `json:"deploy_items"`
	IconUrl     string            `json:"icon_url"`
	Ticket      string            `json:"ticket"`
}

// backend start
type Header struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value"`
}
type Query struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value"`
}
type MoreMatch struct {
	Header []Header `json:"header"`
	Method []string `json:"method"`
	Query  []Query  `json:"query"`
}
type Rewrite struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

type Backend struct {
	Name      string    `json:"name"`
	Port      int       `json:"port"`
	SvcName   string    `json:"svc_name"` // 每个后端服务取名 (后端服务名称) 后期追加
	IngName   string    `json:"ing_name"` //	每个后端服务对应的ingress取名 后期追加
	Match     string    `json:"match"`
	MoreMatch MoreMatch `json:"moreMatch"`
	Rewrite   Rewrite   `json:"rewrite"`
}
type Routes struct {
	Backend Backend `json:"backend"`
	Path    string  `json:"path"`
}

// backend end
func (m *Routes) GetAnnatationsKey(pathType string, mkey string, queryOrHeader string) string {
	switch strings.ToLower(pathType) {
	case "exact":
		return "exact-match-" + queryOrHeader + "-" + mkey
	case "prefix":
		return "prefix-match-" + queryOrHeader + "-" + mkey
	case "regex":
		return "regex-match-" + queryOrHeader + "-" + mkey
	}
	return ""
}
func (m *Routes) GetPathType() string {
	if m.Backend.Match == "" {
		return "Prefix"
	}
	if strings.ToLower(m.Backend.Match) == "prefix" {
		return "Prefix"
	}
	if strings.ToLower(m.Backend.Match) == "exact" {
		return "Exact"
	}
	return "Prefix"
}
func (m *Routes) GetAnnatations() map[string]string {
	result := make(map[string]string)
	if strings.ToLower(m.Backend.Match) == "exact" {
		result["higress.io/use-regex"] = "true"
	}
	if m.Backend.Rewrite.Host != "" {
		result["higress.io/enable-rewrite"] = "true"
		result["higress.io/upstream-vhost"] = "/" + m.Backend.Rewrite.Host
	}
	if m.Backend.Rewrite.Path != "" {
		result["higress.io/enable-rewrite"] = "true"
		result["higress.io/rewrite-target"] = m.Backend.Rewrite.Path
	}
	if m.Backend.MoreMatch.Method != nil {
		result["higress.io/match-method"] = strings.Join(m.Backend.MoreMatch.Method, " ")
	}
	if m.Backend.MoreMatch.Query != nil {
		for _, v := range m.Backend.MoreMatch.Query {
			aKey := m.GetAnnatationsKey(v.Type, v.Key, "query")
			result[aKey] = v.Value
		}
	}
	if m.Backend.MoreMatch.Header != nil {
		for _, v := range m.Backend.MoreMatch.Header {
			aKey := m.GetAnnatationsKey(v.Type, v.Key, "header")
			result[aKey] = v.Value
		}
	}

	return result
}
func (m *Routes) GetBackendName() string {
	return m.Backend.Name
}
func (m *Routes) GetBackendPort() int32 {
	return int32(m.Backend.Port)
}
func (m *Routes) GetPath() string {
	return m.Path
}

// ingress 名称
func (m *Routes) GetIngName() string {
	return m.Backend.IngName
}

type Ingress struct {
	Name   string   `json:"name"`
	Routes []Routes `json:"routes"`
}

func (m *Ingress) ReplaceSvcName(resource zpktype.K8sResourceIngressInterface) {
	for i := 0; i < len(m.Routes); i++ {
		backend := m.Routes[i].Backend
		m.Routes[i].Backend.SvcName = resource.GetIngressSvcName(backend.Name)
		m.Routes[i].Backend.IngName = "ing-" + helper.RandomString(10)
	}
}

func (m *Manifest) RequireCreateDb() (bool, string) {
	for _, v := range m.Platform.Container.StartParams {
		if v.ModuleName == "w7_mysql.DB_NAME" || v.ModuleName == "w7_mysql5.DB_NAME" || v.ModuleName == "w7_mysql57.DB_NAME" {
			return true, v.ValuesText
		}
	}
	return false, ""
}

func (m *Manifest) RequireCreateDbUser() (bool, string, string) {
	uname := ""
	password := ""
	for _, v := range m.Platform.Container.StartParams {
		if v.ModuleName == "w7_mysql.DB_USERNAME" || v.ModuleName == "w7_mysql5.DB_USERNAME" {
			uname = v.ValuesText
		}
		if v.ModuleName == "w7_mysql.DB_PASSWORD" || v.ModuleName == "w7_mysql5.DB_PASSWORD" {
			password = v.ValuesText
		}
	}
	return uname != "" && password != "", uname, password
}

// 给shell 添加搜索标签名称
func (m *Manifest) ReplaceShellJobName(name string) {
	for k, v := range m.Platform.Container.Shells {
		v.SearchJob = name + "-" + v.Type
		m.Platform.Container.Shells[k] = v
	}

}

func (m *Manifest) GetStartParamsModuleNames() []string {
	result := []string{}
	for _, v := range m.Platform.Container.StartParams {
		if v.ModuleName != "" {
			vm := v.ModuleName
			if strings.Contains(v.ModuleName, ".") {
				vm = strings.Split(v.ModuleName, ".")[0]
			}
			result = append(result, vm)
		}
	}
	return lo.Uniq(result)

}

func (m *Manifest) GetFirstPort() int32 {
	if len(m.Platform.Container.Ports) == 0 {
		return int32(m.Platform.Container.ContainerPort)
	}
	return int32(m.Platform.Container.Ports[0].Port)
}

func (m *Manifest) IsHelm() bool {
	return m.Application.Type == "helm"
}

func (m *Manifest) IsOnce() bool {
	return m.Application.Once
}

func (m *Manifest) GetOutDepends() []Depends {
	depends := m.Platform.Depends
	depends = lo.Filter(depends, func(item Depends, index int) bool {
		return item.Type == "out"
	})
	dependsName := lo.Map(depends, func(item Depends, index int) string {
		return item.Name
	})

	for _, on := range m.Platform.DependsOn {
		if lo.Contains(dependsName, on.Name) {
			continue
		}
		dependObj := Depends{
			Name:         on.Name,
			Type:         "out",
			Required:     on.Required,
			SubIdentifie: on.SubIdentifie,
			SubName:      on.SubName,
			Identifie:    on.Identifie,
		}
		depends = append(depends, dependObj)
	}
	return depends
}

func (m *Manifest) GetHelmNamespce() string {

	for _, param := range m.Platform.Container.Env {
		if param.Name == "HELM_NAMESPACE" {
			return param.Value
		}
	}
	return ""
}
