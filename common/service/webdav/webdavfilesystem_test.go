package webdav2

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"golang.org/x/net/webdav"
)

type renameEXDEVFS struct {
	webdav.Dir
}

func (fs renameEXDEVFS) Rename(ctx context.Context, oldName, newName string) error {
	return &os.LinkError{Op: "rename", Old: oldName, New: newName, Err: syscall.EXDEV}
}

func TestWebDAVFileSystemRename_FallbackCopiesAndRemovesFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "webdav-rename-file")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	srcRel := "/src.txt"
	dstRel := "/dst.txt"
	srcAbs := filepath.Join(tmpDir, "src.txt")
	dstAbs := filepath.Join(tmpDir, "dst.txt")

	if err := os.WriteFile(srcAbs, []byte("hello"), 0o640); err != nil {
		t.Fatal(err)
	}

	fs := NewWebDAVFileSystem(renameEXDEVFS{Dir: webdav.Dir(tmpDir)}, tmpDir)
	if err := fs.Rename(context.Background(), srcRel, dstRel); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(srcAbs); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source should be removed, got err=%v", err)
	}
	data, err := os.ReadFile(dstAbs)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected destination content: %q", string(data))
	}
	info, err := os.Stat(dstAbs)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("unexpected destination perm: %o", info.Mode().Perm())
	}
}

func TestWebDAVFileSystemRename_FallbackCopiesAndRemovesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "webdav-rename-dir")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	srcDirRel := "/srcdir"
	dstDirRel := "/dstdir"
	srcDirAbs := filepath.Join(tmpDir, "srcdir")
	dstDirAbs := filepath.Join(tmpDir, "dstdir")

	if err := os.MkdirAll(filepath.Join(srcDirAbs, "nested"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDirAbs, "nested", "file.txt"), []byte("nested"), 0o600); err != nil {
		t.Fatal(err)
	}

	fs := NewWebDAVFileSystem(renameEXDEVFS{Dir: webdav.Dir(tmpDir)}, tmpDir)
	if err := fs.Rename(context.Background(), srcDirRel, dstDirRel); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(srcDirAbs); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source dir should be removed, got err=%v", err)
	}
	data, err := os.ReadFile(filepath.Join(dstDirAbs, "nested", "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "nested" {
		t.Fatalf("unexpected nested content: %q", string(data))
	}
}

func TestWebDAVFileSystemRename_FallbackCopiesSymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "webdav-rename-symlink")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	targetAbs := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(targetAbs, []byte("target"), 0o644); err != nil {
		t.Fatal(err)
	}

	srcRel := "/link"
	dstRel := "/link2"
	srcAbs := filepath.Join(tmpDir, "link")
	dstAbs := filepath.Join(tmpDir, "link2")
	if err := os.Symlink("target.txt", srcAbs); err != nil {
		t.Fatal(err)
	}

	fs := NewWebDAVFileSystem(renameEXDEVFS{Dir: webdav.Dir(tmpDir)}, tmpDir)
	if err := fs.Rename(context.Background(), srcRel, dstRel); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(srcAbs); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source symlink should be removed, got err=%v", err)
	}
	info, err := os.Lstat(dstAbs)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("destination should remain symlink")
	}
	target, err := os.Readlink(dstAbs)
	if err != nil {
		t.Fatal(err)
	}
	if target != "target.txt" {
		t.Fatalf("unexpected symlink target: %s", target)
	}
}
