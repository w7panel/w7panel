package webdav2

import (
	"context"
	"log/slog"
	"os"
	"os/exec"

	"golang.org/x/net/webdav"
)

type WebDAVFileSystem struct {
	webdav.FileSystem
	dir string
}

func (fs WebDAVFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	file, err := fs.FileSystem.OpenFile(ctx, name, flag, perm)
	if err != nil {
		slog.Error("webdav OpenFile failed", "name", name, "flag", flag, "perm", perm, "error", err)
		return nil, err
	}
	return NewWebDAVFile(file), err
}

func (fs WebDAVFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	stat, err := fs.FileSystem.Stat(ctx, name)
	if err != nil {
		return stat, err
	}
	// isSymlink := stat.Mode()&syscall.S_IFMT == syscall.S_IFLNK
	// fullPath := filepath.Join(fs.dir, stat.Name())
	// resolvedTarget, err := filepath.EvalSymlinks(fullPath)
	// symlinkTarget := ""
	// if err == nil {
	// 	resolvedTarget = filepath.Clean(resolvedTarget)
	// 	rootDir := filepath.Clean(fs.dir)
	// 	if strings.HasPrefix(resolvedTarget, rootDir) {
	// 		symlinkTarget, _ = os.Readlink(fullPath)
	// 	} else {
	// 		slog.Warn("symlink escapes container root in Stat - blocked",
	// 			"link", fullPath,
	// 			"resolved", resolvedTarget,
	// 			"root", rootDir)
	// 	}
	// }
	// filestat, ok := stat.(*WebDAVFileInfo)
	// if ok {
	// 	filestat.isSymlink = isSymlink
	// 	filestat.symlinkTarget = symlinkTarget
	// 	filestat.pem = fmt.Sprintf("%d", stat.Mode().Perm())
	// 	syscall, ok := stat.Sys().(*syscall.Stat_t)
	// 	if ok {
	// 		filestat.uid = fmt.Sprintf("%d", syscall.Uid)
	// 		filestat.gid = fmt.Sprintf("%d", syscall.Gid)
	// 	}
	// 	return filestat, nil
	// }
	return stat, nil

}

//	func (fs WebDAVFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
//		return fs.FileSystem.Mkdir(ctx, name, perm)
//	}
//
//	func (fs WebDAVFileSystem) RemoveAll(ctx context.Context, name string) error {
//		return fs.FileSystem.RemoveAll(ctx, name)
//	}
func (fs WebDAVFileSystem) Rename(ctx context.Context, oldName, newName string) error {
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
func NewWebDAVFileSystem(fs webdav.FileSystem, dir string) WebDAVFileSystem {
	return WebDAVFileSystem{fs, dir}
}
