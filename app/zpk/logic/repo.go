package logic

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type repo struct {
	repoUrl        string `json:"repo_url"`
	token          string `json:"token"`
	panelToken     string `json:"panel_token"`
	scheme         string `json:"scheme"`
	IsConsole      bool   `json:"is_console"` //是否需要从console
	baseConsoleUrl string `json:"base_console_url"`
	upgrade        bool   `json:"upgrade"`       // 是否升级包
	checkUpgrade   bool   `json:"check_upgrade"` // 是否升级包
	curVersion     string `json:"cur_version"`
	// loadInnerDepends bool   `json:"load_inner_depends"` // 是否加载内部依赖
}

func NewRepo(repoUrl, token, baseConsoleUrl string) *repo {
	return &repo{
		repoUrl:        repoUrl,
		token:          token,
		baseConsoleUrl: "https://console.w7.cc/api/thirdparty-cd/k8s-offline/",
	}
}

func LoadPackage(url string) (*types.ManifestPackage, error) {
	// slog.Info("load package from %v", url)
	return NewRepo(url, "", "").Load()

}

func LoadPackage2(uri string, token string, checkUpgrade bool) (*types.ManifestPackage, error) {
	//获取source scheme
	scheme := getSourceUri(uri)
	if scheme == "" {
		return nil, errors.New("uri is not valid")
	}
	repo := NewRepo(uri, token, "")
	repo.SetCheckUpgrade(checkUpgrade)
	return repo.Load()
	// if (scheme == "http" || scheme == "https") && strings.HasPrefix(uri, "http") {
	// 	return LoadPackageByHttp(uri)
	// } else {
	// 	return LoadPackageFromConsole(uri)
	// }

	// return nil, errors.New("scheme is not supported")
}

func (self *repo) SetUpgrade(upgrade bool) string {
	self.upgrade = upgrade
	return self.repoUrl
}

func (self *repo) SetCheckUpgrade(upgrade bool) string {
	self.checkUpgrade = upgrade
	return self.repoUrl
}

func (self *repo) SetPanelToken(token string) {
	self.panelToken = token
}

func (self *repo) SetCurVersion(version string) {
	self.curVersion = version
}

func (self *repo) getConsoleUrl() string {
	return self.baseConsoleUrl + "config?url=" + self.repoUrl
}

func (self *repo) getPreInstallUrl() string {
	return self.baseConsoleUrl + "/pre-install" + "?url=" + self.repoUrl
}

func (self *repo) loadPackageFromConsole() (*types.ManifestPackage, error) {
	self.IsConsole = true
	consoleUrl := self.getConsoleUrl()
	token := self.token
	return self.loadPackageByHttp(consoleUrl, token, true)
}
func (self *repo) Load() (*types.ManifestPackage, error) {
	scheme := getSourceUri(self.repoUrl)
	if scheme == "" {
		return nil, errors.New("uri is not valid")
	}

	if scheme == "http" || scheme == "https" {
		return self.loadPackageByHttp(self.repoUrl, self.token, true)
	} else if scheme == "memory" {
		return self.loadPackageByHelmMemory(self.repoUrl)
	} else {
		return self.loadPackageFromConsole()
	}

}

func (self *repo) loadPackageByHelmMemory(uri string) (*types.ManifestPackage, error) {
	// 发送http请求 从uri获取json 数据

	kv := NewManifestSingleton()
	host, err := getSourceHost(uri)
	if err != nil {
		return nil, err
	}
	manifset, ok := kv.Get(host)
	if !ok {
		return nil, errors.New("not found")
	}
	m := *manifset
	p := &types.ManifestPackage{
		Manifest: m,
		// Title :    manifest.Application.Name,
		// Identifie: manifest.Application.Identifie,
		ZpkUrl: self.repoUrl,
		ZipUrl: "",
		Version: types.Version{
			Name: m.Platform.Helm.Version,
		},

		ConsoleReleaseName: "",
	}
	p.RequireInstall = true
	return p, nil
}

func (self *repo) PreInstall(clusterId string) (*console.PreInstall, error) {
	if !self.IsConsole {
		return nil, errors.New("not console url")
	}
	consoleClient := console.NewConsoleCdClient(self.token)
	return consoleClient.PreInstall(self.repoUrl, clusterId)

}

func (self *repo) loadPackageByHttp(uri string, token string, isParent bool) (*types.ManifestPackage, error) {
	// 发送http请求 从uri获取json 数据
	req := helper.RetryHttpClient().R().SetAuthToken(token)
	if self.panelToken != "" {
		req.SetHeader("X-W7Panel-Token", self.panelToken)
	}
	if self.upgrade {
		req.SetQueryParam("is_upgrade", "1")
	}
	if self.checkUpgrade {
		req.SetQueryParam("check_upgrade", "1")
	}
	if self.curVersion != "" {
		req.SetQueryParam("cur_version", self.curVersion)
	}
	resp, err := req.Get(uri)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New(resp.String())
	}

	body := resp.Body()
	// if err != nil {
	// 	return nil, err
	// }
	// 解析json 数据
	var zpkInfo types.ZpkInfo
	if err := json.Unmarshal(body, &zpkInfo); err != nil {
		return nil, err
	}
	manifestStr := zpkInfo.Data.Manifest

	manifestJson, err := yaml.ToJSON([]byte(manifestStr))
	if err != nil {
		return nil, err
	}
	var manifest types.Manifest
	err = json.Unmarshal(manifestJson, &manifest)
	if err != nil {
		return nil, err
	}
	manifest.GenVolumesName(make(map[string]string))
	if self.IsConsole {
		manifest.AppendDomainStartParams()
	}
	// manifest.Application.Identifie = strings.ToLower(manifest.Application.Identifie)
	p := &types.ManifestPackage{
		Manifest: manifest,
		// Title :    manifest.Application.Name,
		// Identifie: manifest.Application.Identifie,
		HelmUrl:            zpkInfo.Data.HelmUrl,
		ZpkUrl:             self.repoUrl,
		ZipUrl:             zpkInfo.Data.ZipURL,
		OciUrl:             zpkInfo.Data.OciURL,
		WebZipUrl:          zpkInfo.Data.WebZipURL,
		Version:            zpkInfo.Data.Version,
		ConsoleReleaseName: zpkInfo.Data.ReleaseName,
		DeployItems:        zpkInfo.Data.DeployItems,
		IconUrl:            zpkInfo.Data.IconUrl,
		Ticket:             zpkInfo.Data.Ticket,
	}

	// if p.HelmUrl != "" {
	// 	p.Manifest.Application.Type = "helm"
	// 	p.Manifest.Platform.Helm.ChartName = p.HelmUrl
	// 	p.Manifest.Platform.Helm.Version = p.Version.Name
	// }

	// if p.OciUrl != "" {
	// 	ociUrl := p.OciUrl
	// 	sdk := k8s.NewK8sClient().Sdk
	// 	oci, err := NewOCI(sdk, ociUrl)
	// 	if err != nil {
	// 		slog.Warn("NewOCI", "err", err)
	// 	}
	// 	if oci != nil {
	// 		p.ZipUrl = oci.GetCodeZipUrl()
	// 		p.WebZipUrl = map[string]string{
	// 			p.Manifest.Application.Identifie: oci.GetWebCodeZipUrl(),
	// 		}
	// 	}

	// }

	uri2, err := parseUri(self.repoUrl)
	if err != nil {
		slog.Warn("parseUri", "err", err)
	}
	// 如果是http 或者 https 并且不是父包 则认为是子包 需要加载父包
	if uri2.Scheme == "http" || uri2.Scheme == "https" && isParent {
		path := uri2.Path
		splitPath := strings.Split(path, "/")
		defaultLen := 4
		isNewZpk := strings.Contains(path, "/zpk/respo")
		if isNewZpk {
			defaultLen = 5
		}
		if len(splitPath) > defaultLen {
			//p.Manifest.Application.Identifie = splitPath[len(splitPath)-2]
			rootIdentifie := splitPath[3]
			if isNewZpk {
				rootIdentifie = splitPath[4]
			}
			ldurl := uri2.Scheme + "://" + uri2.Host + "/respo/info/" + rootIdentifie
			if isNewZpk {
				ldurl = uri2.Scheme + "://" + uri2.Host + "/zpk/respo/info/" + rootIdentifie
			}
			parent, err := self.loadPackageByHttp(ldurl, self.token, false)
			if err != nil {
				slog.Error("LoadParentErr", "err", err)
			}
			p.RequireInstall = true
			if parent != nil {
				p.Manifest.Application.Icon = parent.Manifest.Application.Icon
				p.Parent = parent
				p.RequireParentReleaseName = true
			}
			// if isParent {
			// 	_ = self.LoadDependsByPackage(p)
			// }
			// return p, nil
		}
	}
	if isParent {
		p.RequireInstall = true
		_ = self.LoadDependsByPackage(p)
	}

	return p, nil
}

func (self repo) LoadDependsByPackage(p *types.ManifestPackage) error {
	p.RequireInstall = true
	// 初始化依赖列表
	p.Children = make(map[string]*types.ManifestPackage)
	// 遍历依赖 下载依赖
	depends := p.Manifest.Platform.Depends
	for _, depend := range depends {
		if depend.Type == "out" {
			continue
		}
		// 下载依赖
		uri := depend.GetLoadUrl(p)
		child, err := self.loadPackageByHttp(uri, self.token, false)
		if err != nil {
			slog.Warn("LoadDependsByPackage", "err", err)
			continue
		}
		child.RequireInstall = depend.Required
		// 添加到依赖列表
		p.Children[depend.Identifie] = child
	}
	return nil
}

func getSourceUri(uri string) string {
	//获取source scheme
	sourceUri := strings.Split(uri, "://")
	if len(sourceUri) != 2 {
		return ""
	}
	return sourceUri[0]
}

func getSourceHost(uri string) (string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}

func parseUri(uri string) (*url.URL, error) {
	parsedURL, err := url.Parse(uri)
	return parsedURL, err
}
