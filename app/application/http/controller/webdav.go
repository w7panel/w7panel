package controller

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"gitee.com/we7coreteam/k8s-offline/common/service/procpath"
	webdav1 "gitee.com/we7coreteam/k8s-offline/common/service/webdav"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	"golang.org/x/net/webdav"
)

type crossDeviceFileSystem struct {
	webdav.FileSystem
	dir string
}

func (fs *crossDeviceFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	err := fs.FileSystem.Rename(ctx, oldName, newName)
	if err != nil {
		errStr := err.Error()
		if len(errStr) >= 25 && errStr[len(errStr)-25:] == "invalid cross-device link" {
			slog.Warn("webdav Rename failed, trying cp+rm fallback", "oldName", oldName, "newName", newName, "error", err)

			srcPath := fs.dir + oldName
			dstPath := fs.dir + newName

			if err := exec.Command("cp", "-a", srcPath, dstPath).Run(); err != nil {
				slog.Error("cp failed", "src", srcPath, "dst", dstPath, "error", err)
				return err
			}

			if err := exec.Command("rm", "-rf", srcPath).Run(); err != nil {
				slog.Error("rm failed", "src", srcPath, "error", err)
				return err
			}

			slog.Info("webdav Rename cp+rm fallback success", "oldName", oldName, "newName", newName)
			return nil
		}
		slog.Error("webdav Rename failed", "oldName", oldName, "newName", newName, "error", err)
	}
	return err
}

func (fs *crossDeviceFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	info, err := fs.FileSystem.Stat(ctx, name)
	if err != nil {
		slog.Debug("webdav Stat error", "name", name, "error", err)
	}
	return info, err
}

func (fs *crossDeviceFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	err := fs.FileSystem.Mkdir(ctx, name, perm)
	if err != nil {
		slog.Error("webdav Mkdir failed", "name", name, "perm", perm, "error", err)
	}
	return err
}

func (fs *crossDeviceFileSystem) RemoveAll(ctx context.Context, name string) error {
	err := fs.FileSystem.RemoveAll(ctx, name)
	if err != nil {
		slog.Error("webdav RemoveAll failed", "name", name, "error", err)
	}
	return err
}

func (fs *crossDeviceFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	file, err := fs.FileSystem.OpenFile(ctx, name, flag, perm)
	if err != nil {
		slog.Error("webdav OpenFile failed", "name", name, "flag", flag, "perm", perm, "error", err)
	}
	return file, err
}

type Webdav struct {
	controller.Abstract
}

func (c Webdav) handleWithPermissionPreservation(ctx *gin.Context, prefix string, fs webdav.FileSystem, pid string, rootDir string) {
	userGroup := webdav1.GetUserGroup(pid)

	relPath := ctx.Request.URL.Path[len(prefix):]
	if relPath == "" {
		relPath = "/"
	}

	// 特殊目录使用高效的 PROPFIND 实现，返回标准 XML
	if ctx.Request.Method == "PROPFIND" && isSpecialDir(relPath) {
		fullPath := filepath.Join(rootDir, relPath)
		c.listSpecialDirectoryXML(ctx, fullPath, prefix, relPath)
		return
	}

	if ctx.Request.Method == "PROPFIND" || ctx.Request.Method == "GET" {
		permFs := &webdav1.PermissionFileSystem{
			FileSystem: fs,
			UserGroup:  userGroup,
			RootDir:    rootDir,
		}
		handler := webdav.Handler{
			Prefix:     prefix,
			FileSystem: permFs,
			LockSystem: webdav.NewMemLS(),
		}
		handler.ServeHTTP(ctx.Writer, ctx.Request)
		return
	} else if ctx.Request.Method == "PUT" {
		path := ctx.Request.URL.Path[len(prefix):]
		var fullPath string
		if dir, ok := fs.(webdav.Dir); ok {
			fullPath = string(dir) + path
		} else {
			handler := webdav.Handler{
				Prefix:     prefix,
				FileSystem: fs,
				LockSystem: webdav.NewMemLS(),
			}
			handler.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		var originalMode os.FileMode
		var originalUid, originalGid int
		if stat, err := os.Stat(fullPath); err == nil {
			originalMode = stat.Mode()
			if sysStat, ok := stat.Sys().(*syscall.Stat_t); ok {
				originalUid = int(sysStat.Uid)
				originalGid = int(sysStat.Gid)
			}
		}

		handler := webdav.Handler{
			Prefix:     prefix,
			FileSystem: fs,
			LockSystem: webdav.NewMemLS(),
		}
		handler.ServeHTTP(ctx.Writer, ctx.Request)

		if originalMode != 0 {
			os.Chmod(fullPath, originalMode)
			if originalUid != 0 && originalGid != 0 {
				os.Chown(fullPath, originalUid, originalGid)
			}
		}
	} else if ctx.Request.Method == "MOVE" || ctx.Request.Method == "COPY" {
		dirStr := ""
		if dir, ok := fs.(webdav.Dir); ok {
			dirStr = string(dir)
		}

		handler := webdav.Handler{
			Prefix:     prefix,
			FileSystem: &crossDeviceFileSystem{FileSystem: fs, dir: dirStr},
			LockSystem: webdav.NewMemLS(),
		}
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	} else {
		handler := webdav.Handler{
			Prefix:     prefix,
			FileSystem: fs,
			LockSystem: webdav.NewMemLS(),
		}
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func isSpecialDir(path string) bool {
	cleanPath := filepath.Clean(path)
	specialDirs := []string{"/proc", "/sys", "/dev", "/run"}
	for _, dir := range specialDirs {
		if cleanPath == dir || strings.HasPrefix(cleanPath, dir+"/") {
			return true
		}
	}
	return false
}

func (c Webdav) listSpecialDirectoryXML(ctx *gin.Context, fullPath string, prefix string, relPath string) {
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		slog.Warn("failed to list special directory", "path", fullPath, "error", err)
		ctx.Data(207, "application/xml; charset=utf-8", []byte(`<?xml version="1.0" encoding="UTF-8"?><D:multistatus xmlns:D="DAV:"></D:multistatus>`))
		return
	}

	maxEntries := webdav1.GetMaxDirEntries()
	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	var xmlBuilder strings.Builder
	xmlBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?><D:multistatus xmlns:D="DAV:">`)

	// 添加目录本身
	dirHref := prefix + relPath
	if !strings.HasSuffix(dirHref, "/") {
		dirHref += "/"
	}
	xmlBuilder.WriteString(fmt.Sprintf(`<D:response><D:href>%s</D:href><D:propstat><D:prop><D:displayname>%s</D:displayname><D:resourcetype><D:collection/></D:resourcetype></D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat></D:response>`,
		escapeXML(dirHref), escapeXML(filepath.Base(relPath))))

	for _, entry := range entries {
		name := entry.Name()
		href := dirHref + escapeXML(name)
		if entry.IsDir() {
			href += "/"
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		xmlBuilder.WriteString(`<D:response>`)
		xmlBuilder.WriteString(fmt.Sprintf(`<D:href>%s</D:href>`, href))
		xmlBuilder.WriteString(`<D:propstat><D:prop>`)
		xmlBuilder.WriteString(fmt.Sprintf(`<D:displayname>%s</D:displayname>`, escapeXML(name)))
		xmlBuilder.WriteString(fmt.Sprintf(`<D:getlastmodified>%s</D:getlastmodified>`, info.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")))
		xmlBuilder.WriteString(fmt.Sprintf(`<D:getcontentlength>%d</D:getcontentlength>`, info.Size()))

		if entry.IsDir() {
			xmlBuilder.WriteString(`<D:resourcetype><D:collection/></D:resourcetype>`)
		} else {
			xmlBuilder.WriteString(`<D:resourcetype/>`)
		}

		xmlBuilder.WriteString(`</D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat></D:response>`)
	}

	xmlBuilder.WriteString(`</D:multistatus>`)

	ctx.Data(207, "application/xml; charset=utf-8", []byte(xmlBuilder.String()))
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func (c Webdav) HandlePid(ctx *gin.Context) {
	pid := ctx.Param("pid")
	webDirPath := procpath.GetRootPath(pid)
	c.handleWithPermissionPreservation(ctx,
		"/panel-api/v1/files/webdav-agent/"+pid+"/agent",
		webdav.Dir(webDirPath), pid, webDirPath)
}

func (c Webdav) HandlePidSubPid(ctx *gin.Context) {
	pid := ctx.Param("pid")
	subpid := ctx.Param("subpid")
	webDirPath := procpath.GetRootPathWithSubPid(pid, subpid)
	prefix := "/panel-api/v1/files/webdav-agent/" + pid + "/agent"
	if subpid != "" {
		prefix = "/panel-api/v1/files/webdav-agent/" + pid + "/subagent/" + subpid + "/agent"
	}
	c.handleWithPermissionPreservation(ctx,
		prefix,
		webdav.Dir(webDirPath), pid, webDirPath)
}

func (c Webdav) Handle(ctx *gin.Context) {
	c.handleWithPermissionPreservation(ctx,
		"/k8s/webdav",
		webdav.Dir("/tmp/webdav"), "1", "/tmp/webdav")
}

func (c Webdav) Test(ctx *gin.Context) {
	pid := ctx.Param("pid")
	subpid := ctx.Param("subpid")
	webDirPath := procpath.GetRootPathWithSubPid(pid, subpid)
	prefix := "/k8s/webdav-agent/" + pid + "/agent"
	if subpid != "" {
		prefix = "/k8s/webdav-agent/" + pid + "/subagent/" + subpid + "/agent"
	}
	c.JsonResponseWithoutError(ctx, gin.H{
		"prefix":     prefix,
		"webDirPath": webDirPath,
	})
}

func (c Webdav) jsonError(ctx *gin.Context, message string, code int) {
	ctx.JSON(code, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
