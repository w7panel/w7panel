package console

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/cgroups"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// username password register
type MetricsCgroup struct {
	console2.Abstract
}

type metricsGroupOption struct {
	Path string
}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443
var cro = metricsGroupOption{}

func (c MetricsCgroup) GetName() string {
	return "metrics:cgroup"
}

func (c MetricsCgroup) Configure(cmd *cobra.Command) {
	// username password register
	//
	cmd.Flags().StringVar(&cro.Path, "path", "", "path")

}

func (c MetricsCgroup) GetDescription() string {
	return "cgroup metrics cgroup"
}

func (c MetricsCgroup) Handle(cmd *cobra.Command, args []string) {

	currentPath := cro.Path
	if cro.Path == "self" {
		cPath, err := procToCgroup("/proc/self/cgroup")
		if err != nil {
			slog.Error("read /proc/self/cgroup error", "error", err)
			return
		}
		currentPath = cPath
	}
	_, err := strconv.ParseInt(cro.Path, 10, 64)
	if err == nil {
		cPath, err := procToCgroup("/proc/" + cro.Path + "/cgroup")
		if err != nil {
			slog.Error("read /proc/"+cPath+"/cgroup error", "error", err)
			return
		}
		currentPath = cPath

	}
	if currentPath == "" {
		slog.Error("path is empty")
		return
	}
	slog.Info("cgroupPath", "path", currentPath)
	stat(currentPath)
}

func stat(path string) {
	manager, err := cgroups.Load(path)
	if err != nil {
		slog.Error("load cgroup error", "error", err)
		return
	}
	stat, err := manager.Stat()
	if err != nil {
		slog.Error("stat cgroup error", "error", err)
		return
	}
	slog.Info("cgroupStat", "stat", stat)
}
func procToCgroup(procfile string) (string, error) {
	currentPath := ""
	ct, err := os.ReadFile(procfile)
	if err != nil {
		slog.Error("read /proc/self/cgroup error", "error", err)
		return "", err
	}
	spath := string(ct)
	spath = strings.ReplaceAll(spath, "\n", "")
	if strings.HasPrefix(spath, "0::/") {
		currentPath = spath[3:]
		return currentPath, nil
	}
	return "", errors.New("not found cgroup path")
}
