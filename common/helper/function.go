package helper

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto"
	"crypto/md5"
	"crypto/sha1"
	"slices"

	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"

	// "golang.org/x/mod/zip"

	localJson "encoding/json"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/tools/cache"
)

var clusterD = "cluster.local"

const K3K_AGENT_PREFIX = "w7panel-k3k-agent"

func ChangeClusterDns(domain string) {
	clusterD = domain
}

func ClusterDomain(name, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.%s", name, namespace, clusterD)
}

func CreateDirIfNotExist(dirName string, perm os.FileMode) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.MkdirAll(dirName, perm)
		if err != nil {
			panic(err)
		}
	}
}

func GetCurUsrHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDir
}

func GetAppHomeDir() string {
	homeDir := GetCurUsrHomeDir()

	appDir := homeDir + "/w7_k8s"
	CreateDirIfNotExist(appDir, os.ModePerm)

	return appDir + "/"
}

// func CreateZipFromDir(source, target string) error {
// 	version := module.Version{
// 		Path:    "gitee.com/we7coreteam/k8s-offline",
// 		Version: "v0.1.2",
// 	}
// 	//判断target 是否存在

// 	file, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
// 	if err != nil {
// 		return err
// 	}

// 	return zip.CreateFromDir(file, version, source)
// }

// RandomString generates a random string of the specified length
func RandomString(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = 'a' + byte(rand.Intn(26)) // Simple example, generates lowercase letters
	}
	return string(bytes)
}

func RandomByte(length int) []byte {
	bytes := make([]byte, length)
	_, err := crand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return bytes
}

func LaravelAppKey(length int) string {

	bytes := RandomByte(32)
	return "base64:" + base64.StdEncoding.EncodeToString(bytes)

}

func K8sObjToYaml(obj runtime.Object) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	buf := bytes.NewBuffer([]byte{})
	if err := s.Encode(obj, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func StringToMD5(src string) string {
	m := md5.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

// 验证验证码是否正确

func MyIp() (string, error) {
	// if true {
	// 	return "118.25.145.25", nil
	// }
	res, err := http.Get("https://ifconfig.io")
	if err != nil {
		return "", err
	}
	body := res.Body
	defer body.Close()
	ip, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(ip), "\n", ""), nil
}

func RetryHttpClient() *resty.Client {
	useragent, ok := os.LookupEnv("USER_AGENT")
	client := resty.New()
	client.Debug = IsDebug()
	req := client.SetTimeout(time.Duration(10)*time.Second).SetRetryCount(0).SetHeader("Accept", "application/json")
	if ok {
		req.SetHeader("User-Agent", useragent).EnableTrace()
	}

	return req
}

func YamlParse(data []byte) (map[string]interface{}, error) {
	// yaml 解析
	var yamlData map[string]interface{}
	err := yaml.Unmarshal(data, &yamlData)
	if err != nil {
		return nil, err
	}
	return yamlData, nil
}

func RunCmdBinsh(args ...string) (string, error) {
	// Bug 修复：将 args 作为切片传递给 exec.Command
	cmd := exec.Command("/bin/sh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

// nsenter -t 1 --mount --uts --ipc --net --pid -- /bin/bash
func RunNcenterBinsh(shell string) (string, error) {
	// Bug 修复：将 args 作为切片传递给 exec.Command
	// defaultArgs := []string{} //[]string{"-t", "1", "--mount", "--uts", "--ipc", "--net", "--pid", "--"}
	defaultArgs := []string{"-t", "1", "--mount", "--uts", "--ipc", "--net", "--pid", "--", "sh", "-c"}
	args := append(defaultArgs, shell)
	slog.Info("nsenter args", "args", args)
	// cmd := exec.Command("nsenter", args...)
	cmd := exec.Command("nsenter", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		slog.Error("run cmd err" + stderr.String())
		return stderr.String(), err
	}

	return stdout.String(), nil
}

var fileMutex sync.Mutex

func WriterRegistries(data []byte) error {

	fileMutex.Lock()
	defer fileMutex.Unlock()
	return WriteFileAtomic("/host/etc/rancher/k3s/registries.yaml", data)
}

func ReadRegistries() ([]byte, error) {
	// 1. 判断文件是否存在 /host/etc/containers/registries2.yaml
	// 2. 如果不存在创建文件 并写入data
	// 3. 如果存在 则写入data
	filePath := "/host/etc/rancher/k3s/registries.yaml"
	// if _, err := os.Stat(filePath); os.IsNotExist(err) {
	// 	return nil, err
	// }
	return os.ReadFile(filePath)
}

func ReadK3sConfig() ([]byte, error) {
	filePath := "/host/etc/rancher/k3s/config.yaml"
	return os.ReadFile(filePath)
}

func WriteK3sConfig(config []byte) error {
	filePath := "/host/etc/rancher/k3s/config.yaml"
	// return os.WriteFile(filePath, config, 0644)
	return WriteFileAtomic(filePath, config)
}

func ReadK3sEnvFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func NvidiaReadyFileExites() bool {
	string, ok := os.LookupEnv("GPU_MOCK")
	if ok && string == "true" {
		return true
	}
	filePath := "/host/run/nvidia/validations/.driver-ctr-ready"
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func YamlToBytes(data map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(data)
}

func WriteFileAtomic(filePath string, data []byte) error {
	// 获取目标文件所在的目录
	dir := filepath.Dir(filePath)

	// 创建临时文件
	tmpFile, err := os.CreateTemp(dir, "registries-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // 确保临时文件最终被删除

	// 写入数据到临时文件
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write data to temp file: %v", err)
	}

	// 确保数据写入磁盘
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync data to disk: %v", err)
	}

	// 关闭临时文件
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// 重命名临时文件为目标文件
	if err := os.Rename(tmpFile.Name(), filePath); err != nil {
		return fmt.Errorf("failed to rename temp file to target file: %v", err)
	}

	return nil
}

func ValidateCertificate(certData []byte, host string) (bool, error) {
	// 解码 PEM 格式的证书
	block, _ := pem.Decode(certData)
	if block == nil || block.Type != "CERTIFICATE" {
		return false, fmt.Errorf("无效的 PEM 格式证书")
	}

	// 解析证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("解析证书失败: %v", err)
	}

	// 检查证书是否过期
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return false, fmt.Errorf("证书尚未生效")
	}
	if now.After(cert.NotAfter) {
		return false, fmt.Errorf("证书已过期")
	}

	// 验证证书链（可选）
	// 如果需要验证证书链，可以使用 x509.VerifyOptions
	// opts := x509.VerifyOptions{
	// 	CurrentTime: now,
	// }
	// if _, err := cert.Verify(opts); err != nil {
	// 	return false, fmt.Errorf("证书链验证失败: %v", err)
	// }
	if !IsDomainInCertificate(cert, host) {
		return false, nil
	}

	return true, nil
}

func IsDomainInCertificate(cert *x509.Certificate, domain string) bool {
	// 检查 Subject Alternative Name (SAN)
	for _, san := range cert.DNSNames {
		if strings.EqualFold(san, domain) {
			return true
		}
	}

	// 检查 Common Name (CN)
	if strings.EqualFold(cert.Subject.CommonName, domain) {
		return true
	}

	return false
}

func CreateDatabase(host, port, username, password, dbName string) error {
	// 构建连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port)

	// 连接 MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// 检查连接是否成功
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	slog.Info("Connected to MySQL successfully!")

	// 创建数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}

	slog.Info("Database '%s' created successfully!\n", "dbName", dbName)
	return nil
}

func Unzip(src, dest string, decodeGBk bool) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// 创建目标文件路径
		fname := (f.Name)
		if IsGBKCoding([]byte(fname)) {
			fname2, err := DecodeGBK((f.Name))
			if err == nil {
				// fmt.Println(err)
				// return err
				fname = fname2
			}

		}

		fpath := filepath.Join(dest, fname)

		// 检查目录是否存在
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// 打开zip文件中的文件
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		// print(string(cccc([]byte("你好世界"))))
		// print(fpath + "\n")
		// 创建目标文件
		out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer out.Close()

		// 复制文件内容
		_, err = io.Copy(out, rc)
		if err != nil {
			return err
		}
	}
	return nil
}

func DecodeGBK(s string) (string, error) {
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	b, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// https://blog.csdn.net/tianshi418/article/details/105194869
// 需要说明的是，isGBK()是通过双字节是否落在gbk的编码范围内实现的，
// 而utf-8编码格式的每个字节都是落在gbk的编码范围内，
// 所以只有先调用isUtf8()先判断不是utf-8编码，再调用isGBK()才有意义
func IsGBKCoding(data []byte) bool {
	if isUtf8(data) == true {
		return false
	} else if isGBK(data) == true {
		return true
	} else {
		return false
	}
}

func isGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0x7f {
			//编码0~127,只有一个字节的编码，兼容ASCII码
			i++
			continue
		} else {
			//大于127的使用双字节编码，落在gbk编码范围内的字符
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

//UTF-8编码格式的判断

func preNUm(data byte) int {
	var mask byte = 0x80
	var num int = 0
	// 8bit中首个0bit前有多少个1bits
	for i := 0; i < 8; i++ {
		if (data & mask) == mask {
			num++
			mask = mask >> 1
		} else {
			break
		}
	}
	return num
}
func isUtf8(data []byte) bool {
	i := 0
	for i < len(data) {
		if (data[i] & 0x80) == 0x00 {
			// 0XXX_XXXX
			i++
			continue
		} else if num := preNUm(data[i]); num > 2 {
			// 110X_XXXX 10XX_XXXX
			// 1110_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_0XXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_10XX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_110X 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// preNUm() 返回首个字节的8个bits中首个0bit前面1bit的个数，该数量也是该字符所使用的字节数
			i++
			for j := 0; j < num-1; j++ {
				//判断后面的 num - 1 个字节是不是都是10开头
				if (data[i] & 0xc0) != 0x80 {
					return false
				}
				i++
			}
		} else {
			//其他情况说明不是utf-8
			return false
		}
	}
	return true
}

// 关闭输出
func WaitForNamedCacheSync(controllerName string, stopCh <-chan struct{}, cacheSyncs ...cache.InformerSynced) bool {
	// klog.Infof("Waiting for caches to sync for %s", controllerName)

	if !cache.WaitForCacheSync(stopCh, cacheSyncs...) {
		// utilruntime.HandleError(fmt.Errorf("unable to sync caches for %s", controllerName))
		return false
	}

	// klog.Infof("Caches are synced for %s", controllerName)
	return true
}

// 镜像地址
func SelfImage() string {
	version, ok := os.LookupEnv("HELM_VERSION")
	if !ok {
		version = "1.0.107"
	}
	baseImage, ok1 := os.LookupEnv("IMAGE_REPO")
	if !ok1 {
		baseImage = "ccr.ccs.tencentyun.com/afan-public/w7panel"
	}
	return baseImage + ":" + version
}

func ExtractSingleFileFromTgz(url, fileName string) ([]byte, error) {
	// 发起 HTTP 请求获取远程 tgz 文件的响应
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("无法获取远程文件: %w", err)
	}
	defer resp.Body.Close()

	// 创建 gzip 读取器来解压响应体
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("无法解压 gzip 文件: %w", err)
	}
	defer gzr.Close()

	// 创建 tar 读取器来处理解压后的 tar 文件
	tr := tar.NewReader(gzr)

	// 遍历 tar 文件中的每个条目
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// 到达 tar 文件末尾，结束循环
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取 tar 文件时出错: %w", err)
		}

		// 检查当前条目是否为普通文件且文件名与目标文件路径匹配
		if hdr.Typeflag == tar.TypeReg && strings.Contains(hdr.Name, fileName) {
			// if hdr.Typeflag == tar.TypeReg && hdr.Name == targetFilePath {
			// 创建本地文件用于保存提取的内容
			// 读取 tar 文件条目的内容
			content, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("无法读取文件内容: %w", err)
			}
			return content, nil
		}
	}

	return nil, errors.New("未找到chart.yaml文件")
}

var envVarNameRegex = regexp.MustCompile(`^[-._a-zA-Z][-._a-zA-Z0-9]*$`)

// 验证环境变量名称是否有效
func IsValidEnvVarName(name string) bool {
	return envVarNameRegex.MatchString(name)
}

func GetNodeInnertIp(node *v1.Node) (string, error) {
	// if true {
	// 	return "218.23.2.55", nil
	// }
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address, nil
		}
	}
	return "", nil
}

func Runsh(name string, arg ...string) (string, string, error) {
	cmd := exec.Command(name, arg...)

	// 创建一个 bytes.Buffer 用于存储命令的输出
	var out bytes.Buffer
	var errOut bytes.Buffer

	// 设置命令的标准输出和标准错误输出
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	// 执行命令
	err := cmd.Run()
	if err != nil {
		// slog.Info("Command failed with error: %s\n", err)
		// print(errOut.String())
		// fmt.Print(errOut.String())
		// fmt.Printf("Command failed with error: %s\n", err)
		// fmt.Printf("Error output:\n%s\n", errOut.String())
		return "", errOut.String(), err
	}
	// fmt.Print(out.String())
	// slog.Info("Command failed with error: %s\n", out.String())
	// print(out.String())
	return out.String(), errOut.String(), nil
}

func IsLocalMock() bool {
	return os.Getenv("LOCAL_MOCK") == "true" || os.Getenv("LOCAL_MOCK") == "1"
}

func CleanStaticDir(releaseName string) error {
	dir := facade.Config.GetString("app.microapp_path") + "/" + releaseName
	return os.RemoveAll(dir)
}

func IsDebug() bool {
	return os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1"
}

func ParseX509(cert []byte) (*x509.Certificate, error) {
	// 尝试解码PEM格式的证书
	block, _ := pem.Decode(cert)
	if block == nil {
		// 如果不是PEM格式，直接尝试解析DER格式
		return x509.ParseCertificate(cert)
	}

	// 验证是否是证书类型
	if block.Type != "CERTIFICATE" {
		return nil, errors.New("invalid PEM block type: " + block.Type)
	}

	// 解析证书
	return x509.ParseCertificate(block.Bytes)
}

func EncryptWithPublicKey(pubKey *rsa.PublicKey, data []byte, base64Encoding bool) ([]byte, error) {
	// RSA加密有长度限制，对于长数据需要分段加密或使用混合加密
	// 这里使用OAEP填充方式更安全
	ciphertext, err := rsa.EncryptOAEP(
		sha1.New(),
		crand.Reader,
		pubKey,
		data,
		nil)
	// ciphertext, err := rsa.EncryptPKCS1v15(
	// 	crand.Reader,
	// 	pubKey,
	// 	data,
	// )
	if err != nil {
		return nil, fmt.Errorf("加密错误: %v", err)
	}
	if base64Encoding {
		return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
	}
	return ciphertext, nil
}

func VerifyDataWithPublicKey(pubKey *rsa.PublicKey, data, signature []byte) error {
	// 计算数据的哈希值
	hashed := sha256.Sum256(data)

	// 验证签名
	err := rsa.VerifyPKCS1v15(
		pubKey,
		crypto.SHA256,
		hashed[:],
		signature,
	)
	if err != nil {
		return fmt.Errorf("签名验证失败: %v", err)
	}

	return nil
}

func GetK3kAgentName(name string) string {
	return "w7panel-k3k-agent-" + name
}

func GetK3kServer0Name(name string) string {
	return "k3k-" + name + "-server-0"
}

func GetK3kServer0ContainerName(name string) string {
	return "k3k-" + name + "-server"
}

func GetApiServerHost(k3kNamespce string) string {
	if IsLocalMock() {
		// return "218.23.2.55:31150"
	}
	return k3kNamespce + "-service" + "." + k3kNamespce
}

func IsChildAgent() bool {
	return os.Getenv("IS_CHILD") == "true"
}

func SelfReqUrl() string {
	if IsLocalMock() {
		return "http://127.0.0.1:9007/"
	}
	return "http://" + os.Getenv("POD_IP") + ":8000/"
}

func ServiceAccountName() string {
	sa, ok := os.LookupEnv("SERVICE_ACCOUNT_NAME")
	if !ok {
		sa = "w7panel"
	}
	return sa
}

func ProxyUrl(proxyUrl string, path string, host string, headers map[string]string, query map[string]string) (*httputil.ReverseProxy, error) {
	// 记录函数调用的基本参数
	slog.Info("Creating reverse proxy", "proxyUrl", proxyUrl, "path", path)

	remote, err := url.Parse(proxyUrl)
	if err != nil {
		slog.Error("Failed to parse proxy URL", "error", err)
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	if IsDebug() {
		// 创建自定义Transport来记录请求和响应
		originalTransport := http.DefaultTransport
		proxy.Transport = &loggingTransport{
			transport: originalTransport,
		}
	}
	//curl http://your-go-proxy:8080 -H 'Host: w7panel-k3k-agent-console-98655.w7panel.xyz'
	proxy.Director = func(req *http.Request) {
		originalURL := *req.URL
		originalHost := req.Host

		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host //
		req.Host = remote.Host     //2个host有区别
		if headers != nil {
			for k, v := range headers {
				req.Header.Add(k, v)
			}
		}
		if host != "" {
			// req.URL.Host = host
			req.Host = host
		}

		if path != "" {
			// 检查path是否包含查询参数
			if strings.Contains(path, "?") {
				// 如果path包含查询参数，解析整个path
				pathURL, err := url.Parse(path)
				if err == nil {
					// 设置路径部分
					req.URL.Path = pathURL.Path
					// 设置查询参数，覆盖原始URL的查询参数
					req.URL.RawQuery = pathURL.RawQuery
				} else {
					// 如果解析失败，直接使用整个path作为路径
					req.URL.Path = path
				}
			} else {
				// 如果path不包含查询参数，只替换路径部分
				req.URL.Path = path
			}
		}
		if query != nil {
			if req.URL.RawQuery == "" {
				// req.URL.RawQuery = "?"
			}
			for k, v := range query {
				if req.URL.RawQuery == "" {
					req.URL.RawQuery = req.URL.RawQuery + k + "=" + v
				} else {
					req.URL.RawQuery = req.URL.RawQuery + "&" + k + "=" + v
				}

			}
		}

		if IsDebug() {

		}
		// 记录修改后的请求信息
		slog.Info("Proxying request",
			"method", req.Method,
			"originalURL", originalURL.String(),
			"originalHost", originalHost,
			"newURL", req.URL.String(),
			"newHost", req.Host,
			"headers", fmt.Sprintf("%v", req.Header),
		)
	}

	proxy.ModifyResponse = func(res *http.Response) error {
		// 记录响应信息
		// slog.Info("Received response",
		// 	"status", res.Status,
		// 	"statusCode", res.StatusCode,
		// 	"headers", fmt.Sprintf("%v", res.Header),
		// 	"contentLength", res.ContentLength,
		// )

		res.Header.Del("Access-Control-Allow-Origin")
		res.Header.Del("access-control-allow-headers")
		res.Header.Del("access-control-allow-methods")
		res.Header.Del("access-control-allow-credentials")
		res.Header.Del("access-control-expose-headers")
		return nil
	}

	slog.Info("Reverse proxy created successfully")
	return proxy, nil
}

// 自定义Transport结构体，用于记录请求和响应
type loggingTransport struct {
	transport http.RoundTripper
}

// 实现RoundTripper接口
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 记录发送请求前的信息
	slog.Info("Sending request to remote server",
		"method", req.Method,
		"url", req.URL.String(),
		"headers", fmt.Sprintf("%v", req.Header),
	)

	// 使用原始Transport发送请求
	resp, err := t.transport.RoundTrip(req)

	// 记录响应信息
	if err != nil {
		slog.Error("Error from remote server", "error", err)
	} else {
		slog.Info("Response from remote server",
			"status", resp.Status,
			"statusCode", resp.StatusCode,
			"headers", fmt.Sprintf("%v", resp.Header),
			"contentLength", resp.ContentLength,
		)
	}

	return resp, err
}

// 计算两个字符串切片的交集
func Intersection(a, b []string) []string {
	// 创建一个 map 来存储第一个切片的元素
	set := make(map[string]bool)
	for _, item := range a {
		set[item] = true
	}

	var result []string
	// 检查第二个切片的元素是否存在于 map 中
	for _, item := range b {
		if set[item] {
			result = append(result, item)
			// 避免重复，如果切片中有重复元素
			set[item] = false
		}
	}
	return result
}

// 计算两个字符串切片的差集
func Difference(a, b []string) []string {
	set := make(map[string]bool)
	for _, item := range b {
		set[item] = true
	}

	var diff []string
	for _, item := range a {
		if !set[item] {
			diff = append(diff, item)
		}
	}
	return diff
}

func ImageDigest(ref string, opt ...crane.Option) (string, error) {

	return crane.Digest(ref, opt...)
}

func IsK3kShared() bool {
	mode, ok := os.LookupEnv("K3K_MODE")
	if ok && mode == "shared" {
		return true
	}
	return false
}

func IsK3kVirtual() bool {
	mode, ok := os.LookupEnv("K3K_MODE")
	if ok && mode == "virtual" {
		return true
	}
	return false
}

func ParseResourceLimit(str string) resource.Quantity {
	quantity, err := resource.ParseQuantity(str)
	if err != nil {
		return resource.Quantity{}
	}
	return quantity
}

func ParseFloat64(str string) float64 {
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return val
}

func StringToInt64(str string) int64 {
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func StringToFloat64(str string) float64 {
	val, err := strconv.ParseFloat(str, 10)
	if err != nil {
		return 0
	}
	return val
}

func FloatStringToInt64(str string) int64 {
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return int64(val)
}

func IpCity(ipaddr string) (string, error) {
	koPath, ok := os.LookupEnv("KO_DATA_PATH")
	if !ok {
		return "", errors.New("KO_DATA_PATH not set")
	}

	dbPath := koPath + "/ip2region_v4.xdb" // 或者你的 ipv4 xdb 的路径
	version := xdb.IPv4

	searcher, err := xdb.NewWithFileOnly(version, dbPath)
	if err != nil {
		slog.Error("Failed to create searcher", "error", err)
		// fmt.Printf("failed to create searcher: %s\n", err.Error())
		return "", err
	}

	defer searcher.Close()

	region, err := searcher.SearchByStr(ipaddr)
	if err != nil {
		slog.Error("Failed to search IP address", "error", err)
		// fmt.Printf("failed to search ip: %s\n", err.Error())
		return "", err
	}
	if region != "" {
		regionSplit := strings.Split(region, "|")
		if len(regionSplit) > 2 {
			rstr := regionSplit[1] + "/" + regionSplit[2]
			rstr = strings.ReplaceAll(rstr, "省", "")
			rstr = strings.ReplaceAll(rstr, "市", "")
			return rstr, nil
		}
	}
	return region, nil

}

func MyCity() (string, error) {
	ip, err := MyIp()
	if err != nil {
		slog.Error("Failed to get IP address", "error", err)
		return "", err
	}
	return IpCity(ip)
}

func SafeConcatName(maxLength int, name ...string) string {
	name = slices.DeleteFunc(name, func(s string) bool {
		return s == ""
	})

	fullPath := strings.Join(name, "-")
	if len(fullPath) < maxLength {
		return fullPath
	}

	digest := sha256.Sum256([]byte(fullPath))

	// since we cut the string in the middle, the last char may not be compatible with what is expected in k8s
	// we are checking and if necessary removing the last char
	c := fullPath[maxLength-8]
	if 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return fullPath[0:maxLength-7] + "-" + hex.EncodeToString(digest[0:])[0:5]
	}

	return fullPath[0:maxLength-8] + "-" + hex.EncodeToString(digest[0:])[0:6]
}

func IsLxcfsEnabled() bool {
	return os.Getenv("LXCFS_ENABLED") == "true"
}

func After2SecondRun(f func()) {
	time.AfterFunc(time.Second*2, f)
}

// func CheckLogo() error {
// 	sdk := k8s.NewK8sClient()
// 	configMap, err := sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "logo.config", metav1.GetOptions{})
// 	if err != nil {
// 		slog.Error("Failed to get logo config", "error", err)
// 		return err
// 	}
// 	return WriteLogo(configMap)
// }

func WriteLogo(configMap *v1.ConfigMap) error {
	if configMap.Namespace == "kube-system" && configMap.Name == "k3k.logo.config" {
		kodata, ok := os.LookupEnv("KO_DATA_PATH")
		if ok {
			logoData, ok := configMap.BinaryData["logo"]
			if ok {
				// save configMap
				err := os.WriteFile(kodata+"/assets/logo.png", logoData, 0644)
				if err != nil {
					slog.Error("Failed to write logo", "error", err)
					return err
				}
			} // save configMap
		}
	}
	return nil
}

func GetToken(ctx *gin.Context) string {
	apiToken := ctx.Request.URL.Query().Get("api-token")
	if apiToken != "" {
		return apiToken
	}
	auth := ctx.Request.Header.Get("AuthorizationX")
	if auth == "" || !strings.Contains(auth, " ") {
		auth = ctx.Request.Header.Get("Authorization")
		if auth == "" || !strings.Contains(auth, " ") {
			return ""
		}
	}
	bearertoken := strings.Split(auth, " ")[1]
	return bearertoken
}
func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func ToJsonNoErr(b interface{}) string {
	json, err := localJson.Marshal(b)
	if err != nil {
		return ""
	}
	return string(json)
}

func ToJson(b interface{}) (string, error) {
	json, err := localJson.Marshal(b)
	if err != nil {
		return "", err
	}
	return string(json), nil
}
