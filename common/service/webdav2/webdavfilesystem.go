package webdav2

import (
	"context"
	"log/slog"
	"os"

	"golang.org/x/net/webdav"
)

type WebDAVFileSystem struct {
	webdav.FileSystem
}

func (fs WebDAVFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	file, err := fs.FileSystem.OpenFile(ctx, name, flag, perm)
	if err != nil {
		slog.Error("webdav OpenFile failed", "name", name, "flag", flag, "perm", perm, "error", err)
		return nil, err
	}
	return &WebDAVFile{File: file}, err
}

// func (fs WebDAVFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
// 	return fs.FileSystem.Stat(ctx, name)
// }
// func (fs WebDAVFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
// 	return fs.FileSystem.Mkdir(ctx, name, perm)
// }
// func (fs WebDAVFileSystem) RemoveAll(ctx context.Context, name string) error {
// 	return fs.FileSystem.RemoveAll(ctx, name)
// }
// func (fs WebDAVFileSystem) Rename(ctx context.Context, oldName, newName string) error {
// 	return fs.FileSystem.Rename(ctx, oldName, newName)
// }

func NewWebDAVFileSystem(fs webdav.FileSystem) WebDAVFileSystem {
	return WebDAVFileSystem{fs}
}
