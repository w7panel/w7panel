package webdav2

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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
			slog.Warn("webdav Rename failed, trying copy+remove fallback", "oldName", oldName, "newName", newName, "error", err)

			srcPath := fs.dir + oldName
			dstPath := fs.dir + newName

			if err := copyPath(srcPath, dstPath); err != nil {
				slog.Error("copy fallback failed", "src", srcPath, "dst", dstPath, "error", err)
				return err
			}

			if err := os.RemoveAll(srcPath); err != nil {
				slog.Error("remove fallback failed", "src", srcPath, "error", err)
				return err
			}

			slog.Info("webdav Rename copy+remove fallback success", "oldName", oldName, "newName", newName)
			return nil
		}
		slog.Error("webdav Rename failed", "oldName", oldName, "newName", newName, "error", err)
	}
	return err
}

func copyPath(srcPath, dstPath string) error {
	info, err := os.Lstat(srcPath)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(srcPath)
		if err != nil {
			return err
		}
		_ = os.Remove(dstPath)
		return os.Symlink(target, dstPath)
	}

	if info.IsDir() {
		return copyDir(srcPath, dstPath, info.Mode().Perm())
	}

	return copyFile(srcPath, dstPath, info.Mode())
}

func copyDir(srcPath, dstPath string, perm os.FileMode) error {
	if err := os.MkdirAll(dstPath, perm); err != nil {
		return err
	}

	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcChild := filepath.Join(srcPath, entry.Name())
		dstChild := filepath.Join(dstPath, entry.Name())
		if err := copyPath(srcChild, dstChild); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(srcPath, dstPath string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dstPath, mode.Perm())
}

func NewWebDAVFileSystem(fs webdav.FileSystem, dir string) WebDAVFileSystem {
	return WebDAVFileSystem{fs, dir}
}
