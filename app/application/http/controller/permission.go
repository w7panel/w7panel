package controller

import (
	"os"
	"path/filepath"
	"strconv"

	"gitee.com/we7coreteam/k8s-offline/common/service/procpath"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type PermissionAgent struct {
	controller.Abstract
}

func (c PermissionAgent) Chmod(http *gin.Context) {
	pid := http.Param("pid")
	subpid := http.Param("subpid")

	type Params struct {
		Path      string `form:"path" json:"path" binding:"required"`
		Mode      string `form:"mode" json:"mode" binding:"required"`
		Recursive bool   `form:"recursive" json:"recursive"`
	}

	params := Params{}
	if err := http.ShouldBind(&params); err != nil {
		c.JsonResponseWithError(http, err, 400)
		return
	}

	targetPid := pid
	if subpid != "" {
		targetPid = subpid
	}

	mode, err := strconv.ParseUint(params.Mode, 8, 32)
	if err != nil {
		c.JsonResponseWithError(http, err, 400)
		return
	}

	fullPath := procpath.GetFilePath(targetPid, "", params.Path)

	if params.Recursive {
		err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return os.Chmod(path, os.FileMode(mode))
		})
		if err != nil {
			c.JsonResponseWithError(http, err, 500)
			return
		}
	} else {
		if err := os.Chmod(fullPath, os.FileMode(mode)); err != nil {
			c.JsonResponseWithError(http, err, 500)
			return
		}
	}

	c.JsonSuccessResponse(http)
}

func (c PermissionAgent) Chown(http *gin.Context) {
	pid := http.Param("pid")
	subpid := http.Param("subpid")

	type Params struct {
		Path      string `form:"path" json:"path" binding:"required"`
		Owner     string `form:"owner" json:"owner" binding:"required"`
		Recursive bool   `form:"recursive" json:"recursive"`
	}

	params := Params{}
	if err := http.ShouldBind(&params); err != nil {
		c.JsonResponseWithError(http, err, 400)
		return
	}

	targetPid := pid
	if subpid != "" {
		targetPid = subpid
	}

	fullPath := procpath.GetFilePath(targetPid, "", params.Path)

	owner := params.Owner
	var uid, gid int

	if owner != "" {
		parts := owner
		_, err := os.Stat(fullPath)
		if err != nil {
			c.JsonResponseWithError(http, err, 500)
			return
		}

		parsed := parseOwner(parts)
		if parsed.uid != -1 {
			uid = parsed.uid
		}
		if parsed.gid != -1 {
			gid = parsed.gid
		}
	}

	if params.Recursive {
		err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return os.Chown(path, uid, gid)
		})
		if err != nil {
			c.JsonResponseWithError(http, err, 500)
			return
		}
	} else {
		if err := os.Chown(fullPath, uid, gid); err != nil {
			c.JsonResponseWithError(http, err, 500)
			return
		}
	}

	c.JsonSuccessResponse(http)
}

type ownerInfo struct {
	uid int
	gid int
}

func parseOwner(owner string) ownerInfo {
	info := ownerInfo{-1, -1}
	
	if owner == "" {
		return info
	}

	if n, err := strconv.Atoi(owner); err == nil {
		info.uid = n
		info.gid = n
		return info
	}

	if parts := parseUserGroup(owner); parts != nil {
		if parts[0] != "" {
			info.uid = getUidByName(parts[0])
		}
		if len(parts) > 1 && parts[1] != "" {
			info.gid = getGidByName(parts[1])
		}
	}

	return info
}

func parseUserGroup(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' || s[i] == '.' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s, ""}
}

func getUidByName(name string) int {
	if name == "root" {
		return 0
	}
	return -1
}

func getGidByName(name string) int {
	if name == "root" {
		return 0
	}
	return -1
}
