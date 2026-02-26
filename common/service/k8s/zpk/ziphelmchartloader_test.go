package zpk

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZipHelmChartLoader_Load(t *testing.T) {
	// tempDir, err := os.MkdirTemp(os.TempDir(), "helm-charts-test-")
	// require.NoError(t, err)
	// defer os.RemoveAll(tempDir)
	// err = copyChart(tempDir)
	// require.NoError(t, err)
	zcl := ZipHelmChartLoader{zipFile: filepath.Join("./testdata/demo.zip")}
	chart, err := zcl.Load()
	require.NoError(t, err)
	assert.Equal(t, "demo", chart.Name())
	assert.Equal(t, "A Helm chart for Kubernetes", chart.Metadata.Description)
	assert.Equal(t, "0.1.0", chart.Metadata.Version)
	assert.Len(t, chart.Templates, 7)

	// assert.Equal(t, "templates/NOTES.txt", chart.Templates[0].Name)
	// assert.Len(t, chart.Files, 1)
	// assert.Equal(t, "templates/empty-dir", chart.Files[0].Name)
	// assert.True(t, chart.Files[0].IsDir)
}
