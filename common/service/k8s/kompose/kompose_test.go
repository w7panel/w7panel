package kompose

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_kompose_Covert(t *testing.T) {
	dockerComposerYaml := []byte(`version: '3'
services:
  webapp:
    image: nginx
`)
	// sdk := NewSdk("", "", "", "", "", "")
	// kompose.WithSdk(sdk)
	result, err := ConvertToK8sYaml(dockerComposerYaml)
	assert.Nil(t, err)
	assert.Contains(t, result, "webapp-deployment.yaml")
}

func Test_kompose_Covert_error(t *testing.T) {
	dockerComposerYaml := []byte(`version: '3'
services:
  webapp:
    image: nginx
`)
	os.Setenv("HOME", "/") // make kompose.createDir() return error
	defer os.Unsetenv("HOME")
	// sdk := NewSdk("", "", "", "", "", "")
	// kompose.WithSdk(sdk)
	result, err := ConvertToK8sYaml(dockerComposerYaml)
	print(result)
	assert.NotNil(t, err)
}

// func Test_ApplyAsHelmNew(t *testing.T) {
// 	dockerComposerYaml := []byte(`version: '3'
// services:
//   webapp:
//     image: nginx
// `)
// 	chart, err := ConvertToHelmChart(dockerComposerYaml)
// 	if err != nil {
// 		return
// 	}

// 	sdk := k8s.NewK8sClient()
// 	helm := k8s.NewHelm(sdk)
// 	var values = make(map[string]interface{})
// 	releaseName := "kompose" + uuid.New().String()
// 	res, err := helm.Install(context.TODO(), chart, values, releaseName, "default")

// 	if err != nil {
// 		panic(err)
// 	}
// 	assert.NotNil(t, res)
// }
