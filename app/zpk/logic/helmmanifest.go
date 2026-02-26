package logic

import (
	"sync"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	zpktypes "gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
)

type manifestsingleton struct {
	manifestes map[string]*zpktypes.Manifest
	mu         sync.Mutex
}

// instance 是一个包级别的变量，用于保存唯一的单例对象
var instance *manifestsingleton

// once 是一个用于确保初始化操作只执行一次的 sync.Once 对象
var once sync.Once

// GetInstance 方法返回唯一的单例对象
func NewManifestSingleton() *manifestsingleton {
	// 使用 sync.Once 确保初始化操作只执行一次
	once.Do(func() {
		instance = &manifestsingleton{}
		instance.manifestes = make(map[string]*zpktypes.Manifest)
	})
	return instance
}

func (s *manifestsingleton) Put(key string, manifest *zpktypes.Manifest) {
	if len(s.manifestes) > 100 {
		s.manifestes = make(map[string]*zpktypes.Manifest)
	}
	s.manifestes[key] = manifest
}

func (s *manifestsingleton) Get(key string) (*zpktypes.Manifest, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	manifest, ok := s.manifestes[key]
	return manifest, ok
}

type HelmMemory struct {
	Identifie   string        `form:"identifie"`
	Title       string        `form:"title" `
	Icon        string        `form:"icon" `
	Description string        `form:"description" `
	ChartName   string        `form:"chartName" binding:"required"`
	Repository  string        `form:"repository"`
	Version     string        `form:"version" `
	Kv          []types.EnvKv `form:"kv" `
}

func HelmManifestApp(param *HelmMemory) zpktypes.Manifest {
	if param.Icon == "" {
		param.Icon = "https://cdn.w7.cc/images/2022/05/23/tr7mqnuRykrCohkQ5jjNC5AE0sjhlgIObfuaLaHS.png"
	}
	if param.Description == "" {
		param.Description = "helm应用"
	}
	if param.Title == "" {
		param.Title = "helm应用"
	}

	params := []zpktypes.StartParams{}
	for _, v := range param.Kv {
		startParamsA := zpktypes.StartParams{
			Type:        "text",
			Title:       v.Name,
			Name:        v.Name,
			Description: "helm参数",
			Required:    true,
			ValuesText:  v.Value,
		}
		params = append(params, startParamsA)
	}
	manifest := zpktypes.Manifest{
		Application: zpktypes.Application{
			Author:      "helm",
			Description: param.Description,
			Identifie:   param.Identifie,
			Name:        param.Title,
			Type:        "helm",
			Icon:        param.Icon,
		},
		Platform: zpktypes.Platform{
			Container: zpktypes.Container{
				Build: zpktypes.Build{},
				Cmd:   []string{},
				Ports: []zpktypes.Ports{{
					Name: "默认", Port: 80,
				}},
				InitialDelaySeconds: 2,
				MaxNum:              10,
				MinNum:              1,
				PolicyThreshold:     80,
				PolicyType:          "cpu",
				StartParams:         params,
			},
			Helm: zpktypes.Helm{
				ChartName:  param.ChartName,
				Repository: param.Repository,
				Version:    param.Version,
			},
			// Supports: []string{"notapp"},
		},
	}
	return manifest

}
