package webdav

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/net/webdav"
)

const (
	MaxFileSize   = 50 * 1024 * 1024
	MaxDirEntries = 5000
)

var specialDirs = map[string]bool{
	"/proc": true,
	"/sys":  true,
	"/dev":  true,
	"/run":  true,
}

func isSpecialDir(path string) bool {
	cleanPath := filepath.Clean(path)
	for dir := range specialDirs {
		if cleanPath == dir || strings.HasPrefix(cleanPath, dir+"/") {
			return true
		}
	}
	return false
}

type OFile struct {
	webdav.File
	UserGroup     *UserGroup
	mu            sync.Mutex
	buffer        *bytes.Reader
	readError     error
	once          sync.Once
	filePath      string
	isSymlink     bool
	symlinkTarget string
}

func (n *OFile) tryReadToBuffer() {
	n.once.Do(func() {
		content, err := io.ReadAll(n.File)
		if err != nil {
			n.readError = err
			return
		}
		if len(content) > MaxFileSize {
			n.readError = fmt.Errorf("file too large (max %dMB)", MaxFileSize/1024/1024)
			return
		}
		n.buffer = bytes.NewReader(content)
	})
}

func (n *OFile) DeadProps() (map[xml.Name]webdav.Property, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	user := webdav.Property{
		XMLName:  xml.Name{Local: "user", Space: "w7panel"},
		InnerXML: []byte("unknown"),
	}
	group := webdav.Property{
		XMLName:  xml.Name{Local: "group", Space: "w7panel"},
		InnerXML: []byte("unknown"),
	}
	perm := webdav.Property{
		XMLName:  xml.Name{Local: "mode", Space: "w7panel"},
		InnerXML: []byte("unknown"),
	}
	symlink := webdav.Property{
		XMLName:  xml.Name{Local: "is_symlink", Space: "w7panel"},
		InnerXML: []byte("false"),
	}
	symlinkTarget := webdav.Property{
		XMLName:  xml.Name{Local: "symlink_target", Space: "w7panel"},
		InnerXML: []byte(""),
	}

	ret := make(map[xml.Name]webdav.Property)
	stat, err := n.File.Stat()
	if err != nil {
		return ret, err
	}
	filestat, ok := stat.Sys().(*syscall.Stat_t)
	if ok {
		user.InnerXML = []byte(fmt.Sprintf("%d", filestat.Uid))
		group.InnerXML = []byte(fmt.Sprintf("%d", filestat.Gid))

		groupName, err := n.UserGroup.GetGroupName(int(filestat.Gid))
		if err == nil {
			group.InnerXML = []byte(groupName)
		}
		userName, err := n.UserGroup.GetUserName(int(filestat.Uid))
		if err == nil {
			user.InnerXML = []byte(userName)
		}
		perm.InnerXML = []byte(fmt.Sprintf("%o", stat.Mode().Perm()))
	}

	if n.isSymlink {
		symlink.InnerXML = []byte("true")
		if n.symlinkTarget != "" {
			symlinkTarget.InnerXML = []byte(n.symlinkTarget)
		}
	}

	ret[user.XMLName] = user
	ret[group.XMLName] = group
	ret[perm.XMLName] = perm
	ret[symlink.XMLName] = symlink
	ret[symlinkTarget.XMLName] = symlinkTarget
	return ret, nil
}

func (f *OFile) Patch(patches []webdav.Proppatch) ([]webdav.Propstat, error) {
	return []webdav.Propstat{{Props: nil}}, nil
}

func (f *OFile) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	stat, err := f.File.Stat()
	if err != nil {
		return 0, err
	}

	if stat.IsDir() {
		return f.File.Seek(offset, whence)
	}

	f.tryReadToBuffer()
	if f.readError != nil {
		return 0, f.readError
	}

	if f.buffer != nil {
		return f.buffer.Seek(offset, whence)
	}

	return f.File.Seek(offset, whence)
}

func (f *OFile) Read(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	stat, err := f.File.Stat()
	if err != nil {
		return 0, err
	}

	if stat.IsDir() {
		return f.File.Read(p)
	}

	f.tryReadToBuffer()
	if f.readError != nil {
		return 0, f.readError
	}

	if f.buffer != nil {
		return f.buffer.Read(p)
	}

	return f.File.Read(p)
}

func (f *OFile) Readdir(count int) ([]os.FileInfo, error) {
	if count > 0 && count > MaxDirEntries {
		count = MaxDirEntries
	}

	entries, err := f.File.Readdir(count)
	if err != nil {
		return nil, err
	}

	if len(entries) > MaxDirEntries {
		slog.Warn("directory entries truncated", "path", f.filePath, "total", len(entries), "max", MaxDirEntries)
		entries = entries[:MaxDirEntries]
	}

	return entries, nil
}

type PermissionFileSystem struct {
	webdav.FileSystem
	UserGroup *UserGroup
	RootDir   string
}

func (fs *PermissionFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	cleanName := filepath.Clean("/" + name)

	file, err := fs.FileSystem.OpenFile(ctx, name, flag, perm)
	if err != nil {
		if isSpecialDir(filepath.Dir(cleanName)) && os.IsNotExist(err) {
			slog.Debug("special directory access", "path", name, "error", err)
		}
		return nil, err
	}

	ofile := &OFile{
		File:      file,
		UserGroup: fs.UserGroup,
		filePath:  cleanName,
	}

	if fs.RootDir != "" && flag == os.O_RDONLY {
		fullPath := filepath.Join(fs.RootDir, name)
		if err := fs.handleSpecialFile(ofile, fullPath); err != nil {
			slog.Debug("special file handling", "path", fullPath, "error", err)
		}
	}

	return ofile, nil
}

func (fs *PermissionFileSystem) handleSpecialFile(ofile *OFile, fullPath string) error {
	fi, err := os.Lstat(fullPath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return nil
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		ofile.isSymlink = true
		target, err := os.Readlink(fullPath)
		if err != nil {
			return err
		}
		ofile.symlinkTarget = target
		content, err := fs.readLinkTarget(fullPath, target)
		if err != nil {
			return err
		}
		ofile.buffer = bytes.NewReader(content)
		ofile.once = sync.Once{}
		return nil
	}

	if isSpecialDir(filepath.Dir(fullPath)) {
		content, err := fs.readSpecialFileContent(fullPath)
		if err != nil {
			slog.Debug("failed to read special file", "path", fullPath, "error", err)
			return err
		}
		ofile.buffer = bytes.NewReader(content)
		ofile.once = sync.Once{}
	}

	return nil
}

func (fs *PermissionFileSystem) readLinkTarget(fullPath, target string) ([]byte, error) {
	if isSpecialDir(filepath.Dir(target)) {
		return fs.readSpecialFileContent(target)
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(fullPath), target)
	}

	targetInfo, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	if targetInfo.IsDir() {
		return nil, os.ErrInvalid
	}

	return fs.readSpecialFileContent(target)
}

func (fs *PermissionFileSystem) readSpecialFileContent(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	limitedReader := &io.LimitedReader{R: file, N: MaxFileSize}
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (fs *PermissionFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	fi, err := fs.FileSystem.Stat(ctx, name)
	if err != nil {
		return nil, err
	}

	if sysStat, ok := fi.Sys().(*syscall.Stat_t); ok {
		isSymlink := sysStat.Mode&syscall.S_IFMT == syscall.S_IFLNK
		var symlinkTarget string
		if isSymlink && fs.RootDir != "" {
			fullPath := filepath.Join(fs.RootDir, name)
			symlinkTarget, _ = os.Readlink(fullPath)
		}
		return &fileInfoWithPermission{fi, sysStat, symlinkTarget, isSymlink}, nil
	}

	defaultStat := &syscall.Stat_t{
		Uid: uint32(os.Getuid()),
		Gid: uint32(os.Getgid()),
	}
	return &fileInfoWithPermission{fi, defaultStat, "", false}, nil
}

type fileInfoWithPermission struct {
	os.FileInfo
	stat          *syscall.Stat_t
	symlinkTarget string
	isSymlink     bool
}

func (fi *fileInfoWithPermission) Sys() interface{} {
	return map[string]interface{}{
		"mode":           fi.stat.Mode,
		"uid":            fi.stat.Uid,
		"gid":            fi.stat.Gid,
		"user":           getUserName(fi.stat.Uid),
		"group":          getGroupName(fi.stat.Gid),
		"uid_gid":        fmt.Sprintf("%d:%d", fi.stat.Uid, fi.stat.Gid),
		"is_symlink":     fi.isSymlink,
		"symlink_target": fi.symlinkTarget,
	}
}

func getUserName(uid uint32) string {
	return fmt.Sprintf("%d", uid)
}

func getGroupName(gid uint32) string {
	return fmt.Sprintf("%d", gid)
}

func IsSpecialDirectory(path string) bool {
	return isSpecialDir(path)
}

func GetMaxFileSize() int64 {
	return MaxFileSize
}

func GetMaxDirEntries() int {
	return MaxDirEntries
}
