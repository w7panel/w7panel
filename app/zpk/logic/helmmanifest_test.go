package logic

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"github.com/stretchr/testify/assert"
)

func TestHelmManifestApp(t *testing.T) {
	helmMemory := &HelmMemory{
		Identifie:   "helm-test",
		Title:       "helm-test",
		Icon:        "helm-test",
		Description: "helm-test",
		ChartName:   "helm-test",
		Repository:  "helm-test",
		Version:     "helm-test",
		Kv: []types.EnvKv{
			{Name: "helm-test7", Value: "helm-test8"},
		},
	}
	manifest := HelmManifestApp(helmMemory)
	assert.Equal(t, "helm-test", manifest.Application.Identifie)
	assert.Equal(t, "helm-test", manifest.Application.Name)
	assert.Equal(t, "helm-test", manifest.Application.Icon)
	assert.Equal(t, "helm-test", manifest.Application.Description)
	assert.Equal(t, "helm-test", manifest.Platform.Helm.ChartName)
	assert.Equal(t, "helm-test", manifest.Platform.Helm.Repository)
	assert.Equal(t, "helm-test", manifest.Platform.Helm.Version)
	assert.Equal(t, 1, len(manifest.Platform.Container.StartParams))
	assert.Equal(t, "test", manifest.Platform.Container.StartParams[0].Name)
	assert.Equal(t, "test", manifest.Platform.Container.StartParams[0].ValuesText)
}
