package logic

import (
	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
)

func downStatic(app *types.PackageApp) {
	webzipUrl := app.ManifestPackage.WebZipUrl
	microappPath := facade.Config.GetString("static.microapp_path")
	if len(webzipUrl) > 0 {
		appgroup.DownStaticMap(webzipUrl, app.GetReleaseName(), microappPath)
	}
}
