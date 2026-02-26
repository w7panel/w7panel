package procpath

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
)

const (
	ProcPath      = "/proc"
	HostProcPath  = "/host/proc"
	DefaultPid    = "1"
)

func GetBasePath() string {
	localMock := facade.Config.GetBool("app.local_mock")
	if localMock {
		return HostProcPath
	}
	return ProcPath
}

func GetRootPath(pid string) string {
	return filepath.Join(GetBasePath(), pid, "root")
}

func GetRootPathWithSubPid(pid, subPid string) string {
	basePath := GetRootPath(pid)
	if subPid != "" {
		basePath = filepath.Join(basePath, ProcPath, subPid, "root")
	}
	return basePath
}

func GetFilePath(pid, subPid, relativePath string) string {
	basePath := GetRootPathWithSubPid(pid, subPid)
	return filepath.Join(basePath, strings.TrimPrefix(relativePath, "/"))
}

func GetEtcPasswdPath(pid string) string {
	return filepath.Join(GetRootPath(pid), "etc", "passwd")
}

func GetEtcGroupPath(pid string) string {
	return filepath.Join(GetRootPath(pid), "etc", "group")
}

func IsProcessAlive(pid string) bool {
	_, err := os.Stat(filepath.Join(GetBasePath(), pid))
	return !os.IsNotExist(err)
}

func ConvertToLocalPath(path string) string {
	localMock := facade.Config.GetBool("app.local_mock")
	if !localMock {
		return path
	}
	if strings.HasPrefix(path, "/proc/") {
		return HostProcPath + strings.TrimPrefix(path, ProcPath)
	}
	return path
}
