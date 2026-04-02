package webdav2

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strings"

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
	return NewWebDAVFile(file, fs.dir, name), err
}

func (fs WebDAVFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	stat, err := fs.FileSystem.Stat(ctx, name)
	if err != nil {
		return stat, err
	}
	return stat, nil
}

func (fs WebDAVFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	err := fs.FileSystem.Rename(ctx, oldName, newName)
	if err != nil {
		if strings.Contains(err.Error(), "invalid cross-device link") {
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
