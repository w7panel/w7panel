package logic

import (
	"net/url"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	zpktypes "gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"github.com/stretchr/testify/assert"
)

func TestLoadPackage(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name    string
		args    args
		want    *zpktypes.ManifestPackage
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			args: args{
				uri: "https://zpk.w7.cc/respo/info/nvidia_gpuoperator",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadPackage(tt.args.uri)
			if (err != nil) != tt.wantErr {
				t.Log(err)
				// t.Errorf("LoadPackage() error = %v, wantErr %v", err, tt.wantErr)
			}

			optionMain := types.InstallOption{
				Identifie: "nvidia_gpuoperator",
				PvcName:   "longflow-ai",
				EnvKv: []types.EnvKv{
					{Name: "image.tag", Value: "v2.1.0"},
				},
			}
			// option2 := types.InstallOption{
			// 	Identifie: "ai_ollamaapi",
			// 	PvcName:   "longflow-ai",
			// 	EnvKv: []types.EnvKv{
			// 		{Name: "image.tag", Value: "v2.1.0"},
			// 	},
			// }
			options := []types.InstallOption{optionMain}

			var apps = types.NewPackage(got, options, "ai-ollamaui", "install-id", "default", "ollama.cc", "ollama", "traefik")
			apps.Root.GetServiceLbPort()
			// got.GetServiceLbPort()
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("LoadPackage() error = %v, wantErr %v", err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("LoadPackage() = %v, want %v", got, tt.want)
			// }
			if got.Manifest.Application.Identifie != "longflow_ai" {
				t.Errorf("got.Manifest.Application.Identifie = %v, want %v", got.Manifest, tt.want)
			}
			if len(got.Children) == 0 {
				t.Errorf("got.Children = %v, want %v", len(got.Children), tt.want)
			}
		})
	}
}

func TestToPackageAddConfig(t *testing.T) {

	// uri := "https://zpk.w7.cc/respo/info/ai_ollamaui/ai_ollamaapi"
	uri := "https://zpk.w7.cc/zpk/respo/info/w7_minio"
	manifestPackage, err := LoadPackage(uri)
	if err != nil {
		t.Error(err)
	}
	config := manifestPackage.ToPackageAddConfig("", false)
	t.Log(config)
}

func TestGen(t *testing.T) {

	single := NewManifestSingleton()
	helmMemory := &HelmMemory{
		Identifie:   "helm-test",
		Title:       "helm-test1",
		Icon:        "helm-test2",
		Description: "helm-test3",
		ChartName:   "helm-test4",
		Repository:  "helm-test5",
		Version:     "helm-test6",
		Kv: []types.EnvKv{
			{Name: "helm-test7", Value: "helm-test8"},
		},
	}
	app := HelmManifestApp(helmMemory)
	single.Put(app.Application.Identifie, &app)

	newApp, ok := single.Get("helm-test")
	if !ok {
		t.Error("not found")
	}
	assert.Equal(t, "helm-test", newApp.Application.Identifie)
	assert.Equal(t, "helm-test1", newApp.Application.Name)

	uri := "memory://helm-test"
	u, err := url.Parse(uri)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "helm-test", u.Host)

}

func TestUrl(t *testing.T) {
	uri := "https://zpk.w7.cc/respo/info/ai_ollamaui/ai_ollamaapi"
	u, err := url.Parse(uri)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "ai_ollamaui", u.Path)
}
