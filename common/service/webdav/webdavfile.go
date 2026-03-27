package webdav2

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/net/webdav"
)

type WebDAVFile struct {
	webdav.File
	rootDir  string
	fileInfo *WebDAVFileInfo
	statOnce sync.Once
	statErr  error
}

const (
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

	return entries, nil
}

func NewWebDAVFile(file webdav.File, rootDir string) *WebDAVFile {
	return &WebDAVFile{File: file, rootDir: rootDir}
}

func (f *WebDAVFile) ensureStat() error {
	f.statOnce.Do(func() {
		stat, err := f.File.Stat()
		if err != nil {
			f.statErr = err
			return
		}
		f.fileInfo = &WebDAVFileInfo{FileInfo: stat}

		f.fileInfo.pem = fmt.Sprintf("%o", stat.Mode().Perm())
		sysstat, ok := stat.Sys().(*syscall.Stat_t)
		if ok {
			f.fileInfo.gid = fmt.Sprintf("%d", sysstat.Gid)
			f.fileInfo.uid = fmt.Sprintf("%d", sysstat.Uid)
		}

		isSymlink := stat.Mode()&syscall.S_IFMT == syscall.S_IFLNK
		if isSymlink {
			fullPath := filepath.Join(f.rootDir, stat.Name())
			resolvedTarget, err := filepath.EvalSymlinks(fullPath)
			if err == nil {
				resolvedTarget = filepath.Clean(resolvedTarget)
				rootDir := filepath.Clean(f.rootDir)
				if !strings.HasPrefix(resolvedTarget, rootDir) {
					slog.Warn("symlink escapes container root",
						"link", fullPath,
						"resolved", resolvedTarget,
						"root", rootDir)
					f.fileInfo.fileType = "file"
					f.fileInfo.editable = false
					return
				}
			}
		}

		f.fileInfo.fileType, f.fileInfo.editable = getFileTypeAndEditable(stat.Mode())
	})
	return f.statErr
}

func getFileTypeAndEditable(mode os.FileMode) (string, bool) {
	switch mode & syscall.S_IFMT {
	case syscall.S_IFREG:
		return "file", true
	case syscall.S_IFDIR:
		return "directory", false
	case syscall.S_IFLNK:
		return "symlink", true
	case syscall.S_IFBLK, syscall.S_IFCHR:
		return "device", false
	case syscall.S_IFIFO:
		return "fifo", false
	case syscall.S_IFSOCK:
		return "socket", false
	default:
		return "file", true
	}
}

func (f *WebDAVFile) DeadProps() (map[xml.Name]webdav.Property, error) {
	if err := f.ensureStat(); err != nil {
		return nil, err
	}
	ret := make(map[xml.Name]webdav.Property)
	info := f.fileInfo
	if info == nil {
		return ret, nil
	}

	ret[xml.Name{Local: "uid", Space: "w7panel"}] = webdav.Property{
		XMLName:  xml.Name{Local: "uid", Space: "w7panel"},
		InnerXML: []byte(info.uid),
	}
	ret[xml.Name{Local: "gid", Space: "w7panel"}] = webdav.Property{
		XMLName:  xml.Name{Local: "gid", Space: "w7panel"},
		InnerXML: []byte(info.gid),
	}
	ret[xml.Name{Local: "mode", Space: "w7panel"}] = webdav.Property{
		XMLName:  xml.Name{Local: "mode", Space: "w7panel"},
		InnerXML: []byte(info.pem),
	}
	ret[xml.Name{Local: "type", Space: "w7panel"}] = webdav.Property{
		XMLName:  xml.Name{Local: "type", Space: "w7panel"},
		InnerXML: []byte(info.fileType),
	}
	ret[xml.Name{Local: "editable", Space: "w7panel"}] = webdav.Property{
		XMLName:  xml.Name{Local: "editable", Space: "w7panel"},
		InnerXML: []byte(strconv.FormatBool(info.editable)),
	}

	return ret, nil
}

func (f *WebDAVFile) Patch(patches []webdav.Proppatch) ([]webdav.Propstat, error) {
	return []webdav.Propstat{{Props: nil}}, nil
}

func (f *WebDAVFile) Read(p []byte) (n int, err error) {
	if err := f.ensureStat(); err != nil {
		return 0, err
	}
	if !f.fileInfo.editable {
		return 0, nil
	}
	return f.File.Read(p)
}

func (f *WebDAVFile) Seek(offset int64, whence int) (n int64, err error) {
	if err := f.ensureStat(); err != nil {
		return 0, err
	}
	if !f.fileInfo.editable {
		return 0, nil
	}
	return f.File.Seek(offset, whence)
}

type WebDAVFileInfo struct {
	os.FileInfo
	uid      string
	gid      string
	pem      string
	editable bool
	fileType string
}

func (info WebDAVFileInfo) ContentType(ctx context.Context) (string, error) {
	if !info.Mode().IsRegular() {
		return "application/linux-" + info.fileType, nil
	}
	ctype := mime.TypeByExtension(filepath.Ext(info.Name()))
	if ctype != "" {
		return ctype, nil
	}
	return "application/octet-stream", nil // fallback
}
