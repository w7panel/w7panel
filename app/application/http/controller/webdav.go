package controller

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/common/service/procpath"
	webdav1 "github.com/w7panel/w7panel/common/service/webdav"
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
		"/panel-api/v1/files/webdav",
		webdav.Dir("/tmp/webdav"), "1", "/tmp/webdav")
}

func (c Webdav) Test(ctx *gin.Context) {
	pid := ctx.Param("pid")
	subpid := ctx.Param("subpid")
	webDirPath := procpath.GetRootPathWithSubPid(pid, subpid)
	prefix := "/panel-api/v1/files/webdav-agent/" + pid + "/agent"
	if subpid != "" {
		prefix = "/panel-api/v1/files/webdav-agent/" + pid + "/subagent/" + subpid + "/agent"
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

// WebDAV 分片上传相关结构体
type WebdavChunkUploadRequest struct {
	ChunkIndex   int    `json:"chunkIndex"`
	ChunkTotal   int    `json:"chunkTotal"`
	ChunkSize    int64  `json:"chunkSize"`
	FileSize     int64  `json:"fileSize"`
	Identifier   string `json:"identifier"`
	RelativePath string `json:"relativePath"`
	FileName     string `json:"fileName"`
	Pid          string `json:"pid"`
	SubPid       string `json:"subpid"`
}

type WebdavChunkMergeRequest struct {
	Identifier   string `json:"identifier"`
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
	TotalChunks  int    `json:"totalChunks"`
	FileSize     int64  `json:"fileSize"`
	Pid          string `json:"pid"`
	SubPid       string `json:"subpid"`
}

type WebdavChunkCheckResponse struct {
	ChunkExists bool `json:"chunkExists"`
}

type WebdavUploadResponse struct {
	FileURL  string `json:"fileUrl"`
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
}

// handleChunkUpload 处理 WebDAV 分片上传
func (c Webdav) handleChunkUpload(ctx *gin.Context, pid string, subpid string, rootDir string) {
	type Params struct {
		ChunkIndex   int    `form:"chunkIndex" binding:"required"`
		ChunkTotal   int    `form:"chunkTotal" binding:"required"`
		Identifier   string `form:"identifier" binding:"required"`
		FileName     string `form:"fileName" binding:"required"`
		RelativePath string `form:"relativePath"`
		FileSize     int64  `form:"fileSize"`
	}

	var params Params
	if err := ctx.ShouldBind(&params); err != nil {
		c.jsonError(ctx, fmt.Sprintf("invalid parameters: %v", err), 400)
		return
	}

	// 获取文件
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		c.jsonError(ctx, fmt.Sprintf("failed to get file: %v", err), 400)
		return
	}
	defer file.Close()

	// 创建临时分片目录
	chunkDir := filepath.Join(rootDir, ".webdav_chunks")
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		c.jsonError(ctx, fmt.Sprintf("failed to create chunk directory: %v", err), 500)
		return
	}

	// 使用 identifier + fileName 的 MD5 作为分片目录名
	fileMD5 := md5.Sum([]byte(params.Identifier + params.FileName))
	fileMD5Str := hex.EncodeToString(fileMD5[:])
	userChunkDir := filepath.Join(chunkDir, fileMD5Str)

	if err := os.MkdirAll(userChunkDir, 0755); err != nil {
		c.jsonError(ctx, fmt.Sprintf("failed to create user chunk directory: %v", err), 500)
		return
	}

	// 分片文件路径
	chunkFilename := fmt.Sprintf("%d_%d", params.ChunkIndex, params.ChunkTotal)
	chunkFilePath := filepath.Join(userChunkDir, chunkFilename)

	// 检查分片是否已上传
	if _, err := os.Stat(chunkFilePath); err == nil {
		ctx.JSON(200, gin.H{
			"chunkExists": true,
			"chunkIndex":  params.ChunkIndex,
		})
		return
	}

	// 创建临时文件
	tmpFile, err := os.Create(chunkFilePath + ".tmp")
	if err != nil {
		c.jsonError(ctx, fmt.Sprintf("failed to create temp file: %v", err), 500)
		return
	}
	defer tmpFile.Close()

	// 复制文件内容
	written, err := io.Copy(tmpFile, file)
	if err != nil {
		os.Remove(tmpFile.Name())
		c.jsonError(ctx, fmt.Sprintf("failed to write chunk: %v", err), 500)
		return
	}
	tmpFile.Close()

	// 重命名为正式文件
	if err := os.Rename(tmpFile.Name(), chunkFilePath); err != nil {
		os.Remove(tmpFile.Name())
		c.jsonError(ctx, fmt.Sprintf("failed to save chunk: %v", err), 500)
		return
	}

	slog.Info("webdav chunk uploaded successfully",
		"identifier", params.Identifier,
		"chunkIndex", params.ChunkIndex,
		"chunkTotal", params.ChunkTotal,
		"written", written,
		"fileName", params.FileName,
		"pid", pid)

	ctx.JSON(200, gin.H{
		"chunkExists": false,
		"chunkIndex":  params.ChunkIndex,
		"chunkTotal":  params.ChunkTotal,
		"written":     written,
	})
}

// handleCheckChunk 检查 WebDAV 分片是否已上传
func (c Webdav) handleCheckChunk(ctx *gin.Context, pid string, subpid string, rootDir string) {
	type Params struct {
		Identifier   string `form:"identifier" binding:"required"`
		ChunkIndex   int    `form:"chunkIndex" binding:"required"`
		ChunkTotal   int    `form:"chunkTotal" binding:"required"`
		FileName     string `form:"fileName"`
		RelativePath string `form:"relativePath"`
	}

	var params Params
	if err := ctx.ShouldBind(&params); err != nil {
		c.jsonError(ctx, fmt.Sprintf("invalid parameters: %v", err), 400)
		return
	}
	// os.TempDir()
	chunkDir := filepath.Join(rootDir, ".webdav_chunks")

	// 计算文件 MD5
	fileMD5 := md5.Sum([]byte(params.Identifier + params.FileName))
	fileMD5Str := hex.EncodeToString(fileMD5[:])
	userChunkDir := filepath.Join(chunkDir, fileMD5Str)

	chunkFilename := fmt.Sprintf("%d_%d", params.ChunkIndex, params.ChunkTotal)
	chunkFilePath := filepath.Join(userChunkDir, chunkFilename)

	// 检查分片是否存在
	_, err := os.Stat(chunkFilePath)
	chunkExists := (err == nil)

	ctx.JSON(200, WebdavChunkCheckResponse{
		ChunkExists: chunkExists,
	})
}

// handleMergeChunks 合并 WebDAV 分片
func (c Webdav) handleMergeChunks(ctx *gin.Context, pid string, subpid string, rootDir string) {
	type Params struct {
		Identifier   string `json:"identifier" binding:"required"`
		FileName     string `json:"fileName" binding:"required"`
		RelativePath string `json:"relativePath"`
		TotalChunks  int    `json:"totalChunks" binding:"required"`
		FileSize     int64  `json:"fileSize"`
	}

	var params Params
	if err := ctx.ShouldBindJSON(&params); err != nil {
		c.jsonError(ctx, fmt.Sprintf("invalid parameters: %v", err), 400)
		return
	}

	chunkDir := filepath.Join(rootDir, ".webdav_chunks")

	// 计算文件 MD5
	fileMD5 := md5.Sum([]byte(params.Identifier + params.FileName))
	fileMD5Str := hex.EncodeToString(fileMD5[:])
	userChunkDir := filepath.Join(chunkDir, fileMD5Str)

	// 获取锁，避免并发合并
	lock := getChunkLock(fileMD5Str)
	lock.Lock()
	defer lock.Unlock()

	// 检查分片目录是否存在
	if _, err := os.Stat(userChunkDir); os.IsNotExist(err) {
		c.jsonError(ctx, "chunk directory not found", 400)
		return
	}

	// 收集所有分片文件
	var chunkFiles []string
	for i := 0; i < params.TotalChunks; i++ {
		chunkFilename := fmt.Sprintf("%d_%d", i, params.TotalChunks)
		chunkFilePath := filepath.Join(userChunkDir, chunkFilename)
		if _, err := os.Stat(chunkFilePath); os.IsNotExist(err) {
			c.jsonError(ctx, fmt.Sprintf("missing chunk %d", i), 400)
			return
		}
		chunkFiles = append(chunkFiles, chunkFilePath)
	}

	// 排序分片文件
	sort.Strings(chunkFiles)

	// 确定最终文件路径
	finalFileName := params.FileName
	if params.RelativePath != "" {
		finalFileName = filepath.Join(params.RelativePath, params.FileName)
	}

	// 确保路径在 rootDir 内
	finalFilePath := filepath.Join(rootDir, finalFileName)

	// 安全检查：确保最终路径在 rootDir 内
	if !strings.HasPrefix(finalFilePath, rootDir) {
		c.jsonError(ctx, "invalid file path", 400)
		return
	}

	// 创建目标文件目录
	if err := os.MkdirAll(filepath.Dir(finalFilePath), 0755); err != nil {
		c.jsonError(ctx, fmt.Sprintf("failed to create directory: %v", err), 500)
		return
	}

	// 创建目标文件
	destFile, err := os.Create(finalFilePath)
	if err != nil {
		c.jsonError(ctx, fmt.Sprintf("failed to create destination file: %v", err), 500)
		return
	}
	defer destFile.Close()

	// 合并分片
	var totalWritten int64
	for _, chunkFile := range chunkFiles {
		srcFile, err := os.Open(chunkFile)
		if err != nil {
			c.jsonError(ctx, fmt.Sprintf("failed to open chunk: %v", err), 500)
			return
		}

		written, err := io.Copy(destFile, srcFile)
		srcFile.Close()
		if err != nil {
			c.jsonError(ctx, fmt.Sprintf("failed to merge chunk: %v", err), 500)
			return
		}
		totalWritten += written
	}

	// 清理分片目录
	if err := os.RemoveAll(userChunkDir); err != nil {
		slog.Warn("failed to clean up chunk directory", "dir", userChunkDir, "err", err)
	}

	// 移除锁
	removeChunkLock(fileMD5Str)

	slog.Info("webdav chunks merged successfully",
		"identifier", params.Identifier,
		"fileName", params.FileName,
		"totalChunks", params.TotalChunks,
		"totalWritten", totalWritten,
		"pid", pid)

	ctx.JSON(200, WebdavUploadResponse{
		FileURL:  "/panel-api/v1/files/webdav" + strings.TrimPrefix(finalFilePath, rootDir),
		FileName: params.FileName,
		FileSize: totalWritten,
	})
}

// HandlePidChunkUpload 处理 PID 分片上传
func (c Webdav) HandlePidChunkUpload(ctx *gin.Context) {
	pid := ctx.Param("pid")
	webDirPath := procpath.GetRootPath(pid)
	c.handleChunkUpload(ctx, pid, "", webDirPath)
}

// HandlePidChunkCheck 检查 PID 分片
func (c Webdav) HandlePidChunkCheck(ctx *gin.Context) {
	pid := ctx.Param("pid")
	webDirPath := procpath.GetRootPath(pid)
	c.handleCheckChunk(ctx, pid, "", webDirPath)
}

// HandlePidChunkMerge 合并 PID 分片
func (c Webdav) HandlePidChunkMerge(ctx *gin.Context) {
	pid := ctx.Param("pid")
	webDirPath := procpath.GetRootPath(pid)
	c.handleMergeChunks(ctx, pid, "", webDirPath)
}

// HandlePidSubPidChunkUpload 处理子 PID 分片上传
func (c Webdav) HandlePidSubPidChunkUpload(ctx *gin.Context) {
	pid := ctx.Param("pid")
	subpid := ctx.Param("subpid")
	webDirPath := procpath.GetRootPathWithSubPid(pid, subpid)
	c.handleChunkUpload(ctx, pid, subpid, webDirPath)
}

// HandlePidSubPidChunkCheck 检查子 PID 分片
func (c Webdav) HandlePidSubPidChunkCheck(ctx *gin.Context) {
	pid := ctx.Param("pid")
	subpid := ctx.Param("subpid")
	webDirPath := procpath.GetRootPathWithSubPid(pid, subpid)
	c.handleCheckChunk(ctx, pid, subpid, webDirPath)
}

// HandlePidSubPidChunkMerge 合并子 PID 分片
func (c Webdav) HandlePidSubPidChunkMerge(ctx *gin.Context) {
	pid := ctx.Param("pid")
	subpid := ctx.Param("subpid")
	webDirPath := procpath.GetRootPathWithSubPid(pid, subpid)
	c.handleMergeChunks(ctx, pid, subpid, webDirPath)
}
