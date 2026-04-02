package webdav2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFileTypeAndEditable(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		wantType string
		wantEdit bool
	}{
		{name: "regular file", mode: 0, wantType: "file", wantEdit: true},
		{name: "directory", mode: os.ModeDir, wantType: "directory", wantEdit: false},
		{name: "symlink", mode: os.ModeSymlink, wantType: "symlink", wantEdit: true},
		{name: "char device", mode: os.ModeDevice | os.ModeCharDevice, wantType: "device", wantEdit: false},
		{name: "fifo", mode: os.ModeNamedPipe, wantType: "fifo", wantEdit: false},
		{name: "socket", mode: os.ModeSocket, wantType: "socket", wantEdit: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotEdit := getFileTypeAndEditable(tt.mode)
			if gotType != tt.wantType || gotEdit != tt.wantEdit {
				t.Fatalf("type/editable mismatch: got=(%s,%v), want=(%s,%v)", gotType, gotEdit, tt.wantType, tt.wantEdit)
			}
		})
	}
}

func TestWebDAVFileEnsureStat_UsesReqPathForLstat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "webdav-file-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	targetPath := filepath.Join(tmpDir, "dir", "child.txt")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetPath, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })

	wf := NewWebDAVFile(f, tmpDir, "/dir/child.txt")
	if err := wf.ensureStat(); err != nil {
		t.Fatal(err)
	}
	if wf.fileInfo == nil {
		t.Fatal("fileInfo should not be nil")
	}
	if wf.fileInfo.fileType != "file" || !wf.fileInfo.editable {
		t.Fatalf("unexpected file type/editable: got=(%s,%v)", wf.fileInfo.fileType, wf.fileInfo.editable)
	}
}
