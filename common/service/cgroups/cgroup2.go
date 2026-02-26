package cgroups

import (
	// manager "github.com/opencontainers/cgroups/manager"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	cgroup2 "github.com/containerd/cgroups/v3/cgroup2"
	"github.com/containerd/cgroups/v3/cgroup2/stats"
)

var currentcgroup string

func init() {
	// manager.New("cpu", "cpu.cfs_quota_us")
	// manager.New
	path, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		slog.Error("canot open cgroup file")
	}
	if err == nil {
		spath := string(path)
		spath = strings.ReplaceAll(spath, "\n", "")
		if strings.HasPrefix(spath, "0::/") {
			currentcgroup = spath[3:]
			// 获取currentcgroup dir
			currentcgroup = filepath.Dir(currentcgroup)
			// slog.Info("current cgroup is ", "cgrouppath", currentcgroup)
			// currentcgroup = currentcgroup[:strings.Index(currentcgroup, "/")]
		}
	}
}

func Load(cgroupRoot string) (*cgroup2.Manager, error) {
	return cgroup2.Load(cgroupRoot)
}

func Current() (*cgroup2.Manager, error) {
	return Load(currentcgroup)
}

func CurrentStat() (*stats.Metrics, error) {
	cgroup, err := Load(currentcgroup)
	if err != nil {
		return nil, err
	}
	return cgroup.Stat()
}
