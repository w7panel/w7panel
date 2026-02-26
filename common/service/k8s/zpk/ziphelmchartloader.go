package zpk

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

type ZipHelmChartLoader struct {
	zipFile string `json:"zip_file"`
}

func NewZipHelmChartLoader(zipFile string) *ZipHelmChartLoader {
	return &ZipHelmChartLoader{zipFile: zipFile}
}

func (self ZipHelmChartLoader) Load() (*chart.Chart, error) {

	filename, err := self.getZipPath()
	if err != nil {
		return nil, err
	}
	zrc, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	if len(zrc.File) == 0 {
		return nil, errors.New("zip file does not contain any files")
	}

	defer zrc.Close()

	files := []*loader.BufferedFile{}
	for _, file := range zrc.File {
		if file.FileInfo().IsDir() {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		filename := file.Name
		if filename == "" {
			continue
		}

		files = append(files, &loader.BufferedFile{Name: filename, Data: content})
	}
	return loader.LoadFiles(files)
}

func (self ZipHelmChartLoader) getZipPath() (string, error) {
	url, err := url.Parse(self.zipFile)
	if err != nil {
		return "", err
	}
	if (url.Scheme != "http" && url.Scheme != "https") || url.Host == "" {
		return self.zipFile, nil
	}
	if url.Scheme == "https" || url.Scheme == "http" {
		resp, err := http.Get(self.zipFile)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		file, err := os.CreateTemp(os.TempDir(), "*.zip")
		if err != nil {
			return "", err
		}
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		_, err = file.Write(bytes)
		if err != nil {
			return "", err
		}
		return file.Name(), err
	}
	return "", errors.New("not implemented yet")
}
