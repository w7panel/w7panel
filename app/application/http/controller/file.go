package controller

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/common/service/procpath"
	"github.com/w7panel/w7panel/common/service/s3"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type File struct {
	controller.Abstract
}

// ChunkUploadRequest 分片上传请求
type ChunkUploadRequest struct {
	ChunkIndex   int    `json:"chunkIndex"`   // 分片索引（从 0 开始）
	ChunkTotal   int    `json:"chunkTotal"`   // 总分片数
	ChunkSize    int64  `json:"chunkSize"`    // 分片大小
	FileSize     int64  `json:"fileSize"`     // 文件总大小
	Identifier   string `json:"identifier"`   // 文件唯一标识（通常使用文件 MD5）
	RelativePath string `json:"relativePath"` // 相对路径
	FileName     string `json:"fileName"`     // 文件名
}

// ChunkMergeRequest 合并分片请求
type ChunkMergeRequest struct {
	Identifier   string `json:"identifier"`   // 文件唯一标识
	FileName     string `json:"fileName"`     // 文件名
	RelativePath string `json:"relativePath"` // 相对路径
	TotalChunks  int    `json:"totalChunks"`  // 总分片数
	FileSize     int64  `json:"fileSize"`     // 文件总大小
}

// ChunkCheckResponse 分片检查响应
type ChunkCheckResponse struct {
	ChunkExists bool `json:"chunkExists"` // 分片是否存在
}

// UploadResponse 上传响应
type UploadResponse struct {
	FileURL  string `json:"fileUrl"`  // 文件访问 URL
	FileName string `json:"fileName"` // 文件名
	FileSize int64  `json:"fileSize"` // 文件大小
}

// chunkLockMap 分片上传锁（避免同一文件并发合并）
var chunkLockMap = make(map[string]*sync.Mutex)
var chunkLockMutex sync.Mutex

// getChunkLock 获取文件锁
func getChunkLock(identifier string) *sync.Mutex {
	chunkLockMutex.Lock()
	defer chunkLockMutex.Unlock()
	if _, ok := chunkLockMap[identifier]; !ok {
		chunkLockMap[identifier] = &sync.Mutex{}
	}
	return chunkLockMap[identifier]
}

// removeChunkLock 移除文件锁
func removeChunkLock(identifier string) {
	chunkLockMutex.Lock()
	defer chunkLockMutex.Unlock()
	delete(chunkLockMap, identifier)
}

func (self File) Download(http *gin.Context) {
	r := http.Request
	filename := filepath.Join(facade.GetConfig().GetString("s3.base_dir"),
		strings.TrimPrefix(r.URL.Path, "/panel-api/v1/download/"),
	)
	fs, err := os.Stat(filename)
	if os.IsNotExist(err) {
		self.JsonResponseWithError(http, fmt.Errorf("file not found"), 404)
		return
	}
	if fs.IsDir() {
	}
	file, err := os.Open(filename)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	defer file.Close()

	http.Header("Content-Type", "application/octet-stream")
	http.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name()))
	http.File(file.Name())
}

func (self File) Upload(http *gin.Context) {
	server := s3.NewS3Server(facade.Config.GetString("s3.base_dir"), "/tmp/metadata", "s3bucket")
	server.Server().ServeHTTP(http.Writer, http.Request)
}

// UploadChunk 分片上传
func (self File) UploadChunk(http *gin.Context) {
	baseDir := os.TempDir()
	chunkDir := filepath.Join(baseDir, ".chunks") // 临时分片存储目录

	// 解析表单
	file, _, err := http.Request.FormFile("file")
	if err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("failed to get file: %v", err), 400)
		return
	}
	defer file.Close()

	// 获取分片信息
	chunkIndexStr := http.PostForm("chunkIndex")
	chunkTotalStr := http.PostForm("chunkTotal")
	identifier := http.PostForm("identifier")
	fileName := http.PostForm("fileName")
	_ = http.PostForm("relativePath") // relativePath 用于后续扩展
	fileSizeStr := http.PostForm("fileSize")

	if identifier == "" || chunkIndexStr == "" || chunkTotalStr == "" {
		self.JsonResponseWithError(http, fmt.Errorf("missing required parameters"), 400)
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("invalid chunkIndex: %v", err), 400)
		return
	}

	chunkTotal, err := strconv.Atoi(chunkTotalStr)
	if err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("invalid chunkTotal: %v", err), 400)
		return
	}

	fileSize := int64(0)
	if fileSizeStr != "" {
		fileSize, _ = strconv.ParseInt(fileSizeStr, 10, 64)
		_ = fileSize // 用于后续验证
	}

	// 计算文件 MD5 作为分片目录
	// fileMD5 := md5.Sum([]byte(identifier + fileName))
	// fileMD5Str := hex.EncodeToString(fileMD5[:])
	userChunkDir := filepath.Join(chunkDir, identifier)

	// 创建分片目录
	if err := os.MkdirAll(userChunkDir, 0755); err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("failed to create chunk directory: %v", err), 500)
		return
	}

	// 分片文件路径
	chunkFilename := fmt.Sprintf("%d_%d", chunkIndex, chunkTotal)
	chunkFilePath := filepath.Join(userChunkDir, chunkFilename)

	// 检查分片是否已上传
	if _, err := os.Stat(chunkFilePath); err == nil {
		slog.Info("chunk already exists", "identifier", identifier, "chunkIndex", chunkIndex)
		self.JsonResponse(http, gin.H{
			"chunkExists": true,
			"chunkIndex":  chunkIndex,
		}, nil, 200)
		return
	}

	// 创建临时文件
	tmpFile, err := os.Create(chunkFilePath + ".tmp")
	if err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("failed to create temp file: %v", err), 500)
		return
	}
	defer tmpFile.Close()

	// 复制文件内容
	written, err := io.Copy(tmpFile, file)
	if err != nil {
		os.Remove(tmpFile.Name())
		self.JsonResponseWithError(http, fmt.Errorf("failed to write chunk: %v", err), 500)
		return
	}
	tmpFile.Close()

	// 重命名为正式文件
	if err := os.Rename(tmpFile.Name(), chunkFilePath); err != nil {
		os.Remove(tmpFile.Name())
		self.JsonResponseWithError(http, fmt.Errorf("failed to save chunk: %v", err), 500)
		return
	}

	slog.Info("chunk uploaded successfully",
		"identifier", identifier,
		"chunkIndex", chunkIndex,
		"chunkTotal", chunkTotal,
		"written", written,
		"fileName", fileName)

	self.JsonResponse(http, gin.H{
		"chunkExists": false,
		"chunkIndex":  chunkIndex,
		"chunkTotal":  chunkTotal,
		"written":     written,
	}, nil, 200)
}

// CheckChunk 检查分片是否已上传
func (self File) CheckChunk(http *gin.Context) {
	baseDir := os.TempDir()
	chunkDir := filepath.Join(baseDir, ".chunks")

	type ParamsValidate struct {
		Identifier   string `form:"identifier" binding:"required"`
		ChunkIndex   string `form:"chunkIndex" binding:"required"`
		ChunkTotal   string `form:"chunkTotal" binding:"required"`
		FileName     string `form:"fileName"`
		RelativePath string `form:"relativePath"`
	}

	params := ParamsValidate{}
	if err := http.ShouldBindQuery(&params); err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("invalid parameters: %v", err), 400)
		return
	}

	chunkIndex, err := strconv.Atoi(params.ChunkIndex)
	if err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("invalid chunkIndex: %v", err), 400)
		return
	}

	chunkTotal, err := strconv.Atoi(params.ChunkTotal)
	if err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("invalid chunkTotal: %v", err), 400)
		return
	}

	// 计算文件 MD5
	// fileMD5 := md5.Sum([]byte(params.Identifier + params.FileName))
	// fileMD5Str := hex.EncodeToString(fileMD5[:])
	userChunkDir := filepath.Join(chunkDir, params.Identifier)

	chunkFilename := fmt.Sprintf("%d_%d", chunkIndex, chunkTotal)
	chunkFilePath := filepath.Join(userChunkDir, chunkFilename)

	// 检查分片是否存在
	_, err = os.Stat(chunkFilePath)
	chunkExists := (err == nil)

	self.JsonResponse(http, ChunkCheckResponse{
		ChunkExists: chunkExists,
	}, nil, 200)
}

// MergeChunks 合并分片
func (self File) MergeChunks(http *gin.Context) {
	baseDir := facade.GetConfig().GetString("s3.base_dir")
	chunkDir := filepath.Join(baseDir, ".chunks")

	type ParamsValidate struct {
		Identifier   string `json:"identifier" binding:"required"`
		FileName     string `json:"fileName" binding:"required"`
		RelativePath string `json:"relativePath"`
		TotalChunks  int    `json:"totalChunks" binding:"required"`
		FileSize     int64  `json:"fileSize"`
	}

	params := ParamsValidate{}
	if err := http.ShouldBindJSON(&params); err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("invalid parameters: %v", err), 400)
		return
	}

	// 计算文件 MD5
	// fileMD5 := md5.Sum([]byte(params.Identifier + params.FileName))
	// fileMD5Str := hex.EncodeToString(fileMD5[:])
	userChunkDir := filepath.Join(chunkDir, params.Identifier)

	// 获取锁，避免并发合并
	lock := getChunkLock(params.Identifier)
	lock.Lock()
	defer lock.Unlock()

	// 检查分片目录是否存在
	if _, err := os.Stat(userChunkDir); os.IsNotExist(err) {
		self.JsonResponseWithError(http, fmt.Errorf("chunk directory not found"), 400)
		return
	}

	// 收集所有分片文件
	var chunkFiles []string
	for i := 0; i < params.TotalChunks; i++ {
		chunkFilename := fmt.Sprintf("%d_%d", i, params.TotalChunks)
		chunkFilePath := filepath.Join(userChunkDir, chunkFilename)
		if _, err := os.Stat(chunkFilePath); os.IsNotExist(err) {
			self.JsonResponseWithError(http, fmt.Errorf("missing chunk %d", i), 400)
			return
		}
		chunkFiles = append(chunkFiles, chunkFilePath)
	}

	// 排序分片文件（确保按正确顺序合并）
	sort.Strings(chunkFiles)

	// 确定最终文件路径
	finalFileName := params.FileName
	if params.RelativePath != "" {
		finalFileName = filepath.Join(params.RelativePath, params.FileName)
	}
	finalFilePath := filepath.Join(baseDir, finalFileName)

	// 创建目标文件目录
	if err := os.MkdirAll(filepath.Dir(finalFilePath), 0755); err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("failed to create directory: %v", err), 500)
		return
	}

	// 创建目标文件
	destFile, err := os.Create(finalFilePath)
	if err != nil {
		self.JsonResponseWithError(http, fmt.Errorf("failed to create destination file: %v", err), 500)
		return
	}
	defer destFile.Close()

	// 合并分片
	var totalWritten int64
	for _, chunkFile := range chunkFiles {
		srcFile, err := os.Open(chunkFile)
		if err != nil {
			self.JsonResponseWithError(http, fmt.Errorf("failed to open chunk %s: %v", chunkFile, err), 500)
			return
		}

		written, err := io.Copy(destFile, srcFile)
		srcFile.Close()
		if err != nil {
			self.JsonResponseWithError(http, fmt.Errorf("failed to merge chunk %s: %v", chunkFile, err), 500)
			return
		}
		totalWritten += written
	}

	// 清理分片目录
	if err := os.RemoveAll(userChunkDir); err != nil {
		slog.Warn("failed to clean up chunk directory", "dir", userChunkDir, "err", err)
	}

	// 移除锁
	removeChunkLock(params.Identifier)

	slog.Info("chunks merged successfully",
		"identifier", params.Identifier,
		"fileName", params.FileName,
		"totalChunks", params.TotalChunks,
		"totalWritten", totalWritten)

	self.JsonResponse(http, UploadResponse{
		FileURL:  "/panel-api/v1/download/" + strings.TrimPrefix(finalFilePath, baseDir+"/"),
		FileName: params.FileName,
		FileSize: totalWritten,
	}, nil, 200)
}

// CpPidFile 复制 Pod 文件
func (self File) CpPidFile(http *gin.Context) {
	baseDir := facade.Config.GetString("s3.base_dir")
	type ParamsValidate struct {
		From   string `form:"from"      binding:"required"`
		To     string `form:"to"        binding:"required"`
		Upload string `form:"upload"    binding:"required"`
		Pid    string `form:"pid"       binding:"required"`
	}

	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	if params.Upload == "1" {
		params.From = filepath.Join(baseDir, params.From)
		params.To = procpath.ConvertToLocalPath(params.To)
	} else {
		params.To = filepath.Join(baseDir, params.To)
		params.From = procpath.ConvertToLocalPath(params.From)
	}
	err := os.Mkdir(filepath.Dir(params.To), 0755)
	if err != nil && !os.IsExist(err) {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	slog.Info("cp", "from", params.From, "to", params.To)
	if err = exec.Command("cp", "-r", params.From, params.To).Run(); err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.JsonSuccessResponse(http)
}
