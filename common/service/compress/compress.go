package compress

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/procpath"
	"github.com/bodgit/sevenzip"
	"github.com/klauspost/pgzip"
	"github.com/ulikunitz/xz"
)

type Compressor struct {
	rootPath string
}

func NewCompressor(pid string) *Compressor {
	return &Compressor{
		rootPath: procpath.GetRootPath(pid),
	}
}

// Compress 压缩文件/目录
func (c *Compressor) Compress(sources []string, output string) error {
	outputPath := filepath.Join(c.rootPath, output)
	slog.Info("Compressing files", "sources", sources, "output", outputPath)

	// 根据扩展名确定格式
	format := detectFormat(output)

	// 创建输出目录
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	switch format {
	case "zip":
		return c.compressZip(sources, outputPath)
	case "tar":
		return c.compressTar(sources, outputPath, false, "")
	case "tar.gz", "tgz":
		return c.compressTar(sources, outputPath, true, "gzip")
	case "tar.bz2", "tbz2":
		return c.compressTar(sources, outputPath, true, "bzip2")
	case "tar.xz", "txz":
		return c.compressTar(sources, outputPath, true, "xz")
	default:
		// 默认使用 zip
		return c.compressZip(sources, outputPath)
	}
}

// Extract 解压文件
func (c *Compressor) Extract(source, target string) error {
	sourcePath := filepath.Join(c.rootPath, source)
	targetPath := filepath.Join(c.rootPath, target)
	slog.Info("Extracting file", "source", sourcePath, "target", targetPath)

	// 检测压缩格式
	format := detectFormat(source)

	// 创建目标目录
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	switch format {
	case "zip":
		return c.extractZip(sourcePath, targetPath)
	case "tar":
		return c.extractTar(sourcePath, targetPath)
	case "tar.gz", "tgz":
		return c.extractTarGz(sourcePath, targetPath)
	case "tar.bz2", "tbz2":
		return c.extractTarBz2(sourcePath, targetPath)
	case "tar.xz", "txz":
		return c.extractTarXz(sourcePath, targetPath)
	case "7z":
		return c.extract7z(sourcePath, targetPath)
	default:
		// 尝试自动检测
		return c.extractAuto(sourcePath, targetPath)
	}
}

// detectFormat 检测压缩格式
func detectFormat(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz2"):
		return "tar.bz2"
	case strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".txz"):
		return "tar.xz"
	case strings.HasSuffix(lower, ".tar"):
		return "tar"
	case strings.HasSuffix(lower, ".7z"):
		return "7z"
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	case strings.HasSuffix(lower, ".rar"):
		return "rar"
	default:
		return "zip"
	}
}

// compressZip 压缩为 ZIP 格式
func (c *Compressor) compressZip(sources []string, output string) error {
	zipFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, src := range sources {
		srcPath := filepath.Join(c.rootPath, src)
		if err := c.addToZip(zipWriter, srcPath, c.rootPath); err != nil {
			return err
		}
	}

	return nil
}

// addToZip 添加文件到 ZIP
func (c *Compressor) addToZip(zipWriter *zip.Writer, filePath, rootPath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			subPath := filepath.Join(filePath, entry.Name())
			if err := c.addToZip(zipWriter, subPath, rootPath); err != nil {
				return err
			}
		}
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	}

	return nil
}

// compressTar 压缩为 TAR 格式（可选压缩）
func (c *Compressor) compressTar(sources []string, output string, compress bool, compressType string) error {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	var writer io.Writer = file

	if compress {
		switch compressType {
		case "gzip":
			gzWriter := pgzip.NewWriter(file)
			defer gzWriter.Close()
			writer = gzWriter
		case "bzip2":
			// bzip2 只有读取器，需要外部实现
			return fmt.Errorf("bzip2 compression not supported, use gzip or xz")
		case "xz":
			xzWriter, err := xz.NewWriter(file)
			if err != nil {
				return err
			}
			defer xzWriter.Close()
			writer = xzWriter
		}
	}

	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	for _, src := range sources {
		srcPath := filepath.Join(c.rootPath, src)
		if err := c.addToTar(tarWriter, srcPath, c.rootPath); err != nil {
			return err
		}
	}

	return nil
}

// addToTar 添加文件到 TAR
func (c *Compressor) addToTar(tarWriter *tar.Writer, filePath, rootPath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(relPath)

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	if !info.IsDir() {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(tarWriter, file); err != nil {
			return err
		}
	} else {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			subPath := filepath.Join(filePath, entry.Name())
			if err := c.addToTar(tarWriter, subPath, rootPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// extractZip 解压 ZIP 文件
func (c *Compressor) extractZip(source, target string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)

		// 安全检查：防止目录遍历
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(target)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// extractTar 解压 TAR 文件
func (c *Compressor) extractTar(source, target string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	return c.extractTarReader(file, target)
}

// extractTarGz 解压 TAR.GZ 文件
func (c *Compressor) extractTarGz(source, target string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := pgzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	return c.extractTarReader(gzReader, target)
}

// extractTarBz2 解压 TAR.BZ2 文件
func (c *Compressor) extractTarBz2(source, target string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	bz2Reader := bzip2.NewReader(file)
	return c.extractTarReader(bz2Reader, target)
}

// extractTarXz 解压 TAR.XZ 文件
func (c *Compressor) extractTarXz(source, target string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	xzReader, err := xz.NewReader(file)
	if err != nil {
		return err
	}

	return c.extractTarReader(xzReader, target)
}

// extractTarReader 从 TAR reader 解压
func (c *Compressor) extractTarReader(reader io.Reader, target string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)

		// 安全检查
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(target)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// extract7z 解压 7z 文件
func (c *Compressor) extract7z(source, target string) error {
	reader, err := sevenzip.OpenReader(source)
	if err != nil {
		return fmt.Errorf("failed to open 7z file: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)

		// 安全检查
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(target)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// extractAuto 自动检测格式并解压
func (c *Compressor) extractAuto(source, target string) error {
	// 尝试读取文件头检测格式
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return err
	}
	buf = buf[:n]

	// 检测 ZIP
	if len(buf) >= 4 && buf[0] == 0x50 && buf[1] == 0x4B {
		return c.extractZip(source, target)
	}

	// 检测 GZIP
	if len(buf) >= 2 && buf[0] == 0x1F && buf[1] == 0x8B {
		return c.extractTarGz(source, target)
	}

	// 检测 BZIP2
	if len(buf) >= 3 && buf[0] == 0x42 && buf[1] == 0x5A && buf[2] == 0x68 {
		return c.extractTarBz2(source, target)
	}

	// 检测 XZ
	if len(buf) >= 6 && buf[0] == 0xFD && buf[1] == 0x37 && buf[2] == 0x7A && buf[3] == 0x58 && buf[4] == 0x5A && buf[5] == 0x00 {
		return c.extractTarXz(source, target)
	}

	// 检测 7z
	if len(buf) >= 6 && buf[0] == 0x37 && buf[1] == 0x7A && buf[2] == 0xBC && buf[3] == 0xAF && buf[4] == 0x27 && buf[5] == 0x1C {
		return c.extract7z(source, target)
	}

	// 尝试作为 TAR 处理
	return c.extractTar(source, target)
}
