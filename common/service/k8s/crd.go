package k8s

import (
	"os"
	"strings"
)

type CRD struct {
	*Sdk
	baseDir string
}

func NewCRD(sdk *Sdk) *CRD {
	baseDir, ok := os.LookupEnv("KO_DATA_PATH")
	if !ok {
		baseDir = "../../../kodata"
	}
	return &CRD{
		Sdk:     sdk,
		baseDir: baseDir,
	}
}

func (c *CRD) AppGpuClass() error {
	data, err := os.ReadFile(c.baseDir + "/yaml/nvidia.yaml")
	if err != nil {
		return err
	}
	err = c.ApplyYaml(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *CRD) ApplyCrds() error {
	///读取目录下的所有yaml文件
	dirName := c.baseDir + "/crds/"
	err := c.Sdk.ApplyFile(dirName, *NewApplyOptionsServerSide(c.Sdk.namespace, true))
	if err != nil {
		return err
	}
	return nil
}

func (c *CRD) ApplyCrdsOld() error {
	///读取目录下的所有yaml文件
	files, err := os.ReadDir(c.baseDir + "/crds/")
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(c.baseDir + "/crds/" + file.Name())
		if err != nil {
			return err
		}
		err = c.ApplyYaml(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CRD) ApplyYaml(data []byte) error {
	err := c.Sdk.ApplyBytes(data, *NewApplyOptionsServerSide(c.Sdk.namespace, true))
	if err != nil {
		return err
	}
	return nil
}
