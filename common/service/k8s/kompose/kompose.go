package kompose

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kubernetes/kompose/client"
	stdlog "github.com/sirupsen/logrus"
)

type writer struct {
	io.Writer
	err error
}

func (w *writer) Write(p []byte) (int, error) {
	fmt.Println(string(p))
	w.err = errors.New(string(p))
	return 0, nil
}
func (w *writer) Error() error {
	return w.err
}

func ConvertToK8sYaml(dockerComposerYaml []byte) (map[string]string, error) {
	errCode := make(chan int)
	buffertxt := bytes.NewBuffer([]byte{})
	stdlog.StandardLogger().ExitFunc = func(code int) {
		errCode <- code
	}
	stdlog.SetOutput(buffertxt)

	result := map[string]string{}

	option := client.ConvertOptions{
		PushImage:              false,
		Build:                  nil,
		GenerateJson:           false,
		ToStdout:               false,
		Replicas:               nil,
		VolumeType:             nil,
		PvcRequestSize:         "2Gi",
		WithKomposeAnnotations: nil,
		Profiles:               nil,
	}
	go func() {
		var err error
		result, err = convert(dockerComposerYaml, option)
		if err != nil {
			stdlog.Fatalln(err.Error())
		}
		errCode <- 0
	}()

	select {
	case code := <-errCode:
		if code == 1 {
			return result, errors.New(buffertxt.String())
		}
		return result, nil
	}

}

// 1. 生成随即/tmp/xxx目录
// 2. 读取请求中docker-compose.yaml
// 3. 执行ko.Convert()
// 4. 遍历/tmp/xxx 目录 返回给客户端
func convert(dockerComposerYaml []byte, option client.ConvertOptions) (map[string]string, error) {
	outDir, err := os.MkdirTemp("", "kompose")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(outDir)

	inputFile, err := os.CreateTemp("", "docker-compose")
	if err != nil {
		return nil, err
	}
	// defer os.Remove(inputFile.Name())
	_, err = inputFile.Write(dockerComposerYaml)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	ko, err := client.NewClient()
	if err != nil {
		return nil, err
	}

	option.OutFile = outDir
	option.InputFiles = []string{inputFile.Name()}
	_, err = ko.Convert(option)
	if err != nil {
		return nil, err
	}

	fileNameAndContent := map[string]string{}
	err = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relName, err := filepath.Rel(outDir, path)
			if err != nil {
				return err
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			content, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			fileNameAndContent[relName] = string(content)
		}
		return nil
	})

	return fileNameAndContent, err

}
