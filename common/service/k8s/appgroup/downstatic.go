package appgroup

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

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

// https://zpk.fan.b2.sz.w7.com/zpk/respo/info/nginx_test
type Manifest struct {
	Application struct {
		Identifie string `json:"identifie"`
	}
	V   int64 `json:"v"`
	Web struct {
		Type   string `json:"type"`
		ZipURL string `json:"url"` //示例值: file:///Storage/202601/479d272753da0495156a8e5c1b12d2b6.zip
	} `json:"web"`
}
type Data struct {
	ManifestStr string `json:"manifest"`
	Manifest    Manifest
	Version     Version           `json:"version"`
	ZipURL      string            `json:"zip_url"`
	HelmUrl     string            `json:"helm_url"`
	WebZipURL   map[string]string `json:"webzip_url"`
	ReleaseName string            `json:"app_name"` //控制台接口用这个字段
}

func downStatic(appgroup *v1alpha1.AppGroup) {
	downEnv := os.Getenv("STATIC_DOWN_ENABLED")
	if downEnv != "true" {
		slog.Info("静态资源下载未开启")
		return
	}
	frontTypeStr, ok := appgroup.Annotations["w7.cc/front-type"]
	if !ok {
		return
	}
	if strings.Contains(frontTypeStr, "thirdparty_cd") {
		go k3k.SyncDownStatic(appgroup.Name, appgroup.Spec.ZpkUrl)
		go fetchWebZipAndDownload(appgroup.Spec.ZpkUrl, appgroup.Name)
	}
}
func DownStatic(zpkurl, name string) error {
	return fetchWebZipAndDownload(zpkurl, name)
}

func DownStaticGo(zpkurl, name string) {
	go fetchWebZipAndDownload(zpkurl, name)
}
func fetchWebZipAndDownload(zpkUrl string, releaseName string) error {
	resp, err := helper.RetryHttpClient().R().Get(zpkUrl)
	if err != nil {
		return err
	}
	defer resp.RawBody().Close()
	if resp.StatusCode() != http.StatusOK {
		return errors.New(resp.String())
	}

	body := resp.Body()
	var zpkInfo ZpkInfo
	if err := json.Unmarshal(body, &zpkInfo); err != nil {
		return err
	}
	manifestJson, err := yaml.ToJSON([]byte(zpkInfo.Data.ManifestStr))
	if err != nil {
		return err
	}
	var manifest Manifest
	if err := json.Unmarshal([]byte(manifestJson), &manifest); err != nil {
		return err
	}
	zpkInfo.Data.Manifest = manifest
	webzipUrl := zpkInfo.Data.WebZipURL
	microappPath := os.Getenv("MICROAPP_PATH") //facade.Config.GetString("static.microapp_path")
	if len(webzipUrl) > 0 {
		DownStaticMap(webzipUrl, releaseName, microappPath)
	}
	return nil
	// if zpkInfo.Data.Manifest.V <= 1 {
	// 	webzipUrl := zpkInfo.Data.WebZipURL
	// 	microappPath := facade.Config.GetString("static.microapp_path")
	// 	if len(webzipUrl) > 0 {
	// 		DownStaticMap(webzipUrl, releaseName, microappPath)
	// 	}
	// } else {
	// 	// 1. 下载zpkInfo.HelmUrl tgz文件
	// 	if zpkInfo.Data.HelmUrl == "" {
	// 		slog.Error("HelmUrl为空", "zpkInfo", zpkInfo)
	// 		return errors.New("HelmUrl为空")
	// 	}
	// 	//file:///Storage/202601/479d272753da0495156a8e5c1b12d2b6.zip
	// 	//zpkInfo.Data.Manifest.Web.ZipUR clear file://Storage
	// 	webZipPathClear := strings.Replace(zpkInfo.Data.Manifest.Web.ZipURL, "file:///Storage/", "", 1)
	// 	// webZipPathClear = "" + webZipPathClear
	// 	// 2.读取tgz文件中 files目录中的zip文件 zip文件地址是 files/{zpkInfo.Manifest.Web.ZipURL}
	// 	webZipPath := fmt.Sprintf(zpkInfo.Data.Manifest.Application.Identifie+"/files/Storage/%s", webZipPathClear)
	// 	zipData, err := helper.ExtractSingleFileFromTgz(zpkInfo.Data.HelmUrl, webZipPath)
	// 	if err != nil {
	// 		slog.Error("从tgz提取zip文件失败", "error", err, "helmUrl", zpkInfo.Data.HelmUrl, "zipPath", webZipPath)
	// 		return err
	// 	}

	// 	// 创建临时目录保存zip文件
	// 	tempDir := "/tmp/static-temp"
	// 	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
	// 		slog.Error("创建临时目录失败", "error", err)
	// 		return err
	// 	}

	// 	tempZipPath := fmt.Sprintf("%s/%s", tempDir, filepath.Base(zpkInfo.Data.Manifest.Web.ZipURL))
	// 	if err := os.WriteFile(tempZipPath, zipData, 0644); err != nil {
	// 		slog.Error("保存zip文件失败", "error", err)
	// 		return err
	// 	}

	// 	// 3. 解压zip文件到microappPath目录中
	// 	microappPath := os.Getenv("MICROAPP_PATH")
	// 	if err := os.MkdirAll(microappPath, os.ModePerm); err != nil {
	// 		slog.Error("创建目标目录失败", "error", err)
	// 		return err
	// 	}

	// 	targetPath := fmt.Sprintf("%s/%s", microappPath, releaseName)
	// 	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
	// 		slog.Error("创建应用目录失败", "error", err)
	// 		return err
	// 	}

	// 	// 解压zip文件
	// 	if err := extractZipToDir(tempZipPath, targetPath); err != nil {
	// 		slog.Error("解压zip文件失败", "error", err)
	// 		return err
	// 	}

	// 	// 清理临时文件
	// 	if err := os.RemoveAll(tempDir); err != nil {
	// 		slog.Error("清理临时文件失败", "error", err)
	// 		return err
	// 	}

	// 	slog.Info("静态资源下载成功", "releaseName", releaseName, "targetPath", targetPath)
	// }

	// return nil
}

// extractZipToDir 解压zip文件到指定目录
func extractZipToDir(zipPath, destDir string) error {
	// 打开zip文件
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// 遍历zip文件中的每个文件
	for _, file := range zipReader.File {
		// 构造文件在目标目录中的路径
		filePath := filepath.Join(destDir, file.Name)

		// 如果是目录，创建目录
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// 确保文件的父目录存在
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		// 打开zip中的文件
		fileInArchive, err := file.Open()
		if err != nil {
			return err
		}
		defer fileInArchive.Close()

		// 创建目标文件
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// 复制文件内容
		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}
	}

	return nil
}

func DownStaticMap(webzipUrl map[string]string, releaseName, microappPath string) error {
	if len(webzipUrl) > 0 {
		// 下载静态资源包
		for k, url := range webzipUrl {
			// os.Stat(microappPath + "/" + k)
			err := os.Mkdir(microappPath, os.ModePerm)
			if err != nil {
				slog.Error("创建目录失败", "error", err)
				// continue
			}
			err = os.Mkdir(microappPath+"/"+releaseName, os.ModePerm) // 创建目录，如果不存在则创建 ingore err
			if err != nil {
				slog.Error("创建目录失败", "error", err)
				// continue
			}
			err = downStaticFile(url, microappPath+"/"+releaseName+"/"+k+".zip")
			if err != nil {
				slog.Error("下载静态资源包失败", "error", err)
				continue
			}
		}
	}
	return nil
}

// 下载文件到指定目录
func downStaticFile(url string, zipfile string) error {
	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	defer os.Remove(zipfile)

	// Create the zip file
	zipFile, err := os.Create(zipfile)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// Save the downloaded content to zip file
	_, err = io.Copy(zipFile, resp.Body)
	if err != nil {
		return err
	}

	// Open the zip file for reading
	zipReader, err := zip.OpenReader(zipfile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// Extract each file from the zip archive
	for _, file := range zipReader.File {
		filePath := filepath.Join(filepath.Dir(zipfile), file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := file.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			fileInArchive.Close()
			dstFile.Close()
			return err
		}

		fileInArchive.Close()
		dstFile.Close()
	}

	return nil
}
