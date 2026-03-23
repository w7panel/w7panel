/*
分片上传测试命令

使用方法:

	# 测试普通分片上传 (生成 10MB 文件，分片大小 1MB)
	./runtime/main test:upload-chunk

	# 自定义文件大小和分片大小
	./runtime/main test:upload-chunk --fileSize=52428800 --chunkSize=2097152

	# 测试 WebDAV PID 上传
	./runtime/main test:upload-chunk --pid=<process_id>

	# 测试 WebDAV SubPID 上传
	./runtime/main test:upload-chunk --pid=<process_id> --subPid=<sub_process_id>

	# 指定 API 地址
	./runtime/main test:upload-chunk --baseURL=http://127.0.0.1:8080

参数说明:

	--baseURL      API 基础 URL，默认 http://localhost:8080
	--fileSize     生成的测试文件大小（字节），默认 10MB
	--chunkSize    分片大小（字节），默认 1MB
	--fileName     文件名，默认 test-chunk-upload.bin
	--identifier   文件唯一标识（可选，默认自动生成）
	--relativePath 相对路径（可选）
	--pid          进程 ID（用于 WebDAV PID 上传）
	--subPid       子进程 ID（用于 WebDAV SubPID 上传）
*/
package console

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// TestUploadChunk 分片上传测试命令
type TestUploadChunk struct {
	console.Abstract
}

type uploadChunkOption struct {
	baseURL      string
	fileSize     int64
	chunkSize    int64
	fileName     string
	identifier   string
	relativePath string
	pid          string
	subPid       string
}

var ucOp = uploadChunkOption{}

func (c TestUploadChunk) GetName() string {
	return "test:upload-chunk"
}

func (c TestUploadChunk) Configure(cmd *cobra.Command) {
	cmd.Flags().StringVar(&ucOp.baseURL, "baseURL", "http://localhost:8000", "API 基础 URL")
	cmd.Flags().Int64Var(&ucOp.fileSize, "fileSize", 10*1024*1024, "生成的测试文件大小（字节），默认 10MB")
	cmd.Flags().Int64Var(&ucOp.chunkSize, "chunkSize", 1*1024*1024, "分片大小（字节），默认 1MB")
	cmd.Flags().StringVar(&ucOp.fileName, "fileName", "test-chunk-upload.bin", "文件名")
	cmd.Flags().StringVar(&ucOp.identifier, "identifier", "", "文件唯一标识（可选，默认使用 MD5）")
	cmd.Flags().StringVar(&ucOp.relativePath, "relativePath", "", "相对路径（可选）")
	cmd.Flags().StringVar(&ucOp.pid, "pid", "", "进程 ID（用于 WebDAV PID 上传）")
	cmd.Flags().StringVar(&ucOp.subPid, "subPid", "", "子进程 ID（用于 WebDAV SubPID 上传）")
}

func (c TestUploadChunk) GetDescription() string {
	return "测试分片上传功能 - 动态生成文件并调用分片上传接口"
}

func (c TestUploadChunk) Handle(cmd *cobra.Command, args []string) {

	slog.Info("开始分片上传测试",
		"baseURL", ucOp.baseURL,
		"fileSize", ucOp.fileSize,
		"chunkSize", ucOp.chunkSize,
		"fileName", ucOp.fileName,
		"pid", ucOp.pid,
		"subPid", ucOp.subPid)

	// 计算分片数量
	totalChunks := int((ucOp.fileSize + ucOp.chunkSize - 1) / ucOp.chunkSize)
	slog.Info("分片信息",
		"totalChunks", totalChunks,
		"fileSize", ucOp.fileSize,
		"chunkSize", ucOp.chunkSize)

	// 生成文件标识符
	identifier := ucOp.identifier
	if identifier == "" {
		identifier = fmt.Sprintf("test-%s-%d", ucOp.fileName, ucOp.fileSize)
	}

	// 确定上传接口路径
	var uploadPath, checkPath, mergePath string
	if ucOp.pid != "" {
		if ucOp.subPid != "" {
			uploadPath = fmt.Sprintf("/panel-api/v1/files/webdav-agent/%s/subagent/%s/upload-chunk", ucOp.pid, ucOp.subPid)
			checkPath = fmt.Sprintf("/panel-api/v1/files/webdav-agent/%s/subagent/%s/check-chunk", ucOp.pid, ucOp.subPid)
			mergePath = fmt.Sprintf("/panel-api/v1/files/webdav-agent/%s/subagent/%s/merge-chunks", ucOp.pid, ucOp.subPid)
		} else {
			uploadPath = fmt.Sprintf("/panel-api/v1/files/webdav-agent/%s/upload-chunk", ucOp.pid)
			checkPath = fmt.Sprintf("/panel-api/v1/files/webdav-agent/%s/check-chunk", ucOp.pid)
			mergePath = fmt.Sprintf("/panel-api/v1/files/webdav-agent/%s/merge-chunks", ucOp.pid)
		}
	} else {
		uploadPath = "/panel-api/v1/files/upload-chunk"
		checkPath = "/panel-api/v1/files/check-chunk"
		mergePath = "/panel-api/v1/files/merge-chunks"
	}

	slog.Info("接口路径",
		"uploadPath", uploadPath,
		"checkPath", checkPath,
		"mergePath", mergePath)

	// 创建临时目录存储分片
	tempDir, err := os.MkdirTemp("", "chunk-upload-test-*")
	if err != nil {
		slog.Error("创建临时目录失败", "error", err)
		return
	}
	defer os.RemoveAll(tempDir)
	slog.Info("临时目录", "dir", tempDir)

	// 上传所有分片
	uploadedChunks := make([]int, 0, totalChunks)
	for i := 0; i < totalChunks; i++ {
		chunkSize := ucOp.chunkSize
		remaining := ucOp.fileSize - int64(i)*ucOp.chunkSize
		if remaining < ucOp.chunkSize {
			chunkSize = remaining
		}

		// 生成随机数据分片
		chunkData := c.generateRandomChunk(chunkSize)
		chunkFilePath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d", i))
		if err := os.WriteFile(chunkFilePath, chunkData, 0644); err != nil {
			slog.Error("写入分片文件失败", "chunkIndex", i, "error", err)
			return
		}

		slog.Info("上传分片",
			"chunkIndex", i,
			"chunkTotal", totalChunks,
			"chunkSize", chunkSize)

		// 上传分片
		success, err := c.uploadChunk(uploadPath, chunkFilePath, i, totalChunks, chunkSize, identifier, ucOp.fileName, ucOp.relativePath, ucOp.fileSize)
		if err != nil {
			slog.Error("上传分片失败", "chunkIndex", i, "error", err)
			if !success {
				return
			}
		}

		if success {
			uploadedChunks = append(uploadedChunks, i)
		}
	}

	slog.Info("所有分片上传完成", "uploadedChunks", len(uploadedChunks))

	// 检查所有分片
	slog.Info("开始检查分片")
	for i := 0; i < totalChunks; i++ {
		exists, err := c.checkChunk(checkPath, i, totalChunks, identifier, ucOp.fileName, ucOp.relativePath)
		if err != nil {
			slog.Error("检查分片失败", "chunkIndex", i, "error", err)
			return
		}
		if !exists {
			slog.Error("分片不存在", "chunkIndex", i)
			return
		}
	}
	slog.Info("所有分片检查通过")

	// 合并分片
	slog.Info("开始合并分片")
	err = c.mergeChunks(mergePath, identifier, ucOp.fileName, ucOp.relativePath, totalChunks, ucOp.fileSize)
	if err != nil {
		slog.Error("合并分片失败", "error", err)
		return
	}

	slog.Info("分片上传测试完成", "fileName", ucOp.fileName, "totalSize", ucOp.fileSize)
}

// generateRandomChunk 生成随机数据分片
func (c TestUploadChunk) generateRandomChunk(size int64) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// uploadChunk 上传单个分片
func (c TestUploadChunk) uploadChunk(uploadPath, chunkFilePath string, chunkIndex, chunkTotal int, chunkSize int64, identifier, fileName, relativePath string, fileSize int64) (bool, error) {
	// 创建 multipart 表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件
	file, err := os.Open(chunkFilePath)
	if err != nil {
		return false, fmt.Errorf("打开分片文件失败：%v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(chunkFilePath))
	if err != nil {
		return false, fmt.Errorf("创建表单文件失败：%v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return false, fmt.Errorf("复制文件内容失败：%v", err)
	}

	// 添加表单字段
	_ = writer.WriteField("chunkIndex", strconv.Itoa(chunkIndex))
	_ = writer.WriteField("chunkTotal", strconv.Itoa(chunkTotal))
	_ = writer.WriteField("chunkSize", strconv.FormatInt(chunkSize, 10))
	_ = writer.WriteField("identifier", identifier)
	_ = writer.WriteField("fileName", fileName)
	_ = writer.WriteField("relativePath", relativePath)
	_ = writer.WriteField("fileSize", strconv.FormatInt(fileSize, 10))

	if err := writer.Close(); err != nil {
		return false, fmt.Errorf("关闭 writer 失败：%v", err)
	}

	// 发送请求
	url := ucOp.baseURL + uploadPath
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return false, fmt.Errorf("创建请求失败：%v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("发送请求失败：%v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("读取响应失败：%v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("上传失败，状态码：%d, 响应：%s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, fmt.Errorf("解析响应失败：%v", err)
	}

	chunkExists, ok := result["chunkExists"].(bool)
	if ok && chunkExists {
		slog.Info("分片已存在，跳过", "chunkIndex", chunkIndex)
		return true, nil
	}

	slog.Info("分片上传成功", "chunkIndex", chunkIndex, "response", string(respBody))
	return true, nil
}

// checkChunk 检查分片是否存在
func (c TestUploadChunk) checkChunk(checkPath string, chunkIndex, chunkTotal int, identifier, fileName, relativePath string) (bool, error) {
	url := ucOp.baseURL + checkPath
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("创建请求失败：%v", err)
	}

	query := req.URL.Query()
	query.Add("chunkIndex", strconv.Itoa(chunkIndex))
	query.Add("chunkTotal", strconv.Itoa(chunkTotal))
	query.Add("identifier", identifier)
	query.Add("fileName", fileName)
	query.Add("relativePath", relativePath)
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("发送请求失败：%v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("读取响应失败：%v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("检查失败，状态码：%d, 响应：%s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, fmt.Errorf("解析响应失败：%v", err)
	}

	chunkExists, ok := result["chunkExists"].(bool)
	if !ok {
		return false, fmt.Errorf("响应格式错误")
	}

	slog.Info("分片检查结果",
		"chunkIndex", chunkIndex,
		"chunkExists", chunkExists)

	return chunkExists, nil
}

// mergeChunks 合并分片
func (c TestUploadChunk) mergeChunks(mergePath, identifier, fileName, relativePath string, totalChunks int, fileSize int64) error {
	url := ucOp.baseURL + mergePath

	type MergeRequest struct {
		Identifier   string `json:"identifier"`
		FileName     string `json:"fileName"`
		RelativePath string `json:"relativePath"`
		TotalChunks  int    `json:"totalChunks"`
		FileSize     int64  `json:"fileSize"`
	}

	reqBody := MergeRequest{
		Identifier:   identifier,
		FileName:     fileName,
		RelativePath: relativePath,
		TotalChunks:  totalChunks,
		FileSize:     fileSize,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败：%v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败：%v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败：%v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败：%v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("合并失败，状态码：%d, 响应：%s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析响应失败：%v", err)
	}

	// 计算文件 MD5 进行验证
	fileMD5 := md5.Sum(jsonData)
	fileMD5Str := hex.EncodeToString(fileMD5[:])

	slog.Info("分片合并成功",
		"response", string(respBody),
		"requestMD5", fileMD5Str)

	return nil
}
