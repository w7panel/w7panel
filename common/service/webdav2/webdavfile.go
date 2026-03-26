package webdav2

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/samber/lo"
	"golang.org/x/net/webdav"
)

type WebDAVFile struct {
	webdav.File
	fileInfo *WebDAVFileInfo
	rootDir  string
}

const (
	MaxFileSize   = 50 * 1024 * 1024
	MaxDirEntries = 5000
)

func (f *WebDAVFile) Readdir(count int) ([]os.FileInfo, error) {
	if count > 0 && count > MaxDirEntries {
		count = MaxDirEntries
	}

	entries, err := f.File.Readdir(count)
	if err != nil {
		return nil, err
	}
	filters := lo.Filter(entries, func(info os.FileInfo, index int) bool {
		m := info.Mode()
		if !m.IsRegular() && !m.IsDir() {
			return false
		}
		return info.Name() != "ptmx"
	})

	return filters, nil
}

func NewWebDAVFile(file webdav.File, rootDir string) *WebDAVFile {
	r := &WebDAVFile{File: file, rootDir: rootDir}
	//测试发现 写入文件内容  权限和所有者不会变更
	//如果要加 r.stat记录下 权限所有者
	// webdav.File 的Close方法中恢复权限和所有者
	return r
}

type WebDAVFileInfo struct {
	os.FileInfo
	isSymlink     bool
	symlinkTarget string
	uid           string
	gid           string
	pem           string
	ext           string
}

// 不读文件 获取content-type
func (info WebDAVFileInfo) ContentType(ctx context.Context) (string, error) {
	ctype := mime.TypeByExtension(info.ext)
	if ctype != "" {
		return ctype, nil
	}
	return "application/octet-stream", nil
}

func (f *WebDAVFile) DeadProps() (map[xml.Name]webdav.Property, error) {
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	ret := make(map[xml.Name]webdav.Property)
	info, ok := stat.(*WebDAVFileInfo)
	if !ok {
		return ret, nil
	}
	user := webdav.Property{
		XMLName:  xml.Name{Local: "uid", Space: "w7panel"},
		InnerXML: []byte(info.uid),
	}
	group := webdav.Property{
		XMLName:  xml.Name{Local: "gid", Space: "w7panel"},
		InnerXML: []byte(info.gid),
	}
	perm := webdav.Property{
		XMLName:  xml.Name{Local: "mode", Space: "w7panel"},
		InnerXML: []byte(info.pem),
	}
	symlinkStr := "false"
	if info.isSymlink {
		symlinkStr = "true"
	}
	symlink := webdav.Property{
		XMLName:  xml.Name{Local: "is_symlink", Space: "w7panel"},
		InnerXML: []byte(symlinkStr),
	}
	symlinkTarget := webdav.Property{
		XMLName:  xml.Name{Local: "symlink_target", Space: "w7panel"},
		InnerXML: []byte(info.symlinkTarget),
	}
	// filestat, ok := stat.Sys().(*syscall.Stat_t)
	// if ok {
	// 	user.InnerXML = []byte(fmt.Sprintf("%d", filestat.Uid))
	// 	group.InnerXML = []byte(fmt.Sprintf("%d", filestat.Gid))
	// 	perm.InnerXML = []byte(fmt.Sprintf("%o", stat.Mode().Perm()))
	// }
	ret[user.XMLName] = user
	ret[group.XMLName] = group
	ret[perm.XMLName] = perm
	ret[symlink.XMLName] = symlink
	ret[symlinkTarget.XMLName] = symlinkTarget
	return ret, nil
}
func (n *WebDAVFile) Patch(patches []webdav.Proppatch) ([]webdav.Propstat, error) {

	return []webdav.Propstat{{Props: nil}}, nil
}
func (n *WebDAVFile) Stat() (os.FileInfo, error) {
	if n.fileInfo != nil {
		return n.fileInfo, nil
	}
	stat, err := n.File.Stat()
	if err != nil {
		return nil, err
	}
	n.fileInfo = &WebDAVFileInfo{FileInfo: stat}

	isSymlink := stat.Mode()&syscall.S_IFMT == syscall.S_IFLNK
	fullPath := filepath.Join(n.rootDir, stat.Name())
	resolvedTarget, err := filepath.EvalSymlinks(fullPath)
	symlinkTarget := ""
	if err == nil {
		resolvedTarget = filepath.Clean(resolvedTarget)
		rootDir := filepath.Clean(n.rootDir)
		if strings.HasPrefix(resolvedTarget, rootDir) {
			symlinkTarget, _ = os.Readlink(fullPath)
		} else {
			slog.Warn("symlink escapes container root in Stat - blocked",
				"link", fullPath,
				"resolved", resolvedTarget,
				"root", rootDir)
		}
	}
	n.fileInfo.isSymlink = isSymlink
	n.fileInfo.symlinkTarget = symlinkTarget
	n.fileInfo.pem = fmt.Sprintf("%o", stat.Mode().Perm())
	n.fileInfo.ext = filepath.Ext(n.fileInfo.Name())
	sysstat, ok := stat.Sys().(*syscall.Stat_t)
	if ok {
		n.fileInfo.gid = fmt.Sprintf("%d", sysstat.Gid)
		n.fileInfo.uid = fmt.Sprintf("%d", sysstat.Uid)

	}
	return n.fileInfo, nil
}
