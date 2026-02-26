package logic

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	appv1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	index "helm.sh/helm/v3/pkg/repo"
)

type HelmConfig struct {
	ChartName  string `json:"chartName"`  //helm chart name
	Repository string `json:"repository"` //helm chart repository
	Version    string `json:"version"`    //版本号
}

type UpgradeInfo struct {
	Version     string `json:"version"`
	CanUpgrade  bool   `json:"canUpgrade"`
	Description string `json:"description"`
	ZpkUrl      string `json:"zpkUrl"`
	HelmConfig  *HelmConfig
}

var NotUpgrade = &UpgradeInfo{
	Version:     "",
	CanUpgrade:  false,
	Description: "已是最新版本",
	ZpkUrl:      "",
}

type UpgradeCheck struct {
	sdk               *k8s.Sdk
	groupApi          *appgroup.AppGroupApi
	helmApi           *k8s.Helm
	thirdPartyCDToken string
}

func NewUpgradeCheck(sdk *k8s.Sdk) *UpgradeCheck {
	groupApi, err := appgroup.NewAppGroupApi(sdk)
	if err != nil {
		return nil
	}
	helmApi := k8s.NewHelm(sdk)
	return &UpgradeCheck{
		sdk:      sdk,
		groupApi: groupApi,
		helmApi:  helmApi,
	}
}

func (u *UpgradeCheck) WithCDToken(cdToken string) {
	u.thirdPartyCDToken = cdToken
}

func (u *UpgradeCheck) Check(namespace string, groupname string) *UpgradeInfo {
	group, err := u.groupApi.GetAppGroup(namespace, groupname)
	if err != nil {
		return NotUpgrade
	}
	if group.Spec.ZpkUrl != "" {
		result, err := u.CheckZpk(group)
		if err != nil {
			return NotUpgrade
		}
		return result
	}
	result2, err := u.CheckHelmRepo(group)
	if err != nil {
		return NotUpgrade
	}
	return result2
}

func (u *UpgradeCheck) CheckHelmRepo(group *appv1.AppGroup) (*UpgradeInfo, error) {
	if group.Spec.HelmConfig.Repository != "" {
		res, err := http.Get(group.Spec.HelmConfig.Repository + "/index.yaml")
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		index, err := loadIndex(data, "k8s-offline")
		if err != nil {
			return nil, err
		}
		enties, ok := index.Entries[group.Spec.HelmConfig.ChartName]
		if ok {
			result := &UpgradeInfo{
				Version:     group.Spec.Version,
				CanUpgrade:  false,
				Description: group.Spec.HelmConfig.ChartName,
				ZpkUrl:      "",
				HelmConfig: &HelmConfig{
					ChartName:  group.Spec.HelmConfig.ChartName,
					Repository: group.Spec.HelmConfig.Repository,
					Version:    group.Spec.Version,
				},
			}
			for _, entry := range enties {
				if semver.Compare(entry.Version, "v"+group.Spec.Version) > 0 {
					result.Version = entry.Version
					result.CanUpgrade = true
				}
			}
			return result, nil
		}

	}
	return NotUpgrade, nil
}

func (u *UpgradeCheck) CheckZpk(group *appv1.AppGroup) (*UpgradeInfo, error) {
	pk, err := LoadPackage2(group.Spec.ZpkUrl, u.thirdPartyCDToken, true)
	if err != nil {
		return nil, err
	}
	newVersion := pk.Version.Name
	appVersion := group.Spec.Version
	//compare version
	// slog.Info("newVersion", "newVersion", newVersion, "appVersion", appVersion)
	if newVersion != "" && newVersion[0] != 'v' {
		newVersion = "v" + newVersion
	}
	if appVersion != "" && appVersion[0] != 'v' {
		appVersion = "v" + appVersion
	}
	mockUpgrade := facade.Config.GetBool("upgrade.mock_upgrade")
	compare := semver.Compare(newVersion, appVersion)
	result := &UpgradeInfo{
		Version: newVersion,
		// CanUpgrade:  true || (compare > 0 && newVersion != ""),
		CanUpgrade:  mockUpgrade || (compare > 0 && newVersion != ""),
		Description: pk.Version.Description,
		ZpkUrl:      group.Spec.ZpkUrl,
	}
	return result, nil
}

func loadIndex(data []byte, source string) (*index.IndexFile, error) {
	i := &index.IndexFile{}

	if len(data) == 0 {
		return i, index.ErrEmptyIndexYaml
	}

	if err := jsonOrYamlUnmarshal(data, i); err != nil {
		return i, err
	}

	for name, cvs := range i.Entries {
		for idx := len(cvs) - 1; idx >= 0; idx-- {
			if cvs[idx] == nil {
				log.Printf("skipping loading invalid entry for chart %q from %s: empty entry", name, source)
				continue
			}
			// When metadata section missing, initialize with no data
			if cvs[idx].Metadata == nil {
				cvs[idx].Metadata = &chart.Metadata{}
			}
			if cvs[idx].APIVersion == "" {
				cvs[idx].APIVersion = chart.APIVersionV1
			}
			if err := cvs[idx].Validate(); ignoreSkippableChartValidationError(err) != nil {
				log.Printf("skipping loading invalid entry for chart %q %q from %s: %s", name, cvs[idx].Version, source, err)
				cvs = append(cvs[:idx], cvs[idx+1:]...)
			}
		}
		// adjust slice to only contain a set of valid versions
		i.Entries[name] = cvs
	}
	i.SortEntries()
	if i.APIVersion == "" {
		return i, index.ErrNoAPIVersion
	}
	return i, nil
}

func jsonOrYamlUnmarshal(b []byte, i interface{}) error {
	if json.Valid(b) {
		return json.Unmarshal(b, i)
	}
	return yaml.UnmarshalStrict(b, i)
}

// ignoreSkippableChartValidationError inspect the given error and returns nil if
// the error isn't important for index loading
//
// In particular, charts may introduce validations that don't impact repository indexes
// And repository indexes may be generated by older/non-compliant software, which doesn't
// conform to all validations.
func ignoreSkippableChartValidationError(err error) error {
	verr, ok := err.(chart.ValidationError)
	if !ok {
		return err
	}

	// https://github.com/helm/helm/issues/12748 (JFrog repository strips alias field)
	if strings.HasPrefix(verr.Error(), "validation: more than one dependency with name or alias") {
		return nil
	}

	return err
}
