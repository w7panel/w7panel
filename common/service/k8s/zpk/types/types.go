package types

type ShellInterface interface {
	GetType() string
	GetShell() string
	GetTitle() string
	GetImage() string
}

type HelmConfigInterface interface {
	GetRepository() string
	GetChartName() string
	GetVersion() string
}

type Shell struct {
	Shell     string `json:"shell"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	SearchJob string `json:"searchJob"` //搜索job的名称
	Image     string `json:"image"`     //执行shell job的镜像 空使用默认当前镜像
}

func (s *Shell) GetShell() string {
	return s.Shell
}
func (s *Shell) GetTitle() string {
	if s.Title == "" {
		if s.Type == "install" {
			return "安装脚本"
		}
		if s.Type == "upgrade" {
			return "更新脚本"
		}
	}
	return s.Title
}

func (s *Shell) GetDisployTitle() string {
	if s.Type == "install" {
		return "[应用安装时触发]" + s.GetTitle()
	} else if s.Type == "upgrade" {
		return "[应用更新时触发]" + s.GetTitle()
	} else if s.Type == "uninstall" {
		return "[应用卸载时触发]" + s.GetTitle()
	} else if s.Type == "requireinstall" {
		return "[应用被安装时触发]" + s.GetTitle()
	}
	return "[自定义触发]" + s.GetTitle()
}
func (s *Shell) GetType() string {
	return s.Type
}

func (s *Shell) GetImage() string {
	return s.Image
}

func GetShellByType(shells []Shell, shellType string) ShellInterface {
	for _, param := range shells {
		if param.Type == shellType {
			return &param
		}
	}
	return nil
}
